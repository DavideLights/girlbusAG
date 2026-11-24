#!/usr/bin/env python3
"""
GIRLBUS: NEON STOP (girlbusAG) — Strategic Urban Props & Street-Life Pixel-Art Generator (v2).

Standalone generator that imports Canvas, PALETTE, MAP_PALETTE, and procedural helpers
from pixel_canvas.py and gen_layered_map_assets.py without modifying them.

Generates 9 crisp Pop-Urban Tokyo/Roma props into assets/sprites/map/
(plus zoomed nearest-neighbor previews in tools/previews/map/):
  1. prop_phone_booth.png      (24x40)
  2. prop_food_cart.png        (48x36)
  3. prop_kiosk_umbrella.png   (40x36)
  4. prop_traffic_light.png    (24x48)
  5. prop_bollard_glow.png     (16x24)
  6. prop_planter_neon.png     (48x24)
  7. prop_vending_drink.png    (28x36)
  8. prop_bus_shelter_deco.png (48x36)
  9. car_van_cyber.png         (64x36)
"""

import os
import sys
from typing import Dict, List, Optional, Tuple

TOOLS_DIR = os.path.dirname(os.path.abspath(__file__))
ROOT_DIR = os.path.dirname(TOOLS_DIR)
if TOOLS_DIR not in sys.path:
    sys.path.insert(0, TOOLS_DIR)

from pixel_canvas import Canvas, PALETTE, RGBA, hex_to_rgba, save_asset_and_preview
from gen_layered_map_assets import (
    MAP_PALETTE,
    OUT_DIR,
    apply_1px_outline,
    draw_box,
    draw_leaf_cluster,
    draw_striped_awning,
    draw_text,
    fill_ellipse,
)

# ==============================================================================
# EXTENDED POP-URBAN PROPS PALETTE (Harmonized with Master #1C1124 Outline)
# ==============================================================================
P: Dict[str, RGBA] = {
    **PALETTE,
    **MAP_PALETTE,
    # Hammered Copper Ramp (Street-food takoyaki grill & brass fittings)
    "copper_hi": hex_to_rgba("#FFB584"),
    "copper_base": hex_to_rgba("#D9723B"),
    "copper_sh": hex_to_rgba("#99441E"),
    # Retro Sky-Blue Ramp (Japanese Drink Vending Machine)
    "sky_hi": hex_to_rgba("#9BE6FF"),
    "sky_base": hex_to_rgba("#4BB5F0"),
    "sky_sh": hex_to_rgba("#247EB8"),
    "ice_glow": hex_to_rgba("#DFF9FF"),
    "ice_mid": hex_to_rgba("#A8ECFF"),
    # Vivid Traffic & Neon LEDs
    "led_red": hex_to_rgba("#FF3358"),
    "led_red_hi": hex_to_rgba("#FFA8BC"),
    "led_amber": hex_to_rgba("#FFB833"),
    "led_green": hex_to_rgba("#2BF088"),
    "led_green_hi": hex_to_rgba("#B8FFD8"),
    "led_green_sh": hex_to_rgba("#149E5A"),
    # Midnight transit display & tinted shelter glass
    "screen_bg": hex_to_rgba("#151E33"),
    "screen_grid": hex_to_rgba("#22304F"),
    "glass_tint": hex_to_rgba("#2A3B54", 190),
    "glass_sheen": hex_to_rgba("#6E94BD", 175),
    # Translucent Ambient Light Glows & Steam
    "glow_cyan": hex_to_rgba("#00F0FF", 52),
    "glow_cyan_soft": hex_to_rgba("#52F7FF", 38),
    "glow_magenta": hex_to_rgba("#FF4D94", 46),
    "glow_warm": hex_to_rgba("#FFD166", 70),
    "glow_warm_soft": hex_to_rgba("#FFD166", 44),
    "glow_amber": hex_to_rgba("#FF9E3D", 62),
    "glow_red": hex_to_rgba("#FF3358", 55),
    "glow_green": hex_to_rgba("#2BF088", 52),
    "glow_sky": hex_to_rgba("#4BB5F0", 52),
    "steam_hi": hex_to_rgba("#FFFFFF", 215),
    "steam_soft": hex_to_rgba("#D8EEFF", 145),
    "booth_warm_hi": hex_to_rgba("#FFF4D4"),
    "booth_warm": hex_to_rgba("#FFD98A"),
    "booth_warm_sh": hex_to_rgba("#D99B52"),
}


# ==============================================================================
# LOCAL PIXEL-CRAFT HELPERS
# ==============================================================================

def clear_pixel(c: Canvas, x: int, y: int) -> None:
    if 0 <= x < c.width and 0 <= y < c.height:
        c.pixels[y][x] = (0, 0, 0, 0)


def draw_rounded_box(
    c: Canvas,
    x: int,
    y: int,
    w: int,
    h: int,
    fill: RGBA,
    outline: RGBA = P["O"],
    hi: Optional[RGBA] = None,
    sh: Optional[RGBA] = None,
    round_top: bool = True,
    round_bottom: bool = False,
) -> None:
    """Draw a pixel-clean box with optional 1px chamfered corners and clean 1px outline."""
    draw_box(c, x, y, w, h, fill, outline, hi=hi, sh=sh)
    if round_top and w >= 4 and h >= 4:
        clear_pixel(c, x, y)
        clear_pixel(c, x + w - 1, y)
        c.set_pixel(x + 1, y + 1, outline)
        c.set_pixel(x + w - 2, y + 1, outline)
    if round_bottom and w >= 4 and h >= 4:
        clear_pixel(c, x, y + h - 1)
        clear_pixel(c, x + w - 1, y + h - 1)
        c.set_pixel(x + 1, y + h - 2, outline)
        c.set_pixel(x + w - 2, y + h - 2, outline)


# ==============================================================================
# 1. prop_phone_booth.png (24x40)
# ==============================================================================

