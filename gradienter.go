package main

import (
	"encoding/hex"
	"encoding/json"
	"image/color"
	"math"
	"strconv"

	"github.com/gilmae/interpolation"
)

const ( // Colour Modes
	TrueColouring   = "true"
	SmoothColouring = "smooth"
	NoColouring     = "none"
	BandedColouring = "banded"
)

const (
	PaletteLength = 16
)

var redInterpolant interpolation.MonotonicCubic
var greenInterpolant interpolation.MonotonicCubic
var blueInterpolant interpolation.MonotonicCubic

func initialiseGradient(gradient_str string) {
	var g [][]string

	byt := []byte(gradient_str)
	_ = json.Unmarshal(byt, &g)
	var size = len(g)
	var xSequence = make([]float64, size)
	var redpoints = make([]float64, size)
	var greenpoints = make([]float64, size)
	var bluepoints = make([]float64, size)

	for i, v := range g {
		xSequence[i], _ = strconv.ParseFloat(v[0], 64)
		b, _ := hex.DecodeString(v[1])
		redpoints[i] = float64(b[0])
		greenpoints[i] = float64(b[1])
		bluepoints[i] = float64(b[2])
	}

	redInterpolant = interpolation.CreateMonotonicCubic(xSequence, redpoints)
	greenInterpolant = interpolation.CreateMonotonicCubic(xSequence, greenpoints)
	blueInterpolant = interpolation.CreateMonotonicCubic(xSequence, bluepoints)
}

func getPixelColour(point PlottedPoint, maxIterations int, colour_mode string) color.NRGBA {
	if colour_mode == "true" {

		var gradient_position = float64(point.Iterations) / float64(maxIterations)
		var redpoint = redInterpolant(gradient_position)
		var greenpoint = greenInterpolant(gradient_position)
		var bluepoint = blueInterpolant(gradient_position)

		return color.NRGBA{uint8(redpoint), uint8(greenpoint), uint8(bluepoint), 255}
	} else if colour_mode == "smooth" {

		palette := fillPalette()

		jitteredEscape := jitter(point)
		index1 := int(math.Abs(jitteredEscape))
		t2 := jitteredEscape - float64(index1)
		t1 := 1 - t2

		index1 = index1 % len(palette)
		index2 := (index1 + 1) % len(palette)

		clr1 := palette[index1]
		clr2 := palette[index2]

		r := float64(clr1.R)*t1 + float64(clr2.R)*t2
		g := float64(clr1.G)*t1 + float64(clr2.G)*t2
		b := float64(clr1.B)*t1 + float64(clr2.B)*t2

		return color.NRGBA{uint8(r), uint8(g), uint8(b), 255}

	} else if colour_mode == "banded" {

		palette := fillPalette()
		return palette[point.Iterations%len(palette)]

	} else {

		return color.NRGBA{255, 255, 255, 255}
	}
}

func fillPalette() []color.NRGBA {
	var palette = make([]color.NRGBA, PaletteLength)
	for i := 0; i < PaletteLength; i++ {
		var point = float64(i) / float64(PaletteLength)

		var redpoint = redInterpolant(point)
		var greenpoint = greenInterpolant(point)
		var bluepoint = blueInterpolant(point)

		palette[i] = color.NRGBA{uint8(redpoint), uint8(greenpoint), uint8(bluepoint), 255}
	}

	return palette

}

func jitter(p PlottedPoint) float64 {
	magnitude := math.Sqrt(p.real*p.real + p.imag*p.imag)
	return float64(p.Iterations+1) - (math.Log(math.Log(magnitude)))/math.Log(2.0)
}
