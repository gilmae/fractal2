package main

import (
	"flag"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"fmt"
	"os"
	"sync"
)

const (
	defaultGradient = `[["0.0", "000764"],["0.16", "026bcb"],["0.42", "edffff"],["0.6425", "ffaa00"],["0.8675", "000200"],["1.0","000764"]]`
	MutantMandelbrotAlgoValue = "mutant_mandelbrot"
	MandelbrotAlgoValue = "mandelbrot"
	BurningShipAlgoValue = "ship"
	JuliaAlgoValue = "julia"
)

type Config struct {
	algorithm			string
	maxIterations		int
	bailout				float64
	width				int
	height				int
	pointX				int
	pointY 				int
	midX 				float64
	midY 				float64
	zoom 				float64
	output 				string
	filename 			string
	gradient 			string
	mode 				string
	colourMode 			string
	constR 				float64
	constI				float64
}

type Plane struct {
	rMin float64
	rMax float64
	iMin float64
	iMax float64
}

type Point struct {
	real		float64
	imag		float64
	X       	int
	Y       	int
	
}

type PlottedPoint struct {
	X			int
	Y			int
	real		float64
	imag		float64
	Iterations	int
	Escaped		bool
}


func main() {
	c := Get_Config()

	if (c.algorithm == MandelbrotAlgoValue) {
		m := NewMandelbrot()
		m.Process(c)
	} else if (c.algorithm == MutantMandelbrotAlgoValue) {
		m := NewMutantMandelbrot()
	
		m.Process(c)
	} else if c.algorithm == BurningShipAlgoValue {
		b := NewBurningShip()
	
		b.Process(c)
	} else if c.algorithm == JuliaAlgoValue {
		j := NewJulia()
		j.Process(c)
	}

}



func Get_Config() Config {
	var c Config
	flag.StringVar(&c.algorithm, "a", "mandelbrot", "Fractal algorithm: " + MandelbrotAlgoValue + ", " + MutantMandelbrotAlgoValue + ", " + JuliaAlgoValue)
	flag.Float64Var(&c.midX, "r", -99.0, "Real component of the midpoint.")
	flag.Float64Var(&c.midY, "i", -99.0, "Imaginary component of the midpoint.")
	flag.Float64Var(&c.zoom, "z", 1, "Zoom level.")
	flag.StringVar(&c.output, "o", ".", "Output path.")
	flag.StringVar(&c.filename, "f", "", "Output file name.")
	flag.StringVar(&c.colourMode, "c", "none", "Colour mode: true, smooth, banded, none.")
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

func Max(a float64, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func Min(a float64, b float64) float64 {
	if a > b {
		return b
	}
	return a
}

func Initialise_Image(c Config) *image.NRGBA {
	bounds := image.Rect(0, 0, c.width, c.height)
	mbi := image.NewNRGBA(bounds)
	draw.Draw(mbi, bounds, image.NewUniform(color.Black), image.ZP, draw.Src)
	return mbi
}

func Save_Image(mbi *image.NRGBA, filepath string, filename string){
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

func (p *Plane) Get_Scale(zoom float64, height int, width int) (float64, float64, float64) {
	var pixelScaleRealAxis = (p.rMax - p.rMin) / float64(width-1) / zoom
	var pixelScaleImagAxis = (p.iMax - p.iMin) / float64(height-1) / zoom

	var pixelScale = Min(pixelScaleRealAxis, pixelScaleImagAxis)

	pixelOffsetReal := float64(width-1)/2.0
	pixelOffsetImag := float64(height-1)/2.0

	return pixelScale, pixelOffsetReal, pixelOffsetImag
}

func (p *Plane) Calculate_Coordinates_At_Point(config Config) (float64, float64) {
	var pixelScale, pixelOffsetReal, pixelOffsetImag = p.Get_Scale(config.zoom, config.height, config.width)

	var real = config.midX + (float64(config.pointX) - pixelOffsetReal) * pixelScale
	var imag = config.midY - pixelScale * (-1.0 * float64(config.pointY) + pixelOffsetImag);

	return real, imag
}

func (p *Plane) Iterate_Over_Points(config Config, plotted_channel chan PlottedPoint, calc EscapeCalculator){
	var pixelScale, pixelOffsetReal, pixelOffsetImag = p.Get_Scale(config.zoom, config.height, config.width)

	points_channel := make(chan Point)

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			for p := range points_channel {
				var escaped, iteration = calc(p.real, p.imag, config)
				plotted_channel <- PlottedPoint{p.X, p.Y, p.real, p.imag, iteration, escaped}
			}
			wg.Done()
		}()

	}

	for x := 0; x <  config.width; x++ {
		r:= config.midX + (float64(x) - pixelOffsetReal) * pixelScale
		for y := 0; y < config.height; y++ {
			i := config.midY - pixelScale * (-1.0 * float64(y) + pixelOffsetImag);
			points_channel <- Point{r, i, x, y}
		}
	}

	close(points_channel)

	wg.Wait()
}

type EscapeCalculator func(real float64, imag float64, config Config) (bool,int)