def build_prop_phone_booth_24x40() -> Canvas:
    """
    prop_phone_booth.png (24x40) — Glowing retro-pop cyan/magenta glass phone booth
    with warm interior light, roof dome 'TEL' sign, payphone silhouette & crisp glass highlights.
    """
    c = Canvas(24, 40)

    # Ground drop shadow + subtle cyan sidewalk glow spill
    fill_ellipse(c, 12.0, 37.5, 9.0, 2.0, P["s"])
    fill_ellipse(c, 12.0, 36.8, 9.6, 2.2, P["glow_cyan_soft"])

    # Soft magenta/cyan aura behind roof dome sign
    fill_ellipse(c, 12.0, 5.5, 8.8, 4.5, P["glow_magenta"])

    # Roof dome finial & illuminated 'TEL' roof sign (x=4..19, y=1..9)
    draw_rounded_box(c, 7, 1, 10, 3, P["K"], P["O"], hi=P["Q"], sh=P["k"], round_top=True)
    draw_rounded_box(c, 4, 2, 16, 7, P["K"], P["O"], hi=P["Q"], sh=P["k"], round_top=True)
    # Recessed dark marquee box with crisp 3x5 glowing cyan 'TEL'
    c.fill_rect(5, 3, 14, 5, P["screen_bg"])
    draw_text(c, 7, 3, "TEL", P["X"], shadow=None, spacing=1)
    # Tiny glowing white top-left glint on T & L
    c.set_pixel(7, 3, P["3"])
    c.set_pixel(11, 3, P["3"])
    c.set_pixel(15, 3, P["3"])

    # Cyan roof cornice overhang (x=3..20, y=9..12)
    draw_box(c, 3, 9, 18, 3, P["X"], P["O"], hi=P["3"], sh=P["2"])

    # Main booth outer frame (x=4..19, y=11..35): vibrant Cyan/Teal pillars + Magenta trims
    draw_box(c, 4, 11, 16, 24, P["2"], P["O"], hi=P["X"], sh=P["teal_sh"])
    c.fill_rect(5, 12, 1, 21, P["3"])          # crisp left cyan pillar highlight
    c.fill_rect(18, 12, 1, 21, P["2"])         # crisp right cyan pillar
    # Magenta upper & lower trim bars framing the glass
    c.fill_rect(5, 12, 14, 1, P["K"])
    c.fill_rect(5, 31, 14, 1, P["K"])

    # Glowing warm interior glass enclosure (x=6..17, y=13..31)
    c.fill_rect(6, 13, 12, 18, P["booth_warm"])
    # Bright ceiling lamp glow inside booth
    c.fill_rect(7, 13, 10, 4, P["booth_warm_hi"])
    c.fill_rect(9, 13, 6, 2, P["W"])
    # Lower warm shadow gradient inside glass
    c.fill_rect(6, 26, 12, 5, P["booth_warm_sh"])
    # Glass edge tint inside cyan frame
    c.fill_rect(6, 13, 1, 18, P["X"])
    c.fill_rect(17, 13, 1, 18, P["2"])

    # Retro Payphone unit mounted inside warm glass (x=8..13, y=17..25)
    draw_box(c, 8, 17, 6, 7, P["N"], P["O"], hi=P["M"], sh=P["n"])
    c.fill_rect(9, 18, 4, 2, P["X"])           # backlit cyan payphone LCD
    c.set_pixel(10, 18, P["W"])
    c.fill_rect(10, 21, 3, 2, P["W"])          # metallic keypad
    c.set_pixel(11, 21, P["d"])
    # Handset & coiled cord on left
    c.fill_rect(7, 18, 2, 4, P["O"])
    c.set_pixel(7, 19, P["K"])
    c.set_pixel(8, 24, P["O"])
    c.set_pixel(9, 25, P["O"])
    # Shelf beneath phone
    c.fill_rect(7, 25, 9, 1, P["t"])

    # Thin vertical glass door mullion & lower cross-bar
    c.fill_rect(14, 13, 1, 18, P["2"])
    c.fill_rect(6, 27, 12, 1, P["2"])
    # Chrome door handle on right pane
    c.fill_rect(15, 21, 1, 4, P["O"])
    c.fill_rect(15, 22, 1, 2, P["W"])

    # Crisp diagonal glass reflection streaks
    for dx, dy in ((7, 15), (8, 14), (9, 13), (8, 16), (9, 15), (10, 14), (15, 16), (16, 15)):
        c.set_pixel(dx, dy, P["W"])
    for dx, dy in ((7, 29), (8, 28), (15, 29), (16, 28)):
        c.set_pixel(dx, dy, P["3"])

    # Lower magenta kick panel & slate plinth base (x=3..20, y=32..38)
    draw_box(c, 4, 32, 16, 4, P["K"], P["O"], hi=P["Q"], sh=P["k"])
    draw_box(c, 3, 35, 18, 3, P["N"], P["O"], hi=P["M"], sh=P["n"])
    c.fill_rect(5, 36, 14, 1, P["X"])          # glowing neon base accent strip

    return c


# ==============================================================================
# 2. prop_food_cart.png (48x36)
# ==============================================================================

