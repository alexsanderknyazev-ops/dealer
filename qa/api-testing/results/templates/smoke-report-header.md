# Smoke report — {{RUN_ID}}

| Field | Value |
|-------|-------|
| Run ID | `{{RUN_ID}}` |
| Timestamp (UTC) | {{TIMESTAMP}} |
| Executor | {{EXECUTOR}} |
| Git commit | {{GIT_COMMIT}} |
| Employee API | {{EMPLOYEE_API}} |
| Client Public | {{CLIENT_PUBLIC}} |
| Client Protected | {{CLIENT_PROTECTED}} |
| **Passed** | **{{PASS}}** |
| **Failed** | **{{FAIL}}** |
| **Skipped** | **{{SKIP}}** |
| Pass rate | {{PASS_RATE}}% |

## Environment checklist

- [ ] docker compose up
- [ ] make migrate
- [ ] make full-seed
- [ ] fixtures/apply-fixtures.sh

## Results

| Auto ID | Test | Status | Expected | Actual | Notes |
|---------|------|--------|----------|--------|-------|
