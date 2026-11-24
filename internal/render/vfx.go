package render

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"girlbusAG/internal/game"
)

const (
	maxFXParticles = 512
	maxFXFlashes   = 24
	numNeonRain    = 48
)

type fxShape uint8

const (
	fxShapePixel fxShape = iota
	fxShapeSpark
	fxShapeDiamond
	fxShapeShard
	fxShapeRing
	fxShapeHeart
)

type fxParticle struct {
	x, y       float32
	vx, vy     float32
	drag       float32
	gravity    float32
	size       float32
	life       float32
	maxLife    float32
	r, g, b, a uint8
	shape      fxShape
	additive   bool
	ground     bool
}

type fxLightFlash struct {
	x, y      float32
	radius    float32
	maxRadius float32
	life      float32
	maxLife   float32
	intensity float32
	r, g, b   uint8
}

type fxRainDrop struct {
	x, y        float32
	vx, vy      float32
	length      float32
	splashX     float32
	splashY     float32
	splashTimer float32
}

func (r *Renderer) fxNextRand() float32 {
	if r.fxRandState == 0 {
		r.fxRandState = 0x9E3779B9
	}
	x := r.fxRandState
	x ^= x << 13
	x ^= x >> 17
	x ^= x << 5
	r.fxRandState = x
	return float32(x&0xFFFFFF) / float32(0x1000000)
}

func (r *Renderer) fxRandRange(lo, hi float32) float32 {
	return lo + (hi-lo)*r.fxNextRand()
}

func (r *Renderer) spawnFXParticle(p fxParticle) {
	if p.maxLife <= 0 {
		p.maxLife = 0.25
	}
	if p.life <= 0 {
		p.life = p.maxLife
	}
	if r.fxParticleCount < maxFXParticles {
		r.fxParticles[r.fxParticleCount] = p
		r.fxParticleCount++
		return
	}
	// Recycle the oldest-feeling lowest-life slot in O(1) scan over small cap or index 0.
	victim := 0
	minFrac := float32(2.0)
	for i := 0; i < maxFXParticles; i += 8 {
		f := r.fxParticles[i].life / r.fxParticles[i].maxLife
		if f < minFrac {
			minFrac = f
			victim = i
		}
	}
	r.fxParticles[victim] = p
}

func (r *Renderer) spawnFXFlash(x, y, radius, intensity, life float32, c color.RGBA) {
	if radius <= 2 || intensity <= 0.01 || life <= 0.01 {
		return
	}
	f := fxLightFlash{
		x:         x,
		y:         y,
		radius:    radius * 0.45,
		maxRadius: radius,
		life:      life,
		maxLife:   life,
		intensity: intensity,
		r:         c.R,
		g:         c.G,
		b:         c.B,
	}
	if r.fxFlashCount < maxFXFlashes {
		r.fxFlashes[r.fxFlashCount] = f
		r.fxFlashCount++
		return
	}
	victim := 0
	minLife := r.fxFlashes[0].life
	for i := 1; i < maxFXFlashes; i++ {
		if r.fxFlashes[i].life < minLife {
			minLife = r.fxFlashes[i].life
			victim = i
		}
	}
	r.fxFlashes[victim] = f
}

func (r *Renderer) emitBurst(x, y float32, count int, minSpd, maxSpd, minLife, maxLife, size float32, c color.RGBA, shape fxShape, additive, ground bool, drag, gravity float32) {
	for i := 0; i < count; i++ {
		ang := float64(r.fxRandRange(0, 6.2831853))
		spd := r.fxRandRange(minSpd, maxSpd)
		life := r.fxRandRange(minLife, maxLife)
		r.spawnFXParticle(fxParticle{
			x:        x + r.fxRandRange(-3, 3),
			y:        y + r.fxRandRange(-3, 3),
			vx:       float32(math.Cos(ang)) * spd,
			vy:       float32(math.Sin(ang)) * spd,
			drag:     drag,
			gravity:  gravity,
			size:     size * r.fxRandRange(0.75, 1.30),
			life:     life,
			maxLife:  life,
			r:        c.R,
			g:        c.G,
			b:        c.B,
			a:        c.A,
			shape:    shape,
			additive: additive,
			ground:   ground,
		})
	}
}

