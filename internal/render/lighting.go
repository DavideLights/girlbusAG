package render

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"girlbusAG/internal/game"
)

const (
	numAmbientParticles = 92
	radialLightSize     = 64
)

var colGoldenHourShadow = Premul(33, 22, 53, 77)

type particleKind uint8

const (
	partSakuraPetal particleKind = iota
	partDustMote
	partStoreSteam
	partCanalSparkle
	partLampMote
)

type ambientParticle struct {
	x, y         float32
	baseX, baseY float32
	vx, vy       float32
	phase        float32
	life         float32
	maxLife      float32
	size         float32
	kind         particleKind
	initialized  bool
}

var steamAnchors = [9][2]float32{
	{192, 66},   // Pizzeria oven vent
	{1072, 66},  // Coffee Daily espresso steam
	{1184, 66},  // Ramen broth steam
	{990, 102},  // North Street Food Cart
	{164, 560},  // South Waterfront Food Cart
	{144, 252},  // West-North manhole
	{422, 456},  // West-South manhole
	{852, 244},  // East-North manhole
	{1124, 452}, // East-South manhole
}

var lampMoteAnchors = [12][2]float32{
	{180, 96},
	{316, 96},
	{530, 96},
	{756, 96},
	{1036, 96},
	{1140, 96},
	{340, 356},
	{948, 356},
	{130, 552},
	{452, 552},
	{808, 552},
	{1122, 552},
}

