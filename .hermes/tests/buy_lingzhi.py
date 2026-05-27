#!/usr/bin/env python3
"""买灵芝孢子油 - 完整流程"""
import subprocess, json

B = "http://localhost:8080"
SELLER_ID = "6c48589a-68b7-4d2f-9b88-03d1ad4aaca8"

def api(method, path, data=None, token=None):
    args = ["curl","-s","-w","%{http_code}","--max-time","8","-X",method]
    if data: args+=["-H","Content-Type: application/json","-d",json.dumps(data)]
    if token: args+=["-H",f"Authorization: Bearer {token}"]
    args.append(f"{B}{path}")
    r = subprocess.run(args, capture_output=True, text=True, timeout=10)
    o=r.stdout.strip(); c=int(o[-3:]) if len(o)>=3 and o[-3:].isdigit() else 0
    return (json.loads(o[:-3].strip()) if c else {}), c

# 1. 注册
d1, _ = api("POST","/api/v1/auth/register",{"public_key":"buyer-lingzhi-v3"})
tok = d1["data"]["token"]
uid = d1["data"]["user_id"]
print(f"✅ 注册: {uid[:20]}...")

# 2. 查余额
d, _ = api("GET","/api/v1/wallet",token=tok)
w = d.get("data",{})
print(f"💰 余额: 法币={w.get('fiat_balance',0)}分 积分={w.get('points_balance',0)}")

# 3. 充值 ¥400
d2, c2 = api("POST","/api/v1/recharge",{"amount":40000},token=tok)
print(f"💳 充值结果: HTTP {c2} -> {json.dumps(d2, indent=2)[:150]}")
if c2 != 201:
    print("充值失败，退出")
    exit(1)

# 4. 查余额
d, _ = api("GET","/api/v1/wallet",token=tok)
w = d.get("data",{})
print(f"💰 充值后: 积分={w.get('points_balance',0)}")

# 5. 找到¥350的挂牌并购买
d3, _ = api("GET","/api/v1/listings",token=tok)
target_lid = None
for l in (d3.get("data") or []):
    if l.get("price") == 35000 and l.get("seller_id") == SELLER_ID:
        target_lid = l["id"]
        print(f"\n🎯 找到目标: 挂牌={l['id'][:12]}... 价格=¥{l['price']/100:.0f} 库存={l['quantity']-l.get('sold_quantity',0)}/{l['quantity']}")
        break

if not target_lid:
    print("没找到¥350的挂牌")
    exit(1)

# 6. 购买
d4, c4 = api("POST",f"/api/v1/listings/{target_lid}/buy",{"quantity":1},token=tok)
print(f"\n🛒 购买结果: HTTP {c4}")
print(json.dumps(d4, indent=2))

if c4 == 201:
    oid = d4.get("data",{}).get("order_id","")
    print(f"\n🎉 购买成功！订单ID: {oid[:20]}...")
    print(f"   总价: ¥{d4.get('data',{}).get('total_price',0)/100:.0f}")
    print(f"   状态: {d4.get('data',{}).get('status','')}")
else:
    print(f"\n❌ 购买失败: {d4.get('message','')} - {json.dumps(d4, indent=2)[:200]}")
