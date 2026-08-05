package importer

import (
	"encoding/xml"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"unicode"
)

func ParseSVG(filename string) (*Entity, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", filename, err)
	}
	defer f.Close()

	var svg svgRoot
	if err := xml.NewDecoder(f).Decode(&svg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", filename, err)
	}

	ent := &Entity{}
	xf := identityMatrix()
	parseSVGElements(svg.Children, xf, ent)
	return ent, nil
}

type svgRoot struct {
	XMLName  xml.Name     `xml:"svg"`
	Children []svgElement `xml:",any"`
}

type svgElement struct {
	XMLName   xml.Name     `xml:""`
	Attrs     []xml.Attr   `xml:",any,attr"`
	Children  []svgElement `xml:",any"`
}

func (e svgElement) attr(name string) string {
	for _, a := range e.Attrs {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}

func (e svgElement) floatAttr(name string) float64 {
	v, _ := strconv.ParseFloat(e.attr(name), 64)
	return v
}

// Affine 2D matrix [a b tx; c d ty; 0 0 1] stored as [a, b, c, d, tx, ty]
type matrix [6]float64

func identityMatrix() matrix { return matrix{1, 0, 0, 1, 0, 0} }

func (m matrix) apply(p Point) Point {
	return Point{
		X: m[0]*p.X + m[1]*p.Y + m[4],
		Y: m[2]*p.X + m[3]*p.Y + m[5],
	}
}

func (m matrix) mul(n matrix) matrix {
	return matrix{
		m[0]*n[0] + m[1]*n[2],
		m[0]*n[1] + m[1]*n[3],
		m[2]*n[0] + m[3]*n[2],
		m[2]*n[1] + m[3]*n[3],
		m[0]*n[4] + m[1]*n[5] + m[4],
		m[2]*n[4] + m[3]*n[5] + m[5],
	}
}

func parseTransform(s string) matrix {
	m := identityMatrix()
	s = strings.TrimSpace(s)
	for len(s) > 0 {
		var fn string
		var args []float64
		i := strings.IndexByte(s, '(')
		if i < 0 {
			break
		}
		fn = strings.TrimSpace(s[:i])
		j := strings.IndexByte(s[i:], ')')
		if j < 0 {
			break
		}
		args = parseNumbers(s[i+1 : i+j])
		s = strings.TrimSpace(s[i+j+1:])

		switch fn {
		case "translate":
			tx := getNum(args, 0)
			ty := getNum(args, 1)
			m = m.mul(matrix{1, 0, 0, 1, tx, ty})
		case "scale":
			sx := getNum(args, 0)
			sy := sx
			if len(args) > 1 {
				sy = args[1]
			}
			m = m.mul(matrix{sx, 0, 0, sy, 0, 0})
		case "rotate":
			deg := getNum(args, 0)
			rad := deg * math.Pi / 180
			cos, sin := math.Cos(rad), math.Sin(rad)
			if len(args) >= 3 {
				cx, cy := args[1], args[2]
				m = m.mul(matrix{1, 0, 0, 1, cx, cy})
				m = m.mul(matrix{cos, -sin, sin, cos, 0, 0})
				m = m.mul(matrix{1, 0, 0, 1, -cx, -cy})
			} else {
				m = m.mul(matrix{cos, -sin, sin, cos, 0, 0})
			}
		case "matrix":
			if len(args) >= 6 {
				m = m.mul(matrix{args[0], args[2], args[1], args[3], args[4], args[5]})
			}
		}
	}
	return m
}

func getNum(nums []float64, i int) float64 {
	if i < len(nums) {
		return nums[i]
	}
	return 0
}

func parseSVGElements(elements []svgElement, xf matrix, ent *Entity) {
	for _, el := range elements {
		tf := xf
		if t := el.attr("transform"); t != "" {
			tf = xf.mul(parseTransform(t))
		}

		switch el.XMLName.Local {
		case "g":
			parseSVGElements(el.Children, tf, ent)
		case "path":
			parseSVGPath(el.attr("d"), tf, ent)
		case "line":
			p1 := tf.apply(Point{el.floatAttr("x1"), el.floatAttr("y1")})
			p2 := tf.apply(Point{el.floatAttr("x2"), el.floatAttr("y2")})
			ent.Paths = append(ent.Paths, Path{Points: []Point{p1, p2}})
		case "polyline", "polygon":
			pts := parsePointsList(el.attr("points"), tf)
			if len(pts) > 0 {
				ent.Paths = append(ent.Paths, Path{
					Points: pts,
					Closed: el.XMLName.Local == "polygon",
				})
			}
		case "rect":
			x, y := el.floatAttr("x"), el.floatAttr("y")
			w, h := el.floatAttr("width"), el.floatAttr("height")
			pts := []Point{
				tf.apply(Point{x, y}),
				tf.apply(Point{x + w, y}),
				tf.apply(Point{x + w, y + h}),
				tf.apply(Point{x, y + h}),
			}
			ent.Paths = append(ent.Paths, Path{Points: pts, Closed: true})
		case "circle":
			cx, cy := el.floatAttr("cx"), el.floatAttr("cy")
			r := el.floatAttr("r")
			pts := circleToPoints(Point{cx, cy}, r, 64)
			for i := range pts {
				pts[i] = tf.apply(pts[i])
			}
			ent.Paths = append(ent.Paths, Path{Points: pts, Closed: true})
		case "ellipse":
			cx, cy := el.floatAttr("cx"), el.floatAttr("cy")
			rx, ry := el.floatAttr("rx"), el.floatAttr("ry")
			pts := make([]Point, 64)
			for i := range pts {
				a := 2 * math.Pi * float64(i) / 64
				pts[i] = tf.apply(Point{cx + rx*math.Cos(a), cy + ry*math.Sin(a)})
			}
			ent.Paths = append(ent.Paths, Path{Points: pts, Closed: true})
		}

		if el.XMLName.Local != "g" {
			parseSVGElements(el.Children, tf, ent)
		}
	}
}

func parsePointsList(s string, xf matrix) []Point {
	nums := parseNumbers(s)
	var pts []Point
	for i := 0; i+1 < len(nums); i += 2 {
		pts = append(pts, xf.apply(Point{nums[i], nums[i+1]}))
	}
	return pts
}

// SVG path parser

func parseSVGPath(d string, xf matrix, ent *Entity) {
	if d == "" {
		return
	}
	tokens := tokenizePath(d)
	var (
		paths   []Path
		cur     Point
		start   Point
		pts     []Point
		cmd     byte
		prevCmd byte
		prevCp  Point // previous control point for S/T
	)

	closePath := func() {
		if len(pts) > 0 {
			paths = append(paths, Path{Points: pts, Closed: true})
			pts = nil
		}
		cur = start
	}

	i := 0
	for i < len(tokens) {
		tok := tokens[i]
		if isLetter(tok) {
			cmd = tok[0]
			i++
		} else if prevCmd != 0 {
			cmd = implicitCmd(prevCmd)
		} else {
			i++
			continue
		}

		rel := cmd >= 'a' && cmd <= 'z'
		upper := cmd
		if rel {
			upper = cmd - 32
		}

		switch upper {
		case 'M':
			if len(pts) > 0 {
				paths = append(paths, Path{Points: pts})
				pts = nil
			}
			x, y := nextNum(tokens, &i), nextNum(tokens, &i)
			if rel {
				x += cur.X
				y += cur.Y
			}
			cur = Point{x, y}
			start = cur
			pts = append(pts, xf.apply(cur))
			if rel {
				cmd = 'l'
			} else {
				cmd = 'L'
			}

		case 'L':
			x, y := nextNum(tokens, &i), nextNum(tokens, &i)
			if rel {
				x += cur.X
				y += cur.Y
			}
			cur = Point{x, y}
			pts = append(pts, xf.apply(cur))

		case 'H':
			x := nextNum(tokens, &i)
			if rel {
				x += cur.X
			}
			cur.X = x
			pts = append(pts, xf.apply(cur))

		case 'V':
			y := nextNum(tokens, &i)
			if rel {
				y += cur.Y
			}
			cur.Y = y
			pts = append(pts, xf.apply(cur))

		case 'C':
			x1, y1 := nextNum(tokens, &i), nextNum(tokens, &i)
			x2, y2 := nextNum(tokens, &i), nextNum(tokens, &i)
			x, y := nextNum(tokens, &i), nextNum(tokens, &i)
			if rel {
				x1 += cur.X; y1 += cur.Y
				x2 += cur.X; y2 += cur.Y
				x += cur.X; y += cur.Y
			}
			bez := flattenCubic(cur, Point{x1, y1}, Point{x2, y2}, Point{x, y}, 0.1)
			for _, p := range bez[1:] {
				pts = append(pts, xf.apply(p))
			}
			prevCp = Point{x2, y2}
			cur = Point{x, y}

		case 'S':
			x2, y2 := nextNum(tokens, &i), nextNum(tokens, &i)
			x, y := nextNum(tokens, &i), nextNum(tokens, &i)
			if rel {
				x2 += cur.X; y2 += cur.Y
				x += cur.X; y += cur.Y
			}
			x1, y1 := cur.X, cur.Y
			if prevCmd == 'C' || prevCmd == 'c' || prevCmd == 'S' || prevCmd == 's' {
				x1 = 2*cur.X - prevCp.X
				y1 = 2*cur.Y - prevCp.Y
			}
			bez := flattenCubic(cur, Point{x1, y1}, Point{x2, y2}, Point{x, y}, 0.1)
			for _, p := range bez[1:] {
				pts = append(pts, xf.apply(p))
			}
			prevCp = Point{x2, y2}
			cur = Point{x, y}

		case 'Q':
			x1, y1 := nextNum(tokens, &i), nextNum(tokens, &i)
			x, y := nextNum(tokens, &i), nextNum(tokens, &i)
			if rel {
				x1 += cur.X; y1 += cur.Y
				x += cur.X; y += cur.Y
			}
			bez := flattenQuadratic(cur, Point{x1, y1}, Point{x, y}, 0.1)
			for _, p := range bez[1:] {
				pts = append(pts, xf.apply(p))
			}
			prevCp = Point{x1, y1}
			cur = Point{x, y}

		case 'T':
			x, y := nextNum(tokens, &i), nextNum(tokens, &i)
			if rel {
				x += cur.X; y += cur.Y
			}
			x1, y1 := cur.X, cur.Y
			if prevCmd == 'Q' || prevCmd == 'q' || prevCmd == 'T' || prevCmd == 't' {
				x1 = 2*cur.X - prevCp.X
				y1 = 2*cur.Y - prevCp.Y
			}
			bez := flattenQuadratic(cur, Point{x1, y1}, Point{x, y}, 0.1)
			for _, p := range bez[1:] {
				pts = append(pts, xf.apply(p))
			}
			prevCp = Point{x1, y1}
			cur = Point{x, y}

		case 'A':
			rx := nextNum(tokens, &i)
			ry := nextNum(tokens, &i)
			rot := nextNum(tokens, &i)
			largeArc := nextNum(tokens, &i) != 0
			sweepFlag := nextNum(tokens, &i) != 0
			x, y := nextNum(tokens, &i), nextNum(tokens, &i)
			if rel {
				x += cur.X; y += cur.Y
			}
			arcPts := svgArcToPoints(cur, Point{x, y}, rx, ry, rot, largeArc, sweepFlag)
			for _, p := range arcPts[1:] {
				pts = append(pts, xf.apply(p))
			}
			cur = Point{x, y}

		case 'Z':
			closePath()

		default:
			i++
		}

		prevCmd = cmd
	}

	if len(pts) > 0 {
		paths = append(paths, Path{Points: pts})
	}
	ent.Paths = append(ent.Paths, paths...)
}

func implicitCmd(prev byte) byte {
	switch prev {
	case 'M':
		return 'L'
	case 'm':
		return 'l'
	default:
		return prev
	}
}

func tokenizePath(d string) []string {
	var tokens []string
	i := 0
	for i < len(d) {
		c := d[i]
		if c == ',' || c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			i++
			continue
		}
		if isPathLetter(c) {
			tokens = append(tokens, string(c))
			i++
			continue
		}
		j := i
		if c == '-' || c == '+' {
			j++
		}
		dotSeen := false
		for j < len(d) {
			ch := d[j]
			if ch == '.' && !dotSeen {
				dotSeen = true
				j++
			} else if ch >= '0' && ch <= '9' {
				j++
			} else if ch == 'e' || ch == 'E' {
				j++
				if j < len(d) && (d[j] == '+' || d[j] == '-') {
					j++
				}
			} else {
				break
			}
		}
		if j > i {
			tokens = append(tokens, d[i:j])
		} else {
			i++
			continue
		}
		i = j
	}
	return tokens
}

