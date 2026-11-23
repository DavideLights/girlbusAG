#!/usr/bin/env python3
"""
GIRLBUS: NEON STOP (girlbusAG) — Character & Girl Pixel-Art Animator
Generates all 32x32 character frames, horizontal sprite sheets, 8x previews,
and the VIP aura overlay in assets/sprites/characters/.

Usage:
    python3 tools/gen_characters.py
"""

import os
import sys
from typing import Dict, List

TOOLS_DIR = os.path.dirname(os.path.abspath(__file__))
ROOT_DIR = os.path.dirname(TOOLS_DIR)
if TOOLS_DIR not in sys.path:
    sys.path.insert(0, TOOLS_DIR)

from pixel_canvas import (
    PALETTE,
    Canvas,
    grid_to_canvas,
    hex_to_rgba,
    pack_horizontal_strip,
    save_asset_and_preview,
)

OUT_DIR = os.path.join(ROOT_DIR, "assets", "sprites", "characters")
PREVIEW_DIR = os.path.join(ROOT_DIR, "tools", "previews", "characters")


def build_frame(grid: List[str], name: str) -> Canvas:
    if len(grid) != 32:
        raise ValueError(f"{name}: expected 32 rows, got {len(grid)}")
    for idx, row in enumerate(grid):
        if len(row) != 32:
            raise ValueError(f"{name} row {idx:02d}: expected 32 cols, got {len(row)}: '{row}'")
    return grid_to_canvas(grid, 32, 32, PALETTE)


# ==============================================================================
# 1. ALFREDO PROTAGONISTA BASE (32x32, 8 frames, NO graduation elements)
# ==============================================================================
ALFREDO_IDLE_0: List[str] = [
    "................................",  # 00
    "................................",  # 01
    "..........OOOOOOOOOOOO..........",  # 02 Top of short cropped hair
    ".........OAAHHHHHHHHhhO.........",  # 03 Hair highlight top-left
    "........OAAHHHHHHHHHHhhO........",  # 04 Short dark hair volume
    "........OAHHHHHHHHHHHhhO........",  # 05 Hair mid
    "........OHHHHHHHHHHHHHhO........",  # 06 Neat short front fringe
    "........OHHhSSSSSSSShHhO........",  # 07 Clean hairline with temple taper
    "........OHhSLLLLSSSSScHO........",  # 08 Forehead & short sideburns
    "........OHoSOOOSSOOOScoO........",  # 09 Thick expressive dark eyebrows
    ".......OSSOSSSSSSSSSScOSSO......",  # 10 Ears top + upper eyelids
    ".......OScOSWESSSSWEScOScO......",  # 11 Ears + 2x3 eyes top
    ".......OScOSEeSSSSeeScOScO......",  # 12 Ears + 2x3 eyes mid
    "........OOOSeeSSSSeeScOOO.......",  # 13 2x3 eyes bottom
    ".........OSSSSSSSSSSSScO........",  # 14 Clean mid-cheek gap
    ".........OccSSmmmmoScccO........",  # 15 Asymmetric confident smirk (tapered up-right)
    "..........OcccccccccccO.........",  # 16 Jawline shadow
    "...........OOOOOOOOOOO..........",  # 17 Chin outline
    "..........OJBBoWWWoJjO..........",  # 18 Street jacket collar + white tee
    ".........OJJJBoWWWoJjjO.........",  # 19 Shoulders (slate-blue jacket)
    "........OJJJJBoWWWoJjjjO........",  # 20 Upper torso
    "........OJJJJBoWWWoJjjjO........",  # 21 Mid torso
    "........OJJoJBoWWWoJoJjO........",  # 22 Sleeve separation + open jacket
    "........OSSOJJwwwwjjOSSO........",  # 23 Clean rounded hands + jacket hem
    ".........OcOOOOOOOOOOcO.........",  # 24 Rounded hand bottoms + belt line
    "..........OPPPpOOOPppO..........",  # 25 Dark jeans upper legs
    "..........OPPPpOOOPppO..........",  # 26 Dark jeans mid legs
    "..........OPPPpOOOPppO..........",  # 27 Dark jeans lower legs
    "..........ORWRO..ORWRO..........",  # 28 Red sneakers upper + lace
    "........ssOWwWOssOWwWOss........",  # 29 Crisp white sneaker soles
    ".........ssOOOssssOOOss.........",  # 30 Rounded sneaker bottom outline
    "..........ssssssssssss..........",  # 31 Ground shadow
]

ALFREDO_IDLE_1: List[str] = [
    "................................",  # 00
    "................................",  # 01
    "................................",  # 02 Breathing idle: head/torso 1px down
    "..........OOOOOOOOOOOO..........",  # 03
    ".........OAAHHHHHHHHhhO.........",  # 04
    "........OAAHHHHHHHHHHhhO........",  # 05
    "........OAHHHHHHHHHHHhhO........",  # 06
    "........OHHHHHHHHHHHHHhO........",  # 07
    "........OHHhSSSSSSSShHhO........",  # 08
    "........OHhSLLLLSSSSScHO........",  # 09
    "........OHoSOOOSSOOOScoO........",  # 10
    ".......OSSOSSSSSSSSSScOSSO......",  # 11
    ".......OScOSWESSSSWEScOScO......",  # 12
    ".......OScOSEeSSSSeeScOScO......",  # 13
    "........OOOSeeSSSSeeScOOO.......",  # 14
    ".........OSSSSSSSSSSSScO........",  # 15
    ".........OccSSmmmmoScccO........",  # 16
    "..........OcccccccccccO.........",  # 17
    "...........OOOOOOOOOOO..........",  # 18
    "..........OJBBoWWWoJjO..........",  # 19 Collar compressed 1px down
    ".........OJJJBoWWWoJjjO.........",  # 20
    "........OJJJJBoWWWoJjjjO........",  # 21
    "........OJJoJBoWWWoJoJjO........",  # 22
    "........OSSOJJwwwwjjOSSO........",  # 23 Hands breathe down 1px
    ".........OcOOOOOOOOOOcO.........",  # 24
    "..........OPPPpOOOPppO..........",  # 25 Jeans upper
    "..........OPPPpOOOPppO..........",  # 26 Jeans mid
    "..........OPPPpOOOPppO..........",  # 27 Jeans lower (legs solid)
    "..........ORWRO..ORWRO..........",  # 28 Sneakers
    "........ssOWwWOssOWwWOss........",  # 29
    ".........ssOOOssssOOOss.........",  # 30
    "..........ssssssssssss..........",  # 31
]

