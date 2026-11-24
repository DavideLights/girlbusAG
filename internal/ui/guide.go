package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"girlbusAG/assets"
	"girlbusAG/internal/game"
	"girlbusAG/internal/render"
)

var (
	guideLilac = color.RGBA{216, 172, 246, 255}
	guideSky   = color.RGBA{144, 199, 245, 255}
)

func (m *Manager) drawTutorial(dst *ebiten.Image, r *render.Renderer, mx, my float32) {
	screenShade(dst, 186)
	menuBoard(dst, 24, 14, 592, 330, menuTeal)
	menuHeading(dst, r, "Guida rapida", 38, 16)
	menuChip(dst, r, fmt.Sprintf("%d / 4", m.TutorialTab+1), 543, 27, 59, menuGold)

	tabs := [...]string{"FERMATE E BASI", "MINIGIOCHI", "SLOT E SHOP", "CURE E BOSS"}
	tabAccents := [...]color.RGBA{menuTeal, guideLilac, menuGold, menuCoral}
	for i, label := range tabs {
		rect := tutorialTabRect(i)
		bg, fg := menuSurface, menuMuted
		accent := tabAccents[i]
		if m.TutorialTab == i {
			bg, fg = accent, menuInk
		} else if m.usingPointer && rect.Contains(mx, my) {
			bg, fg = menuRaised, accent
		}
		fill(dst, rect.X, rect.Y, rect.W, rect.H, bg)
		if m.TutorialTab == i && m.SelectedIdx == 0 {
			fill(dst, rect.X, rect.Y+rect.H-2, rect.W, 2, menuCream)
		} else if m.TutorialTab != i {
			fill(dst, rect.X, rect.Y+rect.H-2, rect.W, 1, render.PremulColor(accent, 130))
		}
		if fg == menuInk {
			x := int(rect.X + (rect.W-float32(render.TextPixelWidth(label, 1)))/2)
			r.DrawScaledText(dst, label, x, int(rect.Y+3), 1, fg, color.Transparent)
		} else {
			centerText(dst, r, label, rect.X, rect.Y+3, rect.W, fg)
		}
	}

	switch m.TutorialTab {
	case 0:
		m.drawTutorialTabStops(dst, r)
	case 1:
		m.drawTutorialTabMinigames(dst, r)
	case 2:
		m.drawTutorialTabSlotsShop(dst, r)
	case 3:
		m.drawTutorialTabSurvivalBoss(dst, r)
	}

	type guideButton struct {
		rect  game.Rect
		label string
		focus int
	}
	for _, b := range [...]guideButton{
		{tutorialPrevRect(), "< PREC", -1},
		{tutorialNextRect(), "SUCC >", -1},
		{tutorialPracticeRect(), "PROVA SUL CAMPO", 1},
		{tutorialBackRect(), "TORNA [ESC]", 2},
	} {
		style := m.buttonStyle(b.focus, b.rect, mx, my)
		style.primary = b.focus == 1
		drawMenuButton(dst, r, b.rect, b.label, style)
	}
	r.DrawTextColored(dst, "A / D pagina", 222, 305, menuMuted)
}

func guideCard(dst *ebiten.Image, x, y, w, h float32, accent color.RGBA) {
	fill(dst, x+2, y+2, w, h, menuInk)
	fill(dst, x, y, w, h, menuSurface)
	fill(dst, x, y+1, w, min(h-1, 23), render.PremulColor(accent, 22))
	fill(dst, x, y, w, 1, menuRule)
	fill(dst, x, y, 26, 2, accent)
}

