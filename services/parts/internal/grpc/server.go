package grpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/dealer/dealer/services/parts/internal/domain"
	"github.com/dealer/dealer/services/parts/internal/service"
	partsv1 "github.com/dealer/dealer/pkg/pb/parts/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	partsv1.UnimplementedPartsServiceServer
	svc *service.PartService
}

func NewServer(svc *service.PartService) *Server {
	return &Server{svc: svc}
}

func folderToProto(f *domain.PartFolder) *partsv1.PartFolder {
	if f == nil {
		return nil
	}
	parentID := ""
	if f.ParentID != nil {
		parentID = f.ParentID.String()
	}
	return &partsv1.PartFolder{
		Id:        f.ID.String(),
		Name:      f.Name,
		ParentId:  parentID,
		CreatedAt: f.CreatedAt.Unix(),
		UpdatedAt: f.UpdatedAt.Unix(),
	}
}

func toProto(p *domain.Part) *partsv1.Part {
	if p == nil {
		return nil
	}
	folderID := ""
	if p.FolderID != nil {
		folderID = p.FolderID.String()
	}
	brandID := ""
	if p.BrandID != nil {
		brandID = p.BrandID.String()
	}
	dpID, leID, whID := "", "", ""
	if p.DealerPointID != nil {
		dpID = p.DealerPointID.String()
	}
	if p.LegalEntityID != nil {
		leID = p.LegalEntityID.String()
	}
	if p.WarehouseID != nil {
		whID = p.WarehouseID.String()
	}
	return &partsv1.Part{
		Id:             p.ID.String(),
		Sku:            p.SKU,
		Name:           p.Name,
		Category:       p.Category,
		FolderId:       folderID,
		BrandId:        brandID,
		DealerPointId:  dpID,
		LegalEntityId:  leID,
		WarehouseId:    whID,
		Quantity:       p.Quantity,
		Unit:           p.Unit,
		Price:          p.Price,
		Location:       p.Location,
		Notes:          p.Notes,
		CreatedAt:      p.CreatedAt.Unix(),
		UpdatedAt:      p.UpdatedAt.Unix(),
	}
}

func parseUUIDOpt(s string) *uuid.UUID {
	if s == "" {
		return nil
	}
	u, err := uuid.Parse(s)
	if err != nil {
		return nil
	}
	return &u
}

func strPtr(s string) *string {
	return &s
}

