#!/usr/bin/env python3
"""AIGO 市场快照"""
import subprocess, json

B = "http://localhost:8080"

def api(method, path, data=None, token=None):
    args = ["curl","-s","-w","%{http_code}","--max-time","8","-X",method]
    if data: args += ["-H","Content-Type: application/json","-d",json.dumps(data)]
    if token: args += ["-H",f"Authorization: Bearer {token}"]
    args.append(f"{B}{path}")
    r = subprocess.run(args, capture_output=True, text=True, timeout=10)
    out = r.stdout.strip()
    code = int(out[-3:]) if len(out)>=3 and out[-3:].isdigit() else 0
    body = out[:-3].strip() if code else out
    return (json.loads(body) if body else {}), code

d,_ = api("POST","/api/v1/auth/register",{"public_key":"snap-3"})
tok = d["data"]["token"]

# 挂牌
d,_ = api("GET","/api/v1/listings",token=tok)
ls = d.get("data") or []
print(f"活跃挂牌: {len(ls)} 条")
for i,l in enumerate(ls,1):
    print(f"  #{i}  价格={l.get('Price','?')}  库存={l.get('Quantity','?')}  已售={l.get('SoldQty',0)}  类型={l.get('PriceType','?')}")

# 商品
d,_ = api("GET","/api/v1/products?limit=50",token=tok)
ps = d.get("data") or []
print(f"\n上架商品: {len(ps)} 个")
for i,p in enumerate(ps,1):
    print(f"  #{i}  [{p.get('status','?')}]  {p.get('category','?')}  ¥{p.get('price_min','?')}~{p.get('price_max','?')}")
