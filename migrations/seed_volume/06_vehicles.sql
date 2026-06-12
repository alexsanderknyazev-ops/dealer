-- vehicles: 120 автомобилей
SET search_path TO vehicles, public;

INSERT INTO vehicles (
  id, vin, make, model, year, mileage_km, price, status, color, notes,
  brand_id, dealer_point_id, legal_entity_id, warehouse_id,
  created_at, updated_at
)
SELECT
  ('90000006-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  'VOLVIN' || lpad(g.n::text, 11, '0'),
  'VolumeMake-' || (1 + (g.n % 10)),
  'Model-' || g.n,
  2015 + (g.n % 11),
  (g.n * 1370) % 250000,
  500000 + (g.n % 50) * 25000,
  (ARRAY['available','available','available','reserved','sold'])[1 + (g.n % 5)],
  (ARRAY['белый','чёрный','серый','синий','красный'])[1 + (g.n % 5)],
  'Volume vehicle #' || g.n,
  ('90000004-0000-4000-8000-' || lpad(to_hex(1 + (g.n % 80)), 12, '0'))::uuid,
  ('90000003-0000-4000-8000-' || lpad(to_hex(1 + (g.n % 60)), 12, '0'))::uuid,
  ('90000013-0000-4000-8000-' || lpad(to_hex(1 + (g.n % 60)), 12, '0'))::uuid,
  ('90000023-0000-4000-8000-' || lpad(to_hex((1 + (g.n % 60)) * 2 - 1), 12, '0'))::uuid,
  now() - (g.n || ' days')::interval,
  now()
FROM generate_series(1, 120) AS g(n)
ON CONFLICT (vin) DO UPDATE SET
  make = EXCLUDED.make,
  model = EXCLUDED.model,
  price = EXCLUDED.price,
  status = EXCLUDED.status,
  updated_at = now();
