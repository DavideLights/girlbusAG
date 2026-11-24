package game

import (
	"math"
	"math/rand"
	"testing"
)

func hasPopupText(w *World, want string) bool {
	for _, p := range w.Popups {
		if p.Text == want {
			return true
		}
	}
	return false
}

func TestCompassClampingAndSafeTopY(t *testing.T) {
	w := NewWorldWithSeed(701)
	if got := w.StopSafeTopY(); got != 52 {
		t.Fatalf("normal run StopSafeTopY = %v, want 52", got)
	}

	w.BossSchedule = nil
	if got := w.StopSafeTopY(); got != 30 {
		t.Fatalf("empty schedule StopSafeTopY = %v, want 30", got)
	}

	w.OverdriveActive = true
	if got := w.StopSafeTopY(); got != 52 {
		t.Fatalf("Overdrive StopSafeTopY = %v, want 52", got)
	}

	wTut := NewTutorialWorld(nil)
	if got := wTut.StopSafeTopY(); got != 68 {
		t.Fatalf("Tutorial StopSafeTopY = %v, want 68", got)
	}

	sx, sy := ClampCompassScreenPos(-40, 10, 52)
	if sx != 36 || sy != 66 {
		t.Fatalf("ClampCompassScreenPos(-40,10,52) = (%v,%v), want (36,66)", sx, sy)
	}

	sx2, sy2 := ClampCompassScreenPos(ScreenWidth+90, ScreenHeight+40, 68)
	if sx2 != ScreenWidth-36 || sy2 != ScreenHeight-64 {
		t.Fatalf("ClampCompassScreenPos(max) = (%v,%v), want (%v,%v)", sx2, sy2, ScreenWidth-36, ScreenHeight-64)
	}
}

func TestOverdriveCyclicBossLoop(t *testing.T) {
	w := NewWorldWithSeed(702)
	w.Girls = nil
	w.BossScheduleIdx = len(w.BossSchedule) // exhausted normal schedule

	w.EnterOverdrive()
	if !w.OverdriveActive || w.NextOverdriveBossTimer != 30.0 {
		t.Fatalf("expected OverdriveActive=true NextOverdriveBossTimer=30, got (%v, %v)", w.OverdriveActive, w.NextOverdriveBossTimer)
	}
	if !hasPopupText(w, "MODALITA OVERDRIVE!") {
		t.Fatalf("expected MODALITA OVERDRIVE! popup")
	}

	// Verify cyclic spawns through all 6 BossKinds and wrap-around
	for wave := 1; wave <= int(BossCount)+1; wave++ {
		w.checkTimedBossSpawns(w.NextOverdriveBossTimer)
		if !w.Boss.Active {
			t.Fatalf("wave %d: expected boss to spawn", wave)
		}
		wantKind := BossKind((wave - 1) % int(BossCount))
		if w.Boss.Kind != wantKind {
			t.Fatalf("wave %d: got BossKind %d, want %d", wave, w.Boss.Kind, wantKind)
		}
		if w.OverdriveWave != wave || w.NextOverdriveBossTimer != 55.0 {
			t.Fatalf("wave %d: got OverdriveWave=%d timer=%v, want wave=%d timer=55", wave, w.OverdriveWave, w.NextOverdriveBossTimer, wave)
		}

		// While boss is active, timer must NOT count down or spawn another boss
		w.checkTimedBossSpawns(20.0)
		if w.NextOverdriveBossTimer != 55.0 {
			t.Fatalf("expected NextOverdriveBossTimer to stay 55.0 while boss active, got %v", w.NextOverdriveBossTimer)
		}

		// Clear active boss for next wave
		w.Boss.Active = false
	}

	// Defeating an Overdrive boss must not re-trigger Victory
	w.Victory = false
	w.ForceSpawnBoss(BossMom)
	w.Boss.Shields = 1
	st := NewBusStop(90, w.Boss.X, w.Boss.Y, ArchEmo, 1.0, false, false, false, w.rng)
	st.HookedBoss = true
	w.Boss.HookedByStop = st.ID
	w.completeStopBoarding(&st)
	if w.Victory {
		t.Fatalf("expected Overdrive boss defeat NOT to re-set Victory=true")
	}
}

