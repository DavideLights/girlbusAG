package render

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"girlbusAG/internal/game"
)

const numAmbientParticles = 60

var colGoldenHourShadow = color.RGBA{24, 16, 36, 95}

type particleKind uint8

const (
	partSakuraPetal particleKind = iota
	partDustMote
	partStoreSteam
	partCanalSparkle
)

type ambientParticle struct {
	x, y       float32
	vx, vy     float32
	phase      float32
	life       float32
	maxLife    float32
	size       float32
	kind       particleKind
	initialized bool
}

func buildRadialLightTexture() *ebiten.Image {
	const size = 64
	rgba := image.NewRGBA(image.Rect(0, 0, size, size))
	center := float64(size-1) * 0.5
	maxR := float64(size) * 0.5

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - center
			dy := float64(y) - center
			dist := math.Sqrt(dx*dx+dy*dy) / maxR
			if dist >= 1.0 {
				continue
			}
			// Smooth quadratic-gaussian falloff reaching zero cleanly at the edge
			t := 1.0 - dist*dist
			falloff := t * t * math.Exp(-1.6*dist*dist)
			v := uint8(clampF(float32(falloff*255.0), 0, 255))
			rgba.SetRGBA(x, y, color.RGBA{R: v, G: v, B: v, A: v})
		}
	}
	return ebiten.NewImageFromImage(rgba)
}

func (r *Renderer) initAmbientParticles() {
	// Deterministic pseudo-random seed spread across the 1280x720 Pop-Urban district
	for i := 0; i < numAmbientParticles; i++ {
		fi := float32(i)
		var kind particleKind
		switch {
		case i < 22:
			kind = partSakuraPetal
		case i < 40:
			kind = partDustMote
		case i < 50:
			kind = partStoreSteam
		default:
			kind = partCanalSparkle
		}
		r.respawnParticle(i, kind, fi*0.37)
	}
}

func (r *Renderer) respawnParticle(idx int, kind particleKind, seed float32) {
	p := &r.particles[idx]
	p.kind = kind
	p.initialized = true
	p.phase = seed * 3.17

	// Simple deterministic hash in [0, 1)
	h1 := frac32(float32(math.Sin(float64(float32(idx)*12.9898+seed*78.233)) * 43758.5453))
	h2 := frac32(float32(math.Sin(float64(float32(idx)*39.346+seed*11.135)) * 24634.6345))

	switch kind {
	case partSakuraPetal:
		p.x = h1 * game.ArenaWidth
		p.y = h2 * (game.ArenaHeight - 60)
		p.vx = 14.0 + h1*14.0
		p.vy = 10.0 + h2*9.0
		p.maxLife = 7.0 + h1*5.0
		p.life = p.maxLife * (0.3 + 0.7*h2)
		p.size = 2.0
	case partDustMote:
		p.x = h1 * game.ArenaWidth
		p.y = 80.0 + h2*540.0
		p.vx = -6.0 + h1*12.0
		p.vy = -4.0 + h2*8.0
		p.maxLife = 5.0 + h2*4.0
		p.life = p.maxLife * (0.2 + 0.8*h1)
		p.size = 1.5
	case partStoreSteam:
		// Spawn near north storefront windows (Coffee Daily & Konbini)
		storeXs := [8]float32{80, 192, 368, 480, 800, 912, 1072, 1184}
		sx := storeXs[idx%len(storeXs)]
		p.x = sx + (h1-0.5)*46.0
		p.y = 66.0 + h2*18.0
		p.vx = (h2 - 0.5) * 6.0
		p.vy = -12.0 - h1*8.0
		p.maxLife = 2.2 + h2*1.4
		p.life = p.maxLife * (0.2 + 0.8*h1)
		p.size = 2.0
	case partCanalSparkle:
		p.x = h1 * game.ArenaWidth
		p.y = 682.0 + h2*34.0
		p.vx = 4.0 + h1*6.0
		p.vy = 0
		p.maxLife = 1.6 + h2*1.5
		p.life = p.maxLife * (0.2 + 0.8*h1)
		p.size = 2.0
	}
}

func (r *Renderer) updateAmbientParticles(dt float32) {
	for i := 0; i < numAmbientParticles; i++ {
		p := &r.particles[i]
		if !p.initialized {
			r.initAmbientParticles()
			break
		}
		p.life -= dt
		p.phase += dt * 2.4
		if p.life <= 0 {
			r.respawnParticle(i, p.kind, r.tick+float32(i)*0.19)
			continue
		}

		sway := float32(math.Sin(float64(p.phase))) * 8.0
		switch p.kind {
		case partSakuraPetal:
			p.x += (p.vx + sway) * dt
			p.y += (p.vy + float32(math.Cos(float64(p.phase*1.3)))*4.0) * dt
		case partDustMote:
			p.x += (p.vx + sway*0.5) * dt
			p.y += p.vy * dt
		case partStoreSteam:
			p.x += (p.vx + sway*0.4) * dt
			p.y += p.vy * dt
		case partCanalSparkle:
			p.x += p.vx * dt
		}

		if p.x > game.ArenaWidth+12 {
			p.x = -8
		} else if p.x < -12 {
			p.x = game.ArenaWidth + 8
		}
		if p.y > game.ArenaHeight+8 {
			p.y = 16
		}
	}
}

