package ui

import (
	"testing"

	"girlbusAG/internal/game"
	"girlbusAG/internal/input"
	"girlbusAG/internal/meta"
)

func TestStationaryPointerDoesNotTakeKeyboardFocus(t *testing.T) {
	m := NewManager()
	save := meta.DefaultSaveData()
	pointer := input.MenuInput{MouseX: 40, MouseY: titleButtonRect(3).Y + 4}
	m.Update(pointer, &save)
	if m.SelectedIdx != 3 {
		t.Fatalf("moving the pointer selected %d; want challenges", m.SelectedIdx)
	}
	pointer.Down = true
	m.Update(pointer, &save)
	if m.SelectedIdx != 4 {
		t.Fatalf("keyboard focus selected %d; want options", m.SelectedIdx)
	}
	pointer.Down, pointer.Confirm = false, true
	m.Update(pointer, &save)
	if m.Scene != SceneOptions {
		t.Fatalf("confirm opened scene %d; want options", m.Scene)
	}
}

func TestOptionsVolumeDragStaysOnOriginalSlider(t *testing.T) {
	m := NewManager()
	m.OpenScene(SceneOptions, SceneTitle)
	save := meta.DefaultSaveData()
	oldMusic := save.Settings.MusicVol
	row := optionsRowRect(1)
	barX, barW := optionsSlider(row)
	press := input.MenuInput{MouseX: barX + barW/2, MouseY: row.Y + 5, MouseLeftJustPressed: true, MouseLeftPressed: true}
	if result := m.Update(press, &save); result.Action != ActionSettingsChanged || save.Settings.MasterVol != 0.5 {
		t.Fatalf("click returned %+v and volume %f; want settings change to 0.5", result, save.Settings.MasterVol)
	}

	press.MouseX, press.MouseY = barX+barW+18, optionsRowRect(2).Y+5
	press.MouseLeftJustPressed = false
	m.Update(press, &save)
	if save.Settings.MasterVol != 1 || save.Settings.MusicVol != oldMusic {
		t.Fatalf("drag changed master to %f and music to %f; want 1 and %f", save.Settings.MasterVol, save.Settings.MusicVol, oldMusic)
	}
	if result := m.Update(press, &save); result.Action != ActionNone {
		t.Fatalf("stationary slider returned %+v; want no action", result)
	}
}

func TestGuidePracticeIsAccessibleByKeyboard(t *testing.T) {
	m := NewManager()
	save := meta.DefaultSaveData()
	m.OpenScene(SceneTutorial, SceneTitle)
	m.Update(input.MenuInput{Right: true}, &save)
	if m.TutorialTab != 1 {
		t.Fatalf("right moved to guide tab %d; want 1", m.TutorialTab)
	}
	m.Update(input.MenuInput{Down: true}, &save)
	if result := m.Update(input.MenuInput{Confirm: true}, &save); result.Action != ActionStartTutorial {
		t.Fatalf("confirm returned %+v; want start tutorial", result)
	}
}

func TestTitleMapPickerOnlyStartsPlayableMap(t *testing.T) {
	m := NewManager()
	save := meta.DefaultSaveData()
	save.Profile.UnlockedChallenges["all_6_bosses"] = true
	confirm := input.MenuInput{Confirm: true}
	if res := m.Update(confirm, &save); res.Action != ActionStartNormal {
		t.Fatalf("initial map returned %+v; want start normal", res)
	}
	for _, want := range []int{titleMapBeach, titleMapComingSoon} {
		res := m.Update(input.MenuInput{Right: true, Confirm: true}, &save)
		if m.mapIndex != want || res.Action != ActionNone {
			t.Fatalf("map navigation selected %d and returned %+v; want %d without starting", m.mapIndex, res, want)
		}
		if res := m.Update(confirm, &save); res.Action != ActionNone {
			t.Fatalf("unavailable map %d returned %+v; want no action", want, res)
		}
		m.SelectedIdx = 2
		if res := m.Update(confirm, &save); res.Action != ActionNone {
			t.Fatalf("unavailable map %d returned %+v in Boss Rush", want, res)
		}
	}
	m.Update(input.MenuInput{Right: true}, &save)
	if m.mapIndex != titleMapCity || m.SelectedIdx != 0 {
		t.Fatalf("wrapped to map %d with menu selection %d; want city and play", m.mapIndex, m.SelectedIdx)
	}
	m.SelectedIdx = 2
	if res := m.Update(confirm, &save); res.Action != ActionStartBossRush {
		t.Fatalf("Boss Rush on the city returned %+v; want start Boss Rush", res)
	}
}