func TestEmoSoulmateBlackHeartAndSynergyPopups(t *testing.T) {
	w := NewWorldWithSeed(703)
	w.Popups = nil
	w.HealPickups = nil
	w.Player.SelectedStopArch = ArchEmo
	w.Girls = []Girl{NewGirl(1, w.Player.X+50, w.Player.Y, ArchEmo, false, w.rng)}

	in := baseTestInput()
	in.ThrowJustPressed = true
	in.AimWorldX = w.Girls[0].X
	in.AimWorldY = w.Girls[0].Y
	w.Update(in, 0.016)

	if len(w.Stops) != 1 || !w.Stops[0].SynergyActive {
		t.Fatalf("expected Emo stop with SynergyActive=true")
	}
	if !hasPopupText(w, "SOULMATE! +1 SCUDO & CUORE") {
		t.Fatalf("expected SOULMATE! +1 SCUDO & CUORE popup, got %+v", w.Popups)
	}

	w.Popups = nil
	w.completeStopBoarding(&w.Stops[0])
	if !hasPopupText(w, "BLACK HEART! (+1 HP)") {
		t.Fatalf("expected BLACK HEART! (+1 HP) popup, got %+v", w.Popups)
	}

	foundBlackHeart := false
	for _, hp := range w.HealPickups {
		if hp.BlackHeart && hp.Amount == 1 && hp.Chronos == 25.0 && hp.LifeTimer == 45.0 {
			foundBlackHeart = true
			break
		}
	}
	if !foundBlackHeart {
		t.Fatalf("expected BlackHeart HealPickup (1 HP, 25 Chronos, 45s) on Emo SOULMATE boarding, got %+v", w.HealPickups)
	}

	// Verify clear synergy popups for Student, Kawaii, Gyaru
	cases := []struct {
		arch Archetype
		want string
	}{
		{ArchStudent, "HONOR ROLL! +XP & FREEZE"},
		{ArchKawaii, "MOE CRIT! GRAZE x2"},
		{ArchGyaru, "VIRAL #1! BARRA +35%"},
	}
	for _, tc := range cases {
		w2 := NewWorldWithSeed(704)
		w2.Popups = nil
		w2.Player.SelectedStopArch = tc.arch
		w2.Girls = []Girl{NewGirl(1, w2.Player.X+50, w2.Player.Y, tc.arch, false, w2.rng)}
		in2 := baseTestInput()
		in2.ThrowJustPressed = true
		in2.AimWorldX = w2.Girls[0].X
		in2.AimWorldY = w2.Girls[0].Y
		w2.Update(in2, 0.016)
		if !hasPopupText(w2, tc.want) {
			t.Fatalf("arch %d: expected popup %q, got %+v", tc.arch, tc.want, w2.Popups)
		}
	}
}

func TestPolice3StarHazardsAndDamage(t *testing.T) {
	w := NewWorldWithSeed(705)
	w.Girls = nil
	w.Police.AddStar(3)
	w.Police.HazardTimer = 0.01

	w.Police.Update(0.05, 1.0, w.Player.X, w.Player.Y, w.Stops, &w.Bullets, &w.Telegraphs, w.rng)
	if len(w.Police.Hazards) != 1 {
		t.Fatalf("expected 1 PoliceHazard spawned at 3★, got %d", len(w.Police.Hazards))
	}
	h := w.Police.Hazards[0]
	if math.Abs(float64((h.X2-h.X1)-96)) > 0.01 {
		t.Fatalf("expected hazard width 96, got %v", h.X2-h.X1)
	}
	if h.X1 < 96 || h.X2 > ArenaWidth-96 || h.Y1 < 148 || h.Y1 > BoardwalkRailY-28 {
		t.Fatalf("hazard out of clamped bounds: %+v", h)
	}
	if h.TelegraphTimer <= 0 || h.LifeTimer != 8.0 {
		t.Fatalf("unexpected initial hazard timers: %+v", h)
	}

	// During telegraph, hazard does not damage Alfredo
	w.Player.X = (h.X1 + h.X2) * 0.5
	w.Player.Y = h.Y1
	startHP := w.Player.HP
	w.handleBulletInteractions()
	if w.Player.HP != startHP {
		t.Fatalf("expected no damage while hazard telegraph is active")
	}

	// Once telegraph expires, standing on hazard damages Alfredo with BulletPolice slow
	w.Police.Hazards[0].TelegraphTimer = 0
	w.handleBulletInteractions()
	if w.Player.HP != startHP-1 || w.Player.SlowTimer <= 0 {
		t.Fatalf("expected active hazard to deal 1 HP + slow, HP=%d/%d slow=%v", w.Player.HP, startHP, w.Player.SlowTimer)
	}

	// Reducing stars below 3 clears hazards
	w.Police.ReduceStar(1)
	if len(w.Police.Hazards) != 0 {
		t.Fatalf("expected hazards cleared when stars < 3, got %d", len(w.Police.Hazards))
	}
}

