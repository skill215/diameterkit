#!/usr/bin/env python3
"""Offline source projection, NOT an execution of C++ or an OCS capture.
Run from repo root: python3 interop/smsfw-scap1-fixtures.py ../sms-fw
Uses only Python stdlib and the actual legacy template, never diameterkit codec.
"""
from pathlib import Path
import hashlib, json, struct, sys
root = Path(sys.argv[1]) / 'c-cpp/libsrc/kernel/dialib'
template = root / 'yppdia_avptemplate_scap1.cf'
inputs = dict(SUBSCRIPTIONID_MSISDN='15551234567', TIMESTAMP=1700000000, DCS='8', ORIG_MSISDN='15551234567', DEST_MSISDN='15557654321', VMSC_ORIG='15550000001', TRAFFIC_CASE='MO', CM_SEQ_MSGS='1', UDL='140')
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
  elif typ in ('UINT64','UNSIGNED64'):data=struct.pack('!Q',int(val))
  elif typ=='INTEGER64':data=struct.pack('!q',int(val))
  else:data=struct.pack('!I',int(val)&0xffffffff)
 return avp(c,data,v)
custom=[]
while index<len(rows):custom.append(field([]))

session=avp(263,b'ctf.example;legacy-scap1;1')
def frame(body,flags=0xc0):return bytes([1])+(20+len(body)).to_bytes(3,'big')+bytes([flags])+(271).to_bytes(3,'big')+struct.pack('!III',0,0x51525354,0x61626364)+body
base=session+avp(264,b'ctf.example')+avp(296,b'example.net')+avp(283,b'example.net')
frames={'b45-acr':frame(base+b''.join(custom)),'b45-acr-envelope':frame(base+b''.join(custom)+avp(293,b'ocs.example')+u(278,0))}
def answer(code,extra=b'',error=False):
 body=session+u(268,code)+avp(264,b'ocs.example')+avp(296,b'example.net')
 if not error:body+=u(480,1)+u(485,0)
 return frame(body+extra,0x60 if error else 0x40)
frames.update({'b45-aca-success':answer(2001),'b45-aca-failed':answer(5012),'b45-aca-diagnostic':answer(5005,avp(279,avp(485,bytes(4)))),'b45-aca-error':answer(3004,error=True),'b45-aca-application':answer(2001,custom[0])})
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
r=frames['b45-acr'];a=frames['b45-aca-success']
def project(name,path,fn,raw=r):frames['b45-'+name]=change(raw,[(0,x) if isinstance(x,int) else x for x in path],fn)
def replace(name,path,d,raw=r):project(name,path,lambda k,_,old:avp(k[1],d,k[0],old[4]),raw)
sub=(193,553);rsu=(193,606);uv=(193,612);param=(193,607)
replace('time-zero',[(193,610)],bytes(4))
replace('time-max',[(193,610)],bytes([255])*4)
replace('subscription-empty',[sub,(193,554)],b'')
replace('subscription-raw',[sub,(193,554)],b'\xffraw')
replace('subscription-nul',[sub,(193,554)],b'a\0b')
for name,path in [('vsai',[260]),('acct',[260,259]),('subscription',[sub]),('subscriber',[sub,(193,554)]),('unit-type',[rsu,(193,611)]),('unit-value',[rsu,uv]),('digits',[rsu,uv,(193,617)]),('number',[485]),('record-type',[480]),('action',[(193,615)]),('timestamp',[(193,610)])]:
 project('missing-'+name,path,lambda k,d,a:b'')
replace('bad-digits-width',[rsu,uv,(193,617)],bytes(4))
replace('bad-number-width',[485],bytes(8))
replace('units-zero',[rsu,uv,(193,617)],bytes(8))
replace('units-max',[rsu,uv,(193,617)],bytes([255])*8)
replace('wrong-unit-type',[rsu,(193,611)],struct.pack('!I',3))
replace('wrong-record-type',[480],struct.pack('!I',2))
replace('wrong-number',[485],struct.pack('!I',1))
replace('wrong-action',[(193,615)],struct.pack('!I',1))
replace('wrong-vendor-value',[260,266],struct.pack('!I',10415))
replace('wrong-acct',[260,259],struct.pack('!I',4))
project('auth-in-vsai',[260,259],lambda k,d,a:avp(258,d))
project('wrong-unit-vendor',[rsu,uv,(193,617)],lambda k,d,a:avp(617,d,10415))
project('standard-action',[(193,615)],lambda k,d,a:avp(436,d))
project('standard-time',[(193,610)],lambda k,d,a:avp(55,d))
project('unknown-group-child',[rsu],lambda k,d,a:avp(606,d+avp(9999,b'x',f=0),193))
project('unexpected-cc-type',[485],lambda k,d,a:a+u(416,4))
# Preserve each607 parent instance, never turn parameter types into a map.
def parameters(name,fn):
 parts=list(entries(r[20:]));selected=[x[2] for x in parts if x[0]==param]
 other=b''.join(x[2] for x in parts if x[0]!=param)
 frames['b45-'+name]=frame(other+b''.join(fn(selected)))
