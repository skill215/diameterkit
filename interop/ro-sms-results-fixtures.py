#!/usr/bin/env python3
"""Independent result/2022 projections of frozen standard SMS Ro pairs."""
from pathlib import Path
import hashlib,json
root=Path(__file__).resolve().parents[1];dst=root/'testdata/codec'
def load(n):return bytes.fromhex((dst/(n+'.hex')).read_text())
def avp(c,d,v=0,flags=None):
 if flags is None:flags=192 if v else 64
 n=8+len(d)+(4 if flags&128 else 0)
 return c.to_bytes(4,'big')+bytes([flags])+n.to_bytes(3,'big')+(v.to_bytes(4,'big') if flags&128 else b'')+d+bytes(-n%4)
def u(c,n):return avp(c,n.to_bytes(4,'big'))
def parts(b):
 off=0
 while off<len(b):
  c=int.from_bytes(b[off:off+4],'big');f=b[off+4];n=int.from_bytes(b[off+5:off+8],'big');v=int.from_bytes(b[off+8:off+12],'big') if f&128 else 0;h=12 if f&128 else 8
  yield (v,c),b[off+h:off+n],b[off:off+(n+3)//4*4]
  off+=(n+3)//4*4
 assert off==len(b)
def frame(seed,top,child=None,token=None,grant=False):
 body=b''
 for key,data,raw in parts(seed[20:]):
  if key==(0,268):raw=u(268,top)
  if key==(0,456):continue
  body+=raw
 if child is not None:
  group=u(439,1)+u(268,child)
  if grant:group+=avp(431,avp(417,(1).to_bytes(8,'big')))
  if token is not None:group+=avp(2022,token,10415)
  body+=avp(456,group)
 return seed[:1]+(20+len(body)).to_bytes(3,'big')+seed[4:20]+body
out=[]
def put(name,b,request,expect='pair'):
 file='b71-'+name+'.hex';(dst/file).write_text(b.hex()+'\n');out.append(dict(file=file,bytes=len(b),sha256=hashlib.sha256(b).hexdigest(),request=request+'.hex',expect=expect,source='interop/ro-sms-results-fixtures.py',provenance='Independent Python projection of frozen standardSMS pairs;no OCS or C++execution'))
uses=[('iec','v-iec-cca-01','v-iec-ccr-01'),('initial','v-sms-cca-01','v-sms-ccr-01'),('term','v-term-cca-01','v-term-ccr-01')]
for kind,seed,req in uses:
 for code in [4010,4011,5030]:
  for label,top,child in [('top',code,None),('child',2001,code),('both',code,code),('top-failure-child-success',code,2001)]:put(kind+'-'+label+'-'+str(code),frame(load(seed),top,child),req)
seed=load('v-iec-cca-01');req='v-iec-ccr-01'
for label,token in [('absent',None),('empty',b''),('binary',b'\x00\xff\x00\x01')]:put('token-'+label,frame(seed,2001,2001,token,True),req)
put('token-failure',frame(seed,4010,4010,b'\x00\xff'),req)
put('no-parent',frame(seed,4011),req)
for label,code in [('unselected',2002),('rating',5031)]:put('child-'+label,frame(seed,2001,code),req,'reject')
# Parent/identity ambiguity never becomes first/last-wins token selection.
b=frame(seed,2001,2001,b'\x00\xff',True)
body=b[20:]
group=next(raw for key,data,raw in parts(body) if key==(0,456))
for name,newbody,expect in [('duplicate-parent',body+group,'reject'),('top-token',body+avp(2022,b'\x00\xff',10415,128),'reject')]:put(name,b[:1]+(20+len(newbody)).to_bytes(3,'big')+b[4:20]+newbody,req,expect)
newbody=b''
for key,data,raw in parts(body):
 if key==(0,456):raw=avp(456,b''.join(u(439,2) if k==(0,439) else r for k,d,r in parts(data)))
 newbody+=raw
put('wrong-service',b[:1]+(20+len(newbody)).to_bytes(3,'big')+b[4:20]+newbody,req,'association')
manifest=json.loads((dst/'manifest.json').read_text());manifest=[x for x in manifest if not x['file'].startswith('b71-')]+out
(dst/'manifest.json').write_text(json.dumps(manifest,indent=2)+'\n');print(len(out),'new;',len(manifest),'total')
