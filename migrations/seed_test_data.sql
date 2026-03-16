-- Тестовые клиенты и автомобили (идемпотентно: вставляем только если таблицы пусты)

-- Тестовые клиенты
INSERT INTO customers (id, name, email, phone, customer_type, inn, address, notes, created_at, updated_at)
SELECT v.id, v.name, v.email, v.phone, v.customer_type, v.inn, v.address, v.notes, v.created_at, v.updated_at
FROM (VALUES
  (gen_random_uuid(), 'Иван Петров', 'ivan.petrov@example.com', '+7 900 111-22-33', 'individual', '', 'г. Москва, ул. Ленина 1', 'Постоянный клиент', now(), now()),
  (gen_random_uuid(), 'ООО Рога и копыта', 'office@roga-kopyta.ru', '+7 495 123-45-67', 'legal', '7707123456', 'г. Москва, ул. Тверская 10', 'Корпоративный клиент', now(), now()),
  (gen_random_uuid(), 'Мария Сидорова', 'maria.s@mail.ru', '+7 916 555-44-33', 'individual', '', 'г. Санкт-Петербург, Невский пр. 5', '', now(), now()),
  (gen_random_uuid(), 'Алексей Козлов', 'a.kozlov@gmail.com', '+7 903 777-88-99', 'individual', '', 'г. Казань, ул. Баумана 20', 'Интересуется SUV', now(), now()),
  (gen_random_uuid(), 'ЗАО Автопарк', 'zakaz@avtopark.ru', '+7 812 333-44-55', 'legal', '7816123456', 'г. Санкт-Петербург, пр. Энергетиков 15', 'Флотские закупки', now(), now())
) AS v(id, name, email, phone, customer_type, inn, address, notes, created_at, updated_at)
WHERE NOT EXISTS (SELECT 1 FROM customers LIMIT 1);

-- Тестовые автомобили
INSERT INTO vehicles (id, vin, make, model, year, mileage_km, price, status, color, notes, created_at, updated_at)
SELECT v.id, v.vin, v.make, v.model, v.year, v.mileage_km, v.price, v.status, v.color, v.notes, v.created_at, v.updated_at
FROM (VALUES
  (gen_random_uuid(), 'WVWZZZ3CZWE123456', 'Volkswagen', 'Polo', 2022, 15000, 1450000, 'available', 'Белый', 'Комплектация Comfort', now(), now()),
  (gen_random_uuid(), 'Z94CB41AAGR323020', 'Hyundai', 'Solaris', 2023, 8000, 1320000, 'available', 'Серый', 'Новый в наличии', now(), now()),
  (gen_random_uuid(), 'XW8ZZZ5NZLG012345', 'Skoda', 'Octavia', 2021, 42000, 1650000, 'available', 'Чёрный', 'Полная комплектация', now(), now()),
  (gen_random_uuid(), 'JN1TANT31U0123456', 'Nissan', 'Qashqai', 2022, 28000, 2450000, 'reserved', 'Синий', 'Зарезервирован до 20.03', now(), now()),
  (gen_random_uuid(), 'WF0XXXTTGCHB12345', 'Kia', 'K5', 2023, 5000, 2890000, 'available', 'Красный', 'Премиум', now(), now()),
  (gen_random_uuid(), 'WVWZZZAUZLP123789', 'Volkswagen', 'Tiguan', 2020, 55000, 2200000, 'sold', 'Белый', 'Продан 01.03', now(), now()),
  (gen_random_uuid(), 'TMBJJ9NE2K0123456', 'Skoda', 'Kodiaq', 2022, 32000, 3100000, 'available', 'Зелёный', '7 мест', now(), now()),
  (gen_random_uuid(), 'KMHGH4JH2EU123456', 'Hyundai', 'Creta', 2023, 12000, 1890000, 'available', 'Белый', '', now(), now())
) AS v(id, vin, make, model, year, mileage_km, price, status, color, notes, created_at, updated_at)
WHERE NOT EXISTS (SELECT 1 FROM vehicles LIMIT 1);

