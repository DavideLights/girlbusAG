#!/usr/bin/env python3
"""
GIRLBUS: NEON STOP (girlbusAG) — Costume Skins & Gold Bus Pixel-Art Generator
Generates 32x32 character skin frames (8 poses each) in assets/sprites/characters/:
  - alfredo_drip_<pose>.png       (Neon Drip)
  - alfredo_banana_<pose>.png     (Alfredo Banana)
  - alfredo_chicken_<pose>.png    (Alfredo Aletta di Pollo)
  - alfredo_samurai_<pose>.png    (Cyber-Samurai Zero)
And the 96x48 Gold Bus in assets/sprites/world/:
  - bus_gold.png                  (Autobus Dorato Drift)

Plus 8x/5x inspection previews in tools/previews/.
"""

import os
import sys
from typing import Dict, List, Tuple

TOOLS_DIR = os.path.dirname(os.path.abspath(__file__))
ROOT_DIR = os.path.dirname(TOOLS_DIR)
if TOOLS_DIR not in sys.path:
    sys.path.insert(0, TOOLS_DIR)

from gen_characters import (  # noqa: E402
    ALFREDO_DASH,
    ALFREDO_IDLE_0,
    ALFREDO_IDLE_1,
    ALFREDO_RUN_0,
    ALFREDO_RUN_1,
    ALFREDO_RUN_2,
    ALFREDO_RUN_3,
    ALFREDO_THROW,
)
from pixel_canvas import (  # noqa: E402
    PALETTE,
    Canvas,
    RGBA,
    grid_to_canvas,
    hex_to_rgba,
    save_asset_and_preview,
)

CHAR_OUT_DIR = os.path.join(ROOT_DIR, "assets", "sprites", "characters")
WORLD_OUT_DIR = os.path.join(ROOT_DIR, "assets", "sprites", "world")

POSES = ["idle_0", "idle_1", "run_0", "run_1", "run_2", "run_3", "dash", "throw"]

# Extended unified palette for all 4 costume skins + Gold Drift Bus
# NOTE: No conflicting overrides on PALETTE keys needed by shared layers
# (O, o, S, c, L, m, W, w, H, h, A, E, e, P, p, R, r, X, 2, 3, F, s).
SKIN_PALETTE: Dict[str, RGBA] = {
    **PALETTE,
    # Neon Drip (Vibrant magenta-purple cyber jacket + electric cyan visor/lapels)
    "M": hex_to_rgba("#DC269C"),   # Vibrant magenta-purple jacket base
    "m": hex_to_rgba("#BA665A"),   # Keep Alfredo's warm terracotta smirk!
    "n": hex_to_rgba("#8E1368"),   # Magenta-purple jacket shadow
    "N": hex_to_rgba("#FF72DC"),   # Magenta-purple jacket highlight

    # Banana Suit (Clean 3-tone sunny yellow + dark brown stem)
    "y": hex_to_rgba("#FFE135"),   # Banana vibrant yellow mid
    "Y": hex_to_rgba("#FFF59D"),   # Banana bright lemon highlight
    "d": hex_to_rgba("#E0A214"),   # Banana warm golden shadow
    "K": hex_to_rgba("#6B4226"),   # Banana brown stem base
    "k": hex_to_rgba("#422613"),   # Banana brown stem shadow

    # Alfredo Aletta di Pollo (Warm 3-tone crispy golden fried breading)
    "C": hex_to_rgba("#F4BE58"),   # Crispy fried golden highlight / crumb
    "c": hex_to_rgba("#EEA98C"),   # Keep skin warm shadow!
    "7": hex_to_rgba("#D4822A"),   # Golden fried breading mid
    "f": hex_to_rgba("#9E5216"),   # Deep fried crust shadow

    # Cyber-Samurai Zero (Sleek obsidian armor + crimson menpo/chest + gold plates)
    "Z": hex_to_rgba("#28192C"),   # Obsidian armor base
    "z": hex_to_rgba("#160C18"),   # Obsidian armor shadow
    "V": hex_to_rgba("#4B3354"),   # Obsidian armor specular edge
    "v": hex_to_rgba("#B81D3A"),   # Dark crimson armor plate
    "T": hex_to_rgba("#7D0F24"),   # Crimson plate shadow
    "G": hex_to_rgba("#FFD166"),   # Gold shoulder / horn highlight
    "D": hex_to_rgba("#E09F24"),   # Deep gold shadow
    "g": hex_to_rgba("#FFF0A6"),   # Gold bright specular (Gold Bus AC / trim)

    # Autobus Dorato Drift extras
    "0": hex_to_rgba("#A36A10"),   # Dark metallic gold trim shadow
    "1": hex_to_rgba("#FF6B2B"),   # Blazing orange racing flame
    "4": hex_to_rgba("#222A3B"),   # Dark tinted bus glass
    "5": hex_to_rgba("#3A4761"),   # Bus glass reflection streak
    "6": hex_to_rgba("#FFE5B8"),   # Warm cream bus trim
}


def build_grid(grid: List[str], name: str, w: int = 32, h: int = 32) -> Canvas:
    if len(grid) != h:
        raise ValueError(f"{name}: expected {h} rows, got {len(grid)}")
    for idx, row in enumerate(grid):
        if len(row) != w:
            raise ValueError(f"{name} row {idx:02d}: expected {w} cols, got {len(row)}: '{row}'")
    return grid_to_canvas(grid, w, h, SKIN_PALETTE)


def blank_buf(w: int = 32, h: int = 32) -> List[List[str]]:
    return [["." for _ in range(w)] for _ in range(h)]


def stamp(buf: List[List[str]], lines: List[str], ox: int = 0, oy: int = 0) -> None:
    h = len(buf)
    w = len(buf[0])
    for y, row in enumerate(lines):
        py = oy + y
        if not (0 <= py < h):
            continue
        for x, ch in enumerate(row):
            px = ox + x
            if 0 <= px < w and ch != ".":
                buf[py][px] = ch


