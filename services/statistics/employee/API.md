# employee-statistics-service — API

gRPC-сервис `statistics.employee.v1.EmployeeStatisticsService`. Доступ защищён JWT. Потребляет событие `deal.completed.v1`.

## Endpoints

| gRPC | HTTP | Описание |
|---|---|---|
| `GetOverview` | `GET /api/stats/employee/overview` | Сводная статистика по платформе |

## Сообщения

### GetOverview
Request: пусто

Response: `overview` (EmployeeOverview):
- `customers_count` — клиентов
- `vehicles_count` — автомобилей
- `deals_count` — сделок
- `deals_by_stage[]` — `{stage, count}` по этапам
- `total_revenue` — выручка
- `parts_count` — запчастей
- `dealer_points_count` — дилерских точек
