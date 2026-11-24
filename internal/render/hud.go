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
		H: 176,
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

// TutorialGuideRect returns the compact [GUIDA] button rectangle in the tutorial sandbox bar.
func TutorialGuideRect() game.Rect {
	return game.Rect{X: 500, Y: 33, W: 60, H: 26}
}

// TutorialExitRect returns the compact [MENU] button rectangle in the tutorial sandbox bar.
func TutorialExitRect() game.Rect {
	return game.Rect{X: 566, Y: 33, W: 60, H: 26}
}

// drawHUD delegates to DrawHUDAndOverlays so both legacy and exported entry points stay in sync.
func (r *Renderer) drawHUD(screen *ebiten.Image, w *game.World) {
	r.DrawHUDAndOverlays(screen, w)
}

// DrawHUDAndOverlays renders the clean top HUD bar, pixel-art relic tray, segmented boss shield bar,
// streak/rush badge, tutorial sandbox bar, bottom mouse-friendly hotbar, challenge toast, and active modals.
func (r *Renderer) DrawHUDAndOverlays(screen *ebiten.Image, w *game.World) {
	// 1. Sleek Top Dark-Glass HUD Bar (640x24px, ZERO HP bar — HP lives diegetically beneath Alfredo)
	r.drawTopHUDBar(screen, w)

	// 2. Tutorial Sandbox Bar (when in interactive tutorial)
	subBarY := float32(29)
	if w.TutorialMode {
		r.drawTutorialSandboxBar(screen, w)
		subBarY = 70
	}

	// 3. Intuitive Pixel-Art Relic Tray (top-left below HUD bar)
	r.drawRelicTray(screen, w, subBarY)

	// 4. Centered Boss 3-Shield Bar / Boss Rush Countdown / 5* Chopper Alert
	if !w.TutorialMode {
		r.drawCenterStatusBanner(screen, w)
	}

	// 5. Streak / Rush Hour Badge (top-right below HUD bar)
	r.drawStreakBadge(screen, w, subBarY)

	// 6. Challenge Unlocked Toast Banner
	r.drawChallengeToast(screen, w)

	// 7. Clean Bottom Bus-Stop Hotbar Dock
	r.drawBottomHotbarDock(screen, w)

	// 8. Active Modals (Level-Up / Vending Shop / Rizz-Slots 777)
	if w.PendingLevelUp {
		r.drawLevelUpModal(screen, w)
	} else if w.ShopMenu.Active {
		r.drawShopModal(screen, w)
	} else if w.SlotModal.Active {
		r.drawSlotModal(screen, w)
	}
}

