# client-registration — тест-кейсы

**Сервис:** `client-registration-service`  
**gRPC:** `:50058`

| ID | P | Endpoint | Auth | Steps | Expected | Auto |
|----|---|----------|------|-------|----------|------|
| TC-CR-001 | P0 | POST /api/client/register | public gw | full payload + vin | 200, client, tokens | CR-001 |
| TC-CR-002 | P0 | GET /api/client/profile | protected gw | Bearer | 200, profile + vehicles[] | CR-002 |
| TC-CR-003 | P0 | GET /api/client/vehicles | protected gw | Bearer | 200, vehicles[] | CR-003 |
| TC-CR-004 | P1 | POST /api/client/vehicles | protected gw | vin второго авто | 200, vehicle linked | CR-004 |
| TC-CR-005 | P1 | POST /api/client/register | — | missing full_name | 400 | CR-005 |
| TC-CR-006 | P0 | gRPC GetVehicleByVIN | internal | vin при register | vehicle_id | manual |
| TC-CR-007 | P0 | gRPC IssueTokens | internal | после Kafka publish | tokens в ответе register | manual |
| TC-CR-008 | P0 | Kafka publish | — | register success | client.registration.v1 event | manual |

## Flow register

1. Validate VIN via vehicles gRPC  
2. Save client to Postgres  
3. Publish Kafka  
4. IssueTokens via client-auth (retry до 15×)
