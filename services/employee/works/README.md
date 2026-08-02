# works-service

Сервис контура сотрудников: справочник работ СТО (виды работ, нормо-часы) и папки работ. Доступ защищён JWT.

## Стек

- Go 1.22, gRPC, JWT (golang-jwt/v5)
- PostgreSQL (pgx/v5) — схема `works`

## Порты

| Протокол | Порт |
|----------|------|
| gRPC | 50065 |
| HTTP (health/metrics) | 8098 |

## gRPC API

`works.v1.WorksService`:
- `CreateWork`, `GetWork`, `ListWorks`, `UpdateWork`, `DeleteWork`
- `CreateFolder`, `GetFolder`, `ListFolders`, `UpdateFolder`, `DeleteFolder`

## Взаимодействия

- Исходящие gRPC: —
- Kafka: —
- Хранилища: PostgreSQL (`works`)

## Запуск

```bash
go run ./services/employee/works   # make run-works
```

Docker: `build/works-service.Dockerfile`, compose-сервис `works-service`, версия в `VERSION`.

## API

Полное описание всех эндпоинтов — см. [API.md](API.md).