func (r *Renderer) drawTopHUDBar(screen *ebiten.Image, w *game.World) {
	// Sleek translucent dark-glass backdrop + subtle top sheen + thin neon bottom trim
	vector.DrawFilledRect(screen, 0, 0, game.ScreenWidth, 24, Premul(14, 9, 26, 224), false)
	vector.DrawFilledRect(screen, 0, 0, game.ScreenWidth, 1, Premul(255, 255, 255, 18), false)
	vector.DrawFilledRect(screen, 0, 24, game.ScreenWidth, 1.5, PremulColor(colGyaru, 165), false)

	// --- LEFT (x=10): Bolt Icon + Sleek Chronos Bar (74x9px) + Tiny SH cue ---
	r.DrawSpriteIcon(screen, assets.SpriteIconBolt, 10, 6, 1.0)

	const (
		chrBarX = float32(25)
		chrBarY = float32(7)
		chrBarW = float32(74)
		chrBarH = float32(9)
	)
	chrFrac := float32(0)
	if w.Player.ChronosMax > 0 {
		chrFrac = clampF(w.Player.ChronosMeter/w.Player.ChronosMax, 0, 1)
	}
	vector.DrawFilledRect(screen, chrBarX, chrBarY, chrBarW, chrBarH, Premul(16, 22, 42, 235), false)
	fillW := chrBarW * chrFrac
	if fillW > 0 {
		fillCol := colGyaru
		if w.Player.ChronosActive {
			pulse := uint8(195 + 60*math.Sin(float64(r.tick*14.0)))
			fillCol = color.RGBA{pulse, 255, 245, 255}
		}
		vector.DrawFilledRect(screen, chrBarX, chrBarY, fillW, chrBarH, fillCol, false)
		vector.DrawFilledRect(screen, chrBarX, chrBarY+1, fillW, 2, Premul(255, 255, 255, 95), false)
	}
	chrBorder := PremulColor(colGyaru, 185)
	if w.Player.ChronosActive {
		chrBorder = colGold
	}
	vector.StrokeRect(screen, chrBarX, chrBarY, chrBarW, chrBarH, 1.0, chrBorder, false)

	// Tiny SH cue pill
	shBg := Premul(20, 42, 56, 220)
	shBorder := PremulColor(colGyaru, 150)
	shTxt := color.RGBA{185, 245, 245, 255}
	if w.Player.ChronosActive {
		shBg = Premul(18, 92, 102, 245)
		shBorder = colGold
		shTxt = colGold
	}
	drawPillBadge(screen, 102, 5, 16, 13, shBg, shBorder)
	r.drawTextColored(screen, "SH", 104, 4, shTxt)

	// --- CENTER-LEFT (x=122): 5 Wanted Stars + Compact Multiplier Pill when Stars > 0 ---
	const starStartX = float32(122)
	const starStep = float32(13)
	for s := 0; s < 5; s++ {
		sx := starStartX + float32(s)*starStep
		if s < w.Police.Stars {
			if w.Police.Stars >= 4 && int(r.tick*10)%2 == 0 {
				r.drawSpriteIconTinted(screen, assets.SpriteIconStarWanted, sx, 6, 1.0, colRedAlert)
			} else {
				r.DrawSpriteIcon(screen, assets.SpriteIconStarWanted, sx, 6, 1.0)
			}
		} else {
			r.drawSpriteIconTinted(screen, assets.SpriteIconStarWanted, sx, 6, 1.0, color.RGBA{46, 38, 66, 135})
		}
	}
	if w.Police.Stars > 0 {
		xpMult, _ := w.Police.XPAndCashMultiplier()
		multTxt := fmt.Sprintf("%.1fx", xpMult)
		mw := float32(textPixelWidth(multTxt, 1.0) + 10)
		mx := starStartX + 5*starStep + 3
		multBorder := colGold
		multBg := Premul(58, 40, 8, 232)
		if w.Police.Stars >= 4 && int(r.tick*8)%2 == 0 {
			multBorder = colRedAlert
			multBg = Premul(68, 14, 26, 235)
		}
		drawPillBadge(screen, mx, 5, mw, 14, multBg, multBorder)
		r.drawTextColored(screen, multTxt, int(mx+5), 4, colGold)
	}

	// --- CENTER (x=276): Clean Centered Timer + Difficulty Tier Pill (01:24 • T1) ---
	const (
		centerPillX = float32(276)
		centerPillY = float32(4)
		centerPillW = float32(88)
		centerPillH = float32(16)
	)
	drawPillBadge(screen, centerPillX, centerPillY, centerPillW, centerPillH, Premul(22, 15, 38, 238), PremulColor(colGyaru, 155))
	mins := int(w.RunTime) / 60
	secs := int(w.RunTime) % 60
	timeStr := fmt.Sprintf("%02d:%02d", mins, secs)
	tierStr := fmt.Sprintf("T%d", w.DifficultyTier)
	r.drawScaledText(screen, timeStr, int(centerPillX+10), 4, 1.0, color.White, colOutline)
	vector.DrawFilledCircle(screen, centerPillX+48, 12, 1.7, colGyaru, true)
	r.drawScaledText(screen, tierStr, int(centerPillX+56), 4, 1.0, colGold, colOutline)

	// --- RIGHT (x=426): Coin Icon + Gold Cash, followed at x=500 by LV + Violet-Cyan XP Bar ---
	r.DrawSpriteIcon(screen, assets.SpriteIconCoin, 426, 6, 1.0)
	r.drawScaledText(screen, fmt.Sprintf("$%d", w.Player.Cash), 441, 4, 1.0, colGold, colOutline)

	lvStr := fmt.Sprintf("LV %d", w.Player.Level)
	r.drawScaledText(screen, lvStr, 500, 4, 1.0, color.RGBA{235, 228, 255, 255}, colOutline)

	const (
		xpBarX = float32(534)
		xpBarY = float32(9)
		xpBarW = float32(76)
		xpBarH = float32(7)
	)
	xpFrac := float32(0)
	if w.Player.NextLevelXP > 0 {
		xpFrac = clampF(float32(w.Player.XP)/float32(w.Player.NextLevelXP), 0, 1)
	}
	vector.DrawFilledRect(screen, xpBarX, xpBarY, xpBarW, xpBarH, Premul(24, 16, 44, 238), false)
	xpFillW := xpBarW * xpFrac
	if xpFillW > 0 {
		// Violet base with electric cyan leading sheen
		vector.DrawFilledRect(screen, xpBarX, xpBarY, xpFillW, xpBarH, colEmo, false)
		tipW := xpFillW * 0.35
		if tipW > 14 {
			tipW = 14
		}
		vector.DrawFilledRect(screen, xpBarX+xpFillW-tipW, xpBarY, tipW, xpBarH, colGyaru, false)
		vector.DrawFilledRect(screen, xpBarX, xpBarY+1, xpFillW, 1.5, Premul(255, 255, 255, 95), false)
	}
	vector.StrokeRect(screen, xpBarX, xpBarY, xpBarW, xpBarH, 1.0, Premul(165, 130, 235, 210), false)

	// Optional tiny FPS badge at far right corner only when enabled
	if r.showFPS {
		fpsTxt := fmt.Sprintf("%.0f", ebiten.ActualFPS())
		drawPillBadge(screen, 613, 5, 24, 13, Premul(14, 18, 32, 230), PremulColor(colGyaru, 150))
		tw := textPixelWidth(fpsTxt, 1.0)
		r.drawTextColored(screen, fpsTxt, 613+(24-tw)/2, 4, colGyaru)
	}
}

func relicIconKey(rk game.RelicKind) string {
	switch rk {
	case game.RelicNoiseCancel:
		return assets.SpriteIconRelicHeadphones
	case game.RelicMonthlyPass:
		return assets.SpriteIconRelicPass
	case game.RelicArmoredGlass:
		return assets.SpriteIconRelicShield
	case game.RelicLuckySocks:
		return assets.SpriteIconRelicClover
	case game.RelicDoubleEspresso:
		return assets.SpriteIconRelicEspresso
	case game.RelicFakeVIPPass:
		return assets.SpriteIconRelicVIP
	default:
		return assets.SpriteIconRelicVIP
	}
}

func (r *Renderer) drawRelicTray(screen *ebiten.Image, w *game.World, startY float32) {
	const (
		badgeSize = float32(20)
		step      = float32(23)
	)
	rx := float32(8)
	hoveredIdx := -1
	hoveredX := float32(0)

	for rk := 0; rk < int(game.RelicCount); rk++ {
		if !w.Player.Relics[rk] {
			continue
		}
		rect := game.Rect{X: rx, Y: startY, W: badgeSize, H: badgeSize}
		hovered := rect.Contains(r.mouseScreenX, r.mouseScreenY)
		bg := Premul(22, 14, 40, 224)
		border := PremulColor(colGold, 175)
		if hovered {
			bg = Premul(42, 26, 70, 242)
			border = color.RGBA{255, 255, 255, 255}
			hoveredIdx = rk
			hoveredX = rx
		}
		drawPillBadge(screen, rx, startY, badgeSize, badgeSize, bg, border)
		r.DrawSpriteIcon(screen, relicIconKey(game.RelicKind(rk)), rx+2, startY+2, 1.0)
		rx += step
	}

	// Tidy 1-line tooltip pill on hover
	if hoveredIdx >= 0 {
		spec := game.GetRelicSpec(game.RelicKind(hoveredIdx))
		tipW := float32(textPixelWidth(spec.Name, 1.0) + 12)
		tipX := clampF(hoveredX, 6, game.ScreenWidth-tipW-6)
		tipY := startY + badgeSize + 3
		drawPillBadge(screen, tipX, tipY, tipW, 16, Premul(16, 10, 30, 244), colGold)
		r.drawTextColored(screen, spec.Name, int(tipX+6), int(tipY+1), colGold)
	}
}

