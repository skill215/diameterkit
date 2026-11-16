#!/usr/bin/env python3
"""Synthetic Router ALR/ALA vectors assembled without diameterkit's Go codec."""
import hashlib,json,struct
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1]; D=ROOT/'testdata/codec'
def avp(code,data,vendor=0,flags=None):
    if flags is None: flags=0xc0 if vendor else 0x40
    h=struct.pack('!I',code)+bytes([flags]); n=8+(4 if flags&0x80 else 0)+len(data)
    b=h+n.to_bytes(3,'big')+(struct.pack('!I',vendor) if flags&0x80 else b'')+data
    return b+b'\0'*((-len(b))%4)
def fields(b):
    out=[];i=0
    while i<len(b):
        n=int.from_bytes(b[i+5:i+8],'big');v=int.from_bytes(b[i+8:i+12],'big') if b[i+4]&128 else 0;h=12 if b[i+4]&128 else 8
        assert n>=h and i+n<=len(b)
        out.append((int.from_bytes(b[i:i+4],'big'),v,b[i+4],b[i+h:i+n]));i+=(n+3)&~3
    assert i==len(b);return out
def frame(header,fs):
    b=b''.join(avp(c,d,v,f) for c,v,f,d in fs);h=bytearray(header);h[1:4]=(20+len(b)).to_bytes(3,'big');return bytes(h)+b
def main():
    source=bytes.fromhex((D/'v21-router-alr-r.hex').read_text());base=fields(source[20:]);h=source[:20];cases={}
    def request(name,token=b'router-alert-21',event=2,node=True):
        fs=[]
        for c,v,f,d in base:
            if c==3102:d=avp(1,token)
            if c==3333:
                if event is None:continue
                d=struct.pack('!I',event)
            if c==2401 and not node:continue
            fs.append((c,v,f,d))
        cases[name]=frame(h,fs)
    request('r');request('reachable',event=1,node=False);request('both',event=3);request('move-no-node',node=False)
    request('numeric',b'001010123456789');request('short',b'7');request('unicode','令牌-é'.encode());request('decomposed','令牌-e\u0301'.encode())
    request('max',b'x'*256);request('long',b'x'*257);request('empty',b'');request('nul',b'a\0b');request('invalid-utf8',b'\xff')
    request('zero-event',event=0,node=False);request('missing-event',event=None,node=False);request('reachable-node',event=1)
    session=next(d for c,v,f,d in base if c==263)
    host=next(d for c,v,f,d in base if c==293);realm=next(d for c,v,f,d in base if c==283)
    for name,code,error in [('a',2001,False),('failure',5012,False),('e',3004,True)]:
        hh=bytearray(h);hh[4]=0x60 if error else 0x40
        cases[name]=frame(hh,[(263,0,64,session),(268,0,64,struct.pack('!I',code)),(277,0,64,struct.pack('!I',1)),(264,0,64,host),(296,0,64,realm)])
    for name in ['r','a','failure','e']:
        b=bytearray(cases[name]);b[8:12]=(16777313).to_bytes(4,'big');cases['legacy-'+name]=bytes(b)
    manifest=json.loads((D/'manifest.json').read_text());manifest=[e for e in manifest if not e['file'].startswith('b74-')]
    for name,b in cases.items():
        path='b74-router-'+name+'.hex';(D/path).write_text(b.hex()+'\n')
        manifest.append(dict(file=path,bytes=len(b),sha256=hashlib.sha256(b).hexdigest(),source='spec/batch-74.md',provenance='Independent Python AVP assembly from v21-router-alr-r envelope; synthetic, not captured'))
    (D/'manifest.json').write_text(json.dumps(manifest,ensure_ascii=False,indent=2)+'\n');print(len(cases),'new;',len(manifest),'total')
if __name__=='__main__':main()