func isPathLetter(c byte) bool {
	return (c >= 'A' && c <= 'Z' && c != 'E') || (c >= 'a' && c <= 'z' && c != 'e')
}

func isLetter(s string) bool {
	return len(s) == 1 && isPathLetter(s[0])
}

func nextNum(tokens []string, i *int) float64 {
	for *i < len(tokens) {
		s := tokens[*i]
		if isLetter(s) {
			return 0
		}
		*i++
		v, err := strconv.ParseFloat(s, 64)
		if err == nil {
			return v
		}
	}
	return 0
}

func parseNumbers(s string) []float64 {
	var nums []float64
	for _, field := range strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || unicode.IsSpace(r)
	}) {
		v, err := strconv.ParseFloat(field, 64)
		if err == nil {
			nums = append(nums, v)
		}
	}
	return nums
}

// De Casteljau flattening

func flattenCubic(p0, p1, p2, p3 Point, tol float64) []Point {
	if cubicFlat(p0, p1, p2, p3, tol) {
		return []Point{p0, p3}
	}
	m01 := mid(p0, p1)
	m12 := mid(p1, p2)
	m23 := mid(p2, p3)
	m012 := mid(m01, m12)
	m123 := mid(m12, m23)
	m0123 := mid(m012, m123)
	left := flattenCubic(p0, m01, m012, m0123, tol)
	right := flattenCubic(m0123, m123, m23, p3, tol)
	return append(left, right[1:]...)
}

