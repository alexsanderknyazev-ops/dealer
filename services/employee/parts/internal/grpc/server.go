package grpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/dealer/dealer/services/parts/internal/client"
	"github.com/dealer/dealer/services/parts/internal/domain"
	"github.com/dealer/dealer/services/parts/internal/service"
	partsv1 "github.com/dealer/dealer/pkg/pb/parts/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	partsv1.UnimplementedPartsServiceServer
	svc          *service.PartService
	employees    *client.EmployeeResolver
	dealerPoints *client.DealerPointsChecker
	workOrders   *client.WorkOrdersNotifier
}

func NewServer(
	svc *service.PartService,
	employees *client.EmployeeResolver,
	dealerPoints *client.DealerPointsChecker,
	workOrders *client.WorkOrdersNotifier,
) *Server {
	return &Server{svc: svc, employees: employees, dealerPoints: dealerPoints, workOrders: workOrders}
}

func partWriteErr(err error) error {
	if errors.Is(err, service.ErrNotFound) {
		return status.Error(codes.NotFound, "part not found")
	}
	for _, refErr := range []error{
		service.ErrFolderNotFound,
		service.ErrBrandNotFound,
		service.ErrDealerPointNotFound,
		service.ErrLegalEntityNotFound,
		service.ErrWarehouseNotFound,
	} {
		if errors.Is(err, refErr) {
			return status.Error(codes.NotFound, err.Error())
		}
	}
	return status.Error(codes.Internal, err.Error())
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
		return nil, partWriteErr(err)
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

func (s *Server) ListPartStock(ctx context.Context, req *partsv1.ListPartStockRequest) (*partsv1.ListPartStockResponse, error) {
	rows, err := s.svc.ListStock(ctx, req.PartId)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "part not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	out := make([]*partsv1.PartStockRow, len(rows))
	for i, row := range rows {
		out[i] = &partsv1.PartStockRow{
			WarehouseId: row.WarehouseID.String(),
			Quantity:    row.Quantity,
		}
	}
	return &partsv1.ListPartStockResponse{Stock: out}, nil
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
		return nil, partWriteErr(err)
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

func (s *Server) documentToProto(ctx context.Context, d *domain.MovementDocument) *partsv1.MovementDocument {
	if d == nil {
		return nil
	}
	out := &partsv1.MovementDocument{
		Id:             d.ID.String(),
		DocumentNumber: d.DocumentNumber,
		Status:         d.Status,
		MovementType:   d.MovementType,
		ReferenceType:  d.ReferenceType,
		Notes:          d.Notes,
		CreatedAt:      d.CreatedAt.Unix(),
		UpdatedAt:      d.UpdatedAt.Unix(),
	}
	if d.ReferenceID != nil {
		out.ReferenceId = d.ReferenceID.String()
	}
	if d.ParentDocumentID != nil {
		out.ParentDocumentId = d.ParentDocumentID.String()
		if parent, err := s.svc.GetMovementDocument(ctx, d.ParentDocumentID.String()); err == nil && parent != nil {
			out.ParentDocumentNumber = parent.DocumentNumber
		}
	}
	if d.CreatedBy != nil {
		out.CreatedBy = d.CreatedBy.String()
		if s.employees != nil {
			out.CreatedByName = s.employees.FullName(ctx, *d.CreatedBy)
		}
	}
	if d.ConfirmedBy != nil {
		out.ConfirmedBy = d.ConfirmedBy.String()
		if s.employees != nil {
			out.ConfirmedByName = s.employees.FullName(ctx, *d.ConfirmedBy)
		}
	}
	if d.ConfirmedAt != nil {
		out.ConfirmedAt = d.ConfirmedAt.Unix()
	}
	out.Lines = make([]*partsv1.MovementDocumentLine, len(d.Lines))
	for i, l := range d.Lines {
		line := &partsv1.MovementDocumentLine{
			Id:          l.ID.String(),
			PartId:      l.PartID.String(),
			WarehouseId: l.WarehouseID.String(),
			Quantity:    l.Quantity,
			Notes:       l.Notes,
			SortOrder:   l.SortOrder,
		}
		if l.ReferenceLineID != nil {
			line.ReferenceLineId = l.ReferenceLineID.String()
		}
		if l.DestinationWarehouseID != nil {
			line.DestinationWarehouseId = l.DestinationWarehouseID.String()
			if s.dealerPoints != nil {
				line.DestinationWarehouseName = s.dealerPoints.WarehouseName(ctx, *l.DestinationWarehouseID)
			}
		}
		if part, err := s.svc.Get(ctx, l.PartID.String()); err == nil && part != nil {
			line.PartName = part.Name
			line.PartSku = part.SKU
		}
		if s.dealerPoints != nil {
			line.WarehouseName = s.dealerPoints.WarehouseName(ctx, l.WarehouseID)
		}
		if stockRows, err := s.svc.ListStock(ctx, l.PartID.String()); err == nil {
			for _, row := range stockRows {
				if row.WarehouseID == l.WarehouseID {
					line.SourceStockQuantity = row.Quantity
					break
				}
			}
		}
		out.Lines[i] = line
	}
	if d.ReferenceType == domain.RefWorkOrder && d.ReferenceID != nil && s.workOrders != nil {
		if wo, err := s.workOrders.GetWorkOrder(ctx, d.ReferenceID.String()); err == nil && wo != nil {
			out.ReferenceLabel = wo.OrderNumber
			out.CustomerName = wo.CustomerName
			out.VehicleVin = wo.VehicleVin
			out.VehicleLabel = wo.VehicleLabel
		}
	}
	if d.ReferenceType == domain.RefMovementDocument && d.ReferenceID != nil {
		if ref, err := s.svc.GetMovementDocument(ctx, d.ReferenceID.String()); err == nil && ref != nil {
			out.ReferenceLabel = ref.DocumentNumber
		}
	}
	return out
}

func movementDocumentErr(err error) error {
	switch {
	case errors.Is(err, service.ErrMovementDocumentNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, service.ErrMovementDocumentNotDraft),
		errors.Is(err, service.ErrMovementDocumentNotInProgress),
		errors.Is(err, service.ErrMovementDocumentNotEditable),
		errors.Is(err, service.ErrMovementDocumentNoLines):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, service.ErrInsufficientStock),
		errors.Is(err, service.ErrParentNotClosed),
		errors.Is(err, service.ErrParentNotExtractable),
		errors.Is(err, service.ErrOpenExtractionExists),
		errors.Is(err, service.ErrNothingToExtract),
		errors.Is(err, service.ErrExtractionExceedsBalance),
		errors.Is(err, service.ErrDestinationRequired),
		errors.Is(err, service.ErrSameSourceDestination):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, service.ErrNotFound), errors.Is(err, service.ErrWarehouseNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func (s *Server) CreateMovementDocument(ctx context.Context, req *partsv1.CreateMovementDocumentRequest) (*partsv1.CreateMovementDocumentResponse, error) {
	lines, err := parseMovementDocumentLines(req.Lines)
	if err != nil {
		return nil, err
	}
	var refID *uuid.UUID
	if req.ReferenceId != "" {
		if id, err := uuid.Parse(req.ReferenceId); err == nil {
			refID = &id
		}
	}
	var createdBy *uuid.UUID
	if req.CreatedBy != "" {
		if id, err := uuid.Parse(req.CreatedBy); err == nil {
			createdBy = &id
		}
	}
	doc, err := s.svc.CreateMovementDocument(ctx, domain.CreateMovementDocumentInput{
		MovementType: req.MovementType, ReferenceType: req.ReferenceType, ReferenceID: refID,
		Notes: req.Notes, CreatedBy: createdBy, Lines: lines,
	})
	if err != nil {
		return nil, movementDocumentErr(err)
	}
	return &partsv1.CreateMovementDocumentResponse{Document: s.documentToProto(ctx, doc)}, nil
}

func (s *Server) GetMovementDocument(ctx context.Context, req *partsv1.GetMovementDocumentRequest) (*partsv1.GetMovementDocumentResponse, error) {
	doc, err := s.svc.GetMovementDocument(ctx, req.Id)
	if err != nil {
		return nil, movementDocumentErr(err)
	}
	return &partsv1.GetMovementDocumentResponse{Document: s.documentToProto(ctx, doc)}, nil
}

func parseMovementDocumentLines(reqLines []*partsv1.MovementDocumentLineInput) ([]domain.MovementDocumentLineInput, error) {
	lines := make([]domain.MovementDocumentLineInput, 0, len(reqLines))
	for _, it := range reqLines {
		partID, err := uuid.Parse(it.PartId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid part_id")
		}
		warehouseID, err := uuid.Parse(it.WarehouseId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid warehouse_id")
		}
		var refLine *uuid.UUID
		if it.ReferenceLineId != "" {
			if id, err := uuid.Parse(it.ReferenceLineId); err == nil {
				refLine = &id
			}
		}
		var destWh *uuid.UUID
		if it.DestinationWarehouseId != "" {
			if id, err := uuid.Parse(it.DestinationWarehouseId); err == nil {
				destWh = &id
			}
		}
		lines = append(lines, domain.MovementDocumentLineInput{
			PartID: partID, WarehouseID: warehouseID, DestinationWarehouseID: destWh, Quantity: it.Quantity,
			ReferenceLineID: refLine, Notes: it.Notes, SortOrder: it.SortOrder,
		})
	}
	return lines, nil
}

func (s *Server) UpdateMovementDocument(ctx context.Context, req *partsv1.UpdateMovementDocumentRequest) (*partsv1.UpdateMovementDocumentResponse, error) {
	in := domain.UpdateMovementDocumentInput{ReplaceLines: req.ReplaceLines}
	if req.MovementType != nil {
		v := req.GetMovementType()
		in.MovementType = &v
	}
	if req.Notes != nil {
		v := req.GetNotes()
		in.Notes = &v
	}
	if req.ReplaceLines {
		lines, err := parseMovementDocumentLines(req.Lines)
		if err != nil {
			return nil, err
		}
		in.Lines = lines
	}
	doc, err := s.svc.UpdateMovementDocument(ctx, req.Id, in)
	if err != nil {
		return nil, movementDocumentErr(err)
	}
	return &partsv1.UpdateMovementDocumentResponse{Document: s.documentToProto(ctx, doc)}, nil
}

func (s *Server) ListMovementDocuments(ctx context.Context, req *partsv1.ListMovementDocumentsRequest) (*partsv1.ListMovementDocumentsResponse, error) {
	list, total, err := s.svc.ListMovementDocuments(ctx, req.Limit, req.Offset, req.Status, req.ReferenceType, req.ReferenceId)
	if err != nil {
		return nil, movementDocumentErr(err)
	}
	out := make([]*partsv1.MovementDocument, len(list))
	for i, d := range list {
		full, err := s.svc.GetMovementDocument(ctx, d.ID.String())
		if err != nil {
			out[i] = s.documentToProto(ctx, d)
		} else {
			out[i] = s.documentToProto(ctx, full)
		}
	}
	return &partsv1.ListMovementDocumentsResponse{Documents: out, Total: total}, nil
}

func (s *Server) StartMovementDocument(ctx context.Context, req *partsv1.StartMovementDocumentRequest) (*partsv1.StartMovementDocumentResponse, error) {
	doc, err := s.svc.StartMovementDocument(ctx, req.Id)
	if err != nil {
		return nil, movementDocumentErr(err)
	}
	return &partsv1.StartMovementDocumentResponse{Document: s.documentToProto(ctx, doc)}, nil
}

func (s *Server) CloseMovementDocument(ctx context.Context, req *partsv1.CloseMovementDocumentRequest) (*partsv1.CloseMovementDocumentResponse, error) {
	var closedBy *uuid.UUID
	if req.ClosedBy != "" {
		if id, err := uuid.Parse(req.ClosedBy); err == nil {
			closedBy = &id
		}
	}
	doc, err := s.svc.CloseMovementDocument(ctx, req.Id, closedBy)
	if err != nil {
		return nil, movementDocumentErr(err)
	}
	return &partsv1.CloseMovementDocumentResponse{Document: s.documentToProto(ctx, doc)}, nil
}

func (s *Server) ConfirmMovementDocument(ctx context.Context, req *partsv1.ConfirmMovementDocumentRequest) (*partsv1.ConfirmMovementDocumentResponse, error) {
	var confirmedBy *uuid.UUID
	if req.ConfirmedBy != "" {
		if id, err := uuid.Parse(req.ConfirmedBy); err == nil {
			confirmedBy = &id
		}
	}
	doc, err := s.svc.ConfirmMovementDocument(ctx, req.Id, confirmedBy)
	if err != nil {
		return nil, movementDocumentErr(err)
	}
	return &partsv1.ConfirmMovementDocumentResponse{Document: s.documentToProto(ctx, doc)}, nil
}

func (s *Server) CancelMovementDocument(ctx context.Context, req *partsv1.CancelMovementDocumentRequest) (*partsv1.CancelMovementDocumentResponse, error) {
	doc, err := s.svc.CancelMovementDocument(ctx, req.Id)
	if err != nil {
		return nil, movementDocumentErr(err)
	}
	return &partsv1.CancelMovementDocumentResponse{Document: s.documentToProto(ctx, doc)}, nil
}

func (s *Server) CreateProductionExtraction(ctx context.Context, req *partsv1.CreateProductionExtractionRequest) (*partsv1.CreateProductionExtractionResponse, error) {
	var createdBy *uuid.UUID
	if req.CreatedBy != "" {
		if id, err := uuid.Parse(req.CreatedBy); err == nil {
			createdBy = &id
		}
	}
	doc, err := s.svc.CreateProductionExtraction(ctx, req.Id, createdBy)
	if err != nil {
		return nil, movementDocumentErr(err)
	}
	return &partsv1.CreateProductionExtractionResponse{Document: s.documentToProto(ctx, doc)}, nil
}
