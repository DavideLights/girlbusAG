package game

import "testing"

func TestShopAntiInfiniteLivesAndEscalatingPrices(t *testing.T) {
	w := NewWorldWithSeed(901)
	w.Player.Cash = 5000

	// 1. Espresso blocked at full HP
	w.OpenShop(0)
	inHeal := baseTestInput()
	inHeal.ShopAction = ShopActionHeal
	w.HandleShopInput(inHeal)
	if w.Player.Cash != 5000 || w.Player.EspressoBuys != 0 {
		t.Fatalf("expected Espresso blocked at Full HP, cash=%d buys=%d", w.Player.Cash, w.Player.EspressoBuys)
	}

	// 2. Injured HP -> 1st Espresso costs $35, enters 40s cooldown, 2nd buy blocked during cooldown, next costs $65
	w.Player.HP = 2
	w.HandleShopInput(inHeal)
	if w.Player.Cash != 4965 || w.Player.HP != 4 || w.Player.EspressoBuys != 1 {
		t.Fatalf("expected 1st Espresso ($35 -> Cash=4965, HP=4), got Cash=%d HP=%d buys=%d",
			w.Player.Cash, w.Player.HP, w.Player.EspressoBuys)
	}
	if w.Shops[0].HealCooldown != EspressoCooldownSec {
		t.Fatalf("expected HealCooldown=%.1fs, got %.1f", EspressoCooldownSec, w.Shops[0].HealCooldown)
	}
	w.HandleShopInput(inHeal)
	if w.Player.Cash != 4965 {
		t.Fatalf("expected Espresso blocked while on cooldown")
	}
	w.updateShops(41.0, false)
	if w.Shops[0].HealCooldown != 0 || w.CurrentHealCost() != 65 {
		t.Fatalf("expected cooldown cleared and 2nd Espresso cost $65, got cd=%.1f cost=%d",
			w.Shops[0].HealCooldown, w.CurrentHealCost())
	}

	// 3. Relics hard-cap at 6/6 (no infinite MaxHP farm from Shop!)
	for i := 0; i < int(RelicCount); i++ {
		w.Player.Relics[i] = true
	}
	startMaxHP := w.Player.MaxHP
	cashBeforeRelic := w.Player.Cash
	inRelic := baseTestInput()
	inRelic.ShopAction = ShopActionRelic
	w.HandleShopInput(inRelic)
	if w.Player.Cash != cashBeforeRelic || w.Player.MaxHP != startMaxHP {
		t.Fatalf("expected Relic purchase hard-blocked at 6/6, cash=%d MaxHP=%d", w.Player.Cash, w.Player.MaxHP)
	}

	// 4. Upgrade Crate scales in price and goes on 35s cooldown
	inUpg := baseTestInput()
	inUpg.ShopAction = ShopActionUpgrade
	w.HandleShopInput(inUpg)
	if w.Player.Cash != cashBeforeRelic-140 || w.Shops[0].UpgradeCooldown != UpgradeCooldownSec || w.CurrentUpgradeCost() != 200 {
		t.Fatalf("expected 1st Upgrade $140, cd=%.1f, next $200; got cash=%d cd=%.1f next=$%d",
			UpgradeCooldownSec, w.Player.Cash, w.Shops[0].UpgradeCooldown, w.CurrentUpgradeCost())
	}
	w.PendingLevelUp = false
	w.HandleShopInput(inUpg)
	if w.PendingLevelUp || w.Player.Cash != cashBeforeRelic-140 {
		t.Fatalf("expected Upgrade Crate blocked while on 35s cooldown")
	}
}

