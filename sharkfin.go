package main

import (
	"fmt"
	"math"
	"strconv"
)

type sharkFinPlane struct {
	Plane
}

func newSharkFin() sharkFinPlane {
	return sharkFinPlane{Plane{-2.25, 0.75, -1.5, 1.5}}
}

func (m sharkFinPlane) process(c config) {
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

func (m *sharkFinPlane) image(c config) {
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

	var checkIfPointEscapes escapeCalculator = func(r float64, i float64, config config) (bool, int, float64, float64) {
		var zr = r
		var zi = i
		var zrc float64
		var zic float64
		var count int

		for count = 0; count < config.maxIterations && zr*zr+zi*zi < 4.0; count++ {
			zr = zr + r;
			zi = zi + i;
			zrc = zr * zr - math.Abs(zi) * zi;
			zic = zr * zi * 2;
			zr = zrc;
			zi = zic;
		}

		return count < config.maxIterations, count, zr, zi
	}

	m.iterateOverPoints(c, plottedChannel, checkIfPointEscapes)

	if c.filename == "" {
		c.filename = "sharkFinPlane_mb_" + strconv.FormatFloat(c.midX, 'E', -1, 64) + "_" + strconv.FormatFloat(c.midY, 'E', -1, 64) + "_" + strconv.FormatFloat(c.zoom, 'E', -1, 64) + ".jpg"
	}

	saveimage(mbi, c.output, c.filename)

	fmt.Printf("%s/%s\n", c.output, c.filename)
}
