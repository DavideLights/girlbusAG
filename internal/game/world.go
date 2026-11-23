package game

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// World is the complete, deterministic pure-Go game simulation state.
type World struct {
	Player     Player
	Stops      []BusStop
	Girls      []Girl
	Bullets    BulletPool
	Police     PoliceSystem
	Boss       Boss
	Buses      []NeonBus
	ZTLZones   []Rect
	Shops      []VendingShop
	Popups     []FloatingText
	Telegraphs []TelegraphLine
	Express    []ExpressLine
	WaveRings  []WaveRing

	Props         []MapProp
	Lights        []LightSource
	ParkIslands   []Rect
	BusBays       []Rect
	CrosswalkCols []int
	CrosswalkRows []int
	CanalY        float32
	BoardwalkY    float32

	HitStopTimer      float32
	ScreenShake       float32
	BulletFreezeTimer float32
	RunTime           float32
	Score             int
	GirlsBoarded      int
	DifficultyTier    int
	BossesSpawned     [3]bool
	BossesDefeated    int
	GameOver          bool
	Victory           bool

	PendingLevelUp bool
	UpgradeChoices [3]UpgradeCard

	CameraX, CameraY float32

	girlSpawnTimer float32
	nextEntityID   int
	rng            *rand.Rand
}

// NewWorld creates a new run using a time-based seed.
func NewWorld() *World {
	return NewWorldWithSeed(time.Now().UnixNano())
}

// NewWorldWithSeed creates a deterministic game World from the given RNG seed.
func NewWorldWithSeed(seed int64) *World {
	layout := BuildDistrictLayout()
	w := &World{
		Player:         NewPlayer(),
		Stops:          make([]BusStop, 0, 8),
		Girls:          make([]Girl, 0, 48),
		Police:         NewPoliceSystem(),
		Buses:          make([]NeonBus, 0, 4),
		ZTLZones:       layout.ZTLZones,
		Shops:          layout.Shops,
		Props:          layout.Props,
		Lights:         layout.Lights,
		ParkIslands:    layout.ParkIslands,
		BusBays:        layout.BusBays,
		CrosswalkCols:  layout.CrosswalkCols,
		CrosswalkRows:  layout.CrosswalkRows,
		CanalY:         layout.CanalY,
		BoardwalkY:     layout.BoardwalkY,
		Popups:         make([]FloatingText, 0, 24),
		Telegraphs:     make([]TelegraphLine, 0, 12),
		Express:        make([]ExpressLine, 0, 8),
		WaveRings:      make([]WaveRing, 0, 16),
		DifficultyTier: 1,
		girlSpawnTimer: 1.5,
		nextEntityID:   1,
		rng:            rand.New(rand.NewSource(seed)),
	}

	// Seed initial horde across all 4 archetypes
	for i := 0; i < 8; i++ {
		arch := Archetype(i % int(ArchCount))
		w.spawnGirl(arch, i == 7)
	}

	w.snapCamera()
	return w
}

func (w *World) allocID() int {
	id := w.nextEntityID
	w.nextEntityID++
	return id
}

func (w *World) snapCamera() {
	w.CameraX = clamp32(w.Player.X-ScreenWidth*0.5, 0, ArenaWidth-ScreenWidth)
	w.CameraY = clamp32(w.Player.Y-ScreenHeight*0.5, 0, ArenaHeight-ScreenHeight)
}

// ActiveStopCount returns how many Bus Stops are currently active in the arena.
func (w *World) ActiveStopCount() int {
	n := 0
	for i := range w.Stops {
		if w.Stops[i].Active {
			n++
		}
	}
	return n
}

// ForceSpawnBoss immediately spawns a Boss of the requested kind (useful for testing/challenge hotkeys).
func (w *World) ForceSpawnBoss(kind BossKind) {
	if w.Boss.Active {
		return
	}
	sx, sy := randomEdgePos(w.rng)
	w.Boss.Spawn(kind, sx, sy)
	w.BossesSpawned[kind] = true
	w.addPopup(w.Player.X, w.Player.Y-36, "BOSS: "+w.Boss.Name+"!", 4)
}

func (w *World) addPopup(x, y float32, text string, kind int) {
	if len(w.Popups) >= 32 {
		w.Popups = w.Popups[1:]
	}
	w.Popups = append(w.Popups, FloatingText{
		X:       x,
		Y:       y,
		VY:      -28,
		Life:    1.35,
		MaxLife: 1.35,
		Text:    text,
		Kind:    kind,
	})
}

func (w *World) addWaveRing(x, y, maxR float32, kind int) {
	w.WaveRings = append(w.WaveRings, WaveRing{
		X:         x,
		Y:         y,
		Radius:    8,
		MaxRadius: maxR,
		Life:      0.42,
		MaxLife:   0.42,
		Kind:      kind,
	})
}

