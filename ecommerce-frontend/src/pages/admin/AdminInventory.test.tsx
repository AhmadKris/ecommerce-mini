import { screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";

import { AdminInventory } from "./AdminInventory";
import { config } from "../../lib/config";
import { server } from "../../test/msw-server";
import { renderWithProviders } from "../../test/test-utils";

const productsUrl = `${config.apiBaseUrl}/products`;
const adjustmentsUrl = `${config.apiBaseUrl}/admin/inventory/adjustments`;
const movementsUrl = `${config.apiBaseUrl}/admin/inventory/3/movements`;

const sampleProduct = {
  id: 3,
  name: "Kopi Susu",
  slug: "kopi-susu",
  description: "",
  price: 18000,
  stock: 10,
  category_id: 1,
  image_url: "",
  created_at: "2026-01-01T00:00:00Z",
};

function mockProductList() {
  server.use(
    http.get(productsUrl, () =>
      HttpResponse.json({
        success: true,
        data: { items: [sampleProduct], meta: { page: 1, limit: 100, total: 1, total_pages: 1 } },
      }),
    ),
  );
}

describe("AdminInventory page", () => {
  it("submits a stock adjustment and then shows the product's movement history", async () => {
    mockProductList();
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
              before_quantity: 10,
              after_quantity: 20,
              reason: "Restock mingguan",
              reference: "",
              created_by: 1,
              created_at: "2026-01-01T00:00:00Z",
            },
          },
          { status: 201 },
        );
      }),
      http.get(movementsUrl, () =>
        HttpResponse.json({
          success: true,
          data: {
            items: [
              {
                id: 1,
                product_id: 3,
                type: "in",
                quantity: 10,
                before_quantity: 10,
                after_quantity: 20,
                reason: "Restock mingguan",
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
    const user = userEvent.setup();

    renderWithProviders(<AdminInventory />);

    const productSelect = await screen.findByLabelText(/^produk$/i);
    await within(productSelect).findByRole("option", { name: /kopi susu/i });
    await user.selectOptions(productSelect, "3");
    await user.type(screen.getByLabelText(/kuantitas/i), "10");
    await user.type(screen.getByLabelText(/^alasan$/i), "Restock mingguan");
    await user.click(screen.getByRole("button", { name: /terapkan penyesuaian/i }));

    expect(await screen.findByText("Restock mingguan")).toBeInTheDocument();
    expect(receivedBody).toMatchObject({ product_id: 3, type: "in", quantity: 10, reason: "Restock mingguan" });
  });

  it("shows a hint instead of a table before any product is selected", () => {
    mockProductList();

    renderWithProviders(<AdminInventory />);

    expect(screen.getByText(/pilih produk untuk melihat riwayat/i)).toBeInTheDocument();
  });

  it("surfaces the backend's conflict error for an invalid adjustment", async () => {
    mockProductList();
    server.use(
      http.post(adjustmentsUrl, () =>
        HttpResponse.json(
          { success: false, error: { code: "CONFLICT", message: "Stok tidak mencukupi", details: null } },
          { status: 409 },
        ),
      ),
    );
    const user = userEvent.setup();

    renderWithProviders(<AdminInventory />);

    const productSelect = await screen.findByLabelText(/^produk$/i);
    await within(productSelect).findByRole("option", { name: /kopi susu/i });
    await user.selectOptions(productSelect, "3");
    await user.selectOptions(screen.getByLabelText(/tipe penyesuaian/i), "out");
    await user.type(screen.getByLabelText(/kuantitas/i), "9999");
    await user.type(screen.getByLabelText(/^alasan$/i), "Test");
    await user.click(screen.getByRole("button", { name: /terapkan penyesuaian/i }));

    expect(await screen.findByRole("alert")).toHaveTextContent(/tidak mencukupi/i);
  });
});
