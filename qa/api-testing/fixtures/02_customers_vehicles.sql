-- QA fixtures: customers + vehicles
-- Requires: seed_dealer_brands (brand, dealer_point, warehouse)

SET search_path TO customers, vehicles, public;

-- Customers
INSERT INTO customers (id, name, email, phone, customer_type, inn, address, notes, created_at, updated_at)
VALUES
  (
    'a2200001-0000-4000-8000-000000000001',
    'QA Customer One',
    'qa.customer1@test.local',
    '+79002000001',
    'individual',
    '',
    'QA address 1',
    'Fixture customer for deals/WO',
    now(), now()
  ),
  (
    'a2200001-0000-4000-8000-000000000002',
    'QA Customer Two',
    'qa.customer2@test.local',
    '+79002000002',
    'legal',
    '500100732259',
    'QA address 2',
    'Fixture legal customer',
    now(), now()
  )
ON CONFLICT (id) DO UPDATE SET
  name       = EXCLUDED.name,
  email      = EXCLUDED.email,
  phone      = EXCLUDED.phone,
  updated_at = now();

-- Vehicles (fixed VIN for client registration tests)
INSERT INTO vehicles (
  id, vin, make, model, year, mileage_km, price, status, color, notes,
  brand_id, dealer_point_id, legal_entity_id, warehouse_id,
  created_at, updated_at
)
VALUES
  (
    'a3300001-0000-4000-8000-000000000001',
    'QAVINCLIENT001',
    'Hyundai',
    'QA Solaris',
    2024,
    10000,
    1650000,
    'available',
    'Silver',
    'For client registration E2E',
    '40000000-0000-4000-8000-000000000003',
    '10000000-0000-4000-8000-000000000001',
    '20000000-0000-4000-8000-000000000001',
    '30000000-0000-4000-8000-000000000001',
    now(), now()
  ),
  (
    'a3300001-0000-4000-8000-000000000002',
    'QAVINCLIENT002',
    'Hyundai',
    'QA Creta',
    2023,
    25000,
    1890000,
    'available',
    'White',
    'Second vehicle for addVehicle API',
    '40000000-0000-4000-8000-000000000003',
    '10000000-0000-4000-8000-000000000001',
    '20000000-0000-4000-8000-000000000001',
    '30000000-0000-4000-8000-000000000001',
    now(), now()
  ),
  (
    'a3300001-0000-4000-8000-000000000003',
    'QAVINDEAL001',
    'Volkswagen',
    'QA Polo',
    2022,
    30000,
    1450000,
    'available',
    'Blue',
    'For deal E2E',
    '40000000-0000-4000-8000-000000000001',
    '10000000-0000-4000-8000-000000000001',
    '20000000-0000-4000-8000-000000000001',
    '30000000-0000-4000-8000-000000000001',
    now(), now()
  )
ON CONFLICT (vin) DO UPDATE SET
  model      = EXCLUDED.model,
  status     = EXCLUDED.status,
  brand_id   = EXCLUDED.brand_id,
  updated_at = now();
