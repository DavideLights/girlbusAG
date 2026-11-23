#!/usr/bin/env python3
"""
GIRLBUS: NEON STOP (girlbusAG) — Stratified District, Lighting & Intuitive UI Showcase Generator
Composites the actual generated sprites from assets/sprites/{map,world,characters} into a
320x180 native scene (upscaled 4x to 1280x720 nearest-neighbor) demonstrating:
- North Storefront Boulevard (Konbini Pop! [peak-envref7], Coffee Daily [envref5], Arcade & Books [peak-envref4])
- 45° Golden-Hour Architectural Drop Shadows & Warm Additive Window / Streetlamp Light Pools
- Central Cobblestone Cross-Street Plaza & Lush Pocket-Park Island (Lush Tree, Bench+Plants [peak-npcref6])
- Sidewalk Cats & Clutter, Parked Cars, Legal BUS Bay (+25%) vs ZTL Zone
- Intuitive 3D Keycap Boarding Combo & 4-Slot Bottom Archetype Hotbar Dock
"""

import math
from pixel_canvas import Canvas, PALETTE, hex_to_rgba
from gen_layered_map_assets import (
    build_konbini_96x64 as make_building_konbini,
    build_coffee_96x64 as make_building_coffee,
    build_arcade_96x64 as make_building_arcade,
    build_prop_tree_lush_48 as make_prop_tree_lush,
    build_prop_streetlamp_24x48 as make_prop_streetlamp,
    build_prop_bench_plants_48x32 as make_prop_bench_plants,
    build_prop_cats_clutter_32 as make_prop_cats_clutter,
    build_parked_car_48,
    build_tile_grass_park_32 as make_tile_grass_park,
    build_tile_cobble_plaza_32 as make_tile_cobble_plaza,
    build_tile_boardwalk_32 as make_tile_boardwalk,
    build_tile_canal_water_32 as make_tile_canal_water,
)

def make_car_yellow() -> Canvas:
    return build_parked_car_48("yellow")
from gen_world_assets import (
    STOP_KAWAII_48,
    build_tile_asphalt_32,
    build_tile_sidewalk_32,
    build_tile_crosswalk_32,
    build_tile_ztl_32,
)
from gen_characters import (
    ALFREDO_IDLE_0,
    GIRL_KAWAII_IDLE_0,
    GIRL_EMO_IDLE_0,
    GIRL_STUDENT_IDLE_0,
    GIRL_GYARU_IDLE_0,
)
from pixel_canvas import grid_to_canvas


def add_radial_glow(canvas: Canvas, cx: int, cy: int, rx: int, ry: int, rgb: tuple, strength: float) -> None:
    r_col, g_col, b_col = rgb
    for py in range(max(0, cy - ry), min(canvas.height, cy + ry + 1)):
        for px in range(max(0, cx - rx), min(canvas.width, cx + rx + 1)):
            dx = (px - cx) / float(max(1, rx))
            dy = (py - cy) / float(max(1, ry))
            d2 = dx * dx + dy * dy
            if d2 >= 1.0:
                continue
            falloff = (1.0 - d2) * (1.0 - d2) * strength
            dst = canvas.pixels[py][px]
            nr = min(255, int(dst[0] + r_col * falloff))
            ng = min(255, int(dst[1] + g_col * falloff))
            nb = min(255, int(dst[2] + b_col * falloff))
            canvas.pixels[py][px] = (nr, ng, nb, 255)


