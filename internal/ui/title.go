package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"girlbusAG/internal/meta"
	"girlbusAG/internal/render"
)

var titleLettering = [...]string{
	".#### ##### ####. #.... ####. #...# .####",
	"#.... ..#.. #...# #.... #...# #...# #....",
	"#.... ..#.. #...# #.... #...# #...# #....",
	"#.### ..#.. ####. #.... ####. #...# .###.",
	"#...# ..#.. #.#.. #.... #...# #...# ....#",
	"#...# ..#.. #..#. #.... #...# #...# ....#",
	".#### ##### #...# ##### ####. .###. ####.",
}

func drawPixelLettering(dst *ebiten.Image, lettering []string, x, y, size float32, c color.Color) {
	for gy, row := range lettering {
		for gx, pixel := range row {
			if pixel == '#' {
				fill(dst, x+float32(gx)*size, y+float32(gy)*size, size, size, c)
			}
		}
	}
}

func (m *Manager) drawTitle(dst *ebiten.Image, r *render.Renderer, save *meta.SaveData, mx, my float32) {
	screenShade(dst, 154)
	fill(dst, 0, 0, 310, 330, render.PremulColor(menuInk, 224))
	for i := 0; i < 4; i++ {
		fill(dst, 310+float32(i)*5, 0, 5, 330, render.PremulColor(menuInk, uint8(184-i*45)))
	}
	if m.titleArt == nil {
		m.titleArt = ebiten.NewImage(310, 144)
		drawTitleArt(m.titleArt, r)
	}
	dst.DrawImage(m.titleArt, nil)

	m.drawMapPicker(dst, r, mx, my)

	labels := [...]string{"GIOCA", "IMPARA A GIOCARE", "BOSS RUSH", "SFIDE & SKIN", "OPZIONI", "COMANDI", "ESCI"}
	accents := [...]color.RGBA{menuTeal, menuGold, menuCoral, guideLilac, {}, guideSky, {}}
	count := len(labels)
	if m.isWasm {
		count--
	}
	for i, label := range labels[:count] {
		rect := titleButtonRect(i)
		style := m.buttonStyle(i, rect, mx, my)
		style.accent = accents[i]
		style.primary, style.large = i == 0, i == 0
		if i == 0 && !m.titleMapPlayable() {
			style.locked = true
			label = "MAPPA BLOCCATA"
			if m.mapIndex == titleMapComingSoon {
				label = "MAPPE IN ARRIVO"
			}
		}
		if i == 2 {
			unlocked := save.Profile.IsBossRushUnlocked()
			style.locked = !unlocked || !m.titleMapPlayable()
			if !unlocked {
				label = "BOSS RUSH / BATTI 6 BOSS"
			}
		}
		drawMenuButton(dst, r, rect, label, style)
	}

	fill(dst, 0, 330, 640, 30, menuInk)
	fill(dst, 36, 331, 570, 1, menuRule)
	all := meta.AllChallenges()
	unlocked := 0
	for _, ch := range all {
		if save.Profile.UnlockedChallenges[ch.ID] {
			unlocked++
		}
	}
	stats := fmt.Sprintf("RECORD %d   RUN %d   SFIDE %d/%d", save.Profile.BestScore, save.Profile.TotalRuns, unlocked, len(all))
	r.DrawTextColored(dst, stats, 36, 337, menuGold)
	if save.Profile.SelectedTitle != "" {
		titleX := 52 + render.TextPixelWidth(stats, 1)
		r.DrawTextColored(dst, fitText(save.Profile.SelectedTitle, 470-titleX), titleX, 337, menuTeal)
	}
	const cue = "W/S MENU   A/D MAPPE"
	r.DrawTextColored(dst, cue, 606-render.TextPixelWidth(cue, 1), 337, menuMuted)
}

func drawTitleArt(dst *ebiten.Image, r *render.Renderer) {
	fill(dst, 36, 36, 15, 3, menuCoral)
	fill(dst, 53, 36, 8, 3, menuTeal)
	r.DrawTextColored(dst, "ARCADE DI QUARTIERE", 68, 30, menuMuted)
	drawPixelLettering(dst, titleLettering[:], 40, 61, 6, color.RGBA{162, 75, 111, 255})
	drawPixelLettering(dst, titleLettering[:], 36, 55, 6, menuCream)
	fill(dst, 36, 110, 260, 27, menuCoral)
	fill(dst, 39, 133, 254, 1, render.PremulColor(menuInk, 95))
	r.DrawScaledText(dst, "NEON STOP", 47, 107, 2, menuInk, color.Transparent)
	for i := 0; i < 3; i++ {
		fill(dst, 267+float32(i)*7, 119, 3, 9, menuInk)
	}
}

func drawBusPictogram(dst *ebiten.Image, x, y float32, c color.RGBA) {
	fill(dst, x+1, y, 12, 1, c)
	fill(dst, x, y+1, 14, 9, c)
	fill(dst, x+2, y+2, 4, 4, menuInk)
	fill(dst, x+8, y+2, 4, 4, menuInk)
	fill(dst, x+2, y+10, 3, 2, c)
	fill(dst, x+9, y+10, 3, 2, c)
}
