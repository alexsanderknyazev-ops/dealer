# client-public-gateway — детальные тест-кейсы

**HTTP:** `:8091` — unauthenticated routes only.

---

## TC-CPG-D001 — Register full flow

См. [`../client-registration/detailed-test-cases.md`](../client-registration/detailed-test-cases.md) TC-CR-D001

Gateway-specific checks:
- Response приходит с :8091 (не :8093)
- CORS headers present

---

## TC-CPG-D002 — Login / refresh / logout

См. [`../client-auth/detailed-test-cases.md`](../client-auth/detailed-test-cases.md)

Redis prefix: `client-auth:refresh:`

---

## TC-CPG-D003 — Protected routes NOT exposed

```bash
curl -s -w "\nHTTP:%{http_code}\n" http://127.0.0.1:8091/api/client/profile
curl -s -w "\nHTTP:%{http_code}\n" http://127.0.0.1:8091/api/me
```
- **404** (not registered on public gateway) — profile/me только на :8093

---

## TC-CPG-D004 — Employee login on client gateway (P2)

```bash
curl -s -X POST http://127.0.0.1:8091/api/login \
  -d '{"email":"<employee>","password":"..."}'
```
- Может **200** если email существует в clientauth (negative test: use employee-only email)
- Employee email в `auth.users` но не в `clientauth.users` → **401**

БД: no cross-contamination between auth schemas.