def buf_to_grid(buf: List[List[str]]) -> List[str]:
    return ["".join(row) for row in buf]


def stamp_rounded_hand(
    buf: List[List[str]],
    ox: int,
    oy: int,
    hi: str = "S",
    sh: str = "c",
) -> None:
    """Stamps a clean 4x4 rounded hand/gauntlet peeking out of a costume side."""
    hand = [
        ".OO.",
        f"O{hi}{hi}O",
        f"O{sh}{sh}O",
        ".OO.",
    ]
    stamp(buf, hand, ox=ox, oy=oy)


def add_beacon_and_throw_arm(
    buf: List[List[str]],
    sleeve_col: str,
    sleeve_hi: str,
    sleeve_sh: str,
) -> None:
    """Draws Alfredo's Bus Stop beacon & raised right arm for throw pose."""
    beacon_and_arm = [
        "......3X..",  # 11
        ".....OXXO.",  # 12
        "....OXFFXO",  # 13
        "....OXFFXO",  # 14
        ".....O22O.",  # 15
        "....OSSO..",  # 16
        "...O" + sleeve_hi + "cO...",  # 17
        "..O" + sleeve_col + sleeve_sh + "O....",  # 18
        ".OO" + sleeve_col + sleeve_sh + "O....",  # 19
        "O" + sleeve_col + sleeve_col + sleeve_sh + "OO....",  # 20
    ]
    stamp(buf, beacon_and_arm, ox=22, oy=11)


def add_dash_trails(buf: List[List[str]]) -> None:
    """Draws Chronos neon cyan speed streaks behind the dashing character."""
    trails = [
        (3, 4, "33X2"),
        (4, 2, "3XXXX22"),
        (6, 4, "3X2"),
        (8, 2, "3XXXX22"),
        (9, 6, "22"),
        (11, 1, "3XXXXX22"),
        (12, 3, "22"),
        (15, 4, "3XX2"),
        (18, 2, "3XXXXX22"),
        (19, 4, "22"),
        (20, 8, "32"),
        (21, 2, "3XXX22"),
        (23, 5, "3XX2"),
        (25, 2, "3XXXX2"),
    ]
    for ry, rx, s in trails:
        for i, ch in enumerate(s):
            if 0 <= rx + i < 32 and buf[ry][rx + i] == ".":
                buf[ry][rx + i] = ch


# ==============================================================================
# 1. NEON DRIP (alfredo_drip_<pose>.png)
# Starts from Alfredo Base's exact head/hair/face/body silhouette.
# Sharp neon cyan visor shades (1px dark frame + solid cyan + 2px white glint),
# vibrant magenta-purple jacket with clean cyan collar/lapels, matching sneakers.
# ==============================================================================
def make_neon_drip_frame(base_grid: List[str], pose: str) -> List[str]:
    """Converts Alfredo Base's exact 32x32 frame into Neon Drip cyber gear."""
    buf = [list(row) for row in base_grid]

    # 1. Recolor jacket (J->M, j->n, B->N) and sneakers (R->M, r->n, lace W->X)
    #    while keeping hair (H/h/A), white tee (W/w), and sneaker soles (OWwWO) crisp.
    for y in range(32):
        for x in range(32):
            ch = buf[y][x]
            if 17 <= y <= 24:
                if ch == "J":
                    buf[y][x] = "M"
                elif ch == "j":
                    buf[y][x] = "n"
                elif ch == "B":
                    buf[y][x] = "N"
            elif 27 <= y <= 30:
                if ch == "R":
                    buf[y][x] = "M"
                elif ch == "r":
                    buf[y][x] = "n"
                elif ch == "W":
                    # Only turn sneaker upper lace (between R's in ORWRO) into cyan X
                    left_r = x > 0 and base_grid[y][x - 1] == "R"
                    right_r = x < 31 and base_grid[y][x + 1] == "R"
                    if left_r and right_r:
                        buf[y][x] = "X"

    # 2. Clean cyan collar & lapels flanking the crisp white inner tee (oW...Wo)
    collar_left_x = -1
    collar_right_x = -1
    for y in range(17, 24):
        row_str = base_grid[y]
        if "oW" not in row_str:
            continue
        left_o = row_str.index("oW")
        right_o = row_str.find("o", left_o + 2)
        if left_o >= 3 and row_str[left_o - 3:left_o] == "JBB":
            # Collar row: clean symmetrical flared cyan collar tips (XX left, X2 right)
            collar_left_x = left_o - 1
            buf[y][left_o - 2] = "X"
            buf[y][left_o - 1] = "X"
            if right_o != -1 and right_o + 2 < 32:
                collar_right_x = right_o + 1
                buf[y][right_o + 1] = "X"
                if base_grid[y][right_o + 2] in ("J", "j"):
                    buf[y][right_o + 2] = "2"
        else:
            # Lapels below collar: crisp straight cyan lapel line matching collar column
            lx = collar_left_x if collar_left_x != -1 else left_o - 1
            if lx == left_o - 1 and 0 <= lx < 32 and base_grid[y][lx] in ("B", "J"):
                buf[y][lx] = "X"
                if lx - 1 >= 0 and buf[y][lx - 1] == "M" and y in (18, 19, 20):
                    buf[y][lx - 1] = "N"
            rx = collar_right_x if collar_right_x != -1 else (right_o + 1 if right_o != -1 else -1)
            if right_o != -1 and rx == right_o + 1 and 0 <= rx < 32 and base_grid[y][rx] in ("J", "j"):
                buf[y][rx] = "X"

    # 3. Sharp neon cyan visor shades over Alfredo's eyes (keep clean face & smirk!)
    eye_row = -1
    idx = -1
    for y in range(6, 17):
        row_str = "".join(buf[y])
        if "SWESSSSWES" in row_str:
            eye_row = y
            idx = row_str.index("SWESSSSWES")
            break

    if eye_row != -1:
        # Clear eyebrows on eye_row-2 to clean forehead skin so the 1px dark visor
        # frame stands out razor-sharp with zero clumping.
        if eye_row - 2 >= 0:
            for dx in (1, 2, 3, 6, 7, 8):
                if buf[eye_row - 2][idx + dx] == "O":
                    buf[eye_row - 2][idx + dx] = "S"

        # Top 1px dark frame (cols idx-1 .. idx+11)
        for dx, ch in enumerate("OOOOOOOOOOOOO"):
            buf[eye_row - 1][idx - 1 + dx] = ch
        # Upper lens: solid bright cyan + crisp 2px white highlight glint
        for dx, ch in enumerate("OXWWXXXXXXXXO"):
            buf[eye_row][idx - 1 + dx] = ch
        # Lower lens: sharp 1px tapered sporty shades silhouette
        for dx, ch in enumerate("OOXXXXXXXXXOO"):
            buf[eye_row + 1][idx - 1 + dx] = ch
        # Bottom 1px dark frame + clean warm cheek corners separating ears cleanly
        for dx, ch in enumerate("OSOOOOOOOOOcO"):
            buf[eye_row + 2][idx - 1 + dx] = ch

    return buf_to_grid(buf)


