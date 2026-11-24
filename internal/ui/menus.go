package ui

import (
	"fmt"
	"image/color"
	"math"

	"girlbusAG/assets"
	"girlbusAG/internal/game"
	"girlbusAG/internal/meta"
	"girlbusAG/internal/render"
	"github.com/hajimehoshi/ebiten/v2"
)

func (m *Manager) drawPaused(dst *ebiten.Image, r *render.Renderer, w *game.World, mx, my float32) {
	screenShade(dst, 164)
	menuBoard(dst, 32, 32, 245, 294, menuCoral)
	menuBoard(dst, 290, 32, 318, 294, menuTeal)
	menuHeading(dst, r, "Pausa", 49, 49)
	restart := "RIAVVIA RUN"
	if w != nil && w.TutorialMode {
		restart = "RIAVVIA TUTORIAL"
	}
	labels := [...]string{"RIPRENDI", restart, "GUIDA & TUTORIAL", "OPZIONI", "MENU PRINCIPALE"}
	for i, label := range labels {
		rect := pauseButtonRect(i)
		style := m.buttonStyle(i, rect, mx, my)
		style.primary = i == 0
		drawMenuButton(dst, r, rect, label, style)
	}
	r.DrawTextColored(dst, "ESC  Riprendi", 49, 302, menuMuted)
	menuHeading(dst, r, "Stato run", 305, 49)
	if w == nil {
		return
	}
	if w.TutorialMode {
		menuChip(dst, r, "CAMPO DI PROVA", 493, 55, 98, menuTeal)
	}
	mins, secs := int(w.RunTime)/60, int(w.RunTime)%60
	for i, stat := range [...]struct{ label, value string }{
		{"TEMPO", fmt.Sprintf("%02d:%02d", mins, secs)},
		{"PUNTEGGIO", fmt.Sprint(w.Score)},
		{"TIER", fmt.Sprint(w.DifficultyTier)},
	} {
		x := 306 + float32(i)*96
		fill(dst, x, 88, 91, 46, menuSurface)
		fill(dst, x, 88, 20, 1, menuRule)
		centerText(dst, r, stat.label, x, 92, 91, menuMuted)
		centerText(dst, r, fitText(stat.value, 84), x, 111, 91, menuCream)
	}
	r.DrawTextColored(dst, fmt.Sprintf("Passeggere %d    Boss %d    Ricarica %.1fs", w.GirlsBoarded, w.BossesDefeated, w.Player.EffectiveStopCooldown()), 307, 138, menuTeal)
	fill(dst, 305, 163, 288, 1, menuRule)
	r.DrawTextColored(dst, "RELIQUIE ATTIVE", 307, 166, menuGold)
	count := 0
	for rk := game.RelicKind(0); rk < game.RelicCount; rk++ {
		if !w.Player.Relics[rk] {
			continue
		}
		spec := game.GetRelicSpec(rk)
		y := 183 + count*22
		fill(dst, 307, float32(y+6), 3, 3, menuTeal)
		r.DrawTextColored(dst, spec.Name, 318, y-2, menuCream)
		r.DrawTextColored(dst, fitText(spec.Desc, 282), 318, y+8, menuMuted)
		count++
	}
	if count == 0 {
		r.DrawTextColored(dst, "Acquista una reliquia al distributore.", 307, 190, menuMuted)
	}
}

