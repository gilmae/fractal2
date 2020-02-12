package main

import (
	"fmt"
	"math"
	"strconv"
)

type Mandelbrot struct {
	Plane
}

func newMandelbrot() Mandelbrot {
	return Mandelbrot{Plane{-2.25, 0.75, -1.5, 1.5}}
}

func (m *Mandelbrot) process(c Config) {
	if c.midX == -99.0 {
		c.midX = (m.rMax + m.rMin) / 2.0
	}

	if c.midY == -99.0 {
		c.midY = (m.iMax + m.iMin) / 2.0
	}

	if c.mode == "image" {
		m.image(c)
	} else if c.mode == "coordsAt" {
		var r, i = m.calculateCoordinatesAtPoint(c)
		fmt.Printf("%18.17e, %18.17e\n", r, i)
	}
}

func (m *Mandelbrot) image(c Config) {
	initialiseGradient(c.gradient)

	mbi := initialiseimage(c)

	plotted_channel := make(chan PlottedPoint)

	go func(points <-chan PlottedPoint) {
		for p := range points {
			if p.Escaped {
				mbi.Set(p.X, p.Y, getPixelColour(p, c.maxIterations, c.colourMode))
			}
		}
	}(plotted_channel)

	var checkIfPointEscapes escapeCalculator = func(real float64, imag float64, config Config) (bool, int, float64, float64) {
		// Check that the point isn't in the main cardioid or the period-2 bulb.
		// If it is, just bail out now
		if ((real+1.0)*(real+1.0))+imag*imag <= 0.0625 {
			return false, config.maxIterations, 0.0, 0.0
		}

		var rsquare = 0.0
		var isquare = 0.0
		var zsquare = 0.0
		var x float64
		var y float64
		var iteration int

		rOld := math.MaxFloat64
		iOld := math.MaxFloat64
		periodocity := 20

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
		for iteration = 1; rsquare+isquare <= config.bailout && iteration < config.maxIterations; iteration++ {
			x = rsquare - isquare + real
			y = zsquare - rsquare - isquare + imag

			// If the function has encountered these coordinates before, then we will encounter them again.
			// Which means we will be in a loop, i.e. there is no escape, it's in the mandelbrot set.
			if x == rOld && y == iOld {
				return iteration < config.maxIterations, iteration, x, y
			}
			rsquare = x * x
			isquare = y * y
			zsquare = (x + y) * (x + y)

			// At the end of the length of the periodocity check, set new co-ordinates
			if iteration%periodocity == 0 {
				rOld = x
				iOld = y
			}
		}

		return iteration < config.maxIterations, iteration, x, y
	}

	m.iterateOverPoints(c, plotted_channel, checkIfPointEscapes)

	if c.filename == "" {
		c.filename = "mandelbrot_" + strconv.FormatFloat(c.midX, 'E', -1, 64) + "_" + strconv.FormatFloat(c.midY, 'E', -1, 64) + "_" + strconv.FormatFloat(c.zoom, 'E', -1, 64) + ".jpg"
	}

	saveimage(mbi, c.output, c.filename)

	fmt.Printf("%s/%s\n", c.output, c.filename)
}
