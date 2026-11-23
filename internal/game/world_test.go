package game

import (
	"math/rand"
	"testing"
)

func TestBulletPoolZeroGCAndShield(t *testing.T) {
	var pool BulletPool
	pool.Spawn(Bullet{X: 100, Y: 100, VX: 0, VY: 100, Radius: 4, MaxAge: 5, Kind: BulletProtractor})
	pool.Spawn(Bullet{X: 300, Y: 300, VX: 50, VY: 0, Radius: 4, MaxAge: 5, Kind: BulletFlash})
	if pool.Count != 2 {
		t.Fatalf("expected 2 bullets, got %d", pool.Count)
	}

	// Place a BusStop whose back-glass shield sits around (100, 100) [Y in stop.Y-28..stop.Y-4]
	stops := []BusStop{
		{ID: 1, X: 100, Y: 118, Active: true, CaptureRadius: 90, ShieldHP: 3, MaxShieldHP: 3},
	}
	pool.Update(0.05, 1.0, 500, 500, stops)
	if stops[0].ShieldHP != 2 {
		t.Fatalf("expected back-glass shield to absorb bullet (ShieldHP=2), got %d", stops[0].ShieldHP)
	}
	if pool.Count != 1 {
		t.Fatalf("expected 1 remaining bullet after shield block, got %d", pool.Count)
	}

	cleared := pool.ClearRadius(300, 300, 30)
	if cleared != 1 || pool.Count != 0 {
		t.Fatalf("expected ClearRadius to clear remaining bullet, cleared=%d count=%d", cleared, pool.Count)
	}
}

func TestPoliceMultipliersAndBribe(t *testing.T) {
	w := NewWorldWithSeed(42)
	expected := [6][2]float32{
		{1.0, 1.0},
		{1.3, 1.4},
		{1.6, 1.8},
		{2.0, 2.3},
		{2.5, 3.0},
		{3.5, 4.0},
	}
	for stars := 0; stars <= 5; stars++ {
		w.Police.Stars = stars
		xpM, cashM := w.Police.XPAndCashMultiplier()
		if xpM != expected[stars][0] || cashM != expected[stars][1] {
			t.Fatalf("stars=%d: got (%v, %v), want (%v, %v)", stars, xpM, cashM, expected[stars][0], expected[stars][1])
		}
	}

	// Test VendingShop mini-menu bribe (5* costs $180, 3* costs $100)
	w.Player.Cash = 280
	w.Player.X = w.Shops[0].X
	w.Player.Y = w.Shops[0].Y
	w.Update(InputState{InteractJustPressed: true, SelectArchetype: -1, ActionDir: -1, ChooseUpgrade: -1}, 0.016)
	if !w.ShopMenu.Active {
		t.Fatalf("expected ShopMenu to open when interacting near VendingShop")
	}
	w.Update(InputState{ShopAction: ShopActionBribe, SelectArchetype: -1, ActionDir: -1, ChooseUpgrade: -1}, 0.016)
	if w.Police.Stars != 3 || w.Player.Cash != 100 {
		t.Fatalf("expected 5* bribe to reduce stars to 3 and cash to 100, got stars=%d cash=%d", w.Police.Stars, w.Player.Cash)
	}
	w.Update(InputState{ShopAction: ShopActionBribe, SelectArchetype: -1, ActionDir: -1, ChooseUpgrade: -1}, 0.016)
	if w.Police.Stars != 1 || w.Player.Cash != 0 {
		t.Fatalf("expected 3* bribe to reduce stars to 1 and cash to 0, got stars=%d cash=%d", w.Police.Stars, w.Player.Cash)
	}
	w.Update(InputState{CancelJustPressed: true, SelectArchetype: -1, ActionDir: -1, ChooseUpgrade: -1}, 0.016)
	if w.ShopMenu.Active {
		t.Fatalf("expected ShopMenu to close on CancelJustPressed")
	}
}

