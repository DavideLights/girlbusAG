package ui

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"girlbusAG/assets"
	"girlbusAG/internal/game"
	"girlbusAG/internal/input"
	"girlbusAG/internal/render"
)

const (
	titleMapCity = iota
	titleMapBeach
	titleMapComingSoon
	titleMapCount
)

const (
	mapPreviewWidth  = 272
	mapPreviewHeight = 184
	mapSlideDuration = 12
)

var (
	mapPickerBounds = game.Rect{X: 326, Y: 24, W: 280, H: 300}
	mapPreviousHit  = game.Rect{X: 337, Y: 290, W: 33, H: 26}
	mapNextHit      = game.Rect{X: 562, Y: 290, W: 33, H: 26}
	mapCards        = [titleMapCount]struct {
		title, subtitle, status string
		accent                  color.RGBA
	}{
		{title: "Centro Neon", subtitle: "Quartiere universitario", status: "GIOCABILE", accent: menuTeal},
		{title: "Spiaggia", subtitle: "Lungomare", status: "BLOCCATA", accent: menuGold},
		{title: "Upcoming soon", subtitle: "Altre mappe in arrivo", accent: guideLilac},
	}
)

func titleMapPageHit(i int) game.Rect {
	return game.Rect{X: 427 + float32(i*26), Y: 290, W: 26, H: 26}
}

func (m *Manager) titleMapPlayable() bool {
	return m.mapIndex == titleMapCity
}

func (m *Manager) browseTitleMaps(in input.MenuInput) bool {
	direction := 0
	pointerInPicker := mapPickerBounds.Contains(in.MouseX, in.MouseY)
	switch {
	case in.Left, in.MouseLeftJustPressed && mapPreviousHit.Contains(in.MouseX, in.MouseY),
		pointerInPicker && in.MapScroll < 0:
		direction = -1
	case in.Right, in.MouseLeftJustPressed && mapNextHit.Contains(in.MouseX, in.MouseY),
		pointerInPicker && in.MapScroll > 0:
		direction = 1
	case in.MouseLeftJustPressed && pointerInPicker:
		for i := range mapCards {
			if titleMapPageHit(i).Contains(in.MouseX, in.MouseY) {
				direction = (i - m.mapIndex + titleMapCount) % titleMapCount
				if direction > titleMapCount/2 {
					direction -= titleMapCount
				}
				break
			}
		}
	}
	if direction == 0 {
		return false
	}
	m.mapPrevious = m.mapIndex
	m.mapIndex = (m.mapIndex + direction + titleMapCount) % titleMapCount
	m.mapDirection = direction
	m.mapSlide = mapSlideDuration
	m.SelectedIdx = 0
	return true
}

func (m *Manager) drawMapPicker(dst *ebiten.Image, r *render.Renderer, mx, my float32) {
	box := mapPickerBounds
	card := mapCards[m.mapIndex]
	menuBoard(dst, box.X, box.Y, box.W, box.H, card.accent)
	drawBusPictogram(dst, 339, 37, card.accent)
	r.DrawTextColored(dst, "MAPPE", 361, 35, menuCream)
	menuChip(dst, r, fmt.Sprintf("%d / %d", m.mapIndex+1, titleMapCount), 541, 33, 53, card.accent)

	fill(dst, 329, 59, mapPreviewWidth+2, mapPreviewHeight+2, menuInk)
	preview := m.titleMapPreview(r, m.mapIndex)
	if m.mapSlide > 0 {
		if m.mapViewport == nil {
			m.mapViewport = ebiten.NewImage(mapPreviewWidth, mapPreviewHeight)
		}
		m.mapViewport.Clear()
		progress := 1 - float64(m.mapSlide)/mapSlideDuration
		offset := math.Round((1-progress)*(1-progress)*mapPreviewWidth) * float64(m.mapDirection)
		var op ebiten.DrawImageOptions
		op.GeoM.Translate(offset-float64(m.mapDirection*mapPreviewWidth), 0)
		m.mapViewport.DrawImage(m.titleMapPreview(r, m.mapPrevious), &op)
		op.GeoM.Reset()
		op.GeoM.Translate(offset, 0)
		m.mapViewport.DrawImage(preview, &op)
		preview = m.mapViewport
	}
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(330, 60)
	dst.DrawImage(preview, &op)

	fill(dst, 330, 244, mapPreviewWidth, 41, menuSurface)
	fill(dst, 330, 244, 3, 41, card.accent)
	r.DrawScaledText(dst, card.title, 341, 246, 2, card.accent, menuInk)
	r.DrawTextColored(dst, card.subtitle, 342, 270, menuMuted)
	if card.status != "" {
		fill(dst, 504, 256, 87, 17, render.PremulColor(card.accent, 39))
		centerText(dst, r, card.status, 504, 256, 87, card.accent)
	}

	drawMapArrow(dst, mapPreviousHit, -1, card.accent, m.usingPointer && mapPreviousHit.Contains(mx, my))
	drawMapArrow(dst, mapNextHit, 1, card.accent, m.usingPointer && mapNextHit.Contains(mx, my))
	for i := range mapCards {
		hit := titleMapPageHit(i)
		x := hit.X + 6
		ink := menuRule
		if i == m.mapIndex || (m.usingPointer && hit.Contains(mx, my)) {
			ink = mapCards[i].accent
		}
		fill(dst, x, hit.Y+11, 14, 4, ink)
		if i == m.mapIndex {
			fill(dst, x+2, hit.Y+8, 10, 1, ink)
		}
	}
}

