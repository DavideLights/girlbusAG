package input

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"girlbusAG/internal/game"
)

// MenuInput represents unified navigation/click input for Title, Pause, Options, Controls, and End screens.
type MenuInput struct {
	Up                   bool
	Down                 bool
	Left                 bool
	Right                bool
	Confirm              bool
	Back                 bool
	MouseX, MouseY       float32
	MouseLeftJustPressed bool
	MouseLeftPressed     bool
}

// Reader tracks stateful input helpers (touch drag origin, archetype cycling, aim mode, gamepad buffer) across frames.
type Reader struct {
	touchIDs      []ebiten.TouchID
	gamepadIDs    []ebiten.GamepadID
	moveTouchID   ebiten.TouchID
	hasMoveTouch  bool
	moveOriginX   int
	moveOriginY   int
	currentArch   int
	lastAimWorldX float32
	lastAimWorldY float32
	hasValidAim   bool
	lastCursorX   int
	lastCursorY   int
	hasInitCursor bool
	aimMode       int // 0=mouse, 1=keyboard, 2=gamepad
	gpAimX        float32
	gpAimY        float32
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
	return defaultReader.Read(camX, camY, defaultReader.currentArch, 0, 0, 0, true, false)
}

// Read collects Keyboard + Mouse, Gamepad, and Touch input into a single game.InputState.
// stopPhase: 0=none, 1=1-button minigame (Emo/Kawaii/Gyaru), 2=directional minigame/combo (Pragmata/Phase2).
func (r *Reader) Read(camX, camY float32, activeArch int, stopPhase int, stopScreenX, stopScreenY float32, canThrowExtraStop bool, chronosToggle bool) game.InputState {
	r.currentArch = activeArch
	inComboOrMinigame := stopPhase > 0

	st := game.InputState{
		SelectArchetype:   -1,
		ActionDir:         game.ActionNone,
		ChooseUpgrade:     -1,
		ShopAction:        game.ShopActionNone,
		ChronosToggleMode: chronosToggle,
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

	// 2. Gamepad Twin-Stick + Face Buttons (action dirs without leaving Left Stick)
	r.gamepadIDs = ebiten.AppendGamepadIDs(r.gamepadIDs[:0])
	if len(r.gamepadIDs) > 0 {
		gid := r.gamepadIDs[0]
		if ebiten.IsStandardGamepadLayoutAvailable(gid) {
			const deadzone = 0.22
			ax := float32(ebiten.StandardGamepadAxisValue(gid, ebiten.StandardGamepadAxisLeftStickHorizontal))
			ay := float32(ebiten.StandardGamepadAxisValue(gid, ebiten.StandardGamepadAxisLeftStickVertical))
			if float32(math.Abs(float64(ax))) > deadzone {
				mx += ax
			}
			if float32(math.Abs(float64(ay))) > deadzone {
				my += ay
			}

			// Right Stick analog aiming
			rx := float32(ebiten.StandardGamepadAxisValue(gid, ebiten.StandardGamepadAxisRightStickHorizontal))
			ry := float32(ebiten.StandardGamepadAxisValue(gid, ebiten.StandardGamepadAxisRightStickVertical))
			if rLen := float32(math.Sqrt(float64(rx*rx + ry*ry))); rLen > deadzone {
				r.aimMode = 2
				r.gpAimX = rx / rLen
				r.gpAimY = ry / rLen
			}

			// Bumpers / Triggers: L1 Dash, L2 Chronos, R1 Cycle Archetype, R2 Throw
			if inpututil.IsStandardGamepadButtonJustPressed(gid, ebiten.StandardGamepadButtonFrontTopLeft) {
				st.DashJustPressed = true
			}
			if ebiten.IsStandardGamepadButtonPressed(gid, ebiten.StandardGamepadButtonFrontBottomLeft) {
				st.ChronosHeld = true
			}
			if inpututil.IsStandardGamepadButtonJustPressed(gid, ebiten.StandardGamepadButtonFrontBottomLeft) {
				st.ChronosJustPressed = true
			}
			if inpututil.IsStandardGamepadButtonJustPressed(gid, ebiten.StandardGamepadButtonFrontBottomRight) {
				st.ThrowJustPressed = true
			}
			if inpututil.IsStandardGamepadButtonJustPressed(gid, ebiten.StandardGamepadButtonFrontTopRight) {
				st.SelectArchetype = (r.currentArch + 1) % int(game.ArchCount)
			}

			if stopPhase > 0 {
				// Map Face Buttons (Y/B/A/X) cleanly so Left Thumb stays on Left Stick
				if inpututil.IsStandardGamepadButtonJustPressed(gid, ebiten.StandardGamepadButtonRightTop) {
					st.ActionDir = game.ActionUp
				} else if inpututil.IsStandardGamepadButtonJustPressed(gid, ebiten.StandardGamepadButtonRightRight) {
					st.ActionDir = game.ActionRight
				} else if inpututil.IsStandardGamepadButtonJustPressed(gid, ebiten.StandardGamepadButtonRightBottom) {
					st.ActionDir = game.ActionDown
				} else if inpututil.IsStandardGamepadButtonJustPressed(gid, ebiten.StandardGamepadButtonRightLeft) {
					st.ActionDir = game.ActionLeft
				}
				if ebiten.IsStandardGamepadButtonPressed(gid, ebiten.StandardGamepadButtonRightTop) ||
					ebiten.IsStandardGamepadButtonPressed(gid, ebiten.StandardGamepadButtonRightRight) ||
					ebiten.IsStandardGamepadButtonPressed(gid, ebiten.StandardGamepadButtonRightBottom) ||
					ebiten.IsStandardGamepadButtonPressed(gid, ebiten.StandardGamepadButtonRightLeft) {
					st.ActionHeld = true
				}
			} else {
				if inpututil.IsStandardGamepadButtonJustPressed(gid, ebiten.StandardGamepadButtonRightBottom) {
					st.DashJustPressed = true
					st.SlotSpinJustPressed = true
				}
				if inpututil.IsStandardGamepadButtonJustPressed(gid, ebiten.StandardGamepadButtonRightRight) {
					st.CancelJustPressed = true
				}
				if inpututil.IsStandardGamepadButtonJustPressed(gid, ebiten.StandardGamepadButtonRightLeft) {
					st.InteractJustPressed = true
				}
			}

			// D-Pad fallback for minigames / combo
			if st.ActionDir == game.ActionNone {
				if inpututil.IsStandardGamepadButtonJustPressed(gid, ebiten.StandardGamepadButtonLeftTop) {
					st.ActionDir = game.ActionUp
				} else if inpututil.IsStandardGamepadButtonJustPressed(gid, ebiten.StandardGamepadButtonLeftRight) {
					st.ActionDir = game.ActionRight
				} else if inpututil.IsStandardGamepadButtonJustPressed(gid, ebiten.StandardGamepadButtonLeftBottom) {
					st.ActionDir = game.ActionDown
				} else if inpututil.IsStandardGamepadButtonJustPressed(gid, ebiten.StandardGamepadButtonLeftLeft) {
					st.ActionDir = game.ActionLeft
				}
			}
			if stopPhase > 0 && (ebiten.IsStandardGamepadButtonPressed(gid, ebiten.StandardGamepadButtonLeftTop) ||
				ebiten.IsStandardGamepadButtonPressed(gid, ebiten.StandardGamepadButtonLeftRight) ||
				ebiten.IsStandardGamepadButtonPressed(gid, ebiten.StandardGamepadButtonLeftBottom) ||
				ebiten.IsStandardGamepadButtonPressed(gid, ebiten.StandardGamepadButtonLeftLeft)) {
				st.ActionHeld = true
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
			r.aimMode = 0
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

	// 4. Dash, Chronos, Throw, Recall, Interact, Cancel, Slot Spin
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		st.DashJustPressed = true
	}

	if ebiten.IsKeyPressed(ebiten.KeyShiftLeft) ||
		ebiten.IsKeyPressed(ebiten.KeyQ) ||
		ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
		st.ChronosHeld = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyShiftLeft) ||
		inpututil.IsKeyJustPressed(ebiten.KeyQ) ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonMiddle) {
		st.ChronosJustPressed = true
	}

	kbThrow := inpututil.IsKeyJustPressed(ebiten.KeyF) || inpututil.IsKeyJustPressed(ebiten.KeyEnter)
	mouseThrow := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	if mouseThrow || kbThrow {
		st.ThrowJustPressed = true
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		st.RecallJustPressed = true
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		st.InteractJustPressed = true
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		st.SlotSpinJustPressed = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		st.CancelJustPressed = true
	}

	// 5. Mouse screen & world aiming coordinates + Aim Source Detection
	curX, curY := ebiten.CursorPosition()
	if !r.hasInitCursor {
		r.lastCursorX, r.lastCursorY = curX, curY
		r.hasInitCursor = true
	}
	cursorMoved := curX != r.lastCursorX || curY != r.lastCursorY
	r.lastCursorX, r.lastCursorY = curX, curY

	st.MouseScreenX = float32(curX)
	st.MouseScreenY = float32(curY)
	st.MouseLeftJustPressed = mouseThrow

	mouseClicked := mouseThrow || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight)
	if cursorMoved || mouseClicked {
		r.aimMode = 0
	} else if kbThrow {
		r.aimMode = 1
	}

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

	// 6. Bus Stop archetype selection (Wheel / 1..4 / Tab) & Shop choices (1..4)
	if inpututil.IsKeyJustPressed(ebiten.KeyDigit1) || inpututil.IsKeyJustPressed(ebiten.KeyNumpad1) {
		st.SelectArchetype = int(game.ArchEmo)
		st.ShopAction = game.ShopActionHeal
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDigit2) || inpututil.IsKeyJustPressed(ebiten.KeyNumpad2) {
		st.SelectArchetype = int(game.ArchStudent)
		st.ShopAction = game.ShopActionBribe
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDigit3) || inpututil.IsKeyJustPressed(ebiten.KeyNumpad3) {
		st.SelectArchetype = int(game.ArchKawaii)
		st.ShopAction = game.ShopActionRelic
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDigit4) || inpututil.IsKeyJustPressed(ebiten.KeyNumpad4) {
		st.SelectArchetype = int(game.ArchGyaru)
		st.ShopAction = game.ShopActionUpgrade
	} else if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		st.SelectArchetype = (r.currentArch + 1) % int(game.ArchCount)
	} else {
		_, wy := ebiten.Wheel()
		if wy > 0 {
			st.SelectArchetype = (r.currentArch + int(game.ArchCount) - 1) % int(game.ArchCount)
		} else if wy < 0 {
			st.SelectArchetype = (r.currentArch + 1) % int(game.ArchCount)
		}
	}
	if st.SelectArchetype >= 0 {
		r.currentArch = st.SelectArchetype
	}

	// 7. Minigame & 4-Step Boarding Combo directional actions (Arrows / IJKL / Zero Hand-Lift LMB)
	if st.ActionDir == game.ActionNone {
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyI) {
			st.ActionDir = game.ActionUp
		} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) || inpututil.IsKeyJustPressed(ebiten.KeyL) {
			st.ActionDir = game.ActionRight
		} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyK) {
			st.ActionDir = game.ActionDown
		} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebiten.KeyJ) {
			st.ActionDir = game.ActionLeft
		}
	}
	if stopPhase > 0 && (ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyI) ||
		ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyL) ||
		ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyK) ||
		ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyJ)) {
		st.ActionHeld = true
	}

	if stopPhase > 0 {
		dxCur := float32(curX) - stopScreenX
		dyCur := float32(curY) - stopScreenY
		lmbForMinigame := !canThrowExtraStop || (dxCur*dxCur+dyCur*dyCur <= 145*145)
		if stopPhase == 1 {
			if lmbForMinigame && st.MouseLeftJustPressed {
				if st.ActionDir == game.ActionNone {
					st.ActionDir = game.ActionUp
				}
				st.ThrowJustPressed = false
			}
			if lmbForMinigame && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
				st.ActionHeld = true
			}
		} else if stopPhase == 2 && lmbForMinigame && st.MouseLeftJustPressed {
			if st.ActionDir == game.ActionNone {
				if float32(math.Abs(float64(dxCur))) >= float32(math.Abs(float64(dyCur))) {
					if dxCur >= 0 {
						st.ActionDir = game.ActionRight
					} else {
						st.ActionDir = game.ActionLeft
					}
				} else {
					if dyCur >= 0 {
						st.ActionDir = game.ActionDown
					} else {
						st.ActionDir = game.ActionUp
					}
				}
			}
			st.ThrowJustPressed = false
		}
	}

	switch r.aimMode {
	case 2:
		st.AimIsDirectional = true
		st.AimDirX = r.gpAimX
		st.AimDirY = r.gpAimY
	case 1:
		st.AimIsDirectional = true
		st.AimDirX = mx
		st.AimDirY = my
	default:
		st.AimIsDirectional = false
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

// ReadMenuInput reads keyboard, gamepad, and mouse input for UI menu screens.
func (r *Reader) ReadMenuInput() MenuInput {
	curX, curY := ebiten.CursorPosition()
	mi := MenuInput{
		Up:                   inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW),
		Down:                 inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS),
		Left:                 inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA),
		Right:                inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) || inpututil.IsKeyJustPressed(ebiten.KeyD),
		Confirm:              inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace),
		Back:                 inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyBackspace),
		MouseX:               float32(curX),
		MouseY:               float32(curY),
		MouseLeftJustPressed: inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft),
		MouseLeftPressed:     ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft),
	}

	r.gamepadIDs = ebiten.AppendGamepadIDs(r.gamepadIDs[:0])
	if len(r.gamepadIDs) > 0 {
		gid := r.gamepadIDs[0]
		if ebiten.IsStandardGamepadLayoutAvailable(gid) {
			if inpututil.IsStandardGamepadButtonJustPressed(gid, ebiten.StandardGamepadButtonLeftTop) {
				mi.Up = true
			}
			if inpututil.IsStandardGamepadButtonJustPressed(gid, ebiten.StandardGamepadButtonLeftBottom) {
				mi.Down = true
			}
			if inpututil.IsStandardGamepadButtonJustPressed(gid, ebiten.StandardGamepadButtonLeftLeft) {
				mi.Left = true
			}
			if inpututil.IsStandardGamepadButtonJustPressed(gid, ebiten.StandardGamepadButtonLeftRight) {
				mi.Right = true
			}
			if inpututil.IsStandardGamepadButtonJustPressed(gid, ebiten.StandardGamepadButtonRightBottom) {
				mi.Confirm = true
			}
			if inpututil.IsStandardGamepadButtonJustPressed(gid, ebiten.StandardGamepadButtonRightRight) {
				mi.Back = true
			}
		}
	}

	return mi
}

// IsRestartJustPressed reports whether the rewind/restart key (R) was pressed.
func IsRestartJustPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyR)
}

// IsPauseJustPressed reports whether Escape or P was just pressed to toggle pause.
func IsPauseJustPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyP)
}

// AnyInputJustPressed reports whether any key, mouse button, or touch occurred this frame (for audio unlock).
func AnyInputJustPressed() bool {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		return true
	}
	keys := inpututil.AppendJustPressedKeys(nil)
	if len(keys) > 0 {
		return true
	}
	touches := inpututil.AppendJustPressedTouchIDs(nil)
	return len(touches) > 0
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
