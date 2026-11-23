package render

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"girlbusAG/assets"
	"girlbusAG/internal/game"
)

// HotbarSlotRect returns the virtual screen rectangle for archetype slot i (0..3).
func HotbarSlotRect(i int) game.Rect {
	const (
		cardW = float32(72)
		cardH = float32(38)
		gap   = float32(6)
	)
	totalCardsW := 4*cardW + 3*gap
	dockW := totalCardsW + 76
	dockX := (game.ScreenWidth - dockW) * 0.5
	dockY := float32(game.ScreenHeight) - cardH - 6
	return game.Rect{
		X: dockX + float32(i)*(cardW+gap),
		Y: dockY,
		W: cardW,
		H: cardH,
	}
}

// LevelUpCardRect returns the virtual screen rectangle for upgrade card i (0..2).
func LevelUpCardRect(i int) game.Rect {
	return game.Rect{
		X: float32(44 + i*188),
		Y: 92,
		W: 174,
		H: 174,
	}
}

// ShopChoiceRect returns the virtual screen rectangle for Vending Shop row i (0..3).
func ShopChoiceRect(i int) game.Rect {
	return game.Rect{
		X: 144,
		Y: 104 + float32(i)*36,
		W: 352,
		H: 30,
	}
}

// ShopCloseRect returns the close button rectangle for the Vending Shop modal.
func ShopCloseRect() game.Rect {
	return game.Rect{X: 256, Y: 274, W: 128, H: 24}
}

// SlotSpinRect returns the SPIN button rectangle for Rizz-Slots 777.
func SlotSpinRect() game.Rect {
	return game.Rect{X: 162, Y: 256, W: 196, H: 30}
}

// SlotCloseRect returns the EXIT button rectangle for Rizz-Slots 777.
func SlotCloseRect() game.Rect {
	return game.Rect{X: 372, Y: 256, W: 106, H: 30}
}

