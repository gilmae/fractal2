package main
import (
	"sync"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"os"
	"fmt"
	"math"
)

type BurningShip struct {
	rMin float64
	rMax float64
	iMin float64
	iMax float64
}

func NewBurningShip() BurningShip {
	return BurningShip{-2.5, 1.5, -1, 2.0}
}

func (m *BurningShip) Process(c Config) {
	c.midX = 0.45
	c.midY = 0.5
	if c.mode == "image" {
		m.Image(c)
	} else if c.mode == "coordsAt" {
		var r,i = m.Calculate_Coordinates_At_Point(c)
		fmt.Printf("%18.17e, %18.17e\n", r, i)
	}
}

func (m *BurningShip) Image(c Config) {
	initialise_gradient(c.gradient)
	
	bounds := image.Rect(0, 0, c.width, c.width)
	mbi := image.NewNRGBA(bounds)
	draw.Draw(mbi, bounds, image.NewUniform(color.Black), image.ZP, draw.Src)

	plotted_channel := make(chan PlottedPoint)

	go func (points <- chan PlottedPoint) {
		for p := range points {
			 if p.Escaped {
				mbi.Set(p.X, p.Y, get_colour(p.Iterations, c.maxIterations))
			 }
		}
	}(plotted_channel)

	m.Iterate_Over_Points(c, plotted_channel)

	file, err := os.Create(c.output + "/" + c.filename)
	if err != nil {
		fmt.Println(err)
	}

	if err = jpeg.Encode(file, mbi, &jpeg.Options{jpeg.DefaultQuality}); err != nil {
		fmt.Println(err)
	}

	if err = file.Close(); err != nil {
		fmt.Println(err)
	}

	fmt.Printf("%s/%s\n", c.output, c.filename)
}

func (m *BurningShip) Calculate_Coordinates_At_Point(config Config) (float64, float64) {
	var pixelScale = (m.rMax - m.rMin) / float64(config.width-1)
	pixelOffset := float64(config.width-1)/2.0

	var real = config.midX + (float64(config.pointX) - pixelOffset) * pixelScale
	var imag = config.midY - pixelScale * (-1.0 * float64(config.pointY) + pixelOffset);

	return real, imag
}

func (m *BurningShip) Iterate_Over_Points(config Config, plotted_channel chan PlottedPoint){
	var pixelScale = (m.rMax - m.rMin) / float64(config.width-1)
	pixelOffset := float64(config.width-1)/2.0

	points_channel := make(chan Point)

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			for p := range points_channel {
				var escaped, iteration = m.Check_If_Point_Escapes(p.real, p.imag, config)
				plotted_channel <- PlottedPoint{p.X, p.Y, p.real, p.imag, iteration, escaped}
			}
			wg.Done()
		}()

	}

	for x := 0; x <  config.width; x++ {
		r:= config.midX + (float64(x) - pixelOffset) * pixelScale
		for y := 0; y < config.width; y++ {
			i := config.midY - pixelScale * (-1.0 * float64(y) + pixelOffset);
			points_channel <- Point{r, i, x, y}
		}
	}

	close(points_channel)

	wg.Wait()
}

func (m *BurningShip) Check_If_Point_Escapes(real float64, imag float64, config Config) (bool, int) {
	var zReal = real
	var zImag = imag
    var iteration int 

    for iteration = 0; iteration < config.maxIterations && (zReal * zReal + zImag*zImag) < config.bailout; iteration++ {
      xtemp := zReal * zReal - zImag*zImag - real
      newImag := math.Abs(2 * zReal * zImag + imag)
      newReal := math.Abs(xtemp)
  
	  zReal = newReal
	  zImag = newImag
    }
  
  

  
	 return iteration < config.maxIterations, iteration
}