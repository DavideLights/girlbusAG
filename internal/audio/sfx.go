package audio

import "math"

type sfxRNG uint32

func (r *sfxRNG) next() float64 {
	x := uint32(*r)
	x ^= x << 13
	x ^= x >> 17
	x ^= x << 5
	*r = sfxRNG(x)
	return (float64(x) / 2147483647.5) - 1.0
}

func oscSine(phase float64) float64 {
	return math.Sin(2.0 * math.Pi * phase)
}

func oscSaw(phase float64) float64 {
	_, frac := math.Modf(phase)
	if frac < 0 {
		frac += 1.0
	}
	return 2.0*frac - 1.0
}

func oscSquare(phase float64) float64 {
	_, frac := math.Modf(phase)
	if frac < 0 {
		frac += 1.0
	}
	if frac < 0.5 {
		return 1.0
	}
	return -1.0
}

func oscTri(phase float64) float64 {
	_, frac := math.Modf(phase)
	if frac < 0 {
		frac += 1.0
	}
	return 1.0 - 4.0*math.Abs(frac-0.5)
}

// renderSFX synthesizes a stereo 16-bit signed LE PCM buffer with smooth
// anti-click fade-in/fade-out edges.
func renderSFX(durationSec float64, fn func(i int, t, dt, dur float64) (float64, float64)) []byte {
	frames := int(math.Round(durationSec * float64(SampleRate)))
	if frames < 1 {
		frames = 1
	}
	buf := make([]byte, frames*4)
	const dt = 1.0 / float64(SampleRate)
	fadeInFrames := int(math.Round(0.002 * float64(SampleRate)))
	fadeOutFrames := int(math.Round(0.008 * float64(SampleRate)))
	if fadeInFrames < 1 {
		fadeInFrames = 1
	}
	if fadeOutFrames < 1 {
		fadeOutFrames = 1
	}

	for i := 0; i < frames; i++ {
		t := float64(i) * dt
		l, r := fn(i, t, dt, durationSec)

		edge := 1.0
		if i < fadeInFrames {
			edge *= float64(i) / float64(fadeInFrames)
		}
		if rem := frames - 1 - i; rem < fadeOutFrames {
			edge *= float64(rem) / float64(fadeOutFrames)
		}

		outL := math.Tanh(l * edge)
		outR := math.Tanh(r * edge)

		uL := uint16(floatToInt16(outL))
		uR := uint16(floatToInt16(outR))
		off := i * 4
		buf[off+0] = byte(uL)
		buf[off+1] = byte(uL >> 8)
		buf[off+2] = byte(uR)
		buf[off+3] = byte(uR >> 8)
	}
	return buf
}

func buildComboStepSFX(freq float64, dur float64) []byte {
	var ph1, ph2, ph3 float64
	return renderSFX(dur, func(_ int, t, dt, d float64) (float64, float64) {
		ph1 += freq * dt
		ph2 += (freq * 1.004) * dt
		ph3 += (freq * 2.0) * dt
		env := math.Exp(-t * (7.5 / d * 0.14))
		sparkle := math.Exp(-t * 24.0)
		l := (oscTri(ph1)*0.55 + oscSaw(ph2)*0.25 + oscSine(ph3)*0.20*sparkle) * env * 0.72
		r := (oscTri(ph2)*0.55 + oscSaw(ph1)*0.25 + oscSine(ph3)*0.20*sparkle) * env * 0.72
		return l, r
	})
}

func buildArpSFX(notes []float64, stepDur, tailDur float64, wave func(float64) float64, gain float64) []byte {
	totalDur := float64(len(notes))*stepDur + tailDur
	var phL, phR, phSub float64
	return renderSFX(totalDur, func(_ int, t, dt, _ float64) (float64, float64) {
		idx := int(t / stepDur)
		if idx >= len(notes) {
			idx = len(notes) - 1
		}
		localT := t - float64(idx)*stepDur
		freq := notes[idx]
		phL += (freq * 1.003) * dt
		phR += (freq * 0.997) * dt
		phSub += (freq * 0.5) * dt
		env := math.Exp(-localT * 10.0)
		if idx == len(notes)-1 {
			env = math.Exp(-localT * (4.5 / math.Max(tailDur, 0.08)))
		}
		l := (wave(phL)*0.65 + oscSine(phSub)*0.35) * env * gain
		r := (wave(phR)*0.65 + oscSine(phSub)*0.35) * env * gain
		return l, r
	})
}