// Update advances the entire simulation by dt seconds using the provided InputState.
func (w *World) Update(in InputState, dt float32) {
	dt = clamp32(dt, 0, 0.05)
	if dt <= 0 || w.GameOver {
		return
	}

	// Level-Up 3-Card Selection pauses simulation cleanly
	if w.PendingLevelUp {
		if in.ChooseUpgrade >= 0 && in.ChooseUpgrade < 3 {
			w.applyUpgrade(w.UpgradeChoices[in.ChooseUpgrade])
			w.PendingLevelUp = false
		}
		return
	}

	// Screen shake decay
	if w.ScreenShake > 0 {
		w.ScreenShake -= dt * 26.0
		if w.ScreenShake < 0 {
			w.ScreenShake = 0
		}
	}

	// Brutal Katana ZERO Hit-Stop freeze
	if w.HitStopTimer > 0 {
		w.HitStopTimer -= dt
		if w.HitStopTimer < 0 {
			w.HitStopTimer = 0
		}
		return
	}

	w.RunTime += dt

	w.updateVFX(dt)
	w.updateExpressLanes()

	inExpress := w.isPointInExpressLane(w.Player.X, w.Player.Y)
	inEmoFog := w.isPlayerInEmoFog()

	// Update Alfredo
	if w.Player.Update(in, dt, inExpress, inEmoFog) {
		w.triggerBasslinePulse()
	}
	w.constrainPlayerPos()

	// Time scales: Chronos slows enemy/bullet simulation while keeping Alfredo & minigames crisp
	enemyTimeScale := float32(1.0)
	if w.Player.ChronosActive {
		enemyTimeScale = w.Player.ChronosTimeScale
	}
	bulletTimeScale := enemyTimeScale
	if w.BulletFreezeTimer > 0 {
		w.BulletFreezeTimer -= dt
		if w.BulletFreezeTimer < 0 {
			w.BulletFreezeTimer = 0
		}
		bulletTimeScale = 0.08
	}

	// Throw tactical Bus Stop
	if in.ThrowJustPressed {
		w.tryThrowBusStop(in.AimWorldX, in.AimWorldY)
	}

	// Update Stops, Minigames & Boarding Combos
	w.updateStops(dt, in.ActionDir)

	// Update Neon Buses
	w.updateBuses(dt)

	// Timed Boss Gatekeepers (1:30, 3:30, 5:30)
	w.checkTimedBossSpawns()

	// Update Girls & horde spawning
	w.updateGirlSpawner(dt)
	for i := range w.Girls {
		expressSlow := w.isPointInExpressLane(w.Girls[i].X, w.Girls[i].Y)
		w.Girls[i].Update(dt, enemyTimeScale, w.DifficultyTier, w.Player.X, w.Player.Y, w.Stops, expressSlow, &w.Bullets)
		w.constrainGirlPos(&w.Girls[i])
	}

	// Update Police & Boss
	w.Police.Update(dt, enemyTimeScale, w.Player.X, w.Player.Y, w.Stops, &w.Bullets, &w.Telegraphs, w.rng)
	w.Boss.Update(dt, enemyTimeScale, w.Player.X, w.Player.Y, w.Stops, w.Girls, &w.Bullets, &w.Telegraphs)

	// Update BulletPool & resolve solid prop cover collisions
	w.Bullets.Update(dt, bulletTimeScale, w.Player.X, w.Player.Y, w.Stops)
	w.resolveBulletPropCollisions()

	// Check Tear-Graze & Bullet-Player collisions
	w.handleBulletInteractions()

	// Vending Machine Shop interaction [E]
	w.updateShops(dt, in.InteractJustPressed)

	// Smooth camera follow
	targetCamX := clamp32(w.Player.X-ScreenWidth*0.5, 0, ArenaWidth-ScreenWidth)
	targetCamY := clamp32(w.Player.Y-ScreenHeight*0.5, 0, ArenaHeight-ScreenHeight)
	lerp := clamp32(dt*9.0, 0, 1)
	w.CameraX += (targetCamX - w.CameraX) * lerp
	w.CameraY += (targetCamY - w.CameraY) * lerp
}

func (w *World) updateVFX(dt float32) {
	for i := 0; i < len(w.Popups); {
		p := &w.Popups[i]
		p.Life -= dt
		p.Y += p.VY * dt
		if p.Life <= 0 {
			w.Popups[i] = w.Popups[len(w.Popups)-1]
			w.Popups = w.Popups[:len(w.Popups)-1]
			continue
		}
		i++
	}

	for i := 0; i < len(w.WaveRings); {
		r := &w.WaveRings[i]
		r.Life -= dt
		if r.Life <= 0 {
			w.WaveRings[i] = w.WaveRings[len(w.WaveRings)-1]
			w.WaveRings = w.WaveRings[:len(w.WaveRings)-1]
			continue
		}
		t := 1.0 - (r.Life / r.MaxLife)
		r.Radius = 8 + (r.MaxRadius-8)*t
		i++
	}

	for i := 0; i < len(w.Telegraphs); {
		t := &w.Telegraphs[i]
		t.Timer -= dt
		if t.Timer <= 0 {
			// Professor Laser fires instant check when telegraph completes
			if t.Kind == 1 && !w.Player.IsInvulnerable() {
				if pointToSegmentDistSq(w.Player.X, w.Player.Y, t.X1, t.Y1, t.X2, t.Y2) <= 100.0 {
					w.damagePlayer(1, BulletProtractor)
				}
			}
			w.Telegraphs[i] = w.Telegraphs[len(w.Telegraphs)-1]
			w.Telegraphs = w.Telegraphs[:len(w.Telegraphs)-1]
			continue
		}
		i++
	}
}

