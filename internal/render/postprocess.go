package render

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"girlbusAG/assets"
	"girlbusAG/internal/game"
)

// Post-processing visual filter modes selectable at runtime.
const (
	FilterOff     = 0 // Pixel Crisp / Off
	FilterWarmCRT = 1 // Warm CRT (Default)
	FilterVHS     = 2 // VHS Noir
)

// PostProcessor renders the 640x360 virtual scene into an offscreen buffer and
// composites it onto the destination screen through the embedded Kage shader.
type PostProcessor struct {
	sceneBuf     *ebiten.Image
	shader       *ebiten.Shader
	shaderErr    error
	shaderOpts   ebiten.DrawRectShaderOptions
	drawOpts     ebiten.DrawImageOptions
	screenSize   []float32
	chronosBlend float32
}

// NewPostProcessor allocates the 640x360 scene buffer and compiles the embedded
// Kage post-processing shader, falling back gracefully to a crisp blit on error.
func NewPostProcessor() *PostProcessor {
	sceneBuf := ebiten.NewImage(game.ScreenWidth, game.ScreenHeight)
	shader, err := ebiten.NewShader(assets.PostProcessShaderSrc)
	if err != nil {
		log.Printf("render: failed to compile postprocess shader (falling back to crisp blit): %v", err)
		shader = nil
	}

	screenSize := []float32{float32(game.ScreenWidth), float32(game.ScreenHeight)}
	p := &PostProcessor{
		sceneBuf:   sceneBuf,
		shader:     shader,
		shaderErr:  err,
		screenSize: screenSize,
	}
	p.shaderOpts.Images[0] = sceneBuf
	p.shaderOpts.Uniforms = map[string]any{
		"Time":       float32(0),
		"Mode":       float32(FilterWarmCRT),
		"Chronos":    float32(0),
		"Rewind":     float32(0),
		"ScreenSize": screenSize,
	}
	return p
}

// BeginScene clears the offscreen 640x360 scene buffer and returns it as the
// active render target for the frame.
func (p *PostProcessor) BeginScene() *ebiten.Image {
	if p == nil {
		return nil
	}
	if p.sceneBuf == nil {
		p.sceneBuf = ebiten.NewImage(game.ScreenWidth, game.ScreenHeight)
	}
	p.sceneBuf.Clear()
	return p.sceneBuf
}

// Present composites the offscreen scene buffer onto dst using the selected filter
// mode (0=Pixel Crisp, 1=Warm CRT, 2=VHS Noir) and Chronos/Rewind intensities [0..1].
func (p *PostProcessor) Present(dst *ebiten.Image, mode int, tick, chronosAmt, rewindAmt float32) {
	if p == nil || dst == nil || p.sceneBuf == nil {
		return
	}
	if mode < FilterOff || mode > FilterVHS {
		mode = FilterWarmCRT
	}
	chronosTarget := clampF(chronosAmt, 0, 1)
	rewindAmt = clampF(rewindAmt, 0, 1)

	// Smoothly ease Chronos transitions over ~0.12s instead of a hard 0->1 snap.
	const chronosLerp float32 = 0.18
	diff := chronosTarget - p.chronosBlend
	if diff > -0.002 && diff < 0.002 {
		p.chronosBlend = chronosTarget
	} else {
		p.chronosBlend = clampF(p.chronosBlend+diff*chronosLerp, 0, 1)
	}

	if p.shader == nil {
		p.drawOpts.GeoM.Reset()
		p.drawOpts.ColorScale.Reset()
		dst.DrawImage(p.sceneBuf, &p.drawOpts)
		return
	}

	if len(p.screenSize) != 2 {
		p.screenSize = []float32{float32(game.ScreenWidth), float32(game.ScreenHeight)}
	}
	if p.shaderOpts.Uniforms == nil {
		p.shaderOpts.Uniforms = make(map[string]any, 5)
	}
	p.shaderOpts.Images[0] = p.sceneBuf
	p.shaderOpts.Uniforms["Time"] = tick
	p.shaderOpts.Uniforms["Mode"] = float32(mode)
	p.shaderOpts.Uniforms["Chronos"] = p.chronosBlend
	p.shaderOpts.Uniforms["Rewind"] = rewindAmt
	p.shaderOpts.Uniforms["ScreenSize"] = p.screenSize

	dst.DrawRectShader(game.ScreenWidth, game.ScreenHeight, p.shader, &p.shaderOpts)
}

// HasShaderError returns any Kage shader compilation error encountered during setup.
func (p *PostProcessor) HasShaderError() error {
	if p == nil {
		return nil
	}
	return p.shaderErr
}
