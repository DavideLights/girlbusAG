package render

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sort"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"girlbusAG/assets"
	"girlbusAG/internal/game"
)

var (
	colOutline         = color.RGBA{28, 17, 36, 255}
	colEmo             = color.RGBA{157, 101, 194, 255}
	colStudent         = color.RGBA{75, 156, 226, 255}
	colKawaii          = color.RGBA{255, 42, 133, 255}
	colGyaru           = color.RGBA{0, 245, 212, 255}
	colGold            = color.RGBA{255, 230, 0, 255}
	colRedAlert        = color.RGBA{255, 45, 75, 255}
	colGreenNeon       = color.RGBA{46, 235, 130, 255}
	colHudBg           = color.RGBA{16, 10, 28, 224}
	colHoloBg          = color.RGBA{16, 10, 28, 200}
	colHoloTranslucent = Premul(16, 10, 32, 165)
)

type drawKind uint8

const (
	drawProp drawKind = iota
	drawHealPickup
	drawShop
	drawSlot
	drawStop
	drawGirl
	drawOfficer
	drawCruiser
	drawBoss
	drawPlayer
)

type renderItem struct {
	sortY float32
	kind  drawKind
	idx   int
}

// Renderer draws the stratified Pop-Urban district, Y-sorted entities, Golden-Hour shadows,
// dynamic additive lighting, ambient particles, contextual aim reticle, and intuitive Arcade HUD.
type Renderer struct {
	sprites              *assets.SpriteBank
	aimX                 float32
	aimY                 float32
	mouseScreenX         float32
	mouseScreenY         float32
	mouseHoverDir        int
	tick                 float32
	textCache            map[string]*ebiten.Image
	lightTex             *ebiten.Image
	whitePixel           *ebiten.Image
	particles            [numAmbientParticles]ambientParticle
	drawables            []renderItem
	playerSkin           int
	busSkin              int
	dirInputMode         int
	screenShakeEnabled   bool
	showFPS              bool
	playingActive        bool
	lastCamX, lastCamY   float32
	activeChallengeToast string
	toastTimer           float32
	relicToastKind       game.RelicKind
	relicToastName       string
	relicToastDesc       string
	relicToastTimer      float32
}

// NewRenderer constructs a Renderer backed by the embedded SpriteBank.
func NewRenderer(sprites *assets.SpriteBank) *Renderer {
	wp := ebiten.NewImage(1, 1)
	wp.Fill(color.White)
	r := &Renderer{
		sprites:            sprites,
		mouseHoverDir:      game.ActionNone,
		textCache:          make(map[string]*ebiten.Image, 160),
		lightTex:           buildRadialLightTexture(),
		whitePixel:         wp,
		drawables:          make([]renderItem, 0, 160),
		screenShakeEnabled: true,
		playingActive:      true,
	}
	r.initAmbientParticles()
	return r
}

// SetPlayingActive controls whether live-combat reticle, diegetic minigames, and edge indicators render.
func (r *Renderer) SetPlayingActive(active bool) {
	r.playingActive = active
}

// Sprites returns the underlying embedded SpriteBank.
func (r *Renderer) Sprites() *assets.SpriteBank {
	return r.sprites
}

// SetAim updates the current world-space aim cursor for rendering the tactical Bus Stop reticle.
func (r *Renderer) SetAim(worldX, worldY float32) {
	r.aimX = worldX
	r.aimY = worldY
}

// SetMouseScreen updates virtual screen cursor coordinates for HUD/modal hover states.
func (r *Renderer) SetMouseScreen(sx, sy float32) {
	r.mouseScreenX = sx
	r.mouseScreenY = sy
}

// SetMouseHoverDir updates the live 4-way quadrant hover direction (-1 none, 0..3) for Mouse-Only minigames & combos.
func (r *Renderer) SetMouseHoverDir(dir int) {
	r.mouseHoverDir = dir
}

// SetVisualPrefs updates cosmetic skins and visual preference flags from saved settings/profile.
func (r *Renderer) SetVisualPrefs(playerSkin, busSkin int, screenShake, showFPS bool) {
	r.playerSkin = playerSkin
	r.busSkin = busSkin
	r.screenShakeEnabled = screenShake
	r.showFPS = showFPS
}

// SetDirInputMode sets whether directional mouse actions use the locked box pad (0) or around-stop compass (1).
func (r *Renderer) SetDirInputMode(mode int) {
	r.dirInputMode = mode
}

// ShowChallengeToast triggers a temporary top-center challenge unlock banner.
func (r *Renderer) ShowChallengeToast(title string) {
	r.activeChallengeToast = title
	r.toastTimer = 3.6
}

// ShowRelicToast triggers a clear 5s banner explaining a newly acquired passive Relic.
func (r *Renderer) ShowRelicToast(rk game.RelicKind) {
	spec := game.GetRelicSpec(rk)
	r.relicToastKind = rk
	r.relicToastName = spec.Name
	r.relicToastDesc = spec.Desc
	r.relicToastTimer = 5.0
}

// AdvanceVisuals runs ambient animation and notification timers at the update rate.
func (r *Renderer) AdvanceVisuals(dt float32) {
	r.tick += dt
	r.updateAmbientParticles(dt)
	r.toastTimer = max(0, r.toastTimer-dt)
	r.relicToastTimer = max(0, r.relicToastTimer-dt)
}

// archetypeColor returns the signature neon color for a BusStop / Girl archetype.
func archetypeColor(a game.Archetype) color.RGBA {
	switch a {
	case game.ArchEmo:
		return colEmo
	case game.ArchStudent:
		return colStudent
	case game.ArchKawaii:
		return colKawaii
	case game.ArchGyaru:
		return colGyaru
	default:
		return colGyaru
	}
}

// Draw renders the full gameplay frame onto the 640x360 virtual screen.
func (r *Renderer) Draw(screen *ebiten.Image, w *game.World) {
	r.DrawWorldScene(screen, w, true)
}

// DrawWorldScene renders the full Pop-Urban world scene and optionally calls DrawHUDAndOverlays when showHUD is true.
func (r *Renderer) DrawWorldScene(screen *ebiten.Image, w *game.World, showHUD bool) {
	// Compute clamped camera + screen shake offset
	shakeX, shakeY := float32(0), float32(0)
	if r.screenShakeEnabled && w.ScreenShake > 0.05 {
		shakeX = float32(math.Sin(float64(r.tick*73.0))) * w.ScreenShake
		shakeY = float32(math.Cos(float64(r.tick*91.0))) * w.ScreenShake
	}
	camX := clampF(w.CameraX, 0, game.ArenaWidth-game.ScreenWidth) + shakeX
	camY := clampF(w.CameraY, 0, game.ArenaHeight-game.ScreenHeight) + shakeY
	r.lastCamX = camX
	r.lastCamY = camY

	// Clear backdrop
	screen.Fill(color.RGBA{26, 20, 38, 255})

	// Layer 0: Stratified Pop-Urban Street Arena, Parks, Bus Bays, ZTL, Boardwalk & Canal Water
	r.drawArenaTiles(screen, w, camX, camY)

	// Layer 0.5: Architectural Golden-Hour Drop Shadows
	r.drawGroundShadows(screen, w, camX, camY)
	r.drawDistrictLighting(screen, w, camX, camY)

	// Lay markings and movement effects on the ground before sorting the world props.
	r.drawExpressLanes(screen, w, camX, camY)
	r.drawBusSlipstreams(screen, w, camX, camY)
	r.drawWaveRings(screen, w, camX, camY)
	r.drawStopCirclesAndTethers(screen, w, camX, camY)
	r.drawPoliceHazards(screen, w, camX, camY)
	r.drawAmbientParticles(screen, camX, camY)

	r.drawYSortedWorld(screen, w, camX, camY)
	r.drawLightingPass(screen, w, camX, camY)
	r.drawBuses(screen, w, camX, camY)
	r.drawHeliSpotlight(screen, w, camX, camY)

	if r.playingActive && !w.PendingLevelUp && !w.ShopMenu.Active && !w.SlotModal.Active {
		r.drawAimAndRange(screen, w, camX, camY)
	}

	// Keep dangers legible over the scenery, light pools and targeting fills.
	r.drawTelegraphs(screen, w, camX, camY)
	r.drawBullets(screen, w, camX, camY)
	r.drawPlayerHealthBar(screen, w, camX, camY)

	if r.playingActive {
		r.drawPopups(screen, w, camX, camY)
		r.drawOffscreenIndicators(screen, w, camX, camY)
	}

	// Layer 7: Chronos Vignette / Atmosphere
	r.drawAtmosphereOverlays(screen, w)

	// Layer 8: Crisp Arcade HUD & Modals (only when showHUD is true)
	if showHUD {
		r.DrawHUDAndOverlays(screen, w)
	}
}

