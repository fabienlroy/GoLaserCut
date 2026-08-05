package preview

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"

	"github.com/fabienlroy/GoLaserCut/importer"
)

var layerColors = []color.RGBA{
	{0, 150, 255, 255},
	{255, 50, 50, 255},
	{50, 200, 50, 255},
	{255, 200, 0, 255},
	{200, 50, 200, 255},
	{0, 200, 200, 255},
	{255, 128, 0, 255},
	{128, 128, 255, 255},
}

var (
	colorGrid    = color.RGBA{60, 60, 60, 255}
	colorGridSub = color.RGBA{40, 40, 40, 255}
	colorOrigin  = color.RGBA{80, 80, 80, 255}
	colorBg      = color.RGBA{30, 30, 30, 255}
	colorRapid   = color.RGBA{100, 100, 100, 128}
)

// Viewport defines the visible world-coordinate rectangle.
type Viewport struct {
	MinX, MinY, MaxX, MaxY float64
}

// Point3D is a 3D point for toolpath rendering.
type Point3D struct {
	X, Y, Z float64
}

// Toolpath3D is a 3D toolpath for preview rendering.
type Toolpath3D struct {
	Points []Point3D
	Closed bool
}

// Canvas renders 2D previews of laser paths and toolpaths.
type Canvas struct {
	Width, Height int
	Margin        float64
	img           *image.RGBA
	view          Viewport
	layerMap      map[string]int
}

func NewCanvas(width, height int) *Canvas {
	return &Canvas{
		Width:    width,
		Height:   height,
		Margin:   5.0,
		img:      image.NewRGBA(image.Rect(0, 0, width, height)),
		layerMap: map[string]int{},
	}
}

func (c *Canvas) SetView(v Viewport) {
	c.view = v
}

func (c *Canvas) View() Viewport {
	return c.view
}

// FitBounds sets the viewport to show the given bounds with margin,
// preserving aspect ratio.
func (c *Canvas) FitBounds(minX, minY, maxX, maxY float64) {
	m := c.Margin
	minX -= m
	minY -= m
	maxX += m
	maxY += m

	worldW := maxX - minX
	worldH := maxY - minY
	if worldW <= 0 || worldH <= 0 {
		c.view = Viewport{minX - 10, minY - 10, maxX + 10, maxY + 10}
		return
	}

	canvasAspect := float64(c.Width) / float64(c.Height)
	worldAspect := worldW / worldH

	if worldAspect > canvasAspect {
		newH := worldW / canvasAspect
		cy := (minY + maxY) / 2
		minY = cy - newH/2
		maxY = cy + newH/2
	} else {
		newW := worldH * canvasAspect
		cx := (minX + maxX) / 2
		minX = cx - newW/2
		maxX = cx + newW/2
	}

	c.view = Viewport{minX, minY, maxX, maxY}
}

// FitEntity sets the viewport to fit an entity set.
func (c *Canvas) FitEntity(ent *importer.Entity) {
	min, max := ent.Bounds()
	c.FitBounds(min.X, min.Y, max.X, max.Y)
}

// Clear fills the canvas with the background color.
func (c *Canvas) Clear() {
	for y := 0; y < c.Height; y++ {
		for x := 0; x < c.Width; x++ {
			c.img.SetRGBA(x, y, colorBg)
		}
	}
}

// DrawGrid draws a coordinate grid with automatic spacing.
func (c *Canvas) DrawGrid() {
	vw := c.view.MaxX - c.view.MinX
	if vw <= 0 {
		return
	}

	spacing := gridSpacing(vw)

	// Sub-grid
	sub := spacing / 5
	if sub > 0.01 {
		startX := math.Floor(c.view.MinX/sub) * sub
		for x := startX; x <= c.view.MaxX; x += sub {
			px, _ := c.worldToPixel(x, 0)
			c.drawVerticalLine(px, colorGridSub)
		}
		startY := math.Floor(c.view.MinY/sub) * sub
		for y := startY; y <= c.view.MaxY; y += sub {
			_, py := c.worldToPixel(0, y)
			c.drawHorizontalLine(py, colorGridSub)
		}
	}

	// Main grid
	startX := math.Floor(c.view.MinX/spacing) * spacing
	for x := startX; x <= c.view.MaxX; x += spacing {
		px, _ := c.worldToPixel(x, 0)
		c.drawVerticalLine(px, colorGrid)
	}
	startY := math.Floor(c.view.MinY/spacing) * spacing
	for y := startY; y <= c.view.MaxY; y += spacing {
		_, py := c.worldToPixel(0, y)
		c.drawHorizontalLine(py, colorGrid)
	}

	// Origin axes
	ox, oy := c.worldToPixel(0, 0)
	c.drawVerticalLine(ox, colorOrigin)
	c.drawHorizontalLine(oy, colorOrigin)
}

// DrawPaths draws 2D paths with per-layer colors.
func (c *Canvas) DrawPaths(paths []importer.Path) {
	for _, path := range paths {
		col := c.layerColor(path.Layer)
		c.drawPath(path, col)
	}
}

