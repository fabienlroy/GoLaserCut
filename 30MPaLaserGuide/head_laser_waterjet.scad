// ============================================================
// WATER-JET GUIDED LASER HEAD — v2
// LP cooling channels open on mating face
// ============================================================

// ============ ADJUSTABLE PARAMETERS ========================

// --- Body ---
D_EXT       = 40;    // outer body diameter (enlarged for threaded fittings)
H_A         = 32;    // piece A height
H_B         = 25;    // piece B height

// --- Junction thread (M × 0.7 pitch) ---
D_FILET     = 20;    // nominal thread diameter M20 × 0.7
PAS_FILET   = 0.7;
H_FILET     = 8;     // threaded length

// --- Sapphire window ---
D_FENETRE   = 10;    // sapphire window diameter
EP_FENETRE  = 1.5;   // window thickness
D_FAISCEAU  = 6;     // beam clearance diameter

// --- HP chamber ---
D_CHAMBRE   = 8;     // HP chamber inner diameter
H_CHAMBRE   = 7;     // HP chamber height

// --- Nozzle ---
D_BUSE_EXT  = 4;     // nozzle pocket diameter (nozzle OD)
H_BUSE      = 5;     // nozzle pocket depth

// --- Piece C — Nozzle cap ---
D_FILET_C   = 10;    // cap thread diameter M10 × 0.7
H_C         = 8;     // piece C height
H_FILET_C   = 5;     // threaded length
D_SORTIE    = 2;     // jet exit diameter (cap)

// --- LP cooling (annular channel on piece A mating face) ---
RAIN_D      = 6.0;   // annular chamber height (enlarged)
RAIN_R_INT  = D_FENETRE/2 + 1.5;   // annulus inner radius (6.5mm)
RAIN_R_EXT  = 12.0;                 // annulus outer radius (fixed, thick wall for tapping)
RAIN_W      = RAIN_R_EXT - RAIN_R_INT;  // annulus width (5.5mm)
RAIN_R      = (RAIN_R_INT + RAIN_R_EXT) / 2;  // center radius (9.25mm)
Z_RADIAL    = H_FILET + RAIN_D / 2 + 4;  // radial holes abs. height (+4mm above thread)

// --- G1/8" BSP elbow fittings for 6mm OD tubing ---
D_TARAUD    = 8.8;   // G1/8" BSP pilot hole diameter
L_TARAUD    = 10;    // tapping depth
D_MEPLAT    = 2;     // flat depth (machined flat surface)
H_MEPLAT    = 14;    // flat height (Z)
W_MEPLAT    = 14;    // flat width (Y)

// --- Collimation screws (2 rows of 3 at 120 deg, offset 60 deg) ---
D_VIS       = 2.5;   // M2.5 hole diameter
H_VIS_1     = 3;     // distance from top for row 1
H_VIS_2     = 13;    // distance from top for row 2

// --- O-rings (ISO 3601) ---
// Window: OR 8 × 1.5
OR_W_ID     = 8;
OR_W_CS     = 1.5;
GR_W_W      = 2.0;   // groove width
GR_W_D      = 1.1;   // groove depth

// LP mating face: OR 18 × 1.5 (seals LP circuit between A and B)
OR_BP_ID    = D_FILET - 2;  // 18mm
OR_BP_CS    = 1.5;
GR_BP_W     = 2.0;
GR_BP_D     = 1.1;

// Nozzle: OR 3 × 1.0
OR_N_ID     = 3;
OR_N_CS     = 1.0;
GR_N_W      = 1.4;
GR_N_D      = 0.75;

// --- Red laser guide (SMA905 + focus optics) ---
SMA_D_FERRULE  = 6.35;   // SMA905 ferrule diameter
SMA_D_BODY     = 8.0;    // connector body diameter
SMA_D_NUT      = 9.5;    // hex nut diameter
SMA_H_FERRULE  = 5;      // ferrule length
SMA_H_BODY     = 12;     // connector body length
SMA_H_NUT      = 4;      // nut height
FOCUS_D        = 12;     // focus tube diameter
FOCUS_H        = 25;     // focus tube length
FOCUS_D_LENS   = 8;      // internal lens diameter
FIBER_D        = 3;      // fiber optic cable diameter
FIBER_L        = 40;     // visible fiber length

