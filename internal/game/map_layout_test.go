package game

import "testing"

func TestDistrictTopologyAndNoOverlaps(t *testing.T) {
	w := NewWorldWithSeed(101)

	if w.BoardwalkY != 624.0 || w.CanalY != 680.0 {
		t.Fatalf("expected BoardwalkY=624 and CanalY=680, got (%v, %v)", w.BoardwalkY, w.CanalY)
	}
	if len(w.ParkIslands) != 2 {
		t.Fatalf("expected 2 ParkIslands, got %d", len(w.ParkIslands))
	}
	if len(w.BusBays) != 4 {
		t.Fatalf("expected 4 BusBays, got %d", len(w.BusBays))
	}
	if len(w.ZTLZones) != 2 {
		t.Fatalf("expected 2 ZTLZones, got %d", len(w.ZTLZones))
	}
	if len(w.CrosswalkCols) == 0 || len(w.CrosswalkRows) == 0 {
		t.Fatalf("expected CrosswalkCols and CrosswalkRows to be populated")
	}
	if len(w.Lights) < 20 {
		t.Fatalf("expected rich LightSource setup, got %d", len(w.Lights))
	}

	// Count storefronts, parked cars, trees, benches
	var buildings, cars, trees int
	for _, p := range w.Props {
		switch p.Kind {
		case PropBuildingKonbini, PropBuildingCoffee, PropBuildingArcade, PropBuildingRamen, PropBuildingBookstore, PropBuildingMetro:
			buildings++
		case PropCarRed, PropCarYellow, PropCarMint:
			cars++
		case PropTreeLush:
			trees++
		}
	}
	if buildings != 8 || cars != 4 || trees != 4 {
		t.Fatalf("unexpected prop counts: buildings=%d cars=%d trees=%d", buildings, cars, trees)
	}
	if len(w.SlotMachines) != 2 {
		t.Fatalf("expected 2 SlotMachines, got %d", len(w.SlotMachines))
	}

	// Verify alleys and central cross-street are open (no solid prop inside alleys)
	for _, x := range []float32{280, 640, 992} {
		for _, p := range w.Props {
			if p.Solid && p.CollideRect.Contains(x, 42) {
				t.Fatalf("expected alley/avenue at x=%v to be open, blocked by %+v", x, p)
			}
		}
	}

	// Verify BusBays don't overlap ZTLZones, ParkIslands, or Solid props
	for _, bay := range w.BusBays {
		cx, cy := bay.X+bay.W*0.5, bay.Y+bay.H*0.5
		if w.isPointInZTL(cx, cy) {
			t.Fatalf("BusBay %+v overlaps ZTLZone", bay)
		}
		for _, p := range w.Props {
			if p.Solid && p.CollideRect.Contains(cx, cy) {
				t.Fatalf("BusBay %+v center overlaps solid prop %+v", bay, p)
			}
		}
	}

	// Verify VendingShops and SlotMachines are accessible (not inside solid props)
	for _, sh := range w.Shops {
		for _, p := range w.Props {
			if p.Solid && p.CollideRect.Contains(sh.X, sh.Y) {
				t.Fatalf("VendingShop %+v is inside solid prop %+v", sh, p)
			}
		}
	}
	for _, sm := range w.SlotMachines {
		for _, p := range w.Props {
			if p.Solid && p.CollideRect.Contains(sm.X, sm.Y) {
				t.Fatalf("SlotMachine %+v is inside solid prop %+v", sm, p)
			}
		}
	}
}

