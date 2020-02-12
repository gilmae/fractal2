package main

import "testing"

func TestMandelbrotEscapeAtMandelbrotMidPoint(t *testing.T) {
	m := newMandelbrot()
	var c config
	c.maxIterations = 1000

	escaped, iterations, finalR, finalI := m.calculateEscape(-0.75, 0.0, c)

	expectedEscape := false
	expectedFinalR := 0.0
	expectedFinalI := 0.0

	if escaped {
		t.Errorf("Escaped was incorrect, got: %t, want: %t.", escaped, expectedEscape)
	}

	if iterations != c.maxIterations {
		t.Errorf("Iterations were incorrect, got: %d, want: %d.", iterations, c.maxIterations)
	}

	if finalR != expectedFinalR {
		t.Errorf("Final r was incorrect, got: %f, want: %f.", finalR, expectedFinalR)
	}

	if finalI != expectedFinalI {
		t.Errorf("Final I was oncorrect, got: %f, want: %f.", finalI, expectedFinalI)
	}
}

func TestMandelbrotEscapeAtMandelbrotMidRealEdge(t *testing.T) {
	m := newMandelbrot()
	var c config
	c.maxIterations = 1000
	c.bailout = 4.0

	escaped, iterations, finalR, finalI := m.calculateEscape(-2.25, 0.0, c)

	expectedEscape := true
	expectedFinalR := 5.66015625
	expectedFinalI := 0.0
	expectedIterations := 4

	if escaped != expectedEscape {
		t.Errorf("Escaped was incorrect, got: %t, want: %t.", escaped, expectedEscape)
	}

	if iterations != expectedIterations {
		t.Errorf("Iterations were incorrect, got: %d, want: %d.", iterations, expectedIterations)
	}

	if finalR != expectedFinalR {
		t.Errorf("Final r was incorrect, got: %f, want: %f.", finalR, expectedFinalR)
	}

	if finalI != expectedFinalI {
		t.Errorf("Final I was oncorrect, got: %f, want: %f.", finalI, expectedFinalI)
	}
}
