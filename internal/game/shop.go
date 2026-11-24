package game

import "fmt"

// Slot machine symbol constants.
const (
	SymCherry  = 0
	SymCash    = 1
	SymXP      = 2
	SymHeart   = 3
	SymGoldBus = 4
	SymSiren   = 5
	SymCount   = 6
)

// SlotSymbolName returns a compact arcade display label for each slot symbol.
func SlotSymbolName(sym int) string {
	switch sym {
	case SymCherry:
		return "CHERRY"
	case SymCash:
		return "$$$"
	case SymXP:
		return "XP*"
	case SymHeart:
		return "CUORE"
	case SymGoldBus:
		return "777"
	case SymSiren:
		return "SIRENA"
	default:
		return "?"
	}
}

// ShopMenuState holds the modal state when interacting with a VendingShop.
type ShopMenuState struct {
	Active    bool
	ShopIndex int
	Message   string
}

// SlotMachineState holds the modal and animated 3-reel state for Rizz-Slots 777.
type SlotMachineState struct {
	Active       bool
	MachineIndex int
	Spinning     bool
	SpinTimer    float32
	SpinDuration float32
	Reels        [3]int
	TargetReels  [3]int
	LastTickStep int
	ResultTitle  string
	ResultDesc   string
	ResultKind   int // 0=neutral, 1=win, 2=jackpot, 3=hazard
}

// EspressoHealCost is the standard Vending Shop Caffe Espresso healing price.
const EspressoHealCost = 35

// CurrentHealCost returns the Espresso healing price for the active Vending Shop.
func (w *World) CurrentHealCost() int {
	if w.ShopMenu.ShopIndex >= 0 && w.ShopMenu.ShopIndex < len(w.Shops) && w.Shops[w.ShopMenu.ShopIndex].HealCost > 0 {
		return w.Shops[w.ShopMenu.ShopIndex].HealCost
	}
	return EspressoHealCost
}

// CurrentBribeCost returns the current police bribe price based on Wanted Stars.
func (w *World) CurrentBribeCost() int {
	if w.Police.Stars >= 4 {
		return 180
	}
	return 100
}

// CurrentSlotCost returns the spin cost of the active slot machine (scales slightly per use).
func (w *World) CurrentSlotCost() int {
	idx := w.SlotModal.MachineIndex
	if idx < 0 || idx >= len(w.SlotMachines) {
		return 35
	}
	return w.SlotMachines[idx].SpinCost
}

// OpenShop activates the Vending Shop mini-menu paused cleanly over gameplay.
func (w *World) OpenShop(shopIdx int) {
	if shopIdx < 0 || shopIdx >= len(w.Shops) {
		return
	}
	w.ShopMenu = ShopMenuState{
		Active:    true,
		ShopIndex: shopIdx,
		Message:   "SCEGLI UN'OFFERTA NEON",
	}
}

