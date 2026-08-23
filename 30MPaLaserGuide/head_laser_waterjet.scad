// ============================================================
// TÊTE LASER GUIDÉE PAR JET D'EAU — v2
// Rainures de refroidissement BP ouvertes sur face de joint
// ============================================================

// ============ PARAMÈTRES MODIFIABLES =======================

// --- Corps ---
D_EXT       = 30;    // Ø extérieur corps
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
RAIN_D      = 4.0;   // profondeur chambre annulaire (< H_FILET pour raccord solide)
RAIN_R_INT  = D_FENETRE/2 + 1.5;   // rayon intérieur anneau (6.5mm)
RAIN_R_EXT  = D_EXT/2 - 1.5;       // rayon extérieur anneau (13.5mm, paroi 1.5mm)
RAIN_W      = RAIN_R_EXT - RAIN_R_INT;  // largeur anneau (7mm)
RAIN_R      = (RAIN_R_INT + RAIN_R_EXT) / 2;  // rayon centre (10mm)
D_CANAL_BP  = 3;     // Ø canal radial (anneau → raccord)
D_RACCORD   = 6.2;   // Ø alésage pour tuyau plastique OD 6mm (press-fit)
L_RACCORD   = 8;     // profondeur d'insertion tuyau
Z_RADIAL    = H_FILET + RAIN_D / 2;  // hauteur abs. trous radiaux = milieu chambre dans le corps

// --- Entrée eau HP (raccord HPLC 10-32 UNF) ---
D_ENTREE_HP = 3.5;   // Ø avant-trou taraudage 10-32 UNF (taraudé à la main)
H_TARAUD_HP = 6;     // profondeur taraudage

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

        // --- Raccord entrée eau BP (0°, radial au milieu de la chambre) ---
        // Alésage Ø6.2 pour tuyau plastique OD 6mm (press-fit)
        rotate([0, 0, 0])
            translate([0, 0, Z_RADIAL])
                rotate([0, 90, 0])
                    union() {
                        // Alésage tuyau (depuis l'extérieur)
                        translate([0, 0, R_EXT - L_RACCORD])
                            cylinder(d=D_RACCORD, h=L_RACCORD + 1);
                        // Canal étroit vers la chambre annulaire
                        translate([0, 0, RAIN_R_EXT - 0.1])
                            cylinder(d=D_CANAL_BP, h=R_EXT - RAIN_R_EXT - L_RACCORD + 1);
                    }

        // --- Raccord sortie eau BP (180°, radial au milieu de la chambre) ---
        rotate([0, 0, 180])
            translate([0, 0, Z_RADIAL])
                rotate([0, 90, 0])
                    union() {
                        translate([0, 0, R_EXT - L_RACCORD])
                            cylinder(d=D_RACCORD, h=L_RACCORD + 1);
                        translate([0, 0, RAIN_R_EXT - 0.1])
                            cylinder(d=D_CANAL_BP, h=R_EXT - RAIN_R_EXT - L_RACCORD + 1);
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
        union() {
            // --- Corps plein ---
            cylinder(d=D_EXT, h=H_B);

            // --- Crêtes filetage M20 (solidaires du corps) ---
            translate([0, 0, H_B - H_FILET - 0.01])
                filet_femelle_ridges(D_FILET, PAS_FILET, H_FILET + 0.02);

            // --- Crêtes filetage M10 (solidaires du corps) ---
            translate([0, 0, -0.01])
                filet_femelle_ridges(D_FILET_C, PAS_FILET, H_FILET_C + 0.01);
        }

        // --- Alésage filetage HAUT M20 × 0.7 (reçoit pièce A) ---
        translate([0, 0, H_B - H_FILET - 0.01])
            filet_femelle(D_FILET, PAS_FILET, H_FILET + 0.02);

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

        // --- Entrée eau HP (latérale, taraudage 10-32 UNF pour raccord HPLC) ---
        // Avant-trou Ø3.5 borgne (extérieur → chambre) + taraudage 6mm
        translate([0, 0, H_BUSE + H_CHAMBRE / 2])
            rotate([0, 90, 0]) {
                // Avant-trou borgne jusqu'à la chambre
                translate([0, 0, D_CHAMBRE/2 - 0.1])
                    cylinder(d=D_ENTREE_HP, h=R_EXT - D_CHAMBRE/2 + 1);
                // Alésage taraudage côté extérieur
                translate([0, 0, R_EXT - H_TARAUD_HP])
                    cylinder(d=D_ENTREE_HP + 0.6, h=H_TARAUD_HP + 0.1);
            }

        // --- Logement buse + filetage interne BAS (reçoit pièce C) ---
        // Alésage traversant pour la buse (accessible par le bas quand C retiré)
        translate([0, 0, -0.01])
            cylinder(d=D_BUSE_EXT + 0.2, h=H_BUSE + 0.02);

        // Filetage interne M10 × 0.7 (reçoit bouchon C)
        translate([0, 0, -0.01])
            filet_femelle(D_FILET_C, PAS_FILET, H_FILET_C + 0.01);

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
}


// ============ AFFICHAGE ====================================
// Décommenter UNE des lignes :

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
vue_eclatee_coupe();

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
// piece_C();