func (w *World) updateExpressLanes() {
	w.Express = w.Express[:0]
	maxDistSq := ExpressLinkMaxDist * ExpressLinkMaxDist
	for i := 0; i < len(w.Stops); i++ {
		if !w.Stops[i].Active {
			continue
		}
		for j := i + 1; j < len(w.Stops); j++ {
			if !w.Stops[j].Active {
				continue
			}
			if distSq(w.Stops[i].X, w.Stops[i].Y, w.Stops[j].X, w.Stops[j].Y) <= maxDistSq {
				w.Express = append(w.Express, ExpressLine{
					X1: w.Stops[i].X, Y1: w.Stops[i].Y,
					X2: w.Stops[j].X, Y2: w.Stops[j].Y,
				})
			}
		}
	}
}

func (w *World) isPointInExpressLane(x, y float32) bool {
	for i := range w.Express {
		e := w.Express[i]
		if pointToSegmentDistSq(x, y, e.X1, e.Y1, e.X2, e.Y2) <= 44.0*44.0 {
			return true
		}
	}
	return false
}

func (w *World) isPlayerInEmoFog() bool {
	for i := range w.Stops {
		st := &w.Stops[i]
		if st.Active && st.Arch == ArchEmo {
			if distSq(w.Player.X, w.Player.Y, st.X, st.Y) <= st.CaptureRadius*st.CaptureRadius {
				return true
			}
		}
	}
	return false
}

func (w *World) triggerBasslinePulse() {
	w.addWaveRing(w.Player.X, w.Player.Y, BasslinePulseRadius, 0)
	rSq := BasslinePulseRadius * BasslinePulseRadius
	for i := range w.Girls {
		g := &w.Girls[i]
		if g.HookedByStop == -1 && distSq(g.X, g.Y, w.Player.X, w.Player.Y) <= rSq {
			dx, dy, _ := normalize2D(g.X-w.Player.X, g.Y-w.Player.Y)
			g.X += dx * 24
			g.Y += dy * 24
			w.constrainGirlPos(g)
		}
	}
}

func (w *World) walkableMaxY() float32 {
	if w.CanalY > 0 {
		return w.CanalY - 4
	}
	return BoardwalkRailY
}

func (w *World) constrainCircleToWorld(x, y, radius, minX, maxX, minY, maxY float32) (float32, float32) {
	for i := range w.Props {
		if w.Props[i].Solid {
			x, y = resolveCircleRect(x, y, radius, w.Props[i].CollideRect)
		}
	}
	return clamp32(x, minX, maxX), clamp32(y, minY, maxY)
}

func (w *World) constrainPlayerPos() {
	w.Player.X, w.Player.Y = w.constrainCircleToWorld(w.Player.X, w.Player.Y, 10.0, 18, ArenaWidth-18, 24, w.walkableMaxY())
}

func (w *World) constrainGirlPos(g *Girl) {
	g.X, g.Y = w.constrainCircleToWorld(g.X, g.Y, 9.0, 24, ArenaWidth-24, 28, w.walkableMaxY())
}

func (w *World) resolveBulletPropCollisions() {
	if len(w.Props) == 0 {
		return
	}
	for i := 0; i < w.Bullets.Count; {
		b := &w.Bullets.Items[i]
		hit := false
		for p := range w.Props {
			pr := &w.Props[p]
			if pr.Solid && pr.CollideRect.IntersectsCircle(b.X, b.Y, b.Radius) {
				hit = true
				break
			}
		}
		if hit {
			w.Bullets.RemoveAt(i)
			continue
		}
		i++
	}
}

func (w *World) isPointInZTL(x, y float32) bool {
	for _, z := range w.ZTLZones {
		if z.Contains(x, y) {
			return true
		}
	}
	return false
}

func (w *World) isPointInBusBay(x, y float32) bool {
	for _, bay := range w.BusBays {
		if bay.Contains(x, y) {
			return true
		}
	}
	return false
}

// EvaluateAim computes real-time tactical placement feedback for the UI aim reticle.
func (w *World) EvaluateAim(aimX, aimY float32) AimFeedback {
	spec := GetArchetypeSpec(w.Player.SelectedStopArch)
	maxRange := spec.ThrowRange * w.Player.RangeMult
	dx, dy, d := normalize2D(aimX-w.Player.X, aimY-w.Player.Y)
	inRange := d <= maxRange
	if d > maxRange {
		d = maxRange
	}
	if d < 1 {
		dx = w.Player.FacingX
		dy = 0
		d = 48
	}
	maxThrowY := w.walkableMaxY() - 8
	tx := clamp32(w.Player.X+dx*d, 48, ArenaWidth-48)
	ty := clamp32(w.Player.Y+dy*d, 76, maxThrowY)

	inZTL := w.isPointInZTL(tx, ty)
	inBusBay := w.isPointInBusBay(tx, ty)

	previewR := spec.CaptureRadius * w.Player.RangeMult
	if inZTL {
		previewR *= 1.25
	}

	perfectDrop := false
	perfSq := PerfectHitRadius * PerfectHitRadius
	if w.Boss.Active && w.Boss.HookedByStop == -1 && distSq(tx, ty, w.Boss.X, w.Boss.Y) <= perfSq {
		perfectDrop = true
	}

	capSq := previewR * previewR
	candidates := 0
	synergyCount := 0
	vipCount := 0
	for i := range w.Girls {
		g := &w.Girls[i]
		if g.HookedByStop != -1 {
			continue
		}
		d2 := distSq(tx, ty, g.X, g.Y)
		if d2 <= perfSq {
			perfectDrop = true
		}
		if d2 <= capSq {
			candidates++
			if g.Arch == w.Player.SelectedStopArch {
				synergyCount++
			}
			if g.IsVIP {
				vipCount++
			}
		}
	}
	if w.Boss.Active && w.Boss.HookedByStop == -1 {
		bossR := previewR + 24.0
		if distSq(tx, ty, w.Boss.X, w.Boss.Y) <= bossR*bossR {
			candidates++
		}
	}

	return AimFeedback{
		ClampedX:           tx,
		ClampedY:           ty,
		InRange:            inRange,
		IsPerfectDrop:      perfectDrop,
		SynergyCount:       synergyCount,
		HookCandidateCount: candidates,
		VIPCount:           vipCount,
		InZTL:              inZTL,
		InBusBay:           inBusBay,
		PreviewRadius:      previewR,
	}
}

