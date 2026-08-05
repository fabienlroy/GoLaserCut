package importer

import "math"

type Point struct {
	X, Y float64
}

type Segment struct {
	From, To Point
}

type Arc struct {
	Center     Point
	Radius     float64
	StartAngle float64 // radians
	EndAngle   float64 // radians
}

// Path is a sequence of points forming an open or closed polyline.
type Path struct {
	Points []Point
	Closed bool
	Layer  string
}

// Entity groups all geometry types a file can produce.
type Entity struct {
	Paths   []Path
	Arcs    []Arc
	Circles []Arc
}

// Bounds returns the bounding box of an entity set.
func (e *Entity) Bounds() (min, max Point) {
	first := true
	expand := func(p Point) {
		if first {
			min, max = p, p
			first = false
			return
		}
		if p.X < min.X {
			min.X = p.X
		}
		if p.Y < min.Y {
			min.Y = p.Y
		}
		if p.X > max.X {
			max.X = p.X
		}
		if p.Y > max.Y {
			max.Y = p.Y
		}
	}
	for _, path := range e.Paths {
		for _, p := range path.Points {
			expand(p)
		}
	}
	for _, a := range e.Arcs {
		expand(Point{a.Center.X - a.Radius, a.Center.Y - a.Radius})
		expand(Point{a.Center.X + a.Radius, a.Center.Y + a.Radius})
	}
	for _, c := range e.Circles {
		expand(Point{c.Center.X - c.Radius, c.Center.Y - c.Radius})
		expand(Point{c.Center.X + c.Radius, c.Center.Y + c.Radius})
	}
	return
}

func degToRad(d float64) float64 { return d * math.Pi / 180 }
func radToDeg(r float64) float64 { return r * 180 / math.Pi }

// arcToPoints approximates a circular arc as a polyline.
func arcToPoints(center Point, radius, startAngle, endAngle float64, segments int) []Point {
	if segments < 2 {
		segments = 2
	}
	sweep := endAngle - startAngle
	if sweep < 0 {
		sweep += 2 * math.Pi
	}
	pts := make([]Point, segments+1)
	for i := 0; i <= segments; i++ {
		t := float64(i) / float64(segments)
		a := startAngle + t*sweep
		pts[i] = Point{
			X: center.X + radius*math.Cos(a),
			Y: center.Y + radius*math.Sin(a),
		}
	}
	return pts
}

// circleToPoints approximates a full circle as a closed polyline.
func circleToPoints(center Point, radius float64, segments int) []Point {
	if segments < 8 {
		segments = 8
	}
	pts := make([]Point, segments)
	for i := 0; i < segments; i++ {
		a := 2 * math.Pi * float64(i) / float64(segments)
		pts[i] = Point{
			X: center.X + radius*math.Cos(a),
			Y: center.Y + radius*math.Sin(a),
		}
	}
	return pts
}
