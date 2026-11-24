#!/usr/bin/env python3
"""
GIRLBUS: NEON STOP (girlbusAG) — Upgrade Icons v2 (6 New Level-Up Cards, 20x20).

Generates 1x native PNG sprites (+ 10x zoomed previews in tools/previews/ui/):
  - assets/sprites/ui/icon_upg_dash.png
  - assets/sprites/ui/icon_upg_shield.png
  - assets/sprites/ui/icon_upg_rhythm.png
  - assets/sprites/ui/icon_upg_bass.png
  - assets/sprites/ui/icon_upg_perfect.png
  - assets/sprites/ui/icon_upg_magnet.png
"""

import os
import sys
from typing import Dict, List

TOOLS_DIR = os.path.dirname(os.path.abspath(__file__))
ROOT_DIR = os.path.dirname(TOOLS_DIR)
if TOOLS_DIR not in sys.path:
    sys.path.insert(0, TOOLS_DIR)

from pixel_canvas import PALETTE, RGBA, grid_to_canvas, hex_to_rgba, save_asset_and_preview

UI_OUT_DIR = os.path.join(ROOT_DIR, "assets", "sprites", "ui")

ICON_PALETTE: Dict[str, RGBA] = {
    **PALETTE,
    "O": hex_to_rgba("#1C1124"),
    "o": hex_to_rgba("#382342"),
    "W": hex_to_rgba("#FFFFFF"),
    "M": hex_to_rgba("#8C9EB8"),
    "d": hex_to_rgba("#526178"),
    "P": hex_to_rgba("#303644"),
    "p": hex_to_rgba("#1F232E"),
    "3": hex_to_rgba("#C8FFFF"),
    "X": hex_to_rgba("#00F0FF"),
    "2": hex_to_rgba("#0099B8"),
    "c": hex_to_rgba("#005C78"),
    "q": hex_to_rgba("#FFDCE8"),
    "Q": hex_to_rgba("#FF85B8"),
    "K": hex_to_rgba("#FF2A85"),
    "k": hex_to_rgba("#B5145B"),
    "L": hex_to_rgba("#FFF5B8"),
    "G": hex_to_rgba("#FFD93D"),
    "D": hex_to_rgba("#E0941A"),
    "b": hex_to_rgba("#99580A"),
    "m": hex_to_rgba("#7CF5D6"),
    "Y": hex_to_rgba("#32D468"),
    "y": hex_to_rgba("#188C43"),
    "Z": hex_to_rgba("#C98BFF"),
    "U": hex_to_rgba("#8A3FFC"),
    "u": hex_to_rgba("#521D9E"),
}


def build_dash_icon() -> List[str]:
    return [
        "....................",
        "..............OOOO..",
        ".............OW3X2O.",
        "............OW3X2O..",
        "...........OW3X2O...",
        "..OOOO....OW3X2O....",
        "..O3XO...OW3X2O.....",
        "...OO...OW3X2O......",
        ".......OW3X2O..OOO..",
        "......OW3X2O...OXO..",
        "....OOOWX2O.....O...",
        "...OGGDW2O..........",
        "...OOGDGO...........",
        "..OKkOGOO...........",
        ".OKkOO..............",
        ".OO.................",
        "..OOOOO.............",
        "..O3XXO.............",
        "...OOO..............",
        "....................",
    ]


def build_shield_icon() -> List[str]:
    return [
        "....................",
        "......OOOOOOOO......",
        "....OOLLLLGGGGDO....",
        "...OLGG3XXXX2GGDO...",
        "..OLG33XXXXXX2GGDO..",
        "..OG3XXXWWWWXX2GDO..",
        "..OG3XXW3XX3WX2GDO..",
        "..OG3XXWXXXXWX2GDO..",
        "..OG3XXXWWWWXX2GDO..",
        "..OG2XXXX22XXX2GDO..",
        "..ODG2XX2cc2X2GDDO..",
        "...ODG2XXXXXXGDDO...",
        "...OODG2XXXXGDDO....",
        "....OODG2XXGDDO.....",
        ".....OODGGGGDO......",
        "......OODGGDO.......",
        ".......OODDO........",
        "........OOO.........",
        "....................",
        "....................",
    ]