func (r *Renderer) drawHUD(screen *ebiten.Image, w *game.World) {
	// Top Arcade Bar
	vector.DrawFilledRect(screen, 0, 0, game.ScreenWidth, 26, colHudBg, false)
	vector.DrawFilledRect(screen, 0, 26, game.ScreenWidth, 1.5, PremulColor(colGyaru, 145), false)

	// HP Hearts
	for i := 0; i < w.Player.MaxHP; i++ {
		hx := float32(12 + i*13)
		c := color.RGBA{65, 32, 55, 210}
		filled := i < w.Player.HP
		if filled {
			c = colKawaii
		}
		drawHeartIcon(screen, hx, 12, c, filled)
	}

	// Chronos Gauge + [SH] Keycap (Left Shift / Q)
	chrX := float32(18 + w.Player.MaxHP*13)
	if chrX < 94 {
		chrX = 94
	}
	r.drawKeycap(screen, "SH", chrX, 4, 18, 17, color.RGBA{18, 70, 82, 240}, colGyaru, w.Player.ChronosActive)
	barX := chrX + 22
	chrFrac := clampF(w.Player.ChronosMeter/w.Player.ChronosMax, 0, 1)
	vector.DrawFilledRect(screen, barX, 7, 56, 10, color.RGBA{20, 30, 50, 220}, false)
	vector.DrawFilledRect(screen, barX, 7, 56*chrFrac, 10, colGyaru, false)
	vector.StrokeRect(screen, barX, 7, 56, 10, 1, colGyaru, false)

	// Wanted Stars + Multiplier Badge
	xpMult, cashMult := w.Police.XPAndCashMultiplier()
	starX := float32(barX + 66)
	for s := 0; s < 5; s++ {
		sc := color.RGBA{60, 52, 74, 200}
		if s < w.Police.Stars {
			sc = colGold
			if w.Police.Stars >= 4 && int(r.tick*10)%2 == 0 {
				sc = colRedAlert
			}
		}
		drawStarIcon(screen, starX+float32(s*12), 12, 4.5, sc)
	}
	multX := int(starX) + 64
	if w.Police.Stars > 0 {
		r.drawScaledText(screen, fmt.Sprintf("%.1fx/%.1fx$", xpMult, cashMult), multX, 4, 1.0, colGold, colOutline)
	} else {
		r.drawScaledText(screen, "PULITO", multX, 4, 1.0, color.RGBA{160, 175, 195, 220}, colOutline)
	}

	// Cash, Level/XP, Timer, Boarded
	r.drawScaledText(screen, fmt.Sprintf("$%d", w.Player.Cash), 370, 4, 1.0, colGreenNeon, colOutline)

	xpFrac := clampF(float32(w.Player.XP)/float32(w.Player.NextLevelXP), 0, 1)
	vector.DrawFilledRect(screen, 424, 17, 78, 5, color.RGBA{30, 30, 50, 220}, false)
	vector.DrawFilledRect(screen, 424, 17, 78*xpFrac, 5, colStudent, false)
	r.drawTextColored(screen, fmt.Sprintf("LV%d", w.Player.Level), 424, 1, color.White)

	mins := int(w.RunTime) / 60
	secs := int(w.RunTime) % 60
	r.drawScaledText(screen, fmt.Sprintf("T%d %02d:%02d", w.DifficultyTier, mins, secs), 512, 4, 1.0, colGold, colOutline)
	r.drawScaledText(screen, fmt.Sprintf("R:%d", w.GirlsBoarded), 594, 4, 1.0, colKawaii, colOutline)

	// Owned Relics strip (top-left under HUD bar)
	rx := float32(8)
	relicCodes := [game.RelicCount]string{"NC", "AB", "VB", "CF", "DE", "VP"}
	for rk := 0; rk < int(game.RelicCount); rk++ {
		if !w.Player.Relics[rk] {
			continue
		}
		drawPillBadge(screen, rx, 30, 22, 13, color.RGBA{34, 20, 56, 220}, colGold)
		r.drawTextColored(screen, relicCodes[rk], int(rx)+5, 29, colGold)
		rx += 25
	}

	// Boss HP/Shield Banner
	if w.Boss.Active {
		bannerTxt := fmt.Sprintf("BOSS: %s [%d/%d]", w.Boss.Name, w.Boss.Shields, w.Boss.MaxShields)
		if w.BossRushMode {
			bannerTxt = fmt.Sprintf("RUSH %d/6: %s [%d/%d]", w.BossesDefeated+1, w.Boss.Name, w.Boss.Shields, w.Boss.MaxShields)
		}
		vector.DrawFilledRect(screen, 190, 30, 260, 22, colHudBg, false)
		vector.StrokeRect(screen, 190, 30, 260, 22, 1.5, colKawaii, false)
		r.drawScaledText(screen, bannerTxt, 200, 34, 1.0, colKawaii, colOutline)
	} else if w.BossRushMode && w.BossScheduleIdx < len(w.BossSchedule) {
		drawPillBadge(screen, 220, 30, 200, 16, colHudBg, colGold)
		r.drawTextColored(screen, fmt.Sprintf("BOSS RUSH %d/6 TRA %.0fs", w.BossesDefeated+1, w.NextBossRushTimer), 232, 31, colGold)
	}

	// Optional FPS badge
	if r.showFPS {
		fpsTxt := fmt.Sprintf("FPS %.0f", ebiten.ActualFPS())
		drawPillBadge(screen, game.ScreenWidth-62, 30, 56, 15, colHudBg, colGyaru)
		r.drawTextColored(screen, fpsTxt, game.ScreenWidth-56, 30, colGyaru)
	}

	// Rizz Fever Streak badge (right side below top bar / FPS)
	if w.StreakCount >= 2 {
		sy := float32(30)
		if r.showFPS {
			sy = 48
		}
		bw := float32(96)
		bx := float32(game.ScreenWidth) - bw - 8
		txtCol := colGyaru
		lbl := fmt.Sprintf("STREAK x%d", w.StreakCount)
		bgCol := colHudBg
		if w.StreakCount >= 5 {
			lbl = fmt.Sprintf("RUSH x%d!", w.StreakCount)
			txtCol = colGold
			if int(r.tick*8)%2 == 0 {
				txtCol = colKawaii
			}
			bgCol = Premul(48, 16, 52, 235)
		}
		drawPillBadge(screen, bx, sy, bw, 19, bgCol, txtCol)
		r.drawTextColored(screen, lbl, int(bx+8), int(sy+1), txtCol)
		frac := clampF(w.StreakTimer/game.StreakWindowDuration, 0, 1)
		vector.DrawFilledRect(screen, bx+6, sy+15, (bw-12)*frac, 2.5, txtCol, false)
	}

	// Challenge unlocked live toast banner
	if r.toastTimer > 0 && r.activeChallengeToast != "" {
		msg := "SFIDA SBLOCCATA: " + r.activeChallengeToast
		mw := float32(len(msg)*6 + 22)
		mx := (game.ScreenWidth - mw) * 0.5
		my := float32(56)
		drawPillBadge(screen, mx, my, mw, 20, color.RGBA{38, 24, 6, 240}, colGold)
		drawStarIcon(screen, mx+10, my+10, 4.5, colGold)
		r.drawScaledText(screen, msg, int(mx+19), int(my+3), 1.0, colGold, colOutline)
	}

	// Bottom-Center 4-Slot Archetype Hotbar Dock
	r.drawBottomHotbarDock(screen, w)
}

