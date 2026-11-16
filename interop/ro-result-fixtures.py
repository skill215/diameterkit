#!/usr/bin/env python3
"""Synthetic result/diagnostic cases assembled independently of Go definitions."""
from pathlib import Path
import hashlib,json
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
def frame(seed,top=2001,extra=b'',drop=(),error=False):
 body=b''.join(u(268,top) if c==268 else raw for c,raw in fields(seed) if c not in drop and (not error or c in [263,264,296,268,284]))+extra
 flags=seed[4]|32 if error else seed[4]
 return seed[:1]+(20+len(body)).to_bytes(3,'big')+bytes([flags])+seed[5:20]+body
def exp(v=10415,c=5011):return avp(297,u(266,v)+u(298,c))
def sf(v=10415,ident=1,bits=0x80000000,m=False):return avp(628,u(266,v)+u(629,ident,10415)+u(630,bits,10415),10415,192 if m else 128)
# Quotes remain diagnostic bytes, not parsed rating input or private definitions.
diag=avp(279,avp(461,b'unknown@example.net')+u(9999,0,10415))
out=[]
def save(n,b,req,expect='pair',profile='legacy'):
 f='b49-'+n+'.hex';(dst/f).write_text(b.hex()+'\n')
 out.append(dict(file=f,bytes=len(b),sha256=hashlib.sha256(b).hexdigest(),request=req+'.hex',expect=expect,profile=profile,source='RFC4006 4.1.3/9.2, RFC6733 7.1.5; frozen Ro result safety matrix',provenance='independent synthetic wire; no OCS capture or C++ execution'))
uses=[(k,'b47-'+k+'-success','b47-'+k+'-request','legacy') for k in ['iec','initial','term','general','balance','refund']]
uses += [('std-iec','v-iec-cca-02','v-iec-ccr-01','standard'),('std-initial','v-sms-cca-01','v-sms-ccr-01','standard'),('std-term','v-term-cca-01','v-term-ccr-01','standard')]
# The standard ECUR fixture names are resolved explicitly from the frozen set below.
for k,seed,req,profile in uses:
 b=load(seed)
 for n,top,extra,expect in [('rating',5031,diag,'pair'),('version',5011,b'','pair'),('conflict',2001,exp()+sf(),'conflict'),('feature-failure',5012,exp()+sf(),'unsupported')]:
  save(k+'-'+n,frame(b,top,extra,drop=[422,456]),req,expect,profile)
seed=load('b47-iec-success');req='b47-iec-request'
for n,top,extra,drop,expect in [
 ('rating-missing',5031,b'',[],'reject'),('rating-empty',5031,avp(279,b''),[],'reject'),
 ('rating-duplicate',5031,diag+diag,[],'reject'),('rating-wrong-vendor',5031,avp(279,u(461,0),10415),[],'reject'),
 ('rating-quoted-incomplete',5031,avp(279,b'\x01\x02\x03'),[],'unsupported'),
 ('rating-quoted-missing-example',5031,avp(279,avp(417,bytes(8))),[],'pair'),
 ('exp-only',2001,exp(),[268],'reject'),('exp-duplicate',2001,exp()+exp(),[],'reject'),
 ('exp-missing-vendor',2001,avp(297,u(298,5011)),[],'reject'),('exp-missing-code',2001,avp(297,u(266,10415)),[],'reject'),
 ('exp-width',2001,avp(297,u(266,10415)+avp(298,b'\x01')),[],'reject'),
 ('exp-duplicate-code',2001,avp(297,u(266,10415)+u(298,5011)+u(298,5012)),[],'reject'),
 ('exp-wrong-parent',2001,avp(456,u(268,2001)+exp()),[],'reject'),
 ('exp-unknown-vendor',2001,exp(999,5011),[],'unsupported'),('exp-unknown-code',2001,exp(10415,5999),[],'unsupported'),
 ('exp-base5011',5011,exp(),[],'unsupported'),('exp-no-features',2001,exp(),[],'conflict'),
 ('features-only',2001,sf(),[],'unsupported'),('features-unknown-vendor',2001,sf(999,7),[],'unsupported'),
 ('features-m1',2001,sf(m=True),[],'unsupported'),('features-multiple',2001,sf()+sf(10415,2,1),[],'unsupported'),
 ('features-missing-leaf',2001,avp(628,u(266,10415)+u(629,1,10415),10415,128),[],'reject'),
 ('features-duplicate-leaf',2001,avp(628,u(266,10415)+u(629,1,10415)+u(630,0,10415)+u(630,1,10415),10415,128),[],'reject'),
 ('version-missing-top',5011,b'',[268],'reject')]:save(n,frame(seed,top,extra,drop),req,expect)
for n,code,extra,drop,expect in [('version',5011,b'',[],'pair'),('rating',5031,diag,[],'pair'),('rating-missing',5031,b'',[],'reject'),('exp',5012,exp(),[],'reject'),('exp-only',5012,exp(),[268],'reject')]:
 save('e-'+n,frame(seed,code,extra,drop,error=True),req,expect)
manifest=json.load(open(dst/'manifest.json'));manifest=[v for v in manifest if not v['file'].startswith('b49-')]+out
(dst/'manifest.json').write_text(json.dumps(manifest,ensure_ascii=False,indent=2)+'\n')
print(len(out),'new frames',len(manifest),'total')