func (r *Renderer) drawAmbientParticles(dst *ebiten.Image, camX, camY float32) {
	for i := 0; i < numAmbientParticles; i++ {
		p := &r.particles[i]
		sx := p.x - camX
		sy := p.y - camY
		if sx < -8 || sx > game.ScreenWidth+8 || sy < -8 || sy > game.ScreenHeight+8 {
			continue
		}
		lifeFrac := clampF(p.life/p.maxLife, 0, 1)
		fade := float32(math.Sin(float64(lifeFrac) * math.Pi))

		switch p.kind {
		case partSakuraPetal:
			a := uint8(195 * fade)
			vector.DrawFilledRect(dst, sx, sy, 2.5, 1.8, color.RGBA{255, 172, 206, a}, false)
			vector.DrawFilledRect(dst, sx+1, sy+1, 1.5, 1.2, color.RGBA{255, 226, 238, a}, false)
		case partDustMote:
			a := uint8(165 * fade)
			vector.DrawFilledRect(dst, sx, sy, 1.8, 1.8, color.RGBA{255, 220, 145, a}, false)
		case partStoreSteam:
			a := uint8(110 * fade)
			vector.DrawFilledRect(dst, sx, sy, 2.2, 2.2, color.RGBA{255, 244, 224, a}, false)
		case partCanalSparkle:
			twinkle := float32(0.5 + 0.5*math.Sin(float64(p.phase*3.5)))
			a := uint8(210 * fade * twinkle)
			vector.DrawFilledRect(dst, sx-1, sy, 3.5, 1.2, color.RGBA{190, 248, 255, a}, false)
			vector.DrawFilledRect(dst, sx, sy-1, 1.2, 3.2, color.RGBA{255, 255, 255, a}, false)
		}
	}
}

// drawGroundShadows renders crisp Golden-Hour diagonal drop shadows (top-left sun -> down-right shadows)
// cast by every building, tree, parked car, bench, lamp, vending shop, and active Bus Stop.
func (r *Renderer) drawGroundShadows(dst *ebiten.Image, w *game.World, camX, camY float32) {
	sh := colGoldenHourShadow

	for i := range w.Props {
		pr := &w.Props[i]
		if !pr.CastShadow {
			continue
		}
		px := pr.X - camX
		py := pr.Y - camY
		if px+pr.W < -32 || px > game.ScreenWidth+32 || py+pr.H < -32 || py > game.ScreenHeight+32 {
			continue
		}

		switch pr.Kind {
		case game.PropBuildingKonbini, game.PropBuildingCoffee, game.PropBuildingArcade:
			// Crisp architectural down-right drop shadow spilled across the north promenade
			vector.DrawFilledRect(dst, px+8, py+pr.H-6, pr.W+4, 16, sh, false)
			vector.DrawFilledRect(dst, px+pr.W-2, py+12, 10, pr.H-8, sh, false)
		case game.PropTreeLush:
			cx := px + pr.W*0.5 + 7
			cy := pr.SortY - camY - 2
			vector.DrawFilledCircle(dst, cx, cy, 17, sh, true)
			vector.DrawFilledCircle(dst, cx+6, cy+3, 12, sh, true)
		case game.PropCarRed, game.PropCarYellow, game.PropCarMint:
			cx := px + (pr.W-48)*0.5
			vector.DrawFilledRect(dst, cx+5, py+8, 48, 26, sh, false)
		case game.PropStreetlamp:
			cx := px + pr.W*0.5
			by := pr.SortY - camY
			vector.DrawFilledRect(dst, cx+1, by-4, 14, 6, sh, false)
			vector.DrawFilledCircle(dst, cx+14, by-1, 4, sh, true)
		case game.PropBenchPlants:
			cx := px + pr.W*0.5
			by := pr.SortY - camY
			vector.DrawFilledRect(dst, cx-18, by-12, 42, 14, sh, false)
		case game.PropCatsClutter:
			cx := px + pr.W*0.5
			by := pr.SortY - camY
			vector.DrawFilledRect(dst, cx-10, by-8, 24, 10, sh, false)
		}
	}

	// Vending machine drop shadows
	for _, shop := range w.Shops {
		sx, sy := shop.X-camX, shop.Y-camY
		vector.DrawFilledRect(dst, sx-11, sy-8, 28, 24, sh, false)
	}

	// Active Bus Stop shelter architectural drop shadows
	for i := range w.Stops {
		st := &w.Stops[i]
		if !st.Active {
			continue
		}
		sx, sy := st.X-camX, st.Y-camY
		vector.DrawFilledRect(dst, sx-18, sy+6, 44, 16, sh, false)
		vector.DrawFilledRect(dst, sx+16, sy-12, 10, 32, sh, false)
	}
}