func (r *Renderer) drawOffscreenIndicators(screen *ebiten.Image, w *game.World, camX, camY float32) {
	// Boss offscreen pointer
	if w.Boss.Active {
		r.drawEdgeMarker(screen, w.Boss.X, w.Boss.Y, camX, camY, "BOSS", colKawaii)
	}
	// Active BusStops offscreen pointers
	for i := range w.Stops {
		st := &w.Stops[i]
		if st.Active && !st.BusRushTriggered {
			r.drawEdgeMarker(screen, st.X, st.Y, camX, camY, "FERMATA", archetypeColor(st.Arch))
		}
	}
	// Nearest VendingShop offscreen marker when player has cash >= $50
	if w.Player.Cash >= 50 && len(w.Shops) > 0 {
		bestIdx := -1
		bestD2 := float32(1e12)
		for i, sh := range w.Shops {
			d2 := (sh.X-w.Player.X)*(sh.X-w.Player.X) + (sh.Y-w.Player.Y)*(sh.Y-w.Player.Y)
			if d2 < bestD2 {
				bestD2 = d2
				bestIdx = i
			}
		}
		if bestIdx >= 0 {
			r.drawEdgeMarker(screen, w.Shops[bestIdx].X, w.Shops[bestIdx].Y, camX, camY, "$ SHOP", colGyaru)
		}
	}
	// Nearest SlotMachine offscreen marker when player has cash >= $35
	if len(w.SlotMachines) > 0 && w.Player.Cash >= w.SlotMachines[0].SpinCost {
		bestIdx := -1
		bestD2 := float32(1e12)
		for i, sm := range w.SlotMachines {
			d2 := (sm.X-w.Player.X)*(sm.X-w.Player.X) + (sm.Y-w.Player.Y)*(sm.Y-w.Player.Y)
			if d2 < bestD2 {
				bestD2 = d2
				bestIdx = i
			}
		}
		if bestIdx >= 0 {
			r.drawEdgeMarker(screen, w.SlotMachines[bestIdx].X, w.SlotMachines[bestIdx].Y, camX, camY, "777", colGold)
		}
	}
}

func (r *Renderer) drawEdgeMarker(screen *ebiten.Image, wx, wy, camX, camY float32, label string, col color.RGBA) {
	sx := wx - camX
	sy := wy - camY
	if sx >= 8 && sx <= game.ScreenWidth-8 && sy >= 28 && sy <= game.ScreenHeight-48 {
		return
	}

	cx := float32(game.ScreenWidth * 0.5)
	cy := float32(game.ScreenHeight * 0.5)
	dx := sx - cx
	dy := sy - cy
	if dx == 0 && dy == 0 {
		return
	}

	const marginX = float32(28)
	const minY = float32(40)
	const maxY = float32(game.ScreenHeight - 60)

	// Clamp ray onto screen safe border box
	tX := float32(1e9)
	if dx > 0 {
		tX = (game.ScreenWidth - marginX - cx) / dx
	} else if dx < 0 {
		tX = (marginX - cx) / dx
	}
	tY := float32(1e9)
	if dy > 0 {
		tY = (maxY - cy) / dy
	} else if dy < 0 {
		tY = (minY - cy) / dy
	}
	t := tX
	if tY < t {
		t = tY
	}
	if t > 1 {
		t = 1
	}

	mx := cx + dx*t
	my := cy + dy*t
	ang := math.Atan2(float64(dy), float64(dx))

	tipX := mx + float32(math.Cos(ang))*9
	tipY := my + float32(math.Sin(ang))*9
	leftX := mx + float32(math.Cos(ang+2.4))*6
	leftY := my + float32(math.Sin(ang+2.4))*6
	rightX := mx + float32(math.Cos(ang-2.4))*6
	rightY := my + float32(math.Sin(ang-2.4))*6

	vector.StrokeLine(screen, leftX, leftY, tipX, tipY, 2.2, col, true)
	vector.StrokeLine(screen, rightX, rightY, tipX, tipY, 2.2, col, true)

	bw := float32(len(label)*6 + 8)
	bx := clampF(mx-bw*0.5, 4, game.ScreenWidth-bw-4)
	by := clampF(my-7, 30, game.ScreenHeight-64)
	drawPillBadge(screen, bx, by, bw, 13, colHudBg, col)
	r.drawTextColored(screen, label, int(bx+4), int(by-1), col)
}

