package service

import (
	"context"
)

type invitationRepository interface {
	CreateFromClosedWorkOrders(ctx context.Context, limit int) (int64, error)
	CreateFromCompletedDeals(ctx context.Context, limit int) (int64, error)
	CreateFromClosedGoodsSales(ctx context.Context, limit int) (int64, error)
}

type notificationRepository interface {
	CreateFromClosedCustomerOrderReceipts(ctx context.Context, limit int) (int64, error)
	CreateRepairAppointmentReminders(ctx context.Context, limit int) (int64, error)
}

type RunResult struct {
	WorkOrdersCreated        int64
	DealsCreated             int64
	GoodsSalesCreated        int64
	OrderReceiptsNotified       int64
	AppointmentRemindersSent  int64
}

type SchedulerService struct {
	invitations   invitationRepository
	notifications notificationRepository
	batchSize     int
}

func NewSchedulerService(invitations invitationRepository, notifications notificationRepository, batchSize int) *SchedulerService {
	return &SchedulerService{invitations: invitations, notifications: notifications, batchSize: batchSize}
}

func (s *SchedulerService) RunOnce(ctx context.Context) (RunResult, error) {
	wo, err := s.invitations.CreateFromClosedWorkOrders(ctx, s.batchSize)
	if err != nil {
		return RunResult{}, err
	}
	deals, err := s.invitations.CreateFromCompletedDeals(ctx, s.batchSize)
	if err != nil {
		return RunResult{}, err
	}
	sales, err := s.invitations.CreateFromClosedGoodsSales(ctx, s.batchSize)
	if err != nil {
		return RunResult{}, err
	}
	var orderReceipts, appointmentReminders int64
	if s.notifications != nil {
		orderReceipts, err = s.notifications.CreateFromClosedCustomerOrderReceipts(ctx, s.batchSize)
		if err != nil {
			return RunResult{}, err
		}
		appointmentReminders, err = s.notifications.CreateRepairAppointmentReminders(ctx, s.batchSize)
		if err != nil {
			return RunResult{}, err
		}
	}
	return RunResult{
		WorkOrdersCreated:        wo,
		DealsCreated:             deals,
		GoodsSalesCreated:        sales,
		OrderReceiptsNotified:    orderReceipts,
		AppointmentRemindersSent: appointmentReminders,
	}, nil
}