ALFREDO_RUN_0: List[str] = [
    "................................",  # 00
    "................................",  # 01
    "..........OOOOOOOOOOOO..........",  # 02 Left leg forward, right arm forward
    ".........OAAHHHHHHHHhhO.........",  # 03
    "........OAAHHHHHHHHHHhhO........",  # 04
    "........OAHHHHHHHHHHHhhO........",  # 05
    "........OHHHHHHHHHHHHHhO........",  # 06
    "........OHHhSSSSSSSShHhO........",  # 07
    "........OHhSLLLLSSSSScHO........",  # 08
    "........OHoSOOOSSOOOScoO........",  # 09
    ".......OSSOSSSSSSSSSScOSSO......",  # 10
    ".......OScOSWESSSSWEScOScO......",  # 11
    ".......OScOSEeSSSSeeScOScO......",  # 12
    "........OOOSeeSSSSeeScOOO.......",  # 13
    ".........OSSSSSSSSSSSScO........",  # 14
    ".........OccSSmmmmoScccO........",  # 15
    "..........OcccccccccccO.........",  # 16
    "...........OOOOOOOOOOO..........",  # 17
    "..........OJBBoWWWoJjO..........",  # 18
    ".........OJJJBoWWWoJjjO.........",  # 19
    ".......OOJJJJBoWWWoJjjOO........",  # 20 Left arm back, right arm swings forward
    "......OSSOOJJBoWWWoJJjOSSO......",  # 21 Right hand forward (22..25)
    ".......OO.OJJoWWWWoJjjOccO......",  # 22
    "..........OJJwwwwjjOOOOOO.......",  # 23
    "..........OOOOOOOOOO............",  # 24
    ".........OPPPpOOOPppO...........",  # 25 Left leg forward/left
    ".........OPPPpO.OOPppO..........",  # 26
    ".........ORWRO...OPppO..........",  # 27 Left sneaker lifted up/left
    "........sOWwWOs..ORWRO..........",  # 28 Left sole + right sneaker lower
    "........ssOOOss.ssOWwWOss.......",  # 29 Right sneaker grounded
    "...................OOOss........",  # 30
    ".........ssssssssssssss.........",  # 31
]

ALFREDO_RUN_1: List[str] = [
    "................................",  # 00 Passing contact pose (1px bounce UP)
    "..........OOOOOOOOOOOO..........",  # 01
    ".........OAAHHHHHHHHhhO.........",  # 02
    "........OAAHHHHHHHHHHhhO........",  # 03
    "........OAHHHHHHHHHHHhhO........",  # 04
    "........OHHHHHHHHHHHHHhO........",  # 05
    "........OHHhSSSSSSSShHhO........",  # 06
    "........OHhSLLLLSSSSScHO........",  # 07
    "........OHoSOOOSSOOOScoO........",  # 08
    ".......OSSOSSSSSSSSSScOSSO......",  # 09
    ".......OScOSWESSSSWEScOScO......",  # 10
    ".......OScOSEeSSSSeeScOScO......",  # 11
    "........OOOSeeSSSSeeScOOO.......",  # 12
    ".........OSSSSSSSSSSSScO........",  # 13
    ".........OccSSmmmmoScccO........",  # 14
    "..........OcccccccccccO.........",  # 15
    "...........OOOOOOOOOOO..........",  # 16
    "..........OJBBoWWWoJjO..........",  # 17
    ".........OJJJBoWWWoJjjO.........",  # 18
    "........OJJJJBoWWWoJjjjO........",  # 19
    "........OJJJJBoWWWoJjjjO........",  # 20 Arms tucked in mid-stride
    "........OJJoJBoWWWoJoJjO........",  # 21
    ".........OSSOJJwwjjOSSO.........",  # 22
    "..........OcOOOOOOOOcO..........",  # 23
    "...........OPPPpOPppO...........",  # 24 Passing legs close together
    "...........OPPPpOPppO...........",  # 25
    "...........OPPPpOPppO...........",  # 26
    "...........OPPPpOPppO...........",  # 27
    "...........ORWROORWRO...........",  # 28 Contact stride sneakers
    "........sssOWwWOOWwWOsss........",  # 29
    ".........sssOOOssOOOsss.........",  # 30
    "...........ssssssssss...........",  # 31
]

ALFREDO_RUN_2: List[str] = [
    "................................",  # 00 Right leg forward, left arm forward
    "................................",  # 01
    "..........OOOOOOOOOOOO..........",  # 02
    ".........OAAHHHHHHHHhhO.........",  # 03
    "........OAAHHHHHHHHHHhhO........",  # 04
    "........OAHHHHHHHHHHHhhO........",  # 05
    "........OHHHHHHHHHHHHHhO........",  # 06
    "........OHHhSSSSSSSShHhO........",  # 07
    "........OHhSLLLLSSSSScHO........",  # 08
    "........OHoSOOOSSOOOScoO........",  # 09
    ".......OSSOSSSSSSSSSScOSSO......",  # 10
    ".......OScOSWESSSSWEScOScO......",  # 11
    ".......OScOSEeSSSSeeScOScO......",  # 12
    "........OOOSeeSSSSeeScOOO.......",  # 13
    ".........OSSSSSSSSSSSScO........",  # 14
    ".........OccSSmmmmoScccO........",  # 15
    "..........OcccccccccccO.........",  # 16
    "...........OOOOOOOOOOO..........",  # 17
    "..........OJBBoWWWoJjO..........",  # 18
    ".........OJJJBoWWWoJjjO.........",  # 19
    "........OOJJJBoWWWoJjjjOO.......",  # 20 Left arm swings forward, right arm back
    "......OSSOJJJBoWWWoJJOOSSO......",  # 21 Left hand forward (06..09)
    "......OccOJJJoWWWWoJjO.OO.......",  # 22
    ".......OOOOOOjjwwwwJJO..........",  # 23
    "............OOOOOOOOOO..........",  # 24
    "...........OPPPpOOOPppO.........",  # 25 Right leg forward/right
    "..........OPPPpOO.OPppO.........",  # 26
    "..........OPPPpO...ORWRO........",  # 27 Right sneaker lifted up/right
    "..........ORWRO..ssOWwWOs.......",  # 28 Left sneaker lower + right sole
    ".......ssOWwWOss..ssOOOss.......",  # 29 Left sneaker grounded
    "........ssOOOss.................",  # 30
    ".........ssssssssssssss.........",  # 31
]

ALFREDO_RUN_3: List[str] = [
    "................................",  # 00 Passing contact pose 2 (1px bounce UP)
    "..........OOOOOOOOOOOO..........",  # 01
    ".........OAAHHHHHHHHhhO.........",  # 02
    "........OAAHHHHHHHHHHhhO........",  # 03
    "........OAHHHHHHHHHHHhhO........",  # 04
    "........OHHHHHHHHHHHHHhO........",  # 05
    "........OHHhSSSSSSSShHhO........",  # 06
    "........OHhSLLLLSSSSScHO........",  # 07
    "........OHoSOOOSSOOOScoO........",  # 08
    ".......OSSOSSSSSSSSSScOSSO......",  # 09
    ".......OScOSWESSSSWEScOScO......",  # 10
    ".......OScOSEeSSSSeeScOScO......",  # 11
    "........OOOSeeSSSSeeScOOO.......",  # 12
    ".........OSSSSSSSSSSSScO........",  # 13
    ".........OccSSmmmmoScccO........",  # 14
    "..........OcccccccccccO.........",  # 15
    "...........OOOOOOOOOOO..........",  # 16
    "..........OJBBoWWWoJjO..........",  # 17
    ".........OJJJBoWWWoJjjO.........",  # 18
    "........OJJJJBoWWWoJjjjO........",  # 19
    ".......OOJJJJBoWWWoJjjjOO.......",  # 20 Slight arm/shoulder flare on 2nd pass
    ".......OSSOJJBoWWWoJjOSSO.......",  # 21
    "........OccOJwwwwwwJOOccO.......",  # 22
    ".........OOOOOOOOOOOOOO.........",  # 23
    "..........OPPPpOOOPppO..........",  # 24 Passing knee overlap
    "..........OPPPpOOOPppO..........",  # 25
    "..........OPPPpOOOPppO..........",  # 26
    "..........ORWRO..OPppO..........",  # 27 Left foot lifting slightly
    ".........sOWwWOs.ORWRO..........",  # 28
    ".........ssOOOss.OWwWOss........",  # 29 Right foot grounded
    "..................OOOss.........",  # 30
    "...........ssssssssss...........",  # 31
]