func TestMinigamesAndBoardingBusRush(t *testing.T) {
	rng := rand.New(rand.NewSource(7))

	// 1. Pragmata Grid persistence when Alfredo leaves radius
	mg := NewMinigameState(ArchStudent, false, rng)
	mg.Update(0.016, true, ActionRight, false, 108, rng)
	if mg.GridX != 1 || mg.GridY != 0 {
		t.Fatalf("expected Pragmata cursor at (1,0), got (%d,%d)", mg.GridX, mg.GridY)
	}
	mg.Update(0.1, false, ActionLeft, false, 108, rng) // Outside radius after leaving: cursor must persist!
	if mg.GridX != 1 || mg.GridY != 0 {
		t.Fatalf("expected Pragmata cursor to persist at (1,0) while dashing outside, got (%d,%d)", mg.GridX, mg.GridY)
	}
	// Navigate to (3,3)
	mg.Update(0.016, true, ActionRight, false, 108, rng)
	mg.Update(0.016, true, ActionRight, false, 108, rng)
	mg.Update(0.016, true, ActionDown, false, 108, rng)
	mg.Update(0.016, true, ActionDown, false, 108, rng)
	mg.Update(0.016, true, ActionDown, false, 108, rng)
	if !mg.Completed {
		t.Fatalf("expected Pragmata grid to complete at (3,3), pos=(%d,%d)", mg.GridX, mg.GridY)
	}

	// 2. Full World Throw -> Perfect Drop -> Minigame -> 4-Step Combo -> Neon Bus Rush -> Boarded!
	w := NewWorldWithSeed(99)
	w.Girls = w.Girls[:1]
	w.Girls[0] = NewGirl(1, w.Player.X+60, w.Player.Y, ArchKawaii, true, w.rng)
	w.Player.SelectedStopArch = ArchKawaii

	// Throw directly on top of girl for Perfect Drop
	w.Update(InputState{
		ThrowJustPressed: true,
		AimWorldX:        w.Girls[0].X,
		AimWorldY:        w.Girls[0].Y,
		SelectArchetype:  -1,
		ActionDir:        -1,
		ChooseUpgrade:    -1,
	}, 0.016)

	if len(w.Stops) != 1 || !w.Stops[0].PerfectDrop {
		t.Fatalf("expected 1 PerfectDrop stop, len=%d", len(w.Stops))
	}
	if w.Police.Stars != 1 {
		t.Fatalf("expected VIP hook to trigger +1★ Police Wanted, got %d", w.Police.Stars)
	}

	// Clear HitStop
	for w.HitStopTimer > 0 {
		w.Update(InputState{SelectArchetype: -1, ActionDir: -1, ChooseUpgrade: -1}, 0.02)
	}

	// Move Alfredo into the Kawaii stop and complete Tear-Graze
	w.Player.X = w.Stops[0].X
	w.Player.Y = w.Stops[0].Y
	for i := 0; i < 60 && !w.Stops[0].Minigame.Completed; i++ {
		dir := ActionNone
		if i%16 == 0 {
			dir = ActionUp
		}
		w.Update(InputState{SelectArchetype: -1, ActionDir: dir, ChooseUpgrade: -1}, 0.05)
	}
	if !w.Stops[0].Minigame.Completed || !w.Stops[0].Combo.Active {
		t.Fatalf("expected Kawaii minigame to complete and activate 4-step combo")
	}

	// Execute 4-Step Boarding Combo
	for step := 0; step < 4; step++ {
		dir := w.Stops[0].Combo.Sequence[step]
		w.Update(InputState{SelectArchetype: -1, ActionDir: dir, ChooseUpgrade: -1}, 0.016)
	}
	if !w.Stops[0].BusRushTriggered || len(w.Buses) != 1 {
		t.Fatalf("expected Neon Bus Rush to trigger on 4-step combo completion")
	}

	// Advance simulation until NeonBus crosses stop and boards the VIP girl
	for i := 0; i < 60 && w.GirlsBoarded == 0; i++ {
		w.Update(InputState{SelectArchetype: -1, ActionDir: -1, ChooseUpgrade: -1}, 0.03)
	}
	if w.GirlsBoarded != 1 {
		t.Fatalf("expected 1 VIP girl boarded, got %d", w.GirlsBoarded)
	}
	if w.Player.Cash <= 0 {
		t.Fatalf("expected Cash reward after boarding girl, got %d", w.Player.Cash)
	}
}

func TestBossThreeStopBreakAndVictory(t *testing.T) {
	w := NewWorldWithSeed(123)
	w.ForceSpawnBoss(BossMom)
	if !w.Boss.Active || w.Boss.Shields != 3 {
		t.Fatalf("expected active BossMom with 3 shields")
	}

	for phase := 0; phase < 3; phase++ {
		w.Boss.X = w.Player.X + 40
		w.Boss.Y = w.Player.Y
		w.Player.StopCooldownTimer = 0
		w.Update(InputState{
			ThrowJustPressed: true,
			AimWorldX:        w.Boss.X,
			AimWorldY:        w.Boss.Y,
			SelectArchetype:  -1,
			ActionDir:        -1,
			ChooseUpgrade:    -1,
		}, 0.016)
		for w.HitStopTimer > 0 {
			w.Update(InputState{SelectArchetype: -1, ActionDir: -1, ChooseUpgrade: -1}, 0.02)
		}
		if len(w.Stops) == 0 || !w.Stops[0].HookedBoss {
			t.Fatalf("phase %d: expected stop to hook BossMom", phase)
		}
		// Force minigame completion cleanly and complete 4-step combo
		w.Player.X = w.Stops[0].X
		w.Player.Y = w.Stops[0].Y
		w.Stops[0].Minigame.Progress = 1.0
		w.Stops[0].Minigame.Completed = true
		w.Update(InputState{SelectArchetype: -1, ActionDir: -1, ChooseUpgrade: -1}, 0.016)
		for step := 0; step < 4; step++ {
			dir := w.Stops[0].Combo.Sequence[step]
			w.Update(InputState{SelectArchetype: -1, ActionDir: dir, ChooseUpgrade: -1}, 0.016)
		}
		for i := 0; i < 80 && len(w.Buses) > 0; i++ {
			w.Update(InputState{SelectArchetype: -1, ActionDir: -1, ChooseUpgrade: -1}, 0.03)
			if w.PendingLevelUp {
				w.Update(InputState{SelectArchetype: -1, ActionDir: -1, ChooseUpgrade: 0}, 0.016)
			}
		}
	}

	if w.Boss.Active || !w.Victory || w.BossesDefeated != 1 {
		t.Fatalf("expected BossMom defeated and Victory=true, active=%v victory=%v defeated=%d", w.Boss.Active, w.Victory, w.BossesDefeated)
	}
}
