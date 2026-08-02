#!/usr/bin/env python3
"""Probe all employee-gateway GET endpoints. Discovers IDs from list responses."""
from __future__ import annotations

import json
import os
import sys
import urllib.error
import urllib.request
from datetime import datetime, timezone
from pathlib import Path

BASE_URL = os.environ.get("BASE_URL", "http://192.168.0.27:9080").rstrip("/")
LOGIN_EMAIL = os.environ.get("LOGIN_EMAIL", "qa.master@test.local")
LOGIN_PASSWORD = os.environ.get("LOGIN_PASSWORD", "Test1234!")
CLIENT_ID = os.environ.get("CLIENT_ID", "a7700001-0000-4000-8000-000000000002")
FALLBACK_BRAND = "40000000-0000-4000-8000-000000000003"
FALLBACK_DP = "10000000-0000-4000-8000-000000000001"
TIMEOUT = int(os.environ.get("PROBE_TIMEOUT", "15"))

ROOT = Path(__file__).resolve().parent.parent
OUT_DIR = Path(os.environ.get("OUT_DIR", ROOT / "results"))
RUN_ID = f"probe-{datetime.now(timezone.utc).strftime('%Y%m%d-%H%M%S')}"
RUN_DIR = OUT_DIR / RUN_ID


def request(method: str, path: str, token: str | None = None) -> tuple[int, str]:
    url = f"{BASE_URL}{path}"
    headers = {"Accept": "application/json"}
    if token:
        headers["Authorization"] = f"Bearer {token}"
    req = urllib.request.Request(url, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=TIMEOUT) as resp:
            body = resp.read().decode("utf-8", errors="replace")
            return resp.status, body[:120].replace("\n", " ")
    except urllib.error.HTTPError as e:
        body = e.read().decode("utf-8", errors="replace") if e.fp else ""
        return e.code, body[:120].replace("\n", " ")
    except Exception as e:
        return 0, str(e)[:120]