def build_prop_food_cart_48x36() -> Canvas:
    """
    prop_food_cart.png (48x36) — Steaming Tokyo/Roma street-food & takoyaki cart
    with striped coral/cream canopy, two glowing red/amber paper lanterns, copper grill,
    wheels, and warm wooden counter.
    """
    c = Canvas(48, 36)

    # Ground drop shadow
    fill_ellipse(c, 24.0, 33.5, 19.5, 2.2, P["s"])

    # Push handle extending on far left (x=2..8, y=21..24)
    c.fill_rect(2, 21, 6, 2, P["O"])
    c.fill_rect(3, 21, 5, 1, P["M"])
    c.fill_rect(2, 20, 2, 3, P["O"])
    c.set_pixel(2, 21, P["R"])

    # Subtle shaded interior valance behind counter so rising white steam pops cleanly
    c.fill_rect(10, 11, 28, 8, hex_to_rgba("#1E182B", 130))

    # Canopy vertical corner wood/iron posts (x=8..9 and x=38..39, y=10..21)
    for px in (8, 38):
        c.fill_rect(px, 10, 2, 11, P["O"])
        c.fill_rect(px, 10, 1, 10, P["1"])
        c.fill_rect(px + 1, 10, 1, 10, P["t"])

    # Small hanging menu chalkboard inside left of cart (x=11..19, y=11..16)
    draw_box(c, 11, 11, 8, 5, P["int_dark"], P["O"], hi=P["T"], sh=P["t"])
    c.fill_rect(13, 13, 4, 1, P["G"])
    c.fill_rect(13, 14, 3, 1, P["C"])

    # Warm Wooden Counter Ledge (x=6..41, y=20..23)
    draw_box(c, 6, 20, 36, 3, P["1"], P["O"], hi=P["C"], sh=P["T"])

    # Hammered Copper Takoyaki Grill sitting on center counter (x=16..32, y=16..21)
    draw_box(c, 16, 17, 16, 4, P["copper_base"], P["O"], hi=P["copper_hi"], sh=P["copper_sh"])
    # Polished copper front rim glint
    c.fill_rect(17, 18, 14, 1, P["copper_hi"])
    c.fill_rect(17, 19, 14, 1, P["copper_base"])
    c.set_pixel(18, 18, P["W"])
    c.set_pixel(29, 19, P["copper_sh"])
    # Sizzling cast-iron griddle top & golden takoyaki balls (y=15..17)
    c.fill_rect(17, 16, 14, 2, P["P"])
    for tx in (18, 21, 24, 27, 29):
        c.fill_rect(tx, 15, 2, 2, P["mustard_base"])
        c.set_pixel(tx, 15, P["G"])           # crispy golden top
        c.set_pixel(tx + 1, 16, P["t"])       # savory glaze dot

    # Condiment bottles & food boat tray on counter sides (y=16..20)
    draw_box(c, 10, 16, 3, 4, P["R"], P["O"], hi=P["F"])
    c.set_pixel(11, 15, P["W"])
    draw_box(c, 13, 17, 3, 3, P["C"], P["O"], hi=P["W"])
    draw_box(c, 33, 17, 5, 3, P["C"], P["O"], hi=P["W"], sh=P["a"])
    c.fill_rect(34, 16, 3, 1, P["Y"])         # chopped green scallions

    # Rising translucent white pixel steam wisps curling above copper grill (y=11..15)
    for sx, sy, col in (
        (19, 14, P["steam_hi"]), (20, 13, P["steam_hi"]), (19, 12, P["steam_soft"]),
        (23, 14, P["steam_hi"]), (24, 13, P["steam_hi"]), (25, 12, P["steam_hi"]), (24, 11, P["steam_soft"]),
        (28, 14, P["steam_hi"]), (29, 13, P["steam_hi"]), (28, 12, P["steam_soft"]),
    ):
        c.set_pixel(sx, sy, col)

    # Lower Wooden Cart Body (x=7..40, y=22..32)
    draw_box(c, 7, 22, 34, 10, P["T"], P["O"], hi=P["1"], sh=P["t"])
    for vx in range(11, 39, 4):
        c.fill_rect(vx, 23, 1, 8, P["t"])
        c.fill_rect(vx + 1, 23, 1, 7, P["1"])
    c.fill_rect(8, 30, 32, 1, P["wood_nail"])

    # Front coral/red street-food noren signboard (x=15..32, y=23..29)
    draw_box(c, 15, 23, 18, 7, P["C"], P["O"], hi=P["W"], sh=P["a"])
    c.fill_rect(16, 24, 16, 5, P["R"])
    # Cute round takoyaki/octopus mascot badge on left + warm cream calligraphy bars
    c.fill_rect(17, 25, 4, 3, P["F"])
    c.set_pixel(17, 25, P["W"])
    c.set_pixel(18, 26, P["O"])
    c.set_pixel(20, 26, P["O"])
    c.fill_rect(22, 25, 4, 1, P["C"])
    c.fill_rect(23, 25, 1, 3, P["C"])
    c.fill_rect(27, 25, 4, 1, P["G"])
    c.fill_rect(27, 27, 4, 1, P["C"])
    c.fill_rect(29, 25, 1, 3, P["G"])

    # Two Classic Spoked Cart Wheels (left x=9..15, right x=32..38, y=28..35)
    for wx in (9, 32):
        wheel = Canvas(48, 36)
        fill_ellipse(wheel, wx + 3.5, 31.5, 3.2, 3.2, P["P"])
        fill_ellipse(wheel, wx + 3.5, 31.5, 2.0, 2.0, P["T"])
        wheel.set_pixel(wx + 3, 31, P["G"])
        wheel.set_pixel(wx + 4, 31, P["1"])
        apply_1px_outline(wheel, P["O"])
        c.blit(wheel, 0, 0)
    c.fill_rect(23, 32, 2, 2, P["O"])

    # Striped Coral-Pink & Warm Cream Canopy (x=3..44, y=2..11)
    draw_box(c, 6, 2, 36, 2, P["T"], P["O"], hi=P["1"])
    draw_striped_awning(c, 3, 3, 42, 8, P["F"], P["f"], P["C"], P["a"], stripe_w=4)

    # Two Glowing Red/Amber Paper Lanterns (Chochin) hanging at corners
    for lx in (3, 39):
        fill_ellipse(c, lx + 3.0, 14.5, 5.5, 4.8, P["glow_amber"])
        c.fill_rect(lx + 2, 10, 2, 1, P["O"])
        draw_rounded_box(
            c, lx, 11, 6, 7, P["R"], P["O"], hi=P["led_amber"], sh=P["r"], round_top=True, round_bottom=True
        )
        c.fill_rect(lx + 1, 13, 4, 3, P["G"])
        c.fill_rect(lx + 2, 13, 2, 2, P["W"])
        c.fill_rect(lx + 1, 15, 4, 1, P["F"])
        c.fill_rect(lx + 2, 18, 2, 1, P["O"])

    return c


# ==============================================================================
# 3. prop_kiosk_umbrella.png (40x36)
# ==============================================================================

def build_prop_kiosk_umbrella_40x36() -> Canvas:
    """
    prop_kiosk_umbrella.png (40x36) — Cozy sidewalk cafe patio table with colorful
    coral & mint striped sun umbrella, two bistro chairs, coffee cups & tiny glowing table lamp.
    """
    c = Canvas(40, 36)

    # Soft wide ground shadow under patio set
    fill_ellipse(c, 20.0, 32.0, 16.0, 2.8, P["s"])

    # --- LEFT & RIGHT BISTRO CHAIRS (x=3..11 and x=29..37, y=20..33) ---
    draw_rounded_box(c, 3, 20, 4, 8, P["P"], P["O"], hi=P["M"], sh=P["p"], round_top=True)
    c.fill_rect(4, 22, 2, 4, P["teal_sh"])
    draw_box(c, 4, 26, 7, 3, P["T"], P["O"], hi=P["1"], sh=P["t"])
    c.fill_rect(4, 29, 2, 4, P["O"])
    c.fill_rect(9, 29, 2, 4, P["O"])
    c.set_pixel(4, 30, P["M"])
    c.set_pixel(9, 30, P["M"])

    draw_rounded_box(c, 33, 20, 4, 8, P["P"], P["O"], hi=P["M"], sh=P["p"], round_top=True)
    c.fill_rect(34, 22, 2, 4, P["teal_sh"])
    draw_box(c, 29, 26, 7, 3, P["T"], P["O"], hi=P["1"], sh=P["t"])
    c.fill_rect(29, 29, 2, 4, P["O"])
    c.fill_rect(34, 29, 2, 4, P["O"])
    c.set_pixel(29, 30, P["M"])
    c.set_pixel(34, 30, P["M"])

    # --- CAFE PATIO TABLE (x=11..28, y=21..33) ---
    c.fill_rect(19, 25, 2, 7, P["O"])
    c.fill_rect(19, 26, 1, 5, P["M"])
    draw_box(c, 15, 31, 10, 2, P["P"], P["O"], hi=P["M"])

    # Warm cream/stone oval bistro tabletop (contrasts cleanly with cups & glowing lamp)
    table = Canvas(40, 36)
    fill_ellipse(table, 20.0, 24.5, 8.8, 3.2, P["a"])
    fill_ellipse(table, 19.6, 23.8, 7.6, 2.3, P["C"])
    fill_ellipse(table, 18.5, 23.2, 4.5, 1.2, P["W"])
    apply_1px_outline(table, P["O"])
    c.blit(table, 0, 0)

    # Warm golden table lamp glow pool on right-center tabletop
    fill_ellipse(c, 23.0, 21.0, 4.5, 3.0, P["glow_warm"])

    # Central umbrella wooden/steel pole rising from table (x=19..20, y=13..24)
    c.fill_rect(19, 13, 2, 11, P["O"])
    c.fill_rect(19, 14, 1, 9, P["1"])
    c.fill_rect(20, 14, 1, 9, P["t"])

    # Two crisp outlined Coffee Cups on saucers (left x=13..16, right x=25..28)
    for cx, cup_col in ((13, P["F"]), (25, P["mint_base"])):
        c.fill_rect(cx, 23, 3, 1, P["O"])         # saucer shadow
        c.fill_rect(cx, 21, 3, 2, P["O"])
        c.set_pixel(cx + 1, 22, cup_col)          # colorful ceramic cup
        c.set_pixel(cx + 1, 21, P["W"])           # foam/rim highlight
        c.set_pixel(cx + 2, 22, P["W"])           # tiny handle
        c.set_pixel(cx + 1, 20, P["steam_soft"])  # tiny steam dot

    # Tiny Warm Glowing Bistro Table Lamp (x=21..25, y=18..23)
    c.fill_rect(22, 22, 3, 1, P["O"])             # brass foot
    c.fill_rect(23, 21, 1, 1, P["D"])             # brass stem
    draw_box(c, 21, 18, 5, 3, P["led_amber"], P["O"], hi=P["W"], sh=P["D"])
    c.fill_rect(22, 19, 3, 1, P["G"])
    c.set_pixel(23, 19, P["W"])                   # glowing lamp core

    # --- CORAL & MINT STRIPED SUN UMBRELLA CANOPY (x=4..35, y=2..15) ---
    umb = Canvas(40, 36)
    canopy_rows = [
        (3, 18, 22),
        (4, 15, 25),
        (5, 12, 28),
        (6, 9, 31),
        (7, 7, 33),
        (8, 6, 34),
        (9, 5, 35),
        (10, 4, 36),
        (11, 4, 36),
        (12, 4, 36),
        (13, 5, 35),
    ]
    for py, x0, x1 in canopy_rows:
        span = max(1, x1 - x0)
        for px in range(x0, x1):
            norm = (px - x0) / float(span)
            gore = int(norm * 6)
            is_coral = (gore % 2 == 0)
            if py <= 5 and px < 22:
                col = P["K"] if is_coral else P["mint_hi"]
            elif py >= 12:
                col = P["f"] if is_coral else P["mint_sh"]
            elif py == 11 and px > 24:
                col = P["f"] if is_coral else P["mint_sh"]
            else:
                col = P["F"] if is_coral else P["mint_base"]
            umb.set_pixel(px, py, col)
    for sx in range(5, 35):
        if (sx - 5) % 5 in (1, 2, 3):
            gore = int(((sx - 4) / 32.0) * 6)
            umb.set_pixel(sx, 14, P["f"] if gore % 2 == 0 else P["mint_sh"])
    umb.fill_rect(19, 2, 2, 1, P["G"])
    apply_1px_outline(umb, P["O"])
    c.blit(umb, 0, 0)

    c.set_pixel(16, 5, P["W"])
    c.set_pixel(14, 6, P["W"])
    c.set_pixel(21, 5, P["3"])
    c.set_pixel(22, 6, P["3"])

    return c