func (m *Manager) drawOptions(dst *ebiten.Image, r *render.Renderer, save *meta.SaveData, mx, my float32) {
	screenShade(dst, 186)
	menuBoard(dst, 90, 16, 460, 328, menuTeal)
	menuHeading(dst, r, "Opzioni", 107, 22)
	fill(dst, 109, 53, 423, 1, menuRule)

	settings := save.Settings
	filters := [...]string{"Pixel Crisp", "Warm CRT", "VHS Noir"}
	rows := [...]string{"Filtro video", "Volume generale", "Musica", "Effetti sonori", "Schermo intero", "Vibrazione schermo", "Mostra FPS", "Chronos (Shift / Q)", "Controllo frecce"}
	accents := [...]color.RGBA{guideSky, menuGold, menuGold, menuGold, guideSky, guideSky, guideSky, menuTeal, menuTeal}
	for i, label := range rows {
		rect := optionsRowRect(i)
		style := m.buttonStyle(i, rect, mx, my)
		accent := accents[i]
		bg := menuSurface
		if style.focused || style.hovered {
			bg = menuRaised
		}
		fill(dst, rect.X, rect.Y, rect.W, rect.H, bg)
		fill(dst, rect.X+214, rect.Y, rect.W-214, rect.H, render.PremulColor(accent, 14))
		if style.focused || style.hovered {
			fill(dst, rect.X, rect.Y, 3, rect.H, accent)
		}
		r.DrawTextColored(dst, label, int(rect.X+12), int(rect.Y+4), menuCream)

		switch i {
		case 0, 7, 8:
			value := filters[settings.FilterMode]
			if i == 7 {
				value = "Tieni premuto"
				if settings.ChronosToggle {
					value = "Attiva / disattiva"
				}
			} else if i == 8 {
				value = "Pad nel riquadro"
				if settings.DirInputMode == meta.DirModeAroundStop {
					value = "Attorno alla fermata"
				}
			}
			menuChip(dst, r, value, rect.X+222, rect.Y+4, 172, accent)
			menuChevron(dst, rect.X+405, rect.Y+9, menuMuted)
		case 1, 2, 3:
			value := *optionVolume(&settings, i)
			barX, barW := optionsSlider(rect)
			barY := rect.Y + 10
			fill(dst, barX, barY, barW, 4, menuInk)
			width := float32(math.Round(float64(barW) * value))
			if width > 0 {
				fill(dst, barX, barY, width, 4, accent)
			}
			fill(dst, barX+width-2, barY-3, 4, 10, menuCream)
			pct := fmt.Sprintf("%d%%", int(value*100+0.5))
			r.DrawTextColored(dst, pct, int(rect.X+410)-render.TextPixelWidth(pct, 1), int(rect.Y+4), menuGold)
		case 4, 5, 6:
			flags := [...]bool{settings.Fullscreen, settings.ScreenShake, settings.ShowFPS}
			drawMenuSwitch(dst, r, rect.X+332, rect.Y+4, flags[i-4])
		}
	}
	r.DrawTextColored(dst, "A / D regola   Mouse trascina", 108, 308, menuMuted)
	back := optionsRowRect(9)
	style := m.buttonStyle(9, back, mx, my)
	style.accent = menuTeal
	drawMenuButton(dst, r, back, "SALVA E TORNA", style)
}

func drawMenuSwitch(dst *ebiten.Image, r *render.Renderer, x, y float32, on bool) {
	fill(dst, x, y+2, 30, 12, menuInk)
	col := menuMuted
	label := "OFF"
	thumbX := x + 2
	if on {
		col, label, thumbX = menuTeal, "ON", x+17
	}
	fill(dst, thumbX, y+4, 11, 8, col)
	r.DrawTextColored(dst, label, int(x+38), int(y), col)
}