-- Тестовые запасные части
INSERT INTO parts (id, sku, name, category, quantity, unit, price, location, notes, created_at, updated_at)
SELECT p.id, p.sku, p.name, p.category, p.quantity, p.unit, p.price, p.location, p.notes, p.created_at, p.updated_at
FROM (VALUES
  (gen_random_uuid(), 'FLT-OIL-001', 'Масляный фильтр VW Polo/Skoda Rapid', 'Фильтры', 24, 'шт', 850, 'Склад А, полка 12', 'OE 06A115561B', now(), now()),
  (gen_random_uuid(), 'FLT-OIL-002', 'Масляный фильтр Hyundai/Kia 1.6', 'Фильтры', 18, 'шт', 720, 'Склад А, полка 12', 'Под линейку Gamma', now(), now()),
  (gen_random_uuid(), 'FLT-AIR-001', 'Воздушный фильтр салона VW Group', 'Фильтры', 30, 'шт', 1200, 'Склад А, полка 14', 'Универсальный угольный', now(), now()),
  (gen_random_uuid(), 'FLT-FUEL-001', 'Топливный фильтр дизель VW 2.0 TDI', 'Фильтры', 8, 'шт', 2100, 'Склад А, полка 15', 'Только для дизельных', now(), now()),
  (gen_random_uuid(), 'BRK-PAD-F-001', 'Колодки передние VW Polo (комплект)', 'Тормоза', 12, 'комплект', 4200, 'Склад Б, полка 3', 'Передняя ось, 4 колодки', now(), now()),
  (gen_random_uuid(), 'BRK-PAD-R-001', 'Колодки задние VW Polo (комплект)', 'Тормоза', 10, 'комплект', 3800, 'Склад Б, полка 3', 'Задняя ось', now(), now()),
  (gen_random_uuid(), 'BRK-DISC-001', 'Диск тормозной передний 288mm VW/Skoda', 'Тормоза', 6, 'шт', 5200, 'Склад Б, полка 5', 'Парная замена', now(), now()),
  (gen_random_uuid(), 'OIL-5W30-001', 'Масло моторное 5W-30 синтетика 5л', 'Масла', 48, 'л', 450, 'Склад В, полка 1', 'Допуск VW 502 00', now(), now()),
  (gen_random_uuid(), 'OIL-0W20-001', 'Масло моторное 0W-20 5л', 'Масла', 36, 'л', 520, 'Склад В, полка 1', 'Hyundai/Kia рекомендовано', now(), now()),
  (gen_random_uuid(), 'OIL-ATF-001', 'Трансмиссия ATF Hyundai/Kia 4л', 'Масла', 20, 'л', 680, 'Склад В, полка 2', 'Для АКПП', now(), now()),
  (gen_random_uuid(), 'WIP-BLADE-001', 'Щётка стеклоочистителя 550mm', 'Расходники', 40, 'шт', 380, 'Склад Г, полка 7', 'Универсальная левая', now(), now()),
  (gen_random_uuid(), 'WIP-BLADE-002', 'Щётка стеклоочистителя 450mm', 'Расходники', 35, 'шт', 320, 'Склад Г, полка 7', 'Универсальная правая', now(), now()),
  (gen_random_uuid(), 'BULB-H7-001', 'Лампа H7 55W halogen', 'Расходники', 60, 'шт', 290, 'Склад Г, полка 9', 'Ближний свет', now(), now()),
  (gen_random_uuid(), 'BULB-D1S-001', 'Лампа D1S ксенон', 'Расходники', 12, 'шт', 3200, 'Склад Г, полка 10', 'Ксенон оригинал', now(), now()),
  (gen_random_uuid(), 'BATT-62AH-001', 'Аккумулятор 62 А·ч 12V', 'Расходники', 8, 'шт', 6500, 'Склад Г, зона тяжёлых', 'С обратной полярностью', now(), now())
) AS p(id, sku, name, category, quantity, unit, price, location, notes, created_at, updated_at)
WHERE NOT EXISTS (SELECT 1 FROM parts LIMIT 1);