func (w *World) tryThrowBusStop(aimX, aimY float32) {
	if w.Player.StunTimer > 0 || w.Player.StopCooldownTimer > 0 {
		return
	}
	if w.ActiveStopCount() >= w.Player.EffectiveMaxStops() {
		return
	}

	spec := GetArchetypeSpec(w.Player.SelectedStopArch)
	fb := w.EvaluateAim(aimX, aimY)
	tx, ty := fb.ClampedX, fb.ClampedY
	inZTL := fb.InZTL
	inBusBay := fb.InBusBay
	perfectDrop := fb.IsPerfectDrop

	stop := NewBusStop(w.allocID(), tx, ty, w.Player.SelectedStopArch, w.Player.RangeMult, inZTL, inBusBay, perfectDrop, w.rng)

	w.Player.StopCooldownTimer = w.Player.StopCooldownBase
	w.Player.ThrowAnimTimer = 0.22

	if perfectDrop {
		w.Player.StopCooldownTimer *= 0.65
		w.HitStopTimer = 0.090 // 90ms Perfect Drop Hit-Stop
		w.ScreenShake = 6.5
		w.Bullets.ClearRadius(tx, ty, 130.0)
		w.addWaveRing(tx, ty, 130.0, 1)
		w.addPopup(tx, ty-28, "PERFECT DROP!", 1)
	} else {
		w.HitStopTimer = 0.045 // 45ms standard hook Hit-Stop
		w.ScreenShake = 3.0
	}

	if inZTL {
		added := w.Police.AddStar(1)
		if added > 0 {
			w.addPopup(tx, ty-44, "ZTL! +1★ WANTED!", 2)
		} else {
			w.addPopup(tx, ty-44, "ZTL 3x BONUS!", 2)
		}
	} else if inBusBay {
		w.addPopup(tx, ty-44, "🚏 BUS BAY +25% SPEED!", 1)
	}

	// Hook Boss if inside capture radius
	if w.Boss.Active && w.Boss.HookedByStop == -1 {
		bossR := stop.CaptureRadius + 24.0
		if distSq(tx, ty, w.Boss.X, w.Boss.Y) <= bossR*bossR {
			stop.HookedBoss = true
			w.Boss.HookedByStop = stop.ID
			w.addPopup(w.Boss.X, w.Boss.Y-32, "BOSS HOOKED!", 4)
		}
	}

	// Hook up to MaxMultiHook closest girls inside CaptureRadius
	capSq := stop.CaptureRadius * stop.CaptureRadius
	hookedVIP := false
	for len(stop.HookedGirlIDs) < w.Player.MaxMultiHook {
		bestIdx := -1
		bestDSq := capSq + 1
		for i := range w.Girls {
			g := &w.Girls[i]
			if g.HookedByStop != -1 {
				continue
			}
			d2 := distSq(tx, ty, g.X, g.Y)
			if d2 <= capSq && d2 < bestDSq {
				bestDSq = d2
				bestIdx = i
			}
		}
		if bestIdx == -1 {
			break
		}
		g := &w.Girls[bestIdx]
		g.HookedByStop = stop.ID
		stop.HookedGirlIDs = append(stop.HookedGirlIDs, g.ID)
		if g.IsVIP {
			hookedVIP = true
		}
		if g.Arch == stop.Arch {
			stop.ActivateSynergy()
		}
	}

	if hookedVIP {
		if w.Police.AddStar(1) > 0 {
			w.addPopup(tx, ty-56, "VIP TARGET! +1★!", 2)
		}
	}

	if stop.SynergyActive {
		w.addPopup(tx, ty-16, spec.SynergyName+"!", 1)
		switch stop.Arch {
		case ArchEmo:
			w.Player.Heal(1) // Black Heart heal
		case ArchStudent:
			// HONOR ROLL: Stun Students inside radius for 1.2s
			for i := range w.Girls {
				if w.Girls[i].Arch == ArchStudent && distSq(tx, ty, w.Girls[i].X, w.Girls[i].Y) <= capSq {
					w.Girls[i].StunTimer = 1.2
				}
			}
		case ArchGyaru:
			// VIRAL #1: Mega-Flash stuns horde for 2.0s
			flashR := stop.CaptureRadius * 1.3
			w.addWaveRing(tx, ty, flashR, 4)
			for i := range w.Girls {
				if distSq(tx, ty, w.Girls[i].X, w.Girls[i].Y) <= flashR*flashR {
					w.Girls[i].StunTimer = 2.0
				}
			}
		}
	}

	w.Stops = append(w.Stops, stop)
}