func (m *Manager) drawControls(dst *ebiten.Image, r *render.Renderer, mx, my float32) {
	screenShade(dst, 186)
	menuBoard(dst, 25, 18, 590, 326, menuTeal)
	menuHeading(dst, r, "Comandi", 38, 24)
	r.DrawTextColored(dst, "MOUSE E TASTIERA", 484, 35, menuMuted)
	fill(dst, 38, 58, 564, 1, menuRule)

	items := [...]struct{ key, desc string }{
		{"WASD", "Muovi Alfredo"},
		{"SPAZIO / DX", "Dash e Dash-Parry"},
		{"SHIFT / Q", "Attiva Chronos"},
		{"CLICK SX / F", "Mira e lancia la fermata"},
		{"ROTELLA / 1-4", "Cambia tipo di fermata"},
		{"CLICK SX", "Minigiochi e combo"},
		{"R / E", "Richiama / interagisci"},
		{"ESC / P", "Pausa / chiudi il menu"},
	}
	for i, item := range items {
		x := 38 + float32(i/4)*286
		y := 72 + float32(i%4)*39
		fill(dst, x, y, 278, 33, menuSurface)
		r.DrawKeycap(dst, item.key, x+4, y+5, 88, 23, menuPanel, menuRule, false)
		r.DrawTextColored(dst, item.desc, int(x+99), int(y+8), menuCream)
	}
	fill(dst, 38, 232, 564, 63, menuSurface)
	fill(dst, 38, 232, 3, 63, menuGold)
	r.DrawTextColored(dst, "FRECCE / IJKL", 49, 238, menuGold)
	r.DrawTextColored(dst, "Direzioni nei minigiochi e nelle combo.", 49, 253, menuCream)
	r.DrawTextColored(dst, "Il gamepad supporta mira twin-stick, croce direzionale e tasti frontali.", 49, 270, menuMuted)
	back := controlsBackRect()
	drawMenuButton(dst, r, back, "INDIETRO [ESC]", m.buttonStyle(0, back, mx, my))
}

func (m *Manager) drawEndRun(dst *ebiten.Image, r *render.Renderer, save *meta.SaveData, victory bool, mx, my float32) {
	screenShade(dst, 182)
	accent, title := menuCoral, "Corsa terminata"
	if victory {
		accent, title = menuGold, "Corsa completata"
		if m.LastWorldStats.BossRush {
			title = "Boss Rush completata"
		}
	}
	menuBoard(dst, 36, 12, 568, 336, accent)
	menuHeading(dst, r, title, 50, 15)

	stats := m.LastWorldStats
	seedTag := "SEED " + game.FormatRunSeed(stats.Seed)
	if m.NewRecord {
		menuChip(dst, r, "NUOVO RECORD", 388, 21, 96, menuGold)
	}
	menuChip(dst, r, seedTag, 492, 21, 98, menuTeal)
	if stats.Quip != "" {
		r.DrawTextColored(dst, fitText(stats.Quip, 540), 50, 44, menuMuted)
	}
	fill(dst, 50, 60, 540, 1, menuRule)

	mins, secs := int(stats.RunTime)/60, int(stats.RunTime)%60
	fill(dst, 50, 64, 540, 48, menuSurface)
	fill(dst, 50, 64, 3, 48, accent)
	r.DrawTextColored(dst, "PUNTEGGIO", 62, 66, menuMuted)
	scoreText := fitText(fmt.Sprint(stats.Score), 118)
	r.DrawScaledText(dst, scoreText, 62, 80, 2, menuCream, menuInk)
	fill(dst, 305, 71, 1, 34, menuRule)
	r.DrawTextColored(dst, "RECORD PERSONALE", 318, 70, menuMuted)
	bestScore := 0
	selectedTitle := ""
	if save != nil {
		bestScore = save.Profile.BestScore
		selectedTitle = save.Profile.SelectedTitle
	}
	r.DrawTextColored(dst, fmt.Sprint(bestScore), 318, 88, menuGold)
	if selectedTitle != "" {
		profileTitle := fitText(selectedTitle, 150)
		r.DrawTextColored(dst, profileTitle, 580-render.TextPixelWidth(profileTitle, 1), 88, menuTeal)
	}

	metrics := [...]struct {
		label, value string
		col          color.RGBA
	}{
		{"TEMPO", fmt.Sprintf("%02d:%02d", mins, secs), menuCream},
		{"TIER", fmt.Sprint(stats.Tier), menuCream},
		{"PASSEGGERI", fmt.Sprint(stats.GirlsBoarded), menuCream},
		{"BOSS", fmt.Sprint(stats.Bosses), menuCream},
		{"STREAK / MAX *", fmt.Sprintf("x%d / %d*", stats.MaxStreak, stats.PeakWanted), menuGold},
	}
	for i, metric := range metrics {
		x := 50 + float32(i)*109
		fill(dst, x, 117, 104, 34, menuSurface)
		fill(dst, x, 117, 22, 1, menuRule)
		centerText(dst, r, metric.label, x, 119, 104, menuMuted)
		centerText(dst, r, fitText(metric.value, 96), x, 134, 104, metric.col)
	}

	drawEndRunBuildColumn(dst, r, stats)
	drawEndRunRightColumn(dst, r, stats, m.RunUnlocks)

	labels := []string{"RIGIOCA [R]", "MENU PRINCIPALE"}
	if victory {
		labels = []string{"CONTINUA OVERDRIVE", "NUOVA RUN", "MENU PRINCIPALE"}
	}
	for i, label := range labels {
		rect := endRunButtonRect(i, len(labels))
		style := m.buttonStyle(i, rect, mx, my)
		style.primary = i == 0
		drawMenuButton(dst, r, rect, label, style)
	}
}

