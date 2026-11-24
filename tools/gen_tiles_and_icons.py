#!/usr/bin/env python3
"""
GIRLBUS: NEON STOP (girlbusAG) — Background Tiles & UI Icon Pack Pixel-Art Generator.

Generates 1x native PNG sprites (+ zoomed previews in tools/previews/):
  Part A — 5 Seamless 32x32 Ground/Background Tiles in assets/sprites/map/:
    - tile_plaza_mosaic.png
    - tile_wet_asphalt.png
    - tile_park_path.png
    - tile_bike_lane.png
    - tile_canal_edge.png

  Part B — 31 Crisp Pixel-Art UI & Pickup Icons in assets/sprites/ui/:
    - 6 Upgrade Icons (20x20)
    - 6 Relic Icons (16x16)
    - 4 Vending Shop Icons (16x16)
    - 6 Rizz-Slots 777 Reel Icons (20x20)
    - 9 HUD, Controls & World Pickup Icons (16x16 / 12x12)

Follows 03_ART_STYLE_AND_SPRITE_BIBLE.md:
  - Universal 1px Deep Aubergine outline (#1C1124)
  - Top-Left cel-pixel 3-tone shading
  - Low-noise background tiles so electric bullets pop effortlessly
  - High-luminance, instantly readable UI icon silhouettes
"""

import os
import sys
from typing import Dict, List, Tuple

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

MAP_OUT_DIR = os.path.join(ROOT_DIR, "assets", "sprites", "map")
UI_OUT_DIR = os.path.join(ROOT_DIR, "assets", "sprites", "ui")

# ==============================================================================
# EXTENDED POP-URBAN TILE & UI PALETTE
# ==============================================================================

TILE_COLORS: Dict[str, RGBA] = {
    **PALETTE,
    # Warm travertine plaza stones (low-noise, desaturated for bullet contrast)
    "trav_base": hex_to_rgba("#D6CCBA"),
    "trav_hi": hex_to_rgba("#E6DEC9"),
    "trav_sh": hex_to_rgba("#B8AD99"),
    "trav_grout": hex_to_rgba("#9E9482"),
    # Mid-tone plaza travertine and oxidized copper inlay
    "mosaic_stone": hex_to_rgba("#ADA299"),
    "mosaic_edge": hex_to_rgba("#B7ACA2"),
    "mosaic_grout": hex_to_rgba("#998F88"),
    "mosaic_copper": hex_to_rgba("#8C9390"),
    "mosaic_stud": hex_to_rgba("#B3A58F"),
    # Evening wet asphalt (matches tile_asphalt #524E61 seamlessly)
    "wet_base": hex_to_rgba("#514D60"),
    "wet_dark": hex_to_rgba("#4A465B"),
    "wet_grain": hex_to_rgba("#5B566A"),
    "wet_sheen_c": hex_to_rgba("#5B6078"),
    "wet_sheen_hi": hex_to_rgba("#636E86"),
    "wet_sheen_p": hex_to_rgba("#60566C"),
    # Park grass (matches tile_grass_park.png exactly)
    "grass_base": hex_to_rgba("#58A854"),
    "grass_sh": hex_to_rgba("#448E42"),
    "grass_hi": hex_to_rgba("#78C86C"),
    "dirt_base": hex_to_rgba("#C2B093"),
    "dirt_sh": hex_to_rgba("#A89578"),
    "dirt_hi": hex_to_rgba("#D4C5AA"),
    "stone_base": hex_to_rgba("#DDD4C3"),
    "stone_hi": hex_to_rgba("#EFE8DA"),
    "stone_sh": hex_to_rgba("#B5A996"),
    # Bike lane (muted terracotta/coral + crisp cream line + asphalt edge)
    "bike_base": hex_to_rgba("#A35956"),
    "bike_sh": hex_to_rgba("#8F4A48"),
    "bike_hi": hex_to_rgba("#B56A66"),
    "bike_cream": hex_to_rgba("#EFE5D1"),
    "bike_cream_sh": hex_to_rgba("#D4C7AF"),
    "asphalt_edge": hex_to_rgba("#524E61"),
    "asphalt_edge_sh": hex_to_rgba("#444052"),
    # Canal embankment stone & water (matches tile_canal_water.png seamlessly)
    "wall_base": hex_to_rgba("#8E8780"),
    "wall_hi": hex_to_rgba("#A8A096"),
    "wall_sh": hex_to_rgba("#706964"),
    "wall_mortar": hex_to_rgba("#524C4A"),
    "wall_damp": hex_to_rgba("#5B5B69"),
    "wall_damp_sh": hex_to_rgba("#434657"),
    "ledge_sh": hex_to_rgba("#463F3E"),
    "water_base": hex_to_rgba("#3A7EC2"),
    "water_sh": hex_to_rgba("#2A629E"),
    "water_hi": hex_to_rgba("#68B0F0"),
    "water_foam": hex_to_rgba("#C4E8FF"),
}

# Single-char high-contrast UI icon palette (inherits #1C1124 outline & core tones)
UI_PALETTE: Dict[str, RGBA] = {
    ".": (0, 0, 0, 0),                    # Transparent
    "O": PALETTE["O"],                    # #1C1124 Universal 1px Outer Outline
    "o": PALETTE["o"],                    # #382342 Soft Internal Outline / Dark Plum
    "h": hex_to_rgba("#1C1124", 115),     # Hollow Dark-Plum Translucent Interior
    "W": PALETTE["W"],                    # #FFFFFF Pure White Specular
    "w": PALETTE["w"],                    # #D3DCE8 Cool Silver-White
    "M": hex_to_rgba("#8C9EB8"),          # Metallic Silver Base
    "d": hex_to_rgba("#526178"),          # Metallic Silver Shadow
    "g": hex_to_rgba("#232C40"),          # Dark Glass / Dial Face / Slate
    "z": hex_to_rgba("#3E4F6E"),          # Glass Reflection / Slate Mid
    "P": PALETTE["P"],                    # #303644 Charcoal Base
    "p": PALETTE["p"],                    # #1F232E Deep Charcoal Shadow
    # Electric Cyan Ramp
    "3": hex_to_rgba("#C8FFFF"),          # Pale Cyan Specular
    "X": hex_to_rgba("#00F0FF"),          # Electric Cyan Base
    "2": hex_to_rgba("#0099B8"),          # Deep Cyan Shadow
    "c": hex_to_rgba("#005C78"),          # Dark Cyan Interior Shadow
    # Neon Pink / Magenta Ramp
    "q": hex_to_rgba("#FFDCE8"),          # Sakura Specular
    "Q": hex_to_rgba("#FF85B8"),          # Hot Pink Highlight
    "K": hex_to_rgba("#FF2A85"),          # Shocking Neon Pink Base
    "k": hex_to_rgba("#B5145B"),          # Deep Magenta Shadow
    # Glossy Ruby / Crimson Red Ramp
    "F": hex_to_rgba("#FF7080"),          # Bright Coral-Red Highlight
    "R": hex_to_rgba("#F02E44"),          # Glossy Ruby Red Base
    "r": hex_to_rgba("#A31428"),          # Deep Crimson Shadow
    # Luminous Gold / Amber Ramp
    "L": hex_to_rgba("#FFF5B8"),          # Lemon-White Specular
    "G": hex_to_rgba("#FFD93D"),          # Luminous Gold Base
    "D": hex_to_rgba("#E0941A"),          # Warm Amber Shadow
    "b": hex_to_rgba("#99580A"),          # Deep Bronze Shadow
    # Vibrant Emerald / Mint Green Ramp
    "l": hex_to_rgba("#A4FF9E"),          # Bright Mint-Lime Highlight
    "Y": hex_to_rgba("#32D468"),          # Vibrant Emerald Base
    "y": hex_to_rgba("#188C43"),          # Deep Emerald Shadow
    "e": hex_to_rgba("#0D592B"),          # Dark Forest Shadow
    "m": hex_to_rgba("#7CF5D6"),          # Luminous Mint Accent
    # Neon Violet / Purple Ramp
    "Z": hex_to_rgba("#C98BFF"),          # Lavender Highlight
    "U": hex_to_rgba("#8A3FFC"),          # Vibrant Violet Base
    "u": hex_to_rgba("#521D9E"),          # Deep Violet Shadow
    # Espresso / Warm Cream Ramp
    "C": hex_to_rgba("#FFF0D4"),          # Warm Cream Base
    "a": hex_to_rgba("#DEC095"),          # Warm Cream Shadow
    "T": hex_to_rgba("#7A4426"),          # Rich Espresso Roast
    "t": hex_to_rgba("#4A2412"),          # Dark Espresso Shadow
    "1": hex_to_rgba("#C97C42"),          # Golden Crema / Caramel
    # Electric Royal Blue Ramp
    "B": hex_to_rgba("#54AEFF"),          # Bright Sky-Blue Highlight
    "J": hex_to_rgba("#1E6EE6"),          # Vibrant Cobalt Blue Base
    "j": hex_to_rgba("#113E94"),          # Deep Navy Blue Shadow
    # Luminous Aura Halos (World Heal Pickup)
    "+": hex_to_rgba("#7CF5D6", 150),     # Soft Mint Glow Aura
    "*": hex_to_rgba("#FFFFFF", 220),     # Bright White Inner Rim Glow
}


