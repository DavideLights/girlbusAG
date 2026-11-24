package ui

import (
	"fmt"
	"image/color"

	"girlbusAG/assets"
	"girlbusAG/internal/game"
	"girlbusAG/internal/meta"
	"girlbusAG/internal/render"
	"github.com/hajimehoshi/ebiten/v2"
)

func (m *Manager) drawChallenges(dst *ebiten.Image, r *render.Renderer, save *meta.SaveData, mx, my float32) {
	screenShade(dst, 186)
	menuBoard(dst, 24, 17, 592, 327, menuGold)
	menuHeading(dst, r, "Sfide & skin", 36, 23)
	profile := &save.Profile
	all := meta.AllChallenges()
	count := 0
	for _, ch := range all {
		if profile.UnlockedChallenges[ch.ID] {
			count++
		}
	}
	menuChip(dst, r, fmt.Sprintf("%d / %d SFIDE", count, len(all)), 502, 34, 99, menuTeal)
	fill(dst, 36, 61, 565, 1, menuRule)

	const pageSize = 5
	start := m.ChallengePage * pageSize
	page := all[start:min(start+pageSize, len(all))]
	inspected := page[0]
	for i, ch := range page {
		rect := challengeRowRect(i)
		style := m.buttonStyle(4+i, rect, mx, my)
		unlocked := profile.UnlockedChallenges[ch.ID]
		if style.focused || style.hovered {
			inspected = ch
		}
		drawChallengeRow(dst, r, rect, ch, unlocked, style.focused || style.hovered)
	}
	for i, rect := range [...]game.Rect{challengePrevRect(), challengeNextRect()} {
		style := menuButtonStyle{focused: m.ChallengePage == i, hovered: m.usingPointer && rect.Contains(mx, my)}
		drawMenuButton(dst, r, rect, fmt.Sprintf("PAG. %d", i+1), style)
	}
	r.DrawTextColored(dst, "W / S scegli  A / D cambia", 220, 275, menuMuted)
	fill(dst, 36, 302, 360, 29, menuSurface)
	_, category, accent := challengeRewardMeta(inspected.ID)
	fill(dst, 36, 302, 2, 29, accent)
	status, statusColor := "BLOCCATO", menuMuted
	if profile.UnlockedChallenges[inspected.ID] {
		status, statusColor = "SBLOCCATO", menuTeal
	}
	r.DrawTextColored(dst, category+" / "+status, 45, 301, statusColor)
	r.DrawTextColored(dst, fitText(inspected.RewardIT, 325), 45, 314, menuCream)

	fill(dst, 406, 73, 203, 221, menuSurface)
	r.DrawTextColored(dst, "PERSONALIZZAZIONE", 414, 76, menuCream)
	fill(dst, 414, 95, 187, 34, menuInk)
	fill(dst, 414, 126, 187, 3, menuPanel)
	fill(dst, 506, 99, 1, 26, menuRule)
	busKey := assets.SpriteBusNeon
	if profile.SelectedBusSkin == meta.BusGoldDrift {
		busKey = assets.SpriteBusGold
	}
	drawSpriteIconCentered(dst, r, assets.PlayerSkinSpriteKey(profile.SelectedPlayerSkin, "idle_0"), 460, 112, 1)
	drawSpriteIconCentered(dst, r, busKey, 553, 112, 0.5)
	title := profile.SelectedTitle
	if title == "" {
		title = "Nessuno"
	}
	type cosmetic struct {
		label string
		value string
		rect  game.Rect
	}
	for i, row := range [...]cosmetic{
		{"SKIN ALFREDO", profile.PlayerSkinName(profile.SelectedPlayerSkin), challengeSkinPRect()},
		{"LIVREA BUS", profile.BusSkinName(profile.SelectedBusSkin), challengeSkinBRect()},
		{"TITOLO PROFILO", title, challengeTitleRect()},
	} {
		r.DrawTextColored(dst, row.label, int(row.rect.X), int(row.rect.Y)-14, menuMuted)
		drawMenuButton(dst, r, row.rect, row.value, m.buttonStyle(i, row.rect, mx, my))
	}
	bossCount := 0
	for b := 0; b < int(game.BossCount); b++ {
		c := menuRule
		if profile.IsBossDefeated(b) {
			c = menuTeal
			bossCount++
		}
		fill(dst, 414+float32(b)*32, 280, 27, 4, c)
	}
	r.DrawTextColored(dst, fmt.Sprintf("BOSS BATTUTI  %d / %d", bossCount, game.BossCount), 414, 262, menuGold)
	back := challengeBackRect()
	drawMenuButton(dst, r, back, "INDIETRO [ESC]", m.buttonStyle(3, back, mx, my))
}