# ==============================================================================
# Shared Animated Legs & Sneakers (exact Alfredo Base jeans & red sneakers)
# ==============================================================================
LEGS_IDLE: List[str] = [
    "..........OPPPpOOOPppO..........",  # 25
    "..........OPPPpOOOPppO..........",  # 26
    "..........OPPPpOOOPppO..........",  # 27
    "..........ORWRO..ORWRO..........",  # 28
    "........ssOWwWOssOWwWOss........",  # 29
    ".........ssOOOssssOOOss.........",  # 30
    "..........ssssssssssss..........",  # 31
]

LEGS_RUN_0: List[str] = [
    ".........OPPPpOOOPppO...........",  # 25
    ".........OPPPpO.OOPppO..........",  # 26
    ".........ORWRO...OPppO..........",  # 27
    "........sOWwWOs..ORWRO..........",  # 28
    "........ssOOOss.ssOWwWOss.......",  # 29
    "...................OOOss........",  # 30
    ".........ssssssssssssss.........",  # 31
]

LEGS_RUN_1: List[str] = [
    "...........OPPPpOPppO...........",  # 24
    "...........OPPPpOPppO...........",  # 25
    "...........OPPPpOPppO...........",  # 26
    "...........OPPPpOPppO...........",  # 27
    "...........ORWROORWRO...........",  # 28
    "........sssOWwWOOWwWOsss........",  # 29
    ".........sssOOOssOOOsss.........",  # 30
    "...........ssssssssss...........",  # 31
]

LEGS_RUN_2: List[str] = [
    "...........OPPPpOOOPppO.........",  # 25
    "..........OPPPpOO.OPppO.........",  # 26
    "..........OPPPpO...ORWRO........",  # 27
    "..........ORWRO..ssOWwWOs.......",  # 28
    ".......ssOWwWOss..ssOOOss.......",  # 29
    "........ssOOOss.................",  # 30
    ".........ssssssssssssss.........",  # 31
]

LEGS_RUN_3: List[str] = [
    "..........OPPPpOOOPppO..........",  # 24
    "..........OPPPpOOOPppO..........",  # 25
    "..........OPPPpOOOPppO..........",  # 26
    "..........ORWRO..OPppO..........",  # 27
    ".........sOWwWOs.ORWRO..........",  # 28
    ".........ssOOOss.OWwWOss........",  # 29
    "..................OOOss.........",  # 30
    "...........ssssssssss...........",  # 31
]

LEGS_DASH: List[str] = [
    ".............OPPPpOOOOOPppO.....",  # 23
    "............OPPPpO.....OPppO....",  # 24
    "...........OPPPpO.......OPppO...",  # 25
    "..........OPPPpO.........OPppO..",  # 26
    ".........ORWRO...........ORWRO..",  # 27
    "........sOWwWOs.........sOWwWOs.",  # 28
    "........ssOOOss.........ssOOOss.",  # 29
    "................................",  # 30
    "........sssssssssssssssssssss...",  # 31
]

LEGS_THROW: List[str] = [
    ".........OPPPpOOOPppO...........",  # 24
    ".........OPPPpO.OOPppO..........",  # 25
    "........OPPPpO...OOPppO.........",  # 26
    "........OPPPpO...OOPppO.........",  # 27
    "........ORWRO.....ORWRO.........",  # 28
    "......ssOWwWOss.ssOWwWOss.......",  # 29
    ".......ssOOOss...ssOOOss........",  # 30
    ".......ssssssssssssssssss.......",  # 31
]


