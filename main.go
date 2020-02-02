package main

import (
	"fmt"
	"sync"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"os"
)

const (
	rMin = -2.25
	rMax = 0.75
	iMin = -1.5
	iMax = 1.5
)

type Point struct {
	real		float64
	imag		float64
	X       	int
	Y       	int
	
}

type PlottedPoint struct {
	X			int
	Y			int
	real		float64
	imag		float64
	Iterations	int
	Escaped		bool
}


func main() {
	var width = 1000
	bounds := image.Rect(0, 0, width, width)
	mbi := image.NewNRGBA(bounds)
	draw.Draw(mbi, bounds, image.NewUniform(color.Black), image.ZP, draw.Src)

	plotted_channel := make(chan PlottedPoint)

	go func (points <- chan PlottedPoint) {
		for p := range points {
			 if p.Escaped {
				mbi.Set(p.X, p.Y, color.NRGBA{255, 255, 255, 255})
			 }
		}
	}(plotted_channel)

	Iterate_Over_Points(width, plotted_channel)

	file, err := os.Create("/Users/gilmae/tmp/f.jpeg")
	if err != nil {
		fmt.Println(err)
	}

	if err = jpeg.Encode(file, mbi, &jpeg.Options{jpeg.DefaultQuality}); err != nil {
		fmt.Println(err)
	}

	if err = file.Close(); err != nil {
		fmt.Println(err)
	}
}

func Iterate_Over_Points(width int, plotted_channel chan PlottedPoint){
	var pixelScale = (rMax - rMin) / float64(width-1)
	pixelOffset := float64(width-1)/2.0

	points_channel := make(chan Point)

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			for p := range points_channel {
				var escaped, iteration = Check_If_Point_Escapes(p.real, p.imag)
				plotted_channel <- PlottedPoint{p.X, p.Y, p.real, p.imag, iteration, escaped}
			}
			wg.Done()
		}()

	}

	for x := 0; x <  width; x++ {
		r:= -0.75 + (float64(x) - pixelOffset) * pixelScale
		for y := 0; y < width; y++ {
			i := 0.0 - pixelScale * (-1.0 * float64(y) + pixelOffset);
			points_channel <- Point{r, i, x, y}
		}
	}

	close(points_channel)

	wg.Wait()
}

func Check_If_Point_Escapes(real float64, imag float64) (bool, int) {
	var rsquare = 0.0;
	var isquare = 0.0;
	var zsquare = 0.0;
	var x float64;
	var y float64;
	var _maxIterations = 1000;
	var iteration int

	for iteration = 1; rsquare + isquare <= 4.0 && iteration < _maxIterations; iteration++ {
		x = rsquare - isquare + real;
		y = zsquare - rsquare - isquare + imag;
		rsquare = x * x;
 		isquare = y * y;
		zsquare = (x + y) * (x + y);
	 }
	 
	 return iteration < _maxIterations, iteration
}