// HandleShopInput processes choices [1..4] or cancel/exit while the Vending Shop modal is open.
func (w *World) HandleShopInput(in InputState) {
	if in.CancelJustPressed {
		w.ShopMenu.Active = false
		return
	}

	if in.ShopAction == ShopActionNone {
		return
	}

	switch in.ShopAction {
	case ShopActionHeal:
		cost := w.CurrentHealCost()
		if w.Player.Cash < cost {
			w.ShopMenu.Message = fmt.Sprintf("SERVONO $%d!", cost)
			return
		}
		w.Player.Cash -= cost
		w.Player.Heal(2)
		w.Player.ChronosMeter = w.Player.ChronosMax
		w.TutUsedSlotOrShop = true
		w.ShopMenu.Message = "CAFFE ESPRESSO! +2 HP & CHRONOS FULL!"
		w.addPopup(w.Player.X, w.Player.Y-28, "ESPRESSO +2 HP & CHRONOS!", 3)
		w.emitEvent(EventShopBuy, ShopActionHeal, false)

	case ShopActionBribe:
		if w.Police.Stars <= 0 {
			w.ShopMenu.Message = "SEI GIA PULITO (0* WANTED)!"
			return
		}
		cost := w.CurrentBribeCost()
		if w.Player.Cash < cost {
			w.ShopMenu.Message = fmt.Sprintf("SERVONO $%d PER LA TANGENTE!", cost)
			return
		}
		w.Player.Cash -= cost
		w.Police.ReduceStar(2)
		w.TutUsedSlotOrShop = true
		w.ShopMenu.Message = fmt.Sprintf("TANGENTE PAGATA! -2* WANTED (-$%d)", cost)
		w.addPopup(w.Player.X, w.Player.Y-28, fmt.Sprintf("TANGENTE -$%d (-2*)", cost), 3)
		w.emitEvent(EventShopBuy, ShopActionBribe, false)

	case ShopActionRelic:
		const cost = 120
		if w.Player.Cash < cost {
			w.ShopMenu.Message = fmt.Sprintf("SERVONO $%d PER LA RELIQUIA!", cost)
			return
		}
		var unowned [RelicCount]RelicKind
		count := 0
		for i := 0; i < int(RelicCount); i++ {
			if !w.Player.Relics[i] {
				unowned[count] = RelicKind(i)
				count++
			}
		}
		w.Player.Cash -= cost
		w.TutUsedSlotOrShop = true
		if count > 0 {
			picked := unowned[w.rng.Intn(count)]
			w.Player.Relics[picked] = true
			spec := GetRelicSpec(picked)
			if picked == RelicFakeVIPPass {
				w.Player.MaxMultiHook = clampInt(w.Player.MaxMultiHook+1, 1, 12)
			}
			w.ShopMenu.Message = "RELIQUIA: " + spec.Name + "!"
			w.addPopup(w.Player.X, w.Player.Y-32, spec.Name+"!", 1)
		} else {
			// All 6 relics already owned: bonus vitality boost
			w.Player.MaxHP++
			w.Player.Heal(2)
			w.ShopMenu.Message = "TUTTE LE RELIQUIE! +1 MAX HP & +2 HP!"
			w.addPopup(w.Player.X, w.Player.Y-32, "+1 MAX HP BONUS!", 1)
		}
		w.emitEvent(EventShopBuy, ShopActionRelic, false)

	case ShopActionUpgrade:
		const cost = 140
		if w.Player.Cash < cost {
			w.ShopMenu.Message = fmt.Sprintf("SERVONO $%d PER LA CASSA UPGRADE!", cost)
			return
		}
		w.Player.Cash -= cost
		w.TutUsedSlotOrShop = true
		w.ShopMenu.Active = false
		w.emitEvent(EventShopBuy, ShopActionUpgrade, false)
		w.triggerLevelUp()
	}
}

// OpenSlotMachine opens the interactive Rizz-Slots 777 modal paused cleanly over gameplay.
func (w *World) OpenSlotMachine(machineIdx int) {
	if machineIdx < 0 || machineIdx >= len(w.SlotMachines) {
		return
	}
	w.SlotModal = SlotMachineState{
		Active:       true,
		MachineIndex: machineIdx,
		Reels:        [3]int{SymGoldBus, SymCash, SymGoldBus},
		TargetReels:  [3]int{SymGoldBus, SymCash, SymGoldBus},
		ResultTitle:  "RIZZ-SLOTS 777",
		ResultDesc:   "PREMI [SPAZIO / GIRA] PER TENTARE LA FORTUNA!",
		ResultKind:   0,
	}
}

// StartSlotSpin begins an animated 3-reel spin on the active Slot Machine.
func (w *World) StartSlotSpin() bool {
	if !w.SlotModal.Active || w.SlotModal.Spinning {
		return false
	}
	idx := w.SlotModal.MachineIndex
	if idx < 0 || idx >= len(w.SlotMachines) {
		return false
	}
	sm := &w.SlotMachines[idx]
	if w.Player.Cash < sm.SpinCost {
		w.SlotModal.ResultTitle = "FONDI INSUFFICIENTI!"
		w.SlotModal.ResultDesc = fmt.Sprintf("SERVONO $%d PER GIRARE!", sm.SpinCost)
		w.SlotModal.ResultKind = 3
		return false
	}

	w.Player.Cash -= sm.SpinCost
	sm.SpinsUsed++
	w.TutUsedSlotOrShop = true
	if sm.SpinCost < 95 {
		sm.SpinCost += 10
		if sm.SpinCost > 95 {
			sm.SpinCost = 95
		}
	}

	lucky := w.Player.HasRelic(RelicLuckySocks)
	w.SlotModal.TargetReels = rollSlotOutcome(w.rng, lucky)
	w.SlotModal.Spinning = true
	w.SlotModal.SpinTimer = 0
	w.SlotModal.SpinDuration = 1.15
	w.SlotModal.LastTickStep = -1
	w.SlotModal.ResultTitle = "RULLI IN CORSA..."
	w.SlotModal.ResultDesc = "INCROCIA LE DITA!"
	w.SlotModal.ResultKind = 0
	w.emitEvent(EventSlotTick, 0, false)
	return true
}

