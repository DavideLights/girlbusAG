package game

// PlayerAfterimage stores one ghost trail snapshot for Alfredo's Katana-ZERO style Dash.
type PlayerAfterimage struct {
	X, Y  float32
	Alpha float32
}

// Player represents Alfredo, the nimble on-foot protagonist.
type Player struct {
	X, Y, VX, VY float32
	FacingX      float32 // -1 or +1
	AnimTimer    float32
	AnimState    int // 0=Idle, 1=Run, 2=Dash, 3=Throw

	HP, MaxHP   int
	InvulnTimer float32

	DashTimer    float32
	DashCooldown float32
	DashDirX     float32
	DashDirY     float32
	Afterimages  [4]PlayerAfterimage
	afterimageT  float32

	ChronosMeter  float32
	ChronosMax    float32
	ChronosActive bool

	XP, NextLevelXP, Level int
	Cash                   int

	SelectedStopArch Archetype
	UnlockedArchs    [ArchCount]bool

	// "Castrated" Lv1 stats that scale dramatically via In-Run Level-Up Upgrades
	StopCooldownBase     float32 // Starts at 7.0s, upgrades down to 1.5s
	StopCooldownTimer    float32
	MaxSimultaneousStops int     // Starts at 1, upgrades up to 4
	MaxMultiHook         int     // Starts at 1 girl per stop, upgrades up to 12
	RangeMult            float32 // Starts at 0.85, upgrades up to 1.60
	ChronosTimeScale     float32 // Starts at 0.55, upgrades down to 0.20 ("Katana Zero Time")
	BasslinePulseTimer   float32 // Emits passive rhythm ring every 2.5s

	// Combat status timers
	ThrowAnimTimer float32
	StunTimer      float32
	SlowTimer      float32
	GroundedTimer  float32 // La Mamma "In Punizione!" temporary slot penalty
	SpeedMult      float32
}

// NewPlayer initializes Alfredo at the center of the arena with Lv1 baseline stats.
func NewPlayer() Player {
	return Player{
		X:                    ArenaWidth * 0.5,
		Y:                    ArenaHeight * 0.5,
		FacingX:              1,
		HP:                   5,
		MaxHP:                5,
		ChronosMeter:         100.0,
		ChronosMax:           100.0,
		Level:                1,
		NextLevelXP:          50,
		SelectedStopArch:     ArchEmo,
		UnlockedArchs:        [ArchCount]bool{true, true, true, true},
		StopCooldownBase:     7.0,
		MaxSimultaneousStops: 1,
		MaxMultiHook:         1,
		RangeMult:            0.85,
		ChronosTimeScale:     0.55,
		BasslinePulseTimer:   BasslinePulsePeriod,
		SpeedMult:            1.0,
	}
}

// EffectiveMaxStops returns current allowed simultaneous stops, factoring in La Mamma's Grounded debuff.
func (p *Player) EffectiveMaxStops() int {
	maxS := p.MaxSimultaneousStops
	if p.GroundedTimer > 0 && maxS > 1 {
		maxS--
	}
	if maxS < 1 {
		maxS = 1
	}
	return maxS
}

// IsInvulnerable reports whether Alfredo currently has i-frames from Dash or damage recovery.
func (p *Player) IsInvulnerable() bool {
	return p.DashTimer > 0 || p.InvulnTimer > 0
}

