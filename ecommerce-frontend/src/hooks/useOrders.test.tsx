import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import type { ReactNode } from "react";
import { describe, expect, it } from "vitest";

import { useCheckout, useOrders } from "./useOrders";
import { config } from "../lib/config";
import { server } from "../test/msw-server";

const ordersUrl = `${config.apiBaseUrl}/orders`;

function wrapper({ children }: { children: ReactNode }) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
}

describe("useOrders", () => {
  it("fetches the order list", async () => {
    server.use(
      http.get(ordersUrl, () =>
        HttpResponse.json({
          success: true,
          data: { items: [], meta: { page: 1, limit: 10, total: 0, total_pages: 0 } },
        }),
      ),
    );

    const { result } = renderHook(() => useOrders(), { wrapper });

    await waitFor(() => expect(result.current.data?.items).toEqual([]));
  });
});

describe("useCheckout", () => {
  it("posts shipping_address and returns the created order", async () => {
    let receivedBody: unknown;
    server.use(
      http.post(ordersUrl, async ({ request }) => {
        receivedBody = await request.json();
        return HttpResponse.json(
          {
            success: true,
            data: { id: 1, user_id: 1, status: "pending", total_amount: 18000, shipping_address: "Jl. Merdeka No. 1", created_at: "2026-01-01T00:00:00Z", items: [] },
          },
          { status: 201 },
        );
      }),
    );

    const { result } = renderHook(() => useCheckout(), { wrapper });
    result.current.mutate("Jl. Merdeka No. 1");

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(receivedBody).toEqual({ shipping_address: "Jl. Merdeka No. 1" });
    expect(result.current.data?.id).toBe(1);
  });
});
