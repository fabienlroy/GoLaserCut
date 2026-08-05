package importer

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

type dxfPair struct {
	code  int
	value string
}

func ParseDXF(filename string) (*Entity, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", filename, err)
	}
	defer f.Close()

	pairs, err := readDXFPairs(f)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", filename, err)
	}

	return parseDXFEntities(pairs), nil
}

func readDXFPairs(f *os.File) ([]dxfPair, error) {
	scanner := bufio.NewScanner(f)
	var pairs []dxfPair
	for scanner.Scan() {
		codeLine := strings.TrimSpace(scanner.Text())
		if !scanner.Scan() {
			break
		}
		valueLine := strings.TrimSpace(scanner.Text())
		code, err := strconv.Atoi(codeLine)
		if err != nil {
			continue
		}
		pairs = append(pairs, dxfPair{code: code, value: valueLine})
	}
	return pairs, scanner.Err()
}

func parseDXFEntities(pairs []dxfPair) *Entity {
	ent := &Entity{}

	inEntities := false
	i := 0
	for i < len(pairs) {
		p := pairs[i]
		if p.code == 0 && p.value == "ENDSEC" && inEntities {
			break
		}
		if p.code == 2 && p.value == "ENTITIES" {
			inEntities = true
			i++
			continue
		}
		if !inEntities {
			i++
			continue
		}

		if p.code != 0 {
			i++
			continue
		}

		switch p.value {
		case "LINE":
			i++
			i = parseDXFLine(pairs, i, ent)
		case "CIRCLE":
			i++
			i = parseDXFCircle(pairs, i, ent)
		case "ARC":
			i++
			i = parseDXFArc(pairs, i, ent)
		case "LWPOLYLINE":
			i++
			i = parseDXFLWPolyline(pairs, i, ent)
		default:
			i++
		}
	}
	return ent
}

func parseDXFLine(pairs []dxfPair, i int, ent *Entity) int {
	var x0, y0, x1, y1 float64
	layer := "0"
	for i < len(pairs) && pairs[i].code != 0 {
		switch pairs[i].code {
		case 8:
			layer = pairs[i].value
		case 10:
			x0 = parseFloat(pairs[i].value)
		case 20:
			y0 = parseFloat(pairs[i].value)
		case 11:
			x1 = parseFloat(pairs[i].value)
		case 21:
			y1 = parseFloat(pairs[i].value)
		}
		i++
	}
	ent.Paths = append(ent.Paths, Path{
		Points: []Point{{x0, y0}, {x1, y1}},
		Layer:  layer,
	})
	return i
}

func parseDXFCircle(pairs []dxfPair, i int, ent *Entity) int {
	var cx, cy, r float64
	layer := "0"
	for i < len(pairs) && pairs[i].code != 0 {
		switch pairs[i].code {
		case 8:
			layer = pairs[i].value
		case 10:
			cx = parseFloat(pairs[i].value)
		case 20:
			cy = parseFloat(pairs[i].value)
		case 40:
			r = parseFloat(pairs[i].value)
		}
		i++
	}
	pts := circleToPoints(Point{cx, cy}, r, 64)
	ent.Paths = append(ent.Paths, Path{
		Points: pts,
		Closed: true,
		Layer:  layer,
	})
	return i
}

func parseDXFArc(pairs []dxfPair, i int, ent *Entity) int {
	var cx, cy, r, startDeg, endDeg float64
	layer := "0"
	for i < len(pairs) && pairs[i].code != 0 {
		switch pairs[i].code {
		case 8:
			layer = pairs[i].value
		case 10:
			cx = parseFloat(pairs[i].value)
		case 20:
			cy = parseFloat(pairs[i].value)
		case 40:
			r = parseFloat(pairs[i].value)
		case 50:
			startDeg = parseFloat(pairs[i].value)
		case 51:
			endDeg = parseFloat(pairs[i].value)
		}
		i++
	}
	startRad := degToRad(startDeg)
	endRad := degToRad(endDeg)
	pts := arcToPoints(Point{cx, cy}, r, startRad, endRad, 32)
	ent.Paths = append(ent.Paths, Path{
		Points: pts,
		Layer:  layer,
	})
	return i
}

func parseDXFLWPolyline(pairs []dxfPair, i int, ent *Entity) int {
	layer := "0"
	closed := false
	var xs, ys, bulges []float64

	for i < len(pairs) && pairs[i].code != 0 {
		switch pairs[i].code {
		case 8:
			layer = pairs[i].value
		case 70:
			flags := parseInt(pairs[i].value)
			closed = (flags & 1) != 0
		case 10:
			xs = append(xs, parseFloat(pairs[i].value))
			bulges = append(bulges, 0)
		case 20:
			ys = append(ys, parseFloat(pairs[i].value))
		case 42:
			if len(bulges) > 0 {
				bulges[len(bulges)-1] = parseFloat(pairs[i].value)
			}
		}
		i++
	}

	n := len(xs)
	if n == 0 || len(ys) < n {
		return i
	}

	var points []Point
	for j := 0; j < n; j++ {
		from := Point{xs[j], ys[j]}
		points = append(points, from)

		nextJ := j + 1
		if nextJ >= n {
			if !closed {
				break
			}
			nextJ = 0
		}
		to := Point{xs[nextJ], ys[nextJ]}
		bulge := bulges[j]

		if math.Abs(bulge) > 1e-10 {
			arcPts := bulgeToArc(from, to, bulge)
			points = append(points, arcPts[1:]...)
		}
	}

	ent.Paths = append(ent.Paths, Path{
		Points: points,
		Closed: closed,
		Layer:  layer,
	})
	return i
}

// bulgeToArc converts a bulge segment between two points into arc points.
// bulge = tan(included_angle / 4). Positive bulge = arc to the left of
// travel direction (visually CCW), negative = to the right.
func bulgeToArc(from, to Point, bulge float64) []Point {
	dx := to.X - from.X
	dy := to.Y - from.Y
	chord := math.Sqrt(dx*dx + dy*dy)
	if chord < 1e-12 {
		return []Point{from, to}
	}

	theta := 4 * math.Atan(math.Abs(bulge))
	radius := chord / (2 * math.Sin(theta/2))
	d := radius * math.Cos(theta/2)

	midX := (from.X + to.X) / 2
	midY := (from.Y + to.Y) / 2

	// Left-normal of travel direction
	nx := -dy / chord
	ny := dx / chord

	var center Point
	if bulge > 0 {
		center = Point{midX - d*nx, midY - d*ny}
	} else {
		center = Point{midX + d*nx, midY + d*ny}
	}

	startAngle := math.Atan2(from.Y-center.Y, from.X-center.X)
	endAngle := math.Atan2(to.Y-center.Y, to.X-center.X)

	sweep := endAngle - startAngle
	if bulge > 0 {
		if sweep > 0 {
			sweep -= 2 * math.Pi
		}
	} else {
		if sweep < 0 {
			sweep += 2 * math.Pi
		}
	}

	segments := int(math.Ceil(math.Abs(sweep) / (math.Pi / 16)))
	if segments < 2 {
		segments = 2
	}

	pts := make([]Point, segments+1)
	for k := 0; k <= segments; k++ {
		t := float64(k) / float64(segments)
		a := startAngle + t*sweep
		pts[k] = Point{
			X: center.X + radius*math.Cos(a),
			Y: center.Y + radius*math.Sin(a),
		}
	}
	return pts
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return v
}

func parseInt(s string) int {
	v, _ := strconv.Atoi(strings.TrimSpace(s))
	return v
}
