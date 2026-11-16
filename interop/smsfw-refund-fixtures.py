#!/usr/bin/env python3
"""Offline source projection, NOT an execution of C++ or an OCS capture.
Run from repo root: python3 interop/smsfw-refund-fixtures.py ../sms-fw
Uses only Python stdlib and the actual legacy template, never diameterkit codec.
"""
from pathlib import Path
import hashlib, json, struct, sys
root = Path(sys.argv[1]) / 'c-cpp/libsrc/kernel/dialib'
template = root / 'yppdia_refund_avptemplate.cf'
inputs = dict(SUBSCRIPTIONID_MSISDN='15551234567', SUBSCRIPTIONID_IMSI='001010123456789', TIMESTAMP=1700000000, DCS=8, INTERFACE_TYPE_DEST=2, MESSAGE_TYPE=0, INTERFACE_TYPE_ORIG=1, PID='00ff', REFUND_INFO='00ff800001')
rows=[]
for line in template.read_text().splitlines():
 line=line.split('#',1)[0].strip()
 if not line:continue
 n,rest=line.split('=',1);a=[x.strip() for x in rest.split(',')]
 rows.append((n.strip(),int(a[0]),int(a[1]),a[2].upper(),a[3:]))
index=0;mapping=[]
def avp(c,d,v=0,f=None):
 f=(192 if v else 64) if f is None else f
 n=(12 if f&128 else 8)+len(d)
 b=struct.pack('!I',c)+bytes([f])+n.to_bytes(3,'big')+(struct.pack('!I',v) if f&128 else b'')+d
 return b+bytes((-n)%4)
def u(c,n,v=0,f=None):return avp(c,struct.pack('!I',n&0xffffffff),v,f)
def field(parent):
 global index
 name,c,v,typ,args=rows[index];index+=1;path=parent+[f'{v}/{c}']
 mapping.append(dict(name=name,path='/'.join(path),type=typ,value=args))
 if typ=='GROUPED':
  children=[]
  for expected in args:
   assert rows[index][0]==expected,(expected,rows[index]);children.append(field(path))
  data=b''.join(children)
 else:
  val=inputs[args[0][1:]] if args[0].startswith('$') else args[0]
  if typ=='STRING':data=str(val).encode()
  elif typ=='OCTETSTRING':data=bytes.fromhex(val)
  elif typ=='TIME':data=struct.pack('!I',(int(val)+2208988800)&0xffffffff)
  elif typ=='UINT64':data=struct.pack('!Q',int(val))
  elif typ=='INTEGER64':data=struct.pack('!q',int(val))
  else:data=struct.pack('!I',int(val)&0xffffffff)
 return avp(c,data,v)
custom=[]
while index<len(rows):custom.append(field([]))
session=avp(263,b'ctf.example;legacy-refund;1')
def frame(body,flags=0xc0):return bytes([1])+(20+len(body)).to_bytes(3,'big')+bytes([flags])+(272).to_bytes(3,'big')+struct.pack('!III',4,0x31323334,0x41424344)+body
base=session+avp(264,b'ctf.example')+avp(296,b'example.net')+avp(283,b'example.net')+u(258,4)+avp(461,b'legacy-sms@example.net')+u(416,4)+u(415,0)+u(436,1)
frames={'b42-refund-request':frame(base+b''.join(custom)), 'b42-refund-enhanced':frame(base+avp(293,b'ocs.example')+u(278,0)+b''.join(custom))}
def answer(code,extra=b'',error=False):
 body=session+u(268,code)+avp(264,b'ocs.example')+avp(296,b'example.net')
 if not error:body+=u(258,4)+u(416,4)+u(415,0)
 return frame(body+extra,0x60 if error else 0x40)
frames.update({'b42-cca-success':answer(2001),'b42-cca-denied':answer(4012),'b42-cca-failed':answer(5012),'b42-cca-error':answer(3004,error=True)})
def entries(b):
 o=0
 while o<len(b):
  n=int.from_bytes(b[o+5:o+8],'big');f=b[o+4];v=int.from_bytes(b[o+8:o+12],'big') if f&128 else 0;h=12 if f&128 else 8
  yield (v,int.from_bytes(b[o:o+4],'big')),b[o+h:o+n],b[o:o+((n+3)&~3)]
  o+=(n+3)&~3

def change(raw,path,fn):
 def walk(b,p):
  out=b'';found=False
  for k,d,a in entries(b):
   if k==p[0]:
    found=True;a=fn(k,d,a) if len(p)==1 else avp(k[1],walk(d,p[1:]),k[0],a[4])
   out+=a
  assert found,p
  return out
 b=bytearray(raw[:20]+walk(raw[20:],path));b[1:4]=len(b).to_bytes(3,'big');return bytes(b)
