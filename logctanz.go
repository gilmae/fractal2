package main

// Attempted to mimic formulae found in https://www.reddit.com/r/fractals/comments/o6zf73/cool_fractal_info_about_it_in_the_comments/
// Don't think it is quite succesful
// Best called with iterations turned down to 100


import (
	"fmt"
	"strconv"
	"math/cmplx"
)

// A Mandelbrot represents the strongly typed planar space for the mandelbrot fractal
type LogTanPlane struct {
	Plane
}

func newLogTan() LogTanPlane {
	return LogTanPlane{Plane{-2.25, 0.75, -1.5, 1.5}}
}

func (m *LogTanPlane) process(c config) {
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

func (m *LogTanPlane) image(c config) {
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
		c.filename = "logtan_" + strconv.FormatFloat(c.midX, 'E', -1, 64) + "_" + strconv.FormatFloat(c.midY, 'E', -1, 64) + "_" + strconv.FormatFloat(c.zoom, 'E', -1, 64) + ".jpg"
	}

	saveimage(mbi, c.output, c.filename)

	fmt.Printf("%s/%s\n", c.output, c.filename)
}

func (m *LogTanPlane) calculateEscape(r float64, i float64, config config) (bool, int, float64, float64) {
	//z := complex(1.0, 0.0)
	z := complex(r, i)
	c := complex(r, i)
	
	var iteration int

	var bailout = 1000.0
	for iteration = 1; imag(z) * real(z) <= bailout && iteration < config.maxIterations; iteration++ {
		z = z * cmplx.Log(c) * cmplx.Tan(z) + c
	}

	return iteration < config.maxIterations, iteration, real(z), imag(z)
}