# ==============================================================================
# 4. prop_traffic_light.png (24x48)
# ==============================================================================

def build_prop_traffic_light_24x48() -> Canvas:
    """
    prop_traffic_light.png (24x48) — Sleek urban dark-iron traffic signal pole with
    hooded red/amber/green glowing LED lenses, pedestrian crossing box & zebra-striped base.
    """
    c = Canvas(24, 48)

    # Ground drop shadow
    fill_ellipse(c, 12.0, 45.2, 6.8, 1.8, P["s"])

    # Subtle LED ambient glow halos behind the signal lenses
    fill_ellipse(c, 12.0, 7.0, 6.2, 3.8, P["glow_red"])
    fill_ellipse(c, 12.0, 12.0, 5.5, 3.2, P["glow_amber"])
    fill_ellipse(c, 12.0, 17.0, 6.4, 3.8, P["glow_green"])

    # --- ZEBRA-STRIPED PROTECTIVE BASE (x=6..17, y=36..46) ---
    draw_box(c, 7, 37, 10, 8, P["G"], P["O"], hi=P["C"], sh=P["D"])
    # Bold diagonal/chevron zebra safety stripes
    for py in range(38, 44):
        for px in range(8, 16):
            if ((px + py) // 2) % 2 == 0:
                c.set_pixel(px, py, P["O"])
    draw_box(c, 8, 35, 8, 3, P["P"], P["O"], hi=P["M"], sh=P["p"])
    draw_box(c, 6, 44, 12, 2, P["P"], P["O"], hi=P["M"], sh=P["p"])

    # --- SLEEK DARK-IRON POLE (x=10..13, y=20..36) ---
    c.fill_rect(10, 20, 4, 16, P["O"])
    c.fill_rect(11, 20, 1, 15, P["M"])
    c.fill_rect(12, 20, 1, 15, P["p"])

    # --- PEDESTRIAN CROSSING BOX & PUSH BUTTON (y=24..33) ---
    # Right side illuminated pedestrian walk signal box (x=14..21, y=24..31)
    draw_box(c, 14, 24, 7, 7, P["P"], P["O"], hi=P["M"], sh=P["p"])
    c.fill_rect(15, 25, 5, 5, P["screen_bg"])
    c.set_pixel(17, 26, P["led_green_hi"])
    c.fill_rect(17, 27, 1, 2, P["led_green"])
    c.set_pixel(16, 28, P["led_green"])
    c.set_pixel(18, 29, P["led_green"])
    c.set_pixel(16, 29, P["led_green"])
    # Left side compact yellow crossing push-button unit (x=6..10, y=28..33)
    draw_box(c, 6, 28, 5, 5, P["G"], P["O"], hi=P["C"], sh=P["D"])
    c.set_pixel(8, 29, P["O"])
    c.set_pixel(8, 31, P["X"])

    # --- TRAFFIC SIGNAL HEAD HOUSING (x=7..16, y=1..21) ---
    c.fill_rect(10, 1, 4, 2, P["O"])
    c.fill_rect(11, 2, 2, 1, P["M"])
    draw_rounded_box(
        c, 7, 3, 10, 18, P["P"], P["O"], hi=P["M"], sh=P["p"], round_top=True, round_bottom=True
    )
    c.fill_rect(9, 4, 6, 16, P["int_dark"])

    # Three Hooded Glowing LED Lenses: Red, Amber, Green
    lenses = [
        (5, P["led_red"], P["led_red_hi"], P["r"]),
        (10, P["led_amber"], P["W"], P["D"]),
        (15, P["led_green"], P["led_green_hi"], P["led_green_sh"]),
    ]
    for ly, base_c, hi_c, sh_c in lenses:
        c.fill_rect(8, ly, 8, 1, P["O"])
        c.fill_rect(9, ly, 6, 1, P["M"])
        c.set_pixel(8, ly + 1, P["P"])
        c.set_pixel(15, ly + 1, P["p"])
        c.fill_rect(10, ly + 1, 4, 3, base_c)
        c.fill_rect(10, ly + 1, 2, 2, hi_c)
        c.set_pixel(10, ly + 1, P["W"])
        c.fill_rect(10, ly + 3, 4, 1, sh_c)
        c.set_pixel(13, ly + 2, sh_c)

    return c


# ==============================================================================
# 5. prop_bollard_glow.png (16x24)
# ==============================================================================

def build_prop_bollard_glow_16x24() -> Canvas:
    """
    prop_bollard_glow.png (16x24) — Modern illuminated architectural LED bollard post
    with warm-cyan glowing glass collar and clean ground shadow.
    """
    c = Canvas(16, 24)

    # Clean ground drop shadow + soft cyan floor light pool
    fill_ellipse(c, 8.0, 21.2, 5.2, 1.6, P["s"])
    fill_ellipse(c, 8.0, 19.6, 5.8, 2.2, P["glow_cyan_soft"])

    # Soft cyan-warm halo around glass collar
    fill_ellipse(c, 8.0, 8.5, 6.2, 4.4, P["glow_cyan"])

    # Base anchor flange (x=4..11, y=18..22)
    draw_rounded_box(
        c, 4, 18, 8, 4, P["P"], P["O"], hi=P["M"], sh=P["p"], round_top=False, round_bottom=True
    )
    c.fill_rect(5, 19, 6, 1, P["M"])

    # Sleek Brushed-Gunmetal Architectural Cylindrical Shaft (x=5..10, y=11..19)
    draw_box(c, 5, 11, 6, 8, P["N"], P["O"], hi=P["M"], sh=P["p"])
    c.fill_rect(6, 12, 1, 6, P["M"])          # brushed steel left cylinder highlight
    c.set_pixel(6, 12, P["W"])                # subtle top specular catch
    c.fill_rect(9, 12, 1, 6, P["p"])          # right cylinder shadow
    # Architectural horizontal groove with faint cyan reflection
    c.fill_rect(6, 15, 4, 1, P["O"])
    c.fill_rect(6, 16, 4, 1, P["2"])

    # Warm-Cyan Illuminated Architectural Glass Collar (x=5..10, y=5..12)
    draw_box(c, 5, 5, 6, 7, P["X"], P["O"], hi=P["3"], sh=P["2"])
    c.fill_rect(6, 6, 4, 5, P["3"])
    c.fill_rect(7, 6, 2, 4, P["W"])           # bright luminous vertical core
    c.fill_rect(7, 8, 2, 2, P["C"])           # subtle warm filament tint
    c.fill_rect(6, 6, 1, 5, P["X"])
    c.fill_rect(9, 6, 1, 5, P["2"])
    c.fill_rect(6, 8, 4, 1, P["2"])           # architectural diffuser ring

    # Sleek Bevelled Dark-Iron Top Cap (x=4..11, y=2..6)
    draw_rounded_box(c, 4, 2, 8, 4, P["P"], P["O"], hi=P["M"], sh=P["p"], round_top=True)
    c.fill_rect(6, 3, 3, 1, P["W"])

    return c


# ==============================================================================
# 6. prop_planter_neon.png (48x24)
# ==============================================================================

def build_prop_planter_neon_48x24() -> Canvas:
    """
    prop_planter_neon.png (48x24) — Wide modern concrete & cedar street planter box
    filled with lush green shrubs, pink/yellow blossoms, and a subtle warm LED accent strip
    along the lower recess.
    """
    c = Canvas(48, 24)

    # Ground drop shadow + subtle warm LED floor spill
    fill_ellipse(c, 24.0, 21.8, 21.5, 1.8, P["s"])
    c.fill_rect(6, 20, 36, 2, P["glow_warm_soft"])

    # Lower recessed dark concrete plinth + continuous warm LED strip (x=4..43, y=18..22)
    draw_box(c, 4, 18, 40, 4, P["n"], P["O"], sh=P["int_dark"])
    c.fill_rect(6, 19, 36, 1, P["led_amber"])
    c.fill_rect(8, 19, 32, 1, P["G"])
    for hx in (10, 18, 26, 34):
        c.fill_rect(hx, 19, 3, 1, P["C"])

    # Main Modern Concrete Outer Frame (x=2..45, y=10..19)
    draw_box(c, 2, 10, 44, 9, P["cobble_base"], P["O"], hi=P["cobble_hi"], sh=P["cobble_sh"])
    c.fill_rect(3, 11, 42, 1, P["cobble_hi"])
    c.fill_rect(7, 12, 1, 6, P["cobble_sh"])
    c.fill_rect(40, 12, 1, 6, P["cobble_sh"])

    # Warm Cedar Horizontal Slat Inlay Panel (x=8..39, y=12..18)
    draw_box(c, 8, 12, 32, 6, P["T"], P["O"], hi=P["1"], sh=P["t"])
    c.fill_rect(9, 14, 30, 1, P["t"])
    c.fill_rect(9, 15, 30, 1, P["1"])
    c.fill_rect(9, 16, 30, 1, P["t"])

    # Lush overflowing multi-cluster green shrubs (y=2..12)
    blobs = [
        (8.5, 8.5, 4.8, 3.6),
        (15.5, 6.8, 5.6, 4.0),
        (23.5, 6.2, 6.2, 4.2),
        (31.5, 7.0, 5.6, 4.0),
        (39.0, 8.5, 4.8, 3.6),
        (12.0, 9.5, 4.5, 2.8),
        (20.0, 9.5, 4.8, 2.8),
        (28.0, 9.5, 4.8, 2.8),
        (35.5, 9.5, 4.5, 2.8),
    ]
    draw_leaf_cluster(c, blobs, base_col=P["Y"], hi_col=P["l"], sh_col=P["y"], deep_col=P["leaf_deep"])

    # Crisp Pink & Yellow blossoms across shrubs
    blossoms = [
        (8, 7, P["K"], P["Q"]),
        (14, 5, P["G"], P["W"]),
        (19, 8, P["K"], P["Q"]),
        (24, 5, P["K"], P["W"]),
        (29, 7, P["G"], P["W"]),
        (34, 5, P["K"], P["Q"]),
        (39, 8, P["G"], P["W"]),
        (11, 9, P["G"], P["W"]),
        (32, 9, P["K"], P["Q"]),
    ]
    for bx, by, col_a, col_b in blossoms:
        c.set_pixel(bx, by, col_a)
        c.set_pixel(bx + 1, by, col_b)
        c.set_pixel(bx, by + 1, col_a)

    return c


# ==============================================================================
# 7. prop_vending_drink.png (28x36)
# ==============================================================================

def build_prop_vending_drink_28x36() -> Canvas:
    """
    prop_vending_drink.png (28x36) — Tall glowing retro sky-blue & cream Japanese
    drink vending machine (visually distinct from red shop vending) with illuminated
    colorful bottle rows and cool blue light spill.
    """
    c = Canvas(28, 36)

    # Ground drop shadow + cool sky-blue light spill on sidewalk
    fill_ellipse(c, 14.0, 33.6, 11.0, 2.0, P["s"])
    fill_ellipse(c, 14.0, 32.8, 12.2, 2.6, P["glow_sky"])

    # Cool blue ambient glow halo around lit drink showcase
    fill_ellipse(c, 14.0, 14.0, 12.5, 10.5, P["glow_sky"])

    # Rubber leveling feet (y=32..34)
    c.fill_rect(6, 32, 3, 2, P["O"])
    c.fill_rect(19, 32, 3, 2, P["O"])

    # Tall Retro Sky-Blue & Warm Cream Two-Tone Chassis (x=4..23, y=2..33)
    draw_rounded_box(c, 4, 2, 20, 31, P["sky_base"], P["O"], hi=P["sky_hi"], sh=P["sky_sh"], round_top=True)
    # Lower retro Warm Cream body half (x=5..22, y=20..31) — unmistakable sky-blue & cream styling!
    c.fill_rect(5, 20, 18, 12, P["C"])
    c.fill_rect(5, 20, 1, 12, P["W"])          # top-left cream edge highlight
    c.fill_rect(22, 20, 1, 12, P["a"])         # right cream shadow
    c.fill_rect(5, 31, 18, 1, P["a"])          # lower cream shadow
    # Retro sky-blue belt stripe separating upper & lower halves
    c.fill_rect(5, 20, 18, 1, P["sky_sh"])
    c.fill_rect(5, 21, 18, 1, P["sky_hi"])

    # Backlit Cream/Sky Header Sign Cap (x=6..21, y=4..6)
    c.fill_rect(6, 4, 16, 2, P["W"])
    c.fill_rect(7, 4, 5, 1, P["sky_sh"])
    c.fill_rect(13, 4, 3, 1, P["F"])
    c.fill_rect(17, 4, 4, 1, P["sky_base"])

    # --- ILLUMINATED DRINK SHOWCASE WINDOW (x=6..21, y=7..19) ---
    draw_box(c, 6, 7, 16, 13, P["ice_glow"], P["sky_sh"], hi=P["W"], sh=P["ice_mid"])
    # 3 shelves of crisp colorful miniature Japanese drink bottles (with 1px breathing spacing)
    shelves = [
        (8,  [P["Y"], P["sky_base"], P["F"], P["G"], P["K"], P["teal_base"], P["R"]]),
        (12, [P["R"], P["G"], P["X"], P["mint_base"], P["B"], P["F"], P["Y"]]),
        (16, [P["sky_sh"], P["Y"], P["F"], P["G"], P["U"], P["X"], P["K"]]),
    ]
    for sy, bottles in shelves:
        c.fill_rect(7, sy + 2, 14, 1, P["M"])  # shelf bar
        for i, col in enumerate(bottles):
            bx = 7 + i * 2
            c.set_pixel(bx, sy, P["W"])        # bottle neck cap
            c.set_pixel(bx + 1, sy, col)       # bottle shoulder
            c.fill_rect(bx, sy + 1, 2, 1, col) # colorful bottle label
            # Blue (cold) / red (hot) selection indicator on shelf rim
            c.set_pixel(bx, sy + 2, P["sky_base"] if sy < 15 else (P["R"] if i % 2 == 0 else P["sky_base"]))

    # Subtle diagonal glass sheen streak
    for rx, ry in ((7, 9), (8, 8), (19, 17), (20, 16)):
        c.set_pixel(rx, ry, P["W"])

    # --- MIDDLE CREAM CONTROL PANEL: DIGITAL PRICE + COIN SLOT (y=22..25) ---
    draw_box(c, 7, 22, 8, 4, P["screen_bg"], P["sky_sh"])
    c.fill_rect(8, 23, 5, 1, P["X"])           # glowing cyan price digits
    c.set_pixel(10, 23, P["W"])
    # Chrome coin slot & green indicator LED on right cream panel
    draw_box(c, 16, 22, 3, 4, P["M"], P["O"], hi=P["W"])
    c.set_pixel(17, 23, P["O"])
    c.fill_rect(20, 23, 2, 2, P["g"])
    c.set_pixel(20, 23, P["led_green"])

    # --- LOWER DISPENSER CHUTE & ECO STICKER ON CREAM BODY (y=27..31) ---
    draw_box(c, 7, 27, 11, 4, P["g"], P["O"], hi=P["M"], sh=P["int_dark"])
    c.fill_rect(8, 28, 9, 1, P["z"])           # translucent push-flap highlight
    # Cute green eco-bottle sticker on cream bottom right
    c.fill_rect(19, 27, 3, 4, P["W"])
    c.fill_rect(20, 28, 1, 2, P["Y"])

    return c


# ==============================================================================
# 8. prop_bus_shelter_deco.png (48x36)
# ==============================================================================

def build_prop_bus_shelter_deco_48x36() -> Canvas:
    """
    prop_bus_shelter_deco.png (48x36) — Illuminated City Metro/Bus Map Totem & Info Kiosk:
    sleek slate frame with glowing neon transit map display, subtle warm canopy light,
    tinted glass windbreak, arrival readout & modern cedar/steel bench rail.
    """
    c = Canvas(48, 36)

    # Ground drop shadow
    fill_ellipse(c, 24.0, 33.4, 21.0, 2.2, P["s"])

    # Sleek translucent tinted glass windbreak backing the right bay (x=27..41, y=9..29)
    c.fill_rect(27, 9, 14, 20, P["glass_tint"])
    # Subtle warm under-canopy light wash strictly inside canopy bay (y=8..13)
    c.fill_rect(7, 8, 34, 4, P["glow_warm_soft"])
    # Crisp diagonal glass reflection streaks on windbreak
    for i in range(6):
        c.set_pixel(29 + i, 24 - i, P["glass_sheen"])
        c.set_pixel(33 + i, 24 - i, P["glass_sheen"])

    # --- TOP ROOF BADGE & SLEEK CANTILEVERED CANOPY (x=3..44, y=1..9) ---
    draw_rounded_box(c, 6, 1, 13, 4, P["N"], P["O"], hi=P["X"], sh=P["n"], round_top=True)
    c.fill_rect(8, 2, 9, 2, P["screen_bg"])
    c.fill_rect(9, 2, 3, 1, P["X"])
    c.fill_rect(13, 2, 3, 1, P["K"])

    draw_rounded_box(c, 3, 4, 42, 5, P["N"], P["O"], hi=P["M"], sh=P["n"], round_top=True)
    c.fill_rect(5, 5, 38, 1, P["W"])
    # Subtle warm recessed under-canopy LED soffit light (y=7)
    c.fill_rect(6, 7, 36, 1, P["G"])
    c.fill_rect(8, 7, 32, 1, P["booth_warm_hi"])

    # --- STRUCTURAL SLATE PILLARS & BASE PLATES (y=8..33) ---
    for px in (5, 26, 40):
        draw_box(c, px, 8, 3, 25, P["P"], P["O"], hi=P["M"], sh=P["p"])
    draw_box(c, 4, 31, 5, 3, P["M"], P["O"], sh=P["d"])
    draw_box(c, 25, 31, 5, 3, P["M"], P["O"], sh=P["d"])
    draw_box(c, 39, 31, 5, 3, P["M"], P["O"], sh=P["d"])

    # --- LEFT/CENTER: ILLUMINATED NEON TRANSIT MAP TOTEM (x=7..27, y=9..31) ---
    draw_box(c, 7, 9, 21, 22, P["N"], P["O"], hi=P["M"], sh=P["n"])
    draw_box(c, 9, 11, 17, 16, P["screen_bg"], P["O"])
    for gx in (13, 17, 21):
        c.fill_rect(gx, 12, 1, 14, P["screen_grid"])
    for gy in (15, 19, 23):
        c.fill_rect(10, gy, 15, 1, P["screen_grid"])

    # Neon Cyan Metro Line
    for dx, dy in (
        (10, 14), (11, 14), (12, 15), (13, 16), (14, 16), (15, 16), (16, 16),
        (17, 17), (18, 18), (19, 18), (20, 18), (21, 18), (22, 17), (23, 16), (24, 16),
    ):
        c.set_pixel(dx, dy, P["X"])
    # Neon Magenta Bus Line
    for dx, dy in (
        (14, 12), (14, 13), (14, 14), (14, 15), (15, 17), (15, 18), (16, 19),
        (17, 20), (18, 21), (19, 22), (20, 23), (21, 24), (22, 24),
    ):
        c.set_pixel(dx, dy, P["K"])
    # Neon Gold Express Loop
    for dx, dy in (
        (10, 22), (11, 22), (12, 21), (13, 20), (14, 20), (15, 20), (16, 20),
        (19, 15), (20, 14), (21, 13), (22, 13), (23, 13),
    ):
        c.set_pixel(dx, dy, P["G"])
    # Emerald Park Line
    for dx, dy in ((11, 24), (12, 24), (13, 23), (14, 23), (15, 23), (16, 23), (17, 22)):
        c.set_pixel(dx, dy, P["led_green"])

    # Bright white interchange station nodes & 'YOU ARE HERE' red beacon
    for sx, sy in ((14, 16), (16, 20), (21, 18), (11, 22), (22, 13)):
        c.set_pixel(sx, sy, P["W"])
    c.set_pixel(17, 18, P["led_red"])
    c.set_pixel(18, 18, P["W"])

    # Diagonal glass reflection glint on map screen
    c.set_pixel(10, 12, P["z"])
    c.set_pixel(11, 12, P["3"])
    c.set_pixel(10, 13, P["z"])

    # Lower route color legend bar
    c.fill_rect(10, 28, 3, 1, P["X"])
    c.fill_rect(14, 28, 3, 1, P["K"])
    c.fill_rect(18, 28, 3, 1, P["G"])
    c.fill_rect(22, 28, 2, 1, P["led_green"])

    # --- RIGHT SIDE: DIGITAL BUS ARRIVAL BOARD + WAITING BENCH RAIL (x=28..44) ---
    draw_box(c, 28, 10, 13, 9, P["N"], P["O"], hi=P["M"], sh=P["n"])
    c.fill_rect(29, 11, 11, 7, P["screen_bg"])
    c.fill_rect(30, 12, 2, 1, P["X"])
    c.fill_rect(33, 12, 6, 1, P["W"])
    c.fill_rect(30, 14, 2, 1, P["K"])
    c.fill_rect(33, 14, 5, 1, P["G"])
    c.fill_rect(30, 16, 2, 1, P["led_green"])
    c.fill_rect(33, 16, 6, 1, P["C"])

    # Modern Waiting Bench Rail / Leaning Perch (x=27..44, y=21..33)
    draw_box(c, 28, 21, 15, 2, P["M"], P["O"], hi=P["W"], sh=P["d"])
    draw_box(c, 27, 25, 18, 4, P["T"], P["O"], hi=P["1"], sh=P["t"])
    c.fill_rect(28, 27, 16, 1, P["t"])
    draw_box(c, 29, 29, 2, 4, P["P"], P["O"], hi=P["M"])
    draw_box(c, 41, 29, 2, 4, P["P"], P["O"], hi=P["M"])
    c.fill_rect(31, 30, 10, 1, P["M"])

    return c


# ==============================================================================
# 9. car_van_cyber.png (64x36)
# ==============================================================================

def build_car_van_cyber_64x36() -> Canvas:
    """
    car_van_cyber.png (64x36) — Stylish retro-pop Japanese/Italian kei delivery van
    in cream & teal two-tone paint, chrome trim, roof cargo rack with crates,
    glowing warm headlights and red taillights, ground shadow. Facing RIGHT.
    """
    c = Canvas(64, 36)

    # Ground drop shadow
    fill_ellipse(c, 32.0, 32.2, 26.5, 2.6, P["s"])

    # Subtle glowing warm headlight spill (right) & red taillight glow (left)
    fill_ellipse(c, 59.0, 24.5, 4.5, 3.2, P["glow_warm"])
    fill_ellipse(c, 5.5, 24.0, 3.6, 2.6, P["glow_red"])

    # --- ROOF CARGO RACK WITH COLORFUL CRATES (x=10..44, y=2..11) ---
    c.fill_rect(11, 8, 32, 2, P["O"])
    for sx in (13, 26, 39):
        c.fill_rect(sx, 9, 2, 2, P["O"])
        c.set_pixel(sx, 9, P["W"])

    # Crate 1 (Left): Warm wooden delivery crate (x=13..22, y=3..9)
    draw_box(c, 13, 3, 10, 6, P["T"], P["O"], hi=P["1"], sh=P["t"])
    c.fill_rect(14, 5, 8, 1, P["t"])
    c.fill_rect(14, 7, 8, 1, P["t"])
    c.fill_rect(17, 4, 2, 4, P["1"])

    # Crate 2 (Center): Retro coral-red plastic soda crate (x=23..32, y=4..9)
    draw_box(c, 23, 4, 10, 5, P["F"], P["O"], hi=P["Q"], sh=P["r"])
    for gx in (25, 28, 30):
        c.fill_rect(gx, 6, 1, 2, P["int_dark"])
    c.fill_rect(24, 5, 8, 1, P["C"])

    # Crate 3 (Right): Tied cream parcel box with cyan tape (x=33..41, y=5..9)
    draw_box(c, 33, 5, 8, 4, P["C"], P["O"], hi=P["W"], sh=P["a"])
    c.fill_rect(36, 6, 2, 2, P["X"])

    # Front chrome cargo rack rail bar cleanly separating crates from roof
    c.fill_rect(11, 8, 32, 1, P["M"])
    c.fill_rect(12, 8, 6, 1, P["W"])
    c.fill_rect(25, 8, 6, 1, P["W"])
    c.fill_rect(11, 9, 32, 1, P["d"])
    c.set_pixel(10, 8, P["O"])
    c.set_pixel(43, 8, P["O"])

    # --- KEI VAN UPPER CABIN: WARM CREAM PAINT (x=7..56, y=10..20) ---
    draw_rounded_box(c, 8, 10, 44, 10, P["C"], P["O"], hi=P["W"], sh=P["a"], round_top=True)
    # Steep cab-over front nose slope (x=50..57, y=11..20)
    c.fill_rect(50, 11, 4, 1, P["O"])
    c.fill_rect(51, 12, 4, 2, P["C"])
    c.fill_rect(51, 12, 3, 1, P["W"])
    c.set_pixel(54, 12, P["O"])
    c.fill_rect(52, 13, 4, 7, P["C"])
    c.fill_rect(55, 13, 1, 7, P["O"])
    c.fill_rect(56, 15, 1, 6, P["O"])
    c.fill_rect(54, 14, 2, 6, P["a"])

    # Tinted Side Windows & Steep Front Windshield (y=12..19)
    draw_box(c, 10, 12, 12, 7, P["g"], P["O"], hi=P["a"])
    draw_box(c, 23, 12, 14, 7, P["g"], P["O"], hi=P["a"])
    draw_box(c, 38, 12, 10, 7, P["g"], P["O"], hi=P["a"])
    draw_box(c, 49, 13, 6, 6, P["g"], P["O"], hi=P["a"])
    c.fill_rect(53, 14, 1, 4, P["z"])

    # Crisp diagonal glass reflection streaks on side windows & windshield
    for wx in (12, 25, 40):
        c.set_pixel(wx + 2, 14, P["3"])
        c.set_pixel(wx + 1, 15, P["z"])
        c.set_pixel(wx, 16, P["z"])
        c.set_pixel(wx + 4, 14, P["z"])
        c.set_pixel(wx + 3, 15, P["z"])
    c.set_pixel(51, 14, P["3"])
    c.set_pixel(50, 15, P["z"])

    # Cute chrome side mirror neatly mounted at cab lower front corner (x=47..49, y=17..19)
    draw_box(c, 47, 17, 3, 3, P["W"], P["O"], sh=P["M"])

    # --- KEI VAN LOWER BODY: RICH POP-TEAL PAINT + CHROME TRIM (x=7..57, y=19..29) ---
    draw_box(c, 7, 19, 50, 10, P["teal_base"], P["O"], hi=P["teal_hi"], sh=P["teal_sh"])
    # Crisp double chrome beltline trim stripe separating cream & teal
    c.fill_rect(8, 19, 48, 1, P["W"])
    c.fill_rect(8, 20, 48, 1, P["M"])
    c.fill_rect(8, 21, 48, 1, P["teal_hi"])

    # Door seams & chrome door handles
    c.fill_rect(22, 20, 1, 8, P["teal_sh"])
    c.fill_rect(37, 20, 1, 8, P["teal_sh"])
    c.fill_rect(48, 20, 1, 8, P["teal_sh"])
    c.fill_rect(24, 22, 3, 1, P["W"])
    c.fill_rect(24, 23, 3, 1, P["d"])
    c.fill_rect(39, 22, 3, 1, P["W"])
    c.fill_rect(39, 23, 3, 1, P["d"])

    # Retro Pop Delivery Accent Stripes on side panel
    c.fill_rect(10, 23, 10, 2, P["C"])
    c.fill_rect(10, 25, 10, 1, P["F"])
    c.fill_rect(25, 24, 10, 1, P["G"])
    c.fill_rect(25, 25, 10, 1, P["C"])
    for ly in (22, 24, 26):
        c.fill_rect(9, ly, 2, 1, P["teal_sh"])

    # Front nose chrome grille trim
    c.fill_rect(54, 22, 2, 3, P["M"])
    c.set_pixel(55, 22, P["W"])

    # Glowing Warm Retro Round Headlight (front right x=55..58, y=22..26)
    draw_box(c, 55, 22, 4, 4, P["G"], P["O"], hi=P["W"], sh=P["D"])
    c.fill_rect(56, 23, 2, 2, P["booth_warm_hi"])
    c.set_pixel(56, 23, P["W"])
    c.fill_rect(55, 26, 2, 1, P["led_amber"])

    # Glowing Red & Amber Stacked Taillight (rear left x=6..9, y=21..26)
    draw_box(c, 6, 21, 3, 5, P["led_red"], P["O"], hi=P["led_red_hi"], sh=P["r"])
    c.set_pixel(7, 22, P["W"])
    c.set_pixel(7, 24, P["led_amber"])

    # Chrome Wrap-Around Bumpers (rear left & front right, y=26..29)
    draw_box(c, 5, 26, 4, 3, P["M"], P["O"], hi=P["W"], sh=P["d"])
    draw_box(c, 55, 26, 5, 3, P["M"], P["O"], hi=P["W"], sh=P["d"])

    # --- TWO CHUNKY KEI VAN WHEELS WITH CRISP CHROME HUBCAPS (matching parked cars) ---
    for wx in (13, 43):
        c.fill_rect(wx - 1, 26, 11, 2, P["O"])
        c.fill_rect(wx, 25, 9, 1, P["teal_sh"])
        tire = Canvas(64, 36)
        fill_ellipse(tire, wx + 4.5, 29.5, 4.0, 3.4, P["P"])
        fill_ellipse(tire, wx + 4.5, 30.2, 3.5, 2.5, P["p"])
        apply_1px_outline(tire, P["O"])
        c.blit(tire, 0, 0)
        # Crisp 4x4 chrome hubcap matching build_parked_car_48 style
        draw_box(c, wx + 3, 28, 4, 4, P["M"], P["W"], sh=P["d"])
        c.set_pixel(wx + 4, 29, P["W"])

    return c


# ==============================================================================
# MAIN ASSET GENERATION & VERIFICATION
# ==============================================================================

EXPECTED_SPECS: List[Tuple[str, int, int, int]] = [
    ("prop_phone_booth", 24, 40, 8),
    ("prop_food_cart", 48, 36, 6),
    ("prop_kiosk_umbrella", 40, 36, 6),
    ("prop_traffic_light", 24, 48, 6),
    ("prop_bollard_glow", 16, 24, 8),
    ("prop_planter_neon", 48, 24, 6),
    ("prop_vending_drink", 28, 36, 8),
    ("prop_bus_shelter_deco", 48, 36, 6),
    ("car_van_cyber", 64, 36, 6),
]


def main() -> None:
    os.makedirs(OUT_DIR, exist_ok=True)

    builders = {
        "prop_phone_booth": build_prop_phone_booth_24x40,
        "prop_food_cart": build_prop_food_cart_48x36,
        "prop_kiosk_umbrella": build_prop_kiosk_umbrella_40x36,
        "prop_traffic_light": build_prop_traffic_light_24x48,
        "prop_bollard_glow": build_prop_bollard_glow_16x24,
        "prop_planter_neon": build_prop_planter_neon_48x24,
        "prop_vending_drink": build_prop_vending_drink_28x36,
        "prop_bus_shelter_deco": build_prop_bus_shelter_deco_48x36,
        "car_van_cyber": build_car_van_cyber_64x36,
    }

    for name, exp_w, exp_h, zoom in EXPECTED_SPECS:
        canvas = builders[name]()
        assert (canvas.width, canvas.height) == (exp_w, exp_h), (
            f"{name}: expected {exp_w}x{exp_h}, got {canvas.width}x{canvas.height}"
        )
        save_asset_and_preview(canvas, OUT_DIR, name, zoom=zoom)
        print(f"Generated {name}.png ({canvas.width}x{canvas.height}) + preview_{name}_{zoom}x.png")

    print(f"Successfully generated all {len(EXPECTED_SPECS)} urban prop sprites (+ previews) in {OUT_DIR}")


if __name__ == "__main__":
    main()
