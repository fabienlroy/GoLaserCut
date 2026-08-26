# GoLaserCut

Free, open-source, cross-platform laser cutter software written in Go.

**Import** (SVG, DXF, images) → **CAM** (cut, engrave, 3D relief) → **Control** (GRBL 1.1 over serial)

Works on **macOS**, **Linux** and **Windows**. No runtime dependencies.

## Why?

| Software | Free | macOS | Linux | Windows | CAM + Control |
|----------|------|-------|-------|---------|---------------|
| LaserGRBL | ✅ | ❌ | ❌ | ✅ | ✅ |
| LightBurn | ❌ (60€) | ✅ | ✅ | ✅ | ✅ |
| **GoLaserCut** | **✅** | **✅** | **✅** | **✅** | **✅** |

## Status

Early development — contributions welcome.

See [project_plan.md](project_plan.md) for architecture, roadmap and conventions.

## Build

```bash
go build ./cmd/golasercut
```

Requires Go 1.22+.

## Software

GoLaserCut is a complete laser cutter toolchain:

### Import

- **SVG** — vector paths for cutting and engraving
- **DXF** — CAD interchange format
- **Images** — raster engraving (PNG, JPG, BMP)
- **STL heightmaps** — 3D relief engraving from surface meshes

### CAM (Computer-Aided Manufacturing)