func (r *Renderer) updateFX(dt float32) {
	if dt <= 0 {
		return
	}
	if r.damageFlashTimer > 0 {
		r.damageFlashTimer -= dt
		if r.damageFlashTimer < 0 {
			r.damageFlashTimer = 0
		}
	}

	for i := 0; i < r.fxParticleCount; {
		p := &r.fxParticles[i]
		p.life -= dt
		if p.life <= 0 {
			r.fxParticleCount--
			r.fxParticles[i] = r.fxParticles[r.fxParticleCount]
			continue
		}
		if p.drag > 0 {
			damp := clampF(1.0-p.drag*dt, 0.05, 1.0)
			p.vx *= damp
			p.vy *= damp
		}
		p.vy += p.gravity * dt
		p.x += p.vx * dt
		p.y += p.vy * dt
		i++
	}

	for i := 0; i < r.fxFlashCount; {
		f := &r.fxFlashes[i]
		f.life -= dt
		if f.life <= 0 {
			r.fxFlashCount--
			r.fxFlashes[i] = r.fxFlashes[r.fxFlashCount]
			continue
		}
		frac := 1.0 - clampF(f.life/f.maxLife, 0, 1)
		f.radius = f.maxRadius * (0.40 + 0.60*float32(math.Sqrt(float64(frac))))
		i++
	}
}

func (r *Renderer) initRainIfNeeded() {
	if r.fxRainInit {
		return
	}
	r.fxRainInit = true
	for i := 0; i < numNeonRain; i++ {
		d := &r.fxRain[i]
		d.x = r.fxRandRange(0, game.ScreenWidth)
		d.y = r.fxRandRange(0, game.ScreenHeight)
		d.vx = -65.0 - r.fxRandRange(0, 22.0)
		d.vy = 295.0 + r.fxRandRange(0, 75.0)
		d.length = 7.0 + r.fxRandRange(0, 5.0)
		d.splashTimer = 0
	}
}

