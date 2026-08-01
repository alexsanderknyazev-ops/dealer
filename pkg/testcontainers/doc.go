// Package testcontainers предоставляет утилиты для интеграционных тестов
// сервисов на Testcontainers: поднятие реальных контейнеров Postgres (с
// применением всех миграций), Redis и Kafka, а также проверку доступности
// Docker.
//
// Пакет находится в root-модуле и переиспользуется сервисными модулями через
// replace (см. каждый services/*/go.mod).
package testcontainers
