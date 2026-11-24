package game

import (
	"math"
	"strings"
	"testing"
)

func TestDistrictEventsAllFourAndEffects(t *testing.T) {
	// 1. DistrictEventZTLJam: BusBays act as Golden ZTL with 4x rewards and +1* Wanted
	w := NewWorldWithSeed(2001)
	w.Girls = nil
	w.FirstBoardingDone = true
	bay := w.BusBays[0]
	bx, by := bay.X+bay.W*0.5, bay.Y+bay.H*0.5

	if w.isPointInZTL(bx, by) {
		t.Fatalf("expected BusBay center NOT to be ZTL before DistrictEventZTLJam")
	}

	w.DrainEvents()
	w.TriggerDistrictEvent(DistrictEventZTLJam)
	if w.DistrictEvent.Active != DistrictEventZTLJam || w.DistrictEvent.Duration != 22.0 || w.DistrictEvent.Timer != 22.0 {
		t.Fatalf("unexpected ZTLJam state: %+v", w.DistrictEvent)
	}
	if !w.isPointInZTL(bx, by) {
		t.Fatalf("expected BusBay center to count as ZTL during DistrictEventZTLJam")
	}
	if !hasPopupText(w, "INGORGO ZTL 4x! AREE BUS DORATE!") {
		t.Fatalf("expected ZTLJam popup, got %+v", w.Popups)
	}
	if DistrictEventTitle(DistrictEventZTLJam, 0) != "INGORGO ZTL 4x" || DistrictEventSubtitle(DistrictEventZTLJam, 0, false) == "" {
		t.Fatalf("unexpected ZTLJam title/subtitle")
	}

	sawGoldenRing := false
	for _, r := range w.WaveRings {
		if r.Kind == 1 && (r.X == ArenaWidth*0.5 || r.X == w.Player.X) {
			sawGoldenRing = true
			break
		}
	}
	if !sawGoldenRing {
		t.Fatalf("expected golden WaveRing on player/center on ZTLJam trigger")
	}

	sawStartEvent := false
	for _, ev := range w.DrainEvents() {
		if ev.Kind == EventDistrictEventStart && ev.Value == int(DistrictEventZTLJam) {
			sawStartEvent = true
			break
		}
	}
	if !sawStartEvent {
		t.Fatalf("expected EventDistrictEventStart for DistrictEventZTLJam")
	}

	w.Player.X = bx
	w.Player.Y = by + 40
	w.tryThrowBusStop(bx, by)
	if len(w.Stops) != 1 || !w.Stops[0].GoldenZTL || !w.Stops[0].InZTL || !w.Stops[0].InBusBay {
		t.Fatalf("expected GoldenZTL+InZTL+InBusBay stop on BusBay during ZTLJam, got %+v", w.Stops)
	}
	if w.Police.Stars != 1 {
		t.Fatalf("expected +1* Wanted on GoldenZTL throw, got %d", w.Police.Stars)
	}
	if !hasPopupText(w, "INGORGO ZTL 4x! +1* WANTED!") {
		t.Fatalf("expected INGORGO ZTL 4x! +1* WANTED! popup, got %+v", w.Popups)
	}

	// Verify 4x XP/Cash/Score reward on GoldenZTL boarding (reset Stars=0 to check exact 4x base)
	w.Police.Stars = 0
	w.Player.XP = 0
	w.Player.Cash = 0
	w.Score = 0
	g := NewGirl(9101, bx, by, ArchEmo, false, w.rng)
	g.HookedByStop = w.Stops[0].ID
	w.Girls = []Girl{g}
	w.Stops[0].HookedGirlIDs = []int{g.ID}
	w.completeStopBoarding(&w.Stops[0])
	// Base XP=26*4=104 (lv1->2 costs 50, leaves 54 XP), Cash=18*4=72, Score=100*4=400
	if w.Player.Level != 2 || w.Player.XP != 54 || w.Player.Cash != 72 || w.Score != 400 {
		t.Fatalf("expected 4x GoldenZTL rewards (Lv=2 XP=54 Cash=72 Score=400), got Lv=%d XP=%d Cash=%d Score=%d",
			w.Player.Level, w.Player.XP, w.Player.Cash, w.Score)
	}

	// Expire ZTLJam
	w.DrainEvents()
	w.updateDistrictEvents(22.0)
	if w.DistrictEvent.Active != DistrictEventNone || w.EventsCompleted != 1 || w.isPointInZTL(bx, by) {
		t.Fatalf("expected ZTLJam to end cleanly, active=%v completed=%d", w.DistrictEvent.Active, w.EventsCompleted)
	}
	sawEndEvent := false
	for _, ev := range w.DrainEvents() {
		if ev.Kind == EventDistrictEventEnd && ev.Value == int(DistrictEventZTLJam) {
			sawEndEvent = true
			break
		}
	}
	if !sawEndEvent {
		t.Fatalf("expected EventDistrictEventEnd for DistrictEventZTLJam")
	}

	// 2. DistrictEventHappyHour: -35% shop discount + 1 free slot spin without Cash or SpinCost hike
	w2 := NewWorldWithSeed(2002)
	w2.OpenShop(0)
	w2.OpenSlotMachine(0)
	if w2.CurrentHealCost() != 35 || w2.CurrentBribeCost() != 100 || w2.CurrentRelicCost() != 120 || w2.CurrentUpgradeCost() != 140 || w2.CurrentSlotCost() != 35 {
		t.Fatalf("unexpected baseline shop/slot costs")
	}

	w2.TriggerDistrictEvent(DistrictEventHappyHour)
	if w2.DistrictEvent.Duration != 20.0 || !w2.DistrictEvent.FreeSlotSpinAvailable {
		t.Fatalf("unexpected HappyHour state: %+v", w2.DistrictEvent)
	}
	if !hasPopupText(w2, "HAPPY HOUR! -35% SHOP + 1 SPIN GRATIS!") {
		t.Fatalf("expected HappyHour popup")
	}
	if w2.CurrentHealCost() != 23 || w2.CurrentBribeCost() != 65 || w2.CurrentRelicCost() != 78 || w2.CurrentUpgradeCost() != 91 {
		t.Fatalf("expected -35%% shop costs (23, 65, 78, 91), got (%d, %d, %d, %d)",
			w2.CurrentHealCost(), w2.CurrentBribeCost(), w2.CurrentRelicCost(), w2.CurrentUpgradeCost())
	}
	w2.Police.Stars = 4
	if w2.CurrentBribeCost() != 117 {
		t.Fatalf("expected 4* HappyHour bribe cost 117, got %d", w2.CurrentBribeCost())
	}
	if w2.CurrentSlotCost() != 0 {
		t.Fatalf("expected CurrentSlotCost=0 during HappyHour free spin, got %d", w2.CurrentSlotCost())
	}
	if !strings.Contains(DistrictEventSubtitle(DistrictEventHappyHour, 0, true), "GRATIS") {
		t.Fatalf("expected HappyHour subtitle to mention free spin when available")
	}

	// Free slot spin works even at Cash == 0 and preserves SpinCost
	w2.Player.Cash = 0
	if !w2.StartSlotSpin() {
		t.Fatalf("expected free HappyHour slot spin to succeed with Cash=0")
	}
	if w2.Player.Cash != 0 || w2.SlotMachines[0].SpinCost != 35 || w2.SlotMachines[0].SpinsUsed != 1 || w2.DistrictEvent.FreeSlotSpinAvailable {
		t.Fatalf("unexpected state after free spin: cash=%d spinCost=%d spins=%d freeAvail=%v",
			w2.Player.Cash, w2.SlotMachines[0].SpinCost, w2.SlotMachines[0].SpinsUsed, w2.DistrictEvent.FreeSlotSpinAvailable)
	}
	w2.SlotModal.Spinning = false
	if w2.CurrentSlotCost() != 35 {
		t.Fatalf("expected slot cost to revert to 35 after consuming free spin, got %d", w2.CurrentSlotCost())
	}
	if w2.StartSlotSpin() {
		t.Fatalf("expected 2nd spin with Cash=0 to be rejected after free spin consumed")
	}

	// Buy discounted Espresso ($23)
	w2.Player.HP = 3
	w2.Player.Cash = 23
	inBuy := baseTestInput()
	inBuy.ShopAction = ShopActionHeal
	w2.HandleShopInput(inBuy)
	if w2.Player.Cash != 0 || w2.Player.HP != 5 {
		t.Fatalf("expected $23 discounted Espresso purchase to leave Cash=0 HP=5, got Cash=%d HP=%d", w2.Player.Cash, w2.Player.HP)
	}
	w2.updateDistrictEvents(20.0)
	if w2.CurrentBribeCost() != 180 {
		t.Fatalf("expected bribe cost to return to 180 after HappyHour ended, got %d", w2.CurrentBribeCost())
	}

	// 3. DistrictEventFlashmob: spawns 8..10 FlashmobArch girls around (640,352), respects <34 cap, 80% spawner bias
	w3 := NewWorldWithSeed(2003)
	w3.Girls = nil
	w3.WaveRings = nil
	w3.TriggerDistrictEvent(DistrictEventFlashmob)
	if w3.DistrictEvent.Duration != 25.0 || len(w3.Girls) < 8 || len(w3.Girls) > 10 {
		t.Fatalf("expected Flashmob duration 25s and 8..10 spawned girls, got dur=%v count=%d",
			w3.DistrictEvent.Duration, len(w3.Girls))
	}
	flashArch := w3.DistrictEvent.FlashmobArch
	for i, gr := range w3.Girls {
		if gr.Arch != flashArch {
			t.Fatalf("flashmob girl %d has arch %d, want %d", i, gr.Arch, flashArch)
		}
		d := dist(gr.X, gr.Y, 640, 352)
		if d < 30.0 || d > 140.0 {
			t.Fatalf("flashmob girl %d distance from center = %.1f, expected ~40..130", i, d)
		}
	}
	if len(w3.WaveRings) == 0 || w3.WaveRings[0].X != 640 || w3.WaveRings[0].Y != 352 {
		t.Fatalf("expected WaveRing at central intersection (640,352), got %+v", w3.WaveRings)
	}
	if !strings.Contains(DistrictEventTitle(DistrictEventFlashmob, flashArch), "FLASHMOB") {
		t.Fatalf("expected Flashmob title to contain FLASHMOB")
	}

	// Verify <34 girl cap
	w3Cap := NewWorldWithSeed(2004)
	w3Cap.Girls = nil
	for i := 0; i < 30; i++ {
		w3Cap.Girls = append(w3Cap.Girls, NewGirl(i+1, 200, 200, ArchEmo, false, w3Cap.rng))
	}
	w3Cap.TriggerDistrictEvent(DistrictEventFlashmob)
	if len(w3Cap.Girls) != 34 {
		t.Fatalf("expected Flashmob spawn to obey len(w.Girls) <= 34 cap, got %d", len(w3Cap.Girls))
	}

	// Verify updateGirlSpawner biases heavily toward FlashmobArch
	w3Spawn := NewWorldWithSeed(2005)
	w3Spawn.TriggerDistrictEvent(DistrictEventFlashmob)
	targetArch := w3Spawn.DistrictEvent.FlashmobArch
	w3Spawn.Girls = nil
	w3Spawn.RunTime = 200 // high targetCap
	matchCount := 0
	const trials = 60
	for i := 0; i < trials; i++ {
		w3Spawn.Girls = w3Spawn.Girls[:0]
		w3Spawn.girlSpawnTimer = 0
		w3Spawn.updateGirlSpawner(0.05)
		if len(w3Spawn.Girls) == 1 && w3Spawn.Girls[0].Arch == targetArch {
			matchCount++
		}
	}
	if matchCount < 40 {
		t.Fatalf("expected ~80%% Flashmob spawner bias over %d trials, got %d/%d", trials, matchCount, trials)
	}

	// 4. DistrictEventBlackout: +50% Chronos on start, accumulates score/XP, awards max(150, score/2) & max(25, xp/2) on end
	w4 := NewWorldWithSeed(2006)
	w4.Girls = nil
	w4.FirstBoardingDone = true
	w4.Player.ChronosMeter = 20.0
	w4.Player.ChronosMax = 100.0
	w4.TriggerDistrictEvent(DistrictEventBlackout)
	if w4.DistrictEvent.Duration != 18.0 || w4.Player.ChronosMeter != 70.0 {
		t.Fatalf("expected Blackout duration=18s and Chronos=70.0, got dur=%v chronos=%v",
			w4.DistrictEvent.Duration, w4.Player.ChronosMeter)
	}
	if !hasPopupText(w4, "BLACKOUT NEON! +50% CHRONOS & +50% BONUS!") {
		t.Fatalf("expected Blackout start popup")
	}

	// Minimum floor payout (earned 0 -> +150 Score, +25 XP)
	w4.Score = 100
	w4.Player.XP = 0
	w4.updateDistrictEvents(18.0)
	if w4.Score != 250 || w4.Player.XP != 25 {
		t.Fatalf("expected minimum Blackout end bonus (+150 Score -> 250, +25 XP -> 25), got Score=%d XP=%d",
			w4.Score, w4.Player.XP)
	}
	if !hasPopupText(w4, "BLACKOUT SUPERATO! +50% SCORE & XP!") {
		t.Fatalf("expected Blackout end popup")
	}

	// Earned > floor payout via completeStopBoarding accumulation
	w4.TriggerDistrictEvent(DistrictEventBlackout)
	w4.Score = 0
	w4.Player.XP = 0
	w4.Player.Level = 1
	w4.Player.NextLevelXP = 500
	st := NewBusStop(50, 500, 350, ArchEmo, 1.0, true, false, false, w4.rng)
	for i := 0; i < 4; i++ {
		gr := NewGirl(9500+i, 500, 350, ArchKawaii, false, w4.rng)
		gr.HookedByStop = st.ID
		w4.Girls = append(w4.Girls, gr)
		st.HookedGirlIDs = append(st.HookedGirlIDs, gr.ID)
	}
	w4.completeStopBoarding(&st)
	// 4 girls in ZTL: score = 4 * 100 * 3 = 1200, xp = 4 * int(26*1.5=39) = 156
	if w4.DistrictEvent.BlackoutScoreEarned != 1200 || w4.DistrictEvent.BlackoutXPEarned != 156 {
		t.Fatalf("expected accumulated Blackout (1200 score, 156 XP), got (%d, %d)",
			w4.DistrictEvent.BlackoutScoreEarned, w4.DistrictEvent.BlackoutXPEarned)
	}
	scoreBeforeEnd := w4.Score
	xpBeforeEnd := w4.Player.XP
	w4.updateDistrictEvents(18.0)
	if w4.Score != scoreBeforeEnd+600 || w4.Player.XP != xpBeforeEnd+78 {
		t.Fatalf("expected +600 Score and +78 XP Blackout bonus, got Score=%d XP=%d", w4.Score, w4.Player.XP)
	}
}

