package service

import (
	"context"
	"testing"
	"time"

	"ecommerce-backend/internal/model"
)

type fakeReportRepo struct {
	revenueByPeriod    map[string]float64
	orderCountByPeriod map[string]int64
	orderCountByStatus map[string]int64
	trend              []model.RevenueTrendPoint
	lowStockTotal      int64
	lowStockCritical   int64
	funnel             model.OrderFunnel
}

func periodKey(start, end time.Time) string {
	return start.Format(time.RFC3339) + "|" + end.Format(time.RFC3339)
}

func (r *fakeReportRepo) RevenueTotal(_ context.Context, start, end time.Time) (float64, error) {
	return r.revenueByPeriod[periodKey(start, end)], nil
}

func (r *fakeReportRepo) OrderCount(_ context.Context, start, end time.Time) (int64, error) {
	return r.orderCountByPeriod[periodKey(start, end)], nil
}

func (r *fakeReportRepo) OrderCountByStatus(_ context.Context, status string) (int64, error) {
	return r.orderCountByStatus[status], nil
}

func (r *fakeReportRepo) RevenueTrend(_ context.Context, _, _ time.Time) ([]model.RevenueTrendPoint, error) {
	return r.trend, nil
}

func (r *fakeReportRepo) LowStockCount(_ context.Context, _, _ int) (int64, int64, error) {
	return r.lowStockTotal, r.lowStockCritical, nil
}

func (r *fakeReportRepo) OrderFunnel(_ context.Context) (model.OrderFunnel, error) {
	return r.funnel, nil
}

func TestReportService_Dashboard_ComputesChangePercent(t *testing.T) {
	now := time.Now()
	periodStart := now.AddDate(0, 0, -30)
	previousStart := periodStart.AddDate(0, 0, -30)

	repo := &fakeReportRepo{
		revenueByPeriod: map[string]float64{
			periodKey(periodStart, now):           150000,
			periodKey(previousStart, periodStart): 100000,
		},
		orderCountByPeriod: map[string]int64{
			periodKey(periodStart, now):           20,
			periodKey(previousStart, periodStart): 10,
		},
		orderCountByStatus: map[string]int64{model.OrderStatusPending: 3},
		lowStockTotal:      5,
		lowStockCritical:   1,
		funnel:             model.OrderFunnel{Pending: 3, Paid: 4, Processing: 2, Shipped: 1, Delivered: 6},
	}
	svc := NewReportService(repo)

	report, err := svc.Dashboard(context.Background(), 30)
	if err != nil {
		t.Fatalf("Dashboard returned error: %v", err)
	}

	if report.Revenue.Total != 150000 || report.Revenue.PreviousTotal != 100000 {
		t.Errorf("Revenue = %+v, want total 150000 previous 100000", report.Revenue)
	}
	if report.Revenue.ChangePercent != 50 {
		t.Errorf("Revenue.ChangePercent = %v, want 50", report.Revenue.ChangePercent)
	}
	if report.Orders.Total != 20 || report.Orders.ChangePercent != 100 {
		t.Errorf("Orders = %+v, want total 20 changePercent 100", report.Orders)
	}
	if report.PendingOrders != 3 {
		t.Errorf("PendingOrders = %d, want 3", report.PendingOrders)
	}
	if report.LowStock.Total != 5 || report.LowStock.Critical != 1 {
		t.Errorf("LowStock = %+v, want total 5 critical 1", report.LowStock)
	}
	if report.OrderFunnel != repo.funnel {
		t.Errorf("OrderFunnel = %+v, want %+v", report.OrderFunnel, repo.funnel)
	}
}

func TestReportService_Dashboard_DefaultsDaysWhenInvalid(t *testing.T) {
	repo := &fakeReportRepo{}
	svc := NewReportService(repo)

	report, err := svc.Dashboard(context.Background(), 0)
	if err != nil {
		t.Fatalf("Dashboard returned error: %v", err)
	}
	// defaultReportDays is 30 — with no seeded data, revenue trend should
	// still have exactly 30 points (one per day), not an empty slice.
	if len(report.RevenueTrend) != defaultReportDays {
		t.Errorf("len(RevenueTrend) = %d, want %d", len(report.RevenueTrend), defaultReportDays)
	}
}

func TestReportService_Dashboard_FillsMissingDaysWithZero(t *testing.T) {
	now := time.Now()
	periodStart := now.AddDate(0, 0, -3)
	repo := &fakeReportRepo{
		trend: []model.RevenueTrendPoint{
			{Date: periodStart.Format("2006-01-02"), Amount: 5000},
		},
	}
	svc := NewReportService(repo)

	report, err := svc.Dashboard(context.Background(), 3)
	if err != nil {
		t.Fatalf("Dashboard returned error: %v", err)
	}

	if len(report.RevenueTrend) != 3 {
		t.Fatalf("len(RevenueTrend) = %d, want 3", len(report.RevenueTrend))
	}
	if report.RevenueTrend[0].Amount != 5000 {
		t.Errorf("RevenueTrend[0].Amount = %v, want 5000", report.RevenueTrend[0].Amount)
	}
	for i, point := range report.RevenueTrend[1:] {
		if point.Amount != 0 {
			t.Errorf("RevenueTrend[%d].Amount = %v, want 0 (day with no orders)", i+1, point.Amount)
		}
	}
}

func TestPercentChange_ZeroPrevious(t *testing.T) {
	if got := percentChange(0, 0); got != 0 {
		t.Errorf("percentChange(0, 0) = %v, want 0", got)
	}
	if got := percentChange(0, 100); got != 100 {
		t.Errorf("percentChange(0, 100) = %v, want 100", got)
	}
}
