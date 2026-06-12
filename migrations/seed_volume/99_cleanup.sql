-- Удаление volume seed (UUID 9000000*)
-- Порядок: дочерние таблицы → родительские

DELETE FROM appointments.repair_appointment_parts WHERE appointment_id::text LIKE '90000011-%' OR id::text LIKE '90000031-%';
DELETE FROM appointments.repair_appointment_labor WHERE appointment_id::text LIKE '90000011-%' OR id::text LIKE '90000021-%';
DELETE FROM appointments.repair_appointments WHERE id::text LIKE '90000011-%';

DELETE FROM client_statistics.review_events WHERE id::text LIKE '90000020-%' OR review_id::text LIKE '9000000d-%';
DELETE FROM client_statistics.client_registration_events WHERE id::text LIKE '90000010-%';

DELETE FROM employee_statistics.deal_events WHERE id::text LIKE '9000000f-%';

DELETE FROM employee_reviews.reviews WHERE id::text LIKE '9000000e-%';

DELETE FROM reviews.reviews WHERE id::text LIKE '9000000d-%';
DELETE FROM reviews.review_invitations WHERE id::text LIKE '9000002d-%';

DELETE FROM clients.client_notifications WHERE id::text LIKE '9000003c-%';
DELETE FROM clients.client_vehicles WHERE id::text LIKE '9000002c-%' OR client_id::text LIKE '9000000c-%';
DELETE FROM clients.clients WHERE id::text LIKE '9000000c-%';

DELETE FROM clientauth.users WHERE id::text LIKE '9000000b-%';

DELETE FROM workorders.work_order_parts WHERE work_order_id::text LIKE '9000000a-%';
DELETE FROM workorders.work_order_labor WHERE id::text LIKE '9000001a-%' OR work_order_id::text LIKE '9000000a-%';
DELETE FROM workorders.work_orders WHERE id::text LIKE '9000000a-%';

DELETE FROM deals.deals WHERE id::text LIKE '90000009-%';

DELETE FROM works.works WHERE id::text LIKE '90000008-%';

DELETE FROM parts.customer_order_lines WHERE order_id::text LIKE '90000037-%' OR id::text LIKE '90000038-%' OR id::text LIKE '90000039-%';
DELETE FROM parts.supplier_order_lines WHERE order_id::text LIKE '90000027-%' OR id::text LIKE '90000028-%' OR id::text LIKE '90000029-%';
DELETE FROM parts.customer_orders WHERE id::text LIKE '90000037-%';
DELETE FROM parts.supplier_orders WHERE id::text LIKE '90000027-%';
DELETE FROM parts.movement_document_lines WHERE document_id::text LIKE '90000047-%';
DELETE FROM parts.movement_documents WHERE id::text LIKE '90000047-%';
DELETE FROM parts.stock_movements WHERE id::text LIKE '90000057-%';
DELETE FROM parts.part_stock WHERE part_id::text LIKE '90000007-%';
DELETE FROM parts.parts WHERE id::text LIKE '90000007-%';
DELETE FROM parts.suppliers WHERE id::text LIKE '90000017-%';

DELETE FROM vehicles.vehicles WHERE id::text LIKE '90000006-%' OR vin LIKE 'VOLVIN%';

DELETE FROM customers.customers WHERE id::text LIKE '90000005-%';

DELETE FROM brands.brand_labor_rates WHERE id::text LIKE '90000014-%';
DELETE FROM brands.brands WHERE id::text LIKE '90000004-%';

DELETE FROM dealerpoints.warehouses WHERE id::text LIKE '90000023-%';
DELETE FROM dealerpoints.dealer_point_legal_entities WHERE dealer_point_id::text LIKE '90000003-%';
DELETE FROM dealerpoints.legal_entities WHERE id::text LIKE '90000013-%';
DELETE FROM dealerpoints.dealer_points WHERE id::text LIKE '90000003-%';

DELETE FROM employees.employees WHERE id::text LIKE '90000002-%';

DELETE FROM auth.users WHERE id::text LIKE '90000001-%';