# ==============================================================================
# PART A — 5 SEAMLESS BACKGROUND / GROUND TILES (32x32 each)
# ==============================================================================

def build_tile_plaza_mosaic_32() -> Canvas:
    """Broad stone slabs with a small oxidized-copper setting."""
    T = TILE_COLORS
    c = Canvas(32, 32, bg=T["mosaic_stone"])
    c.fill_rect(0, 0, 32, 1, T["mosaic_grout"])
    c.fill_rect(0, 0, 1, 32, T["mosaic_grout"])
    c.fill_rect(1, 1, 31, 1, T["mosaic_edge"])
    c.fill_rect(1, 1, 1, 31, T["mosaic_edge"])

    for delta in range(-6, 7):
        inset = 6 - abs(delta)
        y = 16 + delta
        c.fill_rect(16 - inset, y, 2 * inset + 1, 1, T["mosaic_copper"])
        if inset >= 2:
            c.fill_rect(17 - inset, y, 2 * inset - 1, 1, T["mosaic_stone"])

    c.fill_rect(15, 16, 3, 1, T["mosaic_stud"])
    c.set_pixel(16, 15, T["mosaic_edge"])
    c.set_pixel(16, 17, T["mosaic_stud"])
    return c


def build_tile_wet_asphalt_32() -> Canvas:
    """Muted slate pavement with sparse, low-contrast rain sheen."""
    T = TILE_COLORS
    c = Canvas(32, 32, bg=T["wet_base"])

    # Broken wet areas wrap across the tile edges.
    wet_bands = [
        (3, 4, 13), (4, 3, 16), (5, 6, 9),
        (11, 24, 9), (12, 22, 13), (13, 26, 7),
        (23, 8, 11), (24, 5, 14), (25, 7, 8),
    ]
    for py, sx, length in wet_bands:
        for dx in range(length):
            c.set_pixel((sx + dx) % 32, py % 32, T["wet_dark"])

    # Reflections from fixtures are animated by the renderer. This tile supplies the underlying sheen.
    cyan_streaks = [
        (4, 6, 6), (4, 14, 2),
        (12, 26, 5), (13, 30, 3),
        (24, 9, 5),
    ]
    for idx, (py, sx, length) in enumerate(cyan_streaks):
        for dx in range(length):
            is_core = idx in (0, 2) and dx == length // 2
            col = T["wet_sheen_hi"] if is_core else T["wet_sheen_c"]
            c.set_pixel((sx + dx) % 32, py % 32, col)

    mauve_streaks = [
        (5, 9, 3),
        (25, 10, 4),
    ]
    for py, sx, length in mauve_streaks:
        for dx in range(length):
            c.set_pixel((sx + dx) % 32, py % 32, T["wet_sheen_p"])

    # Fine aggregate grain matching tile_asphalt.png pebble density
    grain_dark = [
        (1, 2), (14, 1), (25, 3), (29, 8), (3, 13), (11, 15),
        (22, 11), (8, 20), (15, 27), (27, 23), (2, 30), (13, 31),
    ]
    grain_light = [
        (10, 2), (20, 6), (30, 12), (5, 17), (14, 10), (26, 17),
        (1, 22), (17, 21), (21, 27), (30, 28), (16, 1), (7, 9),
    ]
    for gx, gy in grain_dark:
        c.set_pixel(gx, gy, T["wet_dark"])
    for gx, gy in grain_light:
        c.set_pixel(gx, gy, T["wet_grain"])

    return c


def build_tile_park_path_32() -> Canvas:
    """
    tile_park_path.png (32x32) — Soft warm pebble/stepping-stone garden path
    framed with gentle grass edges for the Pocket Parks (seamless vertically & horizontally).
    """
    T = TILE_COLORS
    c = Canvas(32, 32, bg=T["dirt_base"])

    # Soft grassy outer borders on left & right (x=0..4 and x=27..31) with organic scalloped edge
    left_widths = [
        4, 4, 5, 5, 4, 3, 3, 4,
        4, 5, 5, 4, 4, 3, 4, 4,
        5, 4, 4, 3, 3, 4, 5, 5,
        4, 4, 3, 4, 4, 5, 4, 4,
    ]
    right_widths = [
        4, 5, 5, 4, 3, 3, 4, 4,
        5, 4, 4, 3, 4, 4, 5, 5,
        4, 4, 3, 4, 4, 5, 4, 4,
        3, 3, 4, 5, 5, 4, 4, 4,
    ]
    for y in range(32):
        lw = left_widths[y]
        rw = right_widths[y]
        for x in range(lw):
            col = T["grass_sh"] if x == lw - 1 else T["grass_base"]
            c.set_pixel(x, y, col)
        c.set_pixel(lw, y, T["dirt_sh"])  # soft dirt shadow under grass rim
        for x in range(32 - rw, 32):
            col = T["grass_sh"] if x == 32 - rw else T["grass_base"]
            c.set_pixel(x, y, col)
        c.set_pixel(31 - rw, y, T["dirt_sh"])

    # Subtle grass blade highlights inside the park grass edges
    for gx, gy in ((1, 3), (2, 10), (1, 17), (2, 24), (1, 29), (29, 2), (30, 9), (29, 15), (30, 22), (29, 28)):
        c.set_pixel(gx, gy, T["grass_hi"])
        c.set_pixel(gx, (gy + 1) % 32, T["grass_sh"])

    # Warm sandy dirt subtle variation
    for dx, dy in ((7, 3), (22, 6), (12, 15), (24, 19), (8, 27), (19, 30)):
        c.set_pixel(dx, dy, T["dirt_hi"])
    for dx, dy in ((15, 2), (9, 11), (20, 14), (14, 23), (23, 28), (11, 30)):
        c.set_pixel(dx, dy, T["dirt_sh"])

    # Rounded warm stepping stones (3 main stones staggered nicely along the path)
    def draw_stepping_stone(cx: float, cy: float, rx: float, ry: float) -> None:
        for py in range(int(cy - ry - 1), int(cy + ry + 2)):
            for px in range(int(cx - rx - 1), int(cx + rx + 2)):
                dx = (px + 0.5 - cx) / rx
                dy = (py + 0.5 - cy) / ry
                d2 = dx * dx + dy * dy
                if d2 <= 1.0:
                    wx, wy = px % 32, py % 32
                    if dy > 0.55 or dx > 0.65:
                        c.set_pixel(wx, wy, T["stone_sh"])
                    elif dy < -0.45 or dx < -0.55:
                        c.set_pixel(wx, wy, T["stone_hi"])
                    else:
                        c.set_pixel(wx, wy, T["stone_base"])
                elif d2 <= 1.28 and (dy > 0.25 or dx > 0.25):
                    # Subtle bottom-right earth shadow under stepping stone
                    wx, wy = px % 32, py % 32
                    if 5 <= wx <= 26:
                        c.set_pixel(wx, wy, T["dirt_sh"])

    draw_stepping_stone(13.0, 6.5, 5.2, 3.6)
    draw_stepping_stone(19.5, 16.5, 5.0, 3.8)
    draw_stepping_stone(12.5, 26.5, 4.8, 3.5)

    # Small smooth decorative river pebbles nestled cleanly between stones
    pebbles = [(21, 5), (23, 9), (8, 15), (10, 19), (22, 25), (20, 29), (10, 12)]
    for px, py in pebbles:
        c.set_pixel(px, py, T["stone_hi"])
        c.set_pixel(px + 1, py, T["stone_base"])
        c.set_pixel(px, py + 1, T["stone_base"])
        c.set_pixel(px + 1, py + 1, T["stone_sh"])

    return c


