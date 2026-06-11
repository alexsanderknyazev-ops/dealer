# Результаты тестирования

Все артефакты прогонов хранятся здесь, отдельно от тест-кейсов и фикстур.

## Структура

```
results/
├── README.md                 ← вы здесь
├── INDEX.md                  ← сводка прогонов (обновляется вручную или скриптом)
├── templates/                ← шаблоны протоколов
├── runs/                     ← каждый прогон в своей папке
│   └── run-YYYYMMDD-HHMMSS/
│       ├── smoke-report.md   ← auto: run-api-tests.sh
│       ├── meta.json         ← pass/fail/skip, env
│       └── manual/           ← ручные протоколы фаз (опционально)
├── latest/                   ← копия последнего auto-прогона
│   ├── smoke-report.md
│   └── meta.json
└── manual/                   ← итоговые протоколы фаз / go-no-go
    ├── phase-0-setup.md
    ├── phase-2-employee.md
    …
    └── go-no-go.md
```

## Автоматический smoke-прогон

```bash
export POSTGRES_DSN='...'
./qa/api-testing/scripts/run-api-tests.sh
```

Создаётся:
- `runs/run-<timestamp>/smoke-report.md`
- `runs/run-<timestamp>/meta.json`
- копия в `latest/`

## Ручные протоколы

1. Скопировать шаблон из `templates/` в `manual/` или `runs/<run-id>/manual/`
2. Заполнить статусы ☐ → ✅ / ❌ / ⏭
3. Приложить SQL/curl в комментариях или отдельным файлом в той же папке run

## Именование run-папок

| Тип | Паттерн | Пример |
|-----|---------|--------|
| Auto smoke | `run-YYYYMMDD-HHMMSS` | `run-20260612-143052` |
| Full cycle | `cycle-YYYYMMDD` | `cycle-20260612` + подпапки фаз |
| Hotfix reg | `reg-<ticket>-YYYYMMDD` | `reg-DEAL-142-20260612` |

## Go/No-Go

Итоговое решение — `manual/go-no-go.md` (шаблон в `templates/go-no-go.md`).

## Git

По умолчанию `runs/*/` и `latest/*` в `.gitignore` (локальные прогоны).  
Для CI или релиза — закоммитить конкретную папку run или экспорт в `manual/go-no-go.md`.