func (r *Renderer) drawBottomHotbarDock(screen *ebiten.Image, w *game.World) {
	r0 := HotbarSlotRect(0)
	dockX := r0.X
	dockY := r0.Y
	cardW := r0.W
	cardH := r0.H
	const gap = float32(6)
	totalCardsW := 4*cardW + 3*gap
	dockW := totalCardsW + 76

	vector.DrawFilledRect(screen, dockX-4, dockY-4, dockW+8, cardH+8, colHudBg, false)
	vector.StrokeRect(screen, dockX-4, dockY-4, dockW+8, cardH+8, 1.5, PremulColor(colGyaru, 120), false)

	type slotInfo struct {
		key     string
		name    string
		stopSpr string
		girlSpr string
		arch    game.Archetype
	}
	slots := [4]slotInfo{
		{"1", "EMO", assets.SpriteStopEmo, assets.SpriteGirlEmo0, game.ArchEmo},
		{"2", "UNI", assets.SpriteStopStudent, assets.SpriteGirlStudent0, game.ArchStudent},
		{"3", "KAW", assets.SpriteStopKawaii, assets.SpriteGirlKawaii0, game.ArchKawaii},
		{"4", "GYA", assets.SpriteStopGyaru, assets.SpriteGirlGyaru0, game.ArchGyaru},
	}

	effCD := w.Player.EffectiveStopCooldown()
	cdFrac := float32(0)
	if effCD > 0 && w.Player.StopCooldownTimer > 0 {
		cdFrac = clampF(w.Player.StopCooldownTimer/effCD, 0, 1)
	}

	op := &ebiten.DrawImageOptions{}
	for i, s := range slots {
		rect := HotbarSlotRect(i)
		cx, cy := rect.X, rect.Y
		ac := archetypeColor(s.arch)
		selected := w.Player.SelectedStopArch == s.arch
		hovered := rect.Contains(r.mouseScreenX, r.mouseScreenY)

		cardBg := color.RGBA{26, 18, 42, 235}
		borderCol := PremulColor(ac, 140)
		if hovered && !selected {
			cardBg = color.RGBA{38, 26, 58, 242}
			borderCol = color.RGBA{255, 255, 255, 255}
		}
		if selected {
			cy -= 3
			cardBg = color.RGBA{46, 30, 72, 248}
			borderCol = colGold
			vector.StrokeRect(screen, cx-2, cy-2, cardW+4, cardH+4, 2.0, ac, false)
		}

		vector.DrawFilledRect(screen, cx, cy, cardW, cardH, cardBg, false)
		vector.StrokeRect(screen, cx, cy, cardW, cardH, 1.4, borderCol, false)

		r.drawKeycap(screen, s.key, cx+3, cy+3, 14, 15, color.RGBA{42, 34, 62, 240}, ac, selected)
		r.drawTextColored(screen, s.name, int(cx+20), int(cy+2), ac)

		stopImg := r.sprites.Get(s.stopSpr)
		girlImg := r.sprites.Get(s.girlSpr)

		op.GeoM.Reset()
		op.GeoM.Scale(20.0/float64(stopImg.Bounds().Dx()), 20.0/float64(stopImg.Bounds().Dy()))
		op.GeoM.Translate(float64(cx+6), float64(cy+17))
		op.ColorScale.Reset()
		screen.DrawImage(stopImg, op)

		op.GeoM.Reset()
		op.GeoM.Scale(18.0/float64(girlImg.Bounds().Dx()), 18.0/float64(girlImg.Bounds().Dy()))
		op.GeoM.Translate(float64(cx+44), float64(cy+18))
		screen.DrawImage(girlImg, op)

		r.drawTextColored(screen, "=", int(cx+32), int(cy+19), Premul(230, 230, 245, 220))

		if cdFrac > 0 {
			ovH := cardH * cdFrac
			vector.DrawFilledRect(screen, cx+1, cy+cardH-ovH, cardW-2, ovH, Premul(10, 6, 20, 145), false)
		}
	}

	statusX := dockX + totalCardsW + 6
	r.drawKeycap(screen, "LMB", statusX, dockY+2, 26, 16, color.RGBA{65, 40, 95, 240}, colGold, false)
	r.drawTextColored(screen, "LANCIA", int(statusX+30), int(dockY+2), color.White)

	slotStr := fmt.Sprintf("%d/%d", w.ActiveStopCount(), w.Player.EffectiveMaxStops())
	r.drawScaledText(screen, slotStr, int(statusX+4), int(dockY+20), 1.0, colGyaru, colOutline)

	barX := statusX + 30
	barW := float32(36)
	readyFrac := 1.0 - cdFrac
	barCol := colGreenNeon
	if cdFrac > 0 {
		barCol = colGold
	}
	vector.DrawFilledRect(screen, barX, dockY+24, barW, 8, color.RGBA{26, 20, 40, 220}, false)
	vector.DrawFilledRect(screen, barX, dockY+24, barW*readyFrac, 8, barCol, false)
	vector.StrokeRect(screen, barX, dockY+24, barW, 8, 1, barCol, false)

	if w.Player.GroundedTimer > 0 {
		r.drawScaledText(screen, fmt.Sprintf("PUNIZIONE %.0fs!", w.Player.GroundedTimer), int(dockX+totalCardsW*0.5-46), int(dockY-20), 1.0, colRedAlert, colOutline)
	}
}

