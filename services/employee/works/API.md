# works-service — API

gRPC-сервис `works.v1.WorksService`. Доступ защищён JWT.

## Endpoints

### Работы

| gRPC | HTTP | Описание |
|---|---|---|
| `CreateWork` | `POST /api/works` | Создание работы |
| `GetWork` | `GET /api/works/{id}` | Работа по id |
| `ListWorks` | `GET /api/works` | Список работ |
| `UpdateWork` | `PUT /api/works/{id}` | Обновление работы (частичное) |
| `DeleteWork` | `DELETE /api/works/{id}` | Удаление работы |

### Папки работ

| gRPC | HTTP | Описание |
|---|---|---|
| `CreateFolder` | `POST /api/works/folders` | Создание папки |
| `GetFolder` | `GET /api/works/folders/{id}` | Папка по id |
| `ListFolders` | `GET /api/works/folders` | Список папок |
| `UpdateFolder` | `PUT /api/works/folders/{id}` | Обновление папки |
| `DeleteFolder` | `DELETE /api/works/folders/{id}` | Удаление папки |

## Сообщения

### Work (модель)
`id`, `code`, `name`, `category`, `labor_hours`, `unit_price`, `notes`, `folder_id`, `created_at`, `updated_at`

### WorkFolder (модель)
`id`, `name`, `parent_id`, `created_at`, `updated_at`

### CreateWork
Request: `code`, `name`, `category`, `labor_hours`, `unit_price`, `notes`, `folder_id`
Response: `work`

### GetWork / UpdateWork / DeleteWork
Request: `id` (+ optional поля для update)
Response: `work` / пусто

### ListWorks
Request: `limit`, `offset`, `search`, `category`, `folder_id`
Response: `works[]`, `total`

### CreateFolder
Request: `name`, `parent_id` (пусто = корень)
Response: `folder`

### GetFolder / UpdateFolder / DeleteFolder
Request: `id` (+ optional `name`, `parent_id` для update)
Response: `folder` / пусто

### ListFolders
Request: `parent_id`
Response: `folders[]`
