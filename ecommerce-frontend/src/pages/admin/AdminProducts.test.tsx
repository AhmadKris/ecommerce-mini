import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { Route, Routes } from "react-router-dom";
import { describe, expect, it } from "vitest";

import { AdminProducts } from "./AdminProducts";
import { config } from "../../lib/config";
import { server } from "../../test/msw-server";
import { renderWithProviders } from "../../test/test-utils";

function renderListPage() {
  return renderWithProviders(
    <Routes>
      <Route path="/admin/products" element={<AdminProducts />} />
    </Routes>,
    { route: "/admin/products" },
  );
}

const sampleProduct = {
  id: 1,
  name: "Kopi Susu",
  slug: "kopi-susu",
  sku: "KOPI-001",
  description: "",
  price: 18000,
  stock: 10,
  category_id: 1,
  image_url: "",
  created_at: "2026-01-01T00:00:00Z",
  category: { id: 1, name: "Minuman", slug: "minuman" },
};

describe("AdminProducts page", () => {
  it("renders the SKU column, category filter, and search box", async () => {
    server.use(
      http.get(`${config.apiBaseUrl}/products`, () =>
        HttpResponse.json({
          success: true,
          data: { items: [sampleProduct], meta: { page: 1, limit: 10, total: 1, total_pages: 1 } },
        }),
      ),
      http.get(`${config.apiBaseUrl}/categories`, () =>
        HttpResponse.json({ success: true, data: { items: [{ id: 1, name: "Minuman", slug: "minuman" }] } }),
      ),
    );

    renderListPage();

    expect(await screen.findByText("KOPI-001")).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "SKU" })).toBeInTheDocument();
    expect(screen.getByLabelText(/cari produk/i)).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "Minuman" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /export/i })).toBeEnabled();
  });

  it("re-fetches with the search param when typing in the search box", async () => {
    let capturedSearch: string | null = null;
    server.use(
      http.get(`${config.apiBaseUrl}/products`, ({ request }) => {
        capturedSearch = new URL(request.url).searchParams.get("search");
        return HttpResponse.json({
          success: true,
          data: { items: [sampleProduct], meta: { page: 1, limit: 10, total: 1, total_pages: 1 } },
        });
      }),
      http.get(`${config.apiBaseUrl}/categories`, () =>
        HttpResponse.json({ success: true, data: { items: [] } }),
      ),
    );
    const user = userEvent.setup();

    renderListPage();
    await screen.findByText("Kopi Susu");

    await user.type(screen.getByLabelText(/cari produk/i), "kopi");

    await waitFor(() => expect(capturedSearch).toBe("kopi"));
  });

  it("disables Export when there are no products", async () => {
    server.use(
      http.get(`${config.apiBaseUrl}/products`, () =>
        HttpResponse.json({
          success: true,
          data: { items: [], meta: { page: 1, limit: 10, total: 0, total_pages: 0 } },
        }),
      ),
      http.get(`${config.apiBaseUrl}/categories`, () =>
        HttpResponse.json({ success: true, data: { items: [] } }),
      ),
    );

    renderListPage();

    expect(await screen.findByText(/belum ada produk/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /export/i })).toBeDisabled();
  });
});