func (r *Renderer) drawCenterStatusBanner(screen *ebiten.Image, w *game.World) {
	if w.Boss.Active {
		const (
			barW = float32(268)
			barH = float32(22)
			barX = (game.ScreenWidth - barW) * 0.5
			barY = float32(29)
		)
		drawPillBadge(screen, barX, barY, barW, barH, Premul(26, 10, 30, 236), colKawaii)
		vector.DrawFilledRect(screen, barX+1, barY+1, barW-2, 2, PremulColor(colKawaii, 85), false)

		nameTxt := w.Boss.Name
		if w.BossRushMode {
			nameTxt = fmt.Sprintf("R%d/6 %s", w.BossesDefeated+1, w.Boss.Name)
		}
		r.drawScaledText(screen, nameTxt, int(barX+10), int(barY+4), 1.0, colKawaii, colOutline)

		// Segmented glowing shield blocks (typically 3) on right half of bar
		maxS := w.Boss.MaxShields
		if maxS < 1 {
			maxS = 3
		}
		const (
			blkW = float32(18)
			blkH = float32(10)
			gap  = float32(4)
		)
		totalS := float32(maxS)*blkW + float32(maxS-1)*gap
		startX := barX + barW - totalS - 10
		blkY := barY + (barH-blkH)*0.5

		pulse := uint8(215 + 40*math.Sin(float64(r.tick*8.0)))
		for s := 0; s < maxS; s++ {
			bx := startX + float32(s)*(blkW+gap)
			if s < w.Boss.Shields {
				// Active glowing crimson-gold shield block
				vector.DrawFilledRect(screen, bx, blkY, blkW, blkH, color.RGBA{235, 42, 92, 255}, false)
				vector.DrawFilledRect(screen, bx+1, blkY+1, blkW-2, 3, color.RGBA{255, pulse, 96, 255}, false)
				vector.StrokeRect(screen, bx, blkY, blkW, blkH, 1.2, colGold, false)
			} else {
				// Broken shield block: dark recessed shell + cracked diagonal
				vector.DrawFilledRect(screen, bx, blkY, blkW, blkH, Premul(18, 12, 26, 220), false)
				vector.StrokeRect(screen, bx, blkY, blkW, blkH, 1.0, Premul(110, 62, 88, 180), false)
				vector.StrokeLine(screen, bx+4, blkY+2, bx+blkW-5, blkY+blkH-2, 1.0, Premul(140, 70, 95, 190), true)
			}
		}
		return
	}

	if w.BossRushMode && w.BossScheduleIdx < len(w.BossSchedule) {
		lbl := fmt.Sprintf("BOSS RUSH %d/6 TRA %.0fs", w.BossesDefeated+1, w.NextBossRushTimer)
		bw := float32(textPixelWidth(lbl, 1.0) + 18)
		bx := (game.ScreenWidth - bw) * 0.5
		drawPillBadge(screen, bx, 29, bw, 17, Premul(24, 16, 36, 232), colGold)
		r.drawTextColored(screen, lbl, int(bx+9), 30, colGold)
		return
	}

	if w.Police.Heli.Active {
		lbl := "5* ELICOTTERO IN ZONA"
		bw := float32(textPixelWidth(lbl, 1.0) + 26)
		bx := (game.ScreenWidth - bw) * 0.5
		border := colRedAlert
		if int(r.tick*8)%2 == 0 {
			border = colGold
		}
		drawPillBadge(screen, bx, 29, bw, 17, Premul(44, 10, 20, 234), border)
		vector.DrawFilledCircle(screen, bx+10, 37.5, 3, border, true)
		r.drawTextColored(screen, lbl, int(bx+18), 30, color.White)
	}
}

func (r *Renderer) drawStreakBadge(screen *ebiten.Image, w *game.World, sy float32) {
	if w.StreakCount < 2 {
		return
	}
	const bw = float32(96)
	bx := float32(game.ScreenWidth) - bw - 8
	txtCol := colGyaru
	lbl := fmt.Sprintf("STREAK x%d", w.StreakCount)
	bgCol := Premul(16, 12, 32, 228)
	if w.StreakCount >= 5 {
		lbl = fmt.Sprintf("RUSH x%d!", w.StreakCount)
		txtCol = colGold
		if int(r.tick*8)%2 == 0 {
			txtCol = colKawaii
		}
		bgCol = Premul(52, 16, 54, 238)
	}
	drawPillBadge(screen, bx, sy, bw, 19, bgCol, txtCol)
	tw := textPixelWidth(lbl, 1.0)
	r.drawTextColored(screen, lbl, int(bx+(bw-float32(tw))*0.5), int(sy+1), txtCol)
	frac := clampF(w.StreakTimer/game.StreakWindowDuration, 0, 1)
	vector.DrawFilledRect(screen, bx+6, sy+15, (bw-12)*frac, 2.5, txtCol, false)
}

func (r *Renderer) drawChallengeToast(screen *ebiten.Image, w *game.World) {
	if r.toastTimer <= 0 || r.activeChallengeToast == "" {
		return
	}
	msg := "SFIDA: " + r.activeChallengeToast
	mw := float32(textPixelWidth(msg, 1.0) + 28)
	mx := (game.ScreenWidth - mw) * 0.5
	my := float32(55)
	if w.TutorialMode {
		my = 70
	}
	drawPillBadge(screen, mx, my, mw, 20, Premul(40, 24, 6, 242), colGold)
	r.DrawSpriteIcon(screen, assets.SpriteIconStarWanted, mx+6, my+4, 1.0)
	r.drawScaledText(screen, msg, int(mx+22), int(my+3), 1.0, colGold, colOutline)
}

