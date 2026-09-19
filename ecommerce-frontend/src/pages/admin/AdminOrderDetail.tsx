import { useState } from "react";
import { Link, useParams } from "react-router-dom";

import { Button } from "../../components/ui/Button";
import { useAdminOrder, useUpdateOrderStatus } from "../../hooks/useAdminOrders";
import { formatCurrency } from "../../lib/format";
import { ApiError } from "../../types/api";

// Mirrors orderStatusTransitions in internal/service/order_service.go —
// only these transitions are offered as buttons; the backend still enforces
// them independently (409 on anything else), this is purely UX guidance.
const nextStatusOptions: Record<string, string[]> = {
  pending: ["paid", "cancelled"],
  paid: ["processing", "cancelled"],
  processing: ["shipped", "cancelled"],
  shipped: ["delivered"],
  delivered: [],
  cancelled: [],
};

const statusLabels: Record<string, string> = {
  pending: "Menunggu Pembayaran",
  paid: "Dibayar",
  processing: "Diproses",
  shipped: "Dikirim",
  delivered: "Diterima",
  cancelled: "Dibatalkan",
};

export function AdminOrderDetail() {
  const { id = "" } = useParams<{ id: string }>();
  const orderId = Number(id);
  const { data: order, isLoading, isError } = useAdminOrder(orderId);
  const updateStatus = useUpdateOrderStatus(orderId);
  const [pendingStatus, setPendingStatus] = useState<string | null>(null);

  if (isLoading) {
    return (
      <div className="mx-auto max-w-3xl">
        <p className="text-body-md text-(--ink-secondary)">Memuat order...</p>
      </div>
    );
  }

  if (isError || !order) {
    return (
      <div className="mx-auto max-w-3xl">
        <p role="alert" className="text-body-md text-error-500">
          Order tidak ditemukan.
        </p>
        <Link to="/admin/orders" className="text-body-sm text-(--ink-link) mt-2 inline-block">
          Kembali ke daftar pesanan
        </Link>
      </div>
    );
  }

  const nextStatuses = nextStatusOptions[order.status] ?? [];

  function handleStatusChange(status: string) {
    setPendingStatus(status);
    updateStatus.mutate(status, { onSettled: () => setPendingStatus(null) });
  }

  return (
    <div className="mx-auto max-w-3xl">
      <Link to="/admin/orders" className="text-body-sm text-(--ink-link)">
        ← Kembali ke daftar pesanan
      </Link>

      <div className="mt-4 flex items-center justify-between">
        <h1 className="text-heading-xl text-(--ink-primary)">Order #{order.id}</h1>
        <span className="text-label-sm rounded-full bg-warning-100 px-2.5 py-1 text-warning-500">
          {statusLabels[order.status] ?? order.status}
        </span>
      </div>

      <div className="mt-6 grid gap-4 sm:grid-cols-2">
        <div className="rounded-md border border-(--border-default) bg-(--surface-card) p-4">
          <p className="text-label-sm text-(--ink-secondary)">Customer</p>
          <p className="text-body-md text-(--ink-primary) mt-1">User ID: {order.user_id}</p>
        </div>
        <div className="rounded-md border border-(--border-default) bg-(--surface-card) p-4">
          <p className="text-label-sm text-(--ink-secondary)">Alamat Pengiriman</p>
          <p className="text-body-md text-(--ink-primary) mt-1">{order.shipping_address}</p>
        </div>
      </div>

      <div className="mt-4 rounded-md border border-(--border-default) bg-(--surface-card) p-4">
        <p className="text-label-sm text-(--ink-secondary) mb-3">Item Pesanan</p>
        <div className="flex flex-col gap-2">
          {order.items.map((item) => (
            <div key={item.id} className="text-body-sm flex justify-between text-(--ink-secondary)">
              <span>
                {item.product_name} × {item.quantity}
              </span>
              <span className="text-data-md text-(--ink-primary)">
                {formatCurrency(item.price_at_purchase * item.quantity)}
              </span>
            </div>
          ))}
          <div className="text-body-sm flex justify-between border-t border-(--border-default) pt-2 text-(--ink-secondary)">
            <span>Ongkos kirim</span>
            <span className="text-data-md text-(--ink-primary)">{formatCurrency(order.shipping_cost)}</span>
          </div>
          {order.discount_amount > 0 && (
            <div className="text-body-sm flex justify-between text-(--ink-secondary)">
              <span>Diskon</span>
              <span className="text-data-md text-success-500">-{formatCurrency(order.discount_amount)}</span>
            </div>
          )}
          <div className="text-heading-sm text-(--ink-primary) flex justify-between border-t border-(--border-default) pt-2">
            <span>Total</span>
            <span className="text-data-md">{formatCurrency(order.total_amount)}</span>
          </div>
        </div>
      </div>

      <div className="mt-6">
        <p className="text-label-sm text-(--ink-secondary) mb-3">Ubah Status</p>
        {nextStatuses.length === 0 ? (
          <p className="text-body-sm text-(--ink-secondary)">Order sudah dalam status akhir.</p>
        ) : (
          <div className="flex items-center gap-3">
            {nextStatuses.map((status) => (
              <Button
                key={status}
                variant={status === "cancelled" ? "secondary" : "primary"}
                onClick={() => handleStatusChange(status)}
                disabled={updateStatus.isPending}
              >
                {updateStatus.isPending && pendingStatus === status
                  ? "Memproses..."
                  : `Tandai ${statusLabels[status]}`}
              </Button>
            ))}
          </div>
        )}
        {updateStatus.isError && (
          <p role="alert" className="text-body-sm text-error-500 mt-2">
            {updateStatus.error instanceof ApiError ? updateStatus.error.message : "Gagal mengubah status order."}
          </p>
        )}
      </div>
    </div>
  );
}
