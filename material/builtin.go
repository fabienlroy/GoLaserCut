package material

// BuiltinMaterialLibrary returns the default material library.
// Default feed rates are starting estimates for a ~10W diode laser.
func BuiltinMaterialLibrary() *MaterialLibrary {
	return &MaterialLibrary{
		Materials: []Material{
			// Hardwoods
			{Name: "Mahogany", Category: "hardwood", MinThickness: 3, MaxThickness: 12, DefaultFeed: 800,
				Notes: "Dense tropical hardwood, clean dark edges"},
			{Name: "Walnut", Category: "hardwood", MinThickness: 3, MaxThickness: 10, DefaultFeed: 800,
				Notes: "Hard, dark wood, good contrast"},
			{Name: "Cherry", Category: "hardwood", MinThickness: 3, MaxThickness: 10, DefaultFeed: 900,
				Notes: "Medium-hard, nice reddish tone"},
			{Name: "Maple", Category: "hardwood", MinThickness: 3, MaxThickness: 10, DefaultFeed: 800,
				Notes: "Very hard, light color, high contrast engravings"},
			{Name: "Oak", Category: "hardwood", MinThickness: 3, MaxThickness: 10, DefaultFeed: 700,
				Notes: "Very hard and dense, slow cutting"},
			{Name: "Beech", Category: "hardwood", MinThickness: 3, MaxThickness: 10, DefaultFeed: 750,
				Notes: "Hard, uniform grain"},
			{Name: "Ash", Category: "hardwood", MinThickness: 3, MaxThickness: 10, DefaultFeed: 800,
				Notes: "Hard, prominent grain pattern"},

			// Softwoods
			{Name: "Pine", Category: "softwood", MinThickness: 3, MaxThickness: 20, DefaultFeed: 1500,
				Notes: "Soft, resinous, uneven burning due to grain"},
			{Name: "Cedar", Category: "softwood", MinThickness: 3, MaxThickness: 20, DefaultFeed: 1400,
				Notes: "Soft, aromatic, burns evenly"},
			{Name: "Poplar", Category: "softwood", MinThickness: 3, MaxThickness: 15, DefaultFeed: 1200,
				Notes: "Soft, light color, clean edges"},
			{Name: "Balsa", Category: "softwood", MinThickness: 1, MaxThickness: 10, DefaultFeed: 2000,
				Notes: "Extremely soft, very fast cutting, fire risk"},
			{Name: "Basswood", Category: "softwood", MinThickness: 1, MaxThickness: 10, DefaultFeed: 1800,
				Notes: "Very soft, excellent for detailed engraving"},

			// Plywood and composites
			{Name: "Birch Plywood", Category: "plywood", MinThickness: 3, MaxThickness: 6, DefaultFeed: 1000,
				Notes: "Clean cuts, consistent quality, glue layers may char"},
			{Name: "Poplar Plywood", Category: "plywood", MinThickness: 3, MaxThickness: 6, DefaultFeed: 1100,
				Notes: "Softer than birch, lighter weight"},
			{Name: "MDF", Category: "composite", MinThickness: 3, MaxThickness: 6, DefaultFeed: 1200,
				Notes: "Uniform density, dark edges, releases formaldehyde — ventilate"},
			{Name: "HDF", Category: "composite", MinThickness: 3, MaxThickness: 4, DefaultFeed: 1000,
				Notes: "Denser than MDF, cleaner edges"},
			{Name: "Bamboo", Category: "composite", MinThickness: 3, MaxThickness: 8, DefaultFeed: 900,
				Notes: "Hard, layered structure"},

			// Acrylic (CO₂ laser primarily)
			{Name: "Cast Acrylic", Category: "acrylic", MinThickness: 2, MaxThickness: 10, DefaultFeed: 500,
				Notes: "Best for cutting, flame-polished edges, CO₂ laser recommended"},
			{Name: "Extruded Acrylic", Category: "acrylic", MinThickness: 2, MaxThickness: 6, DefaultFeed: 600,
				Notes: "Cheaper than cast, edges less polished"},

			// Leather
			{Name: "Vegetable-Tanned Leather", Category: "leather", MinThickness: 1, MaxThickness: 4, DefaultFeed: 1500,
				Notes: "Safe to cut, natural finish"},
			{Name: "Chrome-Tanned Leather", Category: "leather", MinThickness: 1, MaxThickness: 3, DefaultFeed: 1500,
				Notes: "WARNING: may release toxic chromium compounds"},
			{Name: "Suede", Category: "leather", MinThickness: 0.5, MaxThickness: 2, DefaultFeed: 2000,
				Notes: "Thin, cuts easily"},

			// Paper and cardboard
			{Name: "Cardboard", Category: "paper", MinThickness: 1, MaxThickness: 5, DefaultFeed: 3000,
				Notes: "Cuts fast, fire risk, keep air assist on"},
			{Name: "Kraft Paper", Category: "paper", MinThickness: 0.1, MaxThickness: 1, DefaultFeed: 4000,
				Notes: "Very fast, minimal power needed"},
			{Name: "Mat Board", Category: "paper", MinThickness: 1, MaxThickness: 3, DefaultFeed: 2500,
				Notes: "Museum/framing board, clean cuts"},

			// Fabric and soft materials
			{Name: "Cotton Fabric", Category: "fabric", MinThickness: 0.2, MaxThickness: 2, DefaultFeed: 3000,
				Notes: "Clean sealed edges, fire risk"},
			{Name: "Felt", Category: "fabric", MinThickness: 1, MaxThickness: 5, DefaultFeed: 2500,
				Notes: "Polyester felt cuts clean, wool felt chars"},
			{Name: "Cork", Category: "natural", MinThickness: 1, MaxThickness: 6, DefaultFeed: 2000,
				Notes: "Engraves and cuts well, nice contrast"},
			{Name: "Foam EVA", Category: "foam", MinThickness: 2, MaxThickness: 10, DefaultFeed: 2000,
				Notes: "Cosplay foam, melted edges, ventilate"},

			// Engineering plastics (CO₂ laser)
			{Name: "Delrin (Acetal)", Category: "plastic", MinThickness: 1, MaxThickness: 6, DefaultFeed: 600,
				Notes: "Cuts clean, no flame, CO₂ laser recommended"},
			{Name: "Mylar (PET)", Category: "plastic", MinThickness: 0.1, MaxThickness: 1, DefaultFeed: 3000,
				Notes: "Thin film, stencil use"},

			// Metals (fiber laser / CO₂ with O₂ assist)
			{Name: "Mild Steel", Category: "metal", MinThickness: 0.3, MaxThickness: 3, DefaultFeed: 300,
				Notes: "Requires fiber laser or high-power CO₂ with O₂ assist"},
			{Name: "Stainless Steel", Category: "metal", MinThickness: 0.3, MaxThickness: 2, DefaultFeed: 250,
				Notes: "Fiber laser, N₂ or O₂ assist for clean edges"},
			{Name: "Aluminum", Category: "metal", MinThickness: 0.3, MaxThickness: 3, DefaultFeed: 400,
				Notes: "Fiber laser only, reflective — be careful"},

			// Stone and ceramic (engraving only)
			{Name: "Slate", Category: "stone", MinThickness: 3, MaxThickness: 10, DefaultFeed: 500,
				Notes: "Engraving only, cannot cut, white contrast on dark surface"},
			{Name: "Ceramic Tile", Category: "stone", MinThickness: 5, MaxThickness: 10, DefaultFeed: 600,
				Notes: "Engraving only with marking compound"},
			{Name: "Anodized Aluminum", Category: "metal", MinThickness: 0.5, MaxThickness: 3, DefaultFeed: 2000,
				Notes: "Diode laser can remove anodizing layer for marking"},
		},
	}
}