def build_tile_bike_lane_32() -> Canvas:
    """
    tile_bike_lane.png (32x32) — Muted terracotta/coral urban cycle-path strip
    with crisp cream edge line and subtle lane dash (seamless vertically).
    """
    T = TILE_COLORS
    c = Canvas(32, 32, bg=T["bike_base"])

    # Narrow evening asphalt transition borders at x=0..1 and x=30..31
    for y in range(32):
        c.set_pixel(0, y, T["asphalt_edge"])
        c.set_pixel(1, y, T["asphalt_edge_sh"])
        c.set_pixel(30, y, T["asphalt_edge_sh"])
        c.set_pixel(31, y, T["asphalt_edge"])

    # Crisp cream solid border lines (x=2..3 and x=28..29)
    for y in range(32):
        c.set_pixel(2, y, T["bike_cream"])
        c.set_pixel(3, y, T["bike_cream_sh"])
        c.set_pixel(28, y, T["bike_cream"])
        c.set_pixel(29, y, T["bike_cream_sh"])

    # Subtle weathered seam along edge paint
    for y in (5, 13, 21, 29):
        c.set_pixel(3, y, T["bike_base"])
        c.set_pixel(28, (y + 3) % 32, T["bike_base"])

    # Centered subtle vertical lane dash (x=15..16, two balanced dashes wrapping cleanly)
    for y in list(range(3, 13)) + list(range(19, 29)):
        c.set_pixel(15, y, T["bike_cream"])
        c.set_pixel(16, y, T["bike_cream_sh"])

    # Fine low-noise terracotta surface aggregate grain
    grain_sh = [
        (6, 2), (11, 7), (22, 4), (25, 11), (8, 15),
        (19, 18), (6, 23), (12, 28), (24, 26), (20, 9),
    ]
    grain_hi = [
        (9, 4), (21, 6), (7, 12), (26, 16), (11, 20),
        (23, 22), (8, 29), (18, 30), (13, 13), (25, 2),
    ]
    for gx, gy in grain_sh:
        c.set_pixel(gx, gy, T["bike_sh"])
    for gx, gy in grain_hi:
        c.set_pixel(gx, gy, T["bike_hi"])

    return c


def build_tile_canal_edge_32() -> Canvas:
    """
    tile_canal_edge.png (32x32) — Detailed stone canal embankment / waterfront
    retaining wall ledge with warm travertine coping stones, tiny moss tufts,
    and dark mooring cleat detail (seamless horizontally).
    """
    T = TILE_COLORS
    c = Canvas(32, 32, bg=T["water_base"])

    # --- 1. TOP WARM TRAVERTINE COPING STONES (y=0..11) ---
    # Two seamless 16px-wide coping slabs (x=0..15 and x=16..31)
    for slab_x in (0, 16):
        c.fill_rect(slab_x, 0, 16, 12, T["trav_base"])
        # Top & left bevel highlight
        c.fill_rect(slab_x + 1, 0, 14, 1, T["trav_hi"])
        c.fill_rect(slab_x + 1, 1, 14, 1, T["trav_hi"])
        c.fill_rect(slab_x + 1, 1, 1, 10, T["trav_hi"])
        # Vertical mortar joint between coping slabs
        c.fill_rect(slab_x, 0, 1, 12, T["trav_grout"])
        # Front rounded coping nose bevel (y=10..11)
        c.fill_rect(slab_x + 1, 10, 15, 1, T["trav_sh"])
        c.fill_rect(slab_x, 11, 16, 1, T["trav_grout"])

    # Subtle weathered stone specks on coping top
    for sx, sy in ((5, 4), (11, 7), (19, 3), (29, 8)):
        c.set_pixel(sx, sy, T["trav_sh"])
    for sx, sy in ((8, 3), (13, 5), (25, 4)):
        c.set_pixel(sx, sy, T["trav_hi"])

    # --- 2. DARK-IRON MOORING CLEAT DETAIL (x=20..28, y=3..8) ---
    # Drop shadow beneath cleat
    c.fill_rect(21, 8, 7, 1, T["trav_grout"])
    # Cleat base plate (bolted to coping stone)
    c.fill_rect(21, 5, 7, 3, T["O"])
    c.fill_rect(22, 6, 5, 1, T["P"])
    # Cleat horns (classic T-shaped waterfront mooring cleat)
    c.fill_rect(20, 3, 9, 3, T["O"])
    c.fill_rect(21, 4, 7, 1, T["M"])
    c.set_pixel(21, 4, T["w"])  # top-left metallic glint
    c.set_pixel(27, 4, T["d"])
    # Bolt studs
    c.set_pixel(22, 7, T["d"])
    c.set_pixel(26, 7, T["d"])

    # --- 3. OVERHANGING COPING SHADOW LEDGE (y=12..13) ---
    c.fill_rect(0, 12, 32, 1, T["O"])
    c.fill_rect(0, 13, 32, 1, T["ledge_sh"])

    # --- 4. COURSED STONE RETAINING WALL MASONRY (y=14..26) ---
    c.fill_rect(0, 14, 32, 13, T["wall_mortar"])
    # Row 1 (y=14..18)
    for bx, bw in ((0, 10), (10, 12), (22, 10)):
        c.fill_rect(bx, 14, bw - 1, 4, T["wall_base"])
        c.fill_rect(bx, 14, bw - 1, 1, T["wall_hi"])
        c.fill_rect(bx, 17, bw - 1, 1, T["wall_sh"])
    # Row 2 (y=19..22, staggered)
    for bx, bw in ((0, 6), (6, 11), (17, 10), (27, 5)):
        c.fill_rect(bx, 19, bw - 1, 3, T["wall_base"])
        c.fill_rect(bx, 19, bw - 1, 1, T["wall_hi"])
        c.fill_rect(bx, 21, bw - 1, 1, T["wall_sh"])
    # Row 3 (y=23..26, damp waterline stone band)
    for bx, bw in ((0, 11), (11, 10), (21, 11)):
        c.fill_rect(bx, 23, bw - 1, 4, T["wall_damp"])
        c.fill_rect(bx, 23, bw - 1, 1, T["wall_sh"])
        c.fill_rect(bx, 26, bw - 1, 1, T["wall_damp_sh"])

    # --- 5. TINY LUSH MOSS TUFTS PEEKING FROM JOINTS ---
    # Moss tuft 1 on coping overhang & joint (x=5..9, y=11..14)
    c.set_pixel(6, 11, T["Y"])
    c.set_pixel(7, 11, T["l"])
    c.fill_rect(6, 12, 3, 2, T["Y"])
    c.set_pixel(7, 12, T["l"])
    c.set_pixel(8, 13, T["y"])
    c.set_pixel(7, 14, T["y"])
    # Moss tuft 2 at center coping joint (x=15..17, y=10..13)
    c.set_pixel(16, 10, T["l"])
    c.set_pixel(15, 11, T["Y"])
    c.set_pixel(16, 11, T["Y"])
    c.set_pixel(16, 12, T["y"])
    # Moss tuft 3 near damp waterline (x=25..28, y=22..25)
    c.set_pixel(26, 22, T["l"])
    c.fill_rect(25, 23, 3, 2, T["Y"])
    c.set_pixel(27, 24, T["y"])
    c.set_pixel(26, 25, T["y"])

    # --- 6. CANAL WATERLINE TRANSITION (y=27..31) ---
    # Deep water reflection shadow right beneath retaining wall
    c.fill_rect(0, 27, 32, 2, T["water_sh"])
    c.fill_rect(0, 29, 32, 3, T["water_base"])
    # Crisp gentle foam / water-lap highlights along wall base (seamless horizontally)
    for fx, flen in ((2, 6), (11, 5), (19, 7), (28, 3)):
        for dx in range(flen):
            c.set_pixel((fx + dx) % 32, 27, T["water_hi"])
        c.set_pixel((fx + 1) % 32, 27, T["water_foam"])
        if flen >= 6:
            c.set_pixel((fx + 2) % 32, 27, T["water_foam"])
    # Soft secondary ripple in canal water at y=30
    for fx, flen in ((6, 5), (22, 6)):
        for dx in range(flen):
            c.set_pixel((fx + dx) % 32, 30, T["water_hi"])

    return c


