package main

import (
	"fmt"
	"math"
	"strconv"
)

/* Good mid point is 0.45, -1 */

type BurningShip struct {
	Plane
}

func newBurningShip() BurningShip {
	return BurningShip{Plane{-1.5, 2.0, -2.0, 1.5}}
}

func (m *BurningShip) process(c Config) {
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

func (m *BurningShip) image(c Config) {
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
		var zReal = real
		var zImag = imag
		var iteration int

		for iteration = 0; iteration < config.maxIterations && (zReal*zReal+zImag*zImag) < config.bailout; iteration++ {
			xtemp := zReal*zReal - zImag*zImag - real
			newImag := math.Abs(2*zReal*zImag + imag)
			newReal := math.Abs(xtemp)

			zReal = newReal
			zImag = newImag
		}

		return iteration < config.maxIterations, iteration, zReal, zImag
	}

	m.iterateOverPoints(c, plotted_channel, checkIfPointEscapes)

	if c.filename == "" {
		c.filename = "ship_" + strconv.FormatFloat(c.midX, 'E', -1, 64) + "_" + strconv.FormatFloat(c.midY, 'E', -1, 64) + "_" + strconv.FormatFloat(c.zoom, 'E', -1, 64) + ".jpg"
	}

	saveimage(mbi, c.output, c.filename)

	fmt.Printf("%s/%s\n", c.output, c.filename)
}
