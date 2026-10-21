import sys
import pymupdf as fitz
pdf, outprefix = sys.argv[1], sys.argv[2]
d = fitz.open(pdf)
pages = [int(x) for x in sys.argv[3].split(',')] if len(sys.argv) > 3 else range(len(d))
for i in pages:
    d[i].get_pixmap(dpi=60).save(f'{outprefix}_{i+1}.png')
print('pages', len(d))