func TestMultiStopChargesGraceDrainAndExpressChain(t *testing.T) {
	w := NewWorldWithSeed(902)
	w.Girls = nil

	// Upgrading MaxStops immediately grants a usable charge for Stop #2
	w.applyUpgrade(UpgradeCard{Kind: UpgradeMaxStops})
	if w.Player.EffectiveMaxStops() != 2 || w.Player.StopCharges != 2 {
		t.Fatalf("expected MaxStops=2 and StopCharges=2 after upgrade, got max=%d charges=%d",
			w.Player.EffectiveMaxStops(), w.Player.StopCharges)
	}

	// Throw Stop #1, step past 0.22s burst delay, then throw Stop #2 inside ExpressLinkMaxDist
	w.tryThrowBusStop(w.Player.X+60, w.Player.Y)
	if len(w.Stops) != 1 || w.Player.StopCharges != 1 {
		t.Fatalf("expected Stop #1 thrown and 1 charge left, stops=%d charges=%d", len(w.Stops), w.Player.StopCharges)
	}
	if w.Player.CanThrowStop(w.ActiveStopCount()) {
		t.Fatalf("expected 0.22s burst delay to guard immediate same-frame 2nd stop throw")
	}
	w.Player.Update(baseTestInput(), 0.25, false, false, 0)
	if !w.Player.CanThrowStop(w.ActiveStopCount()) {
		t.Fatalf("expected Stop #2 to be throwable right after 0.22s burst delay")
	}
	w.tryThrowBusStop(w.Player.X-60, w.Player.Y)
	if len(w.Stops) != 2 || w.Player.StopCharges != 0 {
		t.Fatalf("expected 2 simultaneous stops and 0 charges remaining, stops=%d charges=%d",
			len(w.Stops), w.Player.StopCharges)
	}

	// Grace drain: while inside Stop #1, Stop #2 drains at 0.25x instead of 1.25x
	w.Player.X = w.Stops[0].X
	w.Player.Y = w.Stops[0].Y
	stop2StartLife := w.Stops[1].LifeTimer
	w.updateStops(1.0, ActionNone, false)
	stop2Drop := stop2StartLife - w.Stops[1].LifeTimer
	if stop2Drop < 0.20 || stop2Drop > 0.30 {
		t.Fatalf("expected waiting Stop #2 to drain at ~0.25/s while player in Stop #1, dropped %.3f", stop2Drop)
	}

	// Express Chain: completing Stop #1 refreshes Stop #2 LifeTimer & boosts minigame +35%
	w.Stops[1].Minigame.Progress = 0.20
	w.Stops[1].LifeTimer = 5.0
	w.completeStopBoarding(&w.Stops[0])
	if w.Stops[1].LifeTimer != w.Stops[1].MaxLifeTimer {
		t.Fatalf("expected Express Chain to refresh Stop #2 timer to %.1f, got %.1f",
			w.Stops[1].MaxLifeTimer, w.Stops[1].LifeTimer)
	}
	if w.Stops[1].Minigame.Progress < 0.54 {
		t.Fatalf("expected Express Chain to boost Stop #2 progress >=0.55, got %.2f", w.Stops[1].Minigame.Progress)
	}
}

func TestBossArenaSpotlightAndPoliceThrottle(t *testing.T) {
	w := NewWorldWithSeed(903)
	// Fill arena with bullets and 12 unhooked girls
	w.Girls = nil
	for i := 0; i < 12; i++ {
		w.Girls = append(w.Girls, NewGirl(200+i, 200+float32(i*40), 220, ArchEmo, false, w.rng))
	}
	for i := 0; i < 20; i++ {
		w.Bullets.Spawn(Bullet{X: 300 + float32(i*10), Y: 300, Radius: 4, MaxAge: 5, Kind: BulletTear})
	}

	w.ForceSpawnBoss(BossLover)
	if w.Bullets.Count != 0 {
		t.Fatalf("expected ForceSpawnBoss to clear stray bullets, got %d", w.Bullets.Count)
	}
	if len(w.Girls) > 4 {
		t.Fatalf("expected ForceSpawnBoss to thin unhooked crowd to <=4, got %d", len(w.Girls))
	}
	if w.BossIntroTimer <= 2.0 {
		t.Fatalf("expected BossIntroTimer > 2.0s on spawn, got %.2f", w.BossIntroTimer)
	}

	// Verify all 6 bosses have tactical non-empty hints
	for k := BossKind(0); k < BossCount; k++ {
		if BossGimmickHint(k) == "" {
			t.Fatalf("expected non-empty BossGimmickHint for BossKind %d", k)
		}
	}

	// Verify PoliceSystem throttles officer cap to <= 2 during BossFightMode even at 5*
	w.Police.Stars = 5
	w.Police.BossFightMode = true
	w.Police.SpawnTimer = 0
	for step := 0; step < 30; step++ {
		w.Police.Update(0.5, 1.0, w.Player.X, w.Player.Y, w.Stops, &w.Bullets, &w.Telegraphs, w.rng)
	}
	if len(w.Police.Officers) > 2 {
		t.Fatalf("expected BossFightMode to cap police officers at 2, got %d", len(w.Police.Officers))
	}
}