r=frames['b42-refund-request'];a=frames['b42-cca-success'];si=(10415,873);sms=(10415,2000);mms=(10415,877)
def project(name,path,fn,raw=r):frames['b42-'+name]=change(raw,[(0,x) if isinstance(x,int) else x for x in path],fn)
def replace(name,path,d,raw=r):project(name,path,lambda k,_,old:avp(k[1],d,k[0],old[4]),raw)
money=[456,437,420];uv=money+[445]
for name,digits,exp,cur,units in [('minimum',-2**63,-2**31,-2**31,0),('maximum',2**63-1,2**31-1,2**31-1,2**64-1),('zero',0,0,0,0),('negative',-123,-2,840,1)]:
 raw=change(r,[(0,x) for x in uv+[447]],lambda k,d,a:avp(447,struct.pack('!q',digits)))
 raw=change(raw,[(0,x) for x in uv+[429]],lambda k,d,a:u(429,exp))
 raw=change(raw,[(0,x) for x in money+[425]],lambda k,d,a:u(425,cur))
 raw=change(raw,[(0,x) for x in [456,437,417]],lambda k,d,a:avp(417,struct.pack('!Q',units)))
 frames['b42-money-'+name]=raw
replace('empty-token',[1292],b'')
replace('token-changed',[1292],b'\0\xffchanged')
replace('digits-width4',uv+[447],bytes(4))
replace('exponent-width8',uv+[429],bytes(8))
replace('currency-width8',money+[425],bytes(8))
replace('unit-width4',[456,437,417],bytes(4))
replace('420-scalar',money,struct.pack('!I',1))
for name,path in [('money',money),('unit-value',uv),('digits',uv+[447]),('exponent',uv+[429]),('currency',money+[425]),('units',[456,437,417]),('token',[1292])]:
 project('missing-'+name,path,lambda k,d,a:b'')
project('duplicate-money',money,lambda k,d,a:a+a)
project('duplicate-token',[1292],lambda k,d,a:a+a)
project('413-alias',money,lambda k,d,a:avp(413,d))
project('money-vendor',money,lambda k,d,a:avp(420,d,10415))
project('token-vendor',[1292],lambda k,d,a:avp(1292,d,10415))
project('2022-alias',[1292],lambda k,d,a:avp(2022,d,10415))
project('wrong-parent',[456],lambda k,d,a:avp(456,b''.join(avp(437,b''.join(x for key,dd,x in entries(data) if key!=(0,420))) if key==(0,437) else raw for key,data,raw in entries(d)))+avp(420,avp(445,avp(447,struct.pack('!q',1))+u(429,0))+u(425,840),f=0))
replace('wrong-action',[436],bytes(4))
replace('wrong-number',[415],struct.pack('!I',1))
replace('wrong-node',[si,sms,(10415,2016)],bytes(4))
project('cca-money',[415],lambda k,d,a:a+avp(420,avp(445,avp(447,struct.pack('!q',1))+u(429,0))+u(425,840),f=0),raw=a)
project('cca-token',[415],lambda k,d,a:a+avp(1292,b'\0\xff',f=0),raw=a)
project('cca-gsu',[415],lambda k,d,a:a+avp(431,avp(417,struct.pack('!Q',1))),raw=a)
replace('cca-number',[415],struct.pack('!I',1),a)
manifest=json.load(open('testdata/codec/manifest.json'));manifest=[x for x in manifest if not x['file'].startswith('b42-')]
for name,b in frames.items():
 Path('testdata/codec',name+'.hex').write_text('\n'.join(b[i:i+32].hex(' ') for i in range(0,len(b),32))+'\n')
 manifest.append(dict(file=name+'.hex',bytes=len(b),sha256=hashlib.sha256(b).hexdigest(),source='interop/smsfw-refund-fixtures.py',provenance='Independent Python projection of actual C++ template and supplied input; synthetic paired CCA/counterexamples; not executed C++ or OCS capture'))
Path('testdata/codec/manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
files=['yppdia_refund_avptemplate.cf','dcca-wrapper/AvpsTemplate.cc','dcca-wrapper/AvpImpl.cc','dia-base/dia-message.cc','dia-cc/dia-cc.cc','dcca-wrapper/DccaClientWrapper.cc']
Path('interop/evidence/2026-11-07-smsfw-refund-source.json').write_text(json.dumps(dict(kind='source projection only',inputs=inputs,template_mapping=mapping,source_files=[dict(path='c-cpp/libsrc/kernel/dialib/'+f,sha256=hashlib.sha256((root/f).read_bytes()).hexdigest()) for f in files],fixtures=[x for x in manifest if x['file'].startswith('b42-')]),indent=2)+'\n')
print(len(frames),'frames;',len(mapping),'template nodes;',len(manifest),'total fixtures')
