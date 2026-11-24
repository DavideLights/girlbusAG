package game

import "testing"

func TestChestsSpawnFromMapDefAndSkippedInTutorial(t *testing.T) {
	w := NewWorldWithSeed(801)
	if len(w.Chests) != 3 {
		t.Fatalf("expected 3 NeonChests from DefaultDistrictMap, got %d", len(w.Chests))
	}
	if w.ClosedChestCount() != 3 {
		t.Fatalf("expected all 3 chests initially closed, got %d", w.ClosedChestCount())
	}

	for i, ch := range w.Chests {
		if ch.Opened {
			t.Fatalf("chest %d unexpectedly opened at start", i)
		}
		if !w.isClearPickupPos(ch.X, ch.Y) {
			t.Fatalf("chest %d at (%.1f, %.1f) overlaps solid prop or unwalkable area", i, ch.X, ch.Y)
		}
		d2 := distSq(ch.X, ch.Y, w.Player.X, w.Player.Y)
		if d2 < 100.0*100.0 {
			t.Fatalf("chest %d at (%.1f, %.1f) is too close to player spawn (%.1f, %.1f)", i, ch.X, ch.Y, w.Player.X, w.Player.Y)
		}
	}

	tut := NewTutorialWorld(nil)
	if len(tut.Chests) != 0 {
		t.Fatalf("expected 0 chests in TutorialWorld, got %d", len(tut.Chests))
	}
}

func TestBossDefeatGrantsKeyAndRestocksEmptyChests(t *testing.T) {
	w := NewWorldWithSeed(802)
	for i := range w.Chests {
		w.Chests[i].Opened = true
	}
	if w.ClosedChestCount() != 0 {
		t.Fatalf("expected 0 closed chests before boss defeat")
	}

	w.ForceSpawnBoss(BossBestFriend)
	w.Boss.Shields = 1
	w.Stops = append(w.Stops, BusStop{
		ID:               1,
		Active:           true,
		X:                w.Boss.X,
		Y:                w.Boss.Y,
		Arch:             ArchEmo,
		CaptureRadius:    120,
		LifeTimer:        10,
		HookedBoss:       true,
		BusRushTriggered: true,
	})
	w.Boss.HookedByStop = 1

	st := &w.Stops[0]
	w.Buses = append(w.Buses, NewNeonBus(st))
	w.DrainEvents()
	for step := 0; step < 120 && w.Boss.Active; step++ {
		w.updateBuses(1.0 / 30.0)
	}

	if w.Player.BossKeys != 1 {
		t.Fatalf("expected 1 BossKey after boss defeat, got %d", w.Player.BossKeys)
	}
	if w.ClosedChestCount() != len(w.Chests) {
		t.Fatalf("expected boss defeat to restock all chests when all were opened, got %d/%d closed",
			w.ClosedChestCount(), len(w.Chests))
	}

	events := w.DrainEvents()
	gotKeyEvent := false
	for _, ev := range events {
		if ev.Kind == EventBossKeyGained && ev.Value == 1 {
			gotKeyEvent = true
			break
		}
	}
	if !gotKeyEvent {
		t.Fatalf("expected EventBossKeyGained after boss defeat, events=%+v", events)
	}
}

