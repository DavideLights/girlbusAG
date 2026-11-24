package ui

import (
	"runtime"

	"github.com/hajimehoshi/ebiten/v2"

	"girlbusAG/internal/game"
	"girlbusAG/internal/input"
	"girlbusAG/internal/meta"
	"girlbusAG/internal/render"
)

// Scene identifies the active high-level UI/game state.
type Scene int

const (
	SceneTitle Scene = iota
	ScenePlaying
	ScenePaused
	SceneTutorial
	SceneOptions
	SceneControls
	SceneChallenges
	SceneGameOver
	SceneVictory
)

// Action communicates state transitions or audio/save triggers from UI to main.go.
type Action int

const (
	ActionNone Action = iota
	ActionStartNormal
	ActionStartTutorial
	ActionStartBossRush
	ActionResume
	ActionRestart
	ActionQuit
	ActionOpenOptions
	ActionOpenControls
	ActionOpenChallenges
	ActionBack
	ActionToTitle
	ActionContinueOverdrive
	ActionSettingsChanged
)

// Result summarizes what happened during Manager.Update this frame.
type Result struct {
	Action  Action
	PlaySFX bool
}

// Manager handles all game menus, visual tutorial, options, challenges/skins, and end-of-run overlays.
type Manager struct {
	Scene          Scene
	ReturnScene    Scene
	SelectedIdx    int
	TutorialTab    int
	ChallengePage  int
	NewRecord      bool
	RunUnlocks     []meta.ChallengeDef
	LastWorldStats RunSummary
	isWasm         bool
	pointerX       float32
	pointerY       float32
	usingPointer   bool
	entryTicks     int
	entryLayer     *ebiten.Image
	titleArt       *ebiten.Image
	mapIndex       int
	mapPrevious    int
	mapSlide       int
	mapDirection   int
	mapPreviews    [titleMapCount]*ebiten.Image
	mapViewport    *ebiten.Image
	volumeDrag     int
}

// RunSummary captures end-of-run stats, build loadout, and seed for GameOver and Victory screens.
type RunSummary struct {
	Score           int
	RunTime         float32
	GirlsBoarded    int
	Bosses          int
	Tier            int
	BossRush        bool
	Seed            int64
	Level           int
	MaxStreak       int
	PeakWanted      int
	PerfectDrops    int
	EventsCompleted int
	BoardedByArch   [game.ArchCount]int
	UpgradeRanks    [game.UpgradeCount]int
	Relics          [game.RelicCount]bool
	Quip            string
}

// NewManager initializes the UI starting at SceneTitle.
func NewManager() *Manager {
	return &Manager{
		Scene:  SceneTitle,
		isWasm: runtime.GOOS == "js",
	}
}

// OpenScene switches to targetScene and remembers returnTo when applicable.
func (m *Manager) OpenScene(targetScene, returnTo Scene) {
	m.Scene = targetScene
	m.ReturnScene = returnTo
	m.SelectedIdx = 0
	m.entryTicks = 0
	m.usingPointer = false
	m.volumeDrag = 0
	if targetScene == SceneTutorial {
		m.TutorialTab = 0
	}
}

// SetEndRun displays GameOver or Victory with recorded summary, record flag, and newly unlocked challenges.
func (m *Manager) SetEndRun(victory bool, w *game.World, newRecord bool, unlocks []meta.ChallengeDef) {
	m.LastWorldStats = RunSummary{
		Score:           w.Score,
		RunTime:         w.RunTime,
		GirlsBoarded:    w.GirlsBoarded,
		Bosses:          w.BossesDefeated,
		Tier:            w.DifficultyTier,
		BossRush:        w.BossRushMode,
		Seed:            w.Seed,
		Level:           w.Player.Level,
		MaxStreak:       w.MaxStreak,
		PeakWanted:      w.PeakWanted,
		PerfectDrops:    w.PerfectDrops,
		EventsCompleted: w.EventsCompleted,
		BoardedByArch:   w.BoardedByArch,
		UpgradeRanks:    w.Player.UpgradeRanks,
		Relics:          w.Player.Relics,
		Quip:            game.EndRunQuip(victory, w.BossRushMode, w.PeakWanted, w.BossesDefeated, w.MaxStreak),
	}
	m.NewRecord = newRecord
	m.RunUnlocks = unlocks
	m.SelectedIdx = 0
	m.entryTicks = 0
	m.usingPointer = false
	if victory {
		m.Scene = SceneVictory
	} else {
		m.Scene = SceneGameOver
	}
}

