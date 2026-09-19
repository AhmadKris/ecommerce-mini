import { screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { Route, Routes } from "react-router-dom";
import { beforeEach, describe, expect, it } from "vitest";

import { AdminCustomers } from "./AdminCustomers";
import { config } from "../../lib/config";
import { useAuthStore } from "../../store/auth-store";
import { server } from "../../test/msw-server";
import { renderWithProviders } from "../../test/test-utils";

function renderListPage(route = "/admin/customers") {
  return renderWithProviders(
    <Routes>
      <Route path="/admin/customers" element={<AdminCustomers />} />
    </Routes>,
    { route },
  );
}

const sampleCustomer = {
  id: 1,
  name: "Budi Santoso",
  email: "budi@example.com",
  created_at: "2026-01-01T00:00:00Z",
  roles: [{ id: 1, name: "customer" }],
};

describe("AdminCustomers page", () => {
  beforeEach(() => {
    useAuthStore.setState({ permissions: ["customer:read"] });
  });

  it("renders the customer table with pagination controls", async () => {
    server.use(
      http.get(`${config.apiBaseUrl}/admin/customers`, () =>
        HttpResponse.json({
          success: true,
          data: { items: [sampleCustomer], meta: { page: 1, limit: 10, total: 1, total_pages: 1 } },
        }),
      ),
    );

    renderListPage();

    expect(await screen.findByRole("link", { name: "Budi Santoso" })).toBeInTheDocument();
    expect(screen.getByText("budi@example.com")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /sebelumnya/i })).toBeDisabled();
    expect(screen.getByRole("button", { name: /berikutnya/i })).toBeDisabled();
  });

  it("shows an empty state when there are no customers", async () => {
    server.use(
      http.get(`${config.apiBaseUrl}/admin/customers`, () =>
        HttpResponse.json({
          success: true,
          data: { items: [], meta: { page: 1, limit: 10, total: 0, total_pages: 0 } },
        }),
      ),
    );

    renderListPage();

    expect(await screen.findByText(/belum ada pelanggan/i)).toBeInTheDocument();
  });

  it("redirects away when the viewer lacks customer:read", async () => {
    useAuthStore.setState({ permissions: [] });

    renderWithProviders(
      <Routes>
        <Route path="/admin/customers" element={<AdminCustomers />} />
        <Route path="/admin" element={<div>Dashboard</div>} />
      </Routes>,
      { route: "/admin/customers" },
    );

    expect(await screen.findByText("Dashboard")).toBeInTheDocument();
  });
});
