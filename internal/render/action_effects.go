package render

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"girlbusAG/internal/game"
)

var (
	actionCream = color.RGBA{255, 243, 224, 255}
	actionTeal  = color.RGBA{114, 226, 197, 255}
	actionCoral = color.RGBA{255, 136, 122, 255}
)

func snapPixel(value float32) float32 {
	return float32(math.Round(float64(value)))
}

func drawContactShadow(dst *ebiten.Image, x, y, width float32) {
	x, y = snapPixel(x), snapPixel(y)
	outer := Premul(26, 16, 38, 72)
	inner := Premul(26, 16, 38, 104)
	vector.DrawFilledRect(dst, x-width/2, y, width, 2, outer, false)
	vector.DrawFilledRect(dst, x-width/2+2, y-1, width-4, 4, outer, false)
	vector.DrawFilledRect(dst, x-width/2+3, y, width-6, 2, inner, false)
}

func drawImpactSparks(dst *ebiten.Image, x, y, radius, fade float32, tint color.RGBA, shield bool) {
	vector.StrokeCircle(dst, x, y, max(3, radius*0.72), 1, PremulColor(tint, uint8(114*fade)), true)
	sparks := 8
	if shield {
		sparks = 6
	}
	for i := 0; i < sparks; i++ {
		angle := (float64(i) + 0.5) * 2 * math.Pi / float64(sparks)
		dx, dy := float32(math.Cos(angle)), float32(math.Sin(angle))
		x1, y1 := x+dx*(radius-3), y+dy*(radius-3)
		x2, y2 := x+dx*(radius+3+5*fade), y+dy*(radius+3+5*fade)
		vector.StrokeLine(dst, x1, y1, x2, y2, 2.0, PremulColor(tint, uint8(225*fade)), true)
		vector.DrawFilledRect(dst, snapPixel(x2)-1, snapPixel(y2)-1, 2, 2, PremulColor(actionCream, uint8(235*fade)), false)
	}
}

func drawTether(dst *ebiten.Image, x1, y1, x2, y2, clock float32, tint color.RGBA) {
	vector.StrokeLine(dst, x1, y1, x2, y2, 3, Premul(27, 18, 42, 168), true)
	vector.StrokeLine(dst, x1, y1, x2, y2, 1.5, PremulColor(tint, 235), true)
	for i := 0; i < 2; i++ {
		phase := clock*1.8 + float32(i)*0.5
		phase -= float32(math.Floor(float64(phase)))
		x := snapPixel(x1 + (x2-x1)*phase)
		y := snapPixel(y1 + (y2-y1)*phase)
		vector.DrawFilledRect(dst, x-1, y-1, 3, 3, actionCream, false)
	}
}

func drawHazardWarning(dst *ebiten.Image, x1, y1, x2, y2, progress float32, cruiser bool) {
	dx, dy := x2-x1, y2-y1
	length := float32(math.Hypot(float64(dx), float64(dy)))
	if length < 1 {
		return
	}

	halfWidth := float32(6)
	if cruiser {
		halfWidth = 9
	}
	nx, ny := -dy/length*halfWidth, dx/length*halfWidth
	edge := PremulColor(actionCoral, uint8(92+116*progress))
	vector.StrokeLine(dst, x1, y1, x2, y2, 2*halfWidth, Premul(82, 21, 45, uint8(58+42*progress)), true)
	vector.StrokeLine(dst, x1+nx, y1+ny, x2+nx, y2+ny, 1, edge, true)
	vector.StrokeLine(dst, x1-nx, y1-ny, x2-nx, y2-ny, 1, edge, true)
	vector.StrokeLine(dst, x1, y1, x2, y2, 1, PremulColor(actionCoral, 174), true)
	vector.StrokeLine(dst, x1, y1, x1+dx*progress, y1+dy*progress, 2, PremulColor(actionCream, uint8(82+148*progress)), true)
}

func drawStopDeployment(dst *ebiten.Image, x, y, height, age float32, tint color.RGBA) {
	fade := 1 - clampF(age/0.18, 0, 1)
	if fade <= 0 {
		return
	}
	vector.StrokeLine(dst, x-13, y+11, x-13, y-height-15, 1, PremulColor(tint, uint8(190*fade)), false)
	vector.StrokeLine(dst, x+13, y+11, x+13, y-height-15, 1, PremulColor(tint, uint8(190*fade)), false)
	vector.DrawFilledRect(dst, x-13, y+11, 6, 1, PremulColor(actionCream, uint8(225*fade)), false)
	vector.DrawFilledRect(dst, x+8, y+11, 6, 1, PremulColor(actionCream, uint8(225*fade)), false)
}

