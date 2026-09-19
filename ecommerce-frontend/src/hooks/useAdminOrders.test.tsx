import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import type { ReactNode } from "react";
import { describe, expect, it } from "vitest";

import { useAdminOrder, useUpdateOrderStatus } from "./useAdminOrders";
import { config } from "../lib/config";
import { server } from "../test/msw-server";

const orderUrl = `${config.apiBaseUrl}/admin/orders/9`;

function wrapper({ children }: { children: ReactNode }) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
}

describe("useAdminOrder", () => {
  it("fetches a single order with its items", async () => {
    server.use(
      http.get(orderUrl, () =>
        HttpResponse.json({
          success: true,
          data: {
            id: 9,
            user_id: 10,
            status: "delivered",
            total_amount: 50000,
            shipping_cost: 25000,
            discount_amount: 0,
            promotion_id: null,
            shipping_address: "Jl. Test No 1",
            created_at: "2026-01-01T00:00:00Z",
            items: [],
          },
        }),
      ),
    );

    const { result } = renderHook(() => useAdminOrder(9), { wrapper });

    await waitFor(() => expect(result.current.data?.id).toBe(9));
    expect(result.current.data?.status).toBe("delivered");
  });
});

describe("useUpdateOrderStatus", () => {
  it("rejects an invalid transition with the backend's conflict error", async () => {
    server.use(
      http.patch(`${orderUrl}/status`, () =>
        HttpResponse.json(
          {
            success: false,
            error: {
              code: "CONFLICT",
              message: 'Order dengan status "delivered" tidak bisa diubah ke "pending"',
              details: null,
            },
          },
          { status: 409 },
        ),
      ),
    );

    const { result } = renderHook(() => useUpdateOrderStatus(9), { wrapper });
    result.current.mutate("pending");

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error?.message).toMatch(/tidak bisa diubah/i);
  });
});
