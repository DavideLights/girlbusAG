#!/usr/bin/env python3
"""
GIRLBUS: NEON STOP (girlbusAG) — Layered Urban Environment & Map Pixel-Art Generator.

Generates 1x native PNG sprites in assets/sprites/map/ and previews in tools/previews/map/.

References:
  - peak-envref7.jpg (Konbini storefront, coral/cream awning, red vending, ivy AC, black cat)
  - envref5.jpg      (Coffee Daily facade, forest-green/white awning, blue vending, cats)
  - peak-envref4.jpg (Pop-urban rooftop terraces, colorful cars, boardwalk, canal water)
  - envref6.png      (Pixelade Arcade neon signage, built-in ATM)
  - peak-npcref6.jpg (Lush potted plants, cozy park props)
  - envref3.jpg      (Top-down 3/4 parked cars)
"""

import os
import sys
from typing import Dict, List, Optional, Tuple

TOOLS_DIR = os.path.dirname(os.path.abspath(__file__))
ROOT_DIR = os.path.dirname(TOOLS_DIR)
if TOOLS_DIR not in sys.path:
    sys.path.insert(0, TOOLS_DIR)

from pixel_canvas import Canvas, PALETTE, RGBA, hex_to_rgba, save_asset_and_preview

OUT_DIR = os.path.join(ROOT_DIR, "assets", "sprites", "map")

# ==============================================================================
# EXTENDED ENVIRONMENT PALETTE (Harmonized with Master PALETTE & #1C1124 Outline)
# ==============================================================================
MAP_PALETTE: Dict[str, RGBA] = {
    **PALETTE,
    # Deep Forest Green (Coffee Daily signboard & awning)
    "fg_base": hex_to_rgba("#1E5C34"),
    "fg_dark": hex_to_rgba("#134023"),
    "fg_hi": hex_to_rgba("#2E7D49"),
    # Mint / Pop-Teal (Arcade & Mint Car)
    "mint_hi": hex_to_rgba("#7CF5D6"),
    "mint_base": hex_to_rgba("#38CFA8"),
    "mint_sh": hex_to_rgba("#1F9478"),
    # Deep Interior Shadow
    "int_dark": hex_to_rgba("#171C29"),
    # Terracotta / Brick accents
    "brick_base": hex_to_rgba("#B85A3A"),
    "brick_sh": hex_to_rgba("#873C24"),
    "brick_hi": hex_to_rgba("#D67A56"),
    # Dark Foliage underside
    "leaf_deep": hex_to_rgba("#1D5728"),
    # Ground Tiles
    "grass_base": hex_to_rgba("#58A854"),
    "grass_sh": hex_to_rgba("#448E42"),
    "grass_hi": hex_to_rgba("#78C86C"),
    "cobble_base": hex_to_rgba("#C9BFA8"),
    "cobble_sh": hex_to_rgba("#B5A895"),
    "cobble_hi": hex_to_rgba("#DED5C5"),
    "wood_base": hex_to_rgba("#B87344"),
    "wood_sh": hex_to_rgba("#965930"),
    "wood_hi": hex_to_rgba("#CC8858"),
    "wood_nail": hex_to_rgba("#5E3419"),
    "water_base": hex_to_rgba("#3A7EC2"),
    "water_sh": hex_to_rgba("#2A629E"),
    "water_hi": hex_to_rgba("#68B0F0"),
    "water_glint": hex_to_rgba("#A8DCFF"),
    # Dark Slate Roof & Pop-Urban Bookstore / Ramen / Metro / Asphalt
    "roof_base": hex_to_rgba("#2D344A"),
    "roof_hi": hex_to_rgba("#465170"),
    "roof_sh": hex_to_rgba("#1E2233"),
    "teal_base": hex_to_rgba("#238C82"),
    "teal_hi": hex_to_rgba("#42BAAD"),
    "teal_sh": hex_to_rgba("#165E57"),
    "mustard_base": hex_to_rgba("#E6AC38"),
    "mustard_hi": hex_to_rgba("#FFD166"),
    "mustard_sh": hex_to_rgba("#B57C19"),
    "lantern_glow": hex_to_rgba("#FFD166", 70),
    "asphalt_base": hex_to_rgba("#524E61"),
    "asphalt_dark": hex_to_rgba("#464254"),
    "asphalt_light": hex_to_rgba("#605C70"),
    "asphalt_deep": hex_to_rgba("#3A3647"),
}

P = MAP_PALETTE  # Shorthand


# ==============================================================================
# PROCEDURAL PIXEL-ART HELPERS
# ==============================================================================

FONT_3X5: Dict[str, List[str]] = {
    "A": ["010", "101", "111", "101", "101"],
    "B": ["110", "101", "110", "101", "110"],
    "C": ["011", "100", "100", "100", "011"],
    "D": ["110", "101", "101", "101", "110"],
    "E": ["111", "100", "110", "100", "111"],
    "F": ["111", "100", "110", "100", "100"],
    "G": ["011", "100", "101", "101", "011"],
    "H": ["101", "101", "111", "101", "101"],
    "I": ["111", "010", "010", "010", "111"],
    "K": ["101", "101", "110", "101", "101"],
    "L": ["100", "100", "100", "100", "111"],
    "M": ["101", "111", "111", "101", "101"],
    "N": ["101", "111", "111", "111", "101"],
    "O": ["010", "101", "101", "101", "010"],
    "P": ["110", "101", "110", "100", "100"],
    "R": ["110", "101", "110", "101", "101"],
    "S": ["011", "100", "010", "001", "110"],
    "T": ["111", "010", "010", "010", "010"],
    "U": ["101", "101", "101", "101", "111"],
    "V": ["101", "101", "101", "101", "010"],
    "W": ["101", "101", "111", "111", "101"],
    "X": ["101", "101", "010", "101", "101"],
    "Y": ["101", "101", "010", "010", "010"],
    "Z": ["111", "001", "010", "100", "111"],
    "7": ["111", "001", "010", "010", "010"],
    "&": ["010", "101", "010", "101", "011"],
    " ": ["000", "000", "000", "000", "000"],
}


def draw_text(
    c: Canvas,
    x: int,
    y: int,
    text: str,
    color: RGBA,
    shadow: Optional[RGBA] = None,
    spacing: int = 1,
) -> None:
    cx = x
    for ch in text.upper():
        glyph = FONT_3X5.get(ch, FONT_3X5[" "])
        if ch == " ":
            cx += 2
            continue
        for gy, row in enumerate(glyph):
            for gx, bit in enumerate(row):
                if bit == "1":
                    if shadow is not None:
                        c.set_pixel(cx + gx, y + gy + 1, shadow)
                    c.set_pixel(cx + gx, y + gy, color)
        cx += 3 + spacing


def draw_box(
    c: Canvas,
    x: int,
    y: int,
    w: int,
    h: int,
    fill: RGBA,
    outline: Optional[RGBA] = None,
    hi: Optional[RGBA] = None,
    sh: Optional[RGBA] = None,
) -> None:
    """Draw a pixel-clean rectangle with optional 1px outline, top-left highlight & bottom-right shadow."""
    if outline is not None:
        c.fill_rect(x, y, w, h, outline)
        ix, iy, iw, ih = x + 1, y + 1, max(0, w - 2), max(0, h - 2)
    else:
        ix, iy, iw, ih = x, y, w, h
    if iw <= 0 or ih <= 0:
        return
    c.fill_rect(ix, iy, iw, ih, fill)
    if hi is not None:
        c.fill_rect(ix, iy, iw, 1, hi)
        c.fill_rect(ix, iy, 1, ih, hi)
    if sh is not None:
        c.fill_rect(ix, iy + ih - 1, iw, 1, sh)
        c.fill_rect(ix + iw - 1, iy, 1, ih, sh)


def fill_ellipse(c: Canvas, cx: float, cy: float, rx: float, ry: float, color: RGBA) -> None:
    x0, x1 = int(cx - rx - 1), int(cx + rx + 2)
    y0, y1 = int(cy - ry - 1), int(cy + ry + 2)
    for py in range(y0, y1):
        for px in range(x0, x1):
            dx = (px + 0.5 - cx) / rx
            dy = (py + 0.5 - cy) / ry
            if dx * dx + dy * dy <= 1.0:
                c.set_pixel(px, py, color)


def apply_1px_outline(c: Canvas, outline_color: RGBA = P["O"]) -> None:
    """Add a clean 4-connected 1px outline around all opaque pixels on a temporary canvas."""
    opaque = [
        [c.pixels[y][x][3] >= 200 for x in range(c.width)]
        for y in range(c.height)
    ]
    for y in range(c.height):
        for x in range(c.width):
            if not opaque[y][x] and c.pixels[y][x][3] == 0:
                for nx, ny in ((x - 1, y), (x + 1, y), (x, y - 1), (x, y + 1)):
                    if 0 <= nx < c.width and 0 <= ny < c.height and opaque[ny][nx]:
                        c.set_pixel(x, y, outline_color)
                        break


def draw_leaf_cluster(
    c: Canvas,
    blobs: List[Tuple[float, float, float, float]],
    base_col: RGBA = P["Y"],
    hi_col: RGBA = P["l"],
    sh_col: RGBA = P["y"],
    deep_col: Optional[RGBA] = None,
    outline: RGBA = P["O"],
) -> None:
    """Draw a cel-shaded, outlined multi-blob foliage cluster onto canvas `c`."""
    tmp = Canvas(c.width, c.height)
    for cx, cy, rx, ry in blobs:
        fill_ellipse(tmp, cx, cy, rx, ry, base_col)
    # Shadow underside on each blob
    for cx, cy, rx, ry in blobs:
        x0, x1 = int(cx - rx), int(cx + rx + 1)
        y0, y1 = int(cy - ry), int(cy + ry + 1)
        for py in range(y0, y1):
            for px in range(x0, x1):
                if tmp.pixels[py][px][3] > 0:
                    dx = (px + 0.5 - cx) / rx
                    dy = (py + 0.5 - cy) / ry
                    if dx * dx + dy * dy <= 1.0:
                        if dy > 0.35 or (dx > 0.45 and dy > 0.0):
                            tmp.set_pixel(px, py, sh_col)
                        if deep_col and dy > 0.68:
                            tmp.set_pixel(px, py, deep_col)
    # Top-left highlight cluster
    for cx, cy, rx, ry in blobs:
        fill_ellipse(tmp, cx - rx * 0.25, cy - ry * 0.32, max(1.2, rx * 0.48), max(1.0, ry * 0.38), hi_col)
    apply_1px_outline(tmp, outline)
    c.blit(tmp, 0, 0)


def draw_ac_unit(c: Canvas, x: int, y: int, rusty: bool = False) -> None:
    """Draw an outdoor split AC unit (~15x10) with fan grille, louvers & mounting bracket."""
    # Mounting shelf brackets
    c.fill_rect(x + 2, y + 9, 2, 2, P["d"])
    c.fill_rect(x + 11, y + 9, 2, 2, P["d"])
    draw_box(c, x, y, 15, 9, P["W"], P["O"], hi=P["W"], sh=P["w"])
    c.fill_rect(x + 1, y + 7, 13, 1, P["M"])
    # Left louvers / vents
    c.fill_rect(x + 2, y + 2, 3, 1, P["d"])
    c.fill_rect(x + 2, y + 4, 3, 1, P["d"])
    if rusty:
        c.set_pixel(x + 2, y + 6, P["T"])
        c.set_pixel(x + 12, y + 6, P["T"])
        c.set_pixel(x + 13, y + 2, P["t"])
    # Right circular fan grille
    draw_box(c, x + 7, y + 1, 6, 6, P["g"], P["d"])
    c.set_pixel(x + 9, y + 3, P["M"])
    c.set_pixel(x + 10, y + 3, P["M"])
    c.set_pixel(x + 9, y + 4, P["M"])
    c.set_pixel(x + 10, y + 4, P["M"])
    c.set_pixel(x + 8, y + 2, P["z"])
    c.set_pixel(x + 11, y + 5, P["z"])


def draw_vending_machine(c: Canvas, x: int, y: int, body_col: RGBA, sh_col: RGBA, hi_col: RGBA) -> None:
    """Draw a 13x18 detailed pixel vending machine with lit cans, coin slot & dispenser."""
    draw_box(c, x, y, 13, 18, body_col, P["O"], hi=hi_col, sh=sh_col)
    # Lit glass drink showcase
    draw_box(c, x + 2, y + 2, 9, 8, P["g"], P["W"])
    # Top row cans
    cans_top = [P["R"], P["X"], P["G"], P["F"], P["W"], P["l"], P["B"]]
    for i, col in enumerate(cans_top):
        c.set_pixel(x + 3 + i, y + 3, col)
        c.set_pixel(x + 3 + i, y + 4, col)
    # Shelf bar
    c.fill_rect(x + 3, y + 5, 7, 1, P["M"])
    # Bottom row cans
    cans_bot = [P["K"], P["G"], P["W"], P["X"], P["R"], P["l"], P["C"]]
    for i, col in enumerate(cans_bot):
        c.set_pixel(x + 3 + i, y + 6, col)
        c.set_pixel(x + 3 + i, y + 7, col)
    # Price bar & glowing blue/white panel
    c.fill_rect(x + 2, y + 11, 5, 2, P["W"])
    c.fill_rect(x + 3, y + 11, 3, 1, P["X"])
    # Coin slot
    c.fill_rect(x + 9, y + 11, 2, 2, P["O"])
    c.set_pixel(x + 9, y + 11, P["M"])
    # Bottom dispenser flap
    draw_box(c, x + 2, y + 14, 9, 3, P["g"], P["O"], sh=P["p"])
    c.fill_rect(x + 3, y + 14, 7, 1, P["d"])


