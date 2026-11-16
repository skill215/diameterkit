#!/usr/bin/env python3
"""Verify published modules in two independent cold consumers. Never publish.

Only the public Go proxy and checksum database are accepted. This gate cannot
fall back to a checkout, local proxy, preloaded cache, branch or offline batch.
"""
import argparse
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import tempfile
import urllib.parse
import urllib.request

from release_check import Checks, CORE, ADAPTER, UPSTREAM, ROOT, digest, objects

POLICY = ROOT / "spec/release/version-policy.json"
CONSUMER = "example.com/diameterkit-release-consumer"


def version(value):
    if not re.fullmatch(r"v0\.[1-9][0-9]*\.(?:0|[1-9][0-9]*)(?:-rc\.[1-9][0-9]*)?", value):
        raise ValueError("expected explicit v0.MINOR.PATCH[-rc.N], minor >= 1: " + value)
    return value


def environment(directory):
    # Inherited GOENV, GOFLAGS, private/proxy bypasses, architecture and caches
    # must not turn a public installation into a local checkout success.
    env = {k: v for k, v in os.environ.items() if not k.startswith("GO") and k != "CGO_ENABLED"}
    for name in ("mod", "build", "path", "tmp"):
        (directory / name).mkdir()
    env.update(GOENV="off", GOTOOLCHAIN="local", GOWORK="off", GOFLAGS="-mod=readonly",
               GOPROXY="https://proxy.golang.org", GOSUMDB="sum.golang.org",
               GOPRIVATE="", GONOPROXY="", GONOSUMDB="", GOINSECURE="",
               GOTELEMETRY="off", CGO_ENABLED="0", GOMAXPROCS="2",
               GOMODCACHE=str(directory / "mod"), GOCACHE=str(directory / "build"),
               GOPATH=str(directory / "path"), GOTMPDIR=str(directory / "tmp"))
    return env


def validate_manifest(manifest, path, requirements):
    actual = {r["Path"]: r["Version"] for r in (manifest.get("Require") or [])}
    if manifest.get("Module", {}).get("Path") != path or actual != requirements:
        raise RuntimeError("unexpected module requirements: " + path)
    if manifest.get("Replace") or manifest.get("Exclude"):
        raise RuntimeError("replace/exclude forbidden: " + path)


def validate_graph(raw, expected, cache, consumer):
    rows = list(objects(raw))
    if len(rows) != len(expected) + 1 or {r["Path"] for r in rows} != set(expected) | {CONSUMER}:
        raise RuntimeError("unexpected module graph")
    for row in rows:
        if row.get("Replace"):
            raise RuntimeError("replacement in installed graph")
        path = Path(row.get("Dir", "")).resolve()
        if row["Path"] == CONSUMER:
            if not row.get("Main") or path != consumer.resolve():
                raise RuntimeError("unexpected consumer directory")
        elif row.get("Main") or row.get("Version") != expected[row["Path"]] or not path.is_relative_to(cache.resolve()):
            raise RuntimeError("version upgrade or non-cache module: " + row["Path"])
    return rows


def validate_download(row, module, selected, cache):
    if row.get("Error") or row.get("Path") != module or row.get("Version") != selected:
        raise RuntimeError("wrong/failed module download: " + module)
    for key in ("Sum", "GoModSum"):
        if not re.fullmatch(r"h1:[A-Za-z0-9+/]{43}=", row.get(key, "")):
            raise RuntimeError("missing authenticated checksum: " + module)
    for key in ("Dir", "GoMod", "Zip", "Info"):
        if not Path(row.get(key, "")).resolve().is_relative_to(cache.resolve()):
            raise RuntimeError("download escaped private cache: " + module)
    return row


def tag_commit(tag):
    base = "https://api.github.com/repos/skill215/diameterkit/git/"
    def get(url):
        request = urllib.request.Request(url, headers={"User-Agent": "diameterkit-release-install"})
        with urllib.request.urlopen(request, timeout=30) as response:
            return json.load(response)
    obj = get(base + "ref/tags/" + urllib.parse.quote(tag, safe=""))["object"]
    for _ in range(4):
        if obj["type"] == "commit" and re.fullmatch("[0-9a-f]{40}", obj["sha"]):
            return obj["sha"]
        if obj["type"] != "tag" or not re.fullmatch("[0-9a-f]{40}", obj["sha"]):
            break
        obj = get(base + "tags/" + obj["sha"])["object"]
    raise RuntimeError("tag does not resolve to a commit: " + tag)


