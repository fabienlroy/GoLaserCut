// ============================================================
// TÊTE LASER GUIDÉE PAR JET D'EAU — v2
// Rainures de refroidissement BP ouvertes sur face de joint
// ============================================================

// ============ PARAMÈTRES MODIFIABLES =======================

// --- Corps ---
D_EXT       = 40;    // Ø extérieur corps (augmenté pour raccords filetés)
H_A         = 32;    // hauteur pièce A
H_B         = 25;    // hauteur pièce B

// --- Filetage jonction (M × 0.7 de pas) ---
D_FILET     = 20;    // Ø nominal filetage M20 × 0.7
PAS_FILET   = 0.7;
H_FILET     = 8;     // longueur filetée

// --- Fenêtre saphir ---
D_FENETRE   = 10;    // Ø fenêtre saphir
EP_FENETRE  = 1.5;   // épaisseur fenêtre
D_FAISCEAU  = 6;     // Ø passage libre faisceau

// --- Chambre HP ---
D_CHAMBRE   = 8;     // Ø intérieur chambre HP
H_CHAMBRE   = 7;     // hauteur chambre HP

// --- Buse ---
D_BUSE_EXT  = 4;     // Ø logement buse saphir (extérieur buse)
H_BUSE      = 5;     // profondeur logement buse

// --- Pièce C — Bouchon buse ---
D_FILET_C   = 10;    // Ø filetage bouchon M10 × 0.7
H_C         = 8;     // hauteur pièce C
H_FILET_C   = 5;     // longueur filetée
D_SORTIE    = 2;     // Ø sortie jet (bouchon)

// --- Refroidissement BP (anneau sur face de joint pièce A) ---
RAIN_D      = 6.0;   // hauteur chambre annulaire (augmentée)
RAIN_R_INT  = D_FENETRE/2 + 1.5;   // rayon intérieur anneau (6.5mm)
RAIN_R_EXT  = 12.0;                 // rayon extérieur anneau (fixe, paroi épaisse pour taraudage)
RAIN_W      = RAIN_R_EXT - RAIN_R_INT;  // largeur anneau (5.5mm)
RAIN_R      = (RAIN_R_INT + RAIN_R_EXT) / 2;  // rayon centre (9.25mm)
Z_RADIAL    = H_FILET + RAIN_D / 2 + 4;  // hauteur abs. trous radiaux (+4mm au-dessus du filetage)

// --- Raccords coudés G1/8" BSP pour tubes 6mm OD ---
D_TARAUD    = 8.8;   // Ø avant-trou taraudage G1/8" BSP
L_TARAUD    = 10;    // profondeur taraudage
D_MEPLAT    = 2;     // profondeur du méplat (usinage surface plate)
H_MEPLAT    = 14;    // hauteur du méplat (Z)
W_MEPLAT    = 14;    // largeur du méplat (Y)

// --- Vis de collimation (2 rangées de 3 à 120°, décalées de 60°) ---
D_VIS       = 2.5;   // Ø perçage M2.5
H_VIS_1     = 3;     // distance du haut pour rangée 1
H_VIS_2     = 13;    // distance du haut pour rangée 2

// --- Joints toriques (ISO 3601) ---
// Fenêtre : OR 8 × 1.5
OR_W_ID     = 8;
OR_W_CS     = 1.5;
GR_W_W      = 2.0;   // largeur gorge
GR_W_D      = 1.1;   // profondeur gorge

// Face de joint BP : OR 18 × 1.5 (étanche le circuit BP entre A et B)
OR_BP_ID    = D_FILET - 2;  // 18mm
OR_BP_CS    = 1.5;
GR_BP_W     = 2.0;
GR_BP_D     = 1.1;

// Buse : OR 3 × 1.0
OR_N_ID     = 3;
OR_N_CS     = 1.0;
GR_N_W      = 1.4;
GR_N_D      = 0.75;

