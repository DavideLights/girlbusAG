#!/usr/bin/env python3
"""
GIRLBUS: NEON STOP (girlbusAG) — Hero Storefronts & Landmarks Pixel-Art Generator (v2).

Generates 5 expressive, Pop-Urban Tokyo-meets-Roma hero map sprites (1x native RGBA PNGs)
directly into assets/sprites/map/ (plus zoomed nearest-neighbor previews in tools/previews/map/):
  1. building_pizzeria.png  (96x64) — 24h Neon Pizzeria ("PIZZA")
  2. building_vinyl_bar.png (96x64) — Synthwave Vinyl & Bubble-Tea Lounge ("HI-FI")
  3. building_gelateria.png (96x64) — Pastel Mint & Strawberry Gelato Parlor ("GELATO")
  4. prop_fountain.png      (64x48) — Central Plaza Octagonal Neon Fountain
  5. prop_arch_gate.png     (96x40) — Pop-Urban Neon District Street Arch ("NEON STOP")

Strictly adheres to 03_ART_STYLE_AND_SPRITE_BIBLE.md:
  - Universal 1px Deep Aubergine (#1C1124) outlines (zero pure #000000 outlines)
  - 3-tone Top-Left cel-pixel shading (zero muddy dithering)
  - Vibrant Pop-Urban Tokyo/Roma palette harmonized with master PALETTE & MAP_PALETTE
"""

import math
import os
import sys
from typing import Dict, List, Optional, Tuple

TOOLS_DIR = os.path.dirname(os.path.abspath(__file__))
ROOT_DIR = os.path.dirname(TOOLS_DIR)
if TOOLS_DIR not in sys.path:
    sys.path.insert(0, TOOLS_DIR)

from pixel_canvas import Canvas, PALETTE, RGBA, hex_to_rgba, save_asset_and_preview
from gen_layered_map_assets import (
    FONT_3X5,
    MAP_PALETTE,
    OUT_DIR,
    apply_1px_outline,
    draw_ac_unit,
    draw_box,
    draw_glass_window,
    draw_leaf_cluster,
    draw_striped_awning,
    draw_terracotta_pot,
    fill_ellipse,
)

# ==============================================================================
# EXTENDED HERO PALETTE (Harmonized with Master PALETTE & MAP_PALETTE)
# ==============================================================================
HERO_PALETTE: Dict[str, RGBA] = {
    **PALETTE,
    **MAP_PALETTE,
    # Warm Roman Terracotta & Ochre Stucco (Pizzeria)
    "terra_hi": hex_to_rgba("#EB8359"),
    "terra_base": hex_to_rgba("#C85538"),
    "terra_sh": hex_to_rgba("#8E3422"),
    "stucco_hi": hex_to_rgba("#FFF5DF"),
    "stucco_base": hex_to_rgba("#F2DDBA"),
    "stucco_sh": hex_to_rgba("#D4B488"),
    # Wood-Oven Fire & Amber Bar Glow
    "oven_core": hex_to_rgba("#FFF8CC"),
    "oven_gold": hex_to_rgba("#FFCB47"),
    "oven_orange": hex_to_rgba("#FF6F20"),
    "oven_red": hex_to_rgba("#C93418"),
    "oven_glow": hex_to_rgba("#FFAE33", 68),
    "amber_hi": hex_to_rgba("#FFE9A8"),
    "amber_base": hex_to_rgba("#F2A63B"),
    "amber_sh": hex_to_rgba("#B5621B"),
    "amber_deep": hex_to_rgba("#6E3214"),
    "amber_glow": hex_to_rgba("#FFB84D", 55),
    # Synthwave Slate-Purple Brick & Magenta Neon (Vinyl Bar & Street Arch)
    "slate_hi": hex_to_rgba("#594778"),
    "slate_base": hex_to_rgba("#3D2F57"),
    "slate_sh": hex_to_rgba("#271D3B"),
    "slate_deep": hex_to_rgba("#1C142B"),
    "neon_mag": hex_to_rgba("#FF2A85"),
    "neon_mag_hi": hex_to_rgba("#FF94C8"),
    "neon_mag_sh": hex_to_rgba("#A81256"),
    "neon_mag_glow": hex_to_rgba("#FF2A85", 62),
    "neon_cyan_glow": hex_to_rgba("#00F0FF", 62),
    "neon_gold_glow": hex_to_rgba("#FFD166", 65),
    # Pastel Mint & Strawberry Gelateria Palette
    "pmint_hi": hex_to_rgba("#C9FFF0"),
    "pmint_base": hex_to_rgba("#8BF0D5"),
    "pmint_sh": hex_to_rgba("#52C2A8"),
    "berry_dark": hex_to_rgba("#3D1B3B"),
    "ice_glass": hex_to_rgba("#DCF9FF"),
    "ice_back": hex_to_rgba("#AEEFFF"),
    "gel_pistachio": hex_to_rgba("#7BE07B"),
    "gel_lemon": hex_to_rgba("#FFE95C"),
    "gel_mango": hex_to_rgba("#FF9642"),
    "gel_ube": hex_to_rgba("#A86AD1"),
    # Roman Travertine Stone & Turquoise Fountain Water
    "trav_hi": hex_to_rgba("#FAF1DF"),
    "trav_base": hex_to_rgba("#E0D1B6"),
    "trav_sh": hex_to_rgba("#B5A284"),
    "trav_deep": hex_to_rgba("#857259"),
    "turq_deep": hex_to_rgba("#0D7B94"),
    "turq_base": hex_to_rgba("#1CC8D4"),
    "turq_hi": hex_to_rgba("#63F5FF"),
    "turq_glow": hex_to_rgba("#00F0FF", 95),
    "foam": hex_to_rgba("#F2FFFF"),
}

P = HERO_PALETTE

FONT_EXT: Dict[str, List[str]] = {
    **FONT_3X5,
    # Crisp 4px-wide G and N so 'GELATO' and 'NEON STOP' read with 100% arcade clarity
    "G": ["0111", "1000", "1011", "1001", "0111"],
    "N": ["1001", "1101", "1011", "1001", "1001"],
    "-": ["000", "000", "111", "000", "000"],
    "2": ["110", "001", "010", "100", "111"],
    "4": ["101", "101", "111", "001", "001"],
    "0": ["010", "101", "101", "101", "010"],
}


def draw_text_ext(
    c: Canvas,
    x: int,
    y: int,
    text: str,
    color: RGBA,
    shadow: Optional[RGBA] = None,
    spacing: int = 1,
) -> None:
    """Render crisp 3x5/4x5 pixel text using the extended font table."""
    cx = x
    for ch in text.upper():
        if ch == " ":
            cx += 2
            continue
        glyph = FONT_EXT.get(ch, FONT_EXT[" "])
        gw = len(glyph[0])
        for gy, row in enumerate(glyph):
            for gx, bit in enumerate(row):
                if bit == "1":
                    if shadow is not None:
                        c.set_pixel(cx + gx, y + gy + 1, shadow)
                    c.set_pixel(cx + gx, y + gy, color)
        cx += gw + spacing


def draw_pixel_line(c: Canvas, x0: int, y0: int, x1: int, y1: int, color: RGBA) -> None:
    """Bresenham 1px crisp pixel line."""
    dx = abs(x1 - x0)
    dy = -abs(y1 - y0)
    sx = 1 if x0 < x1 else -1
    sy = 1 if y0 < y1 else -1
    err = dx + dy
    x, y = x0, y0
    while True:
        c.set_pixel(x, y, color)
        if x == x1 and y == y1:
            break
        e2 = 2 * err
        if e2 >= dy:
            err += dy
            x += sx
        if e2 <= dx:
            err += dx
            y += sy