func buildRadialLightTexture() *ebiten.Image {
	rgba := image.NewRGBA(image.Rect(0, 0, radialLightSize, radialLightSize))
	center := float64(radialLightSize-1) * 0.5
	maxR := float64(radialLightSize) * 0.5

	for y := 0; y < radialLightSize; y++ {
		for x := 0; x < radialLightSize; x++ {
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
		case i < 62:
			kind = partStoreSteam
		case i < 74:
			kind = partCanalSparkle
		default:
			kind = partLampMote
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
		anchor := steamAnchors[idx%len(steamAnchors)]
		p.baseX = anchor[0]
		p.baseY = anchor[1]
		p.x = anchor[0] + (h1-0.5)*16.0
		p.y = anchor[1] + (h2-0.5)*6.0
		p.vx = (h2 - 0.5) * 5.0
		p.vy = -10.0 - h1*7.5
		p.maxLife = 2.1 + h2*1.3
		p.life = p.maxLife * (0.15 + 0.85*h1)
		p.size = 2.2 + h1*1.4
	case partCanalSparkle:
		p.x = h1 * game.ArenaWidth
		p.y = 682.0 + h2*34.0
		p.vx = 4.0 + h1*6.0
		p.vy = 0
		p.maxLife = 1.6 + h2*1.5
		p.life = p.maxLife * (0.2 + 0.8*h1)
		p.size = 2.0
	case partLampMote:
		anchor := lampMoteAnchors[idx%len(lampMoteAnchors)]
		p.baseX = anchor[0]
		p.baseY = anchor[1]
		p.x = anchor[0] + (h1-0.5)*20.0
		p.y = anchor[1] + (h2-0.5)*18.0
		p.vx = 10.0 + h1*14.0 // orbit radius X
		p.vy = 6.0 + h2*10.0  // orbit radius Y
		p.maxLife = 4.5 + h2*3.5
		p.life = p.maxLife * (0.2 + 0.8*h1)
		p.size = 1.5
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
			p.x += (p.vx + sway*0.45) * dt
			p.y += p.vy * dt
		case partCanalSparkle:
			p.x += p.vx * dt
		case partLampMote:
			ang := float64(p.phase * 0.72)
			p.x = p.baseX + float32(math.Cos(ang))*p.vx
			p.y = p.baseY + 8.0 + float32(math.Sin(ang*1.35))*p.vy
		}

		if p.kind != partLampMote && p.kind != partStoreSteam {
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
}

func (r *Renderer) drawAmbientParticles(dst *ebiten.Image, camX, camY float32) {
	for i := 0; i < numAmbientParticles; i++ {
		p := &r.particles[i]
		sx := p.x - camX
		sy := p.y - camY
		if sx < -12 || sx > game.ScreenWidth+12 || sy < -12 || sy > game.ScreenHeight+12 {
			continue
		}
		lifeFrac := clampF(p.life/p.maxLife, 0, 1)
		fade := float32(math.Sin(float64(lifeFrac) * math.Pi))

		switch p.kind {
		case partSakuraPetal:
			a := uint8(195 * fade)
			vector.DrawFilledRect(dst, sx, sy, 2.5, 1.8, Premul(255, 172, 206, a), false)
			vector.DrawFilledRect(dst, sx+1, sy+1, 1.5, 1.2, Premul(255, 226, 238, a), false)
		case partDustMote:
			a := uint8(165 * fade)
			vector.DrawFilledRect(dst, sx, sy, 1.8, 1.8, Premul(255, 220, 145, a), false)
		case partStoreSteam:
			// Warm rising culinary & manhole steam wisps that gently expand as they rise
			age := 1.0 - lifeFrac
			wispR := p.size + age*2.8
			aOuter := uint8(62 * fade)
			aInner := uint8(92 * fade)
			vector.DrawFilledCircle(dst, sx, sy, wispR, Premul(255, 240, 218, aOuter), true)
			vector.DrawFilledRect(dst, sx-1.0, sy-1.0, 2.2, 2.0, Premul(255, 248, 232, aInner), false)
		case partCanalSparkle:
			twinkle := float32(0.5 + 0.5*math.Sin(float64(p.phase*3.5)))
			a := uint8(210 * fade * twinkle)
			vector.DrawFilledRect(dst, sx-1, sy, 3.5, 1.2, Premul(190, 248, 255, a), false)
			vector.DrawFilledRect(dst, sx, sy-1, 1.2, 3.2, Premul(255, 255, 255, a), false)
		case partLampMote:
			// Golden twilight light-motes orbiting gently inside streetlamp cones
			twinkle := float32(0.55 + 0.45*math.Sin(float64(p.phase*2.8)))
			a := uint8(205 * fade * twinkle)
			vector.DrawFilledCircle(dst, sx, sy, 2.2, Premul(255, 214, 128, a/2), true)
			vector.DrawFilledRect(dst, sx-0.7, sy-0.7, 1.5, 1.5, Premul(255, 244, 198, a), false)
		}
	}
}

// drawGroundShadows renders crisp Golden-Hour diagonal drop shadows (top-left sun -> down-right shadows)
// cast by every building, landmark, tree, parked vehicle, bench, lamp, vending shop, and active Bus Stop.
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
		case game.PropBuildingKonbini, game.PropBuildingCoffee, game.PropBuildingArcade,
			game.PropBuildingRamen, game.PropBuildingBookstore, game.PropBuildingMetro,
			game.PropBuildingPizzeria, game.PropBuildingVinylBar, game.PropBuildingGelateria:
			// The footprint has a tight contact edge and a stepped penumbra toward the road.
			by := pr.SortY - camY
			contact := Premul(27, 17, 44, 112)
			penumbra := Premul(47, 31, 68, 37)
			vector.DrawFilledRect(dst, px+5, by-6, pr.W-6, 8, contact, false)
			vector.DrawFilledRect(dst, px+9, by+2, pr.W-4, 5, sh, false)
			vector.DrawFilledRect(dst, px+13, by+7, pr.W-5, 4, penumbra, false)
			vector.DrawFilledRect(dst, px+pr.W-5, py+16, 8, pr.H-14, penumbra, false)
		case game.PropBuildingNewsstand, game.PropNeonBillboard, game.PropFoodCart, game.PropBusShelterDeco:
			vector.DrawFilledRect(dst, px+5, py+pr.H-5, pr.W+2, 11, sh, false)
		case game.PropFountain:
			cx := px + pr.W*0.5 + 4
			cy := pr.SortY - camY - 6
			vector.DrawFilledCircle(dst, cx, cy, 24, sh, true)
			vector.DrawFilledRect(dst, cx-26, cy-4, 54, 12, sh, false)
		case game.PropArchGate:
			// Ground shadows beneath the left and right arch pillars
			by := pr.SortY - camY
			vector.DrawFilledRect(dst, px+4, by-8, 18, 10, sh, false)
			vector.DrawFilledRect(dst, px+pr.W-20, by-8, 18, 10, sh, false)
		case game.PropPhoneBooth, game.PropVendingDrink:
			by := pr.SortY - camY
			vector.DrawFilledRect(dst, px+4, by-8, pr.W+2, 10, sh, false)
		case game.PropKioskUmbrella:
			cx := px + pr.W*0.5 + 5
			cy := pr.SortY - camY - 4
			vector.DrawFilledCircle(dst, cx, cy, 16, sh, true)
		case game.PropTreeLush:
			cx := px + pr.W*0.5 + 7
			cy := pr.SortY - camY - 2
			vector.DrawFilledCircle(dst, cx, cy, 17, sh, true)
			vector.DrawFilledCircle(dst, cx+6, cy+3, 12, sh, true)
		case game.PropCarRed, game.PropCarYellow, game.PropCarMint:
			cx := px + (pr.W-48)*0.5
			vector.DrawFilledRect(dst, cx+5, py+8, 48, 26, sh, false)
		case game.PropCarVanCyber:
			vector.DrawFilledRect(dst, px+6, py+10, pr.W-2, pr.H-6, sh, false)
		case game.PropScooter:
			vector.DrawFilledRect(dst, px+4, py+10, 28, 14, sh, false)
		case game.PropBikeRack, game.PropFlowerBox, game.PropPlanterNeon:
			vector.DrawFilledRect(dst, px+3, py+8, pr.W-2, pr.H-6, sh, false)
		case game.PropHydrantBin, game.PropZTLSign, game.PropBollardGlow:
			cx := px + pr.W*0.5
			by := pr.SortY - camY
			vector.DrawFilledRect(dst, cx-6, by-6, 15, 7, sh, false)
		case game.PropStreetlamp, game.PropTrafficLight:
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

	// Vending machine & Slot machine drop shadows
	for _, shop := range w.Shops {
		sx, sy := shop.X-camX, shop.Y-camY
		vector.DrawFilledRect(dst, sx-10, sy+9, 27, 7, sh, false)
		vector.DrawFilledRect(dst, sx-7, sy+16, 25, 3, Premul(47, 31, 68, 36), false)
	}
	for _, sm := range w.SlotMachines {
		sx, sy := sm.X-camX, sm.Y-camY
		vector.DrawFilledRect(dst, sx-10, sy+9, 27, 7, sh, false)
		vector.DrawFilledRect(dst, sx-7, sy+16, 25, 3, Premul(47, 31, 68, 36), false)
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
	r.drawAdditiveBeam(dst, sx, sy, radius, radius, c, intensity)
}

func (r *Renderer) drawAdditiveBeam(dst *ebiten.Image, sx, sy, rx, ry float32, c color.RGBA, intensity float32) {
	if r.lightTex == nil || rx <= 1 || ry <= 1 || intensity <= 0.005 {
		return
	}
	if sx+rx < 0 || sx-rx > game.ScreenWidth || sy+ry < 0 || sy-ry > game.ScreenHeight {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-radialLightSize/2, -radialLightSize/2)
	op.GeoM.Scale(float64(rx*2.0)/radialLightSize, float64(ry*2.0)/radialLightSize)
	op.GeoM.Translate(float64(sx), float64(sy))
	op.Blend = ebiten.BlendLighter
	op.Filter = ebiten.FilterLinear
	op.ColorScale.Scale(
		float32(c.R)/255.0*intensity,
		float32(c.G)/255.0*intensity,
		float32(c.B)/255.0*intensity,
		1.0,
	)
	dst.DrawImage(r.lightTex, op)
}

func brightenLightColor(c color.RGBA, boost float32) color.RGBA {
	mix := func(ch uint8) uint8 {
		v := float32(ch) + (255.0-float32(ch))*boost
		if v > 255 {
			v = 255
		}
		return uint8(v)
	}
	return color.RGBA{R: mix(c.R), G: mix(c.G), B: mix(c.B), A: 255}
}

// drawDistrictLighting projects the district lights onto the ground behind the world sprites.
func (r *Renderer) drawDistrictLighting(dst *ebiten.Image, w *game.World, camX, camY float32) {
	veil := uint8(22)
	if w.Police.Stars >= 5 {
		veil = 32
	}
	ambientScale := float32(1.0)
	if w.DistrictEvent.Active == game.DistrictEventBlackout {
		ambientScale = 0.22
	}
	vector.DrawFilledRect(dst, 0, 0, game.ScreenWidth, game.ScreenHeight, Premul(27, 21, 50, veil), false)
	r.drawAdditiveBeam(dst, 454-camX, 52-camY, 770, 245, color.RGBA{255, 172, 118, 255}, 0.064*ambientScale)
	r.drawAdditiveBeam(dst, 1000-camX, w.CanalY-camY, 720, 191, color.RGBA{104, 170, 229, 255}, 0.046*ambientScale)

	for i := range w.Lights {
		l := &w.Lights[i]
		lx, ly := l.X-camX, l.Y-camY
		// Broad viewport cull including downward pool / cone offsets
		if lx+l.Radius*1.4 < 0 || lx-l.Radius*1.4 > game.ScreenWidth ||
			ly+l.Radius*1.4+48 < 0 || ly-l.Radius*1.4-24 > game.ScreenHeight {
			continue
		}

		flicker := float32(0.97 + 0.03*math.Sin(float64(r.tick*2.1+l.X*0.037)))
		intensity := l.Intensity * ambientScale
		c := color.RGBA{R: l.R, G: l.G, B: l.B, A: 255}

		switch l.Kind {
		case game.LightStreetlamp:
			warmPool := color.RGBA{255, 210, 128, 255}
			r.drawAdditiveBeam(dst, lx+6, ly+33, l.Radius*0.93, l.Radius*0.43, warmPool, 0.22*intensity*flicker)
			r.drawAdditiveBeam(dst, lx+6, ly+32, l.Radius*0.43, l.Radius*0.20, color.RGBA{255, 224, 152, 255}, 0.14*intensity*flicker)
			r.drawAdditiveBeam(dst, lx, ly+15, 18, 30, warmPool, 0.12*intensity*flicker)
			r.drawAdditiveGlow(dst, lx, ly+2, 22, color.RGBA{255, 240, 195, 255}, 0.22*flicker*ambientScale)

		case game.LightShopWindow:
			r.drawAdditiveBeam(dst, lx+6, ly+23, l.Radius*0.90, l.Radius*0.40, c, 0.20*intensity*flicker)
			r.drawAdditiveBeam(dst, lx+4, ly+18, 31, 12, brightenLightColor(c, 0.30), 0.12*intensity*flicker)

		case game.LightNeonSign:
			r.drawAdditiveBeam(dst, lx, ly, l.Radius*0.71, l.Radius*0.33, c, 0.16*intensity*flicker)
			r.drawAdditiveBeam(dst, lx+5, ly+48, l.Radius*0.51, l.Radius*0.19, c, 0.13*intensity*flicker)
			for glint := 0; glint < 3; glint++ {
				step := float32(glint)
				width := l.Radius * (0.26 - 0.058*step)
				x := float32(math.Floor(float64(lx - width/2 + 4*step)))
				y := float32(math.Floor(float64(ly + 46 + step*4)))
				alpha := uint8((52 - step*10) * intensity * flicker)
				vector.DrawFilledRect(dst, x, y, width, 1, Premul(c.R, c.G, c.B, alpha), false)
			}

		case game.LightVending:
			r.drawAdditiveBeam(dst, lx, ly+17, l.Radius*0.46, l.Radius*0.21, c, 0.23*intensity*flicker)
			r.drawAdditiveBeam(dst, lx, ly-3, 22, 21, c, 0.14*intensity*flicker)

		case game.LightSlotMachine:
			pulse := float32(0.91 + 0.09*math.Sin(float64(r.tick*3.3+l.X*0.03)))
			r.drawAdditiveBeam(dst, lx, ly+19, l.Radius*0.49, l.Radius*0.24, color.RGBA{255, 89, 169, 255}, 0.19*intensity*pulse)
			r.drawAdditiveBeam(dst, lx, ly-7, 24, 14, c, 0.20*intensity*pulse)

		case game.LightLantern:
			swayX := float32(math.Sin(float64(r.tick*2.1+l.X*0.045))) * 1.2
			r.drawAdditiveBeam(dst, lx+swayX, ly+18, l.Radius*0.76, l.Radius*0.41, c, 0.22*intensity*flicker)
			r.drawAdditiveGlow(dst, lx+swayX, ly, 19, brightenLightColor(c, 0.22), 0.18*flicker*ambientScale)

		case game.LightFountain:
			ripples := float32(0.92 + 0.08*math.Sin(float64(r.tick*3.6)))
			r.drawAdditiveBeam(dst, lx, ly+6, l.Radius*0.88, l.Radius*0.44, c, 0.23*intensity*ripples)
			r.drawAdditiveBeam(dst, lx, ly+4, l.Radius*0.46, l.Radius*0.28, c, 0.18*intensity*ripples)

		case game.LightBollard:
			r.drawAdditiveBeam(dst, lx, ly+10, l.Radius*0.62, l.Radius*0.31, c, 0.19*intensity*flicker)
			r.drawAdditiveGlow(dst, lx, ly-1, 12, brightenLightColor(c, 0.25), 0.17*flicker*ambientScale)

		case game.LightTraffic:
			// Cycles smoothly between emerald-green, amber, and ruby-red lens bloom + crosswalk spill
			cycle := math.Mod(float64(r.tick*0.28+(l.X+l.Y)*0.0025), 3.0)
			var tc color.RGBA
			switch {
			case cycle < 1.35:
				tc = color.RGBA{56, 255, 148, 255} // Emerald green
			case cycle < 1.85:
				tc = color.RGBA{255, 198, 56, 255} // Warm amber
			default:
				tc = color.RGBA{255, 58, 86, 255} // Ruby red
			}
			r.drawAdditiveBeam(dst, lx, ly+26, l.Radius*0.58, l.Radius*0.31, tc, 0.15*intensity)
			r.drawAdditiveGlow(dst, lx, ly, 17, tc, 0.21*ambientScale)

		case game.LightVehicle:
			// Warm twin headlight spill cones OR soft ruby taillight wash
			isTaillight := l.R > 200 && l.G < 110
			if isTaillight {
				r.drawAdditiveBeam(dst, lx-10, ly+4, l.Radius*0.62, l.Radius*0.28, c, 0.23*intensity)
				r.drawAdditiveGlow(dst, lx, ly-3, 12, c, 0.23*ambientScale)
				r.drawAdditiveGlow(dst, lx, ly+5, 12, c, 0.23*ambientScale)
			} else {
				r.drawAdditiveBeam(dst, lx+20, ly-4, l.Radius*0.78, l.Radius*0.19, c, 0.23*intensity)
				r.drawAdditiveBeam(dst, lx+20, ly+6, l.Radius*0.78, l.Radius*0.19, c, 0.23*intensity)
			}
		}
	}

	r.drawCanalReflections(dst, w, camX, camY)
}

// drawBlackoutOverlay darkens the background arena during DistrictEventBlackout and projects
// boosted additive cones/auras around Alfredo, active Bus Stops, active Boss, and Police patrols.
func (r *Renderer) drawBlackoutOverlay(dst *ebiten.Image, w *game.World, camX, camY float32) {
	if w.DistrictEvent.Active != game.DistrictEventBlackout {
		return
	}

	vector.DrawFilledRect(dst, 0, 0, game.ScreenWidth, game.ScreenHeight, Premul(8, 5, 18, 138), false)

	// Boosted aura around Alfredo so player positioning and nearby dangers stay crystal clear.
	if w.Player.X >= 0 && w.Player.Y >= 0 {
		px, py := w.Player.X-camX, w.Player.Y-camY
		r.drawAdditiveGlow(dst, px, py, 92, color.RGBA{120, 235, 255, 255}, 0.34)
		r.drawAdditiveGlow(dst, px, py, 46, color.RGBA{255, 238, 195, 255}, 0.24)
	}

	// Boosted pools around active Bus Stops.
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
		r.drawAdditiveGlow(dst, sx, sy, st.CaptureRadius*1.15, ac, 0.28)
		r.drawAdditiveGlow(dst, sx, sy, 42, brightenLightColor(ac, 0.35), 0.24)
	}

	// Dramatic spotlight aura around active Gatekeeper Boss.
	if w.Boss.Active {
		bx, by := w.Boss.X-camX, w.Boss.Y-camY
		pulse := float32(0.88 + 0.12*math.Sin(float64(r.tick*6.5)))
		r.drawAdditiveGlow(dst, bx, by, 108, colKawaii, 0.36*pulse)
		r.drawAdditiveGlow(dst, bx, by, 54, colGold, 0.28*pulse)
	}

	// Patrol officers, rushing cruisers, and 5* chopper searchlight.
	for i := range w.Police.Officers {
		o := &w.Police.Officers[i]
		if !o.Active {
			continue
		}
		ox, oy := o.X-camX, o.Y-camY
		col := color.RGBA{95, 180, 255, 255}
		if o.IsSWAT {
			col = color.RGBA{255, 70, 105, 255}
		}
		r.drawAdditiveGlow(dst, ox, oy, 46, col, 0.26)
	}
	sirenMix := float32(0.5 + 0.5*math.Sin(float64(r.tick*12)))
	for _, c := range w.Police.Cruisers {
		if !c.Active {
			continue
		}
		cx, cy := c.X-camX, c.Y-camY
		red := color.RGBA{255, 55, 85, 255}
		blue := color.RGBA{75, 175, 255, 255}
		r.drawAdditiveGlow(dst, cx, cy, 74, blue, 0.24+0.14*(1-sirenMix))
		r.drawAdditiveGlow(dst, cx, cy, 62, red, 0.24+0.14*sirenMix)
		headDir := float32(1)
		if c.VX < 0 {
			headDir = -1
		}
		r.drawAdditiveBeam(dst, cx+headDir*56, cy, 72, 26, color.RGBA{240, 248, 255, 255}, 0.36)
	}
	if w.Police.Stars >= 5 && w.Police.Heli.Active {
		hx, hy := w.Police.Heli.X-camX, w.Police.Heli.Y-camY
		r.drawAdditiveGlow(dst, hx, hy, w.Police.Heli.Radius*1.30, color.RGBA{255, 245, 185, 255}, 0.36)
	}
}

// drawLightingPass adds the lights that follow moving entities and tactical effects.
func (r *Renderer) drawLightingPass(dst *ebiten.Image, w *game.World, camX, camY float32) {
	isBlackout := w.DistrictEvent.Active == game.DistrictEventBlackout
	stopBoost := float32(1.0)
	playerBoost := float32(1.0)
	bulletIntensity := float32(0.20)
	if isBlackout {
		stopBoost = 1.60
		playerBoost = 1.85
		bulletIntensity = 0.36
	}
	// Healing pickups share a mint and gold halo.
	for i := range w.HealPickups {
		hp := &w.HealPickups[i]
		hx, hy := float32(hp.X)-camX, float32(hp.Y)-camY
		if hx < -56 || hx > game.ScreenWidth+56 || hy < -56 || hy > game.ScreenHeight+56 {
			continue
		}
		pulse := float32(0.84 + 0.16*math.Sin(float64(r.tick*5.4)+hp.BobPhase))
		baseR := float32(34)
		if hp.Big {
			baseR = 48
		}
		r.drawAdditiveGlow(dst, hx, hy, baseR, color.RGBA{100, 255, 198, 255}, 0.22*pulse)
		r.drawAdditiveGlow(dst, hx, hy-2, baseR*0.56, color.RGBA{255, 242, 168, 255}, 0.22*pulse)
	}

	// 3b. Armored Neon Chests additive beacons & unlock bursts
	for i := range w.Chests {
		ch := &w.Chests[i]
		cx, cy := ch.X-camX, ch.Y-camY
		if cx < -96 || cx > game.ScreenWidth+96 || cy < -96 || cy > game.ScreenHeight+96 {
			continue
		}
		if ch.Opened {
			if ch.OpenAnimTimer > 0 {
				frac := clampF(ch.OpenAnimTimer/0.65, 0, 1)
				r.drawAdditiveGlow(dst, cx, cy-4, 84, colGold, 0.65*frac)
			}
			continue
		}
		if w.Player.BossKeys > 0 {
			pulse := float32(0.5 + 0.5*math.Sin(float64(r.tick*6.0+ch.BobPhase)))
			r.drawAdditiveGlow(dst, cx, cy-2, 64, colGold, 0.40+0.15*pulse)
			r.drawAdditiveBeam(dst, cx, cy-10, 20, 32, colGyaru, 0.16+0.10*pulse)
		} else {
			r.drawAdditiveGlow(dst, cx, cy, 42, colGyaru, 0.22)
		}
	}

	// 4. Active Bus Stops Soft Archetype Ground Glow + Beacon Pillar
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
		r.drawAdditiveBeam(dst, sx, sy+12, 62, 31, ac, 0.23*pulse*stopBoost)
		r.drawAdditiveBeam(dst, sx, sy-14, 13, 41, brightenLightColor(ac, 0.30), 0.15*pulse*stopBoost)
		r.drawAdditiveGlow(dst, sx, sy, st.CaptureRadius*0.88, ac, 0.10*pulse*stopBoost)
	}

	// 5. Girls VIP Golden Aura & Boss Dramatic Radial Aura
	vipCol := color.RGBA{255, 224, 90, 255}
	for i := range w.Girls {
		g := &w.Girls[i]
		if g.IsVIP {
			gx, gy := g.X-camX, g.Y-camY
			r.drawAdditiveBeam(dst, gx, gy+8, 33, 22, vipCol, 0.26)
		}
	}
	if w.Boss.Active {
		bx, by := w.Boss.X-camX, w.Boss.Y-camY
		pulse := float32(0.84 + 0.16*math.Sin(float64(r.tick*6.5)))
		r.drawAdditiveGlow(dst, bx, by, 72, colKawaii, 0.34*pulse)
		r.drawAdditiveBeam(dst, bx, by+14, 58, 30, colRedAlert, 0.28*pulse)
		r.drawAdditiveGlow(dst, bx, by, 34, colGold, 0.30*pulse)
	}

	// 6. Alfredo Subtle Focus Ring, Dash Trail Glow, Chronos Aura & Heal Flash Burst
	for _, af := range w.Player.Afterimages {
		if af.Alpha <= 0.04 {
			continue
		}
		r.drawAdditiveBeam(dst, af.X-camX, af.Y-camY+10, 26, 13, colGyaru, af.Alpha*0.20)
	}
	if w.Player.X >= 0 && w.Player.Y >= 0 {
		px, py := w.Player.X-camX, w.Player.Y-camY
		// Keep the light on the pavement under Alfredo's shoes.
		r.drawAdditiveBeam(dst, px, py+13, 24, 9, colGyaru, 0.12*playerBoost)
		if isBlackout {
			r.drawAdditiveGlow(dst, px, py, 44, color.RGBA{175, 245, 255, 255}, 0.22)
		}
		if w.Player.ChronosActive {
			r.drawAdditiveBeam(dst, px, py+14, 42, 15, colGyaru, 0.24*playerBoost)
		}
		if w.Player.HealFlashTimer > 0 {
			hf := clampF(float32(w.Player.HealFlashTimer/0.45), 0, 1)
			r.drawAdditiveBeam(dst, px, py+15, 34, 12, color.RGBA{90, 255, 185, 255}, 0.30*hf)
		}
	}

	// 7. Neon Bus Twin Volumetric Headlight Cones, Chassis Underglow & Glowing Magenta Slipstreams
	for _, sl := range w.Slipstreams {
		if sl.Life <= 0 || sl.MaxLife <= 0 {
			continue
		}
		alpha := clampF(sl.Life/sl.MaxLife, 0, 1)
		midX := (sl.X1+sl.X2)*0.5 - camX
		halfW := (sl.X2 - sl.X1) * 0.5
		sy := sl.Y - camY
		if halfW > 4 {
			r.drawAdditiveBeam(dst, midX, sy, halfW, 24, colKawaii, 0.24*alpha)
			r.drawAdditiveBeam(dst, midX, sy, halfW*0.85, 12, colGyaru, 0.18*alpha)
		}
	}
	for _, b := range w.Buses {
		if !b.Active {
			continue
		}
		bx, by := b.X-camX, b.Y-camY
		headCol := color.RGBA{255, 236, 175, 255}
		// Twin volumetric forward headlight cones
		r.drawAdditiveBeam(dst, bx+99, by-8, 71, 15, headCol, 0.33)
		r.drawAdditiveBeam(dst, bx+99, by+8, 71, 15, headCol, 0.33)
		// Bright headlight bulbs
		r.drawAdditiveGlow(dst, bx+44, by-9, 15, color.RGBA{255, 252, 224, 255}, 0.50)
		r.drawAdditiveGlow(dst, bx+44, by+9, 15, color.RGBA{255, 252, 224, 255}, 0.50)
		// Neon chassis underglow + magenta slipstream wake
		r.drawAdditiveBeam(dst, bx, by+7, 68, 24, archetypeColor(b.Arch), 0.26)
		r.drawAdditiveBeam(dst, bx-89, by, 66, 18, colKawaii, 0.24)
	}

	// Crossfade the cruiser lights so the wet road picks up both colors.
	sirenMix := float32(0.5 + 0.5*math.Sin(float64(r.tick*12)))
	for _, c := range w.Police.Cruisers {
		if !c.Active || !c.Dashing {
			continue
		}
		cx, cy := c.X-camX, c.Y-camY
		red := color.RGBA{255, 45, 75, 255}
		blue := color.RGBA{65, 165, 255, 255}
		r.drawAdditiveBeam(dst, cx-11, cy+10, 65, 24, red, 0.14+0.24*sirenMix)
		r.drawAdditiveBeam(dst, cx+11, cy+6, 65, 24, blue, 0.38-0.24*sirenMix)
		r.drawAdditiveGlow(dst, cx-5, cy-6, 21, red, 0.12+0.28*sirenMix)
		r.drawAdditiveGlow(dst, cx+5, cy-6, 21, blue, 0.40-0.28*sirenMix)
		headDir := float32(1)
		if c.VX < 0 {
			headDir = -1
		}
		r.drawAdditiveBeam(dst, cx+headDir*44, cy, 54, 22, color.RGBA{240, 248, 255, 255}, 0.32)
	}
	if w.Police.Stars >= 5 && w.Police.Heli.Active {
		hx, hy := w.Police.Heli.X-camX, w.Police.Heli.Y-camY
		r.drawAdditiveGlow(dst, hx, hy, w.Police.Heli.Radius*1.20, color.RGBA{255, 245, 185, 255}, 0.30)
	}

	// 9. Bullet Clean Crisp Luminance Halos (culled to viewport)
	for i := 0; i < w.Bullets.Count; i++ {
		b := &w.Bullets.Items[i]
		bx, by := b.X-camX, b.Y-camY
		if bx < -20 || bx > game.ScreenWidth+20 || by < -20 || by > game.ScreenHeight+20 {
			continue
		}
		var gc color.RGBA
		rad := float32(15)
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
			rad = 24
		case game.BulletLipstick:
			gc = color.RGBA{255, 52, 148, 255}
		case game.BulletContract:
			gc = color.RGBA{255, 235, 200, 255}
		case game.BulletSeal:
			gc = color.RGBA{255, 215, 70, 255}
		default:
			gc = color.RGBA{255, 90, 120, 255}
		}
		r.drawAdditiveGlow(dst, bx, by, rad, gc, bulletIntensity)
	}
}