ALFREDO_DASH: List[str] = [
    "................................",  # 00 Dynamic forward-leaning Chronos Dash!
    "................................",  # 01
    "..............OOOOOOOOOOOO......",  # 02 Head leaning forward-right
    "....33X2.....OAAHHHHHHHHhhO.....",  # 03 Neon Cyan Chronos speed streak top
    "..3XXXX22...OAAHHHHHHHHHHhhO....",  # 04
    "............OAHHHHHHHHHHHhhO....",  # 05
    "....3X2.....OHHHHHHHHHHHHHhO....",  # 06
    "............OHHhSSSSSSSShHhO....",  # 07
    "..3XXXX22...OHhSLLLLSSSSScHO....",  # 08 Chronos speed trail mid
    "......22....OHoSOOOSSOOOScoO....",  # 09
    "...........OSSOSSSSSSSSSScOSSO..",  # 10
    ".3XXXXX22..OScOSWESSSSWEScOScO..",  # 11 Bright cyan afterimage trail
    "...22......OScOSEeSSSSeeScOScO..",  # 12
    "............OOOSeeSSSSeeScOOO...",  # 13
    ".............OSSSSSSSSSSSScO....",  # 14
    "....3XX2.....OccSSmmmmoScccO....",  # 15
    "..............OcccccccccccO.....",  # 16
    "...............OOOOOOOOOOO......",  # 17
    "..3XXXXX22....OJBBoWWWoJjO......",  # 18 Jacket trail fluttering back
    "....22.......OJJJBoWWWoJjjO.....",  # 19
    "........32..OJJJJBoWWWoJjjjO....",  # 20 Left arm streaming back, right forward
    "..3XXX22...OSSOJJBoWWWoJjjOSSO..",  # 21 Hands streamlined
    "............OccOOOOOOOOOOOOccO..",  # 22
    ".....3XX2....OPPPpOOOOOPppO.....",  # 23 Wide dash stride
    "............OPPPpO.....OPppO....",  # 24
    "..3XXXX2...OPPPpO.......OPppO...",  # 25
    "..........OPPPpO.........OPppO..",  # 26 Extended stride legs
    ".........ORWRO...........ORWRO..",  # 27
    "........sOWwWOs.........sOWwWOs.",  # 28
    "........ssOOOss.........ssOOOss.",  # 29 Grounded dash sneakers
    "................................",  # 30
    "........sssssssssssssssssssss...",  # 31 Wide motion shadow
]

ALFREDO_THROW: List[str] = [
    "................................",  # 00 Action pose throwing Bus Stop Beacon (X/F)
    "................................",  # 01
    "..........OOOOOOOOOOOO..........",  # 02
    ".........OAAHHHHHHHHhhO.........",  # 03
    "........OAAHHHHHHHHHHhhO........",  # 04
    "........OAHHHHHHHHHHHhhO........",  # 05
    "........OHHHHHHHHHHHHHhO........",  # 06
    "........OHHhSSSSSSSShHhO........",  # 07
    "........OHhSLLLLSSSSScHO........",  # 08
    "........OHoSOOOSSOOOScoO........",  # 09
    ".......OSSOSSSSSSSSSScOSSO......",  # 10
    ".......OScOSWESSSSWEScOScO..3X..",  # 11 Floating Bus Stop beacon sparkle
    ".......OScOSEeSSSSeeScOScO.OXXO.",  # 12 Bus Stop beacon body (Neon Cyan + Coral)
    "........OOOSeeSSSSeeScOOO.OXFFXO",  # 13 Beacon core (X/F)
    ".........OSSSSSSSSSSSScO..OXFFXO",  # 14 Beacon core bottom
    ".........OccSSmmmmoScccO...O22O.",  # 15 Beacon base
    "..........OcccccccccccO...OSSO..",  # 16 Right hand raised throwing!
    "...........OOOOOOOOOOO...OJJcO..",  # 17 Raised right sleeve
    "..........OJBBoWWWoJjO..OJJjO...",  # 18 Jacket collar + raised shoulder
    "........OOJJJBoWWWoJjjOOJJjO....",  # 19
    "......OSSOOJJBoWWWoJjjJJjOO.....",  # 20 Left hand out for balance
    "......OccOJJJBoWWWoJjjOOO.......",  # 21
    ".......OOOJJoWWWWoJjjO..........",  # 22
    "..........OOOOOOOOOOO...........",  # 23
    ".........OPPPpOOOPppO...........",  # 24 Wide braced throwing stance
    ".........OPPPpO.OOPppO..........",  # 25
    "........OPPPpO...OOPppO.........",  # 26
    "........OPPPpO...OOPppO.........",  # 27
    "........ORWRO.....ORWRO.........",  # 28
    "......ssOWwWOss.ssOWwWOss.......",  # 29 Grounded throwing stance sneakers
    ".......ssOOOss...ssOOOss........",  # 30
    ".......ssssssssssssssssss.......",  # 31
]

ALFREDO_FRAMES = [
    ("alfredo_idle_0", ALFREDO_IDLE_0),
    ("alfredo_idle_1", ALFREDO_IDLE_1),
    ("alfredo_run_0", ALFREDO_RUN_0),
    ("alfredo_run_1", ALFREDO_RUN_1),
    ("alfredo_run_2", ALFREDO_RUN_2),
    ("alfredo_run_3", ALFREDO_RUN_3),
    ("alfredo_dash", ALFREDO_DASH),
    ("alfredo_throw", ALFREDO_THROW),
]


# ==============================================================================
# 2. GIRL ARCHETYPE 1 — KAWAII (32x32, 3 frames: idle_0, walk_1, walk_2)
# ==============================================================================
GIRL_KAWAII_IDLE_0: List[str] = [
    "................................",  # 00
    "..........OOOOOOOOOOOO..........",  # 01
    "........OOQQKKKKKKKKkkOO........",  # 02
    "......OORWOKKKKKKKKKKkkWROO.....",  # 03
    ".....OORRWOKQQqqqqQQKkkWRROO....",  # 04
    "....OOKKROKqqqKKKKqqqKkkORKOO...",  # 05
    "....OQQKOOxKKKKKKKKKKkkxOOKkO...",  # 06
    "...OQQKOkKKKKKKKKKKKKKkkkOKkkO..",  # 07
    "...OQKOkKKxKSSKKKKSSKxKKkKOKkO..",  # 08
    "...OQKOKKkxSSSSkkSSSSxkKkKOKkO..",  # 09
    "...OQKOkkSSOOOSSSSOOOSSkkKOKkO..",  # 10
    "...OKkOkkSSWIOSSSSOIWSSkkKOKkO..",  # 11
    "..OOKkOkkSSIiOSSSSOiISSkkKOKkOO.",  # 12
    "..OKkkOkkSSbbSSSSSSbbSSkkKOKkkO.",  # 13
    "..OKkkOkkSSSSSSmmSSSSSSkkKOKkkO.",  # 14
    "..OKkkOOkkSccccSSccccSkkOOKkkkO.",  # 15
    "...OKkkOkkxOOOOOOOOOOxkkOOKkkO..",  # 16
    "...OOKkOOkkOWWWrrWWWOkkOOOKkOO..",  # 17
    "....OOKkOOOVVWWrrWWvvOOOoKkOO...",  # 18
    ".....OOKkOVVVVoWWovvvvOOKkOO....",  # 19
    "......OOOOQKVVoGGovvKkOOOOO.....",  # 20
    "........OSSOOVVVvvvvOSSO........",  # 21
    "........OccOOKkKkkKOOccO........",  # 22
    ".........OOKKkkKKkkKKOO.........",  # 23
    "........OOKKkkKKkkKKkkOO........",  # 24
    "........OWWWwWWWWwWWWWwO........",  # 25
    "........OOOOOOOOOOOOOOOO........",  # 26
    "..........OSccO..OSccO..........",  # 27
    "..........OWWwO..OWWwO..........",  # 28
    ".......sssOVvvOssOVvvOsss.......",  # 29
    "........ssOOOOOssOOOOOss........",  # 30
    "..........ssssssssssss..........",  # 31
]

