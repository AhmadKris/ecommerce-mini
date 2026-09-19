import { screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { Route, Routes } from "react-router-dom";
import { describe, expect, it } from "vitest";

import { OrderDetail } from "./OrderDetail";
import { config } from "../../lib/config";
import { server } from "../../test/msw-server";
import { renderWithProviders } from "../../test/test-utils";

function renderDetailPage(id: string) {
  return renderWithProviders(
    <Routes>
      <Route path="/orders/:id" element={<OrderDetail />} />
    </Routes>,
    { route: `/orders/${id}` },
  );
}

const sampleOrder = {
  id: 9,
  user_id: 1,
  status: "shipped",
  total_amount: 41200,
  shipping_cost: 25000,
  discount_amount: 1800,
  promotion_id: 1,
  shipping_address: "Jl. Merdeka No. 1",
  created_at: "2026-01-01T00:00:00Z",
  items: [
    { id: 1, order_id: 9, product_id: 3, product_name: "Kopi Susu", quantity: 1, price_at_purchase: 18000 },
  ],
};

describe("OrderDetail page", () => {
  it("renders the shipping address, items, totals, and status timeline", async () => {
    server.use(http.get(`${config.apiBaseUrl}/orders/9`, () => HttpResponse.json({ success: true, data: sampleOrder })));

    renderDetailPage("9");

    expect(await screen.findByRole("heading", { name: "Order #9" })).toBeInTheDocument();
    expect(screen.getByText("Jl. Merdeka No. 1")).toBeInTheDocument();
    expect(screen.getByText(/kopi susu × 1/i)).toBeInTheDocument();
    expect(screen.getByText(/-rp\s?1\.800/i)).toBeInTheDocument();
    expect(screen.getByText("Dikirim")).toBeInTheDocument();
  });

  it("renders a cancelled badge instead of the timeline for a cancelled order", async () => {
    server.use(
      http.get(`${config.apiBaseUrl}/orders/9`, () =>
        HttpResponse.json({ success: true, data: { ...sampleOrder, status: "cancelled" } }),
      ),
    );

    renderDetailPage("9");

    expect(await screen.findByText(/pesanan dibatalkan/i)).toBeInTheDocument();
  });

  it("shows a not-found state for another user's order", async () => {
    server.use(
      http.get(`${config.apiBaseUrl}/orders/1`, () =>
        HttpResponse.json(
          { success: false, error: { code: "NOT_FOUND", message: "Order tidak ditemukan", details: null } },
          { status: 404 },
        ),
      ),
    );

    renderDetailPage("1");

    expect(await screen.findByRole("alert")).toHaveTextContent(/pesanan tidak ditemukan/i);
  });
});
