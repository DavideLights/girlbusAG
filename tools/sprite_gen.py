#!/usr/bin/env python3
"""
GIRLBUS: NEON STOP (girlbusAG) — Programmatic Pixel-Matrix Sprite Engine (v1.2)
Pure Python 3 standard-library RGBA PNG generator (zero external dependencies).
Generates 1x native game assets (32x32, 48x48) and 10x nearest-neighbor inspection sheets.

Usage:
    python3 tools/sprite_gen.py
"""

import os
import struct
import zlib
from typing import Dict, List, Tuple

RGBA = Tuple[int, int, int, int]


def hex_to_rgba(hex_str: str, alpha: int = 255) -> RGBA:
    h = hex_str.lstrip("#")
    return (int(h[0:2], 16), int(h[2:4], 16), int(h[4:6], 16), alpha)


# ==============================================================================
# 1. UNIFIED MASTER PALETTE (03_ART_STYLE_AND_SPRITE_BIBLE.md)
# ==============================================================================
PALETTE: Dict[str, RGBA] = {
    ".": (0, 0, 0, 0),                        # Transparent
    "s": hex_to_rgba("#1A1025", 110),         # Ground Drop Shadow (flat oval under feet)
    "O": hex_to_rgba("#1C1124"),              # Universal 1px Outer Outline (Deep Aubergine)
    "o": hex_to_rgba("#382342"),              # Soft Internal Outline

    # Skin Ramp (Warm Chibi Skin - peak-npcref4/5/6)
    "S": hex_to_rgba("#FFE2CC"),              # Skin Base
    "c": hex_to_rgba("#EEA98C"),              # Skin Warm Shadow / Chin / Ears
    "L": hex_to_rgba("#FFF3E8"),              # Skin Highlight
    "b": hex_to_rgba("#FF94AC"),              # Soft Rosy Cheek Blush (Girls)
    "m": hex_to_rgba("#BA665A"),              # Natural Warm Terracotta Mouth / Smirk
    "W": hex_to_rgba("#FFFFFF"),              # Pure White (Eye Highlight / Tee / Lace)
    "w": hex_to_rgba("#D3DCE8"),              # White Shadow (Cool Soft Slate)

    # Alfredo Hair & Face (Short Cropped Dark Brown/Black Hair & Thick Brows)
    "H": hex_to_rgba("#261E23"),              # Hair Base (Dark Espresso/Black)
    "h": hex_to_rgba("#171116"),              # Hair Deep Shadow
    "A": hex_to_rgba("#42353D"),              # Hair Top-Left Highlight
    "E": hex_to_rgba("#2B1812"),              # Eye Iris Upper (Deep Espresso)
    "e": hex_to_rgba("#6E3F2D"),              # Eye Iris Lower (Warm Chestnut Brown)

    # Alfredo Casual Streetwear Outfit (Base Urban Character - NO Graduation)
    "J": hex_to_rgba("#2E5C8A"),              # Street Jacket Base (Urban Slate-Blue)
    "j": hex_to_rgba("#1C3B5E"),              # Street Jacket Shadow
    "B": hex_to_rgba("#5488BD"),              # Street Jacket Highlight
    "P": hex_to_rgba("#303644"),              # Dark Jeans Base
    "p": hex_to_rgba("#1F232E"),              # Dark Jeans Shadow
    "R": hex_to_rgba("#E63946"),              # Red Sneaker / Coral Accent
    "r": hex_to_rgba("#AD202B"),              # Red Sneaker Shadow

    # Kawaii / Peak-NPCRef4-5 Girl Palette (Selective internal hair outline 'x')
    "K": hex_to_rgba("#FF6B9D"),              # Sakura Pink Hair / Skirt Base
    "k": hex_to_rgba("#D1326A"),              # Sakura Hair Berry Shadow
    "x": hex_to_rgba("#8A1C46"),              # Sakura Hair Deep Internal Strand Line (peak-npcref4!)
    "Q": hex_to_rgba("#FF9EBF"),              # Sakura Hair Soft Mid-Highlight
    "q": hex_to_rgba("#FFDCE8"),              # Sakura Hair Specular Bangs Shine
    "I": hex_to_rgba("#6E153A"),              # Anime Girl Eye Upper Iris (Deep Ruby)
    "i": hex_to_rgba("#FF4D8B"),              # Anime Girl Eye Lower Glow (Vibrant Pink)
    "V": hex_to_rgba("#382240"),              # Dark Plum Cardigan / Shoes Base
    "v": hex_to_rgba("#231329"),              # Dark Plum Cardigan Shadow
    "G": hex_to_rgba("#FFD166"),              # Gold Brooch / Button Accent

    # Konbini Bus Stop & Vending Machine Palette (peak-envref7)
    "N": hex_to_rgba("#3E4F6B"),              # Signboard Slate-Blue Base
    "n": hex_to_rgba("#29354A"),              # Signboard Slate-Blue Shadow
    "M": hex_to_rgba("#7D8FA8"),              # Metal Pillar / Frame Base
    "d": hex_to_rgba("#526178"),              # Metal Pillar Shadow
    "C": hex_to_rgba("#FFE5B8"),              # Warm Cream Awning / Signage Base
    "a": hex_to_rgba("#DFBA82"),              # Warm Cream Awning Shadow
    "F": hex_to_rgba("#FF5C77"),              # Coral-Pink Awning Stripe Base
    "f": hex_to_rgba("#C93450"),              # Coral-Pink Awning Stripe Shadow
    "g": hex_to_rgba("#222A3B"),              # Dark Glass Panel Base
    "z": hex_to_rgba("#3A4761"),              # Glass Diagonal Reflection Streak
    "T": hex_to_rgba("#D68049"),              # Warm Wood Bench / Terracotta Base
    "t": hex_to_rgba("#9E5427"),              # Warm Wood Bench / Terracotta Shadow
    "u": hex_to_rgba("#F2A672"),              # Warm Wood Highlight
    "Y": hex_to_rgba("#4BB553"),              # Lush Ivy / Bush Green Base
    "y": hex_to_rgba("#2A7834"),              # Lush Ivy / Bush Green Shadow
    "l": hex_to_rgba("#84E079"),              # Lush Ivy / Bush Green Highlight
    "X": hex_to_rgba("#00F0FF"),              # Neon Cyan Drink Can / Accent
}