func (r *Renderer) drawLevelUpModal(screen *ebiten.Image, w *game.World) {
	vector.DrawFilledRect(screen, 0, 0, game.ScreenWidth, game.ScreenHeight, color.RGBA{12, 8, 22, 195}, false)
	r.drawScaledText(screen, fmt.Sprintf("LIVELLO %d! SCEGLI POTENZIAMENTO", w.Player.Level), 166, 48, 1.4, colGold, colOutline)
	r.drawTextColored(screen, "CLICCA UNA CARTA O PREMI [1/2/3]", 222, 72, color.RGBA{200, 220, 245, 220})

	keys := [3][2]string{{"1", "Z"}, {"2", "X"}, {"3", "C"}}
	for i := 0; i < 3; i++ {
		card := w.UpgradeChoices[i]
		rect := LevelUpCardRect(i)
		hovered := rect.Contains(r.mouseScreenX, r.mouseScreenY)
		bg := colHudBg
		border := colGyaru
		if hovered {
			bg = color.RGBA{38, 24, 62, 245}
			border = colGold
		}

		vector.DrawFilledRect(screen, rect.X, rect.Y, rect.W, rect.H, bg, false)
		vector.StrokeRect(screen, rect.X, rect.Y, rect.W, rect.H, 2, border, false)
		r.drawKeycap(screen, keys[i][0], rect.X+58, rect.Y+12, 22, 20, color.RGBA{70, 50, 100, 240}, colGold, hovered)
		r.drawKeycap(screen, keys[i][1], rect.X+92, rect.Y+12, 22, 20, color.RGBA{70, 50, 100, 240}, colGold, hovered)
		r.drawScaledText(screen, card.Title, int(rect.X)+10, int(rect.Y)+46, 1.0, colGyaru, colOutline)

		drawWrappedText(r, screen, card.Desc, int(rect.X)+10, int(rect.Y)+82, 24, color.White)
	}
}