// Update processes mouse hover/click and WASD/Gamepad menu navigation across all scenes.
func (m *Manager) Update(mi input.MenuInput, save *meta.SaveData) Result {
	res := Result{}
	if m.entryTicks < menuEntryDuration {
		m.entryTicks++
	}
	if m.mapSlide > 0 {
		m.mapSlide--
	}
	pointerMoved := mi.MouseX != m.pointerX || mi.MouseY != m.pointerY
	m.pointerX, m.pointerY = mi.MouseX, mi.MouseY
	switch {
	case mi.MouseLeftJustPressed:
		m.usingPointer = true
	case mi.Up || mi.Down || mi.Left || mi.Right || mi.Confirm || mi.Back:
		m.usingPointer = false
	case pointerMoved:
		m.usingPointer = true
	}
	pointerSelect := m.usingPointer && (pointerMoved || mi.MouseLeftJustPressed)

	switch m.Scene {
	case SceneTitle:
		if m.browseTitleMaps(mi) {
			return Result{PlaySFX: true}
		}
		maxItems := 7
		if m.isWasm {
			maxItems = 6
		}
		if mi.Up {
			m.SelectedIdx = (m.SelectedIdx + maxItems - 1) % maxItems
			res.PlaySFX = true
		} else if mi.Down {
			m.SelectedIdx = (m.SelectedIdx + 1) % maxItems
			res.PlaySFX = true
		}

		m.selectButton(&mi, pointerSelect, maxItems, titleButtonRect)

		if mi.Confirm {
			res.PlaySFX = true
			switch m.SelectedIdx {
			case 0:
				if m.titleMapPlayable() {
					res.Action = ActionStartNormal
				}
			case 1:
				m.OpenScene(SceneTutorial, SceneTitle)
			case 2:
				if m.titleMapPlayable() && save.Profile.IsBossRushUnlocked() {
					res.Action = ActionStartBossRush
				}
			case 3:
				m.OpenScene(SceneChallenges, SceneTitle)
			case 4:
				m.OpenScene(SceneOptions, SceneTitle)
			case 5:
				m.OpenScene(SceneControls, SceneTitle)
			case 6:
				if !m.isWasm {
					res.Action = ActionQuit
				}
			}
		}

	case ScenePaused:
		const maxItems = 5
		if mi.Back {
			res.Action = ActionResume
			res.PlaySFX = true
			return res
		}
		if mi.Up {
			m.SelectedIdx = (m.SelectedIdx + maxItems - 1) % maxItems
			res.PlaySFX = true
		} else if mi.Down {
			m.SelectedIdx = (m.SelectedIdx + 1) % maxItems
			res.PlaySFX = true
		}

		m.selectButton(&mi, pointerSelect, maxItems, pauseButtonRect)

		if mi.Confirm {
			res.PlaySFX = true
			switch m.SelectedIdx {
			case 0:
				res.Action = ActionResume
			case 1:
				res.Action = ActionRestart
			case 2:
				m.OpenScene(SceneTutorial, ScenePlaying)
			case 3:
				m.OpenScene(SceneOptions, ScenePaused)
			case 4:
				res.Action = ActionToTitle
			}
		}

	case SceneTutorial:
		if mi.Back {
			res.Action = ActionBack
			res.PlaySFX = true
			return res
		}
		if mi.Left {
			m.TutorialTab = (m.TutorialTab + 3) % 4
			res.PlaySFX = true
		} else if mi.Right {
			m.TutorialTab = (m.TutorialTab + 1) % 4
			res.PlaySFX = true
		}
		if mi.Up {
			m.SelectedIdx = (m.SelectedIdx + 2) % 3
			res.PlaySFX = true
		} else if mi.Down {
			m.SelectedIdx = (m.SelectedIdx + 1) % 3
			res.PlaySFX = true
		}
		if mi.Confirm {
			res.PlaySFX = true
			switch m.SelectedIdx {
			case 0:
				m.SelectedIdx = 1
			case 1:
				res.Action = ActionStartTutorial
			case 2:
				res.Action = ActionBack
			}
		}

		if mi.MouseLeftJustPressed {
			for t := 0; t < 4; t++ {
				if tutorialTabRect(t).Contains(mi.MouseX, mi.MouseY) {
					m.TutorialTab = t
					res.PlaySFX = true
					return res
				}
			}
			if tutorialPrevRect().Contains(mi.MouseX, mi.MouseY) {
				m.TutorialTab = (m.TutorialTab + 3) % 4
				res.PlaySFX = true
				return res
			}
			if tutorialNextRect().Contains(mi.MouseX, mi.MouseY) {
				m.TutorialTab = (m.TutorialTab + 1) % 4
				res.PlaySFX = true
				return res
			}
			if tutorialPracticeRect().Contains(mi.MouseX, mi.MouseY) {
				res.Action = ActionStartTutorial
				res.PlaySFX = true
				return res
			}
			if tutorialBackRect().Contains(mi.MouseX, mi.MouseY) {
				res.Action = ActionBack
				res.PlaySFX = true
				return res
			}
		}

	case SceneOptions:
		const maxItems = 10
		if mi.Back {
			res.Action = ActionBack
			res.PlaySFX = true
			return res
		}
		if mi.Up {
			m.SelectedIdx = (m.SelectedIdx + maxItems - 1) % maxItems
			res.PlaySFX = true
		} else if mi.Down {
			m.SelectedIdx = (m.SelectedIdx + 1) % maxItems
			res.PlaySFX = true
		}

		if !mi.MouseLeftPressed {
			m.volumeDrag = 0
		}
		for i := 0; i < maxItems; i++ {
			rect := optionsRowRect(i)
			if m.volumeDrag != 0 || !pointerSelect || !rect.Contains(mi.MouseX, mi.MouseY) {
				continue
			}
			m.SelectedIdx = i
			if !mi.MouseLeftJustPressed {
				continue
			}
			if i == 9 {
				return Result{Action: ActionBack, PlaySFX: true}
			}
			if i >= 1 && i <= 3 {
				barX, barW := optionsSlider(rect)
				if mi.MouseX >= barX-10 && mi.MouseX <= barX+barW+10 {
					m.volumeDrag = i
					break
				}
			}
			mi.Right = true
		}
		if m.volumeDrag != 0 {
			m.SelectedIdx = m.volumeDrag
			barX, barW := optionsSlider(optionsRowRect(m.volumeDrag))
			value := float64(clampUI((mi.MouseX-barX)/barW, 0, 1))
			volume := optionVolume(&save.Settings, m.volumeDrag)
			if *volume != value {
				*volume = value
				res.Action = ActionSettingsChanged
			}
			res.PlaySFX = mi.MouseLeftJustPressed
			return res
		}

		changed := false
		delta := 0
		if mi.Left {
			delta = -1
		} else if mi.Right || (mi.Confirm && m.SelectedIdx != 9) {
			delta = 1
		}

		if delta != 0 {
			res.PlaySFX = true
			switch m.SelectedIdx {
			case 0:
				save.Settings.FilterMode = (save.Settings.FilterMode + 3 + delta) % 3
				changed = true
			case 1, 2, 3:
				volume := optionVolume(&save.Settings, m.SelectedIdx)
				value := clamp01F64(*volume + float64(delta)*0.10)
				changed = *volume != value
				*volume = value
			case 4:
				save.Settings.Fullscreen = !save.Settings.Fullscreen
				changed = true
			case 5:
				save.Settings.ScreenShake = !save.Settings.ScreenShake
				changed = true
			case 6:
				save.Settings.ShowFPS = !save.Settings.ShowFPS
				changed = true
			case 7:
				save.Settings.ChronosToggle = !save.Settings.ChronosToggle
				changed = true
			case 8:
				save.Settings.DirInputMode = (save.Settings.DirInputMode + 1) % 2
				changed = true
			}
		} else if mi.Confirm && m.SelectedIdx == 9 {
			res.Action = ActionBack
			res.PlaySFX = true
			return res
		}

		if changed {
			save.Normalize()
			res.Action = ActionSettingsChanged
		}

	case SceneControls:
		backRect := controlsBackRect()
		if mi.Back || mi.Confirm || (mi.MouseLeftJustPressed && backRect.Contains(mi.MouseX, mi.MouseY)) {
			res.Action = ActionBack
			res.PlaySFX = true
		}

	case SceneChallenges:
		if mi.Back {
			res.Action = ActionBack
			res.PlaySFX = true
			return res
		}
		// The cosmetic controls and return button come before the five reward cards.
		const maxFocus = 9
		if mi.Up {
			m.SelectedIdx = (m.SelectedIdx + maxFocus - 1) % maxFocus
			res.PlaySFX = true
		} else if mi.Down {
			m.SelectedIdx = (m.SelectedIdx + 1) % maxFocus
			res.PlaySFX = true
		}

		prevBtn := challengePrevRect()
		nextBtn := challengeNextRect()
		skinPBtn := challengeSkinPRect()
		skinBBtn := challengeSkinBRect()
		titleBtn := challengeTitleRect()
		backBtn := challengeBackRect()

		if pointerSelect {
			for i, rect := range [...]game.Rect{skinPBtn, skinBBtn, titleBtn, backBtn} {
				if rect.Contains(mi.MouseX, mi.MouseY) {
					m.SelectedIdx = i
				}
			}
			for i := 0; i < 5; i++ {
				if challengeRowRect(i).Contains(mi.MouseX, mi.MouseY) {
					m.SelectedIdx = 4 + i
				}
			}
		}

		if mi.MouseLeftJustPressed {
			if prevBtn.Contains(mi.MouseX, mi.MouseY) {
				m.ChallengePage = 0
				res.PlaySFX = true
			} else if nextBtn.Contains(mi.MouseX, mi.MouseY) {
				m.ChallengePage = 1
				res.PlaySFX = true
			} else if skinPBtn.Contains(mi.MouseX, mi.MouseY) {
				cyclePlayerSkin(save)
				res.PlaySFX = true
				res.Action = ActionSettingsChanged
			} else if skinBBtn.Contains(mi.MouseX, mi.MouseY) {
				cycleBusSkin(save)
				res.PlaySFX = true
				res.Action = ActionSettingsChanged
			} else if titleBtn.Contains(mi.MouseX, mi.MouseY) {
				cycleProfileTitle(save)
				res.PlaySFX = true
				res.Action = ActionSettingsChanged
			} else if backBtn.Contains(mi.MouseX, mi.MouseY) {
				res.Action = ActionBack
				res.PlaySFX = true
			}
		} else if mi.Left || mi.Right {
			res.PlaySFX = true
			switch m.SelectedIdx {
			case 0:
				cyclePlayerSkin(save)
				res.Action = ActionSettingsChanged
			case 1:
				cycleBusSkin(save)
				res.Action = ActionSettingsChanged
			case 2:
				cycleProfileTitle(save)
				res.Action = ActionSettingsChanged
			default:
				m.ChallengePage = (m.ChallengePage + 1) % 2
			}
		} else if mi.Confirm {
			res.PlaySFX = true
			switch m.SelectedIdx {
			case 0:
				cyclePlayerSkin(save)
				res.Action = ActionSettingsChanged
			case 1:
				cycleBusSkin(save)
				res.Action = ActionSettingsChanged
			case 2:
				cycleProfileTitle(save)
				res.Action = ActionSettingsChanged
			case 3:
				res.Action = ActionBack
			}
		}

	case SceneGameOver:
		const maxItems = 2
		if mi.Up || mi.Left {
			m.SelectedIdx = (m.SelectedIdx + maxItems - 1) % maxItems
			res.PlaySFX = true
		} else if mi.Down || mi.Right {
			m.SelectedIdx = (m.SelectedIdx + 1) % maxItems
			res.PlaySFX = true
		}
		m.selectButton(&mi, pointerSelect, maxItems, func(i int) game.Rect { return endRunButtonRect(i, maxItems) })
		if mi.Confirm {
			res.PlaySFX = true
			if m.SelectedIdx == 0 {
				res.Action = ActionRestart
			} else {
				res.Action = ActionToTitle
			}
		}

	case SceneVictory:
		const maxItems = 3
		if mi.Up || mi.Left {
			m.SelectedIdx = (m.SelectedIdx + maxItems - 1) % maxItems
			res.PlaySFX = true
		} else if mi.Down || mi.Right {
			m.SelectedIdx = (m.SelectedIdx + 1) % maxItems
			res.PlaySFX = true
		}
		m.selectButton(&mi, pointerSelect, maxItems, func(i int) game.Rect { return endRunButtonRect(i, maxItems) })
		if mi.Confirm {
			res.PlaySFX = true
			switch m.SelectedIdx {
			case 0:
				res.Action = ActionContinueOverdrive
			case 1:
				res.Action = ActionRestart
			case 2:
				res.Action = ActionToTitle
			}
		}
	}

	return res
}