// --- Water tank (30-100L PE) ---
TANK_L         = 400;    // length (mm)
TANK_W         = 300;    // width
TANK_H         = 420;    // height (~50L)
TANK_WALL      = 4;      // wall thickness
TANK_R         = 15;     // fillet radius

// --- Reverse osmosis ---
RO_D           = 60;     // membrane diameter
RO_L           = 250;    // housing length
RO_D_CAP       = 50;     // end cap diameter
RO_D_PORT      = 8;      // inlet/outlet port diameter

// --- Rendering ---
$fn = 64;

// ============ COMPUTED VALUES ===============================
R_EXT       = D_EXT / 2;
R_FILET     = D_FILET / 2;

// ============ UTILITY MODULES ==============================

module gorge_torique_face(id, cs, w, d, z) {
    // O-ring groove on horizontal face (seal between two pieces)
    translate([0, 0, z])
        rotate_extrude()
            translate([id/2 + cs/2, 0])
                square([cs + 0.3, w], center=true);
}

module gorge_torique_axiale(id, cs, w, d, z) {
    // O-ring groove in bore (around window)
    translate([0, 0, z])
        rotate_extrude()
            translate([id/2 + cs/2, 0])
                circle(d=cs);
}

module fenetre_saphir() {
    color("LightBlue", 0.5)
    cylinder(d=D_FENETRE, h=EP_FENETRE);
}

module joint_torique(id, cs, z) {
    color("Black")
    translate([0, 0, z])
        rotate_extrude()
            translate([id/2 + cs/2, 0])
                circle(d=cs);
}

// --- ISO metric thread (60 deg profile) ---

module filet_male(d_nom, pitch, h) {
    H = pitch * sqrt(3) / 2;
    dp = 5 * H / 8;
    d_min = d_nom - 2 * dp;
    r_mid = (d_nom + d_min) / 4;
    n = h / pitch;
    sl = round(n * 24);

    union() {
        cylinder(d=d_min, h=h);
        intersection() {
            cylinder(d=d_nom, h=h);
            translate([0, 0, -pitch])
                linear_extrude(height=h + 2*pitch, twist=-(n+2)*360,
                              slices=sl + 48, convexity=10)
                    translate([r_mid, 0])
                        circle(r=dp, $fn=3);
        }
    }
}

module filet_femelle_ridges(d_nom, pitch, h) {
    H = pitch * sqrt(3) / 2;
    dp = 5 * H / 8;
    d_min = d_nom - 2 * dp;
    r_mid = (d_nom + d_min) / 4;
    n = h / pitch;
    sl = round(n * 24);

    intersection() {
        difference() {
            cylinder(d=d_nom + 2, h=h);
            cylinder(d=d_min, h=h + 0.01);
        }
        translate([0, 0, -pitch])
            linear_extrude(height=h + 2*pitch, twist=-(n+2)*360,
                          slices=sl + 48, convexity=10)
                translate([r_mid, 0])
                    circle(r=dp, $fn=3);
    }
}

module filet_femelle(d_nom, pitch, h) {
    H = pitch * sqrt(3) / 2;
    dp = 5 * H / 8;
    d_min = d_nom - 2 * dp;
    r_mid = (d_nom + d_min) / 4;
    cylinder(d=2 * (r_mid - dp * 0.3), h=h);
}

// --- 1mm chamfer modules ---
CHAMFER = 1;  // chamfer size (mm)

module chanfrein_haut(d, h) {
    // 45 deg chamfer on outer top edge of cylinder
    translate([0, 0, h - CHAMFER])
        difference() {
            cylinder(d=d + 1, h=CHAMFER + 0.5);
            translate([0, 0, -0.01])
                cylinder(d1=d, d2=d - 2*CHAMFER, h=CHAMFER + 0.01);
        }
}

module chanfrein_bas(d) {
    // 45 deg chamfer on outer bottom edge of cylinder
    translate([0, 0, -0.5])
        difference() {
            cylinder(d=d + 1, h=CHAMFER + 0.5);
            translate([0, 0, 0.5])
                cylinder(d1=d - 2*CHAMFER, d2=d, h=CHAMFER + 0.01);
            cylinder(d=d - 2*CHAMFER, h=CHAMFER + 0.51);
        }
}

