import { Link, useParams } from "react-router-dom";

import { useOrder } from "../../hooks/useOrders";
import { formatCurrency } from "../../lib/format";

// The main forward path an order takes — mirrors orderStatusTransitions on
// the backend, minus "cancelled" (a branch off this line, not a stage on
// it). Used purely to render a timeline; the backend is still the only
// place transition validity is enforced.
const timelineStages = ["pending", "paid", "processing", "shipped", "delivered"];

const statusLabels: Record<string, string> = {
  pending: "Menunggu Pembayaran",
  paid: "Dibayar",
  processing: "Diproses",
  shipped: "Dikirim",
  delivered: "Diterima",
  cancelled: "Dibatalkan",
};

function StatusTimeline({ status }: { status: string }) {
  if (status === "cancelled") {
    return (
      <p className="text-label-sm rounded-full bg-error-100 px-3 py-1 text-error-500 inline-block">
        Pesanan Dibatalkan
      </p>
    );
  }

  const currentIndex = timelineStages.indexOf(status);

  return (
    <ol className="flex items-center">
      {timelineStages.map((stage, index) => {
        const isDone = index <= currentIndex;
        return (
          <li key={stage} className="flex flex-1 items-center last:flex-none">
            <div className="flex flex-col items-center gap-1.5">
              <div
                className={`h-3 w-3 rounded-full ${isDone ? "bg-primary-600" : "bg-neutral-200"}`}
                aria-hidden="true"
              />
              <span
                className={`text-label-sm whitespace-nowrap ${isDone ? "text-(--ink-primary)" : "text-(--ink-secondary)"}`}
              >
                {statusLabels[stage]}
              </span>
            </div>
            {index < timelineStages.length - 1 && (
              <div className={`mx-2 h-0.5 flex-1 ${index < currentIndex ? "bg-primary-600" : "bg-neutral-200"}`} />
            )}
          </li>
        );
      })}
    </ol>
  );
}

export function OrderDetail() {
  const { id = "" } = useParams<{ id: string }>();
  const { data: order, isLoading, isError } = useOrder(Number(id));

  if (isLoading) {
    return (
      <main className="mx-auto max-w-2xl px-6 py-12">
        <p className="text-body-md text-(--ink-secondary)">Memuat pesanan...</p>
      </main>
    );
  }

  if (isError || !order) {
    return (
      <main className="mx-auto max-w-2xl px-6 py-12">
        <p role="alert" className="text-body-md text-error-500">
          Pesanan tidak ditemukan.
        </p>
        <Link to="/orders" className="text-body-sm text-(--ink-link) mt-2 inline-block">
          Kembali ke riwayat pesanan
        </Link>
      </main>
    );
  }

  return (
    <main className="mx-auto max-w-2xl px-6 py-12">
      <Link to="/orders" className="text-body-sm text-(--ink-link)">
        ← Kembali ke riwayat pesanan
      </Link>

      <h1 className="text-heading-xl text-(--ink-primary) mt-4">Order #{order.id}</h1>
      <p className="text-body-sm text-(--ink-secondary) mt-1">
        {new Date(order.created_at).toLocaleDateString("id-ID", { dateStyle: "long" })}
      </p>

      <div className="mt-6 overflow-x-auto rounded-md border border-(--border-default) bg-(--surface-card) p-4">
        <StatusTimeline status={order.status} />
      </div>

      <div className="mt-4 rounded-md border border-(--border-default) bg-(--surface-card) p-4">
        <p className="text-label-sm text-(--ink-secondary)">Alamat Pengiriman</p>
        <p className="text-body-md text-(--ink-primary) mt-1">{order.shipping_address}</p>
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
    </main>
  );
}
