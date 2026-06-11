-- QA fixtures cleanup — удаляет только QA namespace (UUID prefix a1100001, a2200001, …)
-- Порядок важен из-за FK / логических связей

SET search_path TO workorders, parts, deals, reviews, employee_reviews,
  client_statistics, employee_statistics, clients, clientauth, vehicles, customers,
  employees, works, auth, public;

-- Movement / stock (QA parts only)
DELETE FROM parts.stock_movements
WHERE part_id IN ('a4400001-0000-4000-8000-000000000001', 'a4400001-0000-4000-8000-000000000002');

DELETE FROM parts.movement_document_lines
WHERE document_id IN (
  SELECT id FROM parts.movement_documents
  WHERE reference_id = 'a6600001-0000-4000-8000-000000000001'
);

DELETE FROM parts.movement_documents
WHERE reference_id = 'a6600001-0000-4000-8000-000000000001';

-- Work orders
DELETE FROM workorders.work_order_parts
WHERE work_order_id = 'a6600001-0000-4000-8000-000000000001';

DELETE FROM workorders.work_order_labor
WHERE work_order_id = 'a6600001-0000-4000-8000-000000000001';

DELETE FROM workorders.work_orders
WHERE id = 'a6600001-0000-4000-8000-000000000001';

-- QA works catalog
DELETE FROM works.works WHERE id LIKE 'a8800001-%';

-- QA employees (fixed ids only; migration-seeded rows without prefix stay)
DELETE FROM employees.employees WHERE id LIKE 'a9900001-%';

-- Deals + stats events
DELETE FROM employee_statistics.deal_events
WHERE deal_id LIKE 'a5500001-%';

DELETE FROM deals.deals WHERE id LIKE 'a5500001-%';

-- Reviews
DELETE FROM employee_reviews.reviews WHERE review_id LIKE 'a9900001-%';
DELETE FROM client_statistics.review_events WHERE review_id LIKE 'a9900001-%';
DELETE FROM reviews.reviews WHERE id LIKE 'a9900001-%';

-- Client B2C
DELETE FROM client_statistics.client_registration_events WHERE user_id LIKE 'a7700001-%';
DELETE FROM clients.client_vehicles WHERE id LIKE 'a7700001-%';
DELETE FROM clients.clients WHERE id LIKE 'a7700001-%';
DELETE FROM clientauth.users WHERE id LIKE 'a7700001-%';

-- Parts stock + parts
DELETE FROM parts.part_stock
WHERE part_id IN ('a4400001-0000-4000-8000-000000000001', 'a4400001-0000-4000-8000-000000000002');

DELETE FROM parts.parts WHERE id LIKE 'a4400001-%';

-- Vehicles + customers
DELETE FROM vehicles.vehicles WHERE id LIKE 'a3300001-%';
DELETE FROM customers.customers WHERE id LIKE 'a2200001-%';

-- Employee users
DELETE FROM auth.users WHERE id LIKE 'a1100001-%';
