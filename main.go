package main

import (
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"girlbusAG/assets"
	"girlbusAG/internal/audio"
	"girlbusAG/internal/game"
	"girlbusAG/internal/input"
	"girlbusAG/internal/meta"
	"girlbusAG/internal/render"
	"girlbusAG/internal/ui"
)

// App implements ebiten.Game for GIRLBUS: NEON STOP (girlbusAG v2.1).
type App struct {
	world          *game.World
	bgWorld        *game.World
	renderer       *render.Renderer
	post           *render.PostProcessor
	reader         *input.Reader
	ui             *ui.Manager
	audio          *audio.Engine
	save           meta.SaveData
	tracker        meta.RunTracker
	overdriveAck   bool
	lastPerfectAim bool
	tick           float32
}

// NewApp initializes persisted settings/profile, SpriteBank, PostProcessor, Audio, UI, and World.
func NewApp() *App {
	save := meta.Load()
	if save.Settings.Fullscreen {
		ebiten.SetFullscreen(true)
	}

	sprites := assets.LoadSpriteBank()
	renderer := render.NewRenderer(sprites)
	renderer.SetVisualPrefs(
		save.Profile.SelectedPlayerSkin,
		save.Profile.SelectedBusSkin,
		save.Settings.ScreenShake,
		save.Settings.ShowFPS,
	)

	audioEng := audio.NewEngine()
	audioEng.SetVolumes(save.Settings.MasterVol, save.Settings.MusicVol, save.Settings.SFXVol)

	bgWorld := game.NewWorldWithSeed(777)
	// Place peaceful idle girls along sidewalks for the live animated Title Screen backdrop
	for i := range bgWorld.Girls {
		bgWorld.Girls[i].X = float32(140 + i*145)
		if i%2 == 0 {
			bgWorld.Girls[i].Y = 118
		} else {
			bgWorld.Girls[i].Y = 582
		}
	}
	// Keep idle player off-screen so Title Screen shows only the street & idle girls
	bgWorld.Player.X = -300
	bgWorld.Player.Y = -300

	return &App{
		world:    game.NewWorld(),
		bgWorld:  bgWorld,
		renderer: renderer,
		post:     render.NewPostProcessor(),
		reader:   input.NewReader(),
		ui:       ui.NewManager(),
		audio:    audioEng,
		save:     save,
		tracker:  meta.NewRunTracker(),
	}
}

func (a *App) startRun(bossRush bool) {
	a.world = game.NewWorldWithOptions(int64(a.tick*100000)+12345, bossRush)
	a.tracker = meta.NewRunTracker()
	a.overdriveAck = false
	a.lastPerfectAim = false
	a.ui.Scene = ui.ScenePlaying
	a.applyPrefsToRenderer()
}

func (a *App) startTutorial() {
	a.world = game.NewTutorialWorld(&a.save)
	a.tracker = meta.NewRunTracker()
	a.overdriveAck = false
	a.lastPerfectAim = false
	a.ui.Scene = ui.ScenePlaying
	a.applyPrefsToRenderer()
}

func (a *App) applyPrefsToRenderer() {
	a.renderer.SetVisualPrefs(
		a.save.Profile.SelectedPlayerSkin,
		a.save.Profile.SelectedBusSkin,
		a.save.Settings.ScreenShake,
		a.save.Settings.ShowFPS,
	)
}