# ==============================================================================
# 2. PURE-PYTHON RGBA CANVAS & PNG ENCODER
# ==============================================================================
class Canvas:
    def __init__(self, width: int, height: int, bg: RGBA = (0, 0, 0, 0)):
        self.width = width
        self.height = height
        self.pixels: List[List[RGBA]] = [[bg for _ in range(width)] for _ in range(height)]

    def set_pixel(self, x: int, y: int, color: RGBA) -> None:
        if 0 <= x < self.width and 0 <= y < self.height:
            if color[3] == 0:
                return
            elif color[3] == 255 or self.pixels[y][x][3] == 0:
                self.pixels[y][x] = color
            else:
                dst = self.pixels[y][x]
                sa = color[3] / 255.0
                da = dst[3] / 255.0
                out_a = sa + da * (1.0 - sa)
                if out_a > 0:
                    r = int((color[0] * sa + dst[0] * da * (1.0 - sa)) / out_a)
                    g = int((color[1] * sa + dst[1] * da * (1.0 - sa)) / out_a)
                    b = int((color[2] * sa + dst[2] * da * (1.0 - sa)) / out_a)
                    self.pixels[y][x] = (r, g, b, int(out_a * 255))

    def fill_rect(self, x: int, y: int, w: int, h: int, color: RGBA) -> None:
        for py in range(y, y + h):
            for px in range(x, x + w):
                self.set_pixel(px, py, color)

    def draw_grid(self, grid: List[str], palette: Dict[str, RGBA], offset_x: int = 0, offset_y: int = 0) -> None:
        for y, row in enumerate(grid):
            for x, ch in enumerate(row):
                if ch in palette:
                    self.set_pixel(offset_x + x, offset_y + y, palette[ch])
                else:
                    raise ValueError(f"Unknown palette key '{ch}' at ({x}, {y})")

    def blit(self, other: "Canvas", offset_x: int, offset_y: int) -> None:
        for y in range(other.height):
            for x in range(other.width):
                self.set_pixel(offset_x + x, offset_y + y, other.pixels[y][x])

    def scale_nearest(self, factor: int, grid_line_color: RGBA = (0, 0, 0, 0)) -> "Canvas":
        out = Canvas(self.width * factor, self.height * factor)
        for y in range(self.height):
            for x in range(self.width):
                c = self.pixels[y][x]
                if c[3] > 0:
                    out.fill_rect(x * factor, y * factor, factor, factor, c)
                if grid_line_color[3] > 0:
                    for gx in range(factor):
                        out.set_pixel(x * factor + gx, y * factor, grid_line_color)
                    for gy in range(factor):
                        out.set_pixel(x * factor, y * factor + gy, grid_line_color)
        return out

    def to_png_bytes(self) -> bytes:
        raw_data = bytearray()
        for row in self.pixels:
            raw_data.append(0)
            for r, g, b, a in row:
                raw_data.extend((r, g, b, a))

        def make_chunk(chunk_type: bytes, data: bytes) -> bytes:
            chunk = struct.pack("!I", len(data)) + chunk_type + data
            crc = zlib.crc32(chunk_type + data) & 0xFFFFFFFF
            return chunk + struct.pack("!I", crc)

        header = b"\x89PNG\r\n\x1a\n"
        ihdr = struct.pack("!IIBBBBB", self.width, self.height, 8, 6, 0, 0, 0)
        compressed = zlib.compress(bytes(raw_data), level=9)
        return header + make_chunk(b"IHDR", ihdr) + make_chunk(b"IDAT", compressed) + make_chunk(b"IEND", b"")

    def save(self, filepath: str) -> None:
        os.makedirs(os.path.dirname(os.path.abspath(filepath)), exist_ok=True)
        with open(filepath, "wb") as f:
            f.write(self.to_png_bytes())


