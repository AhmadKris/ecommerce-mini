import { screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { Route, Routes } from "react-router-dom";
import { beforeEach, describe, expect, it } from "vitest";

import { AdminReports } from "./AdminReports";
import { config } from "../../lib/config";
import { useAuthStore } from "../../store/auth-store";
import { server } from "../../test/msw-server";
import { renderWithProviders } from "../../test/test-utils";

function renderReportsPage() {
  return renderWithProviders(
    <Routes>
      <Route path="/admin/reports" element={<AdminReports />} />
      <Route path="/admin" element={<div>Dashboard</div>} />
    </Routes>,
    { route: "/admin/reports" },
  );
}

const sampleReport = {
  revenue: { total: 150000, previous_total: 100000, change_percent: 50 },
  orders: { total: 20, previous_total: 10, change_percent: 100 },
  pending_orders: 3,
  low_stock: { total: 5, critical: 1 },
  revenue_trend: [
    { date: "2026-09-17", amount: 0 },
    { date: "2026-09-18", amount: 150000 },
  ],
  order_funnel: { pending: 3, paid: 4, processing: 2, shipped: 1, delivered: 6 },
};

describe("AdminReports page", () => {
  beforeEach(() => {
    useAuthStore.setState({ permissions: ["report:read"] });
  });

  it("renders KPI tiles and the order funnel", async () => {
    server.use(
      http.get(`${config.apiBaseUrl}/admin/reports/dashboard`, () =>
        HttpResponse.json({ success: true, data: sampleReport }),
      ),
    );

    renderReportsPage();

    expect(await screen.findByText(/rp\s?150\.000/i)).toBeInTheDocument();
    expect(screen.getByText("+50.0%")).toBeInTheDocument();
    expect(screen.getByText("20")).toBeInTheDocument();
    expect(screen.getByText("+100.0%")).toBeInTheDocument();
    expect(screen.getByText("Perlu perhatian")).toBeInTheDocument();
    expect(screen.getByText("1 item kritis")).toBeInTheDocument();
    expect(screen.getByText("Pesanan Pending")).toBeInTheDocument();
    expect(screen.getByText("Menunggu Pembayaran")).toBeInTheDocument();
    expect(screen.getByText("Diterima")).toBeInTheDocument();
  });

  it("re-fetches with the selected period", async () => {
    let capturedDays: string | null = null;
    server.use(
      http.get(`${config.apiBaseUrl}/admin/reports/dashboard`, ({ request }) => {
        capturedDays = new URL(request.url).searchParams.get("days");
        return HttpResponse.json({ success: true, data: sampleReport });
      }),
    );

    renderReportsPage();
    await screen.findByText(/rp\s?150\.000/i);
    expect(capturedDays).toBe("30");
  });

  it("redirects away when the viewer lacks report:read", async () => {
    useAuthStore.setState({ permissions: [] });

    renderReportsPage();

    expect(await screen.findByText("Dashboard")).toBeInTheDocument();
  });

  it("shows an error state when the report fails to load", async () => {
    server.use(
      http.get(`${config.apiBaseUrl}/admin/reports/dashboard`, () =>
        HttpResponse.json(
          { success: false, error: { code: "FORBIDDEN", message: "Tidak punya izin", details: null } },
          { status: 403 },
        ),
      ),
    );

    renderReportsPage();

    expect(await screen.findByRole("alert")).toHaveTextContent(/gagal memuat laporan/i);
  });
});
