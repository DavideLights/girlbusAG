package render

import (
	"image/color"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"girlbusAG/internal/game"
)

// Exported Pop-Urban UI theme colors shared with internal/ui.
var (
	ColOutline   = colOutline
	ColEmo       = colEmo
	ColStudent   = colStudent
	ColKawaii    = colKawaii
	ColGyaru     = colGyaru
	ColGold      = colGold
	ColRedAlert  = colRedAlert
	ColGreenNeon = colGreenNeon
	ColHudBg     = colHudBg
	ColHoloBg    = colHoloBg
)

// Premul returns a premultiplied color.RGBA from non-premultiplied 8-bit channels.
func Premul(r, g, b, a uint8) color.RGBA {
	return color.RGBA{
		R: uint8(uint16(r) * uint16(a) / 255),
		G: uint8(uint16(g) * uint16(a) / 255),
		B: uint8(uint16(b) * uint16(a) / 255),
		A: a,
	}
}

// PremulColor scales an opaque color.RGBA by alpha a (0..255) in premultiplied form.
func PremulColor(c color.RGBA, a uint8) color.RGBA {
	return Premul(c.R, c.G, c.B, a)
}

// DrawSpriteIcon renders a crisp nearest-neighbor pixel-art sprite from the SpriteBank
// at integer-snapped (x, y) with the requested scale.
func (r *Renderer) DrawSpriteIcon(screen *ebiten.Image, spriteKey string, x, y, scale float32) {
	if screen == nil || r == nil || r.sprites == nil || spriteKey == "" {
		return
	}
	if scale <= 0 {
		scale = 1
	}
	img := r.sprites.Get(spriteKey)
	if img == nil {
		return
	}
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
	op.GeoM.Scale(float64(scale), float64(scale))
	op.GeoM.Translate(math.Round(float64(x)), math.Round(float64(y)))
	screen.DrawImage(img, op)
}

// drawSpriteIconTinted renders a crisp nearest-neighbor pixel-art sprite tinted with clr.
func (r *Renderer) drawSpriteIconTinted(screen *ebiten.Image, spriteKey string, x, y, scale float32, clr color.Color) {
	if screen == nil || r == nil || r.sprites == nil || spriteKey == "" {
		return
	}
	if scale <= 0 {
		scale = 1
	}
	img := r.sprites.Get(spriteKey)
	if img == nil {
		return
	}
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
	op.GeoM.Scale(float64(scale), float64(scale))
	op.GeoM.Translate(math.Round(float64(x)), math.Round(float64(y)))
	if clr != nil {
		op.ColorScale.ScaleWithColor(clr)
	}
	screen.DrawImage(img, op)
}

// DrawScaledText renders cached crisp text with 1px dark outline and drop-shadow.
func (r *Renderer) DrawScaledText(dst *ebiten.Image, txt string, x, y int, scale float64, fgColor, shadowColor color.Color) {
	r.drawScaledText(dst, txt, x, y, scale, fgColor, shadowColor)
}

// DrawTextColored renders cached 1x text with subtle dark drop-shadow.
func (r *Renderer) DrawTextColored(dst *ebiten.Image, txt string, x, y int, clr color.Color) {
	r.drawTextColored(dst, txt, x, y, clr)
}

// DrawKeycap renders a tactile 3D pixel-art keyboard button or arrow icon.
func (r *Renderer) DrawKeycap(dst *ebiten.Image, label string, x, y, w, h float32, bgCol, borderCol color.Color, pressed bool) {
	r.drawKeycap(dst, label, x, y, w, h, bgCol, borderCol, pressed)
}

// DrawPillBadge renders a compact dark neon-bordered badge.
func DrawPillBadge(dst *ebiten.Image, x, y, w, h float32, bgCol, borderCol color.Color) {
	drawPillBadge(dst, x, y, w, h, bgCol, borderCol)
}

// TextPixelWidth estimates pixel width of txt at scale.
func TextPixelWidth(txt string, scale float64) int {
	return textPixelWidth(txt, scale)
}