// --- Laser guide rouge (SMA905 + optique de focalisation) ---
SMA_D_FERRULE  = 6.35;   // Ø ferrule SMA905
SMA_D_BODY     = 8.0;    // Ø corps connecteur
SMA_D_NUT      = 9.5;    // Ø écrou hexagonal
SMA_H_FERRULE  = 5;      // longueur ferrule
SMA_H_BODY     = 12;     // longueur corps connecteur
SMA_H_NUT      = 4;      // hauteur écrou
FOCUS_D        = 12;     // Ø tube optique de focalisation
FOCUS_H        = 25;     // longueur tube optique
FOCUS_D_LENS   = 8;      // Ø lentille interne
FIBER_D        = 3;      // Ø câble fibre optique
FIBER_L        = 40;     // longueur visible fibre

// --- Réservoir eau (30-100L PE) ---
TANK_L         = 400;    // longueur (mm)
TANK_W         = 300;    // largeur
TANK_H         = 420;    // hauteur (~50L)
TANK_WALL      = 4;      // épaisseur paroi
TANK_R         = 15;     // rayon congés

// --- Osmose inverse ---
RO_D           = 60;     // Ø membrane
RO_L           = 250;    // longueur housing
RO_D_CAP       = 50;     // Ø bouchons
RO_D_PORT      = 8;      // Ø raccords entrée/sortie

// --- Rendu ---
$fn = 64;

// ============ CALCULS ======================================
R_EXT       = D_EXT / 2;
R_FILET     = D_FILET / 2;

// ============ MODULES UTILITAIRES ==========================

module gorge_torique_face(id, cs, w, d, z) {
    // Gorge sur une face horizontale (pour joint entre 2 pièces)
    translate([0, 0, z])
        rotate_extrude()
            translate([id/2 + cs/2, 0])
                square([cs + 0.3, w], center=true);
}

module gorge_torique_axiale(id, cs, w, d, z) {
    // Gorge dans un alésage (autour de la fenêtre)
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

// (module spirale supprimé — remplacé par anneau simple)

// --- Filetage métrique (profil ISO 60°) ---

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

// --- Modules chanfrein 1mm ---
CHAMFER = 1;  // taille chanfrein (mm)

module chanfrein_haut(d, h) {
    // Chanfrein 45° sur l'arête haute extérieure d'un cylindre
    translate([0, 0, h - CHAMFER])
        difference() {
            cylinder(d=d + 1, h=CHAMFER + 0.5);
            translate([0, 0, -0.01])
                cylinder(d1=d, d2=d - 2*CHAMFER, h=CHAMFER + 0.01);
        }
}

module chanfrein_bas(d) {
    // Chanfrein 45° sur l'arête basse extérieure d'un cylindre
    translate([0, 0, -0.5])
        difference() {
            cylinder(d=d + 1, h=CHAMFER + 0.5);
            translate([0, 0, 0.5])
                cylinder(d1=d - 2*CHAMFER, d2=d, h=CHAMFER + 0.01);
            cylinder(d=d - 2*CHAMFER, h=CHAMFER + 0.51);
        }
}

module chanfrein_bas_z(d, z) {
    // Chanfrein 45° à une hauteur Z donnée
    translate([0, 0, z - 0.5])
        difference() {
            cylinder(d=d + 1, h=CHAMFER + 0.5);
            translate([0, 0, 0.5])
                cylinder(d1=d - 2*CHAMFER, d2=d, h=CHAMFER + 0.01);
            cylinder(d=d - 2*CHAMFER, h=CHAMFER + 0.51);
        }
}

module chanfrein_filet_male(d_filet, z_bas) {
    // Chanfrein sur le bout d'un filetage mâle (facilite l'engagement)
    // Taper 45° × 1mm au bout du filetage
    translate([0, 0, z_bas - 0.01])
        difference() {
            cylinder(d=d_filet + 0.1, h=CHAMFER + 0.01);
            cylinder(d1=d_filet - 2*CHAMFER, d2=d_filet + 0.1, h=CHAMFER + 0.02);
        }
}

module chanfrein_filet_femelle(d_filet, z_entree) {
    // Chanfrein (fraisure) à l'entrée d'un taraudage femelle
    // Cone 45° × 1mm qui élargit l'entrée du filetage
    translate([0, 0, z_entree - 0.01])
        cylinder(d1=d_filet, d2=d_filet + 2*CHAMFER, h=CHAMFER + 0.01);
}


// ============ LASER GUIDE ROUGE (SMA905) ===================

module sma905_connector() {
    // Écrou hexagonal
    color("Silver", 0.9)
    translate([0, 0, SMA_H_BODY])
        cylinder(d=SMA_D_NUT, h=SMA_H_NUT, $fn=6);

