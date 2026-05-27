#!/usr/bin/env python3
"""
Batch publish jianke.com scraped products to AIGO with AES-256-GCM encryption.
Usage: python batch_publish.py [--dry-run]
"""

import os
import sys
import json
import base64
import glob
import csv
import time
import random
import argparse
from concurrent.futures import ThreadPoolExecutor, as_completed
from urllib.request import Request, urlopen
from urllib.error import URLError, HTTPError
import hashlib
import hmac

# ── AIGO credentials (from D:/aigo_jianke_credentials.txt) ──
API_KEY = "ee635ecaa0cccab7246a5b1cfac11a6caa14dc55840e76ca57b52df0b9510a36"
USER_ID = "6ceef0bb-d602-4780-a562-a4b0b9883211"
AIGO_BASE = "http://localhost:8080"
SYSTEM_CRYPTO_KEY = "c03c2fc76363d5296d1f404b41acc3c180c4b45dc62c112c31ea2e5eb5b72eb0"

# ── AES-256-GCM encryption (use pycryptodome) ──
try:
    from Crypto.Cipher import AES
    from Crypto.Random import get_random_bytes
    HAS_CRYPTO = True
except ImportError:
    HAS_CRYPTO = False


def encrypt_aes_gcm(plaintext: str, key_hex: str) -> bytes:
    """Encrypt plaintext with AES-256-GCM. Returns nonce+ciphertext+tag bytes."""
    if not plaintext:
        return b""
    if not HAS_CRYPTO:
        raise RuntimeError("pycryptodome not installed: pip install pycryptodome")
    key = bytes.fromhex(key_hex)
    cipher = AES.new(key, AES.MODE_GCM)
    ciphertext, tag = cipher.encrypt_and_digest(plaintext.encode("utf-8"))
    nonce = cipher.nonce
    return nonce + ciphertext + tag


def api_request(method: str, path: str, data: dict = None, token: str = None) -> dict:
    url = f"{AIGO_BASE}{path}"
    headers = {"Content-Type": "application/json", "X-AIGO-KEY": API_KEY}
    if token:
        headers["Authorization"] = f"Bearer {token}"
    body = json.dumps(data).encode() if data else None
    req = Request(url, method=method, data=body, headers=headers)
    try:
        with urlopen(req, timeout=30) as resp:
            return json.loads(resp.read())
    except HTTPError as e:
        body = e.read().decode() if e.fp else ""
        try:
            return json.loads(body)
        except Exception:
            return {"error": body, "code": e.code}
    except URLError as e:
        return {"error": str(e.reason)}


def get_token() -> str:
    """Exchange API key for JWT token."""
    result = api_request("POST", "/api/v1/auth/token", {"api_key": API_KEY})
    if "data" in result and "token" in result["data"]:
        return result["data"]["token"]
    raise RuntimeError(f"Failed to get token: {result}")


def publish_product(token: str, row: dict) -> dict:
    """Encrypt product fields and POST to AIGO."""
    title = row.get("productName", "").strip()
    common_name = row.get("commonName", "").strip()
    manufacturer = row.get("manufacturer", "").strip()
    formulation = row.get("formulation", "").strip()
    packing = row.get("packing", "").strip()
    our_price = row.get("ourPrice", "0").strip()
    market_price = row.get("marketPrice", "0").strip()
    introduction = row.get("introduction", "").strip()

    # Build metadata JSON
    metadata = json.dumps({
        "commonName": common_name,
        "manufacturer": manufacturer,
        "formulation": formulation,
        "packing": packing,
        "marketPrice": market_price,
        "prescriptionType": row.get("prescriptionType", ""),
        "productComment": row.get("productComment", ""),
        "img": row.get("img", ""),
    }, ensure_ascii=False)

    # Encrypt fields
    enc_title = encrypt_aes_gcm(title, SYSTEM_CRYPTO_KEY) if title else b""
    enc_desc = encrypt_aes_gcm(introduction, SYSTEM_CRYPTO_KEY) if introduction else b""
    enc_meta = encrypt_aes_gcm(metadata, SYSTEM_CRYPTO_KEY) if metadata else b""
    enc_key = encrypt_aes_gcm(SYSTEM_CRYPTO_KEY, SYSTEM_CRYPTO_KEY)  # dummy, not used for system decryption

    # Price range (in points, 1 point = 0.01 CNY)
    try:
        price_min = max(100, int(float(our_price) * 100))
        price_max = max(price_min, int(float(market_price) * 100))
    except (ValueError, TypeError):
        price_min = 100
        price_max = 500

    category = "药品"
    tags = [manufacturer[:20], formulation[:20]] if manufacturer or formulation else ["药品"]

    payload = {
        "encrypted_title": base64.b64encode(enc_title).decode(),
        "encrypted_description": base64.b64encode(enc_desc).decode(),
        "encrypted_metadata": base64.b64encode(enc_meta).decode(),
        "encrypted_key": base64.b64encode(enc_key).decode(),
        "price_min": price_min,
        "price_max": price_max,
        "category": category,
        "tags": tags,
    }

    result = api_request("POST", "/api/v1/products", payload, token)

    # Auto-activate the product
    if result.get("code") == 0 and "data" in result:
        product_id = result["data"].get("product_id", result["data"].get("id", ""))
        if product_id:
            api_request("PUT", f"/api/v1/products/{product_id}",
                        {"status": "active"}, token)
            return {"product_id": product_id, "title": title, "price_min": price_min}
    return {"error": str(result), "title": title}


