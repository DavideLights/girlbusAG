package input

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"girlbusAG/internal/game"
)

// Reader tracks stateful input helpers (touch drag origin, archetype cycling, gamepad buffer) across frames.
type Reader struct {
	touchIDs       []ebiten.TouchID
	gamepadIDs     []ebiten.GamepadID
	moveTouchID    ebiten.TouchID
	hasMoveTouch   bool
	moveOriginX    int
	moveOriginY    int
	currentArch    int
	lastAimWorldX  float32
	lastAimWorldY  float32
	hasValidAim    bool
}

var defaultReader = NewReader()

// NewReader creates a stateful cross-platform input reader.
func NewReader() *Reader {
	return &Reader{
		currentArch: int(game.ArchEmo),
	}
}

// ReadInput returns the unified InputState using default reader state.
func ReadInput(camX, camY float32) game.InputState {
	return defaultReader.Read(camX, camY, defaultReader.currentArch, false)
}

// Read collects Keyboard + Mouse, Gamepad, and Touch input into a single game.InputState.
func (r *Reader) Read(camX, camY float32, activeArch int, inComboOrMinigame bool) game.InputState {
	r.currentArch = activeArch

	st := game.InputState{
		SelectArchetype: -1,
		ActionDir:       game.ActionNone,
		ChooseUpgrade:   -1,
	}

	var mx, my float32

	// 1. Keyboard movement (WASD always; Arrow keys when not in combo/minigame)
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		my -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		my += 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		mx -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		mx += 1
	}
	if !inComboOrMinigame {
		if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
			my -= 1
		}
		if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
			my += 1
		}
		if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
			mx -= 1
		}
		if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
			mx += 1
		}
	}

	// 2. Gamepad left stick & buttons
	r.gamepadIDs = ebiten.AppendGamepadIDs(r.gamepadIDs[:0])
	if len(r.gamepadIDs) > 0 {
		gid := r.gamepadIDs[0]
		if ebiten.IsStandardGamepadLayoutAvailable(gid) {
			ax := float32(ebiten.StandardGamepadAxisValue(gid, ebiten.StandardGamepadAxisLeftStickHorizontal))
			ay := float32(ebiten.StandardGamepadAxisValue(gid, ebiten.StandardGamepadAxisLeftStickVertical))
			const deadzone = 0.22
			if float32(math.Abs(float64(ax))) > deadzone {
				mx += ax
			}
			if float32(math.Abs(float64(ay))) > deadzone {
				my += ay
			}

			if inpututil.IsStandardGamepadButtonJustPressed(gid, ebiten.StandardGamepadButtonRightBottom) {
				st.DashJustPressed = true
			}
			if inpututil.IsStandardGamepadButtonJustPressed(gid, ebiten.StandardGamepadButtonRightRight) ||
				inpututil.IsStandardGamepadButtonJustPressed(gid, ebiten.StandardGamepadButtonFrontBottomRight) {
				st.ThrowJustPressed = true
			}
			if ebiten.IsStandardGamepadButtonPressed(gid, ebiten.StandardGamepadButtonFrontBottomLeft) ||
				ebiten.IsStandardGamepadButtonPressed(gid, ebiten.StandardGamepadButtonFrontTopLeft) {
				st.ChronosHeld = true
			}
			if inpututil.IsStandardGamepadButtonJustPressed(gid, ebiten.StandardGamepadButtonRightLeft) {
				st.InteractJustPressed = true
			}
			if inpututil.IsStandardGamepadButtonJustPressed(gid, ebiten.StandardGamepadButtonLeftTop) {
				st.ActionDir = game.ActionUp
			} else if inpututil.IsStandardGamepadButtonJustPressed(gid, ebiten.StandardGamepadButtonLeftRight) {
				st.ActionDir = game.ActionRight
			} else if inpututil.IsStandardGamepadButtonJustPressed(gid, ebiten.StandardGamepadButtonLeftBottom) {
				st.ActionDir = game.ActionDown
			} else if inpututil.IsStandardGamepadButtonJustPressed(gid, ebiten.StandardGamepadButtonLeftLeft) {
				st.ActionDir = game.ActionLeft
			}
			if inpututil.IsStandardGamepadButtonJustPressed(gid, ebiten.StandardGamepadButtonFrontTopRight) {
				st.SelectArchetype = (r.currentArch + 1) % int(game.ArchCount)
			}
		}
	}

	// 3. Touch drag (left half = virtual stick, right half = aim & throw)
	r.touchIDs = inpututil.AppendJustPressedTouchIDs(r.touchIDs[:0])
	for _, tid := range r.touchIDs {
		tx, ty := ebiten.TouchPosition(tid)
		if tx < game.ScreenWidth/2 && !r.hasMoveTouch {
			r.moveTouchID = tid
			r.hasMoveTouch = true
			r.moveOriginX = tx
			r.moveOriginY = ty
		} else {
			st.AimWorldX = float32(tx) + camX
			st.AimWorldY = float32(ty) + camY
			r.lastAimWorldX = st.AimWorldX
			r.lastAimWorldY = st.AimWorldY
			r.hasValidAim = true
			st.ThrowJustPressed = true
		}
	}
	if r.hasMoveTouch {
		if inpututil.IsTouchJustReleased(r.moveTouchID) {
			r.hasMoveTouch = false
		} else {
			tx, ty := ebiten.TouchPosition(r.moveTouchID)
			dx := float32(tx - r.moveOriginX)
			dy := float32(ty - r.moveOriginY)
			const dragRadius = 28.0
			if dx*dx+dy*dy > 16 {
				mx += clampF(dx/dragRadius, -1, 1)
				my += clampF(dy/dragRadius, -1, 1)
			}
		}
	}

	// Normalize movement vector so diagonal movement matches cardinal speed.
	if length := float32(math.Sqrt(float64(mx*mx + my*my))); length > 1.0 {
		mx /= length
		my /= length
	}
	st.MoveX = mx
	st.MoveY = my

	// 4. Dash, Chronos, Throw, Interact
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		inpututil.IsKeyJustPressed(ebiten.KeyShiftLeft) ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		st.DashJustPressed = true
	}

	if ebiten.IsKeyPressed(ebiten.KeyQ) ||
		ebiten.IsKeyPressed(ebiten.KeyC) ||
		ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
		st.ChronosHeld = true
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) ||
		inpututil.IsKeyJustPressed(ebiten.KeyF) ||
		inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		st.ThrowJustPressed = true
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		st.InteractJustPressed = true
	}

	// 5. Mouse aiming in world coordinates
	curX, curY := ebiten.CursorPosition()
	if curX != 0 || curY != 0 || !r.hasValidAim {
		st.AimWorldX = float32(curX) + camX
		st.AimWorldY = float32(curY) + camY
		r.lastAimWorldX = st.AimWorldX
		r.lastAimWorldY = st.AimWorldY
		r.hasValidAim = true
	} else {
		st.AimWorldX = r.lastAimWorldX
		st.AimWorldY = r.lastAimWorldY
	}

	// 6. Bus Stop archetype selection (1..4 or Tab cycle)
	if inpututil.IsKeyJustPressed(ebiten.KeyDigit1) || inpututil.IsKeyJustPressed(ebiten.KeyNumpad1) {
		st.SelectArchetype = int(game.ArchEmo)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDigit2) || inpututil.IsKeyJustPressed(ebiten.KeyNumpad2) {
		st.SelectArchetype = int(game.ArchStudent)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDigit3) || inpututil.IsKeyJustPressed(ebiten.KeyNumpad3) {
		st.SelectArchetype = int(game.ArchKawaii)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDigit4) || inpututil.IsKeyJustPressed(ebiten.KeyNumpad4) {
		st.SelectArchetype = int(game.ArchGyaru)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		st.SelectArchetype = (r.currentArch + 1) % int(game.ArchCount)
	}
	if st.SelectArchetype >= 0 {
		r.currentArch = st.SelectArchetype
	}

	// 7. Minigame & 4-Step Boarding Combo directional actions (Arrows OR I/J/K/L)
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyI) {
		st.ActionDir = game.ActionUp
	} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) || inpututil.IsKeyJustPressed(ebiten.KeyL) {
		st.ActionDir = game.ActionRight
	} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyK) {
		st.ActionDir = game.ActionDown
	} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebiten.KeyJ) {
		st.ActionDir = game.ActionLeft
	}

	// 8. Level-Up 3-Card upgrade selection (1/2/3 or Z/X/C)
	if inpututil.IsKeyJustPressed(ebiten.KeyDigit1) || inpututil.IsKeyJustPressed(ebiten.KeyZ) {
		st.ChooseUpgrade = 0
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDigit2) || inpututil.IsKeyJustPressed(ebiten.KeyX) {
		st.ChooseUpgrade = 1
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDigit3) || inpututil.IsKeyJustPressed(ebiten.KeyC) {
		st.ChooseUpgrade = 2
	}

	return st
}

// IsRestartJustPressed reports whether the rewind/restart key (R) was pressed.
func IsRestartJustPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyR)
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
