import { Link } from "react-router-dom";

import { useOrders } from "../../hooks/useOrders";
import { formatCurrency } from "../../lib/format";

export function OrderHistory() {
  const { data, isLoading, isError } = useOrders();

  if (isLoading) {
    return (
      <main className="mx-auto max-w-3xl px-6 py-12">
        <p className="text-body-md text-(--ink-secondary)">Memuat riwayat pesanan...</p>
      </main>
    );
  }

  if (isError || !data) {
    return (
      <main className="mx-auto max-w-3xl px-6 py-12">
        <p role="alert" className="text-body-md text-error-500">
          Gagal memuat riwayat pesanan.
        </p>
      </main>
    );
  }

  if (data.items.length === 0) {
    return (
      <main className="mx-auto max-w-3xl px-6 py-12">
        <h1 className="text-heading-xl text-(--ink-primary)">Belum ada pesanan</h1>
        <Link to="/products" className="text-body-sm text-(--ink-link) mt-2 inline-block">
          Lihat produk
        </Link>
      </main>
    );
  }

  return (
    <main className="mx-auto max-w-3xl px-6 py-12">
      <h1 className="text-heading-xl text-(--ink-primary) mb-6">Riwayat Pesanan</h1>
      <div className="flex flex-col gap-4">
        {data.items.map((order) => (
          <Link
            key={order.id}
            to={`/orders/${order.id}`}
            className="block rounded-md border border-(--border-default) bg-(--surface-card) p-4 hover:bg-neutral-50"
          >
            <div className="flex items-center justify-between">
              <span className="text-heading-sm text-(--ink-primary)">Order #{order.id}</span>
              <span className="text-label-sm rounded-full bg-warning-100 px-2.5 py-1 text-warning-500">
                {order.status}
              </span>
            </div>
            <p className="text-body-sm text-(--ink-secondary) mt-1">
              {new Date(order.created_at).toLocaleDateString("id-ID", { dateStyle: "long" })} ·{" "}
              {order.items.length} produk
            </p>
            <p className="text-data-md text-(--ink-primary) mt-2">{formatCurrency(order.total_amount)}</p>
          </Link>
        ))}
      </div>
    </main>
  );
}
