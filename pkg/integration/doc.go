// Package integration содержит интеграционные тесты инфраструктуры проекта:
// PostgreSQL с миграциями, Redis и Kafka. Тесты поднимают реальные контейнеры
// через Testcontainers и активируются build-тегом `integration`.
//
// Запуск:
//
//	go test -tags=integration -timeout 30m ./pkg/integration/...
//
// Требуется работающий Docker. Без Docker тесты корректно пропускаются.
package integration