// UpdateWorldVFX advances continuous combat emitters (footsteps, bus sparks, boss embers, neon rain).
func (r *Renderer) UpdateWorldVFX(w *game.World, dt float32) {
	if w == nil || dt <= 0 {
		return
	}
	r.initRainIfNeeded()

	// Subtle asphalt dust & wet puddle ripples beneath Alfredo while running.
	moving := math.Abs(float64(w.Player.VX)) > 20 || math.Abs(float64(w.Player.VY)) > 20
	if moving && w.Player.DashTimer <= 0 && w.Player.X >= 0 {
		r.footstepTimer -= dt
		if r.footstepTimer <= 0 {
			r.footstepTimer = 0.11
			col := color.RGBA{205, 195, 220, 130}
			if w.StreakCount >= 5 {
				col = color.RGBA{0, 245, 212, 165}
			}
			r.spawnFXParticle(fxParticle{
				x:        w.Player.X + r.fxRandRange(-4, 4),
				y:        w.Player.Y + 11 + r.fxRandRange(-1, 2),
				vx:       -w.Player.VX*0.18 + r.fxRandRange(-14, 14),
				vy:       r.fxRandRange(-18, -6),
				drag:     4.0,
				gravity:  22,
				size:     2.2,
				life:     0.26,
				maxLife:  0.26,
				r:        col.R,
				g:        col.G,
				b:        col.B,
				a:        col.A,
				shape:    fxShapePixel,
				additive: w.StreakCount >= 5,
				ground:   true,
			})
		}
	}

	// Continuous neon exhaust sparks trailing active Neon Buses.
	r.busEmitTimer -= dt
	if r.busEmitTimer <= 0 {
		r.busEmitTimer = 0.032
		trailCol := colGyaru
		sparkCol := colGold
		if r.busSkin == 1 {
			trailCol = colGold
			sparkCol = color.RGBA{255, 130, 65, 255}
		}
		for i := range w.Buses {
			b := &w.Buses[i]
			if !b.Active {
				continue
			}
			r.emitBurst(b.X-42, b.Y+r.fxRandRange(-6, 10), 2, 40, 115, 0.16, 0.34, 2.2, trailCol, fxShapeSpark, true, false, 3.0, 0)
			r.spawnFXParticle(fxParticle{
				x:        b.X - 44,
				y:        b.Y + 8,
				vx:       -140 + r.fxRandRange(-35, 20),
				vy:       r.fxRandRange(-28, 22),
				drag:     2.5,
				size:     2.6,
				life:     0.28,
				maxLife:  0.28,
				r:        sparkCol.R,
				g:        sparkCol.G,
				b:        sparkCol.B,
				a:        235,
				shape:    fxShapeDiamond,
				additive: true,
			})
		}
	}

	// Boss ember aura and Synergized Bus Stop ambient sparkles.
	r.bossEmberTimer -= dt
	if r.bossEmberTimer <= 0 {
		r.bossEmberTimer = 0.08
		if w.Boss.Active {
			ang := float64(r.fxRandRange(0, 6.2831853))
			rad := r.fxRandRange(14, 28)
			bx := w.Boss.X + float32(math.Cos(ang))*rad
			by := w.Boss.Y + float32(math.Sin(ang))*rad*0.7
			c := colKawaii
			if w.Boss.EnrageFlashTimer > 0 {
				c = colGold
			}
			r.spawnFXParticle(fxParticle{
				x:        bx,
				y:        by,
				vx:       r.fxRandRange(-16, 16),
				vy:       r.fxRandRange(-36, -18),
				drag:     1.2,
				size:     2.4,
				life:     0.48,
				maxLife:  0.48,
				r:        c.R,
				g:        c.G,
				b:        c.B,
				a:        210,
				shape:    fxShapeDiamond,
				additive: true,
			})
		}
		for i := range w.Stops {
			st := &w.Stops[i]
			if !st.Active || !st.SynergyActive {
				continue
			}
			ang := float64(r.fxRandRange(0, 6.2831853))
			rad := r.fxRandRange(st.CaptureRadius*0.35, st.CaptureRadius*0.92)
			r.spawnFXParticle(fxParticle{
				x:        st.X + float32(math.Cos(ang))*rad,
				y:        st.Y + float32(math.Sin(ang))*rad*0.75,
				vx:       0,
				vy:       r.fxRandRange(-24, -10),
				size:     2.0,
				life:     0.42,
				maxLife:  0.42,
				r:        colGold.R,
				g:        colGold.G,
				b:        colGold.B,
				a:        195,
				shape:    fxShapeDiamond,
				additive: true,
			})
		}
	}

	// Dynamic neon acid rain intensity driven by Wanted Heat, Tier, and Blackout.
	targetRain := float32(0)
	switch {
	case w.Police.Stars >= 5 || w.DistrictEvent.Active == game.DistrictEventBlackout:
		targetRain = 0.90
	case w.Police.Stars >= 3:
		targetRain = 0.58
	case w.DifficultyTier >= 2 || w.Police.Stars >= 2:
		targetRain = 0.28
	}
	if w.TutorialMode {
		targetRain = 0
	}
	r.rainIntensity += (targetRain - r.rainIntensity) * clampF(dt*2.2, 0, 1)

	if r.rainIntensity > 0.02 {
		for i := 0; i < numNeonRain; i++ {
			d := &r.fxRain[i]
			d.x += d.vx * dt
			d.y += d.vy * dt
			if d.splashTimer > 0 {
				d.splashTimer -= dt
			}
			if d.y > game.ScreenHeight+10 {
				if r.fxNextRand() < 0.35 {
					d.splashX = d.x
					d.splashY = game.ScreenHeight - r.fxRandRange(8, 120)
					d.splashTimer = 0.18
				}
				d.x = r.fxRandRange(0, game.ScreenWidth+90)
				d.y = -r.fxRandRange(6, 40)
			}
			if d.x < -20 {
				d.x += game.ScreenWidth + 30
			}
		}
	}
}

