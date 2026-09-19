import { useState } from "react";

import { Button } from "../../components/ui/Button";
import { useAdminReviews, useUpdateReviewStatus } from "../../hooks/useReviews";
import type { Review } from "../../types/review";

const statusStyles: Record<Review["status"], string> = {
  pending: "bg-warning-100 text-warning-500",
  approved: "bg-success-100 text-success-500",
  rejected: "bg-error-100 text-error-500",
};

const statusLabels: Record<Review["status"], string> = {
  pending: "Menunggu",
  approved: "Disetujui",
  rejected: "Ditolak",
};

function ReviewRow({ review }: { review: Review }) {
  const updateStatus = useUpdateReviewStatus();

  return (
    <tr className="border-t border-(--border-default) align-top">
      <td className="text-body-sm text-(--ink-primary) px-4 py-3">{review.product?.name ?? `#${review.product_id}`}</td>
      <td className="px-4 py-3">
        <p className="text-label-md text-(--ink-primary)">{review.title}</p>
        <p className="text-body-sm text-(--ink-secondary) mt-1 max-w-sm">{review.body}</p>
      </td>
      <td className="text-data-md text-(--ink-primary) px-4 py-3">{review.rating}/5</td>
      <td className="px-4 py-3">
        <span className={`text-label-sm rounded-full px-2.5 py-1 ${statusStyles[review.status]}`}>
          {statusLabels[review.status]}
        </span>
      </td>
      <td className="px-4 py-3">
        {review.status === "pending" ? (
          <div className="flex items-center gap-3">
            <button
              type="button"
              className="text-label-sm text-(--ink-link)"
              onClick={() => updateStatus.mutate({ id: review.id, status: "approved" })}
              disabled={updateStatus.isPending}
            >
              Setujui
            </button>
            <button
              type="button"
              className="text-label-sm text-error-500"
              onClick={() => updateStatus.mutate({ id: review.id, status: "rejected" })}
              disabled={updateStatus.isPending}
            >
              Tolak
            </button>
          </div>
        ) : (
          <span className="text-body-sm text-(--ink-secondary)">—</span>
        )}
      </td>
    </tr>
  );
}

export function AdminReviews() {
  const [page, setPage] = useState(1);
  const { data, isLoading, isError } = useAdminReviews(page);

  return (
    <div className="mx-auto max-w-4xl">
      <h1 className="text-heading-xl text-(--ink-primary) mb-6">Moderasi Ulasan</h1>

      {isLoading && <p className="text-body-md text-(--ink-secondary)">Memuat ulasan...</p>}
      {isError && (
        <p role="alert" className="text-body-md text-error-500">
          Gagal memuat daftar ulasan.
        </p>
      )}
      {data && data.items.length === 0 && (
        <p className="text-body-md text-(--ink-secondary)">Belum ada ulasan.</p>
      )}

      {data && data.items.length > 0 && (
        <>
          <div className="overflow-x-auto rounded-md border border-(--border-default)">
            <table className="w-full text-left">
              <thead className="bg-neutral-50">
                <tr>
                  <th className="text-label-sm text-(--ink-secondary) px-4 py-3">Produk</th>
                  <th className="text-label-sm text-(--ink-secondary) px-4 py-3">Ulasan</th>
                  <th className="text-label-sm text-(--ink-secondary) px-4 py-3">Rating</th>
                  <th className="text-label-sm text-(--ink-secondary) px-4 py-3">Status</th>
                  <th className="text-label-sm text-(--ink-secondary) px-4 py-3">Aksi</th>
                </tr>
              </thead>
              <tbody>
                {data.items.map((review) => (
                  <ReviewRow key={review.id} review={review} />
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
    </div>
  );
}
