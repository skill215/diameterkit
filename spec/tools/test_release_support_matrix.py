"""Release-metadata refusal tests, isolated from the checkout and network."""
import copy
import hashlib
import json
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import unittest

import check_support_matrix as check


class SupportMatrixTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.matrix = check.read_json(check.ROOT / "spec/release/support-matrix.json")
        for row in self.matrix["profiles"]:
            path = self.root / row["definition"]
            path.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(check.ROOT / row["definition"], path)

    def test_current_inventory(self):
        self.assertEqual(check.validate(self.root, self.matrix), 12)

    def test_stale_hash_and_definition_change_are_rejected_without_repair(self):
        for mutate_file in (False, True):
            with self.subTest(mutate_file=mutate_file):
                matrix = copy.deepcopy(self.matrix)
                path = self.root / matrix["profiles"][0]["definition"]
                if mutate_file:
                    path.write_bytes(path.read_bytes() + b"\n")
                else:
                    matrix["profiles"][0]["definition_sha256"] = "0" * 64
                before, baseline = path.read_bytes(), copy.deepcopy(matrix)
                with self.assertRaisesRegex(ValueError, "definition_sha256 differs"):
                    check.validate(self.root, matrix)
                self.assertEqual(path.read_bytes(), before)
                self.assertEqual(matrix, baseline)

    def test_inventory_metadata_and_paths(self):
        for kind in ("missing", "duplicate", "unknown", "profile", "version", "requests", "counts", "path", "absolute"):
            with self.subTest(kind=kind):
                matrix = copy.deepcopy(self.matrix)
                row = matrix["profiles"][0]
                if kind == "missing":
                    matrix["profiles"].pop()
                elif kind == "duplicate":
                    matrix["profiles"].append(copy.deepcopy(row))
                elif kind == "unknown":
                    row["binding"] = "unreviewed"
                elif kind in ("profile", "version"):
                    row[kind] += "-stale"
                elif kind == "requests":
                    row[kind][0]["command"] += 1
                elif kind == "counts":
                    matrix[kind]["generated_definitions"] += 1
                elif kind == "path":
                    row["definition"] = "../spec/base.json"
                else:
                    row["definition"] = str(self.root / row["definition"])
                with self.assertRaises(ValueError):
                    check.validate(self.root, matrix)

    def test_missing_file_and_escape_symlink(self):
        path = self.root / self.matrix["profiles"][0]["definition"]
        path.unlink()
        with self.assertRaises(OSError):
            check.validate(self.root, self.matrix)
        with tempfile.TemporaryDirectory() as outside:
            target = Path(outside) / "base.json"
            shutil.copyfile(check.ROOT / "spec/base.json", target)
            path.symlink_to(target)
            with self.assertRaisesRegex(ValueError, "escapes repository"):
                check.validate(self.root, self.matrix)

    def test_refreshed_hash_does_not_hide_semantic_drift(self):
        for binding, old, new in (("base", '"code": 257', '"code": 258'),
                                  ("sgdalert", 'const Version = "legacy-sgd-alert-v1.1"', 'const Version = "changed"'),
                                  ("sgdalert", 'Code: 8388648, Request: true', 'Code: 8388649, Request: true'),
                                  ("sgdalert", 'Code: 8388648, Request: true', 'Code: 8388648 + 0, Request: true')):
            with self.subTest(binding=binding, new=new):
                matrix = copy.deepcopy(self.matrix)
                row = next(r for r in matrix["profiles"] if r["binding"] == binding)
                path = self.root / row["definition"]
                original = (check.ROOT / row["definition"]).read_text()
                self.assertIn(old, original)
                path.write_text(original.replace(old, new))
                row["definition_sha256"] = hashlib.sha256(path.read_bytes()).hexdigest()
                with self.assertRaises(ValueError):
                    check.validate(self.root, matrix)
                path.write_text(original)

    def test_duplicate_json_keys(self):
        path = self.root / "duplicate.json"
        path.write_text('{"profiles": [], "profiles": []}')
        with self.assertRaisesRegex(ValueError, "duplicate JSON key"):
            check.read_json(path)

    def test_actual_cli_refuses_stale_baseline_without_rewriting(self):
        path = self.root / "stale.json"
        self.matrix["profiles"][0]["definition_sha256"] = "0" * 64
        path.write_text(json.dumps(self.matrix))
        before = path.read_bytes()
        result = subprocess.run([sys.executable, str(check.ROOT / "spec/tools/check_support_matrix.py"),
                                 "--matrix", str(path)], capture_output=True, text=True, timeout=30)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("definition_sha256 differs", result.stderr)
        self.assertEqual(path.read_bytes(), before)


if __name__ == "__main__":
    unittest.main()
