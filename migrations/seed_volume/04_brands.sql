-- brands: 80 брендов
SET search_path TO brands, public;

INSERT INTO brands (id, name, created_at, updated_at)
SELECT
  ('90000004-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  'VolumeBrand-' || g.n,
  now() - (g.n || ' days')::interval,
  now()
FROM generate_series(1, 80) AS g(n)
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  updated_at = now();

INSERT INTO brand_labor_rates (id, brand_id, dealer_point_id, warranty_hour_price, commercial_hour_price, created_at, updated_at)
SELECT
  ('90000014-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('90000004-0000-4000-8000-' || lpad(to_hex(1 + (g.n % 80)), 12, '0'))::uuid,
  COALESCE(
    ('90000003-0000-4000-8000-' || lpad(to_hex(1 + (g.n % 60)), 12, '0'))::uuid,
    '10000000-0000-4000-8000-000000000001'::uuid
  ),
  1200 + (g.n % 15) * 50,
  1800 + (g.n % 20) * 75,
  now(), now()
FROM generate_series(1, 80) AS g(n)
ON CONFLICT (brand_id, dealer_point_id) DO NOTHING;