func TestSWATAntiStopAndShieldFlash(t *testing.T) {
	// 1. Regular downward bullet blocked by back-glass shield sets ShieldFlashTimer and records event
	w := NewWorldWithSeed(706)
	w.Girls = nil
	stop := NewBusStop(1, 500, 300, ArchEmo, 1.0, false, false, false, w.rng)
	stop.ShieldHP = 1
	stop.MaxShieldHP = 1
	w.Stops = append(w.Stops, stop)
	w.Bullets.Spawn(Bullet{X: 500, Y: 280, VX: 0, VY: 120, Radius: 4, MaxAge: 3, Kind: BulletFlash})
	w.Update(baseTestInput(), 0.016)
	if w.Stops[0].ShieldHP != 0 || w.Stops[0].ShieldFlashTimer <= 0 {
		t.Fatalf("expected regular shield hit to drop HP=0 and set ShieldFlashTimer>0, got hp=%d flash=%v",
			w.Stops[0].ShieldHP, w.Stops[0].ShieldFlashTimer)
	}
	sawBreak := false
	for _, ev := range w.DrainEvents() {
		if ev.Kind == EventStopShieldHit && ev.Flag {
			sawBreak = true
			break
		}
	}
	if !sawBreak {
		t.Fatalf("expected EventStopShieldHit(Broken=true) when shield hits 0")
	}

	// 2. At 4★, SWAT officers target active BusStop and fire AntiStop bullets
	w2 := NewWorldWithSeed(707)
	w2.Girls = nil
	w2.Police.AddStar(4)
	st := NewBusStop(1, 600, 350, ArchKawaii, 1.0, false, false, false, w2.rng)
	st.ShieldHP = 1
	st.MaxShieldHP = 1
	w2.Stops = append(w2.Stops, st)
	w2.Police.Officers = []PoliceOfficer{{
		ID:         1,
		X:          600,
		Y:          450,
		ShootTimer: 0.01,
		HP:         6,
		IsSWAT:     true,
		Active:     true,
	}}
	w2.Police.Update(0.05, 1.0, 200, 200, w2.Stops, &w2.Bullets, &w2.Telegraphs, w2.rng)
	if w2.Bullets.Count == 0 || !w2.Bullets.Items[0].AntiStop || w2.Bullets.Items[0].Kind != BulletPolice {
		t.Fatalf("expected 4★ SWAT officer to spawn AntiStop BulletPolice, count=%d", w2.Bullets.Count)
	}

	// 3. AntiStop bullet hitting from BELOW (VY < 0) within 26px damages shield, then shaves 2.5s off LifeTimer
	w2.Bullets.Count = 0
	w2.DrainEvents()
	w2.Bullets.Spawn(Bullet{X: 600, Y: 365, VX: 0, VY: -120, Radius: 4, MaxAge: 3, Kind: BulletPolice, AntiStop: true})
	w2.Update(baseTestInput(), 0.016)
	if w2.Stops[0].ShieldHP != 0 || w2.Stops[0].ShieldFlashTimer <= 0 {
		t.Fatalf("expected AntiStop bullet from below to break shield, got hp=%d flash=%v",
			w2.Stops[0].ShieldHP, w2.Stops[0].ShieldFlashTimer)
	}

	lifeBefore := w2.Stops[0].LifeTimer
	w2.Bullets.Spawn(Bullet{X: 600, Y: 365, VX: 0, VY: -120, Radius: 4, MaxAge: 3, Kind: BulletPolice, AntiStop: true})
	w2.Update(baseTestInput(), 0.016)
	if w2.Stops[0].LifeTimer > lifeBefore-2.4 {
		t.Fatalf("expected AntiStop bullet at 0 shield to shave 2.5s off LifeTimer: before=%v after=%v",
			lifeBefore, w2.Stops[0].LifeTimer)
	}
	if !hasPopupText(w2, "SWAT ATTACCA FERMATA!") {
		t.Fatalf("expected SWAT ATTACCA FERMATA! popup, got %+v", w2.Popups)
	}
}

