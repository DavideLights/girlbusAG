package ui

import (
	"fmt"
	"image/color"
	"math"

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
		r.DrawTextColored(dst, spec.Desc, 318, y+8, menuMuted)
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
	screenShade(dst, 178)
	accent, title := menuCoral, "Corsa terminata"
	if victory {
		accent, title = menuGold, "Corsa completata"
		if m.LastWorldStats.BossRush {
			title = "Boss Rush completata"
		}
	}
	menuBoard(dst, 78, 28, 484, 301, accent)
	menuHeading(dst, r, title, 93, 41)
	if m.NewRecord {
		menuChip(dst, r, "NUOVO RECORD", 448, 47, 99, menuGold)
	}
	stats := m.LastWorldStats
	mins, secs := int(stats.RunTime)/60, int(stats.RunTime)%60
	fill(dst, 93, 80, 454, 60, menuSurface)
	fill(dst, 93, 80, 3, 60, accent)
	r.DrawTextColored(dst, "PUNTEGGIO", 107, 84, menuMuted)
	value := fitText(fmt.Sprint(stats.Score), 118)
	r.DrawScaledText(dst, value, 107, 99, 2, menuCream, menuInk)
	fill(dst, 300, 91, 1, 39, menuRule)
	r.DrawTextColored(dst, "RECORD PERSONALE", 315, 89, menuMuted)
	r.DrawTextColored(dst, fmt.Sprint(save.Profile.BestScore), 315, 108, menuGold)
	if save.Profile.SelectedTitle != "" {
		title := fitText(save.Profile.SelectedTitle, 127)
		r.DrawTextColored(dst, title, 540-render.TextPixelWidth(title, 1), 111, menuTeal)
	}
	for i, metric := range [...]struct{ label, value string }{
		{"TEMPO", fmt.Sprintf("%02d:%02d", mins, secs)},
		{"TIER", fmt.Sprint(stats.Tier)},
		{"PASSEGGERI", fmt.Sprint(stats.GirlsBoarded)},
		{"BOSS", fmt.Sprint(stats.Bosses)},
	} {
		x := 93 + float32(i)*115
		fill(dst, x, 146, 108, 38, menuSurface)
		centerText(dst, r, metric.label, x, 147, 108, menuMuted)
		centerText(dst, r, metric.value, x, 164, 108, menuCream)
	}

	if len(m.RunUnlocks) > 0 {
		r.DrawTextColored(dst, "SFIDE SBLOCCATE", 93, 194, menuTeal)
		for i, challenge := range m.RunUnlocks[:min(3, len(m.RunUnlocks))] {
			fill(dst, 93, float32(216+i*15), 3, 3, menuTeal)
			line := fitText(challenge.TitleIT+" / "+challenge.RewardIT, 442)
			r.DrawTextColored(dst, line, 103, 209+i*15, menuCream)
		}
		if len(m.RunUnlocks) > 3 {
			r.DrawTextColored(dst, fmt.Sprintf("+%d in Sfide & Skin", len(m.RunUnlocks)-3), 103, 254, menuMuted)
		}
	} else {
		r.DrawTextColored(dst, "Raccogli le Cure Espresso in strada.", 93, 210, menuMuted)
		r.DrawTextColored(dst, "Le zone ZTL pagano di più e attirano la polizia.", 93, 230, menuMuted)
	}

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