type accentMark struct {
	cellIdx int
	acute   bool
	upper   bool
	isI     bool
}

func mapItalianAccent(r rune) (base rune, hasAccent, acute, upper, isI bool) {
	switch r {
	case 'à':
		return 'a', true, false, false, false
	case 'è':
		return 'e', true, false, false, false
	case 'é':
		return 'e', true, true, false, false
	case 'ì':
		return 'i', true, false, false, true
	case 'ò':
		return 'o', true, false, false, false
	case 'ù':
		return 'u', true, false, false, false
	case 'À':
		return 'A', true, false, true, false
	case 'È':
		return 'E', true, false, true, false
	case 'É':
		return 'E', true, true, true, false
	case 'Ì':
		return 'I', true, false, true, true
	case 'Ò':
		return 'O', true, false, true, false
	case 'Ù':
		return 'U', true, false, true, false
	default:
		return r, false, false, false, false
	}
}

func (r *Renderer) getTextImage(txt string) *ebiten.Image {
	img, ok := r.textCache[txt]
	if ok {
		return img
	}
	if len(r.textCache) >= 640 {
		for k := range r.textCache {
			delete(r.textCache, k)
			break
		}
	}

	var sb strings.Builder
	sb.Grow(len(txt))
	var accents []accentMark
	cellIdx := 0
	for _, ch := range txt {
		base, hasAcc, acute, upper, isI := mapItalianAccent(ch)
		sb.WriteRune(base)
		if hasAcc {
			accents = append(accents, accentMark{
				cellIdx: cellIdx,
				acute:   acute,
				upper:   upper,
				isI:     isI,
			})
		}
		cellIdx++
	}

	w := cellIdx*6 + 2
	if w < 8 {
		w = 8
	}
	img = ebiten.NewImage(w, 16)
	ebitenutil.DebugPrintAt(img, sb.String(), 0, 0)
	for _, acc := range accents {
		// DebugPrintAt renders at ox+1=1 with 6px cell width.
		cx := float32(acc.cellIdx*6 + 1)
		topY := float32(2)
		if acc.upper || acc.isI {
			topY = 0
		}
		if acc.acute {
			vector.DrawFilledRect(img, cx+3, topY, 1, 1, color.White, false)
			vector.DrawFilledRect(img, cx+2, topY+1, 1, 1, color.White, false)
		} else {
			vector.DrawFilledRect(img, cx+1, topY, 1, 1, color.White, false)
			vector.DrawFilledRect(img, cx+2, topY+1, 1, 1, color.White, false)
		}
	}
	r.textCache[txt] = img
	return img
}

// drawScaledText renders cached text with a crisp 1px dark outline/drop-shadow and clean scaling.
func (r *Renderer) drawScaledText(dst *ebiten.Image, txt string, x, y int, scale float64, fgColor, shadowColor color.Color) {
	if txt == "" {
		return
	}
	if scale <= 0 {
		scale = 1.0
	}
	img := r.getTextImage(txt)
	op := &ebiten.DrawImageOptions{}

	_, _, _, sa := shadowColor.RGBA()
	if sa > 0 {
		offsets := [5][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}, {1, 1}}
		for _, off := range offsets {
			op.GeoM.Reset()
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(float64(x+off[0]), float64(y+off[1]))
			op.ColorScale.Reset()
			op.ColorScale.ScaleWithColor(shadowColor)
			dst.DrawImage(img, op)
		}
	}

	op.GeoM.Reset()
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.Reset()
	op.ColorScale.ScaleWithColor(fgColor)
	dst.DrawImage(img, op)
}

func (r *Renderer) drawTextColored(dst *ebiten.Image, txt string, x, y int, clr color.Color) {
	if txt == "" {
		return
	}
	img := r.getTextImage(txt)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x+1), float64(y+1))
	op.ColorScale.ScaleWithColor(color.RGBA{12, 6, 18, 225})
	dst.DrawImage(img, op)

	op.GeoM.Reset()
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.Reset()
	op.ColorScale.ScaleWithColor(clr)
	dst.DrawImage(img, op)
}

