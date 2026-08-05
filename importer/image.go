package importer

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
)

// GrayImage holds a grayscale representation of an imported image.
type GrayImage struct {
	Width  int
	Height int
	Pixels []uint8 // row-major, 0=black 255=white
	DPI    float64
}

// ParseImage loads a PNG or JPEG and converts to 8-bit grayscale.
func ParseImage(filename string, dpi float64) (*GrayImage, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", filename, err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decoding %s: %w", filename, err)
	}

	if dpi <= 0 {
		dpi = 254 // ~10 lines/mm, common for laser engraving
	}

	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	pixels := make([]uint8, w*h)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, _ := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			// ITU-R BT.601 luma
			luma := (299*r + 587*g + 114*b) / 1000
			pixels[y*w+x] = uint8(luma >> 8)
		}
	}

	return &GrayImage{
		Width:  w,
		Height: h,
		Pixels: pixels,
		DPI:    dpi,
	}, nil
}

// WidthMM returns the image width in millimeters at the configured DPI.
func (g *GrayImage) WidthMM() float64 {
	return float64(g.Width) / g.DPI * 25.4
}

// HeightMM returns the image height in millimeters at the configured DPI.
func (g *GrayImage) HeightMM() float64 {
	return float64(g.Height) / g.DPI * 25.4
}

// At returns the grayscale value at pixel (x, y). Returns 255 if out of bounds.
func (g *GrayImage) At(x, y int) uint8 {
	if x < 0 || x >= g.Width || y < 0 || y >= g.Height {
		return 255
	}
	return g.Pixels[y*g.Width+x]
}
