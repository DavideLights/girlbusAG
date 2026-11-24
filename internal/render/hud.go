package render

import (
	"fmt"
	"image/color"
	"math"
	"strings"

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

// DrawHUDAndOverlays renders the HUD, diegetic controls, and active modal.
func (r *Renderer) DrawHUDAndOverlays(screen *ebiten.Image, w *game.World) {
	r.drawTopHUDBar(screen, w)

	subBarY := float32(29)
	if w.TutorialMode {
		r.drawTutorialSandboxBar(screen, w)
		subBarY = 70
	} else {
		r.drawCenterStatusBanner(screen, w)
	}
	r.drawStreakBadge(screen, w, subBarY)

	toastY := subBarY + 26
	if r.drawChallengeToast(screen, toastY) {
		toastY += 39
	}
	r.drawRelicToast(screen, w, toastY)
	r.drawRelicTray(screen, w, subBarY)

	r.drawBottomHotbarDock(screen, w)

	if r.playingActive && !w.PendingLevelUp && !w.ShopMenu.Active && !w.SlotModal.Active {
		r.drawDiegeticStopUI(screen, w, r.lastCamX, r.lastCamY)
		r.drawCriticalOffscreenArrows(screen, w)
	}

	if w.PendingLevelUp {
		r.drawLevelUpModal(screen, w)
	} else if w.ShopMenu.Active {
		r.drawShopModal(screen, w)
	} else if w.SlotModal.Active {
		r.drawSlotModal(screen, w)
	}
}

var (
	runCream = color.RGBA{255, 240, 216, 255}
	runMuted = color.RGBA{200, 181, 190, 255}
	runInk   = color.RGBA{28, 19, 34, 255}
	runPlum  = color.RGBA{37, 28, 45, 255}
	runFace  = color.RGBA{48, 36, 56, 255}
	runLine  = color.RGBA{102, 76, 104, 255}
	runTeal  = color.RGBA{114, 226, 197, 255}
	runCoral = color.RGBA{255, 136, 122, 255}
	runGold  = color.RGBA{255, 209, 133, 255}
	runLilac = color.RGBA{190, 143, 213, 255}
	runSky   = color.RGBA{136, 194, 237, 255}
	runPink  = color.RGBA{255, 153, 188, 255}
)

func drawRunPanel(screen *ebiten.Image, box game.Rect, face, border color.Color) {
	x, y, w, h := box.X, box.Y, box.W, box.H
	vector.DrawFilledRect(screen, x+2, y+3, w, h, Premul(20, 12, 27, 182), false)
	vector.DrawFilledRect(screen, x+1, y, w-2, h, border, false)
	vector.DrawFilledRect(screen, x, y+1, w, h-2, border, false)
	vector.DrawFilledRect(screen, x+1, y+1, w-2, h-2, face, false)
}

func drawRunMeter(screen *ebiten.Image, x, y, width, height, fraction float32, accent color.RGBA) {
	vector.DrawFilledRect(screen, x, y, width, height, runInk, false)
	vector.DrawFilledRect(screen, x, y+height-1, width, 1, runLine, false)
	fill := float32(math.Round(float64((width - 2) * clampF(fraction, 0, 1))))
	if fill > 0 {
		vector.DrawFilledRect(screen, x+1, y+1, fill, height-2, accent, false)
	}
}

func (r *Renderer) drawRunCentered(screen *ebiten.Image, label string, centerX float32, y int, tint color.Color) {
	r.drawTextColored(screen, label, int(centerX-float32(textPixelWidth(label, 1))*0.5), y, tint)
}

func (r *Renderer) drawTopHUDBar(screen *ebiten.Image, w *game.World) {
	const cashX = 395
	cash := fmt.Sprintf("$%d", w.Player.Cash)
	hp := fmt.Sprintf("%d/%d", w.Player.HP, w.Player.MaxHP)
	hpX := max(466, cashX+textPixelWidth(cash, 1)+27)
	hpDivider := float32(hpX - 23)
	hpWidth := max(textPixelWidth(hp, 1), textPixelWidth("SALUTE", 1))
	xpDivider := float32(hpX + hpWidth + 8)
	hpTint := runPink
	if w.Player.HP <= 2 {
		hpTint = runCoral
	}
	vector.DrawFilledRect(screen, 0, 0, game.ScreenWidth, 24, Premul(37, 28, 45, 248), false)
	vector.DrawFilledRect(screen, 0, 24, game.ScreenWidth, 2, runInk, false)
	vector.DrawFilledRect(screen, 0, 24, game.ScreenWidth, 1, runLine, false)
	for _, segment := range [...]struct {
		left, right float32
		accent      color.RGBA
	}{
		{0, 136, runTeal}, {136, 270, runCoral}, {270, 378, runSky},
		{378, hpDivider, runGold}, {hpDivider, xpDivider, hpTint}, {xpDivider, game.ScreenWidth, runLilac},
	} {
		width := segment.right - segment.left
		vector.DrawFilledRect(screen, segment.left+1, 0, width-1, 23, PremulColor(segment.accent, 21), false)
		vector.DrawFilledRect(screen, segment.left+1, 23, width-1, 1, PremulColor(segment.accent, 118), false)
		vector.DrawFilledRect(screen, segment.left, 4, 1, 16, runLine, false)
	}

	chronos := float32(0)
	if w.Player.ChronosMax > 0 {
		chronos = w.Player.ChronosMeter / w.Player.ChronosMax
	}
	chronosColor := runTeal
	chronosHint := "SHIFT/Q"
	if w.Player.ChronosActive {
		chronosColor = runGold
		chronosHint = "ATTIVO"
	}
	r.DrawSpriteIcon(screen, assets.SpriteIconBolt, 12, 5, 1)
	r.drawTextColored(screen, "CHRONOS", 27, -2, chronosColor)
	r.drawTextColored(screen, chronosHint, 87, -2, runMuted)
	drawRunMeter(screen, 28, 15, 96, 6, chronos, chronosColor)

	r.drawTextColored(screen, "RICERCATO", 146, -2, runCoral)
	for i := 0; i < 5; i++ {
		x := 148 + float32(i)*14
		if i < w.Police.Stars {
			r.DrawSpriteIcon(screen, assets.SpriteIconStarWanted, x, 11, 1)
		} else {
			r.drawSpriteIconTinted(screen, assets.SpriteIconStarWanted, x, 11, 1, Premul(201, 181, 199, 135))
		}
	}
	r.drawTextColored(screen, fmt.Sprintf("%d/5", w.Police.Stars), 225, 10, runCream)
	if w.Police.Stars > 0 {
		multiplier, _ := w.Police.XPAndCashMultiplier()
		tint := runGold
		if w.Police.Stars >= 4 {
			tint = runCoral
		}
		r.drawTextColored(screen, fmt.Sprintf("%.1fx", multiplier), 225, -2, tint)
	}

	elapsed := int(w.RunTime)
	runTime := fmt.Sprintf("%02d:%02d", elapsed/60, elapsed%60)
	tier := fmt.Sprintf("FASE %d", w.DifficultyTier)
	timerWidth := textPixelWidth(runTime, 2)
	if timerWidth+5+textPixelWidth(tier, 1) > 102 {
		tier = fmt.Sprintf("F%d", w.DifficultyTier)
	}
	timerX := 323 - (timerWidth+5+textPixelWidth(tier, 1))/2
	r.drawScaledText(screen, runTime, timerX, -3, 2, runCream, runInk)
	r.drawTextColored(screen, tier, timerX+timerWidth+5, 4, runSky)

	r.DrawSpriteIcon(screen, assets.SpriteIconCoin, cashX-15, 6, 1)
	r.drawTextColored(screen, "CASSA", cashX, -2, runGold)
	r.drawTextColored(screen, cash, cashX, 10, runCream)

	r.DrawSpriteIcon(screen, assets.SpriteIconHeartFull, float32(hpX-15), 6, 1)
	r.drawTextColored(screen, "SALUTE", hpX, -2, hpTint)
	r.drawTextColored(screen, hp, hpX, 10, hpTint)

	xpX := xpDivider + 12
	xpRight := float32(626)
	if r.showFPS {
		xpRight = 602
		r.drawRunCentered(screen, fmt.Sprintf("%.0f", ebiten.ActualFPS()), 621, 4, runMuted)
	}
	level := fmt.Sprintf("LIV %d", w.Player.Level)
	r.drawTextColored(screen, level, int(xpX), -1, runCream)
	r.drawTextColored(screen, "XP", int(xpRight)-12, -1, runLilac)
	xp := float32(0)
	if w.Player.NextLevelXP > 0 {
		xp = float32(w.Player.XP) / float32(w.Player.NextLevelXP)
	}
	drawRunMeter(screen, xpX, 15, xpRight-xpX, 6, xp, runLilac)
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

func (r *Renderer) drawRelicTray(screen *ebiten.Image, w *game.World, y float32) {
	const badgeSize = float32(20)
	x := float32(8)
	hovered := -1
	hoveredX := float32(0)

	for i, owned := range w.Player.Relics {
		if !owned {
			continue
		}
		box := game.Rect{X: x, Y: y, W: badgeSize, H: badgeSize}
		edge := runLine
		if box.Contains(r.mouseScreenX, r.mouseScreenY) {
			edge = runGold
			hovered, hoveredX = i, x
		}
		drawRunPanel(screen, box, runPlum, edge)
		r.DrawSpriteIcon(screen, relicIconKey(game.RelicKind(i)), x+2, y+2, 1)
		x += badgeSize + 3
	}

	if hovered >= 0 {
		spec := game.GetRelicSpec(game.RelicKind(hovered))
		width := float32(max(textPixelWidth(spec.Name, 1), textPixelWidth(spec.Desc, 1)) + 20)
		tip := game.Rect{X: hoveredX, Y: y + 25, W: width, H: 32}
		drawRunPanel(screen, tip, runPlum, runGold)
		r.drawTextColored(screen, spec.Name, int(tip.X+10), int(tip.Y+1), runGold)
		r.drawTextColored(screen, spec.Desc, int(tip.X+10), int(tip.Y+16), runCream)
	}
}

func (r *Renderer) drawRunNotice(screen *ebiten.Image, label string, centerX, y float32, edge color.RGBA) {
	width := float32(textPixelWidth(label, 1) + 24)
	x := float32(math.Floor(float64(centerX - width/2)))
	drawRunPanel(screen, game.Rect{X: x, Y: y, W: width, H: 18}, runPlum, runLine)
	vector.DrawFilledRect(screen, x+1, y+2, 3, 14, edge, false)
	r.drawRunCentered(screen, label, centerX+1, int(y+1), runCream)
}

func (r *Renderer) drawCenterStatusBanner(screen *ebiten.Image, w *game.World) {
	const centerX = float32(game.ScreenWidth / 2)
	if w.Boss.Active {
		box := game.Rect{X: centerX - 150, Y: 28, W: 300, H: 21}
		drawRunPanel(screen, box, runPlum, runLine)
		r.drawTextColored(screen, "BOSS", int(box.X+8), int(box.Y+3), runCoral)
		name := w.Boss.Name
		if w.BossRushMode {
			name = fmt.Sprintf("%s  %d/6", name, w.BossesDefeated+1)
		}
		r.drawTextColored(screen, name, int(box.X+42), int(box.Y+3), runCream)
		shieldCount := w.Boss.MaxShields
		if shieldCount < 1 {
			shieldCount = 3
		}
		r.drawTextColored(screen, "SCUDI", int(box.X+188), int(box.Y+3), runMuted)
		for i := 0; i < shieldCount; i++ {
			x := box.X + box.W - 8 - float32(shieldCount-i)*20
			y := box.Y + 7
			vector.DrawFilledRect(screen, x, y, 17, 8, runInk, false)
			if i < w.Boss.Shields {
				vector.DrawFilledRect(screen, x+1, y, 15, 7, runCoral, false)
				vector.DrawFilledRect(screen, x+2, y, 13, 1, runGold, false)
			} else {
				vector.DrawFilledRect(screen, x+1, y+3, 15, 1, runLine, false)
			}
		}
		return
	}

	if w.BossRushMode && w.BossScheduleIdx < len(w.BossSchedule) {
		lbl := fmt.Sprintf("BOSS RUSH %d/6  /  TRA %.0fs", w.BossesDefeated+1, w.NextBossRushTimer)
		r.drawRunNotice(screen, lbl, centerX, 28, runGold)
		return
	}
	if w.Police.Heli.Active {
		r.drawRunNotice(screen, "5/5 WANTED  /  ELICOTTERO IN ZONA", centerX, 28, runCoral)
		return
	}
	if w.OverdriveActive {
		rem := max(0, int(math.Ceil(float64(w.NextOverdriveBossTimer))))
		label := fmt.Sprintf("OVERDRIVE ONDA %d  /  BOSS TRA %02d:%02d", w.OverdriveWave, rem/60, rem%60)
		r.drawRunNotice(screen, label, centerX, 28, runGold)
		return
	}

	totalBosses := len(w.BossSchedule)
	if totalBosses == 0 {
		return
	}
	if w.BossScheduleIdx < len(w.BossScheduleTimes) {
		rem := max(0, int(w.BossScheduleTimes[w.BossScheduleIdx]-w.RunTime))
		label := fmt.Sprintf("BOSS %d/%d  /  PROSSIMO %02d:%02d", w.BossesDefeated, totalBosses, rem/60, rem%60)
		accent := runTeal
		if w.RunTime < 6.5 && w.BossesDefeated == 0 {
			label = fmt.Sprintf("BATTI %d BOSS  /  PRIMO BOSS TRA %02d:%02d", totalBosses, rem/60, rem%60)
			accent = runGold
		}
		r.drawRunNotice(screen, label, centerX, 28, accent)
	} else if w.Victory {
		label := fmt.Sprintf("OVERDRIVE  /  BOSS BATTUTI %d", w.BossesDefeated)
		r.drawRunNotice(screen, label, centerX, 28, runGold)
	}
}

func (r *Renderer) drawStreakBadge(screen *ebiten.Image, w *game.World, y float32) {
	if w.StreakCount < 2 {
		return
	}
	label := fmt.Sprintf("STREAK x%d", w.StreakCount)
	accent := runTeal
	if w.StreakCount >= 5 {
		label = fmt.Sprintf("RUSH x%d", w.StreakCount)
		accent = runGold
	}
	width := float32(textPixelWidth(label, 1) + 20)
	box := game.Rect{X: game.ScreenWidth - width - 8, Y: y, W: width, H: 20}
	drawRunPanel(screen, box, runPlum, runLine)
	r.drawRunCentered(screen, label, box.X+box.W/2, int(box.Y), accent)
	drawRunMeter(screen, box.X+9, box.Y+14, box.W-18, 4, w.StreakTimer/game.StreakWindowDuration, accent)
}

func (r *Renderer) drawChallengeToast(screen *ebiten.Image, y float32) bool {
	if r.toastTimer <= 0 || r.activeChallengeToast == "" {
		return false
	}
	label := "SFIDA SBLOCCATA"
	width := max(textPixelWidth(label, 1), textPixelWidth(r.activeChallengeToast, 1))
	box := game.Rect{X: 8, Y: y, W: float32(width + 40), H: 33}
	drawRunPanel(screen, box, runPlum, runLine)
	vector.DrawFilledRect(screen, box.X+1, box.Y+1, 3, box.H-2, runGold, false)
	r.DrawSpriteIcon(screen, assets.SpriteIconStarWanted, box.X+11, box.Y+10, 1)
	r.drawTextColored(screen, label, int(box.X+30), int(y+1), runGold)
	r.drawTextColored(screen, r.activeChallengeToast, int(box.X+30), int(y+15), runCream)
	return true
}

func (r *Renderer) drawRelicToast(screen *ebiten.Image, w *game.World, y float32) {
	if r.relicToastTimer <= 0 || r.relicToastName == "" || w.PendingLevelUp || w.ShopMenu.Active || w.SlotModal.Active {
		return
	}
	label := "RELIQUIA / " + r.relicToastName
	width := max(textPixelWidth(label, 1), textPixelWidth(r.relicToastDesc, 1))
	box := game.Rect{X: 8, Y: y, W: float32(width + 42), H: 33}
	drawRunPanel(screen, box, runPlum, runLine)
	vector.DrawFilledRect(screen, box.X+1, box.Y+1, 3, box.H-2, runTeal, false)
	r.DrawSpriteIcon(screen, relicIconKey(r.relicToastKind), box.X+10, box.Y+8, 1)
	r.drawTextColored(screen, label, int(box.X+31), int(y+1), runTeal)
	r.drawTextColored(screen, r.relicToastDesc, int(box.X+31), int(y+15), runCream)
}

func (r *Renderer) drawTutorialSandboxBar(screen *ebiten.Image, w *game.World) {
	box := game.Rect{X: 8, Y: 28, W: 624, H: 36}
	drawRunPanel(screen, box, runPlum, runLine)
	vector.DrawFilledRect(screen, box.X+1, box.Y+1, 3, box.H-2, runTeal, false)

	var tip string
	switch w.TutorialStep {
	case 0:
		tip = "1/4  BATTI 4 BOSS IN PARTITA  /  WASD + CLICK SX PER LANCIARE"
	case 1:
		tip = "2/4  ENTRA NEL CERCHIO, COMPLETA MINIGIOCO E COMBO BUS"
	case 2:
		tip = "3/4  AVVICINATI E CLICCA SU SLOT 777 O DISTRIBUTORE"
	default:
		tip = "4/4  IN PARTITA BATTI 4 BOSS PER VINCERE"
	}
	r.drawTextColored(screen, tip, 19, 28, runCream)

	checkpoints := [...]struct {
		label string
		done  bool
	}{
		{"MOVIMENTO", w.TutMoved},
		{"FERMATA", w.TutThrewStop},
		{"COMBO BUS", w.TutBoardedBus},
		{"SHOP/SLOT", w.TutUsedSlotOrShop},
	}
	x := float32(20)
	for _, checkpoint := range checkpoints {
		tint := runMuted
		if checkpoint.done {
			tint = runTeal
			vector.DrawFilledRect(screen, x, 51, 1, 2, runTeal, false)
			vector.DrawFilledRect(screen, x+1, 53, 1, 1, runTeal, false)
			vector.DrawFilledRect(screen, x+2, 52, 1, 1, runTeal, false)
			vector.DrawFilledRect(screen, x+3, 50, 1, 2, runTeal, false)
			vector.DrawFilledRect(screen, x+4, 49, 1, 1, runTeal, false)
		} else {
			vector.StrokeRect(screen, x+0.5, 49.5, 4, 4, 1, runLine, false)
		}
		r.drawTextColored(screen, checkpoint.label, int(x+10), 44, tint)
		x += 91
	}

	r.drawRunButton(screen, TutorialGuideRect(), "GUIDA", runTeal, true, false)
	r.drawRunButton(screen, TutorialExitRect(), "MENU", runCoral, true, false)
}

func (r *Renderer) drawRunButton(screen *ebiten.Image, box game.Rect, label string, accent color.RGBA, enabled, primary bool) {
	face, edge, tint := runFace, runLine, runCream
	hovered := box.Contains(r.mouseScreenX, r.mouseScreenY)
	if !enabled {
		face, tint = runPlum, runMuted
	} else if primary || hovered {
		face, edge, tint = accent, accent, runInk
	}
	drawRunPanel(screen, box, face, edge)
	x := int(box.X + (box.W-float32(textPixelWidth(label, 1)))/2)
	y := int(box.Y + (box.H-16)/2)
	if tint == runInk {
		vector.DrawFilledRect(screen, box.X+1, box.Y+box.H-3, box.W-2, 2, Premul(28, 19, 34, 48), false)
		if hovered {
			vector.DrawFilledRect(screen, box.X+1, box.Y+1, box.W-2, 1, PremulColor(runCream, 140), false)
		}
		r.drawScaledText(screen, label, x, y, 1, tint, color.Transparent)
		return
	}
	r.drawTextColored(screen, label, x, y, tint)
}

func (r *Renderer) drawCriticalOffscreenArrows(screen *ebiten.Image, w *game.World) {
	safeTop := w.StopSafeTopY()
	if w.Boss.Active {
		r.drawOffscreenEdgeArrow(screen, w.Boss.X, w.Boss.Y, safeTop, "BOSS", runCoral)
	}
	if w.Player.HP <= 2 && len(w.HealPickups) > 0 {
		bestIdx := -1
		bestD2 := float32(1e12)
		for i := range w.HealPickups {
			hx, hy := float32(w.HealPickups[i].X), float32(w.HealPickups[i].Y)
			d2 := (hx-w.Player.X)*(hx-w.Player.X) + (hy-w.Player.Y)*(hy-w.Player.Y)
			if d2 < bestD2 {
				bestD2 = d2
				bestIdx = i
			}
		}
		if bestIdx >= 0 {
			hp := &w.HealPickups[bestIdx]
			r.drawOffscreenEdgeArrow(screen, float32(hp.X), float32(hp.Y), safeTop, "CURA", runTeal)
		}
	}
}

func (r *Renderer) drawOffscreenEdgeArrow(screen *ebiten.Image, wx, wy, safeTop float32, label string, accent color.RGBA) {
	if wx >= r.lastCamX+12 && wx <= r.lastCamX+game.ScreenWidth-12 &&
		wy >= r.lastCamY+26 && wy <= r.lastCamY+game.ScreenHeight-48 {
		return
	}

	sx := wx - r.lastCamX
	sy := wy - r.lastCamY
	cx := float32(game.ScreenWidth * 0.5)
	cy := float32(game.ScreenHeight * 0.5)
	dx := sx - cx
	dy := sy - cy
	if dx == 0 && dy == 0 {
		return
	}

	const marginX = float32(18)
	minY := safeTop + 11
	const maxY = float32(game.ScreenHeight - 58)

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

	mx := clampF(cx+dx*t, marginX, game.ScreenWidth-marginX)
	my := clampF(cy+dy*t, minY, maxY)
	ang := math.Atan2(float64(dy), float64(dx))

	pulse := float32(0.5 + 0.5*math.Sin(float64(r.tick*10.0)))
	scale := float32(1.0 + 0.16*pulse)

	tipX := mx + float32(math.Cos(ang))*9*scale
	tipY := my + float32(math.Sin(ang))*9*scale
	leftX := mx + float32(math.Cos(ang+2.45))*7*scale
	leftY := my + float32(math.Sin(ang+2.45))*7*scale
	rightX := mx + float32(math.Cos(ang-2.45))*7*scale
	rightY := my + float32(math.Sin(ang-2.45))*7*scale

	r.drawFilledTriangle(screen, tipX, tipY, leftX, leftY, rightX, rightY, accent)
	vector.StrokeLine(screen, leftX, leftY, tipX, tipY, 1.4, runInk, true)
	vector.StrokeLine(screen, rightX, rightY, tipX, tipY, 1.4, runInk, true)
	vector.StrokeLine(screen, leftX, leftY, rightX, rightY, 1.4, runInk, true)

	bw := float32(textPixelWidth(label, 1.0) + 10)
	px := clampF(mx-float32(math.Cos(ang))*16-bw*0.5, 4, game.ScreenWidth-bw-4)
	py := clampF(my-float32(math.Sin(ang))*14-7, safeTop, game.ScreenHeight-62)
	box := game.Rect{X: float32(math.Round(float64(px))), Y: float32(math.Round(float64(py))), W: bw, H: 14}
	drawRunPanel(screen, box, runPlum, accent)
	r.drawTextColored(screen, label, int(box.X+5), int(box.Y-1), runCream)
}

func (r *Renderer) drawFilledTriangle(screen *ebiten.Image, x1, y1, x2, y2, x3, y3 float32, clr color.RGBA) {
	if r.whitePixel == nil {
		return
	}
	rf := float32(clr.R) / 255.0
	gf := float32(clr.G) / 255.0
	bf := float32(clr.B) / 255.0
	af := float32(clr.A) / 255.0
	verts := [3]ebiten.Vertex{
		{DstX: x1, DstY: y1, SrcX: 0, SrcY: 0, ColorR: rf, ColorG: gf, ColorB: bf, ColorA: af},
		{DstX: x2, DstY: y2, SrcX: 0, SrcY: 0, ColorR: rf, ColorG: gf, ColorB: bf, ColorA: af},
		{DstX: x3, DstY: y3, SrcX: 0, SrcY: 0, ColorR: rf, ColorG: gf, ColorB: bf, ColorA: af},
	}
	screen.DrawTriangles(verts[:], []uint16{0, 1, 2}, r.whitePixel, nil)
}

func (r *Renderer) drawOffscreenIndicators(screen *ebiten.Image, w *game.World, camX, camY float32) {
	// Active BusStops offscreen pointers (Boss & low-HP Heal arrows render in DrawHUDAndOverlays)
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

	bw := float32(textPixelWidth(label, 1.0) + 8)
	bx := clampF(mx-bw*0.5, 4, game.ScreenWidth-bw-4)
	by := clampF(my-7, 30, game.ScreenHeight-64)
	drawPillBadge(screen, bx, by, bw, 13, colHudBg, col)
	r.drawTextColored(screen, label, int(bx+4), int(by-1), col)
}

func (r *Renderer) drawBottomHotbarDock(screen *ebiten.Image, w *game.World) {
	first := HotbarSlotRect(0)
	last := HotbarSlotRect(int(game.ArchCount) - 1)
	dock := game.Rect{X: 94, Y: first.Y - 4, W: 452, H: first.H + 8}
	drawRunPanel(screen, dock, runPlum, runLine)
	vector.DrawFilledRect(screen, dock.X+7, dock.Y+dock.H-2, dock.W-14, 1, runInk, false)

	r.DrawSpriteIcon(screen, assets.SpriteIconMouseWheel, first.X-26, first.Y+2, 1)
	r.drawTextColored(screen, "1-4", int(first.X-27), int(first.Y+19), runMuted)

	routes := [...]struct {
		label, stop, rider string
		accent             color.RGBA
	}{
		{"EMO", assets.SpriteStopEmo, assets.SpriteGirlEmo0, runLilac},
		{"UNI", assets.SpriteStopStudent, assets.SpriteGirlStudent0, runSky},
		{"KAWAII", assets.SpriteStopKawaii, assets.SpriteGirlKawaii0, runPink},
		{"GYARU", assets.SpriteStopGyaru, assets.SpriteGirlGyaru0, runTeal},
	}

	for i, route := range routes {
		box := HotbarSlotRect(i)
		x, y := box.X, box.Y
		accent := route.accent
		selected := w.Player.SelectedStopArch == game.Archetype(i)
		face, edge := runPlum, runLine
		if selected {
			face, edge = runFace, accent
		} else if box.Contains(r.mouseScreenX, r.mouseScreenY) {
			face, edge = runFace, runMuted
		}
		drawRunPanel(screen, box, face, edge)
		labelTint := runCream
		if selected {
			vector.DrawFilledRect(screen, x+1, y+1, box.W-2, box.H-2, PremulColor(accent, 38), false)
			labelTint = accent
		}
		vector.DrawFilledRect(screen, x+2, y+2, 14, 12, accent, false)
		r.drawScaledText(screen, fmt.Sprint(i+1), int(x+6), int(y), 1, runInk, color.Transparent)
		r.drawTextColored(screen, route.label, int(x+21), int(y), labelTint)

		vector.DrawFilledRect(screen, x+12, y+27, 47, 1, accent, false)
		vector.DrawFilledRect(screen, x+33, y+25, 5, 5, accent, false)
		vector.DrawFilledRect(screen, x+34, y+26, 3, 3, face, false)
		if selected {
			vector.DrawFilledRect(screen, x+26, y+35, 20, 2, accent, false)
		}
		r.drawHotbarSprite(screen, route.stop, x+8, y+16, 20)
		r.drawHotbarSprite(screen, route.rider, x+43, y+16, 20)
	}

	active, capacity := w.ActiveStopCount(), w.Player.EffectiveMaxStops()
	statusX := last.X + last.W + 10
	label, accent := "LANCIA", runTeal
	cooldown := w.Player.EffectiveStopCooldown()
	ready := float32(1)
	if cooldown > 0 && w.Player.StopCooldownTimer > 0 {
		ready -= clampF(w.Player.StopCooldownTimer/cooldown, 0, 1)
		label, accent = "RICARICA", runGold
	}
	if active >= capacity {
		label, accent = "FERMATE MAX", runCoral
	}

	r.DrawSpriteIcon(screen, assets.SpriteIconMouseLMB, statusX, first.Y+1, 1)
	r.drawTextColored(screen, label, int(statusX+18), int(first.Y), accent)
	usage := fmt.Sprintf("%d/%d", active, capacity)
	r.drawTextColored(screen, usage, int(statusX+2), int(first.Y+19), runCream)
	meterX := statusX + float32(textPixelWidth(usage, 1)) + 9
	drawRunMeter(screen, meterX, first.Y+24, dock.X+dock.W-meterX-12, 7, ready, accent)

	if w.Player.GroundedTimer > 0 {
		label := fmt.Sprintf("PUNIZIONE  /  %.0fs", w.Player.GroundedTimer)
		r.drawRunNotice(screen, label, game.ScreenWidth/2, first.Y-24, runCoral)
	}
}

func (r *Renderer) drawHotbarSprite(screen *ebiten.Image, key string, x, y, size float32) {
	if r.sprites == nil {
		return
	}
	sprite := r.sprites.Get(key)
	if sprite == nil {
		return
	}
	op := ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
	op.GeoM.Scale(float64(size)/float64(sprite.Bounds().Dx()), float64(size)/float64(sprite.Bounds().Dy()))
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(sprite, &op)
}

func upgradeTheme(kind game.UpgradeKind) (iconKey string, accent color.RGBA) {
	switch kind {
	case game.UpgradeCooldown:
		return assets.SpriteIconUpgCooldown, runTeal
	case game.UpgradeMultiHook:
		return assets.SpriteIconUpgMultiHook, runPink
	case game.UpgradeMaxStops:
		return assets.SpriteIconUpgMaxStops, runGold
	case game.UpgradeRange:
		return assets.SpriteIconUpgRange, runSky
	case game.UpgradeChronos:
		return assets.SpriteIconUpgChronos, runLilac
	case game.UpgradeVitality:
		return assets.SpriteIconUpgVitality, runCoral
	default:
		return assets.SpriteIconUpgCooldown, runTeal
	}
}

func drawRunScrim(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, game.ScreenWidth, game.ScreenHeight, Premul(23, 15, 30, 194), false)
}

func (r *Renderer) drawLevelUpModal(screen *ebiten.Image, w *game.World) {
	drawRunScrim(screen)
	marquee := game.Rect{X: 44, Y: 33, W: 550, H: 40}
	drawRunPanel(screen, marquee, runPlum, runLine)
	vector.DrawFilledRect(screen, marquee.X+1, marquee.Y+1, 4, marquee.H-2, runTeal, false)
	r.drawScaledText(screen, "SCEGLI UN UPGRADE", 60, 37, 2, runCream, runInk)

	level := fmt.Sprintf("LIVELLO %02d", w.Player.Level)
	r.drawTextColored(screen, level, 579-textPixelWidth(level, 1), 45, runTeal)
	r.drawRunCentered(screen, "CLICK SX O TASTI 1-3", game.ScreenWidth/2, 75, runCream)

	for i, card := range w.UpgradeChoices {
		ticket := LevelUpCardRect(i)
		hovered := ticket.Contains(r.mouseScreenX, r.mouseScreenY)
		icon, accent := upgradeTheme(card.Kind)
		face, edge := runPlum, runLine
		if hovered {
			face, edge = runFace, accent
		}
		drawRunPanel(screen, ticket, face, edge)
		x, y := ticket.X, ticket.Y
		center := x + ticket.W/2
		vector.DrawFilledRect(screen, x+1, y+1, ticket.W-2, 3, accent, false)

		art := game.Rect{X: x + 12, Y: y + 13, W: ticket.W - 24, H: 61}
		drawRunPanel(screen, art, runInk, runLine)
		rail := PremulColor(accent, 96)
		vector.DrawFilledRect(screen, art.X+11, art.Y+40, art.W-22, 1, rail, false)
		for _, sx := range [...]float32{art.X + 20, art.X + 40, art.X + art.W - 44, art.X + art.W - 24} {
			vector.DrawFilledRect(screen, sx, art.Y+38, 5, 5, rail, false)
			vector.DrawFilledRect(screen, sx+1, art.Y+39, 3, 3, runInk, false)
		}
		drawRunPanel(screen, game.Rect{X: center - 24, Y: art.Y + 6, W: 48, H: 48}, runFace, accent)
		r.DrawSpriteIcon(screen, icon, center-20, art.Y+10, 2)

		r.drawRunCentered(screen, card.Title, center, int(y+80), runCream)
		drawRunWrappedText(r, screen, card.Desc, center, int(y+101), 25, runMuted)
		for dash := x + 9; dash < x+ticket.W-9; dash += 6 {
			vector.DrawFilledRect(screen, dash, y+135, 3, 1, runLine, false)
		}
		action := game.Rect{X: x + 12, Y: y + 144, W: ticket.W - 24, H: 22}
		r.drawRunButton(screen, action, fmt.Sprintf("[%d] SCEGLI", i+1), accent, true, hovered)
	}
}

func (r *Renderer) drawRunCounter(screen *ebiten.Image, box game.Rect, sprite, label string, tint color.RGBA) {
	drawRunPanel(screen, box, runInk, runLine)
	r.DrawSpriteIcon(screen, sprite, box.X+5, box.Y+(box.H-12)/2, 1)
	r.drawTextColored(screen, label, int(box.X+20), int(box.Y+(box.H-16)/2), tint)
}

func (r *Renderer) drawShopModal(screen *ebiten.Image, w *game.World) {
	drawRunScrim(screen)
	box := game.Rect{X: 124, Y: 38, W: 392, H: 270}
	drawRunPanel(screen, box, runPlum, runTeal)
	vector.DrawFilledRect(screen, box.X+1, box.Y+1, box.W-2, 2, runTeal, false)

	r.drawTextColored(screen, "DISTRIBUTORE", 144, 45, runCream)
	r.drawScaledText(screen, "NEON", 224, 39, 2, runTeal, runInk)
	r.drawRunCounter(screen, game.Rect{X: 282, Y: 44, W: 64, H: 18}, assets.SpriteIconHeartFull, fmt.Sprintf("%d/%d", w.Player.HP, w.Player.MaxHP), runCream)
	r.drawRunCounter(screen, game.Rect{X: 352, Y: 44, W: 98, H: 18}, assets.SpriteIconCoin, fmt.Sprintf("$%d", w.Player.Cash), runGold)
	r.drawRunCounter(screen, game.Rect{X: 456, Y: 44, W: 40, H: 18}, assets.SpriteIconStarWanted, fmt.Sprint(w.Police.Stars), runCoral)

	message := game.Rect{X: 144, Y: 69, W: 352, H: 29}
	drawRunPanel(screen, message, runInk, runLine)
	msg := w.ShopMenu.Message
	if textPixelWidth(msg, 1) <= 332 {
		r.drawRunCentered(screen, msg, 320, 76, runCream)
	} else if title, desc, ok := strings.Cut(msg, ": "); ok && textPixelWidth(desc, 1) <= 332 {
		r.drawRunCentered(screen, title, 320, 69, runTeal)
		r.drawRunCentered(screen, desc, 320, 82, runCream)
	} else {
		drawRunWrappedText(r, screen, msg, 320, 69, 55, runCream)
	}

	relicEffect := fmt.Sprintf("%d/%d raccolte", w.Player.RelicCountOwned(), game.RelicCount)
	if w.Player.RelicCountOwned() == int(game.RelicCount) {
		relicEffect = "+1 Max HP, +2 HP"
	}
	stock := [...]struct {
		sprite, label, effect string
		price                 int
		accent                color.RGBA
	}{
		{assets.SpriteIconShopEspresso, "Caffè Espresso", "+2 HP e Chronos", w.CurrentHealCost(), runCoral},
		{assets.SpriteIconShopBribe, "Tangente Polizia", "-2 stelle Wanted", w.CurrentBribeCost(), runGold},
		{assets.SpriteIconShopRelic, "Reliquia Casuale", relicEffect, 120, runTeal},
		{assets.SpriteIconShopCrate, "Cassa Upgrade", "Scegli fra 3 carte", 140, runTeal},
	}
	for i, item := range stock {
		row := ShopChoiceRect(i)
		x, y := row.X, row.Y
		hasCash := w.Player.Cash >= item.price
		available := hasCash && (i != 1 || w.Police.Stars > 0)
		hovered := row.Contains(r.mouseScreenX, r.mouseScreenY)
		face, edge, title := runPlum, runLine, runMuted
		if available {
			face, title = runFace, runCream
		}
		if available && hovered {
			edge = item.accent
		}
		drawRunPanel(screen, row, face, edge)
		vector.DrawFilledRect(screen, x+4, y+4, 14, 22, runInk, false)
		r.drawTextColored(screen, fmt.Sprint(i+1), int(x+8), int(y+7), item.accent)
		if available {
			r.DrawSpriteIcon(screen, item.sprite, x+25, y+7, 1)
		} else {
			r.drawSpriteIconTinted(screen, item.sprite, x+25, y+7, 1, Premul(228, 210, 228, 153))
		}
		r.drawTextColored(screen, item.label, int(x+48), int(y+7), title)
		r.drawTextColored(screen, item.effect, int(x+162), int(y+7), runMuted)

		priceBox := game.Rect{X: x + row.W - 58, Y: y + 5, W: 52, H: 20}
		priceAccent := runMuted
		if available {
			priceAccent = runTeal
		} else if !hasCash {
			priceAccent = runCoral
		}
		drawRunPanel(screen, priceBox, runInk, runLine)
		r.drawRunCentered(screen, fmt.Sprintf("$%d", item.price), priceBox.X+priceBox.W/2, int(y+7), priceAccent)
	}

	r.drawRunCentered(screen, "CLICK SX O TASTI 1-4 PER ACQUISTARE", 320, 252, runMuted)
	r.drawRunButton(screen, ShopCloseRect(), "[ESC] CHIUDI", runCoral, true, false)
}

func slotSymbolMeta(sym int) (iconKey string, label string, clr color.RGBA) {
	switch sym % game.SymCount {
	case game.SymCherry:
		return assets.SpriteIconSlotCherry, "CILIEGIA", runPink
	case game.SymCash:
		return assets.SpriteIconSlotCash, "CASH", runTeal
	case game.SymXP:
		return assets.SpriteIconSlotXP, "XP+", runSky
	case game.SymHeart:
		return assets.SpriteIconSlotHeart, "CUORE", runPink
	case game.SymGoldBus:
		return assets.SpriteIconSlot777, "777", runGold
	case game.SymSiren:
		return assets.SpriteIconSlotSiren, "SIRENA", runCoral
	default:
		return assets.SpriteIconSlot777, "777", runGold
	}
}

func (r *Renderer) drawSlotModal(screen *ebiten.Image, w *game.World) {
	const cabinetTitle = "RIZZ-SLOTS 777"
	drawRunScrim(screen)
	cabinet := game.Rect{X: 132, Y: 31, W: 376, H: 270}
	drawRunPanel(screen, cabinet, runPlum, runGold)
	vector.DrawFilledRect(screen, cabinet.X+1, cabinet.Y+1, cabinet.W-2, 2, runGold, false)
	r.drawScaledText(screen, cabinetTitle, 151, 34, 2, runCream, runInk)
	r.drawRunCounter(screen, game.Rect{X: 394, Y: 40, W: 95, H: 19}, assets.SpriteIconCoin, fmt.Sprintf("$%d", w.Player.Cash), runGold)

	lockTimes := [...]float32{0.40, 0.75, 1.10}
	for i, symbol := range w.SlotModal.Reels {
		icon, label, tint := slotSymbolMeta(symbol)
		reel := game.Rect{X: 153 + float32(i)*114, Y: 66, W: 106, H: 76}
		rolling := w.SlotModal.Spinning && w.SlotModal.SpinTimer < lockTimes[i]
		edge := runLine
		if rolling || w.SlotModal.ResultKind == 2 {
			edge = runGold
		}
		drawRunPanel(screen, reel, runInk, edge)

		// The light insert sits inside a recessed cabinet opening.
		vector.DrawFilledRect(screen, reel.X+4, reel.Y+4, reel.W-8, 46, runFace, false)
		vector.DrawFilledRect(screen, reel.X+4, reel.Y+4, reel.W-8, 1, runLine, false)
		vector.DrawFilledRect(screen, reel.X+4, reel.Y+49, reel.W-8, 1, tint, false)

		jitter := float32(0)
		if rolling {
			jitter = float32(math.Round(math.Sin(float64(r.tick*40+float32(i)*2.1)) * 2))
		}
		r.DrawSpriteIcon(screen, icon, reel.X+(reel.W-40)/2, reel.Y+7+jitter, 2)
		r.drawRunCentered(screen, label, reel.X+reel.W/2, int(reel.Y+54), runCream)
	}

	payouts := [...]struct {
		symbol int
		text   string
	}{
		{game.SymGoldBus, "$250 + XP"},
		{game.SymCash, "$140"},
		{game.SymHeart, "+3 HP"},
		{game.SymCherry, "$90"},
		{game.SymXP, "+110 XP"},
		{game.SymSiren, "$110 +1*"},
	}
	legend := game.Rect{X: 151, Y: 148, W: 338, H: 43}
	drawRunPanel(screen, legend, runInk, runLine)
	vector.DrawFilledRect(screen, legend.X+36, legend.Y+6, 1, legend.H-12, runLine, false)
	r.drawTextColored(screen, "TRIS", int(legend.X+6), int(legend.Y+14), runGold)
	for i, payout := range payouts {
		x, y := legend.X+43+float32(i%3)*97, legend.Y+2+float32(i/3)*19
		icon, _, _ := slotSymbolMeta(payout.symbol)
		r.DrawSpriteIcon(screen, icon, x, y, 1)
		r.drawTextColored(screen, payout.text, int(x+23), int(y+2), runMuted)
	}

	accent := runTeal
	switch w.SlotModal.ResultKind {
	case 2:
		accent = runGold
	case 3:
		accent = runCoral
	}
	result := game.Rect{X: 151, Y: 197, W: 338, H: 50}
	drawRunPanel(screen, result, runFace, runLine)
	vector.DrawFilledRect(screen, result.X+1, result.Y+1, 3, result.H-2, accent, false)
	title, description := w.SlotModal.ResultTitle, w.SlotModal.ResultDesc
	if title == cabinetTitle {
		title, description = "COMBINA I SIMBOLI", "VINCI CON UNA COPPIA O UN TRIS"
	}
	r.drawRunCentered(screen, title, result.X+result.W/2, int(result.Y+5), accent)
	drawRunWrappedText(r, screen, description, result.X+result.W/2, int(result.Y+21), 51, runCream)

	cost := w.CurrentSlotCost()
	label := fmt.Sprintf("[SPAZIO] GIRA  $%d", cost)
	spinning := w.SlotModal.Spinning
	canSpin := !spinning && w.Player.Cash >= cost
	if spinning {
		label = "GIRANDO..."
	}
	r.drawRunButton(screen, SlotSpinRect(), label, runGold, canSpin, true)
	r.drawRunButton(screen, SlotCloseRect(), "[ESC] ESCI", runCoral, !spinning, false)
}

func drawRunWrappedText(r *Renderer, screen *ebiten.Image, text string, centerX float32, y, columns int, tint color.Color) {
	line := ""
	limit := columns * textPixelWidth(" ", 1)
	for _, word := range strings.Fields(text) {
		if line == "" {
			line = word
			continue
		}
		next := line + " " + word
		if textPixelWidth(next, 1) > limit {
			r.drawRunCentered(screen, line, centerX, y, tint)
			y += 13
			line = word
		} else {
			line = next
		}
	}
	if line != "" {
		r.drawRunCentered(screen, line, centerX, y, tint)
	}
}
