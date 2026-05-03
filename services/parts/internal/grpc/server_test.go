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

	"github.com/dealer/dealer/services/parts/internal/domain"
	"github.com/dealer/dealer/services/parts/internal/service"
	partsv1 "github.com/dealer/dealer/pkg/pb/parts/v1"
)

type gprPart struct {
	parts map[uuid.UUID]*domain.Part
}

func (f *gprPart) Create(_ context.Context, p *domain.Part) error {
	if f.parts == nil {
		f.parts = make(map[uuid.UUID]*domain.Part)
	}
	cp := *p
	f.parts[p.ID] = &cp
	return nil
}

func (f *gprPart) GetByID(_ context.Context, id uuid.UUID) (*domain.Part, error) {
	p, ok := f.parts[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	cp := *p
	return &cp, nil
}

func (f *gprPart) List(_ context.Context, _ domain.PartListFilter) ([]*domain.Part, int32, error) {
	var out []*domain.Part
	for _, p := range f.parts {
		cp := *p
		out = append(out, &cp)
	}
	return out, int32(len(out)), nil
}

func (f *gprPart) Update(_ context.Context, p *domain.Part) error {
	if _, ok := f.parts[p.ID]; !ok {
		return pgx.ErrNoRows
	}
	cp := *p
	f.parts[p.ID] = &cp
	return nil
}

func (f *gprPart) Delete(_ context.Context, id uuid.UUID) error {
	delete(f.parts, id)
	return nil
}

type gprStock struct{ repo *gprPart }

func (gprStock) ListByPart(context.Context, uuid.UUID) ([]*domain.PartStock, error) { return nil, nil }

func (f *gprStock) Upsert(_ context.Context, partID, _ uuid.UUID, quantity int32) error {
	if p, ok := f.repo.parts[partID]; ok {
		p.Quantity = quantity
	}
	return nil
}

func (f *gprStock) ReplaceForPart(_ context.Context, partID uuid.UUID, rows []struct {
	WarehouseID uuid.UUID
	Quantity    int32
}) error {
	var sum int32
	for _, r := range rows {
		sum += r.Quantity
	}
	if p, ok := f.repo.parts[partID]; ok {
		p.Quantity = sum
	}
	return nil
}

type gprFolder struct{ folders map[uuid.UUID]*domain.PartFolder }

func (f *gprFolder) Create(_ context.Context, folder *domain.PartFolder) error {
	if f.folders == nil {
		f.folders = make(map[uuid.UUID]*domain.PartFolder)
	}
	cp := *folder
	f.folders[folder.ID] = &cp
	return nil
}

func (f *gprFolder) GetByID(_ context.Context, id uuid.UUID) (*domain.PartFolder, error) {
	x, ok := f.folders[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	cp := *x
	return &cp, nil
}

func (gprFolder) ListByParent(context.Context, *uuid.UUID) ([]*domain.PartFolder, error) { return nil, nil }

func (f *gprFolder) Update(_ context.Context, folder *domain.PartFolder) error {
	if _, ok := f.folders[folder.ID]; !ok {
		return pgx.ErrNoRows
	}
	cp := *folder
	f.folders[folder.ID] = &cp
	return nil
}

func (f *gprFolder) Delete(_ context.Context, id uuid.UUID) error {
	delete(f.folders, id)
	return nil
}

func dialParts(t *testing.T, srv *grpc.Server) partsv1.PartsServiceClient {
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
	return partsv1.NewPartsServiceClient(c)
}

func TestPartsGRPC_PartFlow(t *testing.T) {
	pr := &gprPart{parts: map[uuid.UUID]*domain.Part{}}
	svc := service.NewPartService(pr, &gprFolder{folders: map[uuid.UUID]*domain.PartFolder{}}, &gprStock{repo: pr})
	s := grpc.NewServer()
	partsv1.RegisterPartsServiceServer(s, NewServer(svc))
	cli := dialParts(t, s)
	ctx := context.Background()
	cr, err := cli.CreatePart(ctx, &partsv1.CreatePartRequest{Sku: "S", Name: "N", Category: "c"})
	if err != nil {
		t.Fatal(err)
	}
	id := cr.GetPart().GetId()
	if _, err := cli.GetPart(ctx, &partsv1.GetPartRequest{Id: id}); err != nil {
		t.Fatal(err)
	}
	if _, err := cli.ListParts(ctx, &partsv1.ListPartsRequest{}); err != nil {
		t.Fatal(err)
	}
	n := "N2"
	if _, err := cli.UpdatePart(ctx, &partsv1.UpdatePartRequest{Id: id, Name: &n}); err != nil {
		t.Fatal(err)
	}
	if _, err := cli.DeletePart(ctx, &partsv1.DeletePartRequest{Id: id}); err != nil {
		t.Fatal(err)
	}
}

func TestPartsGRPC_FolderFlow(t *testing.T) {
	pr := &gprPart{parts: map[uuid.UUID]*domain.Part{}}
	fr := &gprFolder{folders: map[uuid.UUID]*domain.PartFolder{}}
	svc := service.NewPartService(pr, fr, &gprStock{repo: pr})
	s := grpc.NewServer()
	partsv1.RegisterPartsServiceServer(s, NewServer(svc))
	cli := dialParts(t, s)
	ctx := context.Background()
	f, err := cli.CreateFolder(ctx, &partsv1.CreateFolderRequest{Name: "F"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cli.GetFolder(ctx, &partsv1.GetFolderRequest{Id: f.GetFolder().GetId()}); err != nil {
		t.Fatal(err)
	}
	if _, err := cli.ListFolders(ctx, &partsv1.ListFoldersRequest{}); err != nil {
		t.Fatal(err)
	}
	n := "F2"
	if _, err := cli.UpdateFolder(ctx, &partsv1.UpdateFolderRequest{Id: f.GetFolder().GetId(), Name: &n}); err != nil {
		t.Fatal(err)
	}
	if _, err := cli.DeleteFolder(ctx, &partsv1.DeleteFolderRequest{Id: f.GetFolder().GetId()}); err != nil {
		t.Fatal(err)
	}
}
