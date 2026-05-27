#!/usr/bin/env python3
"""给卖家发私信 - 测试消息系统"""
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

# 注册
d,_ = api("POST","/api/v1/auth/register",{"public_key":"dongge-msg-test"})
tok = d["data"]["token"]
uid = d["data"]["user_id"]
print(f"✅ 注册: {uid[:20]}...")

# 查看消息系统是否在线（未读数）
d,_ = api("GET","/api/v1/messages/unread",token=tok)
print(f"📬 未读数: {d}")

# 给卖家发消息
d,c = api("POST","/api/v1/messages",{
    "receiver_id": SELLER_ID,
    "subject": "关于臣能灵芝孢子油的咨询",
    "body": "你好！我从你的商业广播中看到了臣能灵芝孢子油，已经下单购买了¥350的商品（订单ID: b3aa483c-8c20-4cdd-953e-6fd8d49d620e）。请问这个商品具体是哪款灵芝孢子油？是否可以提供更多产品信息和解密密钥？期待回复。"
},token=tok)
print(f"\n📤 发送消息: HTTP {c}")
if c == 201:
    mid = d.get("data",{}).get("message_id","")[:16]
    print(f"✅ 发送成功！消息ID: {mid}...")
    print(f"   收件人: {SELLER_ID[:20]}...")
else:
    print(f"❌ 发送失败: {json.dumps(d, indent=2)[:200]}")

# 查看已发送
d,_ = api("GET","/api/v1/messages/sent",token=tok)
sent = d.get("data") or []
print(f"\n📤 已发送: {len(sent)} 条")
for m in sent:
    print(f"  收件人={m['receiver_id'][:12]}... 主题={m.get('subject','(无)')} 时间={m.get('created_at','')[:19]}")
