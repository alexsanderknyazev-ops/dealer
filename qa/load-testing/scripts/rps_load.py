#!/usr/bin/env python3
"""Fixed-RPS GET load test (stdlib only)."""
from __future__ import annotations

import argparse
import json
import random
import statistics
import sys
import threading
import time
import urllib.error
import urllib.request
from concurrent.futures import FIRST_COMPLETED, Future, ThreadPoolExecutor, wait
from datetime import datetime, timezone
from pathlib import Path


def login(base_url: str, email: str, password: str, retries: int = 12, delay: float = 10.0) -> str:
    payload = json.dumps({"email": email, "password": password}).encode()
    last_err: Exception | None = None
    for attempt in range(1, retries + 1):
        try:
            req = urllib.request.Request(
                f"{base_url}/api/login",
                data=payload,
                headers={"Content-Type": "application/json"},
                method="POST",
            )
            with urllib.request.urlopen(req, timeout=30) as resp:
                data = json.loads(resp.read())
            token = data.get("access_token")
            if not token:
                raise SystemExit("login failed: no access_token")
            if attempt > 1:
                print(f"[login] ok on attempt {attempt}", flush=True)
            return token
        except Exception as e:
            last_err = e
            print(f"[login] attempt {attempt}/{retries} failed: {e}", flush=True)
            if attempt < retries:
                time.sleep(delay)
    raise SystemExit(f"login failed after {retries} attempts: {last_err}")


def load_manifest(path: Path) -> tuple[str, list[dict]]:
    data = json.loads(path.read_text(encoding="utf-8"))
    endpoints = [e for e in data.get("endpoints", []) if e.get("auth")]
    if not endpoints:
        raise SystemExit(f"no auth endpoints in {path}")
    return data.get("base_url", ""), endpoints


def wait_any(futures: set[Future]) -> tuple[set[Future], set[Future]]:
    done, not_done = wait(futures, return_when=FIRST_COMPLETED)
    return done, not_done


def pct(values: list[float], p: float) -> float:
    if not values:
        return 0.0
    values = sorted(values)
    idx = max(0, min(len(values) - 1, int(len(values) * p / 100) - 1))
    return values[idx]


class LatencyReservoir:
    """Keep at most N latency samples for percentile estimates."""

    def __init__(self, capacity: int = 20000) -> None:
        self.capacity = capacity
        self.samples: list[float] = []
        self.seen = 0
        self._lock = threading.Lock()

    def add(self, value: float) -> None:
        with self._lock:
            self.seen += 1
            if len(self.samples) < self.capacity:
                self.samples.append(value)
            else:
                j = random.randint(0, self.seen - 1)
                if j < self.capacity:
                    self.samples[j] = value

    def snapshot(self) -> list[float]:
        with self._lock:
            return list(self.samples)


class TokenHolder:
    def __init__(
        self,
        base_url: str,
        email: str,
        password: str,
        refresh_sec: int = 1200,
        login_retries: int = 30,
        login_delay: float = 10.0,
    ) -> None:
        self.base_url = base_url
        self.email = email
        self.password = password
        self.refresh_sec = refresh_sec
        self.login_retries = login_retries
        self.login_delay = login_delay
        self._token = login(base_url, email, password, login_retries, login_delay)
        self._lock = threading.Lock()
        self._stop = threading.Event()
        self.refreshes = 0
        threading.Thread(target=self._refresh_loop, daemon=True).start()

    def get(self) -> str:
        with self._lock:
            return self._token

    def _refresh_loop(self) -> None:
        while not self._stop.wait(self.refresh_sec):
            try:
                new = login(self.base_url, self.email, self.password)
                with self._lock:
                    self._token = new
                    self.refreshes += 1
                print(f"[token] refreshed (#{self.refreshes})", flush=True)
            except Exception as e:
                print(f"[token] refresh failed: {e}", flush=True)

    def stop(self) -> None:
        self._stop.set()


def do_get(base_url: str, token_holder: TokenHolder, path: str) -> tuple[int, float, str | None]:
    url = f"{base_url}{path}"
    headers = {"Authorization": f"Bearer {token_holder.get()}", "Accept": "application/json"}
    req = urllib.request.Request(url, headers=headers, method="GET")
    t0 = time.monotonic()
    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            resp.read()
            return resp.status, time.monotonic() - t0, None
    except urllib.error.HTTPError as e:
        return e.code, time.monotonic() - t0, None
    except Exception as e:
        return 0, time.monotonic() - t0, type(e).__name__


