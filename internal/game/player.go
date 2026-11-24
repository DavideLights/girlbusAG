package game

// PlayerAfterimage stores one ghost trail snapshot for Alfredo's Katana-ZERO style Dash.
type PlayerAfterimage struct {
	X, Y    float32
	Alpha   float32
	FacingX float32
}

// Player represents Alfredo, the nimble on-foot protagonist.
type Player struct {
	X, Y, VX, VY float32
	FacingX      float32 // -1 or +1
	AnimTimer    float32
	AnimState    int // 0=Idle, 1=Run, 2=Dash, 3=Throw

	HP, MaxHP      int
	InvulnTimer    float32
	HealFlashTimer float64

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
	BossKeys               int

	SelectedStopArch Archetype
	UnlockedArchs    [ArchCount]bool
	Relics           [RelicCount]bool

	// Balanced Lv1 baseline stats that scale dramatically via In-Run Level-Up Upgrades
	StopCooldownBase     float32 // Starts at 6.5s, upgrades down to 1.5s
	StopCooldownTimer    float32
	MaxSimultaneousStops int     // Starts at 1, upgrades up to 4
	MaxMultiHook         int     // Starts at 1 girl per stop, upgrades up to 12
	RangeMult            float32 // Starts at 0.85, upgrades up to 1.60
	ChronosTimeScale     float32 // Starts at 0.55, upgrades down to 0.20 ("Katana Zero Time")
	BasslinePulseTimer   float32 // Emits passive rhythm ring every 2.5s

	// Combat status & game-feel timers
	ThrowAnimTimer       float32
	ThrowFacingX         float32
	StunTimer            float32
	SlowTimer            float32
	GroundedTimer        float32 // La Mamma "In Punizione!" temporary slot penalty
	SpeedMult            float32
	DashBufferTimer      float32
	ThrowBufferTimer     float32
	DashParryHitStopDone bool
	DashParryShaveUsed   float32
	ChronosToggledOn     bool
	SlipstreamTimer      float32
	SlipstreamPopupCD    float32
}

// NewPlayer initializes Alfredo at the center of the arena with Lv1 baseline stats.
func NewPlayer() Player {
	return Player{
		X:                    ArenaWidth * 0.5,
		Y:                    ArenaHeight * 0.5,
		FacingX:              1,
		HP:                   6,
		MaxHP:                6,
		ChronosMeter:         100.0,
		ChronosMax:           100.0,
		Level:                1,
		NextLevelXP:          50,
		SelectedStopArch:     ArchEmo,
		UnlockedArchs:        [ArchCount]bool{true, true, true, true},
		StopCooldownBase:     6.5,
		MaxSimultaneousStops: 1,
		MaxMultiHook:         1,
		RangeMult:            0.85,
		ChronosTimeScale:     0.55,
		BasslinePulseTimer:   BasslinePulsePeriod,
		SpeedMult:            1.0,
	}
}

// HasRelic reports whether Alfredo owns the specified passive relic this run.
func (p *Player) HasRelic(r RelicKind) bool {
	if r < 0 || r >= RelicCount {
		return false
	}
	return p.Relics[r]
}

// RelicCountOwned returns how many passive relics Alfredo has collected this run.
func (p *Player) RelicCountOwned() int {
	n := 0
	for _, has := range p.Relics {
		if has {
			n++
		}
	}
	return n
}

// EffectiveStopCooldown returns the base throw cooldown adjusted by passive relics.
func (p *Player) EffectiveStopCooldown() float32 {
	cd := p.StopCooldownBase
	if p.HasRelic(RelicNoiseCancel) {
		cd *= 0.85
	}
	if cd < 1.2 {
		cd = 1.2
	}
	return cd
}

// EffectiveRangeMult returns ThrowRange / CaptureRadius multiplier adjusted by relics.
func (p *Player) EffectiveRangeMult() float32 {
	rm := p.RangeMult
	if p.HasRelic(RelicMonthlyPass) {
		rm += 0.15
	}
	return rm
}

// EffectiveMaxMultiHook returns max girls hooked per stop including relics.
func (p *Player) EffectiveMaxMultiHook() int {
	return clampInt(p.MaxMultiHook, 1, 12)
}

// ShieldBonus returns extra BusStop back-glass shield pips from passive relics.
func (p *Player) ShieldBonus() int {
	bonus := 0
	if p.HasRelic(RelicMonthlyPass) {
		bonus++
	}
	if p.HasRelic(RelicArmoredGlass) {
		bonus += 3
	}
	return bonus
}

// MinigameSpeedMult returns minigame charge speed multiplier from passive relics.
func (p *Player) MinigameSpeedMult() float32 {
	if p.HasRelic(RelicDoubleEspresso) {
		return 1.25
	}
	return 1.0
}

// CashBonusMult returns passive cash multiplier from relics.
func (p *Player) CashBonusMult() float32 {
	m := float32(1.0)
	if p.HasRelic(RelicLuckySocks) {
		m += 0.15
	}
	if p.HasRelic(RelicFakeVIPPass) {
		m += 0.25
	}
	return m
}

