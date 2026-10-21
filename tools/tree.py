"""Дерево файлов репозитория по git ls-files: python tree.py <корень> [подпапка] [глубина]"""
import io
import subprocess
import sys

sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding="utf-8")

root = sys.argv[1]
sub = sys.argv[2] if len(sys.argv) > 2 and sys.argv[2] != "." else ""
depth = int(sys.argv[3]) if len(sys.argv) > 3 else 99
cmd = ["git", "-C", root, "ls-files"] + ([sub] if sub else [])
files = subprocess.run(cmd, capture_output=True, text=True, encoding="utf-8").stdout.split("\n")
tree = {}
dirs = set()
for f in filter(None, files):
    parts = f[len(sub):].strip("/").split("/") if sub else f.split("/")
    node, path = tree, ""
    for i, p in enumerate(parts[:depth]):
        path += "/" + p
        node = node.setdefault(p, {})
        if i < len(parts) - 1:
            dirs.add(path)


def show(node, prefix="", path=""):
    items = sorted(node.items(), key=lambda kv: (path + "/" + kv[0] not in dirs, kv[0].lower()))
    for i, (name, child) in enumerate(items):
        last = i == len(items) - 1
        is_dir = path + "/" + name in dirs
        print(prefix + ("└── " if last else "├── ") + name + ("/" if is_dir else ""))
        show(child, prefix + ("    " if last else "│   "), path + "/" + name)


print((sub.rstrip("/") or ".") + "/")
show(tree)