module chanfrein_bas_z(d, z) {
    // 45 deg chamfer at given Z height
    translate([0, 0, z - 0.5])
        difference() {
            cylinder(d=d + 1, h=CHAMFER + 0.5);
            translate([0, 0, 0.5])
                cylinder(d1=d - 2*CHAMFER, d2=d, h=CHAMFER + 0.01);
            cylinder(d=d - 2*CHAMFER, h=CHAMFER + 0.51);
        }
}

module chanfrein_filet_male(d_filet, z_bas) {
    // Chamfer on male thread tip (eases engagement)
    translate([0, 0, z_bas - 0.01])
        difference() {
            cylinder(d=d_filet + 0.1, h=CHAMFER + 0.01);
            cylinder(d1=d_filet - 2*CHAMFER, d2=d_filet + 0.1, h=CHAMFER + 0.02);
        }
}

module chanfrein_filet_femelle(d_filet, z_entree) {
    // Countersink at female thread entry
    translate([0, 0, z_entree - 0.01])
        cylinder(d1=d_filet, d2=d_filet + 2*CHAMFER, h=CHAMFER + 0.01);
}


// ============ RED LASER GUIDE (SMA905) =====================

module sma905_connector() {
    // Hex nut
    color("Silver", 0.9)
    translate([0, 0, SMA_H_BODY])
        cylinder(d=SMA_D_NUT, h=SMA_H_NUT, $fn=6);

    // Connector body
    color("DimGray", 0.9)
    cylinder(d=SMA_D_BODY, h=SMA_H_BODY);

    // Ferrule (extends below)
    color("Silver")
    translate([0, 0, -SMA_H_FERRULE])
        cylinder(d=SMA_D_FERRULE, h=SMA_H_FERRULE);

    // Fiber optic cable (exits from top)
    color("Orange", 0.8)
    translate([0, 0, SMA_H_BODY + SMA_H_NUT])
        cylinder(d=FIBER_D, h=FIBER_L);
}

module focus_tube() {
    color("DarkSlateGray", 0.9)
    difference() {
        cylinder(d=FOCUS_D, h=FOCUS_H);
        // Internal bore
        translate([0, 0, -0.01])
            cylinder(d=FOCUS_D_LENS, h=FOCUS_H + 0.02);
    }
    // Focus lens (centered)
    color("LightBlue", 0.4)
    translate([0, 0, FOCUS_H/2 - 1])
        cylinder(d=FOCUS_D_LENS - 0.5, h=2);
}

module laser_guide() {
    // Focus tube (screws onto optical plate A)
    focus_tube();

    // SMA905 connector on top of tube
    translate([0, 0, FOCUS_H])
        sma905_connector();

    // Red beam (visible ray, symbolic)
    color("Red", 0.3)
    translate([0, 0, -10])
        cylinder(d1=0.5, d2=3, h=10);
}


// ============ PE WATER TANK ================================

module water_tank() {
    color("SteelBlue", 0.4)
    difference() {
        // Outer shell with rounded edges
        minkowski() {
            cube([TANK_L - 2*TANK_R, TANK_W - 2*TANK_R, TANK_H - 2*TANK_R]);
            sphere(r=TANK_R);
        }
        // Interior cavity
        translate([TANK_WALL, TANK_WALL, TANK_WALL])
            minkowski() {
                cube([TANK_L - 2*TANK_R - 2*TANK_WALL,
                      TANK_W - 2*TANK_R - 2*TANK_WALL,
                      TANK_H - 2*TANK_R - 2*TANK_WALL]);
                sphere(r=TANK_R);
            }
        // Cap opening (top)
        translate([(TANK_L - 2*TANK_R)/2, (TANK_W - 2*TANK_R)/2, TANK_H - TANK_R])
            cylinder(d=50, h=TANK_R + 1);
    }
    // Screw cap
    color("DarkBlue", 0.7)
    translate([(TANK_L - 2*TANK_R)/2, (TANK_W - 2*TANK_R)/2, TANK_H - TANK_R])
        cylinder(d=55, h=8);

