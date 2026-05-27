#!/usr/bin/env python3
"""
AIGO 全功能自动化测试套件 v2 — 所有 HTTP 通过 subprocess curl 调用
修复: 处理 Paginated 响应(顶层 data), 商品 draft→active 流程, 空数据保护
"""
import subprocess, json, sys, time, os

BASE = "http://localhost:8080"
rc = 0
pass_count, fail_count = 0, 0
errors = []

def test(name, condition, detail=""):
    global pass_count, fail_count
    if condition:
        pass_count += 1
        print(f"  ✅ PASS | {name}")
    else:
        fail_count += 1
        msg = f"  ❌ FAIL | {name}  {detail}"
        print(msg)
        errors.append(msg)

def api(method, path, data=None, token=None):
    """curl 封装: 返回 (json_dict, http_status_code)"""
    args = ["curl", "-s", "-X", method, "-w", "%{http_code}", "--max-time", "8"]
    if data:
        args += ["-H", "Content-Type: application/json", "-d", json.dumps(data)]
    if token:
        args += ["-H", f"Authorization: Bearer {token}"]
    args.append(f"{BASE}{path}")
    
    try:
        r = subprocess.run(args, capture_output=True, text=True, timeout=10)
        out = r.stdout.strip()
        if not out:
            return {"error": "empty"}, 0
        # 最后3字符是状态码
        if len(out) >= 3 and out[-3:].isdigit():
            code = int(out[-3:])
            body = out[:-3].strip()
        else:
            code, body = 0, out
        try:
            return (json.loads(body), code) if body else ({}, code)
        except json.JSONDecodeError:
            return {"raw": body[:200]}, code
    except Exception as e:
        return {"error": str(e)}, 0

def safe_data(d, key, default=None):
    """安全获取 data 中的字段 (兼容普通响应和 PaginatedResponse)"""
    if d is None:
        return default
    if key in d:
        return d[key]
    # PaginatedResponse 直接放 data 在顶层
    if key == "data":
        return d.get("data", default)
    return d.get(key, default)

print("=" * 68)
print("  🧪 AIGO - 全功能自动化测试报告")
print(f"  {time.strftime('%Y-%m-%d %H:%M:%S')}")
print("=" * 68)

# === 0. 健康检查 =============================================
print("\n\033[1m[0] 📡 健康检查\033[0m")
data, code = api("GET", "/health")
test("HTTP 200", code == 200, f"got {code}")
test("status=ok", data.get("data",{}).get("status")=="ok", str(data)[:80])
test("version 存在", bool(data.get("data",{}).get("version")))
test("code/message/data 格式", all(k in data for k in ["code","message","data"]))

# === 1. 认证系统 =============================================
print("\n\033[1m[1] 🔐 认证系统\033[0m")

data, code = api("POST", "/api/v1/auth/register", {})
test("注册-缺public_key→400", code == 400, str(code))

data, code = api("POST", "/api/v1/auth/register", {"public_key":"test-pk-1"})
test("注册-201", code == 201, f"got {code}")
test("注册返回user_id", bool(data.get("data",{}).get("user_id")))
test("注册返回api_key", bool(data.get("data",{}).get("api_key")))
test("注册返回token", bool(data.get("data",{}).get("token")))

uid1 = data["data"]["user_id"]
ak1  = data["data"]["api_key"]
tok1 = data["data"]["token"]
test("user_id 非空", bool(uid1))
test("api_key 是64位hex", len(ak1) == 64)
test("token 是JWT格式", tok1.count(".") == 2)

# 用户2
data2, _ = api("POST", "/api/v1/auth/register", {"public_key":"test-pk-2"})
uid2 = data2["data"]["user_id"]
tok2 = data2["data"]["token"]
test("用户2注册成功", bool(uid2) and bool(tok2))

# Token 交换
data, code = api("POST", "/api/v1/auth/token", {"api_key": ak1})
test("Token交换-200", code == 200)
test("Token交换返回新token", bool(data.get("data",{}).get("token")))

data, code = api("POST", "/api/v1/auth/token", {"api_key":"badkey"})
test("错误key→401", code == 401, str(code))

