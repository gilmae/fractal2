package main
import (
	"sync"

	"fmt"
	"strconv"
)

type Mandelbrot struct {
	rMin float64
	rMax float64
	iMin float64
	iMax float64
}

func NewMandelbrot() Mandelbrot {
	return Mandelbrot{-2.25, 0.75, -1.5, 1.5}
}

func (m *Mandelbrot) Process(c Config) {
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

func (m *Mandelbrot) Image(c Config) {
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
		c.filename = "mandelbrot_" + strconv.FormatFloat(c.midX, 'E', -1, 64) + "_" + strconv.FormatFloat(c.midY, 'E', -1, 64) + "_" + strconv.FormatFloat(c.zoom, 'E', -1, 64) + ".jpg"
	}
	
	Save_Image(mbi, c.output, c.filename)

	fmt.Printf("%s/%s\n", c.output, c.filename)
}

func (m *Mandelbrot) Calculate_Coordinates_At_Point(config Config) (float64, float64) {
	var pixelScale = ((m.rMax - m.rMin) / float64(config.width-1)) / config.zoom
	pixelOffset := float64(config.width-1)/2.0

	var real = config.midX + (float64(config.pointX) - pixelOffset) * pixelScale
	var imag = config.midY - pixelScale * (-1.0 * float64(config.pointY) + pixelOffset);

	return real, imag
}

func (m *Mandelbrot) Iterate_Over_Points(config Config, plotted_channel chan PlottedPoint){
	var pixelScaleRealAxis = (m.rMax - m.rMin) / float64(config.width-1) / config.zoom
	var pixelScaleImagAxis = (m.iMax - m.iMin) / float64(config.height-1) / config.zoom

	var pixelScale = Min(pixelScaleRealAxis, pixelScaleImagAxis)

	pixelOffsetReal := float64(config.width-1)/2.0
	pixelOffsetImag := float64(config.height-1)/2.0

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

func (m *Mandelbrot) Check_If_Point_Escapes(real float64, imag float64, config Config) (bool, int) {
	if ((real + 1.0) * (real + 1.0)) + imag * imag <= 0.0625 {
    	return false, config.maxIterations
    }

	var rsquare = 0.0;
	var isquare = 0.0;
	var zsquare = 0.0;
	var x float64;
	var y float64;
	var iteration int


	for iteration = 1; rsquare + isquare <= config.bailout && iteration < config.maxIterations; iteration++ {
		x = rsquare - isquare + real;
		y = zsquare - rsquare - isquare + imag;
		rsquare = x * x;
 		isquare = y * y;
		zsquare = (x + y) * (x + y);
	 }
	 
	 return iteration < config.maxIterations, iteration
}