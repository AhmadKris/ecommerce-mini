import { useState } from "react";
import { Navigate } from "react-router-dom";

import { usePermission } from "../../hooks/usePermission";
import { useDashboardReport } from "../../hooks/useReports";
import { formatCurrency } from "../../lib/format";
import type { OrderFunnel } from "../../types/report";

const periodOptions = [
  { days: 7, label: "7 hari terakhir" },
  { days: 30, label: "30 hari terakhir" },
  { days: 90, label: "90 hari terakhir" },
];

const funnelStages: { key: keyof OrderFunnel; label: string }[] = [
  { key: "pending", label: "Menunggu Pembayaran" },
  { key: "paid", label: "Dibayar" },
  { key: "processing", label: "Diproses" },
  { key: "shipped", label: "Dikirim" },
  { key: "delivered", label: "Diterima" },
];

function ChangeBadge({ percent }: { percent: number }) {
  const isPositive = percent >= 0;
  return (
    <span className={`text-label-sm ${isPositive ? "text-success-500" : "text-error-500"}`}>
      {isPositive ? "+" : ""}
      {percent.toFixed(1)}%
    </span>
  );
}

function KpiCard({
  label,
  value,
  change,
  hint,
}: {
  label: string;
  value: string;
  change?: number;
  hint?: string;
}) {
  return (
    <div className="rounded-md border border-(--border-default) bg-(--surface-card) p-4">
      <p className="text-label-sm text-(--ink-secondary)">{label}</p>
      <p className="text-heading-lg text-(--ink-primary) mt-2">{value}</p>
      {change !== undefined && (
        <p className="mt-1">
          <ChangeBadge percent={change} /> <span className="text-label-sm text-(--ink-secondary)">vs periode sebelumnya</span>
        </p>
      )}
      {hint && <p className="text-label-sm text-(--ink-secondary) mt-1">{hint}</p>}
    </div>
  );
}

export function AdminReports() {
  const canReadReports = usePermission("report:read");
  const [days, setDays] = useState(30);
  const { data, isLoading, isError } = useDashboardReport(days);

  if (!canReadReports) {
    return <Navigate to="/admin" replace />;
  }

  if (isLoading) {
    return (
      <div className="mx-auto max-w-4xl">
        <p className="text-body-md text-(--ink-secondary)">Memuat laporan...</p>
      </div>
    );
  }

  if (isError || !data) {
    return (
      <div className="mx-auto max-w-4xl">
        <p role="alert" className="text-body-md text-error-500">
          Gagal memuat laporan.
        </p>
      </div>
    );
  }

  const maxTrendAmount = Math.max(...data.revenue_trend.map((point) => point.amount), 1);
  const funnelMax = Math.max(...funnelStages.map((stage) => data.order_funnel[stage.key]), 1);

  return (
    <div className="mx-auto max-w-4xl">
      <div className="flex items-center justify-between">
        <h1 className="text-heading-xl text-(--ink-primary)">Laporan</h1>
        <label className="text-body-sm text-(--ink-secondary) flex items-center gap-2">
          Periode
          <select
            value={days}
            onChange={(event) => setDays(Number(event.target.value))}
            className="text-body-sm rounded-md border border-(--border-default) px-2 py-1.5"
          >
            {periodOptions.map((option) => (
              <option key={option.days} value={option.days}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
      </div>

      <div className="mt-6 grid grid-cols-2 gap-4 sm:grid-cols-4">
        <KpiCard label="Pendapatan" value={formatCurrency(data.revenue.total)} change={data.revenue.change_percent} />
        <KpiCard label="Pesanan" value={String(data.orders.total)} change={data.orders.change_percent} />
        <KpiCard label="Pesanan Pending" value={String(data.pending_orders)} hint="Perlu perhatian" />
        <KpiCard
          label="Stok Rendah"
          value={String(data.low_stock.total)}
          hint={`${data.low_stock.critical} item kritis`}
        />
      </div>

      <div className="mt-6 rounded-md border border-(--border-default) bg-(--surface-card) p-4">
        <p className="text-label-sm text-(--ink-secondary) mb-4">Tren Pendapatan</p>
        <div className="flex h-40 items-end gap-1">
          {data.revenue_trend.map((point) => (
            <div
              key={point.date}
              title={`${point.date}: ${formatCurrency(point.amount)}`}
              className="flex-1 rounded-t bg-primary-500/70"
              style={{ height: `${Math.max((point.amount / maxTrendAmount) * 100, 2)}%` }}
            />
          ))}
        </div>
        <div className="text-label-sm text-(--ink-secondary) mt-2 flex justify-between">
          <span>{data.revenue_trend[0]?.date}</span>
          <span>{data.revenue_trend[data.revenue_trend.length - 1]?.date}</span>
        </div>
      </div>

      <div className="mt-6 rounded-md border border-(--border-default) bg-(--surface-card) p-4">
        <p className="text-label-sm text-(--ink-secondary) mb-4">Funnel Status Order</p>
        <div className="flex flex-col gap-3">
          {funnelStages.map((stage) => {
            const count = data.order_funnel[stage.key];
            return (
              <div key={stage.key} className="grid grid-cols-[160px_1fr_40px] items-center gap-3">
                <span className="text-body-sm text-(--ink-secondary)">{stage.label}</span>
                <div className="h-2 overflow-hidden rounded-full bg-neutral-100">
                  <div
                    className="h-full rounded-full bg-primary-500"
                    style={{ width: `${(count / funnelMax) * 100}%` }}
                  />
                </div>
                <span className="text-data-md text-(--ink-primary) text-right">{count}</span>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
