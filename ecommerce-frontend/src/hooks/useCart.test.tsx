import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import type { ReactNode } from "react";
import { afterEach, describe, expect, it } from "vitest";

import { useAddCartItem, useCart, useRemoveCartItem, useUpdateCartItem } from "./useCart";
import { config } from "../lib/config";
import { useAuthStore } from "../store/auth-store";
import { fakeAccessToken } from "../test/fake-jwt";
import { server } from "../test/msw-server";

const cartUrl = `${config.apiBaseUrl}/cart`;

function wrapper({ children }: { children: ReactNode }) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
}

afterEach(() => useAuthStore.getState().clearSession());

function logIn() {
  useAuthStore.getState().setSession({ access_token: fakeAccessToken(), refresh_token: "r", expires_in: 900 });
}

describe("useCart", () => {
  it("does not fetch when logged out", () => {
    const { result } = renderHook(() => useCart(), { wrapper });
    expect(result.current.fetchStatus).toBe("idle");
  });

  it("fetches the cart when logged in", async () => {
    logIn();
    server.use(
      http.get(cartUrl, () => HttpResponse.json({ success: true, data: { items: [], total: 0 } })),
    );

    const { result } = renderHook(() => useCart(), { wrapper });

    await waitFor(() => expect(result.current.data).toEqual({ items: [], total: 0 }));
  });
});

describe("cart mutations", () => {
  it("useAddCartItem posts product_id/quantity and returns the updated cart", async () => {
    logIn();
    let receivedBody: unknown;
    server.use(
      http.post(`${cartUrl}/items`, async ({ request }) => {
        receivedBody = await request.json();
        return HttpResponse.json({
          success: true,
          data: { items: [{ id: 1, product: { id: 5 }, quantity: 2, subtotal: 100 }], total: 100 },
        });
      }),
    );

    const { result } = renderHook(() => useAddCartItem(), { wrapper });
    result.current.mutate({ productId: 5, quantity: 2 });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(receivedBody).toEqual({ product_id: 5, quantity: 2 });
    expect(result.current.data?.total).toBe(100);
  });

  it("useUpdateCartItem PUTs the new quantity", async () => {
    logIn();
    let receivedBody: unknown;
    server.use(
      http.put(`${cartUrl}/items/1`, async ({ request }) => {
        receivedBody = await request.json();
        return HttpResponse.json({ success: true, data: { items: [], total: 0 } });
      }),
    );

    const { result } = renderHook(() => useUpdateCartItem(), { wrapper });
    result.current.mutate({ itemId: 1, quantity: 3 });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(receivedBody).toEqual({ quantity: 3 });
  });

  it("useRemoveCartItem DELETEs the item", async () => {
    logIn();
    let called = false;
    server.use(
      http.delete(`${cartUrl}/items/1`, () => {
        called = true;
        return HttpResponse.json({ success: true, data: { items: [], total: 0 } });
      }),
    );

    const { result } = renderHook(() => useRemoveCartItem(), { wrapper });
    result.current.mutate(1);

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(called).toBe(true);
  });
});