    // Corps connecteur
    color("DimGray", 0.9)
    cylinder(d=SMA_D_BODY, h=SMA_H_BODY);

    // Ferrule (dépasse en bas)
    color("Silver")
    translate([0, 0, -SMA_H_FERRULE])
        cylinder(d=SMA_D_FERRULE, h=SMA_H_FERRULE);

    // Câble fibre optique (sort par le haut)
    color("Orange", 0.8)
    translate([0, 0, SMA_H_BODY + SMA_H_NUT])
        cylinder(d=FIBER_D, h=FIBER_L);
}

module focus_tube() {
    color("DarkSlateGray", 0.9)
    difference() {
        cylinder(d=FOCUS_D, h=FOCUS_H);
        // Alésage interne
        translate([0, 0, -0.01])
            cylinder(d=FOCUS_D_LENS, h=FOCUS_H + 0.02);
    }
    // Lentille de focalisation (au milieu)
    color("LightBlue", 0.4)
    translate([0, 0, FOCUS_H/2 - 1])
        cylinder(d=FOCUS_D_LENS - 0.5, h=2);
}

module laser_guide() {
    // Tube optique de focalisation (se visse sur la platine A)
    focus_tube();

    // Connecteur SMA905 sur le dessus du tube
    translate([0, 0, FOCUS_H])
        sma905_connector();

    // Faisceau rouge (rayon visible, symbolique)
    color("Red", 0.3)
    translate([0, 0, -10])
        cylinder(d1=0.5, d2=3, h=10);
}


// ============ RÉSERVOIR EAU PE ==============================

module water_tank() {
    color("SteelBlue", 0.4)
    difference() {
        // Coque extérieure arrondie
        minkowski() {
            cube([TANK_L - 2*TANK_R, TANK_W - 2*TANK_R, TANK_H - 2*TANK_R]);
            sphere(r=TANK_R);
        }
        // Évidement intérieur
        translate([TANK_WALL, TANK_WALL, TANK_WALL])
            minkowski() {
                cube([TANK_L - 2*TANK_R - 2*TANK_WALL,
                      TANK_W - 2*TANK_R - 2*TANK_WALL,
                      TANK_H - 2*TANK_R - 2*TANK_WALL]);
                sphere(r=TANK_R);
            }
        // Ouverture bouchon (dessus)
        translate([(TANK_L - 2*TANK_R)/2, (TANK_W - 2*TANK_R)/2, TANK_H - TANK_R])
            cylinder(d=50, h=TANK_R + 1);
    }
    // Bouchon vissable
    color("DarkBlue", 0.7)
    translate([(TANK_L - 2*TANK_R)/2, (TANK_W - 2*TANK_R)/2, TANK_H - TANK_R])
        cylinder(d=55, h=8);

    // Raccord sortie (bas, face avant)
    color("Gray")
    translate([(TANK_L - 2*TANK_R)/2, -TANK_R, TANK_R + 20])
        rotate([-90, 0, 0])
            cylinder(d=12, h=15);

    // Eau visible
    color("CornflowerBlue", 0.3)
    translate([TANK_WALL + 2, TANK_WALL + 2, TANK_WALL + 2])
        cube([TANK_L - 2*TANK_R - 2*TANK_WALL - 4,
              TANK_W - 2*TANK_R - 2*TANK_WALL - 4,
              (TANK_H - 2*TANK_R) * 0.7]);
}


// ============ OSMOSE INVERSE ================================

module ro_unit() {
    // Housing membrane (tube principal)
    color("White", 0.8)
    difference() {
        union() {
            cylinder(d=RO_D, h=RO_L);
            // Bouchon entrée
            translate([0, 0, -5])
                cylinder(d=RO_D_CAP, h=8);
            // Bouchon sortie
            translate([0, 0, RO_L - 3])
                cylinder(d=RO_D_CAP, h=8);
        }
        // Alésage interne
        translate([0, 0, 2])
            cylinder(d=RO_D - 6, h=RO_L - 4);
    }

