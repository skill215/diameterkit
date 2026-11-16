#!/usr/bin/env python3
"""Independent fixed-header projection of previously audited S6c wire vectors.

No Go codec/generator is used. These are synthetic compatibility vectors, not
captures or proof that the legacy application handled every selected S6c field.
"""
import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
DIR = ROOT / 'testdata/codec'
SOURCES = [
    'v-alr-01', 'v-ala-01', 'v-ala-02',
    'v-alr-move-01', 'v-ala-move-01', 'v-ala-move-02',
    'v21-sgsn-alr-r', 'v21-sgsn-alr-a', 'v21-sgsn-alr-f',
    'v21-ims-alr-r', 'v21-ims-alr-a', 'v21-ims-alr-f',
    'd21-t4-alr-r', 'd21-t4-alr-a', 'd21-t4-alr-f',
    'b23-alr-reachable', 'b23-alr-move-nonode', 'b23-alr-both',
    'b23-alr-future-bits', 'b23-alr-zero', 'b23-alr-noevent',
    'b23-alr-reachable-node', 'v21-router-alr-r',
    'n26-alr-originuri', 'n26-alr-hssid-nested', 'n27-alr-uri3327',
]

def main():
    manifest = json.loads((DIR/'manifest.json').read_text())
    manifest = [e for e in manifest if not e['file'].startswith('b73-')]
    for name in SOURCES:
        source = DIR/(name+'.hex')
        raw = bytearray.fromhex(source.read_text())
        assert int.from_bytes(raw[5:8], 'big') == 8388648
        assert int.from_bytes(raw[8:12], 'big') == 16777312
        raw[8:12] = (16777313).to_bytes(4, 'big')
        output = 'b73-'+name+'.hex'
        (DIR/output).write_text(raw.hex()+'\n')
        manifest.append(dict(file=output, bytes=len(raw), sha256=hashlib.sha256(raw).hexdigest(),
                             source='spec/batch-73.md',
                             provenance='Synthetic fixed Application-Id projection from '+name+'.hex; no Go codec/generator; not a peer capture',
                             source_file_sha256=hashlib.sha256(source.read_bytes()).hexdigest()))
    (DIR/'manifest.json').write_text(json.dumps(manifest, ensure_ascii=False, indent=2)+'\n')
    print(len(SOURCES), 'projected frames;', len(manifest), 'total frames')

if __name__ == '__main__':
    main()
