package game

import (
	"math"
	"math/rand"
)

// Girl represents one of the 4 archetype NPCs (or VIP Off-Limits targets) in the arena.
type Girl struct {
	ID           int
	X, Y, VX, VY float32
	Arch         Archetype
	IsVIP        bool
	HookedByStop int // -1 if free
	ShootTimer   float32
	AnimTimer    float32
	EnragedTimer float32
	StunTimer    float32
	SpiralAngle  float32
}

// NewGirl creates a Girl spawned at (x, y) with initial randomized timers.
func NewGirl(id int, x, y float32, arch Archetype, isVIP bool, rng *rand.Rand) Girl {
	return Girl{
		ID:           id,
		X:            x,
		Y:            y,
		Arch:         arch,
		IsVIP:        isVIP,
		HookedByStop: -1,
		ShootTimer:   0.8 + rng.Float32()*1.4,
		SpiralAngle:  rng.Float32() * 6.283,
	}
}

// Enrage triggers a punitive bullet burst and temporary speed boost when a stop fails or escapes.
func (g *Girl) Enrage(playerX, playerY float32, bullets *BulletPool) {
	g.HookedByStop = -1
	g.EnragedTimer = 3.0
	dx, dy, _ := normalize2D(playerX-g.X, playerY-g.Y)
	baseAng := float32(math.Atan2(float64(dy), float64(dx)))
	for k := -2; k <= 2; k++ {
		ang := baseAng + float32(k)*0.24
		bullets.Spawn(Bullet{
			X:      g.X,
			Y:      g.Y,
			VX:     float32(math.Cos(float64(ang))) * 175,
			VY:     float32(math.Sin(float64(ang))) * 175,
			Radius: 4.5,
			MaxAge: 3.5,
			Kind:   BulletKind(g.Arch),
		})
	}
}

// Update advances movement, tether constraint, and archetype bullet-hell patterns.
func (g *Girl) Update(dt, timeScale float32, tier int, playerX, playerY float32, stops []BusStop, expressSlow bool, bullets *BulletPool) {
	effDT := dt * timeScale
	g.AnimTimer += effDT

	if g.EnragedTimer > 0 {
		g.EnragedTimer -= effDT
		if g.EnragedTimer < 0 {
			g.EnragedTimer = 0
		}
	}
	if g.StunTimer > 0 {
		g.StunTimer -= effDT
		if g.StunTimer < 0 {
			g.StunTimer = 0
		}
		g.VX = 0
		g.VY = 0
		return
	}

	dx, dy, d := normalize2D(playerX-g.X, playerY-g.Y)
	tierMult := float32(1.0) + float32(tier-1)*0.15

	if g.HookedByStop != -1 {
		// Find tethered BusStop and keep girl inside tether leash
		var stop *BusStop
		for i := range stops {
			if stops[i].Active && stops[i].ID == g.HookedByStop {
				stop = &stops[i]
				break
			}
		}
		if stop == nil {
			g.HookedByStop = -1
		} else {
			sdx, sdy, sd := normalize2D(stop.X-g.X, stop.Y-g.Y)
			maxLeash := stop.CaptureRadius * 0.78
			if sd > maxLeash {
				g.VX = sdx * 130
				g.VY = sdy * 130
			} else {
				// Gentle orbit around shelter
				g.VX = -sdy * 26
				g.VY = sdx * 26
			}
			g.X += g.VX * effDT
			g.Y += g.VY * effDT
		}
	} else {
		// Free roaming AI per archetype
		baseSpd := float32(52.0) * tierMult
		if g.IsVIP {
			baseSpd *= 1.15
		}
		if g.EnragedTimer > 0 {
			baseSpd *= 1.45
		}
		if expressSlow {
			baseSpd *= 0.75
		}

		preferredDist := float32(155.0)
		switch g.Arch {
		case ArchEmo:
			preferredDist = 130
		case ArchStudent:
			preferredDist = 195
		case ArchKawaii:
			preferredDist = 115
		case ArchGyaru:
			preferredDist = 240
		}

		var moveX, moveY float32
		if d > preferredDist+25 {
			moveX, moveY = dx, dy
		} else if d < preferredDist-25 {
			moveX, moveY = -dx*0.7, -dy*0.7
		} else {
			// Strafe / orbit
			perpSign := float32(1)
			if g.ID%2 == 1 {
				perpSign = -1
			}
			moveX, moveY = -dy*perpSign, dx*perpSign
		}

		g.VX = moveX * baseSpd
		g.VY = moveY * baseSpd
		g.X += g.VX * effDT
		g.Y += g.VY * effDT
	}

	g.X = clamp32(g.X, 24, ArenaWidth-24)
	g.Y = clamp32(g.Y, 28, ArenaHeight-24)

	// Bullet-hell shooting
	shootRate := float32(1.0)
	if g.EnragedTimer > 0 {
		shootRate = 1.75
	} else if g.HookedByStop != -1 {
		shootRate = 0.75
	}
	g.ShootTimer -= effDT * shootRate
	if g.ShootTimer <= 0 && d < 520 {
		g.firePattern(tier, playerX, playerY, bullets)
	}
}