    // Outlet fitting (bottom, front face)
    color("Gray")
    translate([(TANK_L - 2*TANK_R)/2, -TANK_R, TANK_R + 20])
        rotate([-90, 0, 0])
            cylinder(d=12, h=15);

    // Visible water level
    color("CornflowerBlue", 0.3)
    translate([TANK_WALL + 2, TANK_WALL + 2, TANK_WALL + 2])
        cube([TANK_L - 2*TANK_R - 2*TANK_WALL - 4,
              TANK_W - 2*TANK_R - 2*TANK_WALL - 4,
              (TANK_H - 2*TANK_R) * 0.7]);
}


// ============ REVERSE OSMOSIS UNIT =========================

module ro_unit() {
    // Membrane housing (main tube)
    color("White", 0.8)
    difference() {
        union() {
            cylinder(d=RO_D, h=RO_L);
            // Inlet end cap
            translate([0, 0, -5])
                cylinder(d=RO_D_CAP, h=8);
            // Outlet end cap
            translate([0, 0, RO_L - 3])
                cylinder(d=RO_D_CAP, h=8);
        }
        // Internal bore
        translate([0, 0, 2])
            cylinder(d=RO_D - 6, h=RO_L - 4);
    }

    // Raw water inlet (bottom)
    color("Gray")
    translate([0, 0, -15])
        cylinder(d=RO_D_PORT, h=12);

    // Pure water outlet (top)
    color("Gray")
    translate([0, 0, RO_L + 3])
        cylinder(d=RO_D_PORT, h=12);

    // Reject outlet (side, near top)
    color("Gray")
    translate([0, RO_D/2, RO_L - 30])
        rotate([-90, 0, 0])
            cylinder(d=RO_D_PORT, h=15);

    // Internal membrane (symbolic)
    color("Khaki", 0.5)
    translate([0, 0, 10])
        cylinder(d=RO_D - 8, h=RO_L - 20);
}


// ============ PIECE A — OPTICAL PLATE ======================

module piece_A() {
    color("Gold", 0.8)
    difference() {
        union() {
            // Main body (above threaded zone)
            translate([0, 0, H_FILET])
                cylinder(d=D_EXT, h=H_A - H_FILET);

            // M20 × 0.7 threaded zone
            filet_male(D_FILET, PAS_FILET, H_FILET);
        }

        // --- Beam passage (through hole) ---
        translate([0, 0, -1])
            cylinder(d=D_FAISCEAU, h=H_A + 2);

        // Flared laser entry (top)
        translate([0, 0, H_A - 5])
            cylinder(d1=D_FAISCEAU, d2=D_FAISCEAU + 4, h=5.1);

        // --- Window O-ring groove (bottom face, around beam) ---
        gorge_torique_face(OR_W_ID, OR_W_CS, GR_W_W, GR_W_D, -GR_W_W/2);

        // --- LP cooling annular chamber (mating face, above thread) ---
        // Sealed by top face of B when A is screwed in
        translate([0, 0, H_FILET - 0.1])
            rotate_extrude($fn=64)
                translate([RAIN_R_INT, 0])
                    square([RAIN_W, RAIN_D + 0.1]);

        // --- Flat + tapped hole for LP water inlet (0 deg, G1/8" BSP) ---
        rotate([0, 0, 0]) {
            // Flat (machined surface for elbow fitting)
            translate([R_EXT - D_MEPLAT, -W_MEPLAT/2, H_FILET])
                cube([D_MEPLAT + 10, W_MEPLAT, H_MEPLAT]);
            // G1/8" BSP tapped hole (through wall to chamber)
            translate([0, 0, Z_RADIAL])
                rotate([0, 90, 0])
                    translate([0, 0, RAIN_R_EXT - 0.5])
                        cylinder(d=D_TARAUD, h=R_EXT + 1);
        }

        // --- Flat + tapped hole for LP water outlet (180 deg, G1/8" BSP) ---
        rotate([0, 0, 180]) {
            translate([R_EXT - D_MEPLAT, -W_MEPLAT/2, H_FILET])
                cube([D_MEPLAT + 10, W_MEPLAT, H_MEPLAT]);
            translate([0, 0, Z_RADIAL])
                rotate([0, 90, 0])
                    translate([0, 0, RAIN_R_EXT - 0.5])
                        cylinder(d=D_TARAUD, h=R_EXT + 1);
        }

        // --- LP face O-ring groove (seals LP circuit between A and B) ---
        gorge_torique_face(OR_BP_ID, OR_BP_CS, GR_BP_W, GR_BP_D, H_FILET - GR_BP_W/2);

        // --- Collimation screws ---
        // Row 1: 3x M2.5 holes at 120 deg, 3mm from top
        for (i = [0:2]) {
            rotate([0, 0, i * 120])
                translate([0, 0, H_A - H_VIS_1])
                    rotate([0, 90, 0])
                        cylinder(d=D_VIS, h=D_EXT);
        }
        // Row 2: 3x M2.5 holes at 120 deg, offset 60 deg, 13mm from top
        for (i = [0:2]) {
            rotate([0, 0, i * 120 + 60])
                translate([0, 0, H_A - H_VIS_2])
                    rotate([0, 90, 0])
                        cylinder(d=D_VIS, h=D_EXT);
        }

        // --- 2 pin spanner holes (top face) ---
        // At 90 deg and 270 deg to avoid collimation screws
        translate([0, R_EXT - 3, H_A - H_ERGOT])
            cylinder(d=D_ERGOT, h=H_ERGOT + 0.01);
        translate([0, -(R_EXT - 3), H_A - H_ERGOT])
            cylinder(d=D_ERGOT, h=H_ERGOT + 0.01);

        // --- 1mm chamfers on outer edges ---
        chanfrein_haut(D_EXT, H_A);
        chanfrein_bas_z(D_EXT, H_FILET);

        // --- Male thread entry chamfer M20 ---
        chanfrein_filet_male(D_FILET, 0);
    }
}