def draw_fairy_string(
    c: Canvas,
    x0: int,
    x1: int,
    y_top: int,
    sag: float,
    bulb_colors: List[RGBA],
    step: int = 5,
) -> None:
    """Draw a catenary festoon wire with glowing Pop-Urban fairy light bulbs."""
    span = max(1, x1 - x0)
    # Draw wire
    prev_y = y_top
    for px in range(x0, x1 + 1):
        t = (px - x0) / span
        py = int(round(y_top + 4.0 * sag * t * (1.0 - t)))
        c.set_pixel(px, py, P["O"])
        if abs(py - prev_y) > 1:
            c.set_pixel(px, (py + prev_y) // 2, P["O"])
        prev_y = py
    # Draw hanging glowing bulbs
    idx = 0
    for px in range(x0 + 3, x1 - 2, step):
        t = (px - x0) / span
        py = int(round(y_top + 4.0 * sag * t * (1.0 - t))) + 1
        bcol = bulb_colors[idx % len(bulb_colors)]
        halo = (bcol[0], bcol[1], bcol[2], 65)
        fill_ellipse(c, px + 0.5, py + 0.5, 2.2, 2.2, halo)
        c.set_pixel(px, py, P["W"])
        c.set_pixel(px, py + 1, bcol)
        idx += 1


# ==============================================================================
# 1. BUILDING: PIZZERIA (96x64) — Roma-meets-Tokyo 24h Neon Pizzeria ("PIZZA")
# ==============================================================================

def build_pizzeria_96x64() -> Canvas:
    """
    building_pizzeria.png (96x64):
    - Terracotta Roman tiled roof trim, brick chimney with warm wisps & rooftop AC unit
    - Upper warm cream stucco wall with terracotta shutters & glowing neon '24H' badge
    - Hero dark-framed neon 'PIZZA' signboard with pixel pizza-slice icon & side Tokyo blade sign
    - Scalloped warm terracotta & cream striped awning
    - Ground floor: glowing wood-fired brick oven window w/ orange/gold fire glow,
      cozy entrance door, sidewalk chalkboard menu, stacked pizza boxes & tiny basil pot
    """
    c = Canvas(96, 64)
    fill_ellipse(c, 48, 62.5, 45, 2.0, P["s"])

    # --- BRICK WOOD-OVEN CHIMNEY & STEAM WISPS (top-left x=15..24, y=0..6) ---
    draw_box(c, 16, 2, 7, 4, P["terra_base"], P["O"], hi=P["terra_hi"], sh=P["terra_sh"])
    draw_box(c, 15, 1, 9, 3, P["stucco_base"], P["O"], hi=P["W"], sh=P["stucco_sh"])
    # Subtle warm smoke wisps above chimney
    c.set_pixel(19, 0, hex_to_rgba("#FFF5DF", 180))
    c.set_pixel(20, 0, hex_to_rgba("#FFF5DF", 120))

    # The recess on the right gives the terracotta roof a visible return.
    c.fill_rect(82, 9, 9, 52, P["O"])
    c.fill_rect(83, 10, 7, 49, P["stucco_sh"])
    c.fill_rect(86, 12, 4, 45, P["terra_sh"])
    c.fill_rect(85, 44, 3, 14, P["terra_base"])
    for y in (14, 19, 45, 50, 55):
        c.fill_rect(87, y, 3, 1, P["terra_hi"])
    c.fill_rect(81, 3, 7, 7, P["O"])
    c.fill_rect(83, 4, 6, 6, P["terra_sh"])
    c.fill_rect(86, 5, 5, 6, P["O"])
    c.fill_rect(87, 6, 3, 3, P["terra_base"])

    # --- TERRACOTTA ROMAN TILED ROOF & CORNICE TRIM (x=10..84, y=4..10) ---
    draw_box(c, 10, 4, 75, 5, P["terra_base"], P["O"], hi=P["terra_hi"], sh=P["terra_sh"])
    # Scalloped Mediterranean barrel roof tiles
    for rx in range(11, 83, 4):
        c.fill_rect(rx, 5, 2, 3, P["terra_hi"])
        c.set_pixel(rx + 2, 5, P["terra_base"])
        c.fill_rect(rx + 3, 5, 1, 3, P["terra_sh"])
        c.set_pixel(rx + 1, 8, P["O"])
    # Warm cream stucco cornice under tiles
    draw_box(c, 10, 8, 75, 3, P["stucco_base"], P["O"], hi=P["stucco_hi"], sh=P["stucco_sh"])

    # --- UPPER FLOOR ROMAN-TOKYO STUCCO WALL (x=12..83, y=10..24) ---
    draw_box(c, 12, 10, 71, 14, P["stucco_base"], P["O"], hi=P["stucco_hi"], sh=P["stucco_sh"])
    # Subtle exposed terracotta brick accents in plaster
    for bx, by in ((14, 12), (43, 21), (46, 21), (79, 12)):
        c.fill_rect(bx, by, 2, 1, P["terra_base"])
        c.set_pixel(bx + 2, by, P["stucco_sh"])

    # Upper arched warm window with Roman terracotta shutters (x=17..41, y=12..22)
    draw_box(c, 17, 13, 5, 9, P["terra_base"], P["O"], hi=P["terra_hi"], sh=P["terra_sh"])
    draw_box(c, 36, 13, 5, 9, P["terra_base"], P["O"], hi=P["terra_hi"], sh=P["terra_sh"])
    for sy in (15, 17, 19):
        c.fill_rect(18, sy, 3, 1, P["terra_sh"])
        c.fill_rect(37, sy, 3, 1, P["terra_sh"])
    draw_box(c, 22, 12, 14, 10, P["amber_base"], P["O"], hi=P["amber_hi"], sh=P["amber_sh"])
    c.fill_rect(23, 13, 12, 8, P["amber_hi"])
    c.fill_rect(23, 18, 12, 3, P["amber_base"])
    # Mullions
    c.fill_rect(28, 13, 2, 8, P["T"])
    c.fill_rect(23, 16, 12, 1, P["T"])

    # Glowing Tokyo-style '24H' mini neon badge with crisp 1px dark margin (x=44..61, y=12..21)
    fill_ellipse(c, 52.5, 16.5, 9.0, 4.8, P["neon_cyan_glow"])
    draw_box(c, 44, 12, 17, 9, P["int_dark"], P["O"])
    c.fill_rect(45, 13, 15, 1, P["X"])
    c.fill_rect(45, 19, 15, 1, P["2"])
    draw_text_ext(c, 47, 14, "24H", P["W"], shadow=None, spacing=1)

    # Rooftop / Upper wall AC unit + trailing basil/ivy vine (x=64..79, y=12..21)
    draw_ac_unit(c, 64, 12, rusty=True)
    draw_leaf_cluster(c, [(78, 15, 2.2, 2.4), (79, 19, 2.2, 2.4)])

    # --- HERO 'PIZZA' NEON SIGNBOARD (x=10..84, y=23..35) ---
    fill_ellipse(c, 47, 29, 36, 6.5, P["oven_glow"])
    draw_box(c, 10, 23, 75, 12, P["terra_sh"], P["O"], hi=P["terra_hi"], sh=P["int_dark"])
    draw_box(c, 12, 24, 71, 10, P["int_dark"], P["O"])
    # Glowing dual neon trim (coral-orange top, gold bottom)
    c.fill_rect(13, 25, 69, 1, P["oven_orange"])
    c.fill_rect(13, 32, 69, 1, P["oven_gold"])
    c.set_pixel(13, 25, P["W"])
    c.set_pixel(81, 25, P["W"])
    c.set_pixel(13, 32, P["W"])
    c.set_pixel(81, 32, P["W"])

    # Pixel Pizza Slice Icon on left of signboard (x=16..24, y=26..32)
    c.fill_rect(16, 26, 7, 1, P["1"])          # golden crust
    c.fill_rect(17, 27, 5, 1, P["oven_gold"])  # melted cheese
    c.fill_rect(18, 28, 4, 1, P["oven_gold"])
    c.fill_rect(19, 29, 3, 1, P["oven_gold"])
    c.fill_rect(20, 30, 2, 1, P["oven_gold"])
    c.set_pixel(20, 31, P["oven_gold"])        # slice tip
    c.set_pixel(18, 27, P["R"])                # pepperoni 1
    c.set_pixel(20, 28, P["R"])                # pepperoni 2
    c.set_pixel(21, 27, P["Y"])                # basil dot

    # Crisp glowing 'PIZZA' neon lettering (centered x=31..64, y=26..31)
    draw_text_ext(c, 31, 26, "PIZZA", P["oven_core"], shadow=P["oven_orange"], spacing=2)

    # Italian-Tokyo tricolore neon emblem bars on right of signboard (x=70..79, y=27..31)
    c.fill_rect(70, 27, 2, 4, P["l"])
    c.fill_rect(73, 27, 2, 4, P["W"])
    c.fill_rect(76, 27, 2, 4, P["F"])

    # Vertical Tokyo neon blade sign protruding on right (x=85..92, y=24..40)
    c.fill_rect(84, 26, 2, 2, P["O"])
    c.fill_rect(84, 36, 2, 2, P["O"])
    draw_box(c, 85, 24, 7, 16, P["R"], P["O"], hi=P["F"], sh=P["r"])
    c.fill_rect(86, 25, 5, 14, P["int_dark"])
    # Vertical glowing neon dashes / katakana-style strokes
    c.fill_rect(87, 26, 3, 2, P["oven_gold"])
    c.fill_rect(87, 29, 3, 1, P["W"])
    c.fill_rect(87, 31, 3, 2, P["F"])
    c.fill_rect(87, 34, 3, 1, P["X"])
    c.fill_rect(87, 36, 3, 2, P["oven_gold"])

    # --- WARM TERRACOTTA & CREAM SCALLOPED AWNING (x=7..85, y=35..44) ---
    draw_striped_awning(
        c, 7, 35, 78, 10, P["terra_base"], P["terra_sh"], P["stucco_hi"], P["stucco_sh"], stripe_w=5
    )

    # --- GROUND FLOOR PIZZERIA STOREFRONT (x=11..83, y=44..62) ---
    draw_box(c, 11, 44, 73, 18, P["stucco_base"], P["O"], hi=P["stucco_hi"], sh=P["stucco_sh"])
    # Terracotta wainscoting base trim (y=58..61)
    draw_box(c, 11, 58, 73, 4, P["terra_base"], P["O"], hi=P["terra_hi"], sh=P["terra_sh"])

    # --- GLOWING WOOD-OVEN WINDOW (x=14..46, y=45..58) ---
    # Warm exterior light spill on stucco wall around window
    fill_ellipse(c, 30, 52, 18, 7.5, P["oven_glow"])
    draw_box(c, 14, 45, 32, 13, P["t"], P["O"], hi=P["T"], sh=P["int_dark"])
    # Warm dark interior behind wood-fired oven
    c.fill_rect(15, 46, 30, 11, P["amber_deep"])
    # Stepped brick arch above the wood fire.
    c.fill_rect(18, 51, 20, 5, P["terra_sh"])
    c.fill_rect(19, 49, 18, 7, P["terra_base"])
    c.fill_rect(21, 48, 14, 2, P["terra_hi"])
    c.fill_rect(24, 47, 8, 1, P["stucco_sh"])
    c.fill_rect(22, 52, 12, 4, P["oven_red"])
    c.fill_rect(24, 50, 8, 6, P["amber_deep"])
    c.fill_rect(25, 52, 6, 4, P["oven_orange"])
    c.fill_rect(24, 54, 8, 2, P["oven_gold"])
    c.fill_rect(27, 53, 3, 3, P["oven_core"])
    c.set_pixel(26, 51, P["oven_gold"])
    c.set_pixel(29, 52, P["oven_gold"])
    c.set_pixel(24, 50, P["oven_orange"])
    c.set_pixel(30, 51, P["oven_orange"])
    # Wooden pizza peel resting on right side of oven window
    c.fill_rect(39, 53, 5, 2, P["1"])
    c.fill_rect(41, 49, 1, 5, P["T"])
    # Soft translucent warm glass sheen
    sheen = hex_to_rgba("#FFF8CC", 170)
    for i in range(4):
        c.set_pixel(16 + i, 54 - i, sheen)
        c.set_pixel(40 + i, 50 - i, sheen)
    # Takeaway wooden counter shelf & fresh hot pizza box on ledge
    draw_box(c, 13, 56, 34, 3, P["T"], P["O"], hi=P["1"], sh=P["t"])
    draw_box(c, 34, 54, 9, 3, P["W"], P["O"], sh=P["w"])
    c.fill_rect(36, 55, 5, 1, P["R"])

    # --- COZY GLASS ENTRANCE DOOR (x=48..65, y=45..61) ---
    draw_box(c, 48, 45, 17, 17, P["terra_sh"], P["O"], hi=P["terra_base"], sh=P["int_dark"])
    draw_box(c, 50, 47, 13, 10, P["amber_base"], P["O"], hi=P["amber_hi"], sh=P["amber_sh"])
    c.fill_rect(51, 48, 11, 8, P["amber_hi"])
    c.fill_rect(51, 53, 11, 3, P["amber_base"])
    # Diagonal glass reflection
    c.set_pixel(52, 51, P["W"])
    c.set_pixel(53, 50, P["W"])
    c.set_pixel(54, 49, P["W"])
    # Cute hanging 'OPEN' plate on door
    draw_box(c, 53, 50, 7, 4, P["R"], P["O"], hi=P["F"])
    c.fill_rect(54, 51, 5, 1, P["W"])
    # Brass handle & stone threshold
    c.fill_rect(50, 58, 2, 1, P["G"])
    c.fill_rect(49, 60, 15, 2, P["cobble_hi"])

    # --- SIDEWALK CHALKBOARD MENU (x=68..79, y=48..61) ---
    draw_box(c, 68, 48, 11, 12, P["T"], P["O"], hi=P["1"], sh=P["t"])
    c.fill_rect(69, 49, 9, 9, P["g"])
    # Chalk lettering lines & mini pizza sketch
    c.fill_rect(70, 50, 7, 1, P["W"])           # header
    c.fill_rect(70, 52, 5, 1, P["oven_gold"])   # margherita
    c.fill_rect(76, 52, 1, 1, P["R"])
    c.fill_rect(70, 54, 6, 1, P["F"])           # diavola
    c.fill_rect(70, 56, 4, 1, P["l"])           # basilico
    c.set_pixel(76, 55, P["oven_gold"])         # mini pizza dot
    c.set_pixel(76, 56, P["R"])
    # A-frame legs
    c.fill_rect(69, 60, 2, 2, P["O"])
    c.fill_rect(76, 60, 2, 2, P["O"])

    # --- STACKED PIZZA BOXES OUTSIDE ON FAR LEFT (x=2..12, y=49..62) ---
    box_specs = [
        (2, 58, 10, 4, P["W"], P["w"]),
        (3, 55, 10, 4, P["C"], P["a"]),
        (2, 52, 10, 4, P["W"], P["w"]),
        (3, 49, 9, 4, P["C"], P["a"]),
    ]
    for bx, by, bw, bh, bfill, bsh in box_specs:
        draw_box(c, bx, by, bw, bh, bfill, P["O"], hi=P["W"], sh=bsh)
        # Red & green pizzeria stripe on box front edge
        c.fill_rect(bx + 2, by + 1, 2, 1, P["R"])
        c.fill_rect(bx + 5, by + 1, 2, 1, P["Y"])

    # --- TINY BASIL POT OUTSIDE ON FAR RIGHT (x=82..92, y=50..62) ---
    draw_terracotta_pot(
        c,
        83,
        56,
        [
            (85.0, 54.0, 2.8, 2.5),
            (88.5, 54.0, 2.6, 2.3),
            (86.5, 51.5, 2.5, 2.2),
        ],
    )
    # Tiny bright white basil flower buds
    c.set_pixel(86, 50, P["W"])
    c.set_pixel(88, 52, P["l"])

    return c


# ==============================================================================
# 2. BUILDING: VINYL BAR (96x64) — Retro Synthwave Vinyl & Bubble-Tea ("HI-FI")
# ==============================================================================

def build_vinyl_bar_96x64() -> Canvas:
    """
    building_vinyl_bar.png (96x64):
    - Deep slate-purple brick facade with Art-Deco/Tokyo cornice & string fairy lights
    - Glowing electric cyan & coral neon vinyl record emblem on upper-left facade
    - Synthwave 'HI-FI' neon marquee with glowing equalizer bars & tiny boba tea icon
    - Wide warm amber bar window with silhouettes of liquor/syrup bottles, vinyl sleeves & pendant lamps
    - Recessed lounge doorway flanked by a lush potted Monstera plant & vinyl crate
    """
    c = Canvas(96, 64)
    fill_ellipse(c, 48, 62.5, 45, 2.0, P["s"])

    # Stepped stone return, with thin brass reveals echoing the marquee frame.
    c.fill_rect(84, 9, 7, 51, P["O"])
    c.fill_rect(85, 10, 5, 49, P["slate_sh"])
    c.fill_rect(89, 12, 2, 44, P["slate_deep"])
    for y in (14, 19, 37):
        c.fill_rect(85, y, 4, 1, P["slate_hi"])

    # --- ROOFTOP CORNICE & ART-DECO CROWN (x=9..87, y=2..11) ---
    # Art Deco center parapet with a muted brass cap and short cyan inset.
    draw_box(c, 33, 2, 30, 5, P["slate_base"], P["O"], hi=P["slate_hi"], sh=P["slate_sh"])
    c.fill_rect(33, 3, 30, 1, P["T"])
    c.fill_rect(38, 1, 20, 2, P["slate_hi"])
    c.fill_rect(41, 0, 14, 1, P["O"])
    c.fill_rect(41, 4, 14, 1, P["2"])
    # Rooftop small turntable-styled vent box on left
    draw_box(c, 14, 3, 11, 5, P["P"], P["O"], hi=P["M"], sh=P["p"])
    c.fill_rect(16, 4, 7, 2, P["g"])
    c.set_pixel(18, 5, P["neon_mag"])
    # Main roof cornice beam
    draw_box(c, 9, 6, 78, 4, P["P"], P["O"], hi=P["M"], sh=P["p"])
    c.fill_rect(10, 8, 76, 1, P["slate_hi"])

    # --- DEEP SLATE-PURPLE BRICK FACADE (x=10..86, y=10..62) ---
    draw_box(c, 10, 10, 76, 52, P["slate_base"], P["O"], hi=P["slate_hi"], sh=P["slate_sh"])
    # Crisp staggered brick texture clusters (clean, non-noisy)
    for by in range(13, 60, 4):
        offset = 3 if (by // 4) % 2 == 0 else 0
        for bx in range(13 + offset, 82, 9):
            c.fill_rect(bx, by, 4, 1, P["slate_sh"])
            if (bx + by) % 5 == 0:
                c.fill_rect(bx, by - 1, 3, 1, P["slate_hi"])

    # Vertical architectural side pilasters
    draw_box(c, 10, 10, 4, 52, P["slate_sh"], P["O"], hi=P["slate_hi"], sh=P["slate_deep"])
    draw_box(c, 82, 10, 4, 52, P["slate_sh"], P["O"], hi=P["slate_hi"], sh=P["slate_deep"])

    # --- STRING FAIRY LIGHTS ALONG CORNICE (y=9..14, x=11..85) ---
    fairy_cols = [P["G"], P["F"], P["C"]]
    for start, end in ((11, 36), (36, 61), (61, 85)):
        draw_fairy_string(c, start, end, 9, 2.8, fairy_cols, step=7)

    # --- GLOWING CYAN & CORAL NEON VINYL RECORD EMBLEM (cx=26, cy=25) ---
    fill_ellipse(c, 26, 25, 11.5, 10.5, P["neon_cyan_glow"])
    fill_ellipse(c, 26, 25, 5.5, 5.0, P["neon_mag_glow"])
    vinyl = Canvas(96, 64)
    # Outer neon cyan glowing rim
    fill_ellipse(vinyl, 26.0, 25.0, 9.2, 8.8, P["X"])
    # Dark glossy vinyl disc body
    fill_ellipse(vinyl, 26.0, 25.0, 8.0, 7.6, P["int_dark"])
    # Concentric cyan/indigo vinyl groove highlights (top-left & bottom-right arcs)
    for angle_deg in range(200, 265, 6):
        rad = math.radians(angle_deg)
        vinyl.set_pixel(int(round(26 + 6.0 * math.cos(rad))), int(round(25 + 5.7 * math.sin(rad))), P["3"])
        vinyl.set_pixel(int(round(26 + 4.6 * math.cos(rad))), int(round(25 + 4.3 * math.sin(rad))), P["2"])
    for angle_deg in range(20, 85, 6):
        rad = math.radians(angle_deg)
        vinyl.set_pixel(int(round(26 + 6.0 * math.cos(rad))), int(round(25 + 5.7 * math.sin(rad))), P["2"])
    # Coral-pink neon center record label + white spindle hole
    fill_ellipse(vinyl, 26.0, 25.0, 3.3, 3.1, P["F"])
    fill_ellipse(vinyl, 25.5, 24.5, 2.0, 1.8, P["neon_mag_hi"])
    vinyl.set_pixel(26, 25, P["O"])
    vinyl.set_pixel(25, 24, P["W"])
    apply_1px_outline(vinyl, P["O"])
    c.blit(vinyl, 0, 0)
    # Sleek chrome tonearm resting across right edge of record
    draw_box(c, 34, 17, 3, 3, P["M"], P["O"], hi=P["W"])
    draw_pixel_line(c, 35, 19, 31, 26, P["W"])
    draw_pixel_line(c, 36, 19, 32, 26, P["d"])
    c.set_pixel(31, 26, P["neon_mag"])

    # --- SYNTHWAVE 'HI-FI' NEON MARQUEE SIGNBOARD (x=41..81, y=16..32) ---
    fill_ellipse(c, 61, 24, 21, 8.5, P["neon_mag_glow"])
    draw_box(c, 41, 16, 40, 16, P["slate_deep"], P["O"], hi=P["U"], sh=P["int_dark"])
    # Neon glowing double border (cyan top/left, coral-magenta bottom/right)
    c.fill_rect(42, 17, 38, 1, P["X"])
    c.fill_rect(42, 17, 1, 14, P["X"])
    c.fill_rect(42, 30, 38, 1, P["F"])
    c.fill_rect(79, 17, 1, 14, P["F"])
    c.set_pixel(42, 17, P["W"])
    c.set_pixel(79, 30, P["W"])

    # Crisp glowing 'HI-FI' lettering (x=46..69, y=19..24)
    draw_text_ext(c, 46, 19, "HI-FI", P["W"], shadow=P["neon_mag"], spacing=1)
    # Glowing cyan accent bar + neon audio equalizer under HI-FI (y=26..29)
    eq_bars = [(46, 2, P["X"]), (49, 3, P["F"]), (52, 1, P["G"]), (55, 3, P["X"]),
               (58, 2, P["neon_mag"]), (61, 3, P["G"]), (64, 2, P["X"]), (67, 1, P["F"])]
    for ex, eh, ecol in eq_bars:
        c.fill_rect(ex, 29 - eh, 2, eh, ecol)

    # Tiny glowing Bubble-Tea cup icon on right of sign (x=72..78, y=19..29)
    c.fill_rect(75, 19, 1, 3, P["X"])          # straw
    c.fill_rect(72, 22, 6, 1, P["W"])          # cup lid
    c.fill_rect(73, 23, 4, 5, P["C"])          # milk tea body
    c.fill_rect(73, 25, 4, 3, P["F"])          # pink swirl
    c.set_pixel(73, 27, P["O"])                # boba pearl 1
    c.set_pixel(75, 27, P["O"])                # boba pearl 2
    c.set_pixel(74, 26, P["O"])                # boba pearl 3

    # --- SLEEK MID-FACADE NEON CANOPY BAR (x=8..88, y=35..40) ---
    draw_box(c, 8, 35, 80, 5, P["P"], P["O"], hi=P["M"], sh=P["p"])
    c.fill_rect(10, 38, 76, 1, P["X"])         # glowing cyan under-trim
    for lx in range(14, 82, 8):
        c.set_pixel(lx, 38, P["W"])
        c.set_pixel(lx + 4, 38, P["neon_mag"])

    # --- WARM AMBER BAR WINDOW WITH BOTTLES & RECORDS SILHOUETTES (x=14..55, y=41..59) ---
    fill_ellipse(c, 34.5, 50, 22, 9.0, P["amber_glow"])
    draw_box(c, 14, 41, 41, 18, P["P"], P["O"], hi=P["M"], sh=P["p"])
    # Warm glowing amber lounge interior gradient
    c.fill_rect(16, 43, 37, 14, P["amber_base"])
    c.fill_rect(16, 43, 37, 5, P["amber_hi"])
    c.fill_rect(16, 53, 37, 4, P["amber_sh"])

    # Hanging retro pendant bar lamps inside top of window
    for px in (23, 35, 46):
        c.fill_rect(px, 43, 1, 2, P["O"])
        c.fill_rect(px - 1, 45, 3, 1, P["W"])
        c.set_pixel(px, 46, P["oven_core"])

    # Upper bar shelf & silhouettes of tall bottles (y=47..51)
    c.fill_rect(16, 50, 37, 1, P["amber_deep"])
    bottle_specs = [
        (18, 4, 2, P["slate_deep"]), (21, 3, 2, P["neon_mag_sh"]), (24, 4, 2, P["slate_deep"]),
        (28, 3, 2, P["2"]), (31, 4, 2, P["slate_deep"]), (34, 3, 2, P["terra_sh"]),
        (38, 4, 2, P["slate_deep"]), (41, 3, 2, P["U"]), (44, 4, 2, P["slate_deep"]),
        (48, 3, 2, P["neon_mag_sh"]), (50, 4, 2, P["slate_deep"]),
    ]
    for bx, bh, bw, bcol in bottle_specs:
        c.fill_rect(bx, 50 - bh, bw, bh, bcol)
        c.set_pixel(bx, 50 - bh - 1, P["O"])  # bottle neck

    # Lower bar counter & standing vinyl record sleeves + turntable silhouette (y=52..57)
    c.fill_rect(16, 55, 37, 2, P["slate_deep"])
    # Vinyl record sleeves lined up on lower counter
    sleeve_cols = [P["F"], P["X"], P["G"], P["K"], P["W"], P["U"]]
    for idx, sx in enumerate(range(18, 36, 3)):
        c.fill_rect(sx, 52, 2, 3, sleeve_cols[idx % len(sleeve_cols)])
        c.fill_rect(sx + 2, 52, 1, 3, P["O"])
    # Turntable & speaker silhouette on right of bar counter
    c.fill_rect(38, 53, 8, 2, P["O"])
    c.fill_rect(39, 52, 6, 1, P["M"])
    c.set_pixel(40, 54, P["X"])  # power LED
    c.fill_rect(48, 51, 4, 4, P["O"])
    c.fill_rect(49, 52, 2, 2, P["slate_hi"])

    # Window mullion & soft translucent glass reflection streaks
    c.fill_rect(34, 43, 2, 14, P["P"])
    glass_glint = hex_to_rgba("#FFFFFF", 130)
    for i in range(5):
        c.set_pixel(17 + i, 48 - i, glass_glint)
        c.set_pixel(37 + i, 48 - i, glass_glint)

    # --- STYLISH RECESSED LOUNGE DOORWAY (x=58..76, y=41..61) ---
    draw_box(c, 58, 41, 18, 21, P["slate_deep"], P["O"], hi=P["M"], sh=P["int_dark"])
    draw_box(c, 60, 43, 14, 18, P["U"], P["O"], hi=P["Z"], sh=P["u"])
    # Glowing amber-coral tall retro porthole window inside door
    draw_box(c, 63, 45, 8, 9, P["amber_base"], P["O"], hi=P["amber_hi"], sh=P["amber_sh"])
    c.fill_rect(64, 46, 6, 4, P["amber_hi"])
    c.set_pixel(65, 47, P["W"])
    c.set_pixel(66, 46, P["W"])
    # Tiny cyan 'OPEN' neon bar & chrome door handle
    c.fill_rect(64, 51, 6, 2, P["X"])
    c.fill_rect(65, 51, 4, 1, P["W"])
    c.fill_rect(61, 55, 2, 2, P["G"])
    c.fill_rect(61, 59, 12, 2, P["M"])  # chrome kickplate

    # --- LUSH POTTED MONSTERA BY DOORWAY (x=75..94, y=43..62) ---
    # Stylish ribbed cream/teal ceramic planter pot (x=79..90, y=54..62)
    draw_box(c, 79, 54, 11, 3, P["C"], P["O"], hi=P["W"], sh=P["a"])
    draw_box(c, 80, 56, 9, 6, P["teal_base"], P["O"], hi=P["teal_hi"], sh=P["teal_sh"])
    c.fill_rect(82, 57, 1, 4, P["teal_hi"])
    c.fill_rect(86, 57, 1, 4, P["teal_sh"])
    # Dramatic large Monstera split leaves!
    monstera_blobs = [
        (80.5, 50.5, 4.4, 3.6),
        (88.5, 50.5, 4.4, 3.6),
        (84.5, 46.5, 4.8, 3.8),
        (78.2, 46.8, 3.2, 2.6),
        (90.8, 46.8, 3.2, 2.6),
    ]
    draw_leaf_cluster(
        c, monstera_blobs, base_col=P["Y"], hi_col=P["l"], sh_col=P["y"], deep_col=P["leaf_deep"]
    )
    # Crisp 1px Monstera leaf fenestration slits for iconic botanical silhouette
    for mx, my in ((79, 49), (83, 45), (86, 46), (89, 50), (84, 49)):
        c.set_pixel(mx, my, P["leaf_deep"])
        c.set_pixel(mx + 1, my, P["O"])

    # --- FAR LEFT SIDEWALK: VINYL CRATE (x=2..11, y=52..62) ---
    draw_box(c, 2, 56, 9, 6, P["P"], P["O"], hi=P["M"], sh=P["p"])
    c.fill_rect(3, 58, 7, 1, P["O"])
    c.fill_rect(3, 60, 7, 1, P["O"])
    # Colorful vinyl record sleeves peeking out of top of crate
    for rx, rcol in ((3, P["F"]), (5, P["X"]), (7, P["G"]), (9, P["K"])):
        c.fill_rect(rx, 53, 1, 3, rcol)
        c.set_pixel(rx, 52, P["O"])

    return c


# ==============================================================================
# 3. BUILDING: GELATERIA (96x64) — Pastel Mint & Strawberry Parlor ("GELATO")
# ==============================================================================

def build_gelateria_96x64() -> Canvas:
    """
    building_gelateria.png (96x64):
    - Vibrant pastel mint upper facade with scalloped cream & strawberry cornice
    - Glowing neon ice-cream cone sign on upper-right wall
    - Crisp pastel-strawberry 'GELATO' signboard with glowing bulb trim
    - Cream-and-pink striped scalloped awning
    - Glowing glass gelato display case packed with 6 colorful gelato tubs
    - Cute mosaic pastel tiled base, cozy door & sidewalk gelato cone sign
    """
    c = Canvas(96, 64)
    fill_ellipse(c, 48, 62.5, 45, 2.0, P["s"])

    # Mint side wall and tucked-under strawberry cornice cap.
    c.fill_rect(82, 7, 10, 54, P["O"])
    c.fill_rect(83, 8, 8, 52, P["pmint_sh"])
    c.fill_rect(88, 9, 3, 50, P["Y"])
    c.fill_rect(83, 5, 7, 4, P["k"])
    c.fill_rect(88, 6, 4, 3, P["O"])
    c.fill_rect(89, 7, 2, 1, P["Q"])
    for y in (13, 20, 46, 53):
        c.fill_rect(86, y, 3, 1, P["pmint_base"])

    # --- SCALLOPED CREAM & STRAWBERRY ROOF PARAPET (x=10..85, y=2..8) ---
    draw_box(c, 10, 4, 76, 4, P["C"], P["O"], hi=P["W"], sh=P["a"])
    # Decorative pastel finials on roof corners
    for fx in (11, 82):
        draw_box(c, fx, 2, 3, 3, P["K"], P["O"], hi=P["Q"], sh=P["k"])
    # Pink scalloped trim dots along parapet
    for px in range(15, 81, 6):
        c.fill_rect(px, 5, 3, 2, P["K"])
        c.set_pixel(px, 5, P["Q"])

    # --- VIBRANT PASTEL MINT UPPER FACADE (x=11..84, y=8..23) ---
    draw_box(c, 11, 8, 74, 15, P["pmint_base"], P["O"], hi=P["pmint_hi"], sh=P["pmint_sh"])
    # Subtle cream vertical trim molding
    c.fill_rect(13, 9, 1, 13, P["pmint_hi"])
    c.fill_rect(82, 9, 1, 13, P["pmint_sh"])

    # Upper-left cozy pastel window w/ flowerbox (x=17..37, y=9..22)
    draw_box(c, 17, 9, 20, 11, P["W"], P["O"], hi=P["W"], sh=P["w"])
    draw_glass_window(c, 19, 11, 16, 7, P["K"], P["k"])
    # Warm cream curtains inside upper window
    c.fill_rect(20, 12, 3, 5, P["C"])
    c.fill_rect(31, 12, 3, 5, P["C"])
    # Cute strawberry-pink window flowerbox on sill
    draw_box(c, 18, 19, 18, 3, P["K"], P["O"], hi=P["Q"], sh=P["k"])
    draw_leaf_cluster(c, [(22, 18.5, 2.2, 1.3), (27, 18.5, 2.4, 1.3), (32, 18.5, 2.2, 1.3)])
    c.set_pixel(23, 18, P["W"])
    c.set_pixel(27, 18, P["G"])
    c.set_pixel(31, 18, P["Q"])

    # Center upper cute pastel AC unit w/ blossoms (x=42..57, y=11..20)
    draw_ac_unit(c, 42, 11)
    draw_leaf_cluster(c, [(56, 17, 2.0, 2.2)])
    c.set_pixel(56, 16, P["K"])

    # --- GLOWING NEON ICE-CREAM CONE SIGN ON UPPER-RIGHT WALL (x=62..80, y=9..23) ---
    draw_box(c, 62, 9, 18, 14, P["berry_dark"], P["O"], hi=P["K"], sh=P["int_dark"])
    # Glowing cyan top & pink bottom inner neon frame
    c.fill_rect(63, 10, 16, 1, P["X"])
    c.fill_rect(63, 21, 16, 1, P["K"])
    fill_ellipse(c, 71, 15.5, 6.2, 4.8, P["neon_mag_glow"])
    # Pixel Ice-Cream Cone inside plaque (centered x=71, y=11..20)
    # Cherry on top
    c.fill_rect(70, 11, 2, 1, P["R"])
    c.set_pixel(70, 11, P["W"])
    # Top Mint/Cyan Scoop (y=12..14)
    c.fill_rect(69, 12, 4, 2, P["X"])
    c.set_pixel(69, 12, P["W"])
    c.set_pixel(72, 13, P["pmint_sh"])
    # Bottom Strawberry-Pink Scoop (y=14..16)
    c.fill_rect(67, 14, 8, 2, P["K"])
    c.fill_rect(68, 14, 3, 1, P["q"])
    c.fill_rect(73, 15, 2, 1, P["k"])
    # Golden Waffle Cone (y=16..20)
    c.fill_rect(68, 16, 6, 1, P["G"])
    c.fill_rect(69, 17, 4, 1, P["D"])
    c.fill_rect(69, 18, 4, 1, P["G"])
    c.fill_rect(70, 19, 2, 1, P["D"])
    c.set_pixel(70, 20, P["G"])
    # Little neon sparkle dots around cone
    c.set_pixel(65, 13, P["W"])
    c.set_pixel(76, 12, P["G"])
    c.set_pixel(77, 18, P["X"])

    # --- HERO 'GELATO' SIGNBOARD (x=10..85, y=23..34) ---
    draw_box(c, 10, 23, 76, 11, P["K"], P["O"], hi=P["Q"], sh=P["k"])
    draw_box(c, 13, 24, 70, 9, P["berry_dark"], P["W"])
    # Glowing cream marquee bulbs along top & bottom edges of sign
    for bx in range(15, 81, 5):
        c.set_pixel(bx, 24, P["G"])
        c.set_pixel(bx, 32, P["G"])
    # Cute mini gelato cup icons flanking 'GELATO' (x=17..22 & x=73..78)
    for gx in (17, 73):
        c.fill_rect(gx + 1, 26, 3, 2, P["K"])
        c.set_pixel(gx + 1, 26, P["W"])
        c.set_pixel(gx + 3, 26, P["X"])
        c.fill_rect(gx, 28, 5, 2, P["C"])
        c.fill_rect(gx + 1, 30, 3, 1, P["a"])
    # Crisp glowing 'GELATO' lettering with 4px 'G' (centered x=30..65, y=26..31)
    draw_text_ext(c, 30, 26, "GELATO", P["W"], shadow=P["K"], spacing=2)

    # --- CREAM-AND-PINK STRIPED SCALLOPED AWNING (x=7..88, y=34..43) ---
    draw_striped_awning(
        c, 7, 34, 82, 10, P["K"], P["k"], P["C"], P["a"], stripe_w=5
    )

    # --- GROUND FLOOR GELATO PARLOR (x=11..84, y=43..62) ---
    draw_box(c, 11, 43, 74, 19, P["pmint_base"], P["O"], hi=P["pmint_hi"], sh=P["pmint_sh"])

    # --- CUTE PASTEL MOSAIC TILED BASE (x=11..84, y=56..62) ---
    c.fill_rect(11, 56, 74, 6, P["O"])
    tile_cols = [P["pmint_hi"], P["Q"], P["W"], P["C"]]
    for ty in range(57, 61, 2):
        for tx in range(12, 84, 3):
            tcol = tile_cols[((tx // 3) + (ty // 2)) % len(tile_cols)]
            c.fill_rect(tx, ty, 2, 2, tcol)
            c.fill_rect(tx + 2, ty, 1, 2, P["pmint_sh"])
    # Crisp cream grout highlight bar on top of tiled base
    c.fill_rect(12, 56, 72, 1, P["W"])

    # --- GLOWING GLASS GELATO DISPLAY CASE (x=14..57, y=44..57) ---
    fill_ellipse(c, 35.5, 50.5, 22, 7.0, hex_to_rgba("#B8FFFF", 60))
    draw_box(c, 14, 44, 43, 13, P["W"], P["O"], hi=P["W"], sh=P["w"])
    # Chilled glowing ice-cyan illuminated interior
    c.fill_rect(16, 45, 39, 10, P["ice_glass"])
    c.fill_rect(16, 45, 39, 3, P["W"])
    c.fill_rect(16, 52, 39, 3, P["ice_back"])

    # Silver stainless counter bed inside case (y=51..54)
    c.fill_rect(17, 51, 37, 4, P["M"])
    c.fill_rect(17, 51, 37, 1, P["W"])

    # 6 Colorful Mounded Gelato Tubs inside glowing glass showcase!
    gelato_flavors = [
        (18, P["K"], P["q"], P["k"]),              # Strawberry
        (24, P["gel_pistachio"], P["l"], P["Y"]),  # Pistachio
        (30, P["gel_lemon"], P["W"], P["G"]),      # Lemon
        (36, P["C"], P["W"], P["a"]),              # Stracciatella
        (42, P["gel_ube"], P["Z"], P["U"]),        # Blueberry/Ube
        (48, P["gel_mango"], P["G"], P["T"]),      # Mango/Peach
    ]
    for tx, gbase, ghi, gsh in gelato_flavors:
        # Metal tub rim
        c.fill_rect(tx, 52, 5, 2, P["d"])
        c.fill_rect(tx, 51, 5, 1, P["W"])
        # Fluffy swirled gelato mound
        c.fill_rect(tx, 49, 5, 2, gbase)
        c.fill_rect(tx + 1, 48, 3, 1, ghi)
        c.set_pixel(tx + 4, 50, gsh)
        # Tiny white flavor label card behind mound
        c.fill_rect(tx + 1, 46, 3, 2, P["W"])
        c.set_pixel(tx + 2, 46, gbase)
    # Chocolate chips on Stracciatella tub
    c.set_pixel(37, 49, P["O"])
    c.set_pixel(39, 50, P["O"])

    # Diagonal bright specular reflection across display glass
    for i in range(6):
        c.set_pixel(18 + i, 53 - i, P["W"])
        c.set_pixel(44 + i, 53 - i, P["W"])

    # --- COZY PASTEL DOORWAY (x=60..79, y=44..61) ---
    draw_box(c, 60, 44, 19, 18, P["K"], P["O"], hi=P["Q"], sh=P["k"])
    draw_box(c, 62, 46, 15, 10, P["C"], P["O"], hi=P["W"], sh=P["a"])
    c.fill_rect(63, 47, 13, 8, P["stucco_hi"])
    # Subtle glass sheen
    c.set_pixel(64, 50, P["W"])
    c.set_pixel(65, 49, P["W"])
    c.set_pixel(66, 48, P["W"])
    # Cute pink heart/scoop OPEN hanger
    draw_box(c, 66, 49, 7, 4, P["K"], P["O"], hi=P["Q"])
    c.fill_rect(67, 50, 5, 1, P["W"])
    # Cream handle & tiled threshold step
    c.fill_rect(62, 57, 2, 1, P["W"])
    c.fill_rect(61, 60, 17, 2, P["C"])

    # --- SIDEWALK PROPS: CONE A-FRAME SIGN (left) & PASTEL FLOWER POT (right) ---
    # Handmade gelato cone standee with a strawberry scoop and waffle shading.
    c.fill_rect(2, 50, 10, 5, P["O"])
    c.fill_rect(4, 48, 6, 2, P["O"])
    c.fill_rect(3, 51, 8, 3, P["K"])
    c.fill_rect(5, 49, 4, 2, P["Q"])
    c.set_pixel(4, 51, P["q"])
    c.fill_rect(8, 52, 3, 2, P["k"])
    c.fill_rect(4, 54, 6, 3, P["O"])
    c.fill_rect(5, 57, 4, 3, P["O"])
    c.fill_rect(5, 55, 4, 2, P["G"])
    c.fill_rect(6, 57, 2, 2, P["D"])
    c.set_pixel(7, 55, P["T"])
    c.set_pixel(6, 57, P["G"])
    c.fill_rect(3, 60, 8, 2, P["O"])
    c.fill_rect(4, 60, 6, 1, P["T"])

    # Right cute cream planter pot with lush mint-green shrub & pink flowers (x=83..93, y=51..62)
    draw_leaf_cluster(
        c, [(86.5, 54.0, 3.2, 2.8), (90.0, 54.0, 3.0, 2.5), (88.0, 51.0, 2.8, 2.4)]
    )
    c.set_pixel(86, 52, P["K"])
    c.set_pixel(89, 51, P["G"])
    c.set_pixel(90, 54, P["W"])
    draw_box(c, 84, 56, 8, 3, P["W"], P["O"], hi=P["W"], sh=P["w"])
    draw_box(c, 85, 58, 6, 4, P["K"], P["O"], hi=P["Q"], sh=P["k"])

    return c


# ==============================================================================
# 4. PROP: CENTRAL PLAZA NEON FOUNTAIN (64x48)
# ==============================================================================

def _in_octagon(x: float, y: float, cx: float, cy: float, rx: float, ry: float, cut: float = 0.42) -> bool:
    """Return True if point (x, y) lies inside an axis-aligned stretched octagon."""
    nx = abs(x - cx) / rx
    ny = abs(y - cy) / ry
    return nx <= 1.0 and ny <= 1.0 and (nx + ny) <= (2.0 - cut)


def build_prop_fountain_64x48() -> Canvas:
    """
    prop_fountain.png (64x48) — Iconic Pop-Urban Central Plaza Neon Fountain:
    - Octagonal warm travertine stone basin with 1px Deep Aubergine outline
    - Shimmering turquoise/cyan underwater LED glow & concentric water ripples
    - Tiered central travertine pedestal with cascading water sheets & cyan neon ring
    - White & cyan arched parabolic water jets + crisp splash foam
    - 4 tiny decorative planter corners with lush foliage & Pop-Urban flowers
    """
    c = Canvas(64, 48)
    cx, cy = 32.0, 27.5

    # Ground drop shadow tucked cleanly underneath the octagonal base (zero side flaps!)
    fill_ellipse(c, cx, 43.2, 24.5, 3.2, P["s"])

    # --- OCTAGONAL WARM TRAVERTINE STONE BASIN (built on temp canvas for crisp 1px outline) ---
    basin = Canvas(64, 48)

    # 1. Lower extruded outer stone wall (3/4 vertical front & chamfered faces, y=26..42)
    for dy_wall in range(6, -1, -1):
        for py in range(15, 45):
            for px in range(5, 59):
                if _in_octagon(px + 0.5, py + 0.5, cx, cy + dy_wall, 25.0, 12.2, cut=0.44):
                    if px < cx - 10:
                        col = P["trav_base"]
                    elif px > cx + 10:
                        col = P["trav_deep"]
                    else:
                        col = P["trav_sh"]
                    if dy_wall >= 5:
                        col = P["trav_deep"]
                    basin.set_pixel(px, py, col)

    # Decorative carved vertical stone panel ribs on front wall
    for rx in (14, 21, 28, 35, 42, 49):
        for py in range(35, 42):
            if basin.pixels[py][rx][3] > 0:
                basin.set_pixel(rx, py, P["trav_hi"] if rx < cx else P["trav_deep"])

    # 2. Upper flat octagonal stone rim (top bevel)
    for py in range(14, 42):
        for px in range(5, 59):
            if _in_octagon(px + 0.5, py + 0.5, cx, cy, 25.0, 12.2, cut=0.44):
                if py < cy - 4 or px < cx - 12:
                    col = P["trav_hi"]
                elif py > cy + 6 and px > cx + 6:
                    col = P["trav_sh"]
                else:
                    col = P["trav_base"]
                basin.set_pixel(px, py, col)

    # Inner stone lip edge framing the water
    for py in range(16, 39):
        for px in range(8, 56):
            if _in_octagon(px + 0.5, py + 0.5, cx, cy, 21.0, 9.6, cut=0.44):
                basin.set_pixel(px, py, P["trav_deep"] if py < cy else P["trav_hi"])

    # 3. Shimmering Turquoise/Cyan Underwater Glow Pool!
    for py in range(17, 38):
        for px in range(9, 55):
            if _in_octagon(px + 0.5, py + 0.5, cx, cy + 0.5, 19.6, 8.6, cut=0.44):
                dist = math.hypot((px + 0.5 - cx) / 19.6, (py + 0.5 - (cy + 0.5)) / 8.6)
                if dist > 0.78:
                    col = P["turq_hi"] if py > cy else P["X"]
                elif dist < 0.48:
                    col = P["turq_hi"]
                elif dist < 0.65:
                    col = P["turq_base"]
                else:
                    col = P["turq_deep"]
                basin.set_pixel(px, py, col)

    apply_1px_outline(basin, P["O"])
    c.blit(basin, 0, 0)

    # Underwater glowing aura & crisp pixel ripple rings in water
    fill_ellipse(c, cx, cy + 0.5, 14.0, 6.2, P["neon_cyan_glow"])
    for angle_deg in range(0, 360, 14):
        rad = math.radians(angle_deg)
        rx1 = int(round(cx + 12.5 * math.cos(rad)))
        ry1 = int(round(cy + 1.0 + 5.2 * math.sin(rad)))
        if (angle_deg // 14) % 3 != 0:
            c.set_pixel(rx1, ry1, P["3"])

    # --- TIERED CENTRAL TRAVERTINE PEDESTAL (cx=32, y=9..30) ---
    ped = Canvas(64, 48)
    # Submerged octagonal stone base plinth (y=24..29)
    ped.fill_rect(26, 25, 12, 4, P["trav_base"])
    ped.fill_rect(26, 25, 4, 4, P["trav_hi"])
    ped.fill_rect(34, 25, 4, 4, P["trav_sh"])
    ped.fill_rect(27, 28, 10, 1, P["X"])  # glowing cyan waterline reflection
    # Lower fluted travertine column (y=20..25)
    ped.fill_rect(29, 20, 6, 5, P["trav_base"])
    ped.fill_rect(29, 20, 2, 5, P["trav_hi"])
    ped.fill_rect(33, 20, 2, 5, P["trav_sh"])
    # Mid wide scalloped travertine bowl tier (x=21..43, y=15..20)
    fill_ellipse(ped, 32.0, 18.0, 10.2, 2.8, P["trav_sh"])
    fill_ellipse(ped, 31.6, 17.2, 9.8, 2.2, P["trav_hi"])
    fill_ellipse(ped, 32.0, 16.8, 8.2, 1.6, P["turq_hi"])  # turquoise water inside mid tier
    ped.fill_rect(25, 19, 14, 1, P["X"])  # glowing neon cyan LED rim
    # Upper column & top finial chalice (x=28..36, y=9..16)
    ped.fill_rect(30, 12, 4, 5, P["trav_base"])
    ped.fill_rect(30, 12, 1, 5, P["trav_hi"])
    ped.fill_rect(33, 12, 1, 5, P["trav_sh"])
    fill_ellipse(ped, 32.0, 10.8, 4.8, 1.8, P["trav_hi"])
    ped.fill_rect(29, 10, 6, 1, P["G"])  # gold Pop-Urban accent collar
    apply_1px_outline(ped, P["O"])
    c.blit(ped, 0, 0)

    # Cascading water sheets spilling gracefully off mid-tier bowl into main basin
    for sx in (24, 28, 32, 36, 40):
        c.fill_rect(sx, 19, 1, 8, P["3"] if sx % 8 == 0 else P["X"])
        c.set_pixel(sx, 27, P["W"])

    # --- WHITE & CYAN ARCHED WATER JETS & SPLASH FOAM ---
    # 1. Top crown fountain plume spraying upward from finial (y=3..10)
    c.fill_rect(30, 5, 4, 4, P["turq_hi"])
    c.fill_rect(31, 4, 2, 6, P["W"])
    c.set_pixel(30, 4, P["X"])
    c.set_pixel(33, 4, P["X"])
    c.set_pixel(31, 3, P["X"])
    c.set_pixel(32, 3, P["W"])
    # Crown left & right mini arcs falling into mid-tier bowl
    for side in (-1, 1):
        for step in range(1, 9):
            t = step / 8.0
            jx = int(round(32 + side * (step * 1.15)))
            jy = int(round(5.0 + 11.5 * (t * t) - 2.0 * t))
            c.set_pixel(jx, jy, P["W"] if step < 5 else P["3"])
            if step in (2, 4, 6):
                c.set_pixel(jx, jy + 1, P["X"])

    # 2. Four Iconic Parabolic Arched Water Jets curving into the octagonal pool!
    jet_targets = [
        (-1, 17.5, 16.0, 27.5, 7.5),   # Far-left high arc
        (1, 17.5, 48.0, 27.5, 7.5),    # Far-right high arc
        (-1, 18.0, 21.0, 32.5, 6.2),   # Front-left arc
        (1, 18.0, 43.0, 32.5, 6.2),    # Front-right arc
    ]
    for side, y0, tx, ty, peak_h in jet_targets:
        x0 = 32.0 + side * 6.0
        steps = 16
        for i in range(steps + 1):
            t = i / steps
            jx = int(round(x0 + (tx - x0) * t))
            jy = int(round(y0 + (ty - y0) * t - 4.0 * peak_h * t * (1.0 - t)))
            col = P["W"] if i % 3 != 2 else P["3"]
            c.set_pixel(jx, jy, col)
            if 2 <= i <= steps - 2:
                c.set_pixel(jx, jy + 1, P["X"])
        # Crisp bubbly white & cyan splash foam at landing point!
        ix, iy = int(round(tx)), int(round(ty))
        c.fill_rect(ix - 2, iy, 5, 1, P["W"])
        c.fill_rect(ix - 1, iy - 1, 3, 1, P["foam"])
        c.set_pixel(ix - 3, iy + 1, P["3"])
        c.set_pixel(ix + 3, iy + 1, P["3"])
        c.set_pixel(ix, iy - 2, P["W"])

    # --- 4 TINY DECORATIVE PLANTER CORNERS ON OCTAGON CHAMFERS ---
    planter_corners = [
        (8, 17, P["K"]),    # top-left sakura blossom
        (50, 17, P["G"]),   # top-right gold blossom
        (7, 35, P["G"]),    # bottom-left gold blossom
        (51, 35, P["K"]),   # bottom-right sakura blossom
    ]
    for px, py, fcol in planter_corners:
        draw_box(c, px, py + 2, 6, 5, P["trav_base"], P["O"], hi=P["trav_hi"], sh=P["trav_deep"])
        c.fill_rect(px + 1, py + 3, 4, 1, P["terra_base"])
        # Rounded neat shrub cluster (clean silhouette without spiky top pixel)
        draw_leaf_cluster(
            c,
            [(px + 3.0, py + 1.5, 3.0, 1.9)],
            base_col=P["Y"],
            hi_col=P["l"],
            sh_col=P["y"],
        )
        c.set_pixel(px + 2, py + 1, fcol)
        c.set_pixel(px + 4, py + 1, P["W"])

    return c


# ==============================================================================
# 5. PROP: NEON DISTRICT STREET ARCH (96x40) — Wide Walkthrough Gate
# ==============================================================================

def build_prop_arch_gate_96x40() -> Canvas:
    """
    prop_arch_gate.png (96x40) — Experimental Tokyo/Roma Pop-Urban Neon Street Arch:
    - Slender left & right ornamental dark-metal pillars (leaving a wide 68px clear
      transparent walkthrough gap in the center!)
    - Ornate arched overhead neon marquee bar with glowing 'NEON STOP' crest,
      alternating cyan/magenta/gold neon bulbs, and curved wrought-iron scroll brackets
    - 4 tiny glowing rounded hanging street lanterns under the side spans
    """
    c = Canvas(96, 40)

    # Ground drop shadows ONLY under the two pillar bases (center walkthrough stays 100% transparent!)
    fill_ellipse(c, 8.5, 37.5, 5.5, 1.6, P["s"])
    fill_ellipse(c, 87.5, 37.5, 5.5, 1.6, P["s"])

    # --- LEFT & RIGHT SLENDER ORNAMENTAL DARK-METAL PILLARS (x=4..12 & x=83..91) ---
    for px in (4, 83):
        # Fluted heavy plinth base (y=32..38)
        draw_box(c, px, 33, 9, 5, P["P"], P["O"], hi=P["M"], sh=P["p"])
        draw_box(c, px + 1, 30, 7, 4, P["P"], P["O"], hi=P["M"], sh=P["p"])
        c.fill_rect(px + 2, 31, 5, 1, P["X"])  # glowing cyan base collar!

        # Slender fluted dark-metal column shaft (y=11..30)
        c.fill_rect(px + 2, 11, 5, 20, P["O"])
        c.fill_rect(px + 3, 11, 1, 19, P["W"])  # crisp left metallic specular
        c.fill_rect(px + 4, 11, 1, 19, P["M"])  # mid steel
        c.fill_rect(px + 5, 11, 1, 19, P["p"])  # deep shadow

        # Mid ornamental rings on pillar
        draw_box(c, px + 1, 19, 7, 3, P["P"], P["O"], hi=P["M"], sh=P["p"])
        c.fill_rect(px + 3, 20, 3, 1, P["G"])  # gold rivet band

        # Pillar capital & glowing neon finial orb atop pillar (y=2..11)
        draw_box(c, px + 1, 8, 7, 4, P["P"], P["O"], hi=P["W"], sh=P["p"])
        fill_ellipse(c, px + 4.5, 4.5, 3.8, 3.4, P["neon_cyan_glow"])
        c.fill_rect(px + 4, 6, 1, 2, P["O"])
        draw_box(c, px + 2, 2, 5, 5, P["X"], P["O"], hi=P["W"], sh=P["2"])
        c.set_pixel(px + 3, 3, P["W"])

    # --- CURVED WROUGHT-IRON SCROLL BRACKETS BRACING THE ARCH (x=11..18 & x=77..84, y=13..20) ---
    for bx, by in ((11, 19), (12, 18), (13, 17), (14, 16), (15, 15), (16, 14), (12, 15), (13, 14)):
        c.set_pixel(bx, by, P["O"])
        c.set_pixel(bx - 1, by, P["O"])
        c.set_pixel(bx, by - 1, P["M"])
    for bx, by in ((84, 19), (83, 18), (82, 17), (81, 16), (80, 15), (79, 14), (83, 15), (82, 14)):
        c.set_pixel(bx, by, P["O"])
        c.set_pixel(bx + 1, by, P["O"])
        c.set_pixel(bx, by - 1, P["M"])

    # --- ORNATE OVERHEAD NEON MARQUEE SPAN (x=9..86, y=6..14) ---
    # Subtle tight neon glow halo behind marquee
    fill_ellipse(c, 48, 9.5, 30, 5.0, hex_to_rgba("#FF2A85", 45))
    fill_ellipse(c, 48, 8.5, 22, 4.5, hex_to_rgba("#00F0FF", 45))

    # Horizontal dark-metal wing truss beam (x=9..86, y=6..14)
    draw_box(c, 9, 6, 78, 8, P["P"], P["O"], hi=P["M"], sh=P["p"])
    c.fill_rect(11, 8, 74, 4, P["int_dark"])
    # Diagonal iron truss lattice inside left & right wings
    for tx in (13, 18, 75, 80):
        c.set_pixel(tx, 8, P["d"])
        c.set_pixel(tx + 1, 9, P["d"])
        c.set_pixel(tx + 2, 10, P["d"])
        c.set_pixel(tx + 1, 11, P["d"])

    # Glowing Cyan / Magenta / Gold Neon Bulbs along upper & lower marquee rails!
    bulb_cycle = [P["X"], P["neon_mag"], P["G"]]
    for idx, bx in enumerate(range(12, 85, 4)):
        bcol = bulb_cycle[idx % 3]
        c.set_pixel(bx, 6, P["W"])
        c.set_pixel(bx, 7, bcol)
        bcol2 = bulb_cycle[(idx + 1) % 3]
        c.set_pixel(bx, 12, bcol2)
        c.set_pixel(bx, 13, P["W"] if idx % 2 == 0 else bcol2)

    # --- RAISED CENTRAL POP-URBAN NEON CREST: 'NEON STOP' (x=24..71, y=1..17) ---
    # Ornamental gold crown peak
    draw_box(c, 42, 1, 12, 3, P["G"], P["O"], hi=P["W"], sh=P["D"])
    # Main central marquee housing (48px wide x 14px tall for crisp 1px dark padding around text)
    draw_box(c, 24, 3, 48, 14, P["P"], P["O"], hi=P["W"], sh=P["p"])
    draw_box(c, 26, 4, 44, 12, P["int_dark"], P["O"])
    # Dual glowing neon tubing border inside crest (Cyan top/left, Magenta bottom/right)
    c.fill_rect(27, 5, 42, 1, P["X"])
    c.fill_rect(27, 5, 1, 10, P["X"])
    c.fill_rect(27, 14, 42, 1, P["neon_mag"])
    c.fill_rect(68, 5, 1, 10, P["neon_mag"])
    # Corner bright white neon nodes
    for nx, ny in ((27, 5), (68, 5), (27, 14), (68, 14)):
        c.set_pixel(nx, ny, P["W"])

    # Crisp glowing 'NEON STOP' lettering (with 4px 'N's = 34px wide, centered x=31..64, y=7..12)
    draw_text_ext(c, 31, 7, "NEON STOP", P["W"], shadow=P["neon_mag"], spacing=1)

    # --- 4 TINY ROUNDED HANGING STREET LANTERNS UNDER SIDE WING SPANS (y=14..21) ---
    lantern_specs = [
        (16, P["G"], P["C"], P["oven_orange"], hex_to_rgba("#FFD166", 48)),
        (22, P["F"], P["q"], P["neon_mag"], hex_to_rgba("#FF2A85", 48)),
        (73, P["F"], P["q"], P["neon_mag"], hex_to_rgba("#FF2A85", 48)),
        (79, P["G"], P["C"], P["oven_orange"], hex_to_rgba("#FFD166", 48)),
    ]
    for lx, lbase, lhi, lsh, lglow in lantern_specs:
        fill_ellipse(c, lx + 0.5, 18.0, 3.2, 3.0, lglow)
        # Hanging chain
        c.set_pixel(lx, 14, P["O"])
        c.set_pixel(lx, 15, P["M"])
        # Rounded paper/glass lantern silhouette (5x5 with trimmed corners & crisp outline)
        lan = Canvas(96, 40)
        lan.fill_rect(lx - 1, 16, 3, 5, lbase)
        lan.fill_rect(lx - 2, 17, 5, 3, lbase)
        lan.set_pixel(lx - 1, 17, lhi)
        lan.set_pixel(lx, 17, P["W"])
        lan.set_pixel(lx, 18, lhi)
        lan.fill_rect(lx - 1, 19, 3, 1, lsh)
        apply_1px_outline(lan, P["O"])
        c.blit(lan, 0, 0)
        # Bottom tassel
        c.set_pixel(lx, 21, P["O"])

    return c


# ==============================================================================
# MAIN ASSET GENERATION & VERIFICATION
# ==============================================================================

def main() -> None:
    os.makedirs(OUT_DIR, exist_ok=True)

    assets: List[Tuple[str, Canvas, int]] = [
        ("building_pizzeria", build_pizzeria_96x64(), 6),
        ("building_vinyl_bar", build_vinyl_bar_96x64(), 6),
        ("building_gelateria", build_gelateria_96x64(), 6),
        ("prop_fountain", build_prop_fountain_64x48(), 8),
        ("prop_arch_gate", build_prop_arch_gate_96x40(), 6),
    ]

    for name, canvas, zoom in assets:
        save_asset_and_preview(canvas, OUT_DIR, name, zoom=zoom)
        out_path = os.path.join(OUT_DIR, f"{name}.png")
        size_bytes = os.path.getsize(out_path)
        print(
            f"Generated {name}.png ({canvas.width}x{canvas.height}, {size_bytes} bytes) "
            f"+ preview_{name}_{zoom}x.png"
        )

    print(f"Successfully generated all {len(assets)} v2 hero map assets in {OUT_DIR}")


if __name__ == "__main__":
    main()
