package importer

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

const tolerance = 0.01

func closeTo(a, b, tol float64) bool {
	return math.Abs(a-b) < tol
}

func writeTempDXF(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.dxf")
	os.WriteFile(path, []byte(content), 0644)
	return path
}

const dxfLine = `0
SECTION
2
ENTITIES
0
LINE
8
cuts
10
0.0
20
0.0
11
100.0
21
50.0
0
ENDSEC
0
EOF
`

func TestDXFLine(t *testing.T) {
	ent, err := ParseDXF(writeTempDXF(t, dxfLine))
	if err != nil {
		t.Fatal(err)
	}
	if len(ent.Paths) != 1 {
		t.Fatalf("got %d paths, want 1", len(ent.Paths))
	}
	p := ent.Paths[0]
	if p.Layer != "cuts" {
		t.Errorf("layer = %q, want %q", p.Layer, "cuts")
	}
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

const dxfCircle = `0
SECTION
2
ENTITIES
0
CIRCLE
8
holes
10
50.0
20
50.0
40
25.0
0
ENDSEC
0
EOF
`

func TestDXFCircle(t *testing.T) {
	ent, err := ParseDXF(writeTempDXF(t, dxfCircle))
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
	if p.Layer != "holes" {
		t.Errorf("layer = %q, want %q", p.Layer, "holes")
	}
	if len(p.Points) < 8 {
		t.Errorf("circle has %d points, want >= 8", len(p.Points))
	}
	// All points should be ~25 from center (50,50)
	for i, pt := range p.Points {
		d := math.Sqrt((pt.X-50)*(pt.X-50) + (pt.Y-50)*(pt.Y-50))
		if !closeTo(d, 25, tolerance) {
			t.Errorf("point %d distance from center = %f, want 25", i, d)
		}
	}
}

const dxfArc = `0
SECTION
2
ENTITIES
0
ARC
8
0
10
0.0
20
0.0
40
10.0
50
0.0
51
90.0
0
ENDSEC
0
EOF
`

func TestDXFArc(t *testing.T) {
	ent, err := ParseDXF(writeTempDXF(t, dxfArc))
	if err != nil {
		t.Fatal(err)
	}
	if len(ent.Paths) != 1 {
		t.Fatalf("got %d paths, want 1", len(ent.Paths))
	}
	p := ent.Paths[0]
	if p.Closed {
		t.Error("arc should not be closed")
	}
	// First point: (10, 0), last point: (0, 10)
	first := p.Points[0]
	last := p.Points[len(p.Points)-1]
	if !closeTo(first.X, 10, tolerance) || !closeTo(first.Y, 0, tolerance) {
		t.Errorf("arc start = (%f,%f), want (10,0)", first.X, first.Y)
	}
	if !closeTo(last.X, 0, tolerance) || !closeTo(last.Y, 10, tolerance) {
		t.Errorf("arc end = (%f,%f), want (0,10)", last.X, last.Y)
	}
}

const dxfLWPolylineOpen = `0
SECTION
2
ENTITIES
0
LWPOLYLINE
8
outline
70
0
90
3
10
0.0
20
0.0
42
0.0
10
100.0
20
0.0
42
0.0
10
100.0
20
100.0
42
0.0
0
ENDSEC
0
EOF
`

func TestDXFLWPolylineOpen(t *testing.T) {
	ent, err := ParseDXF(writeTempDXF(t, dxfLWPolylineOpen))
	if err != nil {
		t.Fatal(err)
	}
	if len(ent.Paths) != 1 {
		t.Fatalf("got %d paths, want 1", len(ent.Paths))
	}
	p := ent.Paths[0]
	if p.Closed {
		t.Error("should be open")
	}
	if p.Layer != "outline" {
		t.Errorf("layer = %q, want %q", p.Layer, "outline")
	}
	if len(p.Points) != 3 {
		t.Errorf("got %d points, want 3", len(p.Points))
	}
}

const dxfLWPolylineClosed = `0
SECTION
2
ENTITIES
0
LWPOLYLINE
8
0
70
1
90
4
10
0.0
20
0.0
42
0.0
10
100.0
20
0.0
42
0.0
10
100.0
20
100.0
42
0.0
10
0.0
20
100.0
42
0.0
0
ENDSEC
0
EOF
`

func TestDXFLWPolylineClosed(t *testing.T) {
	ent, err := ParseDXF(writeTempDXF(t, dxfLWPolylineClosed))
	if err != nil {
		t.Fatal(err)
	}
	p := ent.Paths[0]
	if !p.Closed {
		t.Error("should be closed")
	}
	if len(p.Points) < 4 {
		t.Errorf("got %d points, want >= 4", len(p.Points))
	}
}

const dxfLWPolylineBulge = `0
SECTION
2
ENTITIES
0
LWPOLYLINE
8
0
70
0
90
2
10
0.0
20
0.0
42
1.0
10
10.0
20
0.0
42
0.0
0
ENDSEC
0
EOF
`

func TestDXFLWPolylineBulge(t *testing.T) {
	ent, err := ParseDXF(writeTempDXF(t, dxfLWPolylineBulge))
	if err != nil {
		t.Fatal(err)
	}
	p := ent.Paths[0]
	// bulge=1.0 means semicircle (tan(pi/4)=1), so arc points should exist
	if len(p.Points) < 3 {
		t.Errorf("bulge arc: got %d points, want >= 3", len(p.Points))
	}
	// The midpoint of the arc should be above the chord (CCW, bulge > 0)
	maxY := 0.0
	for _, pt := range p.Points {
		if pt.Y > maxY {
			maxY = pt.Y
		}
	}
	if maxY < 4.0 {
		t.Errorf("bulge arc maxY = %f, expected > 4 for semicircle", maxY)
	}
}

const dxfMultipleEntities = `0
SECTION
2
ENTITIES
0
LINE
8
layer1
10
0.0
20
0.0
11
10.0
21
10.0
0
LINE
8
layer2
10
20.0
20
20.0
11
30.0
21
30.0
0
CIRCLE
8
layer1
10
50.0
20
50.0
40
5.0
0
ENDSEC
0
EOF
`

func TestDXFMultipleEntities(t *testing.T) {
	ent, err := ParseDXF(writeTempDXF(t, dxfMultipleEntities))
	if err != nil {
		t.Fatal(err)
	}
	if len(ent.Paths) != 3 {
		t.Fatalf("got %d paths, want 3", len(ent.Paths))
	}
	if ent.Paths[0].Layer != "layer1" {
		t.Errorf("path 0 layer = %q, want layer1", ent.Paths[0].Layer)
	}
	if ent.Paths[1].Layer != "layer2" {
		t.Errorf("path 1 layer = %q, want layer2", ent.Paths[1].Layer)
	}
	if ent.Paths[2].Layer != "layer1" {
		t.Errorf("path 2 layer = %q, want layer1", ent.Paths[2].Layer)
	}
}

func TestDXFFileNotFound(t *testing.T) {
	_, err := ParseDXF("/nonexistent/file.dxf")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestDXFBounds(t *testing.T) {
	ent, err := ParseDXF(writeTempDXF(t, dxfMultipleEntities))
	if err != nil {
		t.Fatal(err)
	}
	min, max := ent.Bounds()
	if min.X > 0.1 || min.Y > 0.1 {
		t.Errorf("min = (%f,%f), want near (0,0)", min.X, min.Y)
	}
	if max.X < 54 || max.Y < 54 {
		t.Errorf("max = (%f,%f), want near (55,55)", max.X, max.Y)
	}
}
