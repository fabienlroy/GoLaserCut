package cam

import (
	"math"
	"testing"
)

const tol = 1e-9

func approx(a, b float64) bool {
	return math.Abs(a-b) < tol
}

func TestNewRampProfile(t *testing.T) {
	p := NewRampProfile(6000, 500) // 100 mm/s, 500 mm/s²
	if !approx(p.TargetSpeed, 100) {
		t.Errorf("TargetSpeed = %f, want 100", p.TargetSpeed)
	}
	if !approx(p.PeakAccel, 500) {
		t.Errorf("PeakAccel = %f, want 500", p.PeakAccel)
	}
	if !approx(p.RampTime, 0.4) {
		t.Errorf("RampTime = %f, want 0.4", p.RampTime)
	}
	if !approx(p.Jerk, 2500) {
		t.Errorf("Jerk = %f, want 2500", p.Jerk)
	}
	if !approx(p.RampDistance, 20) {
		t.Errorf("RampDistance = %f, want 20", p.RampDistance)
	}
}

func TestNewRampProfileZeroInputs(t *testing.T) {
	p := NewRampProfile(0, 500)
	if p.RampDistance != 0 {
		t.Error("expected zero profile for zero feed rate")
	}
	p = NewRampProfile(6000, 0)
	if p.RampDistance != 0 {
		t.Error("expected zero profile for zero accel")
	}
	p = NewRampProfile(-100, 500)
	if p.RampDistance != 0 {
		t.Error("expected zero profile for negative feed rate")
	}
}

func TestAccelAt(t *testing.T) {
	p := NewRampProfile(6000, 500)

	if p.AccelAt(0) != 0 {
		t.Error("AccelAt(0) should be 0")
	}
	if p.AccelAt(p.RampTime) != 0 {
		t.Error("AccelAt(T) should be 0")
	}
	if p.AccelAt(-1) != 0 {
		t.Error("AccelAt(-1) should be 0")
	}
	if !approx(p.AccelAt(0.2), 500) {
		t.Errorf("AccelAt(T/2) = %f, want 500", p.AccelAt(0.2))
	}
	if !approx(p.AccelAt(0.1), 250) {
		t.Errorf("AccelAt(T/4) = %f, want 250", p.AccelAt(0.1))
	}
	if !approx(p.AccelAt(0.3), 250) {
		t.Errorf("AccelAt(3T/4) = %f, want 250", p.AccelAt(0.3))
	}
}

func TestAccelSymmetry(t *testing.T) {
	p := NewRampProfile(6000, 500)
	for i := 1; i < 100; i++ {
		frac := float64(i) / 100.0
		t1 := frac * p.RampTime
		t2 := (1 - frac) * p.RampTime
		if !approx(p.AccelAt(t1), p.AccelAt(t2)) {
			t.Errorf("accel not symmetric: a(%f)=%f != a(%f)=%f",
				t1, p.AccelAt(t1), t2, p.AccelAt(t2))
		}
	}
}

func TestVelocityAt(t *testing.T) {
	p := NewRampProfile(6000, 500)

	if p.VelocityAt(0) != 0 {
		t.Error("VelocityAt(0) should be 0")
	}
	if !approx(p.VelocityAt(p.RampTime), 100) {
		t.Errorf("VelocityAt(T) = %f, want 100", p.VelocityAt(p.RampTime))
	}
	if !approx(p.VelocityAt(0.2), 50) {
		t.Errorf("VelocityAt(T/2) = %f, want 50", p.VelocityAt(0.2))
	}
}

func TestVelocityMonotonic(t *testing.T) {
	p := NewRampProfile(6000, 500)
	prev := 0.0
	for i := 1; i <= 1000; i++ {
		tt := float64(i) * p.RampTime / 1000
		v := p.VelocityAt(tt)
		if v < prev-tol {
			t.Errorf("velocity not monotonic at t=%f: %f < %f", tt, v, prev)
		}
		prev = v
	}
}