# 生成新 API Key
data, code = api("POST", "/api/v1/auth/api-key", {"name":"bot-key"}, token=tok1)
test("生成API Key-201", code == 201)
test("API Key名称匹配", data.get("data",{}).get("name")=="bot-key", str(data.get("data",{})))

# 未认证
data, code = api("GET", "/api/v1/wallet")
test("未认证→401", code == 401, str(code))

# === 2. 钱包 & 充值 ==========================================
print("\n\033[1m[2] 💰 钱包 & 充值\033[0m")

data, code = api("GET", "/api/v1/wallet", token=tok1)
test("查余额-200", code == 200)
w = data.get("data",{})
test("余额字段完整", all(k in w for k in ["fiat_balance","points_balance","currency"]), str(w))
init_fiat, init_pts = w["fiat_balance"], w["points_balance"]

data, code = api("POST", "/api/v1/recharge", {"amount":1000}, token=tok1)
test("充值-201", code == 201)
r = data.get("data",{})
test("有order_id", bool(r.get("order_id")))
test("有points_awarded", "points_awarded" in r)
test("积分奖励=100000", r.get("points_awarded") == 100000, str(r))
test("状态completed", r.get("status") == "completed")

data, _ = api("GET", "/api/v1/wallet", token=tok1)
w2 = data.get("data",{})
test("法币增加", w2["fiat_balance"] > init_fiat, f"{init_fiat}→{w2['fiat_balance']}")
test("积分增加", w2["points_balance"] > init_pts, f"{init_pts}→{w2['points_balance']}")
test("积分=100000", w2["points_balance"] == 100000, str(w2))

data, code = api("POST", "/api/v1/recharge", {"amount":-100}, token=tok1)
test("充值负数→400", code == 400, str(code))

data, code = api("POST", "/api/v1/wallet/convert", {"fiat_amount":500}, token=tok1)
test("法币转积分-200", code == 200)

data, code = api("GET", "/api/v1/wallet/transactions?limit=5", token=tok1)
test("交易流水-200", code == 200)
txns = data.get("data") or []
test("交易流水有记录", len(txns) > 0, str(len(txns)))

# === 3. 商品管理 =============================================
print("\n\033[1m[3] 📦 商品管理\033[0m")

data, code = api("POST", "/api/v1/products", {}, token=tok1)
test("创建缺字段→400", code == 400, str(code))

p1 = {
    "encrypted_title": "dGl0bGUx", "encrypted_key": "a2V5MQ==",
    "price_min": 1000, "price_max": 5000,
    "category": "digital", "tags": ["ebook","template"]
}
data, code = api("POST", "/api/v1/products", p1, token=tok1)
test("创建商品-201", code == 201, f"got {code}")
pid1 = data.get("data",{}).get("product_id","")
test("有product_id", bool(pid1))

data, code = api("PUT", f"/api/v1/products/{pid1}", {"status":"active"}, token=tok1)
test("激活商品1-200", code == 200, f"got {code}")

p2 = {
    "encrypted_title": "dGl0bGUy", "encrypted_key": "a2V5Mg==",
    "price_min": 2000, "price_max": 8000,
    "category": "digital", "tags": ["software"]
}
data, code = api("POST", "/api/v1/products", p2, token=tok1)
pid2 = data.get("data",{}).get("product_id","")
test("创建商品2-201", code == 201)
test("ID不同", pid2 != pid1)

data, code = api("PUT", f"/api/v1/products/{pid2}", {"status":"active"}, token=tok1)
test("激活商品2-200", code == 200)

data, code = api("GET", "/api/v1/products?limit=10", token=tok1)
test("商品列表-200", code == 200)
products = data.get("data") or []
test("列表>=2", len(products) >= 2, f"got {len(products)}")

data, code = api("GET", f"/api/v1/products/{pid1}", token=tok1)
test("商品详情-200", code == 200)
test("类别匹配", data.get("data",{}).get("category")=="digital", str(data.get("data",{})))

data, code = api("GET", "/api/v1/products/nonexistent", token=tok1)
test("不存在→404", code == 404, str(code))