func (w *World) updateStops(dt float32, actionDir int) {
	// Route directional input to the closest active stop that Alfredo is standing inside
	primaryStopIdx := -1
	bestDistSq := float32(1e12)
	for i := range w.Stops {
		st := &w.Stops[i]
		if !st.Active || st.BusRushTriggered {
			continue
		}
		d2 := distSq(w.Player.X, w.Player.Y, st.X, st.Y)
		st.PlayerInside = d2 <= st.CaptureRadius*st.CaptureRadius
		if st.PlayerInside && d2 < bestDistSq {
			bestDistSq = d2
			primaryStopIdx = i
		}
	}

	for i := 0; i < len(w.Stops); {
		st := &w.Stops[i]
		if !st.Active {
			w.Stops[i] = w.Stops[len(w.Stops)-1]
			w.Stops = w.Stops[:len(w.Stops)-1]
			continue
		}

		// Archetype passive stop perks
		if st.Arch == ArchStudent {
			st.BellTimer -= dt
			if st.BellTimer <= 0 {
				st.BellTimer = 3.5
				bellR := st.CaptureRadius * 1.25
				w.addWaveRing(st.X, st.Y, bellR, 3)
				for g := range w.Girls {
					gr := &w.Girls[g]
					if gr.HookedByStop == -1 && distSq(gr.X, gr.Y, st.X, st.Y) <= bellR*bellR {
						dx, dy, _ := normalize2D(gr.X-st.X, gr.Y-st.Y)
						gr.X += dx * 42
						gr.Y += dy * 42
						w.constrainGirlPos(gr)
					}
				}
				for o := range w.Police.Officers {
					of := &w.Police.Officers[o]
					if distSq(of.X, of.Y, st.X, st.Y) <= bellR*bellR {
						dx, dy, _ := normalize2D(of.X-st.X, of.Y-st.Y)
						of.X = clamp32(of.X+dx*48, 20, ArenaWidth-20)
						of.Y = clamp32(of.Y+dy*48, 24, ArenaHeight-20)
					}
				}
			}
		} else if st.Arch == ArchGyaru {
			st.HotspotZapTimer -= dt
			if st.HotspotZapTimer <= 0 {
				st.HotspotZapTimer = 0.9
				for g := range w.Girls {
					gr := &w.Girls[g]
					if gr.HookedByStop == -1 && distSq(gr.X, gr.Y, st.X, st.Y) <= st.CaptureRadius*st.CaptureRadius {
						gr.StunTimer = 0.25
					}
				}
			}
		}

		if st.BusRushTriggered {
			i++
			continue
		}

		dir := ActionNone
		if i == primaryStopIdx {
			dir = actionDir
		}

		// LifeTimer drains slower when Alfredo is inside, faster if empty/outside
		if !st.Combo.Active {
			drain := float32(0.55)
			if !st.PlayerInside {
				drain = 1.25
			}
			st.LifeTimer -= dt * drain
			if st.LifeTimer <= 0 {
				w.failStop(st, false)
				i++
				continue
			}
		}

		if !st.Minigame.Completed {
			mgDT := dt
			if st.InBusBay && (st.Minigame.Arch == ArchKawaii || st.Minigame.Arch == ArchGyaru) {
				mgDT *= 1.25
			}
			prevProg := st.Minigame.Progress
			prevGX, prevGY := st.Minigame.GridX, st.Minigame.GridY
			st.Minigame.Update(mgDT, st.PlayerInside, dir, w.rng)
			if st.InBusBay && !st.Minigame.Completed {
				switch st.Minigame.Arch {
				case ArchEmo:
					if st.Minigame.Progress > prevProg {
						st.Minigame.Progress += (st.Minigame.Progress - prevProg) * 0.25
					}
				case ArchKawaii:
					contGain := float32(0.24) * mgDT
					if st.PlayerInside && st.Minigame.Progress > prevProg+contGain+0.01 {
						discrete := st.Minigame.Progress - prevProg - contGain
						st.Minigame.Progress += discrete * 0.25
					}
				case ArchStudent:
					if st.Minigame.GridX != prevGX || st.Minigame.GridY != prevGY {
						boosted := float32(st.Minigame.GridX+st.Minigame.GridY) / 6.0 * 0.92 * 1.25
						if boosted > st.Minigame.Progress {
							st.Minigame.Progress = boosted
						}
					}
				}
				if st.Minigame.Progress >= 1.0 {
					st.Minigame.Progress = 1.0
					st.Minigame.Completed = true
				}
			}
			if st.Minigame.BonusXPTriggered {
				st.Minigame.BonusXPTriggered = false
				xpMult, _ := w.Police.XPAndCashMultiplier()
				xpGain := int(20 * xpMult)
				w.addPopup(st.X, st.Y-24, fmt.Sprintf("+%d XP NODE!", xpGain), 3)
				if w.Player.AddXP(xpGain) {
					w.triggerLevelUp()
				}
			}
			if st.Minigame.FreezeTriggered {
				st.Minigame.FreezeTriggered = false
				w.BulletFreezeTimer = 1.5
				w.addPopup(st.X, st.Y-38, "BULLETS FROZEN!", 1)
			}
		}
		if st.Minigame.Completed && !st.Combo.Active && !st.Combo.Completed && !st.Combo.Failed {
			strict := st.HookedBoss && w.Boss.Active && w.Boss.Kind == BossProfessor
			st.Combo.Start(st.Arch == ArchEmo && st.SynergyActive, strict, w.rng)
		} else if st.Combo.Active && !st.Combo.Completed && !st.Combo.Failed {
			st.Combo.Update(dt, st.PlayerInside, dir)
			if st.Combo.Completed {
				st.BusRushTriggered = true
				w.Buses = append(w.Buses, NewNeonBus(st))
				w.addPopup(st.X, st.Y-32, "NEON BUS RUSH!", 1)
			} else if st.Combo.Failed {
				w.failStop(st, st.Combo.StrictMode)
			}
		}

		i++
	}
}