func (r *Renderer) drawArenaTiles(screen *ebiten.Image, w *game.World, camX, camY float32) {
	tileAsphalt := r.sprites.Get(assets.SpriteTileAsphalt)
	tileWetAsphalt := r.sprites.Get(assets.SpriteTileWetAsphalt)
	tileSidewalk := r.sprites.Get(assets.SpriteTileSidewalk)
	tileCross := r.sprites.Get(assets.SpriteTileCrosswalk)
	tileGrass := r.sprites.Get(assets.SpriteTileGrassPark)
	tileParkPath := r.sprites.Get(assets.SpriteTileParkPath)
	tileCobble := r.sprites.Get(assets.SpriteTileCobblePlaza)
	tilePlazaMosaic := r.sprites.Get(assets.SpriteTilePlazaMosaic)
	tileBikeLane := r.sprites.Get(assets.SpriteTileBikeLane)
	tileBoardwalk := r.sprites.Get(assets.SpriteTileBoardwalk)
	tileCanalEdge := r.sprites.Get(assets.SpriteTileCanalEdge)
	tileWater := r.sprites.Get(assets.SpriteTileCanalWater)

	const tileSize = 32
	startCol := int(math.Floor(float64(camX) / tileSize))
	endCol := int(math.Ceil(float64(camX+game.ScreenWidth) / tileSize))
	startRow := int(math.Floor(float64(camY) / tileSize))
	endRow := int(math.Ceil(float64(camY+game.ScreenHeight) / tileSize))

	if startCol < 0 {
		startCol = 0
	}
	if startRow < 0 {
		startRow = 0
	}
	maxCols := (game.ArenaWidth + tileSize - 1) / tileSize
	maxRows := (game.ArenaHeight + tileSize - 1) / tileSize
	if endCol > maxCols {
		endCol = maxCols
	}
	if endRow > maxRows {
		endRow = maxRows
	}

	op := &ebiten.DrawImageOptions{}

	// 1. Base Grid Tiles (Sidewalks, Central Mosaic & Cobble Plaza, Dry/Wet Asphalt, Bike Lane, Crosswalks, Boardwalk, Canal)
	for row := startRow; row < endRow; row++ {
		for col := startCol; col < endCol; col++ {
			wx := float32(col * tileSize)
			wy := float32(row * tileSize)
			sx := float64(wx - camX)
			sy := float64(wy - camY)

			img := tileAsphalt
			isCentralPlazaCol := col >= 17 && col <= 22 // X = 544..736
			isPlazaMosaicInner := col >= 18 && col <= 21 &&
				((row >= 1 && row <= 4) || (row >= 9 && row <= 12) || (row >= 17 && row <= 18))
			isAlleyCol := (col == 8 || col == 9 || col == 30 || col == 31) && wy < 144

			switch {
			case wy >= w.CanalY-8:
				img = tileWater
			case wy >= w.BoardwalkY:
				img = tileBoardwalk
			case isCentralPlazaCol:
				img = tileCobble
				if isPlazaMosaicInner {
					img = tilePlazaMosaic
				}
				for _, cr := range w.CrosswalkRows {
					if row == cr {
						img = tileCross
						break
					}
				}
			case wy < 144 || wy >= 544:
				img = tileSidewalk
				if isAlleyCol {
					img = tileCobble
				}
			default:
				isCrossCol := false
				for _, cc := range w.CrosswalkCols {
					if col == cc {
						isCrossCol = true
						break
					}
				}
				switch {
				case isCrossCol:
					img = tileCross
				case row == 16:
					// Dedicated Bike Lane transition strip (Y = 512..544) along south avenue curb
					img = tileBikeLane
				default:
					// Blend glossy wet evening asphalt near central neon intersections, gutters, and subtle puddles
					isWetBand := col >= 12 && col <= 27 && ((col+row)%2 == 0 || row == 5 || row == 15)
					isGutterWet := (row == 5 || row == 15) && (col%3 != 1)
					isPuddle := (col*7+row*13)%9 == 0
					if isWetBand || isGutterWet || isPuddle {
						img = tileWetAsphalt
					}
				}
			}

			op.GeoM.Reset()
			op.ColorScale.Reset()
			if img == tileCross {
				op.ColorScale.Scale(0.79, 0.76, 0.79, 1)
			}
			op.GeoM.Translate(sx, sy)
			screen.DrawImage(img, op)
		}
	}
	op.ColorScale.Reset()

	// 1.5 Seamless South Boardwalk (Y=624..656) & Stone Canal Embankment (Y=656..688) above Canal Water
	bwY := w.BoardwalkY - camY
	if bwY < game.ScreenHeight && bwY+tileSize > 0 {
		for col := startCol; col < endCol; col++ {
			sx := float64(float32(col*tileSize) - camX)
			op.GeoM.Reset()
			op.GeoM.Translate(sx, float64(bwY))
			screen.DrawImage(tileBoardwalk, op)
		}
	}
	canalEdgeY := w.CanalY - 24 - camY
	if canalEdgeY < game.ScreenHeight && canalEdgeY+tileSize > 0 {
		for col := startCol; col < endCol; col++ {
			sx := float64(float32(col*tileSize) - camX)
			op.GeoM.Reset()
			op.GeoM.Translate(sx, float64(canalEdgeY))
			screen.DrawImage(tileCanalEdge, op)
		}
	}

	// 2. North & South Sidewalk Curbs (at Y=144 and Y=544)
	northCurbY := float32(144) - camY
	southCurbY := float32(544) - camY
	// Break curbs gently across Central Cobblestone Plaza (544..736)
	drawCurbSegment := func(y float32) {
		vector.DrawFilledRect(screen, -camX, y-2, 544, 4, color.RGBA{226, 218, 198, 255}, false)
		vector.DrawFilledRect(screen, -camX, y+2, 544, 1.5, color.RGBA{46, 36, 58, 130}, false)
		vector.DrawFilledRect(screen, 736-camX, y-2, game.ArenaWidth-736, 4, color.RGBA{226, 218, 198, 255}, false)
		vector.DrawFilledRect(screen, 736-camX, y+2, game.ArenaWidth-736, 1.5, color.RGBA{46, 36, 58, 130}, false)
	}
	drawCurbSegment(northCurbY)
	drawCurbSegment(southCurbY)

	// 3. Avenue Center Yellow Dashed Line (broken around Central Plaza 520..760)
	centerY := float32(344) - camY
	for x := 24; x < game.ArenaWidth-24; x += 48 {
		if x >= 500 && x <= 760 {
			continue
		}
		vector.DrawFilledRect(screen, float32(x)-camX, centerY-1.5, 24, 3, color.RGBA{255, 212, 64, 185}, false)
	}

	// Central Plaza decorative stone medallion
	plazaCX, plazaCY := float32(640)-camX, float32(344)-camY
	vector.StrokeCircle(screen, plazaCX, plazaCY, 42, 2.0, color.RGBA{228, 214, 194, 145}, true)
	vector.StrokeCircle(screen, plazaCX, plazaCY, 24, 1.5, color.RGBA{0, 245, 212, 95}, true)

	// 4. West & East Urban Pocket-Park Islands (Cobble Border + Lush Grass + Stepping-Stone Park Path)
	for _, pk := range w.ParkIslands {
		px, py := pk.X-camX, pk.Y-camY
		if px+pk.W < -32 || px > game.ScreenWidth+32 || py+pk.H < -32 || py > game.ScreenHeight+32 {
			continue
		}
		// Cobblestone perimeter border
		cols := int(math.Ceil(float64(pk.W) / tileSize))
		rows := int(math.Ceil(float64(pk.H) / tileSize))
		for rIdx := 0; rIdx < rows; rIdx++ {
			for cIdx := 0; cIdx < cols; cIdx++ {
				tx := px + float32(cIdx*tileSize)
				ty := py + float32(rIdx*tileSize)
				op.GeoM.Reset()
				op.GeoM.Translate(float64(tx), float64(ty))
				screen.DrawImage(tileCobble, op)
			}
		}
		// Lush grass interior inset by 16px with clean stepping-stone ParkPath walkway through the center
		gx0, gy0 := pk.X+16, pk.Y+16
		gw, gh := pk.W-32, pk.H-32
		gCols := int(math.Ceil(float64(gw) / tileSize))
		gRows := int(math.Ceil(float64(gh) / tileSize))
		for rIdx := 0; rIdx < gRows; rIdx++ {
			for cIdx := 0; cIdx < gCols; cIdx++ {
				tx := gx0 + float32(cIdx*tileSize) - camX
				ty := gy0 + float32(rIdx*tileSize) - camY
				op.GeoM.Reset()
				op.GeoM.Translate(float64(tx), float64(ty))
				tileImg := tileGrass
				// Central stepping-stone walkway across the park heart
				if rIdx == 1 && (cIdx == 1 || cIdx == 2) {
					tileImg = tileParkPath
				}
				screen.DrawImage(tileImg, op)
			}
		}
		// Centered horizontal stepping-stone path connecting west/east park entries cleanly
		pathY := py + pk.H*0.5 - 16
		for pCol := 0; pCol < 4; pCol++ {
			op.GeoM.Reset()
			op.GeoM.Translate(float64(gx0-camX+float32(pCol*tileSize)), float64(pathY))
			screen.DrawImage(tileParkPath, op)
		}
		// Crisp stone border rims
		vector.StrokeRect(screen, px, py, pk.W, pk.H, 2.0, color.RGBA{230, 220, 200, 220}, false)
		vector.StrokeRect(screen, gx0-camX, gy0-camY, gw, gh, 1.5, color.RGBA{72, 130, 68, 220}, false)
	}

	// 5. Designated Yellow BUS Bays (pulsing green/gold border)
	bayPulse := float32(0.5 + 0.5*math.Sin(float64(r.tick*4.5)))
	for _, bay := range w.BusBays {
		bx, by := bay.X-camX, bay.Y-camY
		if bx+bay.W < -20 || bx > game.ScreenWidth+20 || by+bay.H < -20 || by > game.ScreenHeight+20 {
			continue
		}
		vector.DrawFilledRect(screen, bx, by, bay.W, bay.H, Premul(255, 210, 40, 38), false)
		bayBorder := Premul(
			uint8(46+195*bayPulse),
			uint8(230-20*bayPulse),
			uint8(130-90*bayPulse),
			230,
		)
		vector.StrokeRect(screen, bx, by, bay.W, bay.H, 2.0, bayBorder, false)
		// Corner chevrons
		vector.DrawFilledRect(screen, bx+2, by+2, 8, 3, colGold, false)
		vector.DrawFilledRect(screen, bx+bay.W-10, by+bay.H-5, 8, 3, colGold, false)
		r.drawScaledText(screen, "AREA BUS", int(bx+20), int(by+11), 1.0, colGold, colOutline)
	}

	// Restricted bays use the same coral edge language as the police warnings.
	for _, z := range w.ZTLZones {
		zx, zy := snapPixel(z.X-camX), snapPixel(z.Y-camY)
		if zx+z.W < 0 || zx >= game.ScreenWidth || zy+z.H < 0 || zy >= game.ScreenHeight {
			continue
		}
		vector.DrawFilledRect(screen, zx, zy, z.W, z.H, Premul(145, 43, 70, 72), false)
		vector.StrokeRect(screen, zx+0.5, zy+0.5, z.W-1, z.H-1, 1, PremulColor(actionCoral, 205), false)
		for x := zx + 3; x < zx+z.W-6; x += 10 {
			vector.StrokeLine(screen, x, zy+5, x+4, zy+1, 1, PremulColor(actionCoral, 155), false)
			vector.StrokeLine(screen, x, zy+z.H-2, x+4, zy+z.H-6, 1, PremulColor(actionCoral, 155), false)
		}
		vector.DrawFilledRect(screen, zx+5, zy+9, 73, 15, Premul(38, 24, 43, 220), false)
		r.drawTextColored(screen, "ZTL 3x +1", int(zx)+9, int(zy)+9, actionCream)
		r.DrawSpriteIcon(screen, assets.SpriteIconStarWanted, zx+64, zy+10, 1.0)
	}

	// La Rettrice Martial ZTL Half-Arena Overlay
	if w.Boss.Active && w.Boss.MartialZTLActive {
		mz := w.Boss.MartialZTLRect
		zx, zy := mz.X-camX, mz.Y-camY
		vector.DrawFilledRect(screen, zx, zy, mz.W, mz.H, Premul(255, 35, 85, 32), false)
		vector.StrokeRect(screen, zx, zy, mz.W, mz.H, 2.5, Premul(255, 45, 95, 220), false)
		r.drawScaledText(screen, "ZTL MARZIALE (NO PERFECT = +1", int(zx+14), int(zy+10), 1.0, colRedAlert, colOutline)
		r.DrawSpriteIcon(screen, assets.SpriteIconStarWanted, zx+190, zy+11, 1.0)
		r.drawScaledText(screen, ")", int(zx+203), int(zy+10), 1.0, colRedAlert, colOutline)
	}

	// 7. South Boardwalk Wooden Railing & Shimmering Canal Reflections
	bwTopY := w.BoardwalkY - camY
	vector.DrawFilledRect(screen, -camX, bwTopY-2, game.ArenaWidth, 3, Premul(112, 74, 44, 240), false)

	railY := w.CanalY - 6 - camY
	vector.DrawFilledRect(screen, -camX, railY, game.ArenaWidth, 3, color.RGBA{196, 146, 94, 255}, false)
	vector.DrawFilledRect(screen, -camX, railY+3, game.ArenaWidth, 2, color.RGBA{74, 44, 28, 255}, false)
	for rx := 16; rx < game.ArenaWidth; rx += 32 {
		px := float32(rx) - camX
		vector.DrawFilledRect(screen, px-2, railY-5, 5, 11, color.RGBA{164, 112, 68, 255}, false)
		vector.StrokeRect(screen, px-2, railY-5, 5, 11, 1, colOutline, false)
	}

	// Animated canal water reflections along Southern Canal
	waterTopY := w.CanalY - camY
	vector.DrawFilledRect(screen, -camX, waterTopY, game.ArenaWidth, 3, Premul(22, 76, 112, 200), false)
	for i := 0; i < 18; i++ {
		wx := float32(i*74) + float32(math.Sin(float64(r.tick*1.8+float32(i))))*14.0 - camX
		wy := waterTopY + 8 + float32((i*11)%24)
		wLen := float32(18 + (i%3)*10)
		alpha := uint8(95 + 65*math.Sin(float64(r.tick*3.1+float32(i)*1.3)))
		vector.DrawFilledRect(screen, wx, wy, wLen, 2, Premul(185, 245, 255, alpha), false)
	}
}

func (r *Renderer) drawExpressLanes(screen *ebiten.Image, w *game.World, camX, camY float32) {
	for _, e := range w.Express {
		x1, y1 := e.X1-camX, e.Y1-camY
		x2, y2 := e.X2-camX, e.Y2-camY
		vector.StrokeLine(screen, x1, y1, x2, y2, 10, Premul(0, 245, 212, 42), true)
		vector.StrokeLine(screen, x1, y1, x2, y2, 2.5, Premul(0, 245, 212, 200), true)
	}
}

func (r *Renderer) drawWaveRings(screen *ebiten.Image, w *game.World, camX, camY float32) {
	for _, wr := range w.WaveRings {
		if wr.MaxLife <= 0 || wr.Radius <= 0 {
			continue
		}
		alpha := clampF(wr.Life/wr.MaxLife, 0, 1)
		var c color.RGBA
		var impact bool
		switch wr.Kind {
		case 0:
			c = color.RGBA{100, 230, 198, 255}
		case 1:
			c = color.RGBA{255, 211, 131, 255}
			impact = wr.MaxRadius >= 120
		case 2:
			c = color.RGBA{255, 132, 160, 255}
			impact = true
		case 3:
			c = color.RGBA{108, 185, 250, 255}
		case game.WaveBossBreak:
			c = color.RGBA{255, 137, 115, 255}
			impact = true
		default:
			c = color.RGBA{247, 221, 121, 255}
		}
		x, y := wr.X-camX, wr.Y-camY
		vector.StrokeCircle(screen, x, y, wr.Radius, 1.5, PremulColor(c, uint8(190*alpha)), true)
		if impact {
			drawImpactSparks(screen, x, y, wr.Radius, alpha, c, wr.Kind == game.WaveBossBreak)
		}
	}
}

func (r *Renderer) drawStopCirclesAndTethers(screen *ebiten.Image, w *game.World, camX, camY float32) {
	for i := range w.Stops {
		st := &w.Stops[i]
		if !st.Active {
			continue
		}
		sx, sy := st.X-camX, st.Y-camY
		ac := archetypeColor(st.Arch)

		fillAlpha := uint8(13)
		if st.PlayerInside {
			fillAlpha = 22
		}
		vector.DrawFilledCircle(screen, sx, sy, st.CaptureRadius, PremulColor(ac, fillAlpha), true)

		ringC := PremulColor(ac, 176)
		if st.SynergyActive {
			ringC = colGold
		}
		vector.StrokeCircle(screen, sx, sy, st.CaptureRadius, 1.5, ringC, true)

		// Kawaii Tear-Graze ring around Alfredo
		if st.Arch == game.ArchKawaii && st.PlayerInside {
			px, py := w.Player.X-camX, w.Player.Y-camY
			vector.StrokeCircle(screen, px, py, st.GrazeRadius(), 1.5, Premul(255, 160, 210, 190), true)
		}

		// Tether lines to hooked Girls
		for _, gid := range st.HookedGirlIDs {
			for g := range w.Girls {
				if w.Girls[g].ID == gid {
					gx, gy := w.Girls[g].X-camX, w.Girls[g].Y-camY
					drawTether(screen, sx, sy, gx, gy, r.tick, ac)
					break
				}
			}
		}

		// Tether line to hooked Boss
		if st.HookedBoss && w.Boss.Active {
			bx, by := w.Boss.X-camX, w.Boss.Y-camY
			drawTether(screen, sx, sy, bx, by, r.tick, colGold)
		}
	}
}

func (r *Renderer) drawTelegraphs(screen *ebiten.Image, w *game.World, camX, camY float32) {
	for _, t := range w.Telegraphs {
		x1, y1 := t.X1-camX, t.Y1-camY
		x2, y2 := t.X2-camX, t.Y2-camY
		progress := 1 - clampF(t.Timer/max(t.MaxTimer, 0.01), 0, 1)
		drawHazardWarning(screen, x1, y1, x2, y2, progress, t.Kind == 0)
	}
}

