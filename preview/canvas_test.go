package preview

import (
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/fabienlroy/GoLaserCut/importer"
)

func approx(a, b float64) bool {
	return math.Abs(a-b) < 0.01
}

func TestNewCanvas(t *testing.T) {
	c := NewCanvas(800, 600)
	if c.Width != 800 || c.Height != 600 {
		t.Errorf("got %dx%d, want 800x600", c.Width, c.Height)
	}
	img := c.Image()
	bounds := img.Bounds()
	if bounds.Dx() != 800 || bounds.Dy() != 600 {
		t.Errorf("image bounds %dx%d, want 800x600", bounds.Dx(), bounds.Dy())
	}
}

func TestWorldToPixel(t *testing.T) {
	c := NewCanvas(100, 100)
	c.SetView(Viewport{0, 0, 99, 99})

	px, py := c.worldToPixel(0, 99)
	if px != 0 || py != 0 {
		t.Errorf("top-left: got (%d,%d), want (0,0)", px, py)
	}

	px, py = c.worldToPixel(99, 0)
	if px != 99 || py != 99 {
		t.Errorf("bottom-right: got (%d,%d), want (99,99)", px, py)
	}

	px, py = c.worldToPixel(49.5, 49.5)
	if px != 49 || py != 49 {
		t.Errorf("center: got (%d,%d), want (49,49)", px, py)
	}
}

func TestPixelToWorld(t *testing.T) {
	c := NewCanvas(100, 100)
	c.SetView(Viewport{0, 0, 99, 99})

	wx, wy := c.pixelToWorld(0, 0)
	if !approx(wx, 0) || !approx(wy, 99) {
		t.Errorf("pixel (0,0): got world (%.1f,%.1f), want (0,99)", wx, wy)
	}

	wx, wy = c.pixelToWorld(99, 99)
	if !approx(wx, 99) || !approx(wy, 0) {
		t.Errorf("pixel (99,99): got world (%.1f,%.1f), want (99,0)", wx, wy)
	}
}

func TestFitBoundsAspectRatio(t *testing.T) {
	c := NewCanvas(800, 400) // 2:1 aspect
	c.FitBounds(0, 0, 100, 100)

	v := c.View()
	viewW := v.MaxX - v.MinX
	viewH := v.MaxY - v.MinY
	ratio := viewW / viewH

	canvasRatio := float64(c.Width) / float64(c.Height)
	if !approx(ratio, canvasRatio) {
		t.Errorf("viewport ratio %f != canvas ratio %f", ratio, canvasRatio)
	}
}

func TestFitBoundsWideWorld(t *testing.T) {
	c := NewCanvas(400, 400) // 1:1
	c.FitBounds(0, 0, 200, 50)

	v := c.View()
	viewW := v.MaxX - v.MinX
	viewH := v.MaxY - v.MinY
	if !approx(viewW, viewH) {
		t.Errorf("1:1 canvas should have equal viewport W/H, got %f x %f", viewW, viewH)
	}
}

func TestFitEntity(t *testing.T) {
	ent := &importer.Entity{
		Paths: []importer.Path{{
			Points: []importer.Point{{0, 0}, {100, 0}, {100, 50}, {0, 50}},
			Closed: true,
		}},
	}
	c := NewCanvas(800, 600)
	c.FitEntity(ent)

	v := c.View()
	if v.MinX >= 0 || v.MinY >= 0 || v.MaxX <= 100 || v.MaxY <= 50 {
		t.Errorf("viewport should include entity bounds with margin")
	}
}

func TestClear(t *testing.T) {
	c := NewCanvas(10, 10)
	c.Clear()

	pixel := c.Image().RGBAAt(5, 5)
	if pixel != colorBg {
		t.Errorf("pixel after clear = %v, want %v", pixel, colorBg)
	}
}

func TestDrawPaths(t *testing.T) {
	c := NewCanvas(100, 100)
	c.SetView(Viewport{0, 0, 10, 10})
	c.Clear()

	paths := []importer.Path{{
		Points: []importer.Point{{0, 5}, {10, 5}},
		Layer:  "cut",
	}}
	c.DrawPaths(paths)

	// The line goes from world (0,5) to (10,5)
	// In pixels: (0,49) to (99,49) (approximately)
	midPixel := c.Image().RGBAAt(50, 49)
	if midPixel == colorBg {
		t.Error("expected drawn pixel on the path, got background")
	}
}