def load_csv_files(pattern: str):
    """Load all CSV files matching pattern, yield (filename, rows)."""
    for path in sorted(glob.glob(pattern)):
        with open(path, "r", encoding="utf-8-sig") as f:
            reader = csv.DictReader(f)
            rows = list(reader)
        yield os.path.basename(path), rows
        print(f"  Loaded {len(rows)} rows from {os.path.basename(path)}")


def main():
    parser = argparse.ArgumentParser(description="Batch publish jianke products to AIGO")
    parser.add_argument("--dry-run", action="store_true", help="Encrypt but do not POST")
    parser.add_argument("--limit", type=int, default=0, help="Max products per file (0=all)")
    parser.add_argument("--workers", type=int, default=5, help="Parallel workers")
    args = parser.parse_args()

    csv_pattern = "D:/健客网_*.csv"
    csv_files = list(load_csv_files(csv_pattern))

    if not csv_files:
        print("No CSV files found!")
        sys.exit(1)

    total_rows = sum(len(rows) for _, rows in csv_files)
    print(f"\nTotal: {len(csv_files)} files, {total_rows} products")
    print(f"Mode: {'DRY RUN' if args.dry_run else 'LIVE PUBLISH'}\n")

    if args.dry_run:
        all_rows = []
        for _, rows in csv_files:
            limit = args.limit or len(rows)
            all_rows.extend(rows[:limit])
        for i, row in enumerate(all_rows[:10], 1):
            title = row.get("productName", "").strip()
            enc = encrypt_aes_gcm(title, SYSTEM_CRYPTO_KEY)
            print(f"[{i}] {title[:40]} → {base64.b64encode(enc).decode()[:60]}...")
        print(f"\n... (showing first 10 of {len(all_rows)})")
        print("Dry run complete. No products were published.")
        return

    print("Getting JWT token...")
    token = get_token()
    print(f"Token obtained: {token[:20]}...\n")

    success_count = 0
    error_count = 0
    results = []

    for fname, rows in csv_files:
        print(f"Processing {fname} ({len(rows)} products)...")
        limit = args.limit or len(rows)

        with ThreadPoolExecutor(max_workers=args.workers) as pool:
            futures = {
                pool.submit(publish_product, token, row): row
                for row in rows[:limit]
            }
            for future in as_completed(futures):
                result = future.result()
                if "product_id" in result:
                    success_count += 1
                    results.append(result)
                else:
                    error_count += 1
                    print(f"  ERROR: {result.get('title', '?')[:40]} → {result.get('error', '?')}")

                # Progress dot
                total = success_count + error_count
                if total % 50 == 0:
                    print(f"  ... {total} done ({success_count} ok, {error_count} err)")
                time.sleep(0.05)  # gentle rate limit

    print(f"\n{'='*50}")
    print(f"Done!  Success: {success_count}, Errors: {error_count}")

    # Save results
    out_path = "D:/aigo_published_products.json"
    with open(out_path, "w", encoding="utf-8") as f:
        json.dump({
            "published": results,
            "summary": {"success": success_count, "errors": error_count,
                        "total_files": len(csv_files)}
        }, f, ensure_ascii=False, indent=2)
    print(f"Results saved to {out_path}")


if __name__ == "__main__":
    main()