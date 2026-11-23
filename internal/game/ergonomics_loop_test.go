package game

import (
	"math/rand"
	"testing"
)

func baseTestInput() InputState {
	return InputState{
		SelectArchetype: -1,
		ActionDir:       ActionNone,
		ChooseUpgrade:   -1,
		ShopAction:      ShopActionNone,
	}
}

func TestInputBufferDashAndThrow(t *testing.T) {
	w := NewWorldWithSeed(101)
	w.Girls = nil
	w.Player.DashCooldown = 0.08
	w.Player.StopCooldownTimer = 0.08

	// Press Dash and Throw slightly BEFORE cooldowns expire (within 120ms InputBufferDuration).
	in := baseTestInput()
	in.DashJustPressed = true
	in.ThrowJustPressed = true
	in.AimWorldX = w.Player.X + 80
	in.AimWorldY = w.Player.Y
	w.Update(in, 0.02)

	if w.Player.DashTimer > 0 {
		t.Fatalf("expected dash to wait in buffer while DashCooldown > 0")
	}
	if len(w.Stops) != 0 {
		t.Fatalf("expected stop throw to wait in buffer while StopCooldownTimer > 0")
	}

	// Step past remaining 0.06s cooldown -> both buffered actions must trigger automatically!
	for i := 0; i < 6; i++ {
		w.Update(baseTestInput(), 0.02)
	}
	if w.Player.DashCooldown <= 0 {
		t.Fatalf("expected buffered Dash to fire once DashCooldown expired")
	}
	if len(w.Stops) != 1 {
		t.Fatalf("expected buffered Throw to place 1 stop once StopCooldownTimer expired, got %d", len(w.Stops))
	}
}

func TestTetherCoyoteTimeAndGyaruHoldToRise(t *testing.T) {
	// 1. Tether Coyote Time (0.35s)
	w := NewWorldWithSeed(102)
	w.Girls = nil
	stop := NewBusStop(1, 500, 400, ArchGyaru, 1.0, false, false, false, w.rng)
	w.Stops = append(w.Stops, stop)

	// Step inside stop to populate CoyoteTimer
	w.Player.X = 500
	w.Player.Y = 400
	w.Update(baseTestInput(), 0.02)
	if !w.Stops[0].PlayerInside || w.Stops[0].CoyoteTimer <= 0 {
		t.Fatalf("expected PlayerInside=true and positive CoyoteTimer inside stop")
	}

	// Move abruptly outside stop radius: PlayerInside must remain true for 0.35s!
	w.Player.X = 500 + w.Stops[0].CaptureRadius + 30
	w.Update(baseTestInput(), 0.10)
	if !w.Stops[0].PlayerInside {
		t.Fatalf("expected Tether Coyote Time to preserve PlayerInside at 0.10s outside")
	}
	for i := 0; i < 8; i++ {
		w.Update(baseTestInput(), 0.04)
	}
	if w.Stops[0].PlayerInside {
		t.Fatalf("expected PlayerInside to become false once 0.35s Coyote Time elapsed")
	}

	// 2. Gyaru Hold-to-Rise without repeated discrete taps
	rng := rand.New(rand.NewSource(7))
	gy := NewMinigameState(ArchGyaru, false, rng)
	startCursor := gy.BarCursor
	for i := 0; i < 20; i++ {
		gy.Update(0.016, true, ActionNone, true, 108, rng)
	}
	if gy.BarCursor <= startCursor {
		t.Fatalf("expected ActionHeld to smoothly raise Gyaru BarCursor: start=%v got=%v", startCursor, gy.BarCursor)
	}

	// 3. Emo BPM-sync moves cursor faster at higher BPM
	emSlow := NewMinigameState(ArchEmo, false, rng)
	emFast := NewMinigameState(ArchEmo, false, rng)
	emSlow.Update(0.10, true, ActionNone, false, 108, rng)
	emFast.Update(0.10, true, ActionNone, false, 138, rng)
	if emFast.BeatCursor <= emSlow.BeatCursor {
		t.Fatalf("expected 138 BPM Emo cursor (%v) to advance faster than 108 BPM (%v)", emFast.BeatCursor, emSlow.BeatCursor)
	}
}

