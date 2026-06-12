#!/usr/bin/env python3
"""HTTP load test for Dealer stack with target RPS and per-service reporting."""

from __future__ import annotations

import argparse
import json
import os
import statistics
import sys
import threading
import time
import urllib.error
import urllib.request
from collections import defaultdict
from concurrent.futures import ThreadPoolExecutor, as_completed
from dataclasses import dataclass, field
from datetime import datetime, timezone
from pathlib import Path
from typing import Callable


OK_CODES = {200, 204}

# Endpoint label -> backend service (for aggregated report).
SERVICE_BY_ENDPOINT: dict[str, str] = {
    "GET auth /healthz": "auth-service",
    "GET gateway /healthz": "gateway-service",
    "GET client public /healthz": "client-public-gateway",
    "GET client protected /healthz": "client-protected-gateway",
    "GET /api/me": "auth-service",
    "POST /api/login": "auth-service",
    "GET /api/customers": "customers-service",
    "GET /api/vehicles": "vehicles-service",
    "GET /api/deals": "deals-service",
    "GET /api/parts": "parts-service",
    "GET /api/brands": "brands-service",
    "GET /api/works": "works-service",
    "GET /api/employees": "employees-service",
    "GET /api/work-orders": "workorders-service",
    "GET /api/stats/employee/overview": "employee-statistics-service",
    "GET /api/client/profile": "client-registration-service",
    "GET /api/client/vehicles": "client-registration-service",
}


@dataclass
class Sample:
    name: str
    service: str
    status: int
    latency_ms: float
    ts: float
    error: str = ""


@dataclass
class Stats:
    samples: list[Sample] = field(default_factory=list)
    lock: threading.Lock = field(default_factory=threading.Lock)
    started_at: float = field(default_factory=time.perf_counter)

    def add(self, sample: Sample) -> None:
        with self.lock:
            self.samples.append(sample)

    def snapshot_counts(self) -> tuple[int, int, float]:
        with self.lock:
            total = len(self.samples)
            ok = sum(1 for s in self.samples if s.status in OK_CODES)
            elapsed = max(time.perf_counter() - self.started_at, 0.001)
            return total, ok, total / elapsed


class RateLimiter:
    def __init__(self, rps: float) -> None:
        self.interval = 1.0 / rps
        self.lock = threading.Lock()
        self.next_at = time.perf_counter()

    def wait(self) -> None:
        with self.lock:
            now = time.perf_counter()
            if now < self.next_at:
                time.sleep(self.next_at - now)
                now = time.perf_counter()
            self.next_at = max(now, self.next_at) + self.interval


class HttpClient:
    def __init__(self, timeout: float) -> None:
        self.timeout = timeout

    def request(
        self,
        method: str,
        url: str,
        *,
        headers: dict[str, str] | None = None,
        body: dict | None = None,
    ) -> tuple[int, float, str]:
        hdrs = {"Accept": "application/json"}
        if headers:
            hdrs.update(headers)
        data = json.dumps(body).encode() if body is not None else None
        if data is not None:
            hdrs.setdefault("Content-Type", "application/json")

        req = urllib.request.Request(url, data=data, headers=hdrs, method=method)
        start = time.perf_counter()
        try:
            with urllib.request.urlopen(req, timeout=self.timeout) as resp:
                resp.read()
                code = resp.status
        except urllib.error.HTTPError as exc:
            code = exc.code
            try:
                exc.read()
            except Exception:
                pass
        except Exception as exc:  # noqa: BLE001
            elapsed = (time.perf_counter() - start) * 1000
            return 0, elapsed, str(exc)
        elapsed = (time.perf_counter() - start) * 1000
        return code, elapsed, ""


def post_json(client: HttpClient, url: str, body: dict) -> dict:
    req = urllib.request.Request(
        url,
        data=json.dumps(body).encode(),
        headers={"Content-Type": "application/json", "Accept": "application/json"},
        method="POST",
    )
    with urllib.request.urlopen(req, timeout=client.timeout) as resp:
        payload = json.loads(resp.read().decode())
    if resp.status != 200:
        raise RuntimeError(f"POST {url} failed: http={resp.status}")
    return payload


