package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"ecommerce-backend/internal/model"
)

// revenueStatuses are the order statuses that count as realized revenue —
// pending orders haven't been paid yet, and cancelled orders never will be.
var revenueStatuses = []string{
	model.OrderStatusPaid,
	model.OrderStatusProcessing,
	model.OrderStatusShipped,
	model.OrderStatusDelivered,
}

// funnelStatuses are the linear stages shown in the admin dashboard's order
// funnel, in order. Cancelled is deliberately excluded — see
// model.OrderFunnel's doc comment.
var funnelStatuses = []string{
	model.OrderStatusPending,
	model.OrderStatusPaid,
	model.OrderStatusProcessing,
	model.OrderStatusShipped,
	model.OrderStatusDelivered,
}

// ReportRepository aggregates data that already lives in the orders and
// products tables for the admin dashboard — it deliberately doesn't belong
// on OrderRepository/ProductRepository, since these queries exist only to
// serve one read-only reporting concern, not the resources' own CRUD.
type ReportRepository interface {
	RevenueTotal(ctx context.Context, start, end time.Time) (float64, error)
	OrderCount(ctx context.Context, start, end time.Time) (int64, error)
	OrderCountByStatus(ctx context.Context, status string) (int64, error)
	RevenueTrend(ctx context.Context, start, end time.Time) ([]model.RevenueTrendPoint, error)
	LowStockCount(ctx context.Context, lowThreshold, criticalThreshold int) (total int64, critical int64, err error)
	OrderFunnel(ctx context.Context) (model.OrderFunnel, error)
}

type reportRepository struct {
	db *gorm.DB
}

// NewReportRepository builds a ReportRepository backed by db.
func NewReportRepository(db *gorm.DB) ReportRepository {
	return &reportRepository{db: db}
}

func (r *reportRepository) RevenueTotal(ctx context.Context, start, end time.Time) (float64, error) {
	var total float64
	err := r.db.WithContext(ctx).Model(&model.Order{}).
		Where("created_at >= ? AND created_at < ? AND status IN ?", start, end, revenueStatuses).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&total).Error
	if err != nil {
		return 0, fmt.Errorf("repository: revenue total: %w", err)
	}
	return total, nil
}

func (r *reportRepository) OrderCount(ctx context.Context, start, end time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Order{}).
		Where("created_at >= ? AND created_at < ?", start, end).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("repository: order count: %w", err)
	}
	return count, nil
}

func (r *reportRepository) OrderCountByStatus(ctx context.Context, status string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Order{}).Where("status = ?", status).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("repository: order count by status: %w", err)
	}
	return count, nil
}

func (r *reportRepository) RevenueTrend(ctx context.Context, start, end time.Time) ([]model.RevenueTrendPoint, error) {
	var rows []model.RevenueTrendPoint
	err := r.db.WithContext(ctx).Model(&model.Order{}).
		Where("created_at >= ? AND created_at < ? AND status IN ?", start, end, revenueStatuses).
		Select("TO_CHAR(created_at, 'YYYY-MM-DD') AS date, COALESCE(SUM(total_amount), 0) AS amount").
		Group("TO_CHAR(created_at, 'YYYY-MM-DD')").
		Order("date").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("repository: revenue trend: %w", err)
	}
	return rows, nil
}

func (r *reportRepository) LowStockCount(ctx context.Context, lowThreshold, criticalThreshold int) (int64, int64, error) {
	var row struct {
		Total    int64
		Critical int64
	}
	err := r.db.WithContext(ctx).Model(&model.Product{}).
		Select(
			"COUNT(*) FILTER (WHERE stock <= ?) AS total, COUNT(*) FILTER (WHERE stock <= ?) AS critical",
			lowThreshold, criticalThreshold,
		).
		Scan(&row).Error
	if err != nil {
		return 0, 0, fmt.Errorf("repository: low stock count: %w", err)
	}
	return row.Total, row.Critical, nil
}

func (r *reportRepository) OrderFunnel(ctx context.Context) (model.OrderFunnel, error) {
	var rows []struct {
		Status string
		Count  int64
	}
	err := r.db.WithContext(ctx).Model(&model.Order{}).
		Where("status IN ?", funnelStatuses).
		Select("status, COUNT(*) AS count").
		Group("status").
		Scan(&rows).Error
	if err != nil {
		return model.OrderFunnel{}, fmt.Errorf("repository: order funnel: %w", err)
	}

	var funnel model.OrderFunnel
	for _, row := range rows {
		switch row.Status {
		case model.OrderStatusPending:
			funnel.Pending = row.Count
		case model.OrderStatusPaid:
			funnel.Paid = row.Count
		case model.OrderStatusProcessing:
			funnel.Processing = row.Count
		case model.OrderStatusShipped:
			funnel.Shipped = row.Count
		case model.OrderStatusDelivered:
			funnel.Delivered = row.Count
		}
	}
	return funnel, nil
}
