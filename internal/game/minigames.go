package game

import (
	"math"
	"math/rand"
)

// Pragmata grid node types.
const (
	NodeEmpty   = 0
	NodeBlue    = 1 // +XP
	NodeYellow  = 2 // Freeze bullets
	NodeBoard   = 3 // Target (3,3)
	NodeVisited = 4
)

// MinigameState holds the diegetic world-space minigame state for a BusStop.
type MinigameState struct {
	Arch      Archetype
	Progress  float32 // 0.0 to 1.0
	Completed bool
	Synergy   bool

	// Emo: Rhythm-Lock
	BeatCursor    float32
	BeatDir       float32
	BeatWindow    float32
	BeatCooldown  float32
	LastHitResult int // 0=none, 1=hit, -1=miss

	// Student: Pragmata Grid 4x4
	GridX, GridY     int
	GridNodes        [4][4]int
	BlueCollected    int
	YellowCollected  int
	BonusXPTriggered bool
	FreezeTriggered  bool

	// Kawaii: Tear-Graze dance cooldown
	DanceCooldown float32

	// Gyaru: Stardew Valley Tension Bar
	BarCursor       float32
	BarVel          float32
	SweetSpotPos    float32
	SweetSpotTarget float32
	SweetSpotSize   float32
	SweetSpotTimer  float32
}

// NewMinigameState creates a fresh minigame tailored to the stop's archetype.
func NewMinigameState(arch Archetype, perfectDrop bool, rng *rand.Rand) MinigameState {
	m := MinigameState{
		Arch:            arch,
		BeatCursor:      0.15,
		BeatDir:         1.0,
		BeatWindow:      0.18,
		BarCursor:       0.45,
		SweetSpotPos:    0.50,
		SweetSpotTarget: 0.70,
		SweetSpotSize:   0.28,
		SweetSpotTimer:  0.9,
	}
	if perfectDrop {
		m.Progress = 0.30
	}
	if arch == ArchStudent {
		m.initPragmataGrid(rng)
	}
	return m
}

func (m *MinigameState) initPragmataGrid(rng *rand.Rand) {
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			m.GridNodes[y][x] = NodeEmpty
		}
	}
	m.GridX = 0
	m.GridY = 0
	m.GridNodes[0][0] = NodeVisited
	m.GridNodes[3][3] = NodeBoard

	// Place 2 Blue (+XP) and 2 Yellow (Freeze) nodes on intermediate cells.
	blueCandidates := [][2]int{{1, 0}, {0, 1}, {2, 2}, {1, 2}}
	yellowCandidates := [][2]int{{1, 1}, {2, 0}, {0, 2}, {2, 1}, {3, 1}, {1, 3}}

	b1 := blueCandidates[rng.Intn(2)]
	b2 := blueCandidates[2+rng.Intn(2)]
	m.GridNodes[b1[1]][b1[0]] = NodeBlue
	m.GridNodes[b2[1]][b2[0]] = NodeBlue

	y1 := yellowCandidates[rng.Intn(3)]
	y2 := yellowCandidates[3+rng.Intn(3)]
	if m.GridNodes[y1[1]][y1[0]] == NodeEmpty {
		m.GridNodes[y1[1]][y1[0]] = NodeYellow
	}
	if m.GridNodes[y2[1]][y2[0]] == NodeEmpty {
		m.GridNodes[y2[1]][y2[0]] = NodeYellow
	}
}

// ApplySynergy applies archetype-specific Synergy bonuses when matching girls are hooked.
func (m *MinigameState) ApplySynergy() {
	if m.Synergy {
		return
	}
	m.Synergy = true
	switch m.Arch {
	case ArchEmo:
		m.BeatWindow *= 1.40
	case ArchStudent:
		// HONOR ROLL: 1 Blue and 1 Yellow already active!
		m.BlueCollected++
		m.YellowCollected++
		m.BonusXPTriggered = true
		m.FreezeTriggered = true
	case ArchGyaru:
		// VIRAL #1: Stardew cursor +35% larger
		m.SweetSpotSize *= 1.35
	}
}

// ResetPragmataOnHit resets the Pragmata grid path when Alfredo takes damage while tethered.
func (m *MinigameState) ResetPragmataOnHit() {
	if m.Arch != ArchStudent || m.Completed {
		return
	}
	m.GridX = 0
	m.GridY = 0
	m.Progress = 0
}

// AddGrazeCharge boosts Kawaii Tear-Graze progress when grazing a Tear droplet.
func (m *MinigameState) AddGrazeCharge() {
	if m.Arch != ArchKawaii || m.Completed {
		return
	}
	m.Progress += 0.25
	if m.Progress >= 1.0 {
		m.Progress = 1.0
		m.Completed = true
	}
}