func (r *Renderer) drawShopModal(screen *ebiten.Image, w *game.World) {
	vector.DrawFilledRect(screen, 0, 0, game.ScreenWidth, game.ScreenHeight, color.RGBA{10, 6, 18, 190}, false)
	const (
		boxX = float32(128)
		boxY = float32(52)
		boxW = float32(384)
		boxH = float32(258)
	)
	vector.DrawFilledRect(screen, boxX, boxY, boxW, boxH, colHudBg, false)
	vector.StrokeRect(screen, boxX, boxY, boxW, boxH, 2.0, colGyaru, false)

	r.drawScaledText(screen, "DISTRIBUTORE NEON // SHOP", int(boxX+16), int(boxY+12), 1.25, colGyaru, colOutline)
	r.drawScaledText(screen, fmt.Sprintf("$%d", w.Player.Cash), int(boxX+boxW-66), int(boxY+12), 1.25, colGreenNeon, colOutline)
	r.drawTextColored(screen, w.ShopMenu.Message, int(boxX+16), int(boxY+34), colGold)

	bribeCost := w.CurrentBribeCost()
	type shopOption struct {
		key   string
		title string
		cost  int
		col   color.RGBA
	}
	opts := [4]shopOption{
		{"1", "Caffe Espresso (+2 HP & Chronos Full)", 50, colKawaii},
		{"2", "Tangente Polizia (-2* Wanted)", bribeCost, colGold},
		{"3", "Reliquia Casuale (Bonus Passivo)", 120, colGyaru},
		{"4", "Cassa Upgrade Rapido (Scelta 3 Carte)", 140, colGreenNeon},
	}

	for i, opt := range opts {
		rect := ShopChoiceRect(i)
		hovered := rect.Contains(r.mouseScreenX, r.mouseScreenY)
		affordable := w.Player.Cash >= opt.cost
		if i == 1 && w.Police.Stars == 0 {
			affordable = false
		}

		bg := color.RGBA{24, 16, 40, 235}
		border := opt.col
		if hovered && affordable {
			bg = color.RGBA{46, 30, 74, 245}
			border = color.RGBA{255, 255, 255, 255}
		}
		if !affordable {
			border = color.RGBA{85, 80, 100, 200}
		}

		vector.DrawFilledRect(screen, rect.X, rect.Y, rect.W, rect.H, bg, false)
		vector.StrokeRect(screen, rect.X, rect.Y, rect.W, rect.H, 1.5, border, false)
		r.drawKeycap(screen, opt.key, rect.X+6, rect.Y+5, 18, 19, color.RGBA{48, 38, 72, 240}, border, hovered)
		txtCol := color.RGBA{255, 255, 255, 255}
		if !affordable {
			txtCol = color.RGBA{130, 130, 150, 220}
		}
		r.drawTextColored(screen, opt.title, int(rect.X+30), int(rect.Y+7), txtCol)
		costCol := colGreenNeon
		if w.Player.Cash < opt.cost {
			costCol = colRedAlert
		}
		r.drawScaledText(screen, fmt.Sprintf("$%d", opt.cost), int(rect.X+rect.W-42), int(rect.Y+7), 1.0, costCol, colOutline)
	}

	closeRect := ShopCloseRect()
	closeHover := closeRect.Contains(r.mouseScreenX, r.mouseScreenY)
	cBg := color.RGBA{56, 22, 36, 235}
	if closeHover {
		cBg = color.RGBA{92, 34, 54, 245}
	}
	vector.DrawFilledRect(screen, closeRect.X, closeRect.Y, closeRect.W, closeRect.H, cBg, false)
	vector.StrokeRect(screen, closeRect.X, closeRect.Y, closeRect.W, closeRect.H, 1.5, colKawaii, false)
	r.drawScaledText(screen, "[ESC] CHIUDI", int(closeRect.X+22), int(closeRect.Y+4), 1.0, color.White, colOutline)
}

