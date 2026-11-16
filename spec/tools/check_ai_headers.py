#!/usr/bin/env python3
"""Check project-owned Go provenance declarations without changing sources."""
from pathlib import Path
import subprocess

ROOT = Path(__file__).resolve().parents[2]


def main():
    raw = subprocess.check_output(["git", "ls-files", "--cached", "--others", "--exclude-standard", "-z"], cwd=ROOT)
    paths = sorted({p for p in raw.decode().split("\0") if p.endswith(".go")})
    required = ("Disclosure", "Model", "Provider", "Scope", "Date")
    failures = []
    for name in paths:
        text = (ROOT / name).read_text()
        header = text.split("\npackage ", 1)[0]
        if any(header.count("// SPDX-AI-" + key + ": ") != 1 for key in required):
            failures.append(name)
    if not paths or failures:
        raise SystemExit("missing/duplicate AI file headers: " + ", ".join(failures))
    print(f"AI declarations checked: {len(paths)} project Go files; model gaps remain explicit")


if __name__ == "__main__":
    main()
