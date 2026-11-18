"""Failure controls for count-bounded fuzz acceptance and retained artifacts."""
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest
from unittest import mock

import load_checks as load
import release_check as gate


def completed():
    return ("fuzz: elapsed: 0s, gathering baseline coverage: 1/1 completed, now fuzzing with 2 workers\n"
            f"fuzz: elapsed: 1s, execs: {load.FUZZ_EXECUTIONS} (1000/sec), new interesting: 0 (total: 1)\n"
            "PASS\nok  example.test/fuzz 1.0s\n")


class ScriptChecks(gate.Checks):
    """Exercise real subprocess handling while controlling output and exits."""
    def __init__(self, output, scripts, timeout=5):
        super().__init__(output, os.environ.copy())
        self.scripts, self.timeout, self.commands = scripts, timeout, []

    def run(self, name, command, cwd=gate.ROOT, extra=None):
        self.commands.append(command)
        script = self.scripts[len(self.commands)-1]
        return super().run(name, [sys.executable, "-c", script], cwd, extra,
                           timeout=self.timeout)


class FuzzGateTests(unittest.TestCase):
    def test_requires_complete_count_and_real_fuzz_markers(self):
        good = completed()
        self.assertEqual(load.fuzz_measurements(good), {"executions": 1_000_000})
        bad = ["PASS\n", good.replace("1000000", "999999"),
               good.replace("1000000", "1000001"),
               good.replace("now fuzzing with 2 workers", "seed-only"),
               good.replace("PASS", "NOT PASS"), good.replace("ok  ", "no  "),
               good+"--- FAIL: FuzzExample\n", good+"panic: bad\n",
               good+"context deadline exceeded\n"]
        for raw in bad:
            with self.subTest(raw=raw), self.assertRaises(RuntimeError):
                load.fuzz_measurements(raw)

    def test_partial_receipt_preserves_failure_and_corpus_without_retry(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            output = root / "out"; output.mkdir()
            corpus = root / "example/testdata/fuzz/FuzzBad"; corpus.mkdir(parents=True)
            seed = corpus / "input"
            original = b'go test fuzz v1\n[]byte("bad")\n'
            seed.write_bytes(original)
            targets = [("good", ".", "./example", "FuzzGood"),
                       ("bad", ".", "./example", "FuzzBad"),
                       ("never", ".", "./example", "FuzzNever")]
            (output/"fuzz.json").write_text('[{"status":"passed","target":"stale"}]')
            checks = ScriptChecks(output, [
                "print("+repr(completed())+")",
                "print("+repr(completed()+"context deadline exceeded\n")+");raise SystemExit(7)"])
            with mock.patch.object(load, "ROOT", root), mock.patch.object(load, "FUZZ_TARGETS", targets):
                with self.assertRaisesRegex(RuntimeError, "exited 7"):
                    load.run_fuzz(checks)
            receipt = json.loads((output/"fuzz.json").read_text())
            self.assertEqual([r["status"] for r in receipt], ["passed", "failed"])
            self.assertEqual(receipt[-1]["failure_kind"], "command_exit")
            self.assertEqual(len(checks.commands), 2)
            self.assertTrue(all("-fuzztime=1000000x" in c and "-timeout=2m" in c
                                and "-parallel=2" in c and "-fuzzminimizetime=5s" in c
                                for c in checks.commands))
            self.assertEqual(checks.steps[-1]["returncode"], 7)
            artifact = receipt[-1]["corpus"][0]
            self.assertEqual((output/artifact["path"]).read_bytes(), original)
            self.assertEqual(artifact["sha256"], hashlib.sha256(original).hexdigest())
            self.assertEqual(seed.read_bytes(), original)
            self.assertEqual(receipt[-1]["sha256"], gate.digest(output/receipt[-1]["log"]))

    def test_zero_exit_with_insufficient_work_remains_failed(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory); output = root/"out"; output.mkdir()
            checks = ScriptChecks(output, ["print("+repr(completed().replace("1000000","999999"))+")"])
            with mock.patch.object(load, "ROOT", root), mock.patch.object(load, "FUZZ_TARGETS",
                    [("short", ".", ".", "FuzzShort")]):
                with self.assertRaisesRegex(RuntimeError, "incomplete"):
                    load.run_fuzz(checks)
            record = json.loads((output/"fuzz.json").read_text())[0]
            self.assertEqual(record["status"], "failed")
            self.assertEqual(record["failure_kind"], "output")
            self.assertEqual(checks.steps[0]["status"], "failed")
            self.assertEqual(checks.steps[0]["returncode"], 0)

    def test_outer_timeout_keeps_log_hash_and_corpus(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory); output=root/"out";output.mkdir()
            corpus=root/"testdata/fuzz/FuzzSlow";corpus.mkdir(parents=True)
            (corpus/"seed").write_text("preserved")
            checks = ScriptChecks(output, [
                "import time;print('before timeout',flush=True);time.sleep(5)"], timeout=0.2)
            with mock.patch.object(load, "ROOT", root), mock.patch.object(load, "FUZZ_TARGETS",
                    [("slow", ".", ".", "FuzzSlow")]):
                with self.assertRaises(subprocess.TimeoutExpired):
                    load.run_fuzz(checks)
            record=json.loads((output/"fuzz.json").read_text())[0]
            self.assertEqual(record["status"], "failed")
            self.assertEqual(record["failure_kind"], "process_timeout")
            self.assertIn("before timeout", (output/record["log"]).read_text())
            self.assertEqual(record["sha256"], gate.digest(output/record["log"]))
            self.assertEqual((output/record["corpus"][0]["path"]).read_text(), "preserved")


if __name__ == "__main__":
    unittest.main()