func (w *World) failStop(st *BusStop, professorStun bool) {
	for _, gid := range st.HookedGirlIDs {
		for i := range w.Girls {
			if w.Girls[i].ID == gid {
				w.Girls[i].Enrage(w.Player.X, w.Player.Y, &w.Bullets)
				break
			}
		}
	}
	if st.HookedBoss && w.Boss.Active && w.Boss.HookedByStop == st.ID {
		w.Boss.HookedByStop = -1
	}
	if professorStun {
		w.Player.StunTimer = 1.0
		w.addPopup(w.Player.X, w.Player.Y-30, "BOCCIATO! STUN 1s!", 2)
	} else {
		w.addPopup(st.X, st.Y-24, "ESCAPED!", 2)
	}
	st.Active = false
}

func (w *World) updateBuses(dt float32) {
	for i := 0; i < len(w.Buses); {
		bus := &w.Buses[i]
		if !bus.Active {
			w.Buses[i] = w.Buses[len(w.Buses)-1]
			w.Buses = w.Buses[:len(w.Buses)-1]
			continue
		}

		prevX := bus.X
		bus.X += bus.Speed * dt

		// Neon Bus Purification Wave clears bullets along its streak
		w.Bullets.ClearLine(prevX, bus.Y, bus.X+42, bus.Y, 46.0)

		// Knock aside unhooked girls & police officers in bus lane
		for g := range w.Girls {
			gr := &w.Girls[g]
			if gr.HookedByStop == -1 && float32(math.Abs(float64(gr.Y-bus.Y))) < 40 && gr.X >= prevX-30 && gr.X <= bus.X+48 {
				if gr.Y < bus.Y {
					gr.Y -= 56
				} else {
					gr.Y += 56
				}
				w.constrainGirlPos(gr)
			}
		}

		// Collect target stop when NeonBus arrives
		if !bus.Collected {
			for s := range w.Stops {
				st := &w.Stops[s]
				if st.Active && st.ID == bus.TargetStopID && bus.X >= st.X {
					bus.Collected = true
					w.completeStopBoarding(st)
					break
				}
			}
		}

		if bus.X > ArenaWidth+140 {
			bus.Active = false
		}
		i++
	}
}

func (w *World) completeStopBoarding(st *BusStop) {
	w.addWaveRing(st.X, st.Y, 155, 2)
	w.Bullets.ClearRadius(st.X, st.Y, 155)
	w.ScreenShake = 7.0

	xpMult, cashMult := w.Police.XPAndCashMultiplier()
	totalXP := 0
	totalCash := 0
	totalScore := 0

	// Collect all hooked girls
	for _, gid := range st.HookedGirlIDs {
		for i := 0; i < len(w.Girls); i++ {
			if w.Girls[i].ID == gid {
				g := w.Girls[i]
				gXP := float32(26.0)
				gCash := float32(40.0)
				gScore := float32(100.0)
				if g.IsVIP {
					gCash *= 3.0
					gScore *= 3.0
				}
				if st.InZTL {
					gCash *= 3.0
					gScore *= 3.0
				}
				totalXP += int(gXP * xpMult)
				totalCash += int(gCash * cashMult)
				totalScore += int(gScore)
				w.GirlsBoarded++

				// Swap-remove girl
				w.Girls[i] = w.Girls[len(w.Girls)-1]
				w.Girls = w.Girls[:len(w.Girls)-1]
				break
			}
		}
	}

	// Boss 3-Stop Break!
	if st.HookedBoss && w.Boss.Active && w.Boss.HookedByStop == st.ID {
		w.HitStopTimer = 0.140 // 140ms Boss Break Hit-Stop
		w.ScreenShake = 11.5
		w.Bullets.ClearRadius(w.Boss.X, w.Boss.Y, 240)
		defeated := w.Boss.BreakShield()
		if defeated {
			totalXP += int(240 * xpMult)
			totalCash += int(500 * cashMult)
			totalScore += 2500
			w.Police.ReduceStar(2)
			w.DifficultyTier++
			w.BossesDefeated++
			w.addPopup(st.X, st.Y-48, "BOSS BOARDED! TIER UP!", 4)
			if w.Boss.Kind == BossMom {
				w.Victory = true
				w.addPopup(st.X, st.Y-66, "LA MAMMA BOARDED! VICTORY!", 4)
			}
		} else {
			totalXP += int(80 * xpMult)
			totalCash += int(150 * cashMult)
			totalScore += 600
			w.addPopup(st.X, st.Y-48, fmt.Sprintf("BOSS BREAK! (%d/%d)", w.Boss.MaxShields-w.Boss.Shields, w.Boss.MaxShields), 4)
		}
	}

	if totalCash > 0 || totalXP > 0 {
		w.Player.Cash += totalCash
		w.Score += totalScore
		w.addPopup(st.X, st.Y-24, fmt.Sprintf("+$%d  +%dXP", totalCash, totalXP), 3)
		if w.Player.AddXP(totalXP) {
			w.triggerLevelUp()
		}
	}

	st.Active = false
}

