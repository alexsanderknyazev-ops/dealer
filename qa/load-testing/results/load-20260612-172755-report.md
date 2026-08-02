# Load test — 150.0 RPS × 600s

- Target: **150.0 RPS** for **10 min** (~90000 requests)
- Completed: **90000** (actual **149.99 RPS**)
- OK 200: **35470** (39.41%)
- Errors: **54530** (60.59%)
- Latency p50/p95/p99: **5.1/11415.3/30002.1 ms**
- Latency max/mean: **51020.7/1542.1 ms**

## HTTP codes

- `200`: 35470
- `500`: 543

## Exceptions

- `URLError`: 53854
- `timeout`: 107
- `ConnectionResetError`: 25
- `TimeoutError`: 1

## Verdict: **FAIL**

```json
{
  "base_url": "http://192.168.0.27:9080",
  "rps_target": 150.0,
  "duration_sec": 600,
  "max_workers": 400,
  "token_refreshes": 0,
  "total_issued": 90000,
  "completed": 90000,
  "ok": 35470,
  "errors": 54530,
  "error_rate": 0.6059,
  "actual_rps": 149.99,
  "latency_ms": {
    "p50": 5.1,
    "p95": 11415.3,
    "p99": 30002.1,
    "max": 51020.7,
    "mean": 1542.1
  },
  "http_codes": {
    "200": 35470,
    "500": 543
  },
  "exceptions": {
    "URLError": 53854,
    "timeout": 107,
    "TimeoutError": 1,
    "ConnectionResetError": 25
  },
  "finished_at": "2026-06-12T17:38:08.284149+00:00",
  "elapsed_sec": 600.0
}
```