// Update advances menus or active gameplay simulation, audio ducking, and challenge tracking.
func (a *App) Update() error {
	const dt = float32(1.0 / 60.0)
	a.tick += dt

	if input.AnyInputJustPressed() {
		a.audio.Unlock()
	}

	// Keep Title background camera gently panning over the Pop-Urban city
	a.bgWorld.CameraX = (game.ArenaWidth - game.ScreenWidth) * 0.5 * (1.0 + float32(math.Sin(float64(a.tick*0.18))))
	a.bgWorld.CameraY = (game.ArenaHeight - game.ScreenHeight) * 0.5 * (1.0 + float32(math.Cos(float64(a.tick*0.12))))
	for i := range a.bgWorld.Girls {
		a.bgWorld.Girls[i].AnimTimer += dt
	}

	// Sync dynamic Synthwave BPM & Chronos/Pause low-pass filter
	isPlaying := a.ui.Scene == ui.ScenePlaying
	isPaused := a.ui.Scene == ui.ScenePaused ||
		a.ui.Scene == ui.SceneTutorial ||
		a.ui.Scene == ui.SceneOptions ||
		a.ui.Scene == ui.SceneControls ||
		(a.world != nil && (a.world.PendingLevelUp || a.world.ShopMenu.Active || a.world.SlotModal.Active))
	tier := 1
	bossActive := false
	wanted := 0
	chronos := false
	fever := false
	if a.world != nil && (isPlaying || a.ui.Scene == ui.ScenePaused || a.ui.Scene == ui.SceneTutorial) {
		tier = a.world.DifficultyTier
		bossActive = a.world.Boss.Active
		wanted = a.world.Police.Stars
		chronos = a.world.Player.ChronosActive || a.world.BulletFreezeTimer > 0
		fever = isPlaying && a.world.StreakCount >= 5
		a.world.MusicBPM = a.audio.CurrentBPM()
	}
	a.audio.UpdateMusicState(isPlaying, tier, bossActive, wanted, chronos, isPaused)
	a.audio.SetFever(fever)

	// Handle non-gameplay UI scenes
	if a.ui.Scene != ui.ScenePlaying {
		if a.ui.Scene == ui.SceneGameOver && input.IsRestartJustPressed() {
			a.audio.PlaySFX(audio.SFXUIClick)
			a.startRun(a.world != nil && a.world.BossRushMode)
			return nil
		}

		mi := a.reader.ReadMenuInput()
		a.renderer.SetMouseScreen(mi.MouseX, mi.MouseY)
		res := a.ui.Update(mi, &a.save)
		if res.PlaySFX {
			a.audio.PlaySFX(audio.SFXUIClick)
		}

		switch res.Action {
		case ui.ActionStartNormal:
			a.startRun(false)
		case ui.ActionStartTutorial:
			a.startTutorial()
		case ui.ActionStartBossRush:
			a.startRun(true)
		case ui.ActionResume:
			a.ui.Scene = ui.ScenePlaying
		case ui.ActionRestart:
			if a.world != nil && a.world.TutorialMode {
				a.startTutorial()
			} else {
				a.startRun(a.world != nil && a.world.BossRushMode)
			}
		case ui.ActionContinueOverdrive:
			a.overdriveAck = true
			a.ui.Scene = ui.ScenePlaying
		case ui.ActionToTitle:
			a.ui.OpenScene(ui.SceneTitle, ui.SceneTitle)
		case ui.ActionBack:
			_ = meta.Save(a.save)
			a.ui.OpenScene(a.ui.ReturnScene, a.ui.ReturnScene)
		case ui.ActionSettingsChanged:
			a.audio.SetVolumes(a.save.Settings.MasterVol, a.save.Settings.MusicVol, a.save.Settings.SFXVol)
			ebiten.SetFullscreen(a.save.Settings.Fullscreen)
			a.applyPrefsToRenderer()
			_ = meta.Save(a.save)
		case ui.ActionQuit:
			_ = meta.Save(a.save)
			return ebiten.Termination
		}
		return nil
	}

	// Pause toggle only when no in-run modal is capturing [ESC]
	if !a.world.PendingLevelUp && !a.world.ShopMenu.Active && !a.world.SlotModal.Active {
		if input.IsPauseJustPressed() {
			a.audio.PlaySFX(audio.SFXUIClick)
			a.ui.OpenScene(ui.ScenePaused, ui.ScenePlaying)
			return nil
		}
	}

	stopPhase := 0
	var stopScreenX, stopScreenY float32
	canThrowExtraStop := a.world.Player.StopCooldownTimer <= 0 && a.world.ActiveStopCount() < a.world.Player.EffectiveMaxStops()
	if !a.world.PendingLevelUp && !a.world.ShopMenu.Active && !a.world.SlotModal.Active {
		bestD2 := float32(1e12)
		for i := range a.world.Stops {
			st := &a.world.Stops[i]
			if !st.Active || st.BusRushTriggered || !st.PlayerInside {
				continue
			}
			d2 := (a.world.Player.X-st.X)*(a.world.Player.X-st.X) + (a.world.Player.Y-st.Y)*(a.world.Player.Y-st.Y)
			if d2 < bestD2 {
				bestD2 = d2
				stopScreenX = st.X - a.world.CameraX
				stopScreenY = st.Y - a.world.CameraY
				if !st.Minigame.Completed {
					if st.Arch == game.ArchStudent {
						stopPhase = 2
					} else {
						stopPhase = 1
					}
				} else if st.Combo.Active && !st.Combo.Completed {
					stopPhase = 2
				}
			}
		}
	}

	in := a.reader.Read(
		a.world.CameraX,
		a.world.CameraY,
		int(a.world.Player.SelectedStopArch),
		stopPhase,
		stopScreenX,
		stopScreenY,
		canThrowExtraStop,
		a.save.Settings.ChronosToggle,
	)

	// Mouse click routing for Level-Up cards, Shop modal, Slot modal, Tutorial HUD buttons, and Hotbar dock
	if in.MouseLeftJustPressed {
		mx, my := in.MouseScreenX, in.MouseScreenY
		if a.world.PendingLevelUp {
			for i := 0; i < 3; i++ {
				if render.LevelUpCardRect(i).Contains(mx, my) {
					in.ChooseUpgrade = i
					a.audio.PlaySFX(audio.SFXUIClick)
					break
				}
			}
		} else if a.world.ShopMenu.Active {
			if render.ShopCloseRect().Contains(mx, my) {
				in.CancelJustPressed = true
				a.audio.PlaySFX(audio.SFXUIClick)
			} else {
				for i := 0; i < 4; i++ {
					if render.ShopChoiceRect(i).Contains(mx, my) {
						in.ShopAction = i + 1
						break
					}
				}
			}
		} else if a.world.SlotModal.Active {
			if render.SlotSpinRect().Contains(mx, my) {
				in.SlotSpinJustPressed = true
			} else if render.SlotCloseRect().Contains(mx, my) {
				in.CancelJustPressed = true
				a.audio.PlaySFX(audio.SFXUIClick)
			}
		} else {
			if a.world.TutorialMode {
				if render.TutorialExitRect().Contains(mx, my) {
					in.ThrowJustPressed = false
					in.ActionDir = game.ActionNone
					in.ActionHeld = false
					a.audio.PlaySFX(audio.SFXUIClick)
					a.ui.OpenScene(ui.SceneTitle, ui.SceneTitle)
					return nil
				}
				if render.TutorialGuideRect().Contains(mx, my) {
					in.ThrowJustPressed = false
					in.ActionDir = game.ActionNone
					in.ActionHeld = false
					a.audio.PlaySFX(audio.SFXUIClick)
					a.ui.OpenScene(ui.SceneTutorial, ui.ScenePlaying)
					return nil
				}
			}
			for i := 0; i < int(game.ArchCount); i++ {
				if render.HotbarSlotRect(i).Contains(mx, my) {
					in.SelectArchetype = i
					in.ThrowJustPressed = false
					in.ActionDir = game.ActionNone
					in.ActionHeld = false
					a.audio.PlaySFX(audio.SFXUIClick)
					break
				}
			}
		}
	}

	aimX, aimY := a.world.ResolveAim(in)
	if canThrowExtraStop {
		isPerf := a.world.EvaluateAim(aimX, aimY).IsPerfectDrop
		if isPerf && !a.lastPerfectAim {
			a.audio.PlaySFX(audio.SFXSlotTick)
		}
		a.lastPerfectAim = isPerf
	} else {
		a.lastPerfectAim = false
	}

	a.renderer.SetAim(aimX, aimY)
	a.renderer.SetMouseScreen(in.MouseScreenX, in.MouseScreenY)
	a.renderer.SetMouseHoverDir(in.MouseHoverDir)
	a.world.Update(in, dt)

	if !a.world.TutorialMode && !a.world.PendingLevelUp && !a.world.ShopMenu.Active && !a.world.SlotModal.Active && !a.world.GameOver {
		a.tracker.TickWanted(dt, a.world.Police.Stars)
	}

	a.processWorldEvents()

	if !a.world.TutorialMode {
		if a.world.GameOver {
			prevBest := a.save.Profile.BestScore
			a.tracker.FinalizeRun(a.world.Score, int(a.world.RunTime), a.world.GirlsBoarded, &a.save.Profile)
			_ = meta.Save(a.save)
			a.ui.SetEndRun(false, a.world, a.world.Score > prevBest, a.tracker.NewlyUnlocked)
		} else if a.world.Victory && !a.overdriveAck {
			prevBest := a.save.Profile.BestScore
			a.tracker.FinalizeRun(a.world.Score, int(a.world.RunTime), a.world.GirlsBoarded, &a.save.Profile)
			_ = meta.Save(a.save)
			a.ui.SetEndRun(true, a.world, a.world.Score > prevBest, a.tracker.NewlyUnlocked)
		}
	}

	return nil
}

