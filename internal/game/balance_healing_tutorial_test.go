package game

import (
	"math"
	"math/rand"
	"testing"
)

func TestEarlyRunDifficultyBalanceAndRamp(t *testing.T) {
	w := NewWorldWithSeed(501)

	// 1. Player rebalanced baseline stats
	if w.Player.HP != 6 || w.Player.MaxHP != 6 {
		t.Fatalf("expected base Player HP/MaxHP = 6/6, got %d/%d", w.Player.HP, w.Player.MaxHP)
	}
	if w.Player.StopCooldownBase != 6.5 {
		t.Fatalf("expected StopCooldownBase = 6.5, got %v", w.Player.StopCooldownBase)
	}
	if !w.Player.TakeDamage(1) || w.Player.InvulnTimer != 1.15 {
		t.Fatalf("expected InvulnTimer = 1.15 after TakeDamage, got %v", w.Player.InvulnTimer)
	}
	w.Player.Heal(1)
	if w.Player.HealFlashTimer != 0.45 {
		t.Fatalf("expected HealFlashTimer = 0.45 after Heal, got %v", w.Player.HealFlashTimer)
	}
	w.Player.Update(baseTestInput(), 0.20, false, false, 0)
	if w.Player.HealFlashTimer >= 0.45 || w.Player.HealFlashTimer <= 0.20 {
		t.Fatalf("expected HealFlashTimer to decay smoothly (~0.25), got %v", w.Player.HealFlashTimer)
	}

	// 2. Initial wave: 5 girls, 0 VIPs, initial ShootTimer in [2.0, 3.6]
	if len(w.Girls) != 5 {
		t.Fatalf("expected 5 initial girls at t=0, got %d", len(w.Girls))
	}
	for i, g := range w.Girls {
		if g.IsVIP {
			t.Fatalf("expected 0 VIPs at t=0, girl %d is VIP", i)
		}
		if g.ShootTimer < 2.0 || g.ShootTimer > 3.6 {
			t.Fatalf("expected Tier 1 initial ShootTimer in [2.0, 3.6], got %v", g.ShootTimer)
		}
	}

	// 3. EarlyDifficultyScale ramp: 0.55 @ t=0 -> 0.775 @ t=60 -> 1.0 @ t=120 / Tier 2+ / BossRush
	if math.Abs(w.EarlyDifficultyScale()-0.55) > 0.001 {
		t.Fatalf("expected EarlyDifficultyScale()=0.55 at t=0, got %v", w.EarlyDifficultyScale())
	}
	w.RunTime = 60.0
	if math.Abs(w.EarlyDifficultyScale()-0.775) > 0.001 {
		t.Fatalf("expected EarlyDifficultyScale()=0.775 at t=60, got %v", w.EarlyDifficultyScale())
	}
	w.RunTime = 120.0
	if math.Abs(w.EarlyDifficultyScale()-1.0) > 0.001 {
		t.Fatalf("expected EarlyDifficultyScale()=1.0 at t=120, got %v", w.EarlyDifficultyScale())
	}
	w.RunTime = 0
	w.DifficultyTier = 2
	if w.EarlyDifficultyScale() != 1.0 {
		t.Fatalf("expected EarlyDifficultyScale()=1.0 at Tier 2, got %v", w.EarlyDifficultyScale())
	}
	wRush := NewWorldWithOptions(501, true)
	if wRush.EarlyDifficultyScale() != 1.0 {
		t.Fatalf("expected EarlyDifficultyScale()=1.0 in BossRushMode, got %v", wRush.EarlyDifficultyScale())
	}

	// 4. Tier 1 volley density, cooldown division, and bullet speed scaling
	rng := rand.New(rand.NewSource(11))
	checkVolley := func(arch Archetype, tier int, isVIP bool, scale float32, wantCount int) (float32, float32) {
		var pool BulletPool
		g := NewGirl(1, 500, 300, arch, isVIP, rng)
		g.ShootTimer = 0
		g.firePattern(tier, scale, 600, 300, &pool)
		if pool.Count != wantCount {
			t.Fatalf("arch=%d tier=%d vip=%v: got %d bullets, want %d", arch, tier, isVIP, pool.Count, wantCount)
		}
		spd := float32(math.Hypot(float64(pool.Items[0].VX), float64(pool.Items[0].VY)))
		return g.ShootTimer, spd
	}

	cdEarly, spdEarly := checkVolley(ArchKawaii, 1, false, 0.55, 3)
	cdFull, spdFull := checkVolley(ArchKawaii, 1, false, 1.0, 3)
	if cdEarly <= cdFull*1.7 {
		t.Fatalf("expected early fire cooldown (~1.8x longer) cdEarly=%v vs cdFull=%v", cdEarly, cdFull)
	}
	wantEarlySpd := float32(118.0) * (0.78 + 0.22*0.55)
	if math.Abs(float64(spdEarly-wantEarlySpd)) > 0.5 || math.Abs(float64(spdFull-118.0)) > 0.5 {
		t.Fatalf("expected Kawaii Tier 1 speeds early=%v (want %v) full=%v (want 118)", spdEarly, wantEarlySpd, spdFull)
	}
	checkVolley(ArchKawaii, 2, false, 1.0, 5)
	checkVolley(ArchKawaii, 3, false, 1.0, 7)

	checkVolley(ArchGyaru, 1, false, 0.55, 2)
	checkVolley(ArchGyaru, 2, false, 1.0, 3)
	checkVolley(ArchGyaru, 1, true, 1.0, 5)

	checkVolley(ArchEmo, 1, false, 0.55, 1)
	checkVolley(ArchEmo, 2, false, 1.0, 2)
	checkVolley(ArchEmo, 3, false, 1.0, 3)

	checkVolley(ArchStudent, 1, false, 0.55, 1)
	checkVolley(ArchStudent, 2, false, 1.0, 3)
}

