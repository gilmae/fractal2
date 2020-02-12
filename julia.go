package main

import (
	"fmt"
	"math"
	"strconv"
)

type Julia struct {
	Plane
}

func newJulia() Julia {
	return Julia{Plane{-2.0, 2.0, -2.0, 2.0}}
}

func (m *Julia) process(c Config) {
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

func (m *Julia) image(c Config) {
	initialiseGradient(c.gradient)

	c.bailout = determineJuliaBailout(c)

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
		var iteration int
		zR := real
		zI := imag

		for iteration = 0.0; zR*zR+zI*zI < config.bailout && iteration < config.maxIterations; iteration++ {
			tmp := zR*zR - zI*zI
			zI = 2*zR*zI + config.constI
			zR = tmp + config.constR
		}

		return iteration < config.maxIterations, iteration, zR, zI
	}

	m.iterateOverPoints(c, plotted_channel, checkIfPointEscapes)

	if c.filename == "" {
		c.filename = "julia_" + strconv.FormatFloat(c.midX, 'E', -1, 64) + "_" + strconv.FormatFloat(c.midY, 'E', -1, 64) + "_" + strconv.FormatFloat(c.zoom, 'E', -1, 64) + ".jpg"
	}

	saveimage(mbi, c.output, c.filename)

	fmt.Printf("%s/%s\n", c.output, c.filename)
}

func determineJuliaBailout(config Config) float64 {
	/* Where c is the constant in the Julia algorithim, expressed as a complex number,
	Bailout should be R where R**2 - R = |c|.
	That's the quadratic equation, which will give us two values. We'll take the larger.
	*/

	cAbs := math.Sqrt(config.constR*config.constR + config.constI*config.constI)

	a, b := quadratic(1.0, -1.0, -1.0*cAbs)

	/* The bailout test will be testing against the square of R anyway, so doing it now saves
	messing about with absolute values of the returns from the quadratic formula */

	return max(a*a, b*b)
}

func quadratic(a float64, b float64, c float64) (float64, float64) {
	// 0 = ax**2 + bx + c
	// x = (-b ± sqrt(b**2 - 4ac)) / 2a

	d := math.Sqrt(b*b - 4*a*c)

	return (-1.0*b + d) / 2.0 * a, (-1.0*b - d) / 2.0 * a
}