func TestSmartMagneticAimDirectional(t *testing.T) {
	w := NewWorldWithSeed(103)
	w.Player.X = 400
	w.Player.Y = 300
	w.Player.SelectedStopArch = ArchKawaii
	w.Girls = []Girl{
		NewGirl(1, 470, 325, ArchKawaii, false, w.rng), // matching forward-cone girl
		NewGirl(2, 485, 300, ArchEmo, false, w.rng),    // non-matching girl directly ahead
	}

	in := baseTestInput()
	in.AimIsDirectional = true
	in.AimDirX = 1.0
	in.AimDirY = 0.0

	ax, ay := w.ResolveAim(in)
	if ax != w.Girls[0].X || ay != w.Girls[0].Y {
		t.Fatalf("expected Smart Magnetic Aim to snap onto matching Kawaii girl (%v,%v), got (%v,%v)",
			w.Girls[0].X, w.Girls[0].Y, ax, ay)
	}
}

func TestGlobalGrazeAndDashParryCooldownShave(t *testing.T) {
	w := NewWorldWithSeed(104)
	w.Girls = nil
	w.Player.X = 500
	w.Player.Y = 300
	w.Player.ChronosMeter = 50
	w.Player.StopCooldownTimer = 5.0

	// 1. Global Graze at 20px (within b.Radius+GlobalGrazeRadius=22px, outside parryR=16px and hitR=10px)
	w.Bullets.Spawn(Bullet{X: 520, Y: 300, Radius: 4, MaxAge: 5, Kind: BulletPolice})
	w.handleBulletInteractions()
	if w.Player.ChronosMeter != 54 {
		t.Fatalf("expected +4 Chronos on Global Graze, got %v", w.Player.ChronosMeter)
	}
	if w.Player.StopCooldownTimer != 4.75 {
		t.Fatalf("expected -0.25s StopCooldownTimer on Global Graze, got %v", w.Player.StopCooldownTimer)
	}

	// 2. Dash-Parry directly into bullet removes it, adds HitStop, and cuts -1.0s CD
	w.Player.DashTimer = 0.14
	w.Player.DashParryHitStopDone = false
	w.Player.DashParryShaveUsed = 0
	w.Bullets.Spawn(Bullet{X: 503, Y: 300, Radius: 4, MaxAge: 5, Kind: BulletFlash})
	beforeCount := w.Bullets.Count
	w.handleBulletInteractions()
	if w.Bullets.Count != beforeCount-1 {
		t.Fatalf("expected Dash-Parry to remove parried bullet: before=%d after=%d", beforeCount, w.Bullets.Count)
	}
	if w.HitStopTimer < 0.025 {
		t.Fatalf("expected >=25ms micro Hit-Stop on Dash-Parry, got %v", w.HitStopTimer)
	}
	if w.Player.StopCooldownTimer != 3.75 {
		t.Fatalf("expected -1.0s StopCooldownTimer on Dash-Parry (3.75), got %v", w.Player.StopCooldownTimer)
	}
}

func TestStopRecallRefundsCooldownAndClearsBullets(t *testing.T) {
	w := NewWorldWithSeed(105)
	g := NewGirl(99, 520, 300, ArchEmo, false, w.rng)
	g.HookedByStop = 1
	w.Girls = []Girl{g}

	stop := NewBusStop(1, 520, 300, ArchEmo, 1.0, false, false, false, w.rng)
	stop.HookedGirlIDs = []int{99}
	w.Stops = append(w.Stops, stop)
	w.Player.StopCooldownTimer = 6.0
	startHP := w.Player.HP

	w.Bullets.Spawn(Bullet{X: 545, Y: 300, Radius: 4, MaxAge: 5, Kind: BulletRazor})
	in := baseTestInput()
	in.RecallJustPressed = true
	w.Update(in, 0.02)

	if w.ActiveStopCount() != 0 {
		t.Fatalf("expected recalled stop to deactivate")
	}
	if w.Player.StopCooldownTimer > 3.05 || w.Player.StopCooldownTimer < 2.90 {
		t.Fatalf("expected ~50%% cooldown refund (~2.98-3.00s), got %v", w.Player.StopCooldownTimer)
	}
	if w.Girls[0].HookedByStop != -1 {
		t.Fatalf("expected hooked girl to be gently released (HookedByStop=-1)")
	}
	if w.Bullets.Count != 0 {
		t.Fatalf("expected EMP bullet-clear around recalled stop, remaining=%d", w.Bullets.Count)
	}
	if w.Player.HP != startHP {
		t.Fatalf("expected Stop Recall not to injure player")
	}
}