func (r *Renderer) drawPoliceHazards(screen *ebiten.Image, w *game.World, camX, camY float32) {
	for _, h := range w.Police.Hazards {
		if !h.Active {
			continue
		}
		x1, y1 := h.X1-camX, h.Y1-camY
		x2, y2 := h.X2-camX, h.Y2-camY
		if x2 < -32 || x1 > game.ScreenWidth+32 || y1 < -32 || y1 > game.ScreenHeight+32 {
			continue
		}

		if h.TelegraphTimer > 0 {
			pulse := float32(0.5 + 0.5*math.Sin(float64(r.tick*22.0)))
			bandA := uint8(55 + 65*pulse)
			vector.StrokeLine(screen, x1, y1, x2, y2, 14, Premul(255, 55, 45, bandA), false)
			segLen := float32(8)
			gapLen := float32(6)
			for sx := x1; sx < x2; sx += segLen + gapLen {
				ex := sx + segLen
				if ex > x2 {
					ex = x2
				}
				c := colGold
				if int((sx+r.tick*60)/12)%2 == 0 {
					c = colRedAlert
				}
				vector.StrokeLine(screen, sx, y1, ex, y2, 2.2, c, false)
			}
			continue
		}

		// Active electrified red/yellow laser-spike strip
		flicker := int(r.tick*26)%2 == 0
		vector.StrokeLine(screen, x1, y1, x2, y2, 10, Premul(255, 35, 65, 95), false)
		vector.StrokeLine(screen, x1, y1, x2, y2, 4.5, Premul(255, 45, 75, 225), false)
		coreCol := colGold
		if flicker {
			coreCol = color.RGBA{255, 250, 195, 255}
		}
		vector.StrokeLine(screen, x1, y1, x2, y2, 1.8, coreCol, false)

		// Electrified spikes along the barrier strip
		for sx := x1 + 6; sx < x2-4; sx += 11 {
			spikeCol := colGold
			if int(sx)%2 == 0 {
				spikeCol = colRedAlert
			}
			vector.StrokeLine(screen, sx-3, y1+2, sx, y1-5, 1.4, spikeCol, true)
			vector.StrokeLine(screen, sx+3, y1+2, sx, y1-5, 1.4, spikeCol, true)
		}

		// Heavy metallic road anchor pylons at endpoints
		for _, ax := range [2]float32{x1, x2} {
			vector.DrawFilledRect(screen, ax-3, y1-6, 6, 12, color.RGBA{28, 22, 42, 245}, false)
			vector.StrokeRect(screen, ax-3, y1-6, 6, 12, 1.0, colRedAlert, false)
			vector.DrawFilledRect(screen, ax-1.5, y1-4, 3, 3, colGold, false)
		}

		// Tiny "DASH!" prompt when Alfredo is nearby
		nearX := clampF(w.Player.X, h.X1, h.X2)
		dx := w.Player.X - nearX
		dy := w.Player.Y - h.Y1
		if dx*dx+dy*dy <= 68*68 {
			midX := (x1 + x2) * 0.5
			lbl := "DASH!"
			pw := float32(textPixelWidth(lbl, 1.0) + 10)
			px := clampF(midX-pw*0.5, 4, game.ScreenWidth-pw-4)
			py := clampF(y1-22, 28, game.ScreenHeight-36)
			drawPillBadge(screen, px, py, pw, 14, Premul(48, 10, 22, 230), colGold)
			r.drawScaledText(screen, lbl, int(px+5), int(py-1), 1.0, colGold, colOutline)
		}
	}
}

func (r *Renderer) drawYSortedWorld(screen *ebiten.Image, w *game.World, camX, camY float32) {
	r.drawables = r.drawables[:0]

	for i := range w.Props {
		pr := &w.Props[i]
		px, py := pr.X-camX, pr.Y-camY
		if px+pr.W < -48 || px > game.ScreenWidth+48 || py+pr.H < -48 || py > game.ScreenHeight+48 {
			continue
		}
		r.drawables = append(r.drawables, renderItem{sortY: pr.SortY, kind: drawProp, idx: i})
	}
	for i := range w.HealPickups {
		hp := &w.HealPickups[i]
		hx, hy := float32(hp.X)-camX, float32(hp.Y)-camY
		if hx >= -32 && hx <= game.ScreenWidth+32 && hy >= -32 && hy <= game.ScreenHeight+32 {
			r.drawables = append(r.drawables, renderItem{sortY: float32(hp.Y) + 8, kind: drawHealPickup, idx: i})
		}
	}
	for i := range w.Shops {
		r.drawables = append(r.drawables, renderItem{sortY: w.Shops[i].Y + 14, kind: drawShop, idx: i})
	}
	for i := range w.SlotMachines {
		r.drawables = append(r.drawables, renderItem{sortY: w.SlotMachines[i].Y + 14, kind: drawSlot, idx: i})
	}
	for i := range w.Stops {
		if w.Stops[i].Active {
			r.drawables = append(r.drawables, renderItem{sortY: w.Stops[i].Y + 16, kind: drawStop, idx: i})
		}
	}
	for i := range w.Girls {
		gx, gy := w.Girls[i].X-camX, w.Girls[i].Y-camY
		if gx >= -40 && gx <= game.ScreenWidth+40 && gy >= -40 && gy <= game.ScreenHeight+40 {
			r.drawables = append(r.drawables, renderItem{sortY: w.Girls[i].Y + 12, kind: drawGirl, idx: i})
		}
	}
	for i := range w.Police.Officers {
		if w.Police.Officers[i].Active {
			r.drawables = append(r.drawables, renderItem{sortY: w.Police.Officers[i].Y + 12, kind: drawOfficer, idx: i})
		}
	}
	for i := range w.Police.Cruisers {
		if w.Police.Cruisers[i].Active && w.Police.Cruisers[i].Dashing {
			r.drawables = append(r.drawables, renderItem{sortY: w.Police.Cruisers[i].Y + 14, kind: drawCruiser, idx: i})
		}
	}
	if w.Boss.Active {
		r.drawables = append(r.drawables, renderItem{sortY: w.Boss.Y + 15, kind: drawBoss, idx: 0})
	}
	r.drawables = append(r.drawables, renderItem{sortY: w.Player.Y + 12, kind: drawPlayer, idx: 0})

	sort.SliceStable(r.drawables, func(i, j int) bool {
		return r.drawables[i].sortY < r.drawables[j].sortY
	})

	for _, item := range r.drawables {
		switch item.kind {
		case drawProp:
			r.drawSingleProp(screen, &w.Props[item.idx], camX, camY)
		case drawHealPickup:
			r.drawSingleHealPickup(screen, &w.HealPickups[item.idx], camX, camY)
		case drawShop:
			r.drawSingleShop(screen, w, &w.Shops[item.idx], camX, camY)
		case drawSlot:
			r.drawSingleSlotMachine(screen, w, &w.SlotMachines[item.idx], camX, camY)
		case drawStop:
			r.drawSingleStop(screen, &w.Stops[item.idx], camX, camY)
		case drawGirl:
			r.drawSingleGirl(screen, &w.Girls[item.idx], camX, camY)
		case drawOfficer:
			r.drawSingleOfficer(screen, &w.Police.Officers[item.idx], camX, camY)
		case drawCruiser:
			r.drawSingleCruiser(screen, &w.Police.Cruisers[item.idx], camX, camY)
		case drawBoss:
			r.drawBossEntity(screen, w, camX, camY)
		case drawPlayer:
			r.drawPlayerEntity(screen, w, camX, camY)
		}
	}
}

func propSpriteKey(kind game.PropKind) string {
	switch kind {
	case game.PropBuildingKonbini:
		return assets.SpriteBuildingKonbini
	case game.PropBuildingCoffee:
		return assets.SpriteBuildingCoffee
	case game.PropBuildingArcade:
		return assets.SpriteBuildingArcade
	case game.PropBuildingRamen:
		return assets.SpriteBuildingRamen
	case game.PropBuildingBookstore:
		return assets.SpriteBuildingBookstore
	case game.PropBuildingMetro:
		return assets.SpriteBuildingMetro
	case game.PropBuildingNewsstand:
		return assets.SpriteBuildingNewsstand
	case game.PropBuildingPizzeria:
		return assets.SpriteBuildingPizzeria
	case game.PropBuildingVinylBar:
		return assets.SpriteBuildingVinylBar
	case game.PropBuildingGelateria:
		return assets.SpriteBuildingGelateria
	case game.PropFountain:
		return assets.SpritePropFountain
	case game.PropArchGate:
		return assets.SpritePropArchGate
	case game.PropTreeLush:
		return assets.SpritePropTreeLush
	case game.PropStreetlamp:
		return assets.SpritePropStreetlamp
	case game.PropBenchPlants:
		return assets.SpritePropBenchPlants
	case game.PropCatsClutter:
		return assets.SpritePropCatsClutter
	case game.PropScooter:
		return assets.SpritePropScooter
	case game.PropBikeRack:
		return assets.SpritePropBikeRack
	case game.PropHydrantBin:
		return assets.SpritePropHydrantBin
	case game.PropZTLSign:
		return assets.SpritePropZTLSign
	case game.PropCanalBoat:
		return assets.SpritePropCanalBoat
	case game.PropSlotMachine:
		return assets.SpritePropSlotMachine
	case game.PropManhole:
		return assets.SpritePropManhole
	case game.PropAsphaltCrack:
		return assets.SpritePropAsphaltCrack
	case game.PropFlowerBox:
		return assets.SpritePropFlowerBox
	case game.PropNeonBillboard:
		return assets.SpritePropNeonBillboard
	case game.PropPhoneBooth:
		return assets.SpritePropPhoneBooth
	case game.PropFoodCart:
		return assets.SpritePropFoodCart
	case game.PropKioskUmbrella:
		return assets.SpritePropKioskUmbrella
	case game.PropTrafficLight:
		return assets.SpritePropTrafficLight
	case game.PropBollardGlow:
		return assets.SpritePropBollardGlow
	case game.PropPlanterNeon:
		return assets.SpritePropPlanterNeon
	case game.PropVendingDrink:
		return assets.SpritePropVendingDrink
	case game.PropBusShelterDeco:
		return assets.SpritePropBusShelterDeco
	case game.PropCarRed:
		return assets.SpriteCarRed
	case game.PropCarYellow:
		return assets.SpriteCarYellow
	case game.PropCarMint:
		return assets.SpriteCarMint
	case game.PropCarVanCyber:
		return assets.SpriteCarVanCyber
	default:
		return ""
	}
}

func (r *Renderer) drawSingleProp(screen *ebiten.Image, pr *game.MapProp, camX, camY float32) {
	key := propSpriteKey(pr.Kind)
	if key == "" {
		return
	}

	img := r.sprites.Get(key)
	iw := float32(img.Bounds().Dx())
	ih := float32(img.Bounds().Dy())

	cx := pr.X + pr.W*0.5 - camX
	by := pr.SortY - camY
	dx := cx - iw*0.5
	dy := by - ih
	switch pr.Kind {
	case game.PropBuildingKonbini, game.PropBuildingCoffee, game.PropBuildingArcade,
		game.PropBuildingRamen, game.PropBuildingBookstore, game.PropBuildingMetro,
		game.PropBuildingPizzeria, game.PropBuildingVinylBar, game.PropBuildingGelateria,
		game.PropManhole, game.PropAsphaltCrack, game.PropCanalBoat:
		dx = pr.X - camX
		dy = pr.Y - camY
	}

	op := &ebiten.DrawImageOptions{}
	if pr.Kind == game.PropCanalBoat {
		bob := math.Sin(float64(r.tick*2.1+pr.X*0.03)) * 1.5
		dy += float32(bob)
	}
	op.GeoM.Translate(float64(dx), float64(dy))
	screen.DrawImage(img, op)

	// Subtle animated water-jet sparkle pixels above the Central Plaza Fountain
	if pr.Kind == game.PropFountain {
		jetX := dx + iw*0.5
		jetY := dy + 14
		for d := 0; d < 6; d++ {
			phase := frac32(r.tick*1.8 + float32(d)*0.17)
			side := float32(-1)
			if d%2 == 1 {
				side = 1
			}
			spread := float32(4+d%3*3) * side * phase
			arcY := -11.0*phase*(1.0-phase)*4.0 + phase*9.0
			alpha := uint8(230 * math.Sin(float64(phase)*math.Pi))
			vector.DrawFilledRect(screen, jetX+spread-0.75, jetY+arcY-0.75, 1.8, 1.8, Premul(215, 252, 255, alpha), false)
		}
		// Shimmering basin ripple glints
		rip := float32(math.Sin(float64(r.tick * 4.6)))
		vector.DrawFilledRect(screen, jetX-12+rip*3, dy+28, 5, 1.2, Premul(185, 248, 255, 185), false)
		vector.DrawFilledRect(screen, jetX+7-rip*3, dy+31, 6, 1.2, Premul(220, 255, 255, 165), false)
	}
}

