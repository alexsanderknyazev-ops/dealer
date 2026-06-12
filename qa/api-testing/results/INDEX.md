# Индекс прогонов тестирования

| Run ID | Дата (UTC) | Тип | Pass | Fail | Skip | Pass % | Исполнитель | Отчёт |
|--------|------------|-----|------|------|------|--------|-------------|-------|
| full-20260611-214710 | 2026-06-11 | Full API + E2E | ~38 | 5 bugs | 1 | — | Auto QA | [FULL-TEST-REPORT.md](./latest/FULL-TEST-REPORT.md) |

## Статус фаз (текущий цикл)

| Фаза | Статус | Протокол |
|------|--------|----------|
| 0 Подготовка | ✅ | fixtures applied |
| 1 Smoke L0 | ⚠️ | ~82 pass, abort on errors-ingest (fixed script) |
| 2 Employee | ✅ | read/write OK |
| 3 Parts/WO | ⚠️ | E2E OK after parts restart; WO-001 bug |
| 4 Client | ✅ | login, profile OK |
| 5 Stats | ⚠️ | reviews/stats 500 |
| 6 Security | ❌ | SEC-001 client token on :8090 |
| Go/No-Go | **NO-GO** | [bugs.md](./latest/bugs.md) |

См. [`../TEST-PLAN.md`](../TEST-PLAN.md)
