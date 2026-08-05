package cam

import (
	"math"
	"testing"
)

func TestVBitDepthForWidth(t *testing.T) {
	// 90° V-bit, sharp tip: tan(45°)=1, so depth = width/2
	v := VBit{Angle: 90, TipWidth: 0}
	if !approx(v.DepthForWidth(2.0), 1.0) {
		t.Errorf("90° sharp: DepthForWidth(2) = %f, want 1.0", v.DepthForWidth(2.0))
	}
	if !approx(v.DepthForWidth(10.0), 5.0) {
		t.Errorf("90° sharp: DepthForWidth(10) = %f, want 5.0", v.DepthForWidth(10.0))
	}
}

func TestVBitDepthForWidthWithTip(t *testing.T) {
	// 90° V-bit, 0.5mm tip: depth = (width - 0.5) / 2
	v := VBit{Angle: 90, TipWidth: 0.5}
	if !approx(v.DepthForWidth(2.5), 1.0) {
		t.Errorf("DepthForWidth(2.5) = %f, want 1.0", v.DepthForWidth(2.5))
	}
	if !approx(v.DepthForWidth(0.5), 0) {
		t.Errorf("DepthForWidth(0.5) = %f, want 0 (at tip width)", v.DepthForWidth(0.5))
	}
	if !approx(v.DepthForWidth(0.1), 0) {
		t.Errorf("DepthForWidth(0.1) = %f, want 0 (below tip width)", v.DepthForWidth(0.1))
	}
}

func TestVBitWidthAtDepth(t *testing.T) {
	v := VBit{Angle: 90, TipWidth: 0}
	if !approx(v.WidthAtDepth(1.0), 2.0) {
		t.Errorf("WidthAtDepth(1) = %f, want 2.0", v.WidthAtDepth(1.0))
	}
	if !approx(v.WidthAtDepth(0), 0) {
		t.Errorf("WidthAtDepth(0) = %f, want 0", v.WidthAtDepth(0))
	}
	if !approx(v.WidthAtDepth(-1), 0) {
		t.Errorf("WidthAtDepth(-1) = %f, want 0 (negative depth)", v.WidthAtDepth(-1))
	}
}

func TestVBitWidthAtDepthWithTip(t *testing.T) {
	v := VBit{Angle: 90, TipWidth: 0.5}
	if !approx(v.WidthAtDepth(0), 0.5) {
		t.Errorf("WidthAtDepth(0) = %f, want 0.5 (tip width)", v.WidthAtDepth(0))
	}
	if !approx(v.WidthAtDepth(1.0), 2.5) {
		t.Errorf("WidthAtDepth(1) = %f, want 2.5", v.WidthAtDepth(1.0))
	}
}

func TestVBit60Degree(t *testing.T) {
	// 60° V-bit: half-angle = 30°, tan(30°) = 1/√3
	v := VBit{Angle: 60, TipWidth: 0}
	tanHalf := math.Tan(30 * math.Pi / 180)
	expectedDepth := 2.0 / (2 * tanHalf) // = 1/tan(30°) = √3
	if !approx(v.DepthForWidth(2.0), expectedDepth) {
		t.Errorf("60° DepthForWidth(2) = %f, want %f", v.DepthForWidth(2.0), expectedDepth)
	}
}

func TestVBitRoundTrip(t *testing.T) {
	// depth → width → depth should be identity
	for _, angle := range []float64{30, 45, 60, 90, 120} {
		v := VBit{Angle: angle, TipWidth: 0.2}
		for _, w := range []float64{0.5, 1.0, 2.0, 5.0} {
			d := v.DepthForWidth(w)
			w2 := v.WidthAtDepth(d)
			if !approx(w2, w) {
				t.Errorf("angle=%.0f: width %.2f → depth %.4f → width %.4f (want %.2f)",
					angle, w, d, w2, w)
			}
		}
	}
}

func TestVBitMaxWidth(t *testing.T) {
	v := VBit{Angle: 90, TipWidth: 0}
	if !approx(v.MaxWidth(5.0), 10.0) {
		t.Errorf("MaxWidth(5) = %f, want 10", v.MaxWidth(5.0))
	}
}
