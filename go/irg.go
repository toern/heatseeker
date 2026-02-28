// Package irg implements a parser for the TC004 IRG thermal image format.
package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
)

// IrgHeader contains parsed metadata from an IRG file header.
type IrgHeader struct {
	UnknownHeader          int32
	FirstImageSize         uint32
	FirstImageWidth        uint16
	FirstImageHeight       uint16
	Pad1                   int8
	SecondImageSize        uint32
	SecondImageWidth       uint16
	SecondImageHeight      uint16
	Pad2                   uint8
	ThirdImageSize         uint32
	ThirdImageWidth        uint16
	ThirdImageHeight       uint16
	Emissivity             uint32
	ReflectiveTemperature  uint32
	AmbientTemperature     uint32
	Distance               uint32
	Unknown                uint32
	Transmissivity32       uint32
	Padding                uint32
	Transmissivity         uint16
}

// IrgData holds all extracted data from an IRG file.
type IrgData struct {
	Header           IrgHeader
	Grayscale        []uint8
	ThermalRaw       []uint16
	ThermalWidth     int
	ThermalHeight    int
}

// ThermalFahrenheit converts raw thermal data to Fahrenheit.
func (d *IrgData) ThermalFahrenheit() []float64 {
	result := make([]float64, len(d.ThermalRaw))
	for i, v := range d.ThermalRaw {
		kelvin := float64(v) / 10.0
		result[i] = (kelvin-273.15)*9.0/5.0 + 32.0
	}
	return result
}

// ThermalMinMax returns the min and max Fahrenheit values.
func (d *IrgData) ThermalMinMax() (float64, float64) {
	f := d.ThermalFahrenheit()
	if len(f) == 0 {
		return 0, 0
	}
	min, max := math.Inf(1), math.Inf(-1)
	for _, v := range f {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max
}

// ExtractIRG parses an IRG file and returns extracted image data.
func ExtractIRG(path string) (*IrgData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	// Read entire file
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	if len(data) < 0x100 {
		return nil, fmt.Errorf("file too small to be a valid IRG file (%d bytes)", len(data))
	}

	// Parse header using a struct that matches the binary layout.
	// The Python format string is: '<iIHHbIHHBIHHIIIIIIIH14xB'
	// We parse field-by-field for clarity.
	var hdr IrgHeader
	hdr.UnknownHeader = int32(binary.LittleEndian.Uint32(data[0x00:]))
	hdr.FirstImageSize = binary.LittleEndian.Uint32(data[0x04:])
	hdr.FirstImageWidth = binary.LittleEndian.Uint16(data[0x08:])
	hdr.FirstImageHeight = binary.LittleEndian.Uint16(data[0x0A:])
	hdr.Pad1 = int8(data[0x0C])
	hdr.SecondImageSize = binary.LittleEndian.Uint32(data[0x0D:])
	hdr.SecondImageWidth = binary.LittleEndian.Uint16(data[0x11:])
	hdr.SecondImageHeight = binary.LittleEndian.Uint16(data[0x13:])
	hdr.Pad2 = data[0x15]
	hdr.ThirdImageSize = binary.LittleEndian.Uint32(data[0x16:])
	hdr.ThirdImageWidth = binary.LittleEndian.Uint16(data[0x1A:])
	hdr.ThirdImageHeight = binary.LittleEndian.Uint16(data[0x1C:])
	hdr.Emissivity = binary.LittleEndian.Uint32(data[0x1E:])
	hdr.ReflectiveTemperature = binary.LittleEndian.Uint32(data[0x22:])
	hdr.AmbientTemperature = binary.LittleEndian.Uint32(data[0x26:])
	hdr.Distance = binary.LittleEndian.Uint32(data[0x2A:])
	hdr.Unknown = binary.LittleEndian.Uint32(data[0x2E:])
	hdr.Transmissivity = binary.LittleEndian.Uint16(data[0x32:])

	// Determine data start offset
	dataStart := 0x100
	if data[0x7E] == 0xAC && data[0x7F] == 0xCA {
		dataStart = 0x80
	}

	offset := dataStart

	// Extract grayscale image (uint8 array)
	gsSize := int(hdr.FirstImageSize)
	if offset+gsSize > len(data) {
		return nil, fmt.Errorf("file too small for grayscale data")
	}
	grayscale := make([]uint8, gsSize)
	copy(grayscale, data[offset:offset+gsSize])
	offset += gsSize

	// Extract thermal data (uint16 array)
	thermalCount := int(hdr.SecondImageSize)
	thermalByteSize := thermalCount * 2
	if offset+thermalByteSize > len(data) {
		return nil, fmt.Errorf("file too small for thermal data")
	}
	thermalRaw := make([]uint16, thermalCount)
	for i := 0; i < thermalCount; i++ {
		thermalRaw[i] = binary.LittleEndian.Uint16(data[offset+i*2:])
	}

	return &IrgData{
		Header:        hdr,
		Grayscale:     grayscale,
		ThermalRaw:    thermalRaw,
		ThermalWidth:  int(hdr.SecondImageWidth),
		ThermalHeight: int(hdr.SecondImageHeight),
	}, nil
}