func (m *Manager) drawTutorialTabStops(dst *ebiten.Image, r *render.Renderer) {
	// Top Objective Banner: explicit win goal of the game
	guideCard(dst, 38, 80, 564, 28, menuGold)
	r.DrawTextColored(dst, "Fai salire le passeggere sul bus e batti i quattro boss.", 46, 81, menuGold)
	r.DrawTextColored(dst, "Ogni imbarco ti dà esperienza e denaro.", 46, 94, menuCream)

	// Left Card: Controls (WASD + Mouse)
	guideCard(dst, 38, 112, 276, 180, menuTeal)
	r.DrawScaledText(dst, "COMANDI ESSENZIALI", 48, 117, 1.0, menuGold, render.ColOutline)

	// Row 1: WASD
	r.DrawKeycap(dst, "WASD", 46, 134, 48, 22, menuPanel, menuTeal, false)
	r.DrawTextColored(dst, "Muovi Alfredo", 102, 134, menuCream)
	r.DrawTextColored(dst, "Schiva i proiettili in strada", 102, 146, menuMuted)

	// Row 2: LMB
	drawTutorialIconBadge(dst, 46, 165, 48, 24, menuGold)
	drawSpriteIconCentered(dst, r, assets.SpriteIconMouseLMB, 70, 177, 1)
	r.DrawTextColored(dst, "CLICK SX: Mira & Lancia", 102, 164, menuGold)
	r.DrawTextColored(dst, "Fermata + Minigiochi & Combo", 102, 176, menuCream)

	// Row 3: RMB / Space
	drawTutorialIconBadge(dst, 46, 196, 48, 24, menuCoral)
	drawSpriteIconCentered(dst, r, assets.SpriteIconMouseRMB, 70, 208, 1)
	r.DrawTextColored(dst, "CLICK DX / SPAZIO: Dash", 102, 195, menuTeal)
	r.DrawTextColored(dst, "Attraversa i proiettili!", 102, 207, menuCream)

	// Row 4: Wheel / 1-4
	drawTutorialIconBadge(dst, 46, 227, 48, 24, guideLilac)
	drawSpriteIconCentered(dst, r, assets.SpriteIconMouseWheel, 70, 239, 1)
	r.DrawTextColored(dst, "ROTELLA / 1-4:", 102, 226, guideLilac)
	r.DrawTextColored(dst, "Cambia tipo di Fermata", 102, 238, menuCream)

	// Row 5: Shift Chronos
	r.DrawKeycap(dst, "SHIFT", 46, 259, 48, 22, menuPanel, guideSky, false)
	r.DrawTextColored(dst, "Chronos (Slow-Mo tattico)", 102, 258, guideSky)
	r.DrawTextColored(dst, "Rallenta il tempo per mirare", 102, 270, menuMuted)

	// Right Card: Bus Stops & Neon Bus
	guideCard(dst, 322, 112, 280, 180, menuCoral)
	r.DrawScaledText(dst, "COME FARE UN IMBARCO", 332, 117, 1.0, menuGold, render.ColOutline)

	// Visual Stage: Stop -> Neon Bus
	guideCard(dst, 332, 133, 260, 36, menuRule)
	drawSpriteIconCentered(dst, r, assets.SpriteStopEmo, 384, 151, 0.5)
	r.DrawScaledText(dst, ">>>", 436, 144, 1, menuGold, render.ColOutline)
	drawSpriteIconCentered(dst, r, assets.SpriteBusNeon, 526, 151, 0.5)

	// 3 Concise Callouts
	r.DrawTextColored(dst, "1  PIAZZA LA FERMATA", 332, 173, menuGold)
	r.DrawTextColored(dst, "Click sinistro vicino a una ragazza.", 332, 186, menuCream)
	r.DrawTextColored(dst, "Mirino dorato = Perfect Drop.", 332, 197, menuGold)
	r.DrawTextColored(dst, "2  COMPLETA IL MINIGIOCO", 332, 209, menuTeal)
	r.DrawTextColored(dst, "Resta nel cerchio e segui il riquadro.", 332, 222, menuCream)
	r.DrawTextColored(dst, "3  CHIAMA L'AUTOBUS", 332, 245, menuCoral)
	r.DrawTextColored(dst, "Chiudi la combo per farlo arrivare.", 332, 258, menuCream)
}

