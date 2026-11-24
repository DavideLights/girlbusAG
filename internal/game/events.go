package game

import (
	"fmt"
	"math"
	"strings"
)

// DistrictEventKind identifies dynamic urban events triggered mid-run.
type DistrictEventKind int

const (
	DistrictEventNone      DistrictEventKind = 0
	DistrictEventZTLJam    DistrictEventKind = 1 // "INGORGO ZTL 4x", duration 22.0s
	DistrictEventHappyHour DistrictEventKind = 2 // "HAPPY HOUR NEON", duration 20.0s
	DistrictEventFlashmob  DistrictEventKind = 3 // "FLASHMOB CAMPUS", duration 25.0s
	DistrictEventBlackout  DistrictEventKind = 4 // "BLACKOUT NEON", duration 18.0s
)

// DistrictEventState tracks active and scheduled dynamic district events.
type DistrictEventState struct {
	Active                DistrictEventKind
	Timer                 float32
	Duration              float32
	NextEventTimer        float32
	LastEvent             DistrictEventKind
	FlashmobArch          Archetype
	FreeSlotSpinAvailable bool
	BlackoutScoreEarned   int
	BlackoutXPEarned      int
}

// DistrictEventTitle returns the compact uppercase banner title for a dynamic district event.
func DistrictEventTitle(kind DistrictEventKind, flashArch Archetype) string {
	switch kind {
	case DistrictEventZTLJam:
		return "INGORGO ZTL 4x"
	case DistrictEventHappyHour:
		return "HAPPY HOUR NEON"
	case DistrictEventFlashmob:
		return "FLASHMOB " + strings.ToUpper(GetArchetypeSpec(flashArch).Name)
	case DistrictEventBlackout:
		return "BLACKOUT NEON"
	default:
		return ""
	}
}

// DistrictEventSubtitle returns the tactical HUD subtitle explaining an active event's rules.
func DistrictEventSubtitle(kind DistrictEventKind, flashArch Archetype, freeSpin bool) string {
	switch kind {
	case DistrictEventZTLJam:
		return "AREE BUS -> ZTL DORATE! RICOMPENSE 4x!"
	case DistrictEventHappyHour:
		if freeSpin {
			return "-35% NEGOZI NEON + 1 SPIN SLOT GRATIS!"
		}
		return "-35% PREZZI NEI NEGOZI NEON!"
	case DistrictEventFlashmob:
		return fmt.Sprintf("ONDATA %s! SYNERGY RUSH!", strings.ToUpper(GetArchetypeSpec(flashArch).Name))
	case DistrictEventBlackout:
		return "+50% CHRONOS & +50% BONUS SCORE/XP A FINE EVENTO!"
	default:
		return ""
	}
}

