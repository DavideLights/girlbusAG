package game

import "math"

// BossClone represents one of L'Amante's holographic decoys.
type BossClone struct {
	X, Y   float32
	Active bool
}

// BossEscort represents one of La Dirigente Aziendale's 4 orbiting intern bodyguards.
type BossEscort struct {
	X, Y   float32
	Angle  float32
	Active bool
}

// Boss represents any of the 6 Gatekeeper Bosses defeated via Boss 3-Stop Break.
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
	PendingKnockback bool

	// Boss-specific gimmick entities
	Clones           [2]BossClone
	Escorts          [4]BossEscort
	MartialZTLActive bool
	MartialZTLRect   Rect
}

// Spawn initializes a Boss at the arena perimeter with 3 Authority Shields.
func (b *Boss) Spawn(kind BossKind, spawnX, spawnY float32) {
	b.Active = true
	b.Kind = kind
	b.X = clamp32(spawnX, 48, ArenaWidth-48)
	b.Y = clamp32(spawnY, 80, BoardwalkRailY-28)
	b.VX = 0
	b.VY = 0
	b.Shields = 3
	b.MaxShields = 3
	b.PhaseTimer = 0
	b.ShootTimer = 1.4
	b.HookedByStop = -1
	b.EnrageFlashTimer = 0.8
	b.PatternStep = 0
	b.PendingKnockback = false

	b.Clones = [2]BossClone{}
	b.Escorts = [4]BossEscort{}
	b.MartialZTLActive = false
	b.MartialZTLRect = Rect{}

	switch kind {
	case BossBestFriend:
		b.Name = "La Migliore Amica"
	case BossLover:
		b.Name = "L'Amante"
	case BossProfessor:
		b.Name = "La Professoressa"
	case BossExecutive:
		b.Name = "La Dirigente"
	case BossRector:
		b.Name = "La Rettrice"
	case BossMom:
		b.Name = "La Mamma"
	}

	b.setupPhaseEntities()
}

func (b *Boss) setupPhaseEntities() {
	switch b.Kind {
	case BossLover:
		offsets := [2][2]float32{{-120, -45}, {120, 45}}
		for i := 0; i < 2; i++ {
			b.Clones[i] = BossClone{
				X:      clamp32(b.X+offsets[i][0], 48, ArenaWidth-48),
				Y:      clamp32(b.Y+offsets[i][1], 80, BoardwalkRailY-24),
				Active: true,
			}
		}
	case BossExecutive:
		for i := 0; i < 4; i++ {
			ang := float32(i) * (math.Pi * 0.5)
			b.Escorts[i] = BossEscort{
				X:      b.X + float32(math.Cos(float64(ang)))*50,
				Y:      b.Y + float32(math.Sin(float64(ang)))*50,
				Angle:  ang,
				Active: true,
			}
		}
	case BossRector:
		b.MartialZTLActive = true
		if b.Shields%2 == 1 {
			b.MartialZTLRect = Rect{X: 0, Y: 144, W: ArenaWidth * 0.5, H: 400}
		} else {
			b.MartialZTLRect = Rect{X: ArenaWidth * 0.5, Y: 144, W: ArenaWidth * 0.5, H: 400}
		}
	}
}

