package game

// PropKind classifies static and interactive Pop-Urban city props in the arena.
type PropKind int

const (
	PropBuildingKonbini PropKind = iota
	PropBuildingCoffee
	PropBuildingArcade
	PropTreeLush
	PropStreetlamp
	PropBenchPlants
	PropCatsClutter
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
	CrosswalkCols []int
	CrosswalkRows []int
	BoardwalkY    float32
	CanalY        float32
}

// BuildDistrictLayout constructs the stratified 3-layer Pop-Urban city district for the 1280x720 arena.
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
	}

	props := make([]MapProp, 0, 40)
	lights := make([]LightSource, 0, 40)

	// --- Layer 1: North Storefront Boulevard (Y = 0..144) ---
	type storefrontSpec struct {
		x    float32
		kind PropKind
	}
	storefronts := []storefrontSpec{
		// Block 1 (X = 32..240), followed by North Plaza Entrance Alley (240..320)
		{x: 32, kind: PropBuildingKonbini},
		{x: 144, kind: PropBuildingCoffee},
		// Block 2 (X = 320..528), followed by Central Cross-Street North Avenue (528..752)
		{x: 320, kind: PropBuildingArcade},
		{x: 432, kind: PropBuildingKonbini},
		// Block 3 (X = 752..960), followed by East Alley (960..1024)
		{x: 752, kind: PropBuildingCoffee},
		{x: 864, kind: PropBuildingArcade},
		// Block 4 (X = 1024..1232)
		{x: 1024, kind: PropBuildingKonbini},
		{x: 1136, kind: PropBuildingCoffee},
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
		}
		lights = append(lights, LightSource{
			X: b.x + 48, Y: 28, Radius: 115, Kind: LightNeonSign,
			R: nr, G: ng, B: nb, Intensity: 0.85,
		})
	}

	// North Sidewalk Promenade streetlamps, benches, and cats/clutter (Y = 72..144)
	northLampsX := []float32{136, 304, 520, 744, 976, 1128}
	for _, lx := range northLampsX {
		props = append(props, makeStreetlamp(lx, 84))
		lights = append(lights, makeStreetlampLight(lx+8, 92))
	}

	northBenchesX := []float32{64, 348, 888}
	for _, bx := range northBenchesX {
		props = append(props, makeBenchPlants(bx, 92))
	}

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

	// --- Layer 2: Tactical Urban Pocket-Park Islands & Roadside Parked Cars ---
	// West Pocket-Park Island: Rect{X: 256, Y: 288, W: 160, H: 112}
	props = append(props,
		makeTreeLush(272, 296),
		makeTreeLush(356, 320),
		makeBenchPlants(316, 304),
		makeStreetlamp(328, 344),
	)
	lights = append(lights, makeStreetlampLight(336, 352))

	// East Pocket-Park Island: Rect{X: 864, Y: 288, W: 160, H: 112}
	props = append(props,
		makeTreeLush(880, 320),
		makeTreeLush(964, 296),
		makeBenchPlants(924, 304),
		makeStreetlamp(936, 344),
	)
	lights = append(lights, makeStreetlampLight(944, 352))

	// 4 Parked Civilian Cars along the avenue curbs (physical cover, clear of BusBays)
	props = append(props,
		makeParkedCar(PropCarRed, 68, 150),
		makeParkedCar(PropCarYellow, 788, 150),
		makeParkedCar(PropCarMint, 280, 504),
		makeParkedCar(PropCarRed, 928, 504),
	)

	// --- Layer 3: South Waterfront Promenade & Boardwalk (Y = 560..680) ---
	southLampsX := []float32{160, 440, 744, 1120}
	for _, lx := range southLampsX {
		props = append(props, makeStreetlamp(lx, 540))
		lights = append(lights, makeStreetlampLight(lx+8, 548))
	}

	for _, bx := range []float32{220, 620, 880} {
		props = append(props, makeBenchPlants(bx, 584))
	}

	props = append(props, MapProp{
		Kind:       PropCatsClutter,
		X:          480,
		Y:          590,
		W:          24,
		H:          20,
		SortY:      610,
		Solid:      false,
		CastShadow: true,
	})

	// Vending Machine neon lights
	for _, sh := range layout.Shops {
		lights = append(lights, LightSource{
			X: sh.X, Y: sh.Y, Radius: 88, Kind: LightVending,
			R: 80, G: 235, B: 255, Intensity: 0.80,
		})
	}

	layout.Props = props
	layout.Lights = lights
	return layout
}

func makeTreeLush(x, y float32) MapProp {
	// Trunk collider centered near base (~28x28, effective radius ~16px)
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
	// Circle center is inside rect: push out along nearest edge
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