def validate_origin(row, tag, commit):
    origin = row.get("Origin", {})
    if (origin.get("VCS") != "git" or origin.get("URL") != "https://github.com/skill215/diameterkit"
            or origin.get("Hash") != commit or origin.get("Ref", "") not in ("", "refs/tags/" + tag)):
        raise RuntimeError("missing or mismatched tag/download origin: " + tag)


def copy_consumers(consumer, adapter):
    hashes = {}
    for folder in ("examples/contracts", "examples/ro"):
        destination = consumer / Path(folder).name
        destination.mkdir()
        for source in sorted((ROOT / folder).glob("*.go")):
            shutil.copyfile(source, destination / source.name)
            hashes[source.relative_to(ROOT).as_posix()] = digest(source)
    if adapter:
        destination = consumer / "sctp"
        destination.mkdir()
        source = ROOT / "spec/tools/testdata/installconsumer/sctp_test.go"
        shutil.copyfile(source, destination / source.name)
        hashes[source.relative_to(ROOT).as_posix()] = digest(source)
    return hashes


def check_api(checks, consumer, downloads):
    baseline = json.loads((ROOT / "spec/release/public-api.json").read_text())
    for target in baseline["targets"]:
        goos, arch = target["target"].split("/")
        selected = [p for p in target["packages"] if p["module"] in downloads]
        extra = {"GOOS": goos, "GOARCH": arch}
        patterns = [m + "/..." for m in downloads if m in (CORE, ADAPTER)]
        raw = checks.run("api-list-" + goos, ["go", "list", "-json", *patterns], consumer, extra)
        packages = {p["ImportPath"]: p for p in objects(raw)
                    if p["Name"] != "main" and "internal" not in p["ImportPath"].split("/") and p.get("GoFiles")}
        if set(packages) != {p["import_path"] for p in selected}:
            raise RuntimeError("installed public package inventory differs")
        for index, p in enumerate(selected):
            actual = packages[p["import_path"]]
            if actual.get("CgoFiles"):
                raise RuntimeError("cgo in public library")
            expected = {Path(f["path"]).name: f["sha256"] for f in p["files"]}
            if set(actual["GoFiles"]) != set(expected) or any(digest(Path(actual["Dir"]) / f) != h for f, h in expected.items()):
                raise RuntimeError("installed API source differs: " + p["import_path"])
            docs = checks.run(f"api-doc-{goos}-{index}", ["go", "doc", "-all", p["import_path"]], consumer, extra)
            if docs != p["documentation"]:
                raise RuntimeError("installed API documentation differs: " + p["import_path"])


