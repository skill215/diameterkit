#!/usr/bin/env python3
"""Bounded source-layout projections; not execution of sms-fw or a peer capture."""
import json,struct,hashlib
from pathlib import Path

def avp(c,d,flags=64):
 n=8+len(d);return struct.pack('!I',c)+bytes([flags])+n.to_bytes(3,'big')+d+bytes((-n)%4)
def u(c,n):return avp(c,struct.pack('!I',n))
v=u(266,193);a=u(258,19302);c=u(259,19302)
strict=avp(260,v+c);dual=avp(260,v+a+c)
base=avp(264,b'dra.example')+avp(296,b'example')+avp(257,b'\x00\x01\x7f\x00\x00\x01')+u(266,0)+avp(269,b'scap-peer')
def frame(groups,answer=False):
 body=base+(u(268,2001) if answer else b'')+groups
 return bytes([1])+(20+len(body)).to_bytes(3,'big')+bytes([0 if answer else 128])+int(257).to_bytes(3,'big')+struct.pack('!III',0,111,222)+body
cases={'strict':strict,'dual':strict+dual,'orphan':dual,'duplicate':strict+dual+dual,'reversed':dual+strict,'split':strict+u(278,1)+dual,'wrong-vendor':strict+avp(260,u(266,10415)+a+c),'wrong-auth':strict+avp(260,v+u(258,4)+c),'wrong-acct':strict+avp(260,v+a+u(259,4)),'child-flags':strict+avp(260,v+avp(258,struct.pack('!I',19302),0)+c),'outer-flags':strict+avp(260,v+a+c,0),'child-order':strict+avp(260,a+v+c),'extra':strict+avp(260,v+a+c+avp(999,b'x',0)),'width':strict+avp(260,v+avp(258,b'\x00\x4b\x66')+c),'strict-twice':strict+strict+dual}
frames={f'b78-scap-{name}-cer':frame(groups) for name,groups in cases.items()}
frames.update({'b78-scap-strict-cea':frame(strict,True),'b78-scap-dual-cea':frame(strict+dual,True)})
p=Path('testdata/codec/manifest.json');m=json.loads(p.read_text());m=[x for x in m if not x['file'].startswith('b78-scap-')]
for name,b in frames.items():
 Path('testdata/codec',name+'.hex').write_text(b.hex()+'\n')
 m.append(dict(file=name+'.hex',bytes=len(b),sha256=hashlib.sha256(b).hexdigest(),source='spec/batch-78.md',provenance='Independent Python projection of frozen C++ VSAI constructor/encoder layout; no C++ execution or capture',independent_offline_evidence='interop/evidence/2026-11-10-scap1-advertisement.json'))
p.write_text(json.dumps(m,ensure_ascii=False,indent=2)+'\n')
print(len(frames),'frames')