func drawEndRunBuildColumn(dst *ebiten.Image, r *render.Renderer, stats RunSummary) {
	fill(dst, 50, 157, 264, 142, menuSurface)
	fill(dst, 50, 157, 3, 142, menuTeal)
	r.DrawTextColored(dst, "RIEPILOGO BUILD", 60, 159, menuTeal)

	archs := [...]struct {
		kind  game.Archetype
		label string
		col   color.RGBA
	}{
		{game.ArchEmo, "EMO", render.ColEmo},
		{game.ArchStudent, "UNI", render.ColStudent},
		{game.ArchKawaii, "POP", render.ColKawaii},
		{game.ArchGyaru, "5G", render.ColGyaru},
	}
	for i, arch := range archs {
		bx := 60 + float32(i)*62
		fill(dst, bx, 176, 58, 20, menuPanel)
		fill(dst, bx, 176, 58, 2, arch.col)
		r.DrawTextColored(dst, arch.label, int(bx+5), 179, arch.col)
		cnt := fmt.Sprint(stats.BoardedByArch[arch.kind])
		r.DrawTextColored(dst, cnt, int(bx+53)-render.TextPixelWidth(cnt, 1), 179, menuCream)
	}

	r.DrawTextColored(dst, "RELIQUIE", 60, 201, menuMuted)
	relicX := float32(116)
	relicCount := 0
	for rk := game.RelicKind(0); rk < game.RelicCount; rk++ {
		if !stats.Relics[rk] {
			continue
		}
		fill(dst, relicX-1, 200, 18, 18, menuPanel)
		fill(dst, relicX-1, 217, 18, 1, menuGold)
		r.DrawSpriteIcon(dst, endRunRelicIconKey(rk), relicX, 201, 1)
		relicX += 21
		relicCount++
	}
	if relicCount == 0 {
		r.DrawTextColored(dst, "Nessuna reliquia", 116, 201, menuMuted)
	}

	fill(dst, 60, 221, 246, 1, menuRule)
	r.DrawTextColored(dst, "UPGRADE ATTIVI", 60, 223, menuGold)

	type activeUpgrade struct {
		title string
		rank  int
	}
	var active []activeUpgrade
	for kind := game.UpgradeKind(0); kind < game.UpgradeCount; kind++ {
		rank := stats.UpgradeRanks[kind]
		if rank > 0 {
			active = append(active, activeUpgrade{title: endRunUpgradeTitle(kind), rank: rank})
		}
	}
	if len(active) == 0 {
		r.DrawTextColored(dst, "Nessun upgrade", 60, 242, menuMuted)
		return
	}

	const maxShown = 4
	shown := active
	if len(shown) > maxShown {
		shown = shown[:maxShown]
	}
	for i, upg := range shown {
		y := 238 + i*12
		fill(dst, 60, float32(y+6), 3, 3, menuTeal)
		r.DrawTextColored(dst, fitText(upg.title, 204), 68, y, menuCream)
		rankTxt := fmt.Sprintf("x%d", upg.rank)
		r.DrawTextColored(dst, rankTxt, 304-render.TextPixelWidth(rankTxt, 1), y, menuGold)
	}
	if len(active) > maxShown {
		r.DrawTextColored(dst, fmt.Sprintf("+%d altri", len(active)-maxShown), 68, 286, menuMuted)
	}
}

