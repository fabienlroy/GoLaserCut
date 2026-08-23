// ============================================================
// TÊTE LASER GUIDÉE PAR JET D'EAU — 2 PIÈCES
// Brevet Synova expiré — projet open source
// github.com/fabienlroy/GoLaserCut
// ============================================================
// Pièce A (haut) : platine optique + alignement + refroidissement BP
// Pièce B (bas)  : chambre HP 300 bars + buse saphir
// Jonction : filetage M × 0.7 de pas, fenêtre saphir prise en sandwich
// ============================================================

// ============ PARAMÈTRES MODIFIABLES =======================

// --- Corps ---
D_EXT       = 30;    // Ø extérieur du corps (à adapter à la source laser)
H_A         = 28;    // hauteur pièce A
H_B         = 25;    // hauteur pièce B

// --- Filetage jonction (M × 0.7 de pas) ---
D_FILET     = 20;    // Ø nominal filetage (M20 × 0.7)
PAS_FILET   = 0.7;   // pas du filet (pour la fraise mono-dent)
H_FILET     = 8;     // longueur filetée

// --- Fenêtre saphir ---
D_FENETRE   = 10;    // Ø fenêtre saphir
EP_FENETRE  = 1.5;   // épaisseur fenêtre saphir
D_FAISCEAU  = 6;     // Ø passage libre pour le faisceau

// --- Chambre HP ---
D_CHAMBRE   = 8;     // Ø intérieur chambre HP
H_CHAMBRE   = 10;    // hauteur chambre HP

// --- Buse ---
D_BUSE_EXT  = 4;     // Ø logement buse saphir (extérieur buse)
H_BUSE      = 5;     // profondeur logement buse

// --- Canal refroidissement BP ---
D_CANAL_EXT = D_FILET - 2;  // Ø extérieur canal BP (18mm)
D_CANAL_INT = D_FENETRE + 3; // Ø intérieur canal BP (13mm)
H_CANAL     = 8;     // hauteur canal
D_ENTREE_BP = 4;     // Ø entrée/sortie eau BP (latéral)

// --- Entrée eau HP ---
D_ENTREE_HP = 1.6;   // Ø 1/16" = raccord HPLC 10-32

// --- Vis d'alignement ---
N_VIS       = 6;     // nombre de vis (3 push + 3 ressorts)
D_VIS       = 2.5;   // Ø perçage pour vis M2.5
D_CERCLE_VIS = D_EXT - 6; // Ø cercle de répartition

// --- Joints toriques (ISO 3601) ---
// Joint fenêtre : OR 8 × 1.5 (ID=8mm, CS=1.5mm)
OR_W_ID     = 8;     // Ø intérieur joint fenêtre
OR_W_CS     = 1.5;   // section joint fenêtre
GR_W_W      = 2.0;   // largeur gorge (CS × 1.33)
GR_W_D      = 1.1;   // profondeur gorge (CS × 0.75)

// Joint buse : OR 3 × 1.0 (ID=3mm, CS=1.0mm)
OR_N_ID     = 3;     // Ø intérieur joint buse
OR_N_CS     = 1.0;   // section joint buse
GR_N_W      = 1.4;   // largeur gorge
GR_N_D      = 0.75;  // profondeur gorge

// --- Rendu ---
$fn = 64;            // résolution des cylindres

// ============ CALCULS DERIVES ==============================
EP_PAROI    = (D_EXT - D_FILET) / 2;  // épaisseur paroi au filetage
R_EXT       = D_EXT / 2;
R_FILET     = D_FILET / 2;

// ============ PIÈCE A — PLATINE OPTIQUE ====================

