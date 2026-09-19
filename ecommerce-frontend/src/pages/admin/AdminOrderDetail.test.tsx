import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { Route, Routes } from "react-router-dom";
import { describe, expect, it } from "vitest";

import { AdminOrderDetail } from "./AdminOrderDetail";
import { config } from "../../lib/config";
import { server } from "../../test/msw-server";
import { renderWithProviders } from "../../test/test-utils";

function renderDetailPage(id: string) {
  return renderWithProviders(
    <Routes>
      <Route path="/admin/orders/:id" element={<AdminOrderDetail />} />
    </Routes>,
    { route: `/admin/orders/${id}` },
  );
}

const sampleOrder = {
  id: 9,
  user_id: 10,
  status: "paid",
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

describe("AdminOrderDetail page", () => {
  it("renders order items, totals, and next-status actions", async () => {
    server.use(http.get(`${config.apiBaseUrl}/admin/orders/9`, () => HttpResponse.json({ success: true, data: sampleOrder })));

    renderDetailPage("9");

    expect(await screen.findByRole("heading", { name: "Order #9" })).toBeInTheDocument();
    expect(screen.getByText(/kopi susu × 1/i)).toBeInTheDocument();
    expect(screen.getByText(/-rp\s?1\.800/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /tandai diproses/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /tandai dibatalkan/i })).toBeInTheDocument();
  });

  it("shows no actions for an order already in a terminal state", async () => {
    server.use(
      http.get(`${config.apiBaseUrl}/admin/orders/9`, () =>
        HttpResponse.json({ success: true, data: { ...sampleOrder, status: "delivered" } }),
      ),
    );

    renderDetailPage("9");

    expect(await screen.findByText(/status akhir/i)).toBeInTheDocument();
  });

  it("submits a status change and surfaces the backend's conflict error", async () => {
    server.use(
      http.get(`${config.apiBaseUrl}/admin/orders/9`, () => HttpResponse.json({ success: true, data: sampleOrder })),
      http.patch(`${config.apiBaseUrl}/admin/orders/9/status`, () =>
        HttpResponse.json(
          { success: false, error: { code: "CONFLICT", message: "Transisi tidak valid", details: null } },
          { status: 409 },
        ),
      ),
    );
    const user = userEvent.setup();

    renderDetailPage("9");

    await user.click(await screen.findByRole("button", { name: /tandai diproses/i }));

    expect(await screen.findByRole("alert")).toHaveTextContent(/transisi tidak valid/i);
  });

  it("shows a not-found state for a 404 from the backend", async () => {
    server.use(
      http.get(`${config.apiBaseUrl}/admin/orders/999`, () =>
        HttpResponse.json(
          { success: false, error: { code: "NOT_FOUND", message: "Order tidak ditemukan", details: null } },
          { status: 404 },
        ),
      ),
    );

    renderDetailPage("999");

    expect(await screen.findByRole("alert")).toHaveTextContent(/order tidak ditemukan/i);
  });
});
