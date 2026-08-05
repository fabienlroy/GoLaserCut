# GoLaserCut

Port cross-platform de LaserGRBL en Go. Logiciel laser gratuit et open source :
CAM (import SVG/DXF/image → parcours laser) + contrôle machine GRBL en série.

Objectif : remplacer LaserGRBL (C#/WinForms, Windows-only) par un outil natif
macOS / Linux / Windows, zéro dépendance runtime.

## Pourquoi

- LaserGRBL : gratuit mais Windows-only (C#/.NET 3.5/WinForms/GDI+)
- LightBurn : cross-platform mais payant (60 €)
- Rien de gratuit et cross-platform n'existe pour piloter un laser GRBL

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

| Package       | Rôle                                                          |
|--------------|---------------------------------------------------------------|
| `serial/`    | Connexion série, protocole GRBL 1.1 (send/receive, status ?) |
| `grbl/`      | Parser réponses GRBL (ok, error, alarm, status `<...>`)      |
| `import/`    | Parseurs : SVG paths, DXF (LWPOLYLINE, CIRCLE), images       |
| `cam/`       | Parcours laser : mode line (coupe), mode scan (gravure raster)|
| `gcode/`     | Génération G-code : header, M4/M5, overscan, multi-passes    |
| `preview/`   | Rendu 2D du parcours (canvas Gio ou export PNG)              |
| `ui/`        | Interface Gio : fenêtre principale, panels, widgets          |
| `cmd/`       | Point d'entrée CLI + GUI                                     |

## Fonctionnalités (par priorité)

### Phase 1 — Sender GRBL (MVP)
- [ ] Connexion série (auto-detect port, baudrate 115200)
- [ ] Console GRBL interactive (envoi commandes, affichage réponses)
- [ ] Jog XY (flèches, pas configurable)
- [ ] Chargement et envoi de fichier .gcode existant
- [ ] Barre de progression, pause/resume/stop
- [ ] Status polling (`?`) : position, état, vitesse
- [ ] Override feedrate/spindle en temps réel (`0x91`..`0x9D`)

### Phase 2 — Import et prévisualisation
- [ ] Import SVG (paths, polylines, cercles, rectangles, transforms)
- [ ] Import DXF (LWPOLYLINE, CIRCLE, ARC — couche filtrable)
- [ ] Import image (PNG/JPG → niveaux de gris)
- [ ] Canvas 2D avec zoom/pan, affichage des parcours
- [ ] Sélection par couche, réordonnancement

### Phase 3 — CAM laser
- [ ] Mode Line : coupe par contour, puissance/vitesse/passes par couche
- [ ] Mode Scan : gravure raster, espacement lignes, overscan auto
- [ ] Mode 3D : heightmap STL → multi-passes avec Z focus (réutilise laser_stl2gcode)
- [ ] Génération G-code en mémoire + export fichier
- [ ] Estimation temps de coupe

### Phase 4 — Avancé
- [ ] Homing, soft limits, work coordinate systems
- [ ] Macros utilisateur (boutons custom)
- [ ] Calibration laser (grille de test puissance/vitesse)
- [ ] Profils matériaux (contreplaqué 3mm, acrylique, acier 0.3mm)
- [ ] Configuration machine ($$ editor)

## Stack technique

- **Go 1.22+** — zéro CGo sauf serial (go.bug.st/serial)
- **GUI : [Gio](https://gioui.org)** — toolkit Go natif, cross-platform
  (macOS/Linux/Windows/Android/iOS), pas de CGo sur la plupart des
  plateformes, rendu GPU. Alternative : Fyne (plus de widgets prêts
  à l'emploi mais plus lourd).
- **Serial : [go.bug.st/serial](https://pkg.go.dev/go.bug.st/serial)**
  — lib Go standard pour ports série
- **Pas de Electron, pas de webview, pas de Qt**

## Conventions

- Go 1.22+, code lisible, pas de generics inutiles
- Fichiers `snake_case.go`, exports `CamelCase`
- Erreurs : `fmt.Errorf` avec `%w`, pas de `log.Fatal` dans les packages
- Tests : `_test.go` dans chaque package
- Build : `go build ./cmd/golasercut`
- Licence : Apache 2.0

## Protocole GRBL 1.1

Référence : https://github.com/gnea/grbl/wiki/Grbl-v1.1-Interface

### Envoi G-code
- Envoyer ligne par ligne, attendre `ok` ou `error:N` avant la suivante
- Buffer de caractères : max 127 octets par ligne (sans \n)
- Mode streaming : compter les octets en vol, envoyer tant que < 127

### Commandes temps réel (pas de \n, envoi direct)
- `?` → status report `<Idle|MPos:0.000,0.000,0.000|...>`
- `!` → feed hold (pause)
- `~` → cycle resume
- `0x18` → soft reset
- `0x85` → jog cancel
- `0x91`..`0x94` → feedrate override (100%, +10%, -10%, +1%, -1%)
- `0x99`..`0x9D` → spindle override

### Status report
Format : `<State|MPos:X,Y,Z|FS:feed,speed|Ov:f,r,s|...>`
Parser avec state machine, extraire : état, position, feed, speed, overrides.

## Briques réutilisables

Ce projet intègre du code des projets frères :
- Parseur DXF → de `dxf2gcode`
- Parseur SVG → de `svg2gcode`
- Raster 3D + heightmap → de `laser_stl2gcode`

## Machine de référence

- FoxAlien XE-Pro, GRBL 1.1, 3 axes
- Laser Tree 80W (10W optique), diode 450nm, TTL
- $30=1000, $32=1 (laser mode)
- CO₂ assist (détendeur W21.8, 1.5–2 bar)

## TODO (ordre de développement)

1. [ ] `serial/` — connexion, envoi, réception, auto-detect port
2. [ ] `grbl/` — parser réponses (ok, error, alarm, status)
3. [ ] `cmd/` — CLI sender basique (charge .gcode, envoie, progress bar)
4. [ ] `ui/` — fenêtre Gio minimale : console + boutons connect/send
5. [ ] `import/dxf.go` — parser DXF (porter depuis dxf2gcode)
6. [ ] `import/svg.go` — parser SVG paths
7. [ ] `preview/` — canvas 2D rendu des parcours
8. [ ] `cam/line.go` — mode coupe (contour → G-code)
9. [ ] `cam/scan.go` — mode gravure raster avec overscan
10. [ ] `gcode/` — writer complet (header, footer, multi-passes)
11. [ ] Tests d'intégration avec GRBL simulator
