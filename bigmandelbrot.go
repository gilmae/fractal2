package main

import (
	"fmt"
	"strconv"
	"math/big"
)

// A Mandelbrot represents the strongly typed planar space for the mandelbrot fractal
type bigMandelbrotPlane struct {
	Plane
}

func newBigMandelbrot() bigMandelbrotPlane {
	return bigMandelbrotPlane{Plane{-2.25, 0.75, -1.5, 1.5}}
}

func (m *bigMandelbrotPlane) process(c config) {
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

func (m *bigMandelbrotPlane) image(c config) {
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

func (m *bigMandelbrotPlane) calculateEscape(real float64, imag float64, config config) (bool, int, big.Float, big.Float) {
	zR := new(big.Float)
	zR.SetFloat64(real)
	zI := new(big.Float)
	zI.SetFloat64(imag)
	// Check that the point isn't in the main cardioid or the period-2 bulb.
	// If it is, just bail out now

	/*if ((real+1.0)*(real+1.0))+imag*imag <= 0.0625 {
		return false, config.maxIterations, 0.0, 0.0
	}*/

	rSquare := new(big.Float)
	rSquare.SetFloat64(0.0)
	
	iSquare := new(big.Float)
	iSquare.SetFloat64(0.0)
	
	zSquare := new(big.Float)
	zSquare.SetFloat64(0.0)

	rSquarePlusISquare := new(big.Float)
	rSquarePlusISquare.Add(rSquare, iSquare)
	
	x := new(big.Float)
	y  := new(big.Float)
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

	bailout := new(big.Float)
	bailout.SetFloat64(config.bailout)
	bailout.Mul(bailout,bailout)

	for iteration = 1; rSquarePlusISquare.Cmp(bailout) <= 0 && iteration < config.maxIterations; iteration++ {
		*x = GetNewX(rSquare, iSquare, zr)
		*y = GetNewX(zSquare, rSquare, iSquare, zi)

		*rsquare = GetNewRSquare(x)
		*isquare = GetNewRSquare(y)
		*zsquare = GetNewZSquare(x,y)
	}
	return iteration < config.maxIterations, iteration, x, y
}

func GetNewRSquare(x *big.Float) big.Float {
	newRSquare := new(big.Float)
	newRSquare.Copy(x)
	newRSquare.Mul(newRSquare,x)
	return *newRSquare
}

func GetNewISquare(y *big.Float) big.Float {
	newISquare := new(big.Float)
	newISquare.Copy(y)
	newISquare.Mul(newISquare,y)
	return *newISquare
}

func GetNewZSquare(x *big.Float, y *big.Float) big.Float {
	newZSquare := new(big.Float)
	newZSquare.Copy(x)
	newZSquare.Add(newZSquare,y)
	
	newZSquare.Mul(newZSquare, newZSquare)
	return *newZSquare
}

func GetNewX(rsquare *big.Float, isquare *big.Float, real *big.Float) big.Float {
	newX := new(big.Float)
	newX.Copy(rsquare)
	newX.Sub(newX,isquare)
	newX.Add(newX,real)
	return *newX

}

func GetNewY(zsquare *big.Float, rsquare *big.Float, isquare *big.Float, imag *big.Float) big.Float {
	newY := new(big.Float)
	newY.Copy(zsquare)
	newY.Sub(newY,rsquare)
	newY.Sub(newY,isquare)
	newY.Add(newY,imag)
	return *newY

}

func GetRSquare(rsquare *big.Float) big.Float {
	newRSquare := new(big.Float)
	newRSquare.Copy(rsquare)
	newRSquare.Mul(newRSquare, rsquare)
	return *newRSquare
}