func TestPositionAt(t *testing.T) {
	p := NewRampProfile(6000, 500)

	if p.PositionAt(0) != 0 {
		t.Error("PositionAt(0) should be 0")
	}
	if !approx(p.PositionAt(p.RampTime), 20) {
		t.Errorf("PositionAt(T) = %f, want 20", p.PositionAt(p.RampTime))
	}
	expected := p.Jerk * 0.008 / 6.0 // j*(T/2)³/6
	if !approx(p.PositionAt(0.2), expected) {
		t.Errorf("PositionAt(T/2) = %f, want %f", p.PositionAt(0.2), expected)
	}
}

func TestPositionMonotonic(t *testing.T) {
	p := NewRampProfile(6000, 500)
	prev := 0.0
	for i := 1; i <= 1000; i++ {
		tt := float64(i) * p.RampTime / 1000
		x := p.PositionAt(tt)
		if x < prev-tol {
			t.Errorf("position not monotonic at t=%f: %f < %f", tt, x, prev)
		}
		prev = x
	}
}

func TestDecelVelocity(t *testing.T) {
	p := NewRampProfile(6000, 500)

	if !approx(p.DecelVelocityAt(0), 100) {
		t.Errorf("DecelVelocityAt(0) = %f, want 100", p.DecelVelocityAt(0))
	}
	if !approx(p.DecelVelocityAt(p.RampTime), 0) {
		t.Errorf("DecelVelocityAt(T) = %f, want 0", p.DecelVelocityAt(p.RampTime))
	}
	if !approx(p.DecelVelocityAt(0.2), 50) {
		t.Errorf("DecelVelocityAt(T/2) = %f, want 50", p.DecelVelocityAt(0.2))
	}
}

func TestDecelPosition(t *testing.T) {
	p := NewRampProfile(6000, 500)

	if !approx(p.DecelPositionAt(0), 0) {
		t.Errorf("DecelPositionAt(0) = %f, want 0", p.DecelPositionAt(0))
	}
	if !approx(p.DecelPositionAt(p.RampTime), 20) {
		t.Errorf("DecelPositionAt(T) = %f, want 20", p.DecelPositionAt(p.RampTime))
	}
}

func TestAccelSegments(t *testing.T) {
	p := NewRampProfile(6000, 500)
	segs := p.AccelSegments(4)
	if len(segs) != 4 {
		t.Fatalf("got %d segments, want 4", len(segs))
	}

	total := 0.0
	for _, s := range segs {
		total += s.Length
	}
	if !approx(total, 20) {
		t.Errorf("total length = %f, want 20", total)
	}

	for i := 1; i < len(segs); i++ {
		if segs[i].FeedRate <= segs[i-1].FeedRate {
			t.Errorf("feed rate not increasing: seg[%d]=%f, seg[%d]=%f",
				i-1, segs[i-1].FeedRate, i, segs[i].FeedRate)
		}
	}

	for i, s := range segs {
		if s.FeedRate <= 0 || s.Length <= 0 {
			t.Errorf("segment %d: length=%f feedRate=%f, both must be > 0", i, s.Length, s.FeedRate)
		}
	}
}

func TestDecelSegments(t *testing.T) {
	p := NewRampProfile(6000, 500)
	segs := p.DecelSegments(4)
	if len(segs) != 4 {
		t.Fatalf("got %d segments, want 4", len(segs))
	}

	total := 0.0
	for _, s := range segs {
		total += s.Length
	}
	if !approx(total, 20) {
		t.Errorf("total length = %f, want 20", total)
	}

	for i := 1; i < len(segs); i++ {
		if segs[i].FeedRate >= segs[i-1].FeedRate {
			t.Errorf("feed rate not decreasing: seg[%d]=%f, seg[%d]=%f",
				i-1, segs[i-1].FeedRate, i, segs[i].FeedRate)
		}
	}
}

func TestSegmentsNilCases(t *testing.T) {
	p := NewRampProfile(6000, 500)
	if p.AccelSegments(0) != nil {
		t.Error("expected nil for n=0")
	}
	if p.DecelSegments(-1) != nil {
		t.Error("expected nil for n=-1")
	}
	z := NewRampProfile(0, 0)
	if z.AccelSegments(4) != nil {
		t.Error("expected nil for zero profile")
	}
}

