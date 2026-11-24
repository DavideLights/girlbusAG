package ui

import (
	"image/color"
	"strings"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"girlbusAG/internal/game"
	"girlbusAG/internal/render"
)

const menuEntryDuration = 8

var (
	menuInk     = color.RGBA{24, 17, 28, 255}
	menuPanel   = color.RGBA{32, 28, 46, 255}
	menuSurface = color.RGBA{45, 39, 60, 255}
	menuRaised  = color.RGBA{63, 54, 78, 255}
	menuRule    = color.RGBA{100, 77, 109, 255}
	menuMuted   = color.RGBA{211, 191, 203, 255}
	menuCream   = color.RGBA{255, 239, 219, 255}
	menuCoral   = color.RGBA{255, 147, 149, 255}
	menuTeal    = color.RGBA{132, 230, 208, 255}
	menuGold    = color.RGBA{255, 213, 150, 255}
)

func fill(dst *ebiten.Image, x, y, w, h float32, c color.Color) {
	vector.DrawFilledRect(dst, x, y, w, h, c, false)
}

func screenShade(dst *ebiten.Image, opacity uint8) {
	fill(dst, 0, 0, game.ScreenWidth, game.ScreenHeight, render.PremulColor(menuInk, opacity))
}

func menuBoard(dst *ebiten.Image, x, y, w, h float32, accent color.RGBA) {
	fill(dst, x+4, y+6, w, h, render.Premul(10, 8, 19, 196))
	fill(dst, x, y, w, h, menuRule)
	fill(dst, x+1, y+1, w-2, h-2, menuPanel)
	fill(dst, x+1, y+2, w-2, 28, render.PremulColor(accent, 15))
	fill(dst, x+1, y+h-3, w-2, 2, menuInk)
	fill(dst, x, y, 48, 2, accent)
}

func menuHeading(dst *ebiten.Image, r *render.Renderer, title string, x, y int) {
	r.DrawScaledText(dst, title, x, y, 2, menuCream, menuInk)
}

func centerText(dst *ebiten.Image, r *render.Renderer, txt string, x, y, w float32, c color.Color) {
	r.DrawTextColored(dst, txt, int(x+(w-float32(render.TextPixelWidth(txt, 1)))/2), int(y), c)
}

func menuChip(dst *ebiten.Image, r *render.Renderer, txt string, x, y, w float32, c color.RGBA) {
	fill(dst, x, y, w, 16, menuSurface)
	fill(dst, x, y+15, w, 1, render.PremulColor(c, 135))
	centerText(dst, r, txt, x, y, w, c)
}

func menuChevron(dst *ebiten.Image, x, y float32, c color.Color) {
	fill(dst, x, y, 2, 1, c)
	fill(dst, x+1, y+1, 2, 1, c)
	fill(dst, x+2, y+2, 2, 2, c)
	fill(dst, x+1, y+4, 2, 1, c)
	fill(dst, x, y+5, 2, 1, c)
}

func menuLock(dst *ebiten.Image, x, y float32, c color.Color) {
	fill(dst, x+3, y, 4, 1, c)
	fill(dst, x+2, y+1, 1, 3, c)
	fill(dst, x+7, y+1, 1, 3, c)
	fill(dst, x, y+3, 10, 6, c)
	fill(dst, x+4, y+5, 2, 2, menuInk)
}

type menuButtonStyle struct {
	focused bool
	hovered bool
	primary bool
	locked  bool
	accent  color.RGBA
	large   bool
}

func (m *Manager) buttonStyle(index int, rect game.Rect, mx, my float32) menuButtonStyle {
	return menuButtonStyle{
		focused: m.SelectedIdx == index,
		hovered: m.usingPointer && rect.Contains(mx, my),
	}
}