func TestBusBayAndZTLUsabilityUpgrades(t *testing.T) {
	w := NewWorldWithSeed(904)
	w.Girls = nil

	// 1. Margin forgiveness: 12px outside BusBay box (still within 18px) detects InBusBay & clears bullets
	bay := w.BusBays[0] // X:160..248, Y:150..186
	tx, ty := bay.X-12, bay.Y+bay.H*0.5
	w.Player.X = tx - 30
	w.Player.Y = ty
	w.Bullets.Spawn(Bullet{X: tx + 50, Y: ty, Radius: 4, MaxAge: 5, Kind: BulletPolice})

	w.tryThrowBusStop(tx, ty)
	if len(w.Stops) != 1 || !w.Stops[0].InBusBay {
		t.Fatalf("expected forgiving 18px BusBay margin to trigger InBusBay at (%.1f,%.1f)", tx, ty)
	}
	if w.Bullets.Count != 0 {
		t.Fatalf("expected BusBay drop to clear bullets within 105px, remaining=%d", w.Bullets.Count)
	}
	if w.Player.StopCooldownTimer > w.Player.EffectiveStopCooldown()*0.76 {
		t.Fatalf("expected -25%% BusBay cooldown discount, got %.2f vs base %.2f",
			w.Player.StopCooldownTimer, w.Player.EffectiveStopCooldown())
	}

	// 2. Curbside pull draws free girls toward active BusBay stop
	w.Girls = append(w.Girls, NewGirl(950, w.Stops[0].X+160, w.Stops[0].Y, ArchEmo, false, w.rng))
	startDist := w.Girls[0].X - w.Stops[0].X
	w.updateStops(0.5, ActionNone, false)
	if w.Girls[0].X-w.Stops[0].X >= startDist-10 {
		t.Fatalf("expected BusBay curbside pull to draw nearby free girl toward shelter")
	}

	// 3. ZTL throw clears bullets with 110px EMP & grants +50% XP bonus on boarding
	w.Stops = nil
	w.Player.StopCooldownTimer = 0
	w.Player.MaxSimultaneousStops = 1
	ztl := w.ZTLZones[0]
	zx, zy := ztl.X+ztl.W*0.5, ztl.Y+ztl.H*0.5
	w.Player.X = zx
	w.Player.Y = zy + 40
	w.Bullets.Spawn(Bullet{X: zx + 70, Y: zy, Radius: 4, MaxAge: 5, Kind: BulletPolice})
	w.tryThrowBusStop(zx, zy)
	if len(w.Stops) != 1 || !w.Stops[0].InZTL || w.Bullets.Count != 0 {
		t.Fatalf("expected ZTL throw to clear nearby bullets with EMP wave")
	}

	w.Girls = []Girl{NewGirl(951, zx, zy, ArchEmo, false, w.rng)}
	w.Stops[0].HookedGirlIDs = []int{951}
	w.Girls[0].HookedByStop = w.Stops[0].ID
	w.Police.Stars = 0
	w.Player.XP = 0
	w.FirstBoardingDone = true // Isolate ZTL +50% XP bonus from first-60s onboarding +15 XP bonus
	w.completeStopBoarding(&w.Stops[0])
	// Base XP=26, ZTL +50% -> 39 XP
	if w.Player.XP != 39 {
		t.Fatalf("expected ZTL +50%% XP (39 XP), got %d", w.Player.XP)
	}
}