# ==============================================================================
# 2. ALFREDO BANANA (alfredo_banana_<pose>.png)
# Smooth curved yellow banana suit wrapping cleanly around Alfredo's exact
# FULL-WIDTH face/eyes/smirk (4px eye bridge, full eyebrows, full smirk).
# Clean 3-tone yellow shading (Y / y / d), neat dark-brown stem curving at top,
# clean rounded hands peeking out the sides, animated jeans + red sneakers.
# ==============================================================================
BANANA_BODY_UPRIGHT: List[str] = [
    "................................",  # 00 Top margin (unclipped on bounce up)
    "..................OOO...........",  # 01 Neat dark-brown stem tip curving right
    ".................OKkO...........",  # 02 Curved brown stem base
    "..............OOYYydO...........",  # 03 Banana crown curve (14..20)
    "............OOYYYYyydO..........",  # 04 Smooth curved yellow top (12..21)
    ".........OOYYYYyyyyyyddO........",  # 05 Smooth top swell (9..23)
    "........OYYYOOOOOOOOOyddO.......",  # 06 Oval cutout top rim (8..24)
    ".......OYYYOHHHHHHHHHOyddO......",  # 07 Cropped hair inside oval cutout (7..25)
    "......OYYYOHSLLLLSSScHOydO......",  # 08 Forehead + highlight + sideburns (6..25)
    "......OYYyOSOOOSSOOOScOydO......",  # 09 Thick expressive eyebrows (full width!)
    "......OYYyOSSSSSSSSSScOydO......",  # 10 Upper eyelids
    "......OYYyOSWESSSSWEScOydO......",  # 11 Eyes top (exact 4px bridge SSSS!)
    "......OYYyOSEeSSSSeeScOydO......",  # 12 Eyes mid
    "......OYYyOSeeSSSSeeScOydO......",  # 13 Eyes bottom
    "......OYYyOSSSSSSSSSScOydO......",  # 14 Clean cheeks
    "......OYYyOcSSmmmmoSccOydO......",  # 15 Signature confident smirk!
    ".......OYYYOcccccccccOyddO......",  # 16 Chin inside oval cutout (7..25, smooth!)
    ".......OYYYYOOOOOOOOOyddO.......",  # 17 Cutout bottom rim (7..24)
    ".......OYYYYyyyyyyyyyyddO.......",  # 18 Smooth curved banana torso (7..24)
    "........OYYYyyyyyyyyyyddO.......",  # 19 (8..24)
    "........OYYYyyyyyyyyyddO........",  # 20 (8..23)
    ".........OYYYyyyyyyyyddO........",  # 21 (9..23)
    ".........OYYyyyyyyydddO.........",  # 22 (9..22)
    "..........OyyddddddddO..........",  # 23 Banana bottom rim (10..21)
    "..........OOOOOOOOOOOO..........",  # 24 Clean hem meeting jeans (10..21)
]


def _animate_suited_character(
    pose: str,
    body: List[str],
    sleeve_col: str,
    sleeve_hi: str,
    sleeve_sh: str,
    right_hand_dx: int = 0,
) -> List[str]:
    """Animates full-body costume suits (Banana / Chicken) with clean rounded hands & legs."""
    buf = blank_buf(32, 32)
    rx = 23 + right_hand_dx
    if pose == "idle_0":
        stamp(buf, body, ox=0, oy=0)
        stamp_rounded_hand(buf, 5, 19)
        stamp_rounded_hand(buf, rx, 19)
        stamp(buf, LEGS_IDLE, ox=0, oy=25)
    elif pose == "idle_1":
        stamp(buf, body, ox=0, oy=1)
        stamp_rounded_hand(buf, 5, 20)
        stamp_rounded_hand(buf, rx, 20)
        stamp(buf, LEGS_IDLE, ox=0, oy=25)
    elif pose == "run_0":
        stamp(buf, body, ox=0, oy=0)
        stamp_rounded_hand(buf, 5, 18)
        stamp_rounded_hand(buf, rx, 20)
        stamp(buf, LEGS_RUN_0, ox=0, oy=25)
    elif pose == "run_1":
        stamp(buf, body, ox=0, oy=-1)
        stamp_rounded_hand(buf, 6, 18)
        stamp_rounded_hand(buf, rx - 1, 18)
        stamp(buf, LEGS_RUN_1, ox=0, oy=24)
    elif pose == "run_2":
        stamp(buf, body, ox=0, oy=0)
        stamp_rounded_hand(buf, 5, 20)
        stamp_rounded_hand(buf, rx, 18)
        stamp(buf, LEGS_RUN_2, ox=0, oy=25)
    elif pose == "run_3":
        stamp(buf, body, ox=0, oy=-1)
        stamp_rounded_hand(buf, 5, 18)
        stamp_rounded_hand(buf, rx, 18)
        stamp(buf, LEGS_RUN_3, ox=0, oy=24)
    elif pose == "dash":
        stamp(buf, body, ox=3, oy=-1)
        stamp_rounded_hand(buf, 8, 18)
        stamp_rounded_hand(buf, rx + 3, 18)
        stamp(buf, LEGS_DASH, ox=0, oy=23)
        add_dash_trails(buf)
    elif pose == "throw":
        stamp(buf, body, ox=0, oy=0)
        stamp_rounded_hand(buf, 5, 19)
        stamp(buf, LEGS_THROW, ox=0, oy=24)
        add_beacon_and_throw_arm(buf, sleeve_col=sleeve_col, sleeve_hi=sleeve_hi, sleeve_sh=sleeve_sh)
    return buf_to_grid(buf)


def make_banana_frame(pose: str) -> List[str]:
    return _animate_suited_character(pose, BANANA_BODY_UPRIGHT, sleeve_col="y", sleeve_hi="Y", sleeve_sh="d", right_hand_dx=0)


