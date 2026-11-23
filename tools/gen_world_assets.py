#!/usr/bin/env python3
"""
GIRLBUS: NEON STOP (girlbusAG) — World, Bus Stops, Vehicles, Police, Bosses,
Projectiles & Street Tiles Pixel-Art Generator.

Generates 1x native PNG sprites and nearest-neighbor inspection previews in:
    assets/sprites/world/ (1x) and tools/previews/world/ (previews)
"""

import os
import sys
from typing import Dict, List

# Allow running from any working directory
TOOLS_DIR = os.path.dirname(os.path.abspath(__file__))
ROOT_DIR = os.path.dirname(TOOLS_DIR)
if TOOLS_DIR not in sys.path:
    sys.path.insert(0, TOOLS_DIR)

from pixel_canvas import (
    Canvas,
    PALETTE,
    RGBA,
    grid_to_canvas,
    hex_to_rgba,
    save_asset_and_preview,
)

OUT_DIR = os.path.join(ROOT_DIR, "assets", "sprites", "world")

# Extended palette for street tiles (exact bible asphalt/sidewalk hexes)
WORLD_PALETTE: Dict[str, RGBA] = {
    **PALETTE,
    "4": hex_to_rgba("#524E61"),  # Warm slate-purple asphalt base
    "5": hex_to_rgba("#464254"),  # Asphalt darker pebble/grain
    "6": hex_to_rgba("#605C70"),  # Asphalt lighter pebble/grain
    "7": hex_to_rgba("#D8CEBE"),  # Warm stone sidewalk base
    "8": hex_to_rgba("#BCB2A2"),  # Sidewalk 16x16 tile groove
    "9": hex_to_rgba("#E4DCCD"),  # Sidewalk subtle top-left highlight
}


# ==============================================================================
# 1. ALL 4 BUS STOPS (48x48) + STANDALONE VENDING SHOP (32x32)
# ==============================================================================

# 1A. Kawaii Konbini Stop (48x48) - from FERMATA_KONBINI_PEAK_48 ('u' -> '1' for wood highlight)
STOP_KAWAII_48: List[str] = [
    "................................................",  # 00
    "................................................",  # 01
    "................................................",  # 02
    "....OOOOOOOOOOOOOOOOOOOOOOOOOOO.................",  # 03
    "...OMdMMdMMdMMdMMdMMdMMdMMdMMdMO................",  # 04
    "...OtttttttttttttttttttttttttttO................",  # 05
    "...OMMMMMMMMMMMMMMMMMMMMMMMMMOyO................",  # 06
    "...OMNNNNNNNNNNNNNNNNNNNNNNNyYlO................",  # 07
    "...OMNFFNNNNCCCNCNCNCCCNCNCNyyYllOO.............",  # 08
    "...OMNFNNNNNCNCNCNCNCNNNCNCNyyYllO..............",  # 09
    "...OMNNNNNNNCCCNCNCNCCCNCNCNyYYlOO..............",  # 10
    "...OMNNNNNNNCNCNCNCNNNCNCNCNNyYOO...............",  # 11
    "...OMnnnnnnnCCCnCCCnCCCnCnCnnyYO................",  # 12
    "..OOMMMMMMMMMMMMMMMMMMMMMMMMMMyOO...............",  # 13
    ".OOFFFFCCCCFFFFCCCCFFFFCCCCFFFFCOO..............",  # 14
    ".OFFFFFCCCCFFFFCCCCFFFFCCCCFFFFCCOO.............",  # 15
    ".OFFFFFCCCCFFFFCCCCFFFFCCCCFFFFCCCO.............",  # 16
    ".OFFFFFCCCCFFFFCCCCFFFFCCCCFFFFCCCO.............",  # 17
    ".OFFFFFCCCCFFFFCCCCFFFFCCCCFFFFCCCO.............",  # 18
    ".OfffffaaaaffffaaaaffffaaaaffffaaaO.............",  # 19
    ".OfffffaaaaffffaaaaffffaaaaffffaaaOOOOOOOOOOOO..",  # 20
    "..OOOOOaaaaOOOOaaaaOOOOaaaaOOOOaaaORRRRRRRRRRRO.",  # 21
    "....OMOnnnnndMnnnnnndMnnnnnndMOOOOORWWWWWWWWWRO.",  # 22
    "....OMggggggdMggggggdMggggggdMO...ORWggggggggRO.",  # 23
    "....OMggggzgdMggWWWGdMggggzgdMO...ORWgXgwggFgRO.",  # 24
    "....OMgggzggdMggWSSWdMgggzggdMO...ORWgXgwggFgRO.",  # 25
    "....OMggzgggdMggWWWWdMggzgggdMO...ORWMMMMMMMMRO.",  # 26
    "....OMgzggggdMggggggdMgzggggdMO...ORWgYggCggXRO.",  # 27
    "....OMggggggdMggggggdMggggggdMO...ORWgYggCggXRO.",  # 28
    "....OMgWWWWgdMggggggdMggggggdMO...ORWMMMMMMMMRO.",  # 29
    "....OMgWFFWgdMggggggdMggggggdMO...ORWggggggggRO.",  # 30
    "....OMgWWWWgdMggggggdMggggggdMO...ORWWWWWWWWWRO.",  # 31
    "....OMggggggdMggggggdMggggggdMO...ORRRRRRRRRROOO",  # 32
    "....OMggggggdMggggggdMggggggdMO...ORROCCCCOROyYO",  # 33
    "....OMggggggdMggggggdMggggggdMO...ORRRRRRRRROYlO",  # 34
    "....OMO111111111111111111111OMO...ORROnnnnORYllO",  # 35
    "....OMOTTTTTTTTTTTTTTTTTTTTTOMO...ORROnggnOYYllO",  # 36
    "....OMOtttttttttttttttttttttOMO...ORROnnnnOyyYlO",  # 37
    "....OMOOOOOOtMOOOOOOtMOOOOOOOMO...ORRRRRRROyyyYO",  # 38
    "....OMggggggdMggggggdMggggggdMO...OrrrrrrrOOTT1O",  # 39
    "....OMdddddddMdddddddMdddddddMO...OOOOOOOOOOTttO",  # 40
    "...sOOOOOOOOOOOOOOOOOOOOOOOOOOOs.sOOOOOOOOOOTttO",  # 41
    "..sssssssssssssssssssssssssssssssssssssssssOOOOO",  # 42
    "................................................",  # 43
    "................................................",  # 44
    "................................................",  # 45
    "................................................",  # 46
    "................................................",  # 47
]

# 1B. Emo / Gothic Grunge Stop (48x48)
# Grunge Purple & Charcoal awning (U,u & P,p), dark iron roof (P,p), glowing Purple/Pink EMO || signboard,
# purple moss/vines (U,Z), cracked dark glass w/ skull & broken-heart posters, dark iron bench (M,d),
# Purple/Black Gothic Drink Dispenser on the right!
STOP_EMO_48: List[str] = [
    "................................................",  # 00
    "................................................",  # 01
    "................................................",  # 02
    "....OOOOOOOOOOOOOOOOOOOOOOOOOOO.................",  # 03
    "...OPpPPpPPpPPpPPpPPpPPpPPpPPpPO................",  # 04 - Dark iron roof tiles
    "...OvvvvvvvvvvvvvvvvvvvvvvvvvvvO................",  # 05 - Dark plum beam
    "...OPPPPPPPPPPPPPPPPPPPPPPPPPouO................",  # 06 - Iron signboard frame + purple vine tip
    "...OPvvvvvvvvvvvvvvvvvvvvvvvuUZO................",  # 07 - Deep plum board + purple moss
    "...OPvKKvvvZZZvZvZvZZZvZvZvvUZO.................",  # 08 - Pink tag + 'EMO ||' top + purple moss
    "...OPvKvvvvZvvvZZvZZvZvZvZvUZZO.................",  # 09 - 'E M O ||' upper
    "...OPvvvvvvZZZvZvZvZvZvZvZvUUZZOO...............",  # 10 - 'E M O ||' mid
    "...OPvvvvvvZvvvZvvvZvZvZvZvuUZOO................",  # 11 - 'E M O ||' lower
    "...OPvvvvvvZZZvZvvvZvZZZvZvuZOO.................",  # 12 - 'E M O ||' bottom
    "..OOPPPPPPPPPPPPPPPPPPPPPPPPPuOO................",  # 13 - Iron signboard bottom frame
    ".OOUUUUPPPPUUUUPPPPUUUUPPPPUUUUPO...............",  # 14 - Grunge Purple & Charcoal striped awning!
    ".OUUUUUPPPPUUUUUPPPPUUUUUPPPPUUPPO..............",  # 15
    ".OUUUUUPPPPUUUUUPPPPUUUUUPPPPUUPPPO.............",  # 16
    ".OUUUUUPPPPUUUUUPPPPUUUUUPPPPUUPPPO.............",  # 17
    ".OUUUUUPPPPUUUUUPPPPUUUUUPPPPUUPPPO.............",  # 18
    ".OuuuuuppppuuuuuppppuuuuuppppuuuuPO.............",  # 19 - Awning fold crease
    ".OuuuuuppppuuuuuppppuuuuuppppuuuuPOOOOOOOOOOOO..",  # 20 - Awning skirt + Gothic Dispenser top
    "..OOOOOppppOOOOppppOOOOppppOOOOpppOUUUUUUUUUUUO.",  # 21 - Scalloped awning + Dispenser roof
    "....OPOppppppPppppppPppppppPOOOOOOOUZZZZZZZZZUO.",  # 22 - Dark shadow + Dispenser neon trim
    "....OPggggggpPggggggpPggggggpPO...OUZggggggggUO.",  # 23 - Cracked dark glass + display
    "....OPgoggggpPggWWWgpPggggogpPO...OUZgKggZggKUO.",  # 24 - Glass crack + Skull poster top + cans
    "....OPggogggpPggWgWgpPgggoggpPO...OUZgKggZggKUO.",  # 25 - Glass crack + Skull eyes + cans
    "....OPgggoggpPgggWggpPggogggpPO...OUZPPPPPPPPUO.",  # 26 - Glass crack + Skull jaw + shelf
    "....OPggggogpPggggggpPgoggggpPO...OUZgZggKggZUO.",  # 27 - Glass crack + cans row 2
    "....OPggggggpPggggggpPggggggpPO...OUZgZggKggZUO.",  # 28 - Dark glass + cans row 2
    "....OPgKoKggpPggggggpPggggggpPO...OUZPPPPPPPPUO.",  # 29 - Broken-heart poster top + shelf 2
    "....OPgKKKggpPggggggpPggggggpPO...OUZggggggggUO.",  # 30 - Broken-heart poster mid
    "....OPggKogkpPggggggpPggggggpPO...OUZZZZZZZZZUO.",  # 31 - Broken-heart crack + middle bar
    "....OPggggggpPggggggpPggggggpPO...OUUUUUUUUUUOOO",  # 32 - Lower glass + Purple bush top
    "....OPggggggpPggggggpPggggggpPO...OUUOKKKKOUOZuO",  # 33 - Dispenser pink logo + bush
    "....OPggggggpPggggggpPggggggpPO...OUUUUUUUUOUZZO",  # 34 - Dispenser body + bush
    "....OPOMMMMMdMMMMMMdMMMMMMdMMPO...OUUOppppOUZZZO",  # 35 - Dark iron bench top + dispenser
    "....OPOMddddddddddddddddddddMPO...OUUOnggnOUUZZO",  # 36 - Dark iron bench seat + flap
    "....OPOppppppppppppppppppppppPO...OUUOppppOuuZuO",  # 37 - Iron bench shadow + bush shadow
    "....OPOOOOOOdPOOOOOOdPOOOOOOOPO...OuuuuuuuOuuuUO",  # 38 - Bench legs + bush base
    "....OPggggggpPggggggpPggggggpPO...OuuuuuuuOOPPMO",  # 39 - Base + Charcoal pot
    "....OPpppppppPpppppppPpppppppPO...OOOOOOOOOOPppO",  # 40 - Curb + Pot body
    "...sOOOOOOOOOOOOOOOOOOOOOOOOOOOs.sOOOOOOOOOOPppO",  # 41 - Ground outline + shadow
    "..sssssssssssssssssssssssssssssssssssssssssOOOOO",  # 42 - Ground shadow
    "................................................",  # 43
    "................................................",  # 44
    "................................................",  # 45
    "................................................",  # 46
    "................................................",  # 47
]

