#!/usr/bin/env python3
"""Independent RFC4006 single-service EVENT frames; not SMS or live OCS proof."""
import hashlib
import json
import struct
from pathlib import Path

def avp(c,d,flags=64,vendor=0):
    n=8+len(d)+(4 if flags&128 else 0)
    return struct.pack('!I',c)+bytes([flags])+n.to_bytes(3,'big')+(struct.pack('!I',vendor) if flags&128 else b'')+d+bytes(-n%4)
def u(c,v):return avp(c,struct.pack('!I',v))
def units(n):return avp(417,struct.pack('!Q',n))
def money(digits,exponent=-2,currency=978):return avp(413,avp(445,avp(447,struct.pack('!q',digits))+avp(429,struct.pack('!i',exponent)))+u(425,currency))
def frame(parts,flags):
    b=b''.join(parts)
    return b'\x01'+(20+len(b)).to_bytes(3,'big')+bytes([flags])+(272).to_bytes(3,'big')+struct.pack('!III',4,0xb1b2b370,0xc1c2c370)+b
session=avp(263,b'ctf.example;event;70')
sub=avp(443,u(450,0)+avp(444,b'15550000001'))
request=[session,avp(264,b'ctf.example'),avp(296,b'example'),avp(283,b'example'),u(258,4),avp(461,b'event@example.net'),u(416,4),u(415,0),avp(293,b'ocs.example'),sub]
answer=[session,u(268,2001),avp(264,b'ocs.example'),avp(296,b'example'),u(258,4),u(416,4),u(415,0)]
frames={}
def put(n,parts,flags):frames['b70-'+n]=frame(parts,flags)
for action,name in [(0,'debit'),(1,'refund'),(2,'balance')]:
    put('r-'+name,request+[u(436,action),u(439,7)],192)
    for typ,data in [('zero',units(0)),('max',units(2**64-1)),('money',money(-123)),('money-min',money(-2**63,-2**31,2**32-1)),('money-max',money(2**63-1,2**31-1))]:
        put('r-'+name+'-'+typ,request+[u(436,action),u(439,7),avp(437,data)],192)
put('a-balance-enough',answer+[u(422,0)],64)
put('a-balance-no',answer+[u(422,1)],64)
put('a-balance-missing',answer,64)
put('a-balance-unknown',answer+[u(422,2**32-1)],64)
put('a-debit',answer+[avp(431,units(1)),u(428,0)],64)
put('a-debit-zero',answer+[avp(431,units(0))],64)
put('a-debit-money',answer+[avp(431,money(-123))],64)
put('a-refund',answer,64)
put('a-refund-cost',answer+[avp(423,avp(445,avp(447,struct.pack('!q',-123))+avp(429,struct.pack('!i',-2)))+u(425,978)+avp(424,b'event'))],64)
for code in [4010,4011,4012,5011,5012,5030]:put('a-'+str(code),answer[:1]+[u(268,code)]+answer[2:],64)
put('a-5031',answer[:1]+[u(268,5031)]+answer[2:]+[avp(279,u(439,7))],64)
put('a-e',answer[:1]+answer[2:4]+[u(268,3004)],96)
balance=request+[u(436,2),u(439,7)]
for idx,name in [(4,'app'),(5,'context'),(6,'type'),(7,'number'),(9,'subscription'),(10,'action'),(11,'service')]:
    put('r-missing-'+name,balance[:idx]+balance[idx+1:],192)
    put('r-duplicate-'+name,balance+balance[idx:idx+1],192)
for idx,name,value in [(4,'app',u(258,5)),(5,'context-format',avp(461,b'missing-domain')),(5,'sms',avp(461,b'18.32274@3gpp.org')),(6,'initial',u(416,1)),(7,'number',u(415,1)),(10,'price',u(436,3)),(9,'bad-sub',avp(443,u(450,0)+avp(444,b'nondecimal')))]:put('r-wrong-'+name,balance[:idx]+[value]+balance[idx+1:],192)
put('r-duplicate-sub-type',balance[:10]+[sub]+balance[10:],192)
for name,unit in [('empty',b''),('mixed',units(1)+money(1)),('width',avp(417,bytes(4))),('legacy420',avp(420,money(1),0))]:put('r-unit-'+name,balance+[avp(437,unit)],192)
put('r-mscc',balance+[avp(456,u(439,7),0)],192)
put('r-token2022',balance+[avp(2022,b'\x00\xff',128,10415)],192)
put('r-token1292',balance+[avp(1292,b'\x00\xff',0)],192)
put('r-unknown-optional',balance+[avp(999001,b'\x00\xff',0)],192)
put('r-unknown-mandatory',balance+[avp(999001,b'x')],192)
put('a-number',answer[:-1]+[u(415,1),u(422,0)],64)
put('a-type',answer[:5]+[u(416,1)]+answer[6:]+[u(422,0)],64)
put('a-duplicate-balance',answer+[u(422,0),u(422,1)],64)
put('a-5031-missing',answer[:1]+[u(268,5031)]+answer[2:],64)
put('a-experimental',answer+[avp(297,u(266,10415)+u(298,5011),0)],64)
manifest=json.loads(Path('testdata/codec/manifest.json').read_text());manifest=[x for x in manifest if not x['file'].startswith('b70-')]
for name,b in frames.items():
    Path('testdata/codec',name+'.hex').write_text('\n'.join(b[i:i+32].hex(' ') for i in range(0,len(b),32))+'\n')
    manifest.append(dict(file=name+'.hex',bytes=len(b),sha256=hashlib.sha256(b).hexdigest(),source='interop/dcca-event-fixtures.py',provenance='Independent Python RFC4006 single-service EVENT projection;not SMS or OCS capture'))
Path('testdata/codec/manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
print(len(frames),'new;',len(manifest),'total')
