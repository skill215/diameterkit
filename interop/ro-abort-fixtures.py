#!/usr/bin/env python3
"""Independent RFC6733 8.5 App4 wire construction; no diameterkit encoder.

Synthetic protocol evidence, not an OCS capture or session-state oracle.
Run from the repository root. Keeps all older manifest entries unchanged.
"""
import hashlib
import json
import struct
from pathlib import Path

def avp(code, data, flags=64, vendor=0):
    size = 8 + (4 if flags & 128 else 0) + len(data)
    b = struct.pack('!I', code) + bytes([flags]) + size.to_bytes(3, 'big')
    if flags & 128:
        b += struct.pack('!I', vendor)
    return b + data + bytes((-size) % 4)

def u(code, value):
    return avp(code, struct.pack('!I', value))

def frame(parts, flags):
    body = b''.join(parts)
    return b'\x01' + (20+len(body)).to_bytes(3, 'big') + bytes([flags]) + (274).to_bytes(3, 'big') + struct.pack('!III', 4, 0xb1b2b369, 0xc1c2c369) + body

session = avp(263, b'ctf.example;abort;69')
proxy = avp(284, avp(280, b'dra.example') + avp(33, b'\x00\xff'))
request = [session, avp(264,b'ocs.example'), avp(296,b'example'), avp(283,b'example'), avp(293,b'ctf.example'), u(258,4)]
options = [avp(1,'用户@example'.encode()),u(278,0),proxy,avp(282,b'dra.example')]
answer = [session,u(268,2001),avp(264,b'ctf.example'),avp(296,b'example')]
frames = {}
def put(name, parts, flags):
    frames['b69-'+name] = frame(parts,flags)

put('asr',request+options,192)
put('asr-min',request,192)
for code in [2001,5002,5012]:
    parts=answer.copy();parts[1]=u(268,code)
    put('asa-'+str(code),parts+[proxy],64)
put('asa-options',answer+[avp(1,b'user@example'),u(278,7),avp(281,'拒绝说明'.encode(),0),avp(294,b'ctf.example',0),avp(279,u(415,99)),proxy],64)
put('asa-e',answer[:1]+answer[2:]+[u(268,3004)]+[proxy],96)
put('asa-e-diagnostic',answer[:1]+answer[2:]+[u(268,5005)]+[avp(279,avp(293,b'')),proxy],96)
put('asr-unknown-optional',request+options+[avp(999001,b'\x00\xff',0)],192)
put('asa-unknown-optional',answer+[proxy,avp(999001,b'\x00\xff',0)],64)
for index,name in enumerate(['session','origin-host','origin-realm','destination-realm','destination-host','auth-app']):
    put('asr-missing-'+name,request[:index]+request[index+1:],192)
    put('asr-duplicate-'+name,request+request[index:index+1],192)
for index,name in enumerate(['session','result','origin-host','origin-realm']):
    put('asa-missing-'+name,answer[:index]+answer[index+1:],64)
    put('asa-duplicate-'+name,answer+answer[index:index+1],64)
put('asr-wrong-app',request[:-1]+[u(258,3)],192)
put('asr-width',request[:-1]+[avp(258,b'\x00\x00\x04')],192)
put('asr-unknown-mandatory',request+[avp(999001,b'x')],192)
put('asa-unknown-mandatory',answer+[avp(999001,b'x')],64)
put('asr-bad-user',request+[avp(1,b'\xff')],192)
put('asr-empty-destination',request[:4]+[avp(293,b'')]+request[5:],192)
put('asa-zero-result',answer[:1]+[u(268,0)]+answer[2:],64)
put('asa-unselected-result',answer[:1]+[u(268,4012)]+answer[2:],64)
put('asa-normal-3004',answer[:1]+[u(268,3004)]+answer[2:],64)
put('asa-e-success',answer,96)
put('asa-e-missing-diagnostic',answer[:1]+[u(268,5004)]+answer[2:],96)
put('asa-e-experimental',answer[:1]+[u(268,5012)]+answer[2:]+[avp(297,u(266,10415)+u(298,5011))],96)
for code in [258,415,416,436,456,292,297]:
    put('asa-unselected-'+str(code),answer+[avp(code,bytes(4),0)],64)
put('asr-result',request+[avp(268,struct.pack('!I',2001),0)],192)
put('asa-wrong-session',[avp(263,b'ctf.example;other')]+answer[1:]+[proxy],64)
put('asa-wrong-source',answer[:2]+[avp(264,b'other.example')]+answer[3:]+[proxy],64)
put('asa-wrong-proxy',answer+[avp(284,avp(280,b'dra.example')+avp(33,b'changed'))],64)
manifest=json.loads(Path('testdata/codec/manifest.json').read_text())
manifest=[x for x in manifest if not x['file'].startswith('b69-')]
for name,b in frames.items():
    Path('testdata/codec',name+'.hex').write_text('\n'.join(b[i:i+32].hex(' ') for i in range(0,len(b),32))+'\n')
    manifest.append(dict(file=name+'.hex',bytes=len(b),sha256=hashlib.sha256(b).hexdigest(),source='interop/ro-abort-fixtures.py',provenance='Independent Python RFC6733 ASR/ASA App4 construction; synthetic, not OCS evidence'))
Path('testdata/codec/manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
print(len(frames),'ASR/ASA fixtures;',len(manifest),'total')