// ============ PIECE B — HP CHAMBER =========================

module piece_B() {
    color("Peru", 0.8)
    difference() {
        // --- Solid body ---
        cylinder(d=D_EXT, h=H_B);

        // --- Upper internal thread M20 × 0.7 (receives piece A) ---
        translate([0, 0, H_B - H_FILET - 0.01])
            filet_male(D_FILET, PAS_FILET, H_FILET + 0.02);

        // --- Window pocket (shoulder, accessible from top) ---
        Z_FEN = H_B - H_FILET - EP_FENETRE;
        translate([0, 0, Z_FEN])
            cylinder(d=D_FENETRE + 0.2, h=EP_FENETRE + 0.1);

        // --- Window O-ring groove (below window) ---
        gorge_torique_face(OR_W_ID, OR_W_CS, GR_W_W, GR_W_D,
                           Z_FEN - GR_W_W/2);

        // --- Beam passage (window to chamber) ---
        translate([0, 0, H_BUSE + H_CHAMBRE - 0.01])
            cylinder(d=D_FAISCEAU, h=Z_FEN - H_BUSE - H_CHAMBRE + 1);

        // --- HP chamber ---
        translate([0, 0, H_BUSE])
            cylinder(d=D_CHAMBRE, h=H_CHAMBRE);

        // --- Flat + tapped hole for HP water inlet (G1/8" BSP elbow) ---
        // Flat on body
        translate([R_EXT - D_MEPLAT, -W_MEPLAT/2, H_BUSE + H_CHAMBRE/2 - H_MEPLAT/2])
            cube([D_MEPLAT + 10, W_MEPLAT, H_MEPLAT]);
        // G1/8" BSP tapped hole (through wall to HP chamber)
        translate([0, 0, H_BUSE + H_CHAMBRE / 2])
            rotate([0, 90, 0])
                translate([0, 0, D_CHAMBRE/2 - 0.1])
                    cylinder(d=D_TARAUD, h=R_EXT + 1);

        // --- Nozzle pocket + lower internal thread (receives piece C) ---
        // Through bore for nozzle (accessible from bottom when C removed)
        translate([0, 0, -0.01])
            cylinder(d=D_BUSE_EXT + 0.2, h=H_BUSE + 0.02);

        // Internal thread M10 × 0.7 (receives nozzle cap C)
        translate([0, 0, -0.01])
            filet_male(D_FILET_C, PAS_FILET, H_FILET_C + 0.01);

        // --- Nozzle O-ring groove (in bore, above thread) ---
        gorge_torique_face(OR_N_ID, OR_N_CS, GR_N_W, GR_N_D,
                           H_FILET_C + 0.5);

        // --- 1mm chamfers ---
        chanfrein_haut(D_EXT, H_B);
        chanfrein_bas(D_EXT);

        // --- Female thread entry chamfers ---
        // M20 × 0.7 (top, receives piece A)
        chanfrein_filet_femelle(D_FILET + 0.2, H_B);
        // M10 × 0.7 (bottom, receives piece C)
        chanfrein_filet_femelle(D_FILET_C + 0.2, 0);
    }
}