    // Raccord entrée eau brute (bas)
    color("Gray")
    translate([0, 0, -15])
        cylinder(d=RO_D_PORT, h=12);

    // Raccord sortie eau pure (haut)
    color("Gray")
    translate([0, 0, RO_L + 3])
        cylinder(d=RO_D_PORT, h=12);

    // Raccord rejet (latéral, en haut)
    color("Gray")
    translate([0, RO_D/2, RO_L - 30])
        rotate([-90, 0, 0])
            cylinder(d=RO_D_PORT, h=15);

    // Membrane interne (symbolique)
    color("Khaki", 0.5)
    translate([0, 0, 10])
        cylinder(d=RO_D - 8, h=RO_L - 20);
}


// ============ PIÈCE A — PLATINE OPTIQUE ====================

module piece_A() {
    color("Gold", 0.8)
    difference() {
        union() {
            // Corps principal (au-dessus de la zone filetée)
            translate([0, 0, H_FILET])
                cylinder(d=D_EXT, h=H_A - H_FILET);

            // Zone filetée M20 × 0.7
            filet_male(D_FILET, PAS_FILET, H_FILET);
        }

        // --- Passage faisceau (traversant) ---
        translate([0, 0, -1])
            cylinder(d=D_FAISCEAU, h=H_A + 2);

        // Entrée laser évasée (haut)
        translate([0, 0, H_A - 5])
            cylinder(d1=D_FAISCEAU, d2=D_FAISCEAU + 4, h=5.1);

        // --- Gorge joint fenêtre (face basse, autour du faisceau) ---
        gorge_torique_face(OR_W_ID, OR_W_CS, GR_W_W, GR_W_D, -GR_W_W/2);

        // --- Chambre annulaire refroidissement BP (face d'appui, au-dessus du filetage) ---
        // Fermée par face sup de B quand A est vissée
        translate([0, 0, H_FILET - 0.1])
            rotate_extrude($fn=64)
                translate([RAIN_R_INT, 0])
                    square([RAIN_W, RAIN_D + 0.1]);

        // --- Méplat + taraudage entrée eau BP (0°, G1/8" BSP) ---
        rotate([0, 0, 0]) {
            // Méplat (surface plate pour raccord coudé)
            translate([R_EXT - D_MEPLAT, -W_MEPLAT/2, H_FILET])
                cube([D_MEPLAT + 10, W_MEPLAT, H_MEPLAT]);
            // Taraudage G1/8" BSP (traverse la paroi jusqu'à la chambre)
            translate([0, 0, Z_RADIAL])
                rotate([0, 90, 0])
                    translate([0, 0, RAIN_R_EXT - 0.5])
                        cylinder(d=D_TARAUD, h=R_EXT + 1);
        }

        // --- Méplat + taraudage sortie eau BP (180°, G1/8" BSP) ---
        rotate([0, 0, 180]) {
            translate([R_EXT - D_MEPLAT, -W_MEPLAT/2, H_FILET])
                cube([D_MEPLAT + 10, W_MEPLAT, H_MEPLAT]);
            translate([0, 0, Z_RADIAL])
                rotate([0, 90, 0])
                    translate([0, 0, RAIN_R_EXT - 0.5])
                        cylinder(d=D_TARAUD, h=R_EXT + 1);
        }

        // --- Gorge joint BP face (étanche le circuit entre A et B) ---
        gorge_torique_face(OR_BP_ID, OR_BP_CS, GR_BP_W, GR_BP_D, H_FILET - GR_BP_W/2);

        // --- Vis de collimation ---
        // Rangée 1 : 3 trous M2.5 à 120°, à 3mm du haut
        for (i = [0:2]) {
            rotate([0, 0, i * 120])
                translate([0, 0, H_A - H_VIS_1])
                    rotate([0, 90, 0])
                        cylinder(d=D_VIS, h=D_EXT);
        }
        // Rangée 2 : 3 trous M2.5 à 120°, décalés de 60°, à 13mm du haut
        for (i = [0:2]) {
            rotate([0, 0, i * 120 + 60])
                translate([0, 0, H_A - H_VIS_2])
                    rotate([0, 90, 0])
                        cylinder(d=D_VIS, h=D_EXT);
        }

        // --- 2 trous pour clé à ergots (face supérieure) ---
        // Décalés à 90° et 270° pour éviter les vis de collimation (0°/60°/120°/180°/240°/300°)
        translate([0, R_EXT - 3, H_A - H_ERGOT])
            cylinder(d=D_ERGOT, h=H_ERGOT + 0.01);
        translate([0, -(R_EXT - 3), H_A - H_ERGOT])
            cylinder(d=D_ERGOT, h=H_ERGOT + 0.01);

        // --- Chanfreins 1mm sur arêtes extérieures ---
        chanfrein_haut(D_EXT, H_A);
        chanfrein_bas_z(D_EXT, H_FILET);

        // --- Chanfrein entrée filetage mâle M20 (bout du filetage) ---
        chanfrein_filet_male(D_FILET, 0);
    }
}


// ============ PIÈCE B — CHAMBRE HP =========================

module piece_B() {
    color("Peru", 0.8)
    difference() {
        // --- Corps plein ---
        cylinder(d=D_EXT, h=H_B);

        // --- Filetage interne HAUT M20 × 0.7 (reçoit pièce A) ---
        translate([0, 0, H_B - H_FILET - 0.01])
            filet_male(D_FILET, PAS_FILET, H_FILET + 0.02);

        // --- Logement fenêtre (épaulement, accessible par le haut) ---
        Z_FEN = H_B - H_FILET - EP_FENETRE;
        translate([0, 0, Z_FEN])
            cylinder(d=D_FENETRE + 0.2, h=EP_FENETRE + 0.1);

        // --- Gorge joint fenêtre (sous la fenêtre) ---
        gorge_torique_face(OR_W_ID, OR_W_CS, GR_W_W, GR_W_D,
                           Z_FEN - GR_W_W/2);

        // --- Passage faisceau (fenêtre → chambre) ---
        translate([0, 0, H_BUSE + H_CHAMBRE - 0.01])
            cylinder(d=D_FAISCEAU, h=Z_FEN - H_BUSE - H_CHAMBRE + 1);

        // --- Chambre HP ---
        translate([0, 0, H_BUSE])
            cylinder(d=D_CHAMBRE, h=H_CHAMBRE);

        // --- Méplat + taraudage entrée eau HP (G1/8" BSP pour raccord coudé) ---
        // Méplat sur le corps
        translate([R_EXT - D_MEPLAT, -W_MEPLAT/2, H_BUSE + H_CHAMBRE/2 - H_MEPLAT/2])
            cube([D_MEPLAT + 10, W_MEPLAT, H_MEPLAT]);
        // Taraudage G1/8" BSP (traverse la paroi jusqu'à la chambre HP)
        translate([0, 0, H_BUSE + H_CHAMBRE / 2])
            rotate([0, 90, 0])
                translate([0, 0, D_CHAMBRE/2 - 0.1])
                    cylinder(d=D_TARAUD, h=R_EXT + 1);

        // --- Logement buse + filetage interne BAS (reçoit pièce C) ---
        // Alésage traversant pour la buse (accessible par le bas quand C retiré)
        translate([0, 0, -0.01])
            cylinder(d=D_BUSE_EXT + 0.2, h=H_BUSE + 0.02);

        // Filetage interne M10 × 0.7 (reçoit bouchon C)
        translate([0, 0, -0.01])
            filet_male(D_FILET_C, PAS_FILET, H_FILET_C + 0.01);

        // --- Gorge joint buse (dans l'alésage, au-dessus du filetage) ---
        gorge_torique_face(OR_N_ID, OR_N_CS, GR_N_W, GR_N_D,
                           H_FILET_C + 0.5);

        // --- Chanfreins 1mm ---
        chanfrein_haut(D_EXT, H_B);
        chanfrein_bas(D_EXT);

        // --- Chanfreins entrées filetages femelles ---
        // M20 × 0.7 (haut, reçoit pièce A)
        chanfrein_filet_femelle(D_FILET + 0.2, H_B);
        // M10 × 0.7 (bas, reçoit pièce C)
        chanfrein_filet_femelle(D_FILET_C + 0.2, 0);
    }
}


// ============ PIÈCE C — BOUCHON BUSE =======================

// Paramètres clé à ergots
D_ERGOT     = 3.1;   // Ø trous pour ergots
ESP_ERGOT   = 14;    // espacement entre ergots (centre à centre)
H_ERGOT     = 3;     // profondeur des trous (= longueur ergot)

module piece_C() {
    color("Sienna", 0.9)
    difference() {
        union() {
            // Corps du bouchon (collerette extérieure)
            cylinder(d=D_EXT, h=H_C - H_FILET_C);

            // Zone filetée M10 × 0.7 (visse dans B par le bas)
            translate([0, 0, H_C - H_FILET_C])
                filet_male(D_FILET_C, PAS_FILET, H_FILET_C);
        }

        // --- Passage jet d'eau (traversant, axial) ---
        translate([0, 0, -0.01])
            cylinder(d=D_SORTIE, h=H_C + 0.02);

        // --- 2 trous pour clé à ergots (face inférieure) ---
        // À 3mm du bord extérieur, Ø3.1mm, profondeur 3mm
        translate([R_EXT - 3, 0, -0.01])
            cylinder(d=D_ERGOT, h=H_ERGOT + 0.01);
        translate([-(R_EXT - 3), 0, -0.01])
            cylinder(d=D_ERGOT, h=H_ERGOT + 0.01);

        // --- Chanfreins 1mm ---
        chanfrein_bas(D_EXT);

        // --- Chanfrein entrée filetage mâle M10 (bout du filetage) ---
        chanfrein_filet_male(D_FILET_C, H_C);
    }
}


// ============ BUSE SAPHIR ==================================

module buse_saphir() {
    color("LightBlue", 0.6)
    difference() {
        cylinder(d=D_BUSE_EXT, h=3);
        translate([0, 0, -0.01])
            cylinder(d=0.1, h=3.02);  // micro-perçage Ø100µm
    }
}


// ============ ASSEMBLAGE ===================================

module assemblage() {
    // Pièce C (bouchon buse, en bas)
    translate([0, 0, -(H_C - H_FILET_C)])
        piece_C();