func TestAccelContinuity(t *testing.T) {
	p := NewRampProfile(6000, 500)
	steps := 10000
	dt := p.RampTime / float64(steps)
	maxJump := 0.0
	for i := 1; i <= steps; i++ {
		a1 := p.AccelAt(float64(i-1) * dt)
		a2 := p.AccelAt(float64(i) * dt)
		jump := math.Abs(a2 - a1)
		if jump > maxJump {
			maxJump = jump
		}
	}
	limit := p.Jerk * dt * 1.01
	if maxJump > limit {
		t.Errorf("acceleration discontinuity: max jump = %f, limit = %f", maxJump, limit)
	}
}

func TestSCurveShape(t *testing.T) {
	p := NewRampProfile(6000, 500)

	// Phase 1: velocity curve is convex (below linear interpolation)
	vQuarter := p.VelocityAt(0.1)
	vLinear := p.VelocityAt(0) + (p.VelocityAt(0.2)-p.VelocityAt(0))*0.5
	if vQuarter >= vLinear {
		t.Errorf("phase 1 not convex: v(T/4)=%f >= linear=%f", vQuarter, vLinear)
	}

	// Phase 2: velocity curve is concave (above linear interpolation)
	v3Quarter := p.VelocityAt(0.3)
	vLinear2 := p.VelocityAt(0.2) + (p.VelocityAt(0.4)-p.VelocityAt(0.2))*0.5
	if v3Quarter <= vLinear2 {
		t.Errorf("phase 2 not concave: v(3T/4)=%f <= linear=%f", v3Quarter, vLinear2)
	}
}

func TestVsConstantAccel(t *testing.T) {
	v := 100.0    // mm/s
	a := 500.0    // mm/s²
	feedRate := v * 60

	constantDist := v * v / (2 * a)
	p := NewRampProfile(feedRate, a)

	ratio := p.RampDistance / constantDist
	if !approx(ratio, 2.0) {
		t.Errorf("linear/constant distance ratio = %f, want 2.0", ratio)
	}
}

func TestDifferentSpeeds(t *testing.T) {
	tests := []struct {
		feedRate float64
		accel    float64
		wantDist float64
		wantTime float64
	}{
		{3000, 500, 5.0, 0.2},    // 50 mm/s: d=50²/500=5
		{12000, 1000, 40.0, 0.4}, // 200 mm/s: d=200²/1000=40
		{600, 200, 0.5, 0.1},     // 10 mm/s: d=10²/200=0.5
	}
	for _, tt := range tests {
		p := NewRampProfile(tt.feedRate, tt.accel)
		if !approx(p.RampDistance, tt.wantDist) {
			t.Errorf("feedRate=%f accel=%f: dist=%f, want %f",
				tt.feedRate, tt.accel, p.RampDistance, tt.wantDist)
		}
		if !approx(p.RampTime, tt.wantTime) {
			t.Errorf("feedRate=%f accel=%f: time=%f, want %f",
				tt.feedRate, tt.accel, p.RampTime, tt.wantTime)
		}
	}
}

func TestNumericalIntegration(t *testing.T) {
	p := NewRampProfile(6000, 500)
	n := 100000
	dt := p.RampTime / float64(n)

	// Integrate velocity to get position — should match PositionAt(T)
	pos := 0.0
	for i := 0; i < n; i++ {
		v := p.VelocityAt((float64(i) + 0.5) * dt)
		pos += v * dt
	}
	if math.Abs(pos-p.RampDistance) > 1e-6 {
		t.Errorf("integrated position = %f, want %f", pos, p.RampDistance)
	}

	// Integrate acceleration to get velocity — should match TargetSpeed
	vel := 0.0
	for i := 0; i < n; i++ {
		a := p.AccelAt((float64(i) + 0.5) * dt)
		vel += a * dt
	}
	if math.Abs(vel-p.TargetSpeed) > 1e-4 {
		t.Errorf("integrated velocity = %f, want %f", vel, p.TargetSpeed)
	}
}
