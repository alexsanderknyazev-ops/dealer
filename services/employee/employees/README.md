# employees-service

Сервис контура сотрудников: HR — сотрудники дилера, их привязка к пользователям auth (по `user_id`). Доступ защищён JWT.

## Стек

- Go 1.22, gRPC, JWT (golang-jwt/v5)
- PostgreSQL (pgx/v5) — схема `employees`

## Порты

| Протокол | Порт |
|----------|------|
| gRPC | 50066 |
| HTTP (health/metrics) | 8099 |

## gRPC API

`employees.v1.EmployeesService`:
- `CreateEmployee`, `GetEmployee`, `GetEmployeeByUserID`, `ListEmployees`, `UpdateEmployee`, `DeleteEmployee`

## Взаимодействия

- Исходящие gRPC: —
- Kafka: —
- Хранилища: PostgreSQL (`employees`)

## Запуск

```bash
go run ./services/employee/employees   # make run-employees
```

Docker: `build/employees-service.Dockerfile`, compose-сервис `employees-service`, версия в `VERSION`.

## API

Полное описание всех эндпоинтов — см. [API.md](API.md).
