import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { Link } from "react-router-dom";

import { Button } from "../ui/Button";
import { Input } from "../ui/Input";
import { useCreateReview, useProductReviews } from "../../hooks/useReviews";
import { reviewSchema, type ReviewFormInput, type ReviewFormValues } from "../../schemas/review";
import { useAuthStore } from "../../store/auth-store";
import { ApiError } from "../../types/api";

function ReviewForm({ slug }: { slug: string }) {
  const createReview = useCreateReview(slug);
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<ReviewFormInput, unknown, ReviewFormValues>({
    resolver: zodResolver(reviewSchema),
    defaultValues: { rating: 5 },
  });

  if (createReview.isSuccess) {
    return (
      <p className="text-body-sm text-success-500 rounded-md border border-(--border-default) bg-(--surface-card) p-4">
        Terima kasih! Ulasan kamu menunggu persetujuan admin sebelum tampil di sini.
      </p>
    );
  }

  return (
    <form
      onSubmit={handleSubmit((values) => createReview.mutate(values, { onSuccess: () => reset() }))}
      className="flex flex-col gap-4 rounded-md border border-(--border-default) bg-(--surface-card) p-4"
      noValidate
    >
      <div className="flex flex-col gap-1">
        <label className="text-label-md text-(--ink-primary)" htmlFor="review-rating">
          Rating
        </label>
        <select
          id="review-rating"
          className="text-body-md w-32 rounded-md border border-(--border-default) px-3 py-2.5 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/30"
          {...register("rating")}
        >
          {[5, 4, 3, 2, 1].map((rating) => (
            <option key={rating} value={rating}>
              {rating} bintang
            </option>
          ))}
        </select>
      </div>
      <Input label="Judul" error={errors.title?.message} {...register("title")} />
      <div className="flex flex-col gap-1">
        <label className="text-label-md text-(--ink-primary)" htmlFor="review-body">
          Ulasan
        </label>
        <textarea
          id="review-body"
          rows={3}
          className="text-body-md rounded-md border border-(--border-default) px-3 py-2.5 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/30"
          {...register("body")}
        />
        {errors.body?.message && <p className="text-body-sm text-error-500">{errors.body.message}</p>}
      </div>
      {createReview.isError && (
        <p role="alert" className="text-body-sm text-error-500">
          {createReview.error instanceof ApiError
            ? createReview.error.message
            : "Gagal mengirim ulasan."}
        </p>
      )}
      <Button type="submit" disabled={createReview.isPending} className="self-start">
        {createReview.isPending ? "Mengirim..." : "Kirim Ulasan"}
      </Button>
    </form>
  );
}

/**
 * Approved reviews for a product plus, for a logged-in customer, a form to
 * submit one. Eligibility (must have a delivered order for this product) is
 * enforced server-side, not pre-checked here — the form is always offered to
 * a logged-in user, and a 403 from the backend surfaces as the form's error
 * message instead.
 */
export function ProductReviews({ slug }: { slug: string }) {
  const { data, isLoading } = useProductReviews(slug);
  const userId = useAuthStore((state) => state.userId);

  return (
    <section className="mt-12">
      <h2 className="text-heading-lg text-(--ink-primary) mb-4">Ulasan</h2>

      {userId ? (
        <div className="mb-6">
          <ReviewForm slug={slug} />
        </div>
      ) : (
        <p className="text-body-sm text-(--ink-secondary) mb-6">
          <Link to="/login" className="text-(--ink-link)">
            Masuk
          </Link>{" "}
          untuk memberi ulasan setelah pesananmu diterima.
        </p>
      )}

      {isLoading && <p className="text-body-md text-(--ink-secondary)">Memuat ulasan...</p>}
      {data && data.items.length === 0 && (
        <p className="text-body-md text-(--ink-secondary)">Belum ada ulasan untuk produk ini.</p>
      )}
      {data && data.items.length > 0 && (
        <ul className="flex flex-col gap-4">
          {data.items.map((review) => (
            <li key={review.id} className="rounded-md border border-(--border-default) bg-(--surface-card) p-4">
              <div className="flex items-center justify-between">
                <p className="text-label-md text-(--ink-primary)">{review.title}</p>
                <span className="text-data-md text-(--ink-primary)">{review.rating}/5</span>
              </div>
              <p className="text-body-sm text-(--ink-secondary) mt-1">{review.body}</p>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