func (r *Renderer) drawSingleHealPickup(screen *ebiten.Image, hp *game.HealPickup, camX, camY float32) {
	hx, hy := float32(hp.X)-camX, float32(hp.Y)-camY
	if hp.LifeTimer > 0 && hp.LifeTimer < 3.0 && int(r.tick*14)%2 == 0 {
		// Gentle rapid blink warning when about to despawn
		return
	}

	pulse := float32(0.5 + 0.5*math.Sin(float64(r.tick*5.5)+hp.BobPhase))
	bob := float32(math.Sin(float64(r.tick*4.4)+hp.BobPhase)) * 2.5

	scale := 1.0
	shadowR := float32(7.0)
	ringR := float32(10.0 + 2.2*pulse)
	if hp.Big {
		scale = 1.2
		shadowR = 9.0
		ringR = 13.0 + 2.8*pulse
	}

	// Ground contact shadow + soft pulsing ground ring
	vector.DrawFilledCircle(screen, hx, hy+7, shadowR, Premul(20, 12, 32, 115), true)
	if hp.BlackHeart {
		vector.StrokeCircle(screen, hx, hy+6, ringR, 1.8, Premul(160, 38, 210, uint8(155+90*pulse)), true)
		vector.StrokeCircle(screen, hx, hy+6, ringR*0.68, 1.3, Premul(245, 42, 108, uint8(165+85*pulse)), true)

		// Pulsing dark-violet/crimson heart with white sheen highlight
		hScale := float32(scale) * (1.15 + 0.12*pulse)
		hyHeart := hy - 3 + bob
		vector.DrawFilledCircle(screen, hx, hyHeart, 8.5+1.5*pulse, Premul(110, 16, 145, uint8(85+55*pulse)), true)
		heartImg := r.sprites.Get(assets.SpriteIconHeartFull)
		hbw := float64(heartImg.Bounds().Dx())
		hbh := float64(heartImg.Bounds().Dy())
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-hbw*0.5, -hbh*0.5)
		op.GeoM.Scale(float64(hScale), float64(hScale))
		op.GeoM.Translate(float64(hx), float64(hyHeart))
		op.ColorScale.Scale(0.62, 0.14, 0.72, 1.0)
		screen.DrawImage(heartImg, op)
		// Crisp inner dark-violet core & white specular sheen
		drawHeartIcon(screen, hx, hyHeart, Premul(140, 24, 175, 215), true)
		vector.DrawFilledRect(screen, hx-3.6, hyHeart-2.8, 2.2, 1.3, Premul(255, 240, 255, 245), false)
		vector.DrawFilledRect(screen, hx-4.0, hyHeart-1.6, 1.2, 1.4, Premul(255, 240, 255, 220), false)

		lbl := "CUORE +1"
		lw := textPixelWidth(lbl, 1.0)
		r.drawScaledText(screen, lbl, int(hx)-lw/2, int(hyHeart-18), 1.0, colKawaii, colOutline)
	} else {
		vector.StrokeCircle(screen, hx, hy+6, ringR, 1.6, Premul(92, 255, 196, uint8(135+85*pulse)), true)
		vector.StrokeCircle(screen, hx, hy+6, ringR*0.68, 1.2, Premul(255, 232, 108, uint8(150+90*pulse)), true)

		// Gently bobbing pickup sprite
		img := r.sprites.Get(assets.SpriteIconPickupHeal)
		bw := float64(img.Bounds().Dx())
		bh := float64(img.Bounds().Dy())
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-bw*0.5, -bh*0.5)
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(float64(hx), float64(hy-2+bob))
		screen.DrawImage(img, op)
	}

	// Crisp white sparkle orbiting the cure pickup
	ang := float64(r.tick*3.4) + hp.BobPhase
	spX := hx + float32(math.Cos(ang))*(ringR*0.75)
	spY := hy - 3 + bob + float32(math.Sin(ang))*4.0
	spA := uint8(165 + 90*pulse)
	vector.DrawFilledRect(screen, spX-2.0, spY-0.5, 4.0, 1.1, Premul(255, 255, 255, spA), false)
	vector.DrawFilledRect(screen, spX-0.5, spY-2.0, 1.1, 4.0, Premul(255, 252, 210, spA), false)
}

func (r *Renderer) drawMouseLMBIcon(dst *ebiten.Image, x, y float32, scale float64) {
	img := r.sprites.Get(assets.SpriteIconMouseLMB)
	op := &ebiten.DrawImageOptions{}
	if scale > 0 && scale != 1.0 {
		op.GeoM.Scale(scale, scale)
	}
	op.GeoM.Translate(float64(x), float64(y))
	dst.DrawImage(img, op)
}

func (r *Renderer) drawSingleShop(screen *ebiten.Image, w *game.World, sh *game.VendingShop, camX, camY float32) {
	img := r.sprites.Get(assets.SpriteShopVending)
	op := &ebiten.DrawImageOptions{}
	bw := float64(img.Bounds().Dx())
	bh := float64(img.Bounds().Dy())
	sx, sy := sh.X-camX, sh.Y-camY

	vector.DrawFilledCircle(screen, sx, sy+12, 12, color.RGBA{20, 12, 32, 110}, true)
	op.GeoM.Translate(-bw*0.5, -bh*0.5)
	op.GeoM.Translate(float64(sx), float64(sy))
	screen.DrawImage(img, op)

	dx := w.Player.X - sh.X
	dy := w.Player.Y - sh.Y
	if dx*dx+dy*dy <= sh.Radius*sh.Radius {
		lbl := "[E / CLICK SX] DISTRIBUTORE (CURE, TANGENTE, RELIQUIE)"
		pw := float32(textPixelWidth(lbl, 1.0) + 26)
		px := clampF(sx-pw*0.5, 4, game.ScreenWidth-pw-4)
		py := clampF(sy-40, 30, game.ScreenHeight-24)
		drawPillBadge(screen, px, py, pw, 20, colHudBg, colGyaru)
		r.drawMouseLMBIcon(screen, px+3, py+3, 0.88)
		r.drawTextColored(screen, lbl, int(px+20), int(py+2), colGyaru)
	}
}

func (r *Renderer) drawSingleSlotMachine(screen *ebiten.Image, w *game.World, sm *game.SlotMachine, camX, camY float32) {
	img := r.sprites.Get(assets.SpritePropSlotMachine)
	op := &ebiten.DrawImageOptions{}
	bw := float64(img.Bounds().Dx())
	bh := float64(img.Bounds().Dy())
	sx, sy := sm.X-camX, sm.Y-camY

	vector.DrawFilledCircle(screen, sx, sy+12, 13, color.RGBA{20, 12, 32, 115}, true)
	op.GeoM.Translate(-bw*0.5, -bh*0.5)
	op.GeoM.Translate(float64(sx), float64(sy))
	screen.DrawImage(img, op)

	if int(r.tick*6)%2 == 0 {
		vector.DrawFilledCircle(screen, sx, sy-16, 2.5, colGold, true)
	}

	dx := w.Player.X - sm.X
	dy := w.Player.Y - sm.Y
	if dx*dx+dy*dy <= sm.Radius*sm.Radius {
		lbl := fmt.Sprintf("[E / CLICK SX] SLOT 777 ($%d)", sm.SpinCost)
		pw := float32(textPixelWidth(lbl, 1.0) + 26)
		px := clampF(sx-pw*0.5, 4, game.ScreenWidth-pw-4)
		py := clampF(sy-42, 30, game.ScreenHeight-24)
		drawPillBadge(screen, px, py, pw, 20, colHudBg, colGold)
		r.drawMouseLMBIcon(screen, px+3, py+3, 0.88)
		r.drawTextColored(screen, lbl, int(px+20), int(py+2), colGold)
	}
}

func (r *Renderer) drawSingleStop(screen *ebiten.Image, st *game.BusStop, camX, camY float32) {
	var key string
	switch st.Arch {
	case game.ArchEmo:
		key = assets.SpriteStopEmo
	case game.ArchStudent:
		key = assets.SpriteStopStudent
	case game.ArchKawaii:
		key = assets.SpriteStopKawaii
	default:
		key = assets.SpriteStopGyaru
	}
	img := r.sprites.Get(key)
	bw := float64(img.Bounds().Dx())
	bh := float64(img.Bounds().Dy())
	sx, sy := snapPixel(st.X-camX), snapPixel(st.Y-camY)
	drop := 1 - clampF(st.DeployAge/0.16, 0, 1)
	dropHeight := snapPixel(12 * drop * drop)
	accent := archetypeColor(st.Arch)
	if st.PerfectDrop {
		accent = colGold
	}

	drawContactShadow(screen, sx, sy+16, 36)
	drawStopDeployment(screen, sx, sy, dropHeight, st.DeployAge, accent)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-bw*0.5, -bh*0.5)
	op.GeoM.Translate(float64(sx), float64(sy-dropHeight))
	screen.DrawImage(img, op)

	// Back-glass shield pips above shelter
	sy -= dropHeight
	if st.MaxShieldHP > 0 {
		pipW := float32(6)
		totalW := float32(st.MaxShieldHP)*(pipW+2) - 2
		startX := sx - totalW*0.5
		for p := 0; p < st.MaxShieldHP; p++ {
			px := startX + float32(p)*(pipW+2)
			c := Premul(24, 28, 48, 210)
			if p < st.ShieldHP {
				c = colGyaru
				if st.ShieldFlashTimer > 0 {
					c = color.RGBA{215, 255, 255, 255}
				}
			}
			vector.DrawFilledRect(screen, px, sy-28, pipW, 4.0, c, false)
			if p < st.ShieldHP {
				vector.DrawFilledRect(screen, px+1, sy-27.5, pipW-2, 1.2, Premul(230, 255, 255, 225), false)
			}
			vector.StrokeRect(screen, px, sy-28, pipW, 4.0, 0.8, colOutline, false)
		}
	}

	// Crisp cyan-white parry flash arc across top shelter edge on shield hit
	if st.ShieldFlashTimer > 0 {
		flashA := clampF(st.ShieldFlashTimer/0.30, 0, 1)
		cyanCol := Premul(0, 245, 212, uint8(220*flashA))
		whiteCol := Premul(240, 255, 255, uint8(250*flashA))
		const segs = 10
		const arcR = float32(22)
		for s := 0; s < segs; s++ {
			a1 := -math.Pi*0.85 + float64(s)*(math.Pi*0.70/segs)
			a2 := -math.Pi*0.85 + float64(s+1)*(math.Pi*0.70/segs)
			ax1 := sx + float32(math.Cos(a1))*arcR
			ay1 := sy - 6 + float32(math.Sin(a1))*(arcR*0.65)
			ax2 := sx + float32(math.Cos(a2))*arcR
			ay2 := sy - 6 + float32(math.Sin(a2))*(arcR*0.65)
			vector.StrokeLine(screen, ax1, ay1, ax2, ay2, 3.2, cyanCol, true)
			vector.StrokeLine(screen, ax1, ay1, ax2, ay2, 1.4, whiteCol, true)
		}
	}

	// Life timer bar under shelter if minigame is active
	if !st.Combo.Active && st.MaxLifeTimer > 0 {
		frac := clampF(st.LifeTimer/st.MaxLifeTimer, 0, 1)
		vector.DrawFilledRect(screen, sx-18, sy+24, 36, 4, color.RGBA{20, 10, 30, 210}, false)
		vector.DrawFilledRect(screen, sx-18, sy+24, 36*frac, 4, archetypeColor(st.Arch), false)
	}
}

func (r *Renderer) drawSingleGirl(screen *ebiten.Image, g *game.Girl, camX, camY float32) {
	gx, gy := g.X-camX, g.Y-camY
	op := &ebiten.DrawImageOptions{}

	// Ground shadow
	drawContactShadow(screen, gx, gy+12, 18)

	if g.IsVIP {
		vipAura := r.sprites.Get(assets.SpriteVIPAura)
		pulse := 1.0 + 0.12*float64(math.Sin(float64(r.tick*7.0)))
		aw := float64(vipAura.Bounds().Dx())
		ah := float64(vipAura.Bounds().Dy())
		op.GeoM.Reset()
		op.GeoM.Translate(-aw*0.5, -ah*0.5)
		op.GeoM.Scale(pulse, pulse)
		op.GeoM.Translate(float64(gx), float64(gy))
		screen.DrawImage(vipAura, op)
	}

	var sheetKey string
	switch g.Arch {
	case game.ArchEmo:
		sheetKey = assets.SpriteGirlEmoSheet
	case game.ArchStudent:
		sheetKey = assets.SpriteGirlStudentSheet
	case game.ArchKawaii:
		sheetKey = assets.SpriteGirlKawaiiSheet
	default:
		sheetKey = assets.SpriteGirlGyaruSheet
	}
	frame, lift := 0, float32(0)
	if g.VX*g.VX+g.VY*g.VY > 64 {
		phase := int(g.AnimTimer*9) % 4
		frame = [4]int{1, 0, 2, 0}[phase]
		if phase%2 == 1 {
			lift = 1
		}
	} else if int(g.AnimTimer*3+float32(g.ID))%7 == 0 {
		lift = 1
	}
	img := r.sprites.GetSheetFrame(sheetKey, frame, 32, 32)
	bw := float64(img.Bounds().Dx())
	bh := float64(img.Bounds().Dy())

	op.GeoM.Reset()
	op.GeoM.Translate(-bw*0.5, -bh*0.5)
	if g.VX < -4 {
		op.GeoM.Scale(-1, 1)
	}
	op.GeoM.Translate(float64(snapPixel(gx)), float64(snapPixel(gy)-lift))
	if g.EnragedTimer > 0 {
		op.ColorScale.Scale(1.3, 0.6, 0.6, 1.0)
	}
	screen.DrawImage(img, op)

	if g.StunTimer > 0 {
		orbit := float32(math.Sin(float64(g.AnimTimer*8))) * 5
		vector.DrawFilledRect(screen, snapPixel(gx-orbit)-1, gy-18, 2, 2, colGold, false)
		vector.DrawFilledRect(screen, snapPixel(gx+orbit)-1, gy-21, 2, 2, actionCream, false)
	}
}

func (r *Renderer) drawSingleOfficer(screen *ebiten.Image, o *game.PoliceOfficer, camX, camY float32) {
	officerImg := r.sprites.Get(assets.SpritePoliceOfficer)
	ox, oy := o.X-camX, o.Y-camY
	vector.DrawFilledCircle(screen, ox, oy+12, 7, color.RGBA{26, 16, 37, 115}, true)

	bw := float64(officerImg.Bounds().Dx())
	bh := float64(officerImg.Bounds().Dy())
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-bw*0.5, -bh*0.5)
	if o.VX < -4 {
		op.GeoM.Scale(-1, 1)
	}
	op.GeoM.Translate(float64(ox), float64(oy))
	if o.IsSWAT {
		op.ColorScale.Scale(0.75, 0.85, 1.35, 1.0)
	}
	screen.DrawImage(officerImg, op)
}