func (r *Renderer) drawAdditiveGlow(dst *ebiten.Image, sx, sy, radius float32, c color.RGBA, intensity float32) {
	if r.lightTex == nil || radius <= 1 || intensity <= 0.005 {
		return
	}
	if sx+radius < 0 || sx-radius > game.ScreenWidth || sy+radius < 0 || sy-radius > game.ScreenHeight {
		return
	}
	scale := float64(radius*2.0) / 64.0
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-32, -32)
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(float64(sx), float64(sy))
	op.Blend = ebiten.BlendLighter
	op.ColorScale.Scale(
		float32(c.R)/255.0*intensity,
		float32(c.G)/255.0*intensity,
		float32(c.B)/255.0*intensity,
		1.0,
	)
	dst.DrawImage(r.lightTex, op)
}

func (r *Renderer) drawAdditiveBeam(dst *ebiten.Image, sx, sy, rx, ry float32, c color.RGBA, intensity float32) {
	if r.lightTex == nil || rx <= 1 || ry <= 1 || intensity <= 0.005 {
		return
	}
	if sx+rx < 0 || sx-rx > game.ScreenWidth || sy+ry < 0 || sy-ry > game.ScreenHeight {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-32, -32)
	op.GeoM.Scale(float64(rx*2.0)/64.0, float64(ry*2.0)/64.0)
	op.GeoM.Translate(float64(sx), float64(sy))
	op.Blend = ebiten.BlendLighter
	op.ColorScale.Scale(
		float32(c.R)/255.0*intensity,
		float32(c.G)/255.0*intensity,
		float32(c.B)/255.0*intensity,
		1.0,
	)
	dst.DrawImage(r.lightTex, op)
}

