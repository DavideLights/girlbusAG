package meta

import (
	"encoding/json"
	"math"
	"testing"
)

func TestDefaultSaveData(t *testing.T) {
	save := DefaultSaveData()

	if save.Settings.FilterMode != FilterWarmCRT {
		t.Fatalf("FilterMode = %d, want %d", save.Settings.FilterMode, FilterWarmCRT)
	}
	if save.Settings.MasterVol != 0.80 {
		t.Fatalf("MasterVol = %v, want 0.80", save.Settings.MasterVol)
	}
	if save.Settings.MusicVol != 0.65 {
		t.Fatalf("MusicVol = %v, want 0.65", save.Settings.MusicVol)
	}
	if save.Settings.SFXVol != 0.85 {
		t.Fatalf("SFXVol = %v, want 0.85", save.Settings.SFXVol)
	}
	if save.Settings.Fullscreen {
		t.Fatalf("Fullscreen = true, want false")
	}
	if !save.Settings.ScreenShake {
		t.Fatalf("ScreenShake = false, want true")
	}
	if save.Settings.ShowFPS {
		t.Fatalf("ShowFPS = true, want false")
	}
	if save.Settings.ChronosToggle {
		t.Fatalf("ChronosToggle = true, want false")
	}
	if save.Profile.UnlockedChallenges == nil {
		t.Fatalf("UnlockedChallenges is nil, want initialized map")
	}
	if !save.Profile.IsPlayerSkinUnlocked(SkinAlfredoBase) {
		t.Fatalf("SkinAlfredoBase should be unlocked by default")
	}
	if save.Profile.IsPlayerSkinUnlocked(SkinNeonDrip) {
		t.Fatalf("SkinNeonDrip should be locked by default")
	}
	if save.Profile.IsPlayerSkinUnlocked(SkinBanana) {
		t.Fatalf("SkinBanana should be locked by default")
	}
	if save.Profile.IsPlayerSkinUnlocked(SkinChickenWing) {
		t.Fatalf("SkinChickenWing should be locked by default")
	}
	if save.Profile.IsPlayerSkinUnlocked(SkinCyberSamurai) {
		t.Fatalf("SkinCyberSamurai should be locked by default")
	}
	if PlayerSkinCount != 5 {
		t.Fatalf("PlayerSkinCount = %d, want 5", PlayerSkinCount)
	}
	if got := save.Profile.UnlockedTitles(); len(got) != 1 || got[0] != "" {
		t.Fatalf("default UnlockedTitles() = %v, want [\"\"]", got)
	}
	if !save.Profile.IsTitleUnlocked("") || save.Profile.IsTitleUnlocked("Veterano") {
		t.Fatalf("unexpected default IsTitleUnlocked behavior")
	}
	if !save.Profile.IsBusSkinUnlocked(BusNeonClassic) {
		t.Fatalf("BusNeonClassic should be unlocked by default")
	}
	if save.Profile.IsBusSkinUnlocked(BusGoldDrift) {
		t.Fatalf("BusGoldDrift should be locked by default")
	}
	if save.Profile.IsBossRushUnlocked() {
		t.Fatalf("BossRush should be locked by default")
	}
}

