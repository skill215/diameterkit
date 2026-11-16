"""Failure-path tests for the release gate, using no network or Go source edits."""
import json
from pathlib import Path
import tempfile
import unittest

import release_check as gate


class ReleaseGateTests(unittest.TestCase):
    def test_benchmark_requires_fixed_work_and_finite_complete_samples(self):
        line = "BenchmarkExample 200 12.5 ns/op 32 B/op 1 allocs/op\n"
        result = gate.load_checks.bench_samples(line * 5, {"BenchmarkExample"})
        self.assertEqual(result["BenchmarkExample"]["statistics"]["ns/op"]["median"], 12.5)
        for bad in (line * 4, (line * 5).replace("200 ", "100 "), (line * 5).replace("12.5", "NaN"), (line * 5).replace("1 allocs/op", "")):
            with self.subTest(output=bad), self.assertRaises(RuntimeError):
                gate.load_checks.bench_samples(bad, {"BenchmarkExample"})

    def test_pressure_requires_all_windows_and_real_resource_measurement(self):
        records = []
        for window, requests, canceled in [(1,16,8),(8,128,32),(16,256,64)]:
            for _ in range(20):
                records.append(dict(Window=window,Requests=requests,Canceled=canceled,Waves=16,Refusals=32,ClientOutstandingPeak=window,ServerInboundPeak=window,ProcessRSSHighWater=1024))
        def encode(rows):
            return "\n".join(json.dumps({"Output": "PRESSURE " + json.dumps(r)}) for r in rows)
        self.assertEqual(len(gate.load_checks.pressure_samples(encode(records))), 60)
        with self.assertRaises(RuntimeError):
            gate.load_checks.pressure_samples(encode(records[:-1]))
        records[0]["ProcessRSSHighWater"] = 0
        with self.assertRaises(RuntimeError):
            gate.load_checks.pressure_samples(encode(records))

    def test_graph_requires_both_current_checkouts(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            modules = [
                {"Path": gate.CORE, "Main": True, "Dir": str(root)},
                {"Path": gate.ADAPTER, "Main": True, "Dir": str(root / "adapters/sctp")},
                {"Path": gate.UPSTREAM, "Version": "v0.8.0"},
            ]
            raw = lambda rows: "\n".join(json.dumps(m) for m in rows)
            gate.validate_modules(raw(modules), root, "v0.8.0")
            for bad in ("cached-core", "missing-adapter", "replaced-upstream", "upstream-version", "extra-module"):
                with self.subTest(bad=bad):
                    rows = json.loads(json.dumps(modules))
                    if bad == "cached-core":
                        rows[0].update(Main=False, Dir=str(root / "cache/core"))
                    elif bad == "missing-adapter":
                        rows.pop(1)
                    elif bad == "replaced-upstream":
                        rows[2]["Replace"] = {"Dir": "/unreviewed"}
                    elif bad == "upstream-version":
                        rows[2]["Version"] = "v0.9.0"
                    else:
                        rows.append({"Path": "unreviewed.example/dependency"})
                    with self.assertRaises(RuntimeError):
                        gate.validate_modules(raw(rows), root, "v0.8.0")

    def test_failed_skipped_or_empty_test_output_is_not_green(self):
        passed = {"Action": "pass", "Test": "TestNative", "Package": "example.com/test"}
        raw = json.dumps(passed) + "\n" + json.dumps({"Action": "skip", "Package": "example.com/no-tests"})
        self.assertEqual(gate.test_counts(raw), {"test_subtest_seed_passes": 1, "test_skips": 0})
        for bad in ("", json.dumps({"Action": "skip", "Test": "TestNative"}),
                    raw + "\n" + json.dumps({"Action": "fail", "Package": "example.com/test"})):
            with self.subTest(output=bad), self.assertRaises(RuntimeError):
                gate.test_counts(bad)

    def test_failed_command_retains_log_and_failure(self):
        import os
        import sys
        with tempfile.TemporaryDirectory() as directory:
            output = Path(directory)
            checks = gate.Checks(output, os.environ.copy())
            with self.assertRaises(RuntimeError):
                checks.run("failure", [sys.executable, "-c", "print('intentional failure'); raise SystemExit(7)"])
            self.assertEqual(checks.steps[0]["status"], "failed")
            self.assertIn("intentional failure", (output / "failure.log").read_text())
            self.assertEqual(checks.steps[0]["sha256"], gate.digest(output / "failure.log"))


if __name__ == "__main__":
    unittest.main()
