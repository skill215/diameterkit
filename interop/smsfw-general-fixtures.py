#!/usr/bin/env python3
"""Offline source projection, NOT an execution of C++ or an OCS capture.
Run from repo root: python3 interop/smsfw-general-fixtures.py ../sms-fw
Uses only Python stdlib and the actual legacy template, never diameterkit codec.
"""
from pathlib import Path
import hashlib, json, struct, sys
root = Path(sys.argv[1]) / 'c-cpp/libsrc/kernel/dialib'
template = root / 'yppdia_avptemplate.cf'
inputs = dict(SUBSCRIPTIONID_MSISDN='15551234567', SUBSCRIPTIONID_IMSI='001010123456789', TIMESTAMP=1700000000, DCS=8, INTERFACE_TYPE_DEST=2, MESSAGE_TYPE=0, INTERFACE_TYPE_ORIG=1, PID='00ff', ORIG_MSISDN='15551234567')
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
  if typ in ('STRING','UTF8STRING'):data=str(val).encode()
  elif typ=='IPV4':data=b'\x00\x01'+bytes(int(x) for x in str(val).split('.'))
  elif typ=='OCTETSTRING':data=bytes.fromhex(val)
  elif typ=='TIME':data=struct.pack('!I',(int(val)+2208988800)&0xffffffff)
  elif typ=='UINT64':data=struct.pack('!Q',int(val))
  elif typ=='INTEGER64':data=struct.pack('!q',int(val))
  else:data=struct.pack('!I',int(val)&0xffffffff)
 return avp(c,data,v)
custom=[]
while index<len(rows):custom.append(field([]))
session=avp(263,b'ctf.example;legacy-general;1')
def frame(body,flags=0xc0):return bytes([1])+(20+len(body)).to_bytes(3,'big')+bytes([flags])+(272).to_bytes(3,'big')+struct.pack('!III',4,0x31323334,0x41424344)+body
base=session+avp(264,b'ctf.example')+avp(296,b'example.net')+avp(283,b'example.net')+u(258,4)+avp(461,b'legacy-sms@example.net')+u(416,4)+u(415,0)+u(436,0)
frames={'b44-general-request':frame(base+b''.join(custom)), 'b44-general-enhanced':frame(base+avp(293,b'ocs.example')+u(278,0)+b''.join(custom))}
def answer(code,extra=b'',error=False):
 body=session+u(268,code)+avp(264,b'ocs.example')+avp(296,b'example.net')
 if not error:body+=u(258,4)+u(416,4)+u(415,0)
 return frame(body+extra,0x60 if error else 0x40)
frames.update({'b44-cca-success':answer(2001),'b44-cca-denied':answer(4012),'b44-cca-failed':answer(5012),'b44-cca-error':answer(3004,error=True)})
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
r=frames['b44-general-request'];a=frames['b44-cca-success'];si=(10415,873);sms=(10415,2000);mms=(10415,877)
def project(name,path,fn,raw=r):frames['b44-'+name]=change(raw,[(0,x) if isinstance(x,int) else x for x in path],fn)
def replace(name,path,d,raw=r):project(name,path,lambda k,_,old:avp(k[1],d,k[0],old[4]),raw)

# Selected actual answer fields; these answers are synthetic, not OCS captures.
def cost(digits=-123,exp=-2,currency=840):
 return avp(423,avp(445,avp(447,struct.pack('!q',digits))+(u(429,exp) if exp is not None else b''))+u(425,currency))