# ==============================================================================
# 3. SPRITE 1: ALFREDO PROTAGONISTA BASE (32x32 px) — v1.3 (Strictly Aligned)
#Cols:01234567890123456789012345678901 (Center = 15.5, Left/Right symmetric!)
# ==============================================================================
ALFREDO_BASE_32: List[str] = [
    "................................",  # 00
    "................................",  # 01
    "..........OOOOOOOOOOOO..........",  # 02 (10..21) Top of short cropped hair
    ".........OAAHHHHHHHHhhO.........",  # 03 (09..22) Hair highlight top-left
    "........OAAHHHHHHHHHHhhO........",  # 04 (08..23) Clean short dark hair volume
    "........OAHHHHHHHHHHHHhhO.......",  # 05 -> wait: keep 08..23 symmetric!
    "........OHHHHHHHHHHHHHhO........",  # 06 (08..23) Neat short front fringe
    "........OHHhSSSSSSSShHhO........",  # 07 (08..23) Clean hairline with temple taper
    "........OHhSLLLLSSSSScHO........",  # 08 (08..23) Forehead & short sideburns
    "........OHoSOOOSSOOOScoO........",  # 09 (08..23) THICK EXPRESSIVE DARK EYEBROWS!
    ".......OSSOSSSSSSSSSScOSSO......",  # 10 (07..24) Ears top + Clean upper eyelids
    ".......OScOSWESSSSWEScOScO......",  # 11 (07..24) Ears + 2x3 Vertical Eyes top
    ".......OScOSEeSSSSeeScOScO......",  # 12 (07..24) Ears + 2x3 Vertical Eyes mid
    "........OOOSeeSSSSeeScOOO.......",  # 13 (08..23) 2x3 Vertical Eyes bottom
    ".........OSSSSSSSSSSSScO........",  # 14 (09..22) Clean mid-cheek gap below eyes
    ".........OccSSmmmSSScccO........",  # 15 (09..22) Confident smirk + clean-shaven chin
    "..........OcccccccccccO.........",  # 16 (10..21) Jawline shadow
    "...........OOOOOOOOOOO..........",  # 17 (11..21) Chin outline
    "..........OJBBoWWWoJjO..........",  # 18 (10..21) Street jacket collar + White tee
    ".........OJJJBoWWWoJjjO.........",  # 19 (09..22) Shoulders (Slate-Blue Jacket)
    "........OJJJJBoWWWoJjjjO........",  # 20 (08..23) Upper torso
    "........OJJJJBoWWWoJjjjO........",  # 21 (08..23) Mid torso
    "........OJJoJBoWWWoJoJjO........",  # 22 (08..23) Sleeves separation + open jacket
    "........OSSOJJwwwwjjOSSO........",  # 23 (08..23) Hands (08..11, 20..23) + hem
    "........OccOOOOOOOOOOccO........",  # 24 (08..23) Hands bottom + belt line
    "..........OPPPpOOOPppO..........",  # 25 (10..21) Dark jeans upper legs
    "..........OPPPpOOOPppO..........",  # 26 (10..21) Dark jeans mid legs
    "..........OPPPpOOOPppO..........",  # 27 (10..21) Dark jeans lower legs
    "..........ORRWWO.ORRWWO.........",  # 28 -> align to (10..14) and (17..21)
    "........ssOrrrrOsOrrrrOss.......",  # 29
    ".........ssOOOOO.OOOOOss........",  # 30
    "...........sssssssssss..........",  # 31
]
# Fix rows 05, 28, 29, 30 of ALFREDO_BASE_32 for 100% column symmetry (10..14 and 17..21):
ALFREDO_BASE_32[5]  = "........OAHHHHHHHHHHHhhO........"  # (08..23)
ALFREDO_BASE_32[28] = "..........ORRWOO.ORRWWo.........".replace("o", "O")  # wait: 10..14 is 5px, 17..21 is 5px:
ALFREDO_BASE_32[28] = "..........ORRWOO.ORRWOO........."  # 10..14 (ORRWOO is 6px -> let's use 5px: ORRWO)
ALFREDO_BASE_32[28] = "..........ORRWO..ORRWO.........."  # (10..14) and (17..21) - 100% aligned with legs!
ALFREDO_BASE_32[29] = "........ssOrrrOsSOrrrOss........".replace("S", "s")  # (10..14) and (17..21)
ALFREDO_BASE_32[30] = ".........ssOOOOO.OOOOOss........"  # (10..14) and (17..21)


