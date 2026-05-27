# Auth + Deals Monitoring (PromQL)

Базовые метрики берутся из:
- `http_requests_total{service,method,path,status}`
- `http_request_duration_seconds_bucket{service,method,path,...}`

`/metrics` доступен через стандартный middleware (`observe.WrapHTTP` + `metrics.RecordHTTP`).

## Request Count

### Auth login throughput (rps)
```promql
sum(rate(http_requests_total{service="auth-service",method="POST",path="/api/login"}[5m]))
```

### Deals create throughput (rps)
```promql
sum(rate(http_requests_total{service="deals-service",method="POST",path="/api/deals"}[5m]))
```

## Error Rate

### Auth login error_rate (5xx + 4xx)
```promql
sum(rate(http_requests_total{service="auth-service",method="POST",path="/api/login",status=~"4..|5.."}[5m]))
/
sum(rate(http_requests_total{service="auth-service",method="POST",path="/api/login"}[5m]))
```

### Deals create error_rate (5xx + 4xx)
```promql
sum(rate(http_requests_total{service="deals-service",method="POST",path="/api/deals",status=~"4..|5.."}[5m]))
/
sum(rate(http_requests_total{service="deals-service",method="POST",path="/api/deals"}[5m]))
```

## p95 Latency

### Auth login p95
```promql
histogram_quantile(
  0.95,
  sum by (le) (
    rate(http_request_duration_seconds_bucket{service="auth-service",method="POST",path="/api/login"}[5m])
  )
)
```

### Deals create p95
```promql
histogram_quantile(
  0.95,
  sum by (le) (
    rate(http_request_duration_seconds_bucket{service="deals-service",method="POST",path="/api/deals"}[5m])
  )
)
```

## Alert Suggestions

- `error_rate > 0.05` в течение `10m` для `/api/login` и `/api/deals`.
- `p95 > 0.8s` для `/api/login`, `p95 > 1.2s` для `/api/deals` в течение `10m`.