func TestWorldHealingPickupsAndEspressoPrice(t *testing.T) {
	w := NewWorldWithSeed(502)
	w.Girls = nil

	// 1. 2 initial HealPickups spawned at run start in clear walkable spots
	if len(w.HealPickups) != 2 {
		t.Fatalf("expected 2 initial HealPickups, got %d", len(w.HealPickups))
	}
	for i, hp := range w.HealPickups {
		if !w.isClearPickupPos(float32(hp.X), float32(hp.Y)) {
			t.Fatalf("initial HealPickup %d at (%v,%v) collides with solid prop or out of bounds", i, hp.X, hp.Y)
		}
	}

	// 2. Magnet within 56px & collection within 26px heals HP, restores Chronos, flashes green & emits EventHealPickup
	w.Player.HP = 3
	w.Player.ChronosMeter = 50
	w.Player.HealFlashTimer = 0
	w.HealPickups = []HealPickup{{
		X:         float64(w.Player.X + 45),
		Y:         float64(w.Player.Y),
		Amount:    1,
		Chronos:   20.0,
		LifeTimer: 45.0,
	}}
	startX := w.HealPickups[0].X
	w.updateHealPickups(0.05)
	if len(w.HealPickups) != 1 || w.HealPickups[0].X >= startX {
		t.Fatalf("expected pickup within 56px to magnet toward Alfredo, startX=%v now=%v", startX, w.HealPickups[0].X)
	}
	// Continue stepping until collected
	for i := 0; i < 10 && len(w.HealPickups) > 0; i++ {
		w.updateHealPickups(0.05)
	}
	if len(w.HealPickups) != 0 {
		t.Fatalf("expected magnetized HealPickup to be collected")
	}
	if w.Player.HP != 4 || w.Player.ChronosMeter != 70 {
		t.Fatalf("expected HP=4 Chronos=70 after pickup, got HP=%d Chronos=%v", w.Player.HP, w.Player.ChronosMeter)
	}
	if w.Player.HealFlashTimer != 0.45 {
		t.Fatalf("expected HealFlashTimer=0.45 after pickup, got %v", w.Player.HealFlashTimer)
	}
	foundHealEvent := false
	for _, ev := range w.DrainEvents() {
		if ev.Kind == EventHealPickup && ev.Value == 1 && !ev.Flag {
			foundHealEvent = true
			break
		}
	}
	if !foundHealEvent {
		t.Fatalf("expected EventHealPickup to be emitted on pickup collection")
	}

	// 3. Periodic spawn every 24s when fewer than 3 pickups exist
	w.HealPickups = nil
	w.healSpawnTimer = 0.02
	w.updateHealPickups(0.05)
	if len(w.HealPickups) != 1 {
		t.Fatalf("expected periodic HealPickup spawn after 24s timer, got %d", len(w.HealPickups))
	}

	// 4. Boarding 4 girls while injured guarantees a +1 HP pickup drop
	w.HealPickups = nil
	w.boardedSinceHeal = 0
	w.Player.HP = 4
	for i := 0; i < 4; i++ {
		st := NewBusStop(i+1, 500, 340, ArchEmo, 1.0, false, false, false, w.rng)
		g := NewGirl(i+100, 500, 340, ArchEmo, false, w.rng)
		g.HookedByStop = st.ID
		w.Girls = []Girl{g}
		st.HookedGirlIDs = []int{g.ID}
		w.completeStopBoarding(&st)
	}
	if len(w.HealPickups) < 1 {
		t.Fatalf("expected guaranteed HealPickup drop within 4 boarded girls while injured")
	}

	// 5. Boss shield break ALWAYS drops a Big (+2 HP, +35 Chronos) pickup
	w.HealPickups = nil
	w.ForceSpawnBoss(BossBestFriend)
	st := NewBusStop(99, w.Boss.X, w.Boss.Y, ArchEmo, 1.0, false, false, false, w.rng)
	st.HookedBoss = true
	w.Boss.HookedByStop = st.ID
	w.completeStopBoarding(&st)
	if len(w.HealPickups) != 1 || !w.HealPickups[0].Big || w.HealPickups[0].Amount != 2 || w.HealPickups[0].Chronos != 35 {
		t.Fatalf("expected Big (+2 HP, +35 Chronos) pickup on boss shield break, got %+v", w.HealPickups)
	}

	// 6. Vending Shop Espresso price lowered from $50 -> $35
	if w.Shops[0].HealCost != 35 {
		t.Fatalf("expected Shop HealCost = 35, got %d", w.Shops[0].HealCost)
	}
	w.Player.Cash = 35
	w.Player.HP = 3
	w.OpenShop(0)
	inBuy := baseTestInput()
	inBuy.ShopAction = ShopActionHeal
	w.HandleShopInput(inBuy)
	if w.Player.Cash != 0 || w.Player.HP != 5 {
		t.Fatalf("expected $35 Espresso purchase to leave Cash=0 and HP=5, got Cash=%d HP=%d", w.Player.Cash, w.Player.HP)
	}
}

