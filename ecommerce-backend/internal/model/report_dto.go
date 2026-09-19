package model

// ReportPeriodQuery binds GET /api/admin/reports/dashboard query params.
type ReportPeriodQuery struct {
	Days int `form:"days"`
}

// RevenueSummary compares a period's revenue against the immediately
// preceding period of equal length (e.g. last 30 days vs the 30 days
// before that).
type RevenueSummary struct {
	Total         float64 `json:"total"`
	PreviousTotal float64 `json:"previous_total"`
	ChangePercent float64 `json:"change_percent"`
}

// OrderSummary mirrors RevenueSummary's period-over-period comparison, but
// counts every order placed regardless of status — checkout volume, not
// revenue.
type OrderSummary struct {
	Total         int64   `json:"total"`
	PreviousTotal int64   `json:"previous_total"`
	ChangePercent float64 `json:"change_percent"`
}

// LowStockSummary is a current snapshot (not period-bound) — Total counts
// products at or below the low-stock threshold, Critical the subset at or
// below the tighter critical threshold.
type LowStockSummary struct {
	Total    int64 `json:"total"`
	Critical int64 `json:"critical"`
}

// RevenueTrendPoint is one day's revenue in the trend series. Days with no
// orders still appear with Amount 0 (see ReportService.fillMissingDays) so
// the frontend can render a continuous chart without gaps.
type RevenueTrendPoint struct {
	Date   string  `json:"date"`
	Amount float64 `json:"amount"`
}

// OrderFunnel is a current snapshot of how many non-cancelled orders sit at
// each stage. Cancelled is excluded — it's a side branch off the linear
// progression, not a stage in it (same reasoning as the frontend's order
// status timeline, see ecommerce-frontend OrderDetail.tsx).
type OrderFunnel struct {
	Pending    int64 `json:"pending"`
	Paid       int64 `json:"paid"`
	Processing int64 `json:"processing"`
	Shipped    int64 `json:"shipped"`
	Delivered  int64 `json:"delivered"`
}

// DashboardReport is the full payload for GET /api/admin/reports/dashboard.
type DashboardReport struct {
	Revenue       RevenueSummary      `json:"revenue"`
	Orders        OrderSummary        `json:"orders"`
	PendingOrders int64               `json:"pending_orders"`
	LowStock      LowStockSummary     `json:"low_stock"`
	RevenueTrend  []RevenueTrendPoint `json:"revenue_trend"`
	OrderFunnel   OrderFunnel         `json:"order_funnel"`
}
