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
	PropBuildingPizzeria
	PropBuildingVinylBar
	PropBuildingGelateria
	PropFountain
	PropArchGate
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
	PropPhoneBooth
	PropFoodCart
	PropKioskUmbrella
	PropTrafficLight
	PropBollardGlow
	PropPlanterNeon
	PropVendingDrink
	PropBusShelterDeco
	PropCarRed
	PropCarYellow
	PropCarMint
	PropCarVanCyber
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
	LightLantern
	LightFountain
	LightBollard
	LightTraffic
	LightVehicle
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
			{X: 220, Y: 92, Radius: 54, HealCost: EspressoHealCost, BribeCost: 100},
			{X: 500, Y: 92, Radius: 54, HealCost: EspressoHealCost, BribeCost: 100},
			{X: 1040, Y: 592, Radius: 54, HealCost: EspressoHealCost, BribeCost: 100},
		},
		// 2 Rizz-Slots 777 Cabinets: North Arcade & South Waterfront Boardwalk.
		SlotMachines: []SlotMachine{
			{X: 386, Y: 94, Radius: 54, SpinCost: 35},
			{X: 774, Y: 586, Radius: 54, SpinCost: 35},
		},
	}

	props := make([]MapProp, 0, 128)
	lights := make([]LightSource, 0, 96)

	// --- Layer 1: Varied North Storefront Boulevard (Y = 0..144) ---
	type storefrontSpec struct {
		x    float32
		kind PropKind
	}
	storefronts := []storefrontSpec{
		// Block 1 (X = 32..240), West Alley 240..320
		{x: 32, kind: PropBuildingKonbini},
		{x: 144, kind: PropBuildingPizzeria},
		// Block 2 (X = 320..528), Central Cross-Street North Plaza 528..752
		{x: 320, kind: PropBuildingArcade},
		{x: 432, kind: PropBuildingVinylBar},
		// Block 3 (X = 752..960), East Alley 960..1024
		{x: 752, kind: PropBuildingMetro},
		{x: 864, kind: PropBuildingGelateria},
		// Block 4 (X = 1024..1232)
		{x: 1024, kind: PropBuildingCoffee},
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
		// Tailored warm interior shop window glow + themed colored neon sign light
		wr, wg, wb := uint8(255), uint8(222), uint8(165)
		var nr, ng, nb uint8
		switch b.kind {
		case PropBuildingKonbini:
			wr, wg, wb = 248, 240, 210
			nr, ng, nb = 0, 245, 212
		case PropBuildingPizzeria:
			wr, wg, wb = 255, 210, 138
			nr, ng, nb = 255, 86, 62
		case PropBuildingCoffee:
			wr, wg, wb = 255, 216, 148
			nr, ng, nb = 255, 160, 90
		case PropBuildingArcade:
			wr, wg, wb = 240, 185, 255
			nr, ng, nb = 255, 42, 133
		case PropBuildingVinylBar:
			wr, wg, wb = 255, 198, 142
			nr, ng, nb = 192, 88, 255
		case PropBuildingRamen:
			wr, wg, wb = 255, 214, 140
			nr, ng, nb = 255, 92, 74
		case PropBuildingBookstore:
			wr, wg, wb = 255, 224, 158
			nr, ng, nb = 255, 208, 108
		case PropBuildingMetro:
			wr, wg, wb = 212, 238, 255
			nr, ng, nb = 64, 210, 255
		case PropBuildingGelateria:
			wr, wg, wb = 255, 226, 210
			nr, ng, nb = 255, 132, 198
		}
		lights = append(lights,
			LightSource{
				X: b.x + 48, Y: 58, Radius: 96, Kind: LightShopWindow,
				R: wr, G: wg, B: wb, Intensity: 0.68,
			},
			LightSource{
				X: b.x + 48, Y: 28, Radius: 118, Kind: LightNeonSign,
				R: nr, G: ng, B: nb, Intensity: 0.85,
			},
		)
	}

	// Iconic North Central Plaza Landmarks (Fountain + Overhead ArchGate + Flanking PhoneBooth + Bollards)
	props = append(props,
		makeFountain(608, 42),
		makeArchGate(592, 108, 146),
		makePhoneBooth(534, 110),
		makeBollardGlow(572, 124),
		makeBollardGlow(694, 124),
	)
	lights = append(lights,
		LightSource{X: 640, Y: 68, Radius: 110, Kind: LightFountain, R: 60, G: 235, B: 255, Intensity: 1.0},
		LightSource{X: 602, Y: 122, Radius: 72, Kind: LightLantern, R: 255, G: 198, B: 112, Intensity: 0.78},
		LightSource{X: 678, Y: 122, Radius: 72, Kind: LightLantern, R: 255, G: 198, B: 112, Intensity: 0.78},
		LightSource{X: 546, Y: 124, Radius: 76, Kind: LightNeonSign, R: 255, G: 68, B: 210, Intensity: 0.80},
		LightSource{X: 580, Y: 132, Radius: 56, Kind: LightBollard, R: 0, G: 245, B: 212, Intensity: 0.68},
		LightSource{X: 702, Y: 132, Radius: 56, Kind: LightBollard, R: 0, G: 245, B: 212, Intensity: 0.68},
	)

	// North Sidewalk Promenade streetlamps, benches, bike racks, hydrants, drink machines, patio kiosks, food cart, transit kiosk & cats
	northLampsX := []float32{168, 304, 518, 744, 1024, 1128}
	for _, lx := range northLampsX {
		props = append(props, makeStreetlamp(lx, 84))
		lights = append(lights, makeStreetlampLight(lx+12, 94))
	}

	northBenchesX := []float32{60, 1184}
	for _, bx := range northBenchesX {
		props = append(props, makeBenchPlants(bx, 96))
	}

	props = append(props,
		makeBusShelterDeco(108, 108),
		makeVendingDrink(262, 86),
		makeBikeRack(446, 106),
		makeHydrantBin(28, 116),
		makeHydrantBin(722, 116),
		makeNeonBillboard(548, 78),
		makeVendingDrink(760, 86),
		makeBikeRack(812, 102),
		makeKioskUmbrella(904, 86),
		makeFoodCart(966, 96),
		makeKioskUmbrella(1076, 86),
	)
	lights = append(lights,
		LightSource{X: 132, Y: 126, Radius: 82, Kind: LightShopWindow, R: 255, G: 220, B: 150, Intensity: 0.72},
		LightSource{X: 276, Y: 104, Radius: 82, Kind: LightVending, R: 70, G: 215, B: 255, Intensity: 0.78},
		LightSource{X: 572, Y: 94, Radius: 92, Kind: LightNeonSign, R: 0, G: 245, B: 212, Intensity: 0.72},
		LightSource{X: 774, Y: 104, Radius: 82, Kind: LightVending, R: 70, G: 215, B: 255, Intensity: 0.78},
		LightSource{X: 924, Y: 98, Radius: 72, Kind: LightLantern, R: 255, G: 210, B: 145, Intensity: 0.72},
		LightSource{X: 990, Y: 104, Radius: 86, Kind: LightLantern, R: 255, G: 184, B: 96, Intensity: 0.82},
		LightSource{X: 1096, Y: 98, Radius: 72, Kind: LightLantern, R: 255, G: 210, B: 138, Intensity: 0.72},
	)

	for _, cx := range []float32{294, 1016, 1234} {
		props = append(props, MapProp{
			Kind:       PropCatsClutter,
			X:          cx,
			Y:          92,
			W:          24,
			H:          20,
			SortY:      112,
			Solid:      false,
			CastShadow: true,
		})
	}

	// --- Layer 2: Pocket-Park Islands, Planters, Traffic Lights, ZTL Signs, Vehicles & Road Decals ---
	// West Pocket-Park Island: Rect{X: 256, Y: 288, W: 160, H: 112}
	props = append(props,
		makeTreeLush(272, 296),
		makeTreeLush(356, 320),
		makeBenchPlants(316, 304),
		makeStreetlamp(328, 344),
		makeFlowerBox(262, 376),
		makeFlowerBox(378, 292),
		makeBollardGlow(250, 284),
		makeBollardGlow(402, 380),
	)
	lights = append(lights,
		makeStreetlampLight(340, 354),
		LightSource{X: 258, Y: 292, Radius: 56, Kind: LightBollard, R: 0, G: 245, B: 212, Intensity: 0.68},
		LightSource{X: 410, Y: 388, Radius: 56, Kind: LightBollard, R: 0, G: 245, B: 212, Intensity: 0.68},
	)

	// East Pocket-Park Island: Rect{X: 864, Y: 288, W: 160, H: 112}
	props = append(props,
		makeTreeLush(880, 320),
		makeTreeLush(964, 296),
		makeBenchPlants(924, 304),
		makeStreetlamp(936, 344),
		makeFlowerBox(870, 292),
		makeFlowerBox(986, 376),
		makeBollardGlow(862, 380),
		makeBollardGlow(1014, 284),
	)
	lights = append(lights,
		makeStreetlampLight(948, 354),
		LightSource{X: 870, Y: 388, Radius: 56, Kind: LightBollard, R: 0, G: 245, B: 212, Intensity: 0.68},
		LightSource{X: 1022, Y: 292, Radius: 56, Kind: LightBollard, R: 0, G: 245, B: 212, Intensity: 0.68},
	)

	// Tactical Low-Cover Illuminated Planters flanking West & East Avenues
	props = append(props,
		makePlanterNeon(72, 324),
		makePlanterNeon(1152, 324),
	)
	lights = append(lights,
		LightSource{X: 96, Y: 336, Radius: 64, Kind: LightBollard, R: 255, G: 196, B: 112, Intensity: 0.68},
		LightSource{X: 1176, Y: 336, Radius: 64, Kind: LightBollard, R: 255, G: 196, B: 112, Intensity: 0.68},
	)

	// 4 Central Cross-Street Intersection Traffic Lights (non-solid slim poles + traffic signal spill)
	type trafficSpec struct {
		x, y    float32
		r, g, b uint8
	}
	trafficLights := []trafficSpec{
		{x: 530, y: 146, r: 56, g: 255, b: 148},
		{x: 730, y: 146, r: 255, g: 194, b: 56},
		{x: 530, y: 502, r: 255, g: 194, b: 56},
		{x: 730, y: 502, r: 56, g: 255, b: 148},
	}
	for _, tl := range trafficLights {
		props = append(props, makeTrafficLight(tl.x, tl.y))
		lights = append(lights, LightSource{
			X: tl.x + 12, Y: tl.y + 12, Radius: 72, Kind: LightTraffic,
			R: tl.r, G: tl.g, B: tl.b, Intensity: 0.72,
		})
	}

	// ZTL Enforcement Signs flanking the 2 ZTL zones
	props = append(props,
		makeZTLSign(546, 200),
		makeZTLSign(712, 434),
	)
	lights = append(lights,
		LightSource{X: 558, Y: 210, Radius: 64, Kind: LightNeonSign, R: 255, G: 48, B: 88, Intensity: 0.78},
		LightSource{X: 724, Y: 444, Radius: 64, Kind: LightNeonSign, R: 255, G: 48, B: 88, Intensity: 0.78},
	)

	// 4 Parked Civilian Cars + 1 Cyber Van + 2 Scooters along curbs (clear of BusBays & ZTLZones)
	type carSpec struct {
		kind PropKind
		x, y float32
	}
	parkedCars := []carSpec{
		{kind: PropCarRed, x: 68, y: 150},
		{kind: PropCarYellow, x: 788, y: 150},
		{kind: PropCarMint, x: 280, y: 504},
		{kind: PropCarRed, x: 928, y: 504},
	}
	for _, c := range parkedCars {
		props = append(props, makeParkedCar(c.kind, c.x, c.y))
		lights = append(lights, LightSource{
			X: c.x + 48, Y: c.y + 16, Radius: 54, Kind: LightVehicle,
			R: 255, G: 235, B: 175, Intensity: 0.55,
		})
	}
	props = append(props,
		makeCyberVan(1084, 228),
		makeScooter(302, 152),
		makeScooter(724, 508),
	)
	lights = append(lights,
		LightSource{X: 1148, Y: 248, Radius: 68, Kind: LightVehicle, R: 255, G: 236, B: 170, Intensity: 0.65},
		LightSource{X: 1084, Y: 248, Radius: 48, Kind: LightVehicle, R: 255, G: 52, B: 74, Intensity: 0.58},
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

	// --- Layer 3: South Waterfront Promenade, Plaza Arch, Edicola, Boardwalk & Canal Boats ---
	props = append(props,
		makeHydrantBin(92, 556),
		makeFoodCart(140, 554),
		makeFlowerBox(182, 594),
		makeBenchPlants(226, 588),
		makeKioskUmbrella(280, 560),
		makeNewsstand(338, 566),
		makeBenchPlants(548, 586),
		makeBollardGlow(570, 554),
		makeArchGate(592, 538, 576),
		makeBollardGlow(694, 554),
		makeFlowerBox(624, 596),
		makePhoneBooth(726, 556),
		makeNeonBillboard(824, 572),
		makeKioskUmbrella(890, 560),
		makeBenchPlants(944, 588),
		makeHydrantBin(968, 550),
		makeBusShelterDeco(1142, 554),
		makeFlowerBox(1148, 594),
	)
	lights = append(lights,
		LightSource{X: 164, Y: 562, Radius: 86, Kind: LightLantern, R: 255, G: 184, B: 96, Intensity: 0.82},
		LightSource{X: 300, Y: 572, Radius: 72, Kind: LightLantern, R: 255, G: 210, B: 140, Intensity: 0.72},
		LightSource{X: 362, Y: 586, Radius: 88, Kind: LightShopWindow, R: 255, G: 215, B: 150, Intensity: 0.70},
		LightSource{X: 578, Y: 562, Radius: 56, Kind: LightBollard, R: 0, G: 245, B: 212, Intensity: 0.68},
		LightSource{X: 602, Y: 552, Radius: 74, Kind: LightLantern, R: 60, G: 235, B: 255, Intensity: 0.80},
		LightSource{X: 678, Y: 552, Radius: 74, Kind: LightLantern, R: 255, G: 210, B: 105, Intensity: 0.80},
		LightSource{X: 702, Y: 562, Radius: 56, Kind: LightBollard, R: 0, G: 245, B: 212, Intensity: 0.68},
		LightSource{X: 738, Y: 570, Radius: 76, Kind: LightNeonSign, R: 0, G: 235, B: 255, Intensity: 0.80},
		LightSource{X: 848, Y: 588, Radius: 92, Kind: LightNeonSign, R: 255, G: 42, B: 133, Intensity: 0.75},
		LightSource{X: 910, Y: 572, Radius: 72, Kind: LightLantern, R: 255, G: 210, B: 140, Intensity: 0.72},
		LightSource{X: 1166, Y: 572, Radius: 82, Kind: LightShopWindow, R: 255, G: 220, B: 150, Intensity: 0.72},
	)

	southLampsX := []float32{118, 440, 796, 1110}
	for _, lx := range southLampsX {
		props = append(props, makeStreetlamp(lx, 540))
		lights = append(lights, makeStreetlampLight(lx+12, 550))
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

	// Canal Boats moored along South Canal Water (Y=688) with warm bow lanterns
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
			X: bx + 41, Y: 690, Radius: 68, Kind: LightLantern,
			R: 255, G: 205, B: 120, Intensity: 0.70,
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

func makeFountain(x, y float32) MapProp {
	return MapProp{
		Kind:        PropFountain,
		X:           x,
		Y:           y,
		W:           64,
		H:           48,
		SortY:       y + 44,
		Solid:       true,
		CollideRect: Rect{X: x + 6, Y: y + 14, W: 52, H: 28},
		CastShadow:  true,
	}
}

func makeArchGate(x, y, sortY float32) MapProp {
	return MapProp{
		Kind:       PropArchGate,
		X:          x,
		Y:          y,
		W:          96,
		H:          40,
		SortY:      sortY,
		Solid:      false,
		CastShadow: true,
	}
}

func makePhoneBooth(x, y float32) MapProp {
	return MapProp{
		Kind:        PropPhoneBooth,
		X:           x,
		Y:           y,
		W:           24,
		H:           40,
		SortY:       y + 38,
		Solid:       true,
		CollideRect: Rect{X: x + 2, Y: y + 10, W: 20, H: 26},
		CastShadow:  true,
	}
}

func makeFoodCart(x, y float32) MapProp {
	return MapProp{
		Kind:        PropFoodCart,
		X:           x,
		Y:           y,
		W:           48,
		H:           36,
		SortY:       y + 34,
		Solid:       true,
		CollideRect: Rect{X: x + 4, Y: y + 10, W: 40, H: 24},
		CastShadow:  true,
	}
}

func makeKioskUmbrella(x, y float32) MapProp {
	return MapProp{
		Kind:        PropKioskUmbrella,
		X:           x,
		Y:           y,
		W:           40,
		H:           36,
		SortY:       y + 34,
		Solid:       true,
		CollideRect: Rect{X: x + 6, Y: y + 10, W: 28, H: 22},
		CastShadow:  true,
	}
}

func makeTrafficLight(x, y float32) MapProp {
	return MapProp{
		Kind:       PropTrafficLight,
		X:          x,
		Y:          y,
		W:          24,
		H:          48,
		SortY:      y + 48,
		Solid:      false,
		CastShadow: true,
	}
}

func makeBollardGlow(x, y float32) MapProp {
	return MapProp{
		Kind:        PropBollardGlow,
		X:           x,
		Y:           y,
		W:           16,
		H:           24,
		SortY:       y + 22,
		Solid:       true,
		CollideRect: Rect{X: x + 3, Y: y + 10, W: 10, H: 12},
		CastShadow:  true,
	}
}

func makePlanterNeon(x, y float32) MapProp {
	return MapProp{
		Kind:        PropPlanterNeon,
		X:           x,
		Y:           y,
		W:           48,
		H:           24,
		SortY:       y + 24,
		Solid:       true,
		CollideRect: Rect{X: x + 3, Y: y + 6, W: 42, H: 16},
		CastShadow:  true,
	}
}

func makeVendingDrink(x, y float32) MapProp {
	return MapProp{
		Kind:        PropVendingDrink,
		X:           x,
		Y:           y,
		W:           28,
		H:           36,
		SortY:       y + 34,
		Solid:       true,
		CollideRect: Rect{X: x + 2, Y: y + 8, W: 24, H: 26},
		CastShadow:  true,
	}
}

func makeBusShelterDeco(x, y float32) MapProp {
	return MapProp{
		Kind:        PropBusShelterDeco,
		X:           x,
		Y:           y,
		W:           48,
		H:           36,
		SortY:       y + 34,
		Solid:       true,
		CollideRect: Rect{X: x + 4, Y: y + 10, W: 40, H: 24},
		CastShadow:  true,
	}
}

func makeCyberVan(x, y float32) MapProp {
	return MapProp{
		Kind:        PropCarVanCyber,
		X:           x,
		Y:           y,
		W:           64,
		H:           36,
		SortY:       y + 36,
		Solid:       true,
		CollideRect: Rect{X: x + 4, Y: y + 8, W: 56, H: 26},
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
		W:          24,
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
		W:           48,
		H:           32,
		SortY:       y + 32,
		Solid:       true,
		CollideRect: Rect{X: x + 2, Y: y + 4, W: 44, H: 26},
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