func cubicFlat(p0, p1, p2, p3 Point, tol float64) bool {
	dx := p3.X - p0.X
	dy := p3.Y - p0.Y
	d := math.Sqrt(dx*dx + dy*dy)
	if d < 1e-12 {
		return true
	}
	d1 := math.Abs((p1.X-p0.X)*dy-(p1.Y-p0.Y)*dx) / d
	d2 := math.Abs((p2.X-p0.X)*dy-(p2.Y-p0.Y)*dx) / d
	return d1 <= tol && d2 <= tol
}

func flattenQuadratic(p0, p1, p2 Point, tol float64) []Point {
	if quadFlat(p0, p1, p2, tol) {
		return []Point{p0, p2}
	}
	m01 := mid(p0, p1)
	m12 := mid(p1, p2)
	m012 := mid(m01, m12)
	left := flattenQuadratic(p0, m01, m012, tol)
	right := flattenQuadratic(m012, m12, p2, tol)
	return append(left, right[1:]...)
}

func quadFlat(p0, p1, p2 Point, tol float64) bool {
	dx := p2.X - p0.X
	dy := p2.Y - p0.Y
	d := math.Sqrt(dx*dx + dy*dy)
	if d < 1e-12 {
		return true
	}
	return math.Abs((p1.X-p0.X)*dy-(p1.Y-p0.Y)*dx)/d <= tol
}

