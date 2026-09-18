import { screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";

import { ProductList } from "./ProductList";
import { config } from "../../lib/config";
import { server } from "../../test/msw-server";
import { renderWithProviders } from "../../test/test-utils";

const productsUrl = `${config.apiBaseUrl}/products`;

const sampleProduct = {
  id: 1,
  name: "Kopi Susu Gula Aren",
  slug: "kopi-susu-gula-aren",
  description: "",
  price: 18000,
  stock: 10,
  category_id: 1,
  image_url: "",
  created_at: "2026-01-01T00:00:00Z",
};

describe("ProductList page", () => {
  it("renders the product grid once loaded", async () => {
    server.use(
      http.get(productsUrl, () =>
        HttpResponse.json({
          success: true,
          data: { items: [sampleProduct], meta: { page: 1, limit: 10, total: 1, total_pages: 1 } },
        }),
      ),
    );

    renderWithProviders(<ProductList />);

    expect(await screen.findByText("Kopi Susu Gula Aren")).toBeInTheDocument();
  });

  it("shows an empty state when there are no products", async () => {
    server.use(
      http.get(productsUrl, () =>
        HttpResponse.json({
          success: true,
          data: { items: [], meta: { page: 1, limit: 10, total: 0, total_pages: 0 } },
        }),
      ),
    );

    renderWithProviders(<ProductList />);

    expect(await screen.findByText(/belum ada produk/i)).toBeInTheDocument();
  });

  it("shows an error state when the request fails", async () => {
    server.use(
      http.get(productsUrl, () =>
        HttpResponse.json(
          {
            success: false,
            error: {
              code: "INTERNAL_ERROR",
              message: "Terjadi kesalahan pada server",
              details: null,
            },
          },
          { status: 500 },
        ),
      ),
    );

    renderWithProviders(<ProductList />);

    expect(await screen.findByRole("alert")).toHaveTextContent(/gagal memuat produk/i);
  });

  it("disables the previous-page button on the first page", async () => {
    server.use(
      http.get(productsUrl, () =>
        HttpResponse.json({
          success: true,
          data: { items: [sampleProduct], meta: { page: 1, limit: 10, total: 20, total_pages: 2 } },
        }),
      ),
    );

    renderWithProviders(<ProductList />);

    expect(await screen.findByRole("button", { name: /sebelumnya/i })).toBeDisabled();
    expect(screen.getByRole("button", { name: /berikutnya/i })).toBeEnabled();
  });
});
