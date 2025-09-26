package main

import (
	"fmt"
	"math/rand/v2"
	"slices"
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

type queueNode struct {
	multiple, prime int
}

type minHeap []queueNode

func (m *minHeap) push(qn queueNode) {
	*m = append(*m, qn)
	cur := len(*m) - 1
	for cur > 0 && (*m)[cur].multiple < (*m)[(cur-1)/2].multiple {
		(*m)[cur], (*m)[(cur-1)/2] = (*m)[(cur-1)/2], (*m)[cur]
	}
}

func (m *minHeap) pop() queueNode {
	q := (*m)[0]
	(*m)[0] = (*m)[len(*m)-1]
	*m = (*m)[:len(*m)-1]
	cur := 0
	if len(*m) <= 1 {
		return q
	}
	if len(*m) == 2 {
		if (*m)[0].multiple > (*m)[1].multiple {
			(*m)[0], (*m)[1] = (*m)[1], (*m)[0]
		}
		return q
	}
	for (*m)[cur].multiple > (*m)[cur*2+1].multiple || (*m)[cur].multiple > (*m)[cur*2+2].multiple {
		if (*m)[cur*2+1].multiple > (*m)[cur*2+2].multiple {
			(*m)[cur], (*m)[cur*2+2] = (*m)[cur*2+2], (*m)[cur]
			cur = cur*2 + 2
		} else {
			(*m)[cur], (*m)[cur*2+1] = (*m)[cur*2+1], (*m)[cur]
			cur = cur*2 + 1
		}
		if cur*2+1 >= len(*m) || cur*2+2 >= len(*m) {
			break
		}
	}
	if cur*2+1 < len(*m) && (*m)[cur].multiple > (*m)[cur*2+1].multiple {
		(*m)[cur], (*m)[cur*2+1] = (*m)[cur*2+1], (*m)[cur]
	}

	return q
}

func simpleSieve(max int) {
	primes := make([]int, 0, max)
	composites := make(map[int]bool)

	for i := 2; i < max; i++ {
		if composites[i] {
			continue
		}
		primes = append(primes, i)
		for j := i * 2; j < max; j += i {
			composites[j] = true
		}
	}
	fmt.Println(primes)
	fmt.Println(len(primes))
}

func getPrimesAlternating(max int) {
	primesByCurMultiple := make(map[int]int)
	composites := make(map[int]bool)
	heap := minHeap{{4, 2}}
	visited := make(map[int]bool)
	var cur queueNode
	for len(heap) > 0 {
		cur = heap.pop()
		visited[cur.prime] = true
		primesByCurMultiple[cur.prime] = cur.multiple
		composites[cur.multiple] = true
		cur.multiple += cur.prime
		if cur.multiple < max {
			heap.push(cur)
		} else {
			primesByCurMultiple[cur.prime] = cur.multiple
		}
		for i := 2; i < max; i++ {
			primesBelowCheckAbove := true
			if composites[i] || visited[i] {
				continue
			}
			for p := range visited {
				if primesByCurMultiple[p] < i && p < i/2 {
					primesBelowCheckAbove = false
					break
				}
			}
			if primesBelowCheckAbove {
				heap.push(queueNode{multiple: i * 2, prime: i})
				visited[i] = true
				// fmt.Println(i)
				break
			}
		}
	}
	primes := make([]int, 0, len(visited))
	for i := range visited {
		primes = append(primes, i)
	}
	slices.Sort(primes)
	fmt.Println(primes)
}

func main() {
	ap := ansipixels.NewAnsiPixels(100)
	mut := &sync.Mutex{}

	ap.Open()
	ap.HideCursor()
	defer func() {
		ap.MouseClickOff()
		ap.MouseTrackingOff()
		ap.ShowCursor()
		ap.Restore()
	}()
	numbersByColor := [101]*tcolor.RGBColor{}
	updatedAt := ([101]uint64{})
	// curMultiple := update{}
	// primesByCurMultiple := make(map[int]int)
	frame := uint64(0)
	updateChan := make(chan update)
	go func() {
		for i := 2; i < 101; i++ {

			mut.Lock()
			if numbersByColor[i] != nil {
				mut.Unlock()
				continue
			}
			mut.Unlock()
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
	}()
	ap.ClearScreen()

	for i := range 10 {
		for j := range 10 {
			num := i*10 + j
			ap.WriteAtStr(ap.W*j/10, ap.H*i/10, strconv.Itoa(num+1))
		}
	}
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

			ap.WriteAt(ap.W*j/10, ap.H*i/10, "%s%s", update.color.Foreground(), strconv.Itoa(update.num))
			ap.EndSyncMode()
			mut.Unlock()
		}
	}()
	go ap.ReadOrResizeOrSignal()
	for {
		// _, err := ap.ReadOrResizeOrSignalOnce()
		// if err != nil {
		// 	panic(err)
		// }

		if len(ap.Data) > 0 && ap.Data[0] == 'q' {
			return
		}
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
			alpha := 1 - float64((min(float64(timeSince)/5000., 255.) / 255.))
			newClr := ansipixels.BlendLuminance(ap.Background, *clr, alpha)
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
			// ap.WriteFg(newClr.Color())
			fg := ap.ColorOutput.Foreground(newClr.Color())
			// ap.WriteBg(ap.Background.Color())
			ap.WriteAt(ap.W*j/10, ap.H*i/10, "%s%s", fg, strconv.Itoa(num))
			ap.WriteAt(ap.W*j/10, ap.H*i/10, "%s%s", newClr.Color().Foreground(), strconv.Itoa(num))
			ap.EndSyncMode()
		}
		frame++
		mut.Unlock()

	}

}