data, code = api("PUT", f"/api/v1/products/{pid1}", {"price_min":1500}, token=tok1)
test("更新商品-200", code == 200)

data, code = api("DELETE", f"/api/v1/products/{pid2}", token=tok1)
test("删除商品-200", code == 200)

data, code = api("GET", f"/api/v1/products/{pid2}", token=tok1)
test("删除后404", code == 404, str(code))

# === 4. 挂牌交易 =============================================
print("\n\033[1m[4] 🏪 挂牌交易\033[0m")

data, code = api("POST", "/api/v1/listings", {
    "product_id": pid1, "price_type": "fixed", "price": 1500, "quantity": 5
}, token=tok1)
test("创建挂牌-201", code == 201, f"got {code}")
lid = data.get("data",{}).get("listing_id","")
test("有listing_id", bool(lid))

data, code = api("GET", "/api/v1/listings", token=tok1)
test("挂牌列表-200", code == 200)
listings = data.get("data") or []
test("列表>=1", len(listings) >= 1, f"got {len(listings)}")

# 购买前先给买家充值
data, code = api("POST", "/api/v1/recharge", {"amount":10000}, token=tok2)
test("买家充值-201", code == 201, str(code))
test("买家获得积分", data.get("data",{}).get("points_awarded")==1000000, str(data))

data, code = api("POST", f"/api/v1/listings/{lid}/buy", {"quantity":2}, token=tok2)
test("购买-201", code == 201, f"got {code}")
order = data.get("data",{})
oid = order.get("order_id","")
test("有order_id", bool(oid))
test("有total_price", "total_price" in order, str(order))
test("状态pending", order.get("status")=="pending", str(order.get("status")))

data, code = api("POST", "/api/v1/listings/nonexistent/buy", {"quantity":1}, token=tok2)
test("不存在挂牌→错误", code >= 400, str(code))

# === 5. 订单管理 =============================================
print("\n\033[1m[5] 📋 订单管理\033[0m")

data, code = api("GET", "/api/v1/orders", token=tok2)
test("买家订单列表-200", code == 200)
orders = data.get("data") or []
test("买家有订单", len(orders) >= 1, f"got {len(orders)}")

data, code = api("GET", "/api/v1/orders", token=tok1)
orders_seller = data.get("data") or []
test("卖家订单列表-200", code == 200)
test("卖家有订单", len(orders_seller) >= 1)

data, code = api("GET", f"/api/v1/orders/{oid}", token=tok1)
test("订单详情-200", code == 200)
test("状态pending", data.get("data",{}).get("status")=="pending")

data, code = api("GET", "/api/v1/orders/nonexistent", token=tok1)
test("不存在→404", code == 404, str(code))

data, code = api("POST", f"/api/v1/orders/{oid}/confirm", token=tok1)
test("确认订单-200", code == 200)
test("→completed", data.get("data",{}).get("status")=="completed", str(data))

data, code = api("POST", f"/api/v1/orders/{oid}/confirm", token=tok1)
test("重复确认报错", code >= 400, str(code))

# 争议
data, code = api("POST", f"/api/v1/listings/{lid}/buy", {"quantity":1}, token=tok2)
doid = data.get("data",{}).get("order_id","")
test("争议-购买成功", bool(doid))

data, code = api("POST", f"/api/v1/orders/{doid}/dispute",
                 {"reason":"描述不符"}, token=tok2)
test("发起争议-200", code == 200)
test("→disputed", data.get("data",{}).get("status")=="disputed", str(data))

# 取消
data, code = api("POST", f"/api/v1/listings/{lid}/buy", {"quantity":1}, token=tok2)
coid = data.get("data",{}).get("order_id","")
test("取消-购买成功", bool(coid))

data, code = api("POST", f"/api/v1/orders/{coid}/cancel", token=tok2)
test("取消订单-200", code == 200)
test("→cancelled", data.get("data",{}).get("status")=="cancelled", str(data))

# === 6. 广播系统 =============================================
print("\n\033[1m[6] 📢 广播系统\033[0m")

data, code = api("GET", "/api/v1/broadcasts", token=tok1)
test("广播列表-200", code == 200)

