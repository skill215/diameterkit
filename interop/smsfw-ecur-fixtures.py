#!/usr/bin/env python3
"""Offline source projection, NOT an execution of C++ or an OCS capture.
Run from repo root: python3 interop/smsfw-ecur-fixtures.py ../sms-fw
Uses only Python stdlib and the actual legacy template, never diameterkit codec.
"""
from pathlib import Path
import hashlib, json, struct, sys
root = Path(sys.argv[1]) / 'c-cpp/libsrc/kernel/dialib'
inputs = dict(SUBSCRIPTIONID_MSISDN='15551234567', SUBSCRIPTIONID_IMSI='001010123456789',
 TIMESTAMP=1700000000, SMSC_ADDR='15550000001', VMSC_ORIG='15550000002', VMSC_RECP='15550000003',
 DCS=8, INTERFACE_TYPE_DEST=2, PID='00ff', REPLY_PATH=0, UDH='0500037f020100ff',
 ORIG_MSISDN='15551234567', DEST_MSISDN='15557654321', PRIORITY=0, UDL=140,
 SMSKEY='sms-0001', SRR_ON=1, CC_SSU=1, SM_STATUS='00ff', CTIMESTAMP=1700000030)
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
  names=[]
  for _ in args:
   names.append(rows[index][0]);children.append(field(path))
  assert sorted(names)==sorted(args),(names,args)
  data=b''.join(children)
 else:
  val=inputs[args[0][1:]] if args[0].startswith('$') else args[0]
  if typ=='STRING':data=str(val).encode()
  elif typ=='OCTETSTRING':data=bytes.fromhex(val)
  elif typ=='TIME':data=struct.pack('!I',(int(val)+2208988800)&0xffffffff)
  elif typ=='UINT64':data=struct.pack('!Q',int(val))
  else:data=struct.pack('!I',int(val)&0xffffffff)
 return avp(c,data,v)
frames={};mappings={}
for kind,req_type,num,h,e in [('init',1,0,0x31323334,0x41424344),('term',3,1,0xb1b2b3b4,0xc1c2c3c4)]:
 template=root/('yppdia_'+kind+'_avptemplate.cf');rows=[]
 for line in template.read_text().splitlines():
  line=line.split('#',1)[0].strip()
  if not line:continue
  name,rest=line.split('=',1);parts=[x.strip() for x in rest.split(',')]
  rows.append((name.strip(),int(parts[0]),int(parts[1]),parts[2].upper(),parts[3:]))
 index=0;mapping=[];custom=[]
 while index<len(rows):custom.append(field([]))
 mappings[kind]=mapping
 session=avp(263,b'ctf.example;legacy-ecur;1')
 def frame(body,flags=0xc0):return bytes([1])+(20+len(body)).to_bytes(3,'big')+bytes([flags])+(272).to_bytes(3,'big')+struct.pack('!III',4,h,e)+body
 base=session+avp(264,b'ctf.example')+avp(296,b'example.net')+avp(283,b'example.net')+u(258,4)+avp(461,b'legacy-sms@example.net')+u(416,req_type)+u(415,num)
 frames['b43-'+kind+'-request']=frame(base+b''.join(custom))
 frames['b43-'+kind+'-enhanced']=frame(base+avp(293,b'ocs.example')+u(278,0)+b''.join(custom))
 for code in [2001,4012 if kind=='init' else 5002,5012,3004]:
  body=session+u(268,code)+avp(264,b'ocs.example')+avp(296,b'example.net')
  if code!=3004:body+=u(258,4)+u(416,req_type)+u(415,num)
  frames['b43-'+kind+'-a-'+str(code)]=frame(body,0x60 if code==3004 else 0x40)
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
i=frames['b43-init-request'];t=frames['b43-term-request'];ia=frames['b43-init-a-2001'];ta=frames['b43-term-a-2001'];si=(10415,873);sms=(10415,2000);mms=(10415,877)
def project(name,raw,path,fn):frames['b43-'+name]=change(raw,[(0,x) if isinstance(x,int) else x for x in path],fn)
def replace(name,raw,path,data):project(name,raw,path,lambda k,d,a:avp(k[1],data,k[0],a[4]))
for label,b in [('init',i),('term',t)]:
 project(label+'-action',b,[415],lambda k,d,a:a+u(436,0))
 replace(label+'-wrong-sequence',b,[415],u(0,2)[8:12])
 replace(label+'-wrong-destination-interface',b,[si,sms,(10415,2002),(10415,2006)],bytes(4))
 replace(label+'-wrong-origin-interface',b,[si,sms,(10415,2009),(10415,2006)],u(0,1)[8:12])
 project(label+'-origin-sccp',b,[si,sms],lambda k,d,a:avp(2000,d+avp(2008,b'15550000002',10415),10415))
 replace(label+'-wrong-imsi-copy',b,[458,460],b'other')
 replace(label+'-wrong-time-copy',b,[si,mms,(10415,1202)],bytes(4))
