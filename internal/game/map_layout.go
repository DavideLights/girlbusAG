package game

// PropKind classifies static and interactive Pop-Urban city props in the arena.
type PropKind int

const (
	PropBuildingKonbini PropKind = iota
	PropBuildingCoffee
	PropBuildingArcade
	PropBuildingRamen
	PropBuildingBookstore
	PropBuildingMetro
	PropBuildingNewsstand
	PropTreeLush
	PropStreetlamp
	PropBenchPlants
	PropCatsClutter
	PropScooter
	PropBikeRack
	PropHydrantBin
	PropZTLSign
	PropCanalBoat
	PropSlotMachine
	PropManhole
	PropAsphaltCrack
	PropFlowerBox
	PropNeonBillboard
	PropCarRed
	PropCarYellow
	PropCarMint
)

// MapProp represents a placed architectural, foliage, or street-furniture prop in the city district.
type MapProp struct {
	Kind        PropKind
	X, Y        float32 // Top-left world position
	W, H        float32
	SortY       float32 // Painter's Algorithm Y-sort baseline
	Solid       bool
	CollideRect Rect // Soft movement collision & bullet-blocking cover footprint
	CastShadow  bool
}

// LightKind classifies ambient and neon light emitters across the district.
type LightKind int

const (
	LightStreetlamp LightKind = iota
	LightShopWindow
	LightNeonSign
	LightVending
	LightSlotMachine
)

// LightSource defines a colored world-space light emitter for the evening/night lighting pass.
type LightSource struct {
	X, Y, Radius float32
	Kind         LightKind
	R, G, B      uint8
	Intensity    float32
}

// AimFeedback summarizes tactical placement information at Alfredo's clamped throw target.
type AimFeedback struct {
	ClampedX, ClampedY float32
	InRange            bool
	IsPerfectDrop      bool
	SynergyCount       int
	HookCandidateCount int
	VIPCount           int
	InZTL              bool
	InBusBay           bool
	PreviewRadius      float32
}

// DistrictLayout holds the static stratified topology generated for the 1280x720 city district.
type DistrictLayout struct {
	Props         []MapProp
	Lights        []LightSource
	ParkIslands   []Rect
	BusBays       []Rect
	ZTLZones      []Rect
	Shops         []VendingShop
	SlotMachines  []SlotMachine
	CrosswalkCols []int
	CrosswalkRows []int
	BoardwalkY    float32
	CanalY        float32
}

