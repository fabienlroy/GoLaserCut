package importer

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

func writeTempSVG(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.svg")
	os.WriteFile(path, []byte(content), 0644)
	return path
}

const svgLine = `<svg xmlns="http://www.w3.org/2000/svg">
  <line x1="0" y1="0" x2="100" y2="50"/>
</svg>`

func TestSVGLine(t *testing.T) {
	ent, err := ParseSVG(writeTempSVG(t, svgLine))
	if err != nil {
		t.Fatal(err)
	}
	if len(ent.Paths) != 1 {
		t.Fatalf("got %d paths, want 1", len(ent.Paths))
	}
	p := ent.Paths[0]
	if len(p.Points) != 2 {
		t.Fatalf("got %d points, want 2", len(p.Points))
	}
	if !closeTo(p.Points[0].X, 0, tolerance) || !closeTo(p.Points[0].Y, 0, tolerance) {
		t.Errorf("start = (%f,%f), want (0,0)", p.Points[0].X, p.Points[0].Y)
	}
	if !closeTo(p.Points[1].X, 100, tolerance) || !closeTo(p.Points[1].Y, 50, tolerance) {
		t.Errorf("end = (%f,%f), want (100,50)", p.Points[1].X, p.Points[1].Y)
	}
}

const svgRect = `<svg xmlns="http://www.w3.org/2000/svg">
  <rect x="10" y="20" width="100" height="50"/>
</svg>`

func TestSVGRect(t *testing.T) {
	ent, err := ParseSVG(writeTempSVG(t, svgRect))
	if err != nil {
		t.Fatal(err)
	}
	if len(ent.Paths) != 1 {
		t.Fatalf("got %d paths, want 1", len(ent.Paths))
	}
	p := ent.Paths[0]
	if !p.Closed {
		t.Error("rect should be closed")
	}
	if len(p.Points) != 4 {
		t.Fatalf("got %d points, want 4", len(p.Points))
	}
}

const svgCircle = `<svg xmlns="http://www.w3.org/2000/svg">
  <circle cx="50" cy="50" r="25"/>
</svg>`

func TestSVGCircle(t *testing.T) {
	ent, err := ParseSVG(writeTempSVG(t, svgCircle))
	if err != nil {
		t.Fatal(err)
	}
	if len(ent.Paths) != 1 {
		t.Fatalf("got %d paths, want 1", len(ent.Paths))
	}
	p := ent.Paths[0]
	if !p.Closed {
		t.Error("circle should be closed")
	}
	for i, pt := range p.Points {
		d := math.Sqrt((pt.X-50)*(pt.X-50) + (pt.Y-50)*(pt.Y-50))
		if !closeTo(d, 25, tolerance) {
			t.Errorf("point %d distance from center = %f, want 25", i, d)
		}
	}
}

const svgEllipse = `<svg xmlns="http://www.w3.org/2000/svg">
  <ellipse cx="50" cy="50" rx="30" ry="20"/>
</svg>`

func TestSVGEllipse(t *testing.T) {
	ent, err := ParseSVG(writeTempSVG(t, svgEllipse))
	if err != nil {
		t.Fatal(err)
	}
	if len(ent.Paths) != 1 {
		t.Fatalf("got %d paths, want 1", len(ent.Paths))
	}
	p := ent.Paths[0]
	if !p.Closed {
		t.Error("ellipse should be closed")
	}
	min, max := ent.Bounds()
	if !closeTo(min.X, 20, 0.5) || !closeTo(min.Y, 30, 0.5) {
		t.Errorf("min = (%f,%f), want near (20,30)", min.X, min.Y)
	}
	if !closeTo(max.X, 80, 0.5) || !closeTo(max.Y, 70, 0.5) {
		t.Errorf("max = (%f,%f), want near (80,70)", max.X, max.Y)
	}
}

const svgPolygon = `<svg xmlns="http://www.w3.org/2000/svg">
  <polygon points="0,0 100,0 100,100 0,100"/>
</svg>`

