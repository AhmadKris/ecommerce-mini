import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { apiClient } from "../lib/api-client";
import type { PaginatedResponse } from "../types/api";
import type { Review } from "../types/review";

/** Lists a product's approved reviews (public). */
export function useProductReviews(slug: string, page = 1) {
  return useQuery({
    queryKey: ["reviews", slug, { page }],
    queryFn: () =>
      apiClient
        .get<PaginatedResponse<Review>>(`/products/${slug}/reviews`, { params: { page } })
        .then((response) => response.data),
    enabled: slug.length > 0,
  });
}

export interface ReviewInput {
  rating: number;
  title: string;
  body: string;
}

/** Submits a review for a product (auth required — see backend for the
 * delivered-order eligibility check enforced server-side). */
export function useCreateReview(slug: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: ReviewInput) =>
      apiClient.post<Review>(`/products/${slug}/reviews`, input).then((r) => r.data),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ["reviews", slug] }),
  });
}

/** Lists every review regardless of status, for admin moderation. */
export function useAdminReviews(page = 1) {
  return useQuery({
    queryKey: ["admin-reviews", { page }],
    queryFn: () =>
      apiClient
        .get<PaginatedResponse<Review>>("/admin/reviews", { params: { page } })
        .then((response) => response.data),
  });
}

/** Approves or rejects a review (admin, needs review:moderate). */
export function useUpdateReviewStatus() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: number; status: "approved" | "rejected" }) =>
      apiClient.patch<Review>(`/admin/reviews/${id}/status`, { status }).then((r) => r.data),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ["admin-reviews"] }),
  });
}
