package main

import (
	"fmt"
	"strconv"
)

// A Mandelbrot represents the strongly typed planar space for the mandelbrot fractal
type mandelbrotPlane struct {
	Plane
}

func newMandelbrot() mandelbrotPlane {
	return mandelbrotPlane{Plane{-2.25, 0.75, -1.5, 1.5}}
}

func (m *mandelbrotPlane) process(c config) {
	if c.midX == -99.0 {
		c.midX = (m.rMax + m.rMin) / 2.0
	}

	if c.midY == -99.0 {
		c.midY = (m.iMax + m.iMin) / 2.0
	}

	if c.mode == imageMode {
		m.image(c)
	} else if c.mode == coordinatesMode {
		var r, i = m.calculateCoordinatesAtPoint(c)
		fmt.Printf("%18.17e, %18.17e\n", r, i)
	}
}

func (m *mandelbrotPlane) image(c config) {
	initialiseGradient(c.gradient)

	mbi := initialiseimage(c)

	plottedChannel := make(chan PlottedPoint)

	go func(points <-chan PlottedPoint) {
		for p := range points {
			if p.Escaped {
				mbi.Set(p.X, p.Y, getPixelColour(p, c.maxIterations, c.colourMode))
			}
		}
	}(plottedChannel)

	m.iterateOverPoints(c, plottedChannel, m.calculateEscape)

	if c.filename == "" {
		c.filename = "mandelbrot_" + strconv.FormatFloat(c.midX, 'E', -1, 64) + "_" + strconv.FormatFloat(c.midY, 'E', -1, 64) + "_" + strconv.FormatFloat(c.zoom, 'E', -1, 64) + ".jpg"
	}

	saveimage(mbi, c.output, c.filename)

	fmt.Printf("%s/%s\n", c.output, c.filename)
}

func (m *mandelbrotPlane) calculateEscape(real float64, imag float64, config config) (bool, int, float64, float64) {
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

	var bailout = config.bailout * config.bailout
	for iteration = 1; rsquare+isquare <= bailout && iteration < config.maxIterations; iteration++ {
		x = rsquare - isquare + real
		y = zsquare - rsquare - isquare + imag

		rsquare = x * x
		isquare = y * y
		zsquare = (x + y) * (x + y)
	}

	return iteration < config.maxIterations, iteration, x, y
}