// UpdateSlotMachine advances the animated 3-reel spin or handles spin/exit inputs.
func (w *World) UpdateSlotMachine(dt float32, in InputState) {
	if !w.SlotModal.Active {
		return
	}
	if w.SlotModal.Spinning {
		w.SlotModal.SpinTimer += dt
		t := w.SlotModal.SpinTimer
		tickStep := int(t * 18.0)
		if tickStep != w.SlotModal.LastTickStep {
			w.SlotModal.LastTickStep = tickStep
			w.emitEvent(EventSlotTick, tickStep, false)
		}

		// Staggered reel locking: Reel 0 @ 0.40s, Reel 1 @ 0.75s, Reel 2 @ 1.10s
		lockTimes := [3]float32{0.40, 0.75, 1.10}
		for r := 0; r < 3; r++ {
			if t >= lockTimes[r] {
				w.SlotModal.Reels[r] = w.SlotModal.TargetReels[r]
			} else {
				w.SlotModal.Reels[r] = (tickStep + r*2) % SymCount
			}
		}

		if t >= w.SlotModal.SpinDuration {
			w.SlotModal.Spinning = false
			w.SlotModal.Reels = w.SlotModal.TargetReels
			w.resolveSlotPayout()
		}
		return
	}

	if in.CancelJustPressed {
		w.SlotModal.Active = false
		return
	}
	if in.SlotSpinJustPressed {
		w.StartSlotSpin()
	}
}

func rollSlotOutcome(rng interface{ Intn(int) int; Float32() float32 }, lucky bool) [3]int {
	// Weighted probabilities so spins feel punchy, fair, and exciting.
	roll := rng.Float32()
	jackpotChance := float32(0.06)
	tripleChance := float32(0.24)
	pairChance := float32(0.62)
	if lucky {
		jackpotChance = 0.11
		tripleChance = 0.34
		pairChance = 0.76
	}

	if roll < jackpotChance {
		return [3]int{SymGoldBus, SymGoldBus, SymGoldBus}
	}
	if roll < tripleChance {
		others := [5]int{SymCherry, SymCash, SymXP, SymHeart, SymSiren}
		s := others[rng.Intn(len(others))]
		return [3]int{s, s, s}
	}
	if roll < pairChance {
		a := rng.Intn(SymCount)
		b := rng.Intn(SymCount)
		for b == a {
			b = rng.Intn(SymCount)
		}
		pos := rng.Intn(3)
		out := [3]int{a, a, a}
		out[pos] = b
		return out
	}

	// 3 distinct symbols
	r0 := rng.Intn(SymCount)
	r1 := (r0 + 1 + rng.Intn(SymCount-1)) % SymCount
	r2 := (r1 + 1 + rng.Intn(SymCount-2)) % SymCount
	if r2 == r0 {
		r2 = (r2 + 1) % SymCount
	}
	return [3]int{r0, r1, r2}
}

