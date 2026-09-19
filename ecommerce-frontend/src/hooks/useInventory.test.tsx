import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import type { ReactNode } from "react";
import { describe, expect, it } from "vitest";

import { useAdjustInventory, useInventoryMovements } from "./useInventory";
import { config } from "../lib/config";
import { server } from "../test/msw-server";

const adjustmentsUrl = `${config.apiBaseUrl}/admin/inventory/adjustments`;
const movementsUrl = `${config.apiBaseUrl}/admin/inventory/3/movements`;

function wrapper({ children }: { children: ReactNode }) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
}

describe("useAdjustInventory", () => {
  it("posts the adjustment and returns the resulting movement", async () => {
    let receivedBody: unknown;
    server.use(
      http.post(adjustmentsUrl, async ({ request }) => {
        receivedBody = await request.json();
        return HttpResponse.json(
          {
            success: true,
            data: {
              id: 1,
              product_id: 3,
              type: "in",
              quantity: 10,
              before_quantity: 5,
              after_quantity: 15,
              reason: "Restock",
              reference: "",
              created_by: 1,
              created_at: "2026-01-01T00:00:00Z",
            },
          },
          { status: 201 },
        );
      }),
    );

    const { result } = renderHook(() => useAdjustInventory(), { wrapper });
    result.current.mutate({ product_id: 3, type: "in", quantity: 10, reason: "Restock" });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(receivedBody).toMatchObject({ product_id: 3, type: "in", quantity: 10 });
    expect(result.current.data?.after_quantity).toBe(15);
  });

  it("surfaces the backend's conflict error when an out-adjustment would go negative", async () => {
    server.use(
      http.post(adjustmentsUrl, () =>
        HttpResponse.json(
          { success: false, error: { code: "CONFLICT", message: "Stok tidak mencukupi", details: null } },
          { status: 409 },
        ),
      ),
    );

    const { result } = renderHook(() => useAdjustInventory(), { wrapper });
    result.current.mutate({ product_id: 3, type: "out", quantity: 9999, reason: "Test" });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error?.message).toMatch(/tidak mencukupi/i);
  });
});

describe("useInventoryMovements", () => {
  it("fetches a product's movement history", async () => {
    server.use(
      http.get(movementsUrl, () =>
        HttpResponse.json({
          success: true,
          data: {
            items: [
              {
                id: 1,
                product_id: 3,
                type: "correction",
                quantity: 50,
                before_quantity: 15,
                after_quantity: 50,
                reason: "Stock opname",
                reference: "",
                created_by: 1,
                created_at: "2026-01-01T00:00:00Z",
              },
            ],
            meta: { page: 1, limit: 10, total: 1, total_pages: 1 },
          },
        }),
      ),
    );

    const { result } = renderHook(() => useInventoryMovements(3), { wrapper });

    await waitFor(() => expect(result.current.data?.items).toHaveLength(1));
    expect(result.current.data?.items[0].after_quantity).toBe(50);
  });
});
