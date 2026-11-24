package game

import (
	"math"
	"math/rand"
)

// PoliceOfficer represents a foot patrol or SWAT officer spawned at Wanted 1★+.
type PoliceOfficer struct {
	ID         int
	X, Y       float32
	VX, VY     float32
	ShootTimer float32
	HP         int
	IsSWAT     bool
	Active     bool
}

// PoliceCruiser represents a patrol car rushing across a horizontal street lane after a 0.8s telegraph.
type PoliceCruiser struct {
	Active         bool
	X, Y           float32
	VX             float32
	TelegraphTimer float32
	Dashing        bool
	ShootTimer     float32
}

// HelicopterSpotlight tracks Alfredo at 5★ MAX HEAT and rains danmaku when locked on.
type HelicopterSpotlight struct {
	Active     bool
	X, Y       float32
	Radius     float32
	ShootTimer float32
}

// PoliceHazard represents a 3★+ horizontal laser/spike road barrier strip.
type PoliceHazard struct {
	Active         bool
	X1, Y1, X2, Y2 float32
	TelegraphTimer float32
	LifeTimer      float32
}

// PoliceSystem manages GTA-style Wanted Stars (0..5★), multipliers, patrols, cruisers, hazards, and the 5★ chopper.
type PoliceSystem struct {
	Stars          int
	HeatDecayTimer float32
	SpawnTimer     float32
	CruiserTimer   float32
	HazardTimer    float32
	Officers       []PoliceOfficer
	Cruisers       []PoliceCruiser
	Hazards        []PoliceHazard
	Heli           HelicopterSpotlight
	BossFightMode  bool
	nextOfficerID  int
}

// NewPoliceSystem creates a clean 0★ PoliceSystem.
func NewPoliceSystem() PoliceSystem {
	return PoliceSystem{
		Officers:      make([]PoliceOfficer, 0, 16),
		Cruisers:      make([]PoliceCruiser, 0, 6),
		Hazards:       make([]PoliceHazard, 0, 4),
		Heli:          HelicopterSpotlight{Radius: 96.0},
		SpawnTimer:    2.0,
		CruiserTimer:  4.0,
		HazardTimer:   4.5,
		nextOfficerID: 1,
	}
}

// AddStar increases Wanted stars up to 5★ and resets the HeatDecayTimer.
func (ps *PoliceSystem) AddStar(n int) int {
	if n <= 0 {
		return 0
	}
	prev := ps.Stars
	ps.Stars += n
	if ps.Stars > 5 {
		ps.Stars = 5
	}
	ps.resetDecayTimer()
	if ps.Stars >= 5 && !ps.Heli.Active {
		ps.Heli.Active = true
		ps.Heli.X = ArenaWidth * 0.5
		ps.Heli.Y = ArenaHeight * 0.5
		ps.Heli.ShootTimer = 0.8
	}
	return ps.Stars - prev
}

// ReduceStar lowers Wanted stars (e.g. Boss defeat or Vending Machine Bribe) and clears excess units.
func (ps *PoliceSystem) ReduceStar(n int) {
	if n <= 0 {
		return
	}
	ps.Stars -= n
	if ps.Stars < 0 {
		ps.Stars = 0
	}
	if ps.Stars < 5 {
		ps.Heli.Active = false
	}
	if ps.Stars < 3 {
		ps.Hazards = ps.Hazards[:0]
	}
	if ps.Stars == 0 {
		ps.Officers = ps.Officers[:0]
		ps.Cruisers = ps.Cruisers[:0]
		ps.HeatDecayTimer = 0
	} else {
		ps.resetDecayTimer()
	}
}

func (ps *PoliceSystem) resetDecayTimer() {
	switch ps.Stars {
	case 1:
		ps.HeatDecayTimer = 25.0
	case 2:
		ps.HeatDecayTimer = 40.0
	default:
		ps.HeatDecayTimer = 0 // 3★-5★ do not decay passively; require Boss Kill or Bribe
	}
}

// XPAndCashMultiplier returns the global risk/reward multipliers matching the GDD table.
func (ps *PoliceSystem) XPAndCashMultiplier() (xpMult float32, cashMult float32) {
	switch ps.Stars {
	case 0:
		return 1.0, 1.0
	case 1:
		return 1.3, 1.4
	case 2:
		return 1.6, 1.8
	case 3:
		return 2.0, 2.3
	case 4:
		return 2.5, 3.0
	default:
		return 3.5, 4.0
	}
}

