import { screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";

import { OrderHistory } from "./OrderHistory";
import { config } from "../../lib/config";
import { renderWithProviders } from "../../test/test-utils";
import { server } from "../../test/msw-server";

const ordersUrl = `${config.apiBaseUrl}/orders`;

describe("OrderHistory page", () => {
  it("shows an empty state when there are no orders", async () => {
    server.use(
      http.get(ordersUrl, () =>
        HttpResponse.json({
          success: true,
          data: { items: [], meta: { page: 1, limit: 10, total: 0, total_pages: 0 } },
        }),
      ),
    );

    renderWithProviders(<OrderHistory />);

    expect(await screen.findByText(/belum ada pesanan/i)).toBeInTheDocument();
  });

  it("renders each order's id, status, and total", async () => {
    server.use(
      http.get(ordersUrl, () =>
        HttpResponse.json({
          success: true,
          data: {
            items: [
              {
                id: 7,
                user_id: 1,
                status: "pending",
                total_amount: 43000,
                shipping_cost: 25000,
                shipping_address: "Jl. Merdeka No. 1",
                created_at: "2026-01-01T00:00:00Z",
                items: [
                  {
                    id: 1,
                    order_id: 7,
                    product_id: 1,
                    product_name: "Kopi Susu Gula Aren",
                    quantity: 1,
                    price_at_purchase: 18000,
                    product: {},
                  },
                ],
              },
            ],
            meta: { page: 1, limit: 10, total: 1, total_pages: 1 },
          },
        }),
      ),
    );

    renderWithProviders(<OrderHistory />);

    expect(await screen.findByText("Order #7")).toBeInTheDocument();
    expect(screen.getByText("pending")).toBeInTheDocument();
    expect(screen.getByText(/43\.000/)).toBeInTheDocument();
  });

  it("shows an error state when the request fails", async () => {
    server.use(
      http.get(ordersUrl, () =>
        HttpResponse.json(
          { success: false, error: { code: "INTERNAL_ERROR", message: "err", details: null } },
          { status: 500 },
        ),
      ),
    );

    renderWithProviders(<OrderHistory />);

    expect(await screen.findByRole("alert")).toHaveTextContent(/gagal memuat riwayat pesanan/i);
  });
});