func (g *Girl) firePattern(tier int, playerX, playerY float32, bullets *BulletPool) {
	dx, dy, _ := normalize2D(playerX-g.X, playerY-g.Y)
	aimAng := float32(math.Atan2(float64(dy), float64(dx)))

	switch g.Arch {
	case ArchEmo:
		g.ShootTimer = 2.4
		g.SpiralAngle += 0.75
		count := 2
		if g.IsVIP || tier >= 3 {
			count = 3
		}
		for i := 0; i < count; i++ {
			ang := g.SpiralAngle + float32(i)*(6.283/float32(count))
			bullets.Spawn(Bullet{
				X:      g.X,
				Y:      g.Y,
				VX:     float32(math.Cos(float64(ang))) * 95,
				VY:     float32(math.Sin(float64(ang))) * 95,
				Radius: 4.5,
				MaxAge: 3.4,
				Kind:   BulletRazor,
			})
		}

	case ArchStudent:
		g.ShootTimer = 2.1
		spread := 1
		if tier >= 2 || g.IsVIP {
			spread = 3
		}
		for k := -(spread / 2); k <= spread/2; k++ {
			ang := aimAng + float32(k)*0.22
			bullets.Spawn(Bullet{
				X:           g.X,
				Y:           g.Y,
				VX:          float32(math.Cos(float64(ang))) * 155,
				VY:          float32(math.Sin(float64(ang))) * 155,
				Radius:      4.0,
				MaxAge:      4.2,
				Kind:        BulletProtractor,
				BouncesLeft: 1,
			})
		}

	case ArchKawaii:
		g.ShootTimer = 2.2
		ringCount := 5
		if g.IsVIP || tier >= 3 {
			ringCount = 7
		}
		g.SpiralAngle += 0.4
		for i := 0; i < ringCount; i++ {
			ang := g.SpiralAngle + float32(i)*(6.283/float32(ringCount))
			bullets.Spawn(Bullet{
				X:           g.X,
				Y:           g.Y,
				VX:          float32(math.Cos(float64(ang))) * 118,
				VY:          float32(math.Sin(float64(ang))) * 118,
				Radius:      4.5,
				MaxAge:      4.5,
				Kind:        BulletTear,
				BouncesLeft: 1,
			})
		}

	case ArchGyaru:
		g.ShootTimer = 2.0
		// Trifurcated flash spread
		fans := 3
		if g.IsVIP {
			fans = 5
		}
		for k := -(fans / 2); k <= fans/2; k++ {
			ang := aimAng + float32(k)*0.26
			bullets.Spawn(Bullet{
				X:      g.X,
				Y:      g.Y,
				VX:     float32(math.Cos(float64(ang))) * 175,
				VY:     float32(math.Sin(float64(ang))) * 175,
				Radius: 4.0,
				MaxAge: 3.4,
				Kind:   BulletFlash,
			})
		}
	}
}