# ==============================================================================
# 3. ALFREDO ALETTA DI POLLO (alfredo_chicken_<pose>.png)
# Clean, appetizing fried chicken drumstick/wing costume:
# Smooth chunky scalloped outer outline, warm 3-tone golden fried breading
# (#F4BE58 / #D4822A / #9E5216) with only 4 deliberate crisp crumb highlights,
# chunky white double-knob bone sticking out top-right with crisp O outline,
# Alfredo's exact full-width face/smirk framed cleanly in the center cutout.
# ==============================================================================
CHICKEN_BODY_UPRIGHT: List[str] = [
    "................................",  # 00 Top margin (unclipped on bounce up)
    ".....................OOO........",  # 01 Chunky white bone upper knob top (21..23)
    "....................OWWWO.......",  # 02 Bone upper knob (20..24)
    "............OOOOOOOOWWWwOOO.....",  # 03 Drumstick dome top + double-knob dimple (12..27)
    "..........OOCCCCC77OWWwwWWwO....",  # 04 Golden fried dome + bone shaft & lower knob (10..28)
    ".........OC77777777OWwOOwwwO....",  # 05 Bone shaft + underside notch + lower knob (9..27)
    "........OC77OOOOOOOOO777OOOO....",  # 06 Cutout top rim + golden crust under bone (8..26)
    ".......OC77OHHHHHHHHHO77ffO.....",  # 07 Cropped hair + upper puff 1 (7..26)
    "......OC777OHSLLLLSSScHO7fO.....",  # 08 Forehead + sideburns (6..26)
    "......OC7CCOSOOOSSOOOScO7fO.....",  # 09 Eyebrows + deliberate crumb #1 (6..26)
    ".......OC77OSSSSSSSSSScO77O.....",  # 10 Smooth scalloped tuck (7..25)
    "......OC777OSWESSSSWEScO77fO....",  # 11 Full-width eyes + plump mid puff 2 (6..26)
    ".....OC7777OSEeSSSSeeScO7CCfO...",  # 12 Eyes mid + deliberate crumb #2 (5..27)
    ".....OC7777OSeeSSSSeeScO77ffO...",  # 13 Eyes bottom (5..27)
    ".....OC7777OSSSSSSSSSScO77ffO...",  # 14 Clean cheeks (5..27)
    "......OC777OcSSmmmmoSccO7ffO....",  # 15 Signature smirk! (6..26)
    ".......OC77OcccccccccO77ffO.....",  # 16 Chin inside cutout (7..26)
    "......OC7777OOOOOOOOO7777ffO....",  # 17 Cutout bottom rim + plump belly puff 3 (6..27)
    "......OC7CC7777777777777fffO....",  # 18 Plump golden belly + deliberate crumb #3 (6..27)
    ".......OC7777777777CC77fffO.....",  # 19 Deliberate crumb #4 (7..26)
    ".......OC7777777777777fffO......",  # 20 Smooth belly lower curve (7..25)
    "........OC7777777777ffffO.......",  # 21 (8..24)
    ".........OC7777777ffffO.........",  # 22 Roasted underside shadow (9..22)
    "..........O777fffffffO..........",  # 23 Suit hem curve (10..21)
    "..........OOOOOOOOOOOO..........",  # 24 Clean hem meeting jeans (10..21)
]


def make_chicken_frame(pose: str) -> List[str]:
    return _animate_suited_character(pose, CHICKEN_BODY_UPRIGHT, sleeve_col="7", sleeve_hi="C", sleeve_sh="f", right_hand_dx=1)


# ==============================================================================
# 4. CYBER-SAMURAI ZERO (alfredo_samurai_<pose>.png)
# Sleeker, properly proportioned crimson-obsidian armor matching Alfredo Base's
# head & torso scale: dark kabuto helmet with crisp curved gold horns,
# 10px glowing cyan visor slit, clean gold pauldrons, neat katana hilt behind
# left shoulder (zero beacon/helmet collision across all 8 poses).
# ==============================================================================
SAMURAI_HEAD_UPRIGHT: List[str] = [
    "................................",  # 00 Top margin (unclipped on bounce up)
    "........OO............OO........",  # 01 Sharp curved gold horn tips
    "........OGO..OOOOOO..OGO........",  # 02 Horn upper curve + kabuto ridge
    "........OGDOOVVVVVVOODGO........",  # 03 Horn body + obsidian dome specular
    ".........OGGDZZZZZZDGGO.........",  # 04 Horns joining forehead crest
    ".........OZZZGGGGGGZZZO.........",  # 05 Sleek obsidian kabuto + gold crest
    "........OVVZZZZZZZZZZZzO........",  # 06 Obsidian dome (16px wide, matching Alfredo!)
    "........OVZZZZZZZZZZZZzO........",  # 07 Kabuto upper brim
    ".......OZZVZZZZZZZZZZzZZO.......",  # 08 Kabuto side flaps (clean dark obsidian)
    ".......OZZOOOOOOOOOOOOZZO.......",  # 09 1px dark visor brow frame
    ".......OZZO3XXXXXXXX2OZZO.......",  # 10 Razor-sharp glowing cyan visor slit!
    ".......OZZOOOOOOOOOOOOZZO.......",  # 11 1px dark visor bottom frame
    "........OOZvvvvvvvvvvZOO........",  # 12 Crimson menpo face mask top
    ".........OZvvvTTTTvvvZO.........",  # 13 Menpo cheek/nose plating
    ".........OvvvvTTTTvvvvO.........",  # 14 Menpo mid
    "..........OvvTTTTTTvvO..........",  # 15 Menpo lower jaw
    "...........OOTTTTTTOO...........",  # 16 Menpo chin guard
    "............OOOOOOOO............",  # 17 Chin outline
]

# Neat diagonal katana hilt peeking cleanly behind the left shoulder
SAMURAI_KATANA_LEFT: List[str] = [
    "OO.....",  # 13 Pommel tip
    "OXGO...",  # 14 Gold pommel + cyan wrap accent
    ".OZZO..",  # 15 Dark wrapped tsuka hilt
    ".OGGGO.",  # 16 Crisp gold tsuba guard
    "...OZZO",  # 17 Scabbard disappearing behind left pauldron
]

# Sleek crimson-obsidian torso (rows 18..24) without side fists for swing poses
SAMURAI_CHEST: List[str] = [
    ".........OGGOZvXXvZOGGO.........",  # 18 Gold pauldron tops + cyan chest core
    ".......OGGDOZvvvvvvZOGDDO.......",  # 19 Clean gold pauldrons + crimson armor
    "........OZZOZvvGGvvZOZZO........",  # 20 Obsidian sleeves + gold chest crest
    "........OZZOZvvvvvvZOZZO........",  # 21 Obsidian sleeves + crimson breastplate
    "........OVVOZvvTTvvZOVVO........",  # 22 Gauntlet cuffs + lower chest shadow
    "..........OOGGDDDDGGOO..........",  # 23 Gold samurai belt
    "..........OOZZZZZZZZOO..........",  # 24 Armored waist hem (cols 10..21)
]