func TestTitleMapPickerMouseArrowsPagesAndWheel(t *testing.T) {
	m := NewManager()
	save := meta.DefaultSaveData()
	m.Update(input.MenuInput{MouseX: 120, MouseY: 165, MapScroll: 1}, &save)
	if m.mapIndex != titleMapCity {
		t.Fatalf("scrolling over the menu changed map to %d", m.mapIndex)
	}
	next := mapNextHit
	m.Update(input.MenuInput{MouseX: next.X + 5, MouseY: next.Y + 5, MouseLeftJustPressed: true}, &save)
	if m.mapIndex != titleMapBeach || m.mapSlide != mapSlideDuration {
		t.Fatalf("next arrow selected %d with slide %d; want beach and animated transition", m.mapIndex, m.mapSlide)
	}
	m.Update(input.MenuInput{MouseX: 418, MouseY: 126, MapScroll: 1}, &save)
	if m.mapIndex != titleMapComingSoon {
		t.Fatalf("wheel selected %d; want coming soon", m.mapIndex)
	}
	firstPage := titleMapPageHit(titleMapCity)
	m.Update(input.MenuInput{MouseX: firstPage.X + 5, MouseY: firstPage.Y + 5, MouseLeftJustPressed: true}, &save)
	if m.mapIndex != titleMapCity {
		t.Fatalf("first page marker selected %d; want city", m.mapIndex)
	}
}

func TestSetEndRunPopulatesBuildSummaryAndSeed(t *testing.T) {
	m := NewManager()
	w := game.NewWorldWithOptions(0xA1B2C3, true)
	w.Score = 12345
	w.RunTime = 187.5
	w.GirlsBoarded = 42
	w.BossesDefeated = 5
	w.DifficultyTier = 4
	w.Player.Level = 9
	w.MaxStreak = 7
	w.PeakWanted = 5
	w.PerfectDrops = 11
	w.EventsCompleted = 3
	w.BoardedByArch = [game.ArchCount]int{10, 12, 8, 12}
	w.Player.UpgradeRanks[game.UpgradeCooldown] = 2
	w.Player.UpgradeRanks[game.UpgradeDashKatana] = 3
	w.Player.UpgradeRanks[game.UpgradeRizzMagnet] = 1
	w.Player.Relics[game.RelicNoiseCancel] = true
	w.Player.Relics[game.RelicFakeVIPPass] = true

	unlocks := []meta.ChallengeDef{{ID: "first_ride", TitleIT: "Prima Corsa", RewardIT: "Skin Neon"}}
	m.SetEndRun(true, w, true, unlocks)

	if m.Scene != SceneVictory {
		t.Fatalf("scene = %d; want SceneVictory", m.Scene)
	}
	if !m.NewRecord || len(m.RunUnlocks) != 1 {
		t.Fatalf("newRecord=%v unlocks=%d; want true and 1", m.NewRecord, len(m.RunUnlocks))
	}
	got := m.LastWorldStats
	if got.Score != 12345 || got.RunTime != 187.5 || got.GirlsBoarded != 42 || got.Bosses != 5 || got.Tier != 4 || !got.BossRush {
		t.Fatalf("core stats mismatch: %+v", got)
	}
	if got.Seed != 0xA1B2C3 || game.FormatRunSeed(got.Seed) != "#A1B2C3" {
		t.Fatalf("seed=%d formatted=%q; want 0xA1B2C3 / #A1B2C3", got.Seed, game.FormatRunSeed(got.Seed))
	}
	if got.Level != 9 || got.MaxStreak != 7 || got.PeakWanted != 5 || got.PerfectDrops != 11 || got.EventsCompleted != 3 {
		t.Fatalf("highlight stats mismatch: %+v", got)
	}
	if got.BoardedByArch != w.BoardedByArch {
		t.Fatalf("BoardedByArch=%v; want %v", got.BoardedByArch, w.BoardedByArch)
	}
	if got.UpgradeRanks != w.Player.UpgradeRanks {
		t.Fatalf("UpgradeRanks=%v; want %v", got.UpgradeRanks, w.Player.UpgradeRanks)
	}
	if got.Relics != w.Player.Relics {
		t.Fatalf("Relics=%v; want %v", got.Relics, w.Player.Relics)
	}
	if got.Quip == "" {
		t.Fatal("expected non-empty Quip")
	}
	wantQuip := game.EndRunQuip(true, true, 5, 5, 7)
	if got.Quip != wantQuip {
		t.Fatalf("Quip=%q; want %q", got.Quip, wantQuip)
	}

	// Verify endRunButtonRect geometry & click-hit consistency for Victory (3 buttons) and GameOver (2 buttons).
	save := meta.DefaultSaveData()
	b1 := endRunButtonRect(1, 3)
	res := m.Update(input.MenuInput{MouseX: b1.X + 4, MouseY: b1.Y + 4, MouseLeftJustPressed: true}, &save)
	if res.Action != ActionRestart {
		t.Fatalf("victory button 1 click returned %+v; want ActionRestart", res)
	}
	m.SetEndRun(false, w, false, nil)
	if m.Scene != SceneGameOver {
		t.Fatalf("scene = %d; want SceneGameOver", m.Scene)
	}
	b0 := endRunButtonRect(0, 2)
	if b0.Y < 300 || b0.Y+b0.H > 348 {
		t.Fatalf("endRunButtonRect Y range [%.0f, %.0f] outside bottom footer", b0.Y, b0.Y+b0.H)
	}
	res = m.Update(input.MenuInput{MouseX: b0.X + 4, MouseY: b0.Y + 4, MouseLeftJustPressed: true}, &save)
	if res.Action != ActionRestart {
		t.Fatalf("gameover button 0 click returned %+v; want ActionRestart", res)
	}
}