def main() -> None:
    w, h = 320, 180
    scene = Canvas(w, h, hex_to_rgba("#524E61"))

    t_asphalt = build_tile_asphalt_32()
    t_sidewalk = build_tile_sidewalk_32()
    t_cross = build_tile_crosswalk_32()
    t_ztl = build_tile_ztl_32()
    t_grass = make_tile_grass_park()
    t_cobble = make_tile_cobble_plaza()
    t_board = make_tile_boardwalk()
    t_water = make_tile_canal_water()

    # 1. Stratified Ground Layers
    for ty in range(0, h, 32):
        for tx in range(0, w, 32):
            if ty < 64:
                scene.blit(t_sidewalk, tx, ty)
            elif ty >= 160:
                scene.blit(t_water, tx, ty)
            elif ty >= 144:
                scene.blit(t_board, tx, ty)
            elif 128 <= tx < 192:
                scene.blit(t_cobble, tx, ty)
            elif tx == 96:
                scene.blit(t_cross, tx, ty)
            else:
                scene.blit(t_asphalt, tx, ty)

    # North Curb & Boardwalk Railing
    scene.fill_rect(0, 63, w, 2, hex_to_rgba("#9E9484"))
    scene.fill_rect(0, 65, w, 1, hex_to_rgba("#1C1124"))
    scene.fill_rect(0, 143, w, 2, hex_to_rgba("#704A2C"))
    scene.fill_rect(0, 158, w, 2, hex_to_rgba("#C4925E"))

    # Pocket-Park Island on Left-Center (X=16..92, Y=92..136)
    scene.fill_rect(14, 90, 78, 46, hex_to_rgba("#1C1124"))
    scene.fill_rect(15, 91, 76, 44, hex_to_rgba("#D8CEBE"))
    for py in range(96, 128, 32):
        for px in range(20, 84, 32):
            scene.blit(t_grass, px, py)
    scene.fill_rect(19, 95, 68, 34, hex_to_rgba("#448E42", 70))

    # Legal Green/Gold BUS BAY on Right Avenue (X=202..264, Y=68..92)
    scene.fill_rect(202, 68, 62, 22, hex_to_rgba("#2EEB82", 55))
    scene.fill_rect(202, 68, 62, 1, hex_to_rgba("#FFD166"))
    scene.fill_rect(202, 89, 62, 1, hex_to_rgba("#FFD166"))
    scene.fill_rect(202, 68, 1, 22, hex_to_rgba("#FFD166"))
    scene.fill_rect(263, 68, 1, 22, hex_to_rgba("#FFD166"))

    # ZTL Zone at Central Intersection (X=134..186, Y=106..138)
    scene.blit(t_ztl, 144, 106)

    # 2. 45-Degree Golden-Hour Architectural Drop Shadows (envref5 style!)
    shadow_col = hex_to_rgba("#181024", 105)
    for bx in (6, 112, 216):
        scene.fill_rect(bx + 6, 54, 96, 12, shadow_col)
        scene.fill_rect(bx + 94, 12, 8, 52, shadow_col)

    # 3. Storefront Buildings along North Boulevard
    b_konbini = make_building_konbini()
    b_coffee = make_building_coffee()
    b_arcade = make_building_arcade()
    scene.blit(b_konbini, 6, -4)
    scene.blit(b_coffee, 112, -4)
    scene.blit(b_arcade, 216, -4)

    # 4. Props, Trees, Benches, Cats, Streetlamps & Parked Cars
    tree = make_prop_tree_lush()
    bench = make_prop_bench_plants()
    cats = make_prop_cats_clutter()
    lamp = make_prop_streetlamp()
    car_y = make_car_yellow()
    stop_k = grid_to_canvas(STOP_KAWAII_48, 48, 48)

    scene.blit(cats, 102, 42)
    scene.blit(lamp, 96, 22)
    scene.blit(lamp, 200, 22)
    scene.blit(tree, 18, 72)
    scene.blit(bench, 44, 96)
    scene.blit(car_y, 266, 66)

    # Active Bus Stop inside the Legal BUS Bay + Holographic Capture Ring
    cx, cy = 228, 78
    for py in range(cy - 24, cy + 25):
        for px in range(cx - 38, cx + 39):
            dx = (px - cx) / 1.5
            dy = (py - cy)
            d = math.sqrt(dx * dx + dy * dy)
            if 22.5 <= d <= 24.0:
                scene.set_pixel(px, py, hex_to_rgba("#FF6B9D", 220))
            elif d < 22.5 and (px + py) % 3 == 0:
                scene.set_pixel(px, py, hex_to_rgba("#FF6B9D", 45))
    scene.blit(stop_k, 204, 42)

    # Characters (Alfredo + 4 Girl Archetypes)
    c_alf = grid_to_canvas(ALFREDO_IDLE_0, 32, 32)
    c_kaw = grid_to_canvas(GIRL_KAWAII_IDLE_0, 32, 32)
    c_emo = grid_to_canvas(GIRL_EMO_IDLE_0, 32, 32)
    c_stu = grid_to_canvas(GIRL_STUDENT_IDLE_0, 32, 32)
    c_gya = grid_to_canvas(GIRL_GYARU_IDLE_0, 32, 32)

    scene.blit(c_alf, 148, 68)
    scene.blit(c_kaw, 226, 68)
    scene.blit(c_stu, 112, 96)
    scene.blit(c_emo, 68, 110)
    scene.blit(c_gya, 254, 108)

    # Holographic Tether from Stop to Kawaii Girl
    for step in range(16):
        tx = 224 + int((242 - 224) * (step / 16.0))
        ty = 72 + int((84 - 72) * (step / 16.0))
        scene.set_pixel(tx, ty, hex_to_rgba("#00F0FF", 230))

    # 5. Dynamic Additive Lighting Pools (Streetlamps, Shop Windows, Neon Signs, Stop Glow)
    add_radial_glow(scene, 54, 42, 46, 26, (255, 210, 140), 0.30)   # Konbini window warm spill
    add_radial_glow(scene, 160, 42, 46, 26, (255, 215, 145), 0.32)  # Coffee Daily window warm spill
    add_radial_glow(scene, 264, 30, 48, 28, (255, 65, 155), 0.28)   # Arcade neon spill
    add_radial_glow(scene, 108, 34, 32, 32, (255, 225, 150), 0.42)  # Streetlamp 1 warm pool
    add_radial_glow(scene, 212, 34, 32, 32, (255, 225, 150), 0.42)  # Streetlamp 2 warm pool
    add_radial_glow(scene, 228, 78, 42, 26, (255, 90, 155), 0.30)   # Active Bus Stop neon halo

    # Floating Sakura Petals & Golden Dust Motes
    for px, py in [(42, 68), (88, 82), (135, 60), (178, 95), (245, 58), (290, 118), (120, 138)]:
        scene.fill_rect(px, py, 2, 1, hex_to_rgba("#FFA8C5", 220))
        scene.set_pixel(px + 5, py - 4, hex_to_rgba("#FFE596", 200))

    # 6. Intuitive Diegetic 3D Keycap Combo Panel next to Bus Stop
    px, py, pw, ph = 186, 16, 74, 24
    scene.fill_rect(px, py, pw, ph, hex_to_rgba("#120A1C", 215))
    scene.fill_rect(px, py, pw, 1, hex_to_rgba("#00F0FF"))
    scene.fill_rect(px, py + ph - 1, pw, 1, hex_to_rgba("#00F0FF"))
    scene.fill_rect(px, py, 1, ph, hex_to_rgba("#00F0FF"))
    scene.fill_rect(px + pw - 1, py, 1, ph, hex_to_rgba("#00F0FF"))
    # 4 3D Arrow Keycaps inside panel
    key_cols = [hex_to_rgba("#2EEB82"), hex_to_rgba("#2EEB82"), hex_to_rgba("#FFD166"), hex_to_rgba("#3A4761")]
    for k in range(4):
        kx = px + 5 + k * 17
        ky = py + 5 - (1 if k == 2 else 0)
        scene.fill_rect(kx, ky + 11, 14, 2, hex_to_rgba("#1C1124"))
        scene.fill_rect(kx, ky, 14, 11, key_cols[k])
        scene.fill_rect(kx + 1, ky + 1, 12, 1, hex_to_rgba("#FFFFFF", 130))
        # Arrow pixel inside keycap
        scene.fill_rect(kx + 5, ky + 4, 4, 3, hex_to_rgba("#1C1124"))

    # 7. Bottom-Center 4-Slot Intuitive Archetype Hotbar Dock
    dock_w, dock_h = 152, 22
    dock_x, dock_y = (w - dock_w) // 2, h - dock_h - 3
    scene.fill_rect(dock_x, dock_y, dock_w, dock_h, hex_to_rgba("#100A1C", 230))
    slot_cols = ["#9D65C2", "#4B9CE2", "#FF2A85", "#00F5D4"]
    for s in range(4):
        sx = dock_x + 3 + s * 37
        sy = dock_y + 2
        sc = hex_to_rgba(slot_cols[s])
        bg = hex_to_rgba("#2A1836", 240) if s == 2 else hex_to_rgba("#1A1228", 220)
        scene.fill_rect(sx, sy, 34, 18, bg)
        border = hex_to_rgba("#FFD166") if s == 2 else sc
        scene.fill_rect(sx, sy, 34, 1, border)
        scene.fill_rect(sx, sy + 17, 34, 1, border)
        scene.fill_rect(sx, sy, 1, 18, border)
        scene.fill_rect(sx + 33, sy, 1, 18, border)
        # Keycap number badge [1..4]
        scene.fill_rect(sx + 2, sy + 2, 6, 6, border)
        # Mini color indicator for Stop + Matching Girl affinity pair
        scene.fill_rect(sx + 11, sy + 4, 9, 10, sc)
        scene.fill_rect(sx + 22, sy + 5, 8, 9, hex_to_rgba("#FFE2CC"))
        scene.fill_rect(sx + 22, sy + 3, 8, 4, sc)

    out_dir = "/home/davidel/girlbusAG/assets/sprites/map"
    scene.save(f"{out_dir}/district_showcase_1x.png")
    scene.scale_nearest(4).save(f"{out_dir}/preview_district_showcase_4x.png")
    print("Generated stratified district showcase:", f"{out_dir}/preview_district_showcase_4x.png")


if __name__ == "__main__":
    main()