func TestDistrictEventSchedulingGuards(t *testing.T) {
	// 1. Skipped in TutorialMode & BossRushMode
	wTut := NewTutorialWorld(nil)
	wTut.DistrictEvent.NextEventTimer = 0.01
	wTut.updateDistrictEvents(1.0)
	if wTut.DistrictEvent.Active != DistrictEventNone {
		t.Fatalf("expected TutorialMode to skip district events")
	}

	wRush := NewWorldWithOptions(2010, true)
	wRush.DistrictEvent.NextEventTimer = 0.01
	wRush.updateDistrictEvents(1.0)
	if wRush.DistrictEvent.Active != DistrictEventNone {
		t.Fatalf("expected BossRushMode to skip district events")
	}

	// 2. Boss.Active prevents NextEventTimer from dropping below 8.0 and blocks trigger
	w := NewWorldWithSeed(2011)
	if w.DistrictEvent.NextEventTimer != 65.0 {
		t.Fatalf("expected initial NextEventTimer=65.0, got %v", w.DistrictEvent.NextEventTimer)
	}
	w.Boss.Active = true
	w.DistrictEvent.NextEventTimer = 2.0
	w.updateDistrictEvents(1.0)
	if w.DistrictEvent.Active != DistrictEventNone || w.DistrictEvent.NextEventTimer < 8.0 {
		t.Fatalf("expected Boss.Active to clamp NextEventTimer >= 8.0 and block trigger, got active=%v timer=%v",
			w.DistrictEvent.Active, w.DistrictEvent.NextEventTimer)
	}
	w.Boss.Active = false

	// 3. Upcoming scheduled boss within < 15s postpones NextEventTimer to 12.0
	w.RunTime = w.BossScheduleTimes[0] - 10.0 // 10s before boss 0
	w.DistrictEvent.NextEventTimer = 0.05
	w.updateDistrictEvents(0.1)
	if w.DistrictEvent.Active != DistrictEventNone || w.DistrictEvent.NextEventTimer != 12.0 {
		t.Fatalf("expected upcoming boss (<15s) to postpone NextEventTimer to 12.0, got active=%v timer=%v",
			w.DistrictEvent.Active, w.DistrictEvent.NextEventTimer)
	}

	// 4. Sequential events never repeat LastEvent back-to-back & StopSafeTopY adjusts while active
	w.BossSchedule = nil
	w.BossScheduleTimes = nil
	if w.StopSafeTopY() != 30 {
		t.Fatalf("expected StopSafeTopY=30 with no schedule and no event, got %v", w.StopSafeTopY())
	}
	var prev DistrictEventKind
	for step := 0; step < 12; step++ {
		w.DistrictEvent.NextEventTimer = 0.01
		w.updateDistrictEvents(0.05)
		cur := w.DistrictEvent.Active
		if cur < DistrictEventZTLJam || cur > DistrictEventBlackout {
			t.Fatalf("step %d: invalid triggered event %v", step, cur)
		}
		if cur == prev {
			t.Fatalf("step %d: repeated LastEvent %v back-to-back", step, cur)
		}
		if w.StopSafeTopY() != 52 {
			t.Fatalf("expected StopSafeTopY=52 while district event %v active, got %v", cur, w.StopSafeTopY())
		}
		prev = cur
		w.updateDistrictEvents(w.DistrictEvent.Timer + 0.1)
		if w.DistrictEvent.NextEventTimer < 60.0 || w.DistrictEvent.NextEventTimer > 90.0 {
			t.Fatalf("expected NextEventTimer in [60,90], got %v", w.DistrictEvent.NextEventTimer)
		}
	}
}