func (a *App) processWorldEvents() {
	isTutorial := a.world != nil && a.world.TutorialMode
	events := a.world.DrainEvents()
	for _, ev := range events {
		switch ev.Kind {
		case game.EventThrow:
			a.audio.PlaySFX(audio.SFXThrow)
		case game.EventPerfectDrop:
			a.audio.PlaySFX(audio.SFXPerfectDrop)
			if !isTutorial {
				a.tracker.PerfectDrops++
			}
		case game.EventHook:
			a.audio.PlaySFX(audio.SFXHook)
			if !isTutorial && ev.Value > a.tracker.MaxHookInOneStop {
				a.tracker.MaxHookInOneStop = ev.Value
			}
		case game.EventMinigameHit:
			a.audio.PlaySFX(audio.SFXMinigameHit)
		case game.EventMinigameMiss:
			a.audio.PlaySFX(audio.SFXMinigameMiss)
		case game.EventComboStep:
			step := ev.Value
			if step < 0 {
				step = 0
			} else if step > 3 {
				step = 3
			}
			a.audio.PlaySFX(audio.SFXComboStep1 + audio.SFXID(step))
		case game.EventBusRush:
			a.audio.PlaySFX(audio.SFXBusHorn)
		case game.EventDash:
			a.audio.PlaySFX(audio.SFXDash)
		case game.EventPlayerHit:
			a.audio.PlaySFX(audio.SFXPlayerHit)
			if !isTutorial {
				a.tracker.OnPlayerHit(ev.Flag)
			}
		case game.EventPoliceStarUp:
			a.audio.PlaySFX(audio.SFXPoliceSiren)
		case game.EventBossSpawn:
			a.audio.PlaySFX(audio.SFXBossSpawn)
		case game.EventBossBreak:
			a.audio.PlaySFX(audio.SFXBossBreak)
		case game.EventBossDefeated:
			if !isTutorial {
				starsArg := 0
				if ev.Flag {
					starsArg = 3
				}
				a.tracker.RecordBossDefeated(ev.Value, starsArg, &a.save.Profile)
			}
		case game.EventLevelUp:
			a.audio.PlaySFX(audio.SFXLevelUp)
		case game.EventShopBuy:
			a.audio.PlaySFX(audio.SFXShopBuy)
		case game.EventSlotTick:
			a.audio.PlaySFX(audio.SFXSlotTick)
		case game.EventSlotWin:
			a.audio.PlaySFX(audio.SFXSlotWin)
		case game.EventSlotJackpot:
			a.audio.PlaySFX(audio.SFXSlotJackpot)
			if !isTutorial {
				a.tracker.JackpotWon = true
			}
		case game.EventSlotBust:
			a.audio.PlaySFX(audio.SFXSlotBust)
		case game.EventSynergy:
			if !isTutorial && ev.Value >= 0 && ev.Value < 4 {
				a.tracker.SynergyMask |= 1 << uint(ev.Value)
			}
		case game.EventGirlBoarded:
			if !isTutorial {
				a.tracker.GirlsBoarded += ev.Value
			}
		case game.EventGraze:
			a.audio.PlaySFX(audio.SFXSlotTick)
		case game.EventDashParry:
			a.audio.PlaySFX(audio.SFXPerfectDrop)
		case game.EventStopRecall:
			a.audio.PlaySFX(audio.SFXThrow)
		case game.EventSlipstream:
			a.audio.PlaySFX(audio.SFXDash)
		case game.EventFeverUp:
			a.audio.PlaySFX(audio.SFXComboStep4)
		case game.EventHealPickup:
			if ev.Value >= 2 {
				a.audio.PlaySFX(audio.SFXSlotWin)
			} else {
				a.audio.PlaySFX(audio.SFXShopBuy)
			}
		}
	}

	if !isTutorial {
		if unlocked := a.tracker.CheckAndUnlock(&a.save.Profile); len(unlocked) > 0 {
			for _, ch := range unlocked {
				a.audio.PlaySFX(audio.SFXChallengeUnlock)
				a.renderer.ShowChallengeToast(ch.TitleIT)
			}
			_ = meta.Save(a.save)
		}
	}
}