func TestChallengesAndSkinUnlocks(t *testing.T) {
	if got := len(AllChallenges()); got != 10 {
		t.Fatalf("len(AllChallenges()) = %d, want 10", got)
	}

	save := DefaultSaveData()
	p := &save.Profile
	rt := NewRunTracker()

	// 1. first_ride -> unlocks SkinNeonDrip
	rt.GirlsBoarded = 10
	// 2. perfect_surgeon
	rt.PerfectDrops = 5
	// 3. outlaw_3star (30s at >=3*) & 4. chaos_king_5star (45s at 5* no-hit)
	rt.TickWanted(20, 3)
	rt.TickWanted(10, 5)
	// Hit resets 5* streak; 3* total is already 30s
	rt.OnPlayerHit(false)
	if rt.NoHitTimeAt5Star != 0 {
		t.Fatalf("NoHitTimeAt5Star = %v, want 0 after hit", rt.NoHitTimeAt5Star)
	}
	rt.TickWanted(44, 5)
	unlocked := rt.CheckAndUnlock(p)
	if p.UnlockedChallenges["chaos_king_5star"] {
		t.Fatalf("chaos_king_5star should not unlock at 44s")
	}
	if len(unlocked) != 3 {
		t.Fatalf("expected 3 unlocks on first check (first_ride, perfect_surgeon, outlaw_3star), got %d", len(unlocked))
	}
	if !p.IsPlayerSkinUnlocked(SkinNeonDrip) {
		t.Fatalf("SkinNeonDrip should be unlocked after first_ride")
	}

	// Finish 45s no-hit at 5* -> unlocks SkinCyberSamurai
	rt.TickWanted(1, 5)
	// 5. full_bus
	rt.MaxHookInOneStop = 5
	// 6. rizz_jackpot -> unlocks BusGoldDrift
	rt.JackpotWon = true
	// 7. all_synergies
	rt.SynergyMask = 0x0F
	// 8. beat_rector_3star (boss 4 at 2* should NOT unlock; at 3* SHOULD unlock)
	rt.RecordBossDefeated(4, 2, p)
	if !p.IsBossDefeated(4) {
		t.Fatalf("Boss 4 (Rector) should be marked defeated")
	}
	rt.CheckAndUnlock(p)
	if p.UnlockedChallenges["beat_rector_3star"] {
		t.Fatalf("beat_rector_3star should not unlock at 2*")
	}
	rt.RecordBossDefeated(4, 3, p)

	// 9. beat_mom_clean: fails if HitByCiabatta == true
	rt.OnPlayerHit(true)
	rt.RecordBossDefeated(5, 0, p)
	rt.CheckAndUnlock(p)
	if p.UnlockedChallenges["beat_mom_clean"] {
		t.Fatalf("beat_mom_clean must NOT unlock when HitByCiabatta is true")
	}

	// Verify skins unlocked so far
	if !p.IsPlayerSkinUnlocked(SkinBanana) {
		t.Fatalf("SkinBanana should be unlocked after full_bus")
	}
	if !p.IsPlayerSkinUnlocked(SkinChickenWing) {
		t.Fatalf("SkinChickenWing should be unlocked after all_synergies")
	}
	if !p.IsPlayerSkinUnlocked(SkinCyberSamurai) {
		t.Fatalf("SkinCyberSamurai should be unlocked after chaos_king_5star")
	}
	if !p.IsBusSkinUnlocked(BusGoldDrift) {
		t.Fatalf("BusGoldDrift should be unlocked after rizz_jackpot")
	}

	// Verify skin display names
	wantNames := map[int]string{
		SkinAlfredoBase:  "Alfredo Base",
		SkinNeonDrip:     "Neon Drip",
		SkinBanana:       "Alfredo Banana",
		SkinChickenWing:  "Aletta di Pollo",
		SkinCyberSamurai: "Cyber-Samurai Zero",
	}
	for skinID, want := range wantNames {
		if got := p.PlayerSkinName(skinID); got != want {
			t.Fatalf("PlayerSkinName(%d) = %q, want %q", skinID, got, want)
		}
	}

	// Clean run for Mom + remaining bosses (0, 1, 2, 3) -> beat_mom_clean & all_6_bosses -> Boss Rush
	rtClean := NewRunTracker()
	rtClean.RecordBossDefeated(0, 1, p)
	rtClean.RecordBossDefeated(1, 1, p)
	rtClean.RecordBossDefeated(2, 1, p)
	rtClean.RecordBossDefeated(3, 1, p)
	rtClean.RecordBossDefeated(5, 2, p)
	rtClean.FinalizeRun(12500, 180, 14, p)

	for _, ch := range AllChallenges() {
		if !p.UnlockedChallenges[ch.ID] {
			t.Fatalf("challenge %q was not unlocked", ch.ID)
		}
	}
	if !p.AllSixBossesDefeated() {
		t.Fatalf("AllSixBossesDefeated() = false, want true")
	}
	if !p.IsBossRushUnlocked() {
		t.Fatalf("IsBossRushUnlocked() = false, want true")
	}
	if p.BestScore != 12500 || p.BestTimeSec != 180 || p.TotalRuns != 1 || p.TotalGirlsBoarded != 14 {
		t.Fatalf("unexpected finalized profile stats: %+v", *p)
	}

	wantTitles := []string{"", "Veterano", "Gloria Criminale", "Onore della Strada", "Figlio Modello"}
	gotTitles := p.UnlockedTitles()
	if len(gotTitles) != len(wantTitles) {
		t.Fatalf("UnlockedTitles() = %v, want %v", gotTitles, wantTitles)
	}
	for i := range wantTitles {
		if gotTitles[i] != wantTitles[i] {
			t.Fatalf("UnlockedTitles()[%d] = %q, want %q", i, gotTitles[i], wantTitles[i])
		}
		if !p.IsTitleUnlocked(wantTitles[i]) {
			t.Fatalf("IsTitleUnlocked(%q) = false, want true", wantTitles[i])
		}
	}
	if p.IsTitleUnlocked("Titolo Inesistente") {
		t.Fatalf("IsTitleUnlocked(\"Titolo Inesistente\") = true, want false")
	}
}

func TestFinalizeRunVictoryThenOverdriveNoDoubleCount(t *testing.T) {
	save := DefaultSaveData()
	p := &save.Profile
	rt := NewRunTracker()

	// Victory at 14 girls
	rt.FinalizeRun(10000, 150, 14, p)
	if p.TotalRuns != 1 || p.TotalGirlsBoarded != 14 || p.BestScore != 10000 || p.BestTimeSec != 150 {
		t.Fatalf("after victory FinalizeRun: got %+v", *p)
	}

	// Continue in Overdrive, then Game Over at 22 girls
	rt.FinalizeRun(18500, 210, 22, p)
	if p.TotalRuns != 1 {
		t.Fatalf("TotalRuns = %d, want 1 (no double-count)", p.TotalRuns)
	}
	if p.TotalGirlsBoarded != 22 {
		t.Fatalf("TotalGirlsBoarded = %d, want 22", p.TotalGirlsBoarded)
	}
	if p.BestScore != 18500 || p.BestTimeSec != 210 {
		t.Fatalf("BestScore/BestTimeSec not updated after overdrive: %+v", *p)
	}
}

