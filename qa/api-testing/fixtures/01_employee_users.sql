-- QA fixtures: employee users (auth.users)
-- Password for all: Test1234!
-- bcrypt: $2a$10$b4bnj9tAH5g7FsPB3ztaD.12eTlbg1euqCvNi5TwPcteB8wthnQuy

SET search_path TO auth, public;

INSERT INTO users (id, email, password_hash, name, phone, role, created_at, updated_at)
VALUES
  (
    'a1100001-0000-4000-8000-000000000001',
    'qa.admin@test.local',
    '$2a$10$b4bnj9tAH5g7FsPB3ztaD.12eTlbg1euqCvNi5TwPcteB8wthnQuy',
    'QA Admin',
    '+79001000001',
    'admin',
    now(), now()
  ),
  (
    'a1100001-0000-4000-8000-000000000002',
    'qa.sales@test.local',
    '$2a$10$b4bnj9tAH5g7FsPB3ztaD.12eTlbg1euqCvNi5TwPcteB8wthnQuy',
    'QA Sales',
    '+79001000002',
    'sales',
    now(), now()
  ),
  (
    'a1100001-0000-4000-8000-000000000003',
    'qa.master@test.local',
    '$2a$10$b4bnj9tAH5g7FsPB3ztaD.12eTlbg1euqCvNi5TwPcteB8wthnQuy',
    'QA Master',
    '+79001000003',
    'master',
    now(), now()
  ),
  (
    'a1100001-0000-4000-8000-000000000004',
    'qa.parts@test.local',
    '$2a$10$b4bnj9tAH5g7FsPB3ztaD.12eTlbg1euqCvNi5TwPcteB8wthnQuy',
    'QA Parts Manager',
    '+79001000004',
    'parts_manager',
    now(), now()
  ),
  (
    'a1100001-0000-4000-8000-000000000005',
    'qa.storekeeper@test.local',
    '$2a$10$b4bnj9tAH5g7FsPB3ztaD.12eTlbg1euqCvNi5TwPcteB8wthnQuy',
    'QA Storekeeper',
    '+79001000005',
    'storekeeper',
    now(), now()
  )
ON CONFLICT (email) DO UPDATE SET
  password_hash = EXCLUDED.password_hash,
  name          = EXCLUDED.name,
  phone         = EXCLUDED.phone,
  role          = EXCLUDED.role,
  updated_at    = now();
