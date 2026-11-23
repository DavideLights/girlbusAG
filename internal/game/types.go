package game

import "math"

// Virtual screen and arena dimensions (16:9 crisp pixel-art viewport over a 2x2 scrolling arena).
const (
	ScreenWidth  = 640
	ScreenHeight = 360
	ArenaWidth   = 1280
	ArenaHeight  = 720

	MaxBullets          = 4096
	PerfectHitRadius    = float32(24.0)
	PlayerHitboxRadius  = float32(6.0)
	BaseGrazeRadius     = float32(28.0)
	ExpressLinkMaxDist  = float32(450.0)
	BusRushSpeed        = float32(900.0)
	BasslinePulsePeriod = float32(2.5)
	BasslinePulseRadius = float32(72.0)
	BoardwalkRailY      = float32(676.0)
)

// Archetype identifies both Bus Stop shelter types and Girl horde archetypes.
type Archetype int

const (
	ArchEmo Archetype = iota
	ArchStudent
	ArchKawaii
	ArchGyaru
	ArchCount
)

// ArchetypeSpec defines base tactical stats for each Bus Stop archetype.
type ArchetypeSpec struct {
	Arch          Archetype
	Name          string
	StopName      string
	SynergyName   string
	ThrowRange    float32
	CaptureRadius float32
	ShieldHP      int
}

var archetypeSpecs = [ArchCount]ArchetypeSpec{
	ArchEmo: {
		Arch:          ArchEmo,
		Name:          "Emo",
		StopName:      "Grunge Shelter",
		SynergyName:   "SOULMATE",
		ThrowRange:    220,
		CaptureRadius: 140,
		ShieldHP:      4,
	},
	ArchStudent: {
		Arch:          ArchStudent,
		Name:          "Student",
		StopName:      "Campus LED",
		SynergyName:   "HONOR ROLL",
		ThrowRange:    300,
		CaptureRadius: 105,
		ShieldHP:      4,
	},
	ArchKawaii: {
		Arch:          ArchKawaii,
		Name:          "Kawaii",
		StopName:      "Purikura Umbrella",
		SynergyName:   "MOE CRITICAL",
		ThrowRange:    150,
		CaptureRadius: 85,
		ShieldHP:      5,
	},
	ArchGyaru: {
		Arch:          ArchGyaru,
		Name:          "Cyber-Gyaru",
		StopName:      "5G Hologram",
		SynergyName:   "VIRAL #1",
		ThrowRange:    360,
		CaptureRadius: 165,
		ShieldHP:      4,
	},
}

// GetArchetypeSpec returns the static specification for an Archetype.
func GetArchetypeSpec(a Archetype) ArchetypeSpec {
	if a < 0 || a >= ArchCount {
		return archetypeSpecs[ArchEmo]
	}
	return archetypeSpecs[a]
}

// BulletKind classifies enemy and boss projectiles in BulletPool.
type BulletKind int

const (
	BulletRazor BulletKind = iota
	BulletProtractor
	BulletTear
	BulletFlash
	BulletPolice
	BulletCiabatta
)

// BossKind identifies the implemented Gatekeeper Bosses.
type BossKind int

const (
	BossBestFriend BossKind = iota
	BossProfessor
	BossMom
)

// ActionDir constants for Minigames and the 4-Step Boarding Combo.
const (
	ActionNone  = -1
	ActionUp    = 0
	ActionRight = 1
	ActionDown  = 2
	ActionLeft  = 3
)

// Player animation states.
const (
	AnimIdle  = 0
	AnimRun   = 1
	AnimDash  = 2
	AnimThrow = 3
)

// InputState is the platform-agnostic control snapshot passed into World.Update each frame.
type InputState struct {
	MoveX, MoveY        float32
	DashJustPressed     bool
	ChronosHeld         bool
	ThrowJustPressed    bool
	AimWorldX, AimWorldY float32
	SelectArchetype     int // -1 if none, 0..3 to switch active Bus Stop archetype
	ActionDir           int // -1 none, 0=Up, 1=Right, 2=Down, 3=Left
	InteractJustPressed bool
	ChooseUpgrade       int // -1 none, 0..2 when Level-Up 3-Card choice is active
}