func TestSVGPolygon(t *testing.T) {
	ent, err := ParseSVG(writeTempSVG(t, svgPolygon))
	if err != nil {
		t.Fatal(err)
	}
	p := ent.Paths[0]
	if !p.Closed {
		t.Error("polygon should be closed")
	}
	if len(p.Points) != 4 {
		t.Fatalf("got %d points, want 4", len(p.Points))
	}
}

const svgPolyline = `<svg xmlns="http://www.w3.org/2000/svg">
  <polyline points="0,0 50,50 100,0"/>
</svg>`

func TestSVGPolyline(t *testing.T) {
	ent, err := ParseSVG(writeTempSVG(t, svgPolyline))
	if err != nil {
		t.Fatal(err)
	}
	p := ent.Paths[0]
	if p.Closed {
		t.Error("polyline should not be closed")
	}
	if len(p.Points) != 3 {
		t.Fatalf("got %d points, want 3", len(p.Points))
	}
}

const svgPathMLZ = `<svg xmlns="http://www.w3.org/2000/svg">
  <path d="M 0,0 L 100,0 L 100,100 Z"/>
</svg>`

func TestSVGPathLines(t *testing.T) {
	ent, err := ParseSVG(writeTempSVG(t, svgPathMLZ))
	if err != nil {
		t.Fatal(err)
	}
	if len(ent.Paths) != 1 {
		t.Fatalf("got %d paths, want 1", len(ent.Paths))
	}
	p := ent.Paths[0]
	if !p.Closed {
		t.Error("path with Z should be closed")
	}
	if len(p.Points) != 3 {
		t.Fatalf("got %d points, want 3", len(p.Points))
	}
}

const svgPathHV = `<svg xmlns="http://www.w3.org/2000/svg">
  <path d="M 0,0 H 100 V 50"/>
</svg>`

func TestSVGPathHV(t *testing.T) {
	ent, err := ParseSVG(writeTempSVG(t, svgPathHV))
	if err != nil {
		t.Fatal(err)
	}
	p := ent.Paths[0]
	if len(p.Points) != 3 {
		t.Fatalf("got %d points, want 3", len(p.Points))
	}
	if !closeTo(p.Points[1].X, 100, tolerance) || !closeTo(p.Points[1].Y, 0, tolerance) {
		t.Errorf("H point = (%f,%f), want (100,0)", p.Points[1].X, p.Points[1].Y)
	}
	if !closeTo(p.Points[2].X, 100, tolerance) || !closeTo(p.Points[2].Y, 50, tolerance) {
		t.Errorf("V point = (%f,%f), want (100,50)", p.Points[2].X, p.Points[2].Y)
	}
}

const svgPathRelative = `<svg xmlns="http://www.w3.org/2000/svg">
  <path d="M 10,10 l 20,0 l 0,20 l -20,0 z"/>
</svg>`

func TestSVGPathRelative(t *testing.T) {
	ent, err := ParseSVG(writeTempSVG(t, svgPathRelative))
	if err != nil {
		t.Fatal(err)
	}
	p := ent.Paths[0]
	if !p.Closed {
		t.Error("should be closed")
	}
	if len(p.Points) != 4 {
		t.Fatalf("got %d points, want 4", len(p.Points))
	}
	if !closeTo(p.Points[3].X, 10, tolerance) || !closeTo(p.Points[3].Y, 30, tolerance) {
		t.Errorf("last point = (%f,%f), want (10,30)", p.Points[3].X, p.Points[3].Y)
	}
}

const svgPathCubic = `<svg xmlns="http://www.w3.org/2000/svg">
  <path d="M 0,0 C 10,20 30,20 40,0"/>
</svg>`

