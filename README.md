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

## Support this project

If you find GoLaserCut useful, consider buying me a coffee:

[![PayPal](https://img.shields.io/badge/PayPal-Donate-blue?logo=paypal)](https://paypal.me/fabienlroy)

## License

Apache 2.0 — see [LICENSE](LICENSE).
