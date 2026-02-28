package main

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

// infernoColor returns an inferno-style colormap color for t in [0,1].
func infernoColor(t float64) color.RGBA {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}

	var r, g, b float64
	switch {
	case t < 0.25:
		s := t / 0.25
		r = s * 0.34
		g = s * 0.06
		b = 0.01 + s*0.33
	case t < 0.5:
		s := (t - 0.25) / 0.25
		r = 0.34 + s*0.48
		g = 0.06 + s*0.06
		b = 0.34 + s*(-0.04)
	case t < 0.75:
		s := (t - 0.5) / 0.25
		r = 0.82 + s*0.13
		g = 0.12 + s*0.53
		b = 0.30 + s*(-0.26)
	default:
		s := (t - 0.75) / 0.25
		r = 0.95 + s*0.03
		g = 0.65 + s*0.30
		b = 0.04 + s*0.90
	}

	clamp := func(v float64) uint8 {
		if v < 0 {
			v = 0
		}
		if v > 1 {
			v = 1
		}
		return uint8(v * 255)
	}

	return color.RGBA{R: clamp(r), G: clamp(g), B: clamp(b), A: 255}
}

// renderThermalImage creates an RGBA image from Fahrenheit data using the inferno colormap.
func renderThermalImage(fahrenheit []float64, width, height int, vmin, vmax float64) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	rng := vmax - vmin
	if math.Abs(rng) < 1e-6 {
		rng = 1.0
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			if idx >= len(fahrenheit) {
				break
			}
			t := (fahrenheit[idx] - vmin) / rng
			img.SetRGBA(x, y, infernoColor(t))
		}
	}
	return img
}

func main() {
	a := app.New()
	w := a.NewWindow("HeatSeeker")
	w.Resize(fyne.NewSize(900, 700))

	// State
	var (
		irgData      *IrgData
		fahrenheit   []float64
		rangeMin     float64
		rangeMax     float64
		dataMin      float64
		dataMax      float64
	)

	// UI components
	tempLabel := widget.NewLabel("Temperature: —")
	tempLabel.TextStyle = fyne.TextStyle{Bold: true}

	infoLabel := widget.NewLabel("")

	imgCanvas := canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 1, 1)))
	imgCanvas.FillMode = canvas.ImageFillContain
	imgCanvas.ScaleMode = canvas.ImageScalePixels

	minLabel := widget.NewLabel("Min: —")
	maxLabel := widget.NewLabel("Max: —")

	minSlider := widget.NewSlider(0, 100)
	maxSlider := widget.NewSlider(0, 100)
	minSlider.Step = 0.1
	maxSlider.Step = 0.1

	updateImage := func() {
		if irgData == nil || len(fahrenheit) == 0 {
			return
		}
		img := renderThermalImage(fahrenheit, irgData.ThermalWidth, irgData.ThermalHeight, rangeMin, rangeMax)
		imgCanvas.Image = img
		imgCanvas.Refresh()
		minLabel.SetText(fmt.Sprintf("Min: %.1f °F", rangeMin))
		maxLabel.SetText(fmt.Sprintf("Max: %.1f °F", rangeMax))
	}

	minSlider.OnChanged = func(v float64) {
		rangeMin = v
		if rangeMin > rangeMax {
			rangeMin = rangeMax
			minSlider.Value = rangeMin
			minSlider.Refresh()
		}
		updateImage()
	}

	maxSlider.OnChanged = func(v float64) {
		rangeMax = v
		if rangeMax < rangeMin {
			rangeMax = rangeMin
			maxSlider.Value = rangeMax
			maxSlider.Refresh()
		}
		updateImage()
	}

	resetBtn := widget.NewButton("Reset Range", func() {
		rangeMin = dataMin
		rangeMax = dataMax
		minSlider.Value = dataMin
		maxSlider.Value = dataMax
		minSlider.Refresh()
		maxSlider.Refresh()
		updateImage()
	})

	loadFile := func(path string) {
		data, err := ExtractIRG(path)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		irgData = data
		fahrenheit = data.ThermalFahrenheit()
		dataMin, dataMax = data.ThermalMinMax()
		rangeMin = dataMin
		rangeMax = dataMax

		minSlider.Min = dataMin
		minSlider.Max = dataMax
		minSlider.Value = dataMin
		minSlider.Refresh()

		maxSlider.Min = dataMin
		maxSlider.Max = dataMax
		maxSlider.Value = dataMax
		maxSlider.Refresh()

		infoLabel.SetText(fmt.Sprintf(
			"Emissivity: %.4f\nAmbient: %.1f K\nDistance: %.1f m\nSize: %dx%d",
			float64(data.Header.Emissivity)/10000.0,
			float64(data.Header.AmbientTemperature)/10000.0,
			float64(data.Header.Distance)/10000.0,
			data.ThermalWidth, data.ThermalHeight,
		))

		updateImage()
	}

	// Menu
	openItem := fyne.NewMenuItem("Open IRG…", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			reader.Close()
			loadFile(reader.URI().Path())
		}, w)
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".irg"}))
		fd.Show()
	})
	fileMenu := fyne.NewMenu("File", openItem)
	mainMenu := fyne.NewMainMenu(fileMenu)
	w.SetMainMenu(mainMenu)

	// Hover detection on image — Fyne doesn't have pixel-level hover on canvas images
	// natively, so we show a general temperature summary instead.
	// For full hover-to-inspect, a custom widget would be needed.

	// Layout
	controls := container.NewVBox(
		widget.NewLabelWithStyle("Temperature Range", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewLabel("Min:"),
		minSlider,
		minLabel,
		widget.NewLabel("Max:"),
		maxSlider,
		maxLabel,
		layout.NewSpacer(),
		resetBtn,
		widget.NewSeparator(),
		tempLabel,
		widget.NewSeparator(),
		infoLabel,
	)

	content := container.NewBorder(nil, nil, nil, controls, imgCanvas)
	w.SetContent(content)
	w.ShowAndRun()
}
