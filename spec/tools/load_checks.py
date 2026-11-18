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
import subprocess

ROOT = Path(__file__).resolve().parents[2]
FUZZ_TARGETS = [
    ("decode", ".", "./codec", "FuzzDecode"),
    ("answer", ".", "./codec", "FuzzDecodeAnswer"),
    ("canonical", ".", "./codec", "FuzzCanonical"),
    ("missing-diagnostic", ".", "./internal/missingavp", "FuzzMissingQuote"),
    ("excess-diagnostic", ".", "./internal/excessavp", "FuzzExcessQuote"),
    ("invalid-diagnostic", ".", "./internal/excessavp", "FuzzInvalidQuote"),
    ("length-diagnostic", ".", "./internal/excessavp", "FuzzLengthQuote"),
    ("diameter-uri", ".", "./avp", "FuzzDiameterURI"),
    ("manual-diagnostic", ".", "./base", "FuzzManualDiagnostics"),
    ("group-constraint", ".", "./internal/groupavp", "FuzzGroupQuote"),
    ("variable-minimum", ".", "./internal/missingavp", "FuzzVariableMinimum"),
    ("conditional-minimum", ".", "./internal/missingavp", "FuzzConditionalMinimum"),
    ("cardinality-minimum", ".", "./internal/cardinalityavp", "FuzzCardinalityQuote"),
    ("variable-group-minimum", ".", "./internal/missingavp", "FuzzMinimumVariableTree"),
    ("minimum-group", ".", "./internal/missingavp", "FuzzMinimumTree"),
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


def mixed_pressure_samples(raw):
    records = []
    for line in raw.splitlines():
        if line.startswith("{"):
            output = json.loads(line).get("Output", "")
            if "MIXED_PRESSURE {" in output:
                records.append(json.loads(output.split("MIXED_PRESSURE ", 1)[1]))
    if len(records) != 160:
        raise RuntimeError("expected 20 repetitions x 2 windows x 4 stalled applications")
    for window in (1, 4):
        for stalled in ("ro", "s6c", "sh", "sgd"):
            rows = [r for r in records if r.get("ClassWindow") == window and r.get("Stalled") == stalled]
            if len(rows) != 20:
                raise RuntimeError("missing mixed pressure scenario")
            expected = dict(Window=4*window, Waves=8, Requests=80*window,
                            HealthyCompleted=72*window, Canceled=4*((window+1)//2),
                            Refusals=32, ClassRefusals=16, PortRefusals=16,
                            ClientOutstandingPeak=4*window, ServerInboundPeak=4*window)
            for row in rows:
                if any(row.get(key) != value for key, value in expected.items()):
                    raise RuntimeError("mixed pressure workload changed")
                for field in ("ClientLogicalPeak", "ServerLogicalPeak", "SampledHeapAlloc", "ProcessRSSHighWater"):
                    value = row.get(field)
                    if type(value) is not int or not 0 < value <= 1 << 30:
                        raise RuntimeError("missing/excessive mixed resource measurement: " + field)
    return records


FUZZ_EXECUTIONS = 1_000_000


def fuzz_measurements(raw):
    """Require the count budget and actual fuzzing; never reinterpret failures."""
    counts = re.findall(r"^fuzz: elapsed: .*?, execs:\s*(\d+)\b", raw, re.M)
    if (not counts or int(counts[-1]) != FUZZ_EXECUTIONS or
            not re.search(r"^PASS\s*$", raw, re.M) or
            not re.search(r"^ok\s+", raw, re.M) or
            "now fuzzing with 2 workers" not in raw or
            re.search(r"^(?:--- FAIL:|FAIL(?:\s|$)|panic:|fatal error:)", raw, re.M) or
            "context deadline exceeded" in raw):
        raise RuntimeError("incomplete or failed count-bounded fuzz output")
    return {"executions": int(counts[-1])}


def preserve_corpus(checks, label, module, package, target):
    corpus = ROOT / module / package / "testdata/fuzz" / target
    artifacts = []
    for entry in sorted(corpus.glob("*")):
        if entry.is_file():
            destination = checks.output / "fuzz-corpus" / label
            destination.mkdir(parents=True, exist_ok=True)
            copied = destination / entry.name
            shutil.copyfile(entry, copied)
            artifacts.append({"path": copied.relative_to(checks.output).as_posix(),
                              "sha256": hashlib.sha256(copied.read_bytes()).hexdigest()})
    return artifacts


def run_fuzz(checks):
    reports = []
    receipt = checks.output / "fuzz.json"
    receipt.write_text("[]\n", encoding="utf-8")
    for label, module, package, target in FUZZ_TARGETS:
        record = {"target": target, "status": "failed", "budget_kind": "executions",
                  "requested_executions": FUZZ_EXECUTIONS, "workers": 2,
                  "minimization_seconds": 5, "timeout_seconds": 120}
        phase = "process"
        try:
            try:
                raw = checks.run("fuzz-" + label,
                    ["go", "test", package, "-run=^$", "-fuzz=^" + target + "$",
                     f"-fuzztime={FUZZ_EXECUTIONS}x", "-fuzzminimizetime=5s",
                     "-parallel=2", "-timeout=2m"], ROOT / module, ENV)
            finally:
                # Capture newly minimized failures even when the Go command
                # exits nonzero or the outer process timeout fires.
                phase = "corpus"
                record["corpus"] = preserve_corpus(checks, label, module, package, target)
                phase = "process"
            phase = "output"
            record.update(fuzz_measurements(raw))
            record["status"] = "passed"
        except (RuntimeError, OSError, subprocess.SubprocessError) as error:
            record["error"] = str(error)
            record["failure_kind"] = checks.steps[-1].get("failure_kind", phase)
            checks.steps[-1]["status"] = "failed"
            raise
        finally:
            step = checks.steps[-1]
            record["log"] = step["log"]
            if "sha256" in step:
                record["sha256"] = step["sha256"]
            reports.append(record)
            # Partial receipts must retain the failed target, never an old
            # complete success list. No automatic retry or failure suppression.
            receipt.write_text(json.dumps(reports, indent=2) + "\n", encoding="utf-8")


def run(checks, suite):
    if suite in ("all", "fuzz"):
        run_fuzz(checks)
    if suite in ("all", "stress"):
        metrics = {"scope": "fixed loopback workload; logical reservations are not RSS; VmHWM includes the entire test process", "runs": {}, "mixed_runs": {}}
        for mode in ("pure", "race"):
            flags = ["-race"] if mode == "race" else []
            env = dict(ENV, CGO_ENABLED="1" if mode == "race" else "0")
            raw = checks.run("pressure-ro-" + mode, ["go", "test", "./examples/ro", "-run=^TestBoundedRoPressure$", "-count=20", "-json", "-timeout=3m"] + flags, extra=env, tests=True)
            metrics["runs"][mode] = pressure_samples(raw)
            raw = checks.run("pressure-mixed-" + mode, ["go", "test", "./examples/ro", "-run=^TestBoundedMixedApplicationPressure$", "-count=20", "-json", "-timeout=3m"] + flags, extra=env, tests=True)
            metrics["mixed_runs"][mode] = mixed_pressure_samples(raw)
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