def build_rhythm_icon() -> List[str]:
    return [
        "....................",
        "........OOOOOOOOOO..",
        ".......OqQQQKKKKKko.",
        ".......OqQQQKKKKKko.",
        "..OOO..OKKOOOOOKkO..",
        "..OLGO.OKKO...OKkO..",
        "...OO..OKKO...OKkO..",
        ".......OKKO...OKkO..",
        ".......OKKO...OKkO..",
        "....OOOOKKO.OOOKkO..",
        "..OqQQQKKKOoqQQKkO..",
        "..OqQKKKKKOoqQKKkO..",
        "..OKKKKKKkoOKKKKkO..",
        "...OkkkkkO..OkkkOO..",
        "....OOOOO....OOO....",
        "..............OOO...",
        "..............OGLO..",
        "...............OO...",
        "....................",
        "....................",
    ]


def build_bass_icon() -> List[str]:
    return [
        "....................",
        "....OOOOOOOOOOOO....",
        "...OZZUUUUUUUUuuO...",
        "..OZUPPPPPPPPPPpO...",
        "..OUPP3mmmmmm3PpO...",
        "..OUPmOOOOOOOOmpO...",
        "..OUPmOuuuuuuOmpO...",
        "..OUPmOuOOOOuOmpO...",
        "..OUPmOuO3XOuoMpO...",
        "..OUPmOuOX2OuoMpO...",
        "..OUPmOuOOOOuoMpO...",
        "..OUPmOuuuuuuoMpO...",
        "..OUPPooooooooPpO...",
        "..OUPPmmmmmmmmPpO...",
        "..OUPPPPPPPPPPPpO...",
        "...OuuuuuuuuuuuO....",
        "....OOOOOOOOOOOO....",
        "..OO............OO..",
        ".O3O............O3O.",
        "....................",
    ]


def build_perfect_icon() -> List[str]:
    return [
        "....................",
        ".........OO.........",
        "........OLGO........",
        "........OLGO........",
        "....OO..OGDO..OO....",
        "...OGLO.OGDO.OLGO...",
        "....OOGGGGGGGGOO....",
        ".....OGLLLLLLGO.....",
        "..OOOOGL3XX3LGOOOO..",
        ".OLLLLGLXXXXLGDDDDO.",
        "..OOOOGLXWWXLGOOOO..",
        ".....OGLX33XLGO.....",
        "....OOGGGGGGGGOO....",
        "...OGDO.OGDO.OGDO...",
        "....OO..OGDO..OO....",
        "........OGDO........",
        "........ODbO........",
        ".........OO.........",
        "....................",
        "....................",
    ]


def build_magnet_icon() -> List[str]:
    return [
        "....................",
        "....................",
        "...OOOOOOOOOOOOOO...",
        "..OLLLLGGGGGGGGDDO..",
        "..OLGG33XXXX22GDDO..",
        ".OOLLG3XWWWWX2GDDDO.",
        "..OOGGXWW33WWXGDDO..",
        ".OOLLG3XWWWWX2GDDDO.",
        "..OLLG33XXXX22GDDO..",
        "..OLLGGGGGGGGGGDDO..",
        "..ODGDDDDDDDDDDDDO..",
        "..ODbPPPPPPPPPPbDO..",
        "..ODGGGGGGGGGGGGDO..",
        "...OODDDDDDDDDDDO...",
        "....OOOOOOOOOOOO....",
        "........OOOO........",
        ".......O3XX2O.......",
        "........OOOO........",
        "....................",
        "....................",
    ]


def main() -> None:
    specs = [
        ("icon_upg_dash", build_dash_icon()),
        ("icon_upg_shield", build_shield_icon()),
        ("icon_upg_rhythm", build_rhythm_icon()),
        ("icon_upg_bass", build_bass_icon()),
        ("icon_upg_perfect", build_perfect_icon()),
        ("icon_upg_magnet", build_magnet_icon()),
    ]
    for name, grid in specs:
        c = grid_to_canvas(grid, 20, 20, ICON_PALETTE)
        save_asset_and_preview(c, UI_OUT_DIR, name, zoom=10)
    print("Generated 6 new 20x20 upgrade icons in assets/sprites/ui/")


if __name__ == "__main__":
    main()
