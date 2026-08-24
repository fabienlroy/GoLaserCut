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

🚧 **Early development** — contributions welcome.

See [project_plan.md](project_plan.md) for architecture, roadmap and conventions.

## Build

```bash
go build ./cmd/golasercut
```

Requires Go 1.22+.

## Features (planned)

- GRBL 1.1 sender with real-time status, jog, overrides
- SVG / DXF / image import with 2D preview
- Line mode (vector cutting with multi-pass)
- Scan mode (raster engraving with overscan)
- 3D engraving from STL heightmaps
- Material profiles (plywood, acrylic, thin steel)
- Native GUI via [Gio](https://gioui.org) — no Electron, no webview

## 30 MPa Water-Jet Guided Laser Head

Parametric OpenSCAD model of a water-jet guided laser cutting head (30 MPa, 1 kW fiber laser).

- [3D Viewer](https://claude.ai/code/artifact/5d1fa2bf-d95c-4771-8e8f-19089802bc26) — Interactive WebGL viewer (explode, clip, toggle pieces)
- [DIY vs Industrial Comparison](https://claude.ai/code/artifact/5689606b-0189-43bd-8e68-c377c5c5598c) — Cost and specs vs Synova LMJ, Chinese WJGL
- [Cutting Performance](https://claude.ai/code/artifact/58c09f3d-e962-4e3f-971e-371441b1937c) — Speed, depth (single & multi-pass), kerf, cut quality

Source files in [`30MPaLaserGuide/`](30MPaLaserGuide/):
- `head_laser_waterjet.scad` — Parametric CAD model (3 pieces, M20/M10 threads, G1/8" BSP fittings)
- `BillOfMaterial.txt` — Full BOM with suppliers and pricing (~3 000 EUR total)
- `viewer_3d.html` — Standalone WebGL viewer (embedded meshes)
- `comparison.html` — DIY vs industrial comparison
- `cutting_performance.html` — Cutting performance data

## Support this project

If you find GoLaserCut useful, consider buying me a coffee:

[![PayPal](https://img.shields.io/badge/PayPal-Donate-blue?logo=paypal)](https://paypal.me/fabienlroy)

## License

Apache 2.0 — see [LICENSE](LICENSE).
