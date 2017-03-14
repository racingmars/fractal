package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"math/cmplx"
	"os"
	"runtime"
	"sync"
	"time"
)

const iterations = 1000
const width = 1600
const useHistogramMethod bool = true

func main() {
	heighttmp := width / 1.75
	height := int(heighttmp)
	histo := make([]int, iterations)
	histodata := make(chan int, 100)
	fractal := make([][]int, width)
	for i := range fractal {
		fractal[i] = make([]int, height)
	}

	img := image.NewGray16(image.Rect(0, 0, width, height))

	workers := runtime.NumCPU()
	runtime.GOMAXPROCS(workers)

	rows := make(chan int, workers)
	var workerWg, histoWg sync.WaitGroup

	histoWg.Add(1)
	go buildHisto(histo, histodata, &histoWg)

	start := time.Now().UTC().UnixNano()
	for w := 0; w < workers; w++ {
		workerWg.Add(1)
		go worker(fractal, width, height, rows, histodata, &workerWg)
	}
	for y := 0; y < height; y++ {
		rows <- y
	}
	close(rows)
	workerWg.Wait()
	stop := time.Now().UTC().UnixNano()
	runtime := stop - start
	fmt.Println("Time to run:", float32(runtime)/1000000000)

	close(histodata)
	histoWg.Wait()

	total := 0
	for _, count := range histo {
		total += count
	}

	colors := make([]float64, iterations)
	for i := 0; i < iterations; i++ {
		color := 0.0
		for j := 0; j < i; j++ {
			color += float64(histo[j]) / float64(total)
		}
		colors[i] = color
	}

	for x := range fractal {
		for y := range fractal[x] {
			if fractal[x][y] < 0 {
				img.Set(x, y, color.Black)
			} else {
				if !useHistogramMethod {
					gray := float64(fractal[x][y]) / float64(iterations) * float64(math.MaxUint16)
					img.SetGray16(x, y, color.Gray16{uint16(math.MaxUint16 - gray)})
				} else {
					img.SetGray16(x, y, color.Gray16{uint16(colors[fractal[x][y]] * math.MaxUint16)})
				}
			}
		}
	}

	f, err := os.Create("image.png")
	if err != nil {
		log.Fatal(err)
	}

	if err := png.Encode(f, img); err != nil {
		f.Close()
		log.Fatal(err)
	}

	f.Close()
}

func worker(img [][]int, width, height int, row, histodata chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for y := range row {
		for x := 0; x < width; x++ {
			r := float64(x)/(float64(width)-1)*3.5 - 2.5
			i := float64(y)/(float64(height)-1)*2 - 1
			in, iter := inSet(complex(r, i))
			if in {
				img[x][y] = -1
			} else {
				img[x][y] = iter
				histodata <- iter
			}
		}
	}
}

func inSet(c complex128) (bool, int) {
	result := complex128(0)
	for i := 0; i < iterations; i++ {
		result = cmplx.Pow(result, 2) + c
		if cmplx.Abs(result) > 2 {
			return false, i
		}
	}
	return true, 0
}

func buildHisto(histo []int, data chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := range data {
		histo[i]++
	}
}