def draw_striped_awning(
    c: Canvas,
    x: int,
    y: int,
    w: int,
    h: int,
    col_a: RGBA,
    sh_a: RGBA,
    col_b: RGBA,
    sh_b: RGBA,
    stripe_w: int = 5,
) -> None:
    """Draw a scalloped striped storefront awning with top highlight, curve shadow & drop shadow."""
    # Top outline
    c.fill_rect(x + 1, y, w - 2, 1, P["O"])
    for py in range(1, h - 1):
        row_x = x if py >= 2 else x + 1
        row_w = w if py >= 2 else w - 2
        c.set_pixel(row_x, y + py, P["O"])
        c.set_pixel(row_x + row_w - 1, y + py, P["O"])
        for px in range(1, row_w - 1):
            abs_x = row_x + px
            stripe_idx = ((abs_x - x) // stripe_w) % 2
            if py == 1:
                c.set_pixel(abs_x, y + py, col_a if stripe_idx == 0 else col_b)
            elif py >= h - 3:
                c.set_pixel(abs_x, y + py, sh_a if stripe_idx == 0 else sh_b)
            else:
                c.set_pixel(abs_x, y + py, col_a if stripe_idx == 0 else col_b)
    # Scalloped bottom outline & shadow
    for px in range(1, w - 1):
        abs_x = x + px
        local = (abs_x - x) % stripe_w
        scallop_y = y + h - 1 if 1 <= local <= stripe_w - 2 else y + h - 2
        c.set_pixel(abs_x, scallop_y, P["O"])
        c.set_pixel(abs_x, scallop_y + 1, P["s"])


def draw_glass_window(c: Canvas, x: int, y: int, w: int, h: int, frame_col: RGBA, frame_sh: RGBA) -> None:
    """Draw a dark-tinted storefront glass window with diagonal reflections."""
    draw_box(c, x, y, w, h, P["g"], P["O"], hi=frame_col, sh=frame_sh)
    c.fill_rect(x + 1, y + 1, w - 2, h - 2, frame_col)
    c.fill_rect(x + 2, y + 2, w - 4, h - 4, P["g"])
    # Diagonal reflection streaks
    for i in range(min(w - 5, h - 5)):
        c.set_pixel(x + 3 + i, y + h - 4 - i, P["z"])
        if x + 5 + i < x + w - 2:
            c.set_pixel(x + 5 + i, y + h - 4 - i, P["z"])
    if w > 12:
        for i in range(min(4, h - 5)):
            c.set_pixel(x + w - 7 + i, y + h - 4 - i, P["z"])


def draw_terracotta_pot(c: Canvas, x: int, y: int, plant_blobs: List[Tuple[float, float, float, float]]) -> None:
    """Draw a cute terracotta pot (7x6) with a leafy plant cluster above it."""
    draw_leaf_cluster(c, plant_blobs)
    # Rim
    draw_box(c, x, y, 7, 3, P["T"], P["O"], hi=P["1"], sh=P["t"])
    # Pot body tapered
    c.fill_rect(x + 1, y + 3, 5, 3, P["O"])
    c.fill_rect(x + 1, y + 3, 4, 2, P["T"])
    c.set_pixel(x + 4, y + 3, P["t"])
    c.set_pixel(x + 4, y + 4, P["t"])
    c.fill_rect(x + 2, y + 5, 3, 1, P["O"])


# ==============================================================================
# 1. THREE DETAILED STOREFRONT BUILDINGS (96x64 each)
# ==============================================================================

def build_konbini_96x64() -> Canvas:
    """
    building_konbini.png (96x64) — Directly modelled after peak-envref7.jpg:
    - Upper floor: Japanese tiled roof, wooden rafters, slate-blue wall, brown shutter,
      AC unit wrapped in ivy, warm lit window, potted balcony greenery.
    - Mid signboard: Wide slate-blue Japanese signboard with cream glyphs & coral tag.
    - Striped awning: Coral-Pink & Warm Cream scalloped vertical stripes.
    - Storefront: Sliding glass doors w/ OPEN sign, chibi poster, pink blob mascot,
      cardboard boxes w/ black cat on left, red vending machine, nobori flag & bush on right.
    """
    c = Canvas(96, 64)

    # Ground drop shadow along bottom
    fill_ellipse(c, 47, 62.5, 45, 2.0, P["s"])

    # --- UPPER FLOOR (y=2..24, x=11..80) ---
    # Japanese roof tiles (y=2..5)
    c.fill_rect(10, 2, 72, 4, P["O"])
    for rx in range(11, 81, 3):
        c.fill_rect(rx, 3, 2, 2, P["M"])
        c.set_pixel(rx + 1, 4, P["d"])
        c.set_pixel(rx + 2, 3, P["d"])
    # Wooden eaves beam & vertical brackets (y=5..8)
    draw_box(c, 10, 5, 72, 3, P["T"], P["O"], hi=P["1"], sh=P["t"])
    for bx in range(14, 80, 9):
        c.fill_rect(bx, 5, 2, 4, P["t"])
        c.set_pixel(bx, 5, P["1"])
        c.set_pixel(bx, 8, P["O"])

    # Upper slate-blue wall (x=12..79, y=8..24)
    draw_box(c, 12, 8, 68, 16, P["N"], P["O"], hi=P["M"], sh=P["n"])
    # Subtle siding lines
    for sy in (12, 16, 20):
        c.fill_rect(13, sy, 66, 1, P["n"])

    # Left slatted brown shutter window (x=15..38, y=10..22)
    draw_box(c, 15, 10, 23, 12, P["t"], P["O"], hi=P["T"], sh=P["int_dark"])
    for sy in range(12, 21, 2):
        c.fill_rect(17, sy, 19, 1, P["T"])
        c.fill_rect(17, sy + 1, 19, 1, P["int_dark"])

    # Center outdoor AC unit (x=42..57, y=11..20)
    draw_ac_unit(c, 42, 11)

    # Right warm lit window (x=61..71, y=10..21)
    draw_box(c, 61, 10, 10, 11, P["C"], P["O"], hi=P["W"], sh=P["a"])
    c.fill_rect(62, 11, 8, 2, P["M"])
    c.fill_rect(62, 13, 8, 6, P["C"])
    c.fill_rect(62, 16, 8, 2, P["F"])
    c.fill_rect(62, 18, 8, 2, P["f"])

    # Ivy vines on upper wall & around AC unit
    draw_leaf_cluster(c, [(12, 11, 2.2, 3.2), (12, 15, 1.8, 2.5)])
    draw_leaf_cluster(c, [(56, 13, 2.5, 2.8), (57, 18, 2.5, 3.0), (54, 21, 2.0, 2.0)])
    draw_leaf_cluster(c, [(78, 12, 2.2, 3.2), (78, 17, 1.8, 2.4)])

    # Balcony potted greenery on ledge above signboard (y=19..24)
    draw_leaf_cluster(c, [(17, 21, 3.8, 3.0), (22, 21, 3.2, 2.6)])
    draw_leaf_cluster(c, [(31, 22, 3.6, 2.6), (36, 22, 2.8, 2.2)])
    # Long tropical arching leaves on right (x=62..88, y=15..24)
    draw_leaf_cluster(
        c,
        [(66, 21, 4.5, 2.2), (74, 19, 4.5, 1.8), (81, 17, 4.0, 1.5), (78, 22, 2.2, 3.2)],
    )

    # --- MID SIGNBOARD (x=10..81, y=24..34) ---
    draw_box(c, 10, 24, 71, 11, P["N"], P["O"], hi=P["M"], sh=P["n"])
    # Inner metallic trim
    c.fill_rect(11, 25, 69, 9, P["M"])
    c.fill_rect(12, 26, 67, 7, P["N"])
    c.fill_rect(12, 32, 67, 1, P["n"])
    # Coral slanted corner tag (top-left of sign)
    c.fill_rect(13, 27, 3, 2, P["F"])
    c.set_pixel(14, 26, P["F"])
    # Crisp Cream Japanese Katakana/Kanji-style pixel glyphs (y=26..30)
    glyph_blocks = [
        (20, 4, "1111/1001/1111/1001"),
        (26, 4, "0110/1010/1110/0010"),
        (32, 4, "1111/0100/0100/1111"),
        (38, 4, "1011/1101/1001/1011"),
        (44, 4, "1110/1010/1110/1001"),
        (50, 4, "1010/1111/1010/1010"),
        (56, 4, "0111/1000/1110/0001"),
    ]
    for gx, gw, pat in glyph_blocks:
        rows = pat.split("/")
        for r_i, row in enumerate(rows):
            for c_i, bit in enumerate(row):
                if bit == "1":
                    c.set_pixel(gx + c_i, 26 + r_i, P["C"] if r_i < 3 else P["a"])
    # Subtitle dots & small sticker
    for dx in range(20, 60, 2):
        c.set_pixel(dx, 31, P["a"])
    draw_box(c, 67, 29, 6, 4, P["W"], P["N"])
    c.set_pixel(69, 30, P["F"])
    # Ivy trailing down right edge of signboard
    draw_leaf_cluster(c, [(79, 27, 2.0, 2.8), (79, 32, 2.0, 2.6)])

    # --- NOBORI VERTICAL JAPANESE BANNER FLAG (x=82..91, y=32..50) ---
    # Pole
    for i in range(22):
        c.set_pixel(83, 31 + i, P["M"] if i % 2 == 0 else P["d"])
    # Flag cloth
    c.fill_rect(84, 33, 7, 13, P["O"])
    c.fill_rect(85, 34, 5, 11, P["W"])
    c.fill_rect(89, 34, 1, 11, P["w"])
    # Vertical Japanese script marks & red stamp
    for ty in (35, 38, 41):
        c.fill_rect(86, ty, 3, 1, P["H"])
        c.set_pixel(87, ty + 1, P["H"])
    c.set_pixel(87, 43, P["R"])

    # --- STRIPED AWNING: CORAL-PINK & WARM CREAM (x=6..81, y=35..44) ---
    draw_striped_awning(c, 6, 35, 75, 10, P["F"], P["f"], P["C"], P["a"], stripe_w=5)

    # --- GROUND FLOOR STOREFRONT (x=11..64, y=45..62) ---
    # Left brick/wood wall pillar (x=11..28, y=45..62)
    draw_box(c, 11, 45, 18, 17, P["T"], P["O"], hi=P["1"], sh=P["t"])
    draw_glass_window(c, 13, 46, 14, 14, P["t"], P["O"])
    # Flyer inside left window
    draw_box(c, 15, 51, 5, 6, P["W"], P["O"])
    c.set_pixel(17, 53, P["F"])
    c.fill_rect(16, 55, 3, 1, P["N"])

    # Sliding glass doors (x=28..63, y=45..62)
    draw_box(c, 28, 45, 18, 17, P["M"], P["O"], hi=P["W"], sh=P["d"])
    draw_glass_window(c, 29, 46, 16, 15, P["d"], P["O"])
    draw_box(c, 45, 45, 18, 17, P["M"], P["O"], hi=P["W"], sh=P["d"])
    draw_glass_window(c, 46, 46, 16, 15, P["d"], P["O"])
    # Horizontal door mid-bars
    c.fill_rect(29, 53, 16, 2, P["M"])
    c.fill_rect(46, 53, 16, 2, P["M"])

    # Inside left door: Chibi face poster + retro handheld console box
    draw_box(c, 36, 47, 7, 6, P["W"], P["O"])
    c.fill_rect(38, 48, 3, 2, P["H"])
    c.fill_rect(38, 50, 3, 2, P["S"])
    draw_box(c, 33, 55, 7, 6, P["G"], P["O"], sh=P["D"])
    c.fill_rect(35, 56, 3, 2, P["X"])

    # Inside right door: Wooden OPEN sign & Pink Blob Mascot!
    c.set_pixel(51, 46, P["M"])
    c.set_pixel(56, 46, P["M"])
    draw_box(c, 48, 47, 12, 5, P["T"], P["O"], hi=P["1"], sh=P["t"])
    # Crisp 3px tall mini 'OPEN' pattern across 10px inner width
    open_mini = [
        "11.11.11.1.1",
        "11.11.10.111",
        "11.10.11.1.1",
    ]
    for oy, row in enumerate(open_mini):
        for ox, ch in enumerate(row):
            if ch == "1":
                c.set_pixel(49 + ox, 48 + oy, P["C"] if oy < 2 else P["a"])

    # Pink Blob Mascot at bottom of right doorway (x=50..59, y=55..62)
    draw_box(c, 50, 56, 9, 6, P["K"], P["O"], hi=P["Q"], sh=P["k"])
    # Ears
    c.fill_rect(51, 54, 2, 2, P["K"])
    c.set_pixel(51, 54, P["O"])
    c.fill_rect(56, 54, 2, 2, P["K"])
    c.set_pixel(57, 54, P["O"])
    # Cute happy eyes & mouth
    c.set_pixel(52, 58, P["O"])
    c.set_pixel(56, 58, P["O"])
    c.set_pixel(54, 59, P["x"])
    c.set_pixel(53, 59, P["b"])
    c.set_pixel(55, 59, P["b"])

    # Red/White A-frame sidewalk sign in front of left window (x=21..29, y=53..62)
    draw_box(c, 21, 53, 8, 9, P["W"], P["O"], sh=P["w"])
    c.fill_rect(22, 54, 6, 7, P["R"])
    c.fill_rect(23, 55, 4, 2, P["W"])
    c.fill_rect(23, 58, 4, 2, P["C"])

    # Hanging wire & side foliage on left awning corner
    draw_leaf_cluster(c, [(8, 44, 2.8, 3.0), (10, 48, 2.2, 2.2)])

    # Stacked cardboard boxes on far left (x=2..11, y=54..62) + Little Black Cat!
    draw_box(c, 2, 58, 9, 4, P["a"], P["O"], hi=P["C"], sh=P["T"])
    c.fill_rect(4, 59, 4, 1, P["t"])
    draw_box(c, 3, 54, 8, 4, P["C"], P["O"], hi=P["W"], sh=P["a"])
    c.fill_rect(5, 55, 3, 1, P["T"])
    # Little Black Cat sitting on top box (x=3..10, y=49..54)
    c.fill_rect(4, 51, 5, 3, P["O"])
    c.fill_rect(5, 52, 3, 1, P["H"])  # inner dark espresso fur highlight
    c.set_pixel(5, 50, P["O"])       # left ear
    c.set_pixel(8, 50, P["O"])       # right ear
    c.set_pixel(6, 51, P["G"])       # gold eye
    c.set_pixel(8, 51, P["G"])       # gold eye
    c.set_pixel(3, 52, P["O"])       # curled tail
    c.set_pixel(3, 51, P["O"])

    # Red Drink Vending Machine on right (x=66..79, y=44..62)
    draw_vending_machine(c, 66, 44, P["R"], P["r"], P["F"])

    # Lush green potted bush on far right (x=81..93, y=49..62)
    draw_terracotta_pot(c, 84, 57, [(85, 53, 3.8, 3.5), (89, 53, 3.5, 3.2), (87, 49, 3.2, 2.8)])

    return c


def build_coffee_96x64() -> Canvas:
    """
    building_coffee.png (96x64) — Directly modelled after envref5.jpg ("COFFEE DAILY"):
    - Upper roof/wall: Concrete urban rooftop parapet with AC duct pipes & weathered AC unit.
    - Signboard: Warm cream facade framing a deep Forest-Green signboard with coffee cup icon
      and crisp cream COFFEE DAILY lettering.
    - Striped Awning: Deep Forest-Green & Crisp White scalloped vertical stripes.
    - Ground floor: Cream/brick wall, large dark-blue café window w/ reflection & cup sticker,
      wooden door w/ green OPEN sign, and Blue Drink Vending Machine on the left side!
    """
    c = Canvas(96, 64)
    fill_ellipse(c, 48, 62.5, 44, 2.0, P["s"])

    # --- ROOFTOP & PARAPET (x=17..81, y=2..21) ---
    draw_box(c, 17, 2, 64, 19, P["M"], P["O"], hi=P["w"], sh=P["d"])
    # Inner recessed roof slab
    draw_box(c, 20, 4, 58, 14, P["d"], P["O"], sh=P["n"])
    # Concrete seam lines
    c.fill_rect(21, 9, 56, 1, P["M"])
    c.fill_rect(35, 5, 1, 12, P["M"])
    c.fill_rect(63, 5, 1, 12, P["M"])
    # Roof parapet front rim highlight
    c.fill_rect(18, 19, 62, 1, P["w"])

    # Metal AC duct pipe curving from left into weathered AC unit
    c.fill_rect(36, 8, 2, 7, P["O"])
    c.fill_rect(37, 9, 1, 5, P["W"])
    c.fill_rect(37, 13, 6, 2, P["O"])
    c.fill_rect(38, 13, 5, 1, P["M"])
    draw_ac_unit(c, 42, 6, rusty=True)

    # --- SIGNBOARD FACADE (x=16..82, y=21..33) ---
    draw_box(c, 16, 21, 66, 13, P["C"], P["O"], hi=P["W"], sh=P["a"])
    # Subtle plaster specks
    c.set_pixel(18, 22, P["a"])
    c.set_pixel(78, 22, P["a"])
    c.set_pixel(79, 32, P["T"])

    # Deep Forest-Green wooden-framed signboard (x=18..80, y=23..32)
    draw_box(c, 18, 23, 62, 10, P["T"], P["O"], hi=P["1"], sh=P["t"])
    draw_box(c, 19, 24, 60, 8, P["fg_base"], P["fg_dark"], hi=P["fg_hi"], sh=P["fg_dark"])

    # Coffee Cup Icon on left of signboard (x=22..27, y=25..30)
    c.set_pixel(23, 25, P["C"])  # steam
    c.set_pixel(25, 25, P["C"])  # steam
    c.fill_rect(22, 27, 4, 2, P["W"])
    c.set_pixel(26, 27, P["C"])  # handle
    c.fill_rect(21, 30, 6, 1, P["a"])  # saucer

    # Crisp Cream 'COFFEE DAILY' lettering (x=29..77, y=26..30)
    draw_text(c, 30, 26, "COFFEE DAILY", P["W"], shadow=P["fg_dark"], spacing=1)

    # --- STRIPED AWNING: DEEP FOREST-GREEN & CRISP WHITE (x=15..83, y=34..43) ---
    draw_striped_awning(c, 15, 34, 68, 10, P["fg_base"], P["fg_dark"], P["W"], P["w"], stripe_w=6)

    # --- GROUND FLOOR CAFE FACADE (x=16..82, y=44..62) ---
    draw_box(c, 16, 44, 66, 18, P["C"], P["O"], hi=P["W"], sh=P["a"])
    # Warm terracotta/brick upper band under awning & base trim
    c.fill_rect(17, 45, 64, 3, P["brick_base"])
    c.fill_rect(17, 45, 64, 1, P["brick_sh"])
    c.fill_rect(17, 59, 64, 2, P["a"])
    c.set_pixel(20, 60, P["T"])
    c.set_pixel(49, 60, P["T"])

    # Large Dark-Blue Cafe Window (x=20..52, y=47..59)
    draw_glass_window(c, 20, 47, 32, 12, P["T"], P["t"])
    # White coffee-cup / menu sticker inside window
    draw_box(c, 39, 51, 5, 6, P["W"], P["g"], sh=P["w"])
    c.fill_rect(40, 53, 3, 2, P["T"])
    c.set_pixel(41, 52, P["fg_base"])

    # Wooden Cafe Door on Right (x=57..76, y=46..61)
    draw_box(c, 57, 46, 19, 16, P["t"], P["O"], hi=P["T"], sh=P["int_dark"])
    draw_box(c, 59, 47, 15, 14, P["T"], P["t"], hi=P["1"], sh=P["t"])
    draw_glass_window(c, 61, 49, 11, 7, P["t"], P["O"])
    # Door handle
    c.fill_rect(59, 55, 2, 1, P["W"])
    c.set_pixel(59, 56, P["M"])
    # Hanging green OPEN sign on door glass
    c.set_pixel(66, 50, P["W"])
    draw_box(c, 63, 51, 7, 4, P["Y"], P["O"], hi=P["l"], sh=P["y"])
    c.fill_rect(64, 52, 5, 1, P["W"])

    # Small green weed/sprout at bottom-right corner
    draw_leaf_cluster(c, [(83, 60, 1.8, 1.8)])

    # --- BLUE DRINK VENDING MACHINE ON LEFT (x=4..16, y=43..61) ---
    draw_vending_machine(c, 4, 43, P["J"], P["j"], P["B"])

    return c


def build_arcade_96x64() -> Canvas:
    """
    building_arcade.png (96x64) — Modelled after peak-envref4.jpg & envref6.png:
    - Rooftop garden terrace: Lush green hedge, wooden deck, rooftop lounge table/umbrella.
    - Facade: Mint/Cyan & Slate-Purple Pop-Urban facade with glowing 'ARCADE & BOOKS'
      neon sign, arcade/bookstore display windows, and built-in ATM on left wall!
    """
    c = Canvas(96, 64)
    fill_ellipse(c, 48, 62.5, 45, 2.0, P["s"])

    # --- ROOFTOP GARDEN TERRACE (x=9..87, y=4..21) ---
    # Wooden terrace deck & slate parapet
    draw_box(c, 9, 10, 78, 11, P["N"], P["O"], hi=P["M"], sh=P["n"])
    draw_box(c, 12, 12, 72, 7, P["T"], P["O"], hi=P["1"], sh=P["t"])
    for px in range(15, 83, 4):
        c.fill_rect(px, 13, 1, 5, P["t"])

    # Lush rooftop garden hedges along back & left
    draw_leaf_cluster(
        c,
        [
            (13, 12, 4.0, 3.2),
            (21, 10, 4.2, 2.8),
            (30, 10, 4.5, 2.6),
            (39, 10, 4.2, 2.6),
            (48, 10, 4.0, 2.6),
            (83, 12, 3.5, 3.0),
        ],
        deep_col=P["leaf_deep"],
    )

    # Rooftop neon antenna / speaker box on left deck (x=18..27, y=9..17)
    draw_box(c, 18, 11, 9, 6, P["U"], P["O"], hi=P["Z"], sh=P["u"])
    c.fill_rect(20, 12, 5, 3, P["g"])
    c.set_pixel(22, 13, P["X"])
    c.fill_rect(22, 6, 1, 5, P["M"])
    c.set_pixel(22, 5, P["K"])

    # Rooftop lounge table & Coral-Pink/White striped umbrella (x=58..78, y=2..18)
    # Table & two stools
    draw_box(c, 64, 14, 9, 3, P["W"], P["O"], sh=P["w"])
    c.fill_rect(68, 16, 1, 2, P["d"])
    draw_box(c, 60, 15, 3, 3, P["M"], P["O"])
    draw_box(c, 74, 15, 3, 3, P["M"], P["O"])
    # Pole
    c.fill_rect(68, 9, 1, 6, P["O"])
    # Umbrella canopy
    for py in range(5):
        uw = 4 + py * 3
        ux = 68 - uw // 2
        c.fill_rect(ux - 1, 3 + py, uw + 2, 1, P["O"])
        for px in range(uw):
            col = P["F"] if ((px // 2) % 2 == 0) else P["W"]
            if py == 4:
                col = P["f"] if ((px // 2) % 2 == 0) else P["w"]
            c.set_pixel(ux + px, 3 + py, col)
    c.fill_rect(67, 2, 3, 1, P["O"])

    # --- POP-URBAN MINT & SLATE-PURPLE FACADE (x=8..88, y=20..62) ---
    draw_box(c, 8, 20, 80, 42, P["mint_base"], P["O"], hi=P["mint_hi"], sh=P["mint_sh"])
    # Vertical slate-purple structural pilasters
    draw_box(c, 8, 20, 4, 42, P["U"], P["O"], hi=P["Z"], sh=P["u"])
    draw_box(c, 84, 20, 4, 42, P["U"], P["O"], hi=P["Z"], sh=P["u"])
    draw_box(c, 24, 36, 3, 26, P["N"], P["O"], hi=P["M"], sh=P["n"])
    draw_box(c, 66, 36, 3, 26, P["N"], P["O"], hi=P["M"], sh=P["n"])

    # --- GLOWING 'ARCADE & BOOKS' NEON SIGNBOARD (x=11..85, y=21..37) ---
    draw_box(c, 11, 21, 74, 16, P["n"], P["O"], hi=P["U"], sh=P["int_dark"])
    # Neon glowing outer border
    c.fill_rect(12, 22, 72, 1, P["X"])
    c.fill_rect(12, 35, 72, 1, P["K"])
    c.fill_rect(12, 22, 1, 14, P["X"])
    c.fill_rect(83, 22, 1, 14, P["K"])

    # Pixel lettering: 'ARCADE & BOOKS' centered cleanly with 1px dark margin
    draw_text(c, 22, 23, "ARCADE", P["K"], shadow=P["u"], spacing=2)
    draw_text(c, 24, 29, "& BOOKS", P["X"], shadow=P["2"], spacing=1)
    # Little neon stars/sparks on sign corners
    c.set_pixel(15, 25, P["G"])
    c.set_pixel(16, 26, P["W"])
    c.set_pixel(79, 25, P["G"])
    c.set_pixel(78, 26, P["W"])

    # --- LEFT WALL: BUILT-IN ATM MACHINE (x=12..24, y=38..61) ---
    draw_box(c, 12, 38, 12, 23, P["U"], P["O"], hi=P["Z"], sh=P["u"])
    draw_box(c, 13, 39, 10, 7, P["n"], P["O"])
    draw_text(c, 14, 40, "ATM", P["G"], spacing=0)
    # Glowing cyan ATM screen
    draw_box(c, 14, 47, 8, 5, P["X"], P["O"], hi=P["3"], sh=P["2"])
    c.fill_rect(15, 49, 5, 1, P["W"])
    # Keypad & card slot
    draw_box(c, 14, 53, 5, 4, P["M"], P["O"], sh=P["d"])
    c.fill_rect(20, 53, 2, 3, P["g"])
    c.fill_rect(15, 58, 6, 2, P["g"])

    # --- CENTER-LEFT WINDOW: ARCADE CABINETS INSIDE (x=27..49, y=38..59) ---
    draw_glass_window(c, 27, 38, 22, 21, P["N"], P["n"])
    # Two miniature glowing arcade cabinets visible inside the glass!
    for cab_x, screen_col in ((30, P["X"]), (40, P["K"])):
        draw_box(c, cab_x, 45, 7, 12, P["u"], P["O"], hi=P["U"])
        c.fill_rect(cab_x + 1, 46, 5, 2, P["G"])  # marquee
        c.fill_rect(cab_x + 1, 49, 5, 3, screen_col)  # glowing CRT
        c.set_pixel(cab_x + 2, 50, P["W"])
        c.fill_rect(cab_x + 1, 53, 5, 1, P["M"])  # control deck

    # --- CENTER ARCHWAY ENTRANCE (x=49..66, y=38..62) ---
    draw_box(c, 49, 38, 17, 24, P["2"], P["O"], hi=P["X"], sh=P["n"])
    draw_box(c, 51, 41, 13, 20, P["g"], P["O"], hi=P["X"])
    # Glowing doorway depth & neon steps
    c.fill_rect(52, 42, 11, 16, P["int_dark"])
    c.fill_rect(53, 44, 3, 12, P["z"])
    c.fill_rect(52, 58, 11, 2, P["M"])
    c.fill_rect(52, 60, 11, 1, P["d"])
    # Small neon OPEN sign above door
    c.fill_rect(54, 39, 7, 2, P["K"])
    c.fill_rect(55, 39, 5, 1, P["W"])

    # --- RIGHT WINDOW: COZY BOOKSTORE SHELVES (x=69..84, y=38..59) ---
    draw_glass_window(c, 69, 38, 15, 21, P["N"], P["n"])
    # 3 wooden bookshelves with colorful books inside!
    book_colors = [P["R"], P["G"], P["X"], P["K"], P["W"], P["B"], P["F"], P["Y"], P["C"]]
    for shelf_idx, sy in enumerate((43, 49, 55)):
        c.fill_rect(71, sy, 11, 1, P["T"])
        c.fill_rect(71, sy + 1, 11, 1, P["t"])
        for bx in range(71, 82):
            if (bx + shelf_idx) % 4 != 3:
                col = book_colors[(bx + shelf_idx * 3) % len(book_colors)]
                c.fill_rect(bx, sy - 3, 1, 3, col)

    return c


# ==============================================================================
# 1B. ADDITIONAL POP-URBAN BUILDINGS (RAMEN, BOOKSTORE, METRO, NEWSSTAND)
# ==============================================================================

def build_ramen_96x64() -> Canvas:
    """
    building_ramen.png (96x64) — Cozy Japanese Ramen Shop facade:
    - Dark slate Japanese tiled roof & warm wood rafters
    - Warm cream plaster upper floor with shoji-latticed glowing window & AC unit
    - Wooden signboard with mini pixel ramen bowl icon & crisp 'RAMEN' lettering
    - Hanging crimson noren curtains & warm glowing paper chochin lanterns
    - Open counter window w/ warm interior glow, two red bar stools & slatted sliding door
    """
    c = Canvas(96, 64)
    fill_ellipse(c, 48, 62.5, 45, 2.0, P["s"])

    # --- DARK SLATE TILED ROOF (x=8..87, y=2..11) ---
    draw_box(c, 10, 2, 76, 3, P["roof_hi"], P["O"], sh=P["roof_base"])
    draw_box(c, 8, 4, 80, 6, P["roof_base"], P["O"], hi=P["roof_hi"], sh=P["roof_sh"])
    for rx in range(11, 86, 4):
        c.fill_rect(rx, 5, 1, 4, P["roof_sh"])
        c.fill_rect(rx + 1, 5, 1, 3, P["roof_hi"])
    # Wooden eaves beam
    draw_box(c, 8, 9, 80, 3, P["T"], P["O"], hi=P["1"], sh=P["t"])

    # --- UPPER FLOOR WALL (x=11..84, y=12..24) ---
    draw_box(c, 11, 12, 74, 12, P["C"], P["O"], hi=P["W"], sh=P["a"])
    # Timber framing posts
    for tx in (12, 42, 58, 82):
        c.fill_rect(tx, 13, 2, 10, P["t"])
        c.fill_rect(tx, 13, 1, 10, P["T"])
    # Upper warm-lit wooden latticed window (x=16..38, y=14..22)
    draw_box(c, 16, 14, 22, 8, P["G"], P["O"], hi=P["W"], sh=P["D"])
    c.fill_rect(17, 15, 20, 6, P["C"])
    c.fill_rect(18, 16, 18, 4, P["G"])
    for lx in range(21, 37, 5):
        c.fill_rect(lx, 15, 1, 6, P["t"])
    c.fill_rect(17, 18, 20, 1, P["t"])
    # Upper right AC unit & ivy
    draw_ac_unit(c, 63, 13)
    draw_leaf_cluster(c, [(60, 17, 2.2, 2.5), (79, 19, 2.4, 2.6)])

    # --- WOODEN 'RAMEN' SIGNBOARD (x=11..84, y=24..34) ---
    draw_box(c, 11, 24, 74, 11, P["t"], P["O"], hi=P["T"], sh=P["int_dark"])
    draw_box(c, 13, 25, 70, 9, P["roof_sh"], P["T"], hi=P["1"], sh=P["t"])
    # Mini pixel ramen bowl icon (x=20..28, y=26..32)
    c.set_pixel(22, 26, P["W"])
    c.set_pixel(24, 26, P["C"])
    c.set_pixel(26, 26, P["W"])
    c.fill_rect(21, 28, 7, 1, P["G"])  # noodles
    c.fill_rect(21, 29, 7, 2, P["W"])  # bowl
    c.fill_rect(22, 31, 5, 1, P["R"])  # bowl stripe
    c.fill_rect(23, 32, 3, 1, P["a"])  # foot
    # 'RAMEN' crisp neon-cream text
    draw_text(c, 33, 27, "RAMEN", P["G"], shadow=P["R"], spacing=2)
    # Red stamp seal on right of sign
    draw_box(c, 69, 27, 5, 5, P["R"], P["O"], hi=P["F"], sh=P["r"])
    c.set_pixel(71, 29, P["W"])

    # --- HANGING RED NOREN CURTAINS & PAPER LANTERNS (y=35..46) ---
    draw_box(c, 9, 34, 78, 3, P["T"], P["O"], hi=P["1"], sh=P["t"])
    # Ground floor warm wooden storefront behind noren (x=11..84, y=37..62)
    draw_box(c, 11, 37, 74, 25, P["T"], P["O"], hi=P["1"], sh=P["t"])
    # Warm lit counter service window (x=24..56, y=38..55)
    draw_box(c, 24, 38, 32, 16, P["G"], P["O"], hi=P["W"], sh=P["D"])
    c.fill_rect(25, 39, 30, 14, P["C"])
    c.fill_rect(25, 48, 30, 5, P["G"])
    # Hanging steam/pots silhouette inside kitchen
    c.fill_rect(31, 45, 5, 3, P["d"])
    c.fill_rect(42, 46, 6, 3, P["M"])
    # Wooden counter shelf & 2 red-cushioned bar stools (x=23..57, y=53..62)
    draw_box(c, 23, 53, 34, 3, P["1"], P["O"], hi=P["C"], sh=P["t"])
    for sx in (30, 44):
        draw_box(c, sx, 56, 7, 3, P["R"], P["O"], hi=P["F"], sh=P["r"])
        c.fill_rect(sx + 2, 59, 3, 3, P["O"])
        c.fill_rect(sx + 3, 59, 1, 2, P["M"])

    # Right slatted wooden entrance door (x=59..77, y=42..61)
    draw_box(c, 59, 42, 18, 20, P["t"], P["O"], hi=P["T"], sh=P["int_dark"])
    draw_glass_window(c, 61, 44, 14, 9, P["T"], P["t"])
    for sx in range(62, 75, 3):
        c.fill_rect(sx, 54, 1, 6, P["1"])

    # Left menu chalkboard (x=13..22, y=47..59)
    draw_box(c, 13, 47, 9, 12, P["g"], P["O"], hi=P["T"], sh=P["t"])
    c.fill_rect(15, 49, 5, 1, P["W"])
    c.fill_rect(15, 51, 4, 1, P["G"])
    c.fill_rect(15, 53, 5, 1, P["F"])
    c.fill_rect(15, 55, 3, 1, P["W"])

    # 5 Crimson Noren curtain panels hanging over counter & door (x=20..76, y=36..44)
    for idx in range(5):
        nx = 20 + idx * 11
        draw_box(c, nx, 36, 10, 8, P["R"], P["O"], hi=P["F"], sh=P["r"])
        # White crest circle on alternating panels
        if idx in (1, 3):
            c.fill_rect(nx + 4, 38, 2, 3, P["W"])
            c.fill_rect(nx + 3, 39, 4, 1, P["W"])
        elif idx == 2:
            c.fill_rect(nx + 3, 39, 4, 2, P["C"])

    # Left & Right glowing paper chochin lanterns (x=9..17 and x=78..86, y=36..46)
    for lx in (9, 78):
        fill_ellipse(c, lx + 4, 41, 6.5, 5.5, P["lantern_glow"])
        c.fill_rect(lx + 2, 36, 4, 1, P["O"])  # top cap
        draw_box(c, lx + 1, 37, 6, 8, P["G"], P["O"], hi=P["W"], sh=P["R"])
        c.fill_rect(lx + 2, 39, 4, 4, P["C"])
        c.fill_rect(lx + 3, 40, 2, 2, P["W"])
        c.fill_rect(lx + 2, 45, 4, 1, P["O"])  # bottom cap

    # Side potted bamboo/plant & red crate
    draw_terracotta_pot(c, 84, 56, [(86, 51, 3.0, 3.8), (89, 53, 2.5, 2.8)])
    draw_box(c, 3, 55, 8, 7, P["R"], P["O"], hi=P["F"], sh=P["r"])
    c.fill_rect(4, 57, 6, 1, P["O"])
    c.fill_rect(4, 59, 6, 1, P["O"])

    return c


def build_bookstore_96x64() -> Canvas:
    """
    building_bookstore.png (96x64) — Indie Bookstore facade:
    - Terracotta/brick upper facade with two warm upstairs windows
    - Deep teal signboard with open-book icon & crisp 'LIBRI' lettering
    - Teal & mustard striped canopy awning
    - Warm-lit display window packed with colorful book spines, cozy entrance & outdoor book bin
    """
    c = Canvas(96, 64)
    fill_ellipse(c, 48, 62.5, 45, 2.0, P["s"])

    # --- UPPER BRICK PARAPET & WALL (x=10..85, y=3..23) ---
    draw_box(c, 9, 3, 78, 4, P["C"], P["O"], hi=P["W"], sh=P["a"])
    draw_box(c, 11, 7, 74, 16, P["brick_base"], P["O"], hi=P["brick_hi"], sh=P["brick_sh"])
    # Subtle brick course accents
    for by in (10, 14, 18):
        for bx in range(13 + (by % 4), 81, 8):
            c.fill_rect(bx, by, 4, 1, P["brick_sh"])
    # Two cozy upstairs windows (x=17..35 & x=60..78, y=9..21)
    for wx in (17, 60):
        draw_box(c, wx, 9, 18, 12, P["C"], P["O"], hi=P["W"], sh=P["a"])
        draw_glass_window(c, wx + 2, 11, 14, 8, P["teal_base"], P["teal_sh"])
        # Warm curtain glow on one side
        c.fill_rect(wx + 3, 12, 4, 6, P["mustard_base"])
        c.fill_rect(wx + 3, 12, 2, 6, P["mustard_hi"])
        # Flower box on window sill
        draw_box(c, wx + 1, 20, 16, 3, P["T"], P["O"], sh=P["t"])
        draw_leaf_cluster(c, [(wx + 5, 19, 2.2, 1.5), (wx + 11, 19, 2.4, 1.5)])
        c.set_pixel(wx + 6, 18, P["K"])
        c.set_pixel(wx + 12, 18, P["G"])

    # --- DEEP TEAL 'LIBRI' SIGNBOARD (x=12..83, y=23..33) ---
    draw_box(c, 12, 23, 72, 11, P["C"], P["O"], hi=P["W"], sh=P["a"])
    draw_box(c, 14, 24, 68, 9, P["teal_base"], P["O"], hi=P["teal_hi"], sh=P["teal_sh"])
    # Open book pixel icon on left (x=19..28, y=26..31)
    c.fill_rect(19, 27, 4, 4, P["W"])
    c.fill_rect(24, 27, 4, 4, P["C"])
    c.fill_rect(23, 27, 1, 4, P["O"])
    c.fill_rect(18, 31, 11, 1, P["mustard_base"])
    # 'LIBRI' crisp mustard-gold text
    draw_text(c, 35, 26, "LIBRI", P["mustard_hi"], shadow=P["teal_sh"], spacing=2)
    c.set_pixel(67, 28, P["F"])
    c.set_pixel(70, 28, P["W"])
    c.set_pixel(73, 28, P["mustard_hi"])

    # --- TEAL & MUSTARD STRIPED CANOPY (x=8..87, y=34..43) ---
    draw_striped_awning(
        c, 8, 34, 80, 10, P["teal_base"], P["teal_sh"], P["mustard_base"], P["mustard_sh"], stripe_w=5
    )

    # --- GROUND FLOOR BOOKSTORE FACADE (x=11..84, y=44..62) ---
    draw_box(c, 11, 44, 74, 18, P["T"], P["O"], hi=P["1"], sh=P["t"])
    # Large warm-lit display window (x=14..52, y=45..59)
    draw_box(c, 14, 45, 38, 15, P["C"], P["O"], hi=P["W"], sh=P["a"])
    c.fill_rect(15, 46, 36, 13, P["int_dark"])
    # Warm reading-lamp glow inside window
    c.fill_rect(16, 46, 34, 12, hex_to_rgba("#3D2B24"))
    fill_ellipse(c, 33, 51, 14, 5, hex_to_rgba("#FFD166", 75))
    # 2 bookshelves packed with colorful book spines
    spines = [P["R"], P["G"], P["X"], P["K"], P["B"], P["F"], P["Y"], P["mustard_hi"], P["W"], P["U"]]
    for sy in (51, 57):
        c.fill_rect(16, sy, 34, 1, P["T"])
        c.fill_rect(16, sy + 1, 34, 1, P["t"])
        for bx in range(17, 49):
            if bx % 6 != 0:
                bh = 3 if bx % 2 == 0 else 4
                c.fill_rect(bx, sy - bh, 1, bh, spines[(bx * 3 + sy) % len(spines)])

    # Cozy entrance door (x=54..71, y=45..61)
    draw_box(c, 54, 45, 17, 17, P["teal_base"], P["O"], hi=P["teal_hi"], sh=P["teal_sh"])
    draw_glass_window(c, 57, 47, 11, 8, P["C"], P["a"])
    c.fill_rect(56, 56, 2, 1, P["mustard_hi"])  # brass handle
    draw_box(c, 59, 50, 7, 3, P["mustard_base"], P["O"])
    c.fill_rect(60, 51, 5, 1, P["W"])

    # Outdoor wooden book display cart on right (x=73..84, y=52..61)
    draw_box(c, 73, 52, 11, 9, P["T"], P["O"], hi=P["1"], sh=P["t"])
    for bx in range(75, 83):
        c.fill_rect(bx, 50, 1, 3, spines[bx % len(spines)])
    c.fill_rect(74, 53, 9, 1, P["O"])

    # Side potted fern on left (x=3..10, y=52..62)
    draw_terracotta_pot(c, 3, 56, [(6, 52, 3.2, 3.0), (8, 53, 2.5, 2.4)])

    return c


def build_metro_96x64() -> Canvas:
    """
    building_metro.png (96x64) — Modern urban subway station entrance:
    - Glowing Cyan/Blue 'M' badge totem overhead
    - Sleek cantilevered steel & cyan-glass canopy with neon trim
    - Concrete retaining walls, descending perspective stairs & 3 metallic handrails
    - Illuminated underground tunnel depth & sidewalk info totem
    """
    c = Canvas(96, 64)
    fill_ellipse(c, 48, 62.5, 44, 2.0, P["s"])

    # --- OVERHEAD GLOWING 'M' BADGE TOTEM (x=34..61, y=2..17) ---
    fill_ellipse(c, 48, 9, 14, 7, hex_to_rgba("#00F0FF", 55))
    draw_box(c, 35, 2, 26, 13, P["J"], P["O"], hi=P["B"], sh=P["j"])
    draw_box(c, 37, 3, 22, 11, P["n"], P["O"], hi=P["X"], sh=P["int_dark"])
    # Bold glowing 'M' emblem (centered x=44..50, y=5..10)
    m_glyph = [
        "11...11",
        "111.111",
        "1.111.1",
        "1..1..1",
        "1.....1",
        "1.....1",
    ]
    for my, row in enumerate(m_glyph):
        for mx, ch in enumerate(row):
            if ch == "1":
                c.set_pixel(44 + mx, 5 + my, P["W"] if my == 0 else (P["3"] if my < 2 else P["X"]))
    # Glowing side bars
    c.fill_rect(39, 7, 3, 2, P["F"])
    c.fill_rect(54, 7, 3, 2, P["F"])

    # --- CANTILEVERED GLASS & STEEL CANOPY (x=8..87, y=15..29) ---
    draw_box(c, 10, 15, 76, 3, P["M"], P["O"], hi=P["W"], sh=P["d"])
    draw_box(c, 8, 17, 80, 9, P["2"], P["O"], hi=P["3"], sh=P["n"])
    # Glass canopy skylight panes with diagonal reflections
    for gx in range(12, 82, 18):
        draw_glass_window(c, gx, 18, 15, 6, P["M"], P["d"])
        c.fill_rect(gx + 2, 19, 11, 1, P["3"])
    # Neon cyan under-canopy light strip
    draw_box(c, 8, 25, 80, 4, P["M"], P["O"], hi=P["W"], sh=P["d"])
    c.fill_rect(10, 27, 76, 1, P["X"])

    # --- STRUCTURAL PILLARS & RETAINING WALLS (y=28..62) ---
    # Concrete side walls
    draw_box(c, 10, 28, 12, 34, P["cobble_base"], P["O"], hi=P["cobble_hi"], sh=P["cobble_sh"])
    draw_box(c, 74, 28, 12, 34, P["cobble_base"], P["O"], hi=P["cobble_hi"], sh=P["cobble_sh"])
    # Steel front columns with safety stripes
    for px in (11, 80):
        draw_box(c, px, 28, 5, 34, P["M"], P["O"], hi=P["W"], sh=P["d"])
        c.fill_rect(px + 1, 48, 3, 2, P["X"])
        c.fill_rect(px + 1, 52, 3, 2, P["G"])
    # Left wall glowing metro route map screen
    draw_box(c, 17, 36, 5, 13, P["P"], P["O"], hi=P["M"])
    c.fill_rect(18, 38, 3, 8, P["X"])
    c.set_pixel(19, 40, P["R"])
    c.set_pixel(19, 43, P["G"])

    # --- DESCENDING STAIRWELL & HANDRAILS (x=22..73, y=29..62) ---
    draw_box(c, 22, 29, 52, 33, P["int_dark"], P["O"])
    # Underground station glow at top/back of stairs
    c.fill_rect(26, 30, 44, 6, P["n"])
    c.fill_rect(32, 31, 32, 2, P["X"])
    # 7 descending perspective stone steps (wider at bottom, darker deeper)
    for step_idx in range(7):
        sy = 35 + step_idx * 4
        tread_col = P["cobble_hi"] if step_idx >= 4 else (P["cobble_base"] if step_idx >= 2 else P["M"])
        riser_col = P["cobble_sh"] if step_idx >= 3 else P["d"]
        c.fill_rect(23, sy, 50, 2, tread_col)
        c.fill_rect(23, sy + 2, 50, 2, riser_col)
        c.fill_rect(23, sy + 3, 50, 1, P["O"])
    # Bottom sidewalk threshold
    c.fill_rect(22, 60, 52, 2, P["cobble_hi"])
    c.fill_rect(22, 62, 52, 1, P["O"])

    # 3 Sleek metallic handrails (left, center divider, right)
    for rx in (24, 47, 70):
        c.fill_rect(rx, 33, 2, 28, P["O"])
        c.fill_rect(rx, 33, 1, 27, P["W"])
        c.fill_rect(rx + 1, 34, 1, 26, P["M"])
        # Posts
        for py in (36, 48, 58):
            c.fill_rect(rx - 1, py, 4, 2, P["d"])
            c.set_pixel(rx, py, P["W"])

    return c


def build_newsstand_48x40() -> Canvas:
    """
    building_newsstand.png (48x40) — Italian/Pop-Urban street 'EDICOLA' kiosk:
    - Slate-blue roof & signboard with 'NEWS' / 'EDICOLA' lettering
    - Coral-pink & cream striped awning
    - Open service counter & side/front colorful magazine racks
    """
    c = Canvas(48, 40)
    fill_ellipse(c, 24, 38.0, 21, 1.8, P["s"])

    # Roof & Header Signboard (x=5..42, y=2..11)
    draw_box(c, 5, 2, 38, 3, P["M"], P["O"], hi=P["W"], sh=P["d"])
    draw_box(c, 6, 4, 36, 8, P["N"], P["O"], hi=P["M"], sh=P["n"])
    draw_text(c, 10, 6, "EDICOLA", P["C"], shadow=P["O"], spacing=1)

    # Coral & Cream Striped Awning (x=3..44, y=11..18)
    draw_striped_awning(c, 3, 11, 42, 7, P["F"], P["f"], P["C"], P["a"], stripe_w=4)

    # Kiosk body (x=6..41, y=18..37)
    draw_box(c, 6, 18, 36, 19, P["C"], P["O"], hi=P["W"], sh=P["a"])
    draw_box(c, 6, 33, 36, 4, P["N"], P["O"], hi=P["M"], sh=P["n"])

    # Open service window (x=14..33, y=19..28)
    draw_box(c, 14, 19, 20, 9, P["T"], P["O"], sh=P["int_dark"])
    c.fill_rect(15, 20, 18, 7, P["int_dark"])
    # Hanging magazines & warm interior light inside booth
    c.fill_rect(17, 20, 14, 2, P["G"])
    c.fill_rect(16, 22, 4, 4, P["K"])
    c.fill_rect(22, 22, 4, 4, P["X"])
    c.fill_rect(28, 22, 4, 4, P["W"])
    # Wooden counter ledge
    draw_box(c, 13, 27, 22, 2, P["1"], P["O"], sh=P["t"])

    # Colorful Magazine Racks (Left, Right, Bottom front)
    mags = [P["R"], P["G"], P["X"], P["K"], P["B"], P["F"], P["Y"], P["W"]]
    # Left & Right side racks
    for rx in (7, 35):
        draw_box(c, rx, 19, 6, 14, P["M"], P["O"], sh=P["d"])
        for my in (20, 25, 29):
            c.fill_rect(rx + 1, my, 2, 3, mags[(rx + my) % len(mags)])
            c.fill_rect(rx + 3, my, 2, 3, mags[(rx + my + 3) % len(mags)])
    # Front bottom rack below counter
    draw_box(c, 15, 29, 18, 6, P["M"], P["O"], sh=P["d"])
    for mx in range(16, 32, 3):
        c.fill_rect(mx, 30, 2, 4, mags[mx % len(mags)])
        c.set_pixel(mx, 30, P["W"])

    return c


# ==============================================================================
# 2B. ADDITIONAL URBAN STREET & CANAL PROPS
# ==============================================================================

def build_prop_scooter_32x24() -> Canvas:
    """prop_scooter.png (32x24) — Parked retro mint/cream scooter in 3/4 top-down view."""
    c = Canvas(32, 24)
    fill_ellipse(c, 16, 20.5, 11, 2.2, P["s"])

    # Rear & Front tires
    for wx in (6, 21):
        draw_box(c, wx, 15, 6, 5, P["P"], P["O"], sh=P["p"])
        c.fill_rect(wx + 1, 16, 4, 3, P["C"])
        c.fill_rect(wx + 2, 17, 2, 1, P["M"])

    # Rear rounded mint fender & body (x=5..17, y=10..17)
    draw_box(c, 5, 11, 12, 6, P["mint_base"], P["O"], hi=P["mint_hi"], sh=P["mint_sh"])
    c.set_pixel(5, 13, P["R"])  # taillight
    # Dark floorboard bridge (x=14..21, y=14..17)
    draw_box(c, 14, 14, 8, 3, P["p"], P["O"])
    # Warm brown leather saddle (x=8..16, y=8..12)
    draw_box(c, 8, 9, 8, 3, P["T"], P["O"], hi=P["1"], sh=P["t"])

    # Front leg-shield & steering column (x=19..25, y=7..17)
    draw_box(c, 20, 8, 5, 9, P["mint_base"], P["O"], hi=P["mint_hi"], sh=P["mint_sh"])
    c.fill_rect(21, 10, 2, 5, P["C"])  # cream front trim
    # Handlebars & chrome mirrors (y=4..8)
    draw_box(c, 18, 6, 8, 3, P["mint_hi"], P["O"], sh=P["mint_base"])
    c.fill_rect(17, 6, 2, 2, P["P"])  # left grip
    c.fill_rect(24, 4, 1, 2, P["M"])  # mirror stem
    draw_box(c, 23, 2, 3, 3, P["W"], P["O"], sh=P["M"])  # mirror
    # Round glowing headlight
    draw_box(c, 24, 7, 3, 3, P["G"], P["O"], hi=P["W"])

    return c


def build_prop_bike_rack_32x20() -> Canvas:
    """prop_bike_rack.png (32x20) — Metal U-rack with 2 parked urban bicycles."""
    c = Canvas(32, 20)
    fill_ellipse(c, 16, 17.5, 13, 1.8, P["s"])

    # Ground metal rack rail + 2 U-loops
    draw_box(c, 3, 15, 26, 2, P["M"], P["O"], sh=P["d"])
    for lx in (8, 20):
        c.fill_rect(lx, 6, 4, 10, P["O"])
        c.fill_rect(lx + 1, 7, 2, 8, (0, 0, 0, 0))
        c.fill_rect(lx, 6, 4, 1, P["W"])
        c.fill_rect(lx, 7, 1, 8, P["M"])
        c.fill_rect(lx + 3, 7, 1, 8, P["d"])

    # Bike 1 (Left, Coral-Red, x=3..16, y=5..16)
    for wx in (3, 11):
        c.fill_rect(wx, 11, 5, 5, P["O"])
        c.fill_rect(wx + 1, 12, 3, 3, P["W"])
        c.set_pixel(wx + 2, 13, P["M"])
    c.fill_rect(5, 9, 8, 2, P["R"])   # frame crossbar
    c.fill_rect(6, 11, 2, 2, P["r"])  # seat tube
    c.fill_rect(5, 7, 3, 2, P["T"])   # brown saddle
    c.fill_rect(12, 7, 2, 3, P["M"])  # handlebars
    c.fill_rect(13, 8, 3, 3, P["C"])  # wicker basket

    # Bike 2 (Right, Cyan/Blue, x=16..29, y=5..16)
    for wx in (16, 24):
        c.fill_rect(wx, 11, 5, 5, P["O"])
        c.fill_rect(wx + 1, 12, 3, 3, P["W"])
        c.set_pixel(wx + 2, 13, P["M"])
    c.fill_rect(18, 9, 8, 2, P["B"])   # blue frame
    c.fill_rect(19, 11, 2, 2, P["J"])  # seat tube
    c.fill_rect(18, 7, 3, 2, P["H"])   # saddle
    c.fill_rect(25, 7, 3, 2, P["W"])   # silver handlebars

    return c


def build_prop_hydrant_bin_24x24() -> Canvas:
    """prop_hydrant_bin.png (24x24) — Urban street trash bin + red fire hydrant."""
    c = Canvas(24, 24)
    fill_ellipse(c, 7.5, 21.0, 5.5, 1.6, P["s"])
    fill_ellipse(c, 17.5, 21.0, 4.0, 1.4, P["s"])

    # Left: Hooded urban recycling/trash bin (x=2..12, y=5..21)
    draw_box(c, 3, 5, 9, 4, P["N"], P["O"], hi=P["M"], sh=P["n"])  # hooded top
    c.fill_rect(5, 7, 5, 2, P["int_dark"])  # waste slot
    draw_box(c, 3, 9, 9, 12, P["M"], P["O"], hi=P["W"], sh=P["d"])
    c.fill_rect(5, 11, 5, 3, P["Y"])  # eco recycling badge
    c.set_pixel(7, 12, P["W"])
    c.fill_rect(4, 16, 7, 1, P["d"])
    c.fill_rect(4, 18, 7, 1, P["d"])

    # Right: Classic Pop-Urban red hydrant (x=14..21, y=9..21)
    draw_box(c, 16, 9, 4, 2, P["M"], P["O"], hi=P["W"])  # top nut
    draw_box(c, 15, 11, 6, 9, P["R"], P["O"], hi=P["F"], sh=P["r"])
    # Side nozzles
    draw_box(c, 13, 13, 3, 3, P["M"], P["O"], hi=P["W"])
    draw_box(c, 20, 13, 3, 3, P["M"], P["O"], hi=P["W"])
    c.fill_rect(17, 14, 2, 2, P["C"])  # front cap
    draw_box(c, 14, 19, 8, 3, P["r"], P["O"], hi=P["R"])  # base flange

    return c


def build_prop_ztl_sign_24x32() -> Canvas:
    """prop_ztl_sign.png (24x32) — Glowing red/amber ZTL pole with traffic camera & beacon."""
    c = Canvas(24, 32)
    fill_ellipse(c, 12, 29.0, 5.5, 1.6, P["s"])

    # Weighted base & metal pole
    draw_box(c, 8, 26, 8, 4, P["P"], P["O"], hi=P["M"], sh=P["p"])
    c.fill_rect(10, 14, 4, 13, P["O"])
    c.fill_rect(11, 14, 1, 12, P["W"])
    c.fill_rect(12, 14, 1, 12, P["M"])

    # Glowing red/amber warning halo at top
    fill_ellipse(c, 12, 5, 7, 4.5, hex_to_rgba("#FF5C77", 55))

    # Top glowing amber/red beacon dome (x=9..14, y=2..6)
    draw_box(c, 9, 2, 6, 4, P["R"], P["O"], hi=P["G"], sh=P["r"])
    c.set_pixel(11, 3, P["W"])

    # Circular No-Entry ZTL sign (x=6..17, y=6..16)
    sign = Canvas(24, 32)
    fill_ellipse(sign, 12, 11, 4.8, 4.8, P["R"])
    for py in range(7, 16):
        for px in range(7, 18):
            if sign.pixels[py][px][3] > 0 and py >= 13:
                sign.set_pixel(px, py, P["r"])
    sign.fill_rect(9, 10, 6, 2, P["W"])  # horizontal no-entry bar
    apply_1px_outline(sign, P["O"])
    c.blit(sign, 0, 0)

    # Small side CCTV ZTL camera pod (x=15..21, y=7..11)
    draw_box(c, 15, 7, 6, 4, P["P"], P["O"], hi=P["W"], sh=P["p"])
    c.set_pixel(19, 8, P["X"])  # cyan lens
    c.set_pixel(16, 8, P["R"])  # recording LED

    return c


def build_prop_canal_boat_48x24() -> Canvas:
    """prop_canal_boat.png (48x24) — Colorful moored wooden canal boat with warm lantern on bow."""
    c = Canvas(48, 24)

    # Water ripples & hull shadow
    fill_ellipse(c, 24, 19.0, 20, 3.0, P["water_sh"])
    c.fill_rect(8, 21, 12, 1, P["water_glint"])
    c.fill_rect(27, 21, 10, 1, P["water_glint"])

    # Wooden hull (tapered bow on right x=4..43, y=9..19)
    hull = Canvas(48, 24)
    hull.fill_rect(6, 10, 31, 8, P["T"])
    hull.fill_rect(37, 11, 4, 6, P["T"])
    hull.fill_rect(41, 12, 3, 3, P["1"])
    hull.fill_rect(5, 11, 2, 5, P["t"])
    # Lower hull shading
    hull.fill_rect(6, 15, 34, 3, P["t"])
    # Colorful teal & coral gunwale stripe
    hull.fill_rect(6, 11, 34, 2, P["mint_base"])
    hull.fill_rect(6, 10, 33, 1, P["1"])
    hull.fill_rect(12, 13, 22, 1, P["F"])
    # Interior cockpit & wooden benches
    hull.fill_rect(10, 10, 23, 3, P["wood_sh"])
    hull.fill_rect(13, 10, 4, 3, P["C"])
    hull.fill_rect(22, 10, 4, 3, P["C"])
    apply_1px_outline(hull, P["O"])
    c.blit(hull, 0, 0)

    # Coiled mooring rope on stern (left)
    draw_box(c, 7, 8, 4, 3, P["C"], P["O"], sh=P["a"])

    # Warm glowing brass lantern on bow (right x=38..44, y=3..11)
    fill_ellipse(c, 41, 7, 5.5, 4.5, P["lantern_glow"])
    c.fill_rect(40, 8, 2, 4, P["O"])  # post
    draw_box(c, 39, 4, 4, 5, P["G"], P["O"], hi=P["W"], sh=P["D"])
    c.fill_rect(40, 5, 2, 3, P["W"])

    return c


def build_prop_slot_machine_32x32() -> Canvas:
    """Gold-trimmed plum cabinet with cream reels and a ruby pull lever."""
    c = Canvas(32, 32)
    fill_ellipse(c, 16, 29, 12, 2, P["s"])

    # Lever and base sit behind the stepped cabinet silhouette.
    c.fill_rect(24, 16, 6, 4, P["O"])
    c.fill_rect(28, 10, 2, 9, P["O"])
    c.fill_rect(28, 11, 1, 7, P["M"])
    draw_box(c, 26, 7, 5, 5, P["K"], P["O"], hi=P["Q"], sh=P["k"])
    draw_box(c, 5, 8, 22, 21, P["u"], P["O"], hi=P["U"], sh=P["V"])
    draw_box(c, 6, 24, 20, 6, P["V"], P["O"], hi=P["U"], sh=P["v"])

    # An enamel marquee with narrow brass trim frames the luminous 777.
    draw_box(c, 6, 2, 20, 9, P["D"], P["O"], hi=P["G"], sh=P["T"])
    c.fill_rect(8, 4, 16, 5, P["int_dark"])
    draw_text(c, 10, 4, "777", P["G"])
    c.set_pixel(8, 6, P["F"])
    c.set_pixel(23, 6, P["F"])

    # Three distinct symbols read as reels, even in a quick glance at native size.
    c.fill_rect(6, 11, 20, 9, P["O"])
    c.fill_rect(7, 11, 18, 1, P["Z"])
    for x in (8, 14, 20):
        c.fill_rect(x, 12, 5, 6, P["C"])
        c.fill_rect(x, 12, 5, 1, P["W"])
        c.fill_rect(x, 17, 5, 1, P["a"])
    c.fill_rect(9, 15, 4, 2, P["R"])
    c.fill_rect(11, 13, 1, 2, P["y"])
    c.set_pixel(12, 14, P["Y"])
    c.fill_rect(15, 14, 3, 3, P["D"])
    c.set_pixel(16, 13, P["G"])
    c.set_pixel(16, 15, P["G"])
    c.fill_rect(21, 13, 3, 1, P["k"])
    c.set_pixel(23, 14, P["k"])
    c.fill_rect(22, 15, 1, 2, P["k"])

    # Capped input rail and recessed coin tray.
    draw_box(c, 6, 19, 20, 5, P["u"], P["O"], hi=P["Z"], sh=P["V"])
    c.fill_rect(9, 21, 3, 1, P["Q"])
    c.fill_rect(14, 21, 3, 1, P["G"])
    c.fill_rect(19, 21, 3, 1, P["Q"])
    c.fill_rect(10, 25, 12, 3, P["O"])
    c.fill_rect(11, 25, 10, 1, P["T"])
    c.fill_rect(14, 26, 4, 1, P["D"])
    return c


def build_prop_manhole_24x24() -> Canvas:
    """prop_manhole.png (24x24) — Subtle non-solid asphalt-harmonized street manhole cover."""
    c = Canvas(24, 24)
    # 3/4 oval manhole harmonized with #524E61 asphalt (no harsh #1C1124 ring)
    fill_ellipse(c, 12.0, 12.0, 9.0, 6.8, P["asphalt_deep"])
    fill_ellipse(c, 11.7, 11.7, 8.2, 6.0, P["asphalt_light"])
    fill_ellipse(c, 12.0, 12.0, 7.2, 5.2, P["asphalt_dark"])
    # Subtle concentric groove & crisp non-noisy cast-iron waffle pattern
    for angle_y in range(8, 17, 2):
        for px in range(7, 18):
            dx = (px + 0.5 - 12.0) / 6.2
            dy = (angle_y + 0.5 - 12.0) / 4.2
            if dx * dx + dy * dy <= 1.0:
                c.set_pixel(px, angle_y, P["asphalt_deep"] if (px + angle_y) % 2 == 0 else P["asphalt_light"])
    # Center ring & two lift notches
    c.fill_rect(10, 11, 4, 2, P["asphalt_deep"])
    c.fill_rect(11, 11, 2, 1, P["asphalt_light"])
    c.fill_rect(5, 11, 2, 2, P["asphalt_deep"])
    c.fill_rect(17, 11, 2, 2, P["asphalt_deep"])
    return c


def build_prop_asphalt_crack_32x16() -> Canvas:
    """prop_asphalt_crack.png (32x16) — Subtle non-solid branching road crack & tar patch."""
    c = Canvas(32, 16)
    # Subtle repaired tar patch on left
    c.fill_rect(3, 8, 6, 4, P["asphalt_dark"])
    c.fill_rect(4, 7, 4, 1, P["asphalt_dark"])
    c.fill_rect(4, 8, 4, 1, P["asphalt_light"])
    # Main branching road crack in dark/light asphalt tones
    crack_pts = [
        (8, 9), (9, 9), (10, 8), (11, 8), (12, 7), (13, 7), (14, 6), (15, 6),
        (16, 7), (17, 7), (18, 8), (19, 8), (20, 8), (21, 9), (22, 9), (23, 10),
        (24, 10), (25, 11), (26, 11), (27, 11),
        # Upper branch
        (15, 5), (16, 4), (17, 4), (18, 3), (19, 3),
        # Lower branch
        (20, 9), (21, 10), (21, 11), (22, 12),
    ]
    for cx, cy in crack_pts:
        c.set_pixel(cx, cy - 1, P["asphalt_light"])
    for cx, cy in crack_pts:
        c.set_pixel(cx, cy, P["asphalt_deep"])
    return c


def build_prop_flower_box_32x20() -> Canvas:
    """prop_flower_box.png (32x20) — Wooden/stone urban planter box with lush greenery & flowers."""
    c = Canvas(32, 20)
    fill_ellipse(c, 16, 17.5, 13, 1.8, P["s"])

    # Planter box (wood body + stone rim/corners, x=4..27, y=9..17)
    draw_box(c, 4, 9, 24, 8, P["T"], P["O"], hi=P["1"], sh=P["t"])
    # Stone corner pillars & bottom plinth
    draw_box(c, 4, 9, 3, 8, P["cobble_base"], P["O"], hi=P["cobble_hi"], sh=P["cobble_sh"])
    draw_box(c, 25, 9, 3, 8, P["cobble_base"], P["O"], hi=P["cobble_hi"], sh=P["cobble_sh"])
    c.fill_rect(7, 12, 18, 1, P["t"])
    c.fill_rect(7, 15, 18, 1, P["t"])

    # Lush green leaves overflowing box
    draw_leaf_cluster(
        c,
        [
            (9.0, 8.0, 4.2, 3.2),
            (16.0, 7.0, 5.0, 3.5),
            (23.0, 8.0, 4.2, 3.2),
            (12.5, 5.5, 3.8, 2.8),
            (19.5, 5.5, 3.8, 2.8),
        ],
        deep_col=P["leaf_deep"],
    )
    # Pink & yellow urban flowers
    for fx, fy, col in (
        (8, 7, P["K"]), (9, 6, P["Q"]),
        (13, 5, P["G"]), (14, 5, P["W"]),
        (17, 7, P["K"]), (18, 6, P["Q"]),
        (21, 5, P["G"]), (24, 7, P["K"]),
    ):
        c.set_pixel(fx, fy, col)

    return c


def build_prop_neon_billboard_48x32() -> Canvas:
    """prop_neon_billboard.png (48x32) — Pop-Urban sidewalk neon sign/totem ('GIRLBUS')."""
    c = Canvas(48, 32)
    fill_ellipse(c, 24, 29.0, 19, 2.0, P["s"])

    # Twin steel legs & anchor pads (y=22..29)
    for lx in (10, 34):
        draw_box(c, lx, 22, 4, 7, P["P"], P["O"], hi=P["M"], sh=P["p"])
        draw_box(c, lx - 1, 27, 6, 3, P["M"], P["O"], sh=P["d"])

    # Outer sleek neon totem frame (x=4..43, y=2..23)
    draw_box(c, 4, 2, 40, 21, P["M"], P["O"], hi=P["W"], sh=P["d"])
    draw_box(c, 6, 4, 36, 17, P["int_dark"], P["O"])
    # Dual Neon Cyan & Shocking Pink border glow
    c.fill_rect(7, 5, 34, 1, P["X"])
    c.fill_rect(7, 5, 1, 15, P["X"])
    c.fill_rect(7, 19, 34, 1, P["K"])
    c.fill_rect(40, 5, 1, 15, P["K"])
    # Corner bulbs
    for cx, cy in ((7, 5), (40, 5), (7, 19), (40, 19)):
        c.set_pixel(cx, cy, P["W"])

    # Glowing 'GIRLBUS' text centered inside panel
    draw_text(c, 10, 8, "GIRLBUS", P["W"], shadow=P["K"], spacing=1)
    # Sharpen 'G' inner pixel so it never reads as 'O'
    c.set_pixel(12, 9, P["int_dark"])
    c.set_pixel(11, 10, P["X"])
    # Decorative neon equalizer / bus-stop stripes below text
    c.fill_rect(10, 15, 8, 2, P["X"])
    c.fill_rect(20, 15, 8, 2, P["G"])
    c.fill_rect(30, 15, 8, 2, P["K"])

    return c


# ==============================================================================
# 2. URBAN PARK & STREET PROPS
# ==============================================================================

def build_prop_tree_lush_48() -> Canvas:
    """
    prop_tree_lush.png (48x48) — Lush, rounded, multi-clustered deciduous tree
    with 1px #1C1124 outline, warm brown trunk, ground shadow & fallen leaves/flowers.
    """
    c = Canvas(48, 48)

    # Soft ground shadow
    fill_ellipse(c, 24, 42.5, 14, 3.2, P["s"])

    # Trunk & root flare (x=19..28, y=28..43)
    trunk = Canvas(48, 48)
    trunk.fill_rect(21, 28, 6, 13, P["T"])
    trunk.fill_rect(21, 28, 2, 12, P["1"])
    trunk.fill_rect(25, 28, 2, 13, P["t"])
    # Roots flare
    trunk.fill_rect(19, 40, 10, 2, P["T"])
    trunk.fill_rect(20, 39, 2, 2, P["1"])
    trunk.fill_rect(26, 39, 2, 3, P["t"])
    apply_1px_outline(trunk, P["O"])
    c.blit(trunk, 0, 0)

    # Lush multi-clustered canopy
    blobs = [
        (24.0, 11.5, 10.5, 7.5),   # Top crown
        (13.5, 18.5, 8.5, 7.5),    # Left lobe
        (34.5, 18.5, 8.5, 7.5),    # Right lobe
        (18.0, 25.0, 9.0, 7.0),    # Lower-left cluster
        (30.0, 25.0, 9.0, 7.0),    # Lower-right cluster
        (24.0, 21.0, 11.0, 8.5),   # Rich center cluster
    ]
    draw_leaf_cluster(c, blobs, base_col=P["Y"], hi_col=P["l"], sh_col=P["y"], deep_col=P["leaf_deep"])

    # Inner leaf-cluster shading accents for crisp pixel texture (zero noise)
    for lx, ly in ((18, 12), (26, 15), (14, 21), (28, 22), (22, 26)):
        c.fill_rect(lx, ly, 3, 1, P["l"])
        c.fill_rect(lx + 1, ly + 4, 3, 1, P["y"])

    # Fallen leaves & tiny flowers around the roots
    for fx, fy, col in ((14, 42, P["Y"]), (16, 44, P["l"]), (32, 43, P["W"]), (33, 43, P["G"]), (29, 44, P["K"])):
        c.set_pixel(fx, fy, col)

    return c


def build_prop_streetlamp_24x48() -> Canvas:
    """
    prop_streetlamp.png (24x48) — Tall dark-iron urban streetlamp with glowing
    Golden/Cream lantern head at top and fluted base rim at bottom.
    """
    c = Canvas(24, 48)
    fill_ellipse(c, 12, 44.5, 5.5, 1.8, P["s"])

    # Warm soft light halo behind lantern
    fill_ellipse(c, 12, 10, 8.0, 6.5, hex_to_rgba("#FFD166", 45))

    # Lantern finial & cap (y=2..6)
    c.fill_rect(11, 2, 2, 2, P["O"])
    draw_box(c, 8, 4, 8, 3, P["P"], P["O"], hi=P["M"], sh=P["p"])

    # Glowing Golden/Cream lantern glass head (y=7..14)
    draw_box(c, 7, 7, 10, 7, P["G"], P["O"], hi=P["W"], sh=P["D"])
    c.fill_rect(9, 8, 6, 5, P["C"])
    c.fill_rect(10, 9, 4, 3, P["W"])
    # Iron vertical lantern struts
    c.fill_rect(11, 7, 1, 7, P["p"])
    c.fill_rect(8, 13, 8, 1, P["P"])

    # Lantern lower collar & decorative crossbar (y=14..18)
    draw_box(c, 9, 14, 6, 3, P["P"], P["O"], hi=P["M"], sh=P["p"])
    c.fill_rect(6, 16, 12, 2, P["O"])
    c.fill_rect(7, 16, 10, 1, P["M"])

    # Tall iron pole (y=18..39)
    c.fill_rect(10, 18, 4, 21, P["O"])
    c.fill_rect(11, 18, 1, 21, P["M"])
    c.fill_rect(12, 18, 1, 21, P["p"])

    # Mid decorative ring (y=27..29)
    draw_box(c, 9, 27, 6, 3, P["P"], P["O"], hi=P["M"], sh=P["p"])

    # Fluted dark-iron base (y=38..45)
    draw_box(c, 9, 38, 6, 3, P["P"], P["O"], hi=P["M"], sh=P["p"])
    draw_box(c, 8, 41, 8, 4, P["P"], P["O"], hi=P["M"], sh=P["p"])

    return c


def build_prop_bench_plants_48x32() -> Canvas:
    """
    prop_bench_plants.png (48x32) — Warm wooden park bench flanked on both sides
    by cute terracotta potted plants (peak-npcref6.jpg).
    """
    c = Canvas(48, 32)
    fill_ellipse(c, 24, 27.5, 21, 2.5, P["s"])

    # Wooden bench (x=13..34, y=10..27)
    # Backrest iron posts
    c.fill_rect(15, 10, 2, 17, P["O"])
    c.fill_rect(31, 10, 2, 17, P["O"])
    # 2 wooden backrest slats
    draw_box(c, 13, 11, 22, 4, P["T"], P["O"], hi=P["1"], sh=P["t"])
    draw_box(c, 13, 15, 22, 4, P["T"], P["O"], hi=P["1"], sh=P["t"])
    # Wooden seat planks
    draw_box(c, 12, 19, 24, 4, P["1"], P["O"], hi=P["C"], sh=P["T"])
    c.fill_rect(13, 21, 22, 1, P["t"])
    # Iron armrests & legs
    draw_box(c, 12, 17, 3, 4, P["P"], P["O"], hi=P["M"])
    draw_box(c, 33, 17, 3, 4, P["P"], P["O"], hi=P["M"])
    draw_box(c, 14, 23, 3, 4, P["P"], P["O"], hi=P["M"], sh=P["p"])
    draw_box(c, 31, 23, 3, 4, P["P"], P["O"], hi=P["M"], sh=P["p"])

    # Left terracotta potted plant (x=3..11)
    draw_terracotta_pot(
        c,
        4,
        22,
        [(6.5, 19.0, 3.2, 2.5), (9.5, 19.5, 2.6, 2.2), (7.5, 15.5, 2.5, 2.6)],
    )
    c.set_pixel(6, 16, P["W"])  # tiny blossom

    # Right terracotta potted plant (x=37..45)
    draw_terracotta_pot(
        c,
        37,
        22,
        [(39.5, 19.5, 2.8, 2.3), (42.5, 19.0, 3.2, 2.5), (40.5, 15.5, 2.5, 2.6)],
    )
    c.set_pixel(41, 16, P["K"])  # tiny pink blossom

    return c


def build_prop_cats_clutter_32() -> Canvas:
    """
    prop_cats_clutter.png (32x32) — Cute sidewalk micro-scene from envref5.jpg & peak-envref7.jpg:
    an orange tabby cat, a grey cat, stacked cardboard parcel boxes & red/white circular street sign.
    """
    c = Canvas(32, 32)
    fill_ellipse(c, 16, 28.5, 14, 2.2, P["s"])

    # Stacked cardboard parcel boxes (top-left x=3..15, y=7..19)
    draw_box(c, 3, 13, 12, 6, P["a"], P["O"], hi=P["C"], sh=P["T"])
    c.fill_rect(6, 15, 5, 1, P["t"])
    draw_box(c, 5, 8, 9, 5, P["C"], P["O"], hi=P["W"], sh=P["a"])
    c.fill_rect(8, 9, 3, 1, P["T"])

    # Red/white circular street sign on right (x=20..29, y=3..27)
    # Pole
    c.fill_rect(24, 13, 2, 14, P["O"])
    c.fill_rect(24, 13, 1, 13, P["M"])
    c.fill_rect(23, 26, 4, 2, P["P"])
    # Circular red sign
    sign = Canvas(32, 32)
    fill_ellipse(sign, 25, 8.5, 4.5, 4.5, P["R"])
    for py in range(4, 13):
        for px in range(20, 30):
            if sign.pixels[py][px][3] > 0 and py >= 10:
                sign.set_pixel(px, py, P["r"])
    sign.fill_rect(22, 8, 6, 2, P["W"])
    apply_1px_outline(sign, P["O"])
    c.blit(sign, 0, 0)

    # Orange tabby cat facing RIGHT (x=2..12, y=21..29)
    tabby_grid = [
        ".O.O.......",
        "O1T1O....O.",
        "OT1WTO..O1O",
        "O1T11TO.OTO",
        ".O1T1T111O.",
        ".OTTTTTtO..",
        ".OWO.OOWO..",
        ".OOO..OOO..",
    ]
    c.draw_grid(tabby_grid, P, offset_x=2, offset_y=21)
    c.set_pixel(3, 23, P["O"])  # eye
    c.set_pixel(5, 23, P["O"])  # eye

    # Grey pixel cat facing LEFT (x=14..24, y=22..30)
    grey_grid = [
        ".......O.O.",
        ".O....OMdMO",
        "OdO..OMWMdO",
        "OMO.OMMdMdO",
        ".OMMMMdMdO.",
        "..OMdddddO.",
        "..OWO..OWO.",
        "..OOO..OOO.",
    ]
    c.draw_grid(grey_grid, P, offset_x=14, offset_y=21)
    c.set_pixel(20, 23, P["O"])
    c.set_pixel(22, 23, P["O"])

    return c


# ==============================================================================
# 3. THREE PARKED CIVILIAN CARS (48x32 each, matching POLICE_CAR_48 3/4 scale)
# ==============================================================================

def build_parked_car_48(variant: str) -> Canvas:
    """
    Build a 48x32 top-down 3/4 parked car facing RIGHT matching POLICE_CAR_48 scale.
    Variants: 'red' (coupe), 'yellow' (taxi/hatchback), 'mint' (compact).
    """
    if variant == "red":
        hi, base, sh = P["F"], P["R"], P["r"]
        roof_hi = P["F"]
    elif variant == "yellow":
        hi, base, sh = P["C"], P["G"], P["D"]
        roof_hi = P["W"]
    elif variant == "mint":
        hi, base, sh = P["mint_hi"], P["mint_base"], P["mint_sh"]
        roof_hi = P["3"]
    else:
        raise ValueError(f"Unknown car variant: {variant}")

    c = Canvas(48, 32)
    # Ground shadow
    fill_ellipse(c, 24, 25.5, 19, 2.2, P["s"])

    # Roof & cabin (x=13..33, y=8..14)
    if variant == "yellow":
        # Cute taxi roof sign
        draw_box(c, 21, 5, 7, 3, P["G"], P["O"], hi=P["W"])
        c.fill_rect(23, 6, 3, 1, P["O"])

    # Roof outline + fill
    roof_w = 18 if variant != "mint" else 16
    roof_x = 14 if variant != "mint" else 15
    draw_box(c, roof_x, 8, roof_w, 4, base, P["O"], hi=roof_hi)

    # Tinted windshield & side windows (y=10..13)
    win_x = roof_x - 2
    win_w = roof_w + 4
    draw_box(c, win_x, 10, win_w, 4, P["g"], P["O"], hi=hi)
    # Pillar separating side window & windshield
    c.fill_rect(win_x + win_w // 2, 10, 1, 4, P["O"])
    # Reflection streaks
    c.set_pixel(win_x + 3, 11, P["z"])
    c.set_pixel(win_x + 2, 12, P["z"])
    c.set_pixel(win_x + win_w - 5, 11, P["z"])
    c.set_pixel(win_x + win_w - 6, 12, P["z"])

    # Main car body (y=13..21)
    bx = 5 if variant != "mint" else 7
    bw = 38 if variant != "mint" else 34
    draw_box(c, bx, 13, bw, 8, base, P["O"], hi=hi, sh=sh)
    # Hood/trunk curved corners
    c.set_pixel(bx, 13, (0, 0, 0, 0))
    c.set_pixel(bx + bw - 1, 13, (0, 0, 0, 0))
    c.set_pixel(bx + 1, 13, P["O"])
    c.set_pixel(bx + bw - 2, 13, P["O"])

    # Variant details
    if variant == "red":
        # Sporty rear spoiler + crisp white racing beltline highlight
        c.fill_rect(4, 11, 4, 2, P["O"])
        c.fill_rect(5, 12, 2, 1, P["R"])
        c.fill_rect(bx + 2, 16, bw - 4, 1, P["F"])
    elif variant == "yellow":
        # Taxi checkered side stripe!
        for cx in range(14, 34, 2):
            c.set_pixel(cx, 16, P["O"])
            c.set_pixel(cx + 1, 16, P["W"])
            c.set_pixel(cx, 17, P["W"])
            c.set_pixel(cx + 1, 17, P["O"])
    elif variant == "mint":
        # Cute compact white/cream side trim
        c.fill_rect(bx + 3, 16, bw - 6, 1, P["W"])

    # Taillight (left) & Headlights (right)
    c.fill_rect(bx, 15, 2, 2, P["R"] if variant != "red" else P["G"])
    c.fill_rect(bx + bw - 2, 15, 2, 2, P["G"])
    c.set_pixel(bx + bw - 1, 15, P["W"])
    # Bumpers
    c.fill_rect(bx - 1, 18, 2, 2, P["M"])
    c.fill_rect(bx + bw - 1, 18, 2, 2, P["M"])

    # Front & Rear Tires with hubcaps (aligned with POLICE_CAR_48)
    wheel_xs = (9, 31) if variant != "mint" else (10, 29)
    for wx in wheel_xs:
        draw_box(c, wx, 19, 8, 6, P["P"], P["O"], sh=P["p"])
        # Round corners of tire
        for rx, ry in ((wx, 19), (wx + 7, 19), (wx, 24), (wx + 7, 24)):
            c.set_pixel(rx, ry, (0, 0, 0, 0))
        c.set_pixel(wx + 1, 19, P["O"])
        c.set_pixel(wx + 6, 19, P["O"])
        c.set_pixel(wx + 1, 24, P["O"])
        c.set_pixel(wx + 6, 24, P["O"])
        # Hubcap
        draw_box(c, wx + 2, 20, 4, 4, P["M"], P["W"], sh=P["d"])

    return c


# ==============================================================================
# 4. STRATIFIED SEAMLESS GROUND TILES (32x32 each)
# ==============================================================================

def build_tile_grass_park_32() -> Canvas:
    """tile_grass_park.png (32x32) — Cozy moss/leafy park grass matching peak-npcref6 & peak-envref7."""
    c = Canvas(32, 32, bg=P["grass_base"])
    # Subtle grass blade tufts (seamless coordinates)
    tufts = [
        (3, 4), (14, 2), (24, 6),
        (8, 12), (20, 14), (29, 11),
        (2, 21), (13, 24), (25, 20),
        (7, 29), (18, 28), (28, 27),
    ]
    for tx, ty in tufts:
        c.set_pixel(tx % 32, ty % 32, P["grass_sh"])
        c.set_pixel((tx + 1) % 32, (ty - 1) % 32, P["grass_hi"])
        c.set_pixel((tx + 2) % 32, ty % 32, P["grass_sh"])
        c.set_pixel((tx + 1) % 32, (ty + 1) % 32, P["grass_sh"])

    # Clover clusters & tiny cozy park wildflowers
    for cx, cy in ((10, 7), (22, 17), (5, 26), (27, 29)):
        c.set_pixel(cx, cy, P["grass_hi"])
        c.set_pixel(cx + 1, cy, P["grass_hi"])
        c.set_pixel(cx, cy + 1, P["grass_hi"])
    # 2 tiny daisy/clover blossoms
    c.set_pixel(17, 9, P["W"])
    c.set_pixel(18, 9, P["G"])
    c.set_pixel(6, 18, P["K"])
    c.set_pixel(7, 18, P["W"])
    return c


def build_tile_cobble_plaza_32() -> Canvas:
    """Low-contrast taupe setts for the plaza and pocket park borders."""
    stone = hex_to_rgba("#B4A99C")
    grout = hex_to_rgba("#A3968C")
    bevel = hex_to_rgba("#BFB3A6")
    c = Canvas(32, 32, bg=grout)
    # 4 staggered rows of 8px tall cobblestones wrapping cleanly at 32x32
    rows = [
        (0, [(0, 10), (10, 12), (22, 10)]),
        (8, [(0, 6), (6, 11), (17, 10), (27, 5)]),
        (16, [(0, 12), (12, 9), (21, 11)]),
        (24, [(0, 7), (7, 10), (17, 11), (28, 4)]),
    ]
    for ry, stones in rows:
        for sx, sw in stones:
            c.fill_rect(sx, ry, sw - 1, 7, stone)
            c.fill_rect(sx, ry, sw - 1, 1, bevel)
            c.fill_rect(sx, ry, 1, 6, bevel)
    return c


def build_tile_boardwalk_32() -> Canvas:
    """tile_boardwalk.png (32x32) — Warm wooden boardwalk planks (peak-envref4.jpg promenade)."""
    c = Canvas(32, 32, bg=P["wood_base"])
    # 4 horizontal planks (8px each), staggered vertical plank joints
    joints = [(4, 20), (12, 28), (0, 16), (8, 24)]
    for idx in range(4):
        py = idx * 8
        # Top highlight & bottom groove
        c.fill_rect(0, py, 32, 1, P["wood_hi"])
        c.fill_rect(0, py + 7, 32, 1, P["wood_sh"])
        # Subtle wood grain streaks placed safely inside plank spans
        j0, j1 = joints[idx]
        c.fill_rect((j0 + 5) % 32, py + 3, 5, 1, P["wood_sh"])
        c.fill_rect((j1 + 4) % 32, py + 5, 4, 1, P["wood_hi"])
        # Staggered vertical plank joints + nail heads
        for jx in joints[idx]:
            c.fill_rect(jx, py, 1, 8, P["wood_sh"])
            c.fill_rect((jx + 1) % 32, py + 1, 1, 6, P["wood_hi"])
            c.set_pixel((jx - 2) % 32, py + 2, P["wood_nail"])
            c.set_pixel((jx - 2) % 32, py + 5, P["wood_nail"])
            c.set_pixel((jx + 2) % 32, py + 2, P["wood_nail"])
            c.set_pixel((jx + 2) % 32, py + 5, P["wood_nail"])
    return c


def build_tile_canal_water_32() -> Canvas:
    """tile_canal_water.png (32x32) — Deep vibrant urban canal water (peak-envref4.jpg)."""
    c = Canvas(32, 32, bg=P["water_base"])
    # Smooth horizontal wave troughs
    for wy, wx_start, wlen in (
        (3, 2, 9), (3, 18, 8),
        (11, 9, 10), (11, 25, 6),
        (19, 1, 8), (19, 15, 11),
        (27, 6, 9), (27, 22, 8),
    ):
        for dx in range(wlen):
            c.set_pixel((wx_start + dx) % 32, wy, P["water_sh"])
            c.set_pixel((wx_start + dx) % 32, (wy + 1) % 32, P["water_sh"])

    # Crisp ripple highlights + specular glints
    for ry, rx_start, rlen in (
        (6, 5, 7), (6, 21, 5),
        (14, 1, 6), (14, 14, 8),
        (22, 8, 7), (22, 24, 6),
        (30, 3, 6), (30, 17, 7),
    ):
        for dx in range(rlen):
            c.set_pixel((rx_start + dx) % 32, ry, P["water_hi"])
        c.set_pixel((rx_start + 1) % 32, ry, P["water_glint"])
        c.set_pixel((rx_start + 2) % 32, ry, P["water_glint"])
    return c


# ==============================================================================
# MAIN ASSET GENERATION & VERIFICATION
# ==============================================================================

def main() -> None:
    os.makedirs(OUT_DIR, exist_ok=True)

    assets: List[Tuple[str, Canvas, int]] = [
        # 1. Storefront Buildings (96x64 & 48x40)
        ("building_konbini", build_konbini_96x64(), 6),
        ("building_coffee", build_coffee_96x64(), 6),
        ("building_arcade", build_arcade_96x64(), 6),
        ("building_ramen", build_ramen_96x64(), 6),
        ("building_bookstore", build_bookstore_96x64(), 6),
        ("building_metro", build_metro_96x64(), 6),
        ("building_newsstand", build_newsstand_48x40(), 6),
        # 2. Urban Park, Street & Canal Props
        ("prop_tree_lush", build_prop_tree_lush_48(), 6),
        ("prop_streetlamp", build_prop_streetlamp_24x48(), 6),
        ("prop_bench_plants", build_prop_bench_plants_48x32(), 6),
        ("prop_cats_clutter", build_prop_cats_clutter_32(), 8),
        ("prop_scooter", build_prop_scooter_32x24(), 8),
        ("prop_bike_rack", build_prop_bike_rack_32x20(), 8),
        ("prop_hydrant_bin", build_prop_hydrant_bin_24x24(), 8),
        ("prop_ztl_sign", build_prop_ztl_sign_24x32(), 8),
        ("prop_canal_boat", build_prop_canal_boat_48x24(), 6),
        ("prop_slot_machine", build_prop_slot_machine_32x32(), 8),
        ("prop_manhole", build_prop_manhole_24x24(), 8),
        ("prop_asphalt_crack", build_prop_asphalt_crack_32x16(), 8),
        ("prop_flower_box", build_prop_flower_box_32x20(), 8),
        ("prop_neon_billboard", build_prop_neon_billboard_48x32(), 6),
        # 3. Parked Civilian Cars (48x32)
        ("car_red", build_parked_car_48("red"), 6),
        ("car_yellow", build_parked_car_48("yellow"), 6),
        ("car_mint", build_parked_car_48("mint"), 6),
        # 4. Stratified Ground Tiles (32x32)
        ("tile_grass_park", build_tile_grass_park_32(), 6),
        ("tile_cobble_plaza", build_tile_cobble_plaza_32(), 6),
        ("tile_boardwalk", build_tile_boardwalk_32(), 6),
        ("tile_canal_water", build_tile_canal_water_32(), 6),
    ]

    for name, canvas, zoom in assets:
        save_asset_and_preview(canvas, OUT_DIR, name, zoom=zoom)
        print(f"Generated {name}.png ({canvas.width}x{canvas.height}) + preview_{name}_{zoom}x.png")

    print(f"Successfully generated all {len(assets)} map sprites (+ previews) in {OUT_DIR}")


if __name__ == "__main__":
    main()