GIRL_KAWAII_WALK_1: List[str] = [
    "................................",  # 00 Left foot step + twin-tail bounce
    "..........OOOOOOOOOOOO..........",  # 01
    "........OOQQKKKKKKKKkkOO........",  # 02
    "......OORWOKKKKKKKKKKkkWROO.....",  # 03
    "....OOORRWOKQQqqqqQQKkkWRROOO...",  # 04 Twin-tails flare/bounce 1px outward
    "...OOKKKROKqqqKKKKqqqKkkORKKkO..",  # 05
    "...OQQKKOOxKKKKKKKKKKkkxOOKkkO..",  # 06
    "..OQQKKOkKKKKKKKKKKKKKkkkOKkkOO.",  # 07
    "..OQKKOkKKxKSSKKKKSSKxKKkKOKkkO.",  # 08
    "...OQKOKKkxSSSSkkSSSSxkKkKOKkO..",  # 09
    "...OQKOkkSSOOOSSSSOOOSSkkKOKkO..",  # 10
    "...OKkOkkSSWIOSSSSOIWSSkkKOKkO..",  # 11
    "..OOKkOkkSSIiOSSSSOiISSkkKOKkOO.",  # 12
    "..OKkkOkkSSbbSSSSSSbbSSkkKOKkkO.",  # 13
    "..OKkkOkkSSSSSSmmSSSSSSkkKOKkkO.",  # 14
    "..OKkkOOkkSccccSSccccSkkOOKkkkO.",  # 15
    "...OKkkOkkxOOOOOOOOOOxkkOOKkkO..",  # 16
    "...OOKkOOkkOWWWrrWWWOkkOOOKkOO..",  # 17
    "....OOKkOOOVVWWrrWWvvOOOoKkOO...",  # 18
    ".....OOKkOVVVVoWWovvvvOOKkOO....",  # 19
    ".....OOOOOQKVVoGGovvKkOOOOO.....",  # 20 Left arm back, right arm forward
    ".......OSSOOOVVVvvvvOOSSO.......",  # 21
    ".......OccOOKkKkkKOO.OccO.......",  # 22 Skirt sways slightly left
    "........OOKKkkKKkkKKOO..........",  # 23
    ".......OOKKkkKKkkKKkkOO.........",  # 24
    ".......OWWWwWWWWwWWWWwO.........",  # 25
    ".......OOOOOOOOOOOOOOOO.........",  # 26
    ".........OWWwO....OSccO.........",  # 27 Left foot lifted/stepped forward-left
    "........sOVvvOs...OWWwO.........",  # 28 Left Mary-Jane high + right sock
    "........ssOOOOss.sOVvvOss.......",  # 29 Right shoe planted
    "..................OOOOOss.......",  # 30
    ".........ssssssssssssss.........",  # 31
]

GIRL_KAWAII_WALK_2: List[str] = [
    "................................",  # 00 Right foot step + twin-tail bounce
    "..........OOOOOOOOOOOO..........",  # 01
    "........OOQQKKKKKKKKkkOO........",  # 02
    "......OORWOKKKKKKKKKKkkWROO.....",  # 03
    "....OOORRWOKQQqqqqQQKkkWRROOO...",  # 04 Twin-tails flare/bounce
    "...OOKKKROKqqqKKKKqqqKkkORKKkO..",  # 05
    "...OQQKKOOxKKKKKKKKKKkkxOOKkkO..",  # 06
    "..OQQKKOkKKKKKKKKKKKKKkkkOKkkOO.",  # 07
    "..OQKKOkKKxKSSKKKKSSKxKKkKOKkkO.",  # 08
    "...OQKOKKkxSSSSkkSSSSxkKkKOKkO..",  # 09
    "...OQKOkkSSOOOSSSSOOOSSkkKOKkO..",  # 10
    "...OKkOkkSSWIOSSSSOIWSSkkKOKkO..",  # 11
    "..OOKkOkkSSIiOSSSSOiISSkkKOKkOO.",  # 12
    "..OKkkOkkSSbbSSSSSSbbSSkkKOKkkO.",  # 13
    "..OKkkOkkSSSSSSmmSSSSSSkkKOKkkO.",  # 14
    "..OKkkOOkkSccccSSccccSkkOOKkkkO.",  # 15
    "...OKkkOkkxOOOOOOOOOOxkkOOKkkO..",  # 16
    "...OOKkOOkkOWWWrrWWWOkkOOOKkOO..",  # 17
    "....OOKkOOOVVWWrrWWvvOOOoKkOO...",  # 18
    ".....OOKkOVVVVoWWovvvvOOKkOO....",  # 19
    ".....OOOOOQKVVoGGovvKkOOOOO.....",  # 20 Right arm back, left arm forward
    ".......OSSOOOVVVvvvvOOSSO.......",  # 21
    ".......OccO.OOKkKkkKOOccO.......",  # 22 Skirt sways slightly right
    "..........OOKKkkKKkkKKOO........",  # 23
    ".........OOKKkkKKkkKKkkOO.......",  # 24
    ".........OWWWwWWWWwWWWWwO.......",  # 25
    ".........OOOOOOOOOOOOOOOO.......",  # 26
    "..........OSccO....OWWwO........",  # 27 Right foot lifted/stepped forward-right
    "..........OWWwO...sOVvvOs.......",  # 28 Right shoe high + left sock
    "........ssOVvvOs..ssOOOOss......",  # 29 Left shoe planted (aligned cols 10..14)
    "........ssOOOOO.................",  # 30
    ".........ssssssssssssss.........",  # 31
]