func drawMapArrow(dst *ebiten.Image, box game.Rect, direction int, accent color.RGBA, hovered bool) {
	face := menuSurface
	ink := accent
	if hovered {
		face, ink = accent, menuInk
	}
	fill(dst, box.X+1, box.Y+3, box.W, box.H, menuInk)
	fill(dst, box.X, box.Y, box.W, box.H, menuRule)
	fill(dst, box.X+1, box.Y+1, box.W-2, box.H-2, face)
	x := box.X + 14
	y := box.Y + 8
	for row := 0; row < 11; row++ {
		dx := float32(direction * (min(row, 10-row) - 2))
		fill(dst, x+dx, y+float32(row), 3, 1, ink)
	}
}

func (m *Manager) titleMapPreview(r *render.Renderer, index int) *ebiten.Image {
	if m.mapPreviews[index] != nil {
		return m.mapPreviews[index]
	}
	img := ebiten.NewImage(mapPreviewWidth, mapPreviewHeight)
	switch index {
	case titleMapCity:
		drawCityPreview(img, r)
	case titleMapBeach:
		drawBeachPreview(img)
	case titleMapComingSoon:
		drawUnknownPreview(img)
	}
	m.mapPreviews[index] = img
	return img
}

func drawCityPreview(dst *ebiten.Image, r *render.Renderer) {
	fill(dst, 0, 0, mapPreviewWidth, mapPreviewHeight, color.RGBA{78, 56, 104, 255})
	fill(dst, 0, 33, mapPreviewWidth, 112, color.RGBA{115, 64, 116, 255})
	fill(dst, 0, 61, mapPreviewWidth, 75, color.RGBA{174, 82, 121, 255})
	fill(dst, 0, 88, mapPreviewWidth, 36, color.RGBA{216, 113, 129, 255})
	vector.DrawFilledCircle(dst, 169, 63, 52, color.RGBA{255, 159, 136, 255}, false)
	vector.DrawFilledCircle(dst, 169, 63, 46, color.RGBA{255, 201, 152, 255}, false)
	for i, width := range [...]float32{99, 87, 65} {
		fill(dst, 169-width/2, 60+float32(i*14), width, 2+float32(i), color.RGBA{255, 151, 134, 255})
	}
	fill(dst, 21, 22, 13, 1, menuGold)
	fill(dst, 27, 16, 1, 13, menuGold)
	fill(dst, 237, 36, 7, 1, menuCoral)
	fill(dst, 240, 33, 1, 7, menuCoral)
	for _, building := range [...]struct{ x, top, w float32 }{
		{0, 90, 24}, {21, 73, 33}, {59, 86, 34}, {84, 57, 24},
		{180, 95, 26}, {215, 73, 29}, {241, 91, 31},
	} {
		fill(dst, building.x, building.top, building.w, 123-building.top, color.RGBA{82, 56, 93, 255})
		fill(dst, building.x+4, building.top-3, building.w-10, 3, color.RGBA{82, 56, 93, 255})
		fill(dst, building.x+5, building.top+8, building.w-12, 2, color.RGBA{172, 109, 137, 255})
	}
	fill(dst, 0, 115, mapPreviewWidth, 31, color.RGBA{85, 62, 98, 255})
	fill(dst, 0, 131, mapPreviewWidth, 8, color.RGBA{178, 129, 138, 255})
	fill(dst, 0, 139, mapPreviewWidth, 3, color.RGBA{232, 180, 152, 255})
	fill(dst, 0, 142, mapPreviewWidth, 42, color.RGBA{56, 52, 79, 255})
	fill(dst, 6, 168, 254, 2, color.RGBA{213, 150, 135, 255})
	for i := 0; i < 4; i++ {
		fill(dst, 34+float32(i)*62, 175, 38, 2, color.RGBA{105, 82, 118, 255})
	}
	drawSpriteIconCentered(dst, r, assets.SpriteBuildingKonbini, 56, 100, 1)
	drawSpriteIconCentered(dst, r, assets.SpriteBuildingArcade, 220, 104, 1)
	drawSpriteIconCentered(dst, r, assets.SpritePropTrafficLight, 103, 109, 1)
	drawSpriteIconCentered(dst, r, assets.SpriteStopKawaii, 35, 136, 1)
	fill(dst, 63, 158, 158, 9, menuInk)
	drawSpriteIconCentered(dst, r, assets.SpriteBusNeon, 146, 127, 2)
	drawSpriteIconCentered(dst, r, assets.SpriteGirlEmo0, 23, 160, 1)
	drawSpriteIconCentered(dst, r, assets.SpriteGirlKawaii0, 47, 166, 1)
	drawSpriteIconCentered(dst, r, assets.SpriteAlfredoIdle0, 246, 149, 2)
}