// XPBonusMult returns passive XP multiplier from relics.
func (p *Player) XPBonusMult() float32 {
	if p.HasRelic(RelicFakeVIPPass) {
		return 1.25
	}
	return 1.0
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

// ReduceStopCooldown shaves sec seconds off the active Bus Stop throw cooldown.
func (p *Player) ReduceStopCooldown(sec float32) {
	if sec <= 0 || p.StopCooldownTimer <= 0 {
		return
	}
	p.StopCooldownTimer -= sec
	if p.StopCooldownTimer < 0 {
		p.StopCooldownTimer = 0
	}
}

// Update advances Alfredo's movement, dash, Chronos focus, cooldowns, and animations.
func (p *Player) Update(in InputState, dt float32, inExpressLane bool, inEmoFog bool, streakCount int) (basslineTriggered bool, dashed bool) {
	p.AnimTimer += dt

	if p.InvulnTimer > 0 {
		p.InvulnTimer -= dt
		if p.InvulnTimer < 0 {
			p.InvulnTimer = 0
		}
	}
	if p.HealFlashTimer > 0 {
		p.HealFlashTimer -= float64(dt)
		if p.HealFlashTimer < 0 {
			p.HealFlashTimer = 0
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
	if p.DashBufferTimer > 0 {
		p.DashBufferTimer -= dt
		if p.DashBufferTimer < 0 {
			p.DashBufferTimer = 0
		}
	}
	if p.ThrowBufferTimer > 0 {
		p.ThrowBufferTimer -= dt
		if p.ThrowBufferTimer < 0 {
			p.ThrowBufferTimer = 0
		}
	}
	if p.SlipstreamTimer > 0 {
		p.SlipstreamTimer -= dt
		if p.SlipstreamTimer < 0 {
			p.SlipstreamTimer = 0
		}
	}
	if p.SlipstreamPopupCD > 0 {
		p.SlipstreamPopupCD -= dt
		if p.SlipstreamPopupCD < 0 {
			p.SlipstreamPopupCD = 0
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

	// Archetype selection (keys 1-4 / wheel / tab)
	if in.SelectArchetype >= 0 && in.SelectArchetype < int(ArchCount) {
		if p.UnlockedArchs[in.SelectArchetype] {
			p.SelectedStopArch = Archetype(in.SelectArchetype)
		}
	}

	// Chronos ("Rizz-Focus") Hold or Toggle mode
	wantChronos := in.ChronosHeld
	if in.ChronosToggleMode {
		if in.ChronosJustPressed {
			p.ChronosToggledOn = !p.ChronosToggledOn
		}
		if p.ChronosMeter <= 1.0 || p.StunTimer > 0 {
			p.ChronosToggledOn = false
		}
		wantChronos = p.ChronosToggledOn
	} else {
		p.ChronosToggledOn = false
	}

	if wantChronos && p.ChronosMeter > 1.0 && p.StunTimer <= 0 {
		p.ChronosActive = true
		p.ChronosMeter -= 28.0 * dt
		if p.ChronosMeter <= 0 {
			p.ChronosMeter = 0
			p.ChronosActive = false
			p.ChronosToggledOn = false
		}
	} else {
		p.ChronosActive = false
		regenRate := float32(8.0)
		if inEmoFog {
			regenRate = 20.0
		}
		if p.HasRelic(RelicNoiseCancel) {
			regenRate *= 1.20
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
		return basslineTriggered, false
	}

	// Normalized movement vector
	mx, my, mLen := normalize2D(in.MoveX, in.MoveY)
	if mx < -0.1 {
		p.FacingX = -1
	} else if mx > 0.1 {
		p.FacingX = 1
	}

	// Dash input buffer (120 ms) + Dash trigger (with i-frames)
	if in.DashJustPressed {
		p.DashBufferTimer = InputBufferDuration
	}
	if p.DashBufferTimer > 0 && p.DashCooldown <= 0 && p.DashTimer <= 0 {
		p.DashBufferTimer = 0
		p.DashParryHitStopDone = false
		p.DashParryShaveUsed = 0
		p.DashTimer = 0.16
		p.DashCooldown = 0.72
		if mLen > 0.05 {
			p.DashDirX, p.DashDirY = mx, my
		} else {
			p.DashDirX, p.DashDirY = p.FacingX, 0
		}
		p.afterimageT = 0
		dashed = true
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
			for i := len(p.Afterimages) - 1; i > 0; i-- {
				p.Afterimages[i] = p.Afterimages[i-1]
			}
			facing := p.DashDirX
			if facing == 0 {
				facing = p.FacingX
			}
			p.Afterimages[0] = PlayerAfterimage{X: p.X, Y: p.Y, Alpha: 0.75, FacingX: facing}
		}
	} else {
		spdMult := p.SpeedMult
		if p.HasRelic(RelicDoubleEspresso) {
			spdMult += 0.18
		}
		baseSpeed := float32(170.0) * spdMult
		if inExpressLane {
			baseSpeed *= 1.25
		}
		if streakCount >= 3 {
			baseSpeed *= 1.10
		}
		if p.SlipstreamTimer > 0 {
			baseSpeed *= 1.45
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

	p.X = clamp32(p.X, 18, ArenaWidth-18)
	p.Y = clamp32(p.Y, 24, ArenaHeight-18)

	return basslineTriggered, dashed
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
	p.InvulnTimer = 1.15
	return true
}

// Heal restores HP up to MaxHP and triggers the near-player green HP bar flash.
func (p *Player) Heal(amount int) {
	if amount <= 0 {
		return
	}
	p.HP += amount
	if p.HP > p.MaxHP {
		p.HP = p.MaxHP
	}
	p.HealFlashTimer = 0.45
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
