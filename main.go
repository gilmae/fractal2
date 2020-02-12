package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"os"
	"strings"
	"sync"
)

const (
	defaultGradient = `[["0.0", "000764"],["0.16", "026bcb"],["0.42", "edffff"],["0.6425", "ffaa00"],["0.8675", "000200"],["1.0","000764"]]`
)

// Fractals supported
const (
	MutantMandelbrotAlgoValue = "mutant_mandelbrot"
	MandelbrotAlgoValue       = "mandelbrot"
	BurningShipAlgoValue      = "ship"
	JuliaAlgoValue            = "julia"
	Z1ZcZIAlgoValue           = "z1zczi"
)

type Config struct {
	algorithm     string
	maxIterations int
	bailout       float64
	width         int
	height        int
	pointX        int
	pointY        int
	midX          float64
	midY          float64
	zoom          float64
	output        string
	filename      string
	gradient      string
	mode          string
	colourMode    string
	constR        float64
	constI        float64
}

// A Plane represents the base confines of the complex plane for a fractal based
// an escape time function.
type Plane struct {
	rMin float64 // The smallest value of the real component
	rMax float64 // The largest value of the real component
	iMin float64 // The smallest value of the imaginary component
	iMax float64 // The largest value of the imaginary component
}

// A Point represents a set of coordinates in the complex plane,
// and the coresponding point in a bitmap
type Point struct {
	X    int     // The X coordinated in the bitmap, where 0 is the left column
	Y    int     // The Y coordinated in the bitmap, where 0 is the top line
	real float64 // The scaled real component of the complex coordinate
	imag float64 // The scaled real component of the complex coordinate
}

// A PlottedPoint represents the result of the escape time function
type PlottedPoint struct {
	X          int     // The X coordinated in the bitmap, where 0 is the left column
	Y          int     // The Y coordinated in the bitmap, where 0 is the top line
	real       float64 // The real component of final value of z in the escape time calculation
	imag       float64 // The imaginary component of final value of z in the escape time calculation
	Iterations int     // The number of iterations it took to determine a result
	Escaped    bool    // True if the coordinate escaped the escape time function
}

func main() {
	c := getConfig()

	if c.algorithm == MandelbrotAlgoValue {
		m := newMandelbrot()
		m.process(c)
	} else if c.algorithm == MutantMandelbrotAlgoValue {
		m := newMutantMandelbrot()

		m.process(c)
	} else if c.algorithm == BurningShipAlgoValue {
		b := newBurningShip()

		b.process(c)
	} else if c.algorithm == JuliaAlgoValue {
		j := newJulia()
		j.process(c)
	} else if c.algorithm == Z1ZcZIAlgoValue {
		o := newZ1ZcZiMandelbrot()
		o.process(c)
	}

}

func getConfig() Config {
	var c Config

	var supportedAlgorithms = []string{MandelbrotAlgoValue, JuliaAlgoValue, BurningShipAlgoValue, MutantMandelbrotAlgoValue, Z1ZcZIAlgoValue}
	var supportedColourings = []string{TrueColouring, BandedColouring, SmoothColouring, NoColouring}

	flag.StringVar(&c.algorithm, "a", "mandelbrot", "Fractal algorithm: "+strings.Join(supportedAlgorithms, ", "))
	flag.Float64Var(&c.midX, "r", -99.0, "Real component of the midpoint.")
	flag.Float64Var(&c.midY, "i", -99.0, "Imaginary component of the midpoint.")
	flag.Float64Var(&c.zoom, "z", 1, "Zoom level.")
	flag.StringVar(&c.output, "o", ".", "Output path.")
	flag.StringVar(&c.filename, "f", "", "Output file name.")
	flag.StringVar(&c.colourMode, "c", "none", "Colour mode: "+strings.Join(supportedColourings, ", "))
	flag.Float64Var(&c.bailout, "b", 4.0, "Bailout value.")
	flag.IntVar(&c.width, "w", 1600, "Width of render.")
	flag.IntVar(&c.height, "h", 1600, "Height of render.")
	flag.IntVar(&c.maxIterations, "m", 2000, "Maximum Iterations before giving up on finding an escape.")
	flag.StringVar(&c.gradient, "g", defaultGradient, "Gradient to use.")
	flag.StringVar(&c.mode, "mode", "image", "Mode: image, coordsAt")
	flag.IntVar(&c.pointX, "x", 0, "x cordinate of a pixel, used for translating to the real component. 0,0 is top left.")
	flag.IntVar(&c.pointY, "y", 0, "y cordinate of a pixel, used for translating to the real component. 0,0 is top left.")
	flag.Float64Var(&c.constR, "cr", 0.0, "Real component of the const point in a Julia set.")
	flag.Float64Var(&c.constI, "ci", 0.0, "Imaginary component of the const point in a Julia set.")
	flag.Parse()

	return c
}