// BuildDistrictLayout constructs the rich stratified Pop-Urban city district for the 1280x720 arena.
func BuildDistrictLayout() DistrictLayout {
	layout := DistrictLayout{
		BoardwalkY:    624.0,
		CanalY:        680.0,
		CrosswalkCols: []int{16, 23},
		CrosswalkRows: []int{5, 16},
		ParkIslands: []Rect{
			{X: 256, Y: 288, W: 160, H: 112}, // West Pocket-Park Island
			{X: 864, Y: 288, W: 160, H: 112}, // East Pocket-Park Island
		},
		BusBays: []Rect{
			{X: 160, Y: 150, W: 88, H: 36},
			{X: 440, Y: 150, W: 88, H: 36},
			{X: 780, Y: 500, W: 88, H: 36},
			{X: 1060, Y: 500, W: 88, H: 36},
		},
		// 2 Forbidden ZTL Zones flanking the central intersection (640, 352) along the cross-street.
		ZTLZones: []Rect{
			{X: 568, Y: 196, W: 144, H: 80},
			{X: 568, Y: 432, W: 144, H: 80},
		},
		// 3 Konbini Neon Vending Machines on the North & South Sidewalk Promenades.
		Shops: []VendingShop{
			{X: 220, Y: 92, Radius: 54, HealCost: 50, BribeCost: 100},
			{X: 500, Y: 92, Radius: 54, HealCost: 50, BribeCost: 100},
			{X: 1040, Y: 592, Radius: 54, HealCost: 50, BribeCost: 100},
		},
		// 2 Rizz-Slots 777 Cabinets: North Arcade & South Waterfront Boardwalk.
		SlotMachines: []SlotMachine{
			{X: 386, Y: 94, Radius: 54, SpinCost: 35},
			{X: 740, Y: 584, Radius: 54, SpinCost: 35},
		},
	}

	props := make([]MapProp, 0, 84)
	lights := make([]LightSource, 0, 56)

	// --- Layer 1: Varied North Storefront Boulevard (Y = 0..144) ---
	type storefrontSpec struct {
		x    float32
		kind PropKind
	}
	storefronts := []storefrontSpec{
		// Block 1 (X = 32..240), alley 240..320
		{x: 32, kind: PropBuildingKonbini},
		{x: 144, kind: PropBuildingRamen},
		// Block 2 (X = 320..528), Central Cross-Street North Avenue 528..752
		{x: 320, kind: PropBuildingArcade},
		{x: 432, kind: PropBuildingBookstore},
		// Block 3 (X = 752..960), East alley 960..1024
		{x: 752, kind: PropBuildingMetro},
		{x: 864, kind: PropBuildingCoffee},
		// Block 4 (X = 1024..1232)
		{x: 1024, kind: PropBuildingKonbini},
		{x: 1136, kind: PropBuildingRamen},
	}
	for _, b := range storefronts {
		props = append(props, MapProp{
			Kind:        b.kind,
			X:           b.x,
			Y:           8,
			W:           96,
			H:           64,
			SortY:       72,
			Solid:       true,
			CollideRect: Rect{X: b.x, Y: 12, W: 96, H: 60},
			CastShadow:  true,
		})
		// Warm interior shop window glow
		lights = append(lights, LightSource{
			X: b.x + 48, Y: 58, Radius: 96, Kind: LightShopWindow,
			R: 255, G: 222, B: 165, Intensity: 0.65,
		})
		// Colored neon storefront sign light
		var nr, ng, nb uint8
		switch b.kind {
		case PropBuildingKonbini:
			nr, ng, nb = 0, 245, 212
		case PropBuildingCoffee:
			nr, ng, nb = 255, 160, 90
		case PropBuildingArcade:
			nr, ng, nb = 255, 42, 133
		case PropBuildingRamen:
			nr, ng, nb = 255, 92, 74
		case PropBuildingBookstore:
			nr, ng, nb = 255, 208, 108
		case PropBuildingMetro:
			nr, ng, nb = 64, 210, 255
		}
		lights = append(lights, LightSource{
			X: b.x + 48, Y: 28, Radius: 118, Kind: LightNeonSign,
			R: nr, G: ng, B: nb, Intensity: 0.85,
		})
	}

	// North Sidewalk Promenade streetlamps, benches, bike racks, hydrants, billboards & cats
	northLampsX := []float32{136, 304, 520, 744, 976, 1128}
	for _, lx := range northLampsX {
		props = append(props, makeStreetlamp(lx, 84))
		lights = append(lights, makeStreetlampLight(lx+8, 92))
	}

	northBenchesX := []float32{64, 916}
	for _, bx := range northBenchesX {
		props = append(props, makeBenchPlants(bx, 94))
	}

	props = append(props,
		makeBikeRack(458, 102),
		makeBikeRack(812, 102),
		makeHydrantBin(108, 116),
		makeHydrantBin(748, 116),
		makeNeonBillboard(548, 78),
	)
	lights = append(lights, LightSource{
		X: 572, Y: 94, Radius: 92, Kind: LightNeonSign,
		R: 0, G: 245, B: 212, Intensity: 0.72,
	})

	for _, cx := range []float32{264, 984, 1176} {
		props = append(props, MapProp{
			Kind:       PropCatsClutter,
			X:          cx,
			Y:          88,
			W:          24,
			H:          20,
			SortY:      108,
			Solid:      false,
			CastShadow: true,
		})
	}

	// --- Layer 2: Pocket-Park Islands, Flower Boxes, ZTL Signs, Scooters, Cars & Subtle Road Decals ---
	// West Pocket-Park Island: Rect{X: 256, Y: 288, W: 160, H: 112}
	props = append(props,
		makeTreeLush(272, 296),
		makeTreeLush(356, 320),
		makeBenchPlants(316, 304),
		makeStreetlamp(328, 344),
		makeFlowerBox(262, 376),
		makeFlowerBox(378, 292),
	)
	lights = append(lights, makeStreetlampLight(336, 352))

	// East Pocket-Park Island: Rect{X: 864, Y: 288, W: 160, H: 112}
	props = append(props,
		makeTreeLush(880, 320),
		makeTreeLush(964, 296),
		makeBenchPlants(924, 304),
		makeStreetlamp(936, 344),
		makeFlowerBox(870, 292),
		makeFlowerBox(986, 376),
	)
	lights = append(lights, makeStreetlampLight(944, 352))

	// ZTL Enforcement Signs flanking the 2 ZTL zones
	props = append(props,
		makeZTLSign(546, 200),
		makeZTLSign(712, 434),
	)
	lights = append(lights,
		LightSource{X: 558, Y: 210, Radius: 64, Kind: LightNeonSign, R: 255, G: 48, B: 88, Intensity: 0.78},
		LightSource{X: 724, Y: 444, Radius: 64, Kind: LightNeonSign, R: 255, G: 48, B: 88, Intensity: 0.78},
	)

	// 4 Parked Civilian Cars + 2 Scooters along curbs (clear of BusBays)
	props = append(props,
		makeParkedCar(PropCarRed, 68, 150),
		makeParkedCar(PropCarYellow, 788, 150),
		makeParkedCar(PropCarMint, 280, 504),
		makeParkedCar(PropCarRed, 928, 504),
		makeScooter(302, 152),
		makeScooter(708, 508),
	)

	// Subtle non-solid road details (manholes & asphalt cracks) outside BusBays & Crosswalks
	manholes := [][2]float32{{132, 240}, {410, 444}, {840, 232}, {1112, 440}}
	for _, m := range manholes {
		props = append(props, MapProp{
			Kind:  PropManhole,
			X:     m[0],
			Y:     m[1],
			W:     24,
			H:     24,
			SortY: m[1] - 400, // Always draw under entities as road decal
			Solid: false,
		})
	}
	cracks := [][2]float32{{210, 420}, {468, 250}, {796, 420}, {1030, 248}, {340, 224}}
	for _, c := range cracks {
		props = append(props, MapProp{
			Kind:  PropAsphaltCrack,
			X:     c[0],
			Y:     c[1],
			W:     32,
			H:     16,
			SortY: c[1] - 400,
			Solid: false,
		})
	}

	// --- Layer 3: South Waterfront Promenade, Edicola, Boardwalk & Canal Boats ---
	props = append(props,
		makeNewsstand(338, 566),
		makeNeonBillboard(824, 572),
		makeHydrantBin(124, 550),
		makeHydrantBin(968, 550),
		makeFlowerBox(182, 594),
		makeFlowerBox(648, 594),
		makeFlowerBox(1148, 594),
	)
	lights = append(lights,
		LightSource{X: 362, Y: 586, Radius: 88, Kind: LightShopWindow, R: 255, G: 215, B: 150, Intensity: 0.70},
		LightSource{X: 848, Y: 588, Radius: 92, Kind: LightNeonSign, R: 255, G: 42, B: 133, Intensity: 0.75},
	)

	southLampsX := []float32{160, 440, 790, 1120}
	for _, lx := range southLampsX {
		props = append(props, makeStreetlamp(lx, 540))
		lights = append(lights, makeStreetlampLight(lx+8, 548))
	}

	for _, bx := range []float32{236, 566, 902} {
		props = append(props, makeBenchPlants(bx, 584))
	}

	props = append(props, MapProp{
		Kind:       PropCatsClutter,
		X:          488,
		Y:          590,
		W:          24,
		H:          20,
		SortY:      610,
		Solid:      false,
		CastShadow: true,
	})

	// Canal Boats moored along South Canal Water (Y=688)
	for _, bx := range []float32{176, 616, 1020} {
		props = append(props, MapProp{
			Kind:       PropCanalBoat,
			X:          bx,
			Y:          688,
			W:          48,
			H:          24,
			SortY:      712,
			Solid:      false,
			CastShadow: false,
		})
		lights = append(lights, LightSource{
			X: bx + 8, Y: 696, Radius: 68, Kind: LightStreetlamp,
			R: 255, G: 205, B: 120, Intensity: 0.65,
		})
	}

	// Vending Machine neon lights
	for _, sh := range layout.Shops {
		lights = append(lights, LightSource{
			X: sh.X, Y: sh.Y, Radius: 88, Kind: LightVending,
			R: 80, G: 235, B: 255, Intensity: 0.80,
		})
	}

	// Slot Machine golden/magenta arcade lights
	for _, sm := range layout.SlotMachines {
		lights = append(lights, LightSource{
			X: sm.X, Y: sm.Y, Radius: 92, Kind: LightSlotMachine,
			R: 255, G: 215, B: 60, Intensity: 0.85,
		})
	}

	layout.Props = props
	layout.Lights = lights
	return layout
}