SAMURAI_LEGS_IDLE: List[str] = [
    "..........OZZzOOOZzzO...........",  # 25 Obsidian armored thighs
    "..........OZZzOOOZzzO...........",  # 26
    "..........OvvTO..OvvTO..........",  # 27 Crimson shin plates
    "..........OVZzO..OVZzO..........",  # 28 Obsidian sabaton boots
    "........ssOZZzOssOZZzOss........",  # 29
    ".........ssOOOssssOOOss.........",  # 30
    "..........ssssssssssss..........",  # 31
]

SAMURAI_LEGS_RUN_0: List[str] = [
    ".........OZZzOOOZzzO............",  # 25
    ".........OZZzO.OOZzzO...........",  # 26
    ".........OvvTO...OvvTO..........",  # 27
    "........sOVZzOs..OVZzO..........",  # 28
    "........ssOOOss.ssOZZzOss.......",  # 29
    "...................OOOss........",  # 30
    ".........ssssssssssssss.........",  # 31
]

SAMURAI_LEGS_RUN_1: List[str] = [
    "...........OZZzOZzzO............",  # 24
    "...........OZZzOZzzO............",  # 25
    "...........OvvTOvvTO............",  # 26
    "...........OvvTOvvTO............",  # 27
    "...........OVZzOVZzO............",  # 28
    "........sssOZZzOZZzOsss.........",  # 29
    ".........sssOOOssOOOsss.........",  # 30
    "...........ssssssssss...........",  # 31
]

SAMURAI_LEGS_RUN_2: List[str] = [
    "...........OZZzOOOZzzO..........",  # 25
    "..........OZZzOO.OZzzO..........",  # 26
    "..........OvvTO...OvvTO.........",  # 27
    "..........OVZzO..sOVZzOs........",  # 28
    ".......ssOZZzOss..ssOOOss.......",  # 29
    "........ssOOOss.................",  # 30
    ".........ssssssssssssss.........",  # 31
]

SAMURAI_LEGS_RUN_3: List[str] = [
    "..........OZZzOOOZzzO...........",  # 24
    "..........OZZzOOOZzzO...........",  # 25
    "..........OvvTO..OvvTO..........",  # 26
    "..........OVZzO..OvvTO..........",  # 27
    ".........sOZZzOs.OVZzO..........",  # 28
    ".........ssOOOss.OZZzOss........",  # 29
    "..................OOOss.........",  # 30
    "...........ssssssssss...........",  # 31
]

SAMURAI_LEGS_DASH: List[str] = [
    ".............OZZzOOOOOZzzO......",  # 23
    "............OZZzO.....OZzzO.....",  # 24
    "...........OvvTO.......OvvTO....",  # 25
    "..........OZZzO.........OZZzO...",  # 26
    ".........OVZzO...........OVZzO..",  # 27
    "........sOZZzOs.........sOZZzOs.",  # 28
    "........ssOOOss.........ssOOOss.",  # 29
    "................................",  # 30
    "........sssssssssssssssssssss...",  # 31
]

SAMURAI_LEGS_THROW: List[str] = [
    ".........OZZzOOOZzzO............",  # 24
    ".........OZZzO.OOZzzO...........",  # 25
    "........OvvTO...OOvvTO..........",  # 26
    "........OZZzO....OOZZzO.........",  # 27
    "........OVZzO.....OVZzO.........",  # 28
    "......ssOZZzOss.ssOZZzOss.......",  # 29
    ".......ssOOOss...ssOOOss........",  # 30
    ".......ssssssssssssssssss.......",  # 31
]


def make_samurai_frame(pose: str) -> List[str]:
    buf = blank_buf(32, 32)
    if pose == "idle_0":
        stamp(buf, SAMURAI_KATANA_LEFT, ox=4, oy=13)
        stamp(buf, SAMURAI_CHEST, ox=0, oy=18)
        stamp_rounded_hand(buf, 8, 21, hi="V", sh="Z")
        stamp_rounded_hand(buf, 20, 21, hi="V", sh="Z")
        stamp(buf, SAMURAI_HEAD_UPRIGHT, ox=0, oy=0)
        stamp(buf, SAMURAI_LEGS_IDLE, ox=0, oy=25)
    elif pose == "idle_1":
        stamp(buf, SAMURAI_KATANA_LEFT, ox=4, oy=14)
        stamp(buf, SAMURAI_CHEST, ox=0, oy=19)
        stamp_rounded_hand(buf, 8, 22, hi="V", sh="Z")
        stamp_rounded_hand(buf, 20, 22, hi="V", sh="Z")
        stamp(buf, SAMURAI_HEAD_UPRIGHT, ox=0, oy=1)
        stamp(buf, SAMURAI_LEGS_IDLE, ox=0, oy=25)
    elif pose == "run_0":
        stamp(buf, SAMURAI_KATANA_LEFT, ox=4, oy=13)
        stamp(buf, SAMURAI_CHEST, ox=0, oy=18)
        stamp_rounded_hand(buf, 7, 20, hi="V", sh="Z")
        stamp_rounded_hand(buf, 21, 22, hi="V", sh="Z")
        stamp(buf, SAMURAI_HEAD_UPRIGHT, ox=0, oy=0)
        stamp(buf, SAMURAI_LEGS_RUN_0, ox=0, oy=25)
    elif pose == "run_1":
        stamp(buf, SAMURAI_KATANA_LEFT, ox=4, oy=12)
        stamp(buf, SAMURAI_CHEST, ox=0, oy=17)
        stamp_rounded_hand(buf, 8, 20, hi="V", sh="Z")
        stamp_rounded_hand(buf, 20, 20, hi="V", sh="Z")
        stamp(buf, SAMURAI_HEAD_UPRIGHT, ox=0, oy=-1)
        stamp(buf, SAMURAI_LEGS_RUN_1, ox=0, oy=24)
    elif pose == "run_2":
        stamp(buf, SAMURAI_KATANA_LEFT, ox=4, oy=13)
        stamp(buf, SAMURAI_CHEST, ox=0, oy=18)
        stamp_rounded_hand(buf, 7, 22, hi="V", sh="Z")
        stamp_rounded_hand(buf, 21, 20, hi="V", sh="Z")
        stamp(buf, SAMURAI_HEAD_UPRIGHT, ox=0, oy=0)
        stamp(buf, SAMURAI_LEGS_RUN_2, ox=0, oy=25)
    elif pose == "run_3":
        stamp(buf, SAMURAI_KATANA_LEFT, ox=4, oy=12)
        stamp(buf, SAMURAI_CHEST, ox=0, oy=17)
        stamp_rounded_hand(buf, 7, 20, hi="V", sh="Z")
        stamp_rounded_hand(buf, 21, 20, hi="V", sh="Z")
        stamp(buf, SAMURAI_HEAD_UPRIGHT, ox=0, oy=-1)
        stamp(buf, SAMURAI_LEGS_RUN_3, ox=0, oy=24)
    elif pose == "dash":
        stamp(buf, SAMURAI_KATANA_LEFT, ox=7, oy=12)
        stamp(buf, SAMURAI_CHEST, ox=3, oy=17)
        stamp_rounded_hand(buf, 9, 20, hi="V", sh="Z")
        stamp_rounded_hand(buf, 24, 20, hi="V", sh="Z")
        stamp(buf, SAMURAI_HEAD_UPRIGHT, ox=3, oy=-1)
        stamp(buf, SAMURAI_LEGS_DASH, ox=0, oy=23)
        add_dash_trails(buf)
    elif pose == "throw":
        stamp(buf, SAMURAI_KATANA_LEFT, ox=4, oy=13)
        stamp(buf, SAMURAI_CHEST, ox=0, oy=18)
        stamp_rounded_hand(buf, 6, 20, hi="V", sh="Z")
        stamp(buf, SAMURAI_HEAD_UPRIGHT, ox=0, oy=0)
        stamp(buf, SAMURAI_LEGS_THROW, ox=0, oy=24)
        add_beacon_and_throw_arm(buf, sleeve_col="v", sleeve_hi="G", sleeve_sh="Z")
    return buf_to_grid(buf)


