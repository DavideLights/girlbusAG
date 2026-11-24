#!/usr/bin/env python3
"""
GIRLBUS: NEON STOP (girlbusAG) — Boss Key & Armored Neon Chest Pixel-Art Generator.

Generates 1x native PNG sprites (+ zoomed previews in tools/previews/):
  - assets/sprites/ui/icon_boss_key.png   (16x16)
  - assets/sprites/world/chest_closed.png (32x28)
  - assets/sprites/world/chest_open.png   (32x28)

Follows 03_ART_STYLE_AND_SPRITE_BIBLE.md:
  - Universal 1px Deep Aubergine outline (#1C1124)
  - Top-Left cel-pixel 3-tone shading
  - Zero dithering, crisp readable silhouettes
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
    RGBA,
    grid_to_canvas,
    hex_to_rgba,
    save_asset_and_preview,
)

UI_OUT_DIR = os.path.join(ROOT_DIR, "assets", "sprites", "ui")
WORLD_OUT_DIR = os.path.join(ROOT_DIR, "assets", "sprites", "world")

CHEST_PALETTE: Dict[str, RGBA] = {
    **PALETTE,
    "Y": hex_to_rgba("#FFE899"),   # Gold Specular
    "G": hex_to_rgba("#FFD166"),   # Gold Base
    "D": hex_to_rgba("#C48622"),   # Gold Deep Shadow
    "U": hex_to_rgba("#382240"),   # Dark Plum Interior
    "u": hex_to_rgba("#231329"),   # Deep Plum Recess
    "K": hex_to_rgba("#FF9EBF"),   # Soft Pink Rim Accent
}


def build_boss_key_grid() -> List[str]:
    return [
        "....OOOOOOOO....",
        "...OYYYYGGGGO...",
        "..OYGGOOOOOGDO..",
        "..OYGO3XX2OGDO..",
        "..OYGOXXXXOGDO..",
        "..OGDO2XX2OGDO..",
        "...OGDOOOOGDO...",
        "....OOGGGGDO....",
        ".....OOYGDO.....",
        "......OYGDO.....",
        "......OYGDOOOO..",
        "......OYGGGGYGO.",
        "......OYGDOOOOO.",
        "......OYGDOOO...",
        "......OYGGGO....",
        "......OOOOO.....",
    ]


def build_chest_closed_grid() -> List[str]:
    return [
        "................................",
        "................................",
        "................................",
        "................................",
        "................................",
        "................................",
        "....OOOOOOOOOOOOOOOOOOOOOOOO....",
        "...OYYGGGGGGGGGGGGGGGGGGGGDDO...",
        "..OYGMddddddddddddddddddddMGDDO.",
        "..OGMdPPPPPPPPPPPPPPPPPPPPdMGDO.",
        "..OGMdP3XXXXXXXXXXXXXXXX2PdMGDO.",
        "..OGMPdPPPPPPPOOOOPPPPPPPdPMGDO.",
        "..OGMGGGGGGGGGOYYOGGGGGGGGGMGDO.",
        "..ODddddddddddOGGOdddddddddddDO.",
        "..OPPPPPPPPPPPoDDoPPPPPPPPPPPpO.",
        "..OGMP3XXXX2PPoOOoPP3XXXX2PMGpO.",
        "..OGMPPPPPPPPPPPPPPPPPPPPPPMGpO.",
        "..OGMPPPPPPPPPPPPPPPPPPPPPPMGpO.",
        "..OGMPppppppppppppppppppppPMGpO.",
        "..OGMPppppppppppppppppppppPMGpO.",
        "..OGMPPPPPPPPPPPPPPPPPPPPPPMGpO.",
        "..ODGMPPPPPPPPPPPPPPPPPPPPMGDpO.",
        "..OOGGDDDDDDDDDDDDDDDDDDDDGDDpO.",
        "...OODDDDDDDDDDDDDDDDDDDDDDDDO..",
        "....OOOOOOOOOOOOOOOOOOOOOOOOO...",
        "...ssssssssssssssssssssssssss...",
        "..ssssssssssssssssssssssssssss..",
        "................................",
    ]


def build_chest_open_grid() -> List[str]:
    return [
        "....OOOOOOOOOOOOOOOOOOOOOOOO....",
        "...OYYGGGGGGGGGGGGGGGGGGGGDDO...",
        "..OYGMddddddddddddddddddddMGDDO.",
        "..OGMdPPPPPPPPPPPPPPPPPPPPdMGDO.",
        "..OGMdP3XXXXXXXXXXXXXXXX2PdMGDO.",
        "..OGMPddddddddddddddddddddPMGDO.",
        "..OOGGGGGGGGGGGGGGGGGGGGGGGGGDO.",
        "...OuuuuuuuuuuuuuuuuuuuuuuuuuO..",
        "...OuUUUUUUUUUUUUUUUUUUUUUUUuO..",
        "...OuUUUU3XX3YYYYYGG3XX3UUUUuO..",
        "...OuUUU3XXXXYGYGYGXXXX3UUUUuO..",
        "..OOGGGGGGGGGGGGGGGGGGGGGGGGGDO.",
        "..ODddddddddddOYYOdddddddddddDO.",
        "..OPPPPPPPPPPPoGGoPPPPPPPPPPPpO.",
        "..OGMP3XXXX2PPoOOoPP3XXXX2PMGpO.",
        "..OGMPPPPPPPPPPPPPPPPPPPPPPMGpO.",
        "..OGMPPPPPPPPPPPPPPPPPPPPPPMGpO.",
        "..OGMPppppppppppppppppppppPMGpO.",
        "..OGMPppppppppppppppppppppPMGpO.",
        "..OGMPPPPPPPPPPPPPPPPPPPPPPMGpO.",
        "..OGMPPPPPPPPPPPPPPPPPPPPPPMGpO.",
        "..ODGMPPPPPPPPPPPPPPPPPPPPMGDpO.",
        "..OOGGDDDDDDDDDDDDDDDDDDDDGDDpO.",
        "...OODDDDDDDDDDDDDDDDDDDDDDDDO..",
        "....OOOOOOOOOOOOOOOOOOOOOOOOO...",
        "...ssssssssssssssssssssssssss...",
        "..ssssssssssssssssssssssssssss..",
        "................................",
    ]


def main() -> None:
    key_canvas = grid_to_canvas(build_boss_key_grid(), 16, 16, CHEST_PALETTE)
    save_asset_and_preview(key_canvas, UI_OUT_DIR, "icon_boss_key", zoom=12)

    closed_canvas = grid_to_canvas(build_chest_closed_grid(), 32, 28, CHEST_PALETTE)
    save_asset_and_preview(closed_canvas, WORLD_OUT_DIR, "chest_closed", zoom=8)

    open_canvas = grid_to_canvas(build_chest_open_grid(), 32, 28, CHEST_PALETTE)
    save_asset_and_preview(open_canvas, WORLD_OUT_DIR, "chest_open", zoom=8)

    print("Generated icon_boss_key.png (16x16), chest_closed.png (32x28), chest_open.png (32x28)")


if __name__ == "__main__":
    main()
