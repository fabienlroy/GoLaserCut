package cam

import "math"

// VBit defines a conical milling tool for V-carving.
type VBit struct {
	Angle    float64 // full included angle, degrees (e.g. 60, 90)
	TipWidth float64 // flat tip diameter, mm (0 = sharp point)
}

func (v VBit) halfAngleRad() float64 {
	return v.Angle * math.Pi / 360
}

// DepthForWidth returns the plunge depth (positive value) needed
// to achieve cut width w at the material surface.
func (v VBit) DepthForWidth(w float64) float64 {
	if w <= v.TipWidth {
		return 0
	}
	return (w - v.TipWidth) / (2 * math.Tan(v.halfAngleRad()))
}

// WidthAtDepth returns the cut width at plunge depth d (positive value).
func (v VBit) WidthAtDepth(d float64) float64 {
	if d <= 0 {
		return v.TipWidth
	}
	return v.TipWidth + 2*d*math.Tan(v.halfAngleRad())
}

// MaxWidth returns the maximum cut width at the given depth limit.
func (v VBit) MaxWidth(maxDepth float64) float64 {
	return v.WidthAtDepth(maxDepth)
}
