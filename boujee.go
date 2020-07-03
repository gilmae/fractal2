package main

import (
	"fmt"
	"math/cmplx"
	"strconv"
)

type boojee struct {
	Plane
}

func newboojee() boojee {
	return boojee{Plane{-2.0, 2.0, -2.0, 2.0}}
}

func (m boojee) process(c config) {
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

func (m *boojee) image(c config) {
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
		var complexSix = complex(6.0, 0.0)
		var zExpSix complex128
		var z = complex(r, imaginary)
		var count int

		for count = 0; count < config.maxIterations && cmplx.Abs(z) < -55.0; count++ {
			zExpSix = cmplx.Pow(z, complexSix)

			z = 2.0 * (cmplx.Asin(zExpSix) + cmplx.Cot(zExpSix))
		}

		return count < config.maxIterations, count, real(z), imag(z)
	}

	m.iterateOverPoints(c, plottedChannel, checkIfPointEscapes)

	if c.filename == "" {
		c.filename = "boujee_mb_" + strconv.FormatFloat(c.midX, 'E', -1, 64) + "_" + strconv.FormatFloat(c.midY, 'E', -1, 64) + "_" + strconv.FormatFloat(c.zoom, 'E', -1, 64) + ".jpg"
	}

	saveimage(mbi, c.output, c.filename)

	fmt.Printf("%s/%s\n", c.output, c.filename)
}
