package game

import "math"

// Boss represents one of the 3 Gatekeeper Bosses defeated via Boss 3-Stop Break.
type Boss struct {
	Active           bool
	Kind             BossKind
	Name             string
	X, Y, VX, VY     float32
	Shields          int
	MaxShields       int
	PhaseTimer       float32
	ShootTimer       float32
	HookedByStop     int
	EnrageFlashTimer float32
	PatternStep      int
}

// Spawn initializes a Boss at the arena perimeter with 3 Authority Shields.
func (b *Boss) Spawn(kind BossKind, spawnX, spawnY float32) {
	b.Active = true
	b.Kind = kind
	b.X = spawnX
	b.Y = spawnY
	b.VX = 0
	b.VY = 0
	b.Shields = 3
	b.MaxShields = 3
	b.PhaseTimer = 0
	b.ShootTimer = 1.4
	b.HookedByStop = -1
	b.EnrageFlashTimer = 0.8
	b.PatternStep = 0

	switch kind {
	case BossBestFriend:
		b.Name = "La Migliore Amica"
	case BossProfessor:
		b.Name = "La Professoressa"
	case BossMom:
		b.Name = "La Mamma"
	}
}

// BreakShield breaks 1 Boss Shield phase and returns true if all 3 shields are broken.
func (b *Boss) BreakShield() bool {
	if !b.Active {
		return false
	}
	b.Shields--
	b.HookedByStop = -1
	b.EnrageFlashTimer = 1.2
	b.ShootTimer = 0.8
	if b.Shields <= 0 {
		b.Shields = 0
		b.Active = false
		return true
	}
	return false
}

// Update advances Boss movement, special gimmicks, and bullet-hell patterns.
func (b *Boss) Update(dt, timeScale float32, playerX, playerY float32, stops []BusStop, girls []Girl, bullets *BulletPool, telegraphs *[]TelegraphLine) {
	if !b.Active {
		return
	}
	effDT := dt * timeScale
	b.PhaseTimer += effDT
	if b.EnrageFlashTimer > 0 {
		b.EnrageFlashTimer -= dt
		if b.EnrageFlashTimer < 0 {
			b.EnrageFlashTimer = 0
		}
	}

	// Keep Boss tethered gently if hooked by an active BusStop
	if b.HookedByStop != -1 {
		var stop *BusStop
		for i := range stops {
			if stops[i].Active && stops[i].ID == b.HookedByStop {
				stop = &stops[i]
				break
			}
		}
		if stop == nil {
			b.HookedByStop = -1
		} else {
			sdx, sdy, sd := normalize2D(stop.X-b.X, stop.Y-b.Y)
			if sd > stop.CaptureRadius*0.85 {
				b.X += sdx * 110 * effDT
				b.Y += sdy * 110 * effDT
			}
		}
	} else {
		// Free Boss movement
		dx, dy, d := normalize2D(playerX-b.X, playerY-b.Y)
		spd := float32(76.0) + float32(3-b.Shields)*16.0
		if b.Kind == BossBestFriend {
			// Tries to intercept active stops to cockblock!
			for s := range stops {
				if stops[s].Active {
					sdx, sdy, sd := normalize2D(stops[s].X-b.X, stops[s].Y-b.Y)
					if sd > 24 {
						dx, dy, d = sdx, sdy, sd
					}
					break
				}
			}
		}
		if d > 120 {
			b.VX = dx * spd
			b.VY = dy * spd
		} else {
			b.VX = -dy * (spd * 0.65)
			b.VY = dx * (spd * 0.65)
		}
		b.X = clamp32(b.X+b.VX*effDT, 36, ArenaWidth-36)
		b.Y = clamp32(b.Y+b.VY*effDT, 36, ArenaHeight-36)
	}

	// La Migliore Amica pushes unhooked girls away from active stops!
	if b.Kind == BossBestFriend {
		for s := range stops {
			st := &stops[s]
			if !st.Active {
				continue
			}
			if distSq(b.X, b.Y, st.X, st.Y) <= (st.CaptureRadius+60)*(st.CaptureRadius+60) {
				for g := range girls {
					gr := &girls[g]
					if gr.HookedByStop == -1 && distSq(gr.X, gr.Y, st.X, st.Y) <= (st.CaptureRadius+24)*(st.CaptureRadius+24) {
						px, py, _ := normalize2D(gr.X-st.X, gr.Y-st.Y)
						gr.X = clamp32(gr.X+px*95*effDT, 24, ArenaWidth-24)
						gr.Y = clamp32(gr.Y+py*95*effDT, 28, ArenaHeight-24)
					}
				}
			}
		}
	}

	// Enrage fires slightly faster as shields break
	rateMult := float32(1.0) + float32(3-b.Shields)*0.22
	b.ShootTimer -= effDT * rateMult
	if b.ShootTimer <= 0 {
		b.firePattern(playerX, playerY, bullets, telegraphs)
	}
}

