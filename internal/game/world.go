package game

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"girlbusAG/internal/meta"
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
	Chests       []NeonChest
	Popups       []FloatingText
	Telegraphs   []TelegraphLine
	Express      []ExpressLine
	WaveRings    []WaveRing
	Slipstreams  []BusSlipstream
	HealPickups  []HealPickup

	Props         []MapProp
	Lights        []LightSource
	ParkIslands   []Rect
	BusBays       []Rect
	CrosswalkCols []int
	CrosswalkRows []int
	CanalY        float32
	BoardwalkY    float32
	MapID         string
	MapName       string
	ChestSpots    [][2]float32

	HitStopTimer           float32
	ScreenShake            float32
	BulletFreezeTimer      float32
	Seed                   int64
	RunTime                float32
	MusicBPM               float32
	StreakCount            int
	StreakTimer            float32
	MaxStreak              int
	Score                  int
	GirlsBoarded           int
	BoardedByArch          [ArchCount]int
	PeakWanted             int
	PerfectDrops           int
	EventsCompleted        int
	DifficultyTier         int
	BossRushMode           bool
	BossSchedule           []BossKind
	BossScheduleTimes      []float32
	BossScheduleIdx        int
	NextBossRushTimer      float32
	BossesSpawned          [BossCount]bool
	BossesDefeated         int
	ChestsOpened           int
	OverdriveActive        bool
	OverdriveWave          int
	NextOverdriveBossTimer float32
	BossIntroTimer         float32
	DistrictEvent          DistrictEventState
	chestRestockTimer      float32
	GameOver               bool
	Victory                bool

	QuipSpeaker       string
	QuipText          string
	QuipTimer         float32
	QuipMaxTimer      float32
	QuipAccent        int
	synergyQuipCD     float32
	FirstBoardingDone bool
	EarlyMissTipShown bool

	TutorialMode         bool
	TutorialStep         int
	TutMoved             bool
	TutThrewStop         bool
	TutCompletedMinigame bool
	TutBoardedBus        bool
	TutUsedSlotOrShop    bool

	PendingLevelUp bool
	QueuedLevelUps int
	UpgradeChoices [3]UpgradeCard
	ShopMenu       ShopMenuState
	SlotModal      SlotMachineState

	WorldEvents []WorldEvent

	CameraX, CameraY float32

	girlSpawnTimer      float32
	healSpawnTimer      float64
	boardedSinceHeal    int
	tutHealRestoreTimer float32
	nextEntityID        int
	rng                 *rand.Rand
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
	mapDef := DefaultDistrictMap()
	layout := mapDef.Layout
	rng := rand.New(rand.NewSource(seed))

	shops := make([]VendingShop, len(layout.Shops))
	copy(shops, layout.Shops)
	for i := range shops {
		shops[i].HealCost = EspressoHealCost
	}
	slots := make([]SlotMachine, len(layout.SlotMachines))
	copy(slots, layout.SlotMachines)
	ztlZones := make([]Rect, len(layout.ZTLZones))
	copy(ztlZones, layout.ZTLZones)
	chestSpots := make([][2]float32, len(mapDef.ChestSpots))
	copy(chestSpots, mapDef.ChestSpots)

	w := &World{
		Player:         NewPlayer(),
		Stops:          make([]BusStop, 0, 8),
		Girls:          make([]Girl, 0, 48),
		Police:         NewPoliceSystem(),
		Buses:          make([]NeonBus, 0, 4),
		ZTLZones:       ztlZones,
		Shops:          shops,
		SlotMachines:   slots,
		Props:          layout.Props,
		Lights:         layout.Lights,
		ParkIslands:    layout.ParkIslands,
		BusBays:        layout.BusBays,
		CrosswalkCols:  layout.CrosswalkCols,
		CrosswalkRows:  layout.CrosswalkRows,
		CanalY:         layout.CanalY,
		BoardwalkY:     layout.BoardwalkY,
		MapID:          mapDef.ID,
		MapName:        mapDef.NameIT,
		ChestSpots:     chestSpots,
		Popups:         make([]FloatingText, 0, 24),
		Telegraphs:     make([]TelegraphLine, 0, 12),
		Express:        make([]ExpressLine, 0, 8),
		WaveRings:      make([]WaveRing, 0, 16),
		Slipstreams:    make([]BusSlipstream, 0, 4),
		HealPickups:    make([]HealPickup, 0, 8),
		WorldEvents:    make([]WorldEvent, 0, 32),
		Seed:           seed,
		MusicBPM:       108.0,
		DifficultyTier: 1,
		BossRushMode:   bossRush,
		DistrictEvent:  DistrictEventState{NextEventTimer: 65.0},
		girlSpawnTimer: 1.5,
		healSpawnTimer: 24.0,
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

	// Seed initial horde across all 4 archetypes (5 girls, 0 VIPs initially)
	for i := 0; i < 5; i++ {
		arch := Archetype(i % int(ArchCount))
		w.spawnGirl(arch, false)
	}

	// Seed 2 initial world healing pickups on clear avenue/plaza spots
	w.spawnInitialHealPickups()

	// Seed Armored Neon Chests from MapDef.ChestSpots
	w.initChestsFromMap()

	w.snapCamera()
	return w
}

// NewTutorialWorld creates a safe interactive tutorial sandbox world where Alfredo can practice
// movement, throwing Bus Stops, Mouse-Only minigames/combos, boarding buses, healing, and shops/slots.
func NewTutorialWorld(save *meta.SaveData) *World {
	_ = save
	w := NewWorldWithOptions(4242, false)
	w.TutorialMode = true
	w.TutorialStep = 0
	w.BossSchedule = nil
	w.BossScheduleTimes = nil
	w.Chests = nil

	w.Player.X = 540
	w.Player.Y = 360
	w.constrainPlayerPos()
	w.Player.Cash = 180
	w.Player.MaxHP = 6
	w.Player.HP = 6
	w.Player.StopCooldownBase = 3.0
	w.Player.StopCooldownTimer = 0

	// Ensure 1 nearby Vending Shop and 1 nearby Slot Machine are right beside the central plaza
	shopX, shopY := w.constrainCircleToWorld(460, 316, 16, 60, ArenaWidth-60, 96, w.walkableMaxY()-20)
	slotX, slotY := w.constrainCircleToWorld(460, 404, 16, 60, ArenaWidth-60, 96, w.walkableMaxY()-20)
	w.Shops = append(w.Shops, VendingShop{
		X:         shopX,
		Y:         shopY,
		Radius:    60,
		HealCost:  EspressoHealCost,
		BribeCost: 100,
	})
	w.SlotMachines = append(w.SlotMachines, SlotMachine{
		X:        slotX,
		Y:        slotY,
		Radius:   60,
		SpinCost: 35,
	})
	w.Lights = append(w.Lights,
		LightSource{X: shopX, Y: shopY, Radius: 88, Kind: LightVending, R: 80, G: 235, B: 255, Intensity: 0.80},
		LightSource{X: slotX, Y: slotY, Radius: 92, Kind: LightSlotMachine, R: 255, G: 215, B: 60, Intensity: 0.85},
	)

	// Place 1 nearby HealPickup in clear view of the central plaza
	w.HealPickups = w.HealPickups[:0]
	hx, hy := w.findSafePickupSpot(612, 360, true)
	w.HealPickups = append(w.HealPickups, HealPickup{
		X:         hx,
		Y:         hy,
		Amount:    1,
		Chronos:   20.0,
		LifeTimer: 120.0,
		BobPhase:  0.4,
		Big:       false,
	})

	// Spawn 4 gentle practice girls (1 of each archetype, zero VIPs)
	w.Girls = w.Girls[:0]
	w.spawnTutorialPracticeGirls()

	w.snapCamera()
	return w
}