func TestJSONRoundTripAndNormalize(t *testing.T) {
	orig := DefaultSaveData()
	orig.Settings.FilterMode = 99
	orig.Settings.MasterVol = -0.5
	orig.Settings.MusicVol = 1.75
	orig.Settings.SFXVol = math.NaN()
	orig.Settings.DirInputMode = 7
	orig.Profile.UnlockedChallenges = nil
	orig.Profile.SelectedPlayerSkin = SkinCyberSamurai // locked -> must fall back to 0
	orig.Profile.SelectedBusSkin = BusGoldDrift        // locked -> must fall back to 0
	orig.Profile.SelectedTitle = "Veterano"            // locked -> must fall back to ""

	orig.Normalize()

	if orig.Settings.FilterMode != FilterVHS {
		t.Fatalf("FilterMode = %d, want %d", orig.Settings.FilterMode, FilterVHS)
	}
	if orig.Settings.DirInputMode != DirModeAroundStop {
		t.Fatalf("DirInputMode = %d, want %d", orig.Settings.DirInputMode, DirModeAroundStop)
	}
	if orig.Settings.MasterVol != 0 {
		t.Fatalf("MasterVol = %v, want 0", orig.Settings.MasterVol)
	}
	if orig.Settings.MusicVol != 1 {
		t.Fatalf("MusicVol = %v, want 1", orig.Settings.MusicVol)
	}
	if orig.Settings.SFXVol != 0 {
		t.Fatalf("SFXVol = %v, want 0", orig.Settings.SFXVol)
	}
	if orig.Profile.UnlockedChallenges == nil {
		t.Fatalf("UnlockedChallenges should be initialized by Normalize()")
	}
	if orig.Profile.SelectedPlayerSkin != SkinAlfredoBase {
		t.Fatalf("SelectedPlayerSkin = %d, want %d", orig.Profile.SelectedPlayerSkin, SkinAlfredoBase)
	}
	if orig.Profile.SelectedBusSkin != BusNeonClassic {
		t.Fatalf("SelectedBusSkin = %d, want %d", orig.Profile.SelectedBusSkin, BusNeonClassic)
	}
	if orig.Profile.SelectedTitle != "" {
		t.Fatalf("SelectedTitle = %q, want \"\"", orig.Profile.SelectedTitle)
	}

	// Unlock skins/title and test full JSON round-trip
	orig.Settings.FilterMode = FilterOff
	orig.Settings.MasterVol = 0.42
	orig.Settings.Fullscreen = true
	orig.Settings.ShowFPS = true
	orig.Settings.ChronosToggle = true
	orig.Settings.DirInputMode = DirModeAroundStop
	orig.Profile.UnlockedChallenges["full_bus"] = true
	orig.Profile.UnlockedChallenges["all_synergies"] = true
	orig.Profile.UnlockedChallenges["chaos_king_5star"] = true
	orig.Profile.UnlockedChallenges["rizz_jackpot"] = true
	orig.Profile.UnlockedChallenges["perfect_surgeon"] = true
	orig.Profile.SelectedPlayerSkin = SkinChickenWing
	orig.Profile.SelectedBusSkin = BusGoldDrift
	orig.Profile.SelectedTitle = "Veterano"
	orig.Profile.MarkBossDefeated(2)

	raw, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	loaded := DefaultSaveData()
	if err := json.Unmarshal(raw, &loaded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}
	loaded.Normalize()

	if loaded.Settings.FilterMode != FilterOff || !loaded.Settings.Fullscreen || !loaded.Settings.ShowFPS || !loaded.Settings.ChronosToggle || loaded.Settings.DirInputMode != DirModeAroundStop {
		t.Fatalf("unexpected loaded settings: %+v", loaded.Settings)
	}
	if loaded.Settings.MasterVol != 0.42 {
		t.Fatalf("loaded MasterVol = %v, want 0.42", loaded.Settings.MasterVol)
	}
	if loaded.Profile.SelectedPlayerSkin != SkinChickenWing {
		t.Fatalf("loaded SelectedPlayerSkin = %d, want %d", loaded.Profile.SelectedPlayerSkin, SkinChickenWing)
	}
	if loaded.Profile.SelectedBusSkin != BusGoldDrift {
		t.Fatalf("loaded SelectedBusSkin = %d, want %d", loaded.Profile.SelectedBusSkin, BusGoldDrift)
	}
	if loaded.Profile.SelectedTitle != "Veterano" {
		t.Fatalf("loaded SelectedTitle = %q, want \"Veterano\"", loaded.Profile.SelectedTitle)
	}
	if !loaded.Profile.IsBossDefeated(2) {
		t.Fatalf("Boss 2 defeat did not survive JSON round-trip")
	}
}