func drawMenuButton(dst *ebiten.Image, r *render.Renderer, rect game.Rect, label string, style menuButtonStyle) {
	bg, edge, fg := menuSurface, menuRule, menuCream
	accent := style.accent
	if accent.A == 0 {
		accent = menuTeal
	}
	switch {
	case style.locked:
		bg, fg = menuPanel, menuMuted
		if style.focused || style.hovered {
			edge = menuMuted
		}
	case style.primary:
		bg, edge, fg = accent, accent, menuInk
		if style.hovered || style.focused {
			edge = menuCream
		}
	case style.focused || style.hovered:
		bg, edge, fg = menuRaised, accent, accent
	default:
		if style.accent.A != 0 {
			fg = accent
		}
	}
	fill(dst, rect.X+1, rect.Y+3, rect.W, rect.H, menuInk)
	fill(dst, rect.X, rect.Y, rect.W, rect.H, edge)
	fill(dst, rect.X+1, rect.Y+1, rect.W-2, rect.H-2, bg)
	if !style.locked && !style.primary && style.accent.A != 0 {
		fill(dst, rect.X+1, rect.Y+1, rect.W-2, rect.H-2, render.PremulColor(accent, 21))
	}
	if style.primary && !style.locked {
		fill(dst, rect.X+1, rect.Y+1, rect.W-2, 2, render.PremulColor(menuCream, 122))
		fill(dst, rect.X+1, rect.Y+rect.H-3, rect.W-2, 2, render.PremulColor(menuInk, 75))
	}
	if style.focused {
		fill(dst, rect.X, rect.Y, 3, rect.H, edge)
	}
	if style.locked {
		menuLock(dst, rect.X+rect.W-20, rect.Y+(rect.H-9)/2, fg)
	} else if style.primary || style.focused || style.hovered {
		menuChevron(dst, rect.X+rect.W-14, rect.Y+(rect.H-6)/2, fg)
	}
	scale := 1.0
	if style.large && !style.locked {
		scale = 2
	}
	label = fitText(label, int(float64(rect.W-34)/scale))
	if fg == menuInk {
		x := int(rect.X + (rect.W-float32(render.TextPixelWidth(label, scale)))/2)
		r.DrawScaledText(dst, label, x, int(rect.Y+(rect.H-16*float32(scale))/2), scale, fg, color.Transparent)
		return
	}
	centerText(dst, r, label, rect.X, rect.Y+(rect.H-16)/2, rect.W, fg)
}

func fitText(value string, width int) string {
	if render.TextPixelWidth(value, 1) <= width {
		return value
	}
	if width < 18 {
		return ""
	}
	for len(value) > 0 && render.TextPixelWidth(value, 1) > width-18 {
		_, size := utf8.DecodeLastRuneInString(value)
		value = value[:len(value)-size]
	}
	return strings.TrimSpace(value) + "..."
}

func titleButtonRect(i int) game.Rect {
	switch i {
	case 0:
		return game.Rect{X: 36, Y: 151, W: 260, H: 30}
	case 4, 5:
		return game.Rect{X: 36 + float32(i-4)*132, Y: 277, W: 128, H: 24}
	case 6:
		return game.Rect{X: 36, Y: 305, W: 70, H: 18}
	default:
		return game.Rect{X: 36, Y: 187 + float32(i-1)*30, W: 260, H: 25}
	}
}

func pauseButtonRect(i int) game.Rect {
	return game.Rect{X: 48, Y: 100 + float32(i)*41, W: 212, H: 32}
}

func tutorialTabRect(t int) game.Rect {
	return game.Rect{X: 38 + float32(t)*142, Y: 50, W: 136, H: 24}
}

func tutorialPrevRect() game.Rect     { return game.Rect{X: 38, Y: 300, W: 82, H: 28} }
func tutorialNextRect() game.Rect     { return game.Rect{X: 126, Y: 300, W: 82, H: 28} }
func tutorialPracticeRect() game.Rect { return game.Rect{X: 311, Y: 300, W: 161, H: 28} }
func tutorialBackRect() game.Rect     { return game.Rect{X: 480, Y: 300, W: 122, H: 28} }

func optionsRowRect(i int) game.Rect {
	if i == 9 {
		return game.Rect{X: 402, Y: 303, W: 128, H: 28}
	}
	return game.Rect{X: 108, Y: 60 + float32(i)*26, W: 424, H: 24}
}

func optionsSlider(rect game.Rect) (float32, float32) { return rect.X + 223, 139 }
func controlsBackRect() game.Rect                     { return game.Rect{X: 470, Y: 305, W: 132, H: 26} }

func challengeRowRect(i int) game.Rect {
	return game.Rect{X: 36, Y: 73 + float32(i)*39, W: 360, H: 36}
}

func challengePrevRect() game.Rect  { return game.Rect{X: 36, Y: 272, W: 82, H: 22} }
func challengeNextRect() game.Rect  { return game.Rect{X: 126, Y: 272, W: 82, H: 22} }
func challengeSkinPRect() game.Rect { return game.Rect{X: 414, Y: 146, W: 187, H: 24} }
func challengeSkinBRect() game.Rect { return game.Rect{X: 414, Y: 187, W: 187, H: 24} }
func challengeTitleRect() game.Rect { return game.Rect{X: 414, Y: 228, W: 187, H: 24} }
func challengeBackRect() game.Rect  { return game.Rect{X: 465, Y: 304, W: 136, H: 27} }

func endRunButtonRect(i, total int) game.Rect {
	if total == 3 {
		return game.Rect{X: 93 + float32(i)*154, Y: 278, W: 146, H: 30}
	}
	return game.Rect{X: 93 + float32(i)*231, Y: 278, W: 223, H: 30}
}