# 1C. Student Campus LED Blue Stop (48x48)
# Campus Blue & Cream striped awning (J,j & C,a), digital LED signboard (g) with glowing Cyan UNI || text (X,3),
# Campus Bell/Clock on top, glass wall with geometric formula posters, warm wood bench,
# Blue Coffee/Energy Drink Vending Machine (J,B,X) on the right!
STOP_STUDENT_48: List[str] = [
    ".............OOOOOOO............................",  # 00 - Campus Clock top outline
    "............OGWWWWWG0..........................."[:48].replace("0", "O"),  # 01 - Clock face top
    "............OGWWOWWGO...........................",  # 02 - Clock hands 12
    "....OOOOOOOOOGWWOOOGO0OOOOOO...................."[:48].replace("0", "O"),  # 03 - Clock 3 + roof top
    "...OBdBBdBBdBOOOOOOOBdBBdBBdBO..................",  # 04 - Campus steel roof tiles
    "...OjjjjjjjjjjjjjjjjjjjjjjjjjjjO................",  # 05 - Blue roof beam
    "...OMMMMMMMMMMMMMMMMMMMMMMMMMOyO................",  # 06 - LED signboard frame + ivy
    "...OMgggggggggggggggggggggggyYlO................",  # 07 - Digital LED signboard + ivy
    "...OMgBBggggXggXgXXgXgXgXgXgyyYllOO.............",  # 08 - Blue tag + glowing Cyan 'UNI ||'
    "...OMgBgggggXggXgXXgXgXgXgXgyyYllO..............",  # 09 - 'U N I ||'
    "...OMgggggggXggXgXXgXgXgXgXgyYYlOO..............",  # 10 - 'U N I ||'
    "...OMgggggggXggXgXgXXgXgXgXgglYOO...............",  # 11 - 'U N I ||'
    "...OMggggggg3XX3gXgXXgXgXgXggyyO................",  # 12 - 'U N I ||' bottom glow
    "..OOMMMMMMMMMMMMMMMMMMMMMMMMMMyOO...............",  # 13 - Frame bottom
    ".OOJJJJCCCCJJJJCCCCJJJJCCCCJJJJCOO..............",  # 14 - Campus Blue & Cream striped awning!
    ".OJJJJJCCCCJJJJCCCCJJJJCCCCJJJJCCOO.............",  # 15
    ".OJJJJJCCCCJJJJCCCCJJJJCCCCJJJJCCCO.............",  # 16
    ".OJJJJJCCCCJJJJCCCCJJJJCCCCJJJJCCCO.............",  # 17
    ".OJJJJJCCCCJJJJCCCCJJJJCCCCJJJJCCCO.............",  # 18
    ".OjjjjjaaaajjjjaaaajjjjaaaajjjjaaaO.............",  # 19 - Awning fold crease
    ".OjjjjjaaaajjjjaaaajjjjaaaajjjjaaaOOOOOOOOOOOO..",  # 20 - Awning skirt + Blue Vending top
    "..OOOOOaaaaOOOOaaaaOOOOaaaaOOOOaaaOJJJJJJJJJJJO.",  # 21 - Scalloped awning + Vending roof
    "....OMOnnnnndMnnnnnndMnnnnnndMOOOOOJBWWWWWWWWJO.",  # 22 - Shadow + Vending blue/white trim
    "....OMggggggdMggggggdMggggggdMO...OJWggggggggJO.",  # 23 - Shelter glass + display
    "....OMggggzgdMggWWXgdMggggzgdMO...OJWgTgwggXgJO.",  # 24 - Formula triangle poster + coffee cans
    "....OMgggzggdMggWXWgdMgggzggdMO...OJWgTgwggXgJO.",  # 25 - Formula poster mid + coffee cans
    "....OMggzgggdMggXXXgdMggzgggdMO...OJWMMMMMMMMJO.",  # 26 - Formula poster base + shelf 1
    "....OMgzggggdMggggggdMgzggggdMO...OJWgGggBggXJO.",  # 27 - Energy cans row 2
    "....OMggggggdMggggggdMggggggdMO...OJWgGggBggXJO.",  # 28 - Energy cans row 2
    "....OMgWWWWgdMggggggdMggggggdMO...OJWMMMMMMMMJO.",  # 29 - Campus timetable poster + shelf 2
    "....OMgWGXGgdMggggggdMggggggdMO...OJWggggggggJO.",  # 30 - Timetable chart + display bottom
    "....OMgWWWWgdMggggggdMggggggdMO...OJBWWWWWWWWJO.",  # 31 - Poster bottom + middle bar
    "....OMggggggdMggggggdMggggggdMO...OJJJJJJJJJJOOO",  # 32 - Lower glass + Potted Bush top
    "....OMggggggdMggggggdMggggggdMO...OJJOXXXXOJOyYO",  # 33 - Glowing cyan 'COFFEE' panel + bush
    "....OMggggggdMggggggdMggggggdMO...OJJJJJJJJOJYlO",  # 34 - Vending body + bush highlight
    "....OMO111111111111111111111OMO...OJJOnnnnOJYllO",  # 35 - Wooden bench highlight + dispenser
    "....OMOTTTTTTTTTTTTTTTTTTTTTOMO...OJJOnggnOYYllO",  # 36 - Wooden bench seat + flap
    "....OMOtttttttttttttttttttttOMO...OJJOnnnnOyyYlO",  # 37 - Wooden bench shadow + bush shadow
    "....OMOOOOOOtMOOOOOOtMOOOOOOOMO...OjjjjjjjOyyyYO",  # 38 - Bench legs + bush base
    "....OMggggggdMggggggdMggggggdMO...OjjjjjjjOOTT1O",  # 39 - Curb + Terracotta pot
    "....OMdddddddMdddddddMdddddddMO...OOOOOOOOOOTttO",  # 40 - Curb + Pot body
    "...sOOOOOOOOOOOOOOOOOOOOOOOOOOOs.sOOOOOOOOOOTttO",  # 41 - Base outline + shadow
    "..sssssssssssssssssssssssssssssssssssssssssOOOOO",  # 42 - Ground shadow
    "................................................",  # 43
    "................................................",  # 44
    "................................................",  # 45
    "................................................",  # 46
    "................................................",  # 47
]

