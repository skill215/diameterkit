#!/usr/bin/env python3
"""Independent CCA assembly from frozen envelope seeds; never invokes Go codec.
Fields are limited to source-consumed standard keys; private Vendor is not guessed.
"""
from pathlib import Path
import json,hashlib,struct,copy
root=Path(__file__).resolve().parents[1];dst=root/'testdata/codec'
def load(n):return bytes.fromhex((dst/(n+'.hex')).read_text())
def avp(c,d,v=0,flags=None):
 if flags is None:flags=192 if v else 64
 n=(12 if flags&128 else 8)+len(d)
 return c.to_bytes(4,'big')+bytes([flags])+n.to_bytes(3,'big')+(v.to_bytes(4,'big') if flags&128 else b'')+d+b'\0'*((-n)%4)
def u(c,n,v=0):return avp(c,n.to_bytes(4,'big'),v)
def fields(b):
 off=20
 while off<len(b):
  n=int.from_bytes(b[off+5:off+8],'big');end=off+((n+3)&~3);yield int.from_bytes(b[off:off+4],'big'),b[off:end];off=end
 assert off==len(b)
def frame(seed,top=2001,extra=b'',drop=()):
 body=b''.join(u(268,top) if c==268 else raw for c,raw in fields(seed) if c not in drop)+extra
 return seed[:1]+(20+len(body)).to_bytes(3,'big')+seed[4:20]+body
out=[]
def save(name,raw,request,expect='pair'):
 name='b48-'+name+'.hex';(dst/name).write_text(raw.hex()+'\n');out.append({'file':name,'bytes':len(raw),'sha256':hashlib.sha256(raw).hexdigest(),'source':'DIACCParseAnswer source audit; RFC4006/8506; '+request,'provenance':'independent synthetic answer to frozen source request; no OCS capture or C++ execution','request':request+'.hex','expect':expect})
for kind in ['iec','initial','term','general','balance','refund']:
 seed=load('b47-'+kind+'-success');req='b47-'+kind+'-request'
 for code in [4010,4011,5030]:save(kind+'-'+str(code),frame(seed,code,drop=[422]),req)
 save(kind+'-child-success',frame(seed,extra=avp(456,u(268,2001))),req)
 save(kind+'-child-denied',frame(seed,extra=avp(456,u(268,4012)),drop=[422]),req)
 save(kind+'-top-failure',frame(seed,5012,extra=avp(456,u(268,2001)),drop=[422]),req)
# Signed64 digits / optional signed32 exponent / unsigned32 currency.
def cost(digits=-123,exp=None,currency=840):
 value=avp(447,struct.pack('!q',digits))+(b'' if exp is None else avp(429,struct.pack('!i',exp)))
 return avp(423,avp(445,value)+u(425,currency))
for kind in ['iec','initial','term']:
 save(kind+'-cost',frame(load('b47-'+kind+'-success'),extra=cost()),'b47-'+kind+'-request')
seed=load('b47-iec-success');req='b47-iec-request'
save('cost-zero',frame(seed,extra=cost(0,0,0)),req)
save('cost-limits',frame(seed,extra=cost(-(1<<63),-(1<<31),(1<<32)-1)),req)
for value in [0,1]:save('general-ddfh-'+str(value),frame(load('b47-general-success'),extra=u(428,value)),'b47-general-request')
# Width/cardinality/vendor/parent and unresolved private-boundary negatives.
neg={
 'child-empty':avp(456,b''),
 'child-width':avp(456,avp(268,b'\0\0\1')),
 'child-duplicate-result':avp(456,u(268,2001)+u(268,4012)),
 'duplicate-mscc':avp(456,u(268,2001))+avp(456,u(268,4012)),
 'child-vendor':avp(456,u(268,2001,10415)),
 'mscc-vendor':avp(456,u(268,2001),10415),
 'child-unknown':avp(456,u(268,4999)),
 'child-wrong-sequence':avp(456,u(268,5002)),
 'child-service-key':avp(456,u(268,2001)+u(439,1)),
 'child-grant':avp(456,u(268,2001)+avp(431,avp(417,struct.pack('!Q',1)))),
 'child-token':avp(456,u(268,2001)+avp(1292,b'\0\xff')),
 'private-token':avp(1292,b'\0\xff'),
 'private-psp':avp(873,avp(20400,avp(20430,struct.pack('!i',0),10415),10415),10415),
 'cost-wrong-parent':avp(456,u(268,2001)+cost()),
 'cost-child-failure':cost()+avp(456,u(268,4012)),
 'ddfh-unknown':u(428,2),
}
for n,extra in neg.items():save(n,frame(seed,extra=extra),req,'reject')
for code in [2002,5031,5011]:save('unselected-'+str(code),frame(seed,code),req,'reject')
save('experimental-5011',frame(seed,extra=avp(297,u(266,10415)+u(298,5011))),req,'reject')
save('missing-top',frame(seed,extra=avp(456,u(268,2001)),drop=[268]),req,'reject')
save('cost-top-failure',frame(seed,4012,extra=cost()),req,'reject')
# Legal standalone union types with invalid original action; CompleteAnswer must reject.
for kind in ['balance','refund']:
 save(kind+'-ddfh',frame(load('b47-'+kind+'-success'),extra=u(428,1)),'b47-'+kind+'-request','relation-reject')
manifest=json.load(open(dst/'manifest.json'));manifest=[x for x in manifest if not x['file'].startswith('b48-')]+out
(dst/'manifest.json').write_text(json.dumps(manifest,ensure_ascii=False,indent=2)+'\n')
source_paths=['c-cpp/libsrc/kernel/dialib/dia-cc/dia-cc.cc','c-cpp/libsrc/kernel/dialib/dia-cc/dia-cc.hpp','c-cpp/libsrc/kernel/dialib/dia-base/dia-attr.hpp','c-cpp/libsrc/kernel/dialib/dcca-wrapper/DccaClientWrapper.cc','c-cpp/libsrc/kernel/dialib/dcca-wrapper/DccaClientWrapperConfig.cc','c-cpp/libsrc/kernel/dialib/yppdia_dcca.cf','c-cpp/libsrc/kernel/dialib/yppdia_refund_avptemplate.cf']
proof={'source_files':[{'path':f,'sha256':hashlib.sha256((root.parent/'sms-fw'/f).read_bytes()).hexdigest()} for f in source_paths], 'fixtures':out,
 'consumption':{'command_result':'top268 required; old2001/2002 could be overwritten by child268; new values stay separate','service_result':'one fixedVendor0/456/268 result-only group; no implied service key, child grant or token','cost':'top423/445/{447i64,429i32 optional} and425u32; no numeric scaling','ddfh':'top428, preserve Absent/0/1; oldwrapper increment not copied; original direct-debit only','token':'configured code1292 sample, top or MSCC; reader ignoresVendor, so actualVendor and path policy unknown','psp':'873/20400/20430 candidate; noVendor evidence; constructor defaultsoperation_result1 and parser first iteration useswrongparent; no default copied'},
 'limits':['Synthetic answers, no C++ execution or liveOCS','5031 diagnostic subflow,2002 legacy interpretation,5011 namespaces/features remain unimplemented; neither 5011 base nor experimental is silently inferred','Multiple service instances and target keys/grants/tokens remain separate named work']}
(root/'interop/evidence/2026-11-07-smsfw-answer-source.json').write_text(json.dumps(proof,ensure_ascii=False,indent=2)+'\n')
print(len(out),'new frames',len(manifest),'total')
