#!/usr/bin/env python3
"""Reproducible checkout checks for both modules; not a published-package gate.

Python is build tooling only. Go production modules remain CGO-independent.
Dependencies must be available via the caller's GOPROXY/cache; checks use a
temporary, explicit workspace and never edit module files or API baselines.
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import load_checks

ROOT = Path(__file__).resolve().parents[2]
CORE = "github.com/skill215/diameterkit"
ADAPTER = CORE + "/adapters/sctp"
UPSTREAM = "github.com/skill215/sctp"
GENERATED = {
    "base": "base", "sh": "app/sh", "s6c": "app/s6c", "sgd": "app/sgd",
    "ro": "app/ro", "roabort": "app/ro/abort", "dcca": "app/ro/dcca",
    "smsfwro": "compat/smsfwro", "legacy420": "compat/smsfwro/legacy420",
    "combined": "compat/smsfwro/combined", "scap1": "compat/scap1",
}
MODULES = {"core": ROOT, "adapter": ROOT / "adapters/sctp"}


def objects(raw):
    decoder = json.JSONDecoder()
    while raw.strip():
        item, end = decoder.raw_decode(raw.lstrip())
        raw = raw.lstrip()[end:]
        yield item


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def validate_modules(raw, root, upstream_version):
    modules = {m["Path"]: m for m in objects(raw)}
    for name, directory in ((CORE, root), (ADAPTER, root / "adapters/sctp")):
        m = modules.get(name, {})
        if not m.get("Main") or Path(m.get("Dir", "")).resolve() != directory.resolve():
            raise RuntimeError("module is not the current checkout: " + name)
    m = modules.get(UPSTREAM, {})
    if m.get("Version") != upstream_version or m.get("Replace"):
        raise RuntimeError("SCTP dependency differs from the pinned upstream version")
    if set(modules) != {CORE, ADAPTER, UPSTREAM}:
        raise RuntimeError("module graph changed; review dependencies before extending the gate")


def test_counts(raw):
    events = [json.loads(line) for line in raw.splitlines() if line.startswith("{")]
    passed = sum(e.get("Action") == "pass" and bool(e.get("Test")) for e in events)
    skipped = [e.get("Test") for e in events if e.get("Action") == "skip" and e.get("Test")]
    if not passed or skipped or any(e.get("Action") == "fail" for e in events):
        raise RuntimeError(f"test gate failed: passes={passed}, skipped={skipped}")
    return {"test_subtest_seed_passes": passed, "test_skips": 0}


class Checks:
    def __init__(self, output, env):
        self.output, self.env = output, env
        self.steps = []

    def run(self, name, command, cwd=ROOT, extra=None, tests=False, timeout=900):
        print("RUN " + name, flush=True)
        env = self.env.copy()
        env.update(extra or {})
        log = self.output / (name + ".log")
        step = {"name": name, "command": [str(c) for c in command],
                "cwd": str(cwd), "log": log.name, "status": "failed"}
        self.steps.append(step)
        try:
            with log.open("w", encoding="utf-8") as out:
                result = subprocess.run(command, cwd=cwd, env=env, stdout=out,
                                        stderr=subprocess.STDOUT, text=True, timeout=timeout)
        except subprocess.TimeoutExpired:
            step["failure_kind"] = "process_timeout"
            raise
        finally:
            # Timeout is a failure too: retain and hash whatever was written.
            if log.exists():
                step["sha256"] = digest(log)
        step["returncode"] = result.returncode
        raw = log.read_text(encoding="utf-8")
        if result.returncode:
            step["failure_kind"] = "command_exit"
            raise RuntimeError(f"{name} exited {result.returncode}; see {log}")
        if tests:
            step.update(test_counts(raw))
        step["status"] = "passed"
        return raw

    def reject(self, name, command, message, extra=None):
        print("RUN " + name, flush=True)
        log = self.output / (name + ".log")
        result = subprocess.run(command, cwd=ROOT, env={**self.env, **(extra or {})}, capture_output=True, text=True, timeout=900)
        log.write_text(result.stdout + result.stderr, encoding="utf-8")
        step = {"name": name, "command": [str(c) for c in command], "status": "failed",
                "expected_failure": True, "log": log.name, "sha256": digest(log)}
        self.steps.append(step)
        if result.returncode == 0 or message not in result.stdout + result.stderr:
            raise RuntimeError("negative control did not reject the expected condition: " + name)
        step["status"] = "passed"


def check_generated(checks, temp):
    expected = {ROOT / p / "zz_generated_messages.go" for p in GENERATED.values()}
    actual = {p for folder in ("base", "app", "compat") for p in (ROOT / folder).rglob("zz_generated_messages.go")}
    if actual != expected:
        raise RuntimeError("generated file inventory changed; update the explicit gate")
    generator = temp / "diameter-schema"
    checks.run("generator-build", ["go", "build", "-o", str(generator), "./cmd/diameter-schema"])
    for name, folder in GENERATED.items():
        checks.run("generated-" + name, [str(generator), "-in", "spec/" + name + ".json",
                   "-out", folder + "/zz_generated_messages.go", "-check"])
    # Exercise the actual CLI refusal in isolation. It must not repair the file.
    stale = temp / "stale.go"
    stale.write_bytes((ROOT / "base/zz_generated_messages.go").read_bytes() + b"\n// stale\n")
    before = digest(stale)
    checks.reject("generated-negative-control", [str(generator), "-in", str(ROOT / "spec/base.json"),
                  "-out", str(stale), "-check"], "generated output differs")
    if digest(stale) != before:
        raise RuntimeError("generated drift refusal did not fail without rewriting")


def execute(suite, output):
    env = os.environ.copy()
    env.update(GOENV="off", GOTOOLCHAIN="local", GOWORK="off", GOFLAGS="-mod=readonly",
               CGO_ENABLED="0")
    # Runtime gates are deliberately native Linux/amd64; cross builds are separate.
    for key in ("GOOS", "GOARCH"):
        env.pop(key, None)
    output.mkdir(parents=True, exist_ok=True)
    checks = Checks(output, env)
    summary = {"format": 1, "suite": suite, "status": "failed", "scope": "current checkout; not installed release", "steps": checks.steps}
    protected = [ROOT / "go.mod", ROOT / "adapters/sctp/go.mod", ROOT / "adapters/sctp/go.sum",
                 ROOT / "spec/release/public-api.json", ROOT / "spec/release/support-matrix.json"]
    before = {str(p): digest(p) for p in protected}
    try:
        version = (ROOT / ".go-version").read_text().strip()
        actual = checks.run("toolchain", ["go", "env", "GOVERSION", "GOHOSTOS", "GOHOSTARCH"]).splitlines()
        if actual != ["go" + version, "linux", "amd64"]:
            raise RuntimeError(f"requires Go {version} on Linux/amd64; got {actual}")
        summary["toolchain"] = actual
        module = json.loads(checks.run("adapter-manifest", ["go", "mod", "edit", "-json"], MODULES["adapter"]))
        requirements = {r["Path"]: r["Version"] for r in module["Require"]}
        if module.get("Replace") or set(requirements) != {CORE, UPSTREAM}:
            raise RuntimeError("adapter dependency manifest changed; review before running")
        with tempfile.TemporaryDirectory(prefix="diameterkit-ci-") as directory:
            temp = Path(directory)
            checks.run("workspace", ["go", "work", "init", str(ROOT), str(MODULES["adapter"])], temp)
            checks.env["GOWORK"] = str(temp / "go.work")
            # A versioned replacement also avoids looking up the offline-only
            # batch requirement's go.mod while constructing the workspace graph.
            checks.run("checkout-binding", ["go", "work", "edit", "-replace=" + CORE + "@" + requirements[CORE] + "=" + str(ROOT)], temp)
            raw = checks.run("module-graph", ["go", "list", "-m", "-json", "all"], MODULES["adapter"])
            validate_modules(raw, ROOT, requirements[UPSTREAM])
            summary["dependencies"] = {"core": "current checkout", UPSTREAM: requirements[UPSTREAM]}
            # Checksums are supplied by the checked-in adapter go.sum, even in
            # offline mode. No new core candidate or edited sum file is needed.
            checks.run("dependency-download", ["go", "mod", "download", UPSTREAM + "@" + requirements[UPSTREAM]], MODULES["adapter"])
            if suite in ("all", "static"):
                checks.run("ai-disclosure", [sys.executable, "spec/tools/check_ai_headers.py"])
                checks.run("tooling-tests", [sys.executable, "-m", "unittest", "discover", "-s", "spec/tools", "-p", "test_release*.py", "-v"])
                checks.run("support-matrix", [sys.executable, "spec/tools/check_support_matrix.py"])
                stale_matrix = temp / "stale-matrix.json"
                baseline_matrix = json.loads((ROOT / "spec/release/support-matrix.json").read_text())
                baseline_matrix["profiles"][0]["definition_sha256"] = "0" * 64
                stale_matrix.write_text(json.dumps(baseline_matrix))
                matrix_digest = digest(stale_matrix)
                checks.reject("support-matrix-negative-control", [sys.executable, "spec/tools/check_support_matrix.py",
                              "--matrix", str(stale_matrix)], "definition_sha256 differs")
                if digest(stale_matrix) != matrix_digest:
                    raise RuntimeError("support matrix check rewrote its baseline")
                check_generated(checks, temp)
                checks.run("generator-determinism", ["go", "test", "-count=1", "-json", "./internal/codegen", "-run", "^TestGeneratedFilesCurrentAndDeterministic$"], tests=True)
                consumer = temp / "consumer"
                consumer.mkdir()
                isolated = {"GOWORK": "off"}
                checks.run("context-consumer-init", ["go", "mod", "init", "example.org/contextconsumer"], consumer, isolated)
                checks.run("context-consumer-binding", ["go", "mod", "edit", "-require=" + CORE + "@v0.0.0",
                           "-replace=" + CORE + "=" + str(ROOT)], consumer, isolated)
                shutil.copyfile(ROOT / "examples/contracts/context_test.go", consumer / "context_test.go")
                checks.run("context-consumer", ["go", "test", "-mod=readonly", "-count=1", "-json", "."],
                           consumer, isolated, tests=True)
                checks.run("api-snapshot", [sys.executable, "spec/tools/public_api_snapshot.py", "--check", "--workspace", str(temp / "go.work")])
                stale = temp / "stale-api.json"
                baseline = json.loads((ROOT / "spec/release/public-api.json").read_text())
                baseline["targets"][0]["packages"][0]["documentation"] += "\nintentional API drift\n"
                stale.write_text(json.dumps(baseline))
                original = digest(stale)
                checks.reject("api-negative-control", [sys.executable, "spec/tools/public_api_snapshot.py", "--check", "--output", str(stale), "--workspace", str(temp / "go.work")], "public API snapshot differs")
                if digest(stale) != original:
                    raise RuntimeError("API check rewrote its baseline")
                # Exercise the compiler view: change only the hidden alias target.
                source = ROOT / "internal/contextvalue/value.go"
                edited = temp / "value.go"
                edited.write_text(source.read_text() + "\nfunc (*Value) SnapshotProbe() {}\n")
                overlay = temp / "overlay.json"
                overlay.write_text(json.dumps({"Replace": {str(source): str(edited)}}))
                checks.reject("context-method-negative-control", [sys.executable, "spec/tools/public_api_snapshot.py",
                              "--check", "--workspace", str(temp / "go.work")], "public API snapshot differs",
                              {"GOFLAGS": "-mod=readonly -overlay=" + str(overlay)})
                for name, cwd in MODULES.items():
                    checks.run(name + "-vet", ["go", "vet", "./..."], cwd)
            for mode in ("pure", "race"):
                if suite not in ("all", mode):
                    continue
                for name, cwd in MODULES.items():
                    command = ["go", "test", "-count=1", "-timeout=10m", "-json"]
                    if mode == "race":
                        command.append("-race")
                    checks.run(name + "-" + mode, command + ["./..."], cwd,
                               {"CGO_ENABLED": "1" if mode == "race" else "0"}, tests=True)
            if suite in ("all", "cross"):
                # Compilation only; Windows/386 SCTP uses the unsupported stub.
                for target in ("linux/arm64", "linux/386", "windows/amd64"):
                    goos, goarch = target.split("/")
                    for name, cwd in MODULES.items():
                        checks.run(name + "-build-" + target.replace("/", "-"), ["go", "build", "./..."], cwd,
                                   {"GOOS": goos, "GOARCH": goarch, "CGO_ENABLED": "0"})
            load_checks.run(checks, suite)
        summary["status"] = "passed"
    except (RuntimeError, OSError, subprocess.SubprocessError) as error:
        summary["error"] = str(error)
    finally:
        changed = [str(p) for p in protected if digest(p) != before[str(p)]]
        if changed:
            summary.update(status="failed", error="check modified protected inputs: " + ", ".join(changed))
        (output / "summary.json").write_text(json.dumps(summary, indent=2) + "\n")
    print(json.dumps({"status": summary["status"], "summary": str(output / "summary.json"), "error": summary.get("error")}))
    return 0 if summary["status"] == "passed" else 1


def main():
    parser = argparse.ArgumentParser(__doc__)
    parser.add_argument("--suite", choices=("all", "static", "pure", "race", "cross", "fuzz", "stress", "bench"), default="all")
    parser.add_argument("--output", type=Path, help="logs/summary directory (default .cache/release-check/SUITE)")
    args = parser.parse_args()
    return execute(args.suite, (args.output or ROOT / ".cache/release-check" / args.suite).resolve())


if __name__ == "__main__":
    sys.exit(main())