# 1D. Cyber-Gyaru 5G Hologram Stop (48x48)
# Neon Cyan & Sakura Pink striped awning (X,2 & K,k), golden roof trim (G,D),
# glowing '5G  ||' signboard with heart icon (K,X), holographic glass panels, plush pink bench (K,Q),
# Golden/Pink 5G Selfie & Drink Kiosk on the right!
STOP_GYARU_48: List[str] = [
    "................................................",  # 00
    "................OOO.....OOO.....................",  # 01 - Holo antenna / star deco
    "...............OX3O.....OKQO....................",  # 02 - Glowing holo ears/beacons
    "....OOOOOOOOOOOOOOOOOOOOOOOOOOO.................",  # 03 - Golden roof outline
    "...OGDGGDGGDGGDGGDGGDGGDGGDGGDGO................",  # 04 - Golden roof trim!
    "...ODDDDDDDDDDDDDDDDDDDDDDDDDDDO................",  # 05 - Deep gold beam
    "...OGGGGGGGGGGGGGGGGGGGGGGGGGOkO................",  # 06 - Gold signboard frame + sakura
    "...OGVggggggggggggggggggggggKQKOO...............",  # 07 - Holo board + sakura blossom
    "...OGgKKggggXXXgXXggKKKggXgXKQKKOO..............",  # 08 - '5 G' + Heart icon top
    "...OGgKgggggXgggXgggKKKKKXggKQKOO...............",  # 09 - '5 G' + Heart mid
    "...OGgggggggXXXgXgXgKKKKKXggKKOO................",  # 10 - '5 G' + Heart mid
    "...OGgggggggggXgXgXggKKKgXgXKKOO................",  # 11 - '5 G' + Heart bottom
    "...OGgggggggXXXgXXXgggKggXgXKQO.................",  # 12 - '5 G' + Heart tip
    "..OOGGGGGGGGGGGGGGGGGGGGGGGGGQOO................",  # 13 - Gold frame bottom
    ".OOXXXXKKKKXXXXKKKKXXXXKKKKXXXXKOO..............",  # 14 - Neon Cyan & Sakura Pink striped awning!
    ".OXXXXXKKKKXXXXXKKKKXXXXXKKKKXXXKKO.............",  # 15
    ".OXXXXXKKKKXXXXXKKKKXXXXXKKKKXXXKKKO............",  # 16
    ".OXXXXXKKKKXXXXXKKKKXXXXXKKKKXXXKKKO............",  # 17
    ".OXXXXXKKKKXXXXXKKKKXXXXXKKKKXXXKKKO............",  # 18
    ".O22222kkkk22222kkkk22222kkkk222kkkO............",  # 19 - Awning fold crease
    ".O22222kkkk22222kkkk22222kkkk222kkkOOOOOOOOOOO..",  # 20 - Awning skirt + 5G Kiosk top!
    "..OOOOOkkkkOOOOOkkkkOOOOOkkkkOOOkkkOGGGGGGGGGGO.",  # 21 - Scalloped awning + Gold Kiosk roof
    "....OGOnnnnndGnnnnnndGnnnnnndGOOOOOOGKKKKKKKKGO.",  # 22 - Shadow + Gold/Pink Kiosk header
    "....OGggggzgdGggggzgdGggggzgdGO...OGKg333333KGO.",  # 23 - Holo glass + Ring-light selfie screen!
    "....OGgggzggdGggWKWGdGgggzggdGO...OGKg3WWWW3KGO.",  # 24 - Star/heart holo poster + Ring light
    "....OGggzgggdGggKKKgdGggzgggdGO...OGKg3WKKW3KGO.",  # 25 - Heart poster + Selfie heart center
    "....OGgzggggdGggWKWgdGgzggggdGO...OGKg3WWWW3KGO.",  # 26 - Holo streak + Ring light bottom
    "....OGgggg2gdGggggggdGgggg2gdGO...OGKg333333KGO.",  # 27 - Cyan holo shimmer + Screen base
    "....OGggg2ggdGggggggdGggg2ggdGO...OGKGGGGGGGGGO.",  # 28 - Holo shimmer + Gold shelf
    "....OGgGGGGgdGggggggdGggggggdGO...OGKgXgKgGggGO.",  # 29 - Gold fashion poster + Pink/Cyan drinks
    "....OGgGKKGgdGggggggdGggggggdGO...OGKgXgKgGggGO.",  # 30 - Fashion poster + Drinks
    "....OGgGGGGgdGggggggdGggggggdGO...OGGGGGGGGGKGO.",  # 31 - Poster base + Gold bar
    "....OGggggggdGggggggdGggggggdGO...OKKKKKKKKKKOOO",  # 32 - Lower glass + Cyber-Sakura Bush top
    "....OGggggggdGggggggdGggggggdGO...OKKOXXXXOKOKQO",  # 33 - 5G neon panel + Pink/Cyan bush
    "....OGggggggdGggggggdGggggggdGO...OKKKKKKKKOKQqO",  # 34 - Kiosk body + Bush highlight
    "....OGOQQQQQQQQQQQQQQQQQQQQQOGO...OKKOnnnnOKQQqO",  # 35 - Plush pink bench highlight + slot
    "....OGOKKKKKKKKKKKKKKKKKKKKKOGO...OKKOnggnOKKQOO",  # 36 - Plush pink bench cushion + flap
    "....OGOkkkkkkkkkkkkkkkkkkkkkOGO...OKKOnnnnOkkKOO",  # 37 - Plush bench shadow + bush shadow
    "....OGOOOOOOGGOOOOOOGGOOOOOOGGO...OkkkkkkkOkkkKO",  # 38 - Gold bench legs + bush base
    "....OGggggggdGggggggdGggggggdGO...OkkkkkkkOOGDGO",  # 39 - Base + Gold pot
    "....OGDDDDDDDGDDDDDDDGDDDDDDDGO...OOOOOOOOOOGDDO",  # 40 - Gold curb + Gold pot body
    "...sOOOOOOOOOOOOOOOOOOOOOOOOOOOs.sOOOOOOOOOOGDDO",  # 41 - Ground outline + shadow
    "..sssssssssssssssssssssssssssssssssssssssssOOOOO",  # 42 - Ground shadow
    "................................................",  # 43
    "................................................",  # 44
    "................................................",  # 45
    "................................................",  # 46
    "................................................",  # 47
]

# 1E. Standalone Neon Drink Vending Shop (32x32)
# Bright red body (R,r), lighted glass display with colorful drink cans (X,Y,F,C),
# glowing SHOP top header, and terracotta potted bush (Y,y,l,T,t,1) on the right!
SHOP_VENDING_32: List[str] = [
    "................................",  # 00
    "................................",  # 01
    "...OOOOOOOOOOOOOOOOOO...........",  # 02 - Vending top outline
    "..ORRRRRRRRRRRRRRRRRRO..........",  # 03 - Bright red roof
    "..ORggggggggggggggggRO..........",  # 04 - Glowing SHOP header frame
    "..ORgXXXgXgXgXXXgXXXRO..........",  # 05 - S H O P pixel top (3x5 each)
    "..ORgXgggXgXgXgXgXgXRO..........",  # 06 - S H O P pixel upper
    "..ORgXXXgXXXgXgXgXXXRO..........",  # 07 - S H O P pixel mid
    "..ORgggXgXgXgXgXgXggRO..........",  # 08 - S H O P pixel lower
    "..ORgXXXgXgXgXXXgXggRO..........",  # 09 - S H O P pixel base
    "..ORWWWWWWWWWWWWWWWWRO..........",  # 10 - Lighted white inner frame
    "..ORWggggggggggggggWRO..........",  # 11 - Glass display top
    "..ORWgXXggYYggFFggCWRO..........",  # 12 - Drink cans row 1 (Cyan, Green, Coral, Cream)
    "..ORWgXXggYYggFFggCWRO..........",  # 13 - Drink cans row 1
    "..ORWMMMMMMMMMMMMMMWRO..........",  # 14 - Metal shelf 1
    "..ORWgCCggXXggKKggYWRO..........",  # 15 - Drink cans row 2 (Cream, Cyan, Pink, Green)
    "..ORWgCCggXXggKKggYWRO..........",  # 16 - Drink cans row 2
    "..ORWMMMMMMMMMMMMMMWRO..........",  # 17 - Metal shelf 2
    "..ORWgFFggCCggXXggGWRO..........",  # 18 - Drink cans row 3 (Coral, Cream, Cyan, Gold)
    "..ORWWWWWWWWWWWWWWWWRO..OOOOO...",  # 19 - Display bottom + Potted bush top!
    "..ORRRRRRRRRRRRRRRRRRO.OllYYyO..",  # 20 - Red body + Bush foliage
    "..ORROCCCCCCCCOORXGRROOllYYYYyO.",  # 21 - Glowing logo panel + Coin slot + Bush
    "..ORROCCCCCCCCOORrGRROOllYYYYyyO",  # 22 - Logo panel + Bush
    "..ORRRRRRRRRRRRRRRRRROYYYYyyyOO.",  # 23 - Red body + Bush lower
    "..ORRROggggggggOORRRO.OyyyyyOO..",  # 24 - Dispenser slot + Bush base
    "..ORRROgzzzzzzgOORRRO..OT11TO...",  # 25 - Dispenser flap + Terracotta pot rim
    "..ORRROggggggggOORRRO..OTTTtO...",  # 26 - Dispenser base + Terracotta pot
    "..ORRRRRRRRRRRRRRRRRRO.OOTttO...",  # 27 - Red base + Terracotta pot
    "..OrrrrrrrrrrrrrrrrrrO..OTttO...",  # 28 - Red bottom shadow + Pot base
    "..sOOOOOOOOOOOOOOOOOOs..OOOOO...",  # 29 - Base outline
    "...ssssssssssssssssss..sssssss..",  # 30 - Ground shadow
    "................................",  # 31
]