// Update advances heat decay, police spawns, cruiser lane-rushes, and helicopter spotlight danmaku.
func (ps *PoliceSystem) Update(dt, timeScale float32, playerX, playerY float32, stops []BusStop, bullets *BulletPool, telegraphs *[]TelegraphLine, rng *rand.Rand) {
	if ps.Stars <= 0 {
		if len(ps.Hazards) > 0 {
			ps.Hazards = ps.Hazards[:0]
		}
		return
	}

	effDT := dt * timeScale

	// Passive heat decay only for 1★ (25s) and 2★ (40s)
	if ps.Stars == 1 || ps.Stars == 2 {
		ps.HeatDecayTimer -= dt
		if ps.HeatDecayTimer <= 0 {
			ps.ReduceStar(1)
			if ps.Stars <= 0 {
				return
			}
		}
	}

	// Spawn officers up to cap based on Stars (throttled cleanly during Boss Fights)
	maxOfficers := ps.Stars * 2
	if maxOfficers > 10 {
		maxOfficers = 10
	}
	if ps.BossFightMode && maxOfficers > 2 {
		maxOfficers = 2
	}
	ps.SpawnTimer -= effDT
	if ps.SpawnTimer <= 0 && len(ps.Officers) < maxOfficers {
		ps.SpawnTimer = 3.2 - float32(ps.Stars)*0.35
		if ps.SpawnTimer < 1.1 {
			ps.SpawnTimer = 1.1
		}
		if ps.BossFightMode {
			ps.SpawnTimer *= 1.6
		}
		sx, sy := randomEdgePos(rng)
		ps.Officers = append(ps.Officers, PoliceOfficer{
			ID:         ps.nextOfficerID,
			X:          sx,
			Y:          sy,
			ShootTimer: 1.0 + rng.Float32()*0.8,
			HP:         2 + ps.Stars,
			IsSWAT:     ps.Stars >= 4,
			Active:     true,
		})
		ps.nextOfficerID++
	}

	// Update PoliceOfficers
	for i := 0; i < len(ps.Officers); {
		o := &ps.Officers[i]
		if !o.Active {
			ps.Officers[i] = ps.Officers[len(ps.Officers)-1]
			ps.Officers = ps.Officers[:len(ps.Officers)-1]
			continue
		}

		// Target player (or active BusStop if SWAT at 4★+)
		tx, ty := playerX, playerY
		targetingStop := false
		if o.IsSWAT {
			for s := range stops {
				if stops[s].Active {
					tx, ty = stops[s].X, stops[s].Y
					targetingStop = true
					break
				}
			}
		}

		dx, dy, d := normalize2D(tx-o.X, ty-o.Y)
		spd := float32(72.0)
		if o.IsSWAT {
			spd = 86.0
		}
		if d > 130 {
			o.VX = dx * spd
			o.VY = dy * spd
		} else {
			o.VX = -dy * (spd * 0.45)
			o.VY = dx * (spd * 0.45)
		}
		o.X = clamp32(o.X+o.VX*effDT, 20, ArenaWidth-20)
		o.Y = clamp32(o.Y+o.VY*effDT, 24, ArenaHeight-20)

		o.ShootTimer -= effDT
		if o.ShootTimer <= 0 && d < 520 {
			aim := float32(math.Atan2(float64(dy), float64(dx)))
			antiStop := ps.Stars >= 4 && o.IsSWAT && targetingStop
			if ps.Stars >= 3 {
				o.ShootTimer = 1.75
				if ps.BossFightMode {
					o.ShootTimer = 2.8
				}
				for k := -1; k <= 1; k++ {
					ang := aim + float32(k)*0.24
					bullets.Spawn(Bullet{
						X:        o.X,
						Y:        o.Y,
						VX:       float32(math.Cos(float64(ang))) * 155,
						VY:       float32(math.Sin(float64(ang))) * 155,
						Radius:   4.5,
						MaxAge:   4.0,
						Kind:     BulletPolice,
						AntiStop: antiStop,
					})
				}
			} else {
				o.ShootTimer = 2.1
				if ps.BossFightMode {
					o.ShootTimer = 3.2
				}
				bullets.Spawn(Bullet{
					X:        o.X,
					Y:        o.Y,
					VX:       float32(math.Cos(float64(aim))) * 125,
					VY:       float32(math.Sin(float64(aim))) * 125,
					Radius:   5.0,
					MaxAge:   4.2,
					Kind:     BulletPolice,
					AntiStop: antiStop,
				})
			}
		}

		i++
	}

	// 2★+: Police Cruisers lane-rush with 0.8s red telegraph
	if ps.Stars >= 2 {
		hazardDT := effDT
		if ps.BossFightMode {
			hazardDT *= 0.55
		}
		ps.CruiserTimer -= hazardDT
		if ps.CruiserTimer <= 0 && len(ps.Cruisers) < ps.Stars-1 {
			ps.CruiserTimer = 4.6 - float32(ps.Stars)*0.45
			if ps.CruiserTimer < 2.2 {
				ps.CruiserTimer = 2.2
			}
			laneY := clamp32(playerY+(rng.Float32()-0.5)*140, 140, ArenaHeight-140)
			fromLeft := rng.Intn(2) == 0
			startX := float32(-40)
			vx := float32(520.0)
			if !fromLeft {
				startX = ArenaWidth + 40
				vx = -520.0
			}
			ps.Cruisers = append(ps.Cruisers, PoliceCruiser{
				Active:         true,
				X:              startX,
				Y:              laneY,
				VX:             vx,
				TelegraphTimer: 0.80,
				ShootTimer:     0.22,
			})
			*telegraphs = append(*telegraphs, TelegraphLine{
				X1:       0,
				Y1:       laneY,
				X2:       ArenaWidth,
				Y2:       laneY,
				Timer:    0.80,
				MaxTimer: 0.80,
				Kind:     0,
			})
		}
	}

	for i := 0; i < len(ps.Cruisers); {
		c := &ps.Cruisers[i]
		if c.TelegraphTimer > 0 {
			c.TelegraphTimer -= effDT
			if c.TelegraphTimer <= 0 {
				c.Dashing = true
			}
			i++
			continue
		}
		c.X += c.VX * effDT
		c.ShootTimer -= effDT
		if c.ShootTimer <= 0 {
			c.ShootTimer = 0.25
			// Siren radial bursts
			for _, ang := range [4]float32{0.6, 2.5, -0.6, -2.5} {
				bullets.Spawn(Bullet{
					X:      c.X,
					Y:      c.Y,
					VX:     float32(math.Cos(float64(ang))) * 150,
					VY:     float32(math.Sin(float64(ang))) * 150,
					Radius: 4.5,
					MaxAge: 3.0,
					Kind:   BulletPolice,
				})
			}
		}
		if c.X < -100 || c.X > ArenaWidth+100 {
			ps.Cruisers[i] = ps.Cruisers[len(ps.Cruisers)-1]
			ps.Cruisers = ps.Cruisers[:len(ps.Cruisers)-1]
			continue
		}
		i++
	}

	// 3★+: Laser/spike road hazard barrier strips (up to 3 active)
	if ps.Stars >= 3 {
		for i := 0; i < len(ps.Hazards); {
			h := &ps.Hazards[i]
			if !h.Active {
				ps.Hazards[i] = ps.Hazards[len(ps.Hazards)-1]
				ps.Hazards = ps.Hazards[:len(ps.Hazards)-1]
				continue
			}
			if h.TelegraphTimer > 0 {
				h.TelegraphTimer -= effDT
				if h.TelegraphTimer < 0 {
					h.TelegraphTimer = 0
				}
				i++
				continue
			}
			h.LifeTimer -= effDT
			if h.LifeTimer <= 0 {
				ps.Hazards[i] = ps.Hazards[len(ps.Hazards)-1]
				ps.Hazards = ps.Hazards[:len(ps.Hazards)-1]
				continue
			}
			i++
		}

		hazardSpawnDT := effDT
		if ps.BossFightMode {
			hazardSpawnDT *= 0.55
		}
		ps.HazardTimer -= hazardSpawnDT
		if ps.HazardTimer <= 0 && len(ps.Hazards) < 3 {
			ps.HazardTimer = 4.5 - float32(ps.Stars-3)*0.4
			if ps.HazardTimer < 3.2 {
				ps.HazardTimer = 3.2
			}
			cx := clamp32(playerX+(rng.Float32()-0.5)*220, 144, ArenaWidth-144)
			cy := clamp32(playerY+(rng.Float32()-0.5)*140, 148, BoardwalkRailY-28)
			ps.Hazards = append(ps.Hazards, PoliceHazard{
				Active:         true,
				X1:             cx - 48,
				Y1:             cy,
				X2:             cx + 48,
				Y2:             cy,
				TelegraphTimer: 0.85,
				LifeTimer:      8.0,
			})
		}
	} else if len(ps.Hazards) > 0 {
		ps.Hazards = ps.Hazards[:0]
	}

	// 5★: Helicopter spotlight pursues Alfredo and rains 360° danmaku when inside cone
	if ps.Stars >= 5 && ps.Heli.Active {
		dx, dy, d := normalize2D(playerX-ps.Heli.X, playerY-ps.Heli.Y)
		if d > 6 {
			ps.Heli.X += dx * 125 * effDT
			ps.Heli.Y += dy * 125 * effDT
		}
		if d <= ps.Heli.Radius {
			ps.Heli.ShootTimer -= effDT
			if ps.Heli.ShootTimer <= 0 {
				ps.Heli.ShootTimer = 0.65
				for k := 0; k < 10; k++ {
					ang := float32(k) * (6.283 / 10.0)
					bullets.Spawn(Bullet{
						X:      ps.Heli.X,
						Y:      ps.Heli.Y,
						VX:     float32(math.Cos(float64(ang))) * 165,
						VY:     float32(math.Sin(float64(ang))) * 165,
						Radius: 4.5,
						MaxAge: 3.5,
						Kind:   BulletPolice,
					})
				}
			}
		}
	}
}

func randomEdgePos(rng *rand.Rand) (float32, float32) {
	switch rng.Intn(4) {
	case 0:
		return rng.Float32() * ArenaWidth, 24
	case 1:
		return rng.Float32() * ArenaWidth, ArenaHeight - 24
	case 2:
		return 24, rng.Float32() * ArenaHeight
	default:
		return ArenaWidth - 24, rng.Float32() * ArenaHeight
	}
}