func (r *Renderer) drawSlotModal(screen *ebiten.Image, w *game.World) {
	vector.DrawFilledRect(screen, 0, 0, game.ScreenWidth, game.ScreenHeight, color.RGBA{10, 6, 18, 196}, false)
	const (
		boxX = float32(140)
		boxY = float32(50)
		boxW = float32(360)
		boxH = float32(256)
	)
	vector.DrawFilledRect(screen, boxX, boxY, boxW, boxH, colHudBg, false)
	vector.StrokeRect(screen, boxX, boxY, boxW, boxH, 2.0, colGold, false)

	r.drawScaledText(screen, "RIZZ-SLOTS 777", int(boxX+18), int(boxY+12), 1.35, colGold, colOutline)
	r.drawScaledText(screen, fmt.Sprintf("$%d", w.Player.Cash), int(boxX+boxW-66), int(boxY+12), 1.25, colGreenNeon, colOutline)

	// 3 Animated Reels
	reelColors := [game.SymCount]color.RGBA{
		colKawaii,
		colGreenNeon,
		colStudent,
		color.RGBA{255, 110, 150, 255},
		colGold,
		colRedAlert,
	}
	for i := 0; i < 3; i++ {
		rx := boxX + 34 + float32(i)*102
		ry := boxY + 52
		rw := float32(88)
		rh := float32(74)
		sym := w.SlotModal.Reels[i]
		sc := reelColors[sym%game.SymCount]

		vector.DrawFilledRect(screen, rx, ry, rw, rh, color.RGBA{24, 16, 40, 245}, false)
		vector.StrokeRect(screen, rx, ry, rw, rh, 2.0, sc, false)
		lbl := game.SlotSymbolName(sym)
		tw := textPixelWidth(lbl, 1.35)
		r.drawScaledText(screen, lbl, int(rx+(rw-float32(tw))*0.5), int(ry+26), 1.35, sc, colOutline)
	}

	// Result message box
	resCol := colGyaru
	switch w.SlotModal.ResultKind {
	case 1:
		resCol = colGreenNeon
	case 2:
		resCol = colGold
	case 3:
		resCol = colRedAlert
	}
	vector.DrawFilledRect(screen, boxX+22, boxY+142, boxW-44, 48, color.RGBA{18, 12, 32, 225}, false)
	vector.StrokeRect(screen, boxX+22, boxY+142, boxW-44, 48, 1.2, resCol, false)
	r.drawScaledText(screen, w.SlotModal.ResultTitle, int(boxX+32), int(boxY+148), 1.05, resCol, colOutline)
	r.drawTextColored(screen, w.SlotModal.ResultDesc, int(boxX+32), int(boxY+168), color.White)

	spinCost := w.CurrentSlotCost()
	spinRect := SlotSpinRect()
	spinHover := spinRect.Contains(r.mouseScreenX, r.mouseScreenY)
	canSpin := !w.SlotModal.Spinning && w.Player.Cash >= spinCost
	sBg := color.RGBA{34, 88, 52, 240}
	sBorder := colGreenNeon
	if !canSpin {
		sBg = color.RGBA{40, 36, 54, 220}
		sBorder = color.RGBA{100, 95, 115, 220}
	} else if spinHover {
		sBg = color.RGBA{48, 120, 70, 245}
		sBorder = colGold
	}
	vector.DrawFilledRect(screen, spinRect.X, spinRect.Y, spinRect.W, spinRect.H, sBg, false)
	vector.StrokeRect(screen, spinRect.X, spinRect.Y, spinRect.W, spinRect.H, 1.8, sBorder, false)
	spinTxt := fmt.Sprintf("[SPAZIO] GIRA ($%d)", spinCost)
	if w.SlotModal.Spinning {
		spinTxt = "GIRANDO..."
	}
	r.drawScaledText(screen, spinTxt, int(spinRect.X+22), int(spinRect.Y+7), 1.0, color.White, colOutline)

	closeRect := SlotCloseRect()
	closeHover := closeRect.Contains(r.mouseScreenX, r.mouseScreenY)
	cBg := color.RGBA{56, 22, 36, 235}
	if closeHover {
		cBg = color.RGBA{92, 34, 54, 245}
	}
	vector.DrawFilledRect(screen, closeRect.X, closeRect.Y, closeRect.W, closeRect.H, cBg, false)
	vector.StrokeRect(screen, closeRect.X, closeRect.Y, closeRect.W, closeRect.H, 1.5, colKawaii, false)
	r.drawScaledText(screen, "[ESC] ESCI", int(closeRect.X+18), int(closeRect.Y+7), 1.0, color.White, colOutline)
}

func drawWrappedText(r *Renderer, screen *ebiten.Image, desc string, x, startY, cut int, clr color.Color) {
	lineY := startY
	for len(desc) > 0 {
		if len(desc) <= cut {
			r.drawTextColored(screen, desc, x, lineY, clr)
			break
		}
		sp := cut
		for sp > 8 && desc[sp] != ' ' {
			sp--
		}
		if desc[sp] != ' ' {
			sp = cut
		}
		r.drawTextColored(screen, desc[:sp], x, lineY, clr)
		desc = desc[sp:]
		if len(desc) > 0 && desc[0] == ' ' {
			desc = desc[1:]
		}
		lineY += 16
	}
}