func (r *Renderer) drawTutorialSandboxBar(screen *ebiten.Image, w *game.World) {
	const (
		barX = float32(8)
		barY = float32(28)
		barW = float32(624)
		barH = float32(36)
	)
	drawPillBadge(screen, barX, barY, barW, barH, Premul(16, 12, 32, 232), PremulColor(colGyaru, 185))
	vector.DrawFilledRect(screen, barX+1, barY+1, barW-2, 2, PremulColor(colGyaru, 55), false)

	// Concise 1-line mouse-first tip advancing with TutorialStep
	var tip string
	switch w.TutorialStep {
	case 0:
		tip = "1/4  MUOVI [WASD] + [CLICK SX] LANCIA FERMATA"
	case 1:
		tip = "2/4  RESTA NEL CERCHIO & [CLICK SX] PER MINIGIOCO + COMBO BUS"
	case 2:
		tip = "3/4  AVVICINATI & [CLICK SX] SU SLOT 777 O DISTRIBUTORE"
	default:
		tip = "4/4  PERFETTO! CONTINUA A PROVARE O TORNA AL MENU!"
	}
	r.DrawSpriteIcon(screen, assets.SpriteIconMouseLMB, barX+6, barY+2, 0.9)
	r.drawScaledText(screen, tip, int(barX+24), int(barY+2), 1.0, colGold, colOutline)

	// 4 tiny live status pills with crisp green checkmarks
	type tutPill struct {
		label string
		done  bool
		w     float32
	}
	pills := [4]tutPill{
		{"MOVIMENTO", w.TutMoved, 84},
		{"FERMATA", w.TutThrewStop, 74},
		{"BUS COMBO", w.TutBoardedBus, 84},
		{"SHOP/SLOT", w.TutUsedSlotOrShop, 84},
	}
	px := barX + 10
	py := barY + 19
	for _, p := range pills {
		bg := Premul(24, 18, 42, 220)
		border := Premul(115, 110, 145, 175)
		txtCol := color.RGBA{180, 185, 210, 230}
		if p.done {
			bg = Premul(16, 54, 34, 236)
			border = colGreenNeon
			txtCol = colGreenNeon
		}
		drawPillBadge(screen, px, py, p.w, 13, bg, border)
		if p.done {
			// Crisp vector green checkmark
			cx, cy := px+6, py+6.5
			vector.StrokeLine(screen, cx-2.5, cy, cx-0.5, cy+2.5, 1.6, colGreenNeon, true)
			vector.StrokeLine(screen, cx-0.5, cy+2.5, cx+3.5, cy-2.5, 1.6, colGreenNeon, true)
		} else {
			vector.StrokeCircle(screen, px+6, py+6.5, 2.5, 1.0, border, true)
		}
		r.drawTextColored(screen, p.label, int(px+13), int(py-1), txtCol)
		px += p.w + 6
	}

	// Compact clickable [GUIDA] and [MENU] buttons on right side
	guideRect := TutorialGuideRect()
	r.drawTutorialActionButton(screen, guideRect, "GUIDA", colGyaru, guideRect.Contains(r.mouseScreenX, r.mouseScreenY))

	exitRect := TutorialExitRect()
	r.drawTutorialActionButton(screen, exitRect, "MENU", colKawaii, exitRect.Contains(r.mouseScreenX, r.mouseScreenY))
}