# ==============================================================================
# 5. AUTOBUS DORATO DRIFT (bus_gold.png, exact 96x48 matching bus_neon.png)
# Metallic gold bodywork, crimson/orange racing flame stripes,
# glowing cyan neon headlights and roof LED bar.
# ==============================================================================
BUS_GOLD_96: List[str] = [
    "................................................................................................",  # 00
    "................................................................................................",  # 01
    "................................................................................................",  # 02
    "................................................................................................",  # 03
    "................................................................................................",  # 04
    "..............OOOOOOOOOOOOOO..............OOOOOOOOOOOOOO........................................",  # 05 - Roof AC units top
    ".............OGgGGGGGGGGGGGGO............OGgGGGGGGGGGGGGO.......................................",  # 06 - Gold AC highlights
    ".............OGG444444GGGGDGO............OGG444444GGGGDGO.......................................",  # 07 - AC fan grilles
    ".............OGDDDDDDDDDDDDGO............OGDDDDDDDDDDDDGO.......................................",  # 08 - AC bases
    ".......OOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOO............",  # 09 - Bus roof outline
    "......O3XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX3O...........",  # 10 - Glowing Cyan Roof LED Bar!
    ".....OGggggGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGG4GGGGGG4gGGGO..........",  # 11 - Metallic gold roof + LED dest board
    "....OGGGGG4XXGXGGXXGXXGGGDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDD4GXXGXXG4GGGGGOO.......",  # 12 - Side LED 'BUS' + Front LED sign!
    "....OGGGGG444444444444444GGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGG444444444GGGGGGGO......",  # 13 - Gold upper band
    "...OGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGO.....",  # 14 - Metallic gold body top
    "...OGGGGO44444444444OO44444444444OO44444444444OO444444444OO44444444444OO4444444444444OGGGGO.....",  # 15 - Panoramic windows
    "...OGGGGO45555444444OO44555544444OO44555544444OO445445444OO44555544444OO4455555444444OGGGGO.....",  # 16 - Glass reflection streaks
    "...OGGGGO454SS444444OO44444M44444OO44444H44444OO445445444OO44444D44444OO44544444S4444OGGGGGO....",  # 17 - Chibi heads peeking inside windows!
    "...OGGGGO44SSc444444OO4444Mn44444OO4444HHh4444OO445445444OO4444GD44444OO44444444Sc444OGGGGGO....",  # 18
    "...OGGGGO44444444444OO44444444444OO44444444444OO444444444OO44444444444OO4444444444444OGGGGGO....",  # 19
    "...OGGGGO44444444444OO44444444444OO44444444444OO444444444OO44444444444OO4444444444444OGGGGGO....",  # 20
    "...OGGGGO44444455444OO44444445544OO44444445544OO444444444OO44444445544OO4444444455444OGGGGGO....",  # 21
    "...OGGGGO44444444444OO44444444444OO44444444444OO444444444OO44444444444OO4444444444444OGGGGGO....",  # 22
    "...OGGGGOOOOOOOOOOOOOO44444444444OOOOOOOOOOOOOOO444444444OOOOOOOOOOOOOOO4444444444444OGGGGGO....",  # 23 - Window sills
    "...OGGGGGGGGGG1RGGGGGO44444444444OGGGGGGGG1RGGGO444444444OGGGGGGGGGGG1RGGGGGGGG1RGGGGOGGGGGO....",  # 24 - Racing flame tips!
    "...OGGGG11RRRRRRGGGGGO44444444444OGG11RRRRRRGGGO444444444OGG11RRRRRRRRR11RRRRRRRGGGGGOGGGGGO....",  # 25 - Crimson/orange racing flame stripes!
    "...ORRRRRR11yyDDDDDDDO44444444444ORRRR11yyDDDDDO444444444ORRRRRR11yyDDDDDD11yyDDDDDDDO00000O....",  # 26 - Blazing flame sweep + gold shadow
    "...O33333333333333333O33333333333O3333333333333O333333333O333333333333333333333333333O33333O....",  # 27 - Neon cyan drift trim top
    "...OXXXXXXXXXXXXXXXXXOXXXXXXXXXXXOXXXXXXXXXXXXXOXXXXXXXXXOXXXXXXXXXXXXXXXXXXXXXXXXXXXOXXXXXO....",  # 28 - Neon cyan speed stripe core!
    "...ORRRRR1111111RRRRRO22222222222ORRRR11111RRRRO222222222ORRRRR111111111111111RRRRRRRO22222O....",  # 29 - Crimson/orange racing stripe lower!
    "...ODDDDDDDDDDDDDDDDDO44444444444ODDDDDDDDDDDDDO444444444ODDDDDDDDDDDDDDDDDDDDDDDDDDDOO3XX3O....",  # 30 - Gold skirt + Cyan Headlight top!
    "...ORRRDDDDDDDDDDDDDDO44444444444ODDDDDDDDDDDDDO444444444ODDDDDDDDDDDDDDDDDDDDDDDDDDDOOXXXXO....",  # 31 - Taillight (L) + Gold skirt + Cyan Headlight!
    "...ORrrDDDDDDDDDDDDDDO44444444444ODDDDDDDDDDDDDO444444444ODDDDDDDDDDDDDDDDDDDDDDDDDDDOO3XX3O....",  # 32 - Taillight + Gold skirt + Headlight glow!
    "...O00000000000000000O44444444444O0000000000000O444444444O000000000000000000000000000O00000O....",  # 33 - Lower metallic gold trim
    "...O00000000OOOOOOOOO0OOOOOOOOOOOO0000000000000OOOOOOOOOOO0000000000OOOOOOOO000000000OOOOOOO....",  # 34 - Wheel arches top
    "...O0000000OPPPPPPPPO0000000000000000000000000000000000000000000000OPPPPPPPPO00000000OGGGGGGGGO.",  # 35 - Wheel wells + Chrome/Gold bumper!
    "...O000000OPPPGGGGPPPO00000000000000000000000000000000000000000000OPPPGGGGPPPO0000000OGGGGGGGGO.",  # 36 - Rear & Front Wheels w/ Gold rims!
    "...OOOOOOOPPPGRRRRGPPPOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOPPPGRRRRGPPPOOOOOOOOssssssss..",  # 37 - Gold/Crimson drift rims
    "..........OPPGDRRDGPPO............................................OPPGDRRDGPPO..................",  # 38 - Tires & gold hubcaps
    "..........OPPGXXXXGPPO............................................OPPGXXXXGPPO..................",  # 39 - Neon cyan center cap!
    "..........OPPPGGGGPPPO............................................OPPPGGGGPPPO..................",  # 40 - Tires lower
    "...........OPPPPPPPPO..............................................OPPPPPPPPO...................",  # 41 - Tire tread bottom
    "...........ssOOOOOOss..............................................ssOOOOOOss...................",  # 42 - Tire outline + shadow
    ".........ssssssssssssss..........................................ssssssssssssss.................",  # 43 - Ground shadow
    "................................................................................................",  # 44
    "................................................................................................",  # 45
    "................................................................................................",  # 46
    "................................................................................................",  # 47
]