# ==============================================================================
# PART B — COMPLETE PIXEL-ART UI & PICKUP ICON PACK
# ==============================================================================

# ------------------------------------------------------------------------------
# B1. 6 UPGRADE ICONS (20x20 each)
# ------------------------------------------------------------------------------

ICON_UPG_COOLDOWN_20 = [
    ".......OOOOO........",  # 00
    "......OWGGLGO.......",  # 01  top pusher crown
    ".......OODDO........",  # 02  neck
    "......OOX3XOO..OO...",  # 03  top bezel + side clicker
    "....OO33XXX3XOOXO...",  # 04
    "...O3XXOOOOOXXOO....",  # 05
    "..O3XOOggWggOOXO....",  # 06  dial + 12 o'clock tick
    "..OXOgzzggOOOXX2O...",  # 07  golden lightning cuts in!
    ".O3XOgzzgOLGGOX2O...",  # 08
    ".OXOggWggOLGOgO2O...",  # 09  9 o'clock tick + bolt
    ".OXOggggOLGGOgO2O...",  # 10
    ".OXOggggOLGGGOO2O...",  # 11  wide lightning zig-zag
    ".O2OggggOOOGGOO2O...",  # 12
    "..O2OgzzggOGDOX2O...",  # 13
    "..O2XOOgggODOOX2O...",  # 14  lightning tip
    "...O2XXOOODOO22O....",  # 15
    "....OO222OOO22O.....",  # 16
    "......OOOOOOO.......",  # 17
    "....................",  # 18
    "....................",  # 19
]

ICON_UPG_MULTIHOOK_20 = [
    ".......OO.OO........",  # 00  floating attracted hearts!
    "......OWROKRO.......",  # 01
    "..OO..ORRRRRRO..OO..",  # 02
    ".OWKO..ORRRRO..OWKO.",  # 03
    ".OKKKO..ORRO...OKKKO",  # 04
    "..OKO....OO.....OKO.",  # 05
    "...O...X....X....O..",  # 06  electric attraction sparks
    "...OOOO......OOOO...",  # 07  silver magnetic pole tips
    "..OWwwMO....OWwwMO..",  # 08
    "..OMMMdO....OMMMdO..",  # 09
    "..OOOOOO....OOOOOO..",  # 10  divider seam
    "..OQKKkO....OQKKkO..",  # 11  glowing neon-pink U body
    "..OKKKkO....OKKKkO..",  # 12
    "..OKKKkO....OKKKkO..",  # 13
    "..OQKkkOOOOOOQKkkO..",  # 14
    "...OKKKKkkkkKKKKO...",  # 15
    "....OQKKKKKKKKkO....",  # 16
    ".....OOKKKKKkOO.....",  # 17
    ".......OOOOOO.......",  # 18
    "....................",  # 19
]

ICON_UPG_MAXSTOPS_20 = [
    "...............OO...",  # 00  golden '+' sparkle top-right
    "..OOOOOOO.....OLGO..",  # 01  back stop sign (left)
    ".O3XXXXX2O...OOLGOO.",  # 02
    ".OXJJJJJ2O..OLGGGDGO",  # 03
    ".OXJWWWJ2O...OODDOO.",  # 04  mini bus glyph on back sign
    ".OXJBBBJ2O....ODDO..",  # 05
    ".O2OOOOOOO.....OO...",  # 06
    "..O2O3XXXXX2O.......",  # 07  front overlapping neon stop sign
    "...OOXFFFFFXO.......",  # 08  coral/cream pop-urban stop face
    "...OOXFWWWF2O.......",  # 09  crisp white/cyan mini bus emblem
    "...OOXFXWXF2O.......",  # 10
    "...OOXFGGGF2O.......",  # 11
    "...OO2XFFF22O.......",  # 12
    "...OO.OOOOOO........",  # 13  back pole + front pole
    "...OMO..OWMO........",  # 14
    "...OdO..OMdO........",  # 15
    "...OMO..OMdO........",  # 16
    "..OOMOO.OMdO........",  # 17
    "..OOOOOOMMMdO.......",  # 18
    ".......OOOOOO.......",  # 19
]

ICON_UPG_RANGE_20 = [
    ".........OO.........",  # 00  clean cardinal gold ticks
    "......OOOGGOOO......",  # 01
    "....OO3XXGGXX3OO....",  # 02  crisp outer cyan radar ring
    "...O3XOOOGGOOOX3O...",  # 03
    "..O3XOggcGGc3XXO2O..",  # 04  clean radar sweep in top-right
    "..OXOgOOXXXX33OgO2O.",  # 05  inner concentric scope ring
    ".O3XOgXOOOOOO3XgO2O.",  # 06
    ".OXOgXOgccc3XXOgO2O.",  # 07
    ".OGGXOggOOGO3XOgGGO.",  # 08  clean horizontal crosshair
    "OGGGOOggOLGGOXOOOGGO",  # 09  dead-center golden bullseye!
    "OGGGOOggOGDGOgOOOGGO",  # 10
    ".OGGXOggOOGOggOgGGO.",  # 11
    ".OXOgXOgccccggOgO2O.",  # 12
    ".O2XOgXOOOOOOXOgO2O.",  # 13
    "..O2OgOOXXXXOOggO2O.",  # 14
    "..O2XOggcGGcggOX2O..",  # 15
    "...O2XOOOGGOOOX2O...",  # 16
    "....OO222GG222OO....",  # 17
    "......OOOGGOOO......",  # 18
    ".........OO.........",  # 19
]

ICON_UPG_CHRONOS_20 = [
    "...OOOOOOOOOOOOOO...",  # 00  ornate violet-gold top cap
    "..OZUUUGGGGGGUUUuO..",  # 01
    "...OOOOOZZZZOOOOO...",  # 02
    "....OgWzzzzzzzgO....",  # 03  top glass bulb
    "....OgW33XX22ggO....",  # 04  glowing cyan neon sand
    ".....Og3XXXX2gO.....",  # 05
    "......Og3XX2gO......",  # 06
    "..OOOO.OgX2gO.OOOO..",  # 07  sweeping cyan/violet time-slow orbit ring
    ".O3XXOO.O3XO.OOXX2O.",  # 08  pinched hourglass waist
    ".O2X2OO.O3XO.OOX3XO.",  # 09  trickling neon sand beam
    "..OOOO.Og3XgO.OOOO..",  # 10
    "......OgW3XggO......",  # 11  bottom glass bulb
    ".....OgWgg3XggO.....",  # 12
    "....OgWg33XX2ggO....",  # 13  accumulating cyan sand mound
    "....Og33XXXXX22O....",  # 14
    "....O3XXXXXXXX2O....",  # 15
    "...OOOOOZZZZOOOOO...",  # 16  ornate bottom cap
    "..OZUUUGGGGGGUUUuO..",  # 17
    "...OOOOOOOOOOOOOO...",  # 18
    "....................",  # 19
]

ICON_UPG_VITALITY_20 = [
    "....W.....OOO.OOO...",  # 00  rising steam + bright glossy red heart
    "...W.....OWFFOFRFO..",  # 01
    "....W...OWFFFRRRRRO.",  # 02
    ".....W..OFRRRRRRRrO.",  # 03
    "....W...OFRRRRRRRrO.",  # 04
    ".........OFRRRRRrO..",  # 05
    "..OOOOOOO.OFRRRrO...",  # 06  cup rim + heart tip
    ".OCTT11TCO.OFRrO....",  # 07  dark espresso + golden crema
    ".OCtTTTTaO..ORO.....",  # 08
    ".OWCCCCCaOOO.O......",  # 09  cup body + handle
    ".OWRRRRRrOCaO.......",  # 10  vibrant red racing stripe
    ".OWRRRRRrO.OaO......",  # 11
    ".OWCCCCCaO.OaO......",  # 12
    ".OWCCCCCaOOaO.......",  # 13
    "..OCCCCaaOOO........",  # 14
    ".OOOOOOOOOOO........",  # 15  clean connected saucer
    ".OOWCCCCCaaO........",  # 16
    "..OOOOOOOOO.........",  # 17
    "....................",  # 18
    "....................",  # 19
]