func mid(a, b Point) Point {
	return Point{(a.X + b.X) / 2, (a.Y + b.Y) / 2}
}

// SVG arc (endpoint parameterization) to polyline points.
// Implements the SVG spec F.6 conversion.
func svgArcToPoints(from, to Point, rx, ry, rotDeg float64, largeArc, sweep bool) []Point {
	if rx == 0 || ry == 0 {
		return []Point{from, to}
	}
	rx = math.Abs(rx)
	ry = math.Abs(ry)

	phi := rotDeg * math.Pi / 180
	cosPhi := math.Cos(phi)
	sinPhi := math.Sin(phi)

	dx := (from.X - to.X) / 2
	dy := (from.Y - to.Y) / 2
	x1p := cosPhi*dx + sinPhi*dy
	y1p := -sinPhi*dx + cosPhi*dy

	// Scale radii if needed
	lambda := (x1p*x1p)/(rx*rx) + (y1p*y1p)/(ry*ry)
	if lambda > 1 {
		s := math.Sqrt(lambda)
		rx *= s
		ry *= s
	}

	num := rx*rx*ry*ry - rx*rx*y1p*y1p - ry*ry*x1p*x1p
	den := rx*rx*y1p*y1p + ry*ry*x1p*x1p
	sq := 0.0
	if den > 0 && num > 0 {
		sq = math.Sqrt(num / den)
	}
	if largeArc == sweep {
		sq = -sq
	}
	cxp := sq * rx * y1p / ry
	cyp := -sq * ry * x1p / rx

	cx := cosPhi*cxp - sinPhi*cyp + (from.X+to.X)/2
	cy := sinPhi*cxp + cosPhi*cyp + (from.Y+to.Y)/2

	startAngle := vecAngle(1, 0, (x1p-cxp)/rx, (y1p-cyp)/ry)
	dTheta := vecAngle((x1p-cxp)/rx, (y1p-cyp)/ry, (-x1p-cxp)/rx, (-y1p-cyp)/ry)

	if !sweep && dTheta > 0 {
		dTheta -= 2 * math.Pi
	} else if sweep && dTheta < 0 {
		dTheta += 2 * math.Pi
	}

	segments := int(math.Ceil(math.Abs(dTheta) / (math.Pi / 16)))
	if segments < 2 {
		segments = 2
	}

	pts := make([]Point, segments+1)
	for k := 0; k <= segments; k++ {
		t := float64(k) / float64(segments)
		a := startAngle + t*dTheta
		x := rx*math.Cos(a)
		y := ry*math.Sin(a)
		pts[k] = Point{
			X: cosPhi*x - sinPhi*y + cx,
			Y: sinPhi*x + cosPhi*y + cy,
		}
	}
	return pts
}

func vecAngle(ux, uy, vx, vy float64) float64 {
	dot := ux*vx + uy*vy
	cross := ux*vy - uy*vx
	return math.Atan2(cross, dot)
}
