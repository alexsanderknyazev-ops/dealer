# employees-service — API

gRPC-сервис `employees.v1.EmployeesService`. Доступ защищён JWT.

## Endpoints

| gRPC | HTTP | Описание |
|---|---|---|
| `CreateEmployee` | `POST /api/employees` | Создание сотрудника |
| `GetEmployee` | `GET /api/employees/{id}` | Сотрудник по id |
| `GetEmployeeByUserID` | — (внутренний gRPC) | Сотрудник по user_id из auth |
| `ListEmployees` | `GET /api/employees` | Список сотрудников |
| `UpdateEmployee` | `PUT /api/employees/{id}` | Обновление сотрудника (частичное) |
| `DeleteEmployee` | `DELETE /api/employees/{id}` | Удаление сотрудника |

## Сообщения

### Employee (модель)
`id`, `user_id`, `full_name`, `position`, `department`, `phone`, `active`, `created_at`, `updated_at`

### CreateEmployee
Request: `user_id`, `full_name`, `position`, `department`, `phone`, `active`
Response: `employee`

### GetEmployee / GetEmployeeByUserID
Request: `id` или `user_id`
Response: `employee`

### ListEmployees
Request: `limit`, `offset`, `search`, `position`, `active_only`
Response: `employees[]`, `total`

### UpdateEmployee
Request: `id`, optional `user_id`, `full_name`, `position`, `department`, `phone`, `active`
Response: `employee`

### DeleteEmployee
Request: `id`
Response: пусто
