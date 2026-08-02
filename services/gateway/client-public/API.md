# client-public-gateway-service — API

Публичный HTTP-шлюз контура клиентов (порт 8091). Без JWT. grpc-gateway reverse proxy на client-auth и client-registration.

## Маршруты

| HTTP | Описание | Бэкенд |
|---|---|---|
| `POST /api/client/register` | Регистрация клиента | client-registration-service |
| `POST /api/login` | Вход клиента | client-auth-service |
| `POST /api/refresh` | Обновление токенов | client-auth-service |
| `POST /api/logout` | Выход | client-auth-service |