- **Line mode** — vector cutting with configurable multi-pass, power ramp, and lead-in/out
- **Scan mode** — raster engraving with bidirectional scanning, overscan, and dithering
- **3D engraving** — heightmap-to-G-code conversion for variable-depth pocket machining
- **Depth calibration** — generates a grid of test pockets at graduated power levels; measure the pocket depths and the software interpolates power-to-depth curves per material (see [Pocket Machining](#pocket-machining-with-calibrated-depth-control))
- **Material library** — built-in profiles (plywood, acrylic, stainless steel, aluminum, etc.) with user-calibrated depth data stored as JSON

### Control

- **GRBL 1.1** sender with real-time status, jog, feed/speed overrides
- **Serial** communication (USB) with auto-detection
- **GRBL simulator** for offline testing and integration tests
- Native GUI via [Gio](https://gioui.org) — no Electron, no webview

### Material Library and Calibration System

The `material/` package provides:

- **`MaterialLibrary`** — stores material properties (name, category, thickness range, default feed rate) and calibration results as JSON
- **`CalibrationResult`** — power-vs-depth lookup with linear interpolation: `DepthForPower(pct)` and `PowerForDepth(depth)` functions
- **`GenerateCalibrationGCode()`** — outputs a test pattern of N pockets at graduated power levels for any material
- **Built-in materials** — plywood, MDF, acrylic, leather, cardboard, anodized aluminum, stainless steel, and more

## 30 MPa Water-Jet Guided Laser

The 30 MPa water-jet guided laser (WJGL) system is an open-source DIY alternative to industrial systems like Synova LMJ. A 1 kW CW fiber laser beam is guided inside a laminar micro water jet at 30 MPa (300 bar), enabling precision cutting with a ~100 um kerf on metals, ceramics, semiconductors, and composites.

The original Synova patents on water-jet guided laser technology (EP 0762947, US 5,773,791 — B. Richerzhagen, filed 1995) have expired (20-year term), placing the core method in the public domain.

### System Overview

```
RO Water Tank ─── HPLC Pump (30 MPa) ─── HP Chamber ─── Sapphire Nozzle
                                                              │
650nm Red Guide ── WDM Combiner ── 1kW Fiber Laser ──── Sapphire Window
                                                              │
                          Arduino Nano ─── QCW TTL ─── Laser Driver
                               │
                     Safety Interlocks (pressure, flow, leak, door, e-stop)
```

### Key Specifications

| Parameter | Value |
|-----------|-------|
| Laser power | 1 kW CW / ~10 kW peak QCW |
| Wavelength | 1070 nm (IR) + 650 nm red guide |
| Water pressure | 30 MPa (300 bar) |
| Nozzle orifice | ~100 um (sapphire, zirconia, or alumina) |
| Kerf width | ~100 um |
| QCW mode | 1-50 kHz, 5-50% duty cycle |
| Total BOM cost | ~2200-4025 EUR |

### Why 1 kW Is Enough

The water-jet guided laser requires the beam to focus into a ~100 um water jet orifice. This demands near-single-mode beam quality (M² ~ 1.1). Power combining (2x 1 kW via fiber combiner) degrades M² proportionally to the number of inputs — the larger output fiber becomes multimode and can no longer couple efficiently into the micro-jet. The nozzle acts as a spatial filter that rejects degraded beams.

Instead of more average power, peak power matters for cutting:

- **QCW pulsing** delivers ~10 kW peak from 1 kW average (10% duty cycle) — this is what ablates material
- **Industrial Synova** systems use only 50-500 W for the same reason: beam quality into the jet trumps raw wattage
- **Multi-pass** at 1 kW already cuts deeper than Synova at 200 W (see tables below)

The 1 kW CW + QCW configuration already exceeds industrial systems on cutting speed and depth while maintaining the beam quality needed for water-jet guidance.

### Cutting Performance

Performance estimates for the DIY 1 kW system vs industrial Synova LMJ (200 W pulsed):

#### Cutting Speed

| Material | Thickness | DIY 1 kW CW | Synova 200 W |
|----------|-----------|-------------|--------------|
| Stainless steel | 0.5 mm | 400-800 mm/min | 100-200 mm/min |
| Stainless steel | 1.0 mm | 150-400 mm/min | 30-80 mm/min |
| Stainless steel | 3.0 mm | 30-80 mm/min | 5-15 mm/min |
| Aluminum | 1.0 mm | 500-1200 mm/min | 150-300 mm/min |
| Titanium | 1.0 mm | 200-500 mm/min | 50-120 mm/min |
| Silicon wafer | 0.3 mm | 300-600 mm/min | 200-500 mm/min |
| Sapphire | 0.5 mm | 40-100 mm/min | 20-60 mm/min |
| CFRP composite | 2.0 mm | 200-500 mm/min | 50-150 mm/min |
| Glass (boro) | 1.0 mm | 80-200 mm/min | 40-100 mm/min |

#### Maximum Cutting Depth (multi-pass)

| Material | DIY 1 kW | Synova 200 W | Passes |
|----------|----------|--------------|--------|
| Stainless steel | 6-15 mm | 3-6 mm | 5-20 |
| Aluminum | 10-20 mm | 5-10 mm | 5-15 |
| Titanium | 5-12 mm | 2-5 mm | 10-30 |
| Silicon | 7-15 mm | up to 7 mm | 10-30 |
| CFRP | 10-20 mm | 5-10 mm | 5-20 |
| Superalloy (Inconel) | 10-20 mm | up to 15 mm | multi-pass, aspect 100:1 |

#### Cut Quality: CW vs QCW vs Industrial Pulsed

| Quality Factor | DIY CW | DIY QCW | Industrial ns Pulsed |
|----------------|--------|---------|---------------------|
| Heat-affected zone | 20-80 um | 5-25 um | 2-15 um |
| Surface roughness (Ra) | 3-10 um | 1-4 um | 0.5-1 um |
| Recast layer | 5-30 um | 1-10 um | 0-5 um |
| Taper angle | < 1 deg | < 0.5 deg | < 0.5 deg |
| Burr / dross | Minimal | Very low | None |

### Industrial Comparison

| | DIY 1 kW WJGL | Synova LMJ-500 | Chinese WJGL |
|--|---------------|----------------|--------------|
| Cost | ~3000 EUR | >500 000 EUR | ~50 000 EUR |
| Power | 1 kW CW + QCW | 50-500 W pulsed (ns) | 200-500 W |
| Cutting quality | Good (QCW: near-industrial) | Best in class | Good |
| Automation | Manual / FoxAlien CNC | Full CNC, 5-axis | CNC, 3-axis |
| Best for | Prototyping, R&D, makers | Production, semiconductors | Small production |

### Pocket Machining with Calibrated Depth Control

By combining the water-jet guided laser with GoLaserCut's calibration system, the kit enables **precision pocket machining** — controlled material removal to a target depth rather than cutting through.

**How it works:**

1. **Calibrate** — Generate a test pattern of border pockets using `GenerateCalibrationGCode()`:
   - 10 pockets at graduated power levels (10% to 90%)
   - Run the pattern on your target material
   - Measure each pocket depth with a micrometer or depth gauge

2. **Store** — Feed the measurements back into the material library:
   - The software builds a power-to-depth interpolation curve
   - `PowerForDepth(0.3)` returns the exact S-value to remove 0.3 mm

3. **Machine** — Use the calibrated profile for pocket operations:
   - Software converts a desired pocket depth to the correct power/speed
   - Multi-pass with Z-axis stepping for deeper pockets
   - Achievable precision: ~10-20 um depth accuracy with proper calibration

**Applications:**

- Controlled-depth engraving on metals (serial numbers, logos, scales)
- Thin-wall pocketing for weight reduction (aerospace parts)
- Channel milling in glass/ceramic for microfluidics
- PCB depaneling with controlled depth scoring
- Selective coating removal (anodizing, oxide layers)

**Calibration border pockets for each material:**

| Material | Pocket size | Feed rate | Power range | Depth per pass |
|----------|-------------|-----------|-------------|----------------|
| Stainless steel 0.5mm | 10x10 mm | 200 mm/min | 20-90% | 0.05-0.3 mm |
| Aluminum 1.0mm | 10x10 mm | 500 mm/min | 10-80% | 0.1-0.5 mm |
| Titanium 0.5mm | 10x10 mm | 150 mm/min | 20-90% | 0.03-0.2 mm |
| Silicon wafer 0.3mm | 5x5 mm | 300 mm/min | 10-70% | 0.02-0.15 mm |
| CFRP 2.0mm | 10x10 mm | 300 mm/min | 20-90% | 0.1-0.4 mm |
| Glass (boro) 1.0mm | 10x10 mm | 200 mm/min | 15-80% | 0.05-0.25 mm |
| Alumina ceramic 1.0mm | 10x10 mm | 100 mm/min | 30-90% | 0.03-0.15 mm |

### QCW Controller + Safety Interlocks

A single Arduino Nano handles QCW pulse generation and safety monitoring:

- Reads FoxAlien GRBL spindle PWM, outputs QCW pulses at 1-50 kHz
- S-value controls duty cycle, frequency set via serial
- Safety interlocks: HP pressure (QDX50A, 200-350 bar), LP flow (> 1 L/min), water leak, IR flame/thermal detection, door, e-stop
- IR flame detectors (digital) detect material ignition, IR thermal sensor monitors workpiece temperature (kills at > 150 C)
- Optional USB micro camera for real-time cutting observation, triggered via serial command
- Any fault = instant laser kill, serial fault reporting, manual reset required

Source: [`30MPaLaserGuide/qcw_pwm_converter/`](30MPaLaserGuide/qcw_pwm_converter/)

### Interactive Documentation

- [3D Viewer](https://claude.ai/code/artifact/5d1fa2bf-d95c-4771-8e8f-19089802bc26) — Interactive WebGL viewer (explode, clip, toggle pieces)
- [DIY vs Industrial Comparison](https://claude.ai/code/artifact/5689606b-0189-43bd-8e68-c377c5c5598c) — Cost and specs vs Synova LMJ, Chinese WJGL
- [Cutting Performance](https://claude.ai/code/artifact/58c09f3d-e962-4e3f-971e-371441b1937c) — Speed, depth (single & multi-pass), kerf, cut quality

### Source Files

[`30MPaLaserGuide/`](30MPaLaserGuide/):

| File | Description |
|------|-------------|
| `head_laser_waterjet.scad` | Parametric CAD model (3 machined pieces, red laser guide, RO unit, water tank) |
| `export_A/B/C.scad` | Individual STL export scripts for each piece |
| `BillOfMaterial.txt` | Full BOM with part references, suppliers, pricing (~2200-4025 EUR) |
| `qcw_pwm_converter/` | Arduino sketch: QCW pulse generation + safety interlocks |
| `viewer_3d.html` | Standalone WebGL viewer (embedded meshes) |
| `comparison.html` | DIY vs industrial comparison |
| `cutting_performance.html` | Cutting performance data and charts |

## Safety Warning

**This project involves a Class 4 laser (1 kW, 1070 nm) and high-pressure water (30 MPa / 300 bar). It can cause permanent blindness, severe burns, and fatal injuries. Do not attempt to build or operate this system without proper training and safety equipment.**

### Laser hazards

- 1 kW infrared at 1070 nm is **invisible** — you will not see the beam before it damages your eyes or skin
- Scattered and reflected IR radiation can cause injury even without direct beam exposure
- **OD6+ laser safety goggles** rated for 1060-1080 nm are mandatory (CE EN 207)
- A fully enclosed, interlocked laser enclosure is required (IEC 60825 Class 4)
- Never bypass the door interlock or E-stop
- Post laser warning signs on all access points

### High-pressure hazards

- 30 MPa (300 bar) water can penetrate skin and cause injection injuries
- Never point the nozzle at any body part, even when the laser is off
- Inspect all HP fittings, tubing, and seals before each use
- Use HPLC-rated stainless steel tubing and fittings only — no improvised plumbing
- Keep clear of the HP circuit during pressurization

### Fire hazards

- The laser can ignite combustible materials (wood, acrylic, composites, oils)
- Keep a CO2 fire extinguisher within reach at all times
- IR flame detectors (S7) and thermal sensors (S8) are included in the safety system — do not disable them
- Never leave the system unattended while the laser is firing

### Electrical hazards

- The 60V / 25A power supply and laser driver carry lethal energy
- All electrical connections must be properly insulated and grounded
- Disconnect mains power before servicing any component

### Safety interlocks

The Arduino safety controller monitors all critical sensors and kills the laser on any fault. **Do not operate the laser with any interlock bypassed or disabled:**

- HP pressure (200-350 bar operating range)
- LP cooling flow (> 1 L/min)
- Water leak detection
- IR flame / thermal detection (> 150 C)
- Enclosure door interlock
- Emergency stop (E-stop)

**This is a DIY research project. The authors assume no liability for injury, damage, or loss resulting from the construction or use of this system. You are solely responsible for compliance with all applicable laser safety regulations (IEC 60825, ANSI Z136, OSHA) in your jurisdiction.**

## Support this project

If you find GoLaserCut useful, consider buying me a coffee:

[![PayPal](https://img.shields.io/badge/PayPal-Donate-blue?logo=paypal)](https://paypal.me/fabienlroy)

## License

Apache 2.0 — see [LICENSE](LICENSE).