func (m *Manager) drawTutorialTabMinigames(dst *ebiten.Image, r *render.Renderer) {
	type miniCard struct {
		title  string
		line1  string
		line2  string
		sprite string
		accent color.RGBA
		x, y   float32
	}
	cards := [4]miniCard{
		{
			title:  "EMO (RHYTHM)",
			line1:  "Click Sx quando il cursore",
			line2:  "tocca il centro della barra.",
			sprite: assets.SpriteGirlEmo0,
			accent: guideLilac,
			x:      38, y: 80,
		},
		{
			title:  "STUDENTESSA (GRIGLIA 4x4)",
			line1:  "Raggiungi la bandiera verde.",
			line2:  "Bonus  Azzurro XP / Giallo rallenta",
			sprite: assets.SpriteGirlStudent0,
			accent: guideSky,
			x:      324, y: 80,
		},
		{
			title:  "KAWAII (GOCCIOLINE)",
			line1:  "Danza nel cerchio & Click Sx",
			line2:  "mentre sfiori le gocce!",
			sprite: assets.SpriteGirlKawaii0,
			accent: menuCoral,
			x:      38, y: 160,
		},
		{
			title:  "GYARU (BARRA TENSIONE)",
			line1:  "Tieni premuto Click Sx per",
			line2:  "restare nella zona verde.",
			sprite: assets.SpriteGirlGyaru0,
			accent: menuTeal,
			x:      324, y: 160,
		},
	}

	for _, c := range cards {
		guideCard(dst, c.x, c.y, 278, 74, c.accent)

		drawTutorialIconBadge(dst, c.x+8, c.y+9, 44, 56, c.accent)
		drawSpriteIconCentered(dst, r, c.sprite, c.x+30, c.y+37, 1)

		r.DrawScaledText(dst, c.title, int(c.x+60), int(c.y+10), 1.0, c.accent, render.ColOutline)
		r.DrawTextColored(dst, c.line1, int(c.x+60), int(c.y+30), menuCream)
		r.DrawTextColored(dst, c.line2, int(c.x+60), int(c.y+45), menuMuted)
	}

	// Boarding combo summary.
	guideCard(dst, 38, 242, 564, 50, menuGold)
	drawTutorialIconBadge(dst, 46, 249, 36, 36, menuGold)
	drawSpriteIconCentered(dst, r, assets.SpriteIconMouseLMB, 64, 267, 1)

	r.DrawScaledText(dst, "COMBO FINALE / 4 DIREZIONI", 92, 246, 1.0, menuGold, render.ColOutline)
	r.DrawTextColored(dst, "Segui le quattro frecce con il mouse, la tastiera o il gamepad.", 92, 262, menuCream)
	r.DrawTextColored(dst, "Completa la combo per chiamare il Bus Neon.", 92, 276, menuTeal)
}

func (m *Manager) drawTutorialTabSlotsShop(dst *ebiten.Image, r *render.Renderer) {
	// Left Card: Rizz-Slots 777
	guideCard(dst, 38, 80, 278, 212, menuGold)

	drawTutorialIconBadge(dst, 46, 88, 38, 38, menuGold)
	drawSpriteIconCentered(dst, r, assets.SpritePropSlotMachine, 65, 107, 1.0)
	r.DrawScaledText(dst, "RIZZ-SLOTS 777", 92, 91, 1, menuGold, render.ColOutline)
	r.DrawTextColored(dst, "Avvicinati e CLICCA (o E)", 92, 108, menuTeal)

	type iconRow struct {
		icon string
		head string
		desc string
		col  color.RGBA
	}
	slotRows := [4]iconRow{
		{assets.SpriteIconSlot777, "JACKPOT 777", "Vita al massimo e $250", menuGold},
		{assets.SpriteIconSlotHeart, "CUORI", "Cura istantanea", menuCoral},
		{assets.SpriteIconSlotCash, "CONTANTI", "Premio in denaro", menuTeal},
		{assets.SpriteIconSlotSiren, "SIRENA", "+$110 e +1 stella Polizia", menuCoral},
	}
	for i, row := range slotRows {
		ry := float32(134 + i*38)
		drawTutorialIconBadge(dst, 48, ry, 28, 28, row.col)
		drawSpriteIconCentered(dst, r, row.icon, 62, ry+14, 1.0)
		r.DrawTextColored(dst, row.head, 84, int(ry+1), row.col)
		r.DrawTextColored(dst, row.desc, 84, int(ry+14), menuCream)
	}

	// Right Card: Neon Vending Machine
	guideCard(dst, 324, 80, 278, 212, menuTeal)

	drawTutorialIconBadge(dst, 332, 88, 38, 38, menuTeal)
	drawSpriteIconCentered(dst, r, assets.SpriteShopVending, 351, 107, 1.0)
	r.DrawScaledText(dst, "DISTRIBUTORE NEON", 378, 91, 1, menuTeal, render.ColOutline)
	r.DrawTextColored(dst, "Avvicinati e CLICCA (o E)", 378, 108, menuGold)

	shopRows := [4]iconRow{
		{assets.SpriteIconShopEspresso, "ESPRESSO", "+2 HP e ricarica Chronos", menuGold},
		{assets.SpriteIconShopBribe, "TANGENTE", "-2 stelle Polizia", menuTeal},
		{assets.SpriteIconShopRelic, "RELIQUIE", "Bonus passivi unici", guideLilac},
		{assets.SpriteIconShopCrate, "CASSA UPGRADE", "Potenziamento extra", menuTeal},
	}
	for i, row := range shopRows {
		ry := float32(134 + i*38)
		drawTutorialIconBadge(dst, 334, ry, 28, 28, row.col)
		drawSpriteIconCentered(dst, r, row.icon, 348, ry+14, 1)
		r.DrawTextColored(dst, row.head, 370, int(ry+1), row.col)
		r.DrawTextColored(dst, row.desc, 370, int(ry+14), menuCream)
	}
}