func TestBossBestFriendKnockbackAndBossLoverExtraZTL(t *testing.T) {
	// 1. BossBestFriend knockback wave on even PatternStep pushes non-dashing Alfredo within 165px
	w := NewWorldWithSeed(708)
	w.Girls = nil
	w.ForceSpawnBoss(BossBestFriend)
	w.Boss.X = 640
	w.Boss.Y = 360
	w.Boss.PatternStep = 1 // Next firePattern increments to 2 (even) -> PendingKnockback=true
	w.Boss.ShootTimer = 0.01
	w.Player.X = 720 // 80px east of boss (within 165px)
	w.Player.Y = 360
	w.Player.DashTimer = 0

	w.Update(baseTestInput(), 0.02)
	if w.Boss.PendingKnockback {
		t.Fatalf("expected World.Update to consume Boss.PendingKnockback")
	}
	if w.Player.X < 745 {
		t.Fatalf("expected BossBestFriend knockback wave to push Alfredo outward (~748), got X=%v", w.Player.X)
	}
	if !hasPopupText(w, "LASCIALA STARE!") {
		t.Fatalf("expected LASCIALA STARE! popup")
	}
	sawRing := false
	for _, r := range w.WaveRings {
		if r.MaxRadius == 165 {
			sawRing = true
			break
		}
	}
	if !sawRing {
		t.Fatalf("expected WaveRing(MaxRadius=165) on BossBestFriend knockback")
	}

	// 2. BossLover defeat adds 2 extra ZTL zones without duplicates
	w2 := NewWorldWithSeed(709)
	w2.Girls = nil
	initialZTL := len(w2.ZTLZones)
	w2.ForceSpawnBoss(BossLover)
	w2.Boss.Shields = 1
	st := NewBusStop(1, w2.Boss.X, w2.Boss.Y, ArchEmo, 1.0, false, false, false, rand.New(rand.NewSource(1)))
	st.HookedBoss = true
	w2.Boss.HookedByStop = st.ID
	w2.completeStopBoarding(&st)

	if len(w2.ZTLZones) != initialZTL+2 {
		t.Fatalf("expected BossLover defeat to add 2 ZTL zones (%d -> %d), got %d",
			initialZTL, initialZTL+2, len(w2.ZTLZones))
	}
	if !hasPopupText(w2, "ZTL RADDOPPIATE!") {
		t.Fatalf("expected ZTL RADDOPPIATE! popup")
	}

	// Defeating BossLover a second time (e.g. Overdrive) must NOT duplicate ZTLs
	w2.EnterOverdrive()
	w2.ForceSpawnBoss(BossLover)
	w2.Boss.Shields = 1
	st2 := NewBusStop(2, w2.Boss.X, w2.Boss.Y, ArchEmo, 1.0, false, false, false, w2.rng)
	st2.HookedBoss = true
	w2.Boss.HookedByStop = st2.ID
	w2.completeStopBoarding(&st2)
	if len(w2.ZTLZones) != initialZTL+2 {
		t.Fatalf("expected no duplicate ZTL zones on repeat BossLover defeat, got %d", len(w2.ZTLZones))
	}
}

func TestDefaultDistrictMapAndWorldMetadata(t *testing.T) {
	m := DefaultDistrictMap()
	if m.ID == "" || m.NameIT == "" || m.SubtitleIT == "" {
		t.Fatalf("expected populated MapDef metadata, got %+v", m)
	}
	if m.Width != ArenaWidth || m.Height != ArenaHeight {
		t.Fatalf("expected MapDef dimensions (%v,%v), got (%v,%v)", ArenaWidth, ArenaHeight, m.Width, m.Height)
	}
	if len(m.UnlockedArchs) != int(ArchCount) {
		t.Fatalf("expected %d UnlockedArchs, got %d", ArchCount, len(m.UnlockedArchs))
	}
	if len(m.ChestSpots) != 3 {
		t.Fatalf("expected 3 ChestSpots, got %d", len(m.ChestSpots))
	}
	if len(m.Layout.Props) == 0 || len(m.Layout.BusBays) != 4 || len(m.Layout.ZTLZones) != 2 {
		t.Fatalf("expected valid wrapped DistrictLayout in DefaultDistrictMap")
	}

	w := NewWorldWithSeed(710)
	if w.MapID != m.ID || w.MapName != m.NameIT || len(w.ChestSpots) != 3 {
		t.Fatalf("expected World map metadata (%s,%s,3 chests), got (%s,%s,%d)",
			m.ID, m.NameIT, w.MapID, w.MapName, len(w.ChestSpots))
	}
}