func textPixelWidth(txt string, scale float64) int {
	if scale <= 0 {
		scale = 1.0
	}
	return int(float64(utf8.RuneCountInString(txt)*6) * scale)
}

// drawKeycap renders a tactile 3D pixel-art keyboard button with top highlight, bottom bevel, and press offset.
func (r *Renderer) drawKeycap(dst *ebiten.Image, label string, x, y, w, h float32, bgCol, borderCol color.Color, pressed bool) {
	if w < 10 {
		w = 10
	}
	if h < 10 {
		h = 10
	}

	br, bg, bb, ba := rgba8(bgCol)
	rr, rg, rb, ra := rgba8(borderCol)

	bevelH := float32(3)
	pressOffset := float32(0)
	if pressed {
		bevelH = 1
		pressOffset = 2
	}

	faceY := y + pressOffset
	faceH := h - 3
	if faceH < 8 {
		faceH = h - 1
	}

	vector.DrawFilledRect(dst, x-1, y+1, w+2, h+1, color.RGBA{14, 8, 22, 220}, false)

	bevelCol := color.RGBA{
		R: uint8(float32(br) * 0.48),
		G: uint8(float32(bg) * 0.48),
		B: uint8(float32(bb) * 0.48),
		A: ba,
	}
	vector.DrawFilledRect(dst, x, faceY+faceH, w, bevelH, bevelCol, false)

	faceCol := color.RGBA{R: br, G: bg, B: bb, A: ba}
	if pressed {
		faceCol = color.RGBA{
			R: clampU8(int(br) + 28),
			G: clampU8(int(bg) + 28),
			B: clampU8(int(bb) + 28),
			A: ba,
		}
	}
	vector.DrawFilledRect(dst, x, faceY, w, faceH, faceCol, false)
	vector.DrawFilledRect(dst, x+1, faceY+1, w-2, 1.5, color.RGBA{255, 255, 255, 85}, false)
	vector.StrokeRect(dst, x, faceY, w, faceH, 1.0, color.RGBA{R: rr, G: rg, B: rb, A: ra}, false)

	cx := x + w*0.5
	cy := faceY + faceH*0.5
	fg := color.RGBA{248, 250, 255, 255}
	switch label {
	case "UP", "^":
		drawArrowIcon(dst, cx, cy, game.ActionUp, fg)
	case "RIGHT", ">":
		drawArrowIcon(dst, cx, cy, game.ActionRight, fg)
	case "DOWN", "v":
		drawArrowIcon(dst, cx, cy, game.ActionDown, fg)
	case "LEFT", "<":
		drawArrowIcon(dst, cx, cy, game.ActionLeft, fg)
	default:
		tw := float32(textPixelWidth(label, 1.0))
		tx := int(cx - tw*0.5)
		ty := int(cy - 8)
		r.drawScaledText(dst, label, tx, ty, 1.0, fg, color.RGBA{12, 6, 18, 240})
	}
}

func drawPillBadge(dst *ebiten.Image, x, y, w, h float32, bgCol, borderCol color.Color) {
	vector.DrawFilledRect(dst, x+1, y+1, w, h, color.RGBA{10, 5, 16, 175}, false)
	vector.DrawFilledRect(dst, x, y, w, h, bgCol, false)
	vector.StrokeRect(dst, x, y, w, h, 1.0, borderCol, false)
}

