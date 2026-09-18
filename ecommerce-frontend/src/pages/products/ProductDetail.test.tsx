import { screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { Route, Routes } from "react-router-dom";
import { describe, expect, it } from "vitest";

import { ProductDetail } from "./ProductDetail";
import { config } from "../../lib/config";
import { server } from "../../test/msw-server";
import { renderWithProviders } from "../../test/test-utils";

function renderDetailPage(slug: string) {
  return renderWithProviders(
    <Routes>
      <Route path="/products/:slug" element={<ProductDetail />} />
    </Routes>,
    { route: `/products/${slug}` },
  );
}

describe("ProductDetail page", () => {
  it("renders the product's name, price, and description once loaded", async () => {
    server.use(
      http.get(`${config.apiBaseUrl}/products/kopi-susu-gula-aren`, () =>
        HttpResponse.json({
          success: true,
          data: {
            id: 1,
            name: "Kopi Susu Gula Aren",
            slug: "kopi-susu-gula-aren",
            description: "Kopi susu dengan gula aren asli.",
            price: 18000,
            stock: 5,
            category_id: 1,
            image_url: "",
            created_at: "2026-01-01T00:00:00Z",
            category: { id: 1, name: "Minuman", slug: "minuman" },
          },
        }),
      ),
    );

    renderDetailPage("kopi-susu-gula-aren");

    expect(await screen.findByRole("heading", { name: "Kopi Susu Gula Aren" })).toBeInTheDocument();
    expect(screen.getByText(/kopi susu dengan gula aren asli/i)).toBeInTheDocument();
    expect(screen.getByText("Minuman")).toBeInTheDocument();
  });

  it("shows a not-found state for a 404 from the backend", async () => {
    server.use(
      http.get(`${config.apiBaseUrl}/products/tidak-ada`, () =>
        HttpResponse.json(
          {
            success: false,
            error: { code: "NOT_FOUND", message: "Produk tidak ditemukan", details: null },
          },
          { status: 404 },
        ),
      ),
    );

    renderDetailPage("tidak-ada");

    expect(await screen.findByRole("alert")).toHaveTextContent(/produk tidak ditemukan/i);
  });

  it("shows out-of-stock instead of a stock count when stock is 0", async () => {
    server.use(
      http.get(`${config.apiBaseUrl}/products/habis`, () =>
        HttpResponse.json({
          success: true,
          data: {
            id: 2,
            name: "Produk Habis",
            slug: "habis",
            description: "",
            price: 5000,
            stock: 0,
            category_id: 1,
            image_url: "",
            created_at: "2026-01-01T00:00:00Z",
          },
        }),
      ),
    );

    renderDetailPage("habis");

    // "Stok habis" appears twice once stock is 0: the info line and the
    // disabled AddToCartButton — assert both rather than a single ambiguous
    // text match.
    expect(await screen.findAllByText(/stok habis/i)).toHaveLength(2);
    expect(screen.getByRole("button", { name: /stok habis/i })).toBeDisabled();
  });
});
