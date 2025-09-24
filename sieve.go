package main

import (
	"image"
	"math/rand/v2"
	"os"
	"strconv"
	"sync"
	"time"

	"fortio.org/terminal/ansipixels"
	"fortio.org/terminal/ansipixels/tcolor"
)

func randomColor() tcolor.RGBColor {
	return tcolor.HSLToRGB(rand.Float64(), 0.5, 0.5)
}

type update struct {
	num   int
	color tcolor.RGBColor
}

func main() {
	ap := ansipixels.NewAnsiPixels(0)
	mut := &sync.Mutex{}

	ap.Open()
	ap.HideCursor()
	defer func() {
		ap.MouseClickOff()
		ap.MouseTrackingOff()
		ap.ShowCursor()
		ap.Restore()
	}()
	numbersByColor := make(map[int]*tcolor.RGBColor)
	updatedAt := make(map[int]uint64)
	frame := uint64(0)
	updateChan := make(chan update)
	img := image.NewRGBA(image.Rect(0, 0, ap.W, ap.H*2))
	go func() {
		for i := 2; i < 101; i++ {

			if numbersByColor[i] != nil {
				continue
			}
			time.Sleep(100 * time.Duration(i) * time.Millisecond)
			go func() {
				color := randomColor()
				for j := i * 2; j < 101; j += i {
					mut.Lock()
					numbersByColor[j] = &color
					mut.Unlock()
					time.Sleep(100 * time.Duration(i) * time.Millisecond)
					updateChan <- update{j, color}
				}
			}()
		}
		switch {
		}
	}()
	ap.ClearScreen()
	// if ap.Color256 {
	ap.Draw216ColorImage(0, 0, img)
	// }
	// if ap.TrueColor {
	// ap.DrawTrueColorImage(0, 0, img)
	// }
	for i := range 10 {
		for j := range 10 {
			num := i*10 + j
			ap.WriteAtStr(ap.W*j/10, ap.H*i/10, ap.Background.Background()+strconv.Itoa(num+1))
		}
	}
	go func() {
		for {
			ap.ReadOrResizeOrSignalOnce()
			if len(ap.Data) > 0 && ap.Data[0] == 'q' {
				os.Exit(1)
			}
		}
	}()
	go func() {
		for update := range updateChan {
			mut.Lock()
			numString := strconv.Itoa(update.num - 1)

			updatedAt[update.num] = frame
			if len(numString) == 1 {
				numString = "0" + numString
			}
			i, err := strconv.Atoi(numString[0:1])
			if err != nil {
				panic("bad update sent")
			}
			j, err := strconv.Atoi(numString[1:])
			if err != nil {
				panic("bad update sent")
			}
			ap.StartSyncMode()

			ap.WriteAtStr(ap.W*j/10, ap.H*i/10, update.color.Foreground()+strconv.Itoa(update.num))
			ap.EndSyncMode()
			mut.Unlock()
		}
	}()

	for {
		mut.Lock()
		for num := range 101 {
			if updatedAt[num] == 0 {
				continue
			}
			clr := numbersByColor[num]
			timeSince := frame - updatedAt[num]
			if timeSince%100 != 0 {
				continue
			}
			alpha := 1 - float64((min(float64(timeSince)/5000, 255) / 255.))
			newClr := ansipixels.BlendLinear(ap.Background, *clr, alpha)
			numString := strconv.Itoa(num - 1)

			if len(numString) == 1 {
				numString = "0" + numString
			}
			i, err := strconv.Atoi(numString[0:1])
			if err != nil {
				panic("bad update sent")
			}
			j, err := strconv.Atoi(numString[1:])
			if err != nil {
				panic("bad update sent")
			}
			ap.StartSyncMode()
			ap.WriteFg(newClr.Color())
			// ap.WriteBg(ap.Background.Color())
			ap.WriteAtStr(ap.W*j/10, ap.H*i/10, strconv.Itoa(num))
			// ap.WriteAtStr(ap.W*j/10, ap.H*i/10, newClr.Color().Foreground()+ap.Background.Background()+strconv.Itoa(num))
			ap.EndSyncMode()
		}
		frame++
		mut.Unlock()

	}

}