func drawDashStreak(dst *ebiten.Image, x, y, dx, dy, fade float32, tint color.RGBA) {
	vector.StrokeLine(dst, x-dx*34-dy*5, y-dy*34+dx*5, x-dx*8-dy*5, y-dy*8+dx*5, 1.4, PremulColor(tint, uint8(135*fade)), true)
	vector.StrokeLine(dst, x-dx*23+dy*7, y-dy*23-dx*7, x-dx*5+dy*7, y-dy*5-dx*7, 1, PremulColor(actionCream, uint8(160*fade)), true)
}

func drawStopPanel(dst *ebiten.Image, x, y, width, height float32, accent color.RGBA) {
	vector.DrawFilledRect(dst, x+2, y+3, width, height, Premul(15, 10, 22, 132), false)
	vector.DrawFilledRect(dst, x, y, width, height, Premul(35, 26, 44, 225), false)
	vector.DrawFilledRect(dst, x+1, y+1, width-2, 19, Premul(52, 37, 55, 215), false)
	vector.StrokeRect(dst, x+0.5, y+0.5, width-1, height-1, 1, PremulColor(accent, 155), false)
	vector.DrawFilledRect(dst, x, y, 20, 1, accent, false)
	vector.DrawFilledRect(dst, x+1, y+20, width-2, 1, PremulColor(accent, 64), false)
}

const uniCellSize = 11

var uniRouteMarkers = [game.NodeBoard + 1]struct {
	label  string
	fill   color.RGBA
	symbol [5]string
}{
	game.NodeBlue: {
		label: "BONUS XP", fill: color.RGBA{123, 203, 255, 255},
		symbol: [5]string{"#.#.##.", "#.#.#.#", ".#..##.", "#.#.#..", "#.#.#.."},
	},
	game.NodeYellow: {
		label: "COLPI LENTI", fill: color.RGBA{255, 212, 115, 255},
		symbol: [5]string{".#####.", "..#.#..", "...#...", "..#.#..", ".#####."},
	},
	game.NodeBoard: {
		label: "ARRIVO", fill: color.RGBA{145, 239, 171, 255},
		symbol: [5]string{".#####.", ".#.#.#.", ".#####.", ".#.....", ".#....."},
	},
}

func (r *Renderer) drawUniRoute(dst *ebiten.Image, x, y float32, state *game.MinigameState) {
	for gy, row := range state.GridNodes {
		for gx, node := range row {
			cx := x + 8 + float32(gx)*(uniCellSize+1)
			cy := y + 24 + float32(gy)*(uniCellSize+1)
			drawUniRouteNode(dst, cx, cy, node)
			if gx == state.GridX && gy == state.GridY {
				vector.StrokeRect(dst, cx-0.5, cy-0.5, uniCellSize+1, uniCellSize+1, 1, actionCream, false)
			}
		}
	}

	vector.DrawFilledRect(dst, x+58, y+39, 75, 1, Premul(226, 210, 238, 55), false)
	for i, kind := range [...]int{game.NodeBoard, game.NodeBlue, game.NodeYellow} {
		lineY := y + 25 + float32(i)*16
		drawUniRouteNode(dst, x+58, lineY, kind)
		marker := uniRouteMarkers[kind]
		r.drawTextColored(dst, marker.label, int(x+72), int(lineY)-2, marker.fill)
	}
}

func drawUniRouteNode(dst *ebiten.Image, x, y float32, kind int) {
	if kind < game.NodeBlue || kind > game.NodeVisited {
		vector.DrawFilledRect(dst, x, y, uniCellSize, uniCellSize, Premul(54, 59, 86, 230), false)
		return
	}
	if kind == game.NodeVisited {
		vector.DrawFilledRect(dst, x, y, uniCellSize, uniCellSize, Premul(47, 133, 131, 220), false)
		vector.DrawFilledRect(dst, x+4, y+4, 3, 3, Premul(173, 241, 216, 170), false)
		return
	}
	marker := uniRouteMarkers[kind]
	vector.DrawFilledRect(dst, x, y, uniCellSize, uniCellSize, marker.fill, false)
	for sy, row := range marker.symbol {
		for sx, pixel := range row {
			if pixel == '#' {
				vector.DrawFilledRect(dst, x+2+float32(sx), y+3+float32(sy), 1, 1, colOutline, false)
			}
		}
	}
}