def consume(directory, output, core, adapter, upstream):
    directory.mkdir()
    env = environment(directory)
    consumer = directory / "consumer"
    consumer.mkdir()
    checks = Checks(output, env)
    report = {"status": "failed", "steps": checks.steps, "cold_at_start": {}, "environment": {k: v for k, v in env.items() if k.startswith("GO") or k == "CGO_ENABLED"}}
    expected = {CORE: core}
    if adapter:
        expected.update({ADAPTER: adapter, UPSTREAM: upstream})
    try:
        for key in ("GOMODCACHE", "GOCACHE", "GOPATH"):
            report["cold_at_start"][key] = not any(Path(env[key]).iterdir())
        if not all(report["cold_at_start"].values()):
            raise RuntimeError("nonempty cache at start")
        wanted_go = (ROOT / ".go-version").read_text().strip()
        actual = checks.run("toolchain", ["go", "env", "GOVERSION", "GOHOSTOS", "GOHOSTARCH", "GOWORK"], consumer).splitlines()
        if actual != ["go" + wanted_go, "linux", "amd64", "off"]:
            raise RuntimeError("requires pinned Go on Linux/amd64 with GOWORK=off")
        report["consumer_sources"] = copy_consumers(consumer, bool(adapter))
        consumer_mod = "module " + CONSUMER + "\n\ngo 1.26.0\n\nrequire (\n" + "".join("\t" + m + " " + v + "\n" for m, v in expected.items()) + ")\n"
        (consumer / "go.mod").write_text(consumer_mod)
        downloads, failures = {}, []
        # Attempt every selected module, preserving real download failures instead
        # of treating a missing release as a skipped or passed installation.
        for index, (module, selected) in enumerate(expected.items()):
            try:
                raw = checks.run("download-" + str(index), ["go", "mod", "download", "-json", module + "@" + selected], consumer)
                downloads[module] = validate_download(json.loads(raw), module, selected, Path(env["GOMODCACHE"]))
            except (RuntimeError, ValueError) as error:
                failures.append(str(error))
        report["downloads"] = downloads
        if failures:
            raise RuntimeError("; ".join(failures))
        tags = {core: tag_commit(core)}
        if adapter:
            tag = "adapters/sctp/" + adapter
            tags[tag] = tag_commit(tag)
        report["tags"] = tags
        for module, row in downloads.items():
            requirements = {CORE: core, UPSTREAM: upstream} if module == ADAPTER else {}
            manifest = json.loads(checks.run("manifest-" + str(len(report.setdefault("manifests", {}))),
                ["go", "mod", "edit", "-json", row["GoMod"]], consumer))
            validate_manifest(manifest, module, requirements)
            report["manifests"][module] = manifest
            tag = core if module == CORE else "adapters/sctp/" + adapter if module == ADAPTER else None
            if tag:
                validate_origin(row, tag, tags[tag])
        checks.run("download-all", ["go", "mod", "download", "all"], consumer)
        # No further Go command may fetch a missing dependency; only the final
        # independent GitHub tag recheck still uses the network.
        checks.env.update(GOPROXY="off", GOSUMDB="off")
        manifest = json.loads(checks.run("consumer-manifest", ["go", "mod", "edit", "-json"], consumer))
        validate_manifest(manifest, CONSUMER, expected)
        raw = checks.run("module-graph", ["go", "list", "-m", "-json", "all"], consumer)
        report["graph"] = validate_graph(raw, expected, Path(env["GOMODCACHE"]), consumer)
        check_api(checks, consumer, downloads)
        deps = checks.run("dependencies", ["go", "list", "-deps", "-test", "-json", "./..."], consumer)
        for package in objects(deps):
            if package.get("CgoFiles"):
                raise RuntimeError("consumer requires cgo: " + package["ImportPath"])
            if not package.get("Standard") and not (Path(package["Dir"]).resolve().is_relative_to(consumer) or Path(package["Dir"]).resolve().is_relative_to(Path(env["GOMODCACHE"]))):
                raise RuntimeError("source outside installed consumer/cache")
        hashes = {name: digest(consumer / name) for name in ("go.mod", "go.sum")}
        for mode in ("pure", "race"):
            flags = ["-race"] if mode == "race" else []
            checks.run("consumer-" + mode, ["go", "test", *flags, "-count=1", "-timeout=5m", "-json", "./..."], consumer,
                       {"CGO_ENABLED": "1" if mode == "race" else "0"}, tests=True)
        checks.run("module-verify", ["go", "mod", "verify"], consumer)
        if any(digest(consumer / name) != sha for name, sha in hashes.items()):
            raise RuntimeError("consumer manifest/checksums changed during validation")
        for tag, commit in tags.items():
            if tag_commit(tag) != commit:
                raise RuntimeError("tag moved during acceptance")
        for name in ("go.mod", "go.sum"):
            shutil.copyfile(consumer / name, output / ("consumer." + name))
        report.update(status="passed", manifest_sha256=hashes)
    except (RuntimeError, OSError, ValueError, KeyError, subprocess.SubprocessError) as error:
        report["error"] = str(error)
    (output / "summary.json").write_text(json.dumps(report, indent=2) + "\n")
    return report


def execute(core, adapter, output):
    version(core)
    version(adapter)
    policy = json.loads(POLICY.read_text())
    # Fresh output prevents an old green result/artifact from surviving a failed run.
    output.mkdir(parents=True, exist_ok=False)
    report = {"format": 1, "status": "failed", "scope": "published public modules; two independent cold consumers",
              "versions": {CORE: core, ADAPTER: adapter, UPSTREAM: policy["upstream"]["version"]},
              "policy_sha256": digest(POLICY), "api_sha256": digest(ROOT / policy["api_review"]["baseline"]), "consumers": {}}
    with tempfile.TemporaryDirectory(prefix="diameterkit-install-") as temporary:
        for name, selected_adapter in (("core", None), ("adapter", adapter)):
            destination = output / name
            destination.mkdir()
            report["consumers"][name] = consume(Path(temporary) / name, destination, core, selected_adapter, policy["upstream"]["version"])
    if all(r["status"] == "passed" for r in report["consumers"].values()):
        report["status"] = "passed"
    (output / "summary.json").write_text(json.dumps(report, indent=2) + "\n")
    print(json.dumps({"status": report["status"], "summary": str(output / "summary.json")}))
    return 0 if report["status"] == "passed" else 1


def main():
    parser = argparse.ArgumentParser(__doc__)
    parser.add_argument("--core-version", required=True, type=version)
    parser.add_argument("--adapter-version", required=True, type=version)
    parser.add_argument("--output", required=True, type=Path, help="new, nonexistent evidence directory")
    args = parser.parse_args()
    return execute(args.core_version, args.adapter_version, args.output.resolve())


if __name__ == "__main__":
    raise SystemExit(main())
