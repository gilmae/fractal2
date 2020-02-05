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
	"math"
)

type Julia struct {
	rMin float64
	rMax float64
	iMin float64
	iMax float64
}

func NewJulia() Julia {
	return Julia{-2.0, 2.0, -2.0, 2.0}
}

func (m *Julia) Process(c Config) {
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

func (m *Julia) Image(c Config) {
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
		c.filename = "julia_" + strconv.FormatFloat(c.midX, 'E', -1, 64) + "_" + strconv.FormatFloat(c.midY, 'E', -1, 64) + "_" + strconv.FormatFloat(c.zoom, 'E', -1, 64) + ".jpg"
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

func (m *Julia) Calculate_Coordinates_At_Point(config Config) (float64, float64) {
	var pixelScale = (m.rMax - m.rMin) / float64(config.width-1)
	pixelOffset := float64(config.width-1)/2.0

	var real = config.midX + (float64(config.pointX) - pixelOffset) * pixelScale
	var imag = config.midY - pixelScale * (-1.0 * float64(config.pointY) + pixelOffset);

	return real, imag
}

func (m *Julia) Iterate_Over_Points(config Config, plotted_channel chan PlottedPoint){
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

func (m *Julia) Check_If_Point_Escapes(real float64, imag float64, config Config) (bool, int) {
	/* Where c is the constant in the Julia algorithim, expressed as a complex number, 
	Bailout should be R where R**2 - R = |c|. 
	That's the quadratic equation, which will give us two values. We'll take the larger.
	*/
	var iteration int
	zR := real
	zI := imag

	cAbs := math.Sqrt(config.constR * config.constR + config.constI * config.constI)

	a,b := Quadratic(1.0, -1.0, -1.0 * cAbs)
	
	bailout := Max(a*a, b*b)
	  
	for iteration = 0.0; zR * zR + zI * zI < bailout && iteration < config.maxIterations; iteration++ {
	  tmp := zR * zR - zI * zI
	  zI = 2 * zR * zI  + config.constI
	  zR = tmp + config.constR 
	}
	 
	 return iteration < config.maxIterations, iteration
}

func Quadratic(a float64, b float64, c float64) (float64, float64) {
	// 0 = ax**2 + bx + c
	// x = (-b ± sqrt(b**2 - 4ac)) / 2a

	d := math.Sqrt(b*b - 4 * a * c)

	return (-1.0*b + d) / 2.0 * a, (-1.0*b - d) / 2.0 * a
}

func Max(a float64, b float64) float64 {
	if a > b {
		return a
	}
	return b
}