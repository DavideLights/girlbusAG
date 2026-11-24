package game

import (
	"math"
	"math/rand"
)

// Layout constants for the locked spawn-at-click minigame / combo UI box and its integrated pad.
const (
	StopPanelDirW  = float32(212.0)
	StopPanelMaxH  = float32(78.0)
	StopPadOffsetX = float32(170.0)
	StopPadOffsetY = float32(37.0)
)

// BusStop represents a deployed Neon Bus Stop shelter with tactical radius, back-glass shield,
// hooked girls/boss tether, diegetic minigame, and 4-step boarding combo.
type BusStop struct {
	ID               int
	X, Y             float32
	Arch             Archetype
	Active           bool
	CaptureRadius    float32
	ThrowRange       float32
	ShieldHP         int
	MaxShieldHP      int
	LifeTimer        float32
	MaxLifeTimer     float32
	ShieldFlashTimer float32
	DeployAge        float32

	InZTL       bool
	InBusBay    bool
	PerfectDrop bool

	UIAnchorX  float32
	UIAnchorY  float32
	UIAnchored bool

	HookedGirlIDs []int
	HookedBoss    bool

	PlayerInside       bool
	CoyoteTimer        float32
	AlVoloBoosted      bool
	SynergyActive      bool
	EmoAbsorbAvailable bool
	BellTimer          float32
	HotspotZapTimer    float32

	Minigame         MinigameState
	Combo            ComboState
	BusRushTriggered bool
}

// ComputeStopUIAnchor locks a BusStop's HUD panel beside the click position while preserving
// generous clickable margins on all 4 sides for directional mouse actions.
func ComputeStopUIAnchor(clickScreenX, clickScreenY, safeTopY float32) (float32, float32) {
	var px float32
	if clickScreenX <= float32(ScreenWidth)*0.52 {
		px = clickScreenX + 26
	} else {
		px = clickScreenX - 26 - StopPanelDirW
	}
	py := clickScreenY - StopPanelMaxH*0.5

	px = clamp32(px, 24, float32(ScreenWidth)-StopPanelDirW-28)
	py = clamp32(py, safeTopY+12, float32(ScreenHeight)-StopPanelMaxH-58)
	px = float32(math.Round(float64(px)))
	py = float32(math.Round(float64(py)))
	return px, py
}

// PadCenter returns the locked screen-space center of this stop's 4-way directional pad.
func (s *BusStop) PadCenter() (float32, float32) {
	if s != nil && s.UIAnchored {
		return s.UIAnchorX + StopPadOffsetX, s.UIAnchorY + StopPadOffsetY
	}
	return FixedPadCenterX, FixedPadCenterY
}

// NewBusStop constructs an active BusStop at (x, y) using Alfredo's current range multiplier.
func NewBusStop(id int, x, y float32, arch Archetype, rangeMult float32, inZTL, inBusBay, perfectDrop bool, rng *rand.Rand) BusStop {
	spec := GetArchetypeSpec(arch)
	capR := spec.CaptureRadius * rangeMult
	if inZTL {
		capR *= 1.25 // ZTL risk/reward bonus: +25% capture radius
	}
	throwR := spec.ThrowRange * rangeMult
	shieldHP := spec.ShieldHP
	if inBusBay {
		shieldHP++ // Legal BUS Bay bonus: +1 Back-Glass Shield HP
	}

	return BusStop{
		ID:              id,
		X:               x,
		Y:               y,
		Arch:            arch,
		Active:          true,
		CaptureRadius:   capR,
		ThrowRange:      throwR,
		ShieldHP:        shieldHP,
		MaxShieldHP:     shieldHP,
		LifeTimer:       14.0,
		MaxLifeTimer:    14.0,
		InZTL:           inZTL,
		InBusBay:        inBusBay,
		PerfectDrop:     perfectDrop,
		HookedGirlIDs:   make([]int, 0, 8),
		BellTimer:       1.2,
		HotspotZapTimer: 0.75,
		Minigame:        NewMinigameState(arch, perfectDrop, rng),
	}
}

// GrazeRadius returns the effective Kawaii Tear-Graze radius (doubled on MOE CRITICAL synergy).
func (s *BusStop) GrazeRadius() float32 {
	if s.Arch == ArchKawaii && s.SynergyActive {
		return BaseGrazeRadius * 2.0
	}
	return BaseGrazeRadius
}

// ActivateSynergy enables archetype-specific Synergy perks on this BusStop.
func (s *BusStop) ActivateSynergy() {
	if s.SynergyActive {
		return
	}
	s.SynergyActive = true
	s.Minigame.ApplySynergy()
	if s.Arch == ArchEmo {
		s.EmoAbsorbAvailable = true
	}
}

// NeonBus is the unstoppable neon express bus summoned at Combo completion.
type NeonBus struct {
	Active       bool
	X, Y         float32
	TargetStopID int
	Speed        float32
	Arch         Archetype
	InZTL        bool
	Synergy      bool
	Collected    bool
}

// NewNeonBus spawns a NeonBus streaking from the left edge along stop.Y.
func NewNeonBus(stop *BusStop) NeonBus {
	return NeonBus{
		Active:       true,
		X:            -100,
		Y:            stop.Y,
		TargetStopID: stop.ID,
		Speed:        BusRushSpeed,
		Arch:         stop.Arch,
		InZTL:        stop.InZTL,
		Synergy:      stop.SynergyActive,
	}
}
