package game

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// World is the complete, deterministic pure-Go game simulation state.
type World struct {
	Player       Player
	Stops        []BusStop
	Girls        []Girl
	Bullets      BulletPool
	Police       PoliceSystem
	Boss         Boss
	Buses        []NeonBus
	ZTLZones     []Rect
	Shops        []VendingShop
	SlotMachines []SlotMachine
	Popups       []FloatingText
	Telegraphs   []TelegraphLine
	Express      []ExpressLine
	WaveRings    []WaveRing
	Slipstreams  []BusSlipstream

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
	MusicBPM          float32
	StreakCount       int
	StreakTimer       float32
	MaxStreak         int
	Score             int
	GirlsBoarded      int
	DifficultyTier    int
	BossRushMode      bool
	BossSchedule      []BossKind
	BossScheduleTimes []float32
	BossScheduleIdx   int
	NextBossRushTimer float32
	BossesSpawned     [BossCount]bool
	BossesDefeated    int
	GameOver          bool
	Victory           bool

	PendingLevelUp bool
	UpgradeChoices [3]UpgradeCard
	ShopMenu       ShopMenuState
	SlotModal      SlotMachineState

	WorldEvents []WorldEvent

	CameraX, CameraY float32

	girlSpawnTimer float32
	nextEntityID   int
	rng            *rand.Rand
}

// NewWorld creates a new normal run using a time-based seed.
func NewWorld() *World {
	return NewWorldWithOptions(time.Now().UnixNano(), false)
}

// NewWorldWithSeed creates a deterministic normal World from the given RNG seed.
func NewWorldWithSeed(seed int64) *World {
	return NewWorldWithOptions(seed, false)
}

// NewWorldWithOptions creates a deterministic game World for normal or Boss Rush mode.
func NewWorldWithOptions(seed int64, bossRush bool) *World {
	layout := BuildDistrictLayout()
	rng := rand.New(rand.NewSource(seed))

	w := &World{
		Player:         NewPlayer(),
		Stops:          make([]BusStop, 0, 8),
		Girls:          make([]Girl, 0, 48),
		Police:         NewPoliceSystem(),
		Buses:          make([]NeonBus, 0, 4),
		ZTLZones:       layout.ZTLZones,
		Shops:          layout.Shops,
		SlotMachines:   layout.SlotMachines,
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
		Slipstreams:    make([]BusSlipstream, 0, 4),
		WorldEvents:    make([]WorldEvent, 0, 32),
		MusicBPM:       108.0,
		DifficultyTier: 1,
		BossRushMode:   bossRush,
		girlSpawnTimer: 1.5,
		nextEntityID:   1,
		rng:            rng,
	}

	if bossRush {
		w.BossSchedule = []BossKind{
			BossBestFriend,
			BossLover,
			BossProfessor,
			BossExecutive,
			BossRector,
			BossMom,
		}
		w.NextBossRushTimer = 8.0
	} else {
		// Rogue-lite 4-of-6 Boss Pool: shuffle the 4 non-final bosses, take 3, then pick Final Boss A or B
		pool := [4]BossKind{BossBestFriend, BossLover, BossProfessor, BossExecutive}
		for i := len(pool) - 1; i > 0; i-- {
			j := rng.Intn(i + 1)
			pool[i], pool[j] = pool[j], pool[i]
		}
		finalBoss := BossRector
		if rng.Intn(2) == 1 {
			finalBoss = BossMom
		}
		w.BossSchedule = []BossKind{pool[0], pool[1], pool[2], finalBoss}
		w.BossScheduleTimes = []float32{105.0, 225.0, 355.0, 485.0}
	}

	// Seed initial horde across all 4 archetypes
	for i := 0; i < 8; i++ {
		arch := Archetype(i % int(ArchCount))
		w.spawnGirl(arch, i == 7)
	}

	w.snapCamera()
	return w
}

func (w *World) emitEvent(kind WorldEventKind, val int, flag bool) {
	if len(w.WorldEvents) < 64 {
		w.WorldEvents = append(w.WorldEvents, WorldEvent{Kind: kind, Value: val, Flag: flag})
	}
}

