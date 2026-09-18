import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import type { ReactNode } from "react";
import { describe, expect, it } from "vitest";

import { useCategories, useCreateCategory, useDeleteCategory } from "./useCategories";
import { config } from "../lib/config";
import { server } from "../test/msw-server";

const categoriesUrl = `${config.apiBaseUrl}/categories`;
const adminCategoriesUrl = `${config.apiBaseUrl}/admin/categories`;

function wrapper({ children }: { children: ReactNode }) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
}

describe("useCategories", () => {
  it("fetches the category list", async () => {
    server.use(
      http.get(categoriesUrl, () =>
        HttpResponse.json({ success: true, data: { items: [{ id: 1, name: "Minuman", slug: "minuman" }] } }),
      ),
    );

    const { result } = renderHook(() => useCategories(), { wrapper });

    await waitFor(() => expect(result.current.data).toHaveLength(1));
    expect(result.current.data?.[0].name).toBe("Minuman");
  });
});

describe("useCreateCategory", () => {
  it("posts the new category's name", async () => {
    let receivedBody: unknown;
    server.use(
      http.post(adminCategoriesUrl, async ({ request }) => {
        receivedBody = await request.json();
        return HttpResponse.json(
          { success: true, data: { id: 2, name: "Snack", slug: "snack" } },
          { status: 201 },
        );
      }),
    );

    const { result } = renderHook(() => useCreateCategory(), { wrapper });
    result.current.mutate("Snack");

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(receivedBody).toEqual({ name: "Snack" });
  });
});

describe("useDeleteCategory", () => {
  it("surfaces the backend's conflict error when the category is still in use", async () => {
    server.use(
      http.delete(`${adminCategoriesUrl}/1`, () =>
        HttpResponse.json(
          {
            success: false,
            error: { code: "CONFLICT", message: "Kategori masih dipakai oleh produk yang ada", details: null },
          },
          { status: 409 },
        ),
      ),
    );

    const { result } = renderHook(() => useDeleteCategory(), { wrapper });
    result.current.mutate(1);

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error?.message).toMatch(/masih dipakai/i);
  });
});