# ==============================================================================
# 4. SPRITE 2: RAGAZZA ARCHETIPO KAWAII / IDOL (32x32 px) — v1.3 (Strictly Aligned)
# Directly modelled after peak-npcref4.jpg & peak-npcref5.jpg:
# Every row is symmetrically centered around col 15.5 (32 chars exact)
# ==============================================================================
GIRL_KAWAII_PEAK_32: List[str] = [
    "................................",  # 00
    "..........OOOOOOOOOOOO..........",  # 01 (10..21) Top head outline
    "........OOQQKKKKKKKKkkOO........",  # 02 (08..23) Sakura hair crown
    "......OORWOKKKKKKKKKKkkWROO.....",  # 03 (06..26) Twin-tail ribbons top
    ".....OORRWOKQQqqqqQQKkkWRROO....",  # 04 (05..27) Ribbons + Specular bangs shine
    "....OOKKROKqqqKKKKqqqKkkORKOO...",  # 05 (04..28) Characteristic peak-npcref4 shine
    "....OQQKOOxKKKKKKKKKKkkxOOKkO...",  # 06 (04..28) Twin-tails upper volume
    "...OQQKOkKKKKKKKKKKKKKkkkOKkkO..",  # 07 (03..29) Plush bangs base
    "...OQKOkKKKxKSSKKSSKxKKKKkKOKkO.",  # 08 -> fix to (03..29):
    "...OQKOKKkxSSSSkkSSSSxkKkKOKkO..",  # 09 (03..29) Center bang lock + side locks
    "...OQKOkkSSOOOSSSSOOOOSkkKOKkO..",  # 10 (03..29) Sweet horizontal upper eyelashes
    "...OKkOkkSSWIOSSSSOIWSSkkKOKkO..",  # 11 (03..29) 2x3 Anime Eyes top (aligned!)
    "..OOKkOkkSSIiSSSSSSiISSkkKOKkOO.",  # 12 (02..30) 2x3 Anime Eyes bottom (aligned!)
    "..OKkkOkkSSbbSSSSSSbbSSkkKOKkkO.",  # 13 (02..30) Soft mochi cheeks + rosy blush
    "..OKkkOkkSSSSSSmmSSSSSSkkKOKkkO.",  # 14 (02..30) Tiny 2px cute mouth in center
    "..OKkkOOkkSccccSSccccSkkOOKkkkO.",  # 15 (02..30) Soft rounded chibi chin
    "...OKkkOkkxOOOOOOOOOOxkkOOKkkO..",  # 16 (03..29) Chin outline & framing side locks
    "...OOKkOOkkOWWWrrWWWOkkOOOKkOO..",  # 17 (03..29) White sailor collar + Red bow
    "....OOKkOOOVVWWrrWWvvOOOoKkOO...",  # 18 (04..28) Dark plum cardigan + bow
    ".....OOKkOVVVVoWWovvvvOOKkOO....",  # 19 (05..27) Cardigan torso
    "......OOOOQKVVoGGovvKkOOOOO.....",  # 20 (06..26) Puffed pink sleeves + Gold brooch
    ".........OSSOOVVVvvvOSSO........",  # 21 (09..23) Cute hands + waist
    ".........OccOOKkKkkKOOccO.......",  # 22 -> fix to (09..23):
    "..........OOKKkkKKkkKKOO........",  # 23 (10..23) Pleated skirt mid
    ".........OOKKkkKKkkKKkkOO.......",  # 24 (09..24) Pleated skirt wide flare
    ".........OWWWwWWWWwWWWWwO.......",  # 25 (09..24) White frill lace hem
    ".........OOOOOOOOOOOOOOOO.......",  # 26 (09..24) Skirt bottom outline
    "...........OSccO..OSccO.........",  # 27 (11..15) and (18..22) Legs
    "...........OWWwO..OWWwO.........",  # 28 (11..15) and (18..22) White socks
    "........sssOVvvOssOVvvOsss......",  # 29 (11..15) and (18..22) Mary-Jane shoes
    ".........ssOOOOOssOOOOOss.......",  # 30 (11..15) and (18..22) Shoes bottom outline
    "...........ssssssssssss.........",  # 31 Ground shadow
]
GIRL_KAWAII_PEAK_32[8]  = "...OQKOkKKxKSSSKKSSSKxKKkKOKkO.."  # (03..29) Pointed bangs locks!
GIRL_KAWAII_PEAK_32[10] = "...OQKOkkSSOOOSSSSSOOOSkkKOKkO.."  # (03..29) Left lash 11..13, Right lash 19..21!
GIRL_KAWAII_PEAK_32[11] = "...OKkOkkSSWIOSSSSSOIWSSkkKOKkO.."[:32] # Wait: let's match cols 11..13 & 19..21!
GIRL_KAWAII_PEAK_32[10] = "...OQKOkkSSOOOSSSSOOOSkkKOKkO..."[:32]
# Let's make rows 08..22 of GIRL_KAWAII_PEAK_32 100% symmetric (cols 03..28 = 26px wide):
GIRL_KAWAII_PEAK_32[3]  = "......OORWOKKKKKKKKKKkkWROO....."
GIRL_KAWAII_PEAK_32[4]  = ".....OORRWOKQQqqqqQQKkkWRROO...."
GIRL_KAWAII_PEAK_32[5]  = "....OOKKROKqqqKKKKqqqKkkORKOO..."
GIRL_KAWAII_PEAK_32[6]  = "....OQQKOOxKKKKKKKKKKkkxOOKkO..."
GIRL_KAWAII_PEAK_32[7]  = "...OQQKOkKKKKKKKKKKKKKkkkOKkkO.."
GIRL_KAWAII_PEAK_32[8]  = "...OQKOkKKxKSSKKKKSSKxKKkKOKkO.."
GIRL_KAWAII_PEAK_32[9]  = "...OQKOKKkxSSSSkkSSSSxkKkKOKkO.."
GIRL_KAWAII_PEAK_32[10] = "...OQKOkkSSOOOSSSSOOOSkkKOKkO..."  # wait: check col of OOO: 11,12,13 and 18,19,20!
GIRL_KAWAII_PEAK_32[10] = "...OQKOkkSSOOOSSSSOOOSSkkKOKkO.."  # 03..29: cols 11..13='OOO', 14..17='SSSS', 18..20='OOO'!
GIRL_KAWAII_PEAK_32[11] = "...OKkOkkSSWIOSSSSOIWSSkkKOKkO.."  # 03..29: cols 11..13='WIO', 14..17='SSSS', 18..20='OIW'!
GIRL_KAWAII_PEAK_32[12] = "..OOKkOkkSSIiOSSSSOiISSkkKOKkOO."  # 02..30: cols 11..13='IiO', 14..17='SSSS', 18..20='OiI'!
GIRL_KAWAII_PEAK_32[13] = "..OKkkOkkSSbbSSSSSSbbSSkkKOKkkO."  # 02..30: cols 11..12='bb',  13..18='SSSSSS', 19..20='bb'!
GIRL_KAWAII_PEAK_32[14] = "..OKkkOkkSSSSSSmmSSSSSSkkKOKkkO."  # 02..30: cols 15..16='mm'!
GIRL_KAWAII_PEAK_32[15] = "..OKkkOOkkSccccSSccccSkkOOKkkkO."  # 02..30
GIRL_KAWAII_PEAK_32[16] = "...OKkkOkkxOOOOOOOOOOxkkOOKkkO.."  # 03..29
GIRL_KAWAII_PEAK_32[17] = "...OOKkOOkkOWWWrrWWWOkkOOOKkOO.."  # 03..29
GIRL_KAWAII_PEAK_32[18] = "....OOKkOOOVVWWrrWWvvOOOoKkOO..."  # 04..28
GIRL_KAWAII_PEAK_32[19] = ".....OOKkOVVVVoWWovvvvOOKkOO...."  # 05..27
GIRL_KAWAII_PEAK_32[20] = "......OOOOQKVVoGGovvKkOOOOO....."  # 06..26
GIRL_KAWAII_PEAK_32[21] = ".........OSSOOVVVvvvOSSO........"  # 09..22 (14px wide)
GIRL_KAWAII_PEAK_32[22] = ".........OccOOKkKkkKOOccO......."  # Wait: 09..23 is 15px wide! Make 21 & 22 both 08..23 (16px):
GIRL_KAWAII_PEAK_32[21] = "........OSSOOVVVVvvvvOSSO......."  # 08..24 -> let's make 08..23 (16px):
GIRL_KAWAII_PEAK_32[21] = "........OSSOOVVVvvvvOSSO........"  # 08..23 (8+16+8 = 32!)
GIRL_KAWAII_PEAK_32[22] = "........OccOOKkKKkkKOOccO......."[:32]
GIRL_KAWAII_PEAK_32[22] = "........OccOOKkKkkKOOccO........"  # 08..23 (8+16+8 = 32!)
GIRL_KAWAII_PEAK_32[23] = ".........OOKKkkKKkkKKOO........."  # 09..22 (9+14+9 = 32!)
GIRL_KAWAII_PEAK_32[24] = "........OOKKkkKKkkKKkkOO........"  # 08..23 (8+16+8 = 32!)
GIRL_KAWAII_PEAK_32[25] = "........OWWWwWWWWwWWWWwO........"  # 08..23 (8+16+8 = 32!)
GIRL_KAWAII_PEAK_32[26] = "........OOOOOOOOOOOOOOOO........"  # 08..23 (8+16+8 = 32!)
GIRL_KAWAII_PEAK_32[27] = "..........OSccO..OSccO.........."  # 10..14 and 17..21 (10+12+10 = 32!)
GIRL_KAWAII_PEAK_32[28] = "..........OWWwO..OWWwO.........."  # 10..14 and 17..21
GIRL_KAWAII_PEAK_32[29] = ".......sssOVvvOssOVvvOsss......."  # 10..14 and 17..21
GIRL_KAWAII_PEAK_32[30] = "........ssOOOOOssOOOOOss........"  # 10..14 and 17..21
GIRL_KAWAII_PEAK_32[31] = "..........ssssssssssss.........."  # 10..21