// EmitWorldEvent spawns crisp combat particles and transient dynamic light flashes for WorldEvents.
func (r *Renderer) EmitWorldEvent(ev game.WorldEvent, w *game.World) {
	if w == nil {
		return
	}
	px, py := w.Player.X, w.Player.Y

	switch ev.Kind {
	case game.EventThrow:
		sx, sy := px, py
		for i := len(w.Stops) - 1; i >= 0; i-- {
			if w.Stops[i].Active {
				sx, sy = w.Stops[i].X, w.Stops[i].Y
				break
			}
		}
		ac := archetypeColor(w.Player.SelectedStopArch)
		r.emitBurst(sx, sy, 10, 45, 130, 0.20, 0.38, 2.2, ac, fxShapeSpark, true, false, 3.8, 0)
		r.emitBurst(sx, sy+8, 6, 25, 65, 0.22, 0.36, 2.4, color.RGBA{220, 210, 235, 150}, fxShapePixel, false, true, 5.0, 0)
		r.spawnFXFlash(sx, sy, 85, 0.45, 0.22, ac)

	case game.EventPerfectDrop:
		sx, sy := px, py
		for i := len(w.Stops) - 1; i >= 0; i-- {
			if w.Stops[i].Active {
				sx, sy = w.Stops[i].X, w.Stops[i].Y
				break
			}
		}
		r.emitBurst(sx, sy, 18, 95, 225, 0.24, 0.48, 3.0, colGold, fxShapeSpark, true, false, 3.2, 0)
		r.emitBurst(sx, sy, 8, 45, 135, 0.28, 0.52, 2.8, actionCream, fxShapeDiamond, true, false, 2.6, -20)
		r.spawnFXFlash(sx, sy, 170, 0.85, 0.34, colGold)

	case game.EventHook:
		ac := archetypeColor(w.Player.SelectedStopArch)
		r.emitBurst(px, py, 8, 40, 105, 0.18, 0.34, 2.2, ac, fxShapeSpark, true, false, 3.5, 0)

	case game.EventMinigameHit, game.EventComboStep:
		c := colGyaru
		if ev.Kind == game.EventComboStep {
			c = colGold
		}
		r.emitBurst(px, py-10, 7, 35, 95, 0.18, 0.34, 2.2, c, fxShapeDiamond, true, false, 3.0, -40)

	case game.EventBusRush:
		r.emitBurst(px, py, 16, 75, 185, 0.22, 0.42, 2.8, actionTeal, fxShapeSpark, true, false, 3.2, 0)
		r.spawnFXFlash(px, py, 135, 0.60, 0.26, actionTeal)

	case game.EventGirlBoarded:
		count := min(24, 10+ev.Value*2)
		r.emitBurst(px, py-6, count/2, 40, 120, 0.30, 0.60, 2.8, colKawaii, fxShapeHeart, true, false, 2.2, -55)
		r.emitBurst(px, py-4, count/2, 55, 145, 0.25, 0.50, 2.4, colGold, fxShapeDiamond, true, false, 2.8, -35)
		r.spawnFXFlash(px, py, 125, 0.55, 0.28, colKawaii)

	case game.EventSynergy:
		ac := archetypeColor(game.Archetype(ev.Value))
		r.emitBurst(px, py, 12, 65, 175, 0.25, 0.50, 2.8, colGold, fxShapeDiamond, true, false, 2.6, -25)
		r.emitBurst(px, py, 10, 70, 165, 0.22, 0.42, 2.5, ac, fxShapeSpark, true, false, 3.0, 0)
		r.spawnFXFlash(px, py, 165, 0.72, 0.32, colGold)

	case game.EventDash:
		r.emitBurst(px, py+6, 9, 30, 90, 0.16, 0.30, 2.2, colGyaru, fxShapeSpark, true, false, 4.2, 0)

	case game.EventDashParry:
		r.emitBurst(px, py, 12, 110, 240, 0.18, 0.36, 3.0, colGyaru, fxShapeSpark, true, false, 3.5, 0)
		r.emitBurst(px, py, 6, 65, 150, 0.20, 0.38, 2.6, actionCream, fxShapeDiamond, true, false, 3.0, 0)
		r.spawnFXFlash(px, py, 115, 0.75, 0.24, colGyaru)

	case game.EventGraze:
		r.emitBurst(px, py, 5, 45, 105, 0.14, 0.24, 1.8, colGold, fxShapeSpark, true, false, 4.0, 0)

	case game.EventStopRecall:
		r.emitBurst(px, py, 16, 80, 195, 0.20, 0.40, 2.5, colGyaru, fxShapeShard, true, false, 3.4, 0)
		r.spawnFXFlash(px, py, 120, 0.62, 0.24, colGyaru)

	case game.EventStopShieldHit:
		count := 6
		rad := float32(65)
		if ev.Flag {
			count = 18
			rad = 110
			r.spawnFXFlash(px, py, rad, 0.68, 0.24, colGyaru)
		}
		r.emitBurst(px, py, count, 65, 165, 0.18, 0.36, 2.5, color.RGBA{175, 245, 255, 255}, fxShapeShard, true, false, 3.2, 40)

	case game.EventPlayerHit:
		r.damageFlashTimer = 0.28
		r.emitBurst(px, py, 16, 85, 200, 0.22, 0.44, 2.8, colRedAlert, fxShapeSpark, true, false, 3.0, 20)
		r.spawnFXFlash(px, py, 130, 0.75, 0.26, colRedAlert)

	case game.EventBossSpawn:
		bx, by := px, py
		if w.Boss.Active {
			bx, by = w.Boss.X, w.Boss.Y
		}
		r.emitBurst(bx, by, 28, 80, 240, 0.30, 0.60, 3.2, colKawaii, fxShapeShard, true, false, 2.5, 0)
		r.spawnFXFlash(bx, by, 230, 0.85, 0.42, colKawaii)

	case game.EventBossBreak:
		bx, by := px, py
		if w.Boss.Active {
			bx, by = w.Boss.X, w.Boss.Y
		}
		count := 32
		rad := float32(230)
		if ev.Flag {
			count = 52
			rad = 280
		}
		r.emitBurst(bx, by, count/2, 110, 290, 0.28, 0.58, 3.4, colGold, fxShapeSpark, true, false, 2.4, 0)
		r.emitBurst(bx, by, count/2, 85, 230, 0.30, 0.62, 3.0, colKawaii, fxShapeShard, true, false, 2.2, 30)
		r.spawnFXFlash(bx, by, rad, 0.95, 0.45, colGold)

	case game.EventBossDefeated:
		r.emitBurst(px, py, 24, 90, 240, 0.32, 0.66, 3.0, colGold, fxShapeDiamond, true, false, 2.2, -30)
		r.spawnFXFlash(px, py, 260, 0.90, 0.45, colGold)

	case game.EventHealPickup:
		mint := color.RGBA{100, 255, 198, 255}
		r.emitBurst(px, py, 14, 35, 105, 0.28, 0.52, 2.4, mint, fxShapeDiamond, true, false, 2.2, -65)
		r.spawnFXFlash(px, py, 95, 0.55, 0.26, mint)

	case game.EventChestOpened, game.EventRelicAcquired, game.EventLevelUp:
		r.emitBurst(px, py-4, 16, 65, 180, 0.28, 0.56, 2.8, colGold, fxShapeDiamond, true, false, 2.4, -45)
		r.emitBurst(px, py-4, 10, 80, 175, 0.22, 0.44, 2.5, actionCream, fxShapeSpark, true, false, 3.0, 0)
		r.spawnFXFlash(px, py, 165, 0.78, 0.36, colGold)

	case game.EventDistrictEventStart:
		c := colGold
		switch game.DistrictEventKind(ev.Value) {
		case game.DistrictEventFlashmob:
			c = archetypeColor(w.DistrictEvent.FlashmobArch)
		case game.DistrictEventBlackout:
			c = colGyaru
		}
		r.emitBurst(px, py, 20, 75, 200, 0.26, 0.52, 2.8, c, fxShapeDiamond, true, false, 2.6, 0)
		r.spawnFXFlash(px, py, 185, 0.72, 0.35, c)
	}
}