// TriggerDistrictEvent activates a dynamic district event immediately and spawns its initial effects.
func (w *World) TriggerDistrictEvent(kind DistrictEventKind) {
	if kind < DistrictEventZTLJam || kind > DistrictEventBlackout {
		return
	}

	w.DistrictEvent.Active = kind
	w.DistrictEvent.LastEvent = kind
	w.DistrictEvent.FreeSlotSpinAvailable = false

	switch kind {
	case DistrictEventZTLJam:
		w.DistrictEvent.Duration = 22.0
		w.DistrictEvent.Timer = 22.0
		w.addPopup(w.Player.X, w.Player.Y-38, "INGORGO ZTL 4x! AREE BUS DORATE!", 1)
		w.addWaveRing(w.Player.X, w.Player.Y, 150, 1)
		w.addWaveRing(ArenaWidth*0.5, ArenaHeight*0.5, 180, 1)

	case DistrictEventHappyHour:
		w.DistrictEvent.Duration = 20.0
		w.DistrictEvent.Timer = 20.0
		w.DistrictEvent.FreeSlotSpinAvailable = true
		w.addPopup(w.Player.X, w.Player.Y-38, "HAPPY HOUR! -35% SHOP + 1 SPIN GRATIS!", 1)

	case DistrictEventFlashmob:
		w.DistrictEvent.Duration = 25.0
		w.DistrictEvent.Timer = 25.0
		arch := Archetype(w.rng.Intn(int(ArchCount)))
		w.DistrictEvent.FlashmobArch = arch

		count := 8 + w.rng.Intn(3)
		for i := 0; i < count && len(w.Girls) < 34; i++ {
			ang := w.rng.Float64() * 2.0 * math.Pi
			rad := float32(40.0) + w.rng.Float32()*90.0
			gx := float32(640.0) + float32(math.Cos(ang))*rad
			gy := float32(352.0) + float32(math.Sin(ang))*rad
			g := NewGirl(w.allocID(), gx, gy, arch, false, w.rng)
			if w.DifficultyTier >= 2 {
				g.ShootTimer = 0.9 + w.rng.Float32()*1.3
			}
			w.constrainGirlPos(&g)
			w.Girls = append(w.Girls, g)
		}
		w.addWaveRing(640, 352, 165, 4)
		w.addPopup(w.Player.X, w.Player.Y-38, "FLASHMOB "+strings.ToUpper(GetArchetypeSpec(arch).Name)+"! SYNERGY RUSH!", 1)

	case DistrictEventBlackout:
		w.DistrictEvent.Duration = 18.0
		w.DistrictEvent.Timer = 18.0
		w.DistrictEvent.BlackoutScoreEarned = 0
		w.DistrictEvent.BlackoutXPEarned = 0
		w.Player.ChronosMeter = clamp32(w.Player.ChronosMeter+w.Player.ChronosMax*0.50, 0, w.Player.ChronosMax)
		w.addPopup(w.Player.X, w.Player.Y-38, "BLACKOUT NEON! +50% CHRONOS & +50% BONUS!", 1)
	}

	w.emitEvent(EventDistrictEventStart, int(kind), false)
}

func (w *World) updateDistrictEvents(dt float32) {
	if w.TutorialMode || w.BossRushMode {
		return
	}

	if w.DistrictEvent.Active != DistrictEventNone {
		w.DistrictEvent.Timer -= dt
		if w.DistrictEvent.Timer <= 0 {
			ended := w.DistrictEvent.Active
			if ended == DistrictEventBlackout {
				bonusScore := max(150, w.DistrictEvent.BlackoutScoreEarned/2)
				bonusXP := max(25, w.DistrictEvent.BlackoutXPEarned/2)
				w.Score += bonusScore
				w.addPopup(w.Player.X, w.Player.Y-38, "BLACKOUT SUPERATO! +50% SCORE & XP!", 3)
				if w.Player.AddXP(bonusXP) {
					w.triggerLevelUp()
				}
			}
			w.EventsCompleted++
			w.emitEvent(EventDistrictEventEnd, int(ended), false)
			w.DistrictEvent.Active = DistrictEventNone
			w.DistrictEvent.Timer = 0
			w.DistrictEvent.Duration = 0
			w.DistrictEvent.FreeSlotSpinAvailable = false
			w.DistrictEvent.NextEventTimer = 60.0 + w.rng.Float32()*30.0
		}
		return
	}

	if w.Boss.Active {
		if w.DistrictEvent.NextEventTimer < 8.0 {
			w.DistrictEvent.NextEventTimer = 8.0
		}
		return
	}

	w.DistrictEvent.NextEventTimer -= dt
	if w.DistrictEvent.NextEventTimer > 0 {
		return
	}

	if w.BossScheduleIdx < len(w.BossScheduleTimes) && w.BossScheduleTimes[w.BossScheduleIdx]-w.RunTime < 15.0 {
		w.DistrictEvent.NextEventTimer = 12.0
		return
	}

	var candidates [4]DistrictEventKind
	n := 0
	for k := DistrictEventZTLJam; k <= DistrictEventBlackout; k++ {
		if k != w.DistrictEvent.LastEvent {
			candidates[n] = k
			n++
		}
	}
	picked := candidates[w.rng.Intn(n)]
	w.TriggerDistrictEvent(picked)
}
