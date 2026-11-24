package render

import (
	"fmt"
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
	particles            [numAmbientParticles]ambientParticle
	drawables            []renderItem
	playerSkin           int
	busSkin              int
	screenShakeEnabled   bool
	showFPS              bool
	playingActive        bool
	activeChallengeToast string
	toastTimer           float32
}

// NewRenderer constructs a Renderer backed by the embedded SpriteBank.
func NewRenderer(sprites *assets.SpriteBank) *Renderer {
	r := &Renderer{
		sprites:            sprites,
		mouseHoverDir:      game.ActionNone,
		textCache:          make(map[string]*ebiten.Image, 160),
		lightTex:           buildRadialLightTexture(),
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

// ShowChallengeToast triggers a temporary top-center challenge unlock banner.
func (r *Renderer) ShowChallengeToast(title string) {
	r.activeChallengeToast = title
	r.toastTimer = 3.6
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
	const dt = 1.0 / 60.0
	r.tick += dt
	r.updateAmbientParticles(dt)
	if r.toastTimer > 0 {
		r.toastTimer -= dt
		if r.toastTimer < 0 {
			r.toastTimer = 0
		}
	}

	// Compute clamped camera + screen shake offset
	shakeX, shakeY := float32(0), float32(0)
	if r.screenShakeEnabled && w.ScreenShake > 0.05 {
		shakeX = float32(math.Sin(float64(r.tick*73.0))) * w.ScreenShake
		shakeY = float32(math.Cos(float64(r.tick*91.0))) * w.ScreenShake
	}
	camX := clampF(w.CameraX, 0, game.ArenaWidth-game.ScreenWidth) + shakeX
	camY := clampF(w.CameraY, 0, game.ArenaHeight-game.ScreenHeight) + shakeY

	// Clear backdrop
	screen.Fill(color.RGBA{26, 20, 38, 255})

	// Layer 0: Stratified Pop-Urban Street Arena, Parks, Bus Bays, ZTL, Boardwalk & Canal Water
	r.drawArenaTiles(screen, w, camX, camY)

	// Layer 0.5: Architectural Golden-Hour Drop Shadows
	r.drawGroundShadows(screen, w, camX, camY)

	// Layer 1: Express Lanes, Wave Rings, Bus Stop Capture Circles & Tethers
	r.drawExpressLanes(screen, w, camX, camY)
	r.drawWaveRings(screen, w, camX, camY)
	r.drawStopCirclesAndTethers(screen, w, camX, camY)

	// Layer 2: Lane / Laser Telegraphs
	r.drawTelegraphs(screen, w, camX, camY)

	// Layer 3: Y-Sorted 2.5D Props, Heal Pickups & World Entities
	r.drawYSortedWorld(screen, w, camX, camY)

	// Layer 4: Neon Buses, Bullets
	r.drawBuses(screen, w, camX, camY)
	r.drawBullets(screen, w, camX, camY)

	// Layer 4.5: Evening Twilight Veil, Multi-Layer Additive Lighting, Canal Reflections & Ambient Particles
	r.drawLightingPass(screen, w, camX, camY)
	r.drawAmbientParticles(screen, camX, camY)
	r.drawHeliSpotlight(screen, w, camX, camY)

	// Layer 4.8: Diegetic Near-Player HP & Micro-Chronos Bar (hidden when Alfredo is offscreen in Title backdrop)
	r.drawPlayerHealthBar(screen, w, camX, camY)

	// Layer 5: Contextual Aim Reticle & Mouse-Only Diegetic Minigames / 4-Step Combos
	if r.playingActive && !w.PendingLevelUp && !w.ShopMenu.Active && !w.SlotModal.Active {
		r.drawAimAndRange(screen, w, camX, camY)
		r.drawDiegeticStopUI(screen, w, camX, camY)
	}

	// Layer 6: Floating Popups & Offscreen Tactical Pointers
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
	tileZTL := r.sprites.Get(assets.SpriteTileZTL)
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
	maxCols := game.ArenaWidth / tileSize
	maxRows := game.ArenaHeight / tileSize
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
			if img == tileWater {
				waveShift := math.Sin(float64(r.tick*2.2+float32(row)*0.8)) * 2.5
				op.GeoM.Translate(sx+waveShift, sy)
			} else {
				op.GeoM.Translate(sx, sy)
			}
			screen.DrawImage(img, op)
		}
	}

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

	// 6. Forbidden ZTL Zones
	ztlPulse := float32(0.55 + 0.45*math.Sin(float64(r.tick*5.5)))
	for _, z := range w.ZTLZones {
		cols := int(math.Ceil(float64(z.W) / tileSize))
		rows := int(math.Ceil(float64(z.H) / tileSize))
		for rIdx := 0; rIdx < rows; rIdx++ {
			for cIdx := 0; cIdx < cols; cIdx++ {
				tx := z.X + float32(cIdx*tileSize)
				ty := z.Y + float32(rIdx*tileSize)
				op.GeoM.Reset()
				op.GeoM.Translate(float64(tx-camX), float64(ty-camY))
				op.ColorScale.Reset()
				op.ColorScale.ScaleAlpha(0.75)
				screen.DrawImage(tileZTL, op)
			}
		}
		op.ColorScale.Reset()

		zx, zy := z.X-camX, z.Y-camY
		borderCol := Premul(255, uint8(40+60*ztlPulse), uint8(110+40*ztlPulse), 240)
		vector.StrokeRect(screen, zx, zy, z.W, z.H, 2.0, borderCol, false)
		vector.DrawFilledRect(screen, zx+4, zy+4, 78, 14, Premul(30, 8, 22, 220), false)
		vector.StrokeRect(screen, zx+4, zy+4, 78, 14, 1, colGold, false)
		r.drawTextColored(screen, "ZTL 3x +1*", int(zx)+8, int(zy)+3, colGold)
	}

	// La Rettrice Martial ZTL Half-Arena Overlay
	if w.Boss.Active && w.Boss.MartialZTLActive {
		mz := w.Boss.MartialZTLRect
		zx, zy := mz.X-camX, mz.Y-camY
		vector.DrawFilledRect(screen, zx, zy, mz.W, mz.H, Premul(255, 35, 85, uint8(22+16*ztlPulse)), false)
		vector.StrokeRect(screen, zx, zy, mz.W, mz.H, 2.5, Premul(255, 45, 95, 220), false)
		r.drawScaledText(screen, "ZTL MARZIALE (NO PERFECT = +1*)", int(zx+14), int(zy+10), 1.0, colRedAlert, colOutline)
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
		alpha := clampF(wr.Life/wr.MaxLife, 0, 1)
		var c color.RGBA
		switch wr.Kind {
		case 0:
			c = Premul(0, 245, 212, uint8(160*alpha))
		case 1:
			c = Premul(255, 240, 120, uint8(230*alpha))
		case 2:
			c = Premul(255, 42, 133, uint8(230*alpha))
		case 3:
			c = Premul(75, 156, 226, uint8(210*alpha))
		default:
			c = Premul(255, 230, 0, uint8(220*alpha))
		}
		vector.StrokeCircle(screen, wr.X-camX, wr.Y-camY, wr.Radius, 2.5, c, true)
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

		fillAlpha := uint8(24)
		if st.PlayerInside {
			fillAlpha = 38
		}
		vector.DrawFilledCircle(screen, sx, sy, st.CaptureRadius, PremulColor(ac, fillAlpha), true)

		ringC := PremulColor(ac, 225)
		if st.SynergyActive {
			ringC = colGold
		}
		vector.StrokeCircle(screen, sx, sy, st.CaptureRadius, 2.2, ringC, true)

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
					vector.StrokeLine(screen, sx, sy, gx, gy, 3.2, PremulColor(ac, 95), true)
					vector.StrokeLine(screen, sx, sy, gx, gy, 1.5, Premul(255, 255, 255, 225), true)
					break
				}
			}
		}

		// Tether line to hooked Boss
		if st.HookedBoss && w.Boss.Active {
			bx, by := w.Boss.X-camX, w.Boss.Y-camY
			vector.StrokeLine(screen, sx, sy, bx, by, 4.2, Premul(255, 42, 133, 125), true)
			vector.StrokeLine(screen, sx, sy, bx, by, 1.8, colGold, true)
		}
	}
}