// ============ PIECE C — NOZZLE CAP =========================

// Pin spanner parameters
D_ERGOT     = 3.1;   // pin hole diameter
ESP_ERGOT   = 14;    // pin spacing (center to center)
H_ERGOT     = 3;     // pin hole depth

module piece_C() {
    color("Sienna", 0.9)
    difference() {
        union() {
            // Cap body (outer flange)
            cylinder(d=D_EXT, h=H_C - H_FILET_C);

            // M10 × 0.7 threaded zone (screws into B from below)
            translate([0, 0, H_C - H_FILET_C])
                filet_male(D_FILET_C, PAS_FILET, H_FILET_C);
        }

        // --- Water jet passage (through hole, axial) ---
        translate([0, 0, -0.01])
            cylinder(d=D_SORTIE, h=H_C + 0.02);

        // --- 2 pin spanner holes (bottom face) ---
        translate([R_EXT - 3, 0, -0.01])
            cylinder(d=D_ERGOT, h=H_ERGOT + 0.01);
        translate([-(R_EXT - 3), 0, -0.01])
            cylinder(d=D_ERGOT, h=H_ERGOT + 0.01);

        // --- 1mm chamfers ---
        chanfrein_bas(D_EXT);

        // --- Male thread entry chamfer M10 ---
        chanfrein_filet_male(D_FILET_C, H_C);
    }
}


// ============ SAPPHIRE NOZZLE ==============================

module buse_saphir() {
    color("LightBlue", 0.6)
    difference() {
        cylinder(d=D_BUSE_EXT, h=3);
        translate([0, 0, -0.01])
            cylinder(d=0.1, h=3.02);  // micro-drilled orifice 100 um
    }
}


// ============ ASSEMBLY =====================================

module assemblage() {
    // Piece C (nozzle cap, bottom)
    translate([0, 0, -(H_C - H_FILET_C)])
        piece_C();

    // Nozzle O-ring
    joint_torique(OR_N_ID, OR_N_CS, H_FILET_C + 0.5);

    // Sapphire nozzle
    translate([0, 0, H_FILET_C + 1])
        buse_saphir();

    // Piece B (main body)
    piece_B();

    // Sapphire window
    Z_FEN = H_B - H_FILET - EP_FENETRE;
    translate([0, 0, Z_FEN])
        fenetre_saphir();

    // Window O-rings (top and bottom)
    joint_torique(OR_W_ID, OR_W_CS, Z_FEN - 1);
    joint_torique(OR_W_ID, OR_W_CS, Z_FEN + EP_FENETRE + 0.5);

    // LP face O-ring (on A/B mating face)
    joint_torique(OR_BP_ID, OR_BP_CS, H_B - 0.5);

    // Piece A (screwed into B from top)
    translate([0, 0, H_B - H_FILET])
        piece_A();

    // Red laser guide (above piece A, on optical axis)
    translate([0, 0, H_B - H_FILET + H_A + 2])
        laser_guide();

    // PE water tank (offset to right)
    translate([120, -TANK_W/2 + TANK_R, -(H_C)])
        water_tank();

    // Reverse osmosis unit (between tank and head, horizontal)
    translate([60, 0, TANK_H/2 - H_C])
        rotate([0, 0, -90])
            ro_unit();
}

module vue_eclatee() {
    ECART = 15;

