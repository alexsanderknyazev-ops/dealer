-- Dealer QA: шаблоны SQL для проверки после API-вызовов
-- Usage: psql "$POSTGRES_DSN" -v id="'<UUID>'" -f db-queries.sql
-- Или подставить UUID вручную.

\echo '--- auth.users (employee) ---'
-- SELECT id, email, name, role, created_at FROM auth.users WHERE email = '<EMAIL>';

\echo '--- customers ---'
-- SELECT * FROM customers.customers WHERE id = '<CUSTOMER_ID>';
-- SELECT COUNT(*) AS cnt FROM customers.customers;

\echo '--- vehicles ---'
-- SELECT id, vin, model, year, status, brand_id, dealer_point_id FROM vehicles.vehicles WHERE id = '<VEHICLE_ID>';
-- SELECT * FROM vehicles.vehicles WHERE vin = '<VIN>';

\echo '--- deals ---'
-- SELECT id, customer_id, vehicle_id, amount, stage, assigned_to FROM deals.deals WHERE id = '<DEAL_ID>';
-- SELECT COUNT(*) FROM deals.deals WHERE stage = 'completed';

\echo '--- brands ---'
-- SELECT * FROM brands.brands WHERE id = '<BRAND_ID>';

\echo '--- dealerpoints ---'
-- SELECT * FROM dealerpoints.dealer_points WHERE id = '<DP_ID>';
-- SELECT * FROM dealerpoints.warehouses WHERE id = '<WH_ID>';
-- SELECT * FROM dealerpoints.dealer_point_legal_entities WHERE dealer_point_id = '<DP_ID>';

\echo '--- parts ---'
-- SELECT id, sku, name, quantity, brand_id, warehouse_id FROM parts.parts WHERE id = '<PART_ID>';
-- SELECT part_id, warehouse_id, quantity FROM parts.part_stock WHERE part_id = '<PART_ID>';
-- SELECT * FROM parts.movement_documents WHERE id = '<DOC_ID>';
-- SELECT * FROM parts.movement_document_lines WHERE document_id = '<DOC_ID>' ORDER BY sort_order;
-- SELECT * FROM parts.stock_movements WHERE movement_document_id = '<DOC_ID>';

\echo '--- workorders ---'
-- SELECT id, order_number, status, parts_issued, movement_document_id, movement_document_status
-- FROM workorders.work_orders WHERE id = '<WO_ID>';
-- SELECT work_id, executor_id, description FROM workorders.work_order_labor WHERE work_order_id = '<WO_ID>';
-- SELECT * FROM workorders.work_order_parts WHERE work_order_id = '<WO_ID>';

\echo '--- works (STO catalog) ---'
-- SELECT id, code, name, labor_hours, unit_price FROM works.works WHERE id = '<WORK_ID>';
-- SELECT COUNT(*) FROM works.works WHERE code LIKE 'LAB-QA-%';

\echo '--- employees ---'
-- SELECT id, user_id, full_name, position, active FROM employees.employees WHERE id = '<EMP_ID>';
-- SELECT e.* FROM employees.employees e JOIN auth.users u ON u.id = e.user_id WHERE u.email = 'qa.master@test.local';

\echo '--- clients (B2C) ---'
-- SELECT * FROM clients.clients WHERE id = '<CLIENT_ID>';
-- SELECT * FROM clients.client_vehicles WHERE client_id = '<CLIENT_ID>';

\echo '--- clientauth ---'
-- SELECT id, email, full_name FROM clientauth.users WHERE id = '<USER_ID>';

\echo '--- reviews (client) ---'
-- SELECT * FROM reviews.reviews WHERE id = '<REVIEW_ID>';
-- SELECT COUNT(*) FROM reviews.reviews WHERE client_id = '<CLIENT_ID>';

\echo '--- employee_reviews ---'
-- SELECT review_id, client_email, rating, vehicle_vin FROM employee_reviews.reviews WHERE review_id = '<REVIEW_ID>';

\echo '--- employee_statistics ---'
-- SELECT * FROM employee_statistics.deal_events WHERE deal_id = '<DEAL_ID>';

\echo '--- client_statistics ---'
-- SELECT * FROM client_statistics.client_registration_events WHERE user_id = '<USER_ID>';
-- SELECT * FROM client_statistics.review_events WHERE review_id = '<REVIEW_ID>';

-- Пример: проверка цепочки WO + movement
-- SELECT wo.id, wo.movement_document_id, wo.movement_document_status, md.status AS doc_status
-- FROM workorders.work_orders wo
-- LEFT JOIN parts.movement_documents md ON md.id = wo.movement_document_id
-- WHERE wo.id = '<WO_ID>';