func (r *Renderer) drawSingleCruiser(screen *ebiten.Image, c *game.PoliceCruiser, camX, camY float32) {
	carImg := r.sprites.Get(assets.SpritePoliceCar)
	cx, cy := c.X-camX, c.Y-camY
	bw := float64(carImg.Bounds().Dx())
	bh := float64(carImg.Bounds().Dy())

	vector.DrawFilledRect(screen, cx-22, cy+4, 48, 16, colGoldenHourShadow, false)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-bw*0.5, -bh*0.5)
	if c.VX < 0 {
		op.GeoM.Scale(-1, 1)
	}
	op.GeoM.Translate(float64(cx), float64(cy))
	screen.DrawImage(carImg, op)

	sirenCol := colRedAlert
	if int(r.tick*14)%2 == 0 {
		sirenCol = color.RGBA{60, 150, 255, 255}
	}
	vector.DrawFilledCircle(screen, cx, cy-10, 4, sirenCol, true)
}

func (r *Renderer) drawBossEntity(screen *ebiten.Image, w *game.World, camX, camY float32) {
	var key string
	switch w.Boss.Kind {
	case game.BossBestFriend:
		key = assets.SpriteBossBestFriend
	case game.BossLover:
		key = assets.SpriteBossLover
	case game.BossProfessor:
		key = assets.SpriteBossProfessor
	case game.BossExecutive:
		key = assets.SpriteBossExecutive
	case game.BossRector:
		key = assets.SpriteBossRector
	default:
		key = assets.SpriteBossMom
	}
	img := r.sprites.Get(key)
	bw := float64(img.Bounds().Dx())
	bh := float64(img.Bounds().Dy())
	bx, by := w.Boss.X-camX, w.Boss.Y-camY

	// L'Amante holographic clones
	if w.Boss.Kind == game.BossLover {
		opClone := &ebiten.DrawImageOptions{}
		for _, cl := range w.Boss.Clones {
			if !cl.Active {
				continue
			}
			cx, cy := cl.X-camX, cl.Y-camY
			vector.StrokeCircle(screen, cx, cy, 20, 1.2, color.RGBA{0, 245, 212, 160}, true)
			opClone.GeoM.Reset()
			opClone.GeoM.Translate(-bw*0.5, -bh*0.5)
			opClone.GeoM.Translate(float64(cx), float64(cy))
			opClone.ColorScale.Reset()
			opClone.ColorScale.Scale(0.70, 0.95, 1.35, 0.72)
			screen.DrawImage(img, opClone)
		}
	}

	// La Dirigente Aziendale orbiting intern escorts
	if w.Boss.Kind == game.BossExecutive {
		escImg := r.sprites.Get(assets.SpriteGirlStudent0)
		ew := float64(escImg.Bounds().Dx())
		eh := float64(escImg.Bounds().Dy())
		opEsc := &ebiten.DrawImageOptions{}
		for _, es := range w.Boss.Escorts {
			if !es.Active {
				continue
			}
			ex, ey := es.X-camX, es.Y-camY
			vector.StrokeLine(screen, bx, by, ex, ey, 1.5, color.RGBA{255, 220, 60, 180}, true)
			vector.DrawFilledCircle(screen, ex, ey+10, 6, color.RGBA{26, 16, 37, 110}, true)
			opEsc.GeoM.Reset()
			opEsc.GeoM.Translate(-ew*0.5, -eh*0.5)
			opEsc.GeoM.Scale(0.85, 0.85)
			opEsc.GeoM.Translate(float64(ex), float64(ey))
			opEsc.ColorScale.Reset()
			opEsc.ColorScale.Scale(1.0, 0.95, 0.80, 1.0)
			screen.DrawImage(escImg, opEsc)
			vector.StrokeCircle(screen, ex, ey, 12, 1.2, colGold, true)
		}
	}

	drawContactShadow(screen, bx, by+15, 25)
	phaseGlow := clampF(w.Boss.EnrageFlashTimer/0.8, 0, 1)
	vector.StrokeCircle(screen, bx, by, 22, 1.5+phaseGlow, PremulColor(actionCoral, uint8(182+68*phaseGlow)), true)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-bw*0.5, -bh*0.5)
	lift := float32(math.Round(1 + math.Sin(float64(w.Boss.PhaseTimer*8))))
	op.GeoM.Translate(float64(snapPixel(bx)), float64(snapPixel(by)-lift))
	if phaseGlow > 0 {
		op.ColorScale.Scale(1+0.25*phaseGlow, 1-0.22*phaseGlow, 1-0.12*phaseGlow, 1)
	}
	screen.DrawImage(img, op)

	// 3-Shield Authority bar above Boss head
	segW := float32(12)
	totalW := float32(w.Boss.MaxShields)*(segW+2) - 2
	sx := bx - totalW*0.5
	sy := by - 26
	for i := 0; i < w.Boss.MaxShields; i++ {
		c := color.RGBA{45, 30, 60, 210}
		if i < w.Boss.Shields {
			c = colKawaii
		}
		vector.DrawFilledRect(screen, sx+float32(i)*(segW+2), sy, segW, 5, c, false)
		vector.StrokeRect(screen, sx+float32(i)*(segW+2), sy, segW, 5, 1, colOutline, false)
	}
}

func (r *Renderer) drawPlayerEntity(screen *ebiten.Image, w *game.World, camX, camY float32) {
	p := &w.Player
	dashImg := r.sprites.Get(assets.PlayerSkinSpriteKey(r.playerSkin, "dash"))

	op := &ebiten.DrawImageOptions{}
	bw := float64(dashImg.Bounds().Dx())
	bh := float64(dashImg.Bounds().Dy())
	trailColor := [...]color.RGBA{actionTeal, {255, 132, 198, 255}, {255, 216, 115, 255}, {243, 174, 103, 255}, actionCoral}
	trail := trailColor[0]
	if r.playerSkin > 0 && r.playerSkin < len(trailColor) {
		trail = trailColor[r.playerSkin]
	}
	for i := len(p.Afterimages) - 1; i >= 0; i-- {
		af := p.Afterimages[i]
		if af.Alpha <= 0.02 {
			continue
		}
		op.GeoM.Reset()
		op.GeoM.Translate(-bw*0.5, -bh*0.5)
		if af.FacingX < 0 {
			op.GeoM.Scale(-1, 1)
		}
		op.GeoM.Translate(float64(snapPixel(af.X-camX)), float64(snapPixel(af.Y-camY)))
		op.ColorScale.Reset()
		op.ColorScale.ScaleWithColor(PremulColor(trail, uint8(168*af.Alpha)))
		screen.DrawImage(dashImg, op)
	}
	op.ColorScale.Reset()

	px, py := p.X-camX, p.Y-camY
	var pose string
	facing, lift := p.FacingX, float32(0)
	if p.DashTimer > 0 {
		drawDashStreak(screen, px, py, p.DashDirX, p.DashDirY, clampF(p.DashTimer/0.16, 0, 1), trail)
	}
	anim := p.AnimState
	if p.ThrowAnimTimer > 0 && p.DashTimer <= 0 {
		anim = game.AnimThrow
	}
	switch anim {
	case game.AnimDash:
		pose = "dash"
		if p.DashDirX != 0 {
			facing = p.DashDirX
		}
	case game.AnimThrow:
		pose = "throw"
		if p.ThrowFacingX != 0 {
			facing = p.ThrowFacingX
		}
		if p.ThrowAnimTimer > 0.15 {
			lift = 1
		}
	case game.AnimRun:
		step := int(p.AnimTimer*12) % 4
		pose = [4]string{"run_0", "run_1", "run_2", "run_3"}[step]
		if step%2 == 1 {
			lift = 1
		}
	default:
		if int(p.AnimTimer*3)%2 == 0 {
			pose = "idle_0"
		} else {
			pose = "idle_1"
		}
	}

	img := r.sprites.Get(assets.PlayerSkinSpriteKey(r.playerSkin, pose))
	drawContactShadow(screen, px, py+12, 18)

	op.GeoM.Reset()
	op.GeoM.Translate(-float64(img.Bounds().Dx())*0.5, -float64(img.Bounds().Dy())*0.5)
	if facing < 0 {
		op.GeoM.Scale(-1, 1)
	}
	op.GeoM.Translate(float64(snapPixel(px)), float64(snapPixel(py)-lift))
	op.ColorScale.Reset()
	if p.InvulnTimer > 0 && p.DashTimer <= 0 {
		pulse := float32(0.5 + 0.5*math.Sin(float64(r.tick*35)))
		op.ColorScale.ScaleAlpha(0.55 + 0.45*pulse)
	}
	screen.DrawImage(img, op)

	// Tiny honest hitbox dot in bullet-hell center
	vector.DrawFilledCircle(screen, px, py+2, 2.2, colGyaru, true)
}

func (r *Renderer) drawBusSlipstreams(screen *ebiten.Image, w *game.World, camX, camY float32) {
	for _, sl := range w.Slipstreams {
		if sl.Life <= 0 || sl.MaxLife <= 0 {
			continue
		}
		alpha := clampF(sl.Life/sl.MaxLife, 0, 1)
		x1, x2 := max(0, sl.X1-camX), min(game.ScreenWidth, sl.X2-camX)
		y := snapPixel(sl.Y - camY)
		if x2-x1 <= 2 || y < -20 || y > game.ScreenHeight+20 {
			continue
		}
		vector.DrawFilledRect(screen, x1, y-16, x2-x1, 32, PremulColor(actionTeal, uint8(31*alpha)), false)
		for _, edge := range [...]float32{y - 16, y + 15} {
			vector.DrawFilledRect(screen, x1, edge, x2-x1, 1, PremulColor(actionTeal, uint8(114*alpha)), false)
		}
		// Advance the lane dashes in world space so they remain stable as the camera moves.
		phase := float32(math.Mod(float64(r.tick*140-camX), 48))
		for x := x1 + phase - 48; x < x2; x += 48 {
			start, end := max(x1, snapPixel(x)), min(x2, snapPixel(x)+13)
			if end > start {
				vector.DrawFilledRect(screen, start, y, end-start, 1, PremulColor(actionCream, uint8(103*alpha)), false)
			}
		}
	}
}

func (r *Renderer) drawBuses(screen *ebiten.Image, w *game.World, camX, camY float32) {
	busKey := assets.SpriteBusNeon
	trail, tail := actionTeal, actionCoral
	if r.busSkin == 1 {
		busKey = assets.SpriteBusGold
		trail = color.RGBA{255, 210, 133, 255}
		tail = color.RGBA{244, 130, 60, 255}
	}
	busImg := r.sprites.Get(busKey)
	bw := float64(busImg.Bounds().Dx())
	bh := float64(busImg.Bounds().Dy())
	op := &ebiten.DrawImageOptions{}

	for _, b := range w.Buses {
		if !b.Active {
			continue
		}
		bx, by := snapPixel(b.X-camX), snapPixel(b.Y-camY)
		if bx < -115 || bx > game.ScreenWidth+188 || by < -27 || by > game.ScreenHeight+27 {
			continue
		}
		drawContactShadow(screen, bx, by+17, 81)
		r.drawAdditiveBeam(screen, bx+55, by+8, 64, 12, actionCream, 0.24)
		r.drawAdditiveBeam(screen, bx-59, by+11, 75, 10, tail, 0.17)
		for i := 0; i < 5; i++ {
			y := by - 10 + float32(i)*5
			phase := float32(math.Mod(float64(b.X*0.8+float32(i*23)), 50))
			end := bx - 42 - snapPixel(phase)*0.7
			vector.DrawFilledRect(screen, end-18-float32(i)*3, y, 18+float32(i)*3, 1, PremulColor(trail, uint8(70+i*23)), false)
		}

		op.GeoM.Reset()
		op.GeoM.Translate(-bw*0.5, -bh*0.5)
		lift := math.Round(0.5 + math.Sin(float64(b.X*0.06))*0.5)
		op.GeoM.Translate(float64(bx), float64(by)-lift)
		screen.DrawImage(busImg, op)
	}
}

func (r *Renderer) drawBullets(screen *ebiten.Image, w *game.World, camX, camY float32) {
	op := &ebiten.DrawImageOptions{}
	for i := 0; i < w.Bullets.Count; i++ {
		b := &w.Bullets.Items[i]
		bx, by := b.X-camX, b.Y-camY
		if bx < -20 || bx > game.ScreenWidth+20 || by < -20 || by > game.ScreenHeight+20 {
			continue
		}

		var key string
		switch b.Kind {
		case game.BulletRazor:
			key = assets.SpriteBulletRazor
		case game.BulletProtractor:
			key = assets.SpriteBulletProtractor
		case game.BulletTear:
			key = assets.SpriteBulletTear
		case game.BulletFlash:
			key = assets.SpriteBulletFlash
		case game.BulletPolice:
			key = assets.SpriteBulletPolice
		case game.BulletCiabatta:
			key = assets.SpriteBulletCiabatta
		case game.BulletLipstick:
			key = assets.SpriteBulletLipstick
		case game.BulletContract:
			key = assets.SpriteBulletContract
		case game.BulletSeal:
			key = assets.SpriteBulletSeal
		default:
			key = assets.SpriteBulletPolice
		}
		img := r.sprites.Get(key)
		bw := float64(img.Bounds().Dx())
		bh := float64(img.Bounds().Dy())

		op.GeoM.Reset()
		op.GeoM.Translate(-bw*0.5, -bh*0.5)
		if b.Kind == game.BulletCiabatta || b.Kind == game.BulletRazor || b.Kind == game.BulletLipstick || b.Kind == game.BulletContract {
			op.GeoM.Rotate(float64(b.Age * 9.0))
		}
		if b.Kind == game.BulletCiabatta {
			op.GeoM.Scale(1.25, 1.25)
		}
		op.GeoM.Translate(float64(bx), float64(by))
		screen.DrawImage(img, op)
	}
}

