import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";

import { ProductList } from "./ProductList";
import { config } from "../../lib/config";
import { server } from "../../test/msw-server";
import { renderWithProviders } from "../../test/test-utils";

const productsUrl = `${config.apiBaseUrl}/products`;
const categoriesUrl = `${config.apiBaseUrl}/categories`;

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

beforeEach(() => {
  server.use(
    http.get(categoriesUrl, () =>
      HttpResponse.json({ success: true, data: { items: [{ id: 1, name: "Minuman", slug: "minuman" }] } }),
    ),
  );
});

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

  it("renders category pills and sends the selected category to the API", async () => {
    let receivedCategory: string | null = null;
    server.use(
      http.get(productsUrl, ({ request }) => {
        receivedCategory = new URL(request.url).searchParams.get("category");
        return HttpResponse.json({
          success: true,
          data: { items: [sampleProduct], meta: { page: 1, limit: 10, total: 1, total_pages: 1 } },
        });
      }),
    );
    const user = userEvent.setup();

    renderWithProviders(<ProductList />);

    const categoryPill = await screen.findByRole("button", { name: "Minuman" });
    await user.click(categoryPill);

    await screen.findByText("Kopi Susu Gula Aren");
    expect(receivedCategory).toBe("minuman");
  });

  it("sends the selected sort option to the API", async () => {
    let receivedSort: string | null = null;
    server.use(
      http.get(productsUrl, ({ request }) => {
        receivedSort = new URL(request.url).searchParams.get("sort");
        return HttpResponse.json({
          success: true,
          data: { items: [sampleProduct], meta: { page: 1, limit: 10, total: 1, total_pages: 1 } },
        });
      }),
    );
    const user = userEvent.setup();

    renderWithProviders(<ProductList />);
    await screen.findByText("Kopi Susu Gula Aren");

    await user.selectOptions(screen.getByLabelText(/urutkan/i), "price_asc");

    await screen.findByText("Kopi Susu Gula Aren");
    expect(receivedSort).toBe("price_asc");
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
