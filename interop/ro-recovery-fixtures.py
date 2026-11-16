#!/usr/bin/env python3
"""Independent header projection of frozen requests/answers, no Go codec imports."""
from pathlib import Path
import json,hashlib
root=Path(__file__).resolve().parents[1];dst=root/'testdata/codec'
def load(n):return bytes.fromhex((dst/(n+'.hex')).read_text())
out=[]
uses=[(k,'b47-'+k+'-request','b47-'+k+'-success','legacy') for k in ['iec','initial','term','general','balance','refund']]
uses += [('std-iec','v-iec-ccr-01','v-iec-cca-01','standard'),('std-initial','v-sms-ccr-01','v-sms-cca-01','standard'),('std-term','v-term-ccr-01','v-term-cca-01','standard')]
for i,(kind,q,a,profile) in enumerate(uses):
 for suffix,src,t in [('unwritten',q,False),('retransmit',q,True),('answer',a,False)]:
  raw=bytearray(load(src));original=bytes(raw);raw[12:16]=(0x65000000+i).to_bytes(4,'big')
  if t:raw[4]|=16
  else:raw[4]&=~16
  assert raw[16:]==original[16:]
  f='b50-'+kind+'-'+suffix+'.hex';(dst/f).write_text(raw.hex()+'\n')
  out.append(dict(file=f,bytes=len(raw),sha256=hashlib.sha256(raw).hexdigest(),source=src+'.hex',profile=profile,request=q+'.hex',answer=a+'.hex',expect='pair',provenance='independent header-only projection; unchanged application bytes/E2E; T0 assumes provenNotWritten, T1 assumes possiblepriorwrite; not targetOCS capture'))
m=json.load(open(dst/'manifest.json'));m=[v for v in m if not v['file'].startswith('b50-')]+out;(dst/'manifest.json').write_text(json.dumps(m,ensure_ascii=False,indent=2)+'\n');print(len(out),'new frames',len(m),'total')