# ==============================================================================
# 3. GIRL ARCHETYPE 2 — EMO (32x32, 3 frames: idle_0, walk_1, walk_2)
# Grunge Purple & Black hair, asymmetric bangs covering her left eye (viewer's
# right) so only her right glowing magenta/purple eye (I,i,W) is visible!
# ==============================================================================
GIRL_EMO_IDLE_0: List[str] = [
    "................................",  # 00
    "..........OOOOOOOOOOOO..........",  # 01
    "........OOZZUUUUUUUUuuOO........",  # 02 Grunge purple hair crown
    ".......OZZUUUUUUUUUUUUuuO.......",  # 03
    "......OZZUUZZZZwwZZZZUUuuO......",  # 04 Specular bangs highlight band (peak-npcref4!)
    ".....OZZUUZZZUUUUUUZZZUUuuO.....",  # 05 Purple highlight streaks
    ".....OZUuUUUUUUUUUUUUUUUuuO.....",  # 06
    "....OOUuUUUUUUUUuUUUUUUUUuuO....",  # 07 Asymmetric layered bangs sweep right
    "....OUuUUuSSSUUUuUUUUUUUUuuO....",  # 08 Bangs dip over her left eye
    "....OUuUuSSSSSUUuuUUUUUUUuuO....",  # 09 Layered purple bangs over right eye
    "....OUuUuSOOOSSUUuUUUUUUUuuO....",  # 10 Visible right upper lash (cols 10..12)
    "....OUuUuSWIOSSSUUuUUUUUUuuO....",  # 11 Glowing magenta/purple eye top (W,I,O)
    "....OUuUuSIiOSSSUUuuUUUUUuuO....",  # 12 Glowing eye bottom + layered bangs
    "....OUuUuSbbSSSSSUUuuUUUUuuO....",  # 13 Cheek blush + clean bangs sweep
    "....OUuUuSSSSSSmmSSUuUUUUuuO....",  # 14 Cute muted emo expression
    "....OUuUuSSccccSScccuUUUUuuO....",  # 15 Soft chin + long side locks
    ".....OUuUuOOOOOOOOOOuUUUUuO.....",  # 16 Chin outline + framing locks
    ".....OUuUOHHHwHHHHHHHOUUuuO.....",  # 17 Black choker with silver stud
    "......OUuOVVVVVvvvvvvOOUuO......",  # 18 Dark gothic hoodie collar
    "......OOuOUHUHVVvvHUHUOOuO......",  # 19 Purple/black striped sleeves!
    ".......OOOHUHUvvvvUHUHOOOO......",  # 20 Striped sleeves mid + gothic bodice
    "........OSSOOUUuuUUOSSO.........",  # 21 Hands + gothic waist
    "........OccOOVVVvvvOOccO........",  # 22
    ".........OOUUwwUUwwUUOO.........",  # 23 Dark pleated gothic skirt + silver chain!
    "........OOUUuuUUwUUuuUOO........",  # 24 Pleated skirt flare + hanging chain
    "........OUUuuVVVVVVuuUUO........",  # 25 Gothic hem
    "........OOOOOOOOOOOOOOOO........",  # 26 Skirt outline
    "..........OSccO..OSccO..........",  # 27 Legs
    "..........OHhhO..OHhhO..........",  # 28 Platform boots upper
    ".......sssOHHHOssOHHHOsss.......",  # 29 Chunky dark platform boots
    "........ssOOOOOssOOOOOss........",  # 30 Boots bottom outline
    "..........ssssssssssss..........",  # 31 Ground shadow
]

GIRL_EMO_WALK_1: List[str] = [
    "................................",  # 00 Left foot step + hair sway
    "..........OOOOOOOOOOOO..........",  # 01
    "........OOZZUUUUUUUUuuOO........",  # 02
    ".......OZZUUUUUUUUUUUUuuO.......",  # 03
    "......OZZUUZZZZwwZZZZUUuuO......",  # 04
    ".....OZZUUZZZUUUUUUZZZUUuuO.....",  # 05
    "....OOZUuUUUUUUUUUUUUUUUuuOO....",  # 06 Hair flares slightly on bounce
    "...OOUuUUUUUUUUUuUUUUUUUUuuuO...",  # 07
    "...OUuUUUuSSSUUUuUUUUUUUUuuuO...",  # 08
    "....OUuUuSSSSSUUuuUUUUUUUuuO....",  # 09
    "....OUuUuSOOOSSUUuUUUUUUUuuO....",  # 10
    "....OUuUuSWIOSSSUUuUUUUUUuuO....",  # 11
    "....OUuUuSIiOSSSUUuuUUUUUuuO....",  # 12
    "....OUuUuSbbSSSSSUUuuUUUUuuO....",  # 13
    "....OUuUuSSSSSSmmSSUuUUUUuuO....",  # 14
    "....OUuUuSSccccSScccuUUUUuuO....",  # 15
    ".....OUuUuOOOOOOOOOOuUUUUuO.....",  # 16
    ".....OUuUOHHHwHHHHHHHOUUuuO.....",  # 17
    "......OUuOVVVVVvvvvvvOOUuO......",  # 18
    "......OOuOUHUHVVvvHUHUOOuO......",  # 19
    ".......OOOHUHUvvvvUHUHOOOO......",  # 20
    ".......OSSOOOUUuuUUOOSSO........",  # 21 Arms swing
    ".......OccOOVVVvvvOO.OccO.......",  # 22
    "........OOUUwwUUwwUUOO..........",  # 23 Skirt sways slightly left
    ".......OOUUuuUUwUUuuUOO.........",  # 24
    ".......OUUuuVVVVVVuuUUO.........",  # 25
    ".......OOOOOOOOOOOOOOOO.........",  # 26
    ".........OHhhO....OSccO.........",  # 27 Left platform boot lifted
    "........sOHHHOs...OHhhO.........",  # 28
    "........ssOOOOss.sOHHHOss.......",  # 29 Right boot planted
    "..................OOOOOss.......",  # 30
    ".........ssssssssssssss.........",  # 31
]

GIRL_EMO_WALK_2: List[str] = [
    "................................",  # 00 Right foot step + hair sway
    "..........OOOOOOOOOOOO..........",  # 01
    "........OOZZUUUUUUUUuuOO........",  # 02
    ".......OZZUUUUUUUUUUUUuuO.......",  # 03
    "......OZZUUZZZZwwZZZZUUuuO......",  # 04
    ".....OZZUUZZZUUUUUUZZZUUuuO.....",  # 05
    "....OOZUuUUUUUUUUUUUUUUUuuOO....",  # 06 Hair flares slightly on bounce
    "...OOUuUUUUUUUUUuUUUUUUUUuuuO...",  # 07
    "...OUuUUUuSSSUUUuUUUUUUUUuuuO...",  # 08
    "....OUuUuSSSSSUUuuUUUUUUUuuO....",  # 09
    "....OUuUuSOOOSSUUuUUUUUUUuuO....",  # 10
    "....OUuUuSWIOSSSUUuUUUUUUuuO....",  # 11
    "....OUuUuSIiOSSSUUuuUUUUUuuO....",  # 12
    "....OUuUuSbbSSSSSUUuuUUUUuuO....",  # 13
    "....OUuUuSSSSSSmmSSUuUUUUuuO....",  # 14
    "....OUuUuSSccccSScccuUUUUuuO....",  # 15
    ".....OUuUuOOOOOOOOOOuUUUUuO.....",  # 16
    ".....OUuUOHHHwHHHHHHHOUUuuO.....",  # 17
    "......OUuOVVVVVvvvvvvOOUuO......",  # 18
    "......OOuOUHUHVVvvHUHUOOuO......",  # 19
    ".......OOOHUHUvvvvUHUHOOOO......",  # 20
    ".......OSSOOOUUuuUUOOSSO........",  # 21 Arms swing
    ".......OccO.OOVVVvvvOOccO.......",  # 22
    "..........OOUUwwUUwwUUOO........",  # 23 Skirt sways slightly right
    ".........OOUUuuUUwUUuuUOO.......",  # 24
    ".........OUUuuVVVVVVuuUUO.......",  # 25
    ".........OOOOOOOOOOOOOOOO.......",  # 26
    "..........OSccO....OHhhO........",  # 27 Right platform boot lifted
    "..........OHhhO...sOHHHOs.......",  # 28
    "........ssOHHHOs..ssOOOOss......",  # 29 Left boot planted (aligned cols 10..14)
    "........ssOOOOO.................",  # 30
    ".........ssssssssssssss.........",  # 31
]


