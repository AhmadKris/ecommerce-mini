import { useState } from "react";

import { Button } from "../../components/ui/Button";
import { useAdminOrders } from "../../hooks/useAdminOrders";
import { formatCurrency } from "../../lib/format";

export function AdminOrders() {
  const [page, setPage] = useState(1);
  const { data, isLoading, isError } = useAdminOrders(page);

  return (
    <main className="mx-auto max-w-4xl px-6 py-12">
      <h1 className="text-heading-xl text-(--ink-primary) mb-6">Semua Pesanan</h1>

      {isLoading && <p className="text-body-md text-(--ink-secondary)">Memuat pesanan...</p>}
      {isError && (
        <p role="alert" className="text-body-md text-error-500">
          Gagal memuat daftar pesanan.
        </p>
      )}
      {data && data.items.length === 0 && (
        <p className="text-body-md text-(--ink-secondary)">Belum ada pesanan.</p>
      )}

      {data && data.items.length > 0 && (
        <>
          <div className="overflow-x-auto rounded-md border border-(--border-default)">
            <table className="w-full text-left">
              <thead className="bg-neutral-50">
                <tr>
                  <th className="text-label-sm text-(--ink-secondary) px-4 py-3">Order</th>
                  <th className="text-label-sm text-(--ink-secondary) px-4 py-3">User ID</th>
                  <th className="text-label-sm text-(--ink-secondary) px-4 py-3">Tanggal</th>
                  <th className="text-label-sm text-(--ink-secondary) px-4 py-3">Status</th>
                  <th className="text-label-sm text-(--ink-secondary) px-4 py-3">Total</th>
                </tr>
              </thead>
              <tbody>
                {data.items.map((order) => (
                  <tr key={order.id} className="border-t border-(--border-default)">
                    <td className="text-body-sm text-(--ink-primary) px-4 py-3">#{order.id}</td>
                    <td className="text-body-sm text-(--ink-secondary) px-4 py-3">{order.user_id}</td>
                    <td className="text-body-sm text-(--ink-secondary) px-4 py-3">
                      {new Date(order.created_at).toLocaleDateString("id-ID", { dateStyle: "medium" })}
                    </td>
                    <td className="px-4 py-3">
                      <span className="text-label-sm rounded-full bg-warning-100 px-2.5 py-1 text-warning-500">
                        {order.status}
                      </span>
                    </td>
                    <td className="text-data-md text-(--ink-primary) px-4 py-3">
                      {formatCurrency(order.total_amount)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div className="mt-6 flex items-center gap-3">
            <Button variant="secondary" onClick={() => setPage((p) => p - 1)} disabled={page <= 1}>
              Sebelumnya
            </Button>
            <span className="text-body-sm text-(--ink-secondary)">
              Halaman {data.meta.page} dari {data.meta.total_pages}
            </span>
            <Button
              variant="secondary"
              onClick={() => setPage((p) => p + 1)}
              disabled={page >= data.meta.total_pages}
            >
              Berikutnya
            </Button>
          </div>
        </>
      )}
    </main>
  );
}
