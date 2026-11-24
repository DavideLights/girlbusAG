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
	BasslinePulsePeriod   = float32(2.5)
	BasslinePulseRadius   = float32(72.0)
	BoardwalkRailY        = float32(676.0)
	GlobalGrazeRadius     = float32(18.0)
	InputBufferDuration   = float32(0.120)
	TetherCoyoteDuration  = float32(0.35)
	StreakWindowDuration  = float32(8.0)
	BusSlipstreamDuration = float32(1.5)
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
		StopName:      "Pensilina Grunge",
		SynergyName:   "SOULMATE",
		ThrowRange:    220,
		CaptureRadius: 140,
		ShieldHP:      4,
	},
	ArchStudent: {
		Arch:          ArchStudent,
		Name:          "Studentessa",
		StopName:      "Pensilina Campus LED",
		SynergyName:   "HONOR ROLL",
		ThrowRange:    300,
		CaptureRadius: 105,
		ShieldHP:      4,
	},
	ArchKawaii: {
		Arch:          ArchKawaii,
		Name:          "Kawaii",
		StopName:      "Ombrellino Purikura",
		SynergyName:   "MOE CRITICAL",
		ThrowRange:    150,
		CaptureRadius: 85,
		ShieldHP:      5,
	},
	ArchGyaru: {
		Arch:          ArchGyaru,
		Name:          "Cyber-Gyaru",
		StopName:      "Ologramma 5G",
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
	BulletLipstick
	BulletContract
	BulletSeal
)

// BossKind identifies all 6 Gatekeeper Bosses.
type BossKind int

const (
	BossBestFriend BossKind = iota
	BossLover
	BossProfessor
	BossExecutive
	BossRector
	BossMom
	BossCount
)

// RelicKind identifies passive in-run Relics bought at Vending Shops.
type RelicKind int

const (
	RelicNoiseCancel RelicKind = iota
	RelicMonthlyPass
	RelicArmoredGlass
	RelicLuckySocks
	RelicDoubleEspresso
	RelicFakeVIPPass
	RelicCount
)

// RelicSpec holds the Italian name and effect description for a passive Relic.
type RelicSpec struct {
	Kind RelicKind
	Name string
	Desc string
}

var relicSpecs = [RelicCount]RelicSpec{
	RelicNoiseCancel:    {Kind: RelicNoiseCancel, Name: "Cuffie Noise-Cancel", Desc: "-15% Cooldown Fermate, +20% Ricarica Chronos"},
	RelicMonthlyPass:    {Kind: RelicMonthlyPass, Name: "Abbonamento Mensile", Desc: "+1 Scudo Vetro, +15% Raggio Fermata"},
	RelicArmoredGlass:   {Kind: RelicArmoredGlass, Name: "Vetro Blindato", Desc: "+3 Scudi Vetro, Perfect Drop pulisce +35px"},
	RelicLuckySocks:     {Kind: RelicLuckySocks, Name: "Calzini Portafortuna", Desc: "+Fortuna Rizz-Slots 777 & +15% Soldi"},
	RelicDoubleEspresso: {Kind: RelicDoubleEspresso, Name: "Doppio Espresso", Desc: "+18% Velocita Movimento & +25% Minigiochi"},
	RelicFakeVIPPass:    {Kind: RelicFakeVIPPass, Name: "Pass VIP Falso", Desc: "+25% XP/Soldi & +1 Ragazza Multi-Hook"},
}

// GetRelicSpec returns the specification for a RelicKind.
func GetRelicSpec(r RelicKind) RelicSpec {
	if r < 0 || r >= RelicCount {
		return relicSpecs[0]
	}
	return relicSpecs[r]
}

// ActionDir constants for Minigames and the 4-Step Boarding Combo.
const (
	ActionNone  = -1
	ActionUp    = 0
	ActionRight = 1
	ActionDown  = 2
	ActionLeft  = 3
)