// Rect defines an axis-aligned rectangle in world coordinates (e.g. ZTL forbidden zones).
type Rect struct {
	X, Y, W, H float32
}

// Contains reports whether point (px, py) lies inside the rectangle.
func (r Rect) Contains(px, py float32) bool {
	return px >= r.X && px <= r.X+r.W && py >= r.Y && py <= r.Y+r.H
}

// IntersectsCircle reports whether circle (cx, cy, radius) overlaps the rectangle.
func (r Rect) IntersectsCircle(cx, cy, radius float32) bool {
	nearX := clamp32(cx, r.X, r.X+r.W)
	nearY := clamp32(cy, r.Y, r.Y+r.H)
	dx := cx - nearX
	dy := cy - nearY
	return dx*dx+dy*dy <= radius*radius
}

// FloatingText represents a diegetic world-space popup ("PERFECT DROP!", "+$150", etc.).
type FloatingText struct {
	X, Y     float32
	VY       float32
	Life     float32
	MaxLife  float32
	Text     string
	Kind     int // 0=info, 1=perfect/synergy, 2=police/alert, 3=cash/xp, 4=boss
}

// VendingShop is a Konbini Neon Vending Machine on the sidewalk for in-run healing and Police bribes.
type VendingShop struct {
	X, Y      float32
	Radius    float32
	Cooldown  float32
	HealCost  int
	BribeCost int
}

// TelegraphLine represents an upcoming lane rush or laser warning line (0.7s-0.8s).
type TelegraphLine struct {
	X1, Y1   float32
	X2, Y2   float32
	Timer    float32
	MaxTimer float32
	Kind     int // 0=PoliceCruiser, 1=ProfessorLaser
}

// ExpressLine links two simultaneous Bus Stops within ExpressLinkMaxDist.
type ExpressLine struct {
	X1, Y1 float32
	X2, Y2 float32
}

// WaveRing represents an expanding pulse ring for visual feedback (Bassline, Bullet-Cancel, Bus Wave).
type WaveRing struct {
	X, Y      float32
	Radius    float32
	MaxRadius float32
	Life      float32
	MaxLife   float32
	Kind      int // 0=Bassline, 1=PerfectCancel, 2=BusPurify, 3=CampusBell, 4=ViralFlash
}

// UpgradeKind identifies an in-run Level-Up upgrade card.
type UpgradeKind int

const (
	UpgradeCooldown UpgradeKind = iota
	UpgradeMultiHook
	UpgradeMaxStops
	UpgradeRange
	UpgradeChronos
	UpgradeVitality
)

// UpgradeCard is one of the 3 choices offered upon leveling up.
type UpgradeCard struct {
	Kind  UpgradeKind
	Title string
	Desc  string
}

// clamp32 clamps v to [lo, hi].
func clamp32(v, lo, hi float32) float32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// distSq returns the squared Euclidean distance between two points.
func distSq(x1, y1, x2, y2 float32) float32 {
	dx := x2 - x1
	dy := y2 - y1
	return dx*dx + dy*dy
}

// dist returns the Euclidean distance between two points.
func dist(x1, y1, x2, y2 float32) float32 {
	return float32(math.Sqrt(float64(distSq(x1, y1, x2, y2))))
}

// normalize2D normalizes vector (x, y) and returns (nx, ny, length).
func normalize2D(x, y float32) (float32, float32, float32) {
	l := float32(math.Sqrt(float64(x*x + y*y)))
	if l <= 0.0001 {
		return 0, 0, 0
	}
	return x / l, y / l, l
}

// pointToSegmentDistSq returns the squared distance from point (px, py) to segment (x1,y1)-(x2,y2).
func pointToSegmentDistSq(px, py, x1, y1, x2, y2 float32) float32 {
	vx := x2 - x1
	vy := y2 - y1
	lenSq := vx*vx + vy*vy
	if lenSq <= 0.0001 {
		return distSq(px, py, x1, y1)
	}
	t := ((px-x1)*vx + (py-y1)*vy) / lenSq
	t = clamp32(t, 0, 1)
	projX := x1 + t*vx
	projY := y1 + t*vy
	return distSq(px, py, projX, projY)
}