def main() -> None:
    os.makedirs(CHAR_OUT_DIR, exist_ok=True)
    os.makedirs(WORLD_OUT_DIR, exist_ok=True)

    generated: List[Tuple[str, int, int]] = []

    base_map = {
        "idle_0": ALFREDO_IDLE_0,
        "idle_1": ALFREDO_IDLE_1,
        "run_0": ALFREDO_RUN_0,
        "run_1": ALFREDO_RUN_1,
        "run_2": ALFREDO_RUN_2,
        "run_3": ALFREDO_RUN_3,
        "dash": ALFREDO_DASH,
        "throw": ALFREDO_THROW,
    }

    # 1. Neon Drip (8 poses)
    for pose in POSES:
        name = f"alfredo_drip_{pose}"
        grid = make_neon_drip_frame(base_map[pose], pose)
        canvas = build_grid(grid, name, 32, 32)
        save_asset_and_preview(canvas, CHAR_OUT_DIR, name, zoom=8)
        generated.append((os.path.join(CHAR_OUT_DIR, f"{name}.png"), 32, 32))

    # 2. Alfredo Banana (8 poses)
    for pose in POSES:
        name = f"alfredo_banana_{pose}"
        grid = make_banana_frame(pose)
        canvas = build_grid(grid, name, 32, 32)
        save_asset_and_preview(canvas, CHAR_OUT_DIR, name, zoom=8)
        generated.append((os.path.join(CHAR_OUT_DIR, f"{name}.png"), 32, 32))

    # 3. Alfredo Aletta di Pollo (8 poses)
    for pose in POSES:
        name = f"alfredo_chicken_{pose}"
        grid = make_chicken_frame(pose)
        canvas = build_grid(grid, name, 32, 32)
        save_asset_and_preview(canvas, CHAR_OUT_DIR, name, zoom=8)
        generated.append((os.path.join(CHAR_OUT_DIR, f"{name}.png"), 32, 32))

    # 4. Cyber-Samurai Zero (8 poses)
    for pose in POSES:
        name = f"alfredo_samurai_{pose}"
        grid = make_samurai_frame(pose)
        canvas = build_grid(grid, name, 32, 32)
        save_asset_and_preview(canvas, CHAR_OUT_DIR, name, zoom=8)
        generated.append((os.path.join(CHAR_OUT_DIR, f"{name}.png"), 32, 32))

    # 5. Autobus Dorato Drift (96x48)
    bus_canvas = build_grid(BUS_GOLD_96, "bus_gold", 96, 48)
    save_asset_and_preview(bus_canvas, WORLD_OUT_DIR, "bus_gold", zoom=5)
    generated.append((os.path.join(WORLD_OUT_DIR, "bus_gold.png"), 96, 48))

    print(f"Generated {len(generated)} assets successfully.")


if __name__ == "__main__":
    main()