func (w *World) resolveSlotPayout() {
	r := w.SlotModal.Reels
	if r[0] == r[1] && r[1] == r[2] {
		switch r[0] {
		case SymGoldBus:
			w.Player.Cash += 250
			w.Score += 1500
			w.Player.Heal(w.Player.MaxHP)
			w.Player.ChronosMeter = w.Player.ChronosMax
			w.SlotModal.ResultTitle = "JACKPOT 777!! AUTOBUS D'ORO!"
			w.SlotModal.ResultDesc = "+$250  +120 XP  FULL HP & CHRONOS!"
			w.SlotModal.ResultKind = 2
			w.addPopup(w.Player.X, w.Player.Y-36, "JACKPOT 777! +$250 +120XP!", 1)
			w.emitEvent(EventSlotJackpot, 0, true)
			if w.Player.AddXP(120) {
				w.SlotModal.Active = false
				w.triggerLevelUp()
			}
			return
		case SymCash:
			w.Player.Cash += 140
			w.Score += 400
			w.SlotModal.ResultTitle = "TRIS DI DOLLARI!"
			w.SlotModal.ResultDesc = "+$140 IN TASCA!"
			w.SlotModal.ResultKind = 1
			w.addPopup(w.Player.X, w.Player.Y-30, "SLOT TRIS! +$140!", 3)
			w.emitEvent(EventSlotWin, SymCash, false)
			return
		case SymXP:
			w.Score += 450
			w.SlotModal.ResultTitle = "TRIS STELLARE XP!"
			w.SlotModal.ResultDesc = "+110 XP ISTANTANEI!"
			w.SlotModal.ResultKind = 1
			w.addPopup(w.Player.X, w.Player.Y-30, "SLOT TRIS! +110 XP!", 1)
			w.emitEvent(EventSlotWin, SymXP, false)
			if w.Player.AddXP(110) {
				w.SlotModal.Active = false
				w.triggerLevelUp()
			}
			return
		case SymHeart:
			if w.Player.HP >= w.Player.MaxHP {
				w.Player.MaxHP++
			}
			w.Player.Heal(3)
			w.Player.ChronosMeter = w.Player.ChronosMax
			w.SlotModal.ResultTitle = "TRIS DI CUORI!"
			w.SlotModal.ResultDesc = "+3 HP (O +1 MAX HP) & CHRONOS FULL!"
			w.SlotModal.ResultKind = 1
			w.addPopup(w.Player.X, w.Player.Y-30, "SLOT CUORI! +3 HP!", 1)
			w.emitEvent(EventSlotWin, SymHeart, false)
			return
		case SymCherry:
			w.Player.Cash += 90
			w.Player.StopCooldownTimer = 0
			w.SlotModal.ResultTitle = "TRIS CILIEGIE!"
			w.SlotModal.ResultDesc = "+$90 & COOLDOWN FERMATA AZZERATO!"
			w.SlotModal.ResultKind = 1
			w.addPopup(w.Player.X, w.Player.Y-30, "SLOT +$90 & RESET CD!", 1)
			w.emitEvent(EventSlotWin, SymCherry, false)
			return
		case SymSiren:
			w.Player.Cash += 110
			w.Police.AddStar(1)
			w.SlotModal.ResultTitle = "SOLDI SPORCHI! (+1* WANTED)"
			w.SlotModal.ResultDesc = "+$110 MA SCATTA L'ALLARME POLIZIA!"
			w.SlotModal.ResultKind = 3
			w.addPopup(w.Player.X, w.Player.Y-30, "SOLDI SPORCHI +$110 (+1*)", 2)
			w.emitEvent(EventPoliceStarUp, w.Police.Stars, false)
			w.emitEvent(EventSlotWin, SymSiren, false)
			return
		}
	}

	// Check 2 matching symbols
	pairSym := -1
	if r[0] == r[1] || r[0] == r[2] {
		pairSym = r[0]
	} else if r[1] == r[2] {
		pairSym = r[1]
	}

	if pairSym != -1 {
		switch pairSym {
		case SymGoldBus, SymCash:
			w.Player.Cash += 65
			w.SlotModal.ResultTitle = "COPPIA VINCENTE!"
			w.SlotModal.ResultDesc = "+$65 DI VINCITA!"
			w.addPopup(w.Player.X, w.Player.Y-28, "SLOT +$65!", 3)
		case SymXP:
			w.SlotModal.ResultTitle = "COPPIA XP!"
			w.SlotModal.ResultDesc = "+50 XP BONUS!"
			w.addPopup(w.Player.X, w.Player.Y-28, "SLOT +50 XP!", 1)
			if w.Player.AddXP(50) {
				w.SlotModal.Active = false
				w.triggerLevelUp()
			}
		case SymHeart:
			w.Player.Heal(1)
			w.Player.Cash += 30
			w.SlotModal.ResultTitle = "COPPIA CUORI!"
			w.SlotModal.ResultDesc = "+1 HP & +$30!"
			w.addPopup(w.Player.X, w.Player.Y-28, "SLOT +1 HP +$30!", 1)
		default:
			w.Player.ChronosMeter = w.Player.ChronosMax
			w.Player.StopCooldownTimer = 0
			w.SlotModal.ResultTitle = "COPPIA TURBO!"
			w.SlotModal.ResultDesc = "CHRONOS FULL & RESET COOLDOWN!"
			w.addPopup(w.Player.X, w.Player.Y-28, "SLOT CHRONOS & RESET CD!", 1)
		}
		w.SlotModal.ResultKind = 1
		w.emitEvent(EventSlotWin, pairSym, false)
		return
	}

	// Consolation prize
	w.SlotModal.ResultTitle = "RITENTA, SARAI PIU RIZZ!"
	w.SlotModal.ResultDesc = "PREMIO CONSOLAZIONE: +15 XP"
	w.SlotModal.ResultKind = 0
	w.emitEvent(EventSlotBust, 0, false)
	if w.Player.AddXP(15) {
		w.SlotModal.Active = false
		w.triggerLevelUp()
	}
}
