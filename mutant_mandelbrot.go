package main

import (
	"fmt"
	"strconv"
)

type MutantMandelbrot struct {
	Plane
}

func newMutantMandelbrot() MutantMandelbrot {
	return MutantMandelbrot{Plane{-2.25, 0.75, -1.5, 1.5}}
}

func (m MutantMandelbrot) process(c Config) {
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

func (m *MutantMandelbrot) image(c Config) {
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
		var zx = real
		var zy = imag
		var x = real
		var y = imag
		var n = 50
		var p = 100
		var count int

		for count = 0; count < config.maxIterations && zx*zx+zy*zy < 20.0; count++ {
			if 0 == (count+1)%n {
				x += zx * float64(count) / float64(p)
				y += zy * float64(count) / float64(p)
				n--
				p++
			}
			var new_zx = zx*zx - zy*zy + x
			zy = 2*zx*zy + y
			zx = new_zx
		}

		return count < config.maxIterations, count, zx, zy
	}

	m.iterateOverPoints(c, plotted_channel, checkIfPointEscapes)

	if c.filename == "" {
		c.filename = "mutant_mb_" + strconv.FormatFloat(c.midX, 'E', -1, 64) + "_" + strconv.FormatFloat(c.midY, 'E', -1, 64) + "_" + strconv.FormatFloat(c.zoom, 'E', -1, 64) + ".jpg"
	}

	saveimage(mbi, c.output, c.filename)

	fmt.Printf("%s/%s\n", c.output, c.filename)
}