func (r *Renderer) drawTutorialActionButton(screen *ebiten.Image, rect game.Rect, label string, accent color.RGBA, hovered bool) {
	bg := Premul(28, 20, 48, 238)
	border := accent
	txtCol := color.RGBA{255, 255, 255, 255}
	if hovered {
		bg = PremulColor(accent, 95)
		border = colGold
		txtCol = colGold
	}
	drawPillBadge(screen, rect.X, rect.Y, rect.W, rect.H, bg, border)
	tw := textPixelWidth(label, 1.0)
	tx := int(rect.X + (rect.W-float32(tw))*0.5)
	ty := int(rect.Y + (rect.H-14)*0.5)
	r.drawScaledText(screen, label, tx, ty, 1.0, txtCol, colOutline)
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
	// Nearest VendingShop offscreen marker when player has cash >= Espresso cost
	if w.Player.Cash >= game.EspressoHealCost && len(w.Shops) > 0 {
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
	// Nearest SlotMachine offscreen marker when player has cash >= spin cost
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

	// Outer dark-glass dock frame extends slightly left for the mouse-wheel cue while preserving HotbarSlotRect(0..3)
	const wheelPad = float32(26)
	frameX := dockX - wheelPad - 4
	frameY := dockY - 4
	frameW := dockW + wheelPad + 8
	frameH := cardH + 8
	vector.DrawFilledRect(screen, frameX, frameY, frameW, frameH, Premul(14, 9, 26, 224), false)
	vector.DrawFilledRect(screen, frameX+1, frameY+1, frameW-2, 1.5, Premul(255, 255, 255, 22), false)
	vector.StrokeRect(screen, frameX, frameY, frameW, frameH, 1.2, PremulColor(colGyaru, 145), false)

	// Left Mouse-Wheel Cue (swap stop archetype)
	r.DrawSpriteIcon(screen, assets.SpriteIconMouseWheel, dockX-wheelPad+3, dockY+3, 1.0)
	r.drawTextColored(screen, "1-4", int(dockX-wheelPad+2), int(dockY+22), color.RGBA{185, 200, 228, 225})

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

	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
	for i, s := range slots {
		rect := HotbarSlotRect(i)
		cx, cy := rect.X, rect.Y
		ac := archetypeColor(s.arch)
		selected := w.Player.SelectedStopArch == s.arch
		hovered := rect.Contains(r.mouseScreenX, r.mouseScreenY)

		cardBg := Premul(24, 16, 40, 236)
		borderCol := PremulColor(ac, 145)
		if hovered && !selected {
			cardBg = Premul(38, 26, 60, 244)
			borderCol = color.RGBA{255, 255, 255, 255}
		}
		if selected {
			cy -= 3
			cardBg = Premul(44, 28, 72, 248)
			borderCol = colGold
			vector.StrokeRect(screen, cx-2, cy-2, cardW+4, cardH+4, 2.0, PremulColor(ac, 235), false)
		}

		vector.DrawFilledRect(screen, cx, cy, cardW, cardH, cardBg, false)
		vector.DrawFilledRect(screen, cx+1, cy+1, cardW-2, 2, PremulColor(ac, 140), false)
		vector.StrokeRect(screen, cx, cy, cardW, cardH, 1.3, borderCol, false)

		// Compact slot number pill + short archetype tag
		numCol := color.RGBA{185, 190, 215, 230}
		if selected {
			numCol = colGold
		}
		r.drawTextColored(screen, s.key, int(cx+5), int(cy+2), numCol)
		r.drawTextColored(screen, s.name, int(cx+18), int(cy+2), ac)

		// Crisp stop shelter sprite + matching girl sprite
		stopImg := r.sprites.Get(s.stopSpr)
		girlImg := r.sprites.Get(s.girlSpr)

		op.GeoM.Reset()
		op.GeoM.Scale(20.0/float64(stopImg.Bounds().Dx()), 20.0/float64(stopImg.Bounds().Dy()))
		op.GeoM.Translate(math.Round(float64(cx+7)), math.Round(float64(cy+16)))
		op.ColorScale.Reset()
		screen.DrawImage(stopImg, op)

		vector.DrawFilledCircle(screen, cx+36, cy+26, 1.6, PremulColor(ac, 210), true)

		op.GeoM.Reset()
		op.GeoM.Scale(18.0/float64(girlImg.Bounds().Dx()), 18.0/float64(girlImg.Bounds().Dy()))
		op.GeoM.Translate(math.Round(float64(cx+44)), math.Round(float64(cy+17)))
		screen.DrawImage(girlImg, op)

		// Smooth cooldown progress sweep
		if cdFrac > 0 {
			ovH := cardH * cdFrac
			vector.DrawFilledRect(screen, cx+1, cy+cardH-ovH, cardW-2, ovH, Premul(8, 5, 18, 155), false)
			vector.DrawFilledRect(screen, cx+1, cy+cardH-ovH, cardW-2, 1.5, PremulColor(colGold, 200), false)
		}
	}

	// Right Mouse-LMB Launch Cue + Active Stop Counter + Readiness Gauge
	statusX := dockX + totalCardsW + 5
	r.DrawSpriteIcon(screen, assets.SpriteIconMouseLMB, statusX+1, dockY+1, 1.0)
	r.drawTextColored(screen, "LANCIA", int(statusX+20), int(dockY+2), color.White)

	slotStr := fmt.Sprintf("%d/%d", w.ActiveStopCount(), w.Player.EffectiveMaxStops())
	r.drawScaledText(screen, slotStr, int(statusX+2), int(dockY+21), 1.0, colGyaru, colOutline)

	barX := statusX + 28
	barW := float32(39)
	readyFrac := 1.0 - cdFrac
	barCol := colGreenNeon
	if cdFrac > 0 {
		barCol = colGold
	}
	vector.DrawFilledRect(screen, barX, dockY+24, barW, 8, Premul(20, 16, 36, 232), false)
	vector.DrawFilledRect(screen, barX, dockY+24, barW*readyFrac, 8, barCol, false)
	vector.StrokeRect(screen, barX, dockY+24, barW, 8, 1.0, barCol, false)

	// Grounded / Punizione warning pill cleanly above the dock
	if w.Player.GroundedTimer > 0 {
		warnTxt := fmt.Sprintf("PUNIZIONE %.0fs!", w.Player.GroundedTimer)
		ww := float32(textPixelWidth(warnTxt, 1.0) + 18)
		wx := (game.ScreenWidth - ww) * 0.5
		drawPillBadge(screen, wx, dockY-24, ww, 16, Premul(58, 12, 22, 240), colRedAlert)
		r.drawScaledText(screen, warnTxt, int(wx+9), int(dockY-23), 1.0, colRedAlert, colOutline)
	}
}

func upgradeTheme(kind game.UpgradeKind) (iconKey string, accent color.RGBA) {
	switch kind {
	case game.UpgradeCooldown:
		return assets.SpriteIconUpgCooldown, colGyaru
	case game.UpgradeMultiHook:
		return assets.SpriteIconUpgMultiHook, colKawaii
	case game.UpgradeMaxStops:
		return assets.SpriteIconUpgMaxStops, colGold
	case game.UpgradeRange:
		return assets.SpriteIconUpgRange, color.RGBA{64, 224, 208, 255}
	case game.UpgradeChronos:
		return assets.SpriteIconUpgChronos, color.RGBA{176, 108, 255, 255}
	case game.UpgradeVitality:
		return assets.SpriteIconUpgVitality, color.RGBA{255, 76, 118, 255}
	default:
		return assets.SpriteIconUpgCooldown, colGyaru
	}
}

func (r *Renderer) drawLevelUpModal(screen *ebiten.Image, w *game.World) {
	vector.DrawFilledRect(screen, 0, 0, game.ScreenWidth, game.ScreenHeight, Premul(10, 6, 20, 204), false)

	title := fmt.Sprintf("LIVELLO %d - SCEGLI UPGRADE", w.Player.Level)
	tw := textPixelWidth(title, 1.35)
	r.drawScaledText(screen, title, (game.ScreenWidth-tw)/2, 38, 1.35, colGold, colOutline)

	sub := "CLICCA UNA CARTA O PREMI [1 / 2 / 3]"
	sw := textPixelWidth(sub, 1.0)
	subX := (game.ScreenWidth-sw)/2 + 8
	r.DrawSpriteIcon(screen, assets.SpriteIconMouseLMB, float32(subX-20), 63, 1.0)
	r.drawTextColored(screen, sub, subX, 64, color.RGBA{205, 225, 248, 232})

	for i := 0; i < 3; i++ {
		card := w.UpgradeChoices[i]
		rect := LevelUpCardRect(i)
		hovered := rect.Contains(r.mouseScreenX, r.mouseScreenY)
		iconKey, accent := upgradeTheme(card.Kind)

		drawY := rect.Y
		bg := Premul(20, 13, 36, 240)
		border := accent
		if hovered {
			drawY -= 3
			bg = Premul(34, 22, 58, 248)
			border = colGold
			// Subtle outer neon halo on hover
			vector.StrokeRect(screen, rect.X-2, drawY-2, rect.W+4, rect.H+4, 2.2, PremulColor(accent, 195), false)
		}

		// Card dark-glass frame + top theme trim
		vector.DrawFilledRect(screen, rect.X, drawY, rect.W, rect.H, bg, false)
		vector.DrawFilledRect(screen, rect.X, drawY, rect.W, 4, accent, false)
		vector.StrokeRect(screen, rect.X, drawY, rect.W, rect.H, 1.8, border, false)

		// Glowing 2x Icon Medallion (20x20 sprite * 2.0 = 40x40!)
		medCX := rect.X + rect.W*0.5
		medCY := drawY + 40
		vector.DrawFilledCircle(screen, medCX, medCY, 27, PremulColor(accent, 42), true)
		vector.DrawFilledCircle(screen, medCX, medCY, 23, Premul(14, 10, 28, 242), true)
		vector.StrokeCircle(screen, medCX, medCY, 23.5, 1.6, border, true)
		r.DrawSpriteIcon(screen, iconKey, medCX-20, medCY-20, 2.0)

		// Bold Clear Centered Title
		titleW := textPixelWidth(card.Title, 1.0)
		r.drawScaledText(screen, card.Title, int(rect.X+(rect.W-float32(titleW))*0.5), int(drawY+73), 1.0, accent, colOutline)

		// Subtle Divider
		vector.DrawFilledRect(screen, rect.X+16, drawY+91, rect.W-32, 1, PremulColor(accent, 110), false)

		// Concise Centered 1-2 Lines Effect Text
		drawWrappedTextCentered(r, screen, card.Desc, rect.X+rect.W*0.5, int(drawY+98), 24, color.RGBA{240, 244, 255, 255})

		// Single Clean Footer Button: CLICK / [1]
		btnX := rect.X + 18
		btnY := drawY + rect.H - 32
		btnW := rect.W - 36
		btnH := float32(22)
		btnBg := Premul(30, 22, 54, 236)
		btnBorder := PremulColor(accent, 190)
		btnTxtCol := color.RGBA{235, 240, 255, 255}
		if hovered {
			btnBg = PremulColor(accent, 95)
			btnBorder = colGold
			btnTxtCol = colGold
		}
		drawPillBadge(screen, btnX, btnY, btnW, btnH, btnBg, btnBorder)
		btnLbl := fmt.Sprintf("CLICK / [%d]", i+1)
		bw := textPixelWidth(btnLbl, 1.0)
		r.drawScaledText(screen, btnLbl, int(btnX+(btnW-float32(bw))*0.5), int(btnY+4), 1.0, btnTxtCol, colOutline)
	}
}

func (r *Renderer) drawShopModal(screen *ebiten.Image, w *game.World) {
	vector.DrawFilledRect(screen, 0, 0, game.ScreenWidth, game.ScreenHeight, Premul(10, 6, 18, 200), false)
	const (
		boxX = float32(124)
		boxY = float32(40)
		boxW = float32(392)
		boxH = float32(268)
	)
	vector.DrawFilledRect(screen, boxX, boxY, boxW, boxH, Premul(16, 10, 30, 242), false)
	vector.DrawFilledRect(screen, boxX, boxY, boxW, 3, colGyaru, false)
	vector.StrokeRect(screen, boxX, boxY, boxW, boxH, 1.8, colGyaru, false)

	// Top Header + Live Alfredo Status Strip (HP, Cash, Wanted Stars)
	r.drawScaledText(screen, "DISTRIBUTORE NEON", int(boxX+14), int(boxY+10), 1.15, colGyaru, colOutline)

	// HP status pill
	hpTxt := fmt.Sprintf("%d/%d", w.Player.HP, w.Player.MaxHP)
	drawPillBadge(screen, boxX+206, boxY+8, 54, 18, Premul(36, 14, 28, 232), colKawaii)
	r.DrawSpriteIcon(screen, assets.SpriteIconHeartFull, boxX+211, boxY+11, 1.0)
	r.drawScaledText(screen, hpTxt, int(boxX+226), int(boxY+10), 1.0, color.White, colOutline)

	// Cash status pill
	cashTxt := fmt.Sprintf("$%d", w.Player.Cash)
	drawPillBadge(screen, boxX+264, boxY+8, 66, 18, Premul(36, 26, 8, 232), colGold)
	r.DrawSpriteIcon(screen, assets.SpriteIconCoin, boxX+269, boxY+11, 1.0)
	r.drawScaledText(screen, cashTxt, int(boxX+284), int(boxY+10), 1.0, colGold, colOutline)

	// Wanted stars status pill
	starTxt := fmt.Sprintf("%d*", w.Police.Stars)
	starBorder := Premul(115, 110, 145, 180)
	if w.Police.Stars > 0 {
		starBorder = colGold
	}
	drawPillBadge(screen, boxX+334, boxY+8, 44, 18, Premul(28, 20, 42, 232), starBorder)
	if w.Police.Stars > 0 {
		r.DrawSpriteIcon(screen, assets.SpriteIconStarWanted, boxX+339, boxY+11, 1.0)
	} else {
		r.drawSpriteIconTinted(screen, assets.SpriteIconStarWanted, boxX+339, boxY+11, 1.0, color.RGBA{75, 65, 95, 180})
	}
	r.drawScaledText(screen, starTxt, int(boxX+354), int(boxY+10), 1.0, colGold, colOutline)

	// Recessed Feedback Message Strip
	drawPillBadge(screen, 144, 74, 352, 22, Premul(12, 8, 24, 230), PremulColor(colGold, 130))
	msgW := textPixelWidth(w.ShopMenu.Message, 1.0)
	r.drawTextColored(screen, w.ShopMenu.Message, 144+(352-msgW)/2, 77, colGold)

	healCost := w.CurrentHealCost()
	bribeCost := w.CurrentBribeCost()

	type shopOption struct {
		icon   string
		title  string
		effect string
		cost   int
		col    color.RGBA
	}
	opts := [4]shopOption{
		{assets.SpriteIconShopEspresso, "Caffe Espresso", "(+2 HP & Chronos)", healCost, colKawaii},
		{assets.SpriteIconShopBribe, "Tangente Polizia", "(-2* Wanted)", bribeCost, colGold},
		{assets.SpriteIconShopRelic, "Reliquia Casuale", "(Bonus Passivo)", 120, colGyaru},
		{assets.SpriteIconShopCrate, "Cassa Upgrade", "(Scegli 3 Carte)", 140, colGreenNeon},
	}

	for i, opt := range opts {
		rect := ShopChoiceRect(i)
		hovered := rect.Contains(r.mouseScreenX, r.mouseScreenY)
		affordable := w.Player.Cash >= opt.cost
		if i == 1 && w.Police.Stars == 0 {
			affordable = false
		}

		bg := Premul(24, 16, 42, 236)
		border := PremulColor(opt.col, 175)
		if hovered && affordable {
			bg = Premul(44, 28, 72, 248)
			border = colGold
		}
		if !affordable {
			bg = Premul(20, 16, 30, 215)
			border = Premul(85, 80, 102, 185)
		}

		vector.DrawFilledRect(screen, rect.X, rect.Y, rect.W, rect.H, bg, false)
		vector.StrokeRect(screen, rect.X, rect.Y, rect.W, rect.H, 1.4, border, false)

		// Crisp 16x16 pixel icon inside dark-glass icon box
		iconBoxBg := Premul(14, 10, 28, 235)
		drawPillBadge(screen, rect.X+4, rect.Y+3, 24, 24, iconBoxBg, border)
		if affordable {
			r.DrawSpriteIcon(screen, opt.icon, rect.X+8, rect.Y+7, 1.0)
		} else {
			r.drawSpriteIconTinted(screen, opt.icon, rect.X+8, rect.Y+7, 1.0, color.RGBA{130, 130, 145, 190})
		}

		// Clean Title + Effect
		titleCol := color.RGBA{255, 255, 255, 255}
		effCol := opt.col
		if !affordable {
			titleCol = color.RGBA{135, 135, 155, 220}
			effCol = color.RGBA{105, 105, 125, 210}
		}
		tx := int(rect.X + 35)
		ty := int(rect.Y + 8)
		r.drawTextColored(screen, opt.title, tx, ty, titleCol)
		effX := tx + textPixelWidth(opt.title, 1.0) + 6
		r.drawTextColored(screen, opt.effect, effX, ty, effCol)

		// Bright Price Pill on right
		priceStr := fmt.Sprintf("$%d", opt.cost)
		const (
			pillW = float32(48)
			pillH = float32(20)
		)
		pillX := rect.X + rect.W - pillW - 6
		pillY := rect.Y + 5
		priceBg := Premul(16, 62, 36, 238)
		priceBorder := colGreenNeon
		priceCol := colGreenNeon
		if w.Player.Cash < opt.cost {
			priceBg = Premul(62, 16, 26, 232)
			priceBorder = colRedAlert
			priceCol = colRedAlert
		} else if i == 1 && w.Police.Stars == 0 {
			priceBg = Premul(32, 30, 44, 225)
			priceBorder = Premul(110, 105, 130, 190)
			priceCol = color.RGBA{150, 150, 170, 220}
		}
		drawPillBadge(screen, pillX, pillY, pillW, pillH, priceBg, priceBorder)
		pw := textPixelWidth(priceStr, 1.0)
		r.drawScaledText(screen, priceStr, int(pillX+(pillW-float32(pw))*0.5), int(pillY+3), 1.0, priceCol, colOutline)
	}

	// Clean Mouse Hint
	hint := "CLICK SX O [1-4] ACQUISTA"
	hw := textPixelWidth(hint, 1.0)
	hx := (game.ScreenWidth-hw)/2 + 8
	r.DrawSpriteIcon(screen, assets.SpriteIconMouseLMB, float32(hx-19), 251, 0.9)
	r.drawTextColored(screen, hint, hx, 252, color.RGBA{185, 200, 228, 220})

	// Close Button at exact ShopCloseRect()
	closeRect := ShopCloseRect()
	closeHover := closeRect.Contains(r.mouseScreenX, r.mouseScreenY)
	cBg := Premul(58, 20, 36, 238)
	cBorder := colKawaii
	if closeHover {
		cBg = Premul(98, 32, 56, 248)
		cBorder = colGold
	}
	drawPillBadge(screen, closeRect.X, closeRect.Y, closeRect.W, closeRect.H, cBg, cBorder)
	closeLbl := "[X] CHIUDI"
	cw := textPixelWidth(closeLbl, 1.0)
	r.drawScaledText(screen, closeLbl, int(closeRect.X+(closeRect.W-float32(cw))*0.5), int(closeRect.Y+5), 1.0, color.White, colOutline)
}

func slotSymbolMeta(sym int) (iconKey string, label string, clr color.RGBA) {
	switch sym % game.SymCount {
	case game.SymCherry:
		return assets.SpriteIconSlotCherry, "CILIEGIA", colKawaii
	case game.SymCash:
		return assets.SpriteIconSlotCash, "CASH", colGreenNeon
	case game.SymXP:
		return assets.SpriteIconSlotXP, "XP+", colStudent
	case game.SymHeart:
		return assets.SpriteIconSlotHeart, "CUORE", color.RGBA{255, 90, 145, 255}
	case game.SymGoldBus:
		return assets.SpriteIconSlot777, "777", colGold
	case game.SymSiren:
		return assets.SpriteIconSlotSiren, "SIRENA", colRedAlert
	default:
		return assets.SpriteIconSlot777, "777", colGold
	}
}

func (r *Renderer) drawSlotModal(screen *ebiten.Image, w *game.World) {
	vector.DrawFilledRect(screen, 0, 0, game.ScreenWidth, game.ScreenHeight, Premul(10, 6, 18, 204), false)
	const (
		boxX = float32(132)
		boxY = float32(32)
		boxW = float32(376)
		boxH = float32(274)
	)
	vector.DrawFilledRect(screen, boxX, boxY, boxW, boxH, Premul(18, 10, 32, 244), false)
	vector.DrawFilledRect(screen, boxX, boxY, boxW, 3, colGold, false)
	vector.StrokeRect(screen, boxX, boxY, boxW, boxH, 2.0, colGold, false)

	// Cabinet Header
	r.drawScaledText(screen, "RIZZ-SLOTS 777", int(boxX+18), int(boxY+10), 1.3, colGold, colOutline)
	cashTxt := fmt.Sprintf("$%d", w.Player.Cash)
	drawPillBadge(screen, boxX+boxW-84, boxY+9, 68, 19, Premul(36, 26, 8, 236), colGold)
	r.DrawSpriteIcon(screen, assets.SpriteIconCoin, boxX+boxW-78, boxY+12, 1.0)
	r.drawScaledText(screen, cashTxt, int(boxX+boxW-63), int(boxY+11), 1.0, colGold, colOutline)

	// 3 Animated Icon-Driven Reels (2x 20x20 = 40x40 pixel-art sprites!)
	lockTimes := [3]float32{0.40, 0.75, 1.10}
	for i := 0; i < 3; i++ {
		const (
			rw  = float32(92)
			rh  = float32(76)
			gap = float32(14)
		)
		rx := boxX + 36 + float32(i)*(rw+gap)
		ry := boxY + 34
		sym := w.SlotModal.Reels[i]
		iconKey, lbl, sc := slotSymbolMeta(sym)

		// Reel cylinder housing
		vector.DrawFilledRect(screen, rx, ry, rw, rh, Premul(20, 13, 36, 248), false)
		vector.DrawFilledRect(screen, rx+2, ry+2, rw-4, 8, Premul(0, 0, 0, 65), false)
		vector.DrawFilledRect(screen, rx+2, ry+rh-10, rw-4, 8, Premul(0, 0, 0, 75), false)

		borderCol := sc
		jitterY := float32(0)
		if w.SlotModal.Spinning && w.SlotModal.SpinTimer < lockTimes[i] {
			jitterY = float32(math.Sin(float64(r.tick*48.0+float32(i)*2.1))) * 2.5
			borderCol = PremulColor(colGold, 190)
		} else if w.SlotModal.ResultKind == 2 && int(r.tick*8)%2 == 0 {
			borderCol = color.RGBA{255, 255, 255, 255}
		}
		vector.StrokeRect(screen, rx, ry, rw, rh, 2.0, borderCol, false)

		// Centered 40x40 pixel-art icon
		r.DrawSpriteIcon(screen, iconKey, rx+(rw-40)*0.5, ry+9+jitterY, 2.0)

		// Tiny crisp label beneath icon
		lw := textPixelWidth(lbl, 1.0)
		r.drawTextColored(screen, lbl, int(rx+(rw-float32(lw))*0.5), int(ry+55), sc)
	}

	// Compact Visual Prize Legend Strip using tiny pixel icons
	const (
		legX = boxX + 18
		legY = boxY + 116
		legW = boxW - 36
		legH = float32(24)
	)
	drawPillBadge(screen, legX, legY, legW, legH, Premul(14, 9, 26, 232), PremulColor(colGold, 125))
	type legendEntry struct {
		icon string
		text string
		col  color.RGBA
		ox   float32
	}
	legend := [4]legendEntry{
		{assets.SpriteIconSlot777, "3x $250", colGold, 6},
		{assets.SpriteIconSlotCash, "3x $140", colGreenNeon, 92},
		{assets.SpriteIconSlotHeart, "3x +3HP", colKawaii, 178},
		{assets.SpriteIconSlotSiren, "3x +1*", colRedAlert, 264},
	}
	for _, le := range legend {
		r.DrawSpriteIcon(screen, le.icon, legX+le.ox, legY+4, 0.8)
		r.drawTextColored(screen, le.text, int(legX+le.ox+18), int(legY+5), le.col)
	}

	// Result Readout Box
	resCol := colGyaru
	switch w.SlotModal.ResultKind {
	case 1:
		resCol = colGreenNeon
	case 2:
		resCol = colGold
	case 3:
		resCol = colRedAlert
	}
	const (
		resX = boxX + 18
		resY = boxY + 146
		resW = boxW - 36
		resH = float32(64)
	)
	drawPillBadge(screen, resX, resY, resW, resH, Premul(14, 10, 28, 236), resCol)
	rtW := textPixelWidth(w.SlotModal.ResultTitle, 1.05)
	r.drawScaledText(screen, w.SlotModal.ResultTitle, int(resX+(resW-float32(rtW))*0.5), int(resY+10), 1.05, resCol, colOutline)
	drawWrappedTextCentered(r, screen, w.SlotModal.ResultDesc, resX+resW*0.5, int(resY+31), 50, color.White)

	// SPIN Button at exact SlotSpinRect()
	spinCost := w.CurrentSlotCost()
	spinRect := SlotSpinRect()
	spinHover := spinRect.Contains(r.mouseScreenX, r.mouseScreenY)
	canSpin := !w.SlotModal.Spinning && w.Player.Cash >= spinCost
	sBg := Premul(24, 82, 46, 242)
	sBorder := colGreenNeon
	if !canSpin {
		sBg = Premul(36, 32, 50, 224)
		sBorder = Premul(105, 100, 122, 210)
	} else if spinHover {
		sBg = Premul(38, 118, 66, 248)
		sBorder = colGold
	}
	drawPillBadge(screen, spinRect.X, spinRect.Y, spinRect.W, spinRect.H, sBg, sBorder)
	spinTxt := fmt.Sprintf("GIRA ($%d)", spinCost)
	if w.SlotModal.Spinning {
		spinTxt = "GIRANDO..."
	}
	stw := textPixelWidth(spinTxt, 1.1)
	iconOffset := float32(0)
	if !w.SlotModal.Spinning {
		iconOffset = 10
		r.DrawSpriteIcon(screen, assets.SpriteIconMouseLMB, spinRect.X+(spinRect.W-float32(stw))*0.5-14, spinRect.Y+7, 1.0)
	}
	r.drawScaledText(screen, spinTxt, int(spinRect.X+(spinRect.W-float32(stw))*0.5+iconOffset), int(spinRect.Y+7), 1.1, color.White, colOutline)

	// EXIT Button at exact SlotCloseRect()
	closeRect := SlotCloseRect()
	closeHover := closeRect.Contains(r.mouseScreenX, r.mouseScreenY)
	cBg := Premul(58, 20, 36, 238)
	cBorder := colKawaii
	if closeHover {
		cBg = Premul(98, 32, 56, 248)
		cBorder = colGold
	}
	drawPillBadge(screen, closeRect.X, closeRect.Y, closeRect.W, closeRect.H, cBg, cBorder)
	closeTxt := "ESCI"
	ctw := textPixelWidth(closeTxt, 1.1)
	r.drawScaledText(screen, closeTxt, int(closeRect.X+(closeRect.W-float32(ctw))*0.5), int(closeRect.Y+7), 1.1, color.White, colOutline)
}

func drawWrappedTextCentered(r *Renderer, screen *ebiten.Image, desc string, centerX float32, startY, cut int, clr color.Color) {
	lineY := startY
	for len(desc) > 0 {
		var line string
		if len(desc) <= cut {
			line = desc
			desc = ""
		} else {
			sp := cut
			for sp > 8 && desc[sp] != ' ' {
				sp--
			}
			if desc[sp] != ' ' {
				sp = cut
			}
			line = desc[:sp]
			desc = desc[sp:]
			if len(desc) > 0 && desc[0] == ' ' {
				desc = desc[1:]
			}
		}
		lw := textPixelWidth(line, 1.0)
		r.drawTextColored(screen, line, int(centerX-float32(lw)*0.5), lineY, clr)
		lineY += 15
	}
}
