package main

import (
	"math/rand/v2"
	"os"
	"strconv"
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

	ap.Open()

	numbersByColor := make(map[int]*tcolor.RGBColor)
	updateChan := make(chan update)
	go func() {
		for i := 2; i < 100; i++ {

			if numbersByColor[i] != nil {
				continue
			}
			color := randomColor()
			for j := i * 2; j < 100; j += i {
				numbersByColor[j] = &color
				time.Sleep(50 * time.Millisecond)
				updateChan <- update{j, color}
			}
		}
	}()
	ap.ClearScreen()
	for i := range 10 {
		for j := range 10 {
			num := i*10 + j
			ap.WriteAtStr(ap.W*j/10, ap.H*i/10, strconv.Itoa(num))
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
	for {
		for update := range updateChan {
			numString := strconv.Itoa(update.num)
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
		}
	}

}
