package game

import (
	"fmt"
	"math"
)

const (
	chestInteractRadius  = float32(44.0)
	chestKeyPressRadius  = float32(68.0)
	chestClearRadius     = float32(130.0)
	chestOpenAnimSeconds = float32(0.65)
	chestRestockDelaySec = float32(4.0)
	chestBaseCashReward  = 80.0
)

func (w *World) initChestsFromMap() {
	if len(w.ChestSpots) == 0 {
		w.Chests = nil
		return
	}
	chests := make([]NeonChest, 0, len(w.ChestSpots))
	for i, spot := range w.ChestSpots {
		cx, cy := w.findSafePickupSpot(spot[0], spot[1], true)
		chests = append(chests, NeonChest{
			X:        float32(cx),
			Y:        float32(cy),
			Radius:   chestInteractRadius,
			Opened:   false,
			BobPhase: float32(i) * 2.1,
		})
	}
	w.Chests = chests
}

// ClosedChestCount returns how many unopened Neon Chests currently remain in the district.
func (w *World) ClosedChestCount() int {
	n := 0
	for i := range w.Chests {
		if !w.Chests[i].Opened {
			n++
		}
	}
	return n
}

func (w *World) restockChests() {
	if len(w.Chests) == 0 {
		return
	}
	for i := range w.Chests {
		w.Chests[i].Opened = false
		w.Chests[i].OpenAnimTimer = 0
	}
	w.chestRestockTimer = 0
	w.addPopup(w.Player.X, w.Player.Y-40, "FORZIERI NEON RIFORNITI!", 1)
}

// OpenChest unlocks chest idx using one Boss Key, granting a guaranteed Relic (or legendary bonus),
// cash, healing, Chronos energy, and a protective bullet-clear shockwave.
func (w *World) OpenChest(idx int) bool {
	if idx < 0 || idx >= len(w.Chests) {
		return false
	}
	ch := &w.Chests[idx]
	if ch.Opened || w.Player.BossKeys <= 0 {
		return false
	}

	w.Player.BossKeys--
	ch.Opened = true
	ch.OpenAnimTimer = chestOpenAnimSeconds
	w.ChestsOpened++

	// Protective shockwave clears nearby bullets on chest unlock
	w.Bullets.ClearRadius(ch.X, ch.Y, chestClearRadius)
	w.addWaveRing(ch.X, ch.Y, chestClearRadius, 1)
	if w.ScreenShake < 5.5 {
		w.ScreenShake = 5.5
	}

	// Base reward bundle: Cash, Score, +2 HP, +40 Chronos
	cashGain := int(math.Round(chestBaseCashReward * float64(w.Player.CashBonusMult())))
	w.Player.Cash += cashGain
	w.Score += 500
	w.Player.Heal(2)
	w.Player.HealFlashTimer = 0.45
	w.Player.ChronosMeter = clamp32(w.Player.ChronosMeter+40.0, 0, w.Player.ChronosMax)

	if rk, ok := w.grantRandomUnownedRelic(); ok {
		spec := GetRelicSpec(rk)
		w.addPopup(ch.X, ch.Y-32, "BAULE: "+spec.Name+"!", 1)
		w.addPopup(ch.X, ch.Y-18, fmt.Sprintf("+$%d  +2 HP  +40 CHRONOS", cashGain), 3)
		w.emitEvent(EventChestOpened, int(rk), false)
	} else {
		// All 6 relics already owned: legendary vitality + Upgrade choice
		w.Player.MaxHP++
		w.Player.Heal(2)
		w.addPopup(ch.X, ch.Y-32, "TESORO LEGGENDARIO! +1 MAX HP!", 1)
		w.addPopup(ch.X, ch.Y-18, fmt.Sprintf("+$%d  +UPGRADE!", cashGain), 3)
		w.emitEvent(EventChestOpened, -1, true)
		w.triggerLevelUp()
	}

	if w.ClosedChestCount() == 0 && w.Player.BossKeys > 0 {
		w.chestRestockTimer = chestRestockDelaySec
	}
	return true
}

func (w *World) updateChests(dt float32, interactPressed bool) {
	if len(w.Chests) == 0 {
		return
	}

	for i := range w.Chests {
		ch := &w.Chests[i]
		ch.BobPhase += dt * 3.8
		if ch.OpenAnimTimer > 0 {
			ch.OpenAnimTimer -= dt
			if ch.OpenAnimTimer < 0 {
				ch.OpenAnimTimer = 0
			}
		}
	}

	if w.chestRestockTimer > 0 {
		w.chestRestockTimer -= dt
		if w.chestRestockTimer <= 0 {
			w.chestRestockTimer = 0
			if w.ClosedChestCount() == 0 && w.Player.BossKeys > 0 {
				w.restockChests()
			}
		}
	}

	if w.Player.BossKeys <= 0 {
		return
	}

	for i := range w.Chests {
		ch := &w.Chests[i]
		if ch.Opened {
			continue
		}
		rad := ch.Radius
		if rad <= 0 {
			rad = chestInteractRadius
		}
		if interactPressed && chestKeyPressRadius > rad {
			rad = chestKeyPressRadius
		}
		if distSq(w.Player.X, w.Player.Y, ch.X, ch.Y) <= rad*rad {
			w.OpenChest(i)
			return
		}
	}
}