# ==============================================================================
# 4. GIRL ARCHETYPE 3 — STUDENT (32x32, 3 frames: idle_0, walk_1, walk_2)
# Caramel/Chestnut hair (T,t,1) + shine (C), white/cyan headphones (W,X)
# around neck, sapphire anime eyes (J,B,W) 100% aligned with top-left highlight!
# ==============================================================================
GIRL_STUDENT_IDLE_0: List[str] = [
    "................................",  # 00
    "..........OOOOOOOOOOOO..........",  # 01
    "........OO11TTTTTTTTttOO........",  # 02 Caramel hair crown
    ".......O11TTTTTTTTTTTTttO.......",  # 03
    "......O11TTCCC1111CCCTTttO......",  # 04 Specular shine (C) on bangs
    ".....O11TT11TTTTTTTT11TTttO.....",  # 05
    ".....O1TTTtTTTTTTTTTTTTtTTtO....",  # 06 Cute side locks framing face
    "....O1TTtTTTTTSSSSSTTTTTttTtO...",  # 07 Soft parted bangs
    "....OTTtTTTtSSSTTTTSSStTTttTtO..",  # 08
    "....OTtOTttSSSSSTTTSSSSSttOTtO..",  # 09
    "....OTtOttSSOOOSSSSOOOSStoOTtO..",  # 10 Upper eyelashes (cols 12..14 & 19..21)
    "....OTtOttSSWJOSSSSWJOSStoOTtO..",  # 11 Sapphire eyes top (top-left W shine on both!)
    "....OTtOttSSJBOSSSSJBOSStoOTtO..",  # 12 Sapphire eyes bottom (J,B)
    "....OTtOttSSbbSSSSSSbbSStoOTtO..",  # 13 Rosy blush
    "....OTtOttSSSSSSmmSSSSSStoOTtO..",  # 14 Sweet smile
    "....OTtOOttSccccSSccccSttOOTtO..",  # 15 Soft chibi chin
    ".....OTtOtttOOOOOOOOOOtttOOTtO..",  # 16 Chin outline + side locks
    ".....OOTtOWXOWWWrrWWWOWXOOTtOO..",  # 17 White/cyan headphones + white shirt + red tie!
    "......OOTOWWJJWWrrWWjjWWOOTOO...",  # 18 Headphones cups + Campus Blue blazer
    ".......OOJJJJJoWrrWoJjjjjOO.....",  # 19 Blazer + crisp shirt + tie
    ".......OOJJJJJoWrrWoJjjjjOO.....",  # 20 Blazer sleeves & torso
    "........OSSOOJJJjjjjOOSSO.......",  # 21 Hands + cardigan waist
    "........OccOONnNwNnNOOccO.......",  # 22 Pleated check skirt top (N,n,w)
    ".........OONwNNnNwNNnNOO........",  # 23 Checkered pleat pattern
    "........OONnNwNNnNwNNnNOO.......",  # 24 Pleated check skirt flare
    "........ONwNNnNwNNnNwNNnO.......",  # 25 Check hem
    "........OOOOOOOOOOOOOOOO........",  # 26 Skirt bottom outline
    "..........OSccO..OSccO..........",  # 27 Legs
    "..........OWWwO..OWWwO..........",  # 28 White socks
    ".......sssOTttOssOTttOsss.......",  # 29 Brown loafers
    "........ssOOOOOssOOOOOss........",  # 30 Loafers bottom outline
    "..........ssssssssssss..........",  # 31 Ground shadow
]

GIRL_STUDENT_WALK_1: List[str] = [
    "................................",  # 00 Left foot step + side-lock bounce
    "..........OOOOOOOOOOOO..........",  # 01
    "........OO11TTTTTTTTttOO........",  # 02
    ".......O11TTTTTTTTTTTTttO.......",  # 03
    "......O11TTCCC1111CCCTTttO......",  # 04
    ".....O11TT11TTTTTTTT11TTttO.....",  # 05
    "....OO1TTTtTTTTTTTTTTTTtTTtOO...",  # 06 Side locks bounce outward
    "...OO1TTtTTTTTSSSSSTTTTTttTtOO..",  # 07
    "...O1TTtTTTtSSSTTTTSSStTTttTtO..",  # 08
    "....OTtOTttSSSSSTTTSSSSSttOTtO..",  # 09
    "....OTtOttSSOOOSSSSOOOSStoOTtO..",  # 10
    "....OTtOttSSWJOSSSSWJOSStoOTtO..",  # 11
    "....OTtOttSSJBOSSSSJBOSStoOTtO..",  # 12
    "....OTtOttSSbbSSSSSSbbSStoOTtO..",  # 13
    "....OTtOttSSSSSSmmSSSSSStoOTtO..",  # 14
    "....OTtOOttSccccSSccccSttOOTtO..",  # 15
    ".....OTtOtttOOOOOOOOOOtttOOTtO..",  # 16
    ".....OOTtOWXOWWWrrWWWOWXOOTtOO..",  # 17
    "......OOTOWWJJWWrrWWjjWWOOTOO...",  # 18
    ".......OOJJJJJoWrrWoJjjjjOO.....",  # 19
    ".......OOJJJJJoWrrWoJjjjjOO.....",  # 20
    ".......OSSOOOJJJjjjjOOOSSO......",  # 21 Arms swing
    ".......OccOONnNwNnNOO.OccO......",  # 22 Skirt sways slightly left
    "........OONwNNnNwNNnNOO.........",  # 23
    ".......OONnNwNNnNwNNnNOO........",  # 24
    ".......ONwNNnNwNNnNwNNnO........",  # 25
    ".......OOOOOOOOOOOOOOOO.........",  # 26
    ".........OWWwO....OSccO.........",  # 27 Left loafer lifted
    "........sOTttOs...OWWwO.........",  # 28
    "........ssOOOOss.sOTttOss.......",  # 29 Right loafer planted
    "..................OOOOOss.......",  # 30
    ".........ssssssssssssss.........",  # 31
]

GIRL_STUDENT_WALK_2: List[str] = [
    "................................",  # 00 Right foot step + side-lock bounce
    "..........OOOOOOOOOOOO..........",  # 01
    "........OO11TTTTTTTTttOO........",  # 02
    ".......O11TTTTTTTTTTTTttO.......",  # 03
    "......O11TTCCC1111CCCTTttO......",  # 04
    ".....O11TT11TTTTTTTT11TTttO.....",  # 05
    "....OO1TTTtTTTTTTTTTTTTtTTtOO...",  # 06 Side locks bounce outward
    "...OO1TTtTTTTTSSSSSTTTTTttTtOO..",  # 07
    "...O1TTtTTTtSSSTTTTSSStTTttTtO..",  # 08
    "....OTtOTttSSSSSTTTSSSSSttOTtO..",  # 09
    "....OTtOttSSOOOSSSSOOOSStoOTtO..",  # 10
    "....OTtOttSSWJOSSSSWJOSStoOTtO..",  # 11
    "....OTtOttSSJBOSSSSJBOSStoOTtO..",  # 12
    "....OTtOttSSbbSSSSSSbbSStoOTtO..",  # 13
    "....OTtOttSSSSSSmmSSSSSStoOTtO..",  # 14
    "....OTtOOttSccccSSccccSttOOTtO..",  # 15
    ".....OTtOtttOOOOOOOOOOtttOOTtO..",  # 16
    ".....OOTtOWXOWWWrrWWWOWXOOTtOO..",  # 17
    "......OOTOWWJJWWrrWWjjWWOOTOO...",  # 18
    ".......OOJJJJJoWrrWoJjjjjOO.....",  # 19
    ".......OOJJJJJoWrrWoJjjjjOO.....",  # 20
    ".......OSSOOOJJJjjjjOOOSSO......",  # 21 Arms swing
    ".......OccO.OONnNwNnNOOccO......",  # 22 Skirt sways slightly right
    ".........OONwNNnNwNNnNOO........",  # 23
    "........OONnNwNNnNwNNnNOO.......",  # 24
    "........ONwNNnNwNNnNwNNnO.......",  # 25
    "........OOOOOOOOOOOOOOOO........",  # 26
    "..........OSccO....OWWwO........",  # 27 Right loafer lifted
    "..........OWWwO...sOTttOs.......",  # 28
    "........ssOTttOs..ssOOOOss......",  # 29 Left loafer planted (aligned cols 10..14)
    "........ssOOOOO.................",  # 30
    ".........ssssssssssssss.........",  # 31
]