func TestClickPriorityGirlOverShopAndQueuedLevelUp(t *testing.T) {
	// 1. Clicking near a Vending Shop while aiming at a girl nearby prioritizes throwing a BusStop
	w := NewWorldWithSeed(711)
	w.Girls = nil
	sh := w.Shops[0]
	w.Player.X = sh.X
	w.Player.Y = sh.Y + 60
	w.Girls = append(w.Girls, NewGirl(9001, sh.X+8, sh.Y-6, ArchEmo, false, w.rng))

	in := baseTestInput()
	in.MouseLeftJustPressed = true
	in.ThrowJustPressed = true
	in.AimWorldX = sh.X + 8
	in.AimWorldY = sh.Y - 6
	w.Update(in, 0.016)
	if w.ShopMenu.Active {
		t.Fatalf("expected aiming at girl near shop NOT to open ShopMenu")
	}
	if len(w.Stops) != 1 {
		t.Fatalf("expected click priority to throw 1 BusStop at the girl, got stops=%d", len(w.Stops))
	}

	// 2. Level-Up triggered while Alfredo is inside an active stop queues cleanly and opens after exit
	w2 := NewWorldWithSeed(712)
	w2.Girls = nil
	st := NewBusStop(1, 500, 350, ArchStudent, 1.0, false, false, false, w2.rng)
	st.PlayerInside = true
	w2.Stops = append(w2.Stops, st)
	w2.Player.X = 500
	w2.Player.Y = 350

	if !w2.Player.AddXP(w2.Player.NextLevelXP) {
		t.Fatalf("expected AddXP(NextLevelXP) to signal level up")
	}
	w2.triggerLevelUp()
	if w2.PendingLevelUp {
		t.Fatalf("expected PendingLevelUp=false while inside active stop minigame")
	}
	if w2.QueuedLevelUps != 1 {
		t.Fatalf("expected QueuedLevelUps=1 while inside active stop, got %d", w2.QueuedLevelUps)
	}

	// Deactivate stop -> next Update flushes queued level-up automatically
	w2.Stops[0].Active = false
	w2.Update(baseTestInput(), 0.016)
	if !w2.PendingLevelUp || w2.QueuedLevelUps != 0 {
		t.Fatalf("expected queued level-up to open after stop ends (Pending=%v, Queued=%d)", w2.PendingLevelUp, w2.QueuedLevelUps)
	}
}

func TestStopUIAnchorLocksWithDirectionalMargins(t *testing.T) {
	// 1. Extreme screen corners always guarantee ample directional click space on all 4 sides of PadCenter
	corners := [][2]float32{{0, 0}, {640, 0}, {640, 360}, {0, 360}, {320, 180}}
	for _, c := range corners {
		ax, ay := ComputeStopUIAnchor(c[0], c[1], 52)
		st := BusStop{UIAnchorX: ax, UIAnchorY: ay, UIAnchored: true}
		padX, padY := st.PadCenter()
		if padX < 190 || float32(ScreenWidth)-padX < 70 {
			t.Fatalf("corner %v produced cramped horizontal padX=%v (left=%v, right=%v)", c, padX, padX, float32(ScreenWidth)-padX)
		}
		if padY < 75 || float32(ScreenHeight)-padY < 90 {
			t.Fatalf("corner %v produced cramped vertical padY=%v (top=%v, bottom=%v)", c, padY, padY, float32(ScreenHeight)-padY)
		}
	}

	// 2. Throwing a stop locks its UIAnchor & PadCenter even if camera and Alfredo move afterward
	w := NewWorldWithSeed(713)
	w.Girls = nil
	w.Player.X = 500
	w.Player.Y = 320
	w.CameraX = 120
	w.CameraY = 80
	w.tryThrowBusStop(540, 320)
	if len(w.Stops) != 1 || !w.Stops[0].UIAnchored {
		t.Fatalf("expected 1 anchored stop after throw, stops=%d", len(w.Stops))
	}
	initX, initY := w.Stops[0].PadCenter()

	w.CameraX = 480
	w.CameraY = 260
	w.Player.X = 900
	afterX, afterY := w.Stops[0].PadCenter()
	if afterX != initX || afterY != initY {
		t.Fatalf("expected locked PadCenter (%v,%v) to remain fixed when camera/player move, got (%v,%v)",
			initX, initY, afterX, afterY)
	}
}