# ==============================================================================
# 2. THE NEON BUS (96x48 horizontal, facing RIGHT)
# ==============================================================================
# Coral-Pink & Cream upper body (F,f,C,a), Slate-Blue lower skirt (N,n),
# Glowing Neon Cyan speed stripe (X,3), Roof AC units (M,d), LED destination board (g with X/G),
# Panoramic windows (g,z) with cute chibi silhouettes, double folding doors, headlights (G,X),
# dark tires (O,P,M) with ground shadow (s).
BUS_NEON_96: List[str] = [
    "................................................................................................",  # 00
    "................................................................................................",  # 01
    "................................................................................................",  # 02
    "................................................................................................",  # 03
    "................................................................................................",  # 04
    "..............OOOOOOOOOOOOOO..............OOOOOOOOOOOOOO........................................",  # 05 - Roof AC units top
    ".............OMWWWWWWWWWWWWMO............OMWWWWWWWWWWWWMO.......................................",  # 06 - Roof AC highlights
    ".............OMMggggggMMMMdMO............OMMggggggMMMMdMO.......................................",  # 07 - AC fan grilles
    ".............OMddddddddddddMO............OMddddddddddddMO.......................................",  # 08 - AC unit bases
    ".......OOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOO............",  # 09 - Bus roof outline
    "......OCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCO..........",  # 10 - Cream upper roof highlight
    ".....OCCCCCGGGGGGGGGGGGGGGGCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCgGGGGGGgCCCCCO.........",  # 11 - Roof trim + front LED dest board!
    "....OCCCCCCgXXgXggXXgXXgggCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCgGXXGXXGgCCCCCCOO.......",  # 12 - Side LED 'BUS' + Front LED sign!
    "....OCCCCCCgggggggggggggggCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCgggggggggCCCCCCCO.......",  # 13 - Cream upper band
    "...OFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFO......",  # 14 - Coral-Pink body top
    "...OFFFFOgggggggggggOOgggggggggggOOgggggggggggOOgggggggggOOgggggggggggOOgggggggggggggOFFFFO.....",  # 15 - Panoramic windows + door glass top
    "...OFFFFOgzzzzggggggOOggzzzzgggggOOggzzzzgggggOOggzggzgggOOggzzzzgggggOOggzzzzzggggggOFFFFO.....",  # 16 - Glass reflection streaks
    "...OFFFFOgzgSSggggggOOgggggKKggggOOgggggHHggggOOggzggzgggOOgggggTTggggOOggzgggggSggggOFFFFFO....",  # 17 - Chibi heads peeking inside windows!
    "...OFFFFOggSScggggggOOggggKKkggggOOggggHHhggggOOggzggzgggOOggggTTtggggOOggggggggScgggOFFFFFO....",  # 18 - Chibi silhouettes in bus
    "...OFFFFOgggggggggggOOgggggggggggOOgggggggggggOOgggggggggOOgggggggggggOOgggggggggggggOFFFFFO....",  # 19 - Window mid
    "...OFFFFOgggggggggggOOgggggggggggOOgggggggggggOOgggggggggOOgggggggggggOOgggggggggggggOFFFFFO....",  # 20 - Window mid
    "...OFFFFOggggggzzgggOOgggggggzzggOOgggggggzzggOOgggggggggOOgggggggzzggOOggggggggzzgggOFFFFFO....",  # 21 - Lower reflection streaks
    "...OFFFFOgggggggggggOOgggggggggggOOgggggggggggOOgggggggggOOgggggggggggOOgggggggggggggOFFFFFO....",  # 22 - Window lower
    "...OFFFFOOOOOOOOOOOOOOgggggggggggOOOOOOOOOOOOOOOgggggggggOOOOOOOOOOOOOOOgggggggggggggOFFFFFO....",  # 23 - Window sills (double doors stay tall!)
    "...OFFFFFFFFFFFFFFFFFOgggggggggggOFFFFFFFFFFFFFOgggggggggOFFFFFFFFFFFFFFFFFFFFFFFFFFFOFFFFFO....",  # 24 - Coral-Pink mid body + door glass
    "...OFFFFFFFFFFFFFFFFFOgggggggggggOFFFFFFFFFFFFFOgggggggggOFFFFFFFFFFFFFFFFFFFFFFFFFFFOFFFFFO....",  # 25 - Coral-Pink body
    "...OfffffffffffffffffOgggggggggggOfffffffffffffOgggggggggOfffffffffffffffffffffffffffOfffffO....",  # 26 - Coral-Pink body shadow
    "...OCCCCCCCCCCCCCCCCCOCCCCCCCCCCCOCCCCCCCCCCCCCOCCCCCCCCCOCCCCCCCCCCCCCCCCCCCCCCCCCCCOCCCCCO....",  # 27 - Cream beltline trim
    "...O33333333333333333O33333333333O3333333333333O333333333O333333333333333333333333333O33333O....",  # 28 - Neon Cyan speed stripe top glow!
    "...OXXXXXXXXXXXXXXXXXOXXXXXXXXXXXOXXXXXXXXXXXXXOXXXXXXXXXOXXXXXXXXXXXXXXXXXXXXXXXXXXXOXXXXXO....",  # 29 - Neon Cyan speed stripe core!
    "...O22222222222222222O22222222222O2222222222222O222222222O222222222222222222222222222O22222O....",  # 30 - Neon Cyan speed stripe bottom
    "...ORRRNNNNNNNNNNNNNNOgggggggggggONNNNNNNNNNNNNOgggggggggONNNNNNNNNNNNNNNNNNNNNNNNNNNOONGGGO....",  # 31 - Taillight (L) + Slate-Blue skirt + Headlight (R)!
    "...ORrrNNNNNNNNNNNNNNOgggggggggggONNNNNNNNNNNNNOgggggggggONNNNNNNNNNNNNNNNNNNNNNNNNNNOOGXXXO....",  # 32 - Taillight + Skirt + Cyan/Gold Headlight!
    "...ONNNNNNNNNNNNNNNNNOgggggggggggONNNNNNNNNNNNNOgggggggggONNNNNNNNNNNNNNNNNNNNNNNNNNNOOGWWWO....",  # 33 - Lower skirt + Doors + Headlight
    "...ONNNNNNNNOOOOOOOONOOOOOOOOOOOOONNNNNNNNNNNNNOOOOOOOOOOONNNNNNNNNNOOOOOOOONNNNNNNNNOOOOOOO....",  # 34 - Wheel arches top
    "...OnnnnnnnOPPPPPPPPOOnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnOPPPPPPPPOOnnnnnnnOCCCCCCCCO.",  # 35 - Wheel wells + Front bumper!
    "...OnnnnnnOPPPWWWWPPPOOnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnOPPPWWWWPPPOOnnnnnnOCCCCCCCCO.",  # 36 - Rear & Front Wheels + Metallic bumper
    "...OOOOOOOPPPWMMMMWPPPOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOPPPWMMMMWPPPOOOOOOOOssssssss..",  # 37 - Chassis bottom + hubcaps
    "..........OPPWMddMWPPO............................................OPPWMddMWPPO..................",  # 38 - Tires & hubcaps
    "..........OPPWMMMMWPPO............................................OPPWMMMMWPPO..................",  # 39 - Tires & hubcaps
    "..........OPPPWWWWPPPO............................................OPPPWWWWPPPO..................",  # 40 - Tires lower
    "...........OPPPPPPPPO..............................................OPPPPPPPPO...................",  # 41 - Tire tread bottom
    "...........ssOOOOOOss..............................................ssOOOOOOss...................",  # 42 - Tire outline + shadow
    ".........ssssssssssssss..........................................ssssssssssssss.................",  # 43 - Ground shadow
    "................................................................................................",  # 44
    "................................................................................................",  # 45
    "................................................................................................",  # 46
    "................................................................................................",  # 47
]


# ==============================================================================
# 3. POLICE UNITS (1*-5*) — police_officer (32x32) & police_car (48x32)
# ==============================================================================

# 3A. Police Officer / Vigile (32x32) - 1.8-head Chibi scale matching Alfredo
POLICE_OFFICER_32: List[str] = [
    "................................",  # 00
    "..........OOOOOOOOOOOO..........",  # 01 - Peaked police cap crown
    ".........OBBJJJJJJJJjjO.........",  # 02 - Navy cap top
    "........OJJJJJOGGOJJJjjO........",  # 03 - Gold police badge on cap!
    "........OJJJJJOGGOJJJjjO........",  # 04 - Cap band + badge
    "........OBBBBBBBBBBBBBbO........"[:32].replace("b", "B"),  # 05 - Cap band trim
    ".......OOPPPPPPPPPPPPPPOO.......",  # 06 - Dark shiny cap visor
    "........OHHHHHHHHHHHHHhO........",  # 07 - Short hair under cap
    "........OHhSLLLLSSSSScHO........",  # 08 - Forehead & sideburns
    "........OHoSOOOSSOOOScoO........",  # 09 - Stern dark eyebrows
    ".......OSSOSSSSSSSSSScOSSO......",  # 10 - Ears + upper lids
    ".......OScOSWESSSSWEScOScO......",  # 11 - Ears + stern eyes
    ".......OScOSEeSSSSeeScOScO......",  # 12 - Eyes mid
    "........OOOSeeSSSSeeScOOO.......",  # 13 - Eyes bottom
    ".........OSSSSSSSSSSSScO........",  # 14 - Cheeks
    ".........OccSSmmmSSScccO........",  # 15 - Stern expression
    "..........OcccccccccccO.........",  # 16 - Jawline shadow
    "...........OOOOOOOOOOO..........",  # 17 - Chin outline
    "..........OJBoWWJWWoJjO.........",  # 18 - Navy uniform collar + white shirt & tie
    ".........OJJJBoWJWGoJjjO........",  # 19 - Shoulders + tie + gold chest badge!
    "........OJJJJBoWJWGoJjjjO.......",  # 20 - Uniform torso + gold badge
    "........OJJJJJoWJWoJjjjjO.......",  # 21 - Uniform torso
    "........OJJoJJoWWWooJjoJO.......",  # 22 - Sleeves separation
    "........OSSOJJJJJJJJJJOSSO......",  # 23 -Wait: check width! Let's align cols 8..23 (16 chars)!
    "........OccOGGGXOGXGGGOccO......",  # 24 - Hands + High-Vis Yellow/Cyan utility belt!
    "..........OPPPpOOOPppO..........",  # 25 - Dark uniform trousers
    "..........OPPPpOOOPppO..........",  # 26 - Trousers
    "..........OPPPpOOOPppO..........",  # 27 - Trousers lower
    "..........OPPpO..OPppO..........",  # 28 - Dark patrol boots
    "........ssOpppOssOpppOss........",  # 29 - Boots sole
    ".........ssOOOOO.OOOOOss........",  # 30 - Boots outline
    "...........sssssssssss..........",  # 31 - Ground shadow
]
# Align row 23 of POLICE_OFFICER_32 to exactly 32 chars (cols 08..23):
POLICE_OFFICER_32[23] = "........OSSOOJJJJJJJJOSSO......."  # Wait: 8 + 4(OSSO) + 9 + 4(OSSO) + 7 = 32 -> let's keep 8..23 symmetric:
POLICE_OFFICER_32[23] = "........OSSOJJJJJJjjOSSO........"  # 8 + 4 + 8 + 4 + 8 = 32!
POLICE_OFFICER_32[24] = "........OccOGGXOOXGGGOccO......."  # Wait: 8 + 4 + 9 + 4 + 7 = 32 -> make 8..23 symmetric:
POLICE_OFFICER_32[24] = "........OccOGGXGGXGGGOccO......."  # Wait: 8 + 4 + 8 + 4 + 8 = 32 -> 4+9+4=17 -> let's count:
# "........" (8) + "OccO" (4) + "GGXGGXGG" (8) + "OccO" (4) + "........" (8) = 32!
POLICE_OFFICER_32[24] = "........OccOGGXGGXGGOccO........"  # 8 + 4 + 8 + 4 + 8 = 32! Wait: O c c O (4) + G G X G G X G G (8) + O c c O (4) = 16!
# Let's double check: "OccOGGXGGXGG" has OccO(4) + GGXGGXGG(8) + OccO(4)? No, "OccO" + "GGXGGXGG" + "OccO" = 4+8+4=16 -> "OccOGGXGGXGGOccO" is 16?
# O-c-c-O (4) + G-G-X-G-G-X-G-G (8) + O-c-c-O (4) = 16! Wait above I wrote "OccOGGXGGXGGOccO" which is O-c-c-O-G-G-X-G-G-X-G-G-O-c-c-O -> 16! Yes! Wait: in line above I missed one G: "GGXGGXGG" has 8 chars, "GGXGGXG" has 7. Let's write it explicitly:
POLICE_OFFICER_32[24] = "........OccOGGXGGXGGOccO........"  # Let's check: 8 + 16 + 8 = 32! Wait: O-c-c-O (4) + G-G-X-G-G-X-G-G (8) = 12 + O-c-c-O (4) = 16? Look closely: in "OccOGGXGGXGGOccO", between the two O's is G-G-X-G-G-X-G-G? No, G-G-X-G-G-X-G-G is 8, plus O at the end of OccO and O at start of OccO:
# "........OccOGGXGGXGGOccO........" -> count: 8 dots + O(1)c(2)c(3)O(4)G(5)G(6)X(7)G(8)G(9)X(10)G(11)G(12)O(13)c(14)c(15)O(16) + 8 dots = 32! Wait: O(4) G(5) G(6) X(7) G(8) G(9) X(10) G(11) G(12) O(13) -> there's 8 chars between O and O: GGXGGXGG! Wait in my string I typed `GGXGGXGG`? Let's count: `G G X G G X G G` is 8, so `OccOGGXGGXGGOccO` is 16!Wait: earlier I typed `OccOGGXGGXGG`? No, I typed `OccOGGXGGXGGOccO`, let's count characters:
# O(1) c(2) c(3) O(4) G(5) G(6) X(7) G(8) G(9) X(10) G(11) G(12) O(13) c(14) c(15) O(16). Wait: let's make it `OccOGGXOOXGGOccO` -> 4 + 8 + 4 = 16! That puts a belt buckle `OO` or `G` in the middle!
POLICE_OFFICER_32[24] = "........OccOGGXXGGXXGOccO......."[:32]
POLICE_OFFICER_32[24] = "........OccOGGXOOXGGOccO........"  # 8 + 4 + 8 + 4 + 8 = 32!

