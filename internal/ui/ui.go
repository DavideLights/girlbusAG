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

// Manager handles all game menus, options, controls guide, challenges/skins, and end-of-run overlays.
type Manager struct {
	Scene          Scene
	ReturnScene    Scene
	SelectedIdx    int
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

// Update processes menu navigation, sliders, toggles, and mouse clicks.
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
				if m.SelectedIdx != i {
					m.SelectedIdx = i
				}
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
				if save.Profile.IsBossRushUnlocked() {
					res.Action = ActionStartBossRush
				}
			case 2:
				m.OpenScene(SceneChallenges, SceneTitle)
			case 3:
				m.OpenScene(SceneControls, SceneTitle)
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
				m.OpenScene(SceneControls, ScenePaused)
			case 3:
				m.OpenScene(SceneOptions, ScenePaused)
			case 4:
				res.Action = ActionToTitle
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
					// Slider bar on right half of row
					barX := rect.X + 210
					barW := float32(150)
					if mi.MouseX >= barX-8 && mi.MouseX <= barX+barW+8 {
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
		backBtn := game.Rect{X: 432, Y: 306, W: 140, H: 28}

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

// Draw renders the active UI overlay onto the 640x360 scene buffer.
func (m *Manager) Draw(dst *ebiten.Image, r *render.Renderer, save *meta.SaveData, w *game.World, mx, my float32) {
	switch m.Scene {
	case SceneTitle:
		m.drawTitle(dst, r, save, mx, my)
	case ScenePaused:
		m.drawPaused(dst, r, w, mx, my)
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
	vector.DrawFilledRect(dst, 0, 0, game.ScreenWidth, game.ScreenHeight, render.Premul(12, 8, 24, 145), false)

	// Logo header box
	vector.DrawFilledRect(dst, 118, 22, 404, 62, render.ColHudBg, false)
	vector.StrokeRect(dst, 118, 22, 404, 62, 2.0, render.ColKawaii, false)
	vector.StrokeRect(dst, 121, 25, 398, 56, 1.0, render.ColGyaru, false)

	titleTxt := "GIRLBUS: NEON STOP"
	titleX := 320 - render.TextPixelWidth(titleTxt, 1.8)/2
	r.DrawScaledText(dst, titleTxt, titleX, 31, 1.8, render.ColGold, render.ColOutline)

	subTxt := "POP-URBAN ROGUE-LITE BULLET-HELL // v2.0"
	subX := 320 - render.TextPixelWidth(subTxt, 1.0)/2
	r.DrawTextColored(dst, subTxt, subX, 63, render.ColGyaru)

	labels := []string{
		"NUOVA RUN",
		"BOSS RUSH (6 BOSS)",
		"SFIDE & SKIN",
		"COMANDI",
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
		locked := i == 1 && bossRushLocked
		display := lbl
		if locked {
			display = "BOSS RUSH [BLOCCATA]"
		}
		m.drawButton(dst, r, rect, display, selected, locked)
	}

	// Bottom record & unlocks bar
	unlocked := 0
	for _, ch := range meta.AllChallenges() {
		if save.Profile.UnlockedChallenges[ch.ID] {
			unlocked++
		}
	}
	vector.DrawFilledRect(dst, 0, game.ScreenHeight-24, game.ScreenWidth, 24, render.ColHudBg, false)
	vector.DrawFilledRect(dst, 0, game.ScreenHeight-24, game.ScreenWidth, 1.2, render.ColGyaru, false)
	stats := fmt.Sprintf("RECORD: %d   |   RUN: %d   |   SFIDE: %d/10   |   SKIN: %s",
		save.Profile.BestScore, save.Profile.TotalRuns, unlocked, save.Profile.PlayerSkinName(save.Profile.SelectedPlayerSkin))
	r.DrawTextColored(dst, stats, 24, game.ScreenHeight-20, render.ColGold)
}

func (m *Manager) drawPaused(dst *ebiten.Image, r *render.Renderer, w *game.World, mx, my float32) {
	vector.DrawFilledRect(dst, 0, 0, game.ScreenWidth, game.ScreenHeight, color.RGBA{10, 6, 20, 195}, false)

	// Left Menu Panel
	vector.DrawFilledRect(dst, 46, 48, 250, 260, render.ColHudBg, false)
	vector.StrokeRect(dst, 46, 48, 250, 260, 2.0, render.ColGyaru, false)
	r.DrawScaledText(dst, "PAUSA", 142, 62, 1.6, render.ColGold, render.ColOutline)

	labels := [5]string{"RIPRENDI", "RIAVVIA RUN", "COMANDI", "OPZIONI", "MENU PRINCIPALE"}
	for i, lbl := range labels {
		rect := pauseButtonRect(i)
		selected := m.SelectedIdx == i || rect.Contains(mx, my)
		m.drawButton(dst, r, rect, lbl, selected, false)
	}

	// Right Run Status & Relics Panel
	vector.DrawFilledRect(dst, 308, 48, 294, 260, render.ColHudBg, false)
	vector.StrokeRect(dst, 308, 48, 294, 260, 1.8, render.ColKawaii, false)
	r.DrawScaledText(dst, "STATO RUN & RELIQUIE", 320, 60, 1.15, render.ColKawaii, render.ColOutline)

	if w != nil {
		mins := int(w.RunTime) / 60
		secs := int(w.RunTime) % 60
		r.DrawTextColored(dst, fmt.Sprintf("Tempo: %02d:%02d  |  Tier %d  |  Score: %d", mins, secs, w.DifficultyTier, w.Score), 320, 86, color.White)
		r.DrawTextColored(dst, fmt.Sprintf("Ragazze: %d  |  Boss: %d  |  Cooldown: %.1fs", w.GirlsBoarded, w.BossesDefeated, w.Player.EffectiveStopCooldown()), 320, 104, render.ColGyaru)

		r.DrawTextColored(dst, "RELIQUIE ATTIVE:", 320, 130, render.ColGold)
		ry := 150
		anyRelic := false
		for rk := 0; rk < int(game.RelicCount); rk++ {
			if w.Player.Relics[rk] {
				anyRelic = true
				spec := game.GetRelicSpec(game.RelicKind(rk))
				r.DrawTextColored(dst, "* "+spec.Name, 320, ry, render.ColGreenNeon)
				r.DrawTextColored(dst, " "+spec.Desc, 320, ry+12, color.RGBA{200, 210, 230, 220})
				ry += 26
			}
		}
		if !anyRelic {
			r.DrawTextColored(dst, "Nessuna reliquia ancora (Negozio $120)", 320, 152, color.RGBA{165, 165, 190, 220})
		}
	}
}

func (m *Manager) drawOptions(dst *ebiten.Image, r *render.Renderer, save *meta.SaveData, mx, my float32) {
	vector.DrawFilledRect(dst, 0, 0, game.ScreenWidth, game.ScreenHeight, render.Premul(10, 6, 20, 205), false)
	vector.DrawFilledRect(dst, 100, 22, 440, 318, render.ColHudBg, false)
	vector.StrokeRect(dst, 100, 22, 440, 318, 2.0, render.ColGyaru, false)

	r.DrawScaledText(dst, "OPZIONI", 276, 28, 1.4, render.ColGold, render.ColOutline)

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
		bg := color.RGBA{24, 16, 40, 220}
		border := color.RGBA{80, 90, 130, 200}
		if selected {
			bg = color.RGBA{46, 30, 74, 240}
			border = render.ColGold
		}
		vector.DrawFilledRect(dst, rect.X, rect.Y, rect.W, rect.H, bg, false)
		vector.StrokeRect(dst, rect.X, rect.Y, rect.W, rect.H, 1.4, border, false)

		if i == 8 {
			r.DrawScaledText(dst, lbl, int(rect.X+130), int(rect.Y+4), 1.05, render.ColGold, render.ColOutline)
			continue
		}

		r.DrawTextColored(dst, lbl, int(rect.X+12), int(rect.Y+4), color.White)

		switch i {
		case 0:
			val := "< " + filterNames[save.Settings.FilterMode%3] + " >"
			r.DrawTextColored(dst, val, int(rect.X+200), int(rect.Y+4), render.ColGyaru)
		case 1, 2, 3:
			v := save.Settings.MasterVol
			if i == 2 {
				v = save.Settings.MusicVol
			} else if i == 3 {
				v = save.Settings.SFXVol
			}
			barX := rect.X + 210
			barY := rect.Y + 7
			barW := float32(140)
			vector.DrawFilledRect(dst, barX, barY, barW, 10, color.RGBA{18, 22, 38, 230}, false)
			vector.DrawFilledRect(dst, barX, barY, barW*float32(v), 10, render.ColGreenNeon, false)
			vector.StrokeRect(dst, barX, barY, barW, 10, 1, render.ColGyaru, false)
			r.DrawTextColored(dst, fmt.Sprintf("%d%%", int(v*100+0.5)), int(barX+barW+8), int(rect.Y+4), render.ColGold)
		case 4:
			m.drawToggleBadge(dst, r, rect.X+280, rect.Y+3, save.Settings.Fullscreen)
		case 5:
			m.drawToggleBadge(dst, r, rect.X+280, rect.Y+3, save.Settings.ScreenShake)
		case 6:
			m.drawToggleBadge(dst, r, rect.X+280, rect.Y+3, save.Settings.ShowFPS)
		case 7:
			modeTxt := "< TIENI PREMUTO >"
			modeCol := render.ColGyaru
			if save.Settings.ChronosToggle {
				modeTxt = "< TOGGLE (ON/OFF) >"
				modeCol = render.ColGold
			}
			r.DrawTextColored(dst, modeTxt, int(rect.X+210), int(rect.Y+4), modeCol)
		}
	}
}

func (m *Manager) drawToggleBadge(dst *ebiten.Image, r *render.Renderer, x, y float32, on bool) {
	txt := "OFF"
	col := render.ColRedAlert
	if on {
		txt = "ON"
		col = render.ColGreenNeon
	}
	render.DrawPillBadge(dst, x, y, 56, 18, color.RGBA{16, 12, 28, 235}, col)
	r.DrawTextColored(dst, txt, int(x+18), int(y+1), col)
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
		{"LMB / FRECCE", "Minigiochi & Combo", render.ColKawaii},
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
	r.DrawTextColored(dst, "LOOP: 1) Lancia vicino (<=24px = PERFECT DROP). Mouse LMB fa Minigiochi e Combo!", 58, 250, render.ColGold)
	r.DrawTextColored(dst, "2) Graze (<=18px) e Dash-Parry tagliano il cooldown; [R] richiama fermate (-50%CD).", 58, 266, color.White)
	r.DrawTextColored(dst, "3) ZTL/VIP pagano 3x (+Polizia); Streak x5 = Rush Hour & scia Bus per ricaricare!", 58, 282, render.ColGyaru)

	backRect := game.Rect{X: 248, Y: 312, W: 144, H: 28}
	m.drawButton(dst, r, backRect, "INDIETRO [ESC]", backRect.Contains(mx, my), false)
}

func (m *Manager) drawChallenges(dst *ebiten.Image, r *render.Renderer, save *meta.SaveData, mx, my float32) {
	vector.DrawFilledRect(dst, 0, 0, game.ScreenWidth, game.ScreenHeight, render.Premul(10, 6, 20, 210), false)
	vector.DrawFilledRect(dst, 20, 18, 600, 324, render.ColHudBg, false)
	vector.StrokeRect(dst, 20, 18, 600, 324, 2.0, render.ColGold, false)

	chTitle := "SFIDE & SKIN SBLOCCABILI"
	chTitleX := 320 - render.TextPixelWidth(chTitle, 1.35)/2
	r.DrawScaledText(dst, chTitle, chTitleX, 26, 1.35, render.ColGold, render.ColOutline)

	all := meta.AllChallenges()
	start := m.ChallengePage * 5
	end := start + 5
	if end > len(all) {
		end = len(all)
	}

	for idx := start; idx < end; idx++ {
		ch := all[idx]
		row := idx - start
		ry := float32(58 + row*38)
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
	vector.DrawFilledRect(dst, 398, 58, 208, 220, color.RGBA{22, 14, 38, 230}, false)
	vector.StrokeRect(dst, 398, 58, 208, 220, 1.5, render.ColGyaru, false)
	r.DrawScaledText(dst, "PERSONALIZZAZIONE", 422, 66, 1.1, render.ColGyaru, render.ColOutline)

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
	m.drawButton(dst, r, skinPBtn, "< "+save.Profile.PlayerSkinName(save.Profile.SelectedPlayerSkin)+" >", skinPBtn.Contains(mx, my), false)

	r.DrawTextColored(dst, "SKIN AUTOBUS (CLICCA):", 408, 186, render.ColGold)
	skinBBtn := game.Rect{X: 404, Y: 204, W: 196, H: 30}
	m.drawButton(dst, r, skinBBtn, "< "+save.Profile.BusSkinName(save.Profile.SelectedBusSkin)+" >", skinBBtn.Contains(mx, my), false)

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

	backBtn := game.Rect{X: 432, Y: 306, W: 140, H: 28}
	m.drawButton(dst, r, backBtn, "INDIETRO [ESC]", backBtn.Contains(mx, my), false)
}

func (m *Manager) drawEndRun(dst *ebiten.Image, r *render.Renderer, save *meta.SaveData, victory bool, mx, my float32) {
	vector.DrawFilledRect(dst, 0, 0, game.ScreenWidth, game.ScreenHeight, render.Premul(14, 6, 24, 200), false)
	for y := 0; y < game.ScreenHeight; y += 6 {
		vector.DrawFilledRect(dst, 0, float32(y), game.ScreenWidth, 2, render.Premul(0, 245, 212, 22), false)
	}

	boxX, boxY, boxW, boxH := float32(104), float32(46), float32(432), float32(268)
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

	r.DrawScaledText(dst, title, int(boxX+52), int(boxY+14), 1.35, border, render.ColOutline)

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
		r.DrawTextColored(dst, "Suggerimento: osa nelle zone ZTL per moltiplicatori fino a 3.5x!", int(boxX+34), int(boxY+140), color.RGBA{185, 195, 220, 220})
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

func titleButtonRect(i int) game.Rect {
	return game.Rect{
		X: 210,
		Y: 102 + float32(i)*36,
		W: 220,
		H: 29,
	}
}

func pauseButtonRect(i int) game.Rect {
	return game.Rect{
		X: 72,
		Y: 100 + float32(i)*38,
		W: 198,
		H: 30,
	}
}

func optionsRowRect(i int) game.Rect {
	return game.Rect{
		X: 120,
		Y: 58 + float32(i)*29,
		W: 400,
		H: 24,
	}
}

func endRunButtonRect(i, total int) game.Rect {
	if total == 3 {
		return game.Rect{
			X: 124 + float32(i)*134,
			Y: 256,
			W: 124,
			H: 32,
		}
	}
	return game.Rect{
		X: 158 + float32(i)*172,
		Y: 256,
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