func drawEndRunRightColumn(dst *ebiten.Image, r *render.Renderer, stats RunSummary, unlocks []meta.ChallengeDef) {
	accent := guideSky
	if len(unlocks) > 0 {
		accent = menuGold
	}
	fill(dst, 326, 157, 264, 142, menuSurface)
	fill(dst, 326, 157, 3, 142, accent)

	if len(unlocks) > 0 {
		r.DrawTextColored(dst, "SFIDE SBLOCCATE", 336, 159, menuGold)
		for i, challenge := range unlocks[:min(3, len(unlocks))] {
			y := 178 + i*35
			fill(dst, 336, float32(y+6), 3, 3, menuTeal)
			r.DrawTextColored(dst, fitText(challenge.TitleIT, 234), 345, y, menuCream)
			r.DrawTextColored(dst, fitText(challenge.RewardIT, 234), 345, y+13, menuTeal)
		}
		if len(unlocks) > 3 {
			r.DrawTextColored(dst, fmt.Sprintf("+%d in Sfide & Skin", len(unlocks)-3), 345, 283, menuMuted)
		}
		return
	}

	r.DrawTextColored(dst, "HIGHLIGHTS RUN", 336, 159, guideSky)
	highlights := [...]struct {
		label string
		value string
		col   color.RGBA
	}{
		{"LIVELLO", fmt.Sprintf("LIV %d", stats.Level), menuCream},
		{"PERFECT DROP", fmt.Sprint(stats.PerfectDrops), menuTeal},
		{"EVENTI SUPERATI", fmt.Sprint(stats.EventsCompleted), menuGold},
	}
	for i, row := range highlights {
		y := float32(179 + i*24)
		fill(dst, 336, y, 244, 20, menuPanel)
		r.DrawTextColored(dst, row.label, 344, int(y+2), menuMuted)
		r.DrawTextColored(dst, row.value, 572-render.TextPixelWidth(row.value, 1), int(y+2), row.col)
	}
	fill(dst, 336, 254, 244, 1, menuRule)
	tip1 := "TIP: Le zone ZTL pagano di piu,"
	tip2 := "ma attirano la polizia. Usa il Dash!"
	if stats.BossRush {
		tip1 = "TIP: In Boss Rush alterna Chronos"
		tip2 = "e Dash per spezzare gli scudi!"
	}
	r.DrawTextColored(dst, tip1, 336, 261, menuMuted)
	r.DrawTextColored(dst, tip2, 336, 276, menuMuted)
}

func endRunRelicIconKey(rk game.RelicKind) string {
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

func endRunUpgradeTitle(kind game.UpgradeKind) string {
	switch kind {
	case game.UpgradeCooldown:
		return "ORARI TURBO"
	case game.UpgradeMultiHook:
		return "GANCIO DOPPIO PIANO"
	case game.UpgradeMaxStops:
		return "RETE EXPRESS"
	case game.UpgradeRange:
		return "PENSILINA XL"
	case game.UpgradeChronos:
		return "KATANA RIZZ-FOCUS"
	case game.UpgradeVitality:
		return "ESPRESSO OVERDRIVE"
	case game.UpgradeDashKatana:
		return "SCATTO KATANA"
	case game.UpgradeGlassFortress:
		return "VETRATA BLINDATA"
	case game.UpgradeMinigameTurbo:
		return "DITA VELOCI"
	case game.UpgradeBasslineShock:
		return "SUBWOOFER URBANO"
	case game.UpgradePerfectNova:
		return "IMPATTO PERFECT"
	case game.UpgradeRizzMagnet:
		return "BIGLIETTO D'ORO"
	default:
		return "UPGRADE"
	}
}
