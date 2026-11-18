#!/usr/bin/env python3
"""Check current support metadata without rewriting definitions or baselines.

Python is release tooling only. Delegated Go metadata uses deliberately narrow,
reviewed declaration shapes; an unfamiliar shape requires review, not inference.
"""
import argparse
import hashlib
import json
from pathlib import Path
import re

from release_check import GENERATED

ROOT = Path(__file__).resolve().parents[2]


def unique_object(pairs):
    result = {}
    for key, value in pairs:
        if key in result:
            raise ValueError("duplicate JSON key: " + key)
        result[key] = value
    return result


def read_json(path):
    return json.loads(path.read_text(encoding="utf-8"), object_pairs_hook=unique_object)


def one(pattern, source):
    matches = re.findall(pattern, source, re.MULTILINE)
    if len(matches) != 1:
        raise ValueError("delegated declaration changed; review its metadata extractor")
    return matches[0]


def delegated_metadata(path):
    # Only these literal declarations are accepted. Comments cannot satisfy them.
    source = path.read_text(encoding="utf-8")
    source = re.sub(r'/\*.*?\*/|//[^\n]*', '', source, flags=re.DOTALL)
    app = int(one(r'^const ApplicationID uint32 = ([0-9]+)\s*$', source))
    version = one(r'^const Version = "([^"\n]+)"\s*$', source)
    profile = one(r'^\s*return schema.NewProfile\("([^"\n]+)", Version, hex.EncodeToString\(sum\[:\]\), bind\(r\), bind\(a\)\)\s*$', source)
    command = int(one(r'func \(\*ALR\) CommandKey\(\) pdu.CommandKey \{\s*return pdu.CommandKey\{ApplicationID: ApplicationID, Code: ([0-9]+), Request: true\}\s*\}', source))
    answer = int(one(r'func \(\*ALA\) CommandKey\(\) pdu.CommandKey \{\s*return pdu.CommandKey\{ApplicationID: ApplicationID, Code: ([0-9]+)\}\s*\}', source))
    if command != answer:
        raise ValueError("delegated request/answer identity differs")
    return profile, version, [dict(name="ALR", application=app, command=command)]


def validate(root, matrix):
    root = root.resolve()
    inventory = {name: "spec/" + name + ".json" for name in GENERATED}
    inventory["sgdalert"] = "compat/sgdalert/profile.go"
    if matrix.get("format") != 1 or not isinstance(matrix.get("profiles"), list):
        raise ValueError("unsupported support matrix format")
    rows = matrix["profiles"]
    bindings = [row["binding"] for row in rows]
    if len(bindings) != len(set(bindings)) or set(bindings) != set(inventory):
        raise ValueError("support matrix binding inventory differs")
    pairs = 0
    for row in rows:
        binding = row["binding"]
        name = row["definition"]
        if name != inventory[binding]:
            raise ValueError(binding + ": unexpected definition path")
        path = (root / name).resolve(strict=True)
        if not path.is_relative_to(root) or not path.is_file():
            raise ValueError(binding + ": definition escapes repository or is not a file")
        actual = hashlib.sha256(path.read_bytes()).hexdigest()
        if row["definition_sha256"] != actual:
            raise ValueError(binding + ": definition_sha256 differs")
        if binding == "sgdalert":
            profile, version, requests = delegated_metadata(path)
        else:
            definition = read_json(path)
            profile, version = definition["profile"], definition["version"]
            requests = [dict(name=c["name"], application=c["application"], command=c["code"])
                        for c in definition["commands"] if c["request"] is True]
        expected = dict(profile=profile, version=version, requests=requests)
        for key, value in expected.items():
            if row.get(key) != value:
                raise ValueError(binding + ": " + key + " differs")
        pairs += len(requests)
    expected_counts = dict(generated_definitions=len(GENERATED), delegated_profiles=1,
                           request_answer_pairs_including_overlapping_bindings=pairs)
    if matrix.get("counts") != expected_counts:
        raise ValueError("support matrix counts differ")
    return len(rows)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--matrix", type=Path, default=ROOT / "spec/release/support-matrix.json")
    args = parser.parse_args()
    try:
        count = validate(ROOT, read_json(args.matrix))
    except (ValueError, OSError, KeyError, TypeError) as error:
        parser.exit(1, "support matrix check failed: " + str(error) + "\n")
    print(f"Support matrix checked: {count} profiles; hashes, identities, requests and counts match")


if __name__ == "__main__":
    main()
