-- clients: 100 профилей + привязка к авто volume
SET search_path TO clients, public;

INSERT INTO clients (id, user_id, email, full_name, phone, created_at, updated_at)
SELECT
  ('9000000c-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('9000000b-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  'vol.client' || g.n || '@test.dealer.local',
  'Volume Client ' || g.n,
  '+7905' || lpad(g.n::text, 7, '0'),
  now() - (g.n || ' days')::interval,
  now()
FROM generate_series(1, 100) AS g(n)
ON CONFLICT (user_id) DO UPDATE SET
  email = EXCLUDED.email,
  full_name = EXCLUDED.full_name,
  updated_at = now();

INSERT INTO client_vehicles (id, client_id, vehicle_id, vin, make, model, year, added_at)
SELECT
  ('9000002c-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('9000000c-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  ('90000006-0000-4000-8000-' || lpad(to_hex(g.n), 12, '0'))::uuid,
  'VOLVIN' || lpad(g.n::text, 11, '0'),
  'VolumeMake-' || (1 + (g.n % 10)),
  'Model-' || g.n,
  2015 + (g.n % 11),
  now() - (g.n || ' days')::interval
FROM generate_series(1, 100) AS g(n)
ON CONFLICT (vin) DO UPDATE SET
  client_id = EXCLUDED.client_id,
  vehicle_id = EXCLUDED.vehicle_id;