# 3B. Top-Down 3/4 Chibi Police Cruiser facing RIGHT (48x32)
# Black-and-white panda patrol livery (P,p & W,w), tinted windshield (g,z),
# roof siren lightbar with bright Red (R) on left and Neon Blue/Cyan (X) on right, wheels (O,M).
POLICE_CAR_48: List[str] = [
    "................................................",  # 00
    "................................................",  # 01
    "................................................",  # 02
    "................................................",  # 03
    "................................................",  # 04
    "................................................",  # 05
    "...................OOOOOOOOO....................",  # 06 - Siren lightbar top
    "..................ORRRWWXXX3O...................",  # 07 - Red (L) + White + Cyan/Blue (R) Siren!
    "..............OOOOOOOOOOOOOOOOOO................",  # 08 - Roof top outline
    ".............OWWWWWWWWWWWWWWWWWWO...............",  # 09 - White roof
    "............OWWggggggggggggggggWWO..............",  # 10 - Tinted windows top
    "...........OWWgzzzzggOggzzzzzgggWWO.............",  # 11 - Windshield & side glass
    "..........OggggggggggOgggggggggggggO............",  # 12 - Side windows & windshield base
    ".....OOOOOPPPPPPPPPWWWWWWWWWWPPPPPPPOOO.........",  # 13 - Panda patrol hood/trunk/doors!
    "....ORPPPPPPPPPPPPPWWWWWWWWWWPPPPPPPPPPPO.......",  # 14 - Taillight (L) + Panda body + Hood (R)
    "...ORRPPPPPPPPPPPPPWWGGGGGGWWPPPPPPPPPPGGO......",  # 15 - Taillight + Gold star/badge + Headlight!
    "...OPPPPPPPPPPPPPPPWWWWWWWWWWPPPPPPPPPPXXO......",  # 16 - Panda body + Cyan police headlight!
    "...OPPPPPPPPPPPPPPPWWWWWWWWWWPPPPPPPPPPWWO......",  # 17 - Lower body + Bumper
    "...OppppppppOOOOOOWWWWWWWWWWWWOOOOOOpPPPPPO.....",  # 38 -> wait: 18 - Wheel arches
    "...OppppppOPPPWWPPPOwwwwwwwwOPPPWWPPPOppppO.....",  # 19 - Front & Rear Tires + White door trim
    "....OOOOOOPPWMMMMWPPOOOOOOOOOPPWMMMMWPPOOO......",  # 20 - Chassis + Hubcaps
    "..........OPPWMddMWPPO......OPPWMddMWPPO........",  # 21 - Tires & Hubcaps
    "..........OPPWMMMMWPPO......OPPWMMMMWPPO........",  # 22 - Tires
    "...........OPPPWWPPPO........OPPPWWPPPO.........",  # 23 - Tires lower
    "............OOOOOOOO..........OOOOOOOO..........",  # 24 - Tire bottom outline
    "..........ssssssssssss......ssssssssssss........",  # 25 - Tire ground shadow
    "........ssssssssssssssssssssssssssssssssss......",  # 26 - Car ground shadow
    "................................................",  # 27
    "................................................",  # 28
    "................................................",  # 29
    "................................................",  # 30
    "................................................",  # 31
]


# ==============================================================================
# 4. 3 ICONIC BOSS GATEKEEPERS (36x36 each)
# ==============================================================================

# 4A. Boss 1: La Migliore Amica (36x36)
# Fierce coral-red high ponytail (F,f,R), determined angled eyebrows, sporty varsity jacket (R,W),
# holding a loud Megaphone (left) & Broken-Heart shield (right)!
BOSS_BEST_FRIEND_36: List[str] = [
    "....................................",  # 00
    "..............OOOOO.................",  # 01 - High ponytail peak!
    ".............OFFFFfO................",  # 02 - Coral-red ponytail sweep
    "............OFFFFFFfO...............",  # 03 - Ponytail volume
    "..........OOFFOOOOOffO..............",  # 04 - Ponytail band + crown outline
    ".........OFFFOFFFFFFFffO............",  # 05 - Hair crown + ponytail cascade
    "........OFFFFFFFFFFFFFffO...........",  # 06 - Hair volume
    "........OFFfFFFFFFFFFFFFfO..........",  # 07 - Side locks
    ".......OFFfFFSSFFFFFFSSFFfO.........",  # 08 - Bangs framing forehead
    ".......OFFfSOOSSSSSSSSOOSFFfO.......",  # 09 - Fierce angled eyebrows (outer high)!
    ".......OFfOSSOOSSSSSSOOSSOFfO.......",  # 10 - Fierce angled brows (inner low)!
    ".......OFfOSWROSSSSSSWROSOFfO.......",  # 11 - 2x3 Fierce coral/ruby anime eyes (top-left W!)
    ".......OFfOSRIOSSSSSSRIOSSFfO.......",  # 12 - Eyes mid
    "........OOOSbbSSSSSSbbSSSOO.........",  # 13 - Cheeks + blush
    ".........OOSSSSSSmmSSSSSSOO.........",  # 14 - Determined shout/smirk mouth
    "..........OcccccSSccccccO...........",  # 15 -> keep 36 chars exact:
    "...........OOOOOOOOOOOOOO...........",  # 16 - Chin outline
    "......OOO...ORRRWWWWWRRRO...........",  # 17 - Megaphone top (L) + Varsity collar
    ".....OMWMO.ORRRWWGGGGWRRRO..........",  # 18 - Megaphone cone + Varsity jacket w/ gold trim
    "....OMWFFOORRRWWWWWWWWWRRRO..OOOOO..",  # 19 - Megaphone + Jacket + Broken Heart Shield top!
    "....OMWFFOORRoRRRRWWWRRoRRO.OKKOQKO.",  # 20 - Megaphone + Arms + Heart Shield (K,Q,O crack)!
    "....OOMMMOOSSORRRRRRRRROSSO.OKOKKKO.",  # 21 - Hands holding megaphone & shield!
    "......OO..OccORRRGGGGGRROccOOKOKkKO.",  # 22 - Waist + Shield crack
    "...........OOOPPPPPPPPPOOO...OOKKO..",  # 23 - Sporty dark pleated skirt + Shield tip
    "..........OOPPPPPPPPPPPPOO....OOO...",  # 24 - Skirt mid
    "..........OPPPpPPPPpPPPPpO..........",  # 25 - Skirt pleats
    "..........OWWWWWWWWWWWWWWO..........",  # 26 - White varsity trim hem
    "..........OOOOOOOOOOOOOOOO..........",  # 27 - Skirt outline
    "............OSccO..OSccO............",  # 28 - Legs
    "............OWWwO..OWWwO............",  # 29 - Sporty socks
    ".........sssORWROssORWROsss.........",  # 30 - Red/White sneakers
    "..........ssOOOOOssOOOOOss..........",  # 31 - Sneakers outline
    "............ssssssssssss............",  # 32 - Ground shadow
    "....................................",  # 33
    "....................................",  # 34
    "....................................",  # 35
]
BOSS_BEST_FRIEND_36[12] = ".......OFfOSRIOSSSSSSRIOSOFfO......."
BOSS_BEST_FRIEND_36[15] = "..........OSccccSSSSccccSO.........."

# 4B. Boss 2: La Professoressa (36x36)
# Strict dark-plum hair bun (V,v), sharp Cyan-rimmed glasses (X,W), formal academic blazer (N,n)
# with white collar, holding a giant glowing Gold & Cyan Laser Protractor/Set-Square triangle (G,X)!
BOSS_PROFESSOR_36: List[str] = [
    "....................................",  # 00
    "..............OOOOOOOO..............",  # 01 - Hair bun top
    ".............OVVVVVVVvO.............",  # 02 - Strict dark-plum bun
    ".............OVVVVVVVvO.............",  # 03 - Hair bun
    "...........OOOVVVVVVVvOOO...........",  # 04 - Crown outline
    ".........OOVVVVVVVVVVVVVvOO.........",  # 05 - Neat academic hair crown
    "........OVVVVVVVVVVVVVVVVVvO........",  # 06 - Hair volume
    "........OVVvVVVVVVVVVVVVVvVO........",  # 07 - Parted bangs
    "........OVvVSSSVVVVVVSSSVvVO........",  # 08 - Forehead
    "........OVvSSOOOSSSOOOSSSvVO........",  # 09 - Sharp strict brows
    ".......OOSvSXXXXSSSSXXXXSvSOO.......",  # 10 - Cyan-rimmed glasses top frame!
    ".......OScOSXWEXSXXSXWEXSOcSO.......",  # 11 - Symmetrical lenses + nose bridge + top-left W!
    ".......OScOSXXXXSSSSXXXXSOcSO.......",  # 12 - Cyan-rimmed glasses bottom frame!
    "........OOOSccSSSSSSccSSSOO.........",  # 13 - Cheeks
    ".........OOSSSSSSmmSSSSSSOO.........",  # 14 - Stern pursed mouth
    "..........OSccccSSSSccccSO..........",  # 15 - Chin shadow
    "...........OOOOOOOOOOOOOO...........",  # 16 - Chin outline
    "...........ONNNWWWWWWNNNO......OO...",  # 17 - Slate-blue blazer collar + Protractor tip!
    "..........ONNNWWWRRWWWNNNO....OGXO..",  # 18 - White collar + Crimson tie + Laser triangle!
    ".........ONNNNWWRRRRWWNNNNO..OG3XO..",  # 19 - Blazer shoulders + Protractor edge
    ".........ONNNoNNWWWWNNNoNNO.OGXggXO.",  # 20 - Arms + Set-square hollow cutout!
    ".........OSSOONNNNNNNNOOSSOOGXgggXO.",  # 21 - Hands holding giant Laser Protractor!
    ".........OccOONNNGGNNNOOccOG3ggggXO.",  # 22 - Gold blazer buttons + Protractor
    "...........OOOPPPPPPPPPOOOOGXXXXXXO.",  # 23 - Pencil skirt + Laser Protractor base!
    "...........OOPPPPPPPPPPOO..OOOOOOOO.",  # 24 - Pencil skirt mid
    "...........OPPPpPPPPpPPpO...........",  # 25 - Skirt shadow
    "...........OPPPPPPPPPPPPO...........",  # 26 - Skirt lower
    "...........OOOOOOOOOOOOOO...........",  # 27 - Skirt outline
    "............OSccO..OSccO............",  # 28 - Legs
    "............OSccO..OSccO............",  # 29 - Legs
    ".........sssOVvvOssOVvvOsss.........",  # 30 - Dark plum heels
    "..........ssOOOOOssOOOOOss..........",  # 31 - Heels outline
    "............ssssssssssss............",  # 32 - Ground shadow
    "....................................",  # 33
    "....................................",  # 34
    "....................................",  # 35
]

