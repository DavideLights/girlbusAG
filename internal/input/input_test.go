package input

import (
	"testing"

	"girlbusAG/internal/game"
)

func TestComputeMouseCardinalDir(t *testing.T) {
	const ox, oy = float32(320), float32(180)
	cases := []struct {
		name string
		x, y float32
		want int
	}{
		{"Right", ox + 50, oy, game.ActionRight},
		{"Left", ox - 50, oy, game.ActionLeft},
		{"Up", ox, oy - 50, game.ActionUp},
		{"Down", ox, oy + 50, game.ActionDown},
		{"NearCenterRight", ox + 2, oy + 1, game.ActionRight},
		{"NearCenterLeft", ox - 2, oy - 1, game.ActionLeft},
		{"NearCenterDown", ox + 1, oy + 2, game.ActionDown},
		{"NearCenterUp", ox - 1, oy - 2, game.ActionUp},
		{"ExactCenter", ox, oy, game.ActionUp},
	}

	for _, tc := range cases {
		got := ComputeMouseCardinalDir(tc.x, tc.y, ox, oy)
		if got != tc.want {
			t.Errorf("%s: ComputeMouseCardinalDir(%v,%v) = %d, want %d", tc.name, tc.x, tc.y, got, tc.want)
		}
	}
}

func TestResolveStopMouseInput(t *testing.T) {
	const sx, sy = float32(320), float32(180)

	// Phase 0: outside stop -> never drives minigame or suppresses throws
	dir, held, hover, suppress := ResolveStopMouseInput(0, sx+20, sy, sx, sy, true, true)
	if dir != game.ActionNone || held || hover != game.ActionNone || suppress {
		t.Fatalf("phase 0: got (%d,%v,%d,%v), want (-1,false,-1,false)", dir, held, hover, suppress)
	}

	// Phase 1: LMB click & hold ALWAYS drive minigame even far (>250px) from stop center
	dir, held, hover, suppress = ResolveStopMouseInput(1, sx+280, sy+120, sx, sy, true, true)
	if dir != game.ActionUp || !held || !suppress {
		t.Fatalf("phase 1 far click: got dir=%d held=%v suppress=%v, want ActionUp/true/true", dir, held, suppress)
	}
	dir, held, _, suppress = ResolveStopMouseInput(1, sx-260, sy-140, sx, sy, false, true)
	if dir != game.ActionNone || !held || suppress {
		t.Fatalf("phase 1 far hold: got dir=%d held=%v suppress=%v, want ActionNone/true/false", dir, held, suppress)
	}

	// Phase 2: live hover direction BEFORE click, and click maps to hovered quadrant + suppresses throw even far away
	dir, held, hover, suppress = ResolveStopMouseInput(2, sx, sy+220, sx, sy, false, false)
	if dir != game.ActionNone || held || hover != game.ActionDown || suppress {
		t.Fatalf("phase 2 hover down: got (%d,%v,%d,%v), want (-1,false,ActionDown,false)", dir, held, hover, suppress)
	}

	dir, _, hover, suppress = ResolveStopMouseInput(2, sx-240, sy+20, sx, sy, true, true)
	if dir != game.ActionLeft || hover != game.ActionLeft || !suppress {
		t.Fatalf("phase 2 far left click: got dir=%d hover=%d suppress=%v, want ActionLeft/ActionLeft/true", dir, hover, suppress)
	}
}

func TestResolveStopKeyboardFallback(t *testing.T) {
	// Outside active stop: never triggers action or held
	dir, held := ResolveStopKeyboardFallback(0, true, false, false, false, true)
	if dir != game.ActionNone || held {
		t.Fatalf("phase 0: got (%d, %v), want (-1, false)", dir, held)
	}

	// Inside active stop: Up/Right/Down/Left (Arrows / IJKL) map cleanly
	cases := []struct {
		name               string
		up, rt, dn, lf, hd bool
		wantDir            int
		wantHeld           bool
	}{
		{"Up_I", true, false, false, false, true, game.ActionUp, true},
		{"Right_L", false, true, false, false, true, game.ActionRight, true},
		{"Down_K", false, false, true, false, true, game.ActionDown, true},
		{"Left_J", false, false, false, true, true, game.ActionLeft, true},
		{"HeldOnly", false, false, false, false, true, game.ActionNone, true},
		{"None", false, false, false, false, false, game.ActionNone, false},
	}

	for _, phase := range []int{1, 2} {
		for _, tc := range cases {
			gotDir, gotHeld := ResolveStopKeyboardFallback(phase, tc.up, tc.rt, tc.dn, tc.lf, tc.hd)
			if gotDir != tc.wantDir || gotHeld != tc.wantHeld {
				t.Errorf("phase %d %s: got (%d, %v), want (%d, %v)", phase, tc.name, gotDir, gotHeld, tc.wantDir, tc.wantHeld)
			}
		}
	}
}