    // Joint buse
    joint_torique(OR_N_ID, OR_N_CS, H_FILET_C + 0.5);

    // Buse saphir
    translate([0, 0, H_FILET_C + 1])
        buse_saphir();

    // Pièce B (corps principal)
    piece_B();

    // Fenêtre saphir
    Z_FEN = H_B - H_FILET - EP_FENETRE;
    translate([0, 0, Z_FEN])
        fenetre_saphir();

    // Joints fenêtre (haut et bas)
    joint_torique(OR_W_ID, OR_W_CS, Z_FEN - 1);
    joint_torique(OR_W_ID, OR_W_CS, Z_FEN + EP_FENETRE + 0.5);

    // Joint BP face (sur la face d'appui A/B)
    joint_torique(OR_BP_ID, OR_BP_CS, H_B - 0.5);

    // Pièce A (vissée dans B par le haut)
    translate([0, 0, H_B - H_FILET])
        piece_A();

    // Laser guide rouge (au-dessus de pièce A, sur l'axe optique)
    translate([0, 0, H_B - H_FILET + H_A + 2])
        laser_guide();

    // Réservoir eau PE (décalé à droite)
    translate([120, -TANK_W/2 + TANK_R, -(H_C)])
        water_tank();

    // Osmose inverse (entre réservoir et tête, couchée)
    translate([60, 0, TANK_H/2 - H_C])
        rotate([0, 0, -90])
            ro_unit();
}

module vue_eclatee() {
    ECART = 15;