func (m *Manager) drawTutorialTabSurvivalBoss(dst *ebiten.Image, r *render.Renderer) {
	// Row 1: HP & Espresso Heals
	guideCard(dst, 38, 80, 564, 64, menuTeal)
	drawTutorialIconBadge(dst, 46, 88, 62, 48, menuTeal)
	drawSpriteIconCentered(dst, r, assets.SpriteIconHeartFull, 63, 112, 1)
	drawSpriteIconCentered(dst, r, assets.SpriteIconPickupHeal, 90, 112, 1)

	r.DrawScaledText(dst, "VITA & CURE ESPRESSO", 118, 86, 1, menuTeal, render.ColOutline)
	r.DrawTextColored(dst, "Tieni d'occhio la vita di Alfredo durante la corsa.", 118, 104, menuCream)
	r.DrawTextColored(dst, "Raccogli le Cure Espresso in strada o quando le lasciano i boss.", 118, 119, menuTeal)

	// Row 2: Police Stars (1* - 5*)
	guideCard(dst, 38, 154, 564, 64, menuGold)
	drawTutorialIconBadge(dst, 46, 162, 62, 48, menuGold)
	drawSpriteIconCentered(dst, r, assets.SpriteIconStarWanted, 60, 190, 1)
	drawSpriteIconCentered(dst, r, assets.SpriteIconStarWanted, 77, 178, 1)
	drawSpriteIconCentered(dst, r, assets.SpriteIconStarWanted, 94, 190, 1)

	r.DrawScaledText(dst, "POLIZIA (1* - 5* STELLE)", 118, 160, 1, menuGold, render.ColOutline)
	r.DrawTextColored(dst, "Zone ZTL o ragazze VIP aumentano le stelle di ricerca.", 118, 178, menuCream)
	r.DrawTextColored(dst, "Più pericolo e moltiplicatore fino a 3.5x XP e 4x Cash.", 118, 193, menuGold)

	// Row 3: Gatekeeper Bosses (3 Shields) - Core Win Objective
	guideCard(dst, 38, 228, 564, 64, menuCoral)
	drawTutorialIconBadge(dst, 46, 236, 62, 48, menuCoral)
	drawSpriteIconCentered(dst, r, assets.SpriteBossMom, 77, 260, 1)

	r.DrawScaledText(dst, "BOSS GATEKEEPER (OBIETTIVO: BATTI 4 BOSS!)", 118, 234, 1, menuCoral, render.ColOutline)
	r.DrawTextColored(dst, "Aggancia ogni Boss 3 volte con una Fermata per rompere i 3 scudi.", 118, 252, menuCream)
	r.DrawTextColored(dst, "Batti i 4 boss della run. Per la Boss Rush, batti tutti e 6 i boss.", 118, 267, menuGold)
}

func drawTutorialIconBadge(dst *ebiten.Image, x, y, w, h float32, border color.RGBA) {
	fill(dst, x, y, w, h, menuInk)
	fill(dst, x, y+h-1, w, 1, render.PremulColor(border, 175))
}

func drawSpriteIconCentered(dst *ebiten.Image, r *render.Renderer, spriteKey string, cx, cy, scale float32) {
	if r == nil || r.Sprites() == nil {
		return
	}
	img := r.Sprites().Get(spriteKey)
	if img == nil {
		return
	}
	b := img.Bounds()
	w := float32(b.Dx()) * scale
	h := float32(b.Dy()) * scale
	r.DrawSpriteIcon(dst, spriteKey, cx-w*0.5, cy-h*0.5, scale)
}
