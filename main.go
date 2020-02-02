package main

import (
	"fmt"
	"sync"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"os"
	"flag"
	"strconv"
)

const (
	rMin = -2.25
	rMax = 0.75
	iMin = -1.5
	iMax = 1.5
	defaultGradient = `[["0.0", "000764"],["0.16", "026bcb"],["0.42", "edffff"],["0.6425", "ffaa00"],["0.8675", "000200"],["1.0","000764"]]`
)

type Config struct {
	maxIterations int
	bailout float64
	width int
	height int
	pointX int
	pointY int
	midX float64
	midY float64
	zoom float64
	output string
	filename string
	gradient string
	mode string
	colourMode string
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

	initialise_gradient(c.gradient)
	
	bounds := image.Rect(0, 0, c.width, c.width)
	mbi := image.NewNRGBA(bounds)
	draw.Draw(mbi, bounds, image.NewUniform(color.Black), image.ZP, draw.Src)

	plotted_channel := make(chan PlottedPoint)

	go func (points <- chan PlottedPoint) {
		for p := range points {
			 if p.Escaped {
				mbi.Set(p.X, p.Y, get_colour(p.Iterations, c.maxIterations))
			 }
		}
	}(plotted_channel)

	Iterate_Over_Points(c, plotted_channel)

	file, err := os.Create(c.output + "/" + c.filename)
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

func Iterate_Over_Points(config Config, plotted_channel chan PlottedPoint){
	var pixelScale = (rMax - rMin) / float64(config.width-1)
	pixelOffset := float64(config.width-1)/2.0

	points_channel := make(chan Point)

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			for p := range points_channel {
				var escaped, iteration = Check_If_Point_Escapes(p.real, p.imag, config)
				plotted_channel <- PlottedPoint{p.X, p.Y, p.real, p.imag, iteration, escaped}
			}
			wg.Done()
		}()

	}

	for x := 0; x <  config.width; x++ {
		r:= config.midX + (float64(x) - pixelOffset) * pixelScale
		for y := 0; y < config.width; y++ {
			i := config.midY - pixelScale * (-1.0 * float64(y) + pixelOffset);
			points_channel <- Point{r, i, x, y}
		}
	}

	close(points_channel)

	wg.Wait()
}

func Check_If_Point_Escapes(real float64, imag float64, config Config) (bool, int) {
	if ((real + 1.0) * (real + 1.0)) + imag * imag <= 0.0625 {
    	return false, config.maxIterations
    }

	var rsquare = 0.0;
	var isquare = 0.0;
	var zsquare = 0.0;
	var x float64;
	var y float64;
	var iteration int


	for iteration = 1; rsquare + isquare <= config.bailout && iteration < config.maxIterations; iteration++ {
		x = rsquare - isquare + real;
		y = zsquare - rsquare - isquare + imag;
		rsquare = x * x;
 		isquare = y * y;
		zsquare = (x + y) * (x + y);
	 }
	 
	 return iteration < config.maxIterations, iteration
}

func Get_Config() Config {
	var c Config
	flag.Float64Var(&c.midX, "r", -0.75, "Real component of the midpoint.")
	flag.Float64Var(&c.midY, "i", 0.0, "Imaginary component of the midpoint.")
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
	flag.Parse()

	if c.filename == "" {
		c.filename = "mb_" + strconv.FormatFloat(c.midX, 'E', -1, 64) + "_" + strconv.FormatFloat(c.midY, 'E', -1, 64) + "_" + strconv.FormatFloat(c.zoom, 'E', -1, 64) + ".jpg"
	}

	return c
}