    // Pièce C (tout en bas)
    translate([0, 0, -3*ECART])
        piece_C();

    // Joint buse
    translate([0, 0, -2*ECART])
        joint_torique(OR_N_ID, OR_N_CS, 0);

    // Buse saphir
    translate([0, 0, -ECART])
        buse_saphir();

    // Pièce B
    piece_B();

    // Joint fenêtre bas
    translate([0, 0, H_B + ECART - 2])
        joint_torique(OR_W_ID, OR_W_CS, 0);

    // Fenêtre saphir
    translate([0, 0, H_B + ECART])
        fenetre_saphir();

    // Joint fenêtre haut
    translate([0, 0, H_B + ECART + EP_FENETRE + 2])
        joint_torique(OR_W_ID, OR_W_CS, 0);

    // Joint BP face
    translate([0, 0, H_B + 2*ECART])
        joint_torique(OR_BP_ID, OR_BP_CS, 0);

    // Pièce A (haut)
    translate([0, 0, H_B + 2*ECART + 3])
        piece_A();
}


module demi_coupe() {
    translate([0, -D_EXT - 1, -80])
        cube([D_EXT + 5, D_EXT * 2 + 2, 250]);
}

module vue_eclatee_coupe() {
    ECART = 15;

    // Pièce C — bouchon buse (terre de Sienne)
    color("Sienna", 0.9)
    translate([0, 0, -3*ECART])
    difference() {
        piece_C();
        demi_coupe();
    }