// DrainEvents returns and clears all simulation events emitted since the previous call.
func (w *World) DrainEvents() []WorldEvent {
	if len(w.WorldEvents) == 0 {
		return nil
	}
	out := make([]WorldEvent, len(w.WorldEvents))
	copy(out, w.WorldEvents)
	w.WorldEvents = w.WorldEvents[:0]
	return out
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

// ForceSpawnBoss immediately spawns a Boss of the requested kind.
func (w *World) ForceSpawnBoss(kind BossKind) {
	if w.Boss.Active {
		return
	}
	sx, sy := randomEdgePos(w.rng)
	w.Boss.Spawn(kind, sx, sy)
	if kind >= 0 && kind < BossCount {
		w.BossesSpawned[kind] = true
	}
	w.addPopup(w.Player.X, w.Player.Y-36, "BOSS: "+w.Boss.Name+"!", 4)
	w.emitEvent(EventBossSpawn, int(kind), false)
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

	// Vending Shop mini-menu pauses simulation cleanly
	if w.ShopMenu.Active {
		w.HandleShopInput(in)
		return
	}

	// Rizz-Slots 777 modal pauses simulation cleanly
	if w.SlotModal.Active {
		w.UpdateSlotMachine(dt, in)
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

	if w.StreakTimer > 0 {
		w.StreakTimer -= dt
		if w.StreakTimer <= 0 {
			w.StreakTimer = 0
			w.StreakCount = 0
		}
	}

	w.updateVFX(dt)
	w.updateSlipstreams(dt)
	w.updateExpressLanes()

	inExpress := w.isPointInExpressLane(w.Player.X, w.Player.Y)
	inEmoFog := w.isPlayerInEmoFog()

	// Update Alfredo
	bassline, dashed := w.Player.Update(in, dt, inExpress, inEmoFog, w.StreakCount)
	if bassline {
		w.triggerBasslinePulse()
	}
	if dashed {
		w.emitEvent(EventDash, 0, false)
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

	// Smart Magnetic Aim, Tactical Stop Recall, and Throw Input Buffer (120 ms)
	aimX, aimY := w.ResolveAim(in)
	recalled := false
	if in.RecallJustPressed {
		recalled = w.tryRecallStop(aimX, aimY, true)
	}
	if !recalled && in.ThrowJustPressed {
		if w.ActiveStopCount() >= w.Player.EffectiveMaxStops() && w.tryRecallStop(aimX, aimY, false) {
			recalled = true
		} else {
			w.Player.ThrowBufferTimer = InputBufferDuration
		}
	}
	if !recalled && w.Player.ThrowBufferTimer > 0 && w.Player.StunTimer <= 0 && w.Player.StopCooldownTimer <= 0 && w.ActiveStopCount() < w.Player.EffectiveMaxStops() {
		w.Player.ThrowBufferTimer = 0
		w.tryThrowBusStop(aimX, aimY)
	}

	// Update Stops, Minigames & Boarding Combos
	w.updateStops(dt, in.ActionDir, in.ActionHeld)

	// Update Neon Buses
	w.updateBuses(dt)

	// Boss Gatekeeper schedule
	w.checkTimedBossSpawns(dt)

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

	// Vending Machine Shop & Slot Machine interaction [E]
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
			// Professor / Rector Laser fires instant check when telegraph completes
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

func (w *World) updateSlipstreams(dt float32) {
	for i := 0; i < len(w.Slipstreams); {
		s := &w.Slipstreams[i]
		s.Life -= dt
		if s.Life <= 0 {
			w.Slipstreams[i] = w.Slipstreams[len(w.Slipstreams)-1]
			w.Slipstreams = w.Slipstreams[:len(w.Slipstreams)-1]
			continue
		}
		if float32(math.Abs(float64(w.Player.Y-s.Y))) <= 26.0 && w.Player.X >= s.X1 && w.Player.X <= s.X2 {
			w.Player.SlipstreamTimer = 2.0
			if w.Player.DashCooldown > 0 {
				w.Player.DashCooldown = 0
			}
			if w.Player.SlipstreamPopupCD <= 0 {
				w.Player.SlipstreamPopupCD = 1.6
				w.addPopup(w.Player.X, w.Player.Y-28, "SLIPSTREAM! +VEL & DASH!", 1)
				w.emitEvent(EventSlipstream, 0, false)
			}
		}
		i++
	}
}

func (w *World) triggerBasslinePulse() {
	w.addWaveRing(w.Player.X, w.Player.Y, BasslinePulseRadius, 0)
	rSq := BasslinePulseRadius * BasslinePulseRadius

	// Herding & Magnet-Beat: push Police officers away
	for i := range w.Police.Officers {
		of := &w.Police.Officers[i]
		if distSq(of.X, of.Y, w.Player.X, w.Player.Y) <= rSq {
			dx, dy, _ := normalize2D(of.X-w.Player.X, of.Y-w.Player.Y)
			of.X = clamp32(of.X+dx*28, 20, ArenaWidth-20)
			of.Y = clamp32(of.Y+dy*28, 24, w.walkableMaxY())
		}
	}

	// Gently pull matching-archetype girls inward to cluster Synergy drops; push others back
	pullMax := BasslinePulseRadius * 1.45
	pullMaxSq := pullMax * pullMax
	for i := range w.Girls {
		g := &w.Girls[i]
		if g.HookedByStop != -1 {
			continue
		}
		d2 := distSq(g.X, g.Y, w.Player.X, w.Player.Y)
		if g.Arch == w.Player.SelectedStopArch {
			if d2 > 42*42 && d2 <= pullMaxSq {
				dx, dy, _ := normalize2D(w.Player.X-g.X, w.Player.Y-g.Y)
				g.X += dx * 18
				g.Y += dy * 18
				w.constrainGirlPos(g)
			}
		} else if d2 <= rSq {
			dx, dy, _ := normalize2D(g.X-w.Player.X, g.Y-w.Player.Y)
			g.X += dx * 20
			g.Y += dy * 20
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
	if w.Boss.Active && w.Boss.MartialZTLActive && w.Boss.MartialZTLRect.Contains(x, y) {
		return true
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

func (w *World) currentBPM() float32 {
	if w.MusicBPM > 0 {
		return w.MusicBPM
	}
	switch {
	case w.DifficultyTier <= 1:
		return 108.0
	case w.DifficultyTier == 2:
		return 118.0
	case w.DifficultyTier == 3:
		return 128.0
	default:
		return 138.0
	}
}

// ResolveAim resolves world-space throw coordinates, applying Smart Magnetic Aim
// in pure-keyboard or twin-stick gamepad directional modes.
func (w *World) ResolveAim(in InputState) (float32, float32) {
	if !in.AimIsDirectional {
		return in.AimWorldX, in.AimWorldY
	}

	dirX, dirY, dLen := normalize2D(in.AimDirX, in.AimDirY)
	if dLen <= 0.001 {
		dirX, dirY = w.Player.FacingX, 0
	}

	spec := GetArchetypeSpec(w.Player.SelectedStopArch)
	maxRange := spec.ThrowRange * w.Player.EffectiveRangeMult()
	searchMax := maxRange + 12.0

	found := false
	bestScore := float32(-1e9)
	bestX, bestY := float32(0), float32(0)

	evalCandidate := func(cx, cy, bonus float32) {
		dx, dy, dVal := normalize2D(cx-w.Player.X, cy-w.Player.Y)
		if dVal > searchMax {
			return
		}
		dot := float32(1.0)
		if dVal >= 1.0 {
			dot = dx*dirX + dy*dirY
		}
		if dot < 0.35 {
			return
		}
		score := dot*120.0 - dVal*0.25 + bonus
		if score > bestScore {
			bestScore = score
			bestX = cx
			bestY = cy
			found = true
		}
	}

	for i := range w.Girls {
		g := &w.Girls[i]
		if g.HookedByStop != -1 {
			continue
		}
		bonus := float32(0)
		if g.Arch == w.Player.SelectedStopArch {
			bonus = 60.0
		}
		evalCandidate(g.X, g.Y, bonus)
	}
	if w.Boss.Active && w.Boss.HookedByStop == -1 {
		evalCandidate(w.Boss.X, w.Boss.Y, 40.0)
	}

	if found {
		return bestX, bestY
	}
	return clamp32(w.Player.X+dirX*maxRange*0.65, 48, ArenaWidth-48),
		clamp32(w.Player.Y+dirY*maxRange*0.65, 76, w.walkableMaxY()-8)
}

func (w *World) tryRecallStop(aimX, aimY float32, anyNearest bool) bool {
	bestIdx := -1
	bestD2 := float32(1e12)
	const maxAimSq = float32(44.0 * 44.0)

	for i := range w.Stops {
		st := &w.Stops[i]
		if !st.Active || st.BusRushTriggered {
			continue
		}
		if !anyNearest && st.Combo.Active {
			continue
		}
		refX, refY := aimX, aimY
		if anyNearest {
			refX, refY = w.Player.X, w.Player.Y
		}
		d2 := distSq(st.X, st.Y, refX, refY)
		if !anyNearest && d2 > maxAimSq {
			continue
		}
		if d2 < bestD2 {
			bestD2 = d2
			bestIdx = i
		}
	}
	if bestIdx == -1 {
		return false
	}

	st := &w.Stops[bestIdx]
	for _, gid := range st.HookedGirlIDs {
		for j := range w.Girls {
			if w.Girls[j].ID == gid {
				w.Girls[j].HookedByStop = -1
				break
			}
		}
	}
	if st.HookedBoss && w.Boss.Active && w.Boss.HookedByStop == st.ID {
		w.Boss.HookedByStop = -1
	}

	st.Active = false
	w.Player.StopCooldownTimer *= 0.50
	w.Bullets.ClearRadius(st.X, st.Y, 95.0)
	w.addWaveRing(st.X, st.Y, 95.0, 1)
	w.addPopup(st.X, st.Y-26, "RICHIAMO! -50% CD", 1)
	w.emitEvent(EventStopRecall, 0, false)
	return true
}

// EvaluateAim computes real-time tactical placement feedback for the UI aim reticle.
func (w *World) EvaluateAim(aimX, aimY float32) AimFeedback {
	spec := GetArchetypeSpec(w.Player.SelectedStopArch)
	rangeMult := w.Player.EffectiveRangeMult()
	maxRange := spec.ThrowRange * rangeMult
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

	previewR := spec.CaptureRadius * rangeMult
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

	// Check L'Amante holographic clone decoy trap!
	if w.Boss.Active && w.Boss.Kind == BossLover && w.Boss.HookedByStop == -1 {
		bossR := fb.PreviewRadius + 24.0
		realHit := distSq(tx, ty, w.Boss.X, w.Boss.Y) <= bossR*bossR
		if !realHit {
			decoyR := fb.PreviewRadius + 18.0
			for c := range w.Boss.Clones {
				cl := &w.Boss.Clones[c]
				if cl.Active && distSq(tx, ty, cl.X, cl.Y) <= decoyR*decoyR {
					cl.Active = false
					w.Player.StopCooldownTimer = w.Player.EffectiveStopCooldown() * 0.65
					w.Player.ThrowAnimTimer = 0.22
					w.ScreenShake = 5.0
					w.addWaveRing(cl.X, cl.Y, 96.0, 2)
					for k := 0; k < 8; k++ {
						ang := float32(k) * (6.283 / 8.0)
						w.Bullets.Spawn(Bullet{
							X:      cl.X,
							Y:      cl.Y,
							VX:     float32(math.Cos(float64(ang))) * 155,
							VY:     float32(math.Sin(float64(ang))) * 155,
							Radius: 4.5,
							MaxAge: 3.4,
							Kind:   BulletLipstick,
						})
					}
					w.addPopup(cl.X, cl.Y-28, "ESCA OLOGRAMMA! +CD!", 2)
					w.emitEvent(EventMinigameMiss, 0, false)
					return
				}
			}
		}
	}

	stop := NewBusStop(w.allocID(), tx, ty, w.Player.SelectedStopArch, w.Player.EffectiveRangeMult(), inZTL, inBusBay, perfectDrop, w.rng)
	shieldBonus := w.Player.ShieldBonus()
	stop.ShieldHP += shieldBonus
	stop.MaxShieldHP += shieldBonus

	w.Player.StopCooldownTimer = w.Player.EffectiveStopCooldown()
	w.Player.ThrowAnimTimer = 0.22

	if perfectDrop {
		cdFactor := float32(0.65)
		if w.StreakCount >= 5 {
			cdFactor = 0.50
		}
		w.Player.StopCooldownTimer *= cdFactor
		w.HitStopTimer = 0.090 // 90ms Perfect Drop Hit-Stop
		w.ScreenShake = 6.5
		clearR := float32(130.0)
		if w.Player.HasRelic(RelicArmoredGlass) {
			clearR += 35.0
		}
		w.Bullets.ClearRadius(tx, ty, clearR)
		w.addWaveRing(tx, ty, clearR, 1)
		w.addPopup(tx, ty-28, "PERFECT DROP!", 1)
		w.emitEvent(EventPerfectDrop, 0, true)
	} else {
		w.HitStopTimer = 0.045 // 45ms standard hook Hit-Stop
		w.ScreenShake = 3.0
		w.emitEvent(EventThrow, 0, false)
	}

	if inZTL {
		added := w.Police.AddStar(1)
		if added > 0 {
			w.addPopup(tx, ty-44, "ZTL! +1* WANTED!", 2)
			w.emitEvent(EventPoliceStarUp, w.Police.Stars, false)
		} else {
			w.addPopup(tx, ty-44, "BONUS ZTL 3x!", 2)
		}
	} else if inBusBay {
		w.addPopup(tx, ty-44, "BUS BAY +25% VELOCITA!", 1)
	}

	// La Rettrice Martial Law: non-perfect throws raise Wanted Stars!
	if w.Boss.Active && w.Boss.Kind == BossRector && !perfectDrop {
		if w.Police.AddStar(1) > 0 {
			w.addPopup(tx, ty-58, "LEGGE MARZIALE! +1*!", 2)
			w.emitEvent(EventPoliceStarUp, w.Police.Stars, false)
		}
	}

	// Hook Boss if inside capture radius
	if w.Boss.Active && w.Boss.HookedByStop == -1 {
		bossR := stop.CaptureRadius + 24.0
		if distSq(tx, ty, w.Boss.X, w.Boss.Y) <= bossR*bossR {
			stop.HookedBoss = true
			w.Boss.HookedByStop = stop.ID
			w.addPopup(w.Boss.X, w.Boss.Y-32, "BOSS AGGANCIATO!", 4)
			if w.Boss.Kind == BossLover {
				// Hooking the real Lover temporarily disrupts her decoys
				w.Boss.Clones[0].Active = false
				w.Boss.Clones[1].Active = false
			}
		}
	}

	// Hook up to EffectiveMaxMultiHook closest girls inside CaptureRadius
	capSq := stop.CaptureRadius * stop.CaptureRadius
	hookedVIP := false
	maxHook := w.Player.EffectiveMaxMultiHook()
	for len(stop.HookedGirlIDs) < maxHook {
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

	if len(stop.HookedGirlIDs) > 0 || stop.HookedBoss {
		w.emitEvent(EventHook, len(stop.HookedGirlIDs), perfectDrop)
	}

	if hookedVIP {
		if w.Police.AddStar(1) > 0 {
			w.addPopup(tx, ty-56, "BERSAGLIO VIP! +1*!", 2)
			w.emitEvent(EventPoliceStarUp, w.Police.Stars, false)
		}
	}

	if stop.SynergyActive {
		w.addPopup(tx, ty-16, spec.SynergyName+"!", 1)
		w.emitEvent(EventSynergy, int(stop.Arch), false)
		switch stop.Arch {
		case ArchEmo:
			w.Player.Heal(1) // Black Heart heal
		case ArchStudent:
			for i := range w.Girls {
				if w.Girls[i].Arch == ArchStudent && distSq(tx, ty, w.Girls[i].X, w.Girls[i].Y) <= capSq {
					w.Girls[i].StunTimer = 1.2
				}
			}
		case ArchGyaru:
			flashR := stop.CaptureRadius * 1.3
			w.addWaveRing(tx, ty, flashR, 4)
			for i := range w.Girls {
				if distSq(tx, ty, w.Girls[i].X, w.Girls[i].Y) <= flashR*flashR {
					w.Girls[i].StunTimer = 2.0
				}
			}
		}
	}

	for b := range w.Buses {
		bus := &w.Buses[b]
		if bus.Active && bus.X < tx && tx-bus.X <= 420.0 && float32(math.Abs(float64(bus.Y-ty))) <= 52.0 {
			if stop.Minigame.Progress < 0.50 {
				stop.Minigame.Progress = 0.50
			}
			stop.AlVoloBoosted = true
			w.addPopup(tx, ty-42, "AL VOLO! +50%!", 1)
			break
		}
	}

	w.Stops = append(w.Stops, stop)
}

func (w *World) updateStops(dt float32, actionDir int, actionHeld bool) {
	primaryStopIdx := -1
	bestDistSq := float32(1e12)
	for i := range w.Stops {
		st := &w.Stops[i]
		if !st.Active || st.BusRushTriggered {
			continue
		}
		d2 := distSq(w.Player.X, w.Player.Y, st.X, st.Y)
		rawInside := d2 <= st.CaptureRadius*st.CaptureRadius
		if rawInside {
			st.CoyoteTimer = TetherCoyoteDuration
		} else if st.CoyoteTimer > 0 {
			st.CoyoteTimer -= dt
			if st.CoyoteTimer < 0 {
				st.CoyoteTimer = 0
			}
		}
		st.PlayerInside = rawInside || st.CoyoteTimer > 0
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
		held := false
		if i == primaryStopIdx {
			dir = actionDir
			held = actionHeld
		}

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
			mgDT := dt * w.Player.MinigameSpeedMult()
			if w.StreakCount >= 3 {
				mgDT *= 1.15
			}
			if st.InBusBay && (st.Minigame.Arch == ArchKawaii || st.Minigame.Arch == ArchGyaru) {
				mgDT *= 1.25
			}
			prevProg := st.Minigame.Progress
			prevGX, prevGY := st.Minigame.GridX, st.Minigame.GridY
			st.Minigame.Update(mgDT, st.PlayerInside, dir, held, w.currentBPM(), w.rng)

			if st.Minigame.LastHitResult == 1 {
				st.Minigame.LastHitResult = 0
				w.emitEvent(EventMinigameHit, 0, false)
			} else if st.Minigame.LastHitResult == -1 {
				st.Minigame.LastHitResult = 0
				w.emitEvent(EventMinigameMiss, 0, false)
			}
			if st.Minigame.Arch == ArchStudent && (st.Minigame.GridX != prevGX || st.Minigame.GridY != prevGY) {
				w.emitEvent(EventMinigameHit, 0, false)
			}

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
				xpGain := int(20 * xpMult * w.Player.XPBonusMult())
				w.addPopup(st.X, st.Y-24, fmt.Sprintf("+%d XP NODO!", xpGain), 3)
				if w.Player.AddXP(xpGain) {
					w.triggerLevelUp()
				}
			}
			if st.Minigame.FreezeTriggered {
				st.Minigame.FreezeTriggered = false
				w.BulletFreezeTimer = 1.5
				w.addPopup(st.X, st.Y-38, "PROIETTILI CONGELATI!", 1)
			}
		}

		if st.Minigame.Completed && !st.Combo.Active && !st.Combo.Completed && !st.Combo.Failed {
			strict := st.HookedBoss && w.Boss.Active && w.Boss.Kind == BossProfessor
			st.Combo.Start(st.Arch == ArchEmo && st.SynergyActive, strict, w.rng)
		} else if st.Combo.Active && !st.Combo.Completed && !st.Combo.Failed {
			prevStep := st.Combo.ComboIndex
			prevErr := st.Combo.LastErrorFlash
			st.Combo.Update(dt, st.PlayerInside, dir)
			if st.Combo.ComboIndex > prevStep {
				w.emitEvent(EventComboStep, prevStep, false)
			} else if st.Combo.LastErrorFlash > prevErr {
				w.emitEvent(EventMinigameMiss, 0, false)
			}

			if st.Combo.Completed {
				st.BusRushTriggered = true
				w.Buses = append(w.Buses, NewNeonBus(st))
				w.addPopup(st.X, st.Y-32, "ARRIVA IL NEON BUS!", 1)
				w.emitEvent(EventBusRush, int(st.Arch), false)
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
		w.addPopup(st.X, st.Y-24, "SCAPPATA!", 2)
	}
	if w.StreakCount > 0 {
		w.StreakCount /= 2
		if w.StreakCount == 0 {
			w.StreakTimer = 0
		}
	}
	w.emitEvent(EventMinigameMiss, 0, false)
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

		// Neon Bus Purification Wave clears bullets along its streak (Rush Hour converts to Cash!)
		cleared := w.Bullets.ClearLine(prevX, bus.Y, bus.X+42, bus.Y, 46.0)
		if cleared > 0 && w.StreakCount >= 5 {
			w.Player.Cash += cleared * 3
		}

		// Spawn/refresh aerodynamic Bus Slipstream ribbon along the streak lane
		if bus.X > 0 {
			x2 := clamp32(bus.X, 0, ArenaWidth)
			matched := false
			for s := range w.Slipstreams {
				if float32(math.Abs(float64(w.Slipstreams[s].Y-bus.Y))) < 1.0 {
					if x2 > w.Slipstreams[s].X2 {
						w.Slipstreams[s].X2 = x2
					}
					w.Slipstreams[s].Life = BusSlipstreamDuration
					w.Slipstreams[s].MaxLife = BusSlipstreamDuration
					matched = true
					break
				}
			}
			if !matched {
				w.Slipstreams = append(w.Slipstreams, BusSlipstream{
					X1:      0,
					X2:      x2,
					Y:       bus.Y,
					Life:    BusSlipstreamDuration,
					MaxLife: BusSlipstreamDuration,
				})
			}
		}

		// Knock aside unhooked girls in bus lane
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
	xpMult *= w.Player.XPBonusMult()
	cashMult *= w.Player.CashBonusMult()
	totalXP := 0
	totalCash := 0
	totalScore := 0
	boardedHere := 0

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
				boardedHere++

				// Swap-remove girl
				w.Girls[i] = w.Girls[len(w.Girls)-1]
				w.Girls = w.Girls[:len(w.Girls)-1]
				break
			}
		}
	}

	if boardedHere > 0 {
		w.emitEvent(EventGirlBoarded, boardedHere, false)
	}

	// Boss 3-Stop Break!
	bossBroken := false
	if st.HookedBoss && w.Boss.Active && w.Boss.HookedByStop == st.ID {
		canBreak := true

		// La Dirigente Aziendale gimmick: require at least 1 escort caught in stop radius if escorts remain
		if w.Boss.Kind == BossExecutive && w.Boss.ActiveEscortCount() > 0 {
			escortR := st.CaptureRadius + 28.0
			escortSq := escortR * escortR
			escortsCaught := 0
			for e := range w.Boss.Escorts {
				es := &w.Boss.Escorts[e]
				if es.Active && distSq(st.X, st.Y, es.X, es.Y) <= escortSq {
					es.Active = false
					escortsCaught++
				}
			}
			if escortsCaught == 0 {
				canBreak = false
				w.Boss.HookedByStop = -1
				w.addPopup(w.Boss.X, w.Boss.Y-36, "SCORTA ATTIVA! PRENDI LE STAGISTE!", 2)
				w.emitEvent(EventMinigameMiss, 0, false)
			}
		}

		if canBreak {
			bossBroken = true
			w.HitStopTimer = 0.140 // 140ms Boss Break Hit-Stop
			w.ScreenShake = 11.5
			w.Bullets.ClearRadius(w.Boss.X, w.Boss.Y, 240)
			bossKind := w.Boss.Kind
			defeated := w.Boss.BreakShield()
			if defeated {
				starsAtKill := w.Police.Stars
				totalXP += int(240 * xpMult)
				totalCash += int(500 * cashMult)
				totalScore += 2500
				w.Police.ReduceStar(2)
				w.DifficultyTier++
				w.BossesDefeated++
				w.addPopup(st.X, st.Y-48, "BOSS A BORDO! TIER UP!", 4)
				w.emitEvent(EventBossBreak, int(bossKind), true)
				w.emitEvent(EventBossDefeated, int(bossKind), starsAtKill >= 3)

				if w.BossRushMode {
					if w.BossesDefeated >= len(w.BossSchedule) {
						w.Victory = true
						w.addPopup(st.X, st.Y-66, "BOSS RUSH COMPLETATA!", 4)
					} else {
						w.NextBossRushTimer = 12.0
					}
				} else if bossKind == BossMom || bossKind == BossRector || w.BossesDefeated >= len(w.BossSchedule) {
					w.Victory = true
					w.addPopup(st.X, st.Y-66, "VITTORIA! CORSA COMPLETATA!", 4)
				}
			} else {
				totalXP += int(80 * xpMult)
				totalCash += int(150 * cashMult)
				totalScore += 600
				w.addPopup(st.X, st.Y-48, fmt.Sprintf("SCUDO ROTTO! (%d/%d)", w.Boss.MaxShields-w.Boss.Shields, w.Boss.MaxShields), 4)
				w.emitEvent(EventBossBreak, int(bossKind), false)
			}
		}
	}

	if boardedHere > 0 || bossBroken {
		w.StreakCount++
		w.StreakTimer = StreakWindowDuration
		if w.StreakCount > w.MaxStreak {
			w.MaxStreak = w.StreakCount
		}
		w.emitEvent(EventFeverUp, w.StreakCount, false)
		if w.StreakCount == 3 {
			w.addPopup(st.X, st.Y-40, "STREAK x3!", 1)
		} else if w.StreakCount >= 5 {
			w.addPopup(st.X, st.Y-40, fmt.Sprintf("RUSH HOUR x%d!", w.StreakCount), 4)
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

func (w *World) checkTimedBossSpawns(dt float32) {
	if w.Boss.Active || w.BossScheduleIdx >= len(w.BossSchedule) {
		return
	}
	if w.BossRushMode {
		w.NextBossRushTimer -= dt
		if w.NextBossRushTimer <= 0 {
			w.ForceSpawnBoss(w.BossSchedule[w.BossScheduleIdx])
			w.BossScheduleIdx++
		}
		return
	}
	if w.BossScheduleIdx < len(w.BossScheduleTimes) && w.RunTime >= w.BossScheduleTimes[w.BossScheduleIdx] {
		w.ForceSpawnBoss(w.BossSchedule[w.BossScheduleIdx])
		w.BossScheduleIdx++
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

		// Dash-Parry: dashing through a bullet destroys it, triggers micro hit-stop, and shaves cooldown
		if w.Player.DashTimer > 0 {
			parryR := hitR + 6.0
			if d2 <= parryR*parryR {
				w.Bullets.RemoveAt(i)
				w.Score += 40
				w.Player.ChronosMeter = clamp32(w.Player.ChronosMeter+10.0, 0, w.Player.ChronosMax)
				if !w.Player.DashParryHitStopDone {
					w.Player.DashParryHitStopDone = true
					if w.HitStopTimer < 0.025 {
						w.HitStopTimer = 0.025
					}
					w.addWaveRing(w.Player.X, w.Player.Y, 36.0, 1)
					w.addPopup(w.Player.X, w.Player.Y-26, "PARRY! -1.0s CD", 1)
					w.emitEvent(EventDashParry, 0, false)
				}
				if w.Player.DashParryShaveUsed < 1.5 {
					shave := float32(1.0)
					if w.Player.DashParryShaveUsed > 0 {
						shave = 0.5
					}
					if w.Player.DashParryShaveUsed+shave > 1.5 {
						shave = 1.5 - w.Player.DashParryShaveUsed
					}
					w.Player.ReduceStopCooldown(shave)
					w.Player.DashParryShaveUsed += shave
				}
				continue
			}
		}

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

		// Global Bullet Graze (<= 18px): restores +4 Chronos and shaves -0.25s Stop Cooldown
		if !b.GlobalGrazed && !w.Player.IsInvulnerable() {
			gR := b.Radius + GlobalGrazeRadius
			if d2 <= gR*gR && d2 > hitR*hitR {
				b.GlobalGrazed = true
				w.Player.ChronosMeter = clamp32(w.Player.ChronosMeter+4.0, 0, w.Player.ChronosMax)
				w.Player.ReduceStopCooldown(0.25)
				w.Score += 15
				if w.StreakCount > 0 {
					w.StreakTimer = clamp32(w.StreakTimer+0.35, 0, StreakWindowDuration)
				}
				w.emitEvent(EventGraze, 0, false)
			}
		}

		// Player hit collision
		if !w.Player.IsInvulnerable() && d2 <= hitR*hitR {
			absorbed := false
			for s := range w.Stops {
				st := &w.Stops[s]
				if st.Active && st.PlayerInside && st.EmoAbsorbAvailable {
					st.EmoAbsorbAvailable = false
					absorbed = true
					w.addPopup(w.Player.X, w.Player.Y-24, "SCUDO SOULMATE!", 1)
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
	w.StreakCount = 0
	w.StreakTimer = 0
	w.ScreenShake = 8.5
	w.emitEvent(EventPlayerHit, int(kind), kind == BulletCiabatta)

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
		w.addPopup(w.Player.X, w.Player.Y-26, "VERBALE! RALLENTATO!", 2)
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
			w.OpenShop(i)
			return
		}
	}

	for i := range w.SlotMachines {
		sm := &w.SlotMachines[i]
		if distSq(w.Player.X, w.Player.Y, sm.X, sm.Y) <= sm.Radius*sm.Radius {
			w.OpenSlotMachine(i)
			return
		}
	}
}

var allUpgradeCards = []UpgradeCard{
	{Kind: UpgradeCooldown, Title: "ORARI TURBO", Desc: "-22% Cooldown Fermate (min 1.5s)"},
	{Kind: UpgradeMultiHook, Title: "GANCIO DOPPIO PIANO", Desc: "+2 Ragazze per Fermata (max 12)"},
	{Kind: UpgradeMaxStops, Title: "RETE EXPRESS", Desc: "+1 Fermata Simultanea (max 4)"},
	{Kind: UpgradeRange, Title: "PENSILINA XL", Desc: "+15% Gittata Lancio & Raggio"},
	{Kind: UpgradeChronos, Title: "KATANA RIZZ-FOCUS", Desc: "Chronos piu lento (-0.10x) & +25 Barra"},
	{Kind: UpgradeVitality, Title: "ESPRESSO OVERDRIVE", Desc: "+1 Max HP, Cura +2 HP, +12% Vel."},
}

func (w *World) triggerLevelUp() {
	w.PendingLevelUp = true
	w.emitEvent(EventLevelUp, w.Player.Level, false)
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
