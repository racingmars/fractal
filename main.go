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
	"time"
)

const iterations = 256
const width = 1600

func main() {
	heighttmp := width / 1.75
	height := int(heighttmp)
	img := image.NewGray(image.Rect(0, 0, width, height))

	workers := runtime.NumCPU()
	runtime.GOMAXPROCS(workers)

	rows := make(chan int, workers)
	done := make(chan bool, workers)

	start := time.Now().UTC().UnixNano()
	for w := 0; w < workers; w++ {
		go worker(img, width, height, rows, done)
	}
	for y := 0; y < height; y++ {
		rows <- y
	}
	close(rows)
	for w := 0; w < workers; w++ {
		<-done
	}
	stop := time.Now().UTC().UnixNano()
	runtime := stop - start
	fmt.Println("Time to run:", float32(runtime)/1000000000)

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

func worker(img *image.Gray, width, height int, row chan int, done chan bool) {
	for y := range row {
		for x := 0; x < width; x++ {
			r := float64(x)/(float64(width)-1)*3.5 - 2.5
			i := float64(y)/(float64(height)-1)*2 - 1
			in, iter := inSet(complex(r, i))
			if in {
				img.Set(x, y, color.Black)
			} else {
				gray := float64(iter) / float64(iterations) * float64(math.MaxUint8)
				img.SetGray(x, y, color.Gray{uint8(math.MaxUint8 - gray)})
			}
		}
	}
	done <- true
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