func (w *World) spawnTutorialPracticeGirls() {
	spots := [ArchCount][2]float32{
		ArchEmo:     {685, 312},
		ArchStudent: {735, 392},
		ArchKawaii:  {655, 418},
		ArchGyaru:   {725, 294},
	}
	for a := Archetype(0); a < ArchCount; a++ {
		g := NewGirl(w.allocID(), spots[a][0], spots[a][1], a, false, w.rng)
		g.ShootTimer = 3.5 + w.rng.Float32()*2.0
		w.constrainGirlPos(&g)
		w.Girls = append(w.Girls, g)
	}
}

// EarlyDifficultyScale returns an early-run intensity ramp starting at 0.55 at t=0
// and smoothly ramping to 1.0 over the first 120 seconds (or immediately 1.0 in Tier 2+ / Boss Rush).
func (w *World) EarlyDifficultyScale() float64 {
	if w.DifficultyTier >= 2 || w.BossRushMode {
		return 1.0
	}
	t := float64(w.RunTime) / 120.0
	if t <= 0 {
		return 0.55
	}
	if t >= 1.0 {
		return 1.0
	}
	return 0.55 + 0.45*t
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

// StopSafeTopY returns the top screen-space safe margin below the active HUD banners.
func (w *World) StopSafeTopY() float32 {
	if w.TutorialMode || w.Boss.Active {
		return 68
	}
	if w.BossRushMode || w.OverdriveActive || w.Police.Heli.Active || len(w.BossSchedule) > 0 || w.DistrictEvent.Active != DistrictEventNone {
		return 52
	}
	return 30
}

// ClampCompassScreenPos clamps a screen-space compass/HUD indicator inside safe viewport bounds.
func ClampCompassScreenPos(sx, sy, safeTopY float32) (float32, float32) {
	return clamp32(sx, 36, ScreenWidth-36), clamp32(sy, safeTopY+14, ScreenHeight-64)
}

// EnterOverdrive activates the endless post-victory Overdrive loop with cyclic boss waves.
func (w *World) EnterOverdrive() {
	w.OverdriveActive = true
	w.NextOverdriveBossTimer = 30.0
	w.addPopup(w.Player.X, w.Player.Y-38, "MODALITA OVERDRIVE!", 4)
}

// ForceSpawnBoss immediately spawns a Boss of the requested kind and clears arena clutter for readability.
func (w *World) ForceSpawnBoss(kind BossKind) {
	if w.Boss.Active {
		return
	}
	sx, sy := randomEdgePos(w.rng)
	w.Boss.Spawn(kind, sx, sy)
	if kind >= 0 && kind < BossCount {
		w.BossesSpawned[kind] = true
	}
	// Boss Arena Spotlight: clear airborne bullets and thin unhooked crowd down to <= 4
	w.Bullets.Count = 0
	w.BossIntroTimer = 2.6
	w.addWaveRing(w.Boss.X, w.Boss.Y, 180, WaveBossBreak)
	if w.ScreenShake < 8.0 {
		w.ScreenShake = 8.0
	}
	if len(w.Girls) > 4 {
		kept := make([]Girl, 0, len(w.Girls))
		freeKept := 0
		for i := range w.Girls {
			if w.Girls[i].HookedByStop != -1 {
				kept = append(kept, w.Girls[i])
			} else if freeKept < 4 {
				kept = append(kept, w.Girls[i])
				freeKept++
			}
		}
		w.Girls = kept
	}
	w.addPopup(w.Player.X, w.Player.Y-36, "BOSS: "+w.Boss.Name+"!", 4)
	w.ShowQuip(w.Boss.Name, BossIntroQuip(kind), 2, 3.8)
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
		if w.TutorialMode {
			w.syncTutorialStep()
		}
		return
	}

	// Rizz-Slots 777 modal pauses simulation cleanly
	if w.SlotModal.Active {
		w.UpdateSlotMachine(dt, in)
		if w.Police.Stars > w.PeakWanted {
			w.PeakWanted = w.Police.Stars
		}
		if w.TutorialMode {
			w.syncTutorialStep()
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
	if w.QuipTimer > 0 {
		w.QuipTimer -= dt
		if w.QuipTimer < 0 {
			w.QuipTimer = 0
		}
	}
	if w.synergyQuipCD > 0 {
		w.synergyQuipCD -= dt
		if w.synergyQuipCD < 0 {
			w.synergyQuipCD = 0
		}
	}
	if w.Police.Stars > w.PeakWanted {
		w.PeakWanted = w.Police.Stars
	}

	if w.TutorialMode && w.tutHealRestoreTimer > 0 {
		w.tutHealRestoreTimer -= dt
		if w.tutHealRestoreTimer <= 0 {
			w.tutHealRestoreTimer = 0
			w.Player.Heal(w.Player.MaxHP)
			w.addPopup(w.Player.X, w.Player.Y-28, "HP RIPRISTINATI (6/6)!", 3)
		}
	}

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
	if w.TutorialMode && !w.TutMoved {
		if math.Abs(float64(in.MoveX)) > 0.05 || math.Abs(float64(in.MoveY)) > 0.05 || dashed {
			w.TutMoved = true
		}
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

	if w.BossIntroTimer > 0 {
		w.BossIntroTimer -= dt
		if w.BossIntroTimer < 0 {
			w.BossIntroTimer = 0
		}
	}

	// Mouse-click interaction with nearby Vending Shop / Slot Machine (when not inside an active stop)
	if w.tryMouseClickInteract(in) {
		return
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
	if !recalled && w.Player.ThrowBufferTimer > 0 && w.Player.CanThrowStop(w.ActiveStopCount()) {
		w.Player.ThrowBufferTimer = 0
		w.tryThrowBusStop(aimX, aimY)
	}

	// Update Stops, Minigames & Boarding Combos
	w.updateStops(dt, in.ActionDir, in.ActionHeld)

	// Update Neon Buses
	w.updateBuses(dt)

	// Flush any Level-Up that was queued while finishing a minigame/combo
	if !w.PendingLevelUp && !w.ShopMenu.Active && !w.SlotModal.Active && w.QueuedLevelUps > 0 && !w.isPlayerInActiveStop() {
		w.QueuedLevelUps--
		w.openLevelUpModal()
	}

	// Boss Gatekeeper schedule (disabled in TutorialMode)
	if !w.TutorialMode {
		w.checkTimedBossSpawns(dt)
	}
	w.updateDistrictEvents(dt)

	// Update Girls & horde spawning
	w.updateGirlSpawner(dt)
	diffScale := float32(w.EarlyDifficultyScale())
	if w.TutorialMode {
		diffScale = 0.25
	}
	for i := range w.Girls {
		expressSlow := w.isPointInExpressLane(w.Girls[i].X, w.Girls[i].Y)
		if w.Boss.Active && w.Girls[i].ShootTimer > 0 {
			w.Girls[i].ShootTimer += dt * enemyTimeScale * 0.55
		}
		w.Girls[i].Update(dt, enemyTimeScale, w.DifficultyTier, diffScale, w.Player.X, w.Player.Y, w.Stops, expressSlow, &w.Bullets)
		w.constrainGirlPos(&w.Girls[i])
	}

	// Update Police & Boss (skipped in TutorialMode so urban events/bosses do not interrupt)
	if !w.TutorialMode {
		w.Police.BossFightMode = w.Boss.Active
		w.Police.Update(dt, enemyTimeScale, w.Player.X, w.Player.Y, w.Stops, &w.Bullets, &w.Telegraphs, w.rng)
		w.Boss.Update(dt, enemyTimeScale, w.Player.X, w.Player.Y, w.Stops, w.Girls, &w.Bullets, &w.Telegraphs)
		if w.Boss.PendingKnockback {
			w.Boss.PendingKnockback = false
			w.addWaveRing(w.Boss.X, w.Boss.Y, 165, 2)
			if w.Player.DashTimer <= 0 && distSq(w.Player.X, w.Player.Y, w.Boss.X, w.Boss.Y) <= 165.0*165.0 {
				dx, dy, l := normalize2D(w.Player.X-w.Boss.X, w.Player.Y-w.Boss.Y)
				if l <= 0.001 {
					dx, dy = 1, 0
				}
				w.Player.X += dx * 28.0
				w.Player.Y += dy * 28.0
				w.constrainPlayerPos()
				w.addPopup(w.Player.X, w.Player.Y-28, "LASCIALA STARE!", 4)
			}
		}
	}

	// Update BulletPool & resolve solid prop cover collisions
	w.Bullets.Update(dt, bulletTimeScale, w.Player.X, w.Player.Y, w.Stops)
	for _, sh := range w.Bullets.DrainShieldHits() {
		w.emitEvent(EventStopShieldHit, 0, sh.Broken)
		if sh.Broken && w.Player.UpgradeRanks[UpgradeGlassFortress] > 0 {
			empR := float32(90.0) + float32(w.Player.UpgradeRanks[UpgradeGlassFortress])*25.0
			w.Bullets.ClearRadius(sh.StopX, sh.StopY, empR)
			w.addWaveRing(sh.StopX, sh.StopY, empR, 1)
			w.addPopup(sh.StopX, sh.StopY-46, "CONTRO-EMP VETRATA!", 1)
		}
		if sh.TimerShaved {
			w.addPopup(sh.StopX, sh.StopY-34, "SWAT ATTACCA FERMATA!", 2)
		}
	}
	w.resolveBulletPropCollisions()

	// Check Tear-Graze & Bullet-Player collisions
	w.handleBulletInteractions()

	// Update World Healing Pickups ("Cure in giro")
	w.updateHealPickups(dt)

	// Update Armored Neon Chests (proximity / [E] unlock with Boss Keys)
	w.updateChests(dt, in.InteractJustPressed)

	// Vending Machine Shop & Slot Machine interaction [E]
	w.updateShops(dt, in.InteractJustPressed)

	if w.TutorialMode {
		w.syncTutorialStep()
	}
	if w.Police.Stars > w.PeakWanted {
		w.PeakWanted = w.Police.Stars
	}

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
	bassRank := w.Player.UpgradeRanks[UpgradeBasslineShock]
	pulseR := BasslinePulseRadius + float32(bassRank)*30.0
	w.addWaveRing(w.Player.X, w.Player.Y, pulseR, 0)
	rSq := pulseR * pulseR

	if bassRank > 0 {
		w.Bullets.ClearRadius(w.Player.X, w.Player.Y, 44.0+float32(bassRank)*14.0)
	}

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
	pullMax := pulseR * 1.45
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
		if z.IntersectsCircle(x, y, 10.0) {
			return true
		}
	}
	if w.Boss.Active && w.Boss.MartialZTLActive && w.Boss.MartialZTLRect.Contains(x, y) {
		return true
	}
	if w.DistrictEvent.Active == DistrictEventZTLJam && w.isPointInBusBay(x, y) {
		return true
	}
	return false
}

func (w *World) hasZTLRect(r Rect) bool {
	for _, z := range w.ZTLZones {
		if z == r {
			return true
		}
	}
	return false
}

func (w *World) isPointInBusBay(x, y float32) bool {
	for _, bay := range w.BusBays {
		if bay.IntersectsCircle(x, y, 18.0) {
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
	w.Player.SyncStopCharges()
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
		previewR *= 1.30
	}

	perfectDrop := false
	perfR := w.Player.EffectivePerfectRadius()
	perfSq := perfR * perfR
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
	if !w.Player.CanThrowStop(w.ActiveStopCount()) {
		return
	}

	fb := w.EvaluateAim(aimX, aimY)
	tx, ty := fb.ClampedX, fb.ClampedY
	inZTL := fb.InZTL
	inBusBay := fb.InBusBay
	perfectDrop := fb.IsPerfectDrop
	w.Player.ThrowFacingX = w.Player.FacingX
	if tx < w.Player.X-2 {
		w.Player.ThrowFacingX = -1
	} else if tx > w.Player.X+2 {
		w.Player.ThrowFacingX = 1
	}

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
					w.Player.ConsumeStopCharge(w.Player.EffectiveStopCooldown() * 0.65)
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
	if w.DistrictEvent.Active == DistrictEventZTLJam && inBusBay {
		stop.GoldenZTL = true
		stop.InZTL = true
		inZTL = true
	}
	ax, ay := ComputeStopUIAnchor(tx-w.CameraX, ty-w.CameraY, w.StopSafeTopY())
	stop.UIAnchorX = ax
	stop.UIAnchorY = ay
	stop.UIAnchored = true
	shieldBonus := w.Player.ShieldBonus()
	stop.ShieldHP += shieldBonus
	stop.MaxShieldHP += shieldBonus
	if durBonus := w.Player.StopDurationBonus(); durBonus > 0 {
		stop.LifeTimer += durBonus
		stop.MaxLifeTimer += durBonus
	}

	cd := w.Player.EffectiveStopCooldown()
	if perfectDrop {
		cdFactor := float32(0.65)
		if w.StreakCount >= 5 {
			cdFactor = 0.50
		}
		cd *= cdFactor
	}
	if inBusBay {
		cd *= 0.75
	}
	w.Player.ConsumeStopCharge(cd)
	w.Player.ThrowAnimTimer = 0.22

	if perfectDrop {
		w.PerfectDrops++
		w.HitStopTimer = 0.090 // 90ms Perfect Drop Hit-Stop
		w.ScreenShake = 6.5
		novaRank := w.Player.UpgradeRanks[UpgradePerfectNova]
		clearR := float32(130.0) + float32(novaRank)*45.0
		if w.Player.HasRelic(RelicArmoredGlass) {
			clearR += 35.0
		}
		if novaRank > 0 {
			w.Player.ChronosMeter = clamp32(w.Player.ChronosMeter+float32(novaRank)*15.0, 0, w.Player.ChronosMax)
			w.Player.Cash += novaRank * 12
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
		w.Bullets.ClearRadius(tx, ty, 110.0)
		ringKind := 2
		if stop.GoldenZTL {
			ringKind = 1
		}
		w.addWaveRing(tx, ty, 110.0, ringKind)
		if !w.TutorialMode {
			added := w.Police.AddStar(1)
			if added > 0 {
				if stop.GoldenZTL {
					w.addPopup(tx, ty-44, "INGORGO ZTL 4x! +1* WANTED!", 1)
				} else {
					w.addPopup(tx, ty-44, "ZTL! +1* WANTED!", 2)
				}
				w.emitEvent(EventPoliceStarUp, w.Police.Stars, false)
			} else {
				if stop.GoldenZTL {
					w.addPopup(tx, ty-44, "BONUS INGORGO ZTL 4x!", 1)
				} else {
					w.addPopup(tx, ty-44, "BONUS ZTL 3x!", 2)
				}
			}
		} else {
			w.addPopup(tx, ty-44, "BONUS ZTL 3x!", 2)
		}
	} else if inBusBay {
		w.Bullets.ClearRadius(tx, ty, 105.0)
		w.addWaveRing(tx, ty, 105.0, 0)
		w.addPopup(tx, ty-44, "AREA BUS (+40% / SCUDO+2)!", 1)
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

	if hookedVIP && !w.TutorialMode {
		if w.Police.AddStar(1) > 0 {
			w.addPopup(tx, ty-56, "BERSAGLIO VIP! +1*!", 2)
			w.emitEvent(EventPoliceStarUp, w.Police.Stars, false)
		}
	}

	if stop.SynergyActive {
		synPopup := "SYNERGY!"
		switch stop.Arch {
		case ArchEmo:
			synPopup = "SOULMATE! +1 SCUDO & CUORE"
		case ArchStudent:
			synPopup = "HONOR ROLL! +XP & FREEZE"
		case ArchKawaii:
			synPopup = "MOE CRIT! GRAZE x2"
		case ArchGyaru:
			synPopup = "VIRAL #1! BARRA +35%"
		}
		w.addPopup(tx, ty-16, synPopup, 1)
		w.emitEvent(EventSynergy, int(stop.Arch), false)
		if w.synergyQuipCD <= 0 && !w.Boss.Active {
			w.ShowQuip(GetArchetypeSpec(stop.Arch).Name, SynergyQuip(stop.Arch), 0, 2.8)
			w.synergyQuipCD = 9.0
		}
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

	if !w.TutorialMode && !w.BossRushMode && w.RunTime <= 60.0 && !w.FirstBoardingDone && len(stop.HookedGirlIDs) == 0 && !stop.HookedBoss {
		w.Player.StopCooldownTimer *= 0.60
		w.addPopup(tx, ty-38, "LANCIO A VUOTO (-40% CD)", 1)
		if !w.EarlyMissTipShown {
			w.EarlyMissTipShown = true
			w.ShowQuip("CONSIGLIO", "Sfiora i colpi o fai Dash-Parry per ricaricare prima!", 1, 3.6)
		}
	}

	w.Stops = append(w.Stops, stop)
	if w.TutorialMode {
		w.TutThrewStop = true
	}
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

		st.DeployAge = min(st.DeployAge+dt, 1)
		if st.ShieldFlashTimer > 0 {
			st.ShieldFlashTimer -= dt
			if st.ShieldFlashTimer < 0 {
				st.ShieldFlashTimer = 0
			}
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

		// Area Bus tactical perk: gentle curbside pull draws nearby free girls toward the shelter
		if st.InBusBay {
			const pullR2 = float32(220.0 * 220.0)
			for g := range w.Girls {
				gr := &w.Girls[g]
				if gr.HookedByStop == -1 {
					dx, dy, d := normalize2D(st.X-gr.X, st.Y-gr.Y)
					if d > 18 && d*d <= pullR2 {
						gr.X += dx * 42.0 * dt
						gr.Y += dy * 42.0 * dt
						w.constrainGirlPos(gr)
					}
				}
			}
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
				if primaryStopIdx >= 0 && primaryStopIdx != i {
					drain = 0.25
				} else {
					drain = 1.25
				}
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
				mgDT *= 1.40
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
						st.Minigame.Progress += (st.Minigame.Progress - prevProg) * 0.40
					}
				case ArchKawaii:
					contGain := float32(0.24) * mgDT
					if st.PlayerInside && st.Minigame.Progress > prevProg+contGain+0.01 {
						discrete := st.Minigame.Progress - prevProg - contGain
						st.Minigame.Progress += discrete * 0.40
					}
				case ArchStudent:
					if st.Minigame.GridX != prevGX || st.Minigame.GridY != prevGY {
						boosted := float32(st.Minigame.GridX+st.Minigame.GridY) / 6.0 * 0.92 * 1.40
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
				w.addPopup(st.X, st.Y-38, "PROIETTILI RALLENTATI!", 1)
			}
		}

		if st.Minigame.Completed && !st.Combo.Active && !st.Combo.Completed && !st.Combo.Failed {
			if w.TutorialMode {
				w.TutCompletedMinigame = true
			}
			strict := st.HookedBoss && w.Boss.Active && w.Boss.Kind == BossProfessor
			st.Combo.Start(st.Arch == ArchEmo && st.SynergyActive, strict, w.rng)
			if comboBonus := w.Player.ComboDurationBonus(); comboBonus > 0 {
				st.Combo.Timer += comboBonus
				st.Combo.MaxTimer += comboBonus
			}
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

		// Neon Bus Purification Wave clears bullets along its streak (Rush Hour / Biglietto d'Oro convert to Cash!)
		cleared := w.Bullets.ClearLine(prevX, bus.Y, bus.X+42, bus.Y, 46.0)
		if cleared > 0 {
			cashPerBullet := w.Player.UpgradeRanks[UpgradeRizzMagnet] * 2
			if w.StreakCount >= 5 {
				cashPerBullet += 3
			}
			if cashPerBullet > 0 {
				w.Player.Cash += cleared * cashPerBullet
			}
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
				gCash := float32(18.0)
				gScore := float32(100.0)
				if g.IsVIP {
					gCash *= 3.0
					gScore *= 3.0
				}
				if st.GoldenZTL {
					gXP *= 4.0
					gCash *= 4.0
					gScore *= 4.0
				} else if st.InZTL {
					gXP *= 1.5
					gCash *= 3.0
					gScore *= 3.0
				}
				totalXP += int(gXP * xpMult)
				totalCash += int(gCash * cashMult)
				totalScore += int(gScore)
				w.GirlsBoarded++
				if g.Arch >= 0 && g.Arch < ArchCount {
					w.BoardedByArch[g.Arch]++
				}
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
		if w.TutorialMode {
			w.TutBoardedBus = true
		}
		for k := 0; k < boardedHere; k++ {
			w.boardedSinceHeal++
			if w.boardedSinceHeal == 3 && w.rng.Float32() < 0.35 {
				w.boardedSinceHeal = 0
				w.spawnHealPickupNear(st.X, st.Y, 1, 20.0, false)
			} else if w.boardedSinceHeal >= 4 {
				if w.Player.HP < w.Player.MaxHP {
					w.spawnHealPickupNear(st.X, st.Y, 1, 20.0, false)
				}
				w.boardedSinceHeal = 0
			}
		}
	}

	if st.Arch == ArchEmo && st.SynergyActive {
		w.spawnBlackHeartNear(st.X, st.Y)
		w.addPopup(st.X, st.Y-36, "BLACK HEART! (+1 HP)", 1)
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
			bossX, bossY := w.Boss.X, w.Boss.Y
			w.HitStopTimer = 0.140 // 140ms Boss Break Hit-Stop
			w.ScreenShake = 11.5
			w.Bullets.ClearRadius(bossX, bossY, 240)
			w.addWaveRing(bossX, bossY, 116, WaveBossBreak)
			bossKind := w.Boss.Kind
			bossName := w.Boss.Name
			defeated := w.Boss.BreakShield()
			w.ShowQuip(bossName, BossBreakQuip(bossKind, w.Boss.Shields), 2, 3.2)

			// Boss shield break ALWAYS drops a Big (+2 HP, +35 Chronos) healing pickup near the boss!
			w.spawnHealPickupNear(bossX, bossY, 2, 35.0, true)

			if defeated {
				starsAtKill := w.Police.Stars
				totalXP += int(240 * xpMult)
				totalCash += int(200 * cashMult)
				totalScore += 2500
				w.Police.ReduceStar(2)
				w.DifficultyTier++
				w.BossesDefeated++
				w.Player.BossKeys++
				w.addPopup(st.X, st.Y-48, "BOSS A BORDO! TIER UP!", 4)
				w.addPopup(bossX, bossY-24, "+1 CHIAVE BOSS! CERCA UN BAULE!", 1)
				w.emitEvent(EventBossBreak, int(bossKind), true)
				w.emitEvent(EventBossDefeated, int(bossKind), starsAtKill >= 3)
				w.emitEvent(EventBossKeyGained, w.Player.BossKeys, false)
				if len(w.Chests) > 0 && w.ClosedChestCount() == 0 {
					w.restockChests()
				}

				if bossKind == BossLover {
					extraZTLs := [2]Rect{
						{X: 296, Y: 186, W: 112, H: 76},
						{X: 872, Y: 448, W: 112, H: 76},
					}
					addedZTL := false
					for _, zr := range extraZTLs {
						if !w.hasZTLRect(zr) {
							w.ZTLZones = append(w.ZTLZones, zr)
							addedZTL = true
						}
					}
					if addedZTL {
						w.addPopup(st.X, st.Y-64, "ZTL RADDOPPIATE!", 2)
					}
				}

				if !w.OverdriveActive {
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
				}
			} else {
				totalXP += int(80 * xpMult)
				totalCash += int(80 * cashMult)
				totalScore += 600
				w.addPopup(st.X, st.Y-48, fmt.Sprintf("SCUDO ROTTO! (%d/%d)", w.Boss.MaxShields-w.Boss.Shields, w.Boss.MaxShields), 4)
				w.emitEvent(EventBossBreak, int(bossKind), false)
			}
		}
	}

	if boardedHere > 0 || bossBroken {
		w.StreakCount++
		w.StreakTimer = w.Player.EffectiveStreakWindow()
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

	if !w.FirstBoardingDone && boardedHere > 0 {
		w.FirstBoardingDone = true
		if !w.TutorialMode && !w.BossRushMode && w.RunTime <= 60.0 {
			totalXP += 15
			w.Player.StopCooldownTimer = 0
			if w.Player.StopCharges == 0 {
				w.Player.StopCharges = 1
			}
			w.addPopup(st.X, st.Y-54, "PRIMO IMBARCO! +15XP & FERMATA PRONTA!", 1)
		}
	}

	if w.DistrictEvent.Active == DistrictEventBlackout {
		w.DistrictEvent.BlackoutScoreEarned += totalScore
		w.DistrictEvent.BlackoutXPEarned += totalXP
	}

	if totalCash > 0 || totalXP > 0 {
		w.Player.Cash += totalCash
		w.Score += totalScore
		w.addPopup(st.X, st.Y-24, fmt.Sprintf("+$%d  +%dXP", totalCash, totalXP), 3)
		if w.Player.AddXP(totalXP) {
			w.triggerLevelUp()
		}
	}

	// Express Chain Bonus: completing a linked stop refreshes sibling LifeTimer & boosts minigame +35%
	for idx := range w.Stops {
		other := &w.Stops[idx]
		if other.ID == st.ID || !other.Active || other.BusRushTriggered {
			continue
		}
		if distSq(st.X, st.Y, other.X, other.Y) <= ExpressLinkMaxDist*ExpressLinkMaxDist {
			other.LifeTimer = other.MaxLifeTimer
			if !other.Minigame.Completed {
				other.Minigame.Progress += 0.35
				if other.Minigame.Progress >= 1.0 {
					other.Minigame.Progress = 1.0
					other.Minigame.Completed = true
				}
			}
			w.addPopup(other.X, other.Y-32, "EXPRESS CHAIN +35%!", 1)
		}
	}

	st.Active = false
}

func (w *World) checkTimedBossSpawns(dt float32) {
	if w.Boss.Active {
		return
	}
	if w.OverdriveActive {
		w.NextOverdriveBossTimer -= dt
		if w.NextOverdriveBossTimer <= 0 {
			w.OverdriveWave++
			kind := BossKind((w.OverdriveWave - 1) % int(BossCount))
			w.ForceSpawnBoss(kind)
			w.NextOverdriveBossTimer = 55.0
		}
		return
	}
	if w.BossScheduleIdx >= len(w.BossSchedule) {
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
	if w.TutorialMode {
		if len(w.Girls) == 0 {
			w.spawnTutorialPracticeGirls()
			w.addPopup(w.Player.X, w.Player.Y-32, "NUOVE RAGAZZE PRATICA!", 1)
		}
		return
	}

	targetCap := 8 + int(w.RunTime/24.0) + (w.DifficultyTier-1)*3
	if targetCap > 26 {
		targetCap = 26
	}
	interval := float32(2.4) - float32(w.DifficultyTier-1)*0.35 - w.RunTime*0.0025
	if interval < 0.85 {
		interval = 0.85
	}
	if w.Boss.Active {
		targetCap = 5
		interval = 4.5
	}
	w.girlSpawnTimer -= dt
	if w.girlSpawnTimer <= 0 && len(w.Girls) < targetCap {
		w.girlSpawnTimer = interval

		var arch Archetype
		if w.DistrictEvent.Active == DistrictEventFlashmob && w.rng.Float32() < 0.80 {
			arch = w.DistrictEvent.FlashmobArch
		} else {
			arch = Archetype(w.rng.Intn(int(ArchCount)))
		}
		vipAllowed := !w.Boss.Active && (w.RunTime >= 45.0 || w.DifficultyTier >= 2)
		isVIP := false
		if vipAllowed {
			vipChance := float32(0.10)
			if w.DifficultyTier >= 2 {
				vipChance = 0.18
			}
			isVIP = w.rng.Float32() < vipChance
		}
		w.spawnGirl(arch, isVIP)
	}
}

func (w *World) spawnGirl(arch Archetype, isVIP bool) {
	sx, sy := randomEdgePos(w.rng)
	g := NewGirl(w.allocID(), sx, sy, arch, isVIP, w.rng)
	if w.DifficultyTier >= 2 {
		g.ShootTimer = 0.9 + w.rng.Float32()*1.3
	}
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
			parryR := hitR + 6.0 + w.Player.DashParryBonusRadius()
			if d2 <= parryR*parryR {
				w.Bullets.RemoveAt(i)
				w.Score += 40
				w.Player.ChronosMeter = clamp32(w.Player.ChronosMeter+10.0+w.Player.DashParryBonusChronos(), 0, w.Player.ChronosMax)
				bonusShave := w.Player.DashParryBonusShave()
				shaveCap := float32(1.5) + bonusShave
				if !w.Player.DashParryHitStopDone {
					w.Player.DashParryHitStopDone = true
					if w.HitStopTimer < 0.025 {
						w.HitStopTimer = 0.025
					}
					w.addWaveRing(w.Player.X, w.Player.Y, 36.0+w.Player.DashParryBonusRadius(), 1)
					w.addPopup(w.Player.X, w.Player.Y-26, fmt.Sprintf("PARRY! -%.1fs CD", 1.0+bonusShave), 1)
					w.emitEvent(EventDashParry, 0, false)
				}
				if w.Player.DashParryShaveUsed < shaveCap {
					shave := float32(1.0) + bonusShave
					if w.Player.DashParryShaveUsed > 0 {
						shave = 0.5
					}
					if w.Player.DashParryShaveUsed+shave > shaveCap {
						shave = shaveCap - w.Player.DashParryShaveUsed
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
					w.StreakTimer = clamp32(w.StreakTimer+0.35, 0, w.Player.EffectiveStreakWindow())
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

	// Police 3★+ Laser/Spike Road Hazard collisions
	if !w.Player.IsInvulnerable() {
		for i := range w.Police.Hazards {
			h := &w.Police.Hazards[i]
			if h.Active && h.TelegraphTimer <= 0 {
				if pointToSegmentDistSq(w.Player.X, w.Player.Y, h.X1, h.Y1, h.X2, h.Y2) <= 10.0*10.0 {
					w.damagePlayer(1, BulletPolice)
					break
				}
			}
		}
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
		if w.TutorialMode {
			w.Player.HP = 1
			w.tutHealRestoreTimer = 1.0
			w.addPopup(w.Player.X, w.Player.Y-34, "TUTORIAL: RECUPERO HP!", 3)
		} else {
			w.GameOver = true
		}
	}
}

func (w *World) isPlayerInActiveStop() bool {
	for i := range w.Stops {
		st := &w.Stops[i]
		if !st.Active || st.BusRushTriggered {
			continue
		}
		if st.PlayerInside || st.CoyoteTimer > 0 || distSq(w.Player.X, w.Player.Y, st.X, st.Y) <= st.CaptureRadius*st.CaptureRadius {
			return true
		}
	}
	return false
}

// tryMouseClickInteract allows opening a nearby Vending Shop or Slot Machine with a direct LMB click
// when Alfredo is not inside an active Bus Stop minigame/combo.
func (w *World) tryMouseClickInteract(in InputState) bool {
	if !in.MouseLeftJustPressed || w.isPlayerInActiveStop() {
		return false
	}
	clickX, clickY := in.AimWorldX, in.AimWorldY
	if clickX == 0 && clickY == 0 && (in.MouseScreenX != 0 || in.MouseScreenY != 0) {
		clickX = w.CameraX + in.MouseScreenX
		clickY = w.CameraY + in.MouseScreenY
	}

	// Priority: if the player is aiming at an unhooked girl or Boss (or can hook targets here),
	// throwing the Bus Stop takes precedence over opening a Shop or Slot Machine!
	for i := range w.Girls {
		g := &w.Girls[i]
		if g.HookedByStop == -1 && distSq(clickX, clickY, g.X, g.Y) <= 30.0*30.0 {
			return false
		}
	}
	if w.Boss.Active && w.Boss.HookedByStop == -1 && distSq(clickX, clickY, w.Boss.X, w.Boss.Y) <= 38.0*38.0 {
		return false
	}
	if w.Player.CanThrowStop(w.ActiveStopCount()) {
		if fb := w.EvaluateAim(clickX, clickY); fb.HookCandidateCount > 0 {
			return false
		}
	}

	const (
		nearPlayerSq = float32(85.0 * 85.0)
		farPlayerSq  = float32(110.0 * 110.0)
		nearClickSq  = float32(36.0 * 36.0)
		tightClickSq = float32(28.0 * 28.0)
	)

	for i := range w.Shops {
		sh := &w.Shops[i]
		if sh.Cooldown > 0 {
			continue
		}
		pD2 := distSq(w.Player.X, w.Player.Y, sh.X, sh.Y)
		cD2 := distSq(clickX, clickY, sh.X, sh.Y)
		if (pD2 <= nearPlayerSq && cD2 <= nearClickSq) || (pD2 <= farPlayerSq && cD2 <= tightClickSq) {
			w.OpenShop(i)
			return true
		}
	}

	for i := range w.SlotMachines {
		sm := &w.SlotMachines[i]
		pD2 := distSq(w.Player.X, w.Player.Y, sm.X, sm.Y)
		cD2 := distSq(clickX, clickY, sm.X, sm.Y)
		if (pD2 <= nearPlayerSq && cD2 <= nearClickSq) || (pD2 <= farPlayerSq && cD2 <= tightClickSq) {
			w.OpenSlotMachine(i)
			return true
		}
	}

	if w.Player.BossKeys > 0 {
		for i := range w.Chests {
			ch := &w.Chests[i]
			if ch.Opened {
				continue
			}
			pD2 := distSq(w.Player.X, w.Player.Y, ch.X, ch.Y)
			cD2 := distSq(clickX, clickY, ch.X, ch.Y)
			if (pD2 <= nearPlayerSq && cD2 <= nearClickSq) || (pD2 <= farPlayerSq && cD2 <= tightClickSq) {
				if w.OpenChest(i) {
					return true
				}
			}
		}
	}
	return false
}

func (w *World) spawnInitialHealPickups() {
	candidates := [2][2]float32{
		{460, 236},
		{820, 476},
	}
	for _, c := range candidates {
		px, py := w.findSafePickupSpot(c[0], c[1], true)
		w.HealPickups = append(w.HealPickups, HealPickup{
			X:         px,
			Y:         py,
			Amount:    1,
			Chronos:   20.0,
			LifeTimer: 45.0,
			BobPhase:  w.rng.Float64() * 6.283,
			Big:       false,
		})
	}
}

func (w *World) isClearPickupPos(x, y float32) bool {
	if x < 48 || x > ArenaWidth-48 || y < 108 || y > w.walkableMaxY()-18 {
		return false
	}
	for i := range w.Props {
		if w.Props[i].Solid && w.Props[i].CollideRect.IntersectsCircle(x, y, 14.0) {
			return false
		}
	}
	return true
}

func (w *World) findSafePickupSpot(nearX, nearY float32, nearTarget bool) (float64, float64) {
	maxY := w.walkableMaxY() - 20
	if nearTarget {
		// Try exact or tiny radial jitter around target position
		for attempt := 0; attempt < 12; attempt++ {
			ang := w.rng.Float64() * 6.283
			rad := float32(0)
			if attempt > 0 {
				rad = 14.0 + float32(attempt)*4.0
			}
			cx := clamp32(nearX+float32(math.Cos(ang))*rad, 48, ArenaWidth-48)
			cy := clamp32(nearY+float32(math.Sin(ang))*rad, 108, maxY)
			if w.isClearPickupPos(cx, cy) {
				return float64(cx), float64(cy)
			}
		}
		cx, cy := w.constrainCircleToWorld(nearX, nearY, 14.0, 48, ArenaWidth-48, 108, maxY)
		return float64(cx), float64(cy)
	}

	// Random walkable avenue/plaza spot clear of solid props
	for attempt := 0; attempt < 24; attempt++ {
		cx := 80.0 + w.rng.Float32()*(ArenaWidth-160.0)
		cy := 148.0 + w.rng.Float32()*(maxY-168.0)
		if !w.isClearPickupPos(cx, cy) {
			continue
		}
		tooClose := false
		for i := range w.HealPickups {
			if distSq(cx, cy, float32(w.HealPickups[i].X), float32(w.HealPickups[i].Y)) < 96*96 {
				tooClose = true
				break
			}
		}
		if !tooClose {
			return float64(cx), float64(cy)
		}
	}
	cx, cy := w.constrainCircleToWorld(ArenaWidth*0.5, ArenaHeight*0.5, 14.0, 48, ArenaWidth-48, 108, maxY)
	return float64(cx), float64(cy)
}

func (w *World) spawnHealPickupNear(x, y float32, amount int, chronos float64, big bool) {
	px, py := w.findSafePickupSpot(x, y, true)
	w.HealPickups = append(w.HealPickups, HealPickup{
		X:         px,
		Y:         py,
		Amount:    amount,
		Chronos:   chronos,
		LifeTimer: 45.0,
		BobPhase:  w.rng.Float64() * 6.283,
		Big:       big,
	})
}

func (w *World) spawnBlackHeartNear(x, y float32) {
	px, py := w.findSafePickupSpot(x, y, true)
	w.HealPickups = append(w.HealPickups, HealPickup{
		X:          px,
		Y:          py,
		Amount:     1,
		Chronos:    25.0,
		LifeTimer:  45.0,
		BobPhase:   w.rng.Float64() * 6.283,
		BlackHeart: true,
	})
}

func (w *World) updateHealPickups(dt float32) {
	dt64 := float64(dt)

	// Periodic world heal spawning: every 30s if fewer than 2 pickups exist
	w.healSpawnTimer -= dt64
	if w.healSpawnTimer <= 0 {
		w.healSpawnTimer = 30.0
		if len(w.HealPickups) < 2 {
			px, py := w.findSafePickupSpot(0, 0, false)
			w.HealPickups = append(w.HealPickups, HealPickup{
				X:         px,
				Y:         py,
				Amount:    1,
				Chronos:   20.0,
				LifeTimer: 45.0,
				BobPhase:  w.rng.Float64() * 6.283,
				Big:       false,
			})
		}
	}

	px := float64(w.Player.X)
	py := float64(w.Player.Y)
	canCollect := w.Player.HP < w.Player.MaxHP || w.Player.ChronosMeter < 95.0

	for i := 0; i < len(w.HealPickups); {
		p := &w.HealPickups[i]
		p.BobPhase += dt64 * 4.2
		p.LifeTimer -= dt64
		if p.LifeTimer <= 0 {
			w.HealPickups[i] = w.HealPickups[len(w.HealPickups)-1]
			w.HealPickups = w.HealPickups[:len(w.HealPickups)-1]
			continue
		}

		if canCollect {
			dx := px - p.X
			dy := py - p.Y
			d := math.Hypot(dx, dy)
			if d <= 56.0 && d > 0.001 {
				step := 160.0 * dt64
				if step > d {
					step = d
				}
				p.X += (dx / d) * step
				p.Y += (dy / d) * step
				dx = px - p.X
				dy = py - p.Y
				d = math.Hypot(dx, dy)
			}
			if d <= 26.0 {
				w.Player.Heal(p.Amount)
				w.Player.HealFlashTimer = 0.45
				w.Player.ChronosMeter = clamp32(w.Player.ChronosMeter+float32(p.Chronos), 0, w.Player.ChronosMax)
				w.addPopup(w.Player.X, w.Player.Y-28, fmt.Sprintf("+%d HP", p.Amount), 3)
				w.emitEvent(EventHealPickup, p.Amount, p.Big)

				w.HealPickups[i] = w.HealPickups[len(w.HealPickups)-1]
				w.HealPickups = w.HealPickups[:len(w.HealPickups)-1]
				continue
			}
		}

		i++
	}
}

func (w *World) syncTutorialStep() {
	if w.TutorialStep < 1 && w.TutMoved && w.TutThrewStop {
		w.TutorialStep = 1
	}
	if w.TutorialStep < 2 && w.TutCompletedMinigame {
		w.TutorialStep = 2
	}
	if w.TutorialStep < 3 && w.TutBoardedBus {
		w.TutorialStep = 3
	}
}

func (w *World) updateShops(dt float32, interactPressed bool) {
	for i := range w.Shops {
		sh := &w.Shops[i]
		if sh.Cooldown > 0 {
			sh.Cooldown = max(0, sh.Cooldown-dt)
		}
		if sh.HealCooldown > 0 {
			sh.HealCooldown = max(0, sh.HealCooldown-dt)
		}
		if sh.UpgradeCooldown > 0 {
			sh.UpgradeCooldown = max(0, sh.UpgradeCooldown-dt)
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
	{Kind: UpgradeCooldown, Title: "ORARI TURBO", Desc: "Ricarica fermate -22% più veloce (fino a 1.5s)", MaxRank: 5},
	{Kind: UpgradeMultiHook, Title: "GANCIO DOPPIO PIANO", Desc: "Aggancia +2 ragazze insieme per fermata (max 12)", MaxRank: 5},
	{Kind: UpgradeMaxStops, Title: "RETE EXPRESS", Desc: "Piazza +1 fermata in più contemporaneamente (max 4)", MaxRank: 3},
	{Kind: UpgradeRange, Title: "PENSILINA XL", Desc: "Lanci più lontano e cerchio cattura più largo (+15%)", MaxRank: 5},
	{Kind: UpgradeChronos, Title: "KATANA RIZZ-FOCUS", Desc: "Chronos (rallentatore) ancora più forte (-10%) e dura +25% barra", MaxRank: 4},
	{Kind: UpgradeVitality, Title: "ESPRESSO OVERDRIVE", Desc: "+1 Vita max, cura subito +2 cuori e corri +12% più veloce", MaxRank: 0},
	{Kind: UpgradeDashKatana, Title: "SCATTO KATANA", Desc: "Scatto (Dash) più frequente (-20%); attraversare i colpi ricarica prima le fermate", MaxRank: 3},
	{Kind: UpgradeGlassFortress, Title: "VETRATA BLINDATA", Desc: "Vetro fermata blocca +2 colpi; se si rompe pulisce i colpi vicini", MaxRank: 3},
	{Kind: UpgradeMinigameTurbo, Title: "DITA VELOCI", Desc: "Minigiochi +22% più rapidi e +2.5s extra per chiudere fermata e combo", MaxRank: 3},
	{Kind: UpgradeBasslineShock, Title: "SUBWOOFER URBANO", Desc: "L'onda attorno ad Alfredo scatta prima, è più larga e cancella i colpi", MaxRank: 3},
	{Kind: UpgradePerfectNova, Title: "IMPATTO PERFECT", Desc: "Lancio Perfetto (mirino oro) più facile, pulisce più colpi e dà +15 Chronos", MaxRank: 3},
	{Kind: UpgradeRizzMagnet, Title: "BIGLIETTO D'ORO", Desc: "+20% XP e soldi, serie +2.5s più lunga e il bus fa +$2 per ogni colpo pulito", MaxRank: 3},
}

func (w *World) isUpgradeEligible(card UpgradeCard) bool {
	if card.MaxRank > 0 && w.Player.UpgradeRanks[card.Kind] >= card.MaxRank {
		return false
	}
	switch card.Kind {
	case UpgradeCooldown:
		return w.Player.StopCooldownBase > 1.51
	case UpgradeMultiHook:
		return w.Player.MaxMultiHook < 12
	case UpgradeMaxStops:
		return w.Player.MaxSimultaneousStops < 4
	case UpgradeRange:
		return w.Player.RangeMult < 1.59
	case UpgradeChronos:
		return w.Player.ChronosTimeScale > 0.21
	}
	return true
}

func (w *World) triggerLevelUp() {
	w.emitEvent(EventLevelUp, w.Player.Level, false)
	if w.isPlayerInActiveStop() {
		w.QueuedLevelUps++
		w.addPopup(w.Player.X, w.Player.Y-34, "LEVEL UP IN CODA!", 1)
		return
	}
	w.openLevelUpModal()
}

func (w *World) openLevelUpModal() {
	w.PendingLevelUp = true
	eligible := make([]UpgradeCard, 0, len(allUpgradeCards))
	for _, card := range allUpgradeCards {
		if w.isUpgradeEligible(card) {
			eligible = append(eligible, card)
		}
	}
	if len(eligible) < 3 {
		fallbacks := []UpgradeKind{UpgradeVitality, UpgradeRizzMagnet, UpgradePerfectNova}
		for _, fk := range fallbacks {
			already := false
			for _, ex := range eligible {
				if ex.Kind == fk {
					already = true
					break
				}
			}
			if !already {
				for _, card := range allUpgradeCards {
					if card.Kind == fk {
						eligible = append(eligible, card)
						break
					}
				}
			}
			if len(eligible) >= 3 {
				break
			}
		}
	}
	for i := len(eligible) - 1; i > 0; i-- {
		j := w.rng.Intn(i + 1)
		eligible[i], eligible[j] = eligible[j], eligible[i]
	}
	w.UpgradeChoices[0] = eligible[0]
	w.UpgradeChoices[1] = eligible[1]
	w.UpgradeChoices[2] = eligible[2]
}

func (w *World) applyUpgrade(card UpgradeCard) {
	if card.Kind >= 0 && card.Kind < UpgradeCount {
		w.Player.UpgradeRanks[card.Kind]++
	}
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
			w.Player.StopCharges = clampInt(w.Player.StopCharges+1, 1, w.Player.EffectiveMaxStops())
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
	case UpgradeGlassFortress:
		for i := range w.Stops {
			if w.Stops[i].Active {
				w.Stops[i].ShieldHP += 2
				w.Stops[i].MaxShieldHP += 2
			}
		}
	case UpgradeBasslineShock:
		if period := w.Player.EffectiveBasslinePeriod(); w.Player.BasslinePulseTimer > period {
			w.Player.BasslinePulseTimer = period
		}
	case UpgradeRizzMagnet:
		if w.StreakCount > 0 {
			w.StreakTimer = w.Player.EffectiveStreakWindow()
		}
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
