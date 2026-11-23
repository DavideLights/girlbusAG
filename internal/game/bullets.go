package game

import "math"

// Bullet is a flat value-type projectile stored in BulletPool for Zero-GC updates.
type Bullet struct {
	X, Y        float32
	VX, VY      float32
	Radius      float32
	Age, MaxAge float32
	Kind        BulletKind
	BouncesLeft int
	Homing      bool
	Grazed      bool
	SplitDone   bool
}

// BulletPool holds up to MaxBullets (4096) pre-allocated projectiles with O(1) swap-removal.
type BulletPool struct {
	Items [MaxBullets]Bullet
	Count int
}

// Spawn inserts a bullet into the pre-allocated pool if capacity remains.
func (p *BulletPool) Spawn(b Bullet) {
	if p.Count >= MaxBullets {
		return
	}
	if b.MaxAge <= 0 {
		b.MaxAge = 6.0
	}
	if b.Radius <= 0 {
		b.Radius = 4.0
	}
	p.Items[p.Count] = b
	p.Count++
}

// RemoveAt deletes the bullet at index i in O(1) time without heap allocation.
func (p *BulletPool) RemoveAt(i int) {
	if i < 0 || i >= p.Count {
		return
	}
	p.Count--
	if i < p.Count {
		p.Items[i] = p.Items[p.Count]
	}
}

// ClearRadius removes all bullets within radius of (cx, cy) and returns the number cleared.
func (p *BulletPool) ClearRadius(cx, cy, radius float32) int {
	rSq := radius * radius
	cleared := 0
	for i := 0; i < p.Count; {
		b := &p.Items[i]
		if distSq(b.X, b.Y, cx, cy) <= rSq {
			p.RemoveAt(i)
			cleared++
			continue
		}
		i++
	}
	return cleared
}

// ClearLine removes all bullets within halfWidth of segment (x1, y1)-(x2, y2).
func (p *BulletPool) ClearLine(x1, y1, x2, y2, halfWidth float32) int {
	hwSq := halfWidth * halfWidth
	cleared := 0
	for i := 0; i < p.Count; {
		b := &p.Items[i]
		if pointToSegmentDistSq(b.X, b.Y, x1, y1, x2, y2) <= hwSq {
			p.RemoveAt(i)
			cleared++
			continue
		}
		i++
	}
	return cleared
}

// Update advances all active bullets, applying Chronos/Fog slow-downs, Kawaii umbrella
// deflections, Emo razor splits, Boomerang homing, wall bounces, and Bus Stop back-glass shields.
func (p *BulletPool) Update(dt, timeScale float32, playerX, playerY float32, stops []BusStop) {
	initialCount := p.Count
	for i := 0; i < initialCount; {
		b := &p.Items[i]
		localScale := timeScale

		// Bus stop environmental modifiers (Emo Melancholic Fog & Kawaii Umbrella)
		for s := range stops {
			st := &stops[s]
			if !st.Active {
				continue
			}
			dSq := distSq(b.X, b.Y, st.X, st.Y)
			if dSq <= st.CaptureRadius*st.CaptureRadius {
				if st.Arch == ArchEmo {
					localScale *= 0.70
				} else if st.Arch == ArchKawaii && b.VY > 0 && b.Y < st.Y && dSq <= (st.CaptureRadius*0.5)*(st.CaptureRadius*0.5) {
					// Purikura Umbrella deflects downward bullets upward.
					b.VY = -b.VY * 0.9
					b.Y = st.Y - float32(math.Sqrt(float64(dSq))) - 2
				}
			}
		}

		effDT := dt * localScale
		b.Age += effDT
		if b.Age >= b.MaxAge {
			p.RemoveAt(i)
			initialCount--
			continue
		}

		// Boomerang Ciabatta homing & return curve
		if b.Homing {
			speed := float32(math.Sqrt(float64(b.VX*b.VX + b.VY*b.VY)))
			if speed < 140 {
				speed = 185
			}
			if b.Age < b.MaxAge*0.48 {
				dx, dy, l := normalize2D(playerX-b.X, playerY-b.Y)
				if l > 1 {
					steer := float32(2.8) * effDT
					b.VX += dx * speed * steer
					b.VY += dy * speed * steer
					nvx, nvy, nl := normalize2D(b.VX, b.VY)
					if nl > 0 {
						b.VX = nvx * speed
						b.VY = nvy * speed
					}
				}
			} else if b.Age >= b.MaxAge*0.48 && b.Age < b.MaxAge*0.55 {
				// Gentle boomerang turn
				b.VX = -b.VX*0.92 + (-b.VY)*0.2
				b.VY = -b.VY*0.92 + (b.VX)*0.2
			}
		}

		// Emo Razor Cross split after 1.0s
		if b.Kind == BulletRazor && !b.SplitDone && b.Age >= 1.0 {
			b.SplitDone = true
			spd := float32(math.Sqrt(float64(b.VX*b.VX+b.VY*b.VY))) * 1.45
			if spd < 130 {
				spd = 130
			}
			baseAng := float32(math.Atan2(float64(b.VY), float64(b.VX)))
			for _, offset := range [2]float32{0.65, -0.65} {
				ang := baseAng + offset
				p.Spawn(Bullet{
					X:         b.X,
					Y:         b.Y,
					VX:        float32(math.Cos(float64(ang))) * spd,
					VY:        float32(math.Sin(float64(ang))) * spd,
					Radius:    b.Radius * 0.85,
					MaxAge:    2.6,
					Kind:      BulletRazor,
					SplitDone: true,
				})
			}
		}

		b.X += b.VX * effDT
		b.Y += b.VY * effDT

		// Bus Stop back-glass shield blocks bullets approaching from behind/top.
		blockedByShield := false
		for s := range stops {
			st := &stops[s]
			if !st.Active || st.ShieldHP <= 0 {
				continue
			}
			// Shield glass sits on the upper/back side of the shelter: [-26, +26] in X, [-28, -4] in Y.
			if b.VY > 0 &&
				b.X >= st.X-26 && b.X <= st.X+26 &&
				b.Y >= st.Y-28 && b.Y <= st.Y-4 {
				st.ShieldHP--
				blockedByShield = true
				break
			}
		}
		if blockedByShield {
			p.RemoveAt(i)
			initialCount--
			continue
		}

		// Arena wall bounce or out-of-bounds cleanup
		out := false
		if b.X < 4 {
			if b.BouncesLeft > 0 {
				b.X = 4
				b.VX = -b.VX
				b.BouncesLeft--
			} else if b.X < -32 {
				out = true
			}
		} else if b.X > ArenaWidth-4 {
			if b.BouncesLeft > 0 {
				b.X = ArenaWidth - 4
				b.VX = -b.VX
				b.BouncesLeft--
			} else if b.X > ArenaWidth+32 {
				out = true
			}
		}

		if !out {
			if b.Y < 4 {
				if b.BouncesLeft > 0 {
					b.Y = 4
					b.VY = -b.VY
					b.BouncesLeft--
				} else if b.Y < -32 {
					out = true
				}
			} else if b.Y > ArenaHeight-4 {
				if b.BouncesLeft > 0 {
					b.Y = ArenaHeight - 4
					b.VY = -b.VY
					b.BouncesLeft--
				} else if b.Y > ArenaHeight+32 {
					out = true
				}
			}
		}

		if out {
			p.RemoveAt(i)
			initialCount--
			continue
		}

		i++
	}
}
