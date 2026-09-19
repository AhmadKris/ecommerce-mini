package service

import (
	"context"
	"fmt"
	"time"

	"ecommerce-backend/internal/apperror"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/repository"
)

const (
	defaultReportDays = 30

	// lowStockThreshold/criticalStockThreshold are simplification choices
	// for this project's scale — a real inventory system would derive these
	// per-product from reorder lead time and sales velocity, not a single
	// global number.
	lowStockThreshold      = 10
	criticalStockThreshold = 3
)

// ReportService computes the admin dashboard's business metrics from
// existing order/product data — no new infrastructure (see .claude/CLAUDE.md
// Known Issues: this was previously deferred pending Prometheus/Grafana,
// which is a separate concern — system metrics, not business reporting).
type ReportService struct {
	reportRepo repository.ReportRepository
}

// NewReportService builds a ReportService with its dependencies.
func NewReportService(reportRepo repository.ReportRepository) *ReportService {
	return &ReportService{reportRepo: reportRepo}
}

// Dashboard returns revenue/order totals for the last `days` days compared
// against the equal-length period before that, plus current snapshots
// (pending orders, low stock, order funnel) that aren't period-bound.
func (s *ReportService) Dashboard(ctx context.Context, days int) (*model.DashboardReport, error) {
	if days < 1 {
		days = defaultReportDays
	}

	now := time.Now()
	periodStart := now.AddDate(0, 0, -days)
	previousStart := periodStart.AddDate(0, 0, -days)

	revenue, err := s.reportRepo.RevenueTotal(ctx, periodStart, now)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: dashboard revenue: %w", err))
	}
	previousRevenue, err := s.reportRepo.RevenueTotal(ctx, previousStart, periodStart)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: dashboard previous revenue: %w", err))
	}

	orderCount, err := s.reportRepo.OrderCount(ctx, periodStart, now)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: dashboard order count: %w", err))
	}
	previousOrderCount, err := s.reportRepo.OrderCount(ctx, previousStart, periodStart)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: dashboard previous order count: %w", err))
	}

	pendingOrders, err := s.reportRepo.OrderCountByStatus(ctx, model.OrderStatusPending)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: dashboard pending orders: %w", err))
	}

	lowStockTotal, lowStockCritical, err := s.reportRepo.LowStockCount(ctx, lowStockThreshold, criticalStockThreshold)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: dashboard low stock: %w", err))
	}

	trend, err := s.reportRepo.RevenueTrend(ctx, periodStart, now)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: dashboard revenue trend: %w", err))
	}

	funnel, err := s.reportRepo.OrderFunnel(ctx)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: dashboard order funnel: %w", err))
	}

	return &model.DashboardReport{
		Revenue: model.RevenueSummary{
			Total:         revenue,
			PreviousTotal: previousRevenue,
			ChangePercent: percentChange(previousRevenue, revenue),
		},
		Orders: model.OrderSummary{
			Total:         orderCount,
			PreviousTotal: previousOrderCount,
			ChangePercent: percentChange(float64(previousOrderCount), float64(orderCount)),
		},
		PendingOrders: pendingOrders,
		LowStock:      model.LowStockSummary{Total: lowStockTotal, Critical: lowStockCritical},
		RevenueTrend:  fillMissingDays(trend, periodStart, now),
		OrderFunnel:   funnel,
	}, nil
}

// percentChange returns how much current changed relative to previous. A
// previous value of 0 makes a true percentage undefined (division by
// zero) — this treats "0 to something" as +100% and "0 to 0" as 0%, a
// simplification reasonable for this project's scale rather than a
// mathematically rigorous growth-rate formula.
func percentChange(previous, current float64) float64 {
	if previous == 0 {
		if current == 0 {
			return 0
		}
		return 100
	}
	return (current - previous) / previous * 100
}

// fillMissingDays inserts zero-amount points for any day in [start, end)
// the repository query didn't return (no orders that day), so the frontend
// can render a continuous daily series instead of a chart with gaps.
func fillMissingDays(points []model.RevenueTrendPoint, start, end time.Time) []model.RevenueTrendPoint {
	byDate := make(map[string]float64, len(points))
	for _, point := range points {
		byDate[point.Date] = point.Amount
	}

	filled := make([]model.RevenueTrendPoint, 0, int(end.Sub(start).Hours()/24)+1)
	for day := start; day.Before(end); day = day.AddDate(0, 0, 1) {
		date := day.Format("2006-01-02")
		filled = append(filled, model.RevenueTrendPoint{Date: date, Amount: byDate[date]})
	}
	return filled
}
