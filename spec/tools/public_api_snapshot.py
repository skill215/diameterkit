#!/usr/bin/env python3
"""Snapshot public Go documentation from build-selected packages in both modules.

This is a reviewable development API inventory, not a compatibility checker or
stable-API promise. No network, source editing, candidate packaging or publishing.
Use the adapter's already installed pinned dependencies (GOMODCACHE may be set).
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess

ROOT = Path(__file__).resolve().parents[2]
TARGETS = ("linux/amd64", "windows/amd64")


def objects(raw):
    decoder = json.JSONDecoder()
    while raw.strip():
        item, end = decoder.raw_decode(raw.lstrip())
        raw = raw.lstrip()[end:]
        yield item


CONTEXT_PACKAGE = "github.com/skill215/diameterkit/validation"


def context_methods(directory, env):
    """Inspect target archives with a host executable, including aliased methods."""
    archives = subprocess.check_output(
        ["go", "list", "-mod=readonly", "-export", "-deps", "-json=ImportPath,Export", CONTEXT_PACKAGE],
        cwd=directory, env=env)
    host = {k: v for k, v in env.items() if k not in ("GOOS", "GOARCH")}
    return json.loads(subprocess.check_output(
        ["go", "run", str(ROOT / "cmd/internal/methodset/main.go"), CONTEXT_PACKAGE, "Context"],
        input=archives, cwd=directory, env=host))


def snapshot(workspace=None):
    env = os.environ.copy()
    env.update(CGO_ENABLED="0", GOWORK=str(workspace) if workspace else "off", GOTOOLCHAIN="local",
               GOPROXY="off", GOSUMDB="off")
    def run(args, cwd, e):
        return subprocess.check_output(["go"] + args, cwd=cwd, env=e, text=True)
    result = {"format": 2, "scope": "development documentation snapshot;not API stability or semantic compatibility proof",
              "toolchain": run(["env", "GOVERSION"], ROOT, env).strip(),
              "selection": "build-selected GoFiles;exclude internal,main,test-only packages;CGO_ENABLED=0;no network",
              "targets": []}
    for target in TARGETS:
        selected = []
        e = env.copy()
        e["GOOS"], e["GOARCH"] = target.split("/")
        for module_dir in (ROOT, ROOT / "adapters/sctp"):
            packages = list(objects(run(["list", "-mod=readonly", "-json", "./..."], module_dir, e)))
            for p in sorted(packages, key=lambda p: p["ImportPath"]):
                if p["Name"] == "main" or "internal" in p["ImportPath"].split("/") or not p.get("GoFiles"):
                    continue
                if p.get("CgoFiles"):
                    raise RuntimeError("Cgo in public package: " + p["ImportPath"])
                docs = run(["doc", "-all", p["ImportPath"]], module_dir, e)
                files = []
                for name in sorted(p["GoFiles"]):
                    path = Path(p["Dir"]) / name
                    files.append({"path": path.relative_to(ROOT).as_posix(),
                                  "sha256": hashlib.sha256(path.read_bytes()).hexdigest()})
                selected.append({"import_path": p["ImportPath"], "package": p["Name"],
                                 "module": p["Module"]["Path"], "files": files,
                                 "documentation_sha256": hashlib.sha256(docs.encode()).hexdigest(),
                                 "documentation": docs})
                if p["ImportPath"] == CONTEXT_PACKAGE:
                    selected[-1]["method_sets"] = {"Context": context_methods(module_dir, e)}
        result["targets"].append({"target": target, "packages": sorted(selected, key=lambda p:p["import_path"])})
    return result


def main():
    parser = argparse.ArgumentParser(__doc__)
    parser.add_argument("--output", type=Path, default=ROOT / "spec/release/public-api.json")
    parser.add_argument("--check", action="store_true", help="fail if the existing snapshot differs")
    parser.add_argument("--workspace", type=Path, help="explicit temporary checkout workspace; default checks independent pinned modules")
    args = parser.parse_args()
    workspace = args.workspace.resolve(strict=True) if args.workspace else None
    current = snapshot(workspace)
    if args.check:
        if json.loads(args.output.read_text()) != current:
            raise SystemExit("public API snapshot differs; review before updating")
    else:
        args.output.parent.mkdir(parents=True, exist_ok=True)
        args.output.write_text(json.dumps(current, ensure_ascii=False, indent=2) + "\n")
    print(json.dumps({"status": "checked" if args.check else "written",
                      "packages_per_target": {t["target"]:len(t["packages"]) for t in current["targets"]}}))


if __name__ == "__main__":
    main()