func (r *Renderer) drawHeliSpotlight(screen *ebiten.Image, w *game.World, camX, camY float32) {
	if w.Police.Stars < 5 || !w.Police.Heli.Active {
		return
	}
	hx, hy := w.Police.Heli.X-camX, w.Police.Heli.Y-camY
	vector.DrawFilledCircle(screen, hx, hy, w.Police.Heli.Radius, Premul(255, 250, 180, 48), true)
	vector.StrokeCircle(screen, hx, hy, w.Police.Heli.Radius, 2.0, Premul(255, 50, 80, 210), true)
	r.drawScaledText(screen, "ELICOTTERO 5*", int(hx)-38, int(hy)-int(w.Police.Heli.Radius)-14, 1.0, colRedAlert, colOutline)
}

// drawPlayerHealthBar renders a sleek diegetic HP capsule (heart badge + segmented pips)
// and contextual cyan Chronos micro-bar right beneath Alfredo's feet.
func (r *Renderer) drawPlayerHealthBar(screen *ebiten.Image, w *game.World, camX, camY float32) {
	p := &w.Player
	// Skip when Alfredo is offscreen (e.g., Title screen backdrop where Player.X < 0)
	if p.X < 0 || p.Y < 0 || p.MaxHP <= 0 {
		return
	}
	px, py := p.X-camX, p.Y-camY
	if px < -40 || px > game.ScreenWidth+40 || py < -40 || py > game.ScreenHeight+40 {
		return
	}

	maxHP := p.MaxHP
	if maxHP > 12 {
		maxHP = 12
	}
	pipsAreaW := float32(34)
	if maxHP > 6 {
		pipsAreaW = float32(34 + (maxHP-6)*4)
		if pipsAreaW > 46 {
			pipsAreaW = 46
		}
	}
	const (
		leftPad  = float32(11)
		rightPad = float32(3)
		capH     = float32(7)
		pipGap   = float32(1.2)
	)
	capW := leftPad + pipsAreaW + rightPad
	barX := clampF(px-capW*0.5, 4, game.ScreenWidth-capW-4)
	barY := clampF(py+14, 4, game.ScreenHeight-14)

	hitFlash := p.InvulnTimer > 0.85
	healFlash := p.HealFlashTimer > 0
	lowHP := p.HP <= 2
	urgentPulse := float32(0.5 + 0.5*math.Sin(float64(r.tick*12.0)))
	healPulse := float32(0.5 + 0.5*math.Sin(float64(r.tick*16.0)))

	// Sleek dark-glass capsule backdrop & status-reactive border
	borderCol := Premul(145, 130, 185, 200)
	if hitFlash {
		borderCol = color.RGBA{255, 255, 255, 255}
	} else if healFlash {
		borderCol = color.RGBA{78, 255, 186, 255}
	} else if lowHP {
		borderCol = Premul(255, 45, 75, uint8(180+75*urgentPulse))
	}

	vector.DrawFilledRect(screen, barX+1, barY+1, capW, capH, Premul(8, 4, 16, 175), false)
	vector.DrawFilledRect(screen, barX, barY, capW, capH, Premul(14, 10, 26, 220), false)
	vector.StrokeRect(screen, barX, barY, capW, capH, 1.0, borderCol, false)

	// Tiny heart badge on left edge (icon_heart_full.png scaled 0.75x -> 9x9)
	heartImg := r.sprites.Get(assets.SpriteIconHeartFull)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(0.75, 0.75)
	op.GeoM.Translate(float64(barX+1), float64(barY-1))
	if healFlash {
		op.ColorScale.Scale(0.75, 1.35, 1.0, 1.0)
	} else if hitFlash {
		op.ColorScale.Scale(1.4, 1.4, 1.4, 1.0)
	}
	screen.DrawImage(heartImg, op)

	// Segmented health pips
	totalGap := pipGap * float32(maxHP-1)
	pipW := (pipsAreaW - totalGap) / float32(maxHP)
	if pipW < 2.2 {
		pipW = 2.2
	}
	pipH := float32(4.2)
	pipY := barY + 1.4
	for i := 0; i < maxHP; i++ {
		x0 := barX + leftPad + float32(i)*(pipW+pipGap)
		if i < p.HP {
			pipCol := color.RGBA{255, 64, 108, 255} // Glowing ruby/coral
			if hitFlash {
				pipCol = color.RGBA{255, 255, 255, 255}
			} else if healFlash {
				pipCol = color.RGBA{
					R: uint8(72 + 55*healPulse),
					G: 255,
					B: uint8(176 + 45*healPulse),
					A: 255,
				}
			} else if lowHP {
				pipCol = color.RGBA{
					R: 255,
					G: uint8(28 + 56*urgentPulse),
					B: uint8(52 + 34*urgentPulse),
					A: 255,
				}
			}
			vector.DrawFilledRect(screen, x0, pipY, pipW, pipH, pipCol, false)
			// Crisp glossy top highlight on active pip
			vector.DrawFilledRect(screen, x0, pipY, pipW, 1.1, Premul(255, 224, 236, 145), false)
		} else {
			// Dim plum for lost HP
			vector.DrawFilledRect(screen, x0, pipY, pipW, pipH, Premul(52, 32, 66, 205), false)
		}
	}

	// Sleek 2px cyan Chronos micro-bar right beneath HP capsule ONLY when active or recharging
	if p.ChronosActive || p.ChronosMeter < p.ChronosMax-1.0 {
		chrFrac := clampF(p.ChronosMeter/p.ChronosMax, 0, 1)
		cX := barX + 2
		cY := barY + capH + 1.5
		cW := capW - 4
		vector.DrawFilledRect(screen, cX, cY, cW, 2.2, Premul(14, 20, 38, 205), false)
		chrCol := colGyaru
		if p.ChronosActive {
			chrCol = color.RGBA{135, 255, 240, 255}
		}
		vector.DrawFilledRect(screen, cX, cY, cW*chrFrac, 2.2, chrCol, false)
	}

	r.drawPlayerStatusIndicators(screen, w, px, py)
}

func (r *Renderer) drawPlayerStatusIndicators(screen *ebiten.Image, w *game.World, px, py float32) {
	p := &w.Player
	pulse := float32(0.5 + 0.5*math.Sin(float64(r.tick*14.0)))

	// 3 small spinning gold stars orbiting Alfredo's head when stunned
	if p.StunTimer > 0 {
		orbitY := py - 18
		for k := 0; k < 3; k++ {
			ang := float64(r.tick*7.0) + float64(k)*(2.0*math.Pi/3.0)
			sx := px + float32(math.Cos(ang))*11.0
			sy := orbitY + float32(math.Sin(ang))*3.5
			drawStarIcon(screen, sx, sy, 2.6, colGold)
		}
	}

	type statusBadge struct {
		label     string
		bgCol     color.RGBA
		borderCol color.RGBA
		txtCol    color.RGBA
	}
	var badges [4]statusBadge
	n := 0

	if p.StunTimer > 0 {
		bdr := colGold
		if int(r.tick*10)%2 == 0 {
			bdr = colRedAlert
		}
		badges[n] = statusBadge{
			label:     "STORDITO!",
			bgCol:     Premul(56, 16, 24, 232),
			borderCol: bdr,
			txtCol:    colGold,
		}
		n++
	}
	if p.GroundedTimer > 0 {
		remSec := math.Ceil(float64(p.GroundedTimer))
		if remSec < 1 {
			remSec = 1
		}
		badges[n] = statusBadge{
			label:     fmt.Sprintf("PUNIZIONE -1 STOP %.0fs", remSec),
			bgCol:     Premul(64, 10, 24, 236),
			borderCol: Premul(255, 45, 75, uint8(190+65*pulse)),
			txtCol:    color.RGBA{255, 200, 212, 255},
		}
		n++
	}
	if p.SlowTimer > 0 {
		bdr := colStudent
		if int(r.tick*8)%2 == 0 {
			bdr = colRedAlert
		}
		badges[n] = statusBadge{
			label:     "RALLENTATO",
			bgCol:     Premul(18, 28, 62, 232),
			borderCol: bdr,
			txtCol:    color.RGBA{165, 218, 255, 255},
		}
		n++
	}
	if p.SlipstreamTimer > 0 {
		bdr := colGyaru
		if int(r.tick*10)%2 == 0 {
			bdr = colGold
		}
		badges[n] = statusBadge{
			label:     "SCIA +45%",
			bgCol:     Premul(10, 44, 54, 232),
			borderCol: bdr,
			txtCol:    colGold,
		}
		n++
	}

	if n == 0 {
		return
	}

	by := py - 34
	if p.StunTimer > 0 {
		by = py - 38
	}
	for i := 0; i < n; i++ {
		b := badges[i]
		bw := float32(textPixelWidth(b.label, 1.0) + 10)
		bx := clampF(px-bw*0.5, 4, game.ScreenWidth-bw-4)
		ry := clampF(by, 27, game.ScreenHeight-62)
		drawPillBadge(screen, bx, ry, bw, 13, b.bgCol, b.borderCol)
		r.drawTextColored(screen, b.label, int(bx+5), int(ry-1), b.txtCol)
		by -= 15
	}
}

