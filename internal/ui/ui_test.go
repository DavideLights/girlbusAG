package ui

import (
	"testing"

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