func (r *Renderer) drawFXParticlesGround(dst *ebiten.Image, camX, camY float32) {
	for i := 0; i < r.fxParticleCount; i++ {
		p := &r.fxParticles[i]
		if !p.ground {
			continue
		}
		sx, sy := p.x-camX, p.y-camY
		if sx < -12 || sx > game.ScreenWidth+12 || sy < -12 || sy > game.ScreenHeight+12 {
			continue
		}
		fade := clampF(p.life/p.maxLife, 0, 1)
		a := uint8(float32(p.a) * fade)
		vector.DrawFilledRect(dst, sx-p.size*0.5, sy-p.size*0.5, p.size, p.size, Premul(p.r, p.g, p.b, a), false)
	}
}

func (r *Renderer) drawFXParticlesAdditive(dst *ebiten.Image, camX, camY float32) {
	if r.whitePixel == nil {
		return
	}
	for i := 0; i < r.fxParticleCount; i++ {
		p := &r.fxParticles[i]
		if p.ground {
			continue
		}
		sx, sy := p.x-camX, p.y-camY
		if sx < -16 || sx > game.ScreenWidth+16 || sy < -16 || sy > game.ScreenHeight+16 {
			continue
		}
		fade := clampF(p.life/p.maxLife, 0, 1)
		a := float32(p.a) / 255.0 * fade
		rf := float32(p.r) / 255.0 * a
		gf := float32(p.g) / 255.0 * a
		bf := float32(p.b) / 255.0 * a

		r.fxOp.GeoM.Reset()
		r.fxOp.ColorScale.Reset()
		if p.additive {
			r.fxOp.Blend = ebiten.BlendLighter
			r.fxOp.ColorScale.Scale(rf, gf, bf, 1.0)
		} else {
			r.fxOp.Blend = ebiten.BlendSourceOver
			r.fxOp.ColorScale.Scale(rf, gf, bf, a)
		}

		switch p.shape {
		case fxShapeSpark:
			ang := math.Atan2(float64(p.vy), float64(p.vx))
			lenPx := max(2.5, p.size*(1.6+fade*1.8))
			thick := max(1.0, p.size*0.42)
			r.fxOp.GeoM.Translate(-0.5, -0.5)
			r.fxOp.GeoM.Scale(float64(lenPx), float64(thick))
			r.fxOp.GeoM.Rotate(ang)
			r.fxOp.GeoM.Translate(float64(sx), float64(sy))
			dst.DrawImage(r.whitePixel, &r.fxOp)

		case fxShapeDiamond:
			d := max(1.8, p.size*(0.75+0.45*fade))
			r.fxOp.GeoM.Translate(-0.5, -0.5)
			r.fxOp.GeoM.Scale(float64(d), float64(d))
			r.fxOp.GeoM.Rotate(math.Pi * 0.25)
			r.fxOp.GeoM.Translate(float64(sx), float64(sy))
			dst.DrawImage(r.whitePixel, &r.fxOp)

		case fxShapeShard:
			ang := math.Atan2(float64(p.vy), float64(p.vx)) + float64(p.life*8.0)
			r.fxOp.GeoM.Translate(-0.5, -0.5)
			r.fxOp.GeoM.Scale(float64(p.size*1.5), float64(p.size*0.75))
			r.fxOp.GeoM.Rotate(ang)
			r.fxOp.GeoM.Translate(float64(sx), float64(sy))
			dst.DrawImage(r.whitePixel, &r.fxOp)

		case fxShapeHeart:
			// Crisp 5x4 pixel-heart silhouette
			s := max(1.0, p.size*0.45)
			pts := [5][2]float32{{-1, -1}, {1, -1}, {-1, 0}, {1, 0}, {0, 1}}
			for _, pt := range pts {
				r.fxOp.GeoM.Reset()
				r.fxOp.GeoM.Scale(float64(s), float64(s))
				r.fxOp.GeoM.Translate(float64(sx+pt[0]*s), float64(sy+pt[1]*s))
				dst.DrawImage(r.whitePixel, &r.fxOp)
			}

		default:
			r.fxOp.GeoM.Translate(-0.5, -0.5)
			r.fxOp.GeoM.Scale(float64(p.size), float64(p.size))
			r.fxOp.GeoM.Translate(float64(sx), float64(sy))
			dst.DrawImage(r.whitePixel, &r.fxOp)
		}
	}
}