func TestRizzFeverStreakAndBusSlipstreamAndAlVolo(t *testing.T) {
	w := NewWorldWithSeed(106)
	w.Girls = nil

	// Build Streak up to 5 via 5 completed boardings
	for s := 0; s < 5; s++ {
		st := NewBusStop(s+1, 400, 300, ArchEmo, 1.0, false, false, false, w.rng)
		g := NewGirl(s+10, 400, 300, ArchEmo, false, w.rng)
		g.HookedByStop = st.ID
		w.Girls = append(w.Girls, g)
		st.HookedGirlIDs = []int{g.ID}
		w.completeStopBoarding(&st)
	}
	if w.StreakCount != 5 {
		t.Fatalf("expected StreakCount=5, got %d", w.StreakCount)
	}

	// Rush Hour x5 Perfect Drop cooldown cuts to 0.50x
	w.Girls = []Girl{NewGirl(77, w.Player.X+60, w.Player.Y, ArchEmo, false, w.rng)}
	inThrow := baseTestInput()
	inThrow.ThrowJustPressed = true
	inThrow.AimWorldX = w.Girls[0].X
	inThrow.AimWorldY = w.Girls[0].Y
	w.Update(inThrow, 0.016)
	wantCD := w.Player.EffectiveStopCooldown() * 0.50
	if w.Player.StopCooldownTimer > wantCD+0.02 {
		t.Fatalf("expected Rush Hour x5 Perfect Drop CD <= %v, got %v", wantCD, w.Player.StopCooldownTimer)
	}

	// Rush Hour NeonBus clearing bullets converts each bullet to +$3 Cash & spawns Slipstream
	w.HitStopTimer = 0
	cashBefore := w.Player.Cash
	w.Bullets.Count = 0
	w.Bullets.Spawn(Bullet{X: 350, Y: 360, Radius: 4, MaxAge: 5, Kind: BulletPolice})
	w.Bullets.Spawn(Bullet{X: 365, Y: 360, Radius: 4, MaxAge: 5, Kind: BulletPolice})
	w.Buses = append(w.Buses, NeonBus{Active: true, X: 300, Y: 360, Speed: 900, TargetStopID: 999})
	w.updateBuses(0.1)
	if w.Player.Cash != cashBefore+6 {
		t.Fatalf("expected +$6 Rush Hour bus bullet conversion, got +%d", w.Player.Cash-cashBefore)
	}
	if len(w.Slipstreams) == 0 {
		t.Fatalf("expected NeonBus to emit a BusSlipstream ribbon")
	}

	// Throwing ahead of approaching NeonBus triggers Al Volo >=50% minigame charge
	w.Stops = nil
	w.Player.StopCooldownTimer = 0
	w.tryThrowBusStop(w.Buses[0].X+180, w.Buses[0].Y)
	if len(w.Stops) != 1 || !w.Stops[0].AlVoloBoosted || w.Stops[0].Minigame.Progress < 0.50 {
		t.Fatalf("expected Al Volo boost (>=0.50), got stops=%d alVolo=%v prog=%v",
			len(w.Stops), len(w.Stops) > 0 && w.Stops[0].AlVoloBoosted, func() float32 {
				if len(w.Stops) == 0 {
					return 0
				}
				return w.Stops[0].Minigame.Progress
			}())
	}

	// Walking onto Bus Slipstream resets DashCooldown and sets SlipstreamTimer
	w.Player.DashCooldown = 0.65
	w.Player.X = 200
	w.Player.Y = 360
	w.updateSlipstreams(0.02)
	if w.Player.DashCooldown != 0 || w.Player.SlipstreamTimer <= 0 {
		t.Fatalf("expected Bus Slipstream to reset DashCooldown and activate SlipstreamTimer, dashCD=%v slip=%v",
			w.Player.DashCooldown, w.Player.SlipstreamTimer)
	}
}

func TestBasslineMagnetBeatHerding(t *testing.T) {
	w := NewWorldWithSeed(107)
	w.Player.X = 500
	w.Player.Y = 350
	w.Player.SelectedStopArch = ArchKawaii

	matchGirl := NewGirl(1, 580, 350, ArchKawaii, false, w.rng)  // d=80 in (42, 1.45*72=104.4] -> pulls inward!
	otherGirl := NewGirl(2, 540, 350, ArchEmo, false, w.rng)     // d=40 <= 72 -> pushes outward!
	w.Girls = []Girl{matchGirl, otherGirl}
	w.Police.Officers = []PoliceOfficer{{X: 540, Y: 350}} // d=40 <= 72 -> pushes outward!

	w.triggerBasslinePulse()

	if w.Girls[0].X >= 580 {
		t.Fatalf("expected matching-archetype girl to pull inward toward Alfredo, X stayed %v", w.Girls[0].X)
	}
	if w.Girls[1].X <= 540 {
		t.Fatalf("expected non-matching girl to push outward, X stayed %v", w.Girls[1].X)
	}
	if w.Police.Officers[0].X <= 540 {
		t.Fatalf("expected Police officer to push outward, X stayed %v", w.Police.Officers[0].X)
	}
}
