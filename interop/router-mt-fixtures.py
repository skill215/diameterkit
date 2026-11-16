#!/usr/bin/env python3
"""Independent synthetic Router TFR/TFA wire vectors, no diameterkit codec."""
import hashlib,json,struct,runpy
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1];D=ROOT/'testdata/codec'
f=runpy.run_path(str(Path(__file__).with_name('router-alert-fixtures.py')))
avp,fields,frame=(f[k] for k in ('avp','fields','frame'))
def read(n):return bytes.fromhex((D/(n+'.hex')).read_text())
def main():
    r=read('v24-sgd-mt-r');a=read('v24-sgd-mt-wait')
    rf=[(c,v,f,b'router.example' if c==293 else d) for c,v,f,d in fields(r[20:]) if c not in (1645,1489)]
    af=[(c,v,f,b'router.example' if c==264 else d) for c,v,f,d in fields(a[20:])]
    cases={'r':frame(r[:20],rf)}
    original=fields(r[20:])
    # TFR MME-Number-for-MT-SMS is 1645; SGSN-Number is 1489.
    mme=next(x for x in original if x[0]==1645)
    cases['mme-r']=frame(r[:20],rf+[mme]);cases['sgsn-r']=frame(r[:20],rf+[(1489,10415,128,mme[3])])
    cases['both-r']=frame(r[:20],rf+[mme,(1489,10415,128,mme[3])])
    cases['nomax-r']=frame(r[:20],[x for x in rf if x[0]!=3330])
    cases['noaddress-r']=frame(r[:20],[x for x in rf if x[0]!=3332])
    def answer(n,token=b'router-token-75',children=None,fs=None):
        if fs is None:fs=af
        group=avp(1,token) if children is None else children
        cases[n]=frame(a[:20],fs+[(3102,10415,192,group)])
    answer('wait');answer('numeric',b'001010123456789');answer('unicode','令牌-é'.encode());answer('decomposed','令牌-e\u0301'.encode());answer('max',b'x'*256)
    answer('long',b'x'*257);answer('empty',b'');answer('nul',b'a\0b');answer('invalid-utf8',b'\xff')
    answer('missing-name',children=b'');answer('msisdn',children=avp(1,b'token')+avp(701,b'\x21\xf3',10415))
    answer('external',children=avp(1,b'token')+avp(3111,b'x@example',10415))
    answer('lmsi',children=avp(1,b'token')+avp(2400,b'1234',10415))
    answer('duplicate-name',children=avp(1,b'token')*2)
    answer('unknown',children=avp(1,b'token')+avp(9999,b'opaque',10415,128))
    answer('notime',fs=[x for x in af if x[0]!=3331])
    cases['missing-token']=frame(a[:20],af)
    cases['duplicate-token']=frame(a[:20],fields(cases['wait'][20:])+[fields(cases['wait'][20:])[-1]])
    maximum=next(x[3] for x in rf if x[0]==3330)
    for n,value in [('equal',maximum),('late',struct.pack('!I',int.from_bytes(maximum,'big')+1)),('zero',b'\0'*4)]:
        answer(n,fs=[(c,v,f,value if c==3331 else d) for c,v,f,d in af])
    for n,code,error in [('success',2001,False),('failure',5012,False),('e',3004,True)]:
        hh=bytearray(a[:20]);hh[4]=0x60 if error else 0x40
        fs=[x for x in af if x[0] not in (297,3331,3322)]+[(268,0,64,struct.pack('!I',code))]
        cases[n]=frame(hh,fs)
    manifest=json.loads((D/'manifest.json').read_text());manifest=[x for x in manifest if not x['file'].startswith('b75-')]
    for n,b in cases.items():
        name='b75-router-mt-'+n+'.hex';(D/name).write_text(b.hex()+'\n')
        manifest.append(dict(file=name,bytes=len(b),sha256=hashlib.sha256(b).hexdigest(),source='spec/batch-75.md',provenance='Independent Python AVP assembly from v24 SGd MT envelope; synthetic, not captured'))
    (D/'manifest.json').write_text(json.dumps(manifest,ensure_ascii=False,indent=2)+'\n');print(len(cases),'new;',len(manifest),'total')
if __name__=='__main__':main()
