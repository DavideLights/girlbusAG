package render

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"girlbusAG/assets"
	"girlbusAG/internal/game"
)

func TestFXParticlePoolCapAndSwapRemove(t *testing.T) {
	r := &Renderer{}

	// Emit more than maxFXParticles to verify hard cap and recycling.
	for i := 0; i < maxFXParticles+120; i++ {
		r.spawnFXParticle(fxParticle{
			x:       float32(i),
			y:       100,
			vx:      10,
			vy:      -10,
			size:    2,
			life:    0.5,
			maxLife: 0.5,
			a:       255,
		})
	}
	if r.fxParticleCount != maxFXParticles {
		t.Fatalf("expected fxParticleCount capped at %d, got %d", maxFXParticles, r.fxParticleCount)
	}

	r.updateFX(0.25)
	if r.fxParticleCount != maxFXParticles {
		t.Fatalf("expected particles alive at t=0.25, got %d", r.fxParticleCount)
	}

	r.updateFX(0.30)
	if r.fxParticleCount != 0 {
		t.Fatalf("expected all expired particles removed after maxLife, got %d", r.fxParticleCount)
	}
}

func TestFXLightFlashLifecycleAndCap(t *testing.T) {
	r := &Renderer{}

	for i := 0; i < maxFXFlashes+10; i++ {
		r.spawnFXFlash(float32(i*10), 200, 120, 0.8, 0.3, color.RGBA{255, 220, 90, 255})
	}
	if r.fxFlashCount != maxFXFlashes {
		t.Fatalf("expected fxFlashCount capped at %d, got %d", maxFXFlashes, r.fxFlashCount)
	}

	r.updateFX(0.15)
	if r.fxFlashCount != maxFXFlashes {
		t.Fatalf("expected flashes still active at half life, got %d", r.fxFlashCount)
	}

	r.updateFX(0.20)
	if r.fxFlashCount != 0 {
		t.Fatalf("expected flashes cleared after expiry, got %d", r.fxFlashCount)
	}
}

func TestEmitWorldEventAndUpdateWorldVFX(t *testing.T) {
	r := &Renderer{}
	w := game.NewWorldWithSeed(42)

	r.EmitWorldEvent(game.WorldEvent{Kind: game.EventPerfectDrop}, w)
	r.EmitWorldEvent(game.WorldEvent{Kind: game.EventBossBreak, Flag: true}, w)
	r.EmitWorldEvent(game.WorldEvent{Kind: game.EventPlayerHit}, w)

	if r.fxParticleCount == 0 {
		t.Fatalf("expected combat events to emit FX particles")
	}
	if r.fxFlashCount < 3 {
		t.Fatalf("expected combat events to spawn at least 3 light flashes, got %d", r.fxFlashCount)
	}
	if r.damageFlashTimer <= 0 {
		t.Fatalf("expected EventPlayerHit to trigger damageFlashTimer")
	}

	// Verify continuous emitters (footsteps, bus sparks, rain scaling at 5* / Blackout)
	w.Player.VX = 180
	w.Police.Stars = 5
	w.Buses = append(w.Buses, game.NeonBus{Active: true, X: 200, Y: 300, Speed: 900})
	for step := 0; step < 30; step++ {
		r.UpdateWorldVFX(w, 0.05)
	}
	if r.rainIntensity <= 0.7 {
		t.Fatalf("expected high neon rain intensity at 5* Wanted, got %.2f", r.rainIntensity)
	}
}

func TestPostProcessorCurvesAndShaderCompilation(t *testing.T) {
	if len(assets.PostProcessShaderSrc) == 0 {
		t.Fatalf("expected embedded PostProcessShaderSrc to be non-empty")
	}
	if _, err := ebiten.NewShader(assets.PostProcessShaderSrc); err != nil {
		t.Fatalf("expected postprocess.kage to compile cleanly, got error: %v", err)
	}

	var p PostProcessor
	p.TriggerImpact(1.5)
	p.SetFever(2.0)
	if p.impactAmt != 1.0 || p.feverTarget != 1.0 {
		t.Fatalf("expected TriggerImpact/SetFever clamped to 1.0, got impact=%v fever=%v", p.impactAmt, p.feverTarget)
	}

	for i := 0; i < 12; i++ {
		p.Advance(0.03)
	}
	if p.impactAmt != 0 {
		t.Fatalf("expected impactAmt to decay to 0 after 0.36s, got %v", p.impactAmt)
	}
	if p.feverBlend <= 0.5 {
		t.Fatalf("expected feverBlend to ease toward 1.0, got %v", p.feverBlend)
	}
}
