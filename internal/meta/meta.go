package meta

import "math"

// Filter mode constants for post-processing shader selection.
const (
	FilterOff     = 0
	FilterWarmCRT = 1
	FilterVHS     = 2
)

// Player skin constants.
const (
	SkinAlfredoBase  = 0
	SkinNeonDrip     = 1
	SkinCyberSamurai = 2
	PlayerSkinCount  = 3
)

// Bus skin constants.
const (
	BusNeonClassic = 0
	BusGoldDrift   = 1
	BusSkinCount   = 2
)

// Settings holds persisted audio, display, shader, and control options.
type Settings struct {
	FilterMode    int     `json:"filter_mode"`
	MasterVol     float64 `json:"master_vol"`
	MusicVol      float64 `json:"music_vol"`
	SFXVol        float64 `json:"sfx_vol"`
	Fullscreen    bool    `json:"fullscreen"`
	ScreenShake   bool    `json:"screen_shake"`
	ShowFPS       bool    `json:"show_fps"`
	ChronosToggle bool    `json:"chronos_toggle"`
}

// Profile holds persisted player progression, stats, and cosmetic selections.
type Profile struct {
	BestScore          int             `json:"best_score"`
	BestTimeSec        int             `json:"best_time_sec"`
	TotalRuns          int             `json:"total_runs"`
	TotalGirlsBoarded  int             `json:"total_girls_boarded"`
	BossDefeatedMask   uint8           `json:"boss_defeated_mask"`
	UnlockedChallenges map[string]bool `json:"unlocked_challenges"`
	SelectedPlayerSkin int             `json:"selected_player_skin"`
	SelectedBusSkin    int             `json:"selected_bus_skin"`
}

// SaveData is the root persisted save payload.
type SaveData struct {
	Settings Settings `json:"settings"`
	Profile  Profile  `json:"profile"`
}

// DefaultSaveData returns clean initial settings and an empty profile.
func DefaultSaveData() SaveData {
	return SaveData{
		Settings: Settings{
			FilterMode:    FilterWarmCRT,
			MasterVol:     0.80,
			MusicVol:      0.65,
			SFXVol:        0.85,
			Fullscreen:    false,
			ScreenShake:   true,
			ShowFPS:       false,
			ChronosToggle: false,
		},
		Profile: Profile{
			UnlockedChallenges: make(map[string]bool),
			SelectedPlayerSkin: SkinAlfredoBase,
			SelectedBusSkin:    BusNeonClassic,
		},
	}
}

func clamp01(v float64) float64 {
	if math.IsNaN(v) || v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// Normalize clamps settings, ensures non-nil maps, and validates selected skins.
func (s *SaveData) Normalize() {
	if s.Settings.FilterMode < FilterOff {
		s.Settings.FilterMode = FilterOff
	} else if s.Settings.FilterMode > FilterVHS {
		s.Settings.FilterMode = FilterVHS
	}

	s.Settings.MasterVol = clamp01(s.Settings.MasterVol)
	s.Settings.MusicVol = clamp01(s.Settings.MusicVol)
	s.Settings.SFXVol = clamp01(s.Settings.SFXVol)

	if s.Profile.UnlockedChallenges == nil {
		s.Profile.UnlockedChallenges = make(map[string]bool)
	}

	if s.Profile.BestScore < 0 {
		s.Profile.BestScore = 0
	}
	if s.Profile.BestTimeSec < 0 {
		s.Profile.BestTimeSec = 0
	}
	if s.Profile.TotalRuns < 0 {
		s.Profile.TotalRuns = 0
	}
	if s.Profile.TotalGirlsBoarded < 0 {
		s.Profile.TotalGirlsBoarded = 0
	}

	if s.Profile.SelectedPlayerSkin < 0 || s.Profile.SelectedPlayerSkin >= PlayerSkinCount ||
		!s.Profile.IsPlayerSkinUnlocked(s.Profile.SelectedPlayerSkin) {
		s.Profile.SelectedPlayerSkin = SkinAlfredoBase
	}
	if s.Profile.SelectedBusSkin < 0 || s.Profile.SelectedBusSkin >= BusSkinCount ||
		!s.Profile.IsBusSkinUnlocked(s.Profile.SelectedBusSkin) {
		s.Profile.SelectedBusSkin = BusNeonClassic
	}
}

// MarkBossDefeated records that the given BossKind (0..5) has been defeated.
func (p *Profile) MarkBossDefeated(bossKind int) {
	if bossKind < 0 || bossKind >= 6 {
		return
	}
	p.BossDefeatedMask |= 1 << uint(bossKind)
}

// IsBossDefeated reports whether the given BossKind (0..5) has been defeated.
func (p Profile) IsBossDefeated(bossKind int) bool {
	if bossKind < 0 || bossKind >= 6 {
		return false
	}
	return (p.BossDefeatedMask & (1 << uint(bossKind))) != 0
}

// AllSixBossesDefeated reports whether all 6 bosses (bits 0..5) are defeated.
func (p Profile) AllSixBossesDefeated() bool {
	return (p.BossDefeatedMask & 0x3F) == 0x3F
}

// IsPlayerSkinUnlocked reports whether the given player skin index is unlocked.
func (p Profile) IsPlayerSkinUnlocked(skin int) bool {
	switch skin {
	case SkinAlfredoBase:
		return true
	case SkinNeonDrip:
		return p.UnlockedChallenges["first_ride"]
	case SkinCyberSamurai:
		return p.UnlockedChallenges["chaos_king_5star"]
	default:
		return false
	}
}

// IsBusSkinUnlocked reports whether the given bus skin index is unlocked.
func (p Profile) IsBusSkinUnlocked(skin int) bool {
	switch skin {
	case BusNeonClassic:
		return true
	case BusGoldDrift:
		return p.UnlockedChallenges["rizz_jackpot"]
	default:
		return false
	}
}

// IsBossRushUnlocked reports whether Boss Rush mode has been unlocked.
func (p Profile) IsBossRushUnlocked() bool {
	return p.UnlockedChallenges["all_6_bosses"] || p.AllSixBossesDefeated()
}

// PlayerSkinName returns the Italian arcade display name for a player skin.
func (p Profile) PlayerSkinName(skin int) string {
	switch skin {
	case SkinAlfredoBase:
		return "Alfredo Base"
	case SkinNeonDrip:
		return "Neon Drip"
	case SkinCyberSamurai:
		return "Cyber-Samurai Zero"
	default:
		return "Sconosciuta"
	}
}

// BusSkinName returns the Italian arcade display name for a bus skin.
func (p Profile) BusSkinName(skin int) string {
	switch skin {
	case BusNeonClassic:
		return "Autobus Neon"
	case BusGoldDrift:
		return "Autobus Dorato Drift"
	default:
		return "Sconosciuto"
	}
}