func TestChestOpenByProximityAndKeyRequirement(t *testing.T) {
	w := NewWorldWithSeed(803)
	ch := &w.Chests[0]

	// Move player directly beside chest 0 without a key -> must stay locked
	w.Player.X = ch.X + 10
	w.Player.Y = ch.Y
	w.Player.BossKeys = 0
	w.updateChests(0.016, true)
	if ch.Opened || w.ChestsOpened != 0 {
		t.Fatalf("chest should remain closed when BossKeys == 0")
	}

	// Set up wounded HP, lower Chronos, and bullets inside & outside clear radius
	w.Player.HP = 3
	w.Player.ChronosMeter = 20
	w.Player.Cash = 100
	w.Player.BossKeys = 1
	w.Bullets.Count = 0
	w.Bullets.Spawn(Bullet{X: ch.X + 60, Y: ch.Y, Kind: BulletRazor, Radius: 4.0})  // inside 130px clear wave
	w.Bullets.Spawn(Bullet{X: ch.X + 190, Y: ch.Y, Kind: BulletRazor, Radius: 4.0}) // outside 130px clear wave

	w.DrainEvents()
	w.updateChests(0.016, false)

	if !ch.Opened || w.ChestsOpened != 1 {
		t.Fatalf("expected chest 0 to open on proximity with key (Opened=%v ChestsOpened=%d)", ch.Opened, w.ChestsOpened)
	}
	if w.Player.BossKeys != 0 {
		t.Fatalf("expected BossKeys decremented to 0, got %d", w.Player.BossKeys)
	}
	if w.Player.RelicCountOwned() != 1 {
		t.Fatalf("expected 1 Relic granted by chest, got %d", w.Player.RelicCountOwned())
	}
	if w.Player.HP != 5 {
		t.Fatalf("expected +2 HP heal (3 -> 5), got %d", w.Player.HP)
	}
	if w.Player.ChronosMeter < 59.9 {
		t.Fatalf("expected +40 Chronos (20 -> 60), got %.1f", w.Player.ChronosMeter)
	}
	if w.Player.Cash <= 100 {
		t.Fatalf("expected Cash reward > 100, got %d", w.Player.Cash)
	}
	if w.Bullets.Count != 1 {
		t.Fatalf("expected chest shockwave to clear near bullet and preserve far bullet (count=1), got %d", w.Bullets.Count)
	}

	events := w.DrainEvents()
	var gotChest, gotRelic bool
	for _, ev := range events {
		if ev.Kind == EventChestOpened && !ev.Flag {
			gotChest = true
		}
		if ev.Kind == EventRelicAcquired {
			gotRelic = true
		}
	}
	if !gotChest || !gotRelic {
		t.Fatalf("expected EventChestOpened and EventRelicAcquired, got %+v", events)
	}

	// Re-entering or attempting to reopen chest 0 with another key must be rejected
	w.Player.BossKeys = 1
	if w.OpenChest(0) {
		t.Fatalf("expected already-opened chest 0 to reject reopening")
	}
}

func TestChestLegendaryFallbackWhenAllRelicsOwnedAndQueuedLevelUp(t *testing.T) {
	w := NewWorldWithSeed(804)
	for i := 0; i < int(RelicCount); i++ {
		w.Player.Relics[i] = true
	}
	startMaxHP := w.Player.MaxHP
	w.Player.BossKeys = 1

	// Active stop around player queues the bonus Level-Up cleanly
	w.Stops = append(w.Stops, BusStop{
		ID:            10,
		Active:        true,
		X:             w.Player.X,
		Y:             w.Player.Y,
		CaptureRadius: 120,
		PlayerInside:  true,
		LifeTimer:     10,
	})

	w.DrainEvents()
	if !w.OpenChest(0) {
		t.Fatalf("expected OpenChest(0) to succeed")
	}
	if w.Player.MaxHP != startMaxHP+1 {
		t.Fatalf("expected +1 MaxHP legendary fallback (%d -> %d), got %d", startMaxHP, startMaxHP+1, w.Player.MaxHP)
	}
	if w.PendingLevelUp || w.QueuedLevelUps != 1 {
		t.Fatalf("expected Level-Up queued during active stop (Pending=%v Queued=%d)", w.PendingLevelUp, w.QueuedLevelUps)
	}

	events := w.DrainEvents()
	foundLegend := false
	for _, ev := range events {
		if ev.Kind == EventChestOpened && ev.Value == -1 && ev.Flag {
			foundLegend = true
			break
		}
	}
	if !foundLegend {
		t.Fatalf("expected legendary EventChestOpened(Value=-1, Flag=true), got %+v", events)
	}
}

func TestChestAutoRestocksAfterTimerWhenKeysRemain(t *testing.T) {
	w := NewWorldWithSeed(805)
	w.Chests[0].Opened = true
	w.Chests[1].Opened = true
	w.Player.BossKeys = 2

	if !w.OpenChest(2) {
		t.Fatalf("expected OpenChest(2) to succeed")
	}
	if w.ClosedChestCount() != 0 {
		t.Fatalf("expected 0 closed chests after opening the last one")
	}
	if w.Player.BossKeys != 1 {
		t.Fatalf("expected 1 BossKey remaining, got %d", w.Player.BossKeys)
	}

	// Step simulation past restock delay (4.0s)
	for step := 0; step < 300; step++ {
		w.updateChests(1.0/60.0, false)
	}
	if w.ClosedChestCount() != len(w.Chests) {
		t.Fatalf("expected automatic chest restock after delay when BossKeys > 0, got %d closed", w.ClosedChestCount())
	}
}