func TestDrawClosedPath(t *testing.T) {
	c := NewCanvas(100, 100)
	c.SetView(Viewport{0, 0, 10, 10})
	c.Clear()

	paths := []importer.Path{{
		Points: []importer.Point{{2, 2}, {8, 2}, {8, 8}, {2, 8}},
		Closed: true,
		Layer:  "border",
	}}
	c.DrawPaths(paths)

	// Check that at least some non-background pixels exist
	nonBg := 0
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			if c.Image().RGBAAt(x, y) != colorBg {
				nonBg++
			}
		}
	}
	if nonBg == 0 {
		t.Error("closed path should have drawn pixels")
	}
}

func TestLayerColors(t *testing.T) {
	c := NewCanvas(100, 100)
	c.SetView(Viewport{0, 0, 10, 10})
	c.Clear()

	paths := []importer.Path{
		{Points: []importer.Point{{0, 3}, {10, 3}}, Layer: "cut"},
		{Points: []importer.Point{{0, 7}, {10, 7}}, Layer: "engrave"},
	}
	c.DrawPaths(paths)

	// Different layers should get different colors
	_, py1 := c.worldToPixel(5, 3)
	_, py2 := c.worldToPixel(5, 7)
	c1 := c.Image().RGBAAt(50, py1)
	c2 := c.Image().RGBAAt(50, py2)

	if c1 == c2 {
		t.Error("different layers should have different colors")
	}
}

func TestDrawGrid(t *testing.T) {
	c := NewCanvas(200, 200)
	c.SetView(Viewport{-10, -10, 10, 10})
	c.Clear()
	c.DrawGrid()

	// Origin should be drawn (brighter than background)
	ox, oy := c.worldToPixel(0, 0)
	pixel := c.Image().RGBAAt(ox, oy)
	if pixel == colorBg {
		t.Error("origin should be drawn")
	}
}

func TestDrawToolpaths(t *testing.T) {
	c := NewCanvas(100, 100)
	c.SetView(Viewport{0, 0, 10, 10})
	c.Clear()

	toolpaths := []Toolpath3D{{
		Points: []Point3D{{0, 5, 0}, {10, 5, -5}},
	}}
	c.DrawToolpaths(toolpaths, 5.0)

	nonBg := 0
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			if c.Image().RGBAAt(x, y) != colorBg {
				nonBg++
			}
		}
	}
	if nonBg == 0 {
		t.Error("toolpath should have drawn pixels")
	}
}

func TestDrawToolpathClosed(t *testing.T) {
	c := NewCanvas(100, 100)
	c.SetView(Viewport{0, 0, 10, 10})
	c.Clear()

	toolpaths := []Toolpath3D{{
		Points: []Point3D{{2, 2, -1}, {8, 2, -1}, {8, 8, -1}},
		Closed: true,
	}}
	c.DrawToolpaths(toolpaths, 2.0)

	nonBg := 0
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			if c.Image().RGBAAt(x, y) != colorBg {
				nonBg++
			}
		}
	}
	if nonBg == 0 {
		t.Error("closed toolpath should have drawn pixels")
	}
}

func TestDepthColor(t *testing.T) {
	surface := depthColor(0, 5)
	deep := depthColor(-5, 5)

	// Surface should be cyan-ish (low R, high B)
	if surface.R > surface.B {
		t.Errorf("surface should be cyan-ish: R=%d B=%d", surface.R, surface.B)
	}

	// Deep should be red-ish (high R, low B)
	if deep.R < deep.B {
		t.Errorf("deep should be red-ish: R=%d B=%d", deep.R, deep.B)
	}
}

func TestDepthColorGradient(t *testing.T) {
	c1 := depthColor(0, 10)
	c2 := depthColor(-5, 10)
	c3 := depthColor(-10, 10)

	// Red should increase with depth
	if c2.R <= c1.R || c3.R <= c2.R {
		t.Errorf("red should increase: %d → %d → %d", c1.R, c2.R, c3.R)
	}
}