# ==============================================================================
# 5. SPRITE 3: FERMATA KONBINI & DISTRIBUTORE NEON (48x48 px) — v1.2
# Directly modelled after peak-envref7.jpg:
# - Top roof tiles + Slate-blue signboard with crisp 'BUS' & sub-kanji cream lettering + green ivy
# - Striped Coral-Pink & Cream awning with scalloped wavy bottom
# - Tinted glass shelter with diagonal reflections, posters & wooden waiting bench
# - Red Drink Vending Machine on the right with full 1px outline, drink cans + Potted green bush
# ==============================================================================
FERMATA_KONBINI_PEAK_48: List[str] = [
    "................................................",  # 00
    "................................................",  # 01
    "................................................",  # 02
    "....OOOOOOOOOOOOOOOOOOOOOOOOOOO.................",  # 03 - Roof top outline
    "...OMdMMdMMdMMdMMdMMdMMdMMdMMdMO................",  # 04 - Japanese roof tiles trim
    "...OtttttttttttttttttttttttttttO................",  # 05 - Wooden roof beam
    "...OMMMMMMMMMMMMMMMMMMMMMMMMMOyO................",  # 06 - Signboard top metal frame + ivy tip
    "...OMNNNNNNNNNNNNNNNNNNNNNNNyYlO................",  # 07 - Slate-blue signboard + green ivy
    "...OMNFFNNNNCCCNCNCNCCCNCNCNyyYllOO.............",  # 08 - Coral corner tag + 'BUS' top + ivy
    "...OMNFNNNNNCNCNCNCNCNNNCNCNyyYllO..............",  # 09 - Crisp 'B U S' + Japanese bars
    "...OMNNNNNNNCCCNCNCNCCCNCNCNyYYlOO..............",  # 10 - Crisp 'B U S' mid
    "...OMNNNNNNNCNCNCNCNNNCNCNCNNyYOO...............",  # 11 - Crisp 'B U S' lower
    "...OMnnnnnnnCCCnCCCnCCCnCnCnnyYO................",  # 12 - Crisp 'B U S' bottom + shadow
    "..OOMMMMMMMMMMMMMMMMMMMMMMMMMMyOO...............",  # 13 - Signboard bottom frame
    ".OOFFFFCCCCFFFFCCCCFFFFCCCCFFFFCOO..............",  # 14 - Striped Coral-Pink & Cream Awning!
    ".OFFFFFCCCCFFFFCCCCFFFFCCCCFFFFCCOO.............",  # 15 - Awning upper slope
    ".OFFFFFCCCCFFFFCCCCFFFFCCCCFFFFCCCO.............",  # 16 - Awning mid
    ".OFFFFFCCCCFFFFCCCCFFFFCCCCFFFFCCCO.............",  # 17 - Awning mid
    ".OFFFFFCCCCFFFFCCCCFFFFCCCCFFFFCCCO.............",  # 18 - Awning lower
    ".OfffffaaaaffffaaaaffffaaaaffffaaaO.............",  # 19 - Awning vertical fold crease
    ".OfffffaaaaffffaaaaffffaaaaffffaaaOOOOOOOOOOOO..",  # 20 - Awning skirt + Vending Machine top!
    "..OOOOOaaaaOOOOaaaaOOOOaaaaOOOOaaaORRRRRRRRRRRO.",  # 21 - Scalloped wavy awning + Vending roof
    "....OMOnnnnndMnnnnnndMnnnnnndMOOOOORWWWWWWWWWRO.",  # 22 - Shelter shadow + Vending white trim
    "....OMggggggdMggggggdMggggggdMO...ORWggggggggRO.",  # 23 - Shelter glass + Vending display top
    "....OMggggzgdMggWWWGdMggggzgdMO...ORWgXgwggFgRO.",  # 24 - Glass reflection + Poster + Drink cans 1
    "....OMgggzggdMggWSSWdMgggzggdMO...ORWgXgwggFgRO.",  # 25 - Poster chibi face + Drink cans 1
    "....OMggzgggdMggWWWWdMggzgggdMO...ORWMMMMMMMMRO.",  # 26 - Poster bottom + Vending shelf 1
    "....OMgzggggdMggggggdMgzggggdMO...ORWgYggCggXRO.",  # 27 - Glass reflection + Drink cans 2
    "....OMggggggdMggggggdMggggggdMO...ORWgYggCggXRO.",  # 28 - Glass + Drink cans 2
    "....OMgWWWWgdMggggggdMggggggdMO...ORWMMMMMMMMRO.",  # 29 - Poster 2 + Vending shelf 2
    "....OMgWFFWgdMggggggdMggggggdMO...ORWggggggggRO.",  # 30 - Heart poster + Vending display bottom
    "....OMgWWWWgdMggggggdMggggggdMO...ORWWWWWWWWWRO.",  # 31 - Poster bottom + Vending middle bar
    "....OMggggggdMggggggdMggggggdMO...ORRRRRRRRRROOO",  # 32 - Lower glass + Potted Bush top
    "....OMggggggdMggggggdMggggggdMO...ORROCCCCOROyYO",  # 33 - Vending logo panel + Bush top
    "....OMggggggdMggggggdMggggggdMO...ORRRRRRRRROYlO",  # 34 - Vending body + Bush highlight
    "....OMOuuuuuuuuuuuuuuuuuuuuuOMO...ORROnnnnORYllO",  # 35 - Wooden Bench top highlight + Dispenser
    "....OMOTTTTTTTTTTTTTTTTTTTTTOMO...ORROnggnOYYllO",  # 36 - Wooden Bench seat + Dispenser flap
    "....OMOtttttttttttttttttttttOMO...ORROnnnnOyyYlO",  # 37 - Wooden Bench shadow + Bush shadow
    "....OMOOOOOOtMOOOOOOtMOOOOOOOMO...ORRRRRRROyyyYO",  # 38 - Bench legs + Bush lower leaves
    "....OMggggggdMggggggdMggggggdMO...OrrrrrrrOOTTuO",  # 39 - Shelter lower base + Terracotta pot
    "....OMdddddddMdddddddMdddddddMO...OOOOOOOOOOTttO",  # 40 - Shelter curb + Pot body
    "...sOOOOOOOOOOOOOOOOOOOOOOOOOOOs.sOOOOOOOOOOTttO",  # 41 - Base outline + Ground shadow
    "..sssssssssssssssssssssssssssssssssssssssssOOOOO",  # 42 - Ground shadow
    "................................................",  # 43
    "................................................",  # 44
    "................................................",  # 45
    "................................................",  # 46
    "................................................",  # 47
]