func makeNewsstand(x, y float32) MapProp {
	return MapProp{
		Kind:        PropBuildingNewsstand,
		X:           x,
		Y:           y,
		W:           48,
		H:           40,
		SortY:       y + 40,
		Solid:       true,
		CollideRect: Rect{X: x + 2, Y: y + 8, W: 44, H: 30},
		CastShadow:  true,
	}
}

func makeTreeLush(x, y float32) MapProp {
	return MapProp{
		Kind:        PropTreeLush,
		X:           x,
		Y:           y,
		W:           48,
		H:           64,
		SortY:       y + 60,
		Solid:       true,
		CollideRect: Rect{X: x + 10, Y: y + 34, W: 28, H: 28},
		CastShadow:  true,
	}
}

func makeBenchPlants(x, y float32) MapProp {
	return MapProp{
		Kind:        PropBenchPlants,
		X:           x,
		Y:           y,
		W:           40,
		H:           24,
		SortY:       y + 24,
		Solid:       true,
		CollideRect: Rect{X: x + 4, Y: y + 4, W: 32, H: 18},
		CastShadow:  true,
	}
}

func makeStreetlamp(x, y float32) MapProp {
	return MapProp{
		Kind:       PropStreetlamp,
		X:          x,
		Y:          y,
		W:          16,
		H:          48,
		SortY:      y + 48,
		Solid:      false,
		CastShadow: true,
	}
}