data, code = api("POST", "/api/v1/broadcasts/commercial",
                 {"title":"广告","content":"测试","level":"basic"}, token=tok2)
test("积分不足→错误", code >= 400, str(code))

data, code = api("POST", "/api/v1/broadcasts/commercial",
                 {"title":"限时","content":"AI模板5折","level":"basic"}, token=tok1)
test("商业广播-201", code == 201, f"got {code}")
bc = data.get("data",{})
test("有broadcast_id", bool(bc.get("broadcast_id")), str(bc))
test("扣积分500", bc.get("points_cost")==500, str(bc.get("points_cost")))

data, code = api("POST", "/api/v1/broadcasts/commercial",
                 {"title":"标准","content":"置顶1h","level":"standard"}, token=tok1)
test("standard-201", code == 201)
test("扣2000积分", data.get("data",{}).get("points_cost")==2000, str(data))

data, code = api("POST", "/api/v1/broadcasts/commercial",
                 {"title":"高级","content":"置顶4h含链接","level":"premium",
                  "link_url":"https://aigo.test.com"}, token=tok1)
test("premium-201", code == 201)
test("扣5000积分", data.get("data",{}).get("points_cost")==5000, str(data))

data, code = api("POST", "/api/v1/admin/broadcasts/system",
                 {"title":"维护","content":"今晚2-4点维护"}, token=tok1)
test("系统广播-201", code == 201)
test("有broadcast_id", bool(data.get("data",{}).get("broadcast_id")))

data, code = api("GET", "/api/v1/broadcasts", token=tok1)
test("广播列表有内容", len(data.get("data") or []) >= 1)

data, code = api("GET", "/api/v1/broadcasts/unread", token=tok1)
test("未读数-200", code == 200)
test("未读数是整数", isinstance(data.get("data",{}).get("unread_count"), int))

# 标记已读
bcs = data.get("data")  # 来自上面的 broadcasts 变量
# 重新获取广播列表
data2, _ = api("GET", "/api/v1/broadcasts", token=tok1)
bclist = data2.get("data") or []
if bclist:
    bc_id = bclist[0].get("ID","")
    if bc_id:
        data, code = api("PUT", f"/api/v1/broadcasts/{bc_id}/read", token=tok1)
        test("标记已读-200", code == 200)

# === 7. 边缘情况 =============================================
print("\n\033[1m[7] ⚠️ 边缘情况\033[0m")

data, code = api("GET", "/api/v1/nonexistent-route")
test("不存在路由→404", code == 404, str(code))

data, code = api("GET", "/api/v1/wallet", token="")
test("空token→401", code == 401, str(code))

data, code = api("POST", "/api/v1/recharge", {"amount":5000}, token=tok1)
test("大额充值-201", code == 201)
test("积分正确", data.get("data",{}).get("points_awarded")==500000, str(data))

data, code = api("GET", "/api/v1/products?category=digital&limit=5", token=tok1)
test("商品过滤-200", code == 200)

# === 8. 最终积分 =============================================
print("\n\033[1m[8] 🧮 积分经济最终验证\033[0m")
data, _ = api("GET", "/api/v1/wallet", token=tok1)
fp = data.get("data",{}).get("points_balance",0)
test("积分>0", fp > 0, f"points={fp}")
test("积分合理", 100000 <= fp <= 1000000, f"points={fp}")

# === 结果 ====================================================
print("\n" + "=" * 68)
total = pass_count + fail_count
pct = round(pass_count/total*100, 1) if total else 0
print(f"  📊 总用例: {total}  ✅ 通过: {pass_count} ({pct}%)  ❌ 失败: {fail_count}")
if errors:
    print(f"\n  ⚠️  失败详情:")
    for i, e in enumerate(errors):
        print(f"    {i+1}. {e}")
print(f"\n  结论: ", end="")
if fail_count == 0:
    print("✅ 全部通过 — 系统稳定可靠")
elif fail_count <= 3:
    print("⚠️ 基本通过 — 少量边缘问题")
else:
    print("❌ 需要修复 — 存在严重问题")
print("=" * 68)

sys.exit(0 if fail_count == 0 else 1)
