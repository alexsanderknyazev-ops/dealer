# customers-service — API

gRPC-сервис `customers.v1.CustomersService`. Доступ защищён JWT (роли admin/manager/sales).

## Endpoints

| gRPC | HTTP | Описание |
|---|---|---|
| `CreateCustomer` | `POST /api/customers` | Создание клиента |
| `GetCustomer` | `GET /api/customers/{id}` | Клиент по id |
| `ListCustomers` | `GET /api/customers` | Список клиентов (поиск по name/email/phone) |
| `UpdateCustomer` | `PUT /api/customers/{id}` | Обновление клиента (частичное) |
| `DeleteCustomer` | `DELETE /api/customers/{id}` | Удаление клиента |

## Сообщения

### Customer (модель)
`id`, `name`, `email`, `phone`, `customer_type` (`individual`/`legal`), `inn`, `address`, `notes`, `created_at`, `updated_at`

### CreateCustomer
Request: `name`, `email`, `phone`, `customer_type`, `inn`, `address`, `notes`
Response: `customer`

### GetCustomer
Request: `id`
Response: `customer`

### ListCustomers
Request: `limit`, `offset`, `search`
Response: `customers[]`, `total`

### UpdateCustomer
Request: `id`, optional `name`, `email`, `phone`, `customer_type`, `inn`, `address`, `notes`
Response: `customer`

### DeleteCustomer
Request: `id`
Response: пусто