# ------------------------------------------------------------------------------
# B2. 6 RELIC ICONS (16x16 each)
# ------------------------------------------------------------------------------

ICON_RELIC_HEADPHONES_16 = [
    "....OOOOOOOO....",  # 00
    "..OOZZUUUUuuOO..",  # 01  sleek violet arched headband
    ".OZZUOOOOOOuUuO.",  # 02
    ".OZUO......OuUO.",  # 03
    "OZUO........OuUO",  # 04
    "OZUO........OuUO",  # 05
    "OOXO........OXOO",  # 06  cyan earcup joints
    "O3XOOP....POOX2O",  # 07  left & right earcups w/ cyan LED ring
    "O3XOPpO..OpPOX2O",  # 08
    "OX3OPpO..OpPO2XO",  # 09
    "OXWOPpO..OpPO2XO",  # 10
    "O2XOPpO..OpPO22O",  # 11
    "O22OOP....POO22O",  # 12
    ".OOOO......OOOO.",  # 13
    "................",  # 14
    "................",  # 15
]

ICON_RELIC_PASS_16 = [
    "................",  # 00
    "................",  # 01
    ".OOOOOOOOOOOOOO.",  # 02  golden monthly bus pass card
    "OLGGGGGGG3XKGbDO",  # 03  diagonal hologram stripe!
    "OGOOOGGG3XKGbGDO",  # 04
    "OGOBOGG3XKGGDODO",  # 05  photo square + stripe
    "OGOOOG3XKGGDDDDO",  # 06
    "OGGGGWXKGGOOOODO",  # 07  pass info bars
    "OGGG3XKGGGDDDDDO",  # 08
    "OGG3XKGGGOOOOODO",  # 09
    "OGGDDDDDDDDDDDDO",  # 10
    "ODbbbbbbbbbbbbbO",  # 11
    ".OOOOOOOOOOOOOO.",  # 12
    "................",  # 13
    "................",  # 14
    "................",  # 15
]

ICON_RELIC_SHIELD_16 = [
    ".....OOOOOO.....",  # 00  armored glass heater shield
    "..OOOWwwwwMOOO..",  # 01
    ".OWwwOOOOOOdMMO.",  # 02
    ".OwOO333XXXOOdO.",  # 03  translucent cyan armored glass
    ".OwO3WW3XXXXOdO.",  # 04  bold diagonal white reflection
    ".OwO3W3XXXXXOdO.",  # 05
    ".OwO33XXX3X2OdO.",  # 06
    ".OMOXXXX3X22OdO.",  # 07
    "..OMOX3XX22OdO..",  # 08
    "..OMOX3X22cOdO..",  # 09
    "...OMOXX2cOdO...",  # 10
    "....OMOXcOdO....",  # 11
    ".....OMOOdO.....",  # 12
    "......OMdO......",  # 13
    ".......OO.......",  # 14
    "................",  # 15
]

ICON_RELIC_CLOVER_16 = [
    "...........W....",  # 00  gold sparkle star top-right
    "...OOO....WLG...",  # 01
    "..OlYYO..OOOW...",  # 02  top-left & top-right clover leaves
    ".OlYYYO.OlYYO...",  # 03
    ".OYYYYYOYYYYO...",  # 04
    "..OYYYYYYYYO....",  # 05
    "...OOYyyYOO.....",  # 06  center stem hub
    "..OllYYYYYYO....",  # 07  bottom-left & bottom-right leaves
    ".OlYYYOYYYYyO...",  # 08
    ".OYYYYO.OYYyO...",  # 09
    "..OYYyO.OyyOO...",  # 10
    "...OOO...OeyO...",  # 11  curved lucky stem
    "..........OeyO..",  # 12
    "...........OeyO.",  # 13
    "............OO..",  # 14
    "................",  # 15
]

ICON_RELIC_ESPRESSO_16 = [
    "....W.....W.....",  # 00  twin rising steam wisps
    "...W.......W....",  # 01
    "....W.....W.....",  # 02
    "..OOOOOO........",  # 03  back-left demitasse cup w/ left handle
    ".OC11TTaO.......",  # 04  golden crema + dark roast
    "OOOWCCCaaOOOOO..",  # 05  front-right cup w/ right handle
    "OaOWTTTaOC11TTaO",  # 06
    "OOOWCCCaaOWCCCaO",  # 07
    "..OCCCCaOOWTTTaO",  # 08  rich espresso band + handles
    "..OOOOOO.OWCCCaO",  # 09
    "........OOWCCCaO",  # 10
    "........OCCCCaaO",  # 11
    ".......OOaaaaaOO",  # 12  saucer
    "........OOOOOOO.",  # 13
    "................",  # 14
    "................",  # 15
]

ICON_RELIC_VIP_16 = [
    ".......OO.......",  # 00  center peak pearl
    "..OO..OWLO..OO..",  # 01  left & right peak pearls
    ".OWLO.OLLO.OWLO.",  # 02
    "..OO..OGGO..OO..",  # 03
    ".OGGO.OGGO.OGGO.",  # 04  3-point golden VIP crown (crisp symmetric edge)
    ".OGGGOOGGOOGGGO.",  # 05
    ".OGGGGOOOGGGDDO.",  # 06
    ".OGGGGOqKOGGGDO.",  # 07  glowing magenta diamond gem!
    ".OGXXOqKKKOGXDO.",  # 08  cyan side studs + magenta gem
    ".OGGGGOkKkOGGDO.",  # 09
    ".ODDDDDODDDDDDO.",  # 10
    ".OORRRRRRRRRROO.",  # 11  crimson royal base band
    "..OGGGGGGGGGDO..",  # 12
    "..OOOOOOOOOOOO..",  # 13
    "................",  # 14
    "................",  # 15
]

# ------------------------------------------------------------------------------
# B3. 4 VENDING SHOP ICONS (16x16 each)
# ------------------------------------------------------------------------------

ICON_SHOP_ESPRESSO_16 = [
    "..OO.OO....W....",  # 00  red heart steam + '+' sparkle
    ".OWROFRRO.WFW...",  # 01
    ".OFRRRRRRO.W....",  # 02
    "..OFRRRrO.......",  # 03
    "...OFRrO...W....",  # 04
    "....OO....W.....",  # 05
    ".OOOOOOOOO......",  # 06  hot espresso cup
    "OC11TTTT1CO.....",  # 07  golden crema + roast
    "OWCCCCCCCaOOO...",  # 08  cup + curved handle
    "OWRRRRRRRaO.aO..",  # 09
    "OWCCCCCCCaO.aO..",  # 10
    "OWCCCCCCCaOOaO..",  # 11
    ".OCCCCCCaaOOO...",  # 12
    ".OOOOOOOOOOO....",  # 13  clean connected saucer
    ".OOWCCCCCCaO....",  # 14
    "..OOOOOOOOO.....",  # 15
]

ICON_SHOP_BRIBE_16 = [
    "................",  # 00
    ".OOOOOOOOOOOO...",  # 01  crisp emerald cash bill / envelope
    "OllYYYYYYYYyYO..",  # 02
    "OlYCCCCCCYYyYO..",  # 03
    "OlYCYCCYCYYyYO..",  # 04  crisp dollar emblem
    "OlYCCYYCCYYOO...",  # 05  gold police star overlapping bottom-right
    "OlYCYCCYCORGOO..",  # 06  bold crimson slash crossing star!
    "OlYCCCCCOFRGGGO.",  # 07
    "OlYYYYYYOORGGDO.",  # 08
    "OyyyyyyyOGORRO..",  # 09
    ".OOOOOOOOGGORRO.",  # 10
    "........ODDOORO.",  # 11
    "........OO..OOO.",  # 12
    "................",  # 13
    "................",  # 14
    "................",  # 15
]

ICON_SHOP_RELIC_16 = [
    ".W..OOO..OOO..L.",  # 00  fluffy golden bow + magical corner sparkles
    "...OLGGOOGGLO...",  # 01
    "...OGDGOOGDGO...",  # 02
    ".OOOOOOOOOOOOOO.",  # 03  gift box lid
    "OZZUUUOGGOUUUuuO",  # 04
    "OZZZZZOGGOUUUUUO",  # 05
    ".OOOOOOOOOOOOOO.",  # 06
    "..OZZUOGGOUUuO..",  # 07  purple mystery box w/ gold ribbon
    "..OZUUOGGOUuuO..",  # 08
    "..OOGGGGGGGGOO..",  # 09  horizontal gold ribbon
    "..OODDDDDDDDOO..",  # 10
    "..OZUUOGGOUuuO..",  # 11
    "..OUUuOGGOuuuO..",  # 12
    "..OuuuODDOuuuO..",  # 13
    "..OOOOOOOOOOOO..",  # 14
    "................",  # 15
]

