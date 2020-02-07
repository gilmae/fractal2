package main
import (
	"sync"
	"fmt"
	"math"
	"strconv"

)

/* Good mid point is 0.45, -1 */

type BurningShip struct {
	Plane
}

func NewBurningShip() BurningShip {
	return BurningShip{Plane{-1.5, 2.0, -2.0, 1.5}}
}

func (m *BurningShip) Process(c Config) {
	if (c.midX == -99.0) {
		c.midX = (m.rMax + m.rMin) / 2.0
	}

	if (c.midY == -99.0) {
		c.midY = (m.iMax + m.iMin) / 2.0
	}
	if c.mode == "image" {
		m.Image(c)
	} else if c.mode == "coordsAt" {
		var r,i = m.Calculate_Coordinates_At_Point(c)
		fmt.Printf("%18.17e, %18.17e\n", r, i)
	}
}

func (m *BurningShip) Image(c Config) {
	initialise_gradient(c.gradient)
	
	mbi := Initialise_Image(c)

	plotted_channel := make(chan PlottedPoint)

	go func (points <- chan PlottedPoint) {
		for p := range points {
			 if p.Escaped {
				mbi.Set(p.X, p.Y, get_colour(p.Iterations, c.maxIterations))
			 }
		}
	}(plotted_channel)

	m.Iterate_Over_Points(c, plotted_channel)

	if c.filename == "" {
		c.filename = "ship_" + strconv.FormatFloat(c.midX, 'E', -1, 64) + "_" + strconv.FormatFloat(c.midY, 'E', -1, 64) + "_" + strconv.FormatFloat(c.zoom, 'E', -1, 64) + ".jpg"
	}

	Save_Image(mbi, c.output, c.filename)

	fmt.Printf("%s/%s\n", c.output, c.filename)
}

func (m *BurningShip) Iterate_Over_Points(config Config, plotted_channel chan PlottedPoint){
	var pixelScale, pixelOffsetReal, pixelOffsetImag = m.Get_Scale(config.zoom, config.height, config.width)

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
		r:= config.midX + (float64(x) - pixelOffsetReal) * pixelScale
		for y := 0; y < config.height; y++ {
			i := config.midY - pixelScale * (-1.0 * float64(y) + pixelOffsetImag);
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