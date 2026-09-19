import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import type { ReactNode } from "react";
import { describe, expect, it } from "vitest";

import { useAdminCustomer, useAdminCustomers } from "./useAdminCustomers";
import { config } from "../lib/config";
import { server } from "../test/msw-server";

function wrapper({ children }: { children: ReactNode }) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
}

describe("useAdminCustomers", () => {
  it("fetches a page of customers filtered by search", async () => {
    server.use(
      http.get(`${config.apiBaseUrl}/admin/customers`, ({ request }) => {
        const url = new URL(request.url);
        expect(url.searchParams.get("search")).toBe("budi");
        return HttpResponse.json({
          success: true,
          data: {
            items: [{ id: 1, name: "Budi", email: "budi@example.com", created_at: "2026-01-01T00:00:00Z", roles: [] }],
            meta: { page: 1, limit: 10, total: 1, total_pages: 1 },
          },
        });
      }),
    );

    const { result } = renderHook(() => useAdminCustomers(1, "budi"), { wrapper });

    await waitFor(() => expect(result.current.data?.items).toHaveLength(1));
    expect(result.current.data?.items[0].name).toBe("Budi");
  });
});

describe("useAdminCustomer", () => {
  it("fetches a single customer with roles and permissions", async () => {
    server.use(
      http.get(`${config.apiBaseUrl}/admin/customers/1`, () =>
        HttpResponse.json({
          success: true,
          data: {
            id: 1,
            name: "Budi",
            email: "budi@example.com",
            created_at: "2026-01-01T00:00:00Z",
            roles: [{ id: 1, name: "admin", permissions: [{ id: 1, code: "customer:read" }] }],
          },
        }),
      ),
    );

    const { result } = renderHook(() => useAdminCustomer(1), { wrapper });

    await waitFor(() => expect(result.current.data?.id).toBe(1));
    expect(result.current.data?.roles[0].permissions?.[0].code).toBe("customer:read");
  });
});