# 4C. Boss 3: La Mamma (Final Boss, 36x36)
# Hair in a large bun with pink hair-curlers (K,W) on the sides, furious glowing eyes,
# white/checkered kitchen apron (W,w,F) over dress, raising the legendary Boomerang Slipper/Ciabatta (T,t,F)!
BOSS_MOM_36: List[str] = [
    "....................................",  # 00
    "..............OOOOOOOO..............",  # 01 - Large hair bun top
    ".............OHHHHHHHhO.............",  # 02 - Hair bun
    "...........OOHHHHHHHHHhOO...........",  # 03 - Crown
    "......OOOOOHHHHHHHHHHHHHhOOOOO......",  # 04 - Pink curlers top + Hair crown!
    ".....OKWWKOHHHHHHHHHHHHHhOKWWKO.....",  # 05 - Left & Right Pink Hair-Curlers (K,W)!
    ".....OKKKKOHHHHHHHHHHHHHhOKKKKO.....",  # 06 - Hair-curlers mid
    "......OOOOOHHSSHHHHHHSSHHhOOOO......",  # 07 - Curlers bottom + Hair framing
    "........OHhSSSOSSSSSSOSSShHO........",  # 08 - FURIOUS angled brows!
    "........OHoSSOOSSSSSSOOSSoHO........",  # 09 - Furious brows
    ".......OSSOSSWROSSSSWROSSOSSO..OOO..",  # 10 - Furious glowing Ruby/Gold eyes (top-left W!)
    ".......OScOSSRGOSSSSRGOSSOSSO.OT1FO.",  # 11 - Glowing eyes + Raised Boomerang Ciabatta!
    "........OOOSSbbSSSSSSbbSSSOO..OTFFO.",  # 12 - Cheeks + Ciabatta strap!
    ".........OOSSSSSOmmOSSSSSOO...OTTtO.",  # 13 - Angry shout mouth + Ciabatta sole!
    "..........OSccccSSSSccccSO....OTttO.",  # 14 - Jawline + Ciabatta base!
    "...........OOOOOOOOOOOOOO.....OSSO..",  # 15 - Chin + Raised hand gripping Ciabatta!
    "..........OFFFFWWWWWWFFFFO...OSSO...",  # 16 - House dress + White apron + Raised arm!
    ".........OFFFFFWWWWWWFFFFFO.OFFFO...",  # 17 - Dress + Apron + Raised sleeve
    "........OFFFFWWWWWWWWWWFFFFOFFFO....",  # 18 - Broad motherly shoulders + raised sleeve
    "........OFFFoWWWWFFWWWWoFFFOFFO.....",  # 19 - Checkered/heart apron pocket + arm!
    "........OSSOOWWWFFFFWWWOOOOOOO......",  # 20 - Left hand + Apron center
    "........OccOOWWWWWWWWWWOO...........",  # 21 - Apron waist
    "..........OOWWWwWWWWwWWWOO..........",  # 22 - Checkered apron skirt!
    ".........OOWWwWWWWwWWWWwWOO.........",  # 23 - Apron wide flare
    ".........OWWWWWwWWWWwWWWWWO.........",  # 24 - Checkered trim
    ".........OFFFWWWWWWWWWWFFFO.........",  # 25 - Dress hem peeking under apron
    ".........OOOOOOOOOOOOOOOOOO.........",  # 26 - Skirt outline
    "............OSccO..OSccO............",  # 27 - Legs
    "............OSccO..OSccO............",  # 28 - Legs
    ".........sssOTFFOssOTFFOsss.........",  # 29 - House slippers
    "..........ssOOOOOssOOOOOss..........",  # 30 - Slippers outline
    "............ssssssssssss............",  # 31 - Ground shadow
    "....................................",  # 32
    "....................................",  # 33
    "....................................",  # 34
    "....................................",  # 35
]

# 4D. Boss 4: L'Amante (36x36)
# Glamorous chibi 3/4 character in crimson/magenta cocktail dress, dramatic wavy dark/burgundy
# hair, dark sunglasses, holding a crimson lipstick!
BOSS_LOVER_36: List[str] = [
    "....................................",  # 00
    "...........OOOOOOOOOOOOOO...........",  # 01 - Wavy hair crown
    ".........OOHAAHHHHHHHHhhHOO.........",  # 02 - Top-left highlight + dark/burgundy hair
    "........OHAxHHkkHHHHkkHHxhhO........",  # 03 - Burgundy waves
    ".......OHHHxkkkHHHHHHkkkxhHhO.......",  # 04 - Voluminous wavy hair
    "......OHHxkHHHHHHHHHHHHHHkxhHO......",  # 05
    "......OHHxHHHHHHHHHHHHHHHHxhHO......",  # 06
    ".....OOHkHHHSSHHHHHHHHSSHHHkHOO.....",  # 07 - Parted bangs
    ".....OHxkHHSSSSSSOOSSSSSSHHkxHO.....",  # 08 - Forehead
    ".....OHxHSOOOOOOSSSSOOOOOOSHxhO.....",  # 09 - Cat-eye sunglasses top rim
    "....OOHkHSOPMPOSSSSSSOPMPOOSHkHOO...",  # 10 - Sleek dark sunglasses + subtle rim glint
    "....OHxkHSOPPPPOSSSSOPPPPOSHkxHO....",  # 11 - Dark tinted lenses + nose bridge
    ".....OHkHSOOOOOOSSSSOOOOOOSHkHO.....",  # 12 - Sunglasses bottom rim
    ".....OHkHSSbbSSSSSSSSSSbbSSHkHO.....",  # 13 - Rosy glam cheeks
    ".....OHxHSSSSSSSRRmRSSSSSSSHxHO.....",  # 14 - Bold crimson lipstick smirk
    ".....OHxkOScccccSSSScccccSOkxhO.....",  # 15 - Soft chin + wavy side locks
    ".....OOHxOOOOOOOOOOOOOOOOOOkHOO.....",  # 16 - Chin outline + cascading waves
    "......OHhO..OSSSSGGGGSSSSO....OOO...",  # 17 - Gold necklace + Lipstick tip top!
    ".......OO...OSSFWWWWWWFSSO...ORFFO..",  # 18 - Sweetheart neckline + Crimson lipstick tip
    "...........ORRRRRKKKKKRRRRO..ORRFO..",  # 19 - Crimson/magenta cocktail dress + lipstick
    "..........ORRRoRRKKKKKRRoRRO.OMWMO..",  # 20 - Arms + silver lipstick collar
    ".........OSSO.ORRRRRRRRROSSOOGGDGGO.",  # 21 - Left hand on hip + right hand + gold tube
    ".........OccO.OKKKKKKKKOSSO..OSSOO..",  # 22 - Hand + waist
    "..........OO..ORRRGGGGGRRRO...OO....",  # 23 - Gold belt trim
    ".............ORRRRRKKKRRRRRO........",  # 24 - Cocktail dress skirt
    ".............ORRRrRKKKRrRRRO........",  # 25 - Dress folds
    ".............OKKKkRRRRRkKKKO........",  # 26 - Magenta/crimson hem
    ".............OOOOOOOOOOOOOOO........",  # 27 - Dress outline
    "..............OSccO...OSccO.........",  # 28 - Legs
    "..............OSccO...OSccO.........",  # 29 - Legs
    "...........sssORrrOsssORrrOsss......",  # 30 - Crimson heels
    "............ssOOOOOsssOOOOOss.......",  # 31 - Heels outline
    "............ssssssssssssssss........",  # 32 - Ground shadow
    "....................................",  # 33
    "....................................",  # 34
    "....................................",  # 35
]

# 4E. Boss 5: La Dirigente Aziendale (36x36)
# Sharp chibi 3/4 corporate boss in tailored navy/slate suit, glasses, slick bun/bob hair,
# holding a gleaming golden briefcase!
BOSS_EXECUTIVE_36: List[str] = [
    "....................................",  # 00
    "............OOOOOOOOOOOO............",  # 01 - Slick executive bun top
    "...........OAAHHHHHHHHhhO...........",  # 02 - Tight top bun
    "...........OHHHHHHHHHHHhO...........",  # 03 - Bun base
    ".........OOOHHHHHHHHHHHHhhOO........",  # 04 - Bob crown outline
    "........OAAHHHHHHHHHHHHHHhhHO.......",  # 05 - Sleek bob volume
    "........OHHHHHHHHHHHHHHHHHhhO.......",  # 06 - Sleek side bob
    "........OHHhSSSHHHHHHHSSShHhO.......",  # 07 - Sharp parted bangs
    "........OHhSSSSSOOOOOOSSSSShO.......",  # 08 - Forehead
    "........OHoSOOOSSSSSSSSOOOSoO.......",  # 09 - Sharp executive eyebrows
    ".......OOScSMMMMOSSSSOMMMMScoOO.....",  # 10 - Rectangular silver glasses top
    ".......OScOSMWEPOSSSSOMWEPOScSO.....",  # 11 - Stern eyes behind glasses + top-left W!
    ".......OScOSMMMMOSSSSOMMMMOscSO.....",  # 12 - Glasses bottom rim
    "........OOOSccSSSSSSSSSSccSOO.......",  # 13 - Cheeks
    ".........OOSSSSSSmmmmSSSSSSOO.......",  # 14 - Stern corporate mouth
    "..........OSccccSSSSSSccccSO........",  # 15 - Jawline
    "...........OOOOOOOOOOOOOOOO.........",  # 16 - Chin outline
    "..........ONNNWWWWWWWWWWNNNO........",  # 17 - Tailored navy suit collar + crisp white shirt
    ".........ONNNNWWWRRRRWWWNNNNO.......",  # 18 - Suit shoulders + crimson power tie
    "........ONNNNNNNWWRRWWNNNNNNnO......",  # 19 - Structured corporate blazer
    "........ONNNoNNNNWRRWNNNNoNNnO......",  # 20 - Sleeves + blazer lapels
    "........OSSOONNNNGGGNNNNNOOSSO......",  # 21 - Hands + gold blazer button
    "........OccOONNNNNNNNNNNNOOTTO......",  # 22 - Waist + Briefcase handle!
    "..........OOONNNNnnnnNNNNOOGGGGO....",  # 23 - Tailored pencil skirt + Golden Briefcase top!
    "...........ONNNNNnnnnNNNNOGWWGGDO...",  # 24 - Skirt + Briefcase highlight & clasp
    "...........ONNNNnNNNNnNNNOGGOGGDO...",  # 25 - Skirt + Briefcase lock!
    "...........ONNNNNnnnnNNNNOGGGGGDO...",  # 26 - Skirt + Briefcase body
    "...........OOOOOOOOOOOOOOOGDDDDDO...",  # 27 - Skirt outline + Briefcase shadow
    "............OSccO...OSccO.OOOOOOO...",  # 28 - Legs + Briefcase bottom
    "............OSccO...OSccO...........",  # 29 - Legs
    ".........sssONnnOsssONnnOsss........",  # 30 - Tailored navy heels
    "..........ssOOOOOsssOOOOOss.........",  # 31 - Heels outline
    "............sssssssssssss...........",  # 32 - Ground shadow
    "....................................",  # 33
    "....................................",  # 34
    "....................................",  # 35
]

