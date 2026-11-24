package game

import (
	"math"
	"testing"
)

func TestUpgradePoolTwelveCardsAndSmartFiltering(t *testing.T) {
	if int(UpgradeCount) != 12 || len(allUpgradeCards) != 12 {
		t.Fatalf("expected 12 unique upgrade cards matching UpgradeCount=%d, got %d", UpgradeCount, len(allUpgradeCards))
	}
	seen := make(map[UpgradeKind]bool, UpgradeCount)
	for _, c := range allUpgradeCards {
		if c.Title == "" || c.Desc == "" {
			t.Fatalf("upgrade card %+v has empty title or desc", c)
		}
		if seen[c.Kind] {
			t.Fatalf("duplicate UpgradeKind %d in allUpgradeCards", c.Kind)
		}
		seen[c.Kind] = true
	}

	// Max out 5 capped cards -> none of them may appear across 40 Level-Up rolls
	w := NewWorldWithSeed(1001)
	w.Player.StopCooldownBase = 1.5
	w.Player.MaxMultiHook = 12
	w.Player.MaxSimultaneousStops = 4
	w.Player.RangeMult = 1.60
	w.Player.UpgradeRanks[UpgradeDashKatana] = 3

	banned := map[UpgradeKind]bool{
		UpgradeCooldown:   true,
		UpgradeMultiHook:  true,
		UpgradeMaxStops:   true,
		UpgradeRange:      true,
		UpgradeDashKatana: true,
	}
	for roll := 0; roll < 40; roll++ {
		w.openLevelUpModal()
		var pickedInRoll [3]UpgradeKind
		for idx, card := range w.UpgradeChoices {
			if banned[card.Kind] {
				t.Fatalf("roll %d offered maxed-out upgrade %+v", roll, card)
			}
			for prev := 0; prev < idx; prev++ {
				if pickedInRoll[prev] == card.Kind {
					t.Fatalf("roll %d offered duplicate card kind %d", roll, card.Kind)
				}
			}
			pickedInRoll[idx] = card.Kind
		}
	}

	// Max out ALL capped cards -> openLevelUpModal still fills 3 distinct fallback cards cleanly
	for _, c := range allUpgradeCards {
		if c.MaxRank > 0 {
			w.Player.UpgradeRanks[c.Kind] = c.MaxRank
		}
	}
	w.openLevelUpModal()
	for idx, card := range w.UpgradeChoices {
		if card.Title == "" {
			t.Fatalf("fallback choice %d is empty", idx)
		}
	}
}

func TestNewUpgradeDashKatanaAndGlassFortress(t *testing.T) {
	w := NewWorldWithSeed(1002)
	w.Girls = nil

	// 1. UpgradeDashKatana: shorter Dash CD + wider parry + higher Chronos & CD shave
	w.applyUpgrade(UpgradeCard{Kind: UpgradeDashKatana})
	if math.Abs(float64(w.Player.EffectiveDashCooldown()-0.72*0.80)) > 0.005 {
		t.Fatalf("expected EffectiveDashCooldown ~0.576, got %v", w.Player.EffectiveDashCooldown())
	}
	w.Player.StopCooldownTimer = 4.0
	w.Player.ChronosMeter = 20.0
	w.Player.DashTimer = 0.14
	w.Bullets.Spawn(Bullet{X: w.Player.X + 20, Y: w.Player.Y, Radius: 4.0, MaxAge: 5.0, Kind: BulletPolice})
	w.handleBulletInteractions()
	if w.Bullets.Count != 0 {
		t.Fatalf("expected expanded Dash-Parry radius to catch bullet at 20px, remaining=%d", w.Bullets.Count)
	}
	if w.Player.ChronosMeter != 36.0 {
		t.Fatalf("expected 20 + 16 = 36 Chronos on Dash-Parry, got %v", w.Player.ChronosMeter)
	}
	if math.Abs(float64(w.Player.StopCooldownTimer-2.65)) > 0.01 {
		t.Fatalf("expected -1.35s shave on StopCooldownTimer (4.0 -> 2.65), got %v", w.Player.StopCooldownTimer)
	}

	// 2. UpgradeGlassFortress: +2 shield to active & new stops + Contro-EMP on shield break
	w.Stops = append(w.Stops, BusStop{ID: 1, Active: true, X: 500, Y: 360, CaptureRadius: 90, ShieldHP: 2, MaxShieldHP: 2, LifeTimer: 10})
	w.applyUpgrade(UpgradeCard{Kind: UpgradeGlassFortress})
	if w.Stops[0].ShieldHP != 4 || w.Stops[0].MaxShieldHP != 4 {
		t.Fatalf("expected active stop to gain +2 ShieldHP (2 -> 4), got %d/%d", w.Stops[0].ShieldHP, w.Stops[0].MaxShieldHP)
	}
	w.Stops[0].ShieldHP = 1
	w.HitStopTimer = 0
	w.Bullets.Spawn(Bullet{X: 500, Y: 360, Radius: 5, MaxAge: 5, Kind: BulletPolice, AntiStop: true})
	w.Bullets.Spawn(Bullet{X: 560, Y: 360, VX: 0, VY: 0, Radius: 4, MaxAge: 5, Kind: BulletRazor})
	w.Update(baseTestInput(), 0.016)
	if w.Bullets.Count != 0 {
		t.Fatalf("expected Contro-EMP to clear nearby bullets when glass shield broke, remaining=%d", w.Bullets.Count)
	}
}

