// Package configenv — чтение переменных окружения для микросервисов.
package configenv

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// String возвращает env или def, если переменная пуста.
func String(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// Int парсит env как int или возвращает def при пустом/невалидном значении.
func Int(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

// Duration парсит env как time.Duration или возвращает def при ошибке.
func Duration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

// SplitCSV разбивает строку (или значение env по key) на непустые элементы.
func SplitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// Brokers читает KAFKA_BROKERS (CSV) с дефолтом.
func Brokers(key, def string) []string {
	return SplitCSV(String(key, def))
}