// Update advances Alfredo's movement, dash, Chronos focus, cooldowns, and animations.
func (p *Player) Update(in InputState, dt float32, inExpressLane bool, inEmoFog bool) (basslineTriggered bool) {
	p.AnimTimer += dt

	if p.InvulnTimer > 0 {
		p.InvulnTimer -= dt
		if p.InvulnTimer < 0 {
			p.InvulnTimer = 0
		}
	}
	if p.DashCooldown > 0 {
		p.DashCooldown -= dt
		if p.DashCooldown < 0 {
			p.DashCooldown = 0
		}
	}
	if p.StopCooldownTimer > 0 {
		p.StopCooldownTimer -= dt
		if p.StopCooldownTimer < 0 {
			p.StopCooldownTimer = 0
		}
	}
	if p.ThrowAnimTimer > 0 {
		p.ThrowAnimTimer -= dt
		if p.ThrowAnimTimer < 0 {
			p.ThrowAnimTimer = 0
		}
	}
	if p.StunTimer > 0 {
		p.StunTimer -= dt
		if p.StunTimer < 0 {
			p.StunTimer = 0
		}
	}
	if p.SlowTimer > 0 {
		p.SlowTimer -= dt
		if p.SlowTimer < 0 {
			p.SlowTimer = 0
		}
	}
	if p.GroundedTimer > 0 {
		p.GroundedTimer -= dt
		if p.GroundedTimer < 0 {
			p.GroundedTimer = 0
		}
	}

	// Fade dash afterimages
	for i := range p.Afterimages {
		if p.Afterimages[i].Alpha > 0 {
			p.Afterimages[i].Alpha -= dt * 3.8
			if p.Afterimages[i].Alpha < 0 {
				p.Afterimages[i].Alpha = 0
			}
		}
	}

	// Archetype selection (keys 1-4)
	if in.SelectArchetype >= 0 && in.SelectArchetype < int(ArchCount) {
		if p.UnlockedArchs[in.SelectArchetype] {
			p.SelectedStopArch = Archetype(in.SelectArchetype)
		}
	}

	// Chronos ("Rizz-Focus") drain & regen
	if in.ChronosHeld && p.ChronosMeter > 1.0 && p.StunTimer <= 0 {
		p.ChronosActive = true
		p.ChronosMeter -= 28.0 * dt
		if p.ChronosMeter <= 0 {
			p.ChronosMeter = 0
			p.ChronosActive = false
		}
	} else {
		p.ChronosActive = false
		regenRate := float32(8.0)
		if inEmoFog {
			regenRate = 20.0
		}
		p.ChronosMeter += regenRate * dt
		if p.ChronosMeter > p.ChronosMax {
			p.ChronosMeter = p.ChronosMax
		}
	}

	// Passive Bassline Pulse (PokéBonk crowd-control rhythm ring)
	p.BasslinePulseTimer -= dt
	if p.BasslinePulseTimer <= 0 {
		p.BasslinePulseTimer = BasslinePulsePeriod
		basslineTriggered = true
	}

	// Stunned state halts movement
	if p.StunTimer > 0 {
		p.VX = 0
		p.VY = 0
		p.AnimState = AnimIdle
		return basslineTriggered
	}

	// Normalized movement vector
	mx, my, mLen := normalize2D(in.MoveX, in.MoveY)
	if mx < -0.1 {
		p.FacingX = -1
	} else if mx > 0.1 {
		p.FacingX = 1
	}

	// Dash trigger (with i-frames)
	if in.DashJustPressed && p.DashCooldown <= 0 && p.DashTimer <= 0 {
		p.DashTimer = 0.16
		p.DashCooldown = 0.72
		if mLen > 0.05 {
			p.DashDirX, p.DashDirY = mx, my
		} else {
			p.DashDirX, p.DashDirY = p.FacingX, 0
		}
		p.afterimageT = 0
	}

	if p.DashTimer > 0 {
		p.DashTimer -= dt
		dashSpeed := float32(490.0)
		p.VX = p.DashDirX * dashSpeed
		p.VY = p.DashDirY * dashSpeed
		p.X += p.VX * dt
		p.Y += p.VY * dt
		p.AnimState = AnimDash

		p.afterimageT -= dt
		if p.afterimageT <= 0 {
			p.afterimageT = 0.032
			// Shift afterimages
			for i := len(p.Afterimages) - 1; i > 0; i-- {
				p.Afterimages[i] = p.Afterimages[i-1]
			}
			p.Afterimages[0] = PlayerAfterimage{X: p.X, Y: p.Y, Alpha: 0.75}
		}
	} else {
		baseSpeed := float32(170.0) * p.SpeedMult
		if inExpressLane {
			baseSpeed *= 1.25
		}
		if p.SlowTimer > 0 {
			baseSpeed *= 0.68
		}
		if p.GroundedTimer > 0 {
			baseSpeed *= 0.88
		}
		p.VX = mx * baseSpeed
		p.VY = my * baseSpeed
		p.X += p.VX * dt
		p.Y += p.VY * dt

		if p.ThrowAnimTimer > 0 {
			p.AnimState = AnimThrow
		} else if mLen > 0.05 {
			p.AnimState = AnimRun
		} else {
			p.AnimState = AnimIdle
		}
	}

	// Keep Alfredo inside sidewalk/arena bounds
	p.X = clamp32(p.X, 18, ArenaWidth-18)
	p.Y = clamp32(p.Y, 24, ArenaHeight-18)

	return basslineTriggered
}

// TakeDamage applies damage if Alfredo is not invulnerable; returns true if hit landed.
func (p *Player) TakeDamage(amount int) bool {
	if p.IsInvulnerable() || amount <= 0 {
		return false
	}
	p.HP -= amount
	if p.HP < 0 {
		p.HP = 0
	}
	p.InvulnTimer = 0.90
	return true
}

// Heal restores HP up to MaxHP.
func (p *Player) Heal(amount int) {
	if amount <= 0 {
		return
	}
	p.HP += amount
	if p.HP > p.MaxHP {
		p.HP = p.MaxHP
	}
}

// AddXP adds experience and returns true if a Level-Up threshold was crossed.
func (p *Player) AddXP(amount int) bool {
	if amount <= 0 {
		return false
	}
	p.XP += amount
	if p.XP >= p.NextLevelXP {
		p.XP -= p.NextLevelXP
		p.Level++
		p.NextLevelXP = 45 + p.Level*38
		return true
	}
	return false
}