// ActiveEscortCount returns the number of active intern bodyguards orbiting La Dirigente.
func (b *Boss) ActiveEscortCount() int {
	n := 0
	for i := range b.Escorts {
		if b.Escorts[i].Active {
			n++
		}
	}
	return n
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
		b.Clones = [2]BossClone{}
		b.Escorts = [4]BossEscort{}
		b.MartialZTLActive = false
		return true
	}
	b.setupPhaseEntities()
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
		b.Y = clamp32(b.Y+b.VY*effDT, 68, BoardwalkRailY-20)
	}

	// Update L'Amante hologram clones orbiting player from opposite flanks
	if b.Kind == BossLover {
		for i := range b.Clones {
			cl := &b.Clones[i]
			if !cl.Active {
				continue
			}
			ang := b.PhaseTimer*0.85 + float32(i)*math.Pi + 1.1
			tx := clamp32(playerX+float32(math.Cos(float64(ang)))*175, 44, ArenaWidth-44)
			ty := clamp32(playerY+float32(math.Sin(float64(ang)))*145, 76, BoardwalkRailY-20)
			cdx, cdy, cd := normalize2D(tx-cl.X, ty-cl.Y)
			if cd > 4 {
				cl.X = clamp32(cl.X+cdx*92*effDT, 40, ArenaWidth-40)
				cl.Y = clamp32(cl.Y+cdy*92*effDT, 72, BoardwalkRailY-20)
			}
		}
	}

	// Update La Dirigente Aziendale orbiting intern escorts
	if b.Kind == BossExecutive {
		for i := range b.Escorts {
			es := &b.Escorts[i]
			if !es.Active {
				continue
			}
			es.Angle += 1.65 * effDT
			es.X = clamp32(b.X+float32(math.Cos(float64(es.Angle)))*50, 28, ArenaWidth-28)
			es.Y = clamp32(b.Y+float32(math.Sin(float64(es.Angle)))*50, 68, BoardwalkRailY-16)
		}
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
						gr.Y = clamp32(gr.Y+py*95*effDT, 28, BoardwalkRailY-12)
					}
				}
			}
		}
	}

	// Enrage fires faster as shields break
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
		if b.PatternStep%2 == 0 {
			b.PendingKnockback = true
		}
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

	case BossLover:
		b.ShootTimer = 1.50
		// Main boss fires 3-way converging lipstick blade fan
		for k := -1; k <= 1; k++ {
			ang := aimAng + float32(k)*0.22
			bullets.Spawn(Bullet{
				X:      b.X,
				Y:      b.Y,
				VX:     float32(math.Cos(float64(ang))) * 165,
				VY:     float32(math.Sin(float64(ang))) * 165,
				Radius: 4.5,
				MaxAge: 3.8,
				Kind:   BulletLipstick,
			})
		}
		// Active hologram clones fire converging cross-bolts
		for i := range b.Clones {
			cl := &b.Clones[i]
			if !cl.Active {
				continue
			}
			cdx, cdy, _ := normalize2D(playerX-cl.X, playerY-cl.Y)
			cAng := float32(math.Atan2(float64(cdy), float64(cdx)))
			for _, off := range [2]float32{-0.16, 0.16} {
				ang := cAng + off
				bullets.Spawn(Bullet{
					X:      cl.X,
					Y:      cl.Y,
					VX:     float32(math.Cos(float64(ang))) * 145,
					VY:     float32(math.Sin(float64(ang))) * 145,
					Radius: 4.0,
					MaxAge: 3.5,
					Kind:   BulletLipstick,
				})
			}
		}

	case BossProfessor:
		b.ShootTimer = 1.55
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

	case BossExecutive:
		b.ShootTimer = 1.60
		if b.PatternStep%2 == 1 {
			// Red-Tape Contract Corridor Walls flanking the aimed lane
			perpX := -float32(math.Sin(float64(aimAng)))
			perpY := float32(math.Cos(float64(aimAng)))
			vx := float32(math.Cos(float64(aimAng))) * 150
			vy := float32(math.Sin(float64(aimAng))) * 150
			for _, side := range [2]float32{-1, 1} {
				for step := 1; step <= 3; step++ {
					off := side * float32(34+step*22)
					bullets.Spawn(Bullet{
						X:      b.X + perpX*off,
						Y:      b.Y + perpY*off,
						VX:     vx,
						VY:     vy,
						Radius: 5.0,
						MaxAge: 3.8,
						Kind:   BulletContract,
					})
				}
			}
		} else {
			// Corporate briefcase 8-way radial burst
			count := 8
			offset := float32(b.PatternStep) * 0.25
			for i := 0; i < count; i++ {
				ang := offset + float32(i)*(6.283/float32(count))
				bullets.Spawn(Bullet{
					X:      b.X,
					Y:      b.Y,
					VX:     float32(math.Cos(float64(ang))) * 165,
					VY:     float32(math.Sin(float64(ang))) * 165,
					Radius: 5.0,
					MaxAge: 4.0,
					Kind:   BulletContract,
				})
			}
		}

	case BossRector:
		b.ShootTimer = 1.48
		if b.PatternStep%2 == 0 {
			// Orbital Academic Seal Cross-Lasers around player
			for k := 0; k < 2; k++ {
				ang := float32(k)*1.57 + float32(b.PatternStep)*0.24
				lx := float32(math.Cos(float64(ang))) * 320
				ly := float32(math.Sin(float64(ang))) * 320
				*telegraphs = append(*telegraphs, TelegraphLine{
					X1:       playerX - lx,
					Y1:       playerY - ly,
					X2:       playerX + lx,
					Y2:       playerY + ly,
					Timer:    0.75,
					MaxTimer: 0.75,
					Kind:     1,
				})
			}
		}
		// Concentric radial academic seal ring
		count := 12
		offset := float32(b.PatternStep) * 0.20
		for i := 0; i < count; i++ {
			ang := offset + float32(i)*(6.283/float32(count))
			bullets.Spawn(Bullet{
				X:      b.X,
				Y:      b.Y,
				VX:     float32(math.Cos(float64(ang))) * 150,
				VY:     float32(math.Sin(float64(ang))) * 150,
				Radius: 5.0,
				MaxAge: 4.2,
				Kind:   BulletSeal,
			})
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
			// "Torna a casa!" 360 shockwave ring
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