ICON_SHOP_CRATE_16 = [
    "................",  # 00
    "..OOOOOOOOOOOO..",  # 01  golden neon supply crate
    ".OMWGGGGGGGGwMO.",  # 02  metallic corner brackets
    ".OMXGGGOOGGGXMO.",  # 03  cyan neon corner LEDs
    ".OGGGOO33OOGGDO.",  # 04  sharp upward arrow chevron!
    ".OGGO3WWWW3OGDO.",  # 05
    ".OGO3WWXXWW3ODO.",  # 06
    ".OG3WXXXXXXW3DO.",  # 07  wide arrow head
    ".OGOOO3XX3OOODO.",  # 08  clean narrow shaft
    ".OGGGO3XX3OGDDO.",  # 09
    ".OGGGO3XX3OGDDO.",  # 10
    ".OGGGGOOOOGGDDO.",  # 11
    ".OdXDDDDDDDDXdO.",  # 12  cyan neon corner LEDs
    ".OddDDDDDDDDddO.",  # 13
    "..OOOOOOOOOOOO..",  # 14
    "................",  # 15
]

# ------------------------------------------------------------------------------
# B4. 6 RIZZ-SLOTS 777 REEL ICONS (20x20 each)
# ------------------------------------------------------------------------------

ICON_SLOT_CHERRY_20 = [
    ".........OOO........",  # 00  stem apex + lush emerald leaf
    "........OYYlOO......",  # 01
    ".......OYOOllYO.....",  # 02
    "......OYO.OYYYYO....",  # 03
    "......OYO..OYYyO....",  # 04
    ".....OYO.OYOOOO.....",  # 05  twin curved stems
    ".....OYO..OYO.......",  # 06
    "....OYO....OYO......",  # 07
    "....OYO....OYO......",  # 08
    "..OOOOO...OOOOO.....",  # 09  plump twin glossy ruby cherries
    ".OFFRRRO.OFFRRRO....",  # 10
    "OFWWFRRROFWWFRRRO...",  # 11  crisp white specular shines
    "OFWFRRRROFWFRRRrO...",  # 12
    "OFRRRRRROFRRRRRrO...",  # 13
    "OFRRRRRrOFRRRRRrO...",  # 14
    "ORRRRRrrOORRRRrrO...",  # 15
    ".ORRRrrO..ORRRrrO...",  # 16
    "..OOOOO....OOOOO....",  # 17
    "....................",  # 18
    "....................",  # 19
]

ICON_SLOT_CASH_20 = [
    "....................",  # 00
    "....................",  # 01
    "....OOOOOOOOOOOO....",  # 02  fat stack of emerald banknotes
    "...OllYYOGGOYYllO...",  # 03  gold currency wrapper band
    "..OlYYYYOGGOYYYYlO..",  # 04
    ".OlYYCCYOGGOYCCYYlO.",  # 05  top bill seal details
    ".OlYCCYYOGGOYYCCYlO.",  # 06
    ".OYYYYYYOGGOYYYYYYO.",  # 07
    ".OyyyyyyODDOyyyyyyO.",  # 08  layered banknote stack edges
    ".OllYYYYOGGOYYYYlYO.",  # 09
    ".OyyyyyyODDOyyyyyyO.",  # 10
    ".OllYYYYOGGOYYYYlYO.",  # 11
    ".OyyyyyyODDOyyyyyyO.",  # 12
    ".OllYYYYOGGOYYYYlYO.",  # 13
    ".OeeeeeeObbOeeeeeeO.",  # 14
    "..OOOOOOOOOOOOOOOO..",  # 15
    "....................",  # 16
    "....................",  # 17
    "....................",  # 18
    "....................",  # 19
]

ICON_SLOT_XP_20 = [
    ".........OO.........",  # 00  sparkling electric cyan star crystal
    "........O33O........",  # 01
    "..W.....O3XO........",  # 02  corner twinkle
    ".W3W...O3WXXO.......",  # 03
    "..W....O3WXXO.......",  # 04
    "......O3WWXX2O......",  # 05
    "..OOOO3WWXXXX2OOOO..",  # 06
    ".O3333WWXXXXXX2222O.",  # 07
    "..O33WWXXXXXXXX22O..",  # 08
    "...O3WXXXXXXXXX2O...",  # 09
    "....O3XXXXXXXX2O....",  # 10
    "....O3XXX22XXX2O....",  # 11
    "...O3XX22OO22XX2O...",  # 12
    "...O3X22O..O22X2O...",  # 13
    "..O3X22O....O22X2O..",  # 14
    "..O222O..W...O222O..",  # 15  bottom twinkle
    "..OOOO..W3W...OOOO..",  # 16
    ".........W..........",  # 17
    "....................",  # 18
    "....................",  # 19
]

ICON_SLOT_HEART_20 = [
    "....................",  # 00
    "....................",  # 01
    "...OOOOO....OOOOO...",  # 02  big glossy crimson arcade heart
    "..OFFFRRO..OFFFRRO..",  # 03
    ".OFWWFFRROOFFRRRRrO.",  # 04  plump white bubble shine
    ".OFWWFRRROORRRRRRrO.",  # 05
    ".OFWFRRRRRRRRRRRRrO.",  # 06
    ".OFFRRRRRRRRRRRRRrO.",  # 07
    ".OFRRRRRRRRRRRRRrrO.",  # 08
    "..OFRRRRRRRRRRRRrrO.",  # 09
    "..OFRRRRRRRRRRRrrO..",  # 10
    "...OFRRRRRRRRRrrO...",  # 11
    "....OFRRRRRRRrrO....",  # 12
    ".....OFRRRRRrrO.....",  # 13
    "......OFRRRrrO......",  # 14
    ".......OFRrrO.......",  # 15
    "........OrrO........",  # 16
    ".........OO.........",  # 17
    "....................",  # 18
    "....................",  # 19
]

ICON_SLOT_777_20 = [
    "......OOOOOOOO......",  # 00  triple 7 jackpot emblem!
    ".OOOOOLLGGGGGO.OOOO.",  # 01  separated left coral-7, center hero gold-7, right coral-7
    "OFFFFOOGGGGGDOOFFFFO",  # 02
    "OFRRROOOOGGDOOOFRRRO",  # 03
    "OOORRO..OGGDO.OOORRO",  # 04
    "..ORRO..OGDO....ORRO",  # 05
    "..ORRO.OGGDO...ORRO.",  # 06
    ".ORRO..OGDO....ORRO.",  # 07
    ".ORRO.OGGDO...ORRO..",  # 08
    ".ORrO.OGDO....ORrO..",  # 09
    ".ORrO.OGDO...ORrO...",  # 10
    ".ORrO.ODDO...ORrO...",  # 11
    "..OO..ODDO....OO....",  # 12
    "......ODDO..........",  # 13
    "..W...ODDO....W.....",  # 14  jackpot gold sparkles
    ".WGL..OOOO...LGW....",  # 15
    "..W...........W.....",  # 16
    "....................",  # 17
    "....................",  # 18
    "....................",  # 19
]

ICON_SLOT_SIREN_20 = [
    "....W..........W....",  # 00  flashing police siren rays
    ".W..F....OO....X..W.",  # 01
    "..F.....O3WO.....X..",  # 02
    "......OOFWWBOO......",  # 03  rounded flashing red/blue beacon dome
    ".....OFFFWWBBBO.....",  # 04
    "..F.OFWWFWWBWWBO.X..",  # 05  left ruby-red flash, right cobalt-blue flash
    "....OFWFRWWBJWBO....",  # 06
    "....OFRRRWWBJJBO....",  # 07
    "....OFRRRWWBJJBO....",  # 08
    "....OFRRRWWJJJBO....",  # 09
    "....OFRrrWWjjJBO....",  # 10
    "....OFRrrWWjjJBO....",  # 11
    "...OOOOOOOOOOOOOO...",  # 12  heavy silver/charcoal mounting base
    "..OWwwwwMMMMMMdddO..",  # 13
    "..OMMMMMMMMMMMdddO..",  # 14
    "..OPPPPPPPPPPPpppO..",  # 15
    ".OOPPPPPPPPPPPpppOO.",  # 16
    ".OOOOOOOOOOOOOOOOOO.",  # 17
    "....................",  # 18
    "....................",  # 19
]