// drawLightingPass renders all dynamic additive light pools, neon halos, bus headlights,
// bullet micro-glows, and police siren flashes.
func (r *Renderer) drawLightingPass(dst *ebiten.Image, w *game.World, camX, camY float32) {
	// 1. District Streetlamps, Shop Windows, Neon Signs, and Vending Machines
	for i := range w.Lights {
		l := &w.Lights[i]
		lx, ly := l.X-camX, l.Y-camY
		flicker := float32(0.94 + 0.06*math.Sin(float64(r.tick*3.2+l.X*0.03)))

		switch l.Kind {
		case game.LightStreetlamp:
			// Warm Golden-Cream #FFD88A pool on the sidewalk & around bulb
			warm := color.RGBA{255, 216, 138, 255}
			r.drawAdditiveGlow(dst, lx, ly+26, l.Radius*0.85, warm, 0.28*l.Intensity*flicker)
			r.drawAdditiveGlow(dst, lx, ly+4, 34, color.RGBA{255, 236, 185, 255}, 0.42*flicker)
		case game.LightShopWindow:
			warm := color.RGBA{255, 216, 138, 255}
			r.drawAdditiveBeam(dst, lx, ly+14, l.Radius*0.95, l.Radius*0.65, warm, 0.30*l.Intensity*flicker)
		case game.LightNeonSign:
			c := color.RGBA{R: l.R, G: l.G, B: l.B, A: 255}
			pulse := float32(0.88 + 0.12*math.Sin(float64(r.tick*4.5+l.X*0.02)))
			r.drawAdditiveGlow(dst, lx, ly, l.Radius*0.72, c, 0.24*l.Intensity*pulse)
		case game.LightVending:
			c := color.RGBA{R: l.R, G: l.G, B: l.B, A: 255}
			r.drawAdditiveGlow(dst, lx, ly, l.Radius*0.78, c, 0.28*l.Intensity)
		}
	}

	// 2. Active Bus Stops holographic glow
	for i := range w.Stops {
		st := &w.Stops[i]
		if !st.Active {
			continue
		}
		sx, sy := st.X-camX, st.Y-camY
		ac := archetypeColor(st.Arch)
		if st.SynergyActive {
			ac = colGold
		}
		pulse := float32(0.85 + 0.15*math.Sin(float64(r.tick*5.5)))
		r.drawAdditiveGlow(dst, sx, sy, 58, ac, 0.36*pulse)
		r.drawAdditiveGlow(dst, sx, sy, st.CaptureRadius*0.85, ac, 0.12*pulse)
	}

	// 3. Girls VIP golden aura & Boss energy aura
	vipCol := color.RGBA{255, 224, 90, 255}
	for i := range w.Girls {
		g := &w.Girls[i]
		if g.IsVIP {
			gx, gy := g.X-camX, g.Y-camY
			r.drawAdditiveGlow(dst, gx, gy, 34, vipCol, 0.38)
		}
	}
	if w.Boss.Active {
		bx, by := w.Boss.X-camX, w.Boss.Y-camY
		pulse := float32(0.85 + 0.15*math.Sin(float64(r.tick*6.5)))
		r.drawAdditiveGlow(dst, bx, by, 56, colKawaii, 0.42*pulse)
		r.drawAdditiveGlow(dst, bx, by, 30, colGold, 0.25*pulse)
	}

	// 4. Player Chronos Dash trail & active Chronos focus halo
	for _, af := range w.Player.Afterimages {
		if af.Alpha <= 0.04 {
			continue
		}
		r.drawAdditiveGlow(dst, af.X-camX, af.Y-camY, 26, colGyaru, af.Alpha*0.38)
	}
	if w.Player.ChronosActive {
		px, py := w.Player.X-camX, w.Player.Y-camY
		r.drawAdditiveGlow(dst, px, py, 46, colGyaru, 0.32)
	}

	// 5. Neon Bus twin forward headlight cones + underglow
	for _, b := range w.Buses {
		if !b.Active {
			continue
		}
		bx, by := b.X-camX, b.Y-camY
		headCol := color.RGBA{255, 236, 175, 255}
		// Twin forward headlight cones
		r.drawAdditiveBeam(dst, bx+108, by-9, 88, 22, headCol, 0.42)
		r.drawAdditiveBeam(dst, bx+108, by+9, 88, 22, headCol, 0.42)
		// Bright headlight bulbs
		r.drawAdditiveGlow(dst, bx+44, by-9, 18, color.RGBA{255, 250, 220, 255}, 0.65)
		r.drawAdditiveGlow(dst, bx+44, by+9, 18, color.RGBA{255, 250, 220, 255}, 0.65)
		// Neon chassis underglow
		r.drawAdditiveBeam(dst, bx, by, 68, 34, archetypeColor(b.Arch), 0.35)
	}

	// 6. Police Cruiser alternating Red/Blue sirens + 5* Helicopter Spotlight
	sirenRed := int(r.tick*14)%2 == 0
	for _, c := range w.Police.Cruisers {
		if !c.Active || !c.Dashing {
			continue
		}
		cx, cy := c.X-camX, c.Y-camY
		sc := color.RGBA{255, 45, 75, 255}
		if !sirenRed {
			sc = color.RGBA{65, 165, 255, 255}
		}
		r.drawAdditiveGlow(dst, cx, cy-6, 68, sc, 0.48)
		r.drawAdditiveBeam(dst, cx+36, cy, 48, 20, color.RGBA{240, 248, 255, 255}, 0.28)
	}
	if w.Police.Stars >= 5 && w.Police.Heli.Active {
		hx, hy := w.Police.Heli.X-camX, w.Police.Heli.Y-camY
		r.drawAdditiveGlow(dst, hx, hy, w.Police.Heli.Radius*1.15, color.RGBA{255, 245, 185, 255}, 0.25)
	}

	// 7. Bullet micro-glows (culled to viewport)
	for i := 0; i < w.Bullets.Count; i++ {
		b := &w.Bullets.Items[i]
		bx, by := b.X-camX, b.Y-camY
		if bx < -20 || bx > game.ScreenWidth+20 || by < -20 || by > game.ScreenHeight+20 {
			continue
		}
		var gc color.RGBA
		rad := float32(14)
		switch b.Kind {
		case game.BulletRazor:
			gc = color.RGBA{224, 82, 180, 255}
		case game.BulletProtractor:
			gc = color.RGBA{255, 225, 70, 255}
		case game.BulletTear:
			gc = color.RGBA{0, 235, 255, 255}
		case game.BulletFlash:
			gc = color.RGBA{255, 250, 140, 255}
		case game.BulletPolice:
			gc = color.RGBA{255, 70, 95, 255}
		case game.BulletCiabatta:
			gc = color.RGBA{255, 145, 75, 255}
			rad = 22
		default:
			gc = color.RGBA{255, 90, 120, 255}
		}
		r.drawAdditiveGlow(dst, bx, by, rad, gc, 0.26)
	}
}

func frac32(v float32) float32 {
	return v - float32(math.Floor(float64(v)))
}