func max(a float64, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a float64, b float64) float64 {
	if a > b {
		return b
	}
	return a
}

func initialiseimage(c Config) *image.NRGBA {
	bounds := image.Rect(0, 0, c.width, c.height)
	mbi := image.NewNRGBA(bounds)
	draw.Draw(mbi, bounds, image.NewUniform(color.Black), image.ZP, draw.Src)
	return mbi
}

func saveimage(mbi *image.NRGBA, filepath string, filename string) {
	file, err := os.Create(filepath + "/" + filename)
	if err != nil {
		fmt.Println(err)
	}

	if err = jpeg.Encode(file, mbi, &jpeg.Options{jpeg.DefaultQuality}); err != nil {
		fmt.Println(err)
	}

	if err = file.Close(); err != nil {
		fmt.Println(err)
	}
}

func (p *Plane) getScale(zoom float64, height int, width int) (float64, float64, float64) {
	var pixelScaleRealAxis = (p.rMax - p.rMin) / float64(width-1) / zoom
	var pixelScaleImagAxis = (p.iMax - p.iMin) / float64(height-1) / zoom

	var pixelScale = min(pixelScaleRealAxis, pixelScaleImagAxis)

	pixelOffsetReal := float64(width-1) / 2.0
	pixelOffsetImag := float64(height-1) / 2.0

	return pixelScale, pixelOffsetReal, pixelOffsetImag
}

func (p *Plane) calculateCoordinatesAtPoint(config Config) (float64, float64) {
	var pixelScale, pixelOffsetReal, pixelOffsetImag = p.getScale(config.zoom, config.height, config.width)

	var real = config.midX + (float64(config.pointX)-pixelOffsetReal)*pixelScale
	var imag = config.midY - pixelScale*(-1.0*float64(config.pointY)+pixelOffsetImag)

	return real, imag
}

func (p *Plane) iterateOverPoints(config Config, plotted_channel chan PlottedPoint, calc escapeCalculator) {
	var pixelScale, pixelOffsetReal, pixelOffsetImag = p.getScale(config.zoom, config.height, config.width)

	points_channel := make(chan Point)

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			for p := range points_channel {
				var escaped, iteration, finalReal, finalImag = calc(p.real, p.imag, config)
				plotted_channel <- PlottedPoint{p.X, p.Y, finalReal, finalImag, iteration, escaped}
			}
			wg.Done()
		}()

	}

	for x := 0; x < config.width; x++ {
		r := config.midX + (float64(x)-pixelOffsetReal)*pixelScale
		for y := 0; y < config.height; y++ {
			i := config.midY + pixelScale*(-1.0*float64(y)+pixelOffsetImag)
			points_channel <- Point{x, y, r, i}
		}
	}

	close(points_channel)

	wg.Wait()
}

type escapeCalculator func(real float64, imag float64, config Config) (escaped bool, iterations int, finalReal float64, finalImaginary float64)