    // Piece C (bottom)
    translate([0, 0, -3*ECART])
        piece_C();

    // Nozzle O-ring
    translate([0, 0, -2*ECART])
        joint_torique(OR_N_ID, OR_N_CS, 0);

    // Sapphire nozzle
    translate([0, 0, -ECART])
        buse_saphir();

    // Piece B
    piece_B();

    // Lower window O-ring
    translate([0, 0, H_B + ECART - 2])
        joint_torique(OR_W_ID, OR_W_CS, 0);

    // Sapphire window
    translate([0, 0, H_B + ECART])
        fenetre_saphir();

    // Upper window O-ring
    translate([0, 0, H_B + ECART + EP_FENETRE + 2])
        joint_torique(OR_W_ID, OR_W_CS, 0);

    // LP face O-ring
    translate([0, 0, H_B + 2*ECART])
        joint_torique(OR_BP_ID, OR_BP_CS, 0);

    // Piece A (top)
    translate([0, 0, H_B + 2*ECART + 3])
        piece_A();
}


module demi_coupe() {
    translate([0, -D_EXT - 1, -80])
        cube([D_EXT + 5, D_EXT * 2 + 2, 250]);
}

module vue_eclatee_coupe() {
    ECART = 15;

    // Piece C — nozzle cap
    color("Sienna", 0.9)
    translate([0, 0, -3*ECART])
    difference() {
        piece_C();
        demi_coupe();
    }

    // Nozzle O-ring
    translate([0, 0, -2*ECART])
        joint_torique(OR_N_ID, OR_N_CS, 0);

    // Sapphire nozzle
    color("LightCyan", 0.7)
    translate([0, 0, -ECART])
    difference() {
        buse_saphir();
        demi_coupe();
    }

    // Piece B — HP chamber
    color("SaddleBrown", 0.9)
    difference() {
        piece_B();
        demi_coupe();
    }

    // Lower window O-ring
    translate([0, 0, H_B + ECART - 2])
        joint_torique(OR_W_ID, OR_W_CS, 0);

    // Sapphire window
    color("CornflowerBlue", 0.7)
    translate([0, 0, H_B + ECART])
    difference() {
        fenetre_saphir();
        demi_coupe();
    }

    // Upper window O-ring
    translate([0, 0, H_B + ECART + EP_FENETRE + 2])
        joint_torique(OR_W_ID, OR_W_CS, 0);

    // LP face O-ring
    translate([0, 0, H_B + 2*ECART])
        joint_torique(OR_BP_ID, OR_BP_CS, 0);

    // Piece A — optical plate
    color("Goldenrod", 0.9)
    translate([0, 0, H_B + 2*ECART + 3])
    difference() {
        piece_A();
        demi_coupe();
    }

    // Red laser guide (above piece A)
    translate([0, 0, H_B + 2*ECART + 3 + H_A + ECART])
        laser_guide();

    // PE water tank (offset to right)
    translate([120, -TANK_W/2 + TANK_R, -(H_C)])
        water_tank();

    // Reverse osmosis unit (between tank and head)
    translate([60, 0, TANK_H/2 - H_C])
        rotate([0, 0, -90])
            ro_unit();
}


// ============ DISPLAY ======================================
// Uncomment ONE of the lines below:
// (Guarded by EXPORT_PART to prevent rendering during include)

if (is_undef(EXPORT_PART)) {

// Half-section assembly view
// difference() {
//     assemblage();
//     translate([0, -D_EXT, -H_C - 2])
//         cube([D_EXT, D_EXT*2, H_A + H_B + H_C + 4]);
// }

// Full assembly view
// assemblage();

// Exploded view
// vue_eclatee();

// Exploded half-section view
// vue_eclatee_coupe();

// Half-section assembly view
// difference() {
//     assemblage();
//     translate([0, -D_EXT, -H_C - 2])
//         cube([D_EXT, D_EXT*2, H_A + H_B + H_C + 4]);
// }

// Piece A alone (flipped, channel face visible)
// rotate([180, 0, 0]) piece_A();

// Piece B alone
piece_B();

// Piece C alone (nozzle cap)
// piece_C();

} // end EXPORT_PART guard