def login() -> str:
    payload = json.dumps({"email": LOGIN_EMAIL, "password": LOGIN_PASSWORD}).encode()
    req = urllib.request.Request(
        f"{BASE_URL}/api/login",
        data=payload,
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    with urllib.request.urlopen(req, timeout=TIMEOUT) as resp:
        data = json.loads(resp.read())
    token = data.get("access_token")
    if not token:
        raise SystemExit("login failed: no access_token")
    return token


def first_id(data: dict, *keys: str) -> str | None:
    for key in keys:
        val = data.get(key)
        if isinstance(val, list) and val and isinstance(val[0], dict):
            item_id = val[0].get("id")
            if item_id:
                return str(item_id)
    return None


def is_json_api(snippet: str) -> bool:
    s = snippet.lstrip()
    return s.startswith("{") or s.startswith("[")


def probe(token: str | None, eid: str, path: str, auth: bool) -> dict:
    code, snippet = request("GET", path, token if auth else None)
    if code == 200 and auth and path.startswith("/api/") and not is_json_api(snippet):
        status = "WARN"
    elif code == 200:
        status = "PASS"
    else:
        status = "FAIL"
    return {
        "id": eid,
        "method": "GET",
        "path": path,
        "auth": auth,
        "http": code,
        "status": status,
        "body_snippet": snippet,
    }


def main() -> int:
    RUN_DIR.mkdir(parents=True, exist_ok=True)
    token = login()
    results: list[dict] = []
    discovered: dict[str, str] = {}
    load_endpoints: list[dict] = []

    def add(eid: str, path: str, auth: bool = True, weight: int = 1) -> None:
        r = probe(token, eid, path, auth)
        results.append(r)
        if r["status"] == "PASS":
            load_endpoints.append(
                {"id": eid, "method": "GET", "path": path, "auth": auth, "weight": weight}
            )
        if r["http"] == 200 and auth and r["status"] == "PASS":
            try:
                _, full = request("GET", path, token)
                data = json.loads(
                    urllib.request.urlopen(
                        urllib.request.Request(
                            f"{BASE_URL}{path}",
                            headers={
                                "Authorization": f"Bearer {token}",
                                "Accept": "application/json",
                            },
                        ),
                        timeout=TIMEOUT,
                    ).read()
                )
                return data
            except Exception:
                return None
        return None

    def add_list(key: str, path: str, json_keys: tuple[str, ...], weight: int = 3) -> None:
        r = probe(token, f"{key}-list", path, True)
        results.append(r)
        if r["status"] == "PASS":
            load_endpoints.append(
                {"id": f"{key}-list", "method": "GET", "path": path, "auth": True, "weight": weight}
            )
        if r["http"] != 200 or r["status"] != "PASS":
            return
        try:
            req = urllib.request.Request(
                f"{BASE_URL}{path}",
                headers={"Authorization": f"Bearer {token}", "Accept": "application/json"},
            )
            with urllib.request.urlopen(req, timeout=TIMEOUT) as resp:
                data = json.loads(resp.read())
            item_id = first_id(data, *json_keys)
            if item_id:
                discovered[key] = item_id
        except Exception:
            pass

    # Static / list endpoints
    results.append(probe(None, "healthz", "/healthz", False))
    load_endpoints.append(
        {"id": "healthz", "method": "GET", "path": "/healthz", "auth": False, "weight": 1}
    )

    for eid, path in [
        ("me", "/api/me"),
    ]:
        add(eid, path)

    list_specs = [
        ("customers", "/api/customers?limit=20", ("customers",)),
        ("vehicles", "/api/vehicles?limit=20", ("vehicles",)),
        ("deals", "/api/deals?limit=20", ("deals",)),
        ("parts", "/api/parts?limit=20", ("parts",)),
        ("parts-folders", "/api/parts/folders?limit=20", ("folders",)),
        ("brands", "/api/brands?limit=20", ("brands",)),
        ("brand-labor-rates", "/api/brand-labor-rates", ()),
        ("dealer-points", "/api/dealer-points?limit=20", ("dealer_points", "dealerPoints")),
        ("legal-entities", "/api/legal-entities?limit=20", ("legal_entities", "legalEntities")),
        ("warehouses", "/api/warehouses?limit=20", ("warehouses",)),
        ("work-orders", "/api/work-orders?limit=20", ("work_orders", "workOrders")),
        ("works", "/api/works?limit=20", ("works",)),
        ("works-folders", "/api/works/folders?limit=20", ("folders",)),
        ("employees", "/api/employees?limit=20", ("employees",)),
        ("reviews", "/api/reviews?limit=20", ("reviews",)),
        ("movement-documents", "/api/movement-documents?limit=20", ("documents",)),
        ("suppliers", "/api/suppliers?limit=20", ("suppliers",)),
        ("supplier-orders", "/api/supplier-orders?limit=20", ("orders", "supplier_orders")),
        ("customer-orders", "/api/customer-orders?limit=20", ("orders", "customer_orders")),
        ("repair-appointments", "/api/repair-appointments?limit=20", ("appointments",)),
    ]
    for key, path, keys in list_specs:
        if keys:
            add_list(key, path, keys)
        else:
            add(key, path)

    for eid, path in [
        ("stats-employee", "/api/stats/employee/overview"),
        ("stats-client", "/api/stats/client/overview"),
        ("reviews-stats", "/api/reviews/stats"),
        ("client-reviews", f"/api/clients/{CLIENT_ID}/reviews"),
    ]:
        add(eid, path, weight=2)

    brand_id = discovered.get("brands", FALLBACK_BRAND)
    dp_id = discovered.get("dealer-points", FALLBACK_DP)
    add("brand-labor-resolve", f"/api/brand-labor-rates/resolve?brand_id={brand_id}")
    add("repair-slots", f"/api/repair-appointment-slots?dealer_point_id={dp_id}&limit=10")
    add("dp-legal", f"/api/dealer-points/{dp_id}/legal-entities")

    by_id_paths = {
        "customers": "/api/customers/{id}",
        "vehicles": "/api/vehicles/{id}",
        "deals": "/api/deals/{id}",
        "parts": "/api/parts/{id}",
        "parts-folders": "/api/parts/folders/{id}",
        "brands": "/api/brands/{id}",
        "dealer-points": "/api/dealer-points/{id}",
        "legal-entities": "/api/legal-entities/{id}",
        "warehouses": "/api/warehouses/{id}",
        "work-orders": "/api/work-orders/{id}",
        "works": "/api/works/{id}",
        "works-folders": "/api/works/folders/{id}",
        "employees": "/api/employees/{id}",
        "reviews": "/api/reviews/{id}",
        "movement-documents": "/api/movement-documents/{id}",
        "supplier-orders": "/api/supplier-orders/{id}",
        "customer-orders": "/api/customer-orders/{id}",
        "repair-appointments": "/api/repair-appointments/{id}",
    }
    for key, tmpl in by_id_paths.items():
        item_id = discovered.get(key)
        if not item_id:
            results.append(
                {
                    "id": f"{key}-by-id",
                    "method": "GET",
                    "path": "—",
                    "auth": True,
                    "http": 0,
                    "status": "SKIP",
                    "body_snippet": "no id from list",
                }
            )
            continue
        path = tmpl.replace("{id}", item_id)
        add(f"{key}-by-id", path, weight=2)
        if key == "parts":
            add("parts-stock", f"/api/parts/{item_id}/stock", weight=2)

    passed = sum(1 for r in results if r["status"] == "PASS")
    failed = sum(1 for r in results if r["status"] == "FAIL")
    warned = sum(1 for r in results if r["status"] == "WARN")
    skipped = sum(1 for r in results if r["status"] == "SKIP")

    manifest = {
        "base_url": BASE_URL,
        "login_email": LOGIN_EMAIL,
        "probed_at": datetime.now(timezone.utc).isoformat(),
        "discovered_ids": discovered,
        "summary": {"pass": passed, "fail": failed, "warn": warned, "skip": skipped},
        "endpoints": load_endpoints,
        "probe_results": results,
    }

    manifest_path = RUN_DIR / "endpoints.json"
    report_path = RUN_DIR / "probe-report.md"
    manifest_path.write_text(json.dumps(manifest, indent=2) + "\n", encoding="utf-8")

    lines = [
        f"# GET endpoints probe — {RUN_ID}",
        "",
        "| Field | Value |",
        "|-------|-------|",
        f"| Base URL | `{BASE_URL}` |",
        f"| User | `{LOGIN_EMAIL}` |",
        f"| Passed | {passed} |",
        f"| Warn (HTML not JSON) | {warned} |",
        f"| Failed | {failed} |",
        f"| Skipped | {skipped} |",
        "",
        "## Results",
        "",
        "| ID | Method | Path | HTTP | Status | Body (truncated) |",
        "|----|--------|------|------|--------|------------------|",
    ]
    for r in results:
        lines.append(
            f"| {r['id']} | {r['method']} | `{r['path']}` | {r['http']} | {r['status']} | {r['body_snippet']} |"
        )
    report_path.write_text("\n".join(lines) + "\n", encoding="utf-8")

    (OUT_DIR / "latest-endpoints.json").write_text(manifest_path.read_text(encoding="utf-8"))
    (OUT_DIR / "latest-probe-report.md").write_text(report_path.read_text(encoding="utf-8"))

    print(f"Report: {report_path}")
    print(f"Manifest: {manifest_path}")
    print(f"PASS={passed} WARN={warned} FAIL={failed} SKIP={skipped}")
    for r in results:
        mark = {"PASS": "✓", "WARN": "!", "SKIP": "-", "FAIL": "✗"}[r["status"]]
        print(f"  {mark} {r['http']:>3} [{r['status']}] GET {r['path']}")

    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())
