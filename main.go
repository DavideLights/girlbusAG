package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"girlbusAG/assets"
	"girlbusAG/internal/game"
	"girlbusAG/internal/input"
	"girlbusAG/internal/render"
)

// App implements ebiten.Game for GIRLBUS: NEON STOP (girlbusAG v2.0).
type App struct {
	world    *game.World
	renderer *render.Renderer
	reader   *input.Reader
}

// NewApp initializes the SpriteBank, Renderer, Input Reader, and deterministic World.
func NewApp() *App {
	sprites := assets.LoadSpriteBank()
	return &App{
		world:    game.NewWorld(),
		renderer: render.NewRenderer(sprites),
		reader:   input.NewReader(),
	}
}

// Update reads unified cross-platform input and advances the game simulation.
func (a *App) Update() error {
	if (a.world.GameOver || a.world.Victory) && input.IsRestartJustPressed() {
		a.world = game.NewWorld()
		return nil
	}

	// When standing inside an active stop with a minigame or combo, Arrow Keys drive the
	// minigame/combo exclusively while WASD continues dodging.
	inComboOrMinigame := false
	for i := range a.world.Stops {
		st := &a.world.Stops[i]
		if st.Active && !st.BusRushTriggered && st.PlayerInside {
			inComboOrMinigame = true
			break
		}
	}

	in := a.reader.Read(a.world.CameraX, a.world.CameraY, int(a.world.Player.SelectedStopArch), inComboOrMinigame)
	a.renderer.SetAim(in.AimWorldX, in.AimWorldY)
	a.world.Update(in, 1.0/60.0)
	return nil
}

// Draw renders the 640x360 virtual screen scaled crisply to the 1280x720 window.
func (a *App) Draw(screen *ebiten.Image) {
	a.renderer.Draw(screen, a.world)
}

// Layout fixes the internal virtual resolution at 640x360 (16:9 pixel-art).
func (a *App) Layout(_, _ int) (int, int) {
	return game.ScreenWidth, game.ScreenHeight
}

func main() {
	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("GIRLBUS: NEON STOP (girlbusAG v2.0)")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(NewApp()); err != nil {
		log.Fatal(err)
	}
}
