import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import type { ReactNode } from "react";
import { describe, expect, it } from "vitest";

import { useCreatePromotion, useDeletePromotion, usePromotions } from "./usePromotions";
import { config } from "../lib/config";
import { server } from "../test/msw-server";

const promotionsUrl = `${config.apiBaseUrl}/admin/promotions`;

function wrapper({ children }: { children: ReactNode }) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
}

describe("usePromotions", () => {
  it("fetches the promotion list", async () => {
    server.use(
      http.get(promotionsUrl, () =>
        HttpResponse.json({
          success: true,
          data: {
            items: [
              {
                id: 1,
                code: "HEMAT10",
                type: "percentage",
                value: 10,
                minimum_purchase: 10000,
                usage_limit: 5,
                used_count: 1,
                starts_at: "2026-01-01T00:00:00Z",
                ends_at: "2026-12-31T00:00:00Z",
                status: "active",
                created_at: "2026-01-01T00:00:00Z",
                updated_at: "2026-01-01T00:00:00Z",
              },
            ],
          },
        }),
      ),
    );

    const { result } = renderHook(() => usePromotions(), { wrapper });

    await waitFor(() => expect(result.current.data).toHaveLength(1));
    expect(result.current.data?.[0].code).toBe("HEMAT10");
  });
});

describe("useCreatePromotion", () => {
  it("posts the new promotion", async () => {
    let receivedBody: unknown;
    server.use(
      http.post(promotionsUrl, async ({ request }) => {
        receivedBody = await request.json();
        return HttpResponse.json(
          { success: true, data: { id: 2, code: "SAVE20", status: "active" } },
          { status: 201 },
        );
      }),
    );

    const { result } = renderHook(() => useCreatePromotion(), { wrapper });
    result.current.mutate({
      code: "SAVE20",
      type: "fixed",
      value: 5000,
      minimum_purchase: 0,
      usage_limit: 10,
      starts_at: "2026-01-01T00:00:00.000Z",
      ends_at: "2026-12-31T00:00:00.000Z",
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(receivedBody).toMatchObject({ code: "SAVE20", type: "fixed", value: 5000 });
  });
});

describe("useDeletePromotion", () => {
  it("surfaces the backend's not-found error", async () => {
    server.use(
      http.delete(`${promotionsUrl}/1`, () =>
        HttpResponse.json(
          { success: false, error: { code: "NOT_FOUND", message: "Promo tidak ditemukan", details: null } },
          { status: 404 },
        ),
      ),
    );

    const { result } = renderHook(() => useDeletePromotion(), { wrapper });
    result.current.mutate(1);

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error?.message).toMatch(/tidak ditemukan/i);
  });
});