func (m *Manager) selectButton(mi *input.MenuInput, pointerSelect bool, count int, rect func(int) game.Rect) {
	if !pointerSelect {
		return
	}
	for i := 0; i < count; i++ {
		if rect(i).Contains(mi.MouseX, mi.MouseY) {
			m.SelectedIdx = i
			mi.Confirm = mi.Confirm || mi.MouseLeftJustPressed
			return
		}
	}
}

// optionVolume is shared by slider drawing, click and drag, and keyboard changes.
// Rows 1 through 3 display master, music, and SFX volume in this order.
func optionVolume(s *meta.Settings, row int) *float64 {
	return [...]*float64{&s.MasterVol, &s.MusicVol, &s.SFXVol}[row-1]
}

func cyclePlayerSkin(save *meta.SaveData) {
	for step := 1; step <= meta.PlayerSkinCount; step++ {
		cand := (save.Profile.SelectedPlayerSkin + step) % meta.PlayerSkinCount
		if save.Profile.IsPlayerSkinUnlocked(cand) {
			save.Profile.SelectedPlayerSkin = cand
			return
		}
	}
}

func cycleBusSkin(save *meta.SaveData) {
	for step := 1; step <= meta.BusSkinCount; step++ {
		cand := (save.Profile.SelectedBusSkin + step) % meta.BusSkinCount
		if save.Profile.IsBusSkinUnlocked(cand) {
			save.Profile.SelectedBusSkin = cand
			return
		}
	}
}

