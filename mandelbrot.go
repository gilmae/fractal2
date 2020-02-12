package main
import (
	"fmt"
	"strconv"
)

type Mandelbrot struct {
	Plane
}

func NewMandelbrot() Mandelbrot {
	return Mandelbrot{Plane{-2.25, 0.75, -1.5, 1.5}}
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
				mbi.Set(p.X, p.Y, get_colour(p, c.maxIterations, c.colourMode))
			 }
		}
	}(plotted_channel)

	var Check_If_Point_Escapes EscapeCalculator =  func(real float64, imag float64, config Config) (bool, int, float64, float64) {
		// Check that the point isn't in the main cardioid or the period-2 bulb.
		// If it is, just bail out now
		if ((real + 1.0) * (real + 1.0)) + imag * imag <= 0.0625 {
			return false, config.maxIterations, 0.0, 0.0
		}
	
		var rsquare = 0.0;
		var isquare = 0.0;
		var zsquare = 0.0;
		var x float64;
		var y float64;
		var iteration int
	
		/* The Mandelbrot function is to iterate the function z(n+1) = z(n)**2 + c, where z(0) = 0 + 0i,
             * for each complex number c in the plane. The following is a slightly modified version of that
             * function intended to avoid as many floating point multiplications as possible.
             *
             * see https://en.wikipedia.org/wiki/Mandelbrot_set#Escape_time_algorithm for details,
             * and https://en.wikipedia.org/wiki/Mandelbrot_set#Optimizations for details on optimizations
             * 
             * We're also using doubles rather than complex numbers for two reasons:
             * 
             * 1. It makes the cardioid pre-calc check at the beginning of this function easier.
             * 2. It will faciliate later adoption of BigFloat
             * 
			 */
		for iteration = 1; rsquare + isquare <= config.bailout && iteration < config.maxIterations; iteration++ {
			x = rsquare - isquare + real;
			y = zsquare - rsquare - isquare + imag;
			rsquare = x * x;
			isquare = y * y;
			zsquare = (x + y) * (x + y);
		 }
		 
		 return iteration < config.maxIterations, iteration, x, y
	}

	m.Iterate_Over_Points(c, plotted_channel, Check_If_Point_Escapes)
	
	if c.filename == "" {
		c.filename = "mandelbrot_" + strconv.FormatFloat(c.midX, 'E', -1, 64) + "_" + strconv.FormatFloat(c.midY, 'E', -1, 64) + "_" + strconv.FormatFloat(c.zoom, 'E', -1, 64) + ".jpg"
	}
	
	Save_Image(mbi, c.output, c.filename)

	fmt.Printf("%s/%s\n", c.output, c.filename)
}



