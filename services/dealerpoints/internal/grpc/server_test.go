package grpc

import (
	"context"
	"net"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/dealer/dealer/services/dealerpoints/internal/domain"
	"github.com/dealer/dealer/services/dealerpoints/internal/service"
	dealerpointsv1 "github.com/dealer/dealer/pkg/pb/dealerpoints/v1"
)

type gDP struct{ byID map[uuid.UUID]*domain.DealerPoint }

func (m *gDP) Create(_ context.Context, d *domain.DealerPoint) error {
	if m.byID == nil {
		m.byID = make(map[uuid.UUID]*domain.DealerPoint)
	}
	cp := *d
	m.byID[d.ID] = &cp
	return nil
}
func (m *gDP) GetByID(_ context.Context, id uuid.UUID) (*domain.DealerPoint, error) {
	d, ok := m.byID[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	cp := *d
	return &cp, nil
}
func (m *gDP) List(_ context.Context, _, _ int32, _ string) ([]*domain.DealerPoint, int32, error) {
	var out []*domain.DealerPoint
	for _, d := range m.byID {
		cp := *d
		out = append(out, &cp)
	}
	return out, int32(len(out)), nil
}
func (m *gDP) Update(_ context.Context, d *domain.DealerPoint) error {
	if _, ok := m.byID[d.ID]; !ok {
		return pgx.ErrNoRows
	}
	cp := *d
	m.byID[d.ID] = &cp
	return nil
}
func (m *gDP) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.byID, id)
	return nil
}

type gLE struct{ byID map[uuid.UUID]*domain.LegalEntity }

func (m *gLE) Create(_ context.Context, e *domain.LegalEntity) error {
	if m.byID == nil {
		m.byID = make(map[uuid.UUID]*domain.LegalEntity)
	}
	cp := *e
	m.byID[e.ID] = &cp
	return nil
}
func (m *gLE) GetByID(_ context.Context, id uuid.UUID) (*domain.LegalEntity, error) {
	e, ok := m.byID[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	cp := *e
	return &cp, nil
}
func (gLE) List(_ context.Context, _, _ int32, _ string) ([]*domain.LegalEntity, int32, error) { return nil, 0, nil }
func (gLE) ListByDealerPoint(_ context.Context, _ uuid.UUID, _, _ int32) ([]*domain.LegalEntity, int32, error) {
	return nil, 0, nil
}
func (m *gLE) Update(_ context.Context, e *domain.LegalEntity) error {
	if _, ok := m.byID[e.ID]; !ok {
		return pgx.ErrNoRows
	}
	cp := *e
	m.byID[e.ID] = &cp
	return nil
}
func (m *gLE) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.byID, id)
	return nil
}
func (gLE) LinkToDealerPoint(context.Context, uuid.UUID, uuid.UUID) error     { return nil }
func (gLE) UnlinkFromDealerPoint(context.Context, uuid.UUID, uuid.UUID) error { return nil }

type gWH struct{ byID map[uuid.UUID]*domain.Warehouse }

func (m *gWH) Create(_ context.Context, w *domain.Warehouse) error {
	if m.byID == nil {
		m.byID = make(map[uuid.UUID]*domain.Warehouse)
	}
	cp := *w
	m.byID[w.ID] = &cp
	return nil
}
func (m *gWH) GetByID(_ context.Context, id uuid.UUID) (*domain.Warehouse, error) {
	w, ok := m.byID[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	cp := *w
	return &cp, nil
}
func (gWH) List(_ context.Context, _, _ int32, _, _ *uuid.UUID, _ string) ([]*domain.Warehouse, int32, error) {
	return nil, 0, nil
}
func (m *gWH) Update(_ context.Context, w *domain.Warehouse) error {
	if _, ok := m.byID[w.ID]; !ok {
		return pgx.ErrNoRows
	}
	cp := *w
	m.byID[w.ID] = &cp
	return nil
}
func (m *gWH) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.byID, id)
	return nil
}

func dialDP(t *testing.T, srv *grpc.Server) dealerpointsv1.DealerPointsServiceClient {
	t.Helper()
	l := bufconn.Listen(1024 * 1024)
	go func() { _ = srv.Serve(l) }()
	t.Cleanup(func() { srv.Stop() })
	c, err := grpc.DialContext(context.Background(), "b",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return l.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return dealerpointsv1.NewDealerPointsServiceClient(c)
}

func TestDealerPointsGRPC_DealerPointFlow(t *testing.T) {
	svc := service.NewDealerPointsService(
		&gDP{byID: map[uuid.UUID]*domain.DealerPoint{}},
		&gLE{byID: map[uuid.UUID]*domain.LegalEntity{}},
		&gWH{byID: map[uuid.UUID]*domain.Warehouse{}},
	)
	s := grpc.NewServer()
	dealerpointsv1.RegisterDealerPointsServiceServer(s, NewServer(svc))
	cli := dialDP(t, s)
	ctx := context.Background()
	cr, err := cli.CreateDealerPoint(ctx, &dealerpointsv1.CreateDealerPointRequest{Name: "N", Address: "A"})
	if err != nil {
		t.Fatal(err)
	}
	id := cr.GetDealerPoint().GetId()
	if _, err := cli.GetDealerPoint(ctx, &dealerpointsv1.GetDealerPointRequest{Id: id}); err != nil {
		t.Fatal(err)
	}
	if _, err := cli.ListDealerPoints(ctx, &dealerpointsv1.ListDealerPointsRequest{}); err != nil {
		t.Fatal(err)
	}
	n := "N2"
	if _, err := cli.UpdateDealerPoint(ctx, &dealerpointsv1.UpdateDealerPointRequest{Id: id, Name: &n}); err != nil {
		t.Fatal(err)
	}
	if _, err := cli.DeleteDealerPoint(ctx, &dealerpointsv1.DeleteDealerPointRequest{Id: id}); err != nil {
		t.Fatal(err)
	}
}