func (b *Boss) firePattern(playerX, playerY float32, bullets *BulletPool, telegraphs *[]TelegraphLine) {
	b.PatternStep++
	dx, dy, _ := normalize2D(playerX-b.X, playerY-b.Y)
	aimAng := float32(math.Atan2(float64(dy), float64(dx)))

	switch b.Kind {
	case BossBestFriend:
		b.ShootTimer = 1.65
		// Heartbreak bouncing rings + aimed burst
		count := 10
		offset := float32(b.PatternStep) * 0.31
		for i := 0; i < count; i++ {
			ang := offset + float32(i)*(6.283/float32(count))
			bullets.Spawn(Bullet{
				X:           b.X,
				Y:           b.Y,
				VX:          float32(math.Cos(float64(ang))) * 135,
				VY:          float32(math.Sin(float64(ang))) * 135,
				Radius:      5.0,
				MaxAge:      4.2,
				Kind:        BulletTear,
				BouncesLeft: 1,
			})
		}

	case BossProfessor:
		b.ShootTimer = 1.55
		// Protractor Geometric Telegraph + dense chalk fan
		if b.PatternStep%2 == 0 {
			for k := 0; k < 2; k++ {
				ang := aimAng + float32(k)*1.57 + float32(b.PatternStep)*0.18
				lx := float32(math.Cos(float64(ang))) * 460
				ly := float32(math.Sin(float64(ang))) * 460
				*telegraphs = append(*telegraphs, TelegraphLine{
					X1:       b.X - lx,
					Y1:       b.Y - ly,
					X2:       b.X + lx,
					Y2:       b.Y + ly,
					Timer:    0.70,
					MaxTimer: 0.70,
					Kind:     1,
				})
				// Follow-up fast protractor bolts along the geometric line
				for _, sign := range [2]float32{-1, 1} {
					bullets.Spawn(Bullet{
						X:      b.X,
						Y:      b.Y,
						VX:     sign * float32(math.Cos(float64(ang))) * 225,
						VY:     sign * float32(math.Sin(float64(ang))) * 225,
						Radius: 4.5,
						MaxAge: 3.6,
						Kind:   BulletProtractor,
					})
				}
			}
		} else {
			for k := -3; k <= 3; k++ {
				ang := aimAng + float32(k)*0.18
				bullets.Spawn(Bullet{
					X:      b.X,
					Y:      b.Y,
					VX:     float32(math.Cos(float64(ang))) * 175,
					VY:     float32(math.Sin(float64(ang))) * 175,
					Radius: 4.5,
					MaxAge: 3.8,
					Kind:   BulletProtractor,
				})
			}
		}

	case BossMom:
		b.ShootTimer = 1.45
		if b.PatternStep%2 == 1 {
			// "La Ciabatta Boomerang a Ricerca"
			for _, off := range [2]float32{-0.28, 0.28} {
				ang := aimAng + off
				bullets.Spawn(Bullet{
					X:      b.X,
					Y:      b.Y,
					VX:     float32(math.Cos(float64(ang))) * 195,
					VY:     float32(math.Sin(float64(ang))) * 195,
					Radius: 9.5,
					MaxAge: 4.4,
					Kind:   BulletCiabatta,
					Homing: true,
				})
			}
		} else {
			// "Torna a casa!" 360° shockwave ring
			ring := 14
			offset := float32(b.PatternStep) * 0.22
			for i := 0; i < ring; i++ {
				ang := offset + float32(i)*(6.283/float32(ring))
				bullets.Spawn(Bullet{
					X:      b.X,
					Y:      b.Y,
					VX:     float32(math.Cos(float64(ang))) * 155,
					VY:     float32(math.Sin(float64(ang))) * 155,
					Radius: 5.0,
					MaxAge: 4.2,
					Kind:   BulletFlash,
				})
			}
		}
	}
}
