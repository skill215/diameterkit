#!/usr/bin/env python3
"""Independent wire projection of frozen source-template inputs, not a peer capture.
Only ordering is normalized; key/flags/payload/parent/repeated-instance are preserved.
This script does not import diameterkit or read its generated definition.
"""
from pathlib import Path
import copy,hashlib,json,struct
root=Path(__file__).resolve().parents[1];dst=root/'testdata/codec'
# Explicitly audited grouped identities; 2018 is ALWAYS a leaf, never guessed.
groups={(0,x) for x in [443,456,437,446,413,420,445,458,431,423,284,279]}|{(10415,x) for x in [873,2000,877,2002,2009,886,1201]}
def parse(b):
 out=[];off=0
 while off<len(b):
  code=int.from_bytes(b[off:off+4],'big');flags=b[off+4];n=int.from_bytes(b[off+5:off+8],'big');h=12 if flags&128 else 8
  vendor=int.from_bytes(b[off+8:off+12],'big') if h==12 else 0;data=b[off+h:off+n]
  out.append([vendor,code,flags,parse(data) if (vendor,code) in groups else data]);off+=(n+3)&~3
 assert off==len(b);return out

def encode(xs):
 out=b''
 for v,c,f,d in xs:
  data=encode(d) if isinstance(d,list) else d;n=8+(4 if f&128 else 0)+len(data)
  out+=struct.pack('!I',c)+bytes([f])+n.to_bytes(3,'big')+(struct.pack('!I',v) if f&128 else b'')+data+b'\0'*((-n)%4)
 return out

def read(name):
 b=bytes.fromhex((dst/(name+'.hex')).read_text());return b[:20],parse(b[20:])

def frame(h,xs):
 b=encode(xs);return h[:1]+(20+len(b)).to_bytes(3,'big')+h[4:]+b

def get(xs,path):
 for key in path:
  xs=next(x[3] for x in xs if (x[0],x[1])==key)
 return xs

def item(xs,path):return next(x for x in (get(xs,path[:-1]) if len(path)>1 else xs) if (x[0],x[1])==path[-1])
# Closed fixed sender field order. Contradictory legacy positions are recorded.
orders={
 (): [263,264,296,283,258,461,416,415,436,293,278,443,55,455,456,458,873,284,282,439,23,1292],
 (873,2000):[2016,2018,2008,2010,2017,2001,2002,2007,2009,2013,2011,2015,2014,2012],
 (873,877):[886,1201,1202,1209,1212,1210,1216],
}
def reorder(xs,path=()):
 for x in xs:
  if isinstance(x[3],list):reorder(x[3],path+(x[1],))
 if path in orders:xs.sort(key=lambda x:orders[path].index(x[1]))
 return xs
sources={'iec':'b41-iec-request','initial':'b43-init-request','term':'b43-term-request','refund':'b42-refund-request','general':'b44-general-request','balance':'b44-balance-request'}
outputs=[];inputs=[]
def save(name,h,xs,source):
 raw=frame(h,xs);name='b47-'+name+'.hex';(dst/name).write_text(raw.hex()+'\n')
 outputs.append({'file':name,'bytes':len(raw),'sha256':hashlib.sha256(raw).hexdigest(),'source':source,'provenance':'independent source-frame ordering projection or explicit cross-template mutation; not C++ execution or OCS capture'})
for kind,source in sources.items():
 h,xs=read(source);inputs.append({'file':source+'.hex','sha256':hashlib.sha256(frame(h,xs)).hexdigest()});save(kind+'-request',h,reorder(xs),source)
 # Existing independently assembled normal success, refusal and E pairs.
 prefix={'iec':'b41','initial':'b43-init','term':'b43-term','refund':'b42','general':'b44','balance':'b44'}[kind]
 for result in ['success','denied','error']:
  a=prefix+'-cca-'+result
  if kind in ['initial','term']:a=prefix+'-a-'+({'success':'2001','error':'3004','denied':'5002' if kind=='term' else '4012'}[result])
  if kind=='balance' and result=='success':a='b44-balance-enough'
  ah,ax=read(a);save(kind+'-'+result,ah,ax,a)
# Mutations specifically attack the newly combined boundary, with valid widths.
S=[(10415,873),(10415,2000)];M=[(10415,873),(10415,877)];R=[(0,456),(0,437)]
cases={
 'iec-address2018':('iec',lambda x:item(x,S+[(10415,2018)]).__setitem__(3,b'\x00\x01\x7f\0\0\1')),
 'general-text2018':('general',lambda x:item(x,S+[(10415,2018)]).__setitem__(3,b'127.0.0.1')),
 'general-draft2018':('general',lambda x:item(x,S+[(10415,2018)]).__setitem__(3,b'\x7f\0\0\1')),
 'general-smsc':('general',lambda x:get(x,S).append([10415,2017,192,b'15550000001'])),
 'iec-money413':('iec',lambda x:get(x,R).append(copy.deepcopy(item(read(sources['general'])[1],R+[(0,413)])))),
 'refund-unit-only':('refund',lambda x:get(x,R).__setitem__(slice(None),[a for a in get(x,R) if a[1]!=420])),
 'refund-client':('refund',lambda x:get(x,S).append([10415,2018,192,b'\0\1\x7f\0\0\1'])),
 'initial-service':('initial',lambda x:x.append([0,439,64,struct.pack('!i',1)])),
 'term-money':('term',lambda x:get(x,[(0,456)]).append(copy.deepcopy(item(read(sources['refund'])[1],[(0,456),(0,437)])))),
 'general-ue':('general',lambda x:x.append(copy.deepcopy(item(read(sources['iec'])[1],[(0,458)])))),
 'general-mms-key':('general',lambda x:get(x,M).append([10415,1210,192,b'sms-0001'])),
 'iec-unknown-optional':('iec',lambda x:get(x,S).append([10415,2999,128,b'opaque'])),
}
for name,(kind,fn) in cases.items():
 h,x=read(sources[kind]);fn(x);save(name,h,x,sources[kind]+'; explicit negative '+name)
manifest=json.load(open(dst/'manifest.json'));manifest=[x for x in manifest if not x['file'].startswith('b47-')]+outputs
(dst/'manifest.json').write_text(json.dumps(manifest,ensure_ascii=False,indent=2)+'\n')
proof={'inputs':inputs,'outputs':outputs,'limitations':['Six command uses from five source templates; balance reuses general. Source order normalized by explicit fixed field order only, no leaf conversions.','Normal/E answers synthetic, no C++ execution or OCS acceptance; receiver interpretation of 2018 and ordering remains deployment evidence.']}
(root/'interop/evidence/2026-11-07-smsfw-combined-source.json').write_text(json.dumps(proof,indent=2)+'\n')
print(len(outputs),'new frames;',len(manifest),'total')