// Update advances the minigame in real time while Alfredo is inside the Bus Stop radius.
// Note: Pragmata cursor persists automatically if Alfredo leaves the radius to dash!
func (m *MinigameState) Update(dt float32, playerInside bool, actionDir int, rng *rand.Rand) {
	if m.Completed {
		return
	}
	if m.BeatCooldown > 0 {
		m.BeatCooldown -= dt
	}
	if m.DanceCooldown > 0 {
		m.DanceCooldown -= dt
	}

	if !playerInside {
		return
	}

	switch m.Arch {
	case ArchEmo:
		// Rhythm-Lock: beat cursor sweeps across [0, 1]
		m.BeatCursor += m.BeatDir * 1.45 * dt
		if m.BeatCursor >= 1.0 {
			m.BeatCursor = 1.0
			m.BeatDir = -1.0
		} else if m.BeatCursor <= 0.0 {
			m.BeatCursor = 0.0
			m.BeatDir = 1.0
		}
		if actionDir != ActionNone && m.BeatCooldown <= 0 {
			m.BeatCooldown = 0.14
			if float32(math.Abs(float64(m.BeatCursor-0.5))) <= m.BeatWindow {
				m.Progress += 0.36
				m.LastHitResult = 1
			} else {
				m.Progress = clamp32(m.Progress-0.06, 0, 1)
				m.LastHitResult = -1
			}
		}

	case ArchStudent:
		// Pragmata Grid 4x4: directional input moves cursor 1 cell
		if actionDir != ActionNone {
			nx, ny := m.GridX, m.GridY
			switch actionDir {
			case ActionUp:
				ny--
			case ActionRight:
				nx++
			case ActionDown:
				ny++
			case ActionLeft:
				nx--
			}
			if nx >= 0 && nx < 4 && ny >= 0 && ny < 4 {
				m.GridX = nx
				m.GridY = ny
				cell := m.GridNodes[ny][nx]
				switch cell {
				case NodeBlue:
					m.BlueCollected++
					m.BonusXPTriggered = true
					m.GridNodes[ny][nx] = NodeVisited
				case NodeYellow:
					m.YellowCollected++
					m.FreezeTriggered = true
					m.GridNodes[ny][nx] = NodeVisited
				case NodeEmpty:
					m.GridNodes[ny][nx] = NodeVisited
				case NodeBoard:
					m.Progress = 1.0
				}
				stepProg := float32(m.GridX+m.GridY) / 6.0 * 0.92
				if stepProg > m.Progress && m.GridNodes[ny][nx] != NodeBoard {
					m.Progress = stepProg
				}
			}
		}

	case ArchKawaii:
		// Tear-Graze: auto-charges inside umbrella, plus dance boost on ActionDir
		m.Progress += 0.24 * dt
		if actionDir != ActionNone && m.DanceCooldown <= 0 {
			m.DanceCooldown = 0.24
			m.Progress += 0.16
		}

	case ArchGyaru:
		// Stardew Valley Tension Bar: sweet spot oscillates smoothly
		m.SweetSpotTimer -= dt
		if m.SweetSpotTimer <= 0 || float32(math.Abs(float64(m.SweetSpotPos-m.SweetSpotTarget))) < 0.04 {
			m.SweetSpotTimer = 0.75 + rng.Float32()*0.65
			m.SweetSpotTarget = 0.18 + rng.Float32()*0.64
		}
		if m.SweetSpotPos < m.SweetSpotTarget {
			m.SweetSpotPos += 0.36 * dt
		} else {
			m.SweetSpotPos -= 0.36 * dt
		}
		m.SweetSpotPos = clamp32(m.SweetSpotPos, 0.14, 0.86)

		// Cursor physics: gravity pulls down, tapping ActionDir kicks upward
		m.BarVel -= 1.45 * dt
		if actionDir == ActionUp || actionDir == ActionRight || actionDir == ActionLeft || actionDir == ActionDown {
			m.BarVel = 0.56
		}
		m.BarCursor += m.BarVel * dt
		if m.BarCursor < 0 {
			m.BarCursor = 0
			m.BarVel = 0
		} else if m.BarCursor > 1 {
			m.BarCursor = 1
			m.BarVel = -0.1
		}

		half := m.SweetSpotSize * 0.5
		if float32(math.Abs(float64(m.BarCursor-m.SweetSpotPos))) <= half {
			m.Progress += 0.54 * dt
		} else {
			m.Progress -= 0.18 * dt
			if m.Progress < 0 {
				m.Progress = 0
			}
		}
	}

	if m.Progress >= 1.0 {
		m.Progress = 1.0
		m.Completed = true
	}
}

// ComboState handles the 4-Step Boarding Combo after the Minigame completes.
type ComboState struct {
	Active         bool
	Sequence       [4]int
	ComboIndex     int
	Timer          float32
	MaxTimer       float32
	StrictMode     bool // True when hooked to La Professoressa: any mistake fails immediately!
	Failed         bool
	Completed      bool
	LastErrorFlash float32
}

// Start initializes a fresh 4-step directional combo sequence.
func (c *ComboState) Start(synergyEmo bool, strictMode bool, rng *rand.Rand) {
	c.Active = true
	c.ComboIndex = 0
	c.MaxTimer = 3.5
	if synergyEmo {
		c.MaxTimer *= 1.40
	}
	c.Timer = c.MaxTimer
	c.StrictMode = strictMode
	c.Failed = false
	c.Completed = false
	c.LastErrorFlash = 0
	for i := 0; i < 4; i++ {
		c.Sequence[i] = rng.Intn(4)
	}
}

// Update advances the escape timer and processes directional combo inputs.
func (c *ComboState) Update(dt float32, playerInside bool, actionDir int) {
	if !c.Active || c.Completed || c.Failed {
		return
	}
	if c.LastErrorFlash > 0 {
		c.LastErrorFlash -= dt
		if c.LastErrorFlash < 0 {
			c.LastErrorFlash = 0
		}
	}

	// Leaving the radius accelerates escape timer (GDD rule)
	drain := float32(1.0)
	if !playerInside {
		drain = 1.65
	}
	c.Timer -= dt * drain
	if c.Timer <= 0 {
		c.Timer = 0
		c.Failed = true
		return
	}

	if playerInside && actionDir != ActionNone {
		expected := c.Sequence[c.ComboIndex]
		if actionDir == expected {
			c.ComboIndex++
			if c.ComboIndex >= 4 {
				c.Completed = true
			}
		} else {
			c.LastErrorFlash = 0.28
			if c.StrictMode {
				// La Professoressa: "Bocciato!"
				c.Failed = true
			} else {
				c.Timer -= 0.45
				if c.Timer <= 0 {
					c.Timer = 0
					c.Failed = true
				}
			}
		}
	}
}
