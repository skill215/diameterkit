#!/usr/bin/env python3
"""Package current core production sources for offline, no-replace adapter checks.

This local-only version is not a Git tag/publication. It includes production Go
sources and LICENSE, not a complete source release. Re-run on changed sources
with a fresh module cache. No workspace, replace or external upload is created.
"""
import argparse
import hashlib
import json
from pathlib import Path
import zipfile

MODULE = 'github.com/skill215/diameterkit'
VERSION = 'v0.0.0-diameterkit.batch85'

def main():
    p = argparse.ArgumentParser(__doc__)
    p.add_argument('source', type=Path)
    p.add_argument('output', type=Path)
    a = p.parse_args()
    source, output = a.source.resolve(), a.output.resolve()
    if output == source or source in output.parents:
        p.error('output must be outside source')
    files = []
    for f in sorted(source.rglob('*')):
        rel = f.relative_to(source)
        if any(x.startswith('.') for x in rel.parts) or rel.parts[0] in ('adapters', 'interop'):
            continue
        if f.is_symlink():
            p.error(f'symlink not accepted: {rel}')
        if f.is_file() and (str(rel) in ('go.mod', 'LICENSE') or f.suffix == '.go' and not f.name.endswith('_test.go')):
            files.append((rel.as_posix(), f.read_bytes()))
    if dict(files)['go.mod'].splitlines()[0] != ('module '+MODULE).encode():
        p.error('unexpected module')
    dest = output / MODULE / '@v'
    dest.mkdir(parents=True, exist_ok=True)
    archive = dest / (VERSION+'.zip')
    with zipfile.ZipFile(archive, 'w', zipfile.ZIP_DEFLATED, compresslevel=9) as z:
        for name, data in files:
            info = zipfile.ZipInfo(f'{MODULE}@{VERSION}/{name}', (1980,1,1,0,0,0))
            info.create_system = 3
            info.external_attr = 0o100644 << 16
            info.compress_type = zipfile.ZIP_DEFLATED
            z.writestr(info, data)
    (dest/(VERSION+'.mod')).write_bytes(dict(files)['go.mod'])
    (dest/(VERSION+'.info')).write_text(json.dumps({'Version':VERSION,'Time':'2026-11-10T00:00:00Z'})+'\n')
    (dest/'list').write_text(VERSION+'\n')
    evidence = {'module':MODULE,'version':VERSION,'scope':'offline production-source candidate;not a published release','files':[{'path':name,'sha256':hashlib.sha256(data).hexdigest()} for name,data in files], 'zip_sha256':hashlib.sha256(archive.read_bytes()).hexdigest()}
    (output/'core-manifest.json').write_text(json.dumps(evidence,indent=2)+'\n')
    print(json.dumps({'files':len(files),'zip_sha256':evidence['zip_sha256']}))

if __name__ == '__main__':
    main()
