#!/usr/bin/env python3
"""Offline source projection, NOT an execution of C++ or an OCS capture.
Run from repo root: python3 interop/smsfw-iec-fixtures.py ../sms-fw
Uses only Python stdlib and the actual legacy template, never diameterkit codec.
"""
from pathlib import Path
import hashlib, json, struct, sys
root = Path(sys.argv[1]) / 'c-cpp/libsrc/kernel/dialib'
template = root / 'yppdia_iec_avptemplate.cf'
inputs = dict(SUBSCRIPTIONID_MSISDN='15551234567', SUBSCRIPTIONID_IMSI='001010123456789',
 TIMESTAMP=1700000000, SMSC_ADDR='15550000001', VMSC_ORIG='15550000002', VMSC_RECP='15550000003',
 DCS=8, INTERFACE_TYPE_DEST=2, PID='00ff', REPLY_PATH=0, UDH='0500037f020100ff',
 ORIG_MSISDN='15551234567', DEST_MSISDN='15557654321', PRIORITY=0, UDL=140,
 SMSKEY='sms-0001', SRR_ON=1)
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
  else:data=struct.pack('!I',int(val)&0xffffffff)
 return avp(c,data,v)
custom=[]
while index<len(rows):custom.append(field([]))
session=avp(263,b'ctf.example;legacy-iec;1')
def frame(body,flags=0xc0):return bytes([1])+(20+len(body)).to_bytes(3,'big')+bytes([flags])+(272).to_bytes(3,'big')+struct.pack('!III',4,0x31323334,0x41424344)+body
base=session+avp(264,b'ctf.example')+avp(296,b'example.net')+avp(283,b'example.net')+u(258,4)+avp(461,b'legacy-sms@example.net')+u(416,4)+u(415,0)+u(436,0)
frames={'b41-iec-request':frame(base+b''.join(custom)), 'b41-iec-enhanced':frame(base+avp(293,b'ocs.example')+u(278,0)+b''.join(custom))}
def answer(code,extra=b'',error=False):
 body=session+u(268,code)+avp(264,b'ocs.example')+avp(296,b'example.net')
 if not error:body+=u(258,4)+u(416,4)+u(415,0)
 return frame(body+extra,0x60 if error else 0x40)
frames.update({'b41-cca-success':answer(2001),'b41-cca-denied':answer(4012),'b41-cca-failed':answer(5012),'b41-cca-error':answer(3004,error=True),
 'b41-cca-grant-ddfh':answer(2001,avp(431,avp(417,struct.pack('!Q',2**64-1)))+u(428,1)),
 'b41-cca-ddfh-zero':answer(4012,u(428,0))})
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
r=frames['b41-iec-request'];a=frames['b41-cca-success'];si=(10415,873);sms=(10415,2000);mms=(10415,877)
def project(name,path,fn,raw=r):frames['b41-'+name]=change(raw,[(0,x) if isinstance(x,int) else x for x in path],fn)
def replace(name,path,d,raw=r):project(name,path,lambda k,_,old:avp(k[1],d,k[0],old[4]),raw)
replace('signed-min',[si,sms,(10415,2001)],struct.pack('!i',-2147483648))
replace('signed-max',[si,mms,(10415,1209)],struct.pack('!i',2147483647))
replace('signed-width',[si,sms,(10415,2001)],bytes(3))
replace('string-nul',[si,mms,(10415,1210)],b'a\0b')
replace('string-nonutf8',[si,mms,(10415,1210)],b'\xff')
replace('string-empty',[si,mms,(10415,1210)],b'')
replace('wrong-ue',[458,460],b'4901542032375101')
replace('wrong-time',[si,mms,(10415,1202)],bytes(4))
replace('wrong-smsc',[si,sms,(10415,2018)],b'other')
replace('wrong-node',[si,sms,(10415,2016)],struct.pack('!I',1))
replace('wrong-action',[436],struct.pack('!I',1))
replace('wrong-units',[456,437,417],struct.pack('!Q',2))
project('missing-udh',[si,sms,(10415,2015)],lambda k,d,a:b'')
project('duplicate-interface',[si,sms,(10415,2002)],lambda k,d,a:a+a)
project('wrong-vendor',[si,sms,(10415,2001)],lambda k,d,a:avp(k[1],d,0))
project('wrong-parent',[si,sms],lambda k,d,a:avp(2000,b''.join(x for key,v,x in entries(d) if key!=(10415,2010))+avp(2026,avp(2010,b'15550000003',10415),10415),10415))
for name,data in [('scalar',bytes(4)),('grouped',avp(445,u(429,0)))]:
 project('420-'+name,[456,437],lambda k,d,a:avp(437,d+avp(420,data,f=0)))
project('token1292',[415],lambda k,d,a:a+avp(1292,b'\0\xff',f=0))
project('cca-psp',[415],lambda k,d,a:a+avp(873,avp(20400,avp(20430,bytes(4),10415),10415),10415,128),raw=a)
replace('cca-number',[415],struct.pack('!I',1),a)
replace('cca-ddfh-unknown',[428],struct.pack('!I',2),frames['b41-cca-grant-ddfh'])
# Swap the two whole subscription parents: same leaves, different instance meaning.
project('subscriptions-reversed',[443,450],lambda k,d,a:u(450,1-int.from_bytes(d,'big')))
manifest=json.load(open('testdata/codec/manifest.json'));manifest=[x for x in manifest if not x['file'].startswith('b41-')]
for name,b in frames.items():
 Path('testdata/codec',name+'.hex').write_text('\n'.join(b[i:i+32].hex(' ') for i in range(0,len(b),32))+'\n')
 manifest.append(dict(file=name+'.hex',bytes=len(b),sha256=hashlib.sha256(b).hexdigest(),source='interop/smsfw-iec-fixtures.py',provenance='Independent Python projection of actual C++ template and supplied input; synthetic paired CCA/counterexamples; not executed C++ or OCS capture'))
Path('testdata/codec/manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
files=['yppdia_iec_avptemplate.cf','dcca-wrapper/AvpsTemplate.cc','dcca-wrapper/AvpImpl.cc','dia-base/dia-message.cc','dia-cc/dia-cc.cc','dcca-wrapper/DccaClientWrapper.cc']
Path('interop/evidence/2026-11-07-smsfw-iec-source.json').write_text(json.dumps(dict(kind='source projection only',inputs=inputs,template_mapping=mapping,source_files=[dict(path='c-cpp/libsrc/kernel/dialib/'+f,sha256=hashlib.sha256((root/f).read_bytes()).hexdigest()) for f in files],fixtures=[x for x in manifest if x['file'].startswith('b41-')]),indent=2)+'\n')
print(len(frames),'frames;',len(mapping),'template nodes;',len(manifest),'total fixtures')