func TestMouseClickShopAndSlotInteraction(t *testing.T) {
	w := NewWorldWithSeed(503)
	w.Girls = nil

	// 1. Left-click directly on Vending Shop while within 85px opens ShopMenu without throwing a stop
	sh := w.Shops[0]
	w.Player.X = sh.X
	w.Player.Y = sh.Y + 70 // 70px away (<= 85px)
	in := baseTestInput()
	in.MouseLeftJustPressed = true
	in.ThrowJustPressed = true
	in.AimWorldX = sh.X + 10
	in.AimWorldY = sh.Y - 8
	w.Update(in, 0.016)
	if !w.ShopMenu.Active {
		t.Fatalf("expected mouse click on nearby VendingShop to open ShopMenu")
	}
	if len(w.Stops) != 0 {
		t.Fatalf("expected mouse click on VendingShop NOT to throw a BusStop, stops=%d", len(w.Stops))
	}
	w.ShopMenu.Active = false

	// 2. Left-click within 28px of SlotMachine while within 110px of Alfredo opens SlotModal without throwing a stop
	sm := w.SlotMachines[0]
	w.Player.X = sm.X
	w.Player.Y = sm.Y + 100 // 100px away (<= 110px)
	in2 := baseTestInput()
	in2.MouseLeftJustPressed = true
	in2.ThrowJustPressed = true
	in2.AimWorldX = sm.X + 15 // 15px from slot center (<= 28px)
	in2.AimWorldY = sm.Y
	w.Update(in2, 0.016)
	if !w.SlotModal.Active {
		t.Fatalf("expected mouse click within 28px of SlotMachine (100px from Alfredo) to open SlotModal")
	}
	if len(w.Stops) != 0 {
		t.Fatalf("expected mouse click on SlotMachine NOT to throw a BusStop")
	}
	w.SlotModal.Active = false

	// 3. When inside an active BusStop, clicking near a shop does NOT interrupt the minigame
	stop := NewBusStop(1, sm.X, sm.Y+20, ArchKawaii, 1.0, false, false, false, w.rng)
	stop.PlayerInside = true
	w.Stops = append(w.Stops, stop)
	w.Player.X = sm.X
	w.Player.Y = sm.Y + 20
	w.Update(in2, 0.016)
	if w.SlotModal.Active {
		t.Fatalf("expected active BusStop minigame to block mouse-click shop/slot opening")
	}
}