// drawCanalReflections breaks the colored columns into narrow ripples below the embankment.
func (r *Renderer) drawCanalReflections(dst *ebiten.Image, w *game.World, camX, camY float32) {
	const canalBottom = float32(game.ArenaHeight)
	canalTop := w.CanalY + 8 // The canal-edge tile covers the first eight pixels past CanalY.
	if canalTop-camY >= game.ScreenHeight || canalBottom-camY <= 0 {
		return
	}
	ambientScale := float32(1.0)
	if w.DistrictEvent.Active == game.DistrictEventBlackout {
		ambientScale = 0.22
	}

	for i := range w.Lights {
		l := &w.Lights[i]
		if l.Y < w.BoardwalkY-96 {
			continue
		}
		lx := l.X - camX
		if lx < -96 || lx > game.ScreenWidth+96 {
			continue
		}

		rc := color.RGBA{R: l.R, G: l.G, B: l.B, A: 255}
		if l.Kind == game.LightStreetlamp {
			rc = color.RGBA{255, 214, 132, 255}
		}

		proximity := clampF(1-(canalTop-l.Y)/181, 0.22, 1)
		strength := l.Intensity * proximity * ambientScale
		reflectionTop := max(canalTop, l.Y+2)
		depth := canalBottom - reflectionTop
		if depth <= 4 {
			continue
		}
		washY := reflectionTop + depth*0.5 - camY
		r.drawAdditiveBeam(dst, lx, washY, l.Radius*0.39, depth*0.5, rc, strength*0.18)

		for band, offset := range [...]float32{0.12, 0.27, 0.46, 0.65, 0.87} {
			phase := float64(r.tick*2.8 + float32(band)*2.12 + l.X*0.042)
			pulse := float32(0.76 + 0.24*math.Cos(phase))
			width := l.Radius * (0.18 + offset*0.25) * pulse
			sy := float32(math.Floor(float64(reflectionTop + depth*offset - camY)))
			x := lx + float32(math.Sin(phase))*(2+8*offset) - width*0.5
			sx := float32(math.Floor(float64(x)))
			a := uint8(185 * strength * (1 - 0.65*offset) * pulse)
			vector.DrawFilledRect(dst, sx, sy, width, 1, Premul(rc.R, rc.G, rc.B, a), false)
		}
	}
}

func frac32(v float32) float32 {
	return v - float32(math.Floor(float64(v)))
}