module piece_A() {
    color("Gold", 0.8)
    difference() {
        // --- Corps plein ---
        cylinder(d=D_EXT, h=H_A);

        // --- Passage faisceau (traversant, axial) ---
        // Évasé en haut (entrée laser), rétréci en bas (vers fenêtre)
        translate([0, 0, -1])
            cylinder(d=D_FAISCEAU, h=H_A + 2);

        // Entrée laser évasée (cône, haut de la pièce)
        translate([0, 0, H_A - 5])
            cylinder(d1=D_FAISCEAU, d2=D_FAISCEAU + 4, h=5.1);

        // --- Logement fenêtre (bas de pièce A) ---
        // Épaulement pour poser la fenêtre
        translate([0, 0, -0.01])
            cylinder(d=D_FENETRE + 0.2, h=EP_FENETRE + GR_W_D + 0.5);

        // --- Gorge joint torique fenêtre (face inférieure pièce A) ---
        // Le joint est pressé contre le dessus de la fenêtre
        translate([0, 0, EP_FENETRE])
            gorge_torique(OR_W_ID, OR_W_CS, GR_W_W, GR_W_D);

        // --- Canal refroidissement BP (annulaire) ---
        translate([0, 0, EP_FENETRE + GR_W_W + 1])
            difference() {
                cylinder(d=D_CANAL_EXT, h=H_CANAL);
                translate([0, 0, -0.5])
                    cylinder(d=D_CANAL_INT, h=H_CANAL + 1);
            }

        // Entrée eau BP (latérale, 2 trous à 180°)
        H_BP_MID = EP_FENETRE + GR_W_W + 1 + H_CANAL / 2;
        rotate([0, 0, 0])
            translate([0, 0, H_BP_MID])
                rotate([0, 90, 0])
                    cylinder(d=D_ENTREE_BP, h=D_EXT, center=true);

        rotate([0, 0, 90])
            translate([0, 0, H_BP_MID])
                rotate([0, 90, 0])
                    cylinder(d=D_ENTREE_BP, h=D_EXT, center=true);

        // --- Filetage externe (zone basse) ---
        // Modélisé comme un cylindre au Ø nominal
        // Le filetage réel sera coupé par la fraise mono-dent sur la CNC
        // Représenté ici par une réduction de diamètre
        translate([0, 0, -0.01])
            difference() {
                cylinder(d=D_EXT + 1, h=H_FILET);
                cylinder(d=D_FILET, h=H_FILET + 1);
            }

        // --- 6 vis d'alignement (radiales) ---
        for (i = [0:N_VIS-1]) {
            angle = i * 360 / N_VIS;
            H_VIS_POS = H_A - 8; // position en hauteur des vis
            rotate([0, 0, angle])
                translate([0, 0, H_VIS_POS])
                    rotate([0, 90, 0])
                        cylinder(d=D_VIS, h=D_EXT, center=false);
        }
    }
}


// ============ PIÈCE B — CHAMBRE HP =========================

module piece_B() {
    color("Peru", 0.8)
    difference() {
        // --- Corps plein ---
        cylinder(d=D_EXT, h=H_B);

        // --- Filetage interne (zone haute) ---
        // Taraudage M20 × 0.7 (réalisé par fraise mono-dent)
        translate([0, 0, H_B - H_FILET - 0.01])
            cylinder(d=D_FILET + 0.2, h=H_FILET + 0.02);

        // --- Logement fenêtre (haut de la chambre) ---
        // Épaulement : la fenêtre repose ici
        translate([0, 0, H_B - H_FILET - EP_FENETRE - GR_W_D])
            cylinder(d=D_FENETRE + 0.2, h=EP_FENETRE + GR_W_D + 0.1);

        // --- Gorge joint torique fenêtre (face supérieure pièce B) ---
        // Le joint est sous la fenêtre
        translate([0, 0, H_B - H_FILET - GR_W_D - 0.5])
            gorge_torique(OR_W_ID, OR_W_CS, GR_W_W, GR_W_D);

        // --- Chambre HP ---
        translate([0, 0, H_BUSE])
            cylinder(d=D_CHAMBRE, h=H_CHAMBRE);

        // --- Passage faisceau (entre fenêtre et chambre) ---
        translate([0, 0, H_BUSE + H_CHAMBRE - 0.01])
            cylinder(d=D_FAISCEAU, h=H_B - H_FILET - EP_FENETRE - H_CHAMBRE - H_BUSE + 1);

        // --- Entrée eau HP (latérale, raccord HPLC 10-32) ---
        translate([0, 0, H_BUSE + H_CHAMBRE / 2])
            rotate([0, 90, 0])
                cylinder(d=D_ENTREE_HP, h=D_EXT, center=true);

        // --- Logement buse saphir (bas) ---
        translate([0, 0, -0.01])
            cylinder(d=D_BUSE_EXT, h=H_BUSE + 0.01);

        // --- Gorge joint torique buse ---
        translate([0, 0, H_BUSE - GR_N_W - 0.5])
            gorge_torique(OR_N_ID, OR_N_CS, GR_N_W, GR_N_D);

        // --- Sortie buse (traversant) ---
        translate([0, 0, -0.01])
            cylinder(d=0.5, h=H_B + 0.02); // Ø0.5 symbolique (buse = Ø0.1)
    }
}