func drawArrowIcon(screen *ebiten.Image, cx, cy float32, dir int, clr color.Color) {
	switch dir {
	case game.ActionUp:
		vector.StrokeLine(screen, cx, cy-4.5, cx, cy+4.5, 2.2, clr, true)
		vector.StrokeLine(screen, cx, cy-4.5, cx-4, cy-0.5, 2.2, clr, true)
		vector.StrokeLine(screen, cx, cy-4.5, cx+4, cy-0.5, 2.2, clr, true)
	case game.ActionRight:
		vector.StrokeLine(screen, cx-4.5, cy, cx+4.5, cy, 2.2, clr, true)
		vector.StrokeLine(screen, cx+4.5, cy, cx+0.5, cy-4, 2.2, clr, true)
		vector.StrokeLine(screen, cx+4.5, cy, cx+0.5, cy+4, 2.2, clr, true)
	case game.ActionDown:
		vector.StrokeLine(screen, cx, cy-4.5, cx, cy+4.5, 2.2, clr, true)
		vector.StrokeLine(screen, cx, cy+4.5, cx-4, cy+0.5, 2.2, clr, true)
		vector.StrokeLine(screen, cx, cy+4.5, cx+4, cy+0.5, 2.2, clr, true)
	case game.ActionLeft:
		vector.StrokeLine(screen, cx-4.5, cy, cx+4.5, cy, 2.2, clr, true)
		vector.StrokeLine(screen, cx-4.5, cy, cx-0.5, cy-4, 2.2, clr, true)
		vector.StrokeLine(screen, cx-4.5, cy, cx-0.5, cy+4, 2.2, clr, true)
	}
}

func drawBoltIcon(dst *ebiten.Image, cx, cy float32, clr color.Color) {
	vector.StrokeLine(dst, cx+1, cy-5, cx-2.5, cy+0.5, 1.8, clr, true)
	vector.StrokeLine(dst, cx-2.5, cy+0.5, cx+2, cy+0.5, 1.8, clr, true)
	vector.StrokeLine(dst, cx+2, cy+0.5, cx-1, cy+5, 1.8, clr, true)
}

func drawHeartIcon(dst *ebiten.Image, cx, cy float32, clr color.Color, filled bool) {
	if filled {
		vector.DrawFilledCircle(dst, cx-2.2, cy-1.5, 2.4, clr, true)
		vector.DrawFilledCircle(dst, cx+2.2, cy-1.5, 2.4, clr, true)
		vector.DrawFilledRect(dst, cx-3.2, cy-1, 6.4, 3.0, clr, false)
		vector.DrawFilledRect(dst, cx-1.8, cy+1.5, 3.6, 2.0, clr, false)
		vector.DrawFilledRect(dst, cx-0.8, cy+3.0, 1.6, 1.2, clr, false)
	} else {
		vector.StrokeCircle(dst, cx-2.2, cy-1.5, 2.2, 1.0, clr, true)
		vector.StrokeCircle(dst, cx+2.2, cy-1.5, 2.2, 1.0, clr, true)
		vector.StrokeLine(dst, cx-4, cy, cx, cy+4, 1.0, clr, true)
		vector.StrokeLine(dst, cx+4, cy, cx, cy+4, 1.0, clr, true)
	}
}

func drawStarIcon(dst *ebiten.Image, cx, cy, r float32, clr color.Color) {
	vector.DrawFilledCircle(dst, cx, cy, r*0.55, clr, true)
	vector.StrokeLine(dst, cx, cy-r, cx, cy+r, 1.4, clr, true)
	vector.StrokeLine(dst, cx-r, cy, cx+r, cy, 1.4, clr, true)
	d := r * 0.65
	vector.StrokeLine(dst, cx-d, cy-d, cx+d, cy+d, 1.1, clr, true)
	vector.StrokeLine(dst, cx-d, cy+d, cx+d, cy-d, 1.1, clr, true)
}

func drawWarningIcon(dst *ebiten.Image, cx, cy float32, clr color.Color) {
	vector.StrokeLine(dst, cx, cy-5, cx-5, cy+4, 1.6, clr, true)
	vector.StrokeLine(dst, cx, cy-5, cx+5, cy+4, 1.6, clr, true)
	vector.StrokeLine(dst, cx-5, cy+4, cx+5, cy+4, 1.6, clr, true)
	vector.DrawFilledRect(dst, cx-0.8, cy-2, 1.6, 3.5, clr, false)
	vector.DrawFilledRect(dst, cx-0.8, cy+2.2, 1.6, 1.4, clr, false)
}

func rgba8(c color.Color) (uint8, uint8, uint8, uint8) {
	r, g, b, a := c.RGBA()
	return uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)
}

func clampU8(v int) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}