func TestOnboardingFirst60sAndRunTracking(t *testing.T) {
	// 1. Empty throw within first 60s before FirstBoardingDone refunds -40% CD & shows CONSIGLIO tip once
	w := NewWorldWithSeed(2020)
	w.Girls = nil
	w.RunTime = 15.0
	w.tryThrowBusStop(w.Player.X+60, w.Player.Y)

	expectedCD := w.Player.EffectiveStopCooldown() * 0.60
	if math.Abs(float64(w.Player.StopCooldownTimer-expectedCD)) > 0.01 {
		t.Fatalf("expected -40%% CD refund on early empty throw (want ~%.2f), got %.2f", expectedCD, w.Player.StopCooldownTimer)
	}
	if !hasPopupText(w, "LANCIO A VUOTO (-40% CD)") {
		t.Fatalf("expected LANCIO A VUOTO (-40%% CD) popup")
	}
	if !w.EarlyMissTipShown || w.QuipSpeaker != "CONSIGLIO" || w.QuipTimer <= 0 {
		t.Fatalf("expected CONSIGLIO quip on first early empty throw, shown=%v speaker=%q timer=%v",
			w.EarlyMissTipShown, w.QuipSpeaker, w.QuipTimer)
	}

	// Second empty throw does NOT re-trigger the CONSIGLIO quip
	w.QuipTimer = 0
	w.Stops = nil
	w.Player.StopCooldownTimer = 0
	w.tryThrowBusStop(w.Player.X+60, w.Player.Y)
	if w.QuipTimer != 0 {
		t.Fatalf("expected CONSIGLIO tip not to repeat after EarlyMissTipShown=true")
	}

	// 2. First boarding within first 60s adds +15 XP, resets StopCooldownTimer=0 & restores 1 StopCharge
	w.Stops = nil
	w.Player.StopCooldownTimer = 4.2
	w.Player.StopCharges = 0
	w.Player.XP = 0
	w.Police.Stars = 0
	g := NewGirl(8801, w.Player.X+50, w.Player.Y, ArchGyaru, false, w.rng)
	st := NewBusStop(10, g.X, g.Y, ArchGyaru, 1.0, false, false, true, w.rng)
	g.HookedByStop = st.ID
	w.Girls = []Girl{g}
	st.HookedGirlIDs = []int{g.ID}

	w.completeStopBoarding(&st)
	if !w.FirstBoardingDone {
		t.Fatalf("expected FirstBoardingDone=true after boarding")
	}
	// Base XP 26 + onboarding 15 = 41 XP
	if w.Player.XP != 41 {
		t.Fatalf("expected 41 XP (26 + 15 first boarding bonus), got %d", w.Player.XP)
	}
	if w.Player.StopCooldownTimer != 0 || w.Player.StopCharges != 1 {
		t.Fatalf("expected StopCooldownTimer=0 and StopCharges=1 on first boarding, got cd=%v charges=%d",
			w.Player.StopCooldownTimer, w.Player.StopCharges)
	}
	if !hasPopupText(w, "PRIMO IMBARCO! +15XP & FERMATA PRONTA!") {
		t.Fatalf("expected PRIMO IMBARCO! popup")
	}
	if w.BoardedByArch[ArchGyaru] != 1 {
		t.Fatalf("expected BoardedByArch[ArchGyaru]=1, got %v", w.BoardedByArch)
	}

	// Subsequent boarding does NOT grant +15 XP or reset cooldown
	w.Player.StopCooldownTimer = 3.0
	w.Player.XP = 0
	g2 := NewGirl(8802, w.Player.X+50, w.Player.Y, ArchStudent, false, w.rng)
	st2 := NewBusStop(11, g2.X, g2.Y, ArchStudent, 1.0, false, false, false, w.rng)
	g2.HookedByStop = st2.ID
	w.Girls = []Girl{g2}
	st2.HookedGirlIDs = []int{g2.ID}
	w.completeStopBoarding(&st2)
	if w.Player.XP != 26 || w.Player.StopCooldownTimer != 3.0 || w.BoardedByArch[ArchStudent] != 1 {
		t.Fatalf("expected no duplicate first-boarding bonus and Student count=1, got XP=%d CD=%v arch=%v",
			w.Player.XP, w.Player.StopCooldownTimer, w.BoardedByArch)
	}

	// 3. No onboarding bonus after 60s or in BossRushMode
	wLate := NewWorldWithSeed(2021)
	wLate.Girls = nil
	wLate.RunTime = 61.0
	wLate.tryThrowBusStop(wLate.Player.X+60, wLate.Player.Y)
	if wLate.EarlyMissTipShown || math.Abs(float64(wLate.Player.StopCooldownTimer-wLate.Player.EffectiveStopCooldown())) > 0.01 {
		t.Fatalf("expected no empty-throw onboarding refund after 60s")
	}

	// 4. PeakWanted, PerfectDrops, Seed & FormatRunSeed
	wTrack := NewWorldWithSeed(0x12AB34CD)
	if wTrack.Seed != 0x12AB34CD || FormatRunSeed(wTrack.Seed) != "#AB34CD" {
		t.Fatalf("unexpected Seed=%x formatted=%s", wTrack.Seed, FormatRunSeed(wTrack.Seed))
	}
	wTrack.Police.AddStar(4)
	wTrack.Update(baseTestInput(), 0.016)
	wTrack.Police.ReduceStar(3)
	wTrack.Update(baseTestInput(), 0.016)
	if wTrack.PeakWanted != 4 {
		t.Fatalf("expected PeakWanted=4 preserved after reducing stars, got %d", wTrack.PeakWanted)
	}
	wTrack.Girls = []Girl{NewGirl(700, wTrack.Player.X+40, wTrack.Player.Y, ArchKawaii, false, wTrack.rng)}
	wTrack.tryThrowBusStop(wTrack.Girls[0].X, wTrack.Girls[0].Y)
	if wTrack.PerfectDrops != 1 {
		t.Fatalf("expected PerfectDrops=1 after perfect throw, got %d", wTrack.PerfectDrops)
	}
}

