package ui

import (
	"fmt"
	"image/color"
	"runtime"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"girlbusAG/assets"
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
}

// RunSummary captures end-of-run stats for GameOver and Victory screens.
type RunSummary struct {
	Score        int
	RunTime      float32
	GirlsBoarded int
	Bosses       int
	Tier         int
	BossRush     bool
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
	if targetScene == SceneTutorial {
		m.TutorialTab = 0
	}
}

// SetEndRun displays GameOver or Victory with recorded summary, record flag, and newly unlocked challenges.
func (m *Manager) SetEndRun(victory bool, w *game.World, newRecord bool, unlocks []meta.ChallengeDef) {
	m.LastWorldStats = RunSummary{
		Score:        w.Score,
		RunTime:      w.RunTime,
		GirlsBoarded: w.GirlsBoarded,
		Bosses:       w.BossesDefeated,
		Tier:         w.DifficultyTier,
		BossRush:     w.BossRushMode,
	}
	m.NewRecord = newRecord
	m.RunUnlocks = unlocks
	m.SelectedIdx = 0
	if victory {
		m.Scene = SceneVictory
	} else {
		m.Scene = SceneGameOver
	}
}

// Update processes mouse hover/click and WASD/Gamepad menu navigation across all scenes.
func (m *Manager) Update(mi input.MenuInput, save *meta.SaveData) Result {
	res := Result{}

	switch m.Scene {
	case SceneTitle:
		maxItems := 6
		if m.isWasm {
			maxItems = 5
		}
		if mi.Up {
			m.SelectedIdx = (m.SelectedIdx + maxItems - 1) % maxItems
			res.PlaySFX = true
		} else if mi.Down {
			m.SelectedIdx = (m.SelectedIdx + 1) % maxItems
			res.PlaySFX = true
		}

		for i := 0; i < maxItems; i++ {
			rect := titleButtonRect(i)
			if rect.Contains(mi.MouseX, mi.MouseY) {
				m.SelectedIdx = i
				if mi.MouseLeftJustPressed {
					mi.Confirm = true
				}
			}
		}

		if mi.Confirm {
			res.PlaySFX = true
			switch m.SelectedIdx {
			case 0:
				res.Action = ActionStartNormal
			case 1:
				m.OpenScene(SceneTutorial, SceneTitle)
			case 2:
				if save.Profile.IsBossRushUnlocked() {
					res.Action = ActionStartBossRush
				}
			case 3:
				m.OpenScene(SceneChallenges, SceneTitle)
			case 4:
				m.OpenScene(SceneOptions, SceneTitle)
			case 5:
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

		for i := 0; i < maxItems; i++ {
			rect := pauseButtonRect(i)
			if rect.Contains(mi.MouseX, mi.MouseY) {
				m.SelectedIdx = i
				if mi.MouseLeftJustPressed {
					mi.Confirm = true
				}
			}
		}

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
		const maxItems = 9
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

		changed := false
		for i := 0; i < maxItems; i++ {
			rect := optionsRowRect(i)
			if rect.Contains(mi.MouseX, mi.MouseY) {
				m.SelectedIdx = i
				if i >= 1 && i <= 3 && mi.MouseLeftPressed {
					// Direct click or drag anywhere on the slider bar
					barX := rect.X + 206
					barW := float32(144)
					if mi.MouseX >= barX-10 && mi.MouseX <= barX+barW+10 {
						frac := float64(clampUI((mi.MouseX-barX)/barW, 0, 1))
						switch i {
						case 1:
							save.Settings.MasterVol = frac
						case 2:
							save.Settings.MusicVol = frac
						case 3:
							save.Settings.SFXVol = frac
						}
						changed = true
						if mi.MouseLeftJustPressed {
							res.PlaySFX = true
						}
					} else if mi.MouseLeftJustPressed {
						mi.Right = true
					}
				} else if mi.MouseLeftJustPressed {
					if i == 8 {
						res.Action = ActionBack
						res.PlaySFX = true
						return res
					}
					mi.Right = true
				}
			}
		}

		delta := 0
		if mi.Left {
			delta = -1
		} else if mi.Right || (mi.Confirm && m.SelectedIdx != 8) {
			delta = 1
		}

		if delta != 0 {
			res.PlaySFX = true
			switch m.SelectedIdx {
			case 0:
				save.Settings.FilterMode = (save.Settings.FilterMode + 3 + delta) % 3
				changed = true
			case 1:
				save.Settings.MasterVol = clamp01F64(save.Settings.MasterVol + float64(delta)*0.10)
				changed = true
			case 2:
				save.Settings.MusicVol = clamp01F64(save.Settings.MusicVol + float64(delta)*0.10)
				changed = true
			case 3:
				save.Settings.SFXVol = clamp01F64(save.Settings.SFXVol + float64(delta)*0.10)
				changed = true
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
			}
		} else if mi.Confirm && m.SelectedIdx == 8 {
			res.Action = ActionBack
			res.PlaySFX = true
			return res
		}

		if changed {
			save.Normalize()
			res.Action = ActionSettingsChanged
		}

	case SceneControls:
		backRect := game.Rect{X: 248, Y: 312, W: 144, H: 28}
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
		const totalPages = 2 // 5 challenges per page = 10 total
		if mi.Up {
			m.ChallengePage = (m.ChallengePage + totalPages - 1) % totalPages
			res.PlaySFX = true
		} else if mi.Down {
			m.ChallengePage = (m.ChallengePage + 1) % totalPages
			res.PlaySFX = true
		}

		prevBtn := game.Rect{X: 34, Y: 256, W: 90, H: 22}
		nextBtn := game.Rect{X: 132, Y: 256, W: 90, H: 22}
		skinPBtn := game.Rect{X: 404, Y: 132, W: 196, H: 30}
		skinBBtn := game.Rect{X: 404, Y: 204, W: 196, H: 30}
		backBtn := game.Rect{X: 432, Y: 302, W: 140, H: 28}

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
			} else if backBtn.Contains(mi.MouseX, mi.MouseY) {
				res.Action = ActionBack
				res.PlaySFX = true
			}
		} else if mi.Left || mi.Right {
			cyclePlayerSkin(save)
			res.PlaySFX = true
			res.Action = ActionSettingsChanged
		} else if mi.Confirm {
			res.Action = ActionBack
			res.PlaySFX = true
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
		for i := 0; i < maxItems; i++ {
			rect := endRunButtonRect(i, 2)
			if rect.Contains(mi.MouseX, mi.MouseY) {
				m.SelectedIdx = i
				if mi.MouseLeftJustPressed {
					mi.Confirm = true
				}
			}
		}
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
		for i := 0; i < maxItems; i++ {
			rect := endRunButtonRect(i, 3)
			if rect.Contains(mi.MouseX, mi.MouseY) {
				m.SelectedIdx = i
				if mi.MouseLeftJustPressed {
					mi.Confirm = true
				}
			}
		}
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

// Draw renders the active UI overlay directly onto the razor-sharp 640x360 screen.
func (m *Manager) Draw(dst *ebiten.Image, r *render.Renderer, save *meta.SaveData, w *game.World, mx, my float32) {
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

func (m *Manager) drawTitle(dst *ebiten.Image, r *render.Renderer, save *meta.SaveData, mx, my float32) {
	// Soft atmospheric vignette over live panning city street
	vector.DrawFilledRect(dst, 0, 0, game.ScreenWidth, game.ScreenHeight, render.Premul(10, 6, 22, 148), false)

	// Crisp Neon Header Badge
	const hdrX, hdrY, hdrW, hdrH = float32(106), float32(16), float32(428), float32(60)
	vector.DrawFilledRect(dst, hdrX+3, hdrY+3, hdrW, hdrH, render.Premul(6, 2, 14, 175), false)
	vector.DrawFilledRect(dst, hdrX, hdrY, hdrW, hdrH, color.RGBA{16, 10, 30, 238}, false)
	vector.StrokeRect(dst, hdrX, hdrY, hdrW, hdrH, 2.0, render.ColKawaii, false)
	vector.StrokeRect(dst, hdrX+3, hdrY+3, hdrW-6, hdrH-6, 1.0, render.PremulColor(render.ColGyaru, 170), false)
	// Decorative neon top accent strip
	vector.DrawFilledRect(dst, hdrX+20, hdrY, hdrW-40, 2.0, render.ColGold, false)

	titleTxt := "GIRLBUS: NEON STOP"
	titleX := 320 - render.TextPixelWidth(titleTxt, 1.85)/2
	r.DrawScaledText(dst, titleTxt, titleX, 24, 1.85, render.ColGold, render.ColOutline)

	subTxt := "POP-URBAN ROGUE-LITE BULLET-HELL"
	subX := 320 - render.TextPixelWidth(subTxt, 1.0)/2
	r.DrawTextColored(dst, subTxt, subX, 55, render.ColGyaru)

	// Sleek Semi-Transparent Glass Card behind Menu Buttons
	maxItems := 6
	if m.isWasm {
		maxItems = 5
	}
	cardH := float32(maxItems*34 + 22)
	const cardX, cardY, cardW = float32(182), float32(84), float32(276)
	vector.DrawFilledRect(dst, cardX+3, cardY+3, cardW, cardH, render.Premul(6, 2, 14, 155), false)
	vector.DrawFilledRect(dst, cardX, cardY, cardW, cardH, render.Premul(14, 9, 28, 212), false)
	vector.StrokeRect(dst, cardX, cardY, cardW, cardH, 1.4, render.PremulColor(render.ColGyaru, 150), false)

	labels := []string{
		"NUOVA RUN",
		"TUTORIAL",
		"BOSS RUSH (6 BOSS)",
		"SFIDE & SKIN",
		"OPZIONI",
		"ESCI",
	}
	if m.isWasm {
		labels = labels[:5]
	}

	bossRushLocked := !save.Profile.IsBossRushUnlocked()
	for i, lbl := range labels {
		rect := titleButtonRect(i)
		selected := m.SelectedIdx == i || rect.Contains(mx, my)
		locked := i == 2 && bossRushLocked
		display := lbl
		if locked {
			display = "BOSS RUSH [BLOCCATA]"
		}
		if i == 1 {
			m.drawAccentButton(dst, r, rect, display, selected, render.ColGyaru, color.RGBA{14, 42, 56, 242})
		} else {
			m.drawButton(dst, r, rect, display, selected, locked)
		}
	}

	// Clean Footer Bar: Record/Unlocks + Mouse-Only Controls Cue
	unlocked := 0
	for _, ch := range meta.AllChallenges() {
		if save.Profile.UnlockedChallenges[ch.ID] {
			unlocked++
		}
	}
	const footY = float32(game.ScreenHeight - 24)
	vector.DrawFilledRect(dst, 0, footY, game.ScreenWidth, 24, render.ColHudBg, false)
	vector.DrawFilledRect(dst, 0, footY, game.ScreenWidth, 1.4, render.ColGyaru, false)

	stats := fmt.Sprintf("RECORD: %d | RUN: %d | SFIDE: %d/10",
		save.Profile.BestScore, save.Profile.TotalRuns, unlocked)
	r.DrawTextColored(dst, stats, 14, int(footY)+4, render.ColGold)

	// Right-aligned crisp mouse icon + zero-arrow-keys badge
	cueTxt := "MOUSE + WASD | ZERO FRECCE"
	cueW := render.TextPixelWidth(cueTxt, 1.0)
	cueX := game.ScreenWidth - cueW - 14
	drawSpriteIconCentered(dst, r, assets.SpriteIconMouseLMB, float32(cueX-11), footY+12, 1.0)
	r.DrawTextColored(dst, cueTxt, cueX, int(footY)+4, render.ColGyaru)
}

func (m *Manager) drawPaused(dst *ebiten.Image, r *render.Renderer, w *game.World, mx, my float32) {
	vector.DrawFilledRect(dst, 0, 0, game.ScreenWidth, game.ScreenHeight, render.Premul(10, 6, 20, 195), false)

	// Left Menu Panel
	vector.DrawFilledRect(dst, 46, 44, 250, 268, render.ColHudBg, false)
	vector.StrokeRect(dst, 46, 44, 250, 268, 2.0, render.ColGyaru, false)
	r.DrawScaledText(dst, "PAUSA", 142, 56, 1.6, render.ColGold, render.ColOutline)

	restartLbl := "RIAVVIA RUN"
	if w != nil && w.TutorialMode {
		restartLbl = "RIAVVIA TUTORIAL"
	}
	labels := [5]string{"RIPRENDI", restartLbl, "TUTORIAL", "OPZIONI", "MENU PRINCIPALE"}
	for i, lbl := range labels {
		rect := pauseButtonRect(i)
		selected := m.SelectedIdx == i || rect.Contains(mx, my)
		if i == 2 {
			m.drawAccentButton(dst, r, rect, lbl, selected, render.ColGyaru, color.RGBA{14, 42, 56, 242})
		} else {
			m.drawButton(dst, r, rect, lbl, selected, false)
		}
	}

	// Right Run Status & Relics Panel
	vector.DrawFilledRect(dst, 308, 44, 294, 268, render.ColHudBg, false)
	vector.StrokeRect(dst, 308, 44, 294, 268, 1.8, render.ColKawaii, false)
	r.DrawScaledText(dst, "STATO RUN & RELIQUIE", 320, 56, 1.15, render.ColKawaii, render.ColOutline)

	if w != nil {
		if w.TutorialMode {
			render.DrawPillBadge(dst, 320, 82, 268, 18, color.RGBA{14, 44, 54, 235}, render.ColGyaru)
			r.DrawTextColored(dst, "CAMPO DI PROVA INTERATTIVO (TUTORIAL)", 332, 83, render.ColGyaru)
		}
		mins := int(w.RunTime) / 60
		secs := int(w.RunTime) % 60
		statY := 84
		if w.TutorialMode {
			statY = 108
		}
		r.DrawTextColored(dst, fmt.Sprintf("Tempo: %02d:%02d | Tier %d | Score: %d", mins, secs, w.DifficultyTier, w.Score), 320, statY, color.White)
		r.DrawTextColored(dst, fmt.Sprintf("Ragazze: %d | Boss: %d | Cooldown: %.1fs", w.GirlsBoarded, w.BossesDefeated, w.Player.EffectiveStopCooldown()), 320, statY+18, render.ColGyaru)

		relicHdrY := statY + 42
		r.DrawTextColored(dst, "RELIQUIE ATTIVE:", 320, relicHdrY, render.ColGold)
		ry := relicHdrY + 18
		anyRelic := false
		for rk := 0; rk < int(game.RelicCount); rk++ {
			if w.Player.Relics[rk] {
				anyRelic = true
				spec := game.GetRelicSpec(game.RelicKind(rk))
				r.DrawTextColored(dst, "* "+spec.Name, 320, ry, render.ColGreenNeon)
				r.DrawTextColored(dst, "  "+spec.Desc, 320, ry+12, color.RGBA{200, 210, 230, 220})
				ry += 26
			}
		}
		if !anyRelic {
			r.DrawTextColored(dst, "Nessuna reliquia ancora (Negozio $120)", 320, ry, color.RGBA{165, 165, 190, 220})
		}
	}

	// Clean bottom cue
	cue := "CLICK SX SU UN PULSANTE | ESC PER RIPRENDERE"
	cueX := 320 - render.TextPixelWidth(cue, 1.0)/2
	r.DrawTextColored(dst, cue, cueX, 324, render.ColGyaru)
}

func (m *Manager) drawTutorial(dst *ebiten.Image, r *render.Renderer, mx, my float32) {
	// Darkened backdrop over live city / paused world
	vector.DrawFilledRect(dst, 0, 0, game.ScreenWidth, game.ScreenHeight, render.Premul(8, 5, 18, 206), false)

	// Centered Dark-Glass Modal: 584x316 at (28, 22)
	const modX, modY, modW, modH = float32(28), float32(22), float32(584), float32(316)
	vector.DrawFilledRect(dst, modX+3, modY+3, modW, modH, render.Premul(4, 2, 10, 185), false)
	vector.DrawFilledRect(dst, modX, modY, modW, modH, color.RGBA{14, 10, 26, 244}, false)
	vector.StrokeRect(dst, modX, modY, modW, modH, 2.0, render.ColGyaru, false)
	vector.StrokeRect(dst, modX+3, modY+3, modW-6, modH-6, 1.0, render.PremulColor(render.ColKawaii, 125), false)

	// Compact Header Strip
	r.DrawScaledText(dst, "GUIDA RAPIDA // GIRLBUS", 40, 28, 1.15, render.ColGold, render.ColOutline)
	badgeTxt := fmt.Sprintf("PAG %d/4 | SOLO MOUSE + WASD", m.TutorialTab+1)
	badgeW := float32(render.TextPixelWidth(badgeTxt, 1.0) + 26)
	badgeX := modX + modW - badgeW - 10
	render.DrawPillBadge(dst, badgeX, 27, badgeW, 18, color.RGBA{20, 16, 38, 240}, render.ColGyaru)
	drawSpriteIconCentered(dst, r, assets.SpriteIconMouseLMB, badgeX+11, 36, 0.85)
	r.DrawTextColored(dst, badgeTxt, int(badgeX+21), 28, render.ColGyaru)

	// Top Row: 4 Clickable Mouse Tabs
	tabLabels := [4]string{
		"1. FERMATE & BUS",
		"2. MINIGIOCHI",
		"3. SLOT & SHOP",
		"4. CURE & BOSS",
	}
	for t := 0; t < 4; t++ {
		rect := tutorialTabRect(t)
		active := m.TutorialTab == t
		hovered := rect.Contains(mx, my)
		bg := color.RGBA{22, 16, 38, 232}
		border := color.RGBA{75, 80, 118, 210}
		var txtCol color.Color = color.RGBA{190, 198, 222, 240}
		if hovered && !active {
			bg = color.RGBA{34, 26, 56, 242}
			border = render.ColGyaru
			txtCol = color.White
		}
		if active {
			bg = color.RGBA{18, 48, 64, 248}
			border = render.ColGold
			txtCol = render.ColGold
		}
		vector.DrawFilledRect(dst, rect.X, rect.Y, rect.W, rect.H, bg, false)
		vector.StrokeRect(dst, rect.X, rect.Y, rect.W, rect.H, 1.5, border, false)
		if active {
			vector.DrawFilledRect(dst, rect.X+2, rect.Y+rect.H-2.5, rect.W-4, 2.5, render.ColGyaru, false)
		}
		tw := render.TextPixelWidth(tabLabels[t], 1.0)
		tx := int(rect.X + (rect.W-float32(tw))*0.5)
		ty := int(rect.Y + 4)
		r.DrawScaledText(dst, tabLabels[t], tx, ty, 1.0, txtCol, render.ColOutline)
	}

	// Active Tab Illustrated Content (Y = 80..292)
	switch m.TutorialTab {
	case 0:
		m.drawTutorialTabStops(dst, r)
	case 1:
		m.drawTutorialTabMinigames(dst, r)
	case 2:
		m.drawTutorialTabSlotsShop(dst, r)
	case 3:
		m.drawTutorialTabSurvivalBoss(dst, r)
	}

	// Bottom Action Row (100% Mouse Clickable)
	prevR := tutorialPrevRect()
	nextR := tutorialNextRect()
	pracR := tutorialPracticeRect()
	backR := tutorialBackRect()

	m.drawButton(dst, r, prevR, "< PREC", prevR.Contains(mx, my), false)
	m.drawButton(dst, r, nextR, "SUCC >", nextR.Contains(mx, my), false)
	m.drawAccentButton(dst, r, pracR, "PROVA SUL CAMPO", pracR.Contains(mx, my), render.ColGreenNeon, color.RGBA{12, 54, 48, 248})
	m.drawButton(dst, r, backR, "TORNA [ESC]", backR.Contains(mx, my), false)
}

func (m *Manager) drawTutorialTabStops(dst *ebiten.Image, r *render.Renderer) {
	// Left Card: Controls (WASD + Mouse)
	vector.DrawFilledRect(dst, 38, 80, 276, 212, color.RGBA{20, 14, 36, 232}, false)
	vector.StrokeRect(dst, 38, 80, 276, 212, 1.3, render.PremulColor(render.ColGyaru, 185), false)
	r.DrawScaledText(dst, "COMANDI ESSENZIALI", 48, 86, 1.0, render.ColGold, render.ColOutline)

	// Row 1: WASD
	r.DrawKeycap(dst, "WASD", 46, 108, 48, 22, color.RGBA{34, 26, 58, 245}, render.ColGyaru, false)
	r.DrawTextColored(dst, "Muovi Alfredo (360 gradi)", 102, 108, color.White)
	r.DrawTextColored(dst, "Schiva i proiettili in strada", 102, 121, color.RGBA{175, 190, 215, 230})

	// Row 2: LMB
	drawTutorialIconBadge(dst, 46, 144, 48, 26, render.ColGold)
	drawSpriteIconCentered(dst, r, assets.SpriteIconMouseLMB, 70, 157, 1.15)
	r.DrawTextColored(dst, "CLICK SX: Mira & Lancia", 102, 143, render.ColGold)
	r.DrawTextColored(dst, "Fermata + Minigiochi & Combo", 102, 156, color.White)

	// Row 3: RMB / Space
	drawTutorialIconBadge(dst, 46, 180, 48, 26, render.ColKawaii)
	drawSpriteIconCentered(dst, r, assets.SpriteIconMouseRMB, 70, 193, 1.15)
	r.DrawTextColored(dst, "CLICK DX / SPAZIO: Dash", 102, 179, render.ColGyaru)
	r.DrawTextColored(dst, "Attraversa i proiettili!", 102, 192, color.White)

	// Row 4: Wheel / 1-4
	drawTutorialIconBadge(dst, 46, 216, 48, 26, render.ColEmo)
	drawSpriteIconCentered(dst, r, assets.SpriteIconMouseWheel, 70, 229, 1.15)
	r.DrawTextColored(dst, "ROTELLA / 1-4:", 102, 215, render.ColEmo)
	r.DrawTextColored(dst, "Cambia tipo di Fermata", 102, 228, color.White)

	// Row 5: Shift Chronos
	r.DrawKeycap(dst, "SHIFT", 46, 254, 48, 22, color.RGBA{18, 52, 68, 245}, render.ColStudent, false)
	r.DrawTextColored(dst, "Chronos (Slow-Mo tattico)", 102, 254, render.ColStudent)
	r.DrawTextColored(dst, "Rallenta il tempo per mirare", 102, 267, color.RGBA{175, 190, 215, 230})

	// Right Card: Bus Stops & Neon Bus
	vector.DrawFilledRect(dst, 322, 80, 280, 212, color.RGBA{20, 14, 36, 232}, false)
	vector.StrokeRect(dst, 322, 80, 280, 212, 1.3, render.PremulColor(render.ColKawaii, 185), false)
	r.DrawScaledText(dst, "FERMATE & AUTOBUS NEON", 332, 86, 1.0, render.ColGold, render.ColOutline)

	// Visual Stage: Stop -> Neon Bus
	vector.DrawFilledRect(dst, 332, 105, 260, 44, color.RGBA{12, 9, 24, 240}, false)
	vector.StrokeRect(dst, 332, 105, 260, 44, 1.0, render.PremulColor(render.ColGyaru, 120), false)
	drawSpriteIconCentered(dst, r, assets.SpriteStopEmo, 384, 127, 0.78)
	r.DrawScaledText(dst, ">>>", 436, 120, 1.1, render.ColGold, render.ColOutline)
	drawSpriteIconCentered(dst, r, assets.SpriteBusNeon, 526, 127, 0.78)

	// 3 Concise Callouts
	r.DrawTextColored(dst, "PERFECT DROP:", 332, 156, render.ColGold)
	r.DrawTextColored(dst, "Centra la ragazza (mirino oro) =", 332, 169, color.White)
	r.DrawTextColored(dst, "+30% gratis e spazza proiettili!", 332, 182, render.ColGreenNeon)

	r.DrawTextColored(dst, "ZONE BUS (+25%) vs ZONE ZTL:", 332, 202, render.ColGyaru)
	r.DrawTextColored(dst, "ZTL paga 3x Cash, ma +1* Polizia!", 332, 215, color.White)

	r.DrawTextColored(dst, "AUTOBUS NEON:", 332, 237, render.ColKawaii)
	r.DrawTextColored(dst, "Completa la fermata per chiamarlo:", 332, 250, color.White)
	r.DrawTextColored(dst, "raccoglie tutte + lascia scia turbo!", 332, 263, render.ColGold)
}

func (m *Manager) drawTutorialTabMinigames(dst *ebiten.Image, r *render.Renderer) {
	type miniCard struct {
		title  string
		line1  string
		line2  string
		sprite string
		accent color.RGBA
		x, y   float32
	}
	cards := [4]miniCard{
		{
			title:  "EMO (RHYTHM)",
			line1:  "Click Sx quando il cursore",
			line2:  "tocca il centro della barra.",
			sprite: assets.SpriteGirlEmo0,
			accent: render.ColEmo,
			x:      38, y: 80,
		},
		{
			title:  "STUDENTESSA (GRIGLIA 4x4)",
			line1:  "Click Sx nella direzione",
			line2:  "(su/giu/dx/sx) fino a PASS.",
			sprite: assets.SpriteGirlStudent0,
			accent: render.ColStudent,
			x:      324, y: 80,
		},
		{
			title:  "KAWAII (GOCCIOLINE)",
			line1:  "Danza nel cerchio & Click Sx",
			line2:  "mentre sfiori le gocce!",
			sprite: assets.SpriteGirlKawaii0,
			accent: render.ColKawaii,
			x:      38, y: 160,
		},
		{
			title:  "GYARU (BARRA TENSIONE)",
			line1:  "Tieni premuto Click Sx per",
			line2:  "restare nella zona verde.",
			sprite: assets.SpriteGirlGyaru0,
			accent: render.ColGyaru,
			x:      324, y: 160,
		},
	}

	for _, c := range cards {
		vector.DrawFilledRect(dst, c.x, c.y, 278, 74, color.RGBA{20, 14, 36, 234}, false)
		vector.StrokeRect(dst, c.x, c.y, 278, 74, 1.4, c.accent, false)

		drawTutorialIconBadge(dst, c.x+8, c.y+9, 44, 56, c.accent)
		drawSpriteIconCentered(dst, r, c.sprite, c.x+30, c.y+37, 1.15)

		r.DrawScaledText(dst, c.title, int(c.x+60), int(c.y+10), 1.0, c.accent, render.ColOutline)
		r.DrawTextColored(dst, c.line1, int(c.x+60), int(c.y+30), color.White)
		r.DrawTextColored(dst, c.line2, int(c.x+60), int(c.y+45), color.RGBA{200, 210, 232, 235})
	}

	// Bottom Highlight Strip: 4-Step Final Combo (Mouse Only!)
	vector.DrawFilledRect(dst, 38, 242, 564, 50, color.RGBA{24, 20, 44, 242}, false)
	vector.StrokeRect(dst, 38, 242, 564, 50, 1.6, render.ColGold, false)
	drawTutorialIconBadge(dst, 46, 249, 36, 36, render.ColGold)
	drawSpriteIconCentered(dst, r, assets.SpriteIconMouseLMB, 64, 267, 1.35)

	r.DrawScaledText(dst, "COMBO FINALE (4 STEP) - 100% SOLO MOUSE!", 92, 246, 1.0, render.ColGold, render.ColOutline)
	r.DrawTextColored(dst, "Clicca col Mouse nella direzione delle 4 frecce attorno alla fermata", 92, 262, color.White)
	r.DrawTextColored(dst, "per far arrivare il Bus Neon! (Zero frecce da tastiera!)", 92, 276, render.ColGyaru)
}

func (m *Manager) drawTutorialTabSlotsShop(dst *ebiten.Image, r *render.Renderer) {
	// Left Card: Rizz-Slots 777
	vector.DrawFilledRect(dst, 38, 80, 278, 212, color.RGBA{20, 14, 36, 234}, false)
	vector.StrokeRect(dst, 38, 80, 278, 212, 1.5, render.ColGold, false)

	drawTutorialIconBadge(dst, 46, 88, 38, 38, render.ColGold)
	drawSpriteIconCentered(dst, r, assets.SpritePropSlotMachine, 65, 107, 1.0)
	r.DrawScaledText(dst, "RIZZ-SLOTS 777", 92, 91, 1.05, render.ColGold, render.ColOutline)
	r.DrawTextColored(dst, "Avvicinati e CLICCA (o E)", 92, 108, render.ColGyaru)

	type iconRow struct {
		icon string
		head string
		desc string
		col  color.RGBA
	}
	slotRows := [4]iconRow{
		{assets.SpriteIconSlot777, "777 JACKPOT:", "Full HP + $250!", render.ColGold},
		{assets.SpriteIconSlotHeart, "CUORI:", "Cura HP istantanea", render.ColKawaii},
		{assets.SpriteIconSlotCash, "$$$ CASH:", "Premio in contanti", render.ColGreenNeon},
		{assets.SpriteIconSlotSiren, "SIRENA:", "+$110 ma +1* Polizia!", render.ColRedAlert},
	}
	for i, row := range slotRows {
		ry := float32(134 + i*38)
		drawTutorialIconBadge(dst, 48, ry, 28, 28, row.col)
		drawSpriteIconCentered(dst, r, row.icon, 62, ry+14, 1.0)
		r.DrawTextColored(dst, row.head, 84, int(ry+1), row.col)
		r.DrawTextColored(dst, row.desc, 84, int(ry+14), color.White)
	}

	// Right Card: Neon Vending Machine
	vector.DrawFilledRect(dst, 324, 80, 278, 212, color.RGBA{20, 14, 36, 234}, false)
	vector.StrokeRect(dst, 324, 80, 278, 212, 1.5, render.ColGyaru, false)

	drawTutorialIconBadge(dst, 332, 88, 38, 38, render.ColGyaru)
	drawSpriteIconCentered(dst, r, assets.SpriteShopVending, 351, 107, 1.0)
	r.DrawScaledText(dst, "DISTRIBUTORE NEON", 378, 91, 1.05, render.ColGyaru, render.ColOutline)
	r.DrawTextColored(dst, "Avvicinati e CLICCA (o E)", 378, 108, render.ColGold)

	shopRows := [4]iconRow{
		{assets.SpriteIconShopEspresso, "ESPRESSO ($35):", "+2 HP immediati!", render.ColGold},
		{assets.SpriteIconShopBribe, "TANGENTE:", "-2* Polizia subito", render.ColGreenNeon},
		{assets.SpriteIconShopRelic, "RELIQUIE:", "Bonus passivi unici", render.ColEmo},
		{assets.SpriteIconShopCrate, "CASSA UPGRADE:", "Potenziamento extra!", render.ColGyaru},
	}
	for i, row := range shopRows {
		ry := float32(134 + i*38)
		drawTutorialIconBadge(dst, 334, ry, 28, 28, row.col)
		drawSpriteIconCentered(dst, r, row.icon, 348, ry+14, 1.15)
		r.DrawTextColored(dst, row.head, 370, int(ry+1), row.col)
		r.DrawTextColored(dst, row.desc, 370, int(ry+14), color.White)
	}
}

func (m *Manager) drawTutorialTabSurvivalBoss(dst *ebiten.Image, r *render.Renderer) {
	// Row 1: HP & Espresso Heals
	vector.DrawFilledRect(dst, 38, 80, 564, 64, color.RGBA{20, 16, 36, 234}, false)
	vector.StrokeRect(dst, 38, 80, 564, 64, 1.4, render.ColGreenNeon, false)
	drawTutorialIconBadge(dst, 46, 88, 62, 48, render.ColGreenNeon)
	drawSpriteIconCentered(dst, r, assets.SpriteIconHeartFull, 63, 112, 1.4)
	drawSpriteIconCentered(dst, r, assets.SpriteIconPickupHeal, 90, 112, 1.45)

	r.DrawScaledText(dst, "VITA & CURE ESPRESSO", 118, 86, 1.05, render.ColGreenNeon, render.ColOutline)
	r.DrawTextColored(dst, "La barra HP e' sempre visibile sotto i piedi di Alfredo.", 118, 104, color.White)
	r.DrawTextColored(dst, "Raccogli le Cure Espresso luminose in strada e droppate dai Boss!", 118, 119, render.ColGyaru)

	// Row 2: Police Stars (1* - 5*)
	vector.DrawFilledRect(dst, 38, 154, 564, 64, color.RGBA{20, 16, 36, 234}, false)
	vector.StrokeRect(dst, 38, 154, 564, 64, 1.4, render.ColGold, false)
	drawTutorialIconBadge(dst, 46, 162, 62, 48, render.ColGold)
	drawSpriteIconCentered(dst, r, assets.SpriteIconStarWanted, 60, 190, 1.25)
	drawSpriteIconCentered(dst, r, assets.SpriteIconStarWanted, 77, 178, 1.45)
	drawSpriteIconCentered(dst, r, assets.SpriteIconStarWanted, 94, 190, 1.25)

	r.DrawScaledText(dst, "POLIZIA (1* - 5* STELLE)", 118, 160, 1.05, render.ColGold, render.ColOutline)
	r.DrawTextColored(dst, "Zone ZTL o ragazze VIP aumentano le stelle di ricerca.", 118, 178, color.White)
	r.DrawTextColored(dst, "Piu' pericolo, ma moltiplicatore fino a 3.5x XP e 4x Cash!", 118, 193, render.ColGold)

	// Row 3: Gatekeeper Bosses (3 Shields)
	vector.DrawFilledRect(dst, 38, 228, 564, 64, color.RGBA{20, 16, 36, 234}, false)
	vector.StrokeRect(dst, 38, 228, 564, 64, 1.4, render.ColKawaii, false)
	drawTutorialIconBadge(dst, 46, 236, 62, 48, render.ColKawaii)
	drawSpriteIconCentered(dst, r, assets.SpriteBossMom, 77, 260, 1.15)

	r.DrawScaledText(dst, "BOSS GATEKEEPER (3 SCUDI)", 118, 234, 1.05, render.ColKawaii, render.ColOutline)
	r.DrawTextColored(dst, "Non si abbattono sparando: agganciali 3 volte con una Fermata", 118, 252, color.White)
	r.DrawTextColored(dst, "per rompere gli scudi, droppare cure e salire di Tier!", 118, 267, render.ColGold)
}

func drawTutorialIconBadge(dst *ebiten.Image, x, y, w, h float32, border color.RGBA) {
	vector.DrawFilledRect(dst, x, y, w, h, color.RGBA{12, 8, 24, 238}, false)
	vector.StrokeRect(dst, x, y, w, h, 1.0, render.PremulColor(border, 175), false)
}

func drawSpriteIconCentered(dst *ebiten.Image, r *render.Renderer, spriteKey string, cx, cy, scale float32) {
	if r == nil || r.Sprites() == nil {
		return
	}
	img := r.Sprites().Get(spriteKey)
	if img == nil {
		return
	}
	b := img.Bounds()
	w := float32(b.Dx()) * scale
	h := float32(b.Dy()) * scale
	r.DrawSpriteIcon(dst, spriteKey, cx-w*0.5, cy-h*0.5, scale)
}

func (m *Manager) drawOptions(dst *ebiten.Image, r *render.Renderer, save *meta.SaveData, mx, my float32) {
	vector.DrawFilledRect(dst, 0, 0, game.ScreenWidth, game.ScreenHeight, render.Premul(10, 6, 20, 205), false)
	vector.DrawFilledRect(dst, 96, 18, 448, 326, render.ColHudBg, false)
	vector.StrokeRect(dst, 96, 18, 448, 326, 2.0, render.ColGyaru, false)

	title := "OPZIONI"
	r.DrawScaledText(dst, title, 320-render.TextPixelWidth(title, 1.4)/2, 24, 1.4, render.ColGold, render.ColOutline)

	filterNames := [3]string{"PIXEL CRISP (OFF)", "WARM CRT (CONSIGLIATO)", "VHS NOIR"}
	rows := [9]string{
		"FILTRO VIDEO",
		"VOLUME MASTER",
		"MUSICA SYNTH",
		"EFFETTI SFX",
		"SCHERMO INTERO",
		"SCREEN SHAKE",
		"MOSTRA FPS",
		"CHRONOS (SHIFT/Q)",
		"SALVA & INDIETRO",
	}

	for i, lbl := range rows {
		rect := optionsRowRect(i)
		selected := m.SelectedIdx == i || rect.Contains(mx, my)
		bg := color.RGBA{24, 16, 40, 224}
		border := color.RGBA{80, 90, 130, 200}
		if selected {
			bg = color.RGBA{46, 30, 74, 244}
			border = render.ColGold
		}
		vector.DrawFilledRect(dst, rect.X, rect.Y, rect.W, rect.H, bg, false)
		vector.StrokeRect(dst, rect.X, rect.Y, rect.W, rect.H, 1.4, border, false)

		if i == 8 {
			tw := render.TextPixelWidth(lbl, 1.05)
			r.DrawScaledText(dst, lbl, int(rect.X+(rect.W-float32(tw))*0.5), int(rect.Y+4), 1.05, render.ColGold, render.ColOutline)
			continue
		}

		r.DrawTextColored(dst, lbl, int(rect.X+12), int(rect.Y+4), color.White)

		switch i {
		case 0:
			val := filterNames[save.Settings.FilterMode%3]
			render.DrawPillBadge(dst, rect.X+196, rect.Y+3, 192, 18, color.RGBA{16, 12, 30, 238}, render.ColGyaru)
			vw := render.TextPixelWidth(val, 1.0)
			r.DrawTextColored(dst, val, int(rect.X+196+(192-float32(vw))*0.5), int(rect.Y+4), render.ColGyaru)
		case 1, 2, 3:
			v := save.Settings.MasterVol
			if i == 2 {
				v = save.Settings.MusicVol
			} else if i == 3 {
				v = save.Settings.SFXVol
			}
			barX := rect.X + 206
			barY := rect.Y + 7
			barW := float32(144)
			vector.DrawFilledRect(dst, barX, barY, barW, 10, color.RGBA{18, 22, 38, 235}, false)
			vector.DrawFilledRect(dst, barX, barY, barW*float32(v), 10, render.ColGreenNeon, false)
			// Crisp slider knob
			knobX := barX + barW*float32(v)
			vector.DrawFilledRect(dst, knobX-2, barY-2, 4, 14, render.ColGold, false)
			vector.StrokeRect(dst, barX, barY, barW, 10, 1, render.ColGyaru, false)
			r.DrawTextColored(dst, fmt.Sprintf("%d%%", int(v*100+0.5)), int(barX+barW+8), int(rect.Y+4), render.ColGold)
		case 4:
			m.drawToggleBadge(dst, r, rect.X+284, rect.Y+3, save.Settings.Fullscreen)
		case 5:
			m.drawToggleBadge(dst, r, rect.X+284, rect.Y+3, save.Settings.ScreenShake)
		case 6:
			m.drawToggleBadge(dst, r, rect.X+284, rect.Y+3, save.Settings.ShowFPS)
		case 7:
			modeTxt := "TIENI PREMUTO"
			modeCol := render.ColGyaru
			if save.Settings.ChronosToggle {
				modeTxt = "TOGGLE (ON/OFF)"
				modeCol = render.ColGold
			}
			render.DrawPillBadge(dst, rect.X+214, rect.Y+3, 156, 18, color.RGBA{16, 12, 30, 238}, modeCol)
			mw := render.TextPixelWidth(modeTxt, 1.0)
			r.DrawTextColored(dst, modeTxt, int(rect.X+214+(156-float32(mw))*0.5), int(rect.Y+4), modeCol)
		}
	}

	footer := "CLICK / TRASCINA CON IL MOUSE | ESC PER TORNARE"
	r.DrawTextColored(dst, footer, 320-render.TextPixelWidth(footer, 1.0)/2, 326, render.ColGyaru)
}

func (m *Manager) drawToggleBadge(dst *ebiten.Image, r *render.Renderer, x, y float32, on bool) {
	txt := "OFF"
	col := render.ColRedAlert
	if on {
		txt = "ON"
		col = render.ColGreenNeon
	}
	render.DrawPillBadge(dst, x, y, 56, 18, color.RGBA{16, 12, 28, 235}, col)
	tw := render.TextPixelWidth(txt, 1.0)
	r.DrawTextColored(dst, txt, int(x+(56-float32(tw))*0.5), int(y+1), col)
}

func (m *Manager) drawControls(dst *ebiten.Image, r *render.Renderer, mx, my float32) {
	vector.DrawFilledRect(dst, 0, 0, game.ScreenWidth, game.ScreenHeight, render.Premul(10, 6, 20, 210), false)
	vector.DrawFilledRect(dst, 30, 20, 580, 326, render.ColHudBg, false)
	vector.StrokeRect(dst, 30, 20, 580, 326, 2.0, render.ColGyaru, false)

	ctrlTitle := "COMANDI & REGOLE ARCADE"
	ctrlX := 320 - render.TextPixelWidth(ctrlTitle, 1.4)/2
	r.DrawScaledText(dst, ctrlTitle, ctrlX, 28, 1.4, render.ColGold, render.ColOutline)

	type ctrlItem struct {
		key  string
		desc string
		col  color.RGBA
	}
	items := [8]ctrlItem{
		{"WASD", "Movimento 360 gradi", render.ColGyaru},
		{"SPAZIO / DX", "Dash & Dash-Parry", render.ColGyaru},
		{"SHIFT / Q", "Chronos (Hold o Toggle)", render.ColStudent},
		{"CLICK SX / F", "Mira, Lancia & 1-Tasto", render.ColGold},
		{"ROTELLA / 1-4", "Cambia Fermata", render.ColEmo},
		{"CLICK SX", "Minigiochi & Combo (Mouse)", render.ColKawaii},
		{"R / E", "Richiamo (-50%CD) / Shop", render.ColGreenNeon},
		{"ESC / P", "Pausa / Chiudi Menu", color.RGBA{200, 200, 225, 255}},
	}

	for i, it := range items {
		colIdx := i / 4
		rowIdx := i % 4
		x := float32(42 + colIdx*282)
		y := float32(66 + rowIdx*44)

		r.DrawKeycap(dst, it.key, x, y, 92, 26, color.RGBA{38, 28, 62, 240}, it.col, false)
		r.DrawTextColored(dst, it.desc, int(x+98), int(y+5), color.White)
	}

	// Quick gameplay guide box
	vector.DrawFilledRect(dst, 50, 244, 540, 58, color.RGBA{22, 14, 38, 230}, false)
	vector.StrokeRect(dst, 50, 244, 540, 58, 1.2, render.ColKawaii, false)
	r.DrawTextColored(dst, "LOOP: 1) Lancia vicino (<=24px = PERFECT DROP). Solo Mouse SX per Minigiochi!", 58, 250, render.ColGold)
	r.DrawTextColored(dst, "2) Graze (<=18px) e Dash-Parry tagliano il cooldown; [R] richiama (-50%CD).", 58, 266, color.White)
	r.DrawTextColored(dst, "3) ZTL/VIP pagano 3x (+Polizia); Streak x5 = Rush Hour & scia Bus turbo!", 58, 282, render.ColGyaru)

	backRect := game.Rect{X: 248, Y: 312, W: 144, H: 28}
	m.drawButton(dst, r, backRect, "INDIETRO [ESC]", backRect.Contains(mx, my), false)
}

func (m *Manager) drawChallenges(dst *ebiten.Image, r *render.Renderer, save *meta.SaveData, mx, my float32) {
	vector.DrawFilledRect(dst, 0, 0, game.ScreenWidth, game.ScreenHeight, render.Premul(10, 6, 20, 210), false)
	vector.DrawFilledRect(dst, 20, 16, 600, 328, render.ColHudBg, false)
	vector.StrokeRect(dst, 20, 16, 600, 328, 2.0, render.ColGold, false)

	chTitle := "SFIDE & SKIN SBLOCCABILI"
	chTitleX := 320 - render.TextPixelWidth(chTitle, 1.35)/2
	r.DrawScaledText(dst, chTitle, chTitleX, 24, 1.35, render.ColGold, render.ColOutline)

	all := meta.AllChallenges()
	start := m.ChallengePage * 5
	end := start + 5
	if end > len(all) {
		end = len(all)
	}

	for idx := start; idx < end; idx++ {
		ch := all[idx]
		row := idx - start
		ry := float32(56 + row*38)
		unlocked := save.Profile.UnlockedChallenges[ch.ID]
		border := color.RGBA{75, 70, 100, 200}
		titleCol := color.RGBA{180, 180, 205, 230}
		badge := "[ ]"
		if unlocked {
			border = render.ColGreenNeon
			titleCol = render.ColGold
			badge = "[OK]"
		}

		vector.DrawFilledRect(dst, 34, ry, 356, 33, color.RGBA{22, 16, 36, 230}, false)
		vector.StrokeRect(dst, 34, ry, 356, 33, 1.2, border, false)
		r.DrawTextColored(dst, badge+" "+ch.TitleIT+" -> "+ch.RewardIT, 42, int(ry+2), titleCol)
		r.DrawTextColored(dst, ch.DescIT, 42, int(ry+16), render.Premul(215, 220, 240, 225))
	}

	prevBtn := game.Rect{X: 34, Y: 256, W: 90, H: 22}
	nextBtn := game.Rect{X: 132, Y: 256, W: 90, H: 22}
	m.drawButton(dst, r, prevBtn, "PAG. 1", m.ChallengePage == 0 || prevBtn.Contains(mx, my), false)
	m.drawButton(dst, r, nextBtn, "PAG. 2", m.ChallengePage == 1 || nextBtn.Contains(mx, my), false)
	r.DrawTextColored(dst, fmt.Sprintf("PAG. %d / 2", m.ChallengePage+1), 234, 260, render.ColGold)

	// Right Panel: Skin Selector
	vector.DrawFilledRect(dst, 398, 56, 208, 222, color.RGBA{22, 14, 38, 230}, false)
	vector.StrokeRect(dst, 398, 56, 208, 222, 1.5, render.ColGyaru, false)
	r.DrawScaledText(dst, "PERSONALIZZAZIONE", 422, 64, 1.1, render.ColGyaru, render.ColOutline)

	// Live miniature sprite preview strip for currently selected player + bus skins
	vector.DrawFilledRect(dst, 404, 82, 196, 28, color.RGBA{14, 10, 28, 230}, false)
	vector.StrokeRect(dst, 404, 82, 196, 28, 1.0, render.PremulColor(render.ColGyaru, 120), false)
	r.DrawTextColored(dst, "PREVIEW:", 410, 89, color.RGBA{185, 190, 215, 230})

	op := &ebiten.DrawImageOptions{}
	alfImg := r.Sprites().Get(assets.SpriteAlfredoIdle0)
	aw, ah := float64(alfImg.Bounds().Dx()), float64(alfImg.Bounds().Dy())
	aScale := 22.0 / ah
	op.GeoM.Reset()
	op.GeoM.Translate(-aw*0.5, -ah*0.5)
	op.GeoM.Scale(aScale, aScale)
	op.GeoM.Translate(484, 96)
	op.ColorScale.Reset()
	switch save.Profile.SelectedPlayerSkin {
	case 1:
		op.ColorScale.Scale(1.12, 0.92, 1.24, 1.0)
	case 2:
		op.ColorScale.Scale(0.95, 0.90, 0.85, 1.0)
		vector.StrokeCircle(dst, 484, 96, 11, 1.0, render.Premul(255, 220, 70, 140), true)
	}
	dst.DrawImage(alfImg, op)

	vector.DrawFilledRect(dst, 512, 86, 1, 20, render.Premul(255, 255, 255, 55), false)

	busImg := r.Sprites().Get(assets.SpriteBusNeon)
	bw, bh := float64(busImg.Bounds().Dx()), float64(busImg.Bounds().Dy())
	bScale := 20.0 / bh
	if bw*bScale > 64.0 {
		bScale = 64.0 / bw
	}
	op.GeoM.Reset()
	op.GeoM.Translate(-bw*0.5, -bh*0.5)
	op.GeoM.Scale(bScale, bScale)
	op.GeoM.Translate(556, 96)
	op.ColorScale.Reset()
	if save.Profile.SelectedBusSkin == 1 {
		op.ColorScale.Scale(1.30, 1.15, 0.55, 1.0)
	}
	dst.DrawImage(busImg, op)

	r.DrawTextColored(dst, "SKIN ALFREDO (CLICCA):", 408, 114, render.ColGold)
	skinPBtn := game.Rect{X: 404, Y: 132, W: 196, H: 30}
	m.drawButton(dst, r, skinPBtn, save.Profile.PlayerSkinName(save.Profile.SelectedPlayerSkin), skinPBtn.Contains(mx, my), false)

	r.DrawTextColored(dst, "SKIN AUTOBUS (CLICCA):", 408, 186, render.ColGold)
	skinBBtn := game.Rect{X: 404, Y: 204, W: 196, H: 30}
	m.drawButton(dst, r, skinBBtn, save.Profile.BusSkinName(save.Profile.SelectedBusSkin), skinBBtn.Contains(mx, my), false)

	bossCount := 0
	for b := 0; b < 6; b++ {
		if save.Profile.IsBossDefeated(b) {
			bossCount++
		}
	}
	r.DrawTextColored(dst, fmt.Sprintf("BOSS BATTUTI: %d / 6", bossCount), 408, 248, render.ColKawaii)
	for b := 0; b < 6; b++ {
		px := float32(408 + b*31)
		pipCol := color.RGBA{46, 36, 66, 220}
		if save.Profile.IsBossDefeated(b) {
			pipCol = render.ColGreenNeon
		}
		vector.DrawFilledRect(dst, px, 265, 25, 5, pipCol, false)
	}

	r.DrawTextColored(dst, "CLICCA CON IL MOUSE PER CAMBIARE PAGINA O SKIN | ESC INDIETRO", 34, 310, render.ColGyaru)

	backBtn := game.Rect{X: 432, Y: 302, W: 140, H: 28}
	m.drawButton(dst, r, backBtn, "INDIETRO [ESC]", backBtn.Contains(mx, my), false)
}

func (m *Manager) drawEndRun(dst *ebiten.Image, r *render.Renderer, save *meta.SaveData, victory bool, mx, my float32) {
	vector.DrawFilledRect(dst, 0, 0, game.ScreenWidth, game.ScreenHeight, render.Premul(14, 6, 24, 200), false)
	for y := 0; y < game.ScreenHeight; y += 6 {
		vector.DrawFilledRect(dst, 0, float32(y), game.ScreenWidth, 2, render.Premul(0, 245, 212, 20), false)
	}

	boxX, boxY, boxW, boxH := float32(104), float32(42), float32(432), float32(276)
	border := render.ColKawaii
	title := "<< VHS REWIND // GAME OVER >>"
	if victory {
		border = render.ColGold
		title = "VITTORIA! CORSA COMPLETATA!"
		if m.LastWorldStats.BossRush {
			title = "VITTORIA!! BOSS RUSH COMPLETATA!"
		}
	}
	vector.DrawFilledRect(dst, boxX, boxY, boxW, boxH, render.ColHudBg, false)
	vector.StrokeRect(dst, boxX, boxY, boxW, boxH, 2.2, border, false)

	titleW := render.TextPixelWidth(title, 1.3)
	r.DrawScaledText(dst, title, int(boxX+(boxW-float32(titleW))*0.5), int(boxY+14), 1.3, border, render.ColOutline)

	if m.NewRecord {
		render.DrawPillBadge(dst, boxX+156, boxY+40, 120, 16, color.RGBA{56, 40, 8, 240}, render.ColGold)
		r.DrawTextColored(dst, "NUOVO RECORD!", int(boxX+176), int(boxY+40), render.ColGold)
	}

	mins := int(m.LastWorldStats.RunTime) / 60
	secs := int(m.LastWorldStats.RunTime) % 60
	r.DrawScaledText(dst, fmt.Sprintf("PUNTEGGIO: %d   |   RECORD: %d", m.LastWorldStats.Score, save.Profile.BestScore), int(boxX+34), int(boxY+68), 1.0, color.White, render.ColOutline)
	r.DrawScaledText(dst, fmt.Sprintf("TEMPO: %02d:%02d   |   TIER: %d   |   RAGAZZE: %d   |   BOSS: %d",
		mins, secs, m.LastWorldStats.Tier, m.LastWorldStats.GirlsBoarded, m.LastWorldStats.Bosses), int(boxX+34), int(boxY+92), 1.0, render.ColGyaru, render.ColOutline)

	if len(m.RunUnlocks) > 0 {
		r.DrawTextColored(dst, "SFIDE SBLOCCATE IN QUESTA RUN:", int(boxX+34), int(boxY+124), render.ColGold)
		for i := 0; i < len(m.RunUnlocks) && i < 3; i++ {
			ch := m.RunUnlocks[i]
			r.DrawTextColored(dst, "* "+ch.TitleIT+" ("+ch.RewardIT+")", int(boxX+38), int(boxY)+142+i*16, render.ColGreenNeon)
		}
	} else {
		r.DrawTextColored(dst, "Suggerimento: raccogli le Cure Espresso in strada e osa in ZTL!", int(boxX+34), int(boxY+140), color.RGBA{185, 195, 220, 220})
	}

	if victory {
		labels := [3]string{"CONTINUA OVERDRIVE", "NUOVA RUN", "MENU PRINCIPALE"}
		for i, lbl := range labels {
			rect := endRunButtonRect(i, 3)
			selected := m.SelectedIdx == i || rect.Contains(mx, my)
			m.drawButton(dst, r, rect, lbl, selected, false)
		}
	} else {
		labels := [2]string{"RIGIOCA [R]", "MENU PRINCIPALE"}
		for i, lbl := range labels {
			rect := endRunButtonRect(i, 2)
			selected := m.SelectedIdx == i || rect.Contains(mx, my)
			m.drawButton(dst, r, rect, lbl, selected, false)
		}
	}

	cue := "CLICK SX PER SCEGLIERE"
	r.DrawTextColored(dst, cue, 320-render.TextPixelWidth(cue, 1.0)/2, int(boxY+boxH-19), render.ColGyaru)
}

func (m *Manager) drawButton(dst *ebiten.Image, r *render.Renderer, rect game.Rect, label string, selected, locked bool) {
	bg := color.RGBA{26, 18, 44, 235}
	border := render.ColGyaru
	var txtCol color.Color = color.White

	if locked {
		bg = color.RGBA{22, 18, 30, 210}
		border = color.RGBA{90, 85, 105, 180}
		txtCol = color.RGBA{135, 130, 150, 210}
	} else if selected {
		bg = color.RGBA{58, 30, 86, 248}
		border = render.ColGold
		txtCol = render.ColGold
	}

	vector.DrawFilledRect(dst, rect.X, rect.Y, rect.W, rect.H, bg, false)
	vector.StrokeRect(dst, rect.X, rect.Y, rect.W, rect.H, 1.6, border, false)

	tw := render.TextPixelWidth(label, 1.0)
	tx := int(rect.X + (rect.W-float32(tw))*0.5)
	ty := int(rect.Y + (rect.H-14)*0.5)
	r.DrawScaledText(dst, label, tx, ty, 1.0, txtCol, render.ColOutline)
}

func (m *Manager) drawAccentButton(dst *ebiten.Image, r *render.Renderer, rect game.Rect, label string, selected bool, accent, baseBg color.RGBA) {
	bg := baseBg
	border := accent
	var txtCol color.Color = accent

	if selected {
		bg = color.RGBA{24, 72, 82, 250}
		border = render.ColGold
		txtCol = render.ColGold
	}

	vector.DrawFilledRect(dst, rect.X, rect.Y, rect.W, rect.H, bg, false)
	vector.StrokeRect(dst, rect.X, rect.Y, rect.W, rect.H, 1.8, border, false)
	vector.DrawFilledRect(dst, rect.X+2, rect.Y+2, 3, rect.H-4, border, false)

	tw := render.TextPixelWidth(label, 1.0)
	tx := int(rect.X + (rect.W-float32(tw))*0.5)
	ty := int(rect.Y + (rect.H-14)*0.5)
	r.DrawScaledText(dst, label, tx, ty, 1.0, txtCol, render.ColOutline)
}

func titleButtonRect(i int) game.Rect {
	return game.Rect{
		X: 202,
		Y: 96 + float32(i)*34,
		W: 236,
		H: 28,
	}
}

func pauseButtonRect(i int) game.Rect {
	return game.Rect{
		X: 72,
		Y: 96 + float32(i)*39,
		W: 198,
		H: 30,
	}
}

func tutorialTabRect(t int) game.Rect {
	return game.Rect{
		X: 38 + float32(t)*142,
		Y: 50,
		W: 136,
		H: 24,
	}
}

func tutorialPrevRect() game.Rect {
	return game.Rect{X: 38, Y: 300, W: 92, H: 28}
}

func tutorialNextRect() game.Rect {
	return game.Rect{X: 138, Y: 300, W: 92, H: 28}
}

func tutorialPracticeRect() game.Rect {
	return game.Rect{X: 246, Y: 300, W: 212, H: 28}
}

func tutorialBackRect() game.Rect {
	return game.Rect{X: 474, Y: 300, W: 128, H: 28}
}

func optionsRowRect(i int) game.Rect {
	return game.Rect{
		X: 116,
		Y: 54 + float32(i)*29,
		W: 408,
		H: 24,
	}
}

func endRunButtonRect(i, total int) game.Rect {
	if total == 3 {
		return game.Rect{
			X: 124 + float32(i)*134,
			Y: 254,
			W: 124,
			H: 32,
		}
	}
	return game.Rect{
		X: 158 + float32(i)*172,
		Y: 254,
		W: 156,
		H: 32,
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
