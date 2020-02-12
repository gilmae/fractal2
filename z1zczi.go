package main
import (
	"fmt"
	"strconv"
	"math/cmplx"
)

type Z1ZcZiMandelbrot struct {
	Plane
}

func NewZ1ZcZiMandelbrot() Z1ZcZiMandelbrot {
	return Z1ZcZiMandelbrot{Plane{-2.25, 0.75, -1.5, 1.5}}
}

func (m Z1ZcZiMandelbrot) Process(c Config) {
	if (c.midX == -99.0) {
		c.midX = (m.rMax + m.rMin) / 2.0
	}

	if (c.midY == -99.0) {
		c.midY = (m.iMax + m.iMin) / 2.0
	}

	if c.mode == "image" {
		m.Image(c)
	} else if c.mode == "coordsAt" {
		var r,i = m.Calculate_Coordinates_At_Point(c)
		fmt.Printf("%18.17e, %18.17e\n", r, i)
	}
}

func (m *Z1ZcZiMandelbrot) Image(c Config) {
	initialise_gradient(c.gradient)
	
	mbi := Initialise_Image(c)

	plotted_channel := make(chan PlottedPoint)

	go func (points <- chan PlottedPoint) {
		for p := range points {
			 if p.Escaped {
				mbi.Set(p.X, p.Y, get_colour(p.Iterations, c.maxIterations, c.colourMode))
			 }
		}
	}(plotted_channel)

	var Check_If_Point_Escapes EscapeCalculator =  func(r float64, imaginary float64, config Config) (bool, int) {
		var z = complex(0.0, 0.0)
		var c = complex(r, imaginary)
		var count int

		for count = 0; count < config.maxIterations && cmplx.Abs(z) < 4.0; count++ {
			z = (z+1.0) * (z + c) * (z + complex(0.0,1.0))
		}

		return  count < config.maxIterations, count
}

	m.Iterate_Over_Points(c, plotted_channel, Check_If_Point_Escapes)

	if c.filename == "" {
		c.filename = "z1zczi_mb_" + strconv.FormatFloat(c.midX, 'E', -1, 64) + "_" + strconv.FormatFloat(c.midY, 'E', -1, 64) + "_" + strconv.FormatFloat(c.zoom, 'E', -1, 64) + ".jpg"
	}
	
	Save_Image(mbi, c.output, c.filename)

	fmt.Printf("%s/%s\n", c.output, c.filename)
}
