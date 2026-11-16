"""Finite fuzz/stress/benchmark workloads for the release-check runner."""
import hashlib
import json
import math
import os
from pathlib import Path
import platform
import re
import statistics
import shutil

ROOT = Path(__file__).resolve().parents[2]
FUZZ_TARGETS = [
    ("decode", ".", "./codec", "FuzzDecode"),
    ("answer", ".", "./codec", "FuzzDecodeAnswer"),
    ("canonical", ".", "./codec", "FuzzCanonical"),
    ("shdata", ".", "./app/sh/shdata", "FuzzParse"),
    ("sctp-mode", "adapters/sctp", ".", "FuzzSCTPModeSelection"),
]
BENCHMARKS = {
    "codec": (".", "./codec", "BenchmarkCodec", {
        "BenchmarkCodec/base-dwr/Decode", "BenchmarkCodec/base-dwr/AppendReuse",
        "BenchmarkCodec/v-srr-01/Decode", "BenchmarkCodec/v-srr-01/AppendReuse",
        "BenchmarkCodec/v-srr-01/Canonical"}),
    "ro": (".", "./examples/ro", "BenchmarkRoRoundTrip", {"BenchmarkRoRoundTrip"}),
    "sctp": ("adapters/sctp", ".", "BenchmarkSCTPFrame", {"BenchmarkSCTPFrame/256", "BenchmarkSCTPFrame/4096"}),
}
ENV = {"GOMAXPROCS": "2", "CGO_ENABLED": "0"}


def bench_samples(raw, expected):
    samples = {}
    for line in raw.splitlines():
        fields = line.split()
        if not fields or not fields[0].startswith("Benchmark") or len(fields) < 4:
            continue
        name = re.sub(r"-\d+$", "", fields[0])
        if not fields[1].isdigit():
            continue
        if int(fields[1]) != 200 or len(fields[2:]) % 2:
            raise RuntimeError("benchmark work count/format changed: " + line)
        values = {fields[i+1]: float(fields[i]) for i in range(2, len(fields), 2)}
        if not {"ns/op", "B/op", "allocs/op"} <= values.keys() or values["ns/op"] <= 0 or any(not math.isfinite(v) or v < 0 for v in values.values()):
            raise RuntimeError("missing benchmark measurements: " + name)
        samples.setdefault(name, []).append(values)
    if set(samples) != expected or any(len(v) != 5 for v in samples.values()):
        raise RuntimeError("incomplete benchmark samples")
    return {name: {"samples": rows, "statistics": {
        unit: {"min": min(r[unit] for r in rows), "median": statistics.median(r[unit] for r in rows), "max": max(r[unit] for r in rows)}
        for unit in rows[0]}} for name, rows in samples.items()}


def pressure_samples(raw):
    records = []
    for line in raw.splitlines():
        if line.startswith("{"):
            output = json.loads(line).get("Output", "")
            if "PRESSURE {" in output:
                records.append(json.loads(output.split("PRESSURE ", 1)[1]))
    expected = {1: (16, 8), 8: (128, 32), 16: (256, 64)}
    if len(records) != 60:
        raise RuntimeError("expected20 repetitions x3 pressure windows")
    for window, (requests, canceled) in expected.items():
        rows = [r for r in records if r["Window"] == window]
        if len(rows) != 20:
            raise RuntimeError("missing pressure window")
        for r in rows:
            if (r["Requests"], r["Canceled"], r["Waves"], r["Refusals"], r["ClientOutstandingPeak"], r["ServerInboundPeak"]) != (requests, canceled, 16, 32, window, window):
                raise RuntimeError("pressure workload changed")
            if not 0 < r["ProcessRSSHighWater"] <= 1 << 30:
                raise RuntimeError("missing/excessive native process RSS")
    return records


def run(checks, suite):
    if suite in ("all", "fuzz"):
        fuzz = []
        for label, module, package, target in FUZZ_TARGETS:
            try:
                raw = checks.run("fuzz-" + label,
                    ["go", "test", package, "-run=^$", "-fuzz=^" + target + "$", "-fuzztime=30s", "-fuzzminimizetime=5s", "-parallel=2", "-timeout=2m"],
                    ROOT / module, ENV)
            finally:
                # Preserve minimized failures for CI artifacts even when the
                # command raises. Never erase a checked-in regression corpus.
                corpus = ROOT / module / package / "testdata/fuzz" / target
                for entry in corpus.glob("*"):
                    if entry.is_file():
                        destination = checks.output / "fuzz-corpus" / label
                        destination.mkdir(parents=True, exist_ok=True)
                        shutil.copyfile(entry, destination / entry.name)
            executions = re.findall(r"execs:\s*(\d+)", raw)
            if not executions or int(executions[-1]) == 0 or "PASS" not in raw:
                raise RuntimeError("no actual fuzz executions: " + target)
            fuzz.append({"target": target, "seconds": 30, "workers": 2, "executions": int(executions[-1])})
        (checks.output / "fuzz.json").write_text(json.dumps(fuzz, indent=2) + "\n")
    if suite in ("all", "stress"):
        metrics = {"scope": "fixed loopback workload; logical reservations are not RSS; VmHWM includes the entire test process", "runs": {}}
        for mode in ("pure", "race"):
            flags = ["-race"] if mode == "race" else []
            env = dict(ENV, CGO_ENABLED="1" if mode == "race" else "0")
            raw = checks.run("pressure-ro-" + mode, ["go", "test", "./examples/ro", "-run=^TestBoundedRoPressure$", "-count=20", "-json", "-timeout=3m"] + flags, extra=env, tests=True)
            metrics["runs"][mode] = pressure_samples(raw)
            raw = checks.run("pressure-sctp-" + mode, ["go", "test", ".", "-run=^TestBoundedSCTPPressure$", "-count=20", "-json", "-timeout=3m"] + flags, ROOT / "adapters/sctp", env, tests=True)
            if raw.count("SCTP_PRESSURE connections=8 frames=512 fixed_buffer_bytes=") != 20:
                raise RuntimeError("incomplete native SCTP pressure workload")
        (checks.output / "pressure.json").write_text(json.dumps(metrics, indent=2) + "\n")
    if suite in ("all", "bench"):
        cpu = "unknown"
        for line in Path("/proc/cpuinfo").read_text().splitlines():
            if line.startswith("model name"):
                cpu = line.split(":", 1)[1].strip(); break
        report = {"format": 1, "scope": "initial measurement baseline, not a performance improvement or cross-host threshold",
                  "environment": {"platform": platform.platform(), "cpu": cpu, "available_cpus": len(os.sched_getaffinity(0)), "go": (ROOT / ".go-version").read_text().strip(), "benchmark_cpus": 1, "cgo": 0},
                  "iterations_per_sample": 200, "samples_per_benchmark": 5, "benchmarks": {}, "workload_sha256": {}}
        for path in ("codec/benchmark_test.go", "examples/ro/pressure_test.go", "adapters/sctp/pressure_linux_test.go"):
            report["workload_sha256"][path] = hashlib.sha256((ROOT / path).read_bytes()).hexdigest()
        for label, (module, package, target, names) in BENCHMARKS.items():
            raw = checks.run("bench-" + label, ["go", "test", package, "-run=^$", "-bench=^" + target + "$", "-benchtime=200x", "-count=5", "-benchmem", "-cpu=1", "-timeout=3m"], ROOT / module, ENV)
            report["benchmarks"].update(bench_samples(raw, names))
        (checks.output / "performance.json").write_text(json.dumps(report, indent=2) + "\n")