func (r *Renderer) drawAimAndRange(screen *ebiten.Image, w *game.World, camX, camY float32) {
	if st := focusedStop(w); st != nil && st.PlayerInside && r.stopPanelRect(st).W > 0 {
		return // Mouse clicks act on the active minigame or boarding combo here.
	}
	fb := w.EvaluateAim(r.aimX, r.aimY)
	spec := game.GetArchetypeSpec(w.Player.SelectedStopArch)
	maxRange := spec.ThrowRange * w.Player.EffectiveRangeMult()
	px, py := w.Player.X-camX, w.Player.Y-camY
	sx, sy := fb.ClampedX-camX, fb.ClampedY-camY
	ac := archetypeColor(w.Player.SelectedStopArch)

	// When on cooldown or at max active stops, draw a clear blocked/loading indicator directly
	// on the crosshair (radial recharge ring + countdown, or padlock + MAX indicator when slots are full),
	// while still showing the Mouse-LMB recall pill if hovering over an active stop for Stop Recall.
	if w.Player.StopCooldownTimer > 0 || w.ActiveStopCount() >= w.Player.EffectiveMaxStops() {
		showingRecall := false
		for i := range w.Stops {
			st := &w.Stops[i]
			if !st.Active || st.BusRushTriggered || st.Combo.Active {
				continue
			}
			dx := fb.ClampedX - st.X
			dy := fb.ClampedY - st.Y
			if dx*dx+dy*dy <= 44.0*44.0 {
				lbl := "CLICK SX RICHIAMA"
				bw := float32(len(lbl)*6 + 22)
				badgeX := clampF(sx+14, 4, game.ScreenWidth-bw-4)
				badgeY := clampF(sy-14, 28, game.ScreenHeight-70)
				drawPillBadge(screen, badgeX, badgeY, bw, 17, Premul(18, 42, 54, 225), colGyaru)
				r.drawMouseLMBIcon(screen, badgeX+3, badgeY+2, 0.80)
				r.drawScaledText(screen, lbl, int(badgeX+18), int(badgeY+1), 1.0, colGyaru, colOutline)
				showingRecall = true
				break
			}
		}

		// Inner precision dot/cross so aiming stays legible even while blocked
		muted := PremulColor(ac, 150)
		vector.StrokeLine(screen, sx-4, sy, sx+4, sy, 1.1, muted, false)
		vector.StrokeLine(screen, sx, sy-4, sx, sy+4, 1.1, muted, false)

		if w.Player.StopCooldownTimer > 0 {
			effCD := w.Player.EffectiveStopCooldown()
			if effCD < 0.5 {
				effCD = 0.5
			}
			readyFrac := clampF(1.0-w.Player.StopCooldownTimer/effCD, 0, 1)
			const ringR = float32(11.5)
			const arcSteps = 24

			// Backdrop & dark track circle
			vector.StrokeCircle(screen, sx, sy, ringR, 3.0, Premul(14, 10, 24, 195), true)
			vector.StrokeCircle(screen, sx, sy, ringR, 1.6, Premul(95, 90, 130, 185), true)

			// Clockwise recharge progress arc
			arcCol := colGold
			if readyFrac >= 0.80 {
				arcCol = colGyaru
			}
			filledSteps := int(float32(arcSteps) * readyFrac)
			if filledSteps < 1 && readyFrac > 0.02 {
				filledSteps = 1
			}
			startAng := float32(-math.Pi / 2)
			stepAng := float32(2*math.Pi) / float32(arcSteps)
			for s := 0; s < filledSteps; s++ {
				a0 := startAng + float32(s)*stepAng
				a1 := startAng + float32(s+1)*stepAng
				if s == filledSteps-1 {
					endAng := startAng + float32(2*math.Pi)*readyFrac
					if endAng > a0 {
						a1 = endAng
					}
				}
				x0 := sx + float32(math.Cos(float64(a0)))*ringR
				y0 := sy + float32(math.Sin(float64(a0)))*ringR
				x1 := sx + float32(math.Cos(float64(a1)))*ringR
				y1 := sy + float32(math.Sin(float64(a1)))*ringR
				vector.StrokeLine(screen, x0, y0, x1, y1, 2.2, arcCol, true)
			}
			if readyFrac > 0.01 {
				tipAng := startAng + float32(2*math.Pi)*readyFrac
				tx := sx + float32(math.Cos(float64(tipAng)))*ringR
				ty := sy + float32(math.Sin(float64(tipAng)))*ringR
				vector.DrawFilledCircle(screen, tx, ty, 1.8, color.RGBA{255, 255, 255, 245}, true)
			}

			if !showingRecall {
				cdText := fmt.Sprintf("%.1fs", w.Player.StopCooldownTimer)
				bw := float32(len(cdText)*6 + 8)
				bx := clampF(sx-bw/2, 4, game.ScreenWidth-bw-4)
				by := clampF(sy+15, 28, game.ScreenHeight-22)
				drawPillBadge(screen, bx, by, bw, 15, Premul(20, 14, 34, 215), arcCol)
				r.drawTextColored(screen, cdText, int(bx+4), int(by+1), arcCol)
			}
		} else {
			// Max active stops reached: draw locked ring + mini padlock + MAX X/Y badge
			lockCol := colRedAlert
			vector.StrokeCircle(screen, sx, sy, 10.5, 1.6, PremulColor(lockCol, 210), true)

			// Mini padlock at upper-right of crosshair
			lx, ly := sx+7, sy-11
			vector.StrokeLine(screen, lx+1.5, ly-3, lx+1.5, ly, 1.2, lockCol, false)
			vector.StrokeLine(screen, lx+1.5, ly-3, lx+4.5, ly-3, 1.2, lockCol, false)
			vector.StrokeLine(screen, lx+4.5, ly-3, lx+4.5, ly, 1.2, lockCol, false)
			vector.DrawFilledRect(screen, lx, ly, 6, 5, Premul(34, 12, 24, 235), false)
			vector.StrokeRect(screen, lx, ly, 6, 5, 1.1, lockCol, false)

			if !showingRecall {
				maxText := fmt.Sprintf("MAX %d/%d", w.ActiveStopCount(), w.Player.EffectiveMaxStops())
				bw := float32(len(maxText)*6 + 8)
				bx := clampF(sx-bw/2, 4, game.ScreenWidth-bw-4)
				by := clampF(sy+15, 28, game.ScreenHeight-22)
				drawPillBadge(screen, bx, by, bw, 15, Premul(36, 12, 24, 220), lockCol)
				r.drawTextColored(screen, maxText, int(bx+4), int(by+1), lockCol)
			}
		}
		return
	}

	pulse := float32(0.5 + 0.5*math.Sin(float64(r.tick*8.0)))

	vector.StrokeCircle(screen, px, py, maxRange, 1.0, PremulColor(ac, 65), true)
	vector.StrokeLine(screen, px, py, sx, sy, 1.0, PremulColor(ac, 75), true)

	if fb.HookCandidateCount > 0 {
		capSq := fb.PreviewRadius * fb.PreviewRadius
		for i := range w.Girls {
			g := &w.Girls[i]
			if g.HookedByStop != -1 {
				continue
			}
			dx := fb.ClampedX - g.X
			dy := fb.ClampedY - g.Y
			if dx*dx+dy*dy <= capSq {
				gx, gy := g.X-camX, g.Y-camY
				tCol := Premul(255, 255, 255, 175)
				ringCol := ac
				if g.Arch == w.Player.SelectedStopArch {
					tCol = PremulColor(colGold, 225)
					ringCol = colGold
				}
				vector.StrokeLine(screen, sx, sy, gx, gy, 2.2, PremulColor(ac, 90), true)
				vector.StrokeLine(screen, sx, sy, gx, gy, 1.2, tCol, true)
				vector.StrokeCircle(screen, gx, gy, 13+2*pulse, 1.5, ringCol, true)
			}
		}
		if w.Boss.Active && w.Boss.HookedByStop == -1 {
			bossR := fb.PreviewRadius + 24.0
			dx := fb.ClampedX - w.Boss.X
			dy := fb.ClampedY - w.Boss.Y
			if dx*dx+dy*dy <= bossR*bossR {
				bx, by := w.Boss.X-camX, w.Boss.Y-camY
				vector.StrokeLine(screen, sx, sy, bx, by, 2.5, colKawaii, true)
				vector.StrokeCircle(screen, bx, by, 22+2*pulse, 2.0, colGold, true)
			}
		}
	}

	retCol := ac
	ringAlpha := uint8(148)
	fillAlpha := uint8(10)
	if fb.InBusBay {
		retCol = colGreenNeon
		fillAlpha = 14
	}
	if fb.InZTL {
		retCol = actionCoral
		ringAlpha = 184
		fillAlpha = 16
	}
	if fb.IsPerfectDrop {
		retCol = color.RGBA{255, 211, 131, 255}
		ringAlpha = 198
		fillAlpha = 14
	}

	strokeCol := PremulColor(retCol, ringAlpha)
	vector.DrawFilledCircle(screen, sx, sy, fb.PreviewRadius, PremulColor(retCol, fillAlpha), true)
	vector.StrokeCircle(screen, sx, sy, fb.PreviewRadius, 1.25, strokeCol, true)
	vector.StrokeCircle(screen, sx, sy, 8+1.5*pulse, 1.5, strokeCol, true)
	vector.StrokeLine(screen, sx-12, sy, sx+12, sy, 1.2, strokeCol, false)
	vector.StrokeLine(screen, sx, sy-12, sx, sy+12, 1.2, strokeCol, false)
	if fb.IsPerfectDrop {
		const cb = float32(16)
		const ct = float32(5)
		vector.StrokeLine(screen, sx-cb, sy-cb, sx-cb+ct, sy-cb, 1.8, colGold, false)
		vector.StrokeLine(screen, sx-cb, sy-cb, sx-cb, sy-cb+ct, 1.8, colGold, false)
		vector.StrokeLine(screen, sx+cb, sy-cb, sx+cb-ct, sy-cb, 1.8, colGold, false)
		vector.StrokeLine(screen, sx+cb, sy-cb, sx+cb, sy-cb+ct, 1.8, colGold, false)
		vector.StrokeLine(screen, sx-cb, sy+cb, sx-cb+ct, sy+cb, 1.8, colGold, false)
		vector.StrokeLine(screen, sx-cb, sy+cb, sx-cb, sy+cb-ct, 1.8, colGold, false)
		vector.StrokeLine(screen, sx+cb, sy+cb, sx+cb-ct, sy+cb, 1.8, colGold, false)
		vector.StrokeLine(screen, sx+cb, sy+cb, sx+cb, sy+cb-ct, 1.8, colGold, false)
	}

	// Show AT MOST 1 concise contextual badge near cursor:
	// Priority: PERFECT! > SYNERGY! > ZTL 3x > BUS +25% (> HOOK xN)
	badgeX := clampF(sx+16, 4, game.ScreenWidth-116)
	badgeY := clampF(sy-14, 28, game.ScreenHeight-70)

	switch {
	case fb.IsPerfectDrop:
		const lbl = "PERFECT!"
		bw := float32(len(lbl)*6 + 20)
		drawPillBadge(screen, badgeX, badgeY, bw, 15, Premul(48, 32, 6, 220), colGold)
		drawBoltIcon(screen, badgeX+8, badgeY+7.5, colGold)
		r.drawScaledText(screen, lbl, int(badgeX+15), int(badgeY+1), 1.0, colGold, colOutline)
	case fb.SynergyCount > 0:
		lbl := fmt.Sprintf("SYNERGY x%d!", fb.SynergyCount)
		bw := float32(len(lbl)*6 + 20)
		drawPillBadge(screen, badgeX, badgeY, bw, 15, Premul(34, 14, 52, 220), ac)
		drawHeartIcon(screen, badgeX+8, badgeY+7.5, ac, true)
		r.drawScaledText(screen, lbl, int(badgeX+15), int(badgeY+1), 1.0, ac, colOutline)
	case fb.InZTL:
		const lbl = "ZTL 3x +1"
		bw := float32(len(lbl)*6 + 34)
		zCol := colRedAlert
		if pulse > 0.5 {
			zCol = colGold
		}
		drawPillBadge(screen, badgeX, badgeY, bw, 16, Premul(48, 8, 22, 225), zCol)
		drawWarningIcon(screen, badgeX+8, badgeY+8, zCol)
		r.drawScaledText(screen, lbl, int(badgeX+15), int(badgeY+2), 1.0, zCol, colOutline)
		r.DrawSpriteIcon(screen, assets.SpriteIconStarWanted, badgeX+71, badgeY+2, 1.0)
	case fb.InBusBay:
		const lbl = "BUS +25%"
		bw := float32(len(lbl)*6 + 12)
		drawPillBadge(screen, badgeX, badgeY, bw, 15, Premul(10, 42, 26, 220), colGreenNeon)
		r.drawScaledText(screen, lbl, int(badgeX+6), int(badgeY+1), 1.0, colGreenNeon, colOutline)
	case fb.HookCandidateCount > 0:
		lbl := fmt.Sprintf("HOOK x%d", fb.HookCandidateCount)
		bw := float32(len(lbl)*6 + 12)
		drawPillBadge(screen, badgeX, badgeY, bw, 15, colHoloTranslucent, ac)
		r.drawScaledText(screen, lbl, int(badgeX+6), int(badgeY+1), 1.0, color.White, colOutline)
	}
}

// resolveMouseHoverDir returns r.mouseHoverDir if set, otherwise infers the 4-way quadrant
// from the cursor position relative to (originX, originY).
func (r *Renderer) resolveMouseHoverDir(originX, originY float32) int {
	if r.mouseHoverDir >= game.ActionUp && r.mouseHoverDir <= game.ActionLeft {
		return r.mouseHoverDir
	}
	qdx := r.mouseScreenX - originX
	qdy := r.mouseScreenY - originY
	if qdx != 0 || qdy != 0 {
		if math.Abs(float64(qdx)) >= math.Abs(float64(qdy)) {
			if qdx >= 0 {
				return game.ActionRight
			}
			return game.ActionLeft
		}
		if qdy >= 0 {
			return game.ActionDown
		}
		return game.ActionUp
	}
	return game.ActionNone
}

// drawMouseCompass renders a crisp 4-way directional mouse compass (^ > v <)
// highlighting the hovered quadrant in bright electric gold-cyan.
func (r *Renderer) drawMouseCompass(screen *ebiten.Image, cx, cy, dist float32, activeDir, targetDir int) {
	dirs := [4]struct {
		id     int
		dx, dy float32
	}{
		{game.ActionUp, 0, -1},
		{game.ActionRight, 1, 0},
		{game.ActionDown, 0, 1},
		{game.ActionLeft, -1, 0},
	}

	// Subtle center compass hub
	vector.StrokeCircle(screen, cx, cy, 3.0, 1.0, Premul(0, 245, 212, 130), true)

	for _, d := range dirs {
		ax := cx + d.dx*dist
		ay := cy + d.dy*dist
		isHover := d.id == activeDir
		isTarget := d.id == targetDir

		bg := Premul(22, 18, 42, 155)
		bdr := Premul(110, 125, 165, 160)
		fg := Premul(180, 195, 225, 185)
		boxR := float32(7.0)

		if isTarget {
			bg = Premul(90, 68, 16, 190)
			bdr = colGold
			fg = colGold
		}
		if isHover {
			// Electric gold-cyan live hover highlight!
			bg = Premul(0, 132, 136, 220)
			bdr = colGold
			fg = color.RGBA{255, 246, 185, 255}
			boxR = 8.2
			vector.StrokeLine(screen, cx+d.dx*4, cy+d.dy*4, ax-d.dx*(boxR+1), ay-d.dy*(boxR+1), 2.0, colGyaru, true)
		}

		vector.DrawFilledRect(screen, ax-boxR, ay-boxR, boxR*2, boxR*2, bg, false)
		vector.StrokeRect(screen, ax-boxR, ay-boxR, boxR*2, boxR*2, 1.3, bdr, false)
		drawArrowIcon(screen, ax, ay, d.id, fg)
	}
}

func focusedStop(w *game.World) *game.BusStop {
	var focusStop *game.BusStop
	bestFocusD2 := float32(1e12)
	for i := range w.Stops {
		st := &w.Stops[i]
		if !st.Active || st.BusRushTriggered {
			continue
		}
		d2 := (w.Player.X-st.X)*(w.Player.X-st.X) + (w.Player.Y-st.Y)*(w.Player.Y-st.Y)

		if st.PlayerInside || st.CoyoteTimer > 0 {
			if focusStop == nil || (!focusStop.PlayerInside && focusStop.CoyoteTimer <= 0) || d2 < bestFocusD2 {
				focusStop = st
				bestFocusD2 = d2
			}
		} else if focusStop == nil {
			focusStop = st
			bestFocusD2 = d2
		}
	}
	return focusStop
}

func (r *Renderer) stopPanelRect(st *game.BusStop) game.Rect {
	x, y := st.UIAnchorX, st.UIAnchorY
	if !st.UIAnchored {
		x, y = 214, 232
	}
	var height float32
	switch {
	case st.Minigame.Completed && st.Combo.Active && !st.Combo.Completed:
		y += 2
		height = 70
	case !st.Minigame.Completed && st.Arch == game.ArchStudent:
		height = 76
	case !st.Minigame.Completed:
		return game.Rect{X: x + 32, Y: y, W: 148, H: 78}
	default:
		return game.Rect{}
	}
	if r.dirInputMode == 1 {
		return game.Rect{X: x + 35, Y: y, W: 142, H: height}
	}
	return game.Rect{X: x, Y: y, W: game.StopPanelDirW, H: height}
}