func drawBeachPreview(dst *ebiten.Image) {
	fill(dst, 0, 0, mapPreviewWidth, mapPreviewHeight, color.RGBA{67, 91, 130, 255})
	vector.DrawFilledCircle(dst, 187, 62, 50, color.RGBA{174, 133, 121, 255}, false)
	fill(dst, 0, 83, mapPreviewWidth, 59, color.RGBA{52, 111, 137, 255})
	for i := 0; i < 10; i++ {
		fill(dst, float32((i*73)%204), 93+float32(i*5), 30+float32((i%4)*7), 2, color.RGBA{113, 163, 169, 255})
	}
	fill(dst, 0, 137, mapPreviewWidth, 47, color.RGBA{134, 105, 117, 255})
	fill(dst, 0, 137, mapPreviewWidth, 2, color.RGBA{178, 145, 123, 255})
	silhouette := color.RGBA{10, 12, 24, 255}
	for i := 0; i < 8; i++ {
		fill(dst, 43+float32(i*2), 56+float32(i*10), 5, 13, silhouette)
	}
	for _, leaf := range [...]struct{ x, y, w, h float32 }{
		{22, 48, 51, 5}, {15, 53, 28, 4}, {43, 42, 36, 6},
		{26, 39, 24, 5}, {55, 51, 27, 5}, {30, 56, 15, 6},
	} {
		fill(dst, leaf.x, leaf.y, leaf.w, leaf.h, silhouette)
	}
	fill(dst, 219, 109, 3, 43, silhouette)
	fill(dst, 211, 106, 19, 4, silhouette)
	fill(dst, 204, 110, 33, 4, silhouette)
	fill(dst, 197, 114, 47, 5, silhouette)
	fill(dst, 0, 0, mapPreviewWidth, mapPreviewHeight, render.Premul(9, 9, 19, 219))

	// Keep the lock readable over the darkened scene.
	fill(dst, 121, 58, 30, 4, menuGold)
	fill(dst, 117, 62, 6, 20, menuGold)
	fill(dst, 149, 62, 6, 20, menuGold)
	fill(dst, 112, 79, 48, 31, color.RGBA{158, 102, 114, 255})
	fill(dst, 112, 79, 48, 25, menuGold)
	fill(dst, 130, 86, 12, 8, menuInk)
	fill(dst, 134, 93, 4, 7, menuInk)
}

func drawUnknownPreview(dst *ebiten.Image) {
	fill(dst, 0, 0, mapPreviewWidth, mapPreviewHeight, color.RGBA{31, 29, 54, 255})
	for x := 16; x < mapPreviewWidth-12; x += 16 {
		for y := 16; y < mapPreviewHeight-12; y += 16 {
			fill(dst, float32(x), float32(y), 1, 1, color.RGBA{85, 69, 110, 255})
		}
	}
	fill(dst, 101, 53, 78, 80, menuInk)
	fill(dst, 97, 49, 78, 80, color.RGBA{74, 56, 102, 255})
	fill(dst, 99, 51, 74, 76, color.RGBA{47, 40, 74, 255})
	fill(dst, 97, 49, 22, 2, guideLilac)
	mark := [...]string{".###.", "#...#", "....#", "...#.", "..#..", ".....", "..#.."}
	drawPixelLettering(dst, mark[:], 123, 70, 6, color.RGBA{116, 81, 157, 255})
	drawPixelLettering(dst, mark[:], 121, 66, 6, guideLilac)
}