def grant(units=1):return avp(431,avp(417,struct.pack('!Q',units)))
frames['b44-cca-cost-units']=answer(2001,grant()+cost())
frames['b44-cca-cost-absent-exponent']=answer(2001,cost(exp=None))
frames['b44-cca-cost-zero']=answer(2001,grant(0)+cost(0,0,0))
frames['b44-cca-cost-limits']=answer(2001,grant(2**64-1)+cost(-2**63,-2**31,2**32-1))
frames['b44-cca-cost-unit']=answer(2001,avp(423,avp(445,avp(447,struct.pack('!q',1)))+u(425,840)+avp(424,'每条'.encode())))
replace('balance-request',[436],struct.pack('!I',2))
frames['b44-balance-enough']=answer(2001,u(422,0))
frames['b44-balance-no-credit']=answer(2001,u(422,1))
frames['b44-balance-grant']=answer(2001,grant()+u(422,0))
frames['b44-balance-unknown']=answer(2001,u(422,2))
frames['b44-failed-cost']=answer(4012,cost())
frames['b44-failed-balance']=answer(5012,u(422,0))
frames['b44-failed-grant']=answer(4012,grant())
money=[456,437,413];uv=money+[445]
replace('money-negative',uv+[447],struct.pack('!q',-123))
replace('currency-negative',money+[425],struct.pack('!i',-1))
replace('raw-originator',[si,mms,(10415,886),(10415,897)],b'\xfforigin')
replace('nul-originator',[si,mms,(10415,886),(10415,897)],b'bad\0')
replace('digits-width4',uv+[447],bytes(4))
replace('units-width4',[456,437,417],bytes(4))
replace('draft-address',[si,sms,(10415,2018)],bytes([127,0,0,1]))
replace('string-address',[si,sms,(10415,2018)],b'127.0.0.1')
replace('ipv6-address',[si,sms,(10415,2018)],b'\x00\x02'+bytes(16))
replace('timezone-nonzero',[(10425,23)],struct.pack('!I',1))
replace('service-nonliteral',[439],struct.pack('!I',2))
replace('refund-action',[436],struct.pack('!I',1))
replace('price-action',[436],struct.pack('!I',3))
for name,path in [('money',money),('exponent',uv+[429]),('currency',money+[425]),('timezone',[(10425,23)]),('service',[439]),('client',[si,sms,(10415,2018)]),('mms',[si,mms])]:
 project('missing-'+name,path,lambda k,d,a:b'')
project('wrong-timezone-vendor',[(10425,23)],lambda k,d,a:avp(23,d,10415))
project('money420-alias',money,lambda k,d,a:avp(420,d))
project('both-money',money,lambda k,d,a:a+avp(420,d))
project('duplicate-money',money,lambda k,d,a:a+a)
project('unexpected-token',[415],lambda k,d,a:a+avp(1292,b'opaque'))
ca=frames['b44-cca-cost-units']
for name,path in [('unit-value',[423,445]),('digits',[423,445,447]),('currency',[423,425]),('units',[431,417])]:
 project('cost-missing-'+name,path,lambda k,d,a:b'',ca)
replace('cost-bad-digits',[423,445,447],bytes(4),ca)
replace('cost-bad-exponent',[423,445,429],bytes(8),ca)
replace('cost-bad-currency',[423,425],bytes(8),ca)
project('duplicate-cost',[423],lambda k,d,a:a+a,ca)
project('cost-wrong-parent',[423],lambda k,d,a:avp(456,a),ca)
project('unknown-cost-child',[423],lambda k,d,a:avp(423,d+avp(9999,b'x',f=0)),ca)
replace('cca-number',[415],struct.pack('!I',1),a)
manifest=json.load(open('testdata/codec/manifest.json'));manifest=[x for x in manifest if not x['file'].startswith('b44-')]
for name,b in frames.items():
 Path('testdata/codec',name+'.hex').write_text('\n'.join(b[i:i+32].hex(' ') for i in range(0,len(b),32))+'\n')
 manifest.append(dict(file=name+'.hex',bytes=len(b),sha256=hashlib.sha256(b).hexdigest(),source='interop/smsfw-general-fixtures.py',provenance='Actual template source projection with explicit inputs; synthetic answers/negative cases, NOT executed C++ or OCS capture'))
Path('testdata/codec/manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
files=['yppdia_avptemplate.cf','dcca-wrapper/AvpsTemplate.cc','dcca-wrapper/AvpImpl.cc','dia-base/dia-message.cc','dia-base/dia-message.hpp','dia-cc/dia-cc.cc','dcca-wrapper/DccaClientWrapper.cc','../../../appsrc/kernel/yppdia/yppdiaengmx.cc']
Path('interop/evidence/2026-11-07-smsfw-general-source.json').write_text(json.dumps(dict(kind='source projection only; ordinary RFC Address mode',inputs=inputs,template_mapping=mapping,source_files=[dict(path=str((root/f).resolve().relative_to(Path(sys.argv[1]).resolve())),sha256=hashlib.sha256((root/f).read_bytes()).hexdigest()) for f in files],fixtures=[x for x in manifest if x['file'].startswith('b44-')]),indent=2)+'\n')
print(len(frames),'frames;',len(mapping),'template nodes;',len(manifest),'total fixtures')