// ShopAction constants for the Vending Shop mini-menu (0 = none so zero-value InputState is safe).
const (
	ShopActionNone    = 0
	ShopActionHeal    = 1
	ShopActionBribe   = 2
	ShopActionRelic   = 3
	ShopActionUpgrade = 4
)

// Player animation states.
const (
	AnimIdle  = 0
	AnimRun   = 1
	AnimDash  = 2
	AnimThrow = 3
)

// WorldEventKind identifies discrete simulation events consumed by Audio and Challenge trackers.
type WorldEventKind int

const (
	EventThrow WorldEventKind = iota
	EventPerfectDrop
	EventHook
	EventMinigameHit
	EventMinigameMiss
	EventComboStep
	EventBusRush
	EventDash
	EventPlayerHit
	EventPoliceStarUp
	EventBossSpawn
	EventBossBreak
	EventBossDefeated
	EventLevelUp
	EventShopBuy
	EventSlotTick
	EventSlotWin
	EventSlotJackpot
	EventSlotBust
	EventSynergy
	EventGirlBoarded
	EventGraze
	EventDashParry
	EventStopRecall
	EventSlipstream
	EventFeverUp
	EventHealPickup
)

// WorldEvent records a single simulation event emitted during World.Update.
type WorldEvent struct {
	Kind  WorldEventKind
	Value int  // Optional context: BossKind, Archetype, Combo step, HookCount, etc.
	Flag  bool // Optional context: e.g. hit by Ciabatta, PerfectDrop, etc.
}

// InputState is the platform-agnostic control snapshot passed into World.Update each frame.
type InputState struct {
	MoveX, MoveY         float32
	DashJustPressed      bool
	ChronosHeld          bool
	ChronosJustPressed   bool
	ChronosToggleMode    bool
	ThrowJustPressed     bool
	RecallJustPressed    bool
	AimWorldX, AimWorldY float32
	AimIsDirectional     bool
	AimDirX, AimDirY     float32
	SelectArchetype      int // -1 if none, 0..3 to switch active Bus Stop archetype
	ActionDir            int // -1 none, 0=Up, 1=Right, 2=Down, 3=Left
	ActionHeld           bool
	MouseHoverDir        int // -1 none, 0=Up, 1=Right, 2=Down, 3=Left (live quadrant hover in stopPhase == 2)
	InteractJustPressed  bool
	ChooseUpgrade        int // -1 none, 0..2 when Level-Up 3-Card choice is active
	ShopAction           int // 0 none, 1..4 when Vending Shop modal is active
	SlotSpinJustPressed  bool
	CancelJustPressed    bool
	MouseScreenX         float32
	MouseScreenY         float32
	MouseLeftJustPressed bool
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
	X, Y    float32
	VY      float32
	Life    float32
	MaxLife float32
	Text    string
	Kind    int // 0=info, 1=perfect/synergy, 2=police/alert, 3=cash/xp, 4=boss
}

// VendingShop is a Konbini Neon Vending Machine on the sidewalk for in-run items, bribes, and relics.
type VendingShop struct {
	X, Y      float32
	Radius    float32
	Cooldown  float32
	HealCost  int
	BribeCost int
}

// SlotMachine is an interactive Rizz-Slots 777 cabinet placed in the district.
type SlotMachine struct {
	X, Y      float32
	Radius    float32
	SpinCost  int
	SpinsUsed int
}

// TelegraphLine represents an upcoming lane rush or laser warning line (0.7s-0.8s).
type TelegraphLine struct {
	X1, Y1   float32
	X2, Y2   float32
	Timer    float32
	MaxTimer float32
	Kind     int // 0=PoliceCruiser, 1=Professor/Rector Laser
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

// BusSlipstream represents the temporary aerodynamic speed ribbon left behind by a streaking NeonBus.
type BusSlipstream struct {
	X1, X2, Y     float32
	Life, MaxLife float32
}

// HealPickup represents a world-space healing collectible ("Cure in giro") dropped across the district.
type HealPickup struct {
	X, Y      float64
	Amount    int // 1 or 2 HP
	Chronos   float64
	LifeTimer float64
	BobPhase  float64
	Big       bool
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