func makeStreetlampLight(x, y float32) LightSource {
	return LightSource{
		X: x, Y: y, Radius: 112, Kind: LightStreetlamp,
		R: 255, G: 212, B: 138, Intensity: 0.75,
	}
}

func makeParkedCar(kind PropKind, x, y float32) MapProp {
	return MapProp{
		Kind:        kind,
		X:           x,
		Y:           y,
		W:           64,
		H:           32,
		SortY:       y + 32,
		Solid:       true,
		CollideRect: Rect{X: x + 2, Y: y + 4, W: 60, H: 26},
		CastShadow:  true,
	}
}

func makeScooter(x, y float32) MapProp {
	return MapProp{
		Kind:        PropScooter,
		X:           x,
		Y:           y,
		W:           32,
		H:           24,
		SortY:       y + 24,
		Solid:       true,
		CollideRect: Rect{X: x + 3, Y: y + 4, W: 26, H: 18},
		CastShadow:  true,
	}
}

func makeBikeRack(x, y float32) MapProp {
	return MapProp{
		Kind:        PropBikeRack,
		X:           x,
		Y:           y,
		W:           32,
		H:           20,
		SortY:       y + 20,
		Solid:       true,
		CollideRect: Rect{X: x + 2, Y: y + 4, W: 28, H: 14},
		CastShadow:  true,
	}
}

func makeHydrantBin(x, y float32) MapProp {
	return MapProp{
		Kind:        PropHydrantBin,
		X:           x,
		Y:           y,
		W:           24,
		H:           24,
		SortY:       y + 24,
		Solid:       true,
		CollideRect: Rect{X: x + 3, Y: y + 6, W: 18, H: 16},
		CastShadow:  true,
	}
}

func makeZTLSign(x, y float32) MapProp {
	return MapProp{
		Kind:       PropZTLSign,
		X:          x,
		Y:          y,
		W:          24,
		H:          32,
		SortY:      y + 32,
		Solid:      false,
		CastShadow: true,
	}
}

func makeFlowerBox(x, y float32) MapProp {
	return MapProp{
		Kind:        PropFlowerBox,
		X:           x,
		Y:           y,
		W:           32,
		H:           20,
		SortY:       y + 20,
		Solid:       true,
		CollideRect: Rect{X: x + 2, Y: y + 4, W: 28, H: 14},
		CastShadow:  true,
	}
}

func makeNeonBillboard(x, y float32) MapProp {
	return MapProp{
		Kind:        PropNeonBillboard,
		X:           x,
		Y:           y,
		W:           48,
		H:           32,
		SortY:       y + 32,
		Solid:       true,
		CollideRect: Rect{X: x + 4, Y: y + 8, W: 40, H: 22},
		CastShadow:  true,
	}
}

// resolveCircleRect pushes circle (cx, cy, radius) smoothly out of axis-aligned rect r.
func resolveCircleRect(cx, cy, radius float32, r Rect) (float32, float32) {
	nearX := clamp32(cx, r.X, r.X+r.W)
	nearY := clamp32(cy, r.Y, r.Y+r.H)
	dx := cx - nearX
	dy := cy - nearY
	d2 := dx*dx + dy*dy
	if d2 >= radius*radius {
		return cx, cy
	}
	if d2 > 0.0001 {
		nx, ny, d := normalize2D(dx, dy)
		push := radius - d
		return cx + nx*push, cy + ny*push
	}
	leftDist := cx - r.X
	rightDist := (r.X + r.W) - cx
	topDist := cy - r.Y
	botDist := (r.Y + r.H) - cy

	minD := leftDist
	outX, outY := r.X-radius, cy
	if rightDist < minD {
		minD = rightDist
		outX, outY = r.X+r.W+radius, cy
	}
	if topDist < minD {
		minD = topDist
		outX, outY = cx, r.Y-radius
	}
	if botDist < minD {
		outX, outY = cx, r.Y+r.H+radius
	}
	return outX, outY
}