func TestSVGPathCubic(t *testing.T) {
	ent, err := ParseSVG(writeTempSVG(t, svgPathCubic))
	if err != nil {
		t.Fatal(err)
	}
	p := ent.Paths[0]
	if len(p.Points) < 3 {
		t.Errorf("cubic should produce multiple points, got %d", len(p.Points))
	}
	first := p.Points[0]
	last := p.Points[len(p.Points)-1]
	if !closeTo(first.X, 0, tolerance) || !closeTo(first.Y, 0, tolerance) {
		t.Errorf("start = (%f,%f), want (0,0)", first.X, first.Y)
	}
	if !closeTo(last.X, 40, tolerance) || !closeTo(last.Y, 0, tolerance) {
		t.Errorf("end = (%f,%f), want (40,0)", last.X, last.Y)
	}
}

const svgPathArc = `<svg xmlns="http://www.w3.org/2000/svg">
  <path d="M 0,0 A 25,25 0 0,1 50,0"/>
</svg>`

func TestSVGPathArc(t *testing.T) {
	ent, err := ParseSVG(writeTempSVG(t, svgPathArc))
	if err != nil {
		t.Fatal(err)
	}
	p := ent.Paths[0]
	if len(p.Points) < 3 {
		t.Errorf("arc should produce multiple points, got %d", len(p.Points))
	}
	first := p.Points[0]
	last := p.Points[len(p.Points)-1]
	if !closeTo(first.X, 0, tolerance) || !closeTo(first.Y, 0, tolerance) {
		t.Errorf("start = (%f,%f), want (0,0)", first.X, first.Y)
	}
	if !closeTo(last.X, 50, tolerance) || !closeTo(last.Y, 0, tolerance) {
		t.Errorf("end = (%f,%f), want (50,0)", last.X, last.Y)
	}
}

const svgTranslate = `<svg xmlns="http://www.w3.org/2000/svg">
  <g transform="translate(100,200)">
    <line x1="0" y1="0" x2="10" y2="10"/>
  </g>
</svg>`

func TestSVGTranslate(t *testing.T) {
	ent, err := ParseSVG(writeTempSVG(t, svgTranslate))
	if err != nil {
		t.Fatal(err)
	}
	p := ent.Paths[0]
	if !closeTo(p.Points[0].X, 100, tolerance) || !closeTo(p.Points[0].Y, 200, tolerance) {
		t.Errorf("translated start = (%f,%f), want (100,200)", p.Points[0].X, p.Points[0].Y)
	}
	if !closeTo(p.Points[1].X, 110, tolerance) || !closeTo(p.Points[1].Y, 210, tolerance) {
		t.Errorf("translated end = (%f,%f), want (110,210)", p.Points[1].X, p.Points[1].Y)
	}
}

const svgScale = `<svg xmlns="http://www.w3.org/2000/svg">
  <g transform="scale(2)">
    <line x1="10" y1="20" x2="30" y2="40"/>
  </g>
</svg>`

func TestSVGScale(t *testing.T) {
	ent, err := ParseSVG(writeTempSVG(t, svgScale))
	if err != nil {
		t.Fatal(err)
	}
	p := ent.Paths[0]
	if !closeTo(p.Points[0].X, 20, tolerance) || !closeTo(p.Points[0].Y, 40, tolerance) {
		t.Errorf("scaled start = (%f,%f), want (20,40)", p.Points[0].X, p.Points[0].Y)
	}
}

const svgMultiplePaths = `<svg xmlns="http://www.w3.org/2000/svg">
  <path d="M 0,0 L 10,0 M 20,0 L 30,0"/>
</svg>`

func TestSVGMultiplePaths(t *testing.T) {
	ent, err := ParseSVG(writeTempSVG(t, svgMultiplePaths))
	if err != nil {
		t.Fatal(err)
	}
	if len(ent.Paths) != 2 {
		t.Fatalf("got %d paths, want 2", len(ent.Paths))
	}
}

func TestSVGFileNotFound(t *testing.T) {
	_, err := ParseSVG("/nonexistent/file.svg")
	if err == nil {
		t.Error("expected error for missing file")
	}
}