# ==============================================================================
# 5. GIRL ARCHETYPE 4 — GYARU (32x32, 3 frames: idle_0, walk_1, walk_2)
# Golden-blonde twin curls (D,G,C), Neon Cyan 5G clips (X,2,3) & pink heart (K),
# cyan/pink anime eyes (X,i,W) 100% aligned, cropped pink/white jacket (K,W)
# ==============================================================================
GIRL_GYARU_IDLE_0: List[str] = [
    "................................",  # 00
    "..........OOOOOOOOOOOO..........",  # 01
    "........OOCGGGGGGGGGGDDO........",  # 02 Voluminous golden-blonde crown
    "......OO3XOGGGGGGGGGGOX3OO......",  # 03 Neon Cyan 5G hairclips/bows!
    ".....OO3XXOGGCCCCCCGGOGXX2OO....",  # 04 Twin curls + bangs shine (C)
    "....OOGG2XOKKGGGGGGKKOGX2DDO....",  # 05 Pink heart clips (K) + curls
    "...OOCGGGDtGGGGGGGGGGDDtOGGDO...",  # 06 Bouncy twin curls upper volume
    "..OOCGGGDTGGGGGGGGGGGGDDDTGGDO..",  # 07
    "..OCGGGtGGDtGDDGGDDGtDGGDtOGDDO.",  # 08 Side curls + bangs
    "..OCGDoGGGtSSSSDDSSSStGGDoOGDDO.",  # 09
    "..OGGDoGGSSOOOSSSSOOOSSGDoOGDDO.",  # 10 Upper eyelashes (cols 11..13 & 18..20)
    "..OGGDoGGSSWXOSSSSWXOSSGDoOGDDO.",  # 11 Sparkling cyan eyes top (W,X,O aligned!)
    "..OGDDoGGSS2iOSSSS2iOSSGDoOGDDO.",  # 12 Sparkling cyan/pink eyes bottom (2,i,O)
    "..OGDDoGGSSbbSSSSSSbbSSGDoOGDDO.",  # 13 Rosy blush
    "..OGDDoGGSSSSSSmmSSSSSSGDoOGDDO.",  # 14 Cute gyaru smile
    "..OGGDtGGGtccccSScccctGGDtOGGDO.",  # 15 Soft chin + bouncy curls
    "...OGDDtGGtOOOOOOOOOOtGDtOGDDO..",  # 16 Chin outline + twin curls
    "...OOGGDOtOOWWWVVWWWOtOOOGGDOO..",  # 17 Cropped white/pink collar + dark top
    "....OOGGOOOKKWWVVWWKKOOOGGDOO...",  # 18 Trendy cropped pink/white jacket
    ".....OOOOOKKKWWVVWWKKKOOOOOO....",  # 19 Cropped jacket + dark inner top
    "......OOSSOKKKWVVWKKKOSSOO......",  # 20 Short cropped sleeves + hands
    ".......OccOOVVVVVvvvOOccO.......",  # 21 Hands + dark top waist
    "........OOCCCCCCCCaaaaOO........",  # 22 Stylish cream/tan mini-skirt!
    "........OOCaCCaCCaaCCaOO........",  # 23 Mini-skirt pleats
    "........OOOOOOOOOOOOOOOO........",  # 24 Higher mini-skirt hem
    "..........OSccO..OSccO..........",  # 25 Exposed legs above leg warmers!
    ".........OWWWwO..OWWWwO.........",  # 26 Iconic Gyaru loose leg-warmer socks top
    "........OWWWWWwOOWWWWWwO........",  # 27 Wide fluffy leg-warmers mid
    "........OWwwwwWOOWwwwwWO........",  # 28 Loose leg-warmers bunched
    ".......ssOTTTTossOTTTToss.......",  # 29 Chunky caramel platform shoes
    "........ssOOOOOssOOOOOss........",  # 30 Platform shoes bottom outline
    "..........ssssssssssss..........",  # 31 Ground shadow
]

GIRL_GYARU_WALK_1: List[str] = [
    "................................",  # 00 Left foot step + twin-curl bounce
    "..........OOOOOOOOOOOO..........",  # 01
    "........OOCGGGGGGGGGGDDO........",  # 02
    "......OO3XOGGGGGGGGGGOX3OO......",  # 03
    "....OOO3XXOGGCCCCCCGGOGXX2OOO...",  # 04 Twin curls bounce outward
    "...OOGGG2XOKKGGGGGGKKOGX2DDDO...",  # 05
    "..OOCGGGGDtGGGGGGGGGGDDtOGGGDO..",  # 06
    ".OOCGGGGDTGGGGGGGGGGGGDDDTGGGDO.",  # 07
    ".OCGGGGtGGDtGDDGGDDGtDGGDtOGDDO.",  # 08
    "..OCGDoGGGtSSSSDDSSSStGGDoOGDDO.",  # 09
    "..OGGDoGGSSOOOSSSSOOOSSGDoOGDDO.",  # 10
    "..OGGDoGGSSWXOSSSSWXOSSGDoOGDDO.",  # 11
    "..OGDDoGGSS2iOSSSS2iOSSGDoOGDDO.",  # 12
    "..OGDDoGGSSbbSSSSSSbbSSGDoOGDDO.",  # 13
    "..OGDDoGGSSSSSSmmSSSSSSGDoOGDDO.",  # 14
    "..OGGDtGGGtccccSScccctGGDtOGGDO.",  # 15
    "...OGDDtGGtOOOOOOOOOOtGDtOGDDO..",  # 16
    "...OOGGDOtOOWWWVVWWWOtOOOGGDOO..",  # 17
    "....OOGGOOOKKWWVVWWKKOOOGGDOO...",  # 18
    ".....OOOOOKKKWWVVWWKKKOOOOOO....",  # 19
    ".....OOSSOKKKKWVVWKKKKOSSOO.....",  # 20 Arms swing
    "......OccOOVVVVVvvvOO.OccO......",  # 21
    ".......OOCCCCCCCCaaaaOO.........",  # 22 Mini-skirt sways left
    ".......OOCaCCaCCaaCCaOO.........",  # 23
    ".......OOOOOOOOOOOOOOOO.........",  # 24
    "........OWWWwO....OSccO.........",  # 25 Left leg-warmer lifted
    ".......OWWWWWwO..OWWWwO.........",  # 26
    ".......OWwwwwWO.OWWWWWwO........",  # 27
    "......ssOTTTTos.OWwwwwWO........",  # 28 Left platform shoe lifted
    ".......ssOOOOss.sOTTTToss.......",  # 29 Right platform shoe planted
    "..................OOOOOss.......",  # 30
    ".........ssssssssssssss.........",  # 31
]