func buildSFXBank() [sfxCount][]byte {
	var bank [sfxCount][]byte

	// 1. SFXThrow: crisp whoosh + pitch-up blip.
	{
		var ph, lp float64
		rng := sfxRNG(0x13572468)
		bank[SFXThrow] = renderSFX(0.14, func(_ int, t, dt, d float64) (float64, float64) {
			prog := t / d
			freq := 320.0 + 480.0*prog
			ph += freq * dt
			n := rng.next()
			lp += 0.25 * (n - lp)
			whoosh := (n - lp) * math.Sin(math.Pi*prog) * 0.40
			blip := oscTri(ph) * math.Exp(-t*18.0) * 0.48
			return whoosh*0.85 + blip, whoosh*1.05 + blip
		})
	}

	// 2. SFXPerfectDrop: punchy sub-impact + shimmering harmonic chord.
	{
		var phSub, phA, phB, phC, phD float64
		bank[SFXPerfectDrop] = renderSFX(0.32, func(_ int, t, dt, _ float64) (float64, float64) {
			subFreq := 42.0 + 95.0*math.Exp(-t*32.0)
			phSub += subFreq * dt
			phA += 523.25 * dt
			phB += 659.25 * dt
			phC += 783.99 * dt
			phD += 987.77 * dt
			sub := oscSine(phSub) * math.Exp(-t*12.0) * 0.72
			shimmer := (oscSine(phA) + oscSine(phB) + oscSine(phC) + oscTri(phD)*0.7) * 0.16 * math.Exp(-t*7.5)
			return sub + shimmer*1.05, sub + shimmer*0.95
		})
	}

	// 3. SFXHook: energetic cyan lock-on chime.
	{
		var ph1, ph2 float64
		bank[SFXHook] = renderSFX(0.16, func(_ int, t, dt, _ float64) (float64, float64) {
			freq := 660.0
			if t > 0.045 {
				freq = 1320.0
			}
			ph1 += freq * dt
			ph2 += (freq * 1.503) * dt
			env := math.Exp(-t * 14.0)
			s := (oscSine(ph1)*0.65 + oscTri(ph2)*0.35) * env * 0.65
			return s, s * 0.96
		})
	}

	// 4. SFXMinigameHit: bright ping.
	{
		var ph1, ph2 float64
		bank[SFXMinigameHit] = renderSFX(0.13, func(_ int, t, dt, _ float64) (float64, float64) {
			freq := 880.0 + 440.0*(1.0-math.Exp(-t*40.0))
			ph1 += freq * dt
			ph2 += (freq * 2.0) * dt
			env := math.Exp(-t * 20.0)
			s := (oscSine(ph1)*0.72 + oscTri(ph2)*0.28) * env * 0.68
			return s, s
		})
	}

	// 5. SFXMinigameMiss: low buzz.
	{
		var ph1, ph2 float64
		bank[SFXMinigameMiss] = renderSFX(0.20, func(_ int, t, dt, d float64) (float64, float64) {
			freq := 135.0 - 40.0*(t/d)
			ph1 += freq * dt
			ph2 += (freq * 1.05) * dt
			env := math.Exp(-t * 9.0)
			s := (oscSaw(ph1)*0.55 + oscSquare(ph2)*0.30) * env * 0.58
			return s, s
		})
	}

	// 6-9. SFXComboStep1..4: ascending pentatonic steps (C5, E5, G5, C6).
	bank[SFXComboStep1] = buildComboStepSFX(523.25, 0.14)
	bank[SFXComboStep2] = buildComboStepSFX(659.25, 0.14)
	bank[SFXComboStep3] = buildComboStepSFX(783.99, 0.14)
	bank[SFXComboStep4] = buildComboStepSFX(1046.50, 0.22)

	// 10. SFXBusHorn: bold twin-tone retro neon bus horn + rush whoosh.
	{
		var ph1, ph2, lp float64
		rng := sfxRNG(0x55AA33CC)
		bank[SFXBusHorn] = renderSFX(0.38, func(_ int, t, dt, d float64) (float64, float64) {
			ph1 += 220.0 * dt
			ph2 += 277.18 * dt
			// Double-honk cadence: short honk 0..0.09s, main blast 0.12..0.38s.
			hornEnv := 0.0
			if t < 0.09 {
				hornEnv = math.Sin(math.Pi * (t / 0.09))
			} else if t >= 0.12 {
				hornEnv = math.Exp(-(t - 0.12) * 5.0)
			}
			brass := (oscSaw(ph1)*0.52 + oscSaw(ph2)*0.48) * hornEnv * 0.55
			n := rng.next()
			lp += 0.18 * (n - lp)
			whoosh := lp * math.Sin(math.Pi*(t/d)) * 0.35
			return brass + whoosh*0.9, brass*0.95 + whoosh*1.1
		})
	}

	// 11. SFXDash: short filtered air-slice transient.
	{
		var ph, prevN float64
		rng := sfxRNG(0x88776655)
		bank[SFXDash] = renderSFX(0.11, func(_ int, t, dt, d float64) (float64, float64) {
			prog := t / d
			ph += (900.0 - 620.0*prog) * dt
			n := rng.next()
			hp := n - prevN
			prevN = n
			env := math.Exp(-t * 24.0)
			slice := (hp*0.60 + oscTri(ph)*0.30) * env * 0.65
			return slice * 1.05, slice * 0.90
		})
	}

	// 12. SFXPlayerHit: crunchy retro impact + downward pitch drop.
	{
		var ph1, ph2 float64
		rng := sfxRNG(0xCAFEBABE)
		bank[SFXPlayerHit] = renderSFX(0.24, func(_ int, t, dt, _ float64) (float64, float64) {
			freq := 55.0 + 245.0*math.Exp(-t*18.0)
			ph1 += freq * dt
			ph2 += (freq * 0.5) * dt
			crunch := rng.next() * math.Exp(-t*28.0) * 0.50
			thud := (oscSquare(ph1)*0.55 + oscSine(ph2)*0.45) * math.Exp(-t*10.0) * 0.62
			return crunch + thud, crunch*0.9 + thud
		})
	}

	// 13. SFXPoliceSiren: classic alternating hi-lo synth siren chirp.
	{
		var ph float64
		bank[SFXPoliceSiren] = renderSFX(0.34, func(_ int, t, dt, d float64) (float64, float64) {
			// Smooth alternating hi-lo (880 Hz <-> 660 Hz).
			mod := math.Sin(2.0 * math.Pi * (t / 0.17))
			freq := 770.0 + 115.0*mod
			ph += freq * dt
			env := math.Sin(math.Pi * (t / d))
			s := (oscTri(ph)*0.65 + oscSaw(ph)*0.35) * env * 0.52
			return s * 1.04, s * 0.96
		})
	}

	// 14. SFXBossSpawn: ominous sub-bass swell + brassy synth stab.
	{
		var phSub, phA, phB, phC float64
		bank[SFXBossSpawn] = renderSFX(0.52, func(_ int, t, dt, d float64) (float64, float64) {
			subFreq := 55.0 - 12.0*(t/d)
			phSub += subFreq * dt
			phA += 110.0 * dt
			phB += 130.81 * dt // Minor 3rd tension
			phC += 164.81 * dt
			swell := math.Min(1.0, t/0.12) * math.Exp(-t*3.6)
			sub := oscSine(phSub) * swell * 0.60
			brass := (oscSaw(phA)*0.40 + oscSaw(phB)*0.32 + oscSaw(phC)*0.28) * swell * 0.45
			return sub + brass*1.06, sub + brass*0.94
		})
	}

	// 15. SFXBossBreak: glass shatter noise + dramatic chord hit.
	{
		var ph1, ph2, ph3, prevN float64
		rng := sfxRNG(0xDEADBEEF)
		bank[SFXBossBreak] = renderSFX(0.44, func(_ int, t, dt, _ float64) (float64, float64) {
			ph1 += 261.63 * dt
			ph2 += 392.00 * dt
			ph3 += 523.25 * dt
			n := rng.next()
			hp := n - prevN
			prevN = n
			shatter := hp * math.Exp(-t*14.0) * 0.65
			// Sparkling glass shards.
			if int(t*220)%3 == 0 {
				shatter *= 1.25
			}
			chord := (oscSaw(ph1)*0.40 + oscSaw(ph2)*0.32 + oscTri(ph3)*0.35) * math.Exp(-t*5.5) * 0.50
			return shatter*1.05 + chord, shatter*0.92 + chord
		})
	}

	// 16. SFXLevelUp: celebratory rising arcade fanfare arpeggio.
	bank[SFXLevelUp] = buildArpSFX([]float64{523.25, 659.25, 783.99, 1046.50}, 0.075, 0.18, oscTri, 0.68)

	// 17. SFXShopBuy: crisp coin/vending clink (B5 -> E6).
	bank[SFXShopBuy] = buildArpSFX([]float64{987.77, 1318.51}, 0.055, 0.11, oscSine, 0.72)

	// 18. SFXSlotTick: mechanical reel click.
	{
		var ph, prevN float64
		rng := sfxRNG(0x24681357)
		bank[SFXSlotTick] = renderSFX(0.032, func(_ int, t, dt, _ float64) (float64, float64) {
			ph += 1150.0 * dt
			n := rng.next()
			hp := n - prevN
			prevN = n
			env := math.Exp(-t * 140.0)
			s := (oscSine(ph)*0.60 + hp*0.40) * env * 0.65
			return s, s
		})
	}

	// 19. SFXSlotWin: bright coin cascade.
	bank[SFXSlotWin] = buildArpSFX([]float64{659.25, 783.99, 987.77, 1318.51}, 0.06, 0.16, oscSine, 0.70)

	// 20. SFXSlotJackpot: huge 777 fanfare.
	{
		notes := []float64{523.25, 659.25, 783.99, 1046.50, 1318.51, 1567.98}
		stepDur := 0.065
		var phL, phR, phSub, phHi float64
		bank[SFXSlotJackpot] = renderSFX(0.66, func(_ int, t, dt, _ float64) (float64, float64) {
			idx := int(t / stepDur)
			if idx >= len(notes) {
				idx = len(notes) - 1
			}
			freq := notes[idx]
			phL += (freq * 1.004) * dt
			phR += (freq * 0.996) * dt
			phSub += 130.81 * dt
			phHi += 2093.00 * dt
			env := math.Exp(-t * 3.8)
			trem := 0.85 + 0.15*math.Sin(2.0*math.Pi*14.0*t)
			l := (oscSaw(phL)*0.45 + oscSine(phSub)*0.30 + oscSine(phHi)*0.20) * env * trem * 0.72
			r := (oscSaw(phR)*0.45 + oscSine(phSub)*0.30 + oscSine(phHi)*0.20) * env * trem * 0.72
			return l, r
		})
	}

	// 21. SFXSlotBust: sad trombone / wah drop.
	{
		wahNotes := []float64{311.13, 293.66, 277.18, 233.08}
		var ph float64
		bank[SFXSlotBust] = renderSFX(0.50, func(_ int, t, dt, _ float64) (float64, float64) {
			idx := int(t / 0.11)
			if idx >= len(wahNotes) {
				idx = len(wahNotes) - 1
			}
			freq := wahNotes[idx]
			if idx == len(wahNotes)-1 {
				localT := t - 0.33
				freq -= 38.0 * localT
				freq += 6.0 * math.Sin(2.0*math.Pi*6.5*localT)
			}
			ph += freq * dt
			env := math.Exp(-t * 3.2)
			s := (oscSaw(ph)*0.55 + oscTri(ph)*0.35) * env * 0.56
			return s, s
		})
	}

	// 22. SFXUIClick: clean crisp pop.
	{
		var ph float64
		bank[SFXUIClick] = renderSFX(0.045, func(_ int, t, dt, d float64) (float64, float64) {
			freq := 950.0 - 430.0*(t/d)
			ph += freq * dt
			s := oscSine(ph) * math.Exp(-t*75.0) * 0.62
			return s, s
		})
	}

	// 23. SFXChallengeUnlock: triumphant neon trophy chime.
	bank[SFXChallengeUnlock] = buildArpSFX([]float64{783.99, 987.77, 1174.66, 1567.98}, 0.08, 0.22, oscTri, 0.68)

	// 24. SFXShieldBlock: crisp crystalline metallic/glass deflect ping (~0.11s).
	{
		var ph1, ph2, ph3, prevN float64
		rng := sfxRNG(0x61A55B10)
		bank[SFXShieldBlock] = renderSFX(0.11, func(_ int, t, dt, _ float64) (float64, float64) {
			ph1 += 1480.0 * dt
			ph2 += 2217.0 * dt
			ph3 += 3136.0 * dt
			n := rng.next()
			hp := n - prevN
			prevN = n
			ping := (oscSine(ph1)*0.48 + oscSine(ph2)*0.32 + oscTri(ph3)*0.20) * math.Exp(-t*28.0) * 0.68
			tick := hp * math.Exp(-t*95.0) * 0.28
			return ping*1.04 + tick, ping*0.96 + tick
		})
	}

	// 25. SFXShieldBreak: resonant glass shatter burst (~0.24s).
	{
		var ph1, ph2, phSub, prevN float64
		rng := sfxRNG(0x61A55B2E)
		bank[SFXShieldBreak] = renderSFX(0.24, func(_ int, t, dt, d float64) (float64, float64) {
			prog := t / d
			ph1 += (1860.0 - 1140.0*prog) * dt
			ph2 += (2480.0 - 1420.0*prog) * dt
			phSub += (190.0 - 95.0*prog) * dt
			n := rng.next()
			hp := n - prevN
			prevN = n
			shatter := hp * math.Exp(-t*13.0) * 0.66
			if int(t*260)%3 == 0 {
				shatter *= 1.22
			}
			shards := (oscSine(ph1)*0.42 + oscTri(ph2)*0.36) * math.Exp(-t*15.0) * 0.45
			thud := oscSine(phSub) * math.Exp(-t*18.0) * 0.34
			return shatter*1.06 + shards + thud, shatter*0.94 + shards*0.92 + thud
		})
	}

	// 26. SFXKeyPickup: crystalline golden key chime (ascending major triad shimmer).
	bank[SFXKeyPickup] = buildArpSFX([]float64{880.0, 1318.51, 1760.0}, 0.055, 0.16, oscSine, 0.68)

	// 27. SFXChestOpen: crisp mechanical latch click followed by golden synthwave fanfare.
	{
		var ph1, ph2, phSub float64
		notes := [4]float64{523.25, 659.25, 783.99, 1046.50}
		bank[SFXChestOpen] = renderSFX(0.52, func(_ int, t, dt, _ float64) (float64, float64) {
			if t < 0.05 {
				// Crisp latch click
				phSub += 240.0 * dt
				ph1 += 1920.0 * dt
				latch := (oscSine(phSub)*0.45 + oscSquare(ph1)*0.25) * math.Exp(-t*55.0)
				return latch * 0.8, latch * 0.8
			}
			nt := t - 0.05
			idx := int(nt / 0.075)
			if idx >= len(notes) {
				idx = len(notes) - 1
			}
			localT := nt - float64(idx)*0.075
			freq := notes[idx]
			ph1 += (freq * 1.003) * dt
			ph2 += (freq * 0.997) * dt
			phSub += (freq * 0.5) * dt
			env := math.Exp(-localT * 8.0)
			if idx == len(notes)-1 {
				env = math.Exp(-localT * 5.2)
			}
			l := (oscTri(ph1)*0.58 + oscSaw(ph2)*0.22 + oscSine(phSub)*0.20) * env * 0.72
			r := (oscTri(ph2)*0.58 + oscSaw(ph1)*0.22 + oscSine(phSub)*0.20) * env * 0.72
			return l, r
		})
	}

	return bank
}