# ------------------------------------------------------------------------------
# B5. 9 HUD, CONTROLS & WORLD PICKUP ICONS (16x16 & 12x12)
# ------------------------------------------------------------------------------

ICON_MOUSE_LMB_16 = [
    ".......OO.......",  # 00  cord stub
    ".....OOOOOO.....",  # 01
    "...OO3XOOMMOO...",  # 02  LEFT button glowing bright cyan!
    "..O33XXOMwwwMO..",  # 03
    "..O3XXXOOMMMdO..",  # 04
    "..OXXXXOOMMMdO..",  # 05
    "..OXX22OOMMMdO..",  # 06
    "..OOOOOOOOOOOO..",  # 07  button seam
    "..OWwwwwMMMMdO..",  # 08  clean mouse palm body
    "..OwwwwMMMMMdO..",  # 09
    "..OwwwMMMMMMdO..",  # 10
    "..OwwMMMMMMddO..",  # 11
    "...OMMMMMMddO...",  # 12
    "....OMMMdddO....",  # 13
    ".....OOOOOO.....",  # 14
    "................",  # 15
]

ICON_MOUSE_RMB_16 = [
    ".......OO.......",  # 00  cord stub
    ".....OOOOOO.....",  # 01
    "...OOMMOOGLOO...",  # 02  RIGHT button glowing bright gold/coral!
    "..OWwwwMOGGFFO..",  # 03
    "..OMMMMOOOGGFO..",  # 04
    "..OMMMMOOOGGDO..",  # 05
    "..OMdddOOODDFO..",  # 06
    "..OOOOOOOOOOOO..",  # 07
    "..OWwwwwMMMMdO..",  # 08
    "..OwwwwMMMMMdO..",  # 09
    "..OwwwMMMMMMdO..",  # 10
    "..OwwMMMMMMddO..",  # 11
    "...OMMMMMMddO...",  # 12
    "....OMMMdddO....",  # 13
    ".....OOOOOO.....",  # 14
    "................",  # 15
]

ICON_MOUSE_WHEEL_16 = [
    ".......OO.......",  # 00  cord stub
    ".....OOGGOO.....",  # 01  top glow above scroll wheel
    "...OOMOWKOMOO...",  # 02  SCROLL WHEEL glowing magenta/gold!
    "..OWwwOQKOwwMO..",  # 03
    "..OMMMOKGOMMdO..",  # 04
    "..OMMMOKKOMMdO..",  # 05
    "..OMddOGGOmmdO..",  # 06
    "..OOOOOOOOOOOO..",  # 07
    "..OWwwwwMMMMdO..",  # 08
    "..OwwwwMMMMMdO..",  # 09
    "..OwwwMMMMMMdO..",  # 10
    "..OwwMMMMMMddO..",  # 11
    "...OMMMMMMddO...",  # 12
    "....OMMMdddO....",  # 13
    ".....OOOOOO.....",  # 14
    "................",  # 15
]

ICON_HEART_FULL_12 = [
    "............",  # 00
    "..OOO..OOO..",  # 01
    ".OFFROOFFRO.",  # 02
    "OFWFRROFRRrO",  # 03
    "OFRRRRRRRRrO",  # 04
    "OFRRRRRRRrrO",  # 05
    ".OFRRRRRrrO.",  # 06
    "..OFRRRrrO..",  # 07
    "...OFRrrO...",  # 08
    "....OrrO....",  # 09
    ".....OO.....",  # 10
    "............",  # 11
]

ICON_HEART_EMPTY_12 = [
    "............",  # 00
    "..OOO..OOO..",  # 01
    ".OooO..OooO.",  # 02  exact silhouette match to icon_heart_full
    "OohhoOOohhoO",  # 03  dark plum rim + translucent interior
    "OohhhhhhhhoO",  # 04
    "OohhhhhhhhoO",  # 05
    ".OohhhhhhoO.",  # 06
    "..OohhhhoO..",  # 07
    "...OohhoO...",  # 08
    "....OooO....",  # 09
    ".....OO.....",  # 10
    "............",  # 11
]

ICON_STAR_WANTED_12 = [
    ".....OO.....",  # 00  crisp glowing gold-amber 5-point Wanted Star
    "....OLLO....",  # 01
    "....OWGO....",  # 02
    ".OOOOLGOOOO.",  # 03
    "OLGGGGGGGGDO",  # 04
    ".OOGGGGGDOO.",  # 05
    "..OGGGGGDO..",  # 06
    "..OGGDDGDO..",  # 07
    ".OGGDOODGDO.",  # 08
    ".OGDO..ODDO.",  # 09
    ".OOO....OOO.",  # 10
    "............",  # 11
]

ICON_COIN_12 = [
    "...OOOOOO...",  # 00  crisp shiny gold '$' coin
    "..OLLGGDDO..",  # 01
    ".OLGGbWGGDO.",  # 02  vertical top stem of '$'
    "OLGbWWWGbGDO",  # 03  top bar of '$'
    "OGGbWGGbGGDO",  # 04
    "OGGGWWWGGGDO",  # 05  middle bar of '$'
    "OGGGbGGWGGDO",  # 06
    "OGGbWWWGbGDO",  # 07  bottom bar of '$'
    "ODGGGbWGGDDO",  # 08  bottom stem of '$'
    ".ODGGGGGDDO.",  # 09
    "..ODDDDDDO..",  # 10
    "...OOOOOO...",  # 11
]

ICON_BOLT_12 = [
    "......OOOO..",  # 00  crisp electric cyan Chronos lightning bolt
    ".....O3XXO..",  # 01
    "....O3WXO...",  # 02
    "...O3WXO....",  # 03
    "..O3WXXOOOO.",  # 04
    ".O3WXXXXXX2O",  # 05
    ".OOOOOXX22O.",  # 06
    ".....OXX2O..",  # 07
    "....O3X2O...",  # 08
    "...O3X2O....",  # 09
    "...O22O.....",  # 10
    "...OOO......",  # 11
]

ICON_PICKUP_HEAL_16 = [
    "....+++..+++....",  # 00  luminous mint-white aura for instant asphalt pop!
    "...+*OO**OO*+...",  # 01
    "..+*OWR**RWO*+..",  # 02  cute floating mini ruby-red heart above cup
    "..+*OFRRRRFO*+..",  # 03
    "..++*OFRRFO*++..",  # 04
    ".+****OOOO****+.",  # 05  bright white inner halo around espresso-med cup
    "+*OOOOOOOOOOOO*+",  # 06
    "+*O11TTTTTT1CO*+",  # 07  golden crema + espresso inside white med cup
    "+*OWWWRRWWWWCO*+",  # 08  bright glossy red medical cross on cup front!
    "+*OWWRRRRWWCOO*+",  # 09
    "+*OWWWRRWWCOmO*+",  # 10  clean mint-white handle
    "+*OWWWmmWWCOO*+.",  # 11
    ".+*OmmmmmmOO*+..",  # 12  clean mint saucer trim
    "..+**OOOO**++...",  # 13
    "...+++++++++....",  # 14
    "................",  # 15
]


# ==============================================================================
# FIXUP PASSES & CLEANUPS
# ==============================================================================

def build_icon_mouse_wheel_16() -> Canvas:
    """Build icon_mouse_wheel with crisp silver body and bright magenta/gold scroll wheel."""
    grid = [
        ".......OO.......",  # 00
        ".....OOGGOO.....",  # 01  top golden wheel highlight
        "...OOMOWKOMOO...",  # 02  bright magenta/gold glowing scroll wheel
        "..OWwwOQKOwwMO..",  # 03
        "..OMMMOKGOMMdO..",  # 04
        "..OMMMOKKOMMdO..",  # 05
        "..OMddOGGOmmdO..",  # 06
        "..OOOOOOOOOOOO..",  # 07
        "..OWwwwwMMMMdO..",  # 08
        "..OwwwwMMMMMdO..",  # 09
        "..OwwwMMMMMMdO..",  # 10
        "..OwwMMMMMMddO..",  # 11
        "...OMMMMMMddO...",  # 12
        "....OMMMdddO....",  # 13
        ".....OOOOOO.....",  # 14
        "................",  # 15
    ]
    # Keep row 6 right button clean silver-shadow (no mint bleed)
    grid[6] = "..OMddOGGOdddO.."
    return grid_to_canvas(grid, 16, 16, UI_PALETTE)


