package audio

import (
	"math"
	"sync"
	"time"

	ebitenaudio "github.com/hajimehoshi/ebiten/v2/audio"
)

// SampleRate is the fixed PCM sample rate used by the procedural synth and SFX bank.
const SampleRate = 44100

// SFXID identifies a pre-rendered sound effect buffer.
type SFXID int

const (
	SFXThrow SFXID = iota
	SFXPerfectDrop
	SFXHook
	SFXMinigameHit
	SFXMinigameMiss
	SFXComboStep1
	SFXComboStep2
	SFXComboStep3
	SFXComboStep4
	SFXBusHorn
	SFXDash
	SFXPlayerHit
	SFXPoliceSiren
	SFXBossSpawn
	SFXBossBreak
	SFXLevelUp
	SFXShopBuy
	SFXSlotTick
	SFXSlotWin
	SFXSlotJackpot
	SFXSlotBust
	SFXUIClick
	SFXChallengeUnlock
	sfxCount
)

const sfxPoolPerID = 4

// Engine manages procedural synthwave music playback, Katana ZERO Chronos low-pass
// ducking, and zero-allocation pooled playback of pre-rendered arcade SFX.
type Engine struct {
	mu          sync.Mutex
	ctx         *ebitenaudio.Context
	musicPlayer *ebitenaudio.Player
	synth       *synthReader

	sfxBuffers  [sfxCount][]byte
	sfxPlayers  [sfxCount][sfxPoolPerID]*ebitenaudio.Player
	sfxNextSlot [sfxCount]int
	lastPlayed  [sfxCount]time.Time
	cooldowns   [sfxCount]time.Duration

	masterVol float64
	musicVol  float64
	sfxVol    float64
	paused    bool
	unlocked  bool
}

// NewEngine constructs an Engine with pre-rendered SFX buffers and a lazy audio context.
// Safe to call in headless tests and before first user interaction on Wasm/browsers.
func NewEngine() *Engine {
	e := &Engine{
		synth:      newSynthReader(),
		sfxBuffers: buildSFXBank(),
		masterVol:  0.85,
		musicVol:   0.70,
		sfxVol:     0.90,
	}

	for i := range e.cooldowns {
		e.cooldowns[i] = 45 * time.Millisecond
	}
	e.cooldowns[SFXSlotTick] = 35 * time.Millisecond
	e.cooldowns[SFXUIClick] = 35 * time.Millisecond
	e.cooldowns[SFXPlayerHit] = 60 * time.Millisecond
	e.cooldowns[SFXPoliceSiren] = 60 * time.Millisecond
	e.cooldowns[SFXBusHorn] = 60 * time.Millisecond
	e.cooldowns[SFXBossSpawn] = 60 * time.Millisecond
	e.cooldowns[SFXBossBreak] = 60 * time.Millisecond

	return e
}

func clamp01(v float64) float64 {
	if math.IsNaN(v) || v < 0.0 {
		return 0.0
	}
	if v > 1.0 {
		return 1.0
	}
	return v
}

func (e *Engine) ensureContext() *ebitenaudio.Context {
	if e.ctx != nil {
		return e.ctx
	}
	ctx := ebitenaudio.CurrentContext()
	if ctx == nil {
		func() {
			defer func() {
				_ = recover()
			}()
			ctx = ebitenaudio.NewContext(SampleRate)
		}()
	}
	e.ctx = ctx
	return e.ctx
}

// Unlock lazily initializes the Ebitengine audio context and starts the synthwave
// music stream on first user interaction (required for browser/Wasm autoplay).
func (e *Engine) Unlock() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.unlocked = true
	ctx := e.ensureContext()
	if ctx == nil {
		return
	}

	if e.musicPlayer == nil {
		p, err := ctx.NewPlayer(e.synth)
		if err != nil || p == nil {
			return
		}
		p.SetBufferSize(50 * time.Millisecond)
		e.musicPlayer = p
		e.applyMusicVolumeLocked()
		e.musicPlayer.Play()
		return
	}

	if !e.musicPlayer.IsPlaying() {
		e.musicPlayer.Play()
	}
}

// SetVolumes sets master, music, and SFX gains clamped to [0, 1].
func (e *Engine) SetVolumes(master, music, sfx float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.masterVol = clamp01(master)
	e.musicVol = clamp01(music)
	e.sfxVol = clamp01(sfx)
	e.applyMusicVolumeLocked()
}

func (e *Engine) applyMusicVolumeLocked() {
	if e.musicPlayer == nil {
		return
	}
	vol := e.masterVol * e.musicVol
	if e.paused {
		vol *= 0.45
	}
	e.musicPlayer.SetVolume(vol)
}

// UpdateMusicState updates dynamic BPM, layer intensity, pause ducking, and the
// Chronos low-pass filter cutoff.
func (e *Engine) UpdateMusicState(playing bool, tier int, bossActive bool, wantedStars int, chronosActive bool, paused bool) {
	e.mu.Lock()
	e.paused = paused
	e.applyMusicVolumeLocked()
	e.mu.Unlock()

	e.synth.setMusicState(playing, tier, bossActive, wantedStars, chronosActive, paused)
}

// SetFever toggles the octave-up arpeggio shimmer layer used during Rizz Fever (x5 Rush Hour).
func (e *Engine) SetFever(active bool) {
	e.synth.setFever(active)
}

// CurrentBPM returns the live smoothed Synthwave sequencer BPM.
func (e *Engine) CurrentBPM() float32 {
	return float32(e.synth.bpm())
}

// PlaySFX plays a pre-rendered sound effect scaled by master * sfx with per-SFX
// cooldown throttling to avoid clipping/spamming.
func (e *Engine) PlaySFX(id SFXID) {
	if id < 0 || id >= sfxCount {
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.unlocked {
		return
	}
	ctx := e.ensureContext()
	if ctx == nil {
		return
	}

	now := time.Now()
	if !e.lastPlayed[id].IsZero() && now.Sub(e.lastPlayed[id]) < e.cooldowns[id] {
		return
	}

	buf := e.sfxBuffers[id]
	if len(buf) == 0 {
		return
	}

	vol := e.masterVol * e.sfxVol
	if vol <= 0.0001 {
		return
	}

	// Reuse an idle pooled player if available.
	for i := 0; i < sfxPoolPerID; i++ {
		if p := e.sfxPlayers[id][i]; p != nil && !p.IsPlaying() {
			_ = p.Rewind()
			p.SetVolume(vol)
			p.Play()
			e.lastPlayed[id] = now
			return
		}
	}

	// Allocate into an empty pool slot if available.
	for i := 0; i < sfxPoolPerID; i++ {
		if e.sfxPlayers[id][i] == nil {
			p := ctx.NewPlayerFromBytes(buf)
			if p == nil {
				return
			}
			e.sfxPlayers[id][i] = p
			p.SetVolume(vol)
			p.Play()
			e.lastPlayed[id] = now
			return
		}
	}

	// Recycle the oldest active slot when all pool slots are saturated.
	slot := e.sfxNextSlot[id]
	e.sfxNextSlot[id] = (slot + 1) % sfxPoolPerID
	p := e.sfxPlayers[id][slot]
	if p != nil {
		_ = p.Rewind()
		p.SetVolume(vol)
		p.Play()
		e.lastPlayed[id] = now
	}
}
