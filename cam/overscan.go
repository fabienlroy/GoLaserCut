package cam

// RampProfile describes a linear-acceleration motion profile for
// overscan in laser raster engraving. Acceleration is the second
// derivative of position and varies linearly (triangular profile):
// it rises from 0 to PeakAccel at the midpoint, then falls back to 0.
// This S-curve velocity eliminates jerk discontinuities that cause
// mechanical vibrations at scan-line endpoints.
type RampProfile struct {
	TargetSpeed  float64 // mm/s
	PeakAccel    float64 // mm/s² (peak of triangular acceleration)
	Jerk         float64 // mm/s³ (constant da/dt within each half)
	RampTime     float64 // s (time from rest to TargetSpeed)
	RampDistance  float64 // mm (overscan distance per side)
}

// NewRampProfile computes a ramp profile from cutting parameters.
// feedRate is in mm/min, maxAccel is peak acceleration in mm/s².
func NewRampProfile(feedRate, maxAccel float64) RampProfile {
	if feedRate <= 0 || maxAccel <= 0 {
		return RampProfile{}
	}
	v := feedRate / 60.0
	t := 2.0 * v / maxAccel
	return RampProfile{
		TargetSpeed: v,
		PeakAccel:   maxAccel,
		Jerk:        maxAccel * maxAccel / v,
		RampTime:    t,
		RampDistance: v * v / maxAccel,
	}
}

// AccelAt returns acceleration at time t during the acceleration ramp.
func (p RampProfile) AccelAt(t float64) float64 {
	if t <= 0 || t >= p.RampTime {
		return 0
	}
	half := p.RampTime / 2
	if t <= half {
		return p.Jerk * t
	}
	return p.PeakAccel - p.Jerk*(t-half)
}

// VelocityAt returns velocity at time t during the acceleration ramp.
func (p RampProfile) VelocityAt(t float64) float64 {
	if t <= 0 {
		return 0
	}
	if t >= p.RampTime {
		return p.TargetSpeed
	}
	half := p.RampTime / 2
	if t <= half {
		return p.Jerk * t * t / 2
	}
	tau := t - half
	return p.TargetSpeed/2 + p.PeakAccel*tau - p.Jerk*tau*tau/2
}

// PositionAt returns position at time t during the acceleration ramp.
func (p RampProfile) PositionAt(t float64) float64 {
	if t <= 0 {
		return 0
	}
	if t >= p.RampTime {
		return p.RampDistance
	}
	half := p.RampTime / 2
	if t <= half {
		return p.Jerk * t * t * t / 6
	}
	tau := t - half
	xHalf := p.Jerk * half * half * half / 6
	return xHalf + p.TargetSpeed/2*tau + p.PeakAccel*tau*tau/2 - p.Jerk*tau*tau*tau/6
}

// DecelVelocityAt returns velocity at time t during the deceleration ramp.
// At t=0 velocity is TargetSpeed, at t=RampTime velocity is 0.
func (p RampProfile) DecelVelocityAt(t float64) float64 {
	return p.VelocityAt(p.RampTime - t)
}

// DecelPositionAt returns distance traveled at time t during deceleration.
// At t=0 position is 0, at t=RampTime position is RampDistance.
func (p RampProfile) DecelPositionAt(t float64) float64 {
	return p.RampDistance - p.PositionAt(p.RampTime-t)
}

// Segment is one discrete step of a ramp, suitable for G-code generation.
type Segment struct {
	Length   float64 // mm
	FeedRate float64 // mm/min
}

// AccelSegments discretizes the acceleration ramp into n segments.
func (p RampProfile) AccelSegments(n int) []Segment {
	if n <= 0 || p.RampTime <= 0 {
		return nil
	}
	dt := p.RampTime / float64(n)
	segs := make([]Segment, n)
	prev := 0.0
	for i := range segs {
		pos := p.PositionAt(float64(i+1) * dt)
		vel := p.VelocityAt((float64(i) + 0.5) * dt)
		segs[i] = Segment{
			Length:   pos - prev,
			FeedRate: vel * 60,
		}
		prev = pos
	}
	return segs
}

// DecelSegments discretizes the deceleration ramp into n segments.
func (p RampProfile) DecelSegments(n int) []Segment {
	if n <= 0 || p.RampTime <= 0 {
		return nil
	}
	dt := p.RampTime / float64(n)
	segs := make([]Segment, n)
	prev := 0.0
	for i := range segs {
		pos := p.DecelPositionAt(float64(i+1) * dt)
		vel := p.DecelVelocityAt((float64(i) + 0.5) * dt)
		segs[i] = Segment{
			Length:   pos - prev,
			FeedRate: vel * 60,
		}
		prev = pos
	}
	return segs
}
