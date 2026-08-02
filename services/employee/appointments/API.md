# appointments-service — API

gRPC-сервис `appointments.v1.RepairAppointmentsService`. Доступ защищён JWT.

## Endpoints

| gRPC | HTTP | Описание |
|---|---|---|
| `ListRepairAppointmentSlots` | `GET /api/repair-appointment-slots` | Свободные слоты на дату |
| `CreateRepairAppointment` | `POST /api/repair-appointments` | Создание записи на ремонт |
| `GetRepairAppointment` | `GET /api/repair-appointments/{id}` | Запись по id |
| `UpdateRepairAppointment` | `PUT /api/repair-appointments/{id}` | Обновление записи |
| `ListRepairAppointments` | `GET /api/repair-appointments` | Список записей |
| `CancelRepairAppointment` | `POST /api/repair-appointments/{id}/cancel` | Отмена записи |
| `CreateWorkOrderFromRepairAppointment` | `POST /api/repair-appointments/{id}/create-work-order` | Создать заказ-наряд из записи |

## Сообщения

### RepairAppointment (модель)
`id`, `appointment_number`, `customer_id`, `customer_name`, `vehicle_id`, `vehicle_vin`, `vehicle_label`, `dealer_point_id`, `warehouse_id`, `scheduled_start`, `scheduled_end`, `status`, `work_order_id`, `work_order_number`, `complaint`, `notes`, `created_by`, `created_at`, `updated_at`, `labor[]`, `parts[]`

### RepairAppointmentLabor / Part
Позиции работ (`work_id`, `description`, `quantity`, `unit_price`, `sort_order`, `work_code`, `work_name`) и запчастей (`part_id`, `warehouse_id`, `quantity`, `unit_price`, `notes`, `sort_order`, `part_name`, `part_sku`, `warehouse_name`)

### RepairAppointmentSlot
`start_at`, `end_at`, `available`, `label`

### ListRepairAppointmentSlots
Request: `date`, `dealer_point_id`
Response: `slots[]`

### CreateRepairAppointment
Request: `customer_id`, `vehicle_id`, `dealer_point_id`, `warehouse_id`, `scheduled_start`, `scheduled_end`, `complaint`, `notes`, `created_by`, `labor[]`, `parts[]`
Response: `appointment`

### GetRepairAppointment
Request: `id`
Response: `appointment`

### UpdateRepairAppointment
Request: `id`, optional поля, `labor[]`, `parts[]`, `replace_lines`
Response: `appointment`

### ListRepairAppointments
Request: `limit`, `offset`, `status`, `date_from`, `date_to`
Response: `appointments[]`, `total`

### CancelRepairAppointment
Request: `id`
Response: `appointment`

### CreateWorkOrderFromRepairAppointment
Request: `id`
Response: `work_order_id`, `work_order_number`, `appointment`