func TestEvaluateAimAndBusBayBonuses(t *testing.T) {
	w := NewWorldWithSeed(202)
	bay := w.BusBays[0] // {X: 160, Y: 150, W: 88, H: 36}
	bayCX, bayCY := bay.X+bay.W*0.5, bay.Y+bay.H*0.5

	// Place Alfredo near the BusBay and spawn 2 girls inside capture radius (1 Kawaii, 1 Emo VIP)
	w.Player.X = bayCX
	w.Player.Y = bayCY + 40
	w.Player.SelectedStopArch = ArchKawaii
	w.Girls = []Girl{
		NewGirl(100, bayCX+10, bayCY, ArchKawaii, false, w.rng),
		NewGirl(101, bayCX+30, bayCY, ArchEmo, true, w.rng),
	}

	fb := w.EvaluateAim(bayCX, bayCY)
	if !fb.InRange || !fb.InBusBay || fb.InZTL || !fb.IsPerfectDrop {
		t.Fatalf("unexpected AimFeedback: %+v", fb)
	}
	if fb.HookCandidateCount != 2 || fb.SynergyCount != 1 || fb.VIPCount != 1 {
		t.Fatalf("expected 2 candidates, 1 synergy, 1 VIP, got %+v", fb)
	}

	// Throw stop into BusBay
	w.Update(InputState{
		ThrowJustPressed: true,
		AimWorldX:        bayCX,
		AimWorldY:        bayCY,
		SelectArchetype:  -1,
		ActionDir:        -1,
		ChooseUpgrade:    -1,
	}, 0.016)

	if len(w.Stops) != 1 || !w.Stops[0].InBusBay {
		t.Fatalf("expected 1 stop with InBusBay=true, got %+v", w.Stops)
	}
	spec := GetArchetypeSpec(ArchKawaii)
	if w.Stops[0].ShieldHP != spec.ShieldHP+1 || w.Stops[0].MaxShieldHP != spec.ShieldHP+1 {
		t.Fatalf("expected +1 ShieldHP (%d) in BusBay, got %d/%d", spec.ShieldHP+1, w.Stops[0].ShieldHP, w.Stops[0].MaxShieldHP)
	}

	foundPopup := false
	for _, p := range w.Popups {
		if p.Text == "BUS BAY +25% VELOCITA!" {
			foundPopup = true
			break
		}
	}
	if !foundPopup {
		t.Fatalf("expected BusBay popup, popups=%+v", w.Popups)
	}

	// Compare +25% Minigame charge speed: BusBay vs Normal Kawaii stop
	w.HitStopTimer = 0
	w.Stops[0].Minigame.Progress = 0
	w.Player.X = w.Stops[0].X
	w.Player.Y = w.Stops[0].Y

	w2 := NewWorldWithSeed(202)
	w2.Girls = nil
	w2.Player.SelectedStopArch = ArchKawaii
	w2.Update(InputState{
		ThrowJustPressed: true,
		AimWorldX:        w2.Player.X + 40,
		AimWorldY:        w2.Player.Y,
		SelectArchetype:  -1,
		ActionDir:        -1,
		ChooseUpgrade:    -1,
	}, 0.016)
	w2.HitStopTimer = 0
	w2.Stops[0].Minigame.Progress = 0
	w2.Player.X = w2.Stops[0].X
	w2.Player.Y = w2.Stops[0].Y

	w.Update(InputState{SelectArchetype: -1, ActionDir: -1, ChooseUpgrade: -1}, 0.05)
	w2.Update(InputState{SelectArchetype: -1, ActionDir: -1, ChooseUpgrade: -1}, 0.05)

	ratio := w.Stops[0].Minigame.Progress / w2.Stops[0].Minigame.Progress
	if ratio < 1.24 || ratio > 1.26 {
		t.Fatalf("expected +25%% minigame charge speed (ratio ~1.25), got busbay=%v normal=%v ratio=%v",
			w.Stops[0].Minigame.Progress, w2.Stops[0].Minigame.Progress, ratio)
	}

	// Verify ZTL aim preview gives +25% radius
	ztl := w.ZTLZones[0]
	w.Player.X = ztl.X + ztl.W*0.5
	w.Player.Y = ztl.Y + ztl.H*0.5
	ztlFB := w.EvaluateAim(w.Player.X, w.Player.Y)
	if !ztlFB.InZTL {
		t.Fatalf("expected ZTL feedback at ZTL center")
	}
	expectedZTLR := spec.CaptureRadius * w.Player.RangeMult * 1.25
	if ztlFB.PreviewRadius < expectedZTLR-0.5 || ztlFB.PreviewRadius > expectedZTLR+0.5 {
		t.Fatalf("expected ZTL PreviewRadius ~%v, got %v", expectedZTLR, ztlFB.PreviewRadius)
	}
}

func TestPropCollisionsAndBoardwalkRailing(t *testing.T) {
	w := NewWorldWithSeed(303)

	// 1. Alfredo cannot walk into the canal past BoardwalkRailY (676)
	w.Player.X = 640
	w.Player.Y = 674
	w.Update(InputState{MoveY: 1, SelectArchetype: -1, ActionDir: -1, ChooseUpgrade: -1}, 0.05)
	if w.Player.Y > BoardwalkRailY+0.01 {
		t.Fatalf("expected Player to stop at boardwalk railing Y<=%v, got %v", BoardwalkRailY, w.Player.Y)
	}

	// 2. Alfredo is pushed out of solid parked car collider
	var car *MapProp
	for i := range w.Props {
		if w.Props[i].Kind == PropCarRed {
			car = &w.Props[i]
			break
		}
	}
	if car == nil {
		t.Fatalf("expected PropCarRed in layout")
	}
	w.Player.X = car.CollideRect.X + car.CollideRect.W*0.5
	w.Player.Y = car.CollideRect.Y + car.CollideRect.H*0.5
	w.Update(InputState{SelectArchetype: -1, ActionDir: -1, ChooseUpgrade: -1}, 0.016)
	if car.CollideRect.Contains(w.Player.X, w.Player.Y) {
		t.Fatalf("expected Player to be pushed out of car collider %+v, pos=(%v,%v)", car.CollideRect, w.Player.X, w.Player.Y)
	}

	// 3. Solid prop blocks enemy bullets
	w.Bullets.Count = 0
	w.Bullets.Spawn(Bullet{
		X:      car.CollideRect.X + car.CollideRect.W*0.5,
		Y:      car.CollideRect.Y - 2,
		VX:     0,
		VY:     120,
		Radius: 4,
		MaxAge: 3,
		Kind:   BulletFlash,
	})
	w.Update(InputState{SelectArchetype: -1, ActionDir: -1, ChooseUpgrade: -1}, 0.03)
	if w.Bullets.Count != 0 {
		t.Fatalf("expected bullet hitting solid parked car to be blocked, count=%d", w.Bullets.Count)
	}
}
