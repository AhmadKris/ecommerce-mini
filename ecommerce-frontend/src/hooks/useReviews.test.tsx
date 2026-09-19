import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import type { ReactNode } from "react";
import { describe, expect, it } from "vitest";

import { useCreateReview, useProductReviews, useUpdateReviewStatus } from "./useReviews";
import { config } from "../lib/config";
import { server } from "../test/msw-server";

const reviewsUrl = `${config.apiBaseUrl}/products/kopi-hitam/reviews`;
const adminReviewsUrl = `${config.apiBaseUrl}/admin/reviews`;

function wrapper({ children }: { children: ReactNode }) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
}

describe("useProductReviews", () => {
  it("fetches a product's approved review list", async () => {
    server.use(
      http.get(reviewsUrl, () =>
        HttpResponse.json({
          success: true,
          data: {
            items: [
              {
                id: 1,
                product_id: 1,
                user_id: 1,
                order_id: 1,
                rating: 5,
                title: "Enak",
                body: "Kopinya enak sekali.",
                status: "approved",
                created_at: "2026-01-01T00:00:00Z",
                updated_at: "2026-01-01T00:00:00Z",
              },
            ],
            meta: { page: 1, limit: 10, total: 1, total_pages: 1 },
          },
        }),
      ),
    );

    const { result } = renderHook(() => useProductReviews("kopi-hitam"), { wrapper });

    await waitFor(() => expect(result.current.data?.items).toHaveLength(1));
    expect(result.current.data?.items[0].title).toBe("Enak");
  });
});

describe("useCreateReview", () => {
  it("rejects with the backend's forbidden error when the caller never received the product", async () => {
    server.use(
      http.post(reviewsUrl, () =>
        HttpResponse.json(
          {
            success: false,
            error: {
              code: "FORBIDDEN",
              message: "Kamu hanya bisa memberi ulasan untuk produk yang sudah diterima",
              details: null,
            },
          },
          { status: 403 },
        ),
      ),
    );

    const { result } = renderHook(() => useCreateReview("kopi-hitam"), { wrapper });
    result.current.mutate({ rating: 5, title: "Bagus", body: "Testing pengiriman ulasan produk." });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error?.message).toMatch(/sudah diterima/i);
  });
});

describe("useUpdateReviewStatus", () => {
  it("patches the review's moderation status", async () => {
    let receivedBody: unknown;
    server.use(
      http.patch(`${adminReviewsUrl}/1/status`, async ({ request }) => {
        receivedBody = await request.json();
        return HttpResponse.json({ success: true, data: { id: 1, status: "approved" } });
      }),
    );

    const { result } = renderHook(() => useUpdateReviewStatus(), { wrapper });
    result.current.mutate({ id: 1, status: "approved" });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(receivedBody).toEqual({ status: "approved" });
  });
});