def percentile(values: list[float], pct: float) -> float:
    if not values:
        return 0.0
    values = sorted(values)
    idx = int(round((pct / 100) * (len(values) - 1)))
    return values[idx]


def latency_stats(values: list[float]) -> dict[str, float]:
    if not values:
        return {"min": 0, "avg": 0, "p50": 0, "p95": 0, "p99": 0, "max": 0}
    return {
        "min": min(values),
        "avg": statistics.mean(values),
        "p50": percentile(values, 50),
        "p95": percentile(values, 95),
        "p99": percentile(values, 99),
        "max": max(values),
    }


def run_worker(
    stop_at: float,
    task: Callable[[], Sample],
    stats: Stats,
    limiter: RateLimiter | None,
) -> None:
    while time.perf_counter() < stop_at:
        if limiter:
            limiter.wait()
        if time.perf_counter() >= stop_at:
            break
        stats.add(task())


def progress_loop(stop_at: float, stats: Stats, target_rps: float | None, interval: int) -> None:
    while time.perf_counter() < stop_at:
        time.sleep(interval)
        if time.perf_counter() >= stop_at:
            break
        total, ok, rps = stats.snapshot_counts()
        target = f" target={target_rps:.0f}" if target_rps else ""
        print(
            f"[progress] elapsed={int(time.perf_counter() - stats.started_at)}s "
            f"requests={total} ok={ok} rps={rps:.1f}{target}",
            flush=True,
        )


def round_robin(tasks: list[Callable[[], Sample]]) -> Callable[[], Sample]:
    idx = {"i": 0}
    lock = threading.Lock()

    def next_task() -> Sample:
        with lock:
            i = idx["i"]
            idx["i"] += 1
        return tasks[i % len(tasks)]()

    return next_task


def build_tasks(args: argparse.Namespace, client: HttpClient) -> dict[str, list[Callable[[], Sample]]]:
    employee_token = ""
    client_token = ""

    if args.scenario in {"employee-read", "mixed"}:
        payload = post_json(
            client,
            f"{args.employee_api}/api/login",
            {"email": args.employee_email, "password": args.password},
        )
        employee_token = payload["access_token"]
        print(f"Employee login OK: {args.employee_email}")

    if args.scenario in {"client-read", "mixed"}:
        payload = post_json(
            client,
            f"{args.client_public}/api/login",
            {"email": args.client_email, "password": args.password},
        )
        client_token = payload["access_token"]
        print(f"Client login OK: {args.client_email}")

    auth_h = {"Authorization": f"Bearer {employee_token}"}
    client_h = {"Authorization": f"Bearer {client_token}"}

    def get(name: str, url: str, headers: dict[str, str] | None = None) -> Callable[[], Sample]:
        service = SERVICE_BY_ENDPOINT.get(name, "unknown")

        def task() -> Sample:
            status, latency_ms, err = client.request("GET", url, headers=headers)
            return Sample(name, service, status, latency_ms, time.perf_counter(), err)

        return task

    def post(name: str, url: str, body: dict) -> Callable[[], Sample]:
        service = SERVICE_BY_ENDPOINT.get(name, "unknown")

        def task() -> Sample:
            status, latency_ms, err = client.request("POST", url, body=body)
            return Sample(name, service, status, latency_ms, time.perf_counter(), err)

        return task

    health = [
        get("GET auth /healthz", f"{args.employee_auth}/healthz"),
        get("GET gateway /healthz", f"{args.employee_api}/healthz"),
        get("GET client public /healthz", f"{args.client_public}/healthz"),
        get("GET client protected /healthz", f"{args.client_protected}/healthz"),
    ]

    employee_reads = [
        get("GET /api/me", f"{args.employee_api}/api/me", auth_h),
        get("GET /api/customers", f"{args.employee_api}/api/customers?limit=20", auth_h),
        get("GET /api/vehicles", f"{args.employee_api}/api/vehicles?limit=20", auth_h),
        get("GET /api/deals", f"{args.employee_api}/api/deals?limit=20", auth_h),
        get("GET /api/parts", f"{args.employee_api}/api/parts?limit=20", auth_h),
        get("GET /api/brands", f"{args.employee_api}/api/brands?limit=20", auth_h),
        get("GET /api/works", f"{args.employee_api}/api/works?limit=20", auth_h),
        get("GET /api/employees", f"{args.employee_api}/api/employees?limit=20", auth_h),
        get("GET /api/work-orders", f"{args.employee_api}/api/work-orders?limit=20", auth_h),
        get("GET /api/stats/employee/overview", f"{args.employee_api}/api/stats/employee/overview", auth_h),
    ]

    employee_login = [
        post(
            "POST /api/login",
            f"{args.employee_api}/api/login",
            {"email": args.employee_email, "password": args.password},
        )
    ]

    client_reads = [
        get("GET /api/client/profile", f"{args.client_protected}/api/client/profile", client_h),
        get("GET /api/client/vehicles", f"{args.client_protected}/api/client/vehicles", client_h),
    ]

    return {
        "health": health,
        "employee-read": employee_reads,
        "employee-login": employee_login,
        "client-read": client_reads,
        "mixed": employee_reads + client_reads + health[:2],
    }


