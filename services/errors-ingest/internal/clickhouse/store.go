package clickhouse

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	ch "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/google/uuid"

	"github.com/dealer/dealer/pkg/errorevent"
)

// Store пишет события в ClickHouse.
type Store struct {
	conn     ch.Conn
	database string
}

// Open подключается к ClickHouse и создаёт схему при необходимости.
func Open(ctx context.Context, addr, database, user, password string) (*Store, error) {
	conn, err := ch.Open(&ch.Options{
		Addr: []string{addr},
		Auth: ch.Auth{
			Database: database,
			Username: user,
			Password: password,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("clickhouse open: %w", err)
	}
	if err := conn.Ping(ctx); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("clickhouse ping: %w", err)
	}
	s := &Store{conn: conn, database: database}
	if err := s.EnsureSchema(ctx); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return s, nil
}

// EnsureSchema создаёт БД и таблицу error_events.
func (s *Store) EnsureSchema(ctx context.Context) error {
	if err := s.conn.Exec(ctx, `CREATE DATABASE IF NOT EXISTS analytics`); err != nil {
		return fmt.Errorf("create database: %w", err)
	}
	const tableSQL = `
CREATE TABLE IF NOT EXISTS analytics.error_events
(
    occurred_at DateTime64(3, 'UTC'),
    event_id UUID,
    source LowCardinality(String),
    kind LowCardinality(String),
    severity LowCardinality(String),
    message String,
    error_code LowCardinality(String),
    http_status UInt16,
    grpc_code LowCardinality(String),
    trace_id String,
    request_id String,
    user_id String,
    route String,
    service LowCardinality(String),
    environment LowCardinality(String),
    context String,
    stack String,
    client String,
    ingested_at DateTime64(3, 'UTC') DEFAULT now64(3)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(occurred_at)
ORDER BY (service, occurred_at)
TTL occurred_at + INTERVAL 90 DAY`
	if err := s.conn.Exec(ctx, tableSQL); err != nil {
		return fmt.Errorf("create table: %w", err)
	}
	return nil
}

// InsertEvents batch-insert в analytics.error_events.
func (s *Store) InsertEvents(ctx context.Context, events []errorevent.Event) error {
	if len(events) == 0 {
		return nil
	}
	batch, err := s.conn.PrepareBatch(ctx, `INSERT INTO analytics.error_events (
		occurred_at, event_id, source, kind, severity, message, error_code,
		http_status, grpc_code, trace_id, request_id, user_id, route,
		service, environment, context, stack, client
	)`)
	if err != nil {
		return fmt.Errorf("prepare batch: %w", err)
	}
	for _, ev := range events {
		occurredAt, err := parseOccurredAt(ev.OccurredAt)
		if err != nil {
			return err
		}
		eventID, err := parseUUIDString(ev.EventID)
		if err != nil {
			return err
		}
		ctxJSON, _ := json.Marshal(ev.Context)
		clientJSON, _ := json.Marshal(ev.Client)
		if err := batch.Append(
			occurredAt,
			eventID,
			ev.Source,
			ev.Kind,
			ev.Severity,
			ev.Message,
			ev.ErrorCode,
			uint16(ev.HTTPStatus),
			ev.GRPCCode,
			ev.TraceID,
			ev.RequestID,
			ev.UserID,
			ev.Route,
			ev.Service,
			ev.Environment,
			string(ctxJSON),
			ev.Stack,
			string(clientJSON),
		); err != nil {
			return fmt.Errorf("batch append: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("batch send: %w", err)
	}
	return nil
}

// Close закрывает соединение.
func (s *Store) Close() error {
	if s == nil || s.conn == nil {
		return nil
	}
	return s.conn.Close()
}

func parseOccurredAt(raw string) (time.Time, error) {
	if raw == "" {
		return time.Now().UTC(), nil
	}
	t, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		t, err = time.Parse(time.RFC3339, raw)
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("parse occurred_at %q: %w", raw, err)
	}
	return t.UTC(), nil
}

func parseUUIDString(raw string) (uuid.UUID, error) {
	if raw == "" {
		return uuid.UUID{}, fmt.Errorf("empty event_id")
	}
	return uuid.Parse(raw)
}