def pv(a,value):
 children=list(entries(next(entries(a))[1]));return avp(607,children[0][2]+avp(609,value,193),193)
parameters('param-empty',lambda xs:xs[:1]+[pv(xs[1],b'')]+xs[2:])
parameters('param-text',lambda xs:xs[:1]+[pv(xs[1],b'008')]+xs[2:])
parameters('param-raw',lambda xs:xs[:4]+[pv(xs[4],b'\xff')]+xs[5:])
parameters('param-nul',lambda xs:xs[:4]+[pv(xs[4],b'a\0b')]+xs[5:])
parameters('missing-param',lambda xs:xs[:-1])
parameters('extra-param',lambda xs:xs+[xs[0]])
parameters('swapped-params',lambda xs:[xs[1],xs[0]]+xs[2:])
parameters('duplicate-param-type',lambda xs:xs[:2]+[xs[1]]+xs[3:])
parameters('origin-copy-mismatch',lambda xs:xs[:5]+[pv(xs[5],b'other')]+xs[6:])
parameters('wrong-literal',lambda xs:[pv(xs[0],b'2')]+xs[1:])
parameters('wrong-timezone',lambda xs:xs[:-1]+[pv(xs[-1],b'+0000')])
project('missing-param-value',[param,(193,609)],lambda k,d,a:b'')
project('wrong-param-parent',[param],lambda k,d,a:d)
replace('aca-number',[485],struct.pack('!I',1),a)
replace('aca-record-type',[480],struct.pack('!I',2),a)
replace('aca-number-negative',[485],struct.pack('!i',-1),a)
project('aca-missing-number',[485],lambda k,d,a:b'',a)
project('aca-ro-field',[485],lambda k,d,a:a+u(415,0),a)
frames['b45-aca-unselected-result']=answer(4012)
for name,app,code in [('header-ro',4,272),('header-acct',3,271),('header-private',19302,271)]:
 b=bytearray(r);b[5:8]=code.to_bytes(3,'big');b[8:12]=struct.pack('!I',app);frames['b45-'+name]=bytes(b)
manifest=json.load(open('testdata/codec/manifest.json'));manifest=[x for x in manifest if not x['file'].startswith('b45-')]
for name,b in frames.items():
 Path('testdata/codec',name+'.hex').write_text('\n'.join(b[i:i+32].hex(' ') for i in range(0,len(b),32))+'\n')
 manifest.append(dict(file=name+'.hex',bytes=len(b),sha256=hashlib.sha256(b).hexdigest(),source='interop/smsfw-scap1-fixtures.py',provenance='Actual SCAP1 template projection with explicit Getter outputs; synthetic ACA from selected RFC6733 envelope, not C++ execution or target capture'))
Path('testdata/codec/manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
files=['yppdia_avptemplate_scap1.cf','dcca-wrapper/AvpsTemplate.cc','dcca-wrapper/AvpImpl.cc','dia-base/dia-message.cc','dia-base/dia-attr.hpp','dia-cc/dia-cc.cc','dcca-wrapper/DccaClientWrapper.cc','../../../appsrc/kernel/yppdia/yppdiaengmx.cc']
Path('interop/evidence/2026-11-07-smsfw-scap1-source.json').write_text(json.dumps(dict(kind='source projection only; ACA synthetic base grammar, target-private ACA unknown',inputs=inputs,template_mapping=mapping,source_files=[dict(path=str((root/f).resolve().relative_to(Path(sys.argv[1]).resolve())),sha256=hashlib.sha256((root/f).read_bytes()).hexdigest()) for f in files],fixtures=[x for x in manifest if x['file'].startswith('b45-')]),indent=2)+'\n')
assert len(custom)==18 and len(mapping)==47
print(len(frames),'frames;',len(mapping),'template nodes;',len(manifest),'total fixtures')
