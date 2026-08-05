package importer

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func createTestPNG(t *testing.T, w, h int, fill color.Color) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, fill)
		}
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "test.png")
	f, _ := os.Create(path)
	png.Encode(f, img)
	f.Close()
	return path
}

func TestParseImageWhite(t *testing.T) {
	path := createTestPNG(t, 10, 5, color.White)
	img, err := ParseImage(path, 254)
	if err != nil {
		t.Fatal(err)
	}
	if img.Width != 10 || img.Height != 5 {
		t.Errorf("size = %dx%d, want 10x5", img.Width, img.Height)
	}
	for i, v := range img.Pixels {
		if v != 255 {
			t.Errorf("pixel %d = %d, want 255 for white", i, v)
			break
		}
	}
}

func TestParseImageBlack(t *testing.T) {
	path := createTestPNG(t, 4, 4, color.Black)
	img, err := ParseImage(path, 254)
	if err != nil {
		t.Fatal(err)
	}
	for i, v := range img.Pixels {
		if v != 0 {
			t.Errorf("pixel %d = %d, want 0 for black", i, v)
			break
		}
	}
}

func TestParseImageDimensions(t *testing.T) {
	path := createTestPNG(t, 254, 127, color.White)
	img, err := ParseImage(path, 254)
	if err != nil {
		t.Fatal(err)
	}
	if !closeTo(img.WidthMM(), 25.4, 0.01) {
		t.Errorf("WidthMM = %f, want 25.4", img.WidthMM())
	}
	if !closeTo(img.HeightMM(), 12.7, 0.01) {
		t.Errorf("HeightMM = %f, want 12.7", img.HeightMM())
	}
}

func TestParseImageDefaultDPI(t *testing.T) {
	path := createTestPNG(t, 10, 10, color.White)
	img, err := ParseImage(path, 0)
	if err != nil {
		t.Fatal(err)
	}
	if img.DPI != 254 {
		t.Errorf("DPI = %f, want 254 (default)", img.DPI)
	}
}

func TestParseImageAt(t *testing.T) {
	path := createTestPNG(t, 4, 4, color.White)
	img, _ := ParseImage(path, 254)
	if img.At(0, 0) != 255 {
		t.Errorf("At(0,0) = %d, want 255", img.At(0, 0))
	}
	if img.At(-1, 0) != 255 {
		t.Errorf("At(-1,0) should return 255 for out of bounds")
	}
	if img.At(100, 100) != 255 {
		t.Errorf("At(100,100) should return 255 for out of bounds")
	}
}

func TestParseImageGrayscale(t *testing.T) {
	// Pure red: BT.601 luma = 0.299*255 = 76.245
	path := createTestPNG(t, 1, 1, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	img, _ := ParseImage(path, 254)
	v := img.At(0, 0)
	if v < 75 || v > 77 {
		t.Errorf("red pixel luma = %d, want ~76", v)
	}
}

func TestParseImageNotFound(t *testing.T) {
	_, err := ParseImage("/nonexistent/file.png", 254)
	if err == nil {
		t.Error("expected error for missing file")
	}
}