func TestDepthColorClamp(t *testing.T) {
	c := depthColor(-20, 5) // beyond max depth
	if c.R != 255 {
		t.Errorf("beyond max depth R=%d, want 255", c.R)
	}
}

func TestGridSpacing(t *testing.T) {
	tests := []struct {
		viewWidth float64
		wantMax   float64
	}{
		{2, 0.5},
		{10, 1},
		{100, 10},
		{500, 50},
	}
	for _, tt := range tests {
		s := gridSpacing(tt.viewWidth)
		if s > tt.wantMax {
			t.Errorf("gridSpacing(%f) = %f, want <= %f", tt.viewWidth, s, tt.wantMax)
		}
		if s <= 0 {
			t.Errorf("gridSpacing(%f) = %f, want > 0", tt.viewWidth, s)
		}
	}
}

func TestExportPNG(t *testing.T) {
	c := NewCanvas(200, 200)
	c.SetView(Viewport{0, 0, 100, 100})
	c.Clear()
	c.DrawGrid()
	c.DrawPaths([]importer.Path{{
		Points: []importer.Point{{10, 10}, {90, 90}},
		Layer:  "test",
	}})

	dir := t.TempDir()
	path := filepath.Join(dir, "test.png")

	if err := c.ExportPNG(path); err != nil {
		t.Fatalf("ExportPNG: %v", err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open PNG: %v", err)
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		t.Fatalf("decode PNG: %v", err)
	}
	bounds := img.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 200 {
		t.Errorf("PNG size %dx%d, want 200x200", bounds.Dx(), bounds.Dy())
	}
}

func TestExportPNGInvalidPath(t *testing.T) {
	c := NewCanvas(10, 10)
	err := c.ExportPNG("/nonexistent/dir/test.png")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestBresenhamHorizontal(t *testing.T) {
	c := NewCanvas(10, 10)
	c.Clear()
	col := color.RGBA{255, 0, 0, 255}
	bresenham(c.img, 0, 5, 9, 5, col)

	for x := 0; x < 10; x++ {
		if c.img.RGBAAt(x, 5) != col {
			t.Errorf("pixel (%d,5) not drawn", x)
		}
	}
}

func TestBresenhamVertical(t *testing.T) {
	c := NewCanvas(10, 10)
	c.Clear()
	col := color.RGBA{0, 255, 0, 255}
	bresenham(c.img, 5, 0, 5, 9, col)

	for y := 0; y < 10; y++ {
		if c.img.RGBAAt(5, y) != col {
			t.Errorf("pixel (5,%d) not drawn", y)
		}
	}
}

func TestBresenhamDiagonal(t *testing.T) {
	c := NewCanvas(10, 10)
	c.Clear()
	col := color.RGBA{0, 0, 255, 255}
	bresenham(c.img, 0, 0, 9, 9, col)

	for i := 0; i < 10; i++ {
		if c.img.RGBAAt(i, i) != col {
			t.Errorf("pixel (%d,%d) not drawn", i, i)
		}
	}
}

func TestBresenhamOutOfBounds(t *testing.T) {
	c := NewCanvas(10, 10)
	c.Clear()
	// Should not panic when line extends outside image
	bresenham(c.img, -5, -5, 15, 15, color.RGBA{255, 0, 0, 255})
}

func TestDrawEntityIntegration(t *testing.T) {
	ent := &importer.Entity{
		Paths: []importer.Path{
			{Points: []importer.Point{{0, 0}, {50, 0}, {50, 30}, {0, 30}}, Closed: true, Layer: "cut"},
			{Points: []importer.Point{{10, 10}, {40, 10}}, Layer: "engrave"},
		},
	}

	c := NewCanvas(400, 300)
	c.FitEntity(ent)
	c.Clear()
	c.DrawGrid()
	c.DrawEntity(ent)

	dir := t.TempDir()
	path := filepath.Join(dir, "integration.png")
	if err := c.ExportPNG(path); err != nil {
		t.Fatalf("ExportPNG: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Size() == 0 {
		t.Error("PNG file is empty")
	}
}

func TestViewGetSet(t *testing.T) {
	c := NewCanvas(100, 100)
	v := Viewport{-5, -5, 50, 50}
	c.SetView(v)

	got := c.View()
	if got != v {
		t.Errorf("View() = %v, want %v", got, v)
	}
}
