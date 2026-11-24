package game

import "testing"

func TestBossSchedulesAndAllSixBosses(t *testing.T) {
	// 1. Normal run draws 4 distinct bosses ending with BossRector or BossMom
	wNorm := NewWorldWithOptions(55, false)
	if len(wNorm.BossSchedule) != 4 || len(wNorm.BossScheduleTimes) != 4 {
		t.Fatalf("expected 4 scheduled bosses in normal run, got %d", len(wNorm.BossSchedule))
	}
	last := wNorm.BossSchedule[3]
	if last != BossRector && last != BossMom {
		t.Fatalf("expected final boss to be Rector or Mom, got %v", last)
	}

	// 2. Boss Rush includes all 6 bosses in order
	wRush := NewWorldWithOptions(55, true)
	if !wRush.BossRushMode || len(wRush.BossSchedule) != int(BossCount) {
		t.Fatalf("expected 6 bosses in Boss Rush, got %d", len(wRush.BossSchedule))
	}

	// 3. L'Amante decoy trap triggers cooldown & bullet burst without placing a stop
	wLover := NewWorldWithSeed(77)
	wLover.Girls = nil
	wLover.ForceSpawnBoss(BossLover)
	wLover.Boss.X = 900
	wLover.Boss.Y = 300
	wLover.Boss.Clones[0] = BossClone{X: wLover.Player.X + 40, Y: wLover.Player.Y, Active: true}
	wLover.Boss.Clones[1] = BossClone{X: 1000, Y: 400, Active: true}
	wLover.Player.SelectedStopArch = ArchKawaii
	wLover.Update(InputState{
		ThrowJustPressed: true,
		AimWorldX:        wLover.Boss.Clones[0].X,
		AimWorldY:        wLover.Boss.Clones[0].Y,
		SelectArchetype:  -1,
		ActionDir:        -1,
		ChooseUpgrade:    -1,
	}, 0.016)
	if len(wLover.Stops) != 0 || wLover.Boss.Clones[0].Active {
		t.Fatalf("expected decoy clone to absorb throw and vanish, stops=%d cloneActive=%v", len(wLover.Stops), wLover.Boss.Clones[0].Active)
	}
	if wLover.Player.StopCooldownTimer <= 0 {
		t.Fatalf("expected decoy trap to trigger stop cooldown")
	}

	// 4. La Rettrice Martial ZTL triggers +1* Wanted on non-perfect throw
	wRector := NewWorldWithSeed(88)
	wRector.Girls = nil
	wRector.ForceSpawnBoss(BossRector)
	wRector.Boss.X = 1100
	wRector.Boss.Y = 300
	wRector.Update(InputState{
		ThrowJustPressed: true,
		AimWorldX:        wRector.Player.X + 40,
		AimWorldY:        wRector.Player.Y,
		SelectArchetype:  -1,
		ActionDir:        -1,
		ChooseUpgrade:    -1,
	}, 0.016)
	if wRector.Police.Stars < 1 {
		t.Fatalf("expected La Rettrice martial law to add +1* Wanted on non-perfect throw")
	}

	// 5. La Dirigente Aziendale requires catching at least 1 escort to break shield
	wExec := NewWorldWithSeed(99)
	wExec.Girls = nil
	wExec.ForceSpawnBoss(BossExecutive)
	if wExec.Boss.ActiveEscortCount() != 4 {
		t.Fatalf("expected 4 escorts for BossExecutive, got %d", wExec.Boss.ActiveEscortCount())
	}
}

func TestShopRelicsAndSlotMachine(t *testing.T) {
	w := NewWorldWithSeed(333)
	w.Player.Cash = 500
	w.Player.X = w.Shops[0].X
	w.Player.Y = w.Shops[0].Y

	// Open shop & buy random Relic
	w.Update(InputState{InteractJustPressed: true, SelectArchetype: -1, ActionDir: -1, ChooseUpgrade: -1}, 0.016)
	if !w.ShopMenu.Active {
		t.Fatalf("expected shop modal to open")
	}
	w.Update(InputState{ShopAction: ShopActionRelic, SelectArchetype: -1, ActionDir: -1, ChooseUpgrade: -1}, 0.016)
	if w.Player.RelicCountOwned() != 1 || w.Player.Cash != 380 {
		t.Fatalf("expected 1 relic owned and cash=380, got relics=%d cash=%d", w.Player.RelicCountOwned(), w.Player.Cash)
	}
	var ownedRelic RelicKind = -1
	for rk := 0; rk < int(RelicCount); rk++ {
		if w.Player.Relics[rk] {
			ownedRelic = RelicKind(rk)
			break
		}
	}
	spec := GetRelicSpec(ownedRelic)
	if w.ShopMenu.Message != spec.Name+": "+spec.Desc {
		t.Fatalf("expected ShopMenu.Message to explain relic %q, got %q", spec.Name+": "+spec.Desc, w.ShopMenu.Message)
	}
	events := w.DrainEvents()
	sawRelicEvent := false
	for _, ev := range events {
		if ev.Kind == EventRelicAcquired && ev.Value == int(ownedRelic) {
			sawRelicEvent = true
			break
		}
	}
	if !sawRelicEvent {
		t.Fatalf("expected EventRelicAcquired for relic %d (%s)", ownedRelic, spec.Name)
	}
	w.Update(InputState{CancelJustPressed: true, SelectArchetype: -1, ActionDir: -1, ChooseUpgrade: -1}, 0.016)

	// Walk to Slot Machine, open, and spin until completion
	w.Player.X = w.SlotMachines[0].X
	w.Player.Y = w.SlotMachines[0].Y
	w.Update(InputState{InteractJustPressed: true, SelectArchetype: -1, ActionDir: -1, ChooseUpgrade: -1}, 0.016)
	if !w.SlotModal.Active {
		t.Fatalf("expected SlotModal to open when interacting near SlotMachine")
	}

	startCash := w.Player.Cash
	w.Update(InputState{SlotSpinJustPressed: true, SelectArchetype: -1, ActionDir: -1, ChooseUpgrade: -1}, 0.016)
	if !w.SlotModal.Spinning || w.Player.Cash >= startCash {
		t.Fatalf("expected slot spin to start and deduct SpinCost")
	}

	// Step simulation until 1.15s spin completes
	for i := 0; i < 40 && w.SlotModal.Spinning; i++ {
		w.Update(InputState{SelectArchetype: -1, ActionDir: -1, ChooseUpgrade: -1}, 0.04)
	}
	if w.SlotModal.Spinning {
		t.Fatalf("expected slot spin to finish within 1.6s")
	}
	if w.SlotMachines[0].SpinsUsed != 1 {
		t.Fatalf("expected SpinsUsed=1, got %d", w.SlotMachines[0].SpinsUsed)
	}
}