func TestQuipsAndShowQuipLifecycle(t *testing.T) {
	// 1. All 6 Boss intro/break/defeat quips are populated and concise
	for k := BossKind(0); k < BossCount; k++ {
		intro := BossIntroQuip(k)
		if intro == "" || len(intro) > 48 {
			t.Fatalf("BossKind %d intro quip invalid length (%d): %q", k, len(intro), intro)
		}
		brk := BossBreakQuip(k, 2)
		def := BossBreakQuip(k, 0)
		if brk == "" || def == "" || brk == def {
			t.Fatalf("BossKind %d break/defeat quips should be distinct and non-empty: brk=%q def=%q", k, brk, def)
		}
	}

	// 2. All 4 SynergyQuip values
	for a := Archetype(0); a < ArchCount; a++ {
		if SynergyQuip(a) == "" {
			t.Fatalf("expected non-empty SynergyQuip for arch %d", a)
		}
	}

	// 3. EndRunQuip branches
	if EndRunQuip(true, true, 3, 6, 4) == "" ||
		EndRunQuip(true, false, 5, 4, 2) == "" ||
		EndRunQuip(true, false, 2, 4, 5) == "" ||
		EndRunQuip(false, false, 4, 1, 2) == "" ||
		EndRunQuip(false, false, 1, 0, 1) == "" {
		t.Fatalf("expected all EndRunQuip variants to be non-empty")
	}

	// 4. ForceSpawnBoss triggers BossIntroQuip, completeStopBoarding triggers BossBreakQuip, QuipTimer decays
	w := NewWorldWithSeed(2030)
	w.Girls = nil
	w.ForceSpawnBoss(BossProfessor)
	if w.QuipSpeaker != "La Professoressa" || w.QuipText != BossIntroQuip(BossProfessor) || w.QuipTimer <= 0 {
		t.Fatalf("expected BossIntroQuip on ForceSpawnBoss, got speaker=%q text=%q timer=%v",
			w.QuipSpeaker, w.QuipText, w.QuipTimer)
	}
	w.Update(baseTestInput(), 0.5)
	if w.QuipTimer >= 3.8 {
		t.Fatalf("expected QuipTimer to decay during World.Update, got %v", w.QuipTimer)
	}

	// Boss shield break quip vs defeat quip
	st := NewBusStop(1, w.Boss.X, w.Boss.Y, ArchEmo, 1.0, false, false, false, w.rng)
	st.HookedBoss = true
	w.Boss.HookedByStop = st.ID
	w.completeStopBoarding(&st)
	if w.QuipSpeaker != "La Professoressa" || w.QuipText != BossBreakQuip(BossProfessor, 2) {
		t.Fatalf("expected BossBreakQuip(2) after 1st shield break, got speaker=%q text=%q", w.QuipSpeaker, w.QuipText)
	}

	w.Boss.Shields = 1
	st2 := NewBusStop(2, w.Boss.X, w.Boss.Y, ArchEmo, 1.0, false, false, false, w.rng)
	st2.HookedBoss = true
	w.Boss.HookedByStop = st2.ID
	w.completeStopBoarding(&st2)
	if w.QuipSpeaker != "La Professoressa" || w.QuipText != BossBreakQuip(BossProfessor, 0) {
		t.Fatalf("expected BossBreakQuip(0) on defeat, got speaker=%q text=%q", w.QuipSpeaker, w.QuipText)
	}

	// 5. SynergyQuip triggers on matching stop and respects 9s cooldown and !Boss.Active
	wSyn := NewWorldWithSeed(2031)
	wSyn.Player.SelectedStopArch = ArchKawaii
	wSyn.Girls = []Girl{NewGirl(1, wSyn.Player.X+40, wSyn.Player.Y, ArchKawaii, false, wSyn.rng)}
	wSyn.tryThrowBusStop(wSyn.Girls[0].X, wSyn.Girls[0].Y)
	if wSyn.QuipText != SynergyQuip(ArchKawaii) || wSyn.synergyQuipCD != 9.0 {
		t.Fatalf("expected SynergyQuip(ArchKawaii) and 9.0s cooldown, got text=%q cd=%v", wSyn.QuipText, wSyn.synergyQuipCD)
	}
}