    // Joint buse
    translate([0, 0, -2*ECART])
        joint_torique(OR_N_ID, OR_N_CS, 0);

    // Buse saphir (cyan clair)
    color("LightCyan", 0.7)
    translate([0, 0, -ECART])
    difference() {
        buse_saphir();
        demi_coupe();
    }

    // Pièce B — chambre HP (brun foncé)
    color("SaddleBrown", 0.9)
    difference() {
        piece_B();
        demi_coupe();
    }

    // Joint fenêtre bas
    translate([0, 0, H_B + ECART - 2])
        joint_torique(OR_W_ID, OR_W_CS, 0);

    // Fenêtre saphir (bleu)
    color("CornflowerBlue", 0.7)
    translate([0, 0, H_B + ECART])
    difference() {
        fenetre_saphir();
        demi_coupe();
    }

    // Joint fenêtre haut
    translate([0, 0, H_B + ECART + EP_FENETRE + 2])
        joint_torique(OR_W_ID, OR_W_CS, 0);

    // Joint BP face
    translate([0, 0, H_B + 2*ECART])
        joint_torique(OR_BP_ID, OR_BP_CS, 0);

    // Pièce A — platine optique (doré)
    color("Goldenrod", 0.9)
    translate([0, 0, H_B + 2*ECART + 3])
    difference() {
        piece_A();
        demi_coupe();
    }

    // Laser guide rouge (au-dessus de pièce A)
    translate([0, 0, H_B + 2*ECART + 3 + H_A + ECART])
        laser_guide();

    // Réservoir eau PE (décalé à droite)
    translate([120, -TANK_W/2 + TANK_R, -(H_C)])
        water_tank();

    // Osmose inverse (entre réservoir et tête)
    translate([60, 0, TANK_H/2 - H_C])
        rotate([0, 0, -90])
            ro_unit();
}


// ============ AFFICHAGE ====================================
// Décommenter UNE des lignes :
// (Gardé par EXPORT_PART pour éviter le rendu lors d'un include)

if (is_undef(EXPORT_PART)) {

// Vue assemblée demi-coupe
// difference() {
//     assemblage();
//     translate([0, -D_EXT, -H_C - 2])
//         cube([D_EXT, D_EXT*2, H_A + H_B + H_C + 4]);
// }

// Vue assemblée complète
// assemblage();

// Vue éclatée
// vue_eclatee();

// Vue éclatée demi-coupe
// vue_eclatee_coupe();

// Vue assemblée demi-coupe
// difference() {
//     assemblage();
//     translate([0, -D_EXT, -H_C - 2])
//         cube([D_EXT, D_EXT*2, H_A + H_B + H_C + 4]);
// }

// Pièce A seule (retournée, face rainures visible)
// rotate([180, 0, 0]) piece_A();

// Pièce B seule
// piece_B();

// Pièce C seule (bouchon buse)
piece_C();

} // fin garde EXPORT_PART