func (s *Server) CreatePart(ctx context.Context, req *partsv1.CreatePartRequest) (*partsv1.CreatePartResponse, error) {
	p, err := s.svc.Create(ctx, service.CreatePartInput{
		SKU: req.Sku, Name: req.Name, Category: req.Category,
		FolderID: parseUUIDOpt(req.FolderId), BrandID: parseUUIDOpt(req.BrandId), DealerPointID: parseUUIDOpt(req.DealerPointId),
		LegalEntityID: parseUUIDOpt(req.LegalEntityId), WarehouseID: parseUUIDOpt(req.WarehouseId),
		Quantity: req.Quantity, Unit: req.Unit, Price: req.Price, Location: req.Location, Notes: req.Notes,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &partsv1.CreatePartResponse{Part: toProto(p)}, nil
}

func (s *Server) GetPart(ctx context.Context, req *partsv1.GetPartRequest) (*partsv1.GetPartResponse, error) {
	p, err := s.svc.Get(ctx, req.Id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "part not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &partsv1.GetPartResponse{Part: toProto(p)}, nil
}

func (s *Server) ListParts(ctx context.Context, req *partsv1.ListPartsRequest) (*partsv1.ListPartsResponse, error) {
	list, total, err := s.svc.List(ctx, domain.PartListFilter{
		Limit: req.Limit, Offset: req.Offset, Search: req.Search, CategoryFilter: req.Category,
		FolderID: parseUUIDOpt(req.FolderId), BrandID: parseUUIDOpt(req.BrandId), DealerPointID: parseUUIDOpt(req.DealerPointId),
		LegalEntityID: parseUUIDOpt(req.LegalEntityId), WarehouseID: parseUUIDOpt(req.WarehouseId),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out := make([]*partsv1.Part, len(list))
	for i, p := range list {
		out[i] = toProto(p)
	}
	return &partsv1.ListPartsResponse{Parts: out, Total: total}, nil
}

func (s *Server) UpdatePart(ctx context.Context, req *partsv1.UpdatePartRequest) (*partsv1.UpdatePartResponse, error) {
	in := service.UpdatePartInput{
		SKU: req.Sku, Name: req.Name, Category: req.Category, Quantity: req.Quantity,
		Unit: req.Unit, Price: req.Price, Location: req.Location, Notes: req.Notes,
	}
	if req.FolderId != nil {
		in.FolderID = strPtr(req.GetFolderId())
	}
	if req.BrandId != nil {
		in.BrandID = strPtr(req.GetBrandId())
	}
	if req.DealerPointId != nil {
		in.DealerPointID = strPtr(req.GetDealerPointId())
	}
	if req.LegalEntityId != nil {
		in.LegalEntityID = strPtr(req.GetLegalEntityId())
	}
	if req.WarehouseId != nil {
		in.WarehouseID = strPtr(req.GetWarehouseId())
	}
	p, err := s.svc.Update(ctx, req.Id, in)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "part not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &partsv1.UpdatePartResponse{Part: toProto(p)}, nil
}

func (s *Server) DeletePart(ctx context.Context, req *partsv1.DeletePartRequest) (*partsv1.DeletePartResponse, error) {
	if err := s.svc.Delete(ctx, req.Id); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "part not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &partsv1.DeletePartResponse{}, nil
}

// Folders

func (s *Server) CreateFolder(ctx context.Context, req *partsv1.CreateFolderRequest) (*partsv1.CreateFolderResponse, error) {
	f, err := s.svc.CreateFolder(ctx, req.Name, parseUUIDOpt(req.ParentId))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &partsv1.CreateFolderResponse{Folder: folderToProto(f)}, nil
}

func (s *Server) GetFolder(ctx context.Context, req *partsv1.GetFolderRequest) (*partsv1.GetFolderResponse, error) {
	f, err := s.svc.GetFolder(ctx, req.Id)
	if err != nil {
		if errors.Is(err, service.ErrFolderNotFound) {
			return nil, status.Error(codes.NotFound, "folder not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &partsv1.GetFolderResponse{Folder: folderToProto(f)}, nil
}

func (s *Server) ListFolders(ctx context.Context, req *partsv1.ListFoldersRequest) (*partsv1.ListFoldersResponse, error) {
	list, err := s.svc.ListFolders(ctx, parseUUIDOpt(req.ParentId))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out := make([]*partsv1.PartFolder, len(list))
	for i, f := range list {
		out[i] = folderToProto(f)
	}
	return &partsv1.ListFoldersResponse{Folders: out}, nil
}

func (s *Server) UpdateFolder(ctx context.Context, req *partsv1.UpdateFolderRequest) (*partsv1.UpdateFolderResponse, error) {
	var parentIDOpt *string
	if req.ParentId != nil {
		v := req.GetParentId()
		parentIDOpt = &v
	}
	var name *string
	if req.Name != nil {
		v := req.GetName()
		name = &v
	}
	f, err := s.svc.UpdateFolder(ctx, req.Id, name, parentIDOpt)
	if err != nil {
		if errors.Is(err, service.ErrFolderNotFound) {
			return nil, status.Error(codes.NotFound, "folder not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &partsv1.UpdateFolderResponse{Folder: folderToProto(f)}, nil
}

func (s *Server) DeleteFolder(ctx context.Context, req *partsv1.DeleteFolderRequest) (*partsv1.DeleteFolderResponse, error) {
	if err := s.svc.DeleteFolder(ctx, req.Id); err != nil {
		if errors.Is(err, service.ErrFolderNotFound) {
			return nil, status.Error(codes.NotFound, "folder not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &partsv1.DeleteFolderResponse{}, nil
}
