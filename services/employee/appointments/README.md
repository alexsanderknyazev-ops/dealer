# appointments-service

Сервис контура сотрудников: запись клиентов на ремонт/ТО — слоты, создание/изменение/отмена записи, конвертация записи в заказ-наряд. Доступ защищён JWT.

## Стек

- Go 1.22, gRPC, JWT (golang-jwt/v5)
- PostgreSQL (pgx/v5) — схема `appointments`

## Порты

| Протокол | Порт |
|----------|------|
| gRPC | 50067 |
| HTTP (health/metrics) | 8101 |

## gRPC API

`appointments.v1.RepairAppointmentsService`:
- `ListRepairAppointmentSlots`, `CreateRepairAppointment`, `GetRepairAppointment`, `UpdateRepairAppointment`, `ListRepairAppointments`, `CancelRepairAppointment`
- `CreateWorkOrderFromRepairAppointment`

## Взаимодействия

- Исходящие gRPC: customers, vehicles, dealer-points, parts, workorders, works
- Kafka: —
- Хранилища: PostgreSQL (`appointments`)

## Запуск

```bash
go run ./services/employee/appointments   # make run-appointments
```

Docker: `build/appointments-service.Dockerfile`, compose-сервис `appointments-service`, версия в `VERSION`.

## API

Полное описание всех эндпоинтов — см. [API.md](API.md).
