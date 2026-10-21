"""Сжимает PNG в палитру из 256 цветов, если файл становится меньше: python optimize_png.py <папка|файлы>"""
import os
import sys

from PIL import Image

paths = []
for arg in sys.argv[1:]:
    if os.path.isdir(arg):
        paths += [os.path.join(arg, f) for f in os.listdir(arg) if f.lower().endswith(".png")]
    else:
        paths.append(arg)

before = after = 0
for p in paths:
    size = os.path.getsize(p)
    with Image.open(p) as im:
        if im.mode == "P":
            before += size
            after += size
            continue
        q = im.convert("RGB").quantize(colors=256, method=Image.Quantize.MEDIANCUT, dither=Image.Dither.NONE)
    tmp = p + ".tmp.png"
    q.save(tmp, optimize=True)
    if os.path.getsize(tmp) < size:
        os.replace(tmp, p)
    else:
        os.remove(tmp)
    before += size
    after += os.path.getsize(p)
print(f"{len(paths)} файлов: {before / 1e6:.1f} МБ -> {after / 1e6:.1f} МБ")