func TestNewUpgradeMinigameTurboAndBasslineShock(t *testing.T) {
	w := NewWorldWithSeed(1003)
	w.Girls = nil

	// 1. UpgradeMinigameTurbo: +22% speed, +2.5s stop LifeTimer, +1.5s Combo timer
	w.applyUpgrade(UpgradeCard{Kind: UpgradeMinigameTurbo})
	if math.Abs(float64(w.Player.MinigameSpeedMult()-1.22)) > 0.005 {
		t.Fatalf("expected MinigameSpeedMult=1.22, got %v", w.Player.MinigameSpeedMult())
	}
	w.tryThrowBusStop(w.Player.X+40, w.Player.Y)
	if len(w.Stops) != 1 || math.Abs(float64(w.Stops[0].MaxLifeTimer-16.5)) > 0.01 {
		t.Fatalf("expected MaxLifeTimer=16.5 with MinigameTurbo, got %+v", w.Stops)
	}
	w.Stops[0].Minigame.Completed = true
	w.updateStops(0.016, ActionNone, false)
	if !w.Stops[0].Combo.Active || math.Abs(float64(w.Stops[0].Combo.MaxTimer-5.0)) > 0.01 {
		t.Fatalf("expected Combo MaxTimer=5.0 (3.5 + 1.5), got %v", w.Stops[0].Combo.MaxTimer)
	}

	// 2. UpgradeBasslineShock: faster period + bullet clear pulse
	w.applyUpgrade(UpgradeCard{Kind: UpgradeBasslineShock})
	if w.Player.EffectiveBasslinePeriod() != 2.0 {
		t.Fatalf("expected EffectiveBasslinePeriod=2.0, got %v", w.Player.EffectiveBasslinePeriod())
	}
	w.Bullets.Spawn(Bullet{X: w.Player.X + 45, Y: w.Player.Y, Radius: 4, MaxAge: 5, Kind: BulletTear})
	w.triggerBasslinePulse()
	if w.Bullets.Count != 0 {
		t.Fatalf("expected Bassline Shock to clear bullets inside 58px radius, remaining=%d", w.Bullets.Count)
	}
}

func TestNewUpgradePerfectNovaAndRizzMagnet(t *testing.T) {
	w := NewWorldWithSeed(1004)
	w.Girls = []Girl{NewGirl(7001, w.Player.X+60, w.Player.Y, ArchEmo, false, w.rng)}

	// 1. UpgradePerfectNova: expands Perfect radius (+8px), clearR (+45px), +15 Chronos, +$12 Cash
	w.applyUpgrade(UpgradeCard{Kind: UpgradePerfectNova})
	if w.Player.EffectivePerfectRadius() != PerfectHitRadius+8.0 {
		t.Fatalf("expected EffectivePerfectRadius=%v, got %v", PerfectHitRadius+8.0, w.Player.EffectivePerfectRadius())
	}
	w.Player.ChronosMeter = 50
	w.Player.Cash = 10
	w.Bullets.Spawn(Bullet{X: w.Girls[0].X + 135, Y: w.Girls[0].Y, Radius: 4, MaxAge: 5, Kind: BulletTear}) // within 130+45=175
	w.tryThrowBusStop(w.Girls[0].X-28, w.Girls[0].Y)                                                        // 28px <= 32px -> Perfect!
	if len(w.Stops) != 1 || !w.Stops[0].PerfectDrop {
		t.Fatalf("expected 28px drop to count as Perfect Drop with PerfectNova")
	}
	if w.Bullets.Count != 0 {
		t.Fatalf("expected expanded 175px clear wave to clear bullet at 163px, remaining=%d", w.Bullets.Count)
	}
	if w.Player.ChronosMeter != 65 || w.Player.Cash != 22 {
		t.Fatalf("expected +15 Chronos (65) and +$12 Cash (22), got Chronos=%v Cash=%d", w.Player.ChronosMeter, w.Player.Cash)
	}

	// 2. UpgradeRizzMagnet: +20% Cash/XP, +2.5s Streak window, +$2 per bus-cleared bullet
	w.applyUpgrade(UpgradeCard{Kind: UpgradeRizzMagnet})
	if math.Abs(float64(w.Player.CashBonusMult()-1.20)) > 0.005 || math.Abs(float64(w.Player.XPBonusMult()-1.20)) > 0.005 {
		t.Fatalf("expected 1.20x Cash & XP mults, got cash=%v xp=%v", w.Player.CashBonusMult(), w.Player.XPBonusMult())
	}
	if w.Player.EffectiveStreakWindow() != StreakWindowDuration+2.5 {
		t.Fatalf("expected Streak window %v, got %v", StreakWindowDuration+2.5, w.Player.EffectiveStreakWindow())
	}
	cashBeforeBus := w.Player.Cash
	w.Bullets.Count = 0
	w.Bullets.Spawn(Bullet{X: 350, Y: 360, Radius: 4, MaxAge: 5, Kind: BulletPolice})
	w.Bullets.Spawn(Bullet{X: 365, Y: 360, Radius: 4, MaxAge: 5, Kind: BulletPolice})
	w.Buses = append(w.Buses, NeonBus{Active: true, X: 300, Y: 360, Speed: 900, TargetStopID: 999})
	w.updateBuses(0.1)
	if w.Player.Cash != cashBeforeBus+4 {
		t.Fatalf("expected +$4 from Biglietto d'Oro bus bullet clear (2 bullets * $2), got +%d", w.Player.Cash-cashBeforeBus)
	}
}