func cycleProfileTitle(save *meta.SaveData) {
	titles := save.Profile.UnlockedTitles()
	if len(titles) == 0 {
		save.Profile.SelectedTitle = ""
		return
	}
	curIdx := 0
	for i, t := range titles {
		if t == save.Profile.SelectedTitle {
			curIdx = i
			break
		}
	}
	save.Profile.SelectedTitle = titles[(curIdx+1)%len(titles)]
}

// Draw renders the active UI overlay directly onto the razor-sharp 640x360 screen.
func (m *Manager) Draw(dst *ebiten.Image, r *render.Renderer, save *meta.SaveData, w *game.World, mx, my float32) {
	if m.Scene == ScenePlaying {
		return
	}
	if m.entryTicks >= menuEntryDuration {
		m.drawScene(dst, r, save, w, mx, my)
		return
	}

	if m.entryLayer == nil {
		m.entryLayer = ebiten.NewImage(game.ScreenWidth, game.ScreenHeight)
	}
	m.entryLayer.Clear()
	m.drawScene(m.entryLayer, r, save, w, mx, my)
	op := &ebiten.DrawImageOptions{}
	progress := float32(m.entryTicks) / menuEntryDuration
	op.ColorScale.ScaleAlpha(1 - 0.2*(1-progress)*(1-progress))
	dst.DrawImage(m.entryLayer, op)
}

func (m *Manager) drawScene(dst *ebiten.Image, r *render.Renderer, save *meta.SaveData, w *game.World, mx, my float32) {
	switch m.Scene {
	case SceneTitle:
		m.drawTitle(dst, r, save, mx, my)
	case ScenePaused:
		m.drawPaused(dst, r, w, mx, my)
	case SceneTutorial:
		m.drawTutorial(dst, r, mx, my)
	case SceneOptions:
		m.drawOptions(dst, r, save, mx, my)
	case SceneControls:
		m.drawControls(dst, r, mx, my)
	case SceneChallenges:
		m.drawChallenges(dst, r, save, mx, my)
	case SceneGameOver:
		m.drawEndRun(dst, r, save, false, mx, my)
	case SceneVictory:
		m.drawEndRun(dst, r, save, true, mx, my)
	}
}

func clampUI(v, lo, hi float32) float32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func clamp01F64(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
