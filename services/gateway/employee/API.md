# gateway-service — API

HTTP-шлюз контура сотрудников (порт 8090). grpc-gateway reverse proxy: принимает HTTP-запросы и транслирует их в gRPC-вызовы бэкенд-сервисов. Все маршруты требуют JWT-аутентификацию (валидируется на бэкендах), кроме методов auth.

## Маршруты

### Auth (auth-service)
| HTTP | Описание |
|---|---|
| `POST /api/register` | Регистрация сотрудника |
| `POST /api/login` | Вход |
| `POST /api/refresh` | Обновление токенов |
| `POST /api/logout` | Выход |
| `GET /api/me` | Текущий профиль |

### Клиенты (customers-service)
| HTTP | Описание |
|---|---|
| `POST /api/customers` | Создать клиента |
| `GET /api/customers` | Список клиентов |
| `GET /api/customers/{id}` | Клиент по id |
| `PUT /api/customers/{id}` | Обновить клиента |
| `DELETE /api/customers/{id}` | Удалить клиента |

### Автомобили (vehicles-service)
| HTTP | Описание |
|---|---|
| `POST /api/vehicles` | Добавить автомобиль |
| `GET /api/vehicles` | Список автомобилей |
| `GET /api/vehicles/{id}` | Автомобиль по id |
| `PUT /api/vehicles/{id}` | Обновить автомобиль |
| `DELETE /api/vehicles/{id}` | Удалить автомобиль |

### Сделки (deals-service)
| HTTP | Описание |
|---|---|
| `POST /api/deals` | Создать сделку |
| `GET /api/deals` | Список сделок |
| `GET /api/deals/{id}` | Сделка по id |
| `PUT /api/deals/{id}` | Обновить сделку |
| `DELETE /api/deals/{id}` | Удалить сделку |

### Запчасти и склад (parts-service)
| HTTP | Описание |
|---|---|
| `POST /api/parts` · `GET /api/parts` · `GET/PUT/DELETE /api/parts/{id}` | Запчасти |
| `GET /api/parts/{part_id}/stock` | Остатки по складам |
| `POST /api/parts/folders` · `GET /api/parts/folders` · `GET/PUT/DELETE /api/parts/folders/{id}` | Папки запчастей |
| `POST /api/movement-documents` · `GET /api/movement-documents` · `GET/PUT /api/movement-documents/{id}` | Документы перемещения |
| `POST /api/movement-documents/{id}/start\|close\|confirm\|cancel\|create-production-extraction` | Операции с документами |
| `GET /api/suppliers` | Поставщики |
| `POST /api/supplier-orders` · `GET /api/supplier-orders` · `GET/PUT /api/supplier-orders/{id}` | Заказы поставщикам |
| `POST /api/supplier-orders/{id}/cancel\|create-receipt\|create-work-order` | Операции по заказам поставщикам |
| `POST /api/customer-orders` · `GET /api/customer-orders` · `GET/PUT /api/customer-orders/{id}` | Клиентские заказы |
| `POST /api/customer-orders/{id}/cancel\|create-sale\|create-work-order` | Операции по клиентским заказам |

### Бренды (brands-service)
| HTTP | Описание |
|---|---|
| `POST /api/brands` · `GET /api/brands` · `GET/PUT/DELETE /api/brands/{id}` | Бренды |
| `GET /api/brand-labor-rates` · `PUT /api/brand-labor-rates` | Нормо-часы по брендам |
| `DELETE /api/brand-labor-rates/{id}` · `GET /api/brand-labor-rates/resolve` | Удаление/разрешение ставок |

### Дилерские точки, юр. лица, склады (dealer-points-service)
| HTTP | Описание |
|---|---|
| `POST /api/dealer-points` · `GET /api/dealer-points` · `GET/PUT/DELETE /api/dealer-points/{id}` | Точки |
| `POST /api/dealer-points/{id}/legal-entities` · `DELETE /api/dealer-points/{dealer_point_id}/legal-entities/{legal_entity_id}` · `GET /api/dealer-points/{dealer_point_id}/legal-entities` | Привязка юр. лиц |
| `POST /api/legal-entities` · `GET /api/legal-entities` · `GET/PUT/DELETE /api/legal-entities/{id}` | Юр. лица |
| `POST /api/warehouses` · `GET /api/warehouses` · `GET/PUT/DELETE /api/warehouses/{id}` | Склады |

### Заказ-наряды (workorders-service)
| HTTP | Описание |
|---|---|
| `POST /api/work-orders` · `GET /api/work-orders` · `GET/PUT/DELETE /api/work-orders/{id}` | Заказ-наряды |
| `POST /api/work-orders/{id}/move-parts-to-work` | Выдача запчастей в работу |

### Работы СТО (works-service)
| HTTP | Описание |
|---|---|
| `POST /api/works` · `GET /api/works` · `GET/PUT/DELETE /api/works/{id}` | Работы |
| `POST /api/works/folders` · `GET /api/works/folders` · `GET/PUT/DELETE /api/works/folders/{id}` | Папки работ |

### Сотрудники (employees-service)
| HTTP | Описание |
|---|---|
| `POST /api/employees` · `GET /api/employees` · `GET/PUT/DELETE /api/employees/{id}` | Сотрудники |

### Запись на ремонт (appointments-service)
| HTTP | Описание |
|---|---|
| `GET /api/repair-appointment-slots` | Слоты |
| `POST /api/repair-appointments` · `GET /api/repair-appointments` · `GET/PUT /api/repair-appointments/{id}` | Записи |
| `POST /api/repair-appointments/{id}/cancel\|create-work-order` | Отмена/конвертация в заказ-наряд |

### Отзывы (employee-reviews-service)
| HTTP | Описание |
|---|---|
| `GET /api/clients/{client_id}/reviews` | Отзывы клиента |
| `GET /api/reviews` · `GET /api/reviews/{id}` | Список/отзыв |
| `GET /api/reviews/stats` | Статистика отзывов |

### Статистика
| HTTP | Описание |
|---|---|
| `GET /api/stats/employee/overview` | Сводка для сотрудников |
| `GET /api/stats/client/overview` | Сводка по клиентской зоне |