GIRL_GYARU_WALK_2: List[str] = [
    "................................",  # 00 Right foot step + twin-curl bounce
    "..........OOOOOOOOOOOO..........",  # 01
    "........OOCGGGGGGGGGGDDO........",  # 02
    "......OO3XOGGGGGGGGGGOX3OO......",  # 03
    "....OOO3XXOGGCCCCCCGGOGXX2OOO...",  # 04 Twin curls bounce outward
    "...OOGGG2XOKKGGGGGGKKOGX2DDDO...",  # 05
    "..OOCGGGGDtGGGGGGGGGGDDtOGGGDO..",  # 06
    ".OOCGGGGDTGGGGGGGGGGGGDDDTGGGDO.",  # 07
    ".OCGGGGtGGDtGDDGGDDGtDGGDtOGDDO.",  # 08
    "..OCGDoGGGtSSSSDDSSSStGGDoOGDDO.",  # 09
    "..OGGDoGGSSOOOSSSSOOOSSGDoOGDDO.",  # 10
    "..OGGDoGGSSWXOSSSSWXOSSGDoOGDDO.",  # 11
    "..OGDDoGGSS2iOSSSS2iOSSGDoOGDDO.",  # 12
    "..OGDDoGGSSbbSSSSSSbbSSGDoOGDDO.",  # 13
    "..OGDDoGGSSSSSSmmSSSSSSGDoOGDDO.",  # 14
    "..OGGDtGGGtccccSScccctGGDtOGGDO.",  # 15
    "...OGDDtGGtOOOOOOOOOOtGDtOGDDO..",  # 16
    "...OOGGDOtOOWWWVVWWWOtOOOGGDOO..",  # 17
    "....OOGGOOOKKWWVVWWKKOOOGGDOO...",  # 18
    ".....OOOOOKKKWWVVWWKKKOOOOOO....",  # 19
    ".....OOSSOKKKKWVVWKKKKOSSOO.....",  # 20 Arms swing
    "......OccO.OOVVVVVvvvOOccO......",  # 21
    ".........OOCCCCCCCCaaaaOO.......",  # 22 Mini-skirt sways right
    ".........OOCaCCaCCaaCCaOO.......",  # 23
    ".........OOOOOOOOOOOOOOOO.......",  # 24
    "..........OSccO....OWWWwO.......",  # 25 Right leg-warmer lifted
    ".........OWWWwO...OWWWWWwO......",  # 26
    "........OWWWWWwO..OWwwwwWO......",  # 27
    "........OWwwwwWO.sOTTTToss......",  # 28 Right platform shoe lifted
    ".......ssOTTTTos..ssOOOOss......",  # 29 Left platform shoe planted
    "........ssOOOOO.................",  # 30
    ".........ssssssssssssss.........",  # 31
]


# ==============================================================================
# 6. VIP TARGET OVERLAY (32x32)
# Floating Golden Crown (G,D,W with O outline at rows 00..04 center)
# + golden sparkle pixels around the sides
# ==============================================================================
VIP_AURA_32: List[str] = [
    "...........OO..OO..OO...........",  # 00 Crown points top outline
    "..........OWGOOGGOOGDO..........",  # 01 Crown points gold + shine
    "..........OGGGWGGGGDDO..........",  # 02 Crown body
    "..........OGGDGGGGDDDO..........",  # 03 Crown band
    "..........OOOOOOOOOOOO..........",  # 04 Crown base outline
    "................................",  # 05
    "...G........................W...",  # 06 Golden side sparkles
    "..GWG......................GWG..",  # 07
    "...G........................D...",  # 08
    "................................",  # 09
    "................................",  # 10
    ".W............................G.",  # 11
    "................................",  # 12
    "..D..........................W..",  # 13
    "............................GWG.",  # 14
    ".GWG.........................G..",  # 15
    "..G.............................",  # 16
    "................................",  # 17
    ".............................D..",  # 18
    "..W.............................",  # 19
    "..........................G.....",  # 20
    ".G.D.....................GWG....",  # 21
    "..G.......................G.....",  # 22
    "................................",  # 23
    "...W........................D...",  # 24
    "................................",  # 25
    "................................",  # 26
    "................................",  # 27
    "................................",  # 28
    "................................",  # 29
    "................................",  # 30
    "................................",  # 31
]


GIRL_ARCHETYPES: Dict[str, List[List[str]]] = {
    "girl_kawaii": [GIRL_KAWAII_IDLE_0, GIRL_KAWAII_WALK_1, GIRL_KAWAII_WALK_2],
    "girl_emo": [GIRL_EMO_IDLE_0, GIRL_EMO_WALK_1, GIRL_EMO_WALK_2],
    "girl_student": [GIRL_STUDENT_IDLE_0, GIRL_STUDENT_WALK_1, GIRL_STUDENT_WALK_2],
    "girl_gyaru": [GIRL_GYARU_IDLE_0, GIRL_GYARU_WALK_1, GIRL_GYARU_WALK_2],
}


def main() -> None:
    os.makedirs(OUT_DIR, exist_ok=True)
    generated: List[str] = []

    # 1. Alfredo (8 frames + 256x32 sheet + 8x preview)
    alfredo_canvases: List[Canvas] = []
    for name, grid in ALFREDO_FRAMES:
        c = build_frame(grid, name)
        path = os.path.join(OUT_DIR, f"{name}.png")
        c.save(path)
        alfredo_canvases.append(c)
        generated.append(path)

    alfredo_sheet = pack_horizontal_strip(alfredo_canvases)
    save_asset_and_preview(alfredo_sheet, OUT_DIR, "alfredo_sheet", zoom=8)
    generated.append(os.path.join(OUT_DIR, "alfredo_sheet.png"))
    generated.append(os.path.join(PREVIEW_DIR, "preview_alfredo_sheet_8x.png"))

    # 2. All 4 Girl Archetypes (idle_0 individual + 96x32 3-frame sheet + 8x preview)
    for archetype, grids in GIRL_ARCHETYPES.items():
        frames = [build_frame(g, f"{archetype}_{i}") for i, g in enumerate(grids)]
        idle_path = os.path.join(OUT_DIR, f"{archetype}_0.png")
        frames[0].save(idle_path)
        generated.append(idle_path)

        sheet = pack_horizontal_strip(frames)
        save_asset_and_preview(sheet, OUT_DIR, f"{archetype}_sheet", zoom=8)
        generated.append(os.path.join(OUT_DIR, f"{archetype}_sheet.png"))
        generated.append(os.path.join(PREVIEW_DIR, f"preview_{archetype}_sheet_8x.png"))

    # 3. VIP Aura overlay (32x32 + 8x preview)
    vip_canvas = build_frame(VIP_AURA_32, "vip_aura")
    save_asset_and_preview(vip_canvas, OUT_DIR, "vip_aura", zoom=8)
    generated.append(os.path.join(OUT_DIR, "vip_aura.png"))
    generated.append(os.path.join(PREVIEW_DIR, "preview_vip_aura_8x.png"))

    print(f"Generated {len(generated)} files in {OUT_DIR}:")
    for p in generated:
        print(f"  - {p}")


if __name__ == "__main__":
    main()