func (w *World) checkTimedBossSpawns() {
	if w.Boss.Active {
		return
	}
	if !w.BossesSpawned[BossBestFriend] && w.RunTime >= 90.0 {
		w.ForceSpawnBoss(BossBestFriend)
	} else if !w.BossesSpawned[BossProfessor] && w.RunTime >= 210.0 {
		w.ForceSpawnBoss(BossProfessor)
	} else if !w.BossesSpawned[BossMom] && w.RunTime >= 330.0 {
		w.ForceSpawnBoss(BossMom)
	}
}

func (w *World) updateGirlSpawner(dt float32) {
	targetCap := 10 + int(w.RunTime/18.0) + (w.DifficultyTier-1)*4
	if targetCap > 42 {
		targetCap = 42
	}
	w.girlSpawnTimer -= dt
	if w.girlSpawnTimer <= 0 && len(w.Girls) < targetCap {
		interval := float32(2.4) - float32(w.DifficultyTier-1)*0.35 - w.RunTime*0.0025
		if interval < 0.75 {
			interval = 0.75
		}
		w.girlSpawnTimer = interval

		arch := Archetype(w.rng.Intn(int(ArchCount)))
		vipChance := float32(0.10)
		if w.DifficultyTier >= 2 {
			vipChance = 0.18
		}
		isVIP := w.rng.Float32() < vipChance
		w.spawnGirl(arch, isVIP)
	}
}

func (w *World) spawnGirl(arch Archetype, isVIP bool) {
	sx, sy := randomEdgePos(w.rng)
	g := NewGirl(w.allocID(), sx, sy, arch, isVIP, w.rng)
	w.constrainGirlPos(&g)
	w.Girls = append(w.Girls, g)
}

func (w *World) handleBulletInteractions() {
	// Check if Alfredo is standing inside a Kawaii stop for Tear-Graze
	var kawaiiStop *BusStop
	for s := range w.Stops {
		st := &w.Stops[s]
		if st.Active && st.Arch == ArchKawaii && st.PlayerInside {
			kawaiiStop = st
			break
		}
	}

	cancelBudget := 0
	for i := 0; i < w.Bullets.Count; {
		b := &w.Bullets.Items[i]
		d2 := distSq(b.X, b.Y, w.Player.X, w.Player.Y)
		hitR := b.Radius + PlayerHitboxRadius

		// Kawaii Tear-Graze check
		if kawaiiStop != nil && b.Kind == BulletTear && !b.Grazed {
			grazeR := kawaiiStop.GrazeRadius()
			if d2 <= grazeR*grazeR && d2 > hitR*hitR {
				b.Grazed = true
				kawaiiStop.Minigame.AddGrazeCharge()
				if kawaiiStop.InBusBay && !kawaiiStop.Minigame.Completed {
					kawaiiStop.Minigame.Progress += 0.25 * 0.25
					if kawaiiStop.Minigame.Progress >= 1.0 {
						kawaiiStop.Minigame.Progress = 1.0
						kawaiiStop.Minigame.Completed = true
					}
				}
				w.Score += 25
				if kawaiiStop.SynergyActive {
					cancelBudget += 2
				}
			}
		}

		// Player hit collision
		if !w.Player.IsInvulnerable() && d2 <= hitR*hitR {
			// Check Emo SOULMATE free hit absorb
			absorbed := false
			for s := range w.Stops {
				st := &w.Stops[s]
				if st.Active && st.PlayerInside && st.EmoAbsorbAvailable {
					st.EmoAbsorbAvailable = false
					absorbed = true
					w.addPopup(w.Player.X, w.Player.Y-24, "SOULMATE GUARD!", 1)
					break
				}
			}
			if absorbed {
				w.Bullets.RemoveAt(i)
				continue
			}

			kind := b.Kind
			w.Bullets.RemoveAt(i)
			w.damagePlayer(1, kind)
			continue
		}
		i++
	}

	// Kawaii MOE CRITICAL synergy cancels 2 nearby bullets per grazed droplet
	for cancelBudget > 0 && w.Bullets.Count > 0 {
		// Remove bullet closest to Alfredo
		bestIdx := -1
		bestD2 := float32(1e12)
		for i := 0; i < w.Bullets.Count; i++ {
			b := &w.Bullets.Items[i]
			d2 := distSq(b.X, b.Y, w.Player.X, w.Player.Y)
			if d2 < bestD2 {
				bestD2 = d2
				bestIdx = i
			}
		}
		if bestIdx == -1 {
			break
		}
		w.Bullets.RemoveAt(bestIdx)
		cancelBudget--
	}
}

func (w *World) damagePlayer(dmg int, kind BulletKind) {
	if !w.Player.TakeDamage(dmg) {
		return
	}
	w.ScreenShake = 8.5

	// Pragmata Grid punishment on hit
	for s := range w.Stops {
		st := &w.Stops[s]
		if st.Active && st.PlayerInside && st.Arch == ArchStudent {
			st.Minigame.ResetPragmataOnHit()
		}
	}

	switch kind {
	case BulletPolice:
		w.Player.SlowTimer = 1.8
		w.addPopup(w.Player.X, w.Player.Y-26, "VERBALE! SLOWED!", 2)
	case BulletCiabatta:
		w.Player.GroundedTimer = 8.0
		w.addPopup(w.Player.X, w.Player.Y-26, "IN PUNIZIONE! (8s)", 2)
	default:
		w.addPopup(w.Player.X, w.Player.Y-24, "-1 HP", 2)
	}

	if w.Player.HP <= 0 {
		w.GameOver = true
	}
}