for suffix,number in [('zero',0),('max',2**64-1)]:replace('term-usage-'+suffix,t,[456,446,417],struct.pack('!Q',number))
replace('term-width4',t,[456,446,417],bytes(4))
project('term-no-usu',t,[456,446],lambda k,d,a:b'')
project('term-empty-usu',t,[456,446,417],lambda k,d,a:b'')
project('term-no-final',t,[456,(10415,872)],lambda k,d,a:b'')
replace('term-wrong-final',t,[456,(10415,872)],bytes(4))
project('term-final-parent',t,[456],lambda k,d,a:avp(456,b''.join(x for key,dd,x in entries(d) if key!=(10415,872)))+u(872,2,10415))
project('term-rsu',t,[456],lambda k,d,a:avp(456,d+avp(437,avp(417,struct.pack('!Q',1)))))
project('init-usu',i,[456],lambda k,d,a:avp(456,d+avp(446,avp(417,struct.pack('!Q',1)))))
project('init-no-rsu',i,[456,437],lambda k,d,a:b'')
project('init-status',i,[si,sms],lambda k,d,a:avp(2000,d+avp(2014,b'\0',10415),10415))
for name,code in [('status',2014),('discharge',2012)]:project('term-no-'+name,t,[si,sms,(10415,code)],lambda k,d,a:b'')
replace('term-empty-status',t,[si,sms,(10415,2014)],b'')
replace('term-zero-discharge',t,[si,sms,(10415,2012)],bytes(4))
project('term-duplicate-status',t,[si,sms,(10415,2014)],lambda k,d,a:a+a)
project('term-token',t,[415],lambda k,d,a:a+avp(1292,b'\0',f=0))
project('init-420',i,[456,437],lambda k,d,a:avp(437,d+avp(420,avp(445,b''),f=0)))
project('init-a-gsu',ia,[415],lambda k,d,a:a+avp(431,avp(417,struct.pack('!Q',1))))
project('term-a-gsu',ta,[415],lambda k,d,a:a+avp(431,avp(417,struct.pack('!Q',1))))
replace('term-a-wrong-sequence',ta,[415],bytes(4))
# Independent sender-order projection: preserve originals; only swap the two
# distinct MMS leaves whose old row order differs between INITIAL and TERM.
project('term-sender-order',t,[si,mms],lambda k,d,a:avp(877,b''.join(raw for key,dd,raw in sorted(entries(d),key=lambda x:[886,1201,1202,1209,1212,1210,1216].index(x[0][1]))),10415))
manifest=json.load(open('testdata/codec/manifest.json'));manifest=[x for x in manifest if not x['file'].startswith('b43-')]
for name,b in frames.items():
 Path('testdata/codec',name+'.hex').write_text('\n'.join(b[i:i+32].hex(' ') for i in range(0,len(b),32))+'\n')
 manifest.append(dict(file=name+'.hex',bytes=len(b),sha256=hashlib.sha256(b).hexdigest(),source='interop/smsfw-ecur-fixtures.py',provenance='Independent Python projection of actual C++ template and supplied input; synthetic paired CCA/counterexamples; not executed C++ or OCS capture'))
Path('testdata/codec/manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
files=['yppdia_init_avptemplate.cf','yppdia_term_avptemplate.cf','dcca-wrapper/AvpsTemplate.cc','dcca-wrapper/AvpImpl.cc','dia-base/dia-message.cc','dia-cc/dia-cc.cc','dcca-wrapper/DccaClientWrapper.cc']
Path('interop/evidence/2026-11-07-smsfw-ecur-source.json').write_text(json.dumps(dict(kind='source projection only',inputs=inputs,template_mapping=mappings,source_files=[dict(path='c-cpp/libsrc/kernel/dialib/'+f,sha256=hashlib.sha256((root/f).read_bytes()).hexdigest()) for f in files],fixtures=[x for x in manifest if x['file'].startswith('b43-')]),indent=2)+'\n')
print(len(frames),'frames;',{k:len(v) for k,v in mappings.items()},'template nodes;',len(manifest),'total fixtures')
