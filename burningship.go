package main

import (
	"fmt"
	"math"
	"strconv"
)

// A BurningShip represents the strongly typed planar space for the Burning Ship fractal
type burningShipPlane struct {
	Plane
}

func newBurningShip() burningShipPlane {
	return burningShipPlane{Plane{-1.5, 2.0, -2.0, 1.5}}
}

func (m *burningShipPlane) process(c config) {
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

func (m *burningShipPlane) image(c config) {
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

	var checkIfPointEscapes escapeCalculator = func(real float64, imag float64, config config) (bool, int, float64, float64) {
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

	m.iterateOverPoints(c, plottedChannel, checkIfPointEscapes)

	if c.filename == "" {
		c.filename = "ship_" + strconv.FormatFloat(c.midX, 'E', -1, 64) + "_" + strconv.FormatFloat(c.midY, 'E', -1, 64) + "_" + strconv.FormatFloat(c.zoom, 'E', -1, 64) + ".jpg"
	}

	saveimage(mbi, c.output, c.filename)

	fmt.Printf("%s/%s\n", c.output, c.filename)
}