func (w *World) updateShops(dt float32, interactPressed bool) {
	for i := range w.Shops {
		if w.Shops[i].Cooldown > 0 {
			w.Shops[i].Cooldown -= dt
			if w.Shops[i].Cooldown < 0 {
				w.Shops[i].Cooldown = 0
			}
		}
	}
	if !interactPressed {
		return
	}

	for i := range w.Shops {
		sh := &w.Shops[i]
		if sh.Cooldown > 0 {
			continue
		}
		if distSq(w.Player.X, w.Player.Y, sh.X, sh.Y) <= sh.Radius*sh.Radius {
			// Smart interaction: Bribe if high heat & affordable, else Heal/Chronos if needed, else Bribe
			if w.Police.Stars >= 2 && w.Player.Cash >= sh.BribeCost {
				w.Player.Cash -= sh.BribeCost
				w.Police.ReduceStar(2)
				sh.Cooldown = 0.6
				w.addPopup(sh.X, sh.Y-28, "BRIBE -$100 (-2★)", 3)
				return
			}
			if (w.Player.HP < w.Player.MaxHP || w.Player.ChronosMeter < w.Player.ChronosMax*0.85) && w.Player.Cash >= sh.HealCost {
				w.Player.Cash -= sh.HealCost
				w.Player.Heal(2)
				w.Player.ChronosMeter = w.Player.ChronosMax
				sh.Cooldown = 0.6
				w.addPopup(sh.X, sh.Y-28, "ESPRESSO +2HP & CHRONOS!", 1)
				return
			}
			if w.Police.Stars > 0 && w.Player.Cash >= sh.BribeCost {
				w.Player.Cash -= sh.BribeCost
				w.Police.ReduceStar(2)
				sh.Cooldown = 0.6
				w.addPopup(sh.X, sh.Y-28, "BRIBE -$100 (-2★)", 3)
				return
			}
			w.addPopup(sh.X, sh.Y-28, "NEED CASH ($50/$100)", 0)
			return
		}
	}
}

var allUpgradeCards = []UpgradeCard{
	{Kind: UpgradeCooldown, Title: "TURBO SCHEDULE", Desc: "-22% Stop Cooldown (min 1.5s)"},
	{Kind: UpgradeMultiHook, Title: "DOUBLE-DECKER HOOK", Desc: "+2 Max Girls per Stop (up to 12)"},
	{Kind: UpgradeMaxStops, Title: "EXPRESS NETWORK", Desc: "+1 Simultaneous Stop (up to 4)"},
	{Kind: UpgradeRange, Title: "WIDE SHELTER", Desc: "+15% Throw Range & Capture Radius"},
	{Kind: UpgradeChronos, Title: "KATANA RIZZ-FOCUS", Desc: "Chronos slower (-0.10x) & +25 Meter"},
	{Kind: UpgradeVitality, Title: "ESPRESSO OVERDRIVE", Desc: "+1 Max HP, Heal +2 HP, +12% Speed"},
}

func (w *World) triggerLevelUp() {
	w.PendingLevelUp = true
	// Fisher-Yates pick 3 distinct upgrade cards
	pool := [6]UpgradeCard{}
	copy(pool[:], allUpgradeCards)
	for i := len(pool) - 1; i > 0; i-- {
		j := w.rng.Intn(i + 1)
		pool[i], pool[j] = pool[j], pool[i]
	}
	w.UpgradeChoices[0] = pool[0]
	w.UpgradeChoices[1] = pool[1]
	w.UpgradeChoices[2] = pool[2]
}

func (w *World) applyUpgrade(card UpgradeCard) {
	switch card.Kind {
	case UpgradeCooldown:
		w.Player.StopCooldownBase *= 0.78
		if w.Player.StopCooldownBase < 1.5 {
			w.Player.StopCooldownBase = 1.5
		}
	case UpgradeMultiHook:
		w.Player.MaxMultiHook += 2
		if w.Player.MaxMultiHook > 12 {
			w.Player.MaxMultiHook = 12
		}
	case UpgradeMaxStops:
		if w.Player.MaxSimultaneousStops < 4 {
			w.Player.MaxSimultaneousStops++
		} else {
			w.Player.MaxMultiHook = clampInt(w.Player.MaxMultiHook+1, 1, 12)
		}
	case UpgradeRange:
		w.Player.RangeMult += 0.15
		if w.Player.RangeMult > 1.60 {
			w.Player.RangeMult = 1.60
		}
	case UpgradeChronos:
		w.Player.ChronosTimeScale -= 0.10
		if w.Player.ChronosTimeScale < 0.20 {
			w.Player.ChronosTimeScale = 0.20
		}
		w.Player.ChronosMax += 25
		w.Player.ChronosMeter = w.Player.ChronosMax
	case UpgradeVitality:
		w.Player.MaxHP++
		w.Player.Heal(2)
		w.Player.SpeedMult += 0.12
	}
	w.addPopup(w.Player.X, w.Player.Y-32, card.Title+"!", 1)
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