# ==============================================================================
# MAIN ASSET GENERATION & VERIFICATION
# ==============================================================================

def main() -> None:
    os.makedirs(MAP_OUT_DIR, exist_ok=True)
    os.makedirs(UI_OUT_DIR, exist_ok=True)

    # Part A — 5 Seamless Background/Ground Tiles (32x32)
    map_tiles: List[Tuple[str, Canvas]] = [
        ("tile_plaza_mosaic", build_tile_plaza_mosaic_32()),
        ("tile_wet_asphalt", build_tile_wet_asphalt_32()),
        ("tile_park_path", build_tile_park_path_32()),
        ("tile_bike_lane", build_tile_bike_lane_32()),
        ("tile_canal_edge", build_tile_canal_edge_32()),
    ]

    for name, canvas in map_tiles:
        assert (canvas.width, canvas.height) == (32, 32), f"{name} must be 32x32"
        save_asset_and_preview(canvas, MAP_OUT_DIR, name, zoom=6)
        print(f"[MAP] Generated {name}.png ({canvas.width}x{canvas.height})")

    # Part B — 31 Crisp Pixel-Art UI & Pickup Icons
    ui_icons: List[Tuple[str, Canvas, int]] = [
        # 6 Upgrade Icons (20x20)
        ("icon_upg_cooldown", grid_to_canvas(ICON_UPG_COOLDOWN_20, 20, 20, UI_PALETTE), 8),
        ("icon_upg_multihook", grid_to_canvas(ICON_UPG_MULTIHOOK_20, 20, 20, UI_PALETTE), 8),
        ("icon_upg_maxstops", grid_to_canvas(ICON_UPG_MAXSTOPS_20, 20, 20, UI_PALETTE), 8),
        ("icon_upg_range", grid_to_canvas(ICON_UPG_RANGE_20, 20, 20, UI_PALETTE), 8),
        ("icon_upg_chronos", grid_to_canvas(ICON_UPG_CHRONOS_20, 20, 20, UI_PALETTE), 8),
        ("icon_upg_vitality", grid_to_canvas(ICON_UPG_VITALITY_20, 20, 20, UI_PALETTE), 8),
        # 6 Relic Icons (16x16)
        ("icon_relic_headphones", grid_to_canvas(ICON_RELIC_HEADPHONES_16, 16, 16, UI_PALETTE), 10),
        ("icon_relic_pass", grid_to_canvas(ICON_RELIC_PASS_16, 16, 16, UI_PALETTE), 10),
        ("icon_relic_shield", grid_to_canvas(ICON_RELIC_SHIELD_16, 16, 16, UI_PALETTE), 10),
        ("icon_relic_clover", grid_to_canvas(ICON_RELIC_CLOVER_16, 16, 16, UI_PALETTE), 10),
        ("icon_relic_espresso", grid_to_canvas(ICON_RELIC_ESPRESSO_16, 16, 16, UI_PALETTE), 10),
        ("icon_relic_vip", grid_to_canvas(ICON_RELIC_VIP_16, 16, 16, UI_PALETTE), 10),
        # 4 Vending Shop Icons (16x16)
        ("icon_shop_espresso", grid_to_canvas(ICON_SHOP_ESPRESSO_16, 16, 16, UI_PALETTE), 10),
        ("icon_shop_bribe", grid_to_canvas(ICON_SHOP_BRIBE_16, 16, 16, UI_PALETTE), 10),
        ("icon_shop_relic", grid_to_canvas(ICON_SHOP_RELIC_16, 16, 16, UI_PALETTE), 10),
        ("icon_shop_crate", grid_to_canvas(ICON_SHOP_CRATE_16, 16, 16, UI_PALETTE), 10),
        # 6 Rizz-Slots 777 Reel Icons (20x20)
        ("icon_slot_cherry", grid_to_canvas(ICON_SLOT_CHERRY_20, 20, 20, UI_PALETTE), 8),
        ("icon_slot_cash", grid_to_canvas(ICON_SLOT_CASH_20, 20, 20, UI_PALETTE), 8),
        ("icon_slot_xp", grid_to_canvas(ICON_SLOT_XP_20, 20, 20, UI_PALETTE), 8),
        ("icon_slot_heart", grid_to_canvas(ICON_SLOT_HEART_20, 20, 20, UI_PALETTE), 8),
        ("icon_slot_777", grid_to_canvas(ICON_SLOT_777_20, 20, 20, UI_PALETTE), 8),
        ("icon_slot_siren", grid_to_canvas(ICON_SLOT_SIREN_20, 20, 20, UI_PALETTE), 8),
        # 9 HUD, Controls & World Pickup Icons
        ("icon_mouse_lmb", grid_to_canvas(ICON_MOUSE_LMB_16, 16, 16, UI_PALETTE), 10),
        ("icon_mouse_rmb", grid_to_canvas(ICON_MOUSE_RMB_16, 16, 16, UI_PALETTE), 10),
        ("icon_mouse_wheel", build_icon_mouse_wheel_16(), 10),
        ("icon_heart_full", grid_to_canvas(ICON_HEART_FULL_12, 12, 12, UI_PALETTE), 12),
        ("icon_heart_empty", grid_to_canvas(ICON_HEART_EMPTY_12, 12, 12, UI_PALETTE), 12),
        ("icon_star_wanted", grid_to_canvas(ICON_STAR_WANTED_12, 12, 12, UI_PALETTE), 12),
        ("icon_coin", grid_to_canvas(ICON_COIN_12, 12, 12, UI_PALETTE), 12),
        ("icon_bolt", grid_to_canvas(ICON_BOLT_12, 12, 12, UI_PALETTE), 12),
        ("icon_pickup_heal", grid_to_canvas(ICON_PICKUP_HEAL_16, 16, 16, UI_PALETTE), 10),
    ]

    for name, canvas, zoom in ui_icons:
        save_asset_and_preview(canvas, UI_OUT_DIR, name, zoom=zoom)
        print(f"[UI ] Generated {name}.png ({canvas.width}x{canvas.height})")

    # Build visual review contact sheets in tools/previews/ui/ (for artist QA)
    sheet_dir = os.path.join(ROOT_DIR, "tools", "previews", "ui")
    os.makedirs(sheet_dir, exist_ok=True)

    # 1. Tile seamless 2x2 verification strip (5 tiles tiled 2x2 each)
    tile_sheet = Canvas(5 * 68 + 8, 72, bg=hex_to_rgba("#1C1124"))
    for idx, (_, tcanvas) in enumerate(map_tiles):
        ox = 4 + idx * 68
        for ty in range(2):
            for tx in range(2):
                tile_sheet.blit(tcanvas, ox + tx * 32, 4 + ty * 32)
    tile_sheet.scale_nearest(3).save(os.path.join(sheet_dir, "sheet_tiles_2x2_preview.png"))

    # 2. UI Icon pack contact sheet on dark evening asphalt (#4E4A60)
    cols = 8
    rows = (len(ui_icons) + cols - 1) // cols
    cell = 26
    ui_sheet = Canvas(cols * cell + 8, rows * cell + 8, bg=hex_to_rgba("#4E4A60"))
    for idx, (_, icanvas, _) in enumerate(ui_icons):
        col = idx % cols
        row = idx // cols
        cx = 4 + col * cell + (cell - icanvas.width) // 2
        cy = 4 + row * cell + (cell - icanvas.height) // 2
        ui_sheet.blit(icanvas, cx, cy)
    ui_sheet.scale_nearest(4).save(os.path.join(sheet_dir, "sheet_ui_icons_preview.png"))

    print(
        f"\nSuccessfully generated {len(map_tiles)} map tiles in {MAP_OUT_DIR} "
        f"and {len(ui_icons)} UI icons in {UI_OUT_DIR} (total {len(map_tiles) + len(ui_icons)} assets)."
    )


if __name__ == "__main__":
    main()
