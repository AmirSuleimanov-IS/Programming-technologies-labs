import sys, glob, re
from PIL import Image
prefix, out = sys.argv[1], sys.argv[2]
files = sorted(glob.glob(prefix + '_*.png'), key=lambda f: int(re.search(r'_(\d+)\.png$', f).group(1)))
ims = [Image.open(f) for f in files]
cols = int(sys.argv[3]) if len(sys.argv) > 3 else 6
w, h = ims[0].size
rows = (len(ims) + cols - 1) // cols
s = Image.new('RGB', (w * cols, h * rows), 'gray')
for i, im in enumerate(ims):
    s.paste(im, ((i % cols) * w, (i // cols) * h))
s.save(out)
print(len(ims), s.size)
