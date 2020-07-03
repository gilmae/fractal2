package main

import (
	"fmt"
	"math/cmplx"
	"strconv"
)

type z1ZcZiPlane struct {
	Plane
}

func newZ1ZcZi() z1ZcZiPlane {
	return z1ZcZiPlane{Plane{-2.25, 0.75, -1.5, 1.5}}
}

func (m z1ZcZiPlane) process(c config) {
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

func (m *z1ZcZiPlane) image(c config) {
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

	var checkIfPointEscapes escapeCalculator = func(r float64, imaginary float64, config config) (bool, int, float64, float64) {
		var z = complex(0.0, 0.0)
		var c = complex(r, imaginary)
		var count int

		for count = 0; count < config.maxIterations && cmplx.Abs(z) < 4.0; count++ {
			z = (z + 1.0) * (z + c) * (z + complex(0.0, 1.0))
		}

		return count < config.maxIterations, count, real(z), imag(z)
	}

	m.iterateOverPoints(c, plottedChannel, checkIfPointEscapes)

	if c.filename == "" {
		c.filename = "z1zczi_mb_" + strconv.FormatFloat(c.midX, 'E', -1, 64) + "_" + strconv.FormatFloat(c.midY, 'E', -1, 64) + "_" + strconv.FormatFloat(c.zoom, 'E', -1, 64) + ".jpg"
	}

	saveimage(mbi, c.output, c.filename)

	fmt.Printf("%s/%s\n", c.output, c.filename)
}