# 4F. Boss 6: La Rettrice (36x36)
# Regal academic boss in royal purple & gold ceremonial toga/robe, silver/white hair,
# golden medallion & ceremonial staff!
BOSS_RECTOR_36: List[str] = [
    "....................................",  # 00
    "............OOOOOOOOOOOO............",  # 01 - Golden academic coronet/laurel top
    "...........OGGGWWGGGGGGDO......OOO..",  # 02 - Coronet + Staff orb top!
    "..........OOWWWWWWWWWWWwOO....O3XGO.",  # 03 - Silver-white hair + Glowing staff orb!
    "........OOWWWWWWWWWWWWWwwWO...OGXGO.",  # 04 - Regal silver hair + Staff finial
    ".......OWWWWWWWWWWWWWWWWwwWO...OGO..",  # 05 - Hair volume + Staff collar
    ".......OWwWWWWWWWWWWWWWWwwWO...OTO..",  # 06 - Silver waves + Staff shaft
    "......OWwWWWSSSWWWWWWWSSSWwwO..OTO..",  # 07 - Parted silver bangs
    "......OWwWSSSSSSWWWWWSSSSSWwO..OTO..",  # 08 - Forehead
    "......OWwSSOOOSSSWWWSSSOOOSwO..OTO..",  # 09 - Dignified silver brows
    "......OSSOSSWIOSSSSSSOIWSSwSO..OTO..",  # 10 - Commanding ruby/violet eyes (top-left W!)
    "......OScOSSIuOSSSSSSOuISSOSO..OTO..",  # 11 - Eyes mid + Staff shaft
    ".......OOOSSccSSSSSSSSccSSOO...OTO..",  # 12 - Cheeks
    "........OOSSSSSSmmmmmSSSSSOO...OTO..",  # 13 - Regal stern expression
    ".........OSccccSSSSSSSccccSO...OTO..",  # 14 - Jawline
    "..........OOOOOOOOOOOOOOOOO....OTO..",  # 15 - Chin outline
    ".........OUUUGGGWWWWWWGGGUUO...OGO..",  # 16 - Royal purple robe + golden stole + Staff ring
    "........OUUUUGGWWWRRWWWGGUUuO..OTO..",  # 17 - Gold stole + crimson academic tie
    ".......OUUUUUGGWWRGGRWWGGUUuuO.OTO..",  # 18 - Robe shoulders + Gold Medallion!
    "......OUUUUoUUGGWRRRRWGGUUoUUuOOTO..",  # 19 - Wide ceremonial sleeves + Medallion
    "......OSSOOUUUGGGGGGGGGGUUOOSSOGDO..",  # 20 - Hands gripping Golden Staff!
    "......OccOOUUUUUUUGGUUUUUUOOccOOTO..",  # 21 - Stole drape + Staff
    "........OOUUUUUUUGDDGUUUUUUOO..OTO..",  # 22 - Ceremonial robe waist
    "........OUUUUUUUGGDDGGUUUUUuO..OTO..",  # 23 - Flowing royal purple toga
    ".......OUUUUUuuUGGDDGGUuuUUUuO.OTO..",  # 24 - Robe pleats + golden center stole
    ".......OUUUUuuuuGGDDGGuuuuUUuO.OTO..",  # 25 - Lower ceremonial robe
    ".......OGGGGGGGGGGGGGGGGGGGGGO.OTO..",  # 26 - Ornate golden robe hem!
    ".......OOOOOOOOOOOOOOOOOOOOOOO.OTO..",  # 27 - Robe outline
    "...........OSccO.....OSccO.....OTO..",  # 28 - Shoes peek
    "...........OSccO.....OSccO.....OTO..",  # 29
    "........sssODddOsssssODddOsss..OGO..",  # 30 - Gold-trimmed ceremonial shoes + Staff tip
    ".........ssOOOOOsssssOOOOOss...OOO..",  # 31 - Shoes & Staff bottom outline
    "..........ssssssssssssssssss........",  # 32 - Ground shadow
    "....................................",  # 33
    "....................................",  # 34
    "....................................",  # 35
]


# ==============================================================================
# 5. PROJECTILES (16x16 each)
# ==============================================================================

# 5A. Spinning purple/cyan razor cross (U,Z,X,W with O outline)
BULLET_RAZOR_16: List[str] = [
    "................",  # 00
    "......OOOO......",  # 01
    ".....OXXXXO.....",  # 02
    ".....OZXWZO.....",  # 03
    "....OOZWWZOO....",  # 04
    ".OOOOZUUUUZOOOO.",  # 05
    "OXXXZUUWWUUZXXXO",  # 06
    "OXWWUWWRRWWUWWXO",  # 07
    "OXWWUWWXXWWUWWXO",  # 08
    "OXXXZUUWWUUZXXXO",  # 09
    ".OOOOZUUUUZOOOO.",  # 10
    "....OOZWWZOO....",  # 11
    ".....OZXWZO.....",  # 12
    ".....OXXXXO.....",  # 13
    "......OOOO......",  # 14
    "................",  # 15
]

# 5B. Golden/Cyan geometric laser triangle (G,X,W,O)
BULLET_PROTRACTOR_16: List[str] = [
    "................",  # 00
    ".......OO.......",  # 01
    "......OGXO......",  # 02
    ".....OG3XXO.....",  # 03
    ".....OGWXWO.....",  # 04
    "....OGWOOXWO....",  # 05
    "....OGWOOWXO....",  # 06
    "...OGWOOOOWXO...",  # 07
    "...OGWOOOOXWO...",  # 08
    "..OGWOOOOOOWXO..",  # 09
    "..OGWOOOOOOWXO..",  # 10
    ".OGWOOOOOOOOXWO.",  # 11
    ".OGGWWWWWWWWXXO.",  # 12
    ".OGGGGGGGGGXXXO.",  # 13
    ".OOOOOOOOOOOOOO.",  # 14
    "................",  # 15
]

# 5C. Cute Kawaii anime sweatdrop/teardrop (X,3,W,K,O)
BULLET_TEAR_16: List[str] = [
    "................",  # 00
    ".......OO.......",  # 01
    "......O3XO......",  # 02
    ".....O3XX2O.....",  # 03
    "....O3WXXX2O....",  # 04
    "....OWWXXX2O....",  # 05
    "...O3WXXXXX2O...",  # 06
    "...O3XXXXXX2O...",  # 07
    "...OXXXXKKX2O...",  # 08
    "...OXXXKKKK2O...",  # 09
    "...O2XXXKKX2O...",  # 10
    "....O2XXXXX2O...",  # 11
    "....O22XXX22O...",  # 12
    ".....OO222OO....",  # 13
    ".......OOO......",  # 14
    "................",  # 15
]

# 5D. Gyaru ring-light star/heart projectile (K,G,W,O)
BULLET_FLASH_16: List[str] = [
    "................",  # 00
    "......OOOO......",  # 01
    "....OOGGGGOO....",  # 02
    "...OGGWWWWGGO...",  # 03
    "..OGWKKKKKKWGO..",  # 04
    "..OGWKWWWWKWGO..",  # 05
    ".OGWKKWKKWKKWGO.",  # 06
    ".OGWKWWKKWWKWGO.",  # 07
    ".OGWKKWWWWKKWGO.",  # 08
    ".OGWKKWKKWKKWGO.",  # 09
    "..OGWKWWWWKWGO..",  # 10
    "..OGWKKKKKKWGO..",  # 11
    "...OGGWWWWGGO...",  # 12
    "....OOGGGGOO....",  # 13
    "......OOOO......",  # 14
    "................",  # 15
]

# 5E. Red/Blue flashing police siren orb (R,X,W,O)
BULLET_POLICE_16: List[str] = [
    "................",  # 00
    ".....OOOOOO.....",  # 01
    "...OORRRXXXOO...",  # 02
    "..ORRRRWWXXXXO..",  # 03
    "..ORRRWWWWXXXO..",  # 04
    ".ORRRWWXXWWXXXO.",  # 05
    ".ORRRWWRRWWXX2O.",  # 06
    ".ORRRWWWWWWXX2O.",  # 07
    ".ORrrWWWWWWX22O.",  # 08
    ".ORrrrWWWWXX22O.",  # 09
    "..OrrrWWWW222O..",  # 10
    "..OrrrrWW2222O..",  # 11
    "...OOrrr222OO...",  # 12
    ".....OOOOOO.....",  # 13
    "................",  # 14
    "................",  # 15
]

# 5F. La Mamma's giant spinning boomerang slipper / ciabatta (T,t,F,W,O)
BULLET_CIABATTA_16: List[str] = [
    "................",  # 00
    ".........OOOO...",  # 01
    "........OT11TO..",  # 02
    ".......OT1TTtO..",  # 03
    "......OFFFFFtO..",  # 04
    ".....OFWWFFftO..",  # 05
    "....OFFWWFFtO...",  # 06
    "...OT1FFFFtO....",  # 07
    "..OT1TTTTtO.....",  # 08
    "..OTTTTTtO......",  # 09
    "..OTTTTtO.......",  # 10
    "..OTtttO........",  # 11
    "...OOOO.........",  # 12
    "................",  # 13
    "................",  # 14
    "................",  # 15
]

