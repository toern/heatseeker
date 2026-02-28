package main

import (
	"math"
	"testing"
)

func TestInfernoColor(t *testing.T) {
	// Test boundaries
	c0 := infernoColor(0.0)
	if c0.A != 255 {
		t.Errorf("expected alpha 255, got %d", c0.A)
	}

	c1 := infernoColor(1.0)
	if c1.A != 255 {
		t.Errorf("expected alpha 255, got %d", c1.A)
	}

	// Values outside [0,1] should be clamped
	cNeg := infernoColor(-1.0)
	if cNeg != infernoColor(0.0) {
		t.Errorf("negative value should clamp to 0: got %v vs %v", cNeg, infernoColor(0.0))
	}

	cOver := infernoColor(2.0)
	if cOver != infernoColor(1.0) {
		t.Errorf("value >1 should clamp to 1: got %v vs %v", cOver, infernoColor(1.0))
	}
}

func TestThermalFahrenheit(t *testing.T) {
	data := &IrgData{
		ThermalRaw:    []uint16{2731, 3731}, // 273.1K and 373.1K
		ThermalWidth:  2,
		ThermalHeight: 1,
	}
	f := data.ThermalFahrenheit()
	if len(f) != 2 {
		t.Fatalf("expected 2 values, got %d", len(f))
	}

	// 273.1K = -0.05°C ≈ 31.91°F
	expectedF0 := (273.1 - 273.15) * 9.0 / 5.0 + 32.0
	if math.Abs(f[0]-expectedF0) > 0.01 {
		t.Errorf("expected %.2f°F, got %.2f°F", expectedF0, f[0])
	}

	// 373.1K = 99.95°C ≈ 211.91°F
	expectedF1 := (373.1 - 273.15) * 9.0 / 5.0 + 32.0
	if math.Abs(f[1]-expectedF1) > 0.01 {
		t.Errorf("expected %.2f°F, got %.2f°F", expectedF1, f[1])
	}
}

func TestThermalMinMax(t *testing.T) {
	data := &IrgData{
		ThermalRaw:    []uint16{2731, 3731, 3000},
		ThermalWidth:  3,
		ThermalHeight: 1,
	}
	min, max := data.ThermalMinMax()
	f := data.ThermalFahrenheit()

	// Min should be from 2731, max from 3731
	if math.Abs(min-f[0]) > 0.01 {
		t.Errorf("min mismatch: expected %.2f, got %.2f", f[0], min)
	}
	if math.Abs(max-f[1]) > 0.01 {
		t.Errorf("max mismatch: expected %.2f, got %.2f", f[1], max)
	}
}

func TestRenderThermalImage(t *testing.T) {
	fahrenheit := []float64{32.0, 100.0, 150.0, 212.0}
	img := renderThermalImage(fahrenheit, 2, 2, 32.0, 212.0)

	if img.Bounds().Dx() != 2 || img.Bounds().Dy() != 2 {
		t.Errorf("expected 2x2 image, got %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
	}

	// All pixels should have alpha 255
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a>>8 != 255 {
				t.Errorf("pixel (%d,%d) alpha should be 255, got %d", x, y, a>>8)
			}
		}
	}
}