// ============ MODULES UTILITAIRES ==========================

module gorge_torique(id, cs, w, d) {
    // Gorge annulaire pour joint torique
    // Tore de section rectangulaire (approx)
    rotate_extrude()
        translate([id/2 + cs/2, 0, 0])
            square([cs + 0.3, w], center=true);
}

module fenetre_saphir() {
    color("LightBlue", 0.5)
    cylinder(d=D_FENETRE, h=EP_FENETRE);
}

module joint_torique(id, cs) {
    color("Black")
    rotate_extrude()
        translate([id/2 + cs/2, 0, 0])
            circle(d=cs);
}


// ============ ASSEMBLAGE ===================================

module assemblage() {
    // Pièce B en bas (chambre HP)
    piece_B();

    // Fenêtre saphir (entre les deux pièces)
    translate([0, 0, H_B - H_FILET - EP_FENETRE])
        fenetre_saphir();

    // Joint torique fenêtre bas
    translate([0, 0, H_B - H_FILET - EP_FENETRE - 0.5])
        joint_torique(OR_W_ID, OR_W_CS);

    // Joint torique fenêtre haut
    translate([0, 0, H_B - H_FILET + 0.5])
        joint_torique(OR_W_ID, OR_W_CS);

    // Pièce A en haut (platine optique)
    translate([0, 0, H_B - H_FILET])
        piece_A();

    // Joint torique buse
    translate([0, 0, H_BUSE - GR_N_W])
        joint_torique(OR_N_ID, OR_N_CS);
}

module vue_eclatee() {
    // Vue éclatée pour visualisation
    ECART = 10;

    // Pièce B (bas)
    piece_B();

    // Fenêtre + joints
    translate([0, 0, H_B + ECART])
        fenetre_saphir();
    translate([0, 0, H_B + ECART - 2])
        joint_torique(OR_W_ID, OR_W_CS);
    translate([0, 0, H_B + ECART + EP_FENETRE + 2])
        joint_torique(OR_W_ID, OR_W_CS);

    // Pièce A (haut)
    translate([0, 0, H_B + 2*ECART + EP_FENETRE])
        piece_A();

    // Joint buse
    translate([0, 0, -ECART])
        joint_torique(OR_N_ID, OR_N_CS);
}


// ============ AFFICHAGE ====================================
// Décommenter UNE des lignes suivantes :

// Vue assemblée (coupe)
// difference() {
//     assemblage();
//     // Demi-coupe pour voir l'intérieur
//     translate([0, -D_EXT, -1])
//         cube([D_EXT, D_EXT*2, H_A + H_B + 2]);
// }


// Vue assemblée complète
// assemblage();

// Vue éclatée
// vue_eclatee();

// Pièce A seule (pour export STL/usinage)
piece_A();

// Pièce B seule (pour export STL/usinage)
// piece_B();
