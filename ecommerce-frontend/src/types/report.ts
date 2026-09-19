export interface RevenueSummary {
  total: number;
  previous_total: number;
  change_percent: number;
}

export interface OrderSummary {
  total: number;
  previous_total: number;
  change_percent: number;
}

export interface LowStockSummary {
  total: number;
  critical: number;
}

export interface RevenueTrendPoint {
  date: string;
  amount: number;
}

export interface OrderFunnel {
  pending: number;
  paid: number;
  processing: number;
  shipped: number;
  delivered: number;
}

export interface DashboardReport {
  revenue: RevenueSummary;
  orders: OrderSummary;
  pending_orders: number;
  low_stock: LowStockSummary;
  revenue_trend: RevenueTrendPoint[];
  order_funnel: OrderFunnel;
}