def aggregate(samples: list[Sample], key_fn: Callable[[Sample], str]) -> dict[str, dict]:
    groups: dict[str, list[Sample]] = defaultdict(list)
    for s in samples:
        groups[key_fn(s)].append(s)

    out: dict[str, dict] = {}
    for key, items in sorted(groups.items(), key=lambda x: -len(x[1])):
        latencies = [s.latency_ms for s in items]
        ok = sum(1 for s in items if s.status in OK_CODES)
        status_codes: dict[str, int] = defaultdict(int)
        errors: dict[str, int] = defaultdict(int)
        for s in items:
            status_codes[str(s.status)] += 1
            if s.status not in OK_CODES:
                err_key = s.error or f"HTTP {s.status}"
                errors[err_key] += 1
        out[key] = {
            "requests": len(items),
            "success": ok,
            "success_rate": ok / len(items),
            "status_codes": dict(sorted(status_codes.items())),
            "errors": dict(sorted(errors.items(), key=lambda x: -x[1])),
            "latency_ms": latency_stats(latencies),
        }
    return out


def build_timeline(samples: list[Sample], started_at: float, bucket_s: int) -> list[dict]:
    if not samples:
        return []
    buckets: dict[int, list[Sample]] = defaultdict(list)
    for s in samples:
        bucket = int((s.ts - started_at) // bucket_s)
        buckets[bucket].append(s)

    timeline = []
    for bucket in sorted(buckets):
        items = buckets[bucket]
        latencies = [s.latency_ms for s in items]
        ok = sum(1 for s in items if s.status in OK_CODES)
        t_from = int(bucket * bucket_s)
        t_to = t_from + bucket_s
        timeline.append(
            {
                "from_s": t_from,
                "to_s": t_to,
                "requests": len(items),
                "rps": len(items) / bucket_s,
                "success_rate": ok / len(items),
                "latency_ms": {"p50": percentile(latencies, 50), "p95": percentile(latencies, 95)},
                "errors": len(items) - ok,
            }
        )
    return timeline


def build_report(
    args: argparse.Namespace,
    stats: Stats,
    actual_duration: float,
) -> tuple[str, dict]:
    samples = stats.samples
    if not samples:
        return "No samples collected.\n", {}

    latencies = [s.latency_ms for s in samples]
    ok = sum(1 for s in samples if s.status in OK_CODES)
    fail = len(samples) - ok
    actual_rps = len(samples) / max(actual_duration, 0.001)

    by_endpoint = aggregate(samples, lambda s: s.name)
    by_service = aggregate(samples, lambda s: s.service)
    timeline = build_timeline(samples, stats.started_at, args.bucket)

    meta = {
        "run_at": datetime.now(timezone.utc).isoformat(),
        "scenario": args.scenario,
        "target_rps": args.rps,
        "duration_s": args.duration,
        "actual_duration_s": round(actual_duration, 2),
        "concurrency": args.concurrency,
        "employee_api": args.employee_api,
        "employee_auth": args.employee_auth,
        "client_public": args.client_public,
        "client_protected": args.client_protected,
        "total_requests": len(samples),
        "expected_requests": int(args.rps * args.duration) if args.rps else None,
        "actual_rps": round(actual_rps, 2),
        "success": ok,
        "fail": fail,
        "success_rate": ok / len(samples),
        "latency_ms": latency_stats(latencies),
        "services": by_service,
        "endpoints": by_endpoint,
        "timeline": timeline,
    }

    lines = [
        "",
        f"=== Load test: {args.scenario} ===",
        f"Target RPS: {args.rps}  Duration: {args.duration}s  Concurrency: {args.concurrency}",
        f"Actual: {len(samples)} requests in {actual_duration:.1f}s ({actual_rps:.1f} RPS)",
        f"Success: {ok} ({100 * ok / len(samples):.2f}%)  Fail: {fail}",
        "",
        "Overall latency (ms):",
        (
            f"  min={meta['latency_ms']['min']:.0f}  avg={meta['latency_ms']['avg']:.0f}  "
            f"p50={meta['latency_ms']['p50']:.0f}  p95={meta['latency_ms']['p95']:.0f}  "
            f"p99={meta['latency_ms']['p99']:.0f}  max={meta['latency_ms']['max']:.0f}"
        ),
        "",
        "Per service:",
        f"  {'Service':<32} {'Reqs':>7} {'RPS':>7} {'OK%':>6} {'p50':>6} {'p95':>6} {'p99':>6} {'Err':>5}",
    ]

    for service, data in by_service.items():
        svc_rps = data["requests"] / max(actual_duration, 0.001)
        lat = data["latency_ms"]
        err_count = data["requests"] - data["success"]
        lines.append(
            f"  {service:<32} {data['requests']:>7} {svc_rps:>7.1f} "
            f"{100 * data['success_rate']:>5.1f}% {lat['p50']:>6.0f} {lat['p95']:>6.0f} "
            f"{lat['p99']:>6.0f} {err_count:>5}"
        )

    lines.extend(["", "Per endpoint:", f"  {'Endpoint':<42} {'Reqs':>6} {'OK%':>6} {'p95':>6} {'top err':>20}"])
    for name, data in by_endpoint.items():
        top_err = next(iter(data["errors"]), "—") if data["errors"] else "—"
        if len(top_err) > 20:
            top_err = top_err[:17] + "..."
        lines.append(
            f"  {name:<42} {data['requests']:>6} {100 * data['success_rate']:>5.0f}% "
            f"{data['latency_ms']['p95']:>6.0f} {top_err:>20}"
        )

    failed_services = [
        (svc, d) for svc, d in by_service.items() if d["success"] < d["requests"]
    ]
    if failed_services:
        lines.extend(["", "Service errors:"])
        for svc, data in failed_services:
            codes = ", ".join(f"{k}×{v}" for k, v in data["status_codes"].items() if k not in {"200", "204"})
            if codes:
                lines.append(f"  {svc}: {codes}")

    return "\n".join(lines) + "\n", meta


def write_reports(report_dir: Path, text: str, meta: dict) -> None:
    report_dir.mkdir(parents=True, exist_ok=True)
    (report_dir / "report.txt").write_text(text, encoding="utf-8")
    (report_dir / "summary.json").write_text(json.dumps(meta, indent=2, ensure_ascii=False), encoding="utf-8")
    (report_dir / "services.json").write_text(
        json.dumps(meta.get("services", {}), indent=2, ensure_ascii=False),
        encoding="utf-8",
    )
    (report_dir / "endpoints.json").write_text(
        json.dumps(meta.get("endpoints", {}), indent=2, ensure_ascii=False),
        encoding="utf-8",
    )
    (report_dir / "timeline.json").write_text(
        json.dumps(meta.get("timeline", []), indent=2, ensure_ascii=False),
        encoding="utf-8",
    )


def parse_args() -> argparse.Namespace:
    p = argparse.ArgumentParser(description="Dealer HTTP load test")
    p.add_argument(
        "--scenario",
        choices=["health", "employee-read", "employee-login", "client-read", "mixed"],
        default=os.environ.get("LOAD_SCENARIO", "mixed"),
    )
    p.add_argument("--rps", type=float, default=float(os.environ.get("LOAD_RPS", "150")))
    p.add_argument("--duration", type=int, default=int(os.environ.get("LOAD_DURATION", "600")))
    p.add_argument("--concurrency", type=int, default=int(os.environ.get("LOAD_CONCURRENCY", "0")))
    p.add_argument("--timeout", type=float, default=float(os.environ.get("LOAD_TIMEOUT", "15")))
    p.add_argument("--bucket", type=int, default=int(os.environ.get("LOAD_BUCKET", "30")),
                   help="Timeline bucket size in seconds")
    p.add_argument("--progress", type=int, default=int(os.environ.get("LOAD_PROGRESS", "30")),
                   help="Progress log interval in seconds (0=off)")
    p.add_argument("--employee-api", default=os.environ.get("EMPLOYEE_API", "http://127.0.0.1:8090"))
    p.add_argument("--employee-auth", default=os.environ.get("EMPLOYEE_AUTH", "http://127.0.0.1:9080"))
    p.add_argument("--client-public", default=os.environ.get("CLIENT_PUBLIC", "http://127.0.0.1:8091"))
    p.add_argument("--client-protected", default=os.environ.get("CLIENT_PROTECTED", "http://127.0.0.1:8093"))
    p.add_argument(
        "--employee-email",
        default=os.environ.get("LOAD_EMPLOYEE_EMAIL", "vol.employee1@test.dealer.local"),
    )
    p.add_argument(
        "--client-email",
        default=os.environ.get("LOAD_CLIENT_EMAIL", "vol.client1@test.dealer.local"),
    )
    p.add_argument("--password", default=os.environ.get("LOAD_PASSWORD", "Test1234!"))
    p.add_argument("--smoke", action="store_true", help="Quick run: 10s, 50 RPS, scenario health")
    p.add_argument("--report-dir", metavar="DIR", help="Directory for report files")
    p.add_argument("--json", dest="json_out", metavar="FILE", help="Write summary JSON (legacy)")
    return p.parse_args()


def auto_concurrency(rps: float) -> int:
    # Enough workers to sustain target RPS at ~100ms average latency.
    return max(10, min(100, int(rps * 0.15) + 5))


def main() -> int:
    args = parse_args()
    if args.smoke:
        args.duration = 10
        args.rps = 50
        args.scenario = "health"
        args.progress = 5

    if args.concurrency <= 0:
        args.concurrency = auto_concurrency(args.rps)

    for attr in ("employee_api", "employee_auth", "client_public", "client_protected"):
        setattr(args, attr, getattr(args, attr).rstrip("/"))

    client = HttpClient(timeout=args.timeout)
    all_tasks = build_tasks(args, client)
    tasks = all_tasks[args.scenario]
    worker = round_robin(tasks) if len(tasks) > 1 else tasks[0]
    limiter = RateLimiter(args.rps) if args.rps > 0 else None

    stats = Stats()
    stop_at = time.perf_counter() + args.duration

    expected = int(args.rps * args.duration)
    print(
        f"Starting scenario={args.scenario} target_rps={args.rps} duration={args.duration}s "
        f"(~{expected} requests) concurrency={args.concurrency}"
    )
    print(f"Employee API: {args.employee_api}")

    progress_thread = None
    if args.progress > 0:
        progress_thread = threading.Thread(
            target=progress_loop,
            args=(stop_at, stats, args.rps, args.progress),
            daemon=True,
        )
        progress_thread.start()

    with ThreadPoolExecutor(max_workers=args.concurrency) as pool:
        futures = [
            pool.submit(run_worker, stop_at, worker, stats, limiter)
            for _ in range(args.concurrency)
        ]
        for fut in as_completed(futures):
            fut.result()

    actual_duration = time.perf_counter() - stats.started_at
    text, meta = build_report(args, stats, actual_duration)
    print(text, end="")

    if args.report_dir:
        write_reports(Path(args.report_dir), text, meta)
        print(f"Reports written to {args.report_dir}/")
        print("  report.txt  summary.json  services.json  endpoints.json  timeline.json")

    if args.json_out:
        with open(args.json_out, "w", encoding="utf-8") as fh:
            json.dump(meta, fh, indent=2, ensure_ascii=False)

    fail = meta.get("fail", 0)
    return 0 if fail == 0 else 2


if __name__ == "__main__":
    raise SystemExit(main())