// DrawEntity draws all paths from an entity set.
func (c *Canvas) DrawEntity(ent *importer.Entity) {
	c.DrawPaths(ent.Paths)
}

// DrawToolpaths draws 3D toolpaths with Z-depth coloring.
// Shallow cuts are cyan, deep cuts are red.
func (c *Canvas) DrawToolpaths(toolpaths []Toolpath3D, maxDepth float64) {
	for _, tp := range toolpaths {
		for i := 0; i < len(tp.Points)-1; i++ {
			col := depthColor(tp.Points[i].Z, maxDepth)
			c.drawLine3D(tp.Points[i], tp.Points[i+1], col)
		}
		if tp.Closed && len(tp.Points) > 2 {
			col := depthColor(tp.Points[len(tp.Points)-1].Z, maxDepth)
			c.drawLine3D(tp.Points[len(tp.Points)-1], tp.Points[0], col)
		}
	}
}

// Image returns the rendered image.
func (c *Canvas) Image() *image.RGBA {
	return c.img
}

// ExportPNG writes the canvas to a PNG file.
func (c *Canvas) ExportPNG(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("creating %s: %w", filename, err)
	}
	defer f.Close()
	return png.Encode(f, c.img)
}

func (c *Canvas) worldToPixel(wx, wy float64) (int, int) {
	vw := c.view.MaxX - c.view.MinX
	vh := c.view.MaxY - c.view.MinY
	if vw <= 0 || vh <= 0 {
		return 0, 0
	}
	px := int((wx - c.view.MinX) / vw * float64(c.Width-1))
	py := int((c.view.MaxY - wy) / vh * float64(c.Height-1))
	return px, py
}

func (c *Canvas) pixelToWorld(px, py int) (float64, float64) {
	vw := c.view.MaxX - c.view.MinX
	vh := c.view.MaxY - c.view.MinY
	wx := c.view.MinX + float64(px)/float64(c.Width-1)*vw
	wy := c.view.MaxY - float64(py)/float64(c.Height-1)*vh
	return wx, wy
}

func (c *Canvas) layerColor(layer string) color.RGBA {
	if idx, ok := c.layerMap[layer]; ok {
		return layerColors[idx%len(layerColors)]
	}
	idx := len(c.layerMap)
	c.layerMap[layer] = idx
	return layerColors[idx%len(layerColors)]
}

func (c *Canvas) drawPath(path importer.Path, col color.RGBA) {
	for i := 0; i < len(path.Points)-1; i++ {
		c.drawLine2D(path.Points[i], path.Points[i+1], col)
	}
	if path.Closed && len(path.Points) > 2 {
		c.drawLine2D(path.Points[len(path.Points)-1], path.Points[0], col)
	}
}

func (c *Canvas) drawLine2D(from, to importer.Point, col color.RGBA) {
	x0, y0 := c.worldToPixel(from.X, from.Y)
	x1, y1 := c.worldToPixel(to.X, to.Y)
	bresenham(c.img, x0, y0, x1, y1, col)
}

func (c *Canvas) drawLine3D(from, to Point3D, col color.RGBA) {
	x0, y0 := c.worldToPixel(from.X, from.Y)
	x1, y1 := c.worldToPixel(to.X, to.Y)
	bresenham(c.img, x0, y0, x1, y1, col)
}

func (c *Canvas) drawVerticalLine(px int, col color.RGBA) {
	if px < 0 || px >= c.Width {
		return
	}
	for y := 0; y < c.Height; y++ {
		c.img.SetRGBA(px, y, col)
	}
}

func (c *Canvas) drawHorizontalLine(py int, col color.RGBA) {
	if py < 0 || py >= c.Height {
		return
	}
	for x := 0; x < c.Width; x++ {
		c.img.SetRGBA(x, py, col)
	}
}

func bresenham(img *image.RGBA, x0, y0, x1, y1 int, col color.RGBA) {
	bounds := img.Bounds()

	abs := func(v int) int {
		if v < 0 {
			return -v
		}
		return v
	}

	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx := 1
	if x0 > x1 {
		sx = -1
	}
	sy := 1
	if y0 > y1 {
		sy = -1
	}
	err := dx - dy

	for {
		if x0 >= bounds.Min.X && x0 < bounds.Max.X && y0 >= bounds.Min.Y && y0 < bounds.Max.Y {
			img.SetRGBA(x0, y0, col)
		}
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

// depthColor maps Z depth to a color gradient.
// Z=0 (surface) → cyan, Z=-maxDepth → red.
func depthColor(z, maxDepth float64) color.RGBA {
	if maxDepth <= 0 {
		return layerColors[0]
	}
	t := math.Abs(z) / maxDepth
	if t > 1 {
		t = 1
	}
	r := uint8(t * 255)
	g := uint8((1 - t) * 100)
	b := uint8((1 - t) * 255)
	return color.RGBA{r, g, b, 255}
}

func gridSpacing(viewWidth float64) float64 {
	steps := []float64{0.1, 0.5, 1, 5, 10, 50, 100, 500, 1000}
	for _, s := range steps {
		if viewWidth/s <= 20 {
			return s
		}
	}
	return 1000
}