func TestTutorialWorldInitImmortalityAndMilestones(t *testing.T) {
	w := NewTutorialWorld(nil)

	// 1. Initialization checks
	if !w.TutorialMode || w.TutorialStep != 0 {
		t.Fatalf("expected TutorialMode=true TutorialStep=0, got mode=%v step=%d", w.TutorialMode, w.TutorialStep)
	}
	if w.Player.Cash != 180 || w.Player.HP != 6 || w.Player.MaxHP != 6 || w.Player.StopCooldownBase != 3.0 {
		t.Fatalf("unexpected tutorial player stats: cash=%d hp=%d/%d cd=%v",
			w.Player.Cash, w.Player.HP, w.Player.MaxHP, w.Player.StopCooldownBase)
	}
	if len(w.Girls) != 4 {
		t.Fatalf("expected 4 gentle practice girls in tutorial, got %d", len(w.Girls))
	}
	var seen [ArchCount]bool
	for _, g := range w.Girls {
		if g.IsVIP {
			t.Fatalf("expected 0 VIPs in TutorialMode")
		}
		seen[g.Arch] = true
	}
	for a := Archetype(0); a < ArchCount; a++ {
		if !seen[a] {
			t.Fatalf("expected tutorial practice girl of archetype %d", a)
		}
	}
	if len(w.HealPickups) < 1 {
		t.Fatalf("expected nearby HealPickup in TutorialMode")
	}

	// 2. Immortality: lethal damage clamps to 1 HP (no GameOver) and restores to 6/6 after 1.0s
	w.Player.HP = 1
	w.Player.InvulnTimer = 0
	w.damagePlayer(2, BulletRazor)
	if w.GameOver || w.Player.HP != 1 {
		t.Fatalf("expected TutorialMode immortality (HP=1, GameOver=false), got HP=%d GameOver=%v", w.Player.HP, w.GameOver)
	}
	for i := 0; i < 25; i++ {
		w.Update(baseTestInput(), 0.05)
	}
	if w.Player.HP != 6 {
		t.Fatalf("expected TutorialMode auto-restore to 6 HP after 1s, got %d", w.Player.HP)
	}

	// 3. No Boss spawns interrupt TutorialMode even at high RunTime
	w.RunTime = 300
	w.Update(baseTestInput(), 0.05)
	if w.Boss.Active {
		t.Fatalf("expected no Boss to spawn in TutorialMode")
	}

	// 4. Milestones & TutorialStep progression (Move -> Throw -> Minigame -> Board -> Shop/Slot)
	inMove := baseTestInput()
	inMove.MoveX = 1.0
	w.Update(inMove, 0.02)
	if !w.TutMoved {
		t.Fatalf("expected TutMoved=true after moving")
	}

	inThrow := baseTestInput()
	inThrow.ThrowJustPressed = true
	inThrow.AimWorldX = w.Girls[0].X
	inThrow.AimWorldY = w.Girls[0].Y
	w.Update(inThrow, 0.02)
	if !w.TutThrewStop || w.TutorialStep < 1 {
		t.Fatalf("expected TutThrewStop=true and TutorialStep>=1, got flag=%v step=%d", w.TutThrewStop, w.TutorialStep)
	}
	for w.HitStopTimer > 0 {
		w.Update(baseTestInput(), 0.02)
	}

	// Complete minigame & board the girl
	w.Player.X = w.Stops[0].X
	w.Player.Y = w.Stops[0].Y
	w.Stops[0].Minigame.Progress = 1.0
	w.Stops[0].Minigame.Completed = true
	w.Update(baseTestInput(), 0.02)
	if !w.TutCompletedMinigame || w.TutorialStep < 2 {
		t.Fatalf("expected TutCompletedMinigame=true and TutorialStep>=2, got flag=%v step=%d", w.TutCompletedMinigame, w.TutorialStep)
	}
	for step := 0; step < 4; step++ {
		inCombo := baseTestInput()
		inCombo.ActionDir = w.Stops[0].Combo.Sequence[step]
		w.Update(inCombo, 0.016)
	}
	for i := 0; i < 60 && !w.TutBoardedBus; i++ {
		w.Update(baseTestInput(), 0.03)
		if w.PendingLevelUp {
			inUp := baseTestInput()
			inUp.ChooseUpgrade = 0
			w.Update(inUp, 0.016)
		}
	}
	if !w.TutBoardedBus || w.TutorialStep != 3 {
		t.Fatalf("expected TutBoardedBus=true and TutorialStep=3, got flag=%v step=%d", w.TutBoardedBus, w.TutorialStep)
	}

	// Using Shop marks TutUsedSlotOrShop
	w.OpenShop(len(w.Shops) - 1)
	inBuy := baseTestInput()
	inBuy.ShopAction = ShopActionHeal
	w.Update(inBuy, 0.016)
	if !w.TutUsedSlotOrShop {
		t.Fatalf("expected TutUsedSlotOrShop=true after purchasing Espresso in TutorialMode")
	}

	// Clearing all girls respawns 4 fresh practice girls gently
	w.ShopMenu.Active = false
	w.Girls = nil
	w.Update(baseTestInput(), 0.02)
	if len(w.Girls) != 4 {
		t.Fatalf("expected 4 fresh practice girls to respawn when all are boarded, got %d", len(w.Girls))
	}
}
