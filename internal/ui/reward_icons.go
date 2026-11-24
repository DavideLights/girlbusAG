package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"girlbusAG/internal/render"
)

func challengeRewardMeta(id string) (kind string, label string, col color.RGBA) {
	switch id {
	case "first_ride", "full_bus", "all_synergies", "chaos_king_5star":
		return "skin", "SKIN", render.ColGyaru
	case "rizz_jackpot":
		return "bus", "BUS", render.ColGold
	case "all_6_bosses":
		return "boss", "MODO", render.ColRedAlert
	case "outlaw_3star", "beat_rector_3star", "beat_mom_clean":
		return "title", "TITOLO", render.ColKawaii
	default:
		return "badge", "BADGE", render.ColGreenNeon
	}
}

// drawHangerIcon draws a crisp clothes hanger icon inside locked Skin reward squares.
func drawHangerIcon(dst *ebiten.Image, cx, cy float32, col color.RGBA) {
	// Hook
	vector.StrokeLine(dst, cx, cy-8, cx+2.5, cy-8, 1.5, col, false)
	vector.StrokeLine(dst, cx+2.5, cy-8, cx+2.5, cy-5.5, 1.5, col, false)
	vector.StrokeLine(dst, cx+2.5, cy-5.5, cx, cy-4, 1.5, col, false)
	// Shoulders + bottom bar
	vector.StrokeLine(dst, cx, cy-4, cx-9, cy+4, 1.6, col, true)
	vector.StrokeLine(dst, cx, cy-4, cx+9, cy+4, 1.6, col, true)
	vector.StrokeLine(dst, cx-9, cy+4, cx+9, cy+4, 1.6, col, false)
}

// drawBusSilhouetteIcon draws a dark mini bus icon for locked Bus livery rewards.
func drawBusSilhouetteIcon(dst *ebiten.Image, cx, cy float32, col color.RGBA) {
	vector.DrawFilledRect(dst, cx-9, cy-5, 18, 10, color.RGBA{22, 24, 36, 240}, false)
	vector.StrokeRect(dst, cx-9, cy-5, 18, 10, 1.3, col, false)
	vector.DrawFilledRect(dst, cx-6, cy-3, 5, 4, color.RGBA{10, 10, 18, 240}, false)
	vector.DrawFilledRect(dst, cx+1, cy-3, 5, 4, color.RGBA{10, 10, 18, 240}, false)
	vector.DrawFilledCircle(dst, cx-5.5, cy+6, 2.0, col, true)
	vector.DrawFilledCircle(dst, cx+5.5, cy+6, 2.0, col, true)
}

// drawCrownIcon draws a spiky crown silhouette for locked Boss Rush / Mode rewards.
func drawCrownIcon(dst *ebiten.Image, cx, cy float32, col color.RGBA) {
	vector.StrokeLine(dst, cx-8, cy+5, cx+8, cy+5, 1.6, col, false)
	vector.StrokeLine(dst, cx-8, cy+5, cx-9, cy-4, 1.5, col, true)
	vector.StrokeLine(dst, cx-9, cy-4, cx-3, cy+1, 1.5, col, true)
	vector.StrokeLine(dst, cx-3, cy+1, cx, cy-7, 1.6, col, true)
	vector.StrokeLine(dst, cx, cy-7, cx+3, cy+1, 1.5, col, true)
	vector.StrokeLine(dst, cx+3, cy+1, cx+9, cy-4, 1.5, col, true)
	vector.StrokeLine(dst, cx+9, cy-4, cx+8, cy+5, 1.5, col, true)
}

// drawTrophyIcon draws a chalice/trophy icon for Titles (filled gold when unlocked, muted outline when locked).
func drawTrophyIcon(dst *ebiten.Image, cx, cy float32, col color.RGBA, unlocked bool) {
	if unlocked {
		vector.DrawFilledRect(dst, cx-5, cy-6, 10, 7, col, false)
		vector.StrokeRect(dst, cx-5, cy-6, 10, 7, 1.1, render.ColOutline, false)
	} else {
		vector.StrokeRect(dst, cx-5, cy-6, 10, 7, 1.4, col, false)
	}
	vector.StrokeLine(dst, cx-5, cy-4, cx-8, cy-4, 1.2, col, false)
	vector.StrokeLine(dst, cx-8, cy-4, cx-6, cy, 1.2, col, true)
	vector.StrokeLine(dst, cx+5, cy-4, cx+8, cy-4, 1.2, col, false)
	vector.StrokeLine(dst, cx+8, cy-4, cx+6, cy, 1.2, col, true)
	vector.StrokeLine(dst, cx, cy+1, cx, cy+6, 1.6, col, false)
	vector.StrokeLine(dst, cx-5, cy+6, cx+5, cy+6, 1.6, col, false)
}

// drawStarOutlineIcon draws a subtle geometric badge/star icon for locked Badges.
func drawStarOutlineIcon(dst *ebiten.Image, cx, cy float32, col color.RGBA) {
	vector.StrokeCircle(dst, cx, cy, 7.5, 1.3, col, true)
	vector.StrokeLine(dst, cx, cy-5, cx, cy+5, 1.3, col, false)
	vector.StrokeLine(dst, cx-5, cy, cx+5, cy, 1.3, col, false)
	vector.StrokeLine(dst, cx-3.5, cy-3.5, cx+3.5, cy+3.5, 1.1, col, true)
	vector.StrokeLine(dst, cx+3.5, cy-3.5, cx-3.5, cy+3.5, 1.1, col, true)
}
