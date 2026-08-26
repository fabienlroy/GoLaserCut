# GoLaserCut

Cross-platform Go port of LaserGRBL. Free, open-source laser cutter software:
CAM (import SVG/DXF/image → laser toolpath) + GRBL machine control over serial.

Goal: replace LaserGRBL (C#/WinForms, Windows-only) with a native tool for
macOS / Linux / Windows, zero runtime dependencies.

## Why

- LaserGRBL: free but Windows-only (C#/.NET 3.5/WinForms/GDI+)
- LightBurn: cross-platform but paid (60 EUR)
- Nothing free and cross-platform exists to drive a GRBL laser

## Architecture

```
┌──────────────────────────────────────────────┐
│                   GUI (Gio)                  │
│  Canvas 2D  │  Jog pad  │  Console  │  Logs  │
├──────────────────────────────────────────────┤
│                   Engine                      │
│  Import  │  CAM  │  GCode Gen  │  Simulator  │
├──────────────────────────────────────────────┤
│                   Serial                      │
│  GRBL 1.1 protocol  │  Status polling         │
└──────────────────────────────────────────────┘
```

## Packages

| Package       | Role                                                          |
|--------------|---------------------------------------------------------------|
| `serial/`    | Serial connection, GRBL 1.1 protocol (send/receive, status ?) |
| `grbl/`      | GRBL response parser (ok, error, alarm, status `<...>`)       |
| `import/`    | Parsers: SVG paths, DXF (LWPOLYLINE, CIRCLE), images         |
| `cam/`       | Laser toolpaths: line mode (cut), scan mode (raster engrave)  |
| `gcode/`     | G-code generation: header, M4/M5, overscan, multi-pass       |
| `preview/`   | 2D toolpath rendering (Gio canvas or PNG export)              |
| `ui/`        | Gio interface: main window, panels, widgets                   |
| `cmd/`       | CLI + GUI entry point                                         |

## Features (by priority)

### Phase 1 — GRBL Sender (MVP)
- [ ] Serial connection (auto-detect port, baudrate 115200)
- [ ] Interactive GRBL console (send commands, display responses)
- [ ] XY jog (arrow keys, configurable step)
- [ ] Load and send existing .gcode file
- [ ] Progress bar, pause/resume/stop
- [ ] Status polling (`?`): position, state, speed
- [ ] Real-time feedrate/spindle override (`0x91`..`0x9D`)

### Phase 2 — Import and preview
- [ ] SVG import (paths, polylines, circles, rectangles, transforms)
- [ ] DXF import (LWPOLYLINE, CIRCLE, ARC — filterable layers)
- [ ] Image import (PNG/JPG → grayscale)
- [ ] 2D canvas with zoom/pan, toolpath display
- [ ] Layer selection, reordering

### Phase 3 — Laser CAM
- [ ] Line mode: contour cut, power/speed/passes per layer
- [ ] Scan mode: raster engrave, line spacing, auto overscan
- [ ] 3D mode: STL heightmap → multi-pass with Z focus (reuses laser_stl2gcode)
- [ ] In-memory G-code generation + file export
- [ ] Cut time estimation

### Phase 4 — Advanced
- [ ] Homing, soft limits, work coordinate systems
- [ ] User macros (custom buttons)
- [ ] Laser calibration (power/speed test grid)
- [ ] Material profiles (3mm plywood, acrylic, 0.3mm steel)
- [ ] Machine configuration ($$ editor)

## Tech stack

- **Go 1.22+** — zero CGo except serial (go.bug.st/serial)
- **GUI: [Gio](https://gioui.org)** — native Go toolkit, cross-platform
  (macOS/Linux/Windows/Android/iOS), no CGo on most platforms, GPU rendering.
  Alternative: Fyne (more ready-made widgets but heavier).
- **Serial: [go.bug.st/serial](https://pkg.go.dev/go.bug.st/serial)**
  — standard Go serial port library
- **No Electron, no webview, no Qt**

## Conventions

- Go 1.22+, readable code, no unnecessary generics
- Files `snake_case.go`, exports `CamelCase`
- Errors: `fmt.Errorf` with `%w`, no `log.Fatal` in packages
- Tests: `_test.go` in each package
- Build: `go build ./cmd/golasercut`
- License: Apache 2.0

## GRBL 1.1 Protocol

Reference: https://github.com/gnea/grbl/wiki/Grbl-v1.1-Interface

### Sending G-code
- Send line by line, wait for `ok` or `error:N` before next
- Character buffer: max 127 bytes per line (without \n)
- Streaming mode: count in-flight bytes, send while < 127

### Real-time commands (no \n, send directly)
- `?` → status report `<Idle|MPos:0.000,0.000,0.000|...>`
- `!` → feed hold (pause)
- `~` → cycle resume
- `0x18` → soft reset
- `0x85` → jog cancel
- `0x91`..`0x94` → feedrate override (100%, +10%, -10%, +1%, -1%)
- `0x99`..`0x9D` → spindle override

### Status report
Format: `<State|MPos:X,Y,Z|FS:feed,speed|Ov:f,r,s|...>`
Parse with state machine, extract: state, position, feed, speed, overrides.

## Reusable components

This project incorporates code from sibling projects:
- DXF parser → from `dxf2gcode`
- SVG parser → from `svg2gcode`
- 3D raster + heightmap → from `laser_stl2gcode`

## Reference machine

- FoxAlien XE-Pro, GRBL 1.1, 3 axes
- Laser Tree 80W (10W optical), 450nm diode, TTL
- $30=1000, $32=1 (laser mode)
- CO2 assist (W21.8 regulator, 1.5-2 bar)

## TODO (development order)

1. [ ] `serial/` — connection, send, receive, auto-detect port
2. [ ] `grbl/` — response parser (ok, error, alarm, status)
3. [ ] `cmd/` — basic CLI sender (load .gcode, send, progress bar)
4. [ ] `ui/` — minimal Gio window: console + connect/send buttons
5. [ ] `import/dxf.go` — DXF parser (port from dxf2gcode)
6. [ ] `import/svg.go` — SVG path parser
7. [ ] `preview/` — 2D canvas toolpath rendering
8. [ ] `cam/line.go` — cut mode (contour → G-code)
9. [ ] `cam/scan.go` — raster engrave mode with overscan
10. [ ] `gcode/` — complete writer (header, footer, multi-pass)
11. [ ] Integration tests with GRBL simulator
