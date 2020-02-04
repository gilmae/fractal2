package main
import (
	"sync"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"os"
	"fmt"
	"strconv"
)

type MutantMandelbrot struct {
	rMin float64
	rMax float64
	iMin float64
	iMax float64
}

func NewMutantMandelbrot() MutantMandelbrot {
	return MutantMandelbrot{-2.25, 0.75, -1.5, 1.5}
}

func (m MutantMandelbrot) Process(c Config) {
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

func (m *MutantMandelbrot) Image(c Config) {
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

	m.Iterate_Over_Points(c, plotted_channel)

	if c.filename == "" {
		c.filename = "mutant_mb_" + strconv.FormatFloat(c.midX, 'E', -1, 64) + "_" + strconv.FormatFloat(c.midY, 'E', -1, 64) + "_" + strconv.FormatFloat(c.zoom, 'E', -1, 64) + ".jpg"
	}
	
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

	fmt.Printf("%s/%s\n", c.output, c.filename)
}

func (m *MutantMandelbrot) Calculate_Coordinates_At_Point(config Config) (float64, float64) {
	var pixelScale = (m.rMax - m.rMin) / float64(config.width-1)
	pixelOffset := float64(config.width-1)/2.0

	var real = config.midX + (float64(config.pointX) - pixelOffset) * pixelScale
	var imag = config.midY - pixelScale * (-1.0 * float64(config.pointY) + pixelOffset);

	return real, imag
}

func (m *MutantMandelbrot) Iterate_Over_Points(config Config, plotted_channel chan PlottedPoint){
	var pixelScale = (m.rMax - m.rMin) / float64(config.width-1)
	pixelOffset := float64(config.width-1)/2.0

	points_channel := make(chan Point)

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			for p := range points_channel {
				var escaped, iteration = m.Check_If_Point_Escapes(p.real, p.imag, config)
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

func (m *MutantMandelbrot) Check_If_Point_Escapes(real float64, imag float64, config Config) (bool, int) {
            var zx = real;
			var zy = imag;
			var x = real
			var y = imag
            var n = 50;
			var p = 100;
			var count int

            for count = 0; count < config.maxIterations && zx * zx + zy * zy < 20.0; count++ {
                if (0 == (count + 1) % n) {
                    x += zx * float64(count) / float64(p);
                    y += zy * float64(count) / float64(p);
                    n--;
                    p++;
                }
                var new_zx = zx * zx - zy * zy + x;
                zy = 2 * zx * zy + y;
                zx = new_zx;
            }

            return  count < config.maxIterations, count
}