def run(
    base_url: str,
    email: str,
    password: str,
    endpoints: list[dict],
    rps: float,
    duration_sec: int,
    max_workers: int,
    report_path: Path,
    checkpoint_sec: int,
    login_retries: int,
    login_delay: float,
) -> int:
    total = int(rps * duration_sec)
    pool: list[str] = []
    for ep in endpoints:
        pool.extend([ep["path"]] * ep.get("weight", 1))

    latencies = LatencyReservoir()
    codes: dict[int, int] = {}
    exceptions: dict[str, int] = {}
    ok = 0
    issued = 0
    lock = threading.Lock()
    token_holder = TokenHolder(
        base_url, email, password,
        login_retries=login_retries, login_delay=login_delay,
    )

    def record(status: int, latency: float, exc: str | None) -> None:
        nonlocal ok
        latencies.add(latency)
        with lock:
            if exc:
                exceptions[exc] = exceptions.get(exc, 0) + 1
            else:
                codes[status] = codes.get(status, 0) + 1
                if status == 200:
                    ok += 1

    def snapshot() -> tuple[int, int, int, dict, dict]:
        with lock:
            done = sum(codes.values()) + sum(exceptions.values())
            return done, ok, issued, dict(codes), dict(exceptions)

    started = time.monotonic()
    deadline = started + duration_sec
    interval = 1.0 / rps
    next_at = started
    stop_progress = threading.Event()
    prev_done, prev_ok = 0, 0

    def progress_loop() -> None:
        nonlocal prev_done, prev_ok
        while not stop_progress.wait(checkpoint_sec):
            elapsed = time.monotonic() - started
            if elapsed >= duration_sec:
                break
            done, local_ok, local_issued, c, ex = snapshot()
            window_done = done - prev_done
            window_ok = local_ok - prev_ok
            window_err = window_done - window_ok
            err = done - local_ok
            actual = done / elapsed if elapsed else 0
            window_rate = 100 * window_ok / window_done if window_done else 0
            total_rate = 100 * local_ok / done if done else 0
            lat = latencies.snapshot()
            codes_str = ", ".join(f"{k}:{v}" for k, v in sorted(c.items()) if k != 200)
            ex_str = ", ".join(f"{k}:{v}" for k, v in sorted(ex.items(), key=lambda x: -x[1])[:3])
            print(
                f"[{elapsed/60:.0f}m] total={done} ok={local_ok} err={err} ({total_rate:.1f}%) "
                f"| window +{window_done} ok={window_ok} err={window_err} ({window_rate:.1f}%) "
                f"rps={actual:.1f} p95={pct(lat, 95)*1000:.0f}ms "
                f"codes=[{codes_str}] exc=[{ex_str}]",
                flush=True,
            )
            prev_done, prev_ok = done, local_ok

    progress = threading.Thread(target=progress_loop, daemon=True)
    progress.start()

    with ThreadPoolExecutor(max_workers=max_workers) as pool_exec:
        inflight: set[Future] = set()
        while issued < total:
            now = time.monotonic()
            if now >= deadline:
                break
            if now < next_at:
                time.sleep(min(next_at - now, deadline - now))
            if len(inflight) >= max_workers:
                done_set, _ = wait_any(inflight)
                for fut in done_set:
                    inflight.discard(fut)
                    status, lat, exc = fut.result()
                    record(status, lat, exc)
            path = pool[issued % len(pool)]
            fut = pool_exec.submit(do_get, base_url, token_holder, path)
            inflight.add(fut)
            issued += 1
            next_at += interval

        for fut in inflight:
            status, lat, exc = fut.result()
            record(status, lat, exc)

    stop_progress.set()
    token_holder.stop()
    progress.join(timeout=2)

    elapsed = time.monotonic() - started
    done = sum(codes.values()) + sum(exceptions.values())
    err = done - ok
    lat_snap = latencies.snapshot()
    summary = {
        "base_url": base_url,
        "rps_target": rps,
        "duration_sec": duration_sec,
        "max_workers": max_workers,
        "token_refreshes": token_holder.refreshes,
        "total_issued": issued,
        "completed": done,
        "ok": ok,
        "errors": err,
        "error_rate": round(err / done, 4) if done else 0,
        "actual_rps": round(done / elapsed, 2) if elapsed else 0,
        "latency_ms": {
            "p50": round(pct(lat_snap, 50) * 1000, 1),
            "p95": round(pct(lat_snap, 95) * 1000, 1),
            "p99": round(pct(lat_snap, 99) * 1000, 1),
            "max": round(max(lat_snap) * 1000, 1) if lat_snap else 0,
            "mean": round(statistics.mean(lat_snap) * 1000, 1) if lat_snap else 0,
        },
        "http_codes": {str(k): v for k, v in sorted(codes.items())},
        "exceptions": exceptions,
        "finished_at": datetime.now(timezone.utc).isoformat(),
        "elapsed_sec": round(elapsed, 1),
    }

    if done and err / done < 0.01 and summary["actual_rps"] >= rps * 0.9:
        verdict = "PASS"
    elif done and err / done >= 0.05:
        verdict = "FAIL"
    else:
        verdict = "DEGRADED"

    lines = [
        f"# Load test — {rps} RPS × {duration_sec}s",
        "",
        f"- Target: **{rps} RPS** for **{duration_sec // 60} min** (~{total} requests)",
        f"- Completed: **{done}** (actual **{summary['actual_rps']} RPS**)",
        f"- OK 200: **{ok}** ({100*ok/done:.2f}%)" if done else "- OK: 0",
        f"- Errors: **{err}** ({100*err/done:.2f}%)" if done else "- Errors: 0",
        f"- Latency p50/p95/p99: **{summary['latency_ms']['p50']}/{summary['latency_ms']['p95']}/{summary['latency_ms']['p99']} ms**",
        f"- Latency max/mean: **{summary['latency_ms']['max']}/{summary['latency_ms']['mean']} ms**",
        "",
        "## HTTP codes",
        "",
    ]
    for code, count in sorted(codes.items()):
        lines.append(f"- `{code}`: {count}")
    if exceptions:
        lines += ["", "## Exceptions", ""]
        for name, count in sorted(exceptions.items(), key=lambda x: -x[1]):
            lines.append(f"- `{name}`: {count}")
    lines += ["", f"## Verdict: **{verdict}**", "", "```json", json.dumps(summary, indent=2), "```", ""]

    report_path.write_text("\n".join(lines), encoding="utf-8")
    latest = report_path.parent / "latest-load-report.md"
    latest.write_text(report_path.read_text(encoding="utf-8"), encoding="utf-8")

    print("\n=== FINAL ===")
    for line in lines[:8]:
        print(line)
    print(f"Verdict: {verdict}")
    print(f"Report: {report_path}")
    return 0 if verdict != "FAIL" else 1


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--base-url", default="http://192.168.0.27:9080")
    parser.add_argument("--email", default="qa.master@test.local")
    parser.add_argument("--password", default="Test1234!")
    parser.add_argument(
        "--manifest",
        default=str(Path(__file__).resolve().parent.parent / "results/latest-endpoints.json"),
    )
    parser.add_argument("--rps", type=float, default=150)
    parser.add_argument("--duration", type=int, default=600)
    parser.add_argument("--max-workers", type=int, default=400)
    parser.add_argument("--checkpoint-sec", type=int, default=300, help="progress interval")
    parser.add_argument("--login-retries", type=int, default=30)
    parser.add_argument("--login-delay", type=float, default=10.0)
    parser.add_argument("--report", default="")
    args = parser.parse_args()

    base_url = args.base_url.rstrip("/")
    manifest_path = Path(args.manifest)
    manifest_url, endpoints = load_manifest(manifest_path)
    if manifest_url:
        base_url = manifest_url.rstrip("/")

    run_id = datetime.now(timezone.utc).strftime("%Y%m%d-%H%M%S")
    report_path = Path(args.report) if args.report else manifest_path.parent / f"load-{run_id}-report.md"

    print(f"Login {args.email} @ {base_url}")
    print(
        f"Load: {args.rps} RPS × {args.duration}s = {int(args.rps*args.duration)} req, "
        f"workers={args.max_workers}, endpoints={len(endpoints)}, checkpoint={args.checkpoint_sec}s"
    )
    return run(
        base_url, args.email, args.password, endpoints,
        args.rps, args.duration, args.max_workers, report_path, args.checkpoint_sec,
        args.login_retries, args.login_delay,
    )


if __name__ == "__main__":
    sys.exit(main())