# 5G. L'Amante's crimson-magenta lipstick projectile (R,F,K,G,W,O)
BULLET_LIPSTICK_16: List[str] = [
    "................",  # 00
    ".........OOOO...",  # 01
    "........ORFFRO..",  # 02
    ".......ORWFFRRO.",  # 03
    "......ORWFFRKKO.",  # 04
    ".....ORFFRRKKO..",  # 05
    "....ORFFRRKKO...",  # 06
    "....OOGGGGOO....",  # 07
    "...OGGWWWWGO....",  # 08
    "...OGGGGGDDO....",  # 09
    "..OMWWWWMMO.....",  # 10
    "..OMMMMddO......",  # 11
    "..OMMMddO.......",  # 12
    "..OddddO........",  # 13
    "...OOOO.........",  # 14
    "................",  # 15
]

# 5H. La Dirigente's red-tape corporate contract projectile (W,C,w,R,F,d,O)
BULLET_CONTRACT_16: List[str] = [
    "................",  # 00
    "..OOOOOOOOOO....",  # 01
    "..OWWWWWWWWwO...",  # 02
    "..OWddddddWWwO..",  # 03
    "..OWWWWWWWWWwO..",  # 04
    "..OWdddWWddWWwO.",  # 05
    "..ORRRRRRRRRRRO.",  # 06
    "..OFFFFFFFFFFFO.",  # 07
    "..OWdddWWdddWwO.",  # 08
    "..OWWWWWWWWWWwO.",  # 09
    "..OWddddWWGGOwO.",  # 10
    "..OWWWWWWGWROwO.",  # 11
    "..OWCCCCCGGGGwO.",  # 12
    "..OwwwwwwwwwwwO.",  # 13
    "..OOOOOOOOOOOOO.",  # 14
    "................",  # 15
]

# 5I. La Rettrice's gold/crimson academic wax seal orb projectile (G,D,R,F,W,O)
BULLET_SEAL_16: List[str] = [
    "................",  # 00
    ".....OOOOOO.....",  # 01
    "...OOGGGGGGOO...",  # 02
    "..OGGWWWWWWGGO..",  # 03
    "..OGWRRRRRRWGO..",  # 04
    ".OGWRFFFFFFRWGO.",  # 05
    ".OGWRFFWWFFRWGO.",  # 06
    ".OGWRFFFFFFRrDO.",  # 07
    ".OGWRRRRRRRRrDO.",  # 08
    "..OGDRRRRRRDDO..",  # 09
    "..OOGDDDDDDGOO..",  # 10
    "...OORROOORRO...",  # 11
    "..ORRRO..ORRRO..",  # 12
    "..ORrO....OrRO..",  # 13
    "...OO......OO...",  # 14
    "................",  # 15
]


# ==============================================================================
# 6. STREET TILES (32x32 each)
# ==============================================================================

def build_tile_asphalt_32() -> Canvas:
    """Warm slate-purple asphalt (#524E61) with subtle tileable pixel pebbles."""
    grid = ["4" * 32 for _ in range(32)]
    c = grid_to_canvas(grid, 32, 32, WORLD_PALETTE)
    dark = WORLD_PALETTE["5"]
    light = WORLD_PALETTE["6"]
    pebbles_dark = [(3, 4), (11, 2), (24, 6), (7, 14), (19, 12), (29, 15), (4, 23), (15, 21), (26, 26), (12, 29)]
    pebbles_light = [(6, 6), (17, 3), (27, 9), (2, 17), (14, 15), (22, 19), (9, 25), (20, 28), (30, 22)]
    for px, py in pebbles_dark:
        c.set_pixel(px, py, dark)
    for px, py in pebbles_light:
        c.set_pixel(px, py, light)
    return c


def build_tile_sidewalk_32() -> Canvas:
    """Warm stone sidewalk (#D8CEBE) with 16x16 tile grooves (#BCB2A2)."""
    c = Canvas(32, 32, WORLD_PALETTE["7"])
    groove = WORLD_PALETTE["8"]
    hi = WORLD_PALETTE["9"]
    # 16x16 grooves
    for i in range(32):
        c.set_pixel(i, 0, groove)
        c.set_pixel(i, 16, groove)
        c.set_pixel(0, i, groove)
        c.set_pixel(16, i, groove)
    # Subtle top-left bevel inside each 16x16 slab
    for bx in (1, 17):
        for by in (1, 17):
            for k in range(14):
                c.set_pixel(bx + k, by, hi)
                c.set_pixel(bx, by + k, hi)
    return c


def build_tile_ztl_32() -> Canvas:
    """Forbidden ZTL asphalt zone with diagonal Red/Yellow hazard border stripes (R,G) & ZTL stamp."""
    c = build_tile_asphalt_32()
    red = WORLD_PALETTE["R"]
    gold = WORLD_PALETTE["G"]
    outline = WORLD_PALETTE["O"]
    white = WORLD_PALETTE["W"]

    # 3px diagonal Red/Yellow hazard border around all 4 edges
    for y in range(32):
        for x in range(32):
            if x < 3 or x >= 29 or y < 3 or y >= 29:
                c.set_pixel(x, y, red if ((x + y) // 3) % 2 == 0 else gold)
            elif x == 3 or x == 28 or y == 3 or y == 28:
                c.set_pixel(x, y, outline)

    # Centered 'ZTL' stamp (5px tall, 15px wide -> x=8..23, y=13..18)
    ztl_stamp = [
        "GGGG.GGGG.G...",
        "...G..GG..G...",
        "..G...GG..G...",
        ".G....GG..G...",
        "GGGG..GG..GGGG",
    ]
    # Add dark backing shadow & crisp red/gold stamp
    ox, oy = 9, 13
    for ry, row in enumerate(ztl_stamp):
        for rx, ch in enumerate(row):
            if ch == "G":
                c.set_pixel(ox + rx + 1, oy + ry + 1, outline)
    for ry, row in enumerate(ztl_stamp):
        for rx, ch in enumerate(row):
            if ch == "G":
                c.set_pixel(ox + rx, oy + ry, gold if ry < 3 else red)
                if ry == 0:
                    c.set_pixel(ox + rx, oy + ry, white)
    return c


def build_tile_crosswalk_32() -> Canvas:
    """Asphalt with crisp white zebra crossing stripes (W,w)."""
    c = build_tile_asphalt_32()
    white = WORLD_PALETTE["W"]
    shadow = WORLD_PALETTE["w"]
    outline = WORLD_PALETTE["O"]
    # Two crisp vertical zebra bars across the 32x32 tile (seamless vertically)
    for bx in (4, 19):
        c.fill_rect(bx, 2, 9, 28, white)
        c.fill_rect(bx + 8, 2, 1, 28, shadow)
        c.fill_rect(bx, 29, 9, 1, shadow)
        # Subtle curb/paint edge
        c.fill_rect(bx, 1, 9, 1, outline)
        c.fill_rect(bx, 30, 9, 1, outline)
    return c


# ==============================================================================
# MAIN ASSET GENERATION & VERIFICATION
# ==============================================================================

def main() -> None:
    os.makedirs(OUT_DIR, exist_ok=True)

    # 1. Bus Stops (48x48) & Standalone Vending Shop (32x32)
    stops_48 = {
        "stop_kawaii": STOP_KAWAII_48,
        "stop_emo": STOP_EMO_48,
        "stop_student": STOP_STUDENT_48,
        "stop_gyaru": STOP_GYARU_48,
    }
    for name, grid in stops_48.items():
        canvas = grid_to_canvas(grid, 48, 48, PALETTE)
        save_asset_and_preview(canvas, OUT_DIR, name, zoom=6)

    shop_canvas = grid_to_canvas(SHOP_VENDING_32, 32, 32, PALETTE)
    save_asset_and_preview(shop_canvas, OUT_DIR, "shop_vending", zoom=8)

    # 2. The Neon Bus (96x48)
    bus_canvas = grid_to_canvas(BUS_NEON_96, 96, 48, PALETTE)
    save_asset_and_preview(bus_canvas, OUT_DIR, "bus_neon", zoom=5)

    # 3. Police Units (32x32 & 48x32)
    officer_canvas = grid_to_canvas(POLICE_OFFICER_32, 32, 32, PALETTE)
    save_asset_and_preview(officer_canvas, OUT_DIR, "police_officer", zoom=8)

    car_canvas = grid_to_canvas(POLICE_CAR_48, 48, 32, PALETTE)
    save_asset_and_preview(car_canvas, OUT_DIR, "police_car", zoom=6)

    # 4. All 6 Iconic Boss Gatekeepers (36x36)
    bosses_36 = {
        "boss_best_friend": BOSS_BEST_FRIEND_36,
        "boss_professor": BOSS_PROFESSOR_36,
        "boss_mom": BOSS_MOM_36,
        "boss_lover": BOSS_LOVER_36,
        "boss_executive": BOSS_EXECUTIVE_36,
        "boss_rector": BOSS_RECTOR_36,
    }
    for name, grid in bosses_36.items():
        canvas = grid_to_canvas(grid, 36, 36, PALETTE)
        save_asset_and_preview(canvas, OUT_DIR, name, zoom=8)

    # 5. Projectiles (16x16)
    bullets_16 = {
        "bullet_razor": BULLET_RAZOR_16,
        "bullet_protractor": BULLET_PROTRACTOR_16,
        "bullet_tear": BULLET_TEAR_16,
        "bullet_flash": BULLET_FLASH_16,
        "bullet_police": BULLET_POLICE_16,
        "bullet_ciabatta": BULLET_CIABATTA_16,
        "bullet_lipstick": BULLET_LIPSTICK_16,
        "bullet_contract": BULLET_CONTRACT_16,
        "bullet_seal": BULLET_SEAL_16,
    }
    for name, grid in bullets_16.items():
        canvas = grid_to_canvas(grid, 16, 16, PALETTE)
        save_asset_and_preview(canvas, OUT_DIR, name, zoom=10)

    # 6. Street Tiles (32x32)
    tiles_32 = {
        "tile_asphalt": build_tile_asphalt_32(),
        "tile_sidewalk": build_tile_sidewalk_32(),
        "tile_ztl": build_tile_ztl_32(),
        "tile_crosswalk": build_tile_crosswalk_32(),
    }
    for name, canvas in tiles_32.items():
        save_asset_and_preview(canvas, OUT_DIR, name, zoom=6)

    total = len(stops_48) + 1 + 1 + 2 + len(bosses_36) + len(bullets_16) + len(tiles_32)
    print(f"Successfully generated all {total} world sprites (+ previews in tools/previews/world) in {OUT_DIR}")


if __name__ == "__main__":
    main()