func drawChallengeRow(dst *ebiten.Image, r *render.Renderer, rect game.Rect, ch meta.ChallengeDef, unlocked, active bool) {
	kind, _, accent := challengeRewardMeta(ch.ID)
	bg := menuSurface
	if active {
		bg = menuRaised
	}
	x, y := rect.X, rect.Y
	fill(dst, x, y, rect.W, rect.H, bg)
	if active {
		fill(dst, x, y, 3, rect.H, accent)
	}
	if unlocked {
		fill(dst, x+9, y+8, 5, 5, menuTeal)
		fill(dst, x+10, y+9, 3, 3, bg)
	} else {
		fill(dst, x+10, y+9, 3, 3, menuMuted)
	}
	r.DrawTextColored(dst, ch.TitleIT, int(x+20), int(y+3), menuCream)
	r.DrawTextColored(dst, fitText(ch.DescIT, 313), int(x+9), int(y+19), menuMuted)
	boxX, boxY := x+rect.W-34, y+2
	fill(dst, boxX, boxY, 32, 32, menuInk)
	if unlocked {
		fill(dst, boxX, boxY+31, 32, 1, accent)
	}
	drawChallengeReward(dst, r, ch.ID, kind, boxX+16, boxY+16, unlocked)
}

func drawChallengeReward(dst *ebiten.Image, r *render.Renderer, id, kind string, cx, cy float32, unlocked bool) {
	if unlocked {
		switch id {
		case "first_ride":
			drawSpriteIconCentered(dst, r, assets.SpriteAlfredoDripIdle0, cx, cy, 1)
		case "full_bus":
			drawSpriteIconCentered(dst, r, assets.SpriteAlfredoBananaIdle0, cx, cy, 1)
		case "all_synergies":
			drawSpriteIconCentered(dst, r, assets.SpriteAlfredoChickenIdle0, cx, cy, 1)
		case "chaos_king_5star":
			drawSpriteIconCentered(dst, r, assets.SpriteAlfredoSamuraiIdle0, cx, cy, 1)
		case "rizz_jackpot":
			drawSpriteIconCentered(dst, r, assets.SpriteBusGold, cx, cy, 0.3125)
		case "all_6_bosses":
			drawCrownIcon(dst, cx, cy, menuGold)
		default:
			if kind == "title" {
				drawTrophyIcon(dst, cx, cy, menuGold, true)
			} else {
				drawSpriteIconCentered(dst, r, assets.SpriteIconStarWanted, cx, cy, 1)
			}
		}
		return
	}
	muted := color.RGBA{153, 127, 149, 255}
	switch kind {
	case "skin":
		drawHangerIcon(dst, cx, cy, muted)
	case "bus":
		drawBusSilhouetteIcon(dst, cx, cy, muted)
	case "boss":
		drawCrownIcon(dst, cx, cy, muted)
	case "title":
		drawTrophyIcon(dst, cx, cy, muted, false)
	default:
		drawStarOutlineIcon(dst, cx, cy, muted)
	}
	menuLock(dst, cx+5, cy-12, muted)
}