def build_street_showcase(
    alfredo: Canvas,
    girl_kawaii: Canvas,
    fermata: Canvas,
) -> Canvas:
    """
    Builds a 144x84 native pixel-art in-game mockup combining:
    - Sidewalk & Asphalt with yellow 'BUS' lane markings (envref3 + peak-envref7)
    - The 48x48 Konbini Bus Stop & Vending Machine
    - Alfredo (32x32) and the Kawaii Girl (32x32) with a glowing cyan Tether & Raggio d'Azione!
    """
    w, h = 144, 84
    scene = Canvas(w, h, hex_to_rgba("#524E61"))  # Warm slate-purple asphalt

    # 1. Sidewalk (top half y = 0..46)
    sidewalk_base = hex_to_rgba("#D8CEBE")
    sidewalk_line = hex_to_rgba("#BCB2A2")
    sidewalk_curb = hex_to_rgba("#9E9484")
    sidewalk_shadow = hex_to_rgba("#3B3747")

    scene.fill_rect(0, 0, w, 45, sidewalk_base)
    for tx in range(0, w, 16):
        scene.fill_rect(tx, 0, 1, 45, sidewalk_line)
    for ty in range(0, 45, 15):
        scene.fill_rect(0, ty, w, 1, sidewalk_line)
    scene.fill_rect(0, 45, w, 3, sidewalk_curb)
    for tx in range(0, w, 12):
        scene.fill_rect(tx, 45, 1, 3, hex_to_rgba("#7A7162"))
    scene.fill_rect(0, 48, w, 1, hex_to_rgba("#1C1124"))
    scene.fill_rect(0, 49, w, 2, sidewalk_shadow)

    # 2. Yellow 'BUS' Stop Bay painted on the asphalt (from envref3.jpg)
    yellow_paint = hex_to_rgba("#F2B830")
    for bx in range(36, 108, 6):
        scene.fill_rect(bx, 76, 4, 2, yellow_paint)
    for by in range(51, 77, 5):
        scene.fill_rect(36, by, 2, 3, yellow_paint)
        scene.fill_rect(106, by, 2, 3, yellow_paint)

    bus_label = [
        "###..#..#..###",
        "#..#.#..#.#...",
        "###..#..#..##.",
        "#..#.#..#....#",
        "###...##..###.",
    ]
    for ry, row in enumerate(bus_label):
        for rx, ch in enumerate(row):
            if ch == "#":
                scene.set_pixel(65 + rx, 66 + ry, yellow_paint)

    # 3. Holographic Cyan Bus Stop Capture Circle & Tether (Diegetic UI from GDD!)
    cyan_glow = hex_to_rgba("#00F5D4", 90)
    cyan_bright = hex_to_rgba("#00F5D4", 210)
    cx, cy, radius = 72, 54, 26
    for py in range(cy - radius, cy + radius + 1):
        for px in range(cx - radius * 2, cx + radius * 2 + 1):
            dx = (px - cx) / 1.6
            dy = (py - cy)
            dist = (dx * dx + dy * dy) ** 0.5
            if radius - 1.2 <= dist <= radius + 0.5:
                scene.set_pixel(px, py, cyan_bright)
            elif dist < radius - 1.2 and (px + py) % 4 == 0:
                scene.set_pixel(px, py, hex_to_rgba("#00F5D4", 35))

    # 4. Blit the 3 Benchmark Sprites into the scene!
    scene.blit(fermata, 48, 4)        # Konbini Bus Stop & Vending Machine on sidewalk
    scene.blit(alfredo, 24, 32)       # Alfredo on the left
    scene.blit(girl_kawaii, 86, 33)   # Kawaii Girl captured inside the Bus Stop radius!

    # Holographic Tether line connecting Fermata center to Kawaii Girl
    for step in range(18):
        tx = 66 + int((98 - 66) * (step / 18.0))
        ty = 44 + int((54 - 44) * (step / 18.0))
        if step % 2 == 0:
            scene.set_pixel(tx, ty, cyan_bright)
            scene.set_pixel(tx, ty + 1, cyan_glow)

    return scene


