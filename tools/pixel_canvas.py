#!/usr/bin/env python3
"""
GIRLBUS: NEON STOP (girlbusAG) — Shared Pixel-Art Canvas & Master Palette Engine
Pure Python 3 standard-library RGBA PNG & Sprite Sheet builder (zero external dependencies).
Used by all Art Subagents and the Master Atlas Builder.
"""

import os
import struct
import zlib
from typing import Dict, List, Tuple

TOOLS_DIR = os.path.dirname(os.path.abspath(__file__))
ROOT_DIR = os.path.dirname(TOOLS_DIR)

RGBA = Tuple[int, int, int, int]


def hex_to_rgba(hex_str: str, alpha: int = 255) -> RGBA:
    h = hex_str.lstrip("#")
    return (int(h[0:2], 16), int(h[2:4], 16), int(h[4:6], 16), alpha)


# ==============================================================================
# MASTER UNIFIED PALETTE (03_ART_STYLE_AND_SPRITE_BIBLE.md)
# Single-character keys usable across all 32x32, 36x36, 48x48, 96x48 grids.
# ==============================================================================
PALETTE: Dict[str, RGBA] = {
    ".": (0, 0, 0, 0),                        # Transparent
    "s": hex_to_rgba("#1A1025", 110),         # Ground Drop Shadow (43% alpha)
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

    # Alfredo & Campus Blue Outfit / Police Uniform
    "J": hex_to_rgba("#2E5C8A"),              # Slate-Blue Jacket / Uniform Base
    "j": hex_to_rgba("#1C3B5E"),              # Slate-Blue Jacket / Uniform Shadow
    "B": hex_to_rgba("#5488BD"),              # Slate-Blue Jacket / Uniform Highlight
    "P": hex_to_rgba("#303644"),              # Dark Jeans / Charcoal Slacks Base
    "p": hex_to_rgba("#1F232E"),              # Dark Jeans / Charcoal Slacks Shadow
    "R": hex_to_rgba("#E63946"),              # Red Sneaker / Siren / Vending Base
    "r": hex_to_rgba("#AD202B"),              # Red Shadow

    # Kawaii / Peak-NPCRef4-5 Girl Palette
    "K": hex_to_rgba("#FF6B9D"),              # Sakura Pink Base
    "k": hex_to_rgba("#D1326A"),              # Sakura Berry Shadow
    "x": hex_to_rgba("#8A1C46"),              # Sakura Deep Internal Strand Line
    "Q": hex_to_rgba("#FF9EBF"),              # Sakura Soft Mid-Highlight
    "q": hex_to_rgba("#FFDCE8"),              # Sakura Specular Bangs Shine
    "I": hex_to_rgba("#6E153A"),              # Ruby Eye Upper Iris
    "i": hex_to_rgba("#FF4D8B"),              # Vibrant Pink Eye Lower Glow
    "V": hex_to_rgba("#382240"),              # Dark Plum Cardigan / Shoes Base
    "v": hex_to_rgba("#231329"),              # Dark Plum Cardigan Shadow
    "G": hex_to_rgba("#FFD166"),              # Gold Brooch / VIP Crown / Blonde Highlight
    "D": hex_to_rgba("#E09F24"),              # Deep Gold / Blonde Base
    "U": hex_to_rgba("#7B3FA0"),              # Emo Grunge Purple Base
    "u": hex_to_rgba("#4E226B"),              # Emo Grunge Purple Shadow
    "Z": hex_to_rgba("#A86AD1"),              # Emo Grunge Purple Highlight

    # Konbini Bus Stop, Environment & Props (peak-envref7)
    "N": hex_to_rgba("#3E4F6B"),              # Signboard Slate-Blue Base
    "n": hex_to_rgba("#29354A"),              # Signboard Slate-Blue Shadow
    "M": hex_to_rgba("#7D8FA8"),              # Metal Pillar / Frame Base
    "d": hex_to_rgba("#526178"),              # Metal Pillar Shadow
    "C": hex_to_rgba("#FFE5B8"),              # Warm Cream Awning / Blonde Shine
    "a": hex_to_rgba("#DFBA82"),              # Warm Cream Shadow
    "F": hex_to_rgba("#FF5C77"),              # Coral-Pink Awning Stripe Base
    "f": hex_to_rgba("#C93450"),              # Coral-Pink Awning Stripe Shadow
    "g": hex_to_rgba("#222A3B"),              # Dark Glass Panel Base
    "z": hex_to_rgba("#3A4761"),              # Glass Diagonal Reflection Streak
    "T": hex_to_rgba("#D68049"),              # Warm Wood / Caramel Hair Base
    "t": hex_to_rgba("#9E5427"),              # Warm Wood / Caramel Hair Shadow
    "1": hex_to_rgba("#F2A672"),              # Warm Wood / Caramel Hair Highlight
    "Y": hex_to_rgba("#4BB553"),              # Lush Ivy / Bush Green Base
    "y": hex_to_rgba("#2A7834"),              # Lush Ivy / Bush Green Shadow
    "l": hex_to_rgba("#84E079"),              # Lush Ivy / Bush Green Highlight
    "X": hex_to_rgba("#00F0FF"),              # Neon Cyan / 5G Hologram / Chronos
    "2": hex_to_rgba("#0099B8"),              # Deep Cyan Shadow
    "3": hex_to_rgba("#B8FFFF"),              # Pale Cyan Highlight
}


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

    def draw_grid(
        self,
        grid: List[str],
        palette: Dict[str, RGBA] = PALETTE,
        offset_x: int = 0,
        offset_y: int = 0,
    ) -> None:
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


def grid_to_canvas(grid: List[str], width: int, height: int, palette: Dict[str, RGBA] = PALETTE) -> Canvas:
    if len(grid) != height:
        raise ValueError(f"Expected {height} rows, got {len(grid)}")
    for idx, r in enumerate(grid):
        if len(r) != width:
            raise ValueError(f"Row {idx} expected length {width}, got {len(r)}: '{r}'")
    c = Canvas(width, height)
    c.draw_grid(grid, palette)
    return c


def pack_horizontal_strip(frames: List[Canvas]) -> Canvas:
    if not frames:
        raise ValueError("No frames provided")
    fw, fh = frames[0].width, frames[0].height
    strip = Canvas(fw * len(frames), fh)
    for i, frame in enumerate(frames):
        strip.blit(frame, i * fw, 0)
    return strip


def save_asset_and_preview(canvas: Canvas, out_dir: str, name: str, zoom: int = 8) -> None:
    os.makedirs(out_dir, exist_ok=True)
    canvas.save(os.path.join(out_dir, f"{name}.png"))
    subdir = os.path.basename(os.path.normpath(out_dir))
    preview_dir = os.path.join(ROOT_DIR, "tools", "previews", subdir)
    os.makedirs(preview_dir, exist_ok=True)
    canvas.scale_nearest(zoom).save(os.path.join(preview_dir, f"preview_{name}_{zoom}x.png"))