func (r *Renderer) drawTelegraphs(screen *ebiten.Image, w *game.World, camX, camY float32) {
	for _, t := range w.Telegraphs {
		blink := math.Sin(float64(r.tick*28.0)) > 0
		alpha := uint8(130)
		if blink {
			alpha = 235
		}
		x1, y1 := t.X1-camX, t.Y1-camY
		x2, y2 := t.X2-camX, t.Y2-camY
		if t.Kind == 0 {
			vector.StrokeLine(screen, x1, y1, x2, y2, 18, color.RGBA{255, 30, 60, 38}, false)
			vector.StrokeLine(screen, x1, y1, x2, y2, 2.0, color.RGBA{255, 45, 75, alpha}, false)
		} else {
			vector.StrokeLine(screen, x1, y1, x2, y2, 8, color.RGBA{255, 210, 40, 40}, true)
			vector.StrokeLine(screen, x1, y1, x2, y2, 2.0, color.RGBA{255, 70, 90, alpha}, true)
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

	// Ground contact shadow + soft pulsing mint-gold ground ring
	vector.DrawFilledCircle(screen, hx, hy+7, shadowR, Premul(20, 12, 32, 115), true)
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
		lbl := "CLICK / [E] SHOP"
		pw := float32(len(lbl)*6 + 24)
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
		lbl := fmt.Sprintf("CLICK / [E] SLOT 777 ($%d)", sm.SpinCost)
		pw := float32(len(lbl)*6 + 24)
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
	sx, sy := st.X-camX, st.Y-camY

	vector.DrawFilledCircle(screen, sx, sy+16, 16, color.RGBA{20, 12, 32, 120}, true)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-bw*0.5, -bh*0.5)
	op.GeoM.Translate(float64(sx), float64(sy))
	screen.DrawImage(img, op)

	// Back-glass shield pips above shelter
	if st.MaxShieldHP > 0 {
		pipW := float32(6)
		totalW := float32(st.MaxShieldHP)*(pipW+2) - 2
		startX := sx - totalW*0.5
		for p := 0; p < st.MaxShieldHP; p++ {
			c := color.RGBA{50, 50, 70, 190}
			if p < st.ShieldHP {
				c = colGyaru
			}
			vector.DrawFilledRect(screen, startX+float32(p)*(pipW+2), sy-28, pipW, 3.5, c, false)
			vector.StrokeRect(screen, startX+float32(p)*(pipW+2), sy-28, pipW, 3.5, 0.7, colOutline, false)
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
	vector.DrawFilledCircle(screen, gx, gy+12, 7, color.RGBA{26, 16, 37, 115}, true)

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
	frame := int(g.AnimTimer*6.0) % 3
	img := r.sprites.GetSheetFrame(sheetKey, frame, 32, 32)
	bw := float64(img.Bounds().Dx())
	bh := float64(img.Bounds().Dy())

	op.GeoM.Reset()
	op.GeoM.Translate(-bw*0.5, -bh*0.5)
	if g.VX < -4 {
		op.GeoM.Scale(-1, 1)
	}
	op.GeoM.Translate(float64(gx), float64(gy))
	if g.EnragedTimer > 0 {
		op.ColorScale.Scale(1.3, 0.6, 0.6, 1.0)
	}
	screen.DrawImage(img, op)

	if g.StunTimer > 0 {
		vector.DrawFilledCircle(screen, gx-5, gy-18, 2.2, colGold, false)
		vector.DrawFilledCircle(screen, gx+5, gy-18, 2.2, colGold, false)
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

	vector.DrawFilledCircle(screen, bx, by+15, 11, color.RGBA{26, 16, 37, 135}, true)
	vector.StrokeCircle(screen, bx, by, 22, 1.5, color.RGBA{255, 42, 133, 180}, true)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-bw*0.5, -bh*0.5)
	op.GeoM.Translate(float64(bx), float64(by))
	if w.Boss.EnrageFlashTimer > 0 && int(r.tick*20)%2 == 0 {
		op.ColorScale.Scale(1.4, 0.6, 0.8, 1.0)
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
	dashImg := r.sprites.Get(assets.SpriteAlfredoDash)

	op := &ebiten.DrawImageOptions{}
	bw := float64(dashImg.Bounds().Dx())
	bh := float64(dashImg.Bounds().Dy())
	for _, af := range p.Afterimages {
		if af.Alpha <= 0.02 {
			continue
		}
		op.GeoM.Reset()
		op.GeoM.Translate(-bw*0.5, -bh*0.5)
		if p.FacingX < 0 {
			op.GeoM.Scale(-1, 1)
		}
		op.GeoM.Translate(float64(af.X-camX), float64(af.Y-camY))
		op.ColorScale.Reset()
		switch r.playerSkin {
		case 1:
			op.ColorScale.Scale(1.0, 0.30, 0.85, af.Alpha*0.65)
		case 2:
			op.ColorScale.Scale(1.0, 0.82, 0.25, af.Alpha*0.65)
		default:
			op.ColorScale.Scale(0.15, 0.95, 1.0, af.Alpha*0.65)
		}
		screen.DrawImage(dashImg, op)
	}
	op.ColorScale.Reset()

	var key string
	switch p.AnimState {
	case game.AnimDash:
		key = assets.SpriteAlfredoDash
	case game.AnimThrow:
		key = assets.SpriteAlfredoThrow
	case game.AnimRun:
		switch int(p.AnimTimer*10) % 4 {
		case 0:
			key = assets.SpriteAlfredoRun0
		case 1:
			key = assets.SpriteAlfredoRun1
		case 2:
			key = assets.SpriteAlfredoRun2
		default:
			key = assets.SpriteAlfredoRun3
		}
	default:
		if int(p.AnimTimer*3)%2 == 0 {
			key = assets.SpriteAlfredoIdle0
		} else {
			key = assets.SpriteAlfredoIdle1
		}
	}

	img := r.sprites.Get(key)
	px, py := p.X-camX, p.Y-camY
	vector.DrawFilledCircle(screen, px, py+12, 7, color.RGBA{26, 16, 37, 130}, true)

	if p.InvulnTimer > 0 && p.DashTimer <= 0 && int(r.tick*22)%2 == 0 {
		return
	}

	op.GeoM.Reset()
	op.GeoM.Translate(-float64(img.Bounds().Dx())*0.5, -float64(img.Bounds().Dy())*0.5)
	if p.FacingX < 0 {
		op.GeoM.Scale(-1, 1)
	}
	op.GeoM.Translate(float64(px), float64(py))
	op.ColorScale.Reset()
	switch r.playerSkin {
	case 1:
		op.ColorScale.Scale(1.12, 0.92, 1.24, 1.0)
	case 2:
		op.ColorScale.Scale(0.95, 0.90, 0.85, 1.0)
		vector.StrokeCircle(screen, px, py, 14, 1.0, Premul(255, 220, 70, 120), true)
	}
	screen.DrawImage(img, op)

	// Tiny honest hitbox dot in bullet-hell center
	vector.DrawFilledCircle(screen, px, py+2, 2.2, colGyaru, true)
}

func (r *Renderer) drawBuses(screen *ebiten.Image, w *game.World, camX, camY float32) {
	// Draw lingering aerodynamic Bus Slipstream lanes (+45% speed & instant Dash reset)
	for _, sl := range w.Slipstreams {
		if sl.Life <= 0 || sl.MaxLife <= 0 {
			continue
		}
		alpha := sl.Life / sl.MaxLife
		aMain := uint8(float32(58) * alpha)
		aEdge := uint8(float32(130) * alpha)
		sx1 := sl.X1 - camX
		sx2 := sl.X2 - camX
		sy := sl.Y - camY
		sw := sx2 - sx1
		if sw <= 2 || sy < -32 || sy > game.ScreenHeight+32 {
			continue
		}
		vector.DrawFilledRect(screen, sx1, sy-16, sw, 32, Premul(0, 245, 212, aMain), false)
		vector.DrawFilledRect(screen, sx1, sy-17, sw, 1.5, Premul(255, 220, 60, aEdge), false)
		vector.DrawFilledRect(screen, sx1, sy+15.5, sw, 1.5, Premul(255, 220, 60, aEdge), false)
	}

	busImg := r.sprites.Get(assets.SpriteBusNeon)
	bw := float64(busImg.Bounds().Dx())
	bh := float64(busImg.Bounds().Dy())
	op := &ebiten.DrawImageOptions{}

	for _, b := range w.Buses {
		if !b.Active {
			continue
		}
		bx, by := b.X-camX, b.Y-camY
		if r.busSkin == 1 {
			vector.DrawFilledRect(screen, bx-140, by-18, 100, 36, Premul(255, 220, 40, 80), false)
			vector.DrawFilledRect(screen, bx-180, by-8, 120, 16, Premul(255, 140, 30, 105), false)
		} else {
			vector.DrawFilledRect(screen, bx-140, by-18, 100, 36, Premul(0, 245, 212, 70), false)
			vector.DrawFilledRect(screen, bx-180, by-8, 120, 16, Premul(255, 42, 133, 95), false)
		}

		op.GeoM.Reset()
		op.GeoM.Translate(-bw*0.5, -bh*0.5)
		op.GeoM.Translate(float64(bx), float64(by))
		op.ColorScale.Reset()
		if r.busSkin == 1 {
			op.ColorScale.Scale(1.30, 1.15, 0.55, 1.0)
		}
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
}

func (r *Renderer) drawAimAndRange(screen *ebiten.Image, w *game.World, camX, camY float32) {
	fb := w.EvaluateAim(r.aimX, r.aimY)
	spec := game.GetArchetypeSpec(w.Player.SelectedStopArch)
	maxRange := spec.ThrowRange * w.Player.EffectiveRangeMult()
	px, py := w.Player.X-camX, w.Player.Y-camY
	sx, sy := fb.ClampedX-camX, fb.ClampedY-camY
	ac := archetypeColor(w.Player.SelectedStopArch)

	// When on cooldown or at max active stops, draw only a subtle minimal crosshair
	// (plus a concise Mouse-LMB recall pill if hovering over an active stop for Stop Recall).
	if w.Player.StopCooldownTimer > 0 || w.ActiveStopCount() >= w.Player.EffectiveMaxStops() {
		muted := PremulColor(ac, 145)
		vector.StrokeCircle(screen, sx, sy, 7, 1.2, muted, true)
		vector.StrokeLine(screen, sx-10, sy, sx+10, sy, 1.0, muted, false)
		vector.StrokeLine(screen, sx, sy-10, sx, sy+10, 1.0, muted, false)

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
				break
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
	ringAlpha := uint8(200)
	fillAlpha := uint8(18)
	if fb.InBusBay {
		retCol = colGreenNeon
		fillAlpha = 24
	}
	if fb.InZTL {
		retCol = color.RGBA{255, uint8(45 + 90*pulse), 95, 255}
		ringAlpha = 230
		fillAlpha = 26
	}
	if fb.IsPerfectDrop {
		retCol = colGold
		ringAlpha = 245
		fillAlpha = uint8(26 + 14*pulse)
	}

	strokeCol := PremulColor(retCol, ringAlpha)
	vector.DrawFilledCircle(screen, sx, sy, fb.PreviewRadius, PremulColor(retCol, fillAlpha), true)
	vector.StrokeCircle(screen, sx, sy, fb.PreviewRadius, 1.6, strokeCol, true)
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
		const lbl = "ZTL 3x +1*"
		bw := float32(len(lbl)*6 + 20)
		zCol := colRedAlert
		if pulse > 0.5 {
			zCol = colGold
		}
		drawPillBadge(screen, badgeX, badgeY, bw, 15, Premul(48, 8, 22, 225), zCol)
		drawWarningIcon(screen, badgeX+8, badgeY+7.5, zCol)
		r.drawScaledText(screen, lbl, int(badgeX+15), int(badgeY+1), 1.0, zCol, colOutline)
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
	if qdx*qdx+qdy*qdy <= 165.0*165.0 && (qdx != 0 || qdy != 0) {
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

func (r *Renderer) drawDiegeticStopUI(screen *ebiten.Image, w *game.World, camX, camY float32) {
	pxScreen, pyScreen := w.Player.X-camX, w.Player.Y-camY
	outsideWarningDrawn := false

	for i := range w.Stops {
		st := &w.Stops[i]
		if !st.Active || st.BusRushTriggered {
			continue
		}
		sx, sy := st.X-camX, st.Y-camY
		ac := archetypeColor(st.Arch)

		if !st.PlayerInside && !outsideWarningDrawn {
			outsideWarningDrawn = true
			dx, dy := sx-pxScreen, sy-pyScreen
			dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))
			if dist > 1 {
				nx, ny := dx/dist, dy/dist
				ax1, ay1 := pxScreen+nx*20, pyScreen+ny*20
				ax2, ay2 := pxScreen+nx*38, pyScreen+ny*38
				blinkCol := colGold
				if int(r.tick*10)%2 == 0 {
					blinkCol = colRedAlert
				}
				vector.StrokeLine(screen, ax1, ay1, ax2, ay2, 3.0, blinkCol, true)
				vector.DrawFilledCircle(screen, ax2, ay2, 3.5, blinkCol, true)

				msg := "TORNA NEL CERCHIO!"
				mw := float32(len(msg)*6 + 24)
				mx := clampF(pxScreen-mw*0.5, 8, game.ScreenWidth-mw-8)
				my := clampF(pyScreen-46, 30, game.ScreenHeight-70)
				drawPillBadge(screen, mx, my, mw, 18, Premul(56, 10, 20, 220), blinkCol)
				drawWarningIcon(screen, mx+10, my+9, blinkCol)
				r.drawScaledText(screen, msg, int(mx+19), int(my+2), 1.0, blinkCol, colOutline)
			}
		}

		// FASE 2: Mouse-Only 4-Step Boarding Combo with live 4-Way Mouse Compass!
		if st.Minigame.Completed && st.Combo.Active && !st.Combo.Completed {
			hoverDir := r.resolveMouseHoverDir(sx, sy)
			targetDir := game.ActionNone
			if st.Combo.ComboIndex >= 0 && st.Combo.ComboIndex < len(st.Combo.Sequence) {
				targetDir = st.Combo.Sequence[st.Combo.ComboIndex]
			}
			if st.PlayerInside {
				r.drawMouseCompass(screen, sx, sy, 26, hoverDir, targetDir)
			}

			panelW := float32(138)
			panelH := float32(54)
			px := clampF(sx-panelW*0.5, 4, game.ScreenWidth-panelW-4)
			py := clampF(sy-86, 28, game.ScreenHeight-panelH-48)

			border := colGyaru
			if st.Combo.LastErrorFlash > 0 || st.Combo.StrictMode {
				border = colRedAlert
			}
			vector.DrawFilledRect(screen, px, py, panelW, panelH, colHoloTranslucent, false)
			vector.StrokeRect(screen, px, py, panelW, panelH, 1.6, border, false)

			r.drawMouseLMBIcon(screen, px+4, py+2, 0.80)
			r.drawScaledText(screen, "CLICK SX DIREZIONE", int(px+19), int(py+2), 0.9, colGold, colOutline)

			bounce := int(r.tick*10)%2 == 0
			for step := 0; step < 4; step++ {
				kx := px + 14 + float32(step)*29
				ky := py + 19
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
						// Live match glow: player's mouse is hovering the exact needed direction!
						boxCol = Premul(0, 145, 148, 230)
						bdrCol = colGold
						arrowCol = color.RGBA{255, 248, 195, 255}
					} else {
						boxCol = Premul(140, 98, 16, 215)
						bdrCol = colGold
						arrowCol = color.RGBA{255, 255, 255, 255}
					}
					if bounce {
						ky -= 2
					}
				}
				vector.DrawFilledRect(screen, kx, ky, 22, 21, boxCol, false)
				vector.StrokeRect(screen, kx, ky, 22, 21, 1.4, bdrCol, false)
				drawArrowIcon(screen, kx+11, ky+10.5, stepDir, arrowCol)
			}

			frac := clampF(st.Combo.Timer/st.Combo.MaxTimer, 0, 1)
			barCol := colGyaru
			if frac < 0.35 {
				barCol = colRedAlert
			}
			vector.DrawFilledRect(screen, px+8, py+45, (panelW-16)*frac, 4.5, barCol, false)
			continue
		}

		// FASE 1: Sleek Translucent Mouse-Only Diegetic Minigame Hologram
		if !st.Minigame.Completed {
			panelW := float32(126)
			panelH := float32(78)
			px := clampF(sx+30, 4, game.ScreenWidth-panelW-4)
			py := clampF(sy-42, 28, game.ScreenHeight-panelH-48)

			vector.DrawFilledRect(screen, px, py, panelW, panelH, colHoloTranslucent, false)
			vector.StrokeRect(screen, px, py, panelW, panelH, 1.5, ac, false)

			pulse := float32(0.5 + 0.5*math.Sin(float64(r.tick*8.0)))

			switch st.Arch {
			case game.ArchEmo:
				r.drawMouseLMBIcon(screen, px+5, py+4, 0.85)
				r.drawScaledText(screen, "CLICK SX A TEMPO", int(px+20), int(py+5), 0.9, colGold, colOutline)

				barX, barY, barW, barH := px+8, py+26, panelW-16, float32(15)
				vector.DrawFilledRect(screen, barX, barY, barW, barH, Premul(24, 14, 42, 190), false)
				swW := barW * st.Minigame.BeatWindow * 2
				vector.DrawFilledRect(screen, barX+barW*0.5-swW*0.5, barY, swW, barH, Premul(46, 235, 130, 165), false)
				nx := barX + st.Minigame.BeatCursor*barW
				vector.DrawFilledRect(screen, nx-1.5, barY-2, 3, barH+4, colGold, false)
				vector.StrokeRect(screen, barX, barY, barW, barH, 1.0, PremulColor(ac, 180), false)

				// Progress bar
				inSweet := math.Abs(float64(st.Minigame.BeatCursor-0.5)) <= float64(st.Minigame.BeatWindow)
				progY := py + 52
				vector.DrawFilledRect(screen, px+8, progY, panelW-16, 10, Premul(20, 14, 34, 180), false)
				vector.DrawFilledRect(screen, px+8, progY, (panelW-16)*st.Minigame.Progress, 10, colGreenNeon, false)
				bdr := ac
				if inSweet {
					bdr = colGold
				}
				vector.StrokeRect(screen, px+8, progY, panelW-16, 10, 1.2, bdr, false)

			case game.ArchStudent:
				hoverDir := r.resolveMouseHoverDir(sx, sy)
				if st.PlayerInside {
					r.drawMouseCompass(screen, sx, sy, 26, hoverDir, game.ActionNone)
				}
				r.drawMouseLMBIcon(screen, px+4, py+3, 0.80)
				r.drawScaledText(screen, "CLICK SX DIREZIONE", int(px+18), int(py+4), 0.85, colGold, colOutline)

				// 4x4 Pragmata Grid on left
				gx0, gy0 := px+8, py+22
				const cell = float32(11)
				for gy := 0; gy < 4; gy++ {
					for gx := 0; gx < 4; gx++ {
						cx := gx0 + float32(gx)*(cell+1)
						cy := gy0 + float32(gy)*(cell+1)
						c := Premul(38, 46, 72, 200)
						switch st.Minigame.GridNodes[gy][gx] {
						case game.NodeBlue:
							c = color.RGBA{40, 165, 255, 235}
						case game.NodeYellow:
							c = colGold
						case game.NodeBoard:
							c = colGreenNeon
						case game.NodeVisited:
							c = color.RGBA{0, 175, 155, 205}
						}
						vector.DrawFilledRect(screen, cx, cy, cell, cell, c, false)
						if gx == st.Minigame.GridX && gy == st.Minigame.GridY {
							vector.StrokeRect(screen, cx-1, cy-1, cell+2, cell+2, 1.5, color.White, false)
						}
					}
				}
				// Compact Mouse Compass inside right side of hologram
				r.drawMouseCompass(screen, px+92, py+46, 15, hoverDir, game.ActionNone)

			case game.ArchKawaii:
				r.drawMouseLMBIcon(screen, px+4, py+4, 0.82)
				r.drawScaledText(screen, "CLICK SX & SFIORA", int(px+19), int(py+5), 0.88, colGold, colOutline)

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
				r.drawScaledText(screen, fmt.Sprintf("%d%%", int(st.Minigame.Progress*100)), int(cx)-10, int(cy)-7, 1.0, color.White, colOutline)

			case game.ArchGyaru:
				r.drawMouseLMBIcon(screen, px+6, py+4, 0.85)
				r.drawScaledText(screen, "TIENI CLICK SX", int(px+22), int(py+5), 0.92, colGold, colOutline)

				barX, barY, barW, barH := px+14, py+24, float32(16), float32(46)
				vector.DrawFilledRect(screen, barX, barY, barW, barH, Premul(22, 28, 44, 195), false)
				swH := barH * st.Minigame.SweetSpotSize
				swCenterY := barY + (1.0-st.Minigame.SweetSpotPos)*barH
				swTop := clampF(swCenterY-swH*0.5, barY, barY+barH-swH)
				vector.DrawFilledRect(screen, barX+1, swTop, barW-2, swH, Premul(46, 235, 130, 195), false)
				curY := barY + (1.0-st.Minigame.BarCursor)*barH
				vector.DrawFilledRect(screen, barX-2, curY-2, barW+4, 4, colGold, false)
				vector.StrokeRect(screen, barX, barY, barW, barH, 1.0, PremulColor(ac, 180), false)

				pH := barH * st.Minigame.Progress
				vector.DrawFilledRect(screen, px+40, barY, 10, barH, Premul(18, 24, 38, 185), false)
				vector.DrawFilledRect(screen, px+40, barY+barH-pH, 10, pH, colGyaru, false)
				vector.StrokeRect(screen, px+40, barY, 10, barH, 1, ac, false)

				// Pulsing Mouse-LMB hold emblem on right
				holdCX, holdCY := px+84, py+47
				vector.StrokeCircle(screen, holdCX, holdCY, 12+2*pulse, 1.4, PremulColor(colGyaru, uint8(140+100*pulse)), true)
				r.drawMouseLMBIcon(screen, holdCX-8, holdCY-8, 1.0)
			}
		}
	}
}

func (r *Renderer) drawPopups(screen *ebiten.Image, w *game.World, camX, camY float32) {
	for _, p := range w.Popups {
		var c color.RGBA
		switch p.Kind {
		case 1:
			c = colGyaru
		case 2:
			c = colRedAlert
		case 3:
			c = colGreenNeon
		case 4:
			c = colKawaii
		default:
			c = color.RGBA{245, 245, 255, 255}
		}
		txt := sanitizePopupText(p.Text)
		sx := int(p.X-camX) - len(txt)*3
		sy := int(p.Y - camY)
		r.drawScaledText(screen, txt, sx, sy, 1.0, c, colOutline)
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