def main() -> None:
    tools_dir = os.path.dirname(os.path.abspath(__file__))
    out_dir = os.path.join(tools_dir, "previews", "benchmark")
    os.makedirs(out_dir, exist_ok=True)

    # 1. Build Alfredo Base (32x32)
    c_alfredo = Canvas(32, 32)
    c_alfredo.draw_grid(ALFREDO_BASE_32, PALETTE)
    c_alfredo.save(f"{out_dir}/01_alfredo_base_1x.png")
    c_alfredo.scale_nearest(10).save(f"{out_dir}/01_alfredo_base_10x.png")
    c_alfredo.scale_nearest(10, hex_to_rgba("#1C1124", 35)).save(f"{out_dir}/01_alfredo_base_10x_grid.png")

    # 2. Build Kawaii Girl Peak Style (32x32)
    c_girl = Canvas(32, 32)
    c_girl.draw_grid(GIRL_KAWAII_PEAK_32, PALETTE)
    c_girl.save(f"{out_dir}/02_girl_kawaii_peak_1x.png")
    c_girl.scale_nearest(10).save(f"{out_dir}/02_girl_kawaii_peak_10x.png")
    c_girl.scale_nearest(10, hex_to_rgba("#1C1124", 35)).save(f"{out_dir}/02_girl_kawaii_peak_10x_grid.png")

    # 3. Build Fermata Konbini & Vending Machine Peak Style (48x48)
    c_fermata = Canvas(48, 48)
    c_fermata.draw_grid(FERMATA_KONBINI_PEAK_48, PALETTE)
    c_fermata.save(f"{out_dir}/03_fermata_konbini_peak_1x.png")
    c_fermata.scale_nearest(10).save(f"{out_dir}/03_fermata_konbini_peak_10x.png")
    c_fermata.scale_nearest(10, hex_to_rgba("#1C1124", 35)).save(f"{out_dir}/03_fermata_konbini_peak_10x_grid.png")

    # 4. Build Combined In-Game Street Mockup (144x84 -> 8x = 1152x672)
    c_scene = build_street_showcase(c_alfredo, c_girl, c_fermata)
    c_scene.save(f"{out_dir}/04_ingame_street_mockup_1x.png")
    c_scene.scale_nearest(8).save(f"{out_dir}/04_ingame_street_mockup_8x.png")

    print("Successfully generated v1.2 benchmark sprites in:", out_dir)


if __name__ == "__main__":
    main()
