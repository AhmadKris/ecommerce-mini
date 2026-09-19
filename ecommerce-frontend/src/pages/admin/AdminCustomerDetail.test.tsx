import { screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { Route, Routes } from "react-router-dom";
import { beforeEach, describe, expect, it } from "vitest";

import { AdminCustomerDetail } from "./AdminCustomerDetail";
import { config } from "../../lib/config";
import { useAuthStore } from "../../store/auth-store";
import { server } from "../../test/msw-server";
import { renderWithProviders } from "../../test/test-utils";

function renderDetailPage(id: string) {
  return renderWithProviders(
    <Routes>
      <Route path="/admin/customers/:id" element={<AdminCustomerDetail />} />
    </Routes>,
    { route: `/admin/customers/${id}` },
  );
}

const sampleCustomer = {
  id: 1,
  name: "Budi Santoso",
  email: "budi@example.com",
  created_at: "2026-01-01T00:00:00Z",
  roles: [{ id: 1, name: "admin", permissions: [{ id: 1, code: "customer:read" }, { id: 2, code: "order:read_all" }] }],
};

describe("AdminCustomerDetail page", () => {
  beforeEach(() => {
    useAuthStore.setState({ permissions: ["customer:read"] });
  });

  it("renders customer profile, roles, and permissions", async () => {
    server.use(http.get(`${config.apiBaseUrl}/admin/customers/1`, () => HttpResponse.json({ success: true, data: sampleCustomer })));

    renderDetailPage("1");

    expect(await screen.findByRole("heading", { name: "Budi Santoso" })).toBeInTheDocument();
    expect(screen.getByText("budi@example.com")).toBeInTheDocument();
    expect(screen.getByText("admin")).toBeInTheDocument();
    expect(screen.getByText("customer:read")).toBeInTheDocument();
    expect(screen.getByText("order:read_all")).toBeInTheDocument();
  });

  it("shows a not-found state for a 404 from the backend", async () => {
    server.use(
      http.get(`${config.apiBaseUrl}/admin/customers/999`, () =>
        HttpResponse.json(
          { success: false, error: { code: "NOT_FOUND", message: "Customer tidak ditemukan", details: null } },
          { status: 404 },
        ),
      ),
    );

    renderDetailPage("999");

    expect(await screen.findByRole("alert")).toHaveTextContent(/pelanggan tidak ditemukan/i);
  });
});