// Draw renders the world through the Kage post-processing shader, then draws razor-sharp
// HUD, overlays, and UI menus directly onto screen with zero CRT distortion.
func (a *App) Draw(screen *ebiten.Image) {
	buf := a.post.BeginScene()
	a.renderer.SetPlayingActive(a.ui.Scene == ui.ScenePlaying)

	useTitleBg := a.ui.Scene == ui.SceneTitle ||
		((a.ui.Scene == ui.SceneTutorial || a.ui.Scene == ui.SceneOptions || a.ui.Scene == ui.SceneControls || a.ui.Scene == ui.SceneChallenges) &&
			a.ui.ReturnScene == ui.SceneTitle)

	activeWorld := a.world
	if useTitleBg {
		activeWorld = a.bgWorld
	}
	a.renderer.DrawWorldScene(buf, activeWorld, false)

	var chronosAmt float32
	if a.ui.Scene == ui.ScenePlaying && a.world != nil && (a.world.Player.ChronosActive || a.world.BulletFreezeTimer > 0) {
		chronosAmt = 1.0
	}
	var rewindAmt float32
	if a.ui.Scene == ui.SceneGameOver {
		rewindAmt = 0.85
	}

	a.post.Present(screen, a.save.Settings.FilterMode, a.tick, chronosAmt, rewindAmt)

	if !useTitleBg {
		a.renderer.DrawHUDAndOverlays(screen, a.world)
	}
	if a.ui.Scene != ui.ScenePlaying {
		mx, my := ebiten.CursorPosition()
		a.ui.Draw(screen, a.renderer, &a.save, a.world, float32(mx), float32(my))
	}
}

// Layout fixes the internal virtual resolution at 640x360 (16:9 pixel-art).
func (a *App) Layout(_, _ int) (int, int) {
	return game.ScreenWidth, game.ScreenHeight
}

func main() {
	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("GIRLBUS: NEON STOP (v2.1)")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(NewApp()); err != nil {
		log.Fatal(err)
	}
}