func (r *Renderer) drawNeonRain(dst *ebiten.Image, camX, camY float32) {
	if r.rainIntensity <= 0.02 || r.whitePixel == nil {
		return
	}
	_ = camX
	_ = camY
	inten := r.rainIntensity
	for i := 0; i < numNeonRain; i++ {
		d := &r.fxRain[i]
		ang := math.Atan2(float64(d.vy), float64(d.vx))
		a := 0.24 * inten
		r.fxOp.GeoM.Reset()
		r.fxOp.GeoM.Translate(-0.5, -0.5)
		r.fxOp.GeoM.Scale(float64(d.length), 1.0)
		r.fxOp.GeoM.Rotate(ang)
		r.fxOp.GeoM.Translate(float64(d.x), float64(d.y))
		r.fxOp.ColorScale.Reset()
		r.fxOp.Blend = ebiten.BlendLighter
		r.fxOp.ColorScale.Scale(0.35*a, 0.85*a, 1.0*a, 1.0)
		dst.DrawImage(r.whitePixel, &r.fxOp)

		if d.splashTimer > 0 {
			sf := clampF(d.splashTimer/0.18, 0, 1)
			rad := 1.5 + (1.0-sf)*3.5
			vector.StrokeCircle(dst, d.splashX, d.splashY, rad, 1.0, Premul(110, 235, 255, uint8(90*sf*inten)), false)
		}
	}
}
