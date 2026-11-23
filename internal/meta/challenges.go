package meta

const (
	bossKindRector = 4
	bossKindMom    = 5
)

// ChallengeDef describes an in-game challenge and its Italian UI copy.
type ChallengeDef struct {
	ID       string
	TitleIT  string
	DescIT   string
	RewardIT string
}

var allChallenges = []ChallengeDef{
	{
		ID:       "first_ride",
		TitleIT:  "Primo Giro",
		DescIT:   "Carica almeno 10 ragazze in una sola partita",
		RewardIT: "Skin Alfredo: Neon Drip",
	},
	{
		ID:       "perfect_surgeon",
		TitleIT:  "Chirurgo della Pensilina",
		DescIT:   "Esegui 5 Perfect Drop in una sola partita",
		RewardIT: "Badge Veterano",
	},
	{
		ID:       "outlaw_3star",
		TitleIT:  "Fuorilegge",
		DescIT:   "Sopravvivi 30s con almeno 3* Wanted",
		RewardIT: "Gloria Criminale",
	},
	{
		ID:       "chaos_king_5star",
		TitleIT:  "Re del Caos 5*",
		DescIT:   "Sopravvivi 45s a 5* Wanted senza subire danni",
		RewardIT: "Skin: Cyber-Samurai Zero",
	},
	{
		ID:       "full_bus",
		TitleIT:  "Autobus Pieno",
		DescIT:   "Cattura almeno 5 ragazze con una sola Fermata",
		RewardIT: "Titolo Capolinea",
	},
	{
		ID:       "rizz_jackpot",
		TitleIT:  "Re della Slot",
		DescIT:   "Vinci un Jackpot 777 alla Rizz-Slots",
		RewardIT: "Autobus Dorato Drift",
	},
	{
		ID:       "all_synergies",
		TitleIT:  "Affinita Totale",
		DescIT:   "Attiva tutte e 4 le Synergy in una sola partita",
		RewardIT: "Maestria Urbis",
	},
	{
		ID:       "beat_rector_3star",
		TitleIT:  "Ribelle Accademico",
		DescIT:   "Sconfiggi La Rettrice con almeno 3* Wanted",
		RewardIT: "Onore della Strada",
	},
	{
		ID:       "beat_mom_clean",
		TitleIT:  "Niente Ciabatta!",
		DescIT:   "Sconfiggi La Mamma senza farti colpire dalla Ciabatta",
		RewardIT: "Figlio Modello",
	},
	{
		ID:       "all_6_bosses",
		TitleIT:  "Collezionista di Boss",
		DescIT:   "Sconfiggi tutti e 6 i Boss almeno una volta",
		RewardIT: "Modalita Boss Rush",
	},
}

// AllChallenges returns a fresh slice of all 10 Italian challenge definitions.
func AllChallenges() []ChallengeDef {
	out := make([]ChallengeDef, len(allChallenges))
	copy(out, allChallenges)
	return out
}

// RunTracker accumulates per-run challenge progress and evaluates unlocks.
type RunTracker struct {
	GirlsBoarded      int
	PerfectDrops      int
	MaxHookInOneStop  int
	SynergyMask       uint8
	TimeAt3Star       float32
	NoHitTimeAt5Star  float32
	HitByCiabatta     bool
	JackpotWon        bool
	BeatRectorAt3Star bool
	BeatMomThisRun    bool
	NewlyUnlocked     []ChallengeDef
}

// NewRunTracker creates an empty RunTracker for a new run.
func NewRunTracker() RunTracker {
	return RunTracker{}
}

// TickWanted advances wanted-tier survival timers by dt seconds.
func (rt *RunTracker) TickWanted(dt float32, wantedStars int) {
	if dt <= 0 {
		return
	}
	if wantedStars >= 3 {
		rt.TimeAt3Star += dt
	}
	if wantedStars >= 5 {
		rt.NoHitTimeAt5Star += dt
	}
}

// OnPlayerHit resets the no-hit 5* streak and marks ciabatta hits.
func (rt *RunTracker) OnPlayerHit(byCiabatta bool) {
	rt.NoHitTimeAt5Star = 0
	if byCiabatta {
		rt.HitByCiabatta = true
	}
}

// RecordBossDefeated updates boss defeat tracking on both the profile and run tracker.
func (rt *RunTracker) RecordBossDefeated(bossKind int, wantedStars int, profile *Profile) {
	if profile != nil {
		profile.MarkBossDefeated(bossKind)
	}
	if bossKind == bossKindRector && wantedStars >= 3 {
		rt.BeatRectorAt3Star = true
	}
	if bossKind == bossKindMom {
		rt.BeatMomThisRun = true
	}
}

func (rt *RunTracker) isSatisfied(id string, profile *Profile) bool {
	switch id {
	case "first_ride":
		return rt.GirlsBoarded >= 10
	case "perfect_surgeon":
		return rt.PerfectDrops >= 5
	case "outlaw_3star":
		return rt.TimeAt3Star >= 30
	case "chaos_king_5star":
		return rt.NoHitTimeAt5Star >= 45
	case "full_bus":
		return rt.MaxHookInOneStop >= 5
	case "rizz_jackpot":
		return rt.JackpotWon
	case "all_synergies":
		return (rt.SynergyMask & 0x0F) == 0x0F
	case "beat_rector_3star":
		return rt.BeatRectorAt3Star
	case "beat_mom_clean":
		return rt.BeatMomThisRun && !rt.HitByCiabatta
	case "all_6_bosses":
		return profile.AllSixBossesDefeated()
	default:
		return false
	}
}

// CheckAndUnlock evaluates all 10 challenges, persists any new unlocks to profile,
// appends them to rt.NewlyUnlocked, and returns the newly unlocked slice for this check.
func (rt *RunTracker) CheckAndUnlock(profile *Profile) []ChallengeDef {
	if profile == nil {
		return nil
	}
	if profile.UnlockedChallenges == nil {
		profile.UnlockedChallenges = make(map[string]bool)
	}

	var unlocked []ChallengeDef
	for _, ch := range allChallenges {
		if profile.UnlockedChallenges[ch.ID] {
			continue
		}
		if rt.isSatisfied(ch.ID, profile) {
			profile.UnlockedChallenges[ch.ID] = true
			rt.NewlyUnlocked = append(rt.NewlyUnlocked, ch)
			unlocked = append(unlocked, ch)
		}
	}
	return unlocked
}

// FinalizeRun commits run totals to profile and evaluates all challenge unlocks.
func (rt *RunTracker) FinalizeRun(score, timeSec, girlsBoarded int, profile *Profile) {
	if profile == nil {
		return
	}
	if girlsBoarded > rt.GirlsBoarded {
		rt.GirlsBoarded = girlsBoarded
	}
	if score > profile.BestScore {
		profile.BestScore = score
	}
	if timeSec > profile.BestTimeSec {
		profile.BestTimeSec = timeSec
	}
	profile.TotalRuns++
	if girlsBoarded > 0 {
		profile.TotalGirlsBoarded += girlsBoarded
	}
	rt.CheckAndUnlock(profile)
}