func (r *Renderer) drawDiegeticStopUI(screen *ebiten.Image, w *game.World, camX, camY float32) {
	pxScreen, pyScreen := w.Player.X-camX, w.Player.Y-camY
	safeTopY := w.StopSafeTopY()

	for i := range w.Stops {
		st := &w.Stops[i]
		if st.Active && !st.BusRushTriggered && !st.PlayerInside {
			sx, sy := st.X-camX, st.Y-camY
			dx, dy := sx-pxScreen, sy-pyScreen
			dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))
			if dist > 1 {
				nx, ny := dx/dist, dy/dist
				ax1, ay1 := pxScreen+nx*20, pyScreen+ny*20
				ax2, ay2 := pxScreen+nx*38, pyScreen+ny*38
				blinkCol := actionCoral
				vector.StrokeLine(screen, ax1, ay1, ax2, ay2, 3.0, blinkCol, true)
				vector.DrawFilledCircle(screen, ax2, ay2, 3.5, blinkCol, true)

				msg := "TORNA NEL CERCHIO!"
				mw := float32(textPixelWidth(msg, 1.0) + 24)
				mx := clampF(pxScreen-mw*0.5, 8, game.ScreenWidth-mw-8)
				my := clampF(pyScreen-46, safeTopY, game.ScreenHeight-70)
				drawPillBadge(screen, mx, my, mw, 18, Premul(56, 10, 20, 220), blinkCol)
				drawWarningIcon(screen, mx+10, my+9, blinkCol)
				r.drawScaledText(screen, msg, int(mx+19), int(my+2), 1.0, blinkCol, colOutline)
			}
			break
		}
	}

	st := focusedStop(w)
	if st == nil {
		return
	}
	ac := archetypeColor(st.Arch)
	panel := r.stopPanelRect(st)
	px, py, panelW, panelH := panel.X, panel.Y, panel.W, panel.H
	aroundStop := r.dirInputMode == 1
	padX, padY := st.PadCenter()
	if aroundStop {
		padX, padY = game.ClampCompassScreenPos(st.X-camX, st.Y-camY, safeTopY)
	}

	// FASE 2: Locked-in-Place Boarding Combo HUD (either integrated Pad or Compass around Bus Stop)
	if st.Minigame.Completed && st.Combo.Active && !st.Combo.Completed {
		hoverDir := r.resolveMouseHoverDir(padX, padY)
		targetDir := game.ActionNone
		if st.Combo.ComboIndex >= 0 && st.Combo.ComboIndex < len(st.Combo.Sequence) {
			targetDir = st.Combo.Sequence[st.Combo.ComboIndex]
		}

		if aroundStop && st.PlayerInside {
			r.drawMouseCompass(screen, padX, padY, 26, hoverDir, targetDir)
		}

		border := actionTeal
		if st.Combo.LastErrorFlash > 0 || st.Combo.StrictMode {
			border = actionCoral
		}
		drawStopPanel(screen, px, py, panelW, panelH, border)

		r.drawMouseLMBIcon(screen, px+8, py+4, 1.0)
		r.drawTextColored(screen, "COMBO IMBARCO", int(px+26), int(py+4), actionCream)

		for step := 0; step < 4; step++ {
			kx := px + 10 + float32(step)*31
			ky := py + 24
			stepDir := st.Combo.Sequence[step]
			boxCol := Premul(36, 36, 62, 185)
			bdrCol := Premul(125, 138, 175, 200)
			arrowCol := color.RGBA{215, 222, 240, 235}

			if step < st.Combo.ComboIndex {
				boxCol = Premul(18, 125, 85, 215)
				bdrCol = colGreenNeon
				arrowCol = color.RGBA{210, 255, 230, 255}
			} else if step == st.Combo.ComboIndex {
				if hoverDir == stepDir {
					boxCol = Premul(0, 145, 148, 230)
					bdrCol = colGold
					arrowCol = color.RGBA{255, 248, 195, 255}
				} else {
					boxCol = Premul(140, 98, 16, 215)
					bdrCol = colGold
					arrowCol = color.RGBA{255, 255, 255, 255}
				}
				ky--
			}
			vector.DrawFilledRect(screen, kx, ky, 22, 21, boxCol, false)
			vector.StrokeRect(screen, kx+0.5, ky+0.5, 21, 20, 1.4, bdrCol, false)
			drawArrowIcon(screen, kx+11, ky+10.5, stepDir, arrowCol)
		}

		frac := clampF(st.Combo.Timer/st.Combo.MaxTimer, 0, 1)
		barCol := actionTeal
		if frac < 0.35 {
			barCol = actionCoral
		}
		vector.DrawFilledRect(screen, px+10, py+56, 122, 4, Premul(20, 14, 28, 200), false)
		vector.DrawFilledRect(screen, px+10, py+56, snapPixel(122*frac), 4, barCol, false)

		if !aroundStop {
			vector.DrawFilledRect(screen, px+142, py+6, 1, panelH-12, Premul(255, 255, 255, 45), false)
			r.drawMouseCompass(screen, padX, padY, 18, hoverDir, targetDir)
		}
		return
	}

	// FASE 1: Locked-in-Place Semi-Transparent Minigame Box
	if !st.Minigame.Completed {
		if st.Arch == game.ArchStudent {
			hoverDir := r.resolveMouseHoverDir(padX, padY)
			if aroundStop && st.PlayerInside {
				r.drawMouseCompass(screen, padX, padY, 26, hoverDir, game.ActionNone)
			}
			drawStopPanel(screen, px, py, panelW, panelH, ac)
			r.drawMouseLMBIcon(screen, px+8, py+3, 1.0)
			r.drawTextColored(screen, "VAI ALL'ARRIVO", int(px+26), int(py+4), actionCream)
			r.drawUniRoute(screen, px, py, &st.Minigame)
			if !aroundStop {
				vector.DrawFilledRect(screen, px+142, py+6, 1, panelH-12, Premul(255, 255, 255, 45), false)
				r.drawMouseCompass(screen, padX, padY, 18, hoverDir, game.ActionNone)
			}
			return
		}

		drawStopPanel(screen, px, py, panelW, panelH, ac)

		switch st.Arch {
		case game.ArchEmo:
			r.drawMouseLMBIcon(screen, px+8, py+3, 1.0)
			r.drawTextColored(screen, "CLICK SX A TEMPO", int(px+28), int(py+4), actionCream)

			barX, barY, barW, barH := px+10, py+26, panelW-20, float32(15)
			innerX, innerY, innerW, innerH := barX+1, barY+1, barW-2, barH-2
			vector.DrawFilledRect(screen, innerX, innerY, innerW, innerH, Premul(24, 14, 42, 190), false)
			swW := float32(math.Round(float64(innerW * st.Minigame.BeatWindow * 2)))
			swLeft := float32(math.Round(float64(innerX + (innerW-swW)*0.5)))
			vector.DrawFilledRect(screen, swLeft, innerY, swW, innerH, Premul(46, 235, 130, 165), false)
			vector.StrokeRect(screen, barX+0.5, barY+0.5, barW-1, barH-1, 1.0, PremulColor(ac, 180), false)
			nx := float32(math.Round(float64(clampF(innerX+st.Minigame.BeatCursor*innerW-1.5, innerX, innerX+innerW-3))))
			vector.DrawFilledRect(screen, nx, barY-2, 3, barH+4, actionCream, false)

			inSweet := math.Abs(float64(st.Minigame.BeatCursor-0.5)) <= float64(st.Minigame.BeatWindow)
			progX, progY, progW, progH := px+10, py+52, panelW-20, float32(10)
			vector.DrawFilledRect(screen, progX+1, progY+1, progW-2, progH-2, Premul(20, 14, 34, 180), false)
			pW := float32(math.Round(float64((progW - 2) * clampF(st.Minigame.Progress, 0, 1))))
			vector.DrawFilledRect(screen, progX+1, progY+1, pW, progH-2, colGreenNeon, false)
			bdr := ac
			if inSweet {
				bdr = colGold
			}
			vector.StrokeRect(screen, progX+0.5, progY+0.5, progW-1, progH-1, 1.2, bdr, false)

		case game.ArchKawaii:
			r.drawMouseLMBIcon(screen, px+7, py+3, 1.0)
			r.drawTextColored(screen, "CLICK SX E SFIORA", int(px+26), int(py+4), actionCream)

			cx, cy := px+panelW*0.5, py+48
			vector.StrokeCircle(screen, cx, cy, 17, 3.5, Premul(60, 28, 52, 195), true)
			steps := int(st.Minigame.Progress * 20)
			for s := 0; s < steps; s++ {
				a1 := -math.Pi/2 + float64(s)*(2*math.Pi/20)
				a2 := -math.Pi/2 + float64(s+1)*(2*math.Pi/20)
				vector.StrokeLine(
					screen,
					cx+float32(math.Cos(a1)*17), cy+float32(math.Sin(a1)*17),
					cx+float32(math.Cos(a2)*17), cy+float32(math.Sin(a2)*17),
					3.5, colKawaii, true,
				)
			}
			value := fmt.Sprintf("%d%%", int(st.Minigame.Progress*100))
			r.drawTextColored(screen, value, int(cx)-textPixelWidth(value, 1)/2, int(cy)-7, actionCream)

		case game.ArchGyaru:
			r.drawMouseLMBIcon(screen, px+10, py+3, 1.0)
			r.drawTextColored(screen, "TIENI CLICK SX", int(px+30), int(py+4), actionCream)

			barX, barY, barW, barH := px+18, py+24, float32(16), float32(46)
			innerX, innerY, innerW, innerH := barX+1, barY+1, barW-2, barH-2
			vector.DrawFilledRect(screen, innerX, innerY, innerW, innerH, Premul(22, 28, 44, 195), false)
			swH := float32(math.Round(float64(innerH * st.Minigame.SweetSpotSize)))
			swCenterY := innerY + (1.0-st.Minigame.SweetSpotPos)*innerH
			swTop := float32(math.Round(float64(clampF(swCenterY-swH*0.5, innerY, innerY+innerH-swH))))
			vector.DrawFilledRect(screen, innerX, swTop, innerW, swH, Premul(46, 235, 130, 210), false)
			vector.StrokeRect(screen, barX+0.5, barY+0.5, barW-1, barH-1, 1.0, PremulColor(ac, 200), false)
			curTop := float32(math.Round(float64(clampF(innerY+(1.0-st.Minigame.BarCursor)*innerH-2, innerY, innerY+innerH-4))))
			vector.DrawFilledRect(screen, barX-2, curTop, barW+4, 4, colGold, false)

			progX, progW := px+46, float32(10)
			vector.DrawFilledRect(screen, progX+1, innerY, progW-2, innerH, Premul(18, 24, 38, 185), false)
			pH := float32(math.Round(float64(innerH * clampF(st.Minigame.Progress, 0, 1))))
			vector.DrawFilledRect(screen, progX+1, innerY+innerH-pH, progW-2, pH, colGyaru, false)
			vector.StrokeRect(screen, progX+0.5, barY+0.5, progW-1, barH-1, 1.0, ac, false)

			r.drawTextColored(screen, "CARICA", int(px+73), int(py+31), actionTeal)
			r.drawScaledText(screen, fmt.Sprintf("%d%%", int(st.Minigame.Progress*100)), int(px+73), int(py+46), 1.5, actionCream, color.Transparent)
		}
	}
}

func (r *Renderer) drawPopups(screen *ebiten.Image, w *game.World, camX, camY float32) {
	var shown [32]image.Rectangle
	count := 0
	if st := focusedStop(w); st != nil {
		bounds := r.stopPanelRect(st)
		if bounds.W > 0 {
			shown[0] = image.Rect(int(bounds.X)-2, int(bounds.Y)-2, int(bounds.X+bounds.W)+2, int(bounds.Y+bounds.H)+3)
			count++
		}
	}
	top := int(w.StopSafeTopY()) + 3
	for i := len(w.Popups) - 1; i >= 0 && count < len(shown); i-- {
		p := w.Popups[i]
		x, y := int(p.X-camX), int(p.Y-camY)
		if x < -16 || x > game.ScreenWidth+16 || y < top-12 || y >= game.ScreenHeight-53 {
			continue
		}
		var c color.RGBA
		switch p.Kind {
		case 1:
			c = actionTeal
			if p.Text == "PERFECT DROP!" {
				c = color.RGBA{255, 211, 131, 255}
			}
		case 2:
			c = actionCoral
		case 3:
			c = color.RGBA{172, 238, 145, 255}
		case 4:
			c = color.RGBA{255, 173, 195, 255}
		default:
			c = actionCream
		}
		txt := sanitizePopupText(p.Text)
		width := textPixelWidth(txt, 1)
		x = min(max(x-width/2, 5), game.ScreenWidth-width-5)
		y = max(y, top)
		var rect image.Rectangle
		for y >= top {
			rect = image.Rect(x-2, y, x+width+2, y+12)
			nextY := y
			for _, previous := range shown[:count] {
				if rect.Overlaps(previous) {
					nextY = min(nextY, previous.Min.Y-14)
				}
			}
			if nextY == y {
				break
			}
			y = nextY
		}
		if y < top {
			continue
		}
		shown[count] = rect
		count++
		fade := uint8(255 * clampF(p.Life/0.22, 0, 1))
		r.drawScaledText(screen, txt, x, rect.Min.Y, 1, PremulColor(c, fade), PremulColor(colOutline, fade))
	}
}

func sanitizePopupText(s string) string {
	s = strings.ReplaceAll(s, "★", "*")
	s = strings.ReplaceAll(s, "🚏 ", "")
	s = strings.ReplaceAll(s, "🚏", "")
	return s
}

func (r *Renderer) drawAtmosphereOverlays(screen *ebiten.Image, w *game.World) {
	if w.Player.ChronosActive || w.BulletFreezeTimer > 0 {
		vector.StrokeRect(screen, 2, 2, game.ScreenWidth-4, game.ScreenHeight-4, 4, Premul(0, 245, 212, 150), false)
		vector.DrawFilledRect(screen, 0, 0, game.ScreenWidth, game.ScreenHeight, Premul(45, 20, 105, 40), false)
	}
}

func clampF(v, lo, hi float32) float32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
