import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { Route, Routes } from "react-router-dom";
import { beforeEach, describe, expect, it } from "vitest";

import { AdminCustomerDetail } from "./AdminCustomerDetail";
import { config } from "../../lib/config";
import { useAuthStore } from "../../store/auth-store";
import { server } from "../../test/msw-server";
import { renderWithProviders } from "../../test/test-utils";

const availableRoles = [
  { id: 1, name: "admin", permissions: [{ id: 1, code: "user:manage" }] },
  { id: 2, name: "customer", permissions: [] },
];

const sampleCustomer = {
  id: 1,
  name: "Budi Santoso",
  email: "budi@example.com",
  created_at: "2026-01-01T00:00:00Z",
  roles: [{ id: 1, name: "admin", permissions: [{ id: 1, code: "customer:read" }, { id: 2, code: "order:read_all" }] }],
};

function renderDetailPage(id: string) {
  return renderWithProviders(
    <Routes>
      <Route path="/admin/customers/:id" element={<AdminCustomerDetail />} />
    </Routes>,
    { route: `/admin/customers/${id}` },
  );
}

describe("AdminCustomerDetail page", () => {
  beforeEach(() => {
    useAuthStore.setState({ permissions: ["customer:read"], userId: 99 });
    server.use(
      http.get(`${config.apiBaseUrl}/admin/customers/1`, () =>
        HttpResponse.json({ success: true, data: sampleCustomer }),
      ),
    );
  });

  it("renders customer profile, roles, and permissions", async () => {
    renderDetailPage("1");

    expect(await screen.findByRole("heading", { name: "Budi Santoso" })).toBeInTheDocument();
    expect(screen.getByText("budi@example.com")).toBeInTheDocument();
    expect(screen.getByText("admin")).toBeInTheDocument();
    expect(screen.getByText("customer:read")).toBeInTheDocument();
    expect(screen.getByText("order:read_all")).toBeInTheDocument();
  });

  it("does not show role management section when caller lacks user:manage", async () => {
    // customer:read only — no user:manage
    useAuthStore.setState({ permissions: ["customer:read"], userId: 99 });

    renderDetailPage("1");

    await screen.findByRole("heading", { name: "Budi Santoso" });
    expect(screen.queryByText("Manajemen Role")).not.toBeInTheDocument();
  });

  it("shows role management checkboxes pre-checked from customer's current roles", async () => {
    useAuthStore.setState({ permissions: ["customer:read", "user:manage"], userId: 99 });
    server.use(
      http.get(`${config.apiBaseUrl}/admin/roles`, () =>
        HttpResponse.json({ success: true, data: availableRoles }),
      ),
    );

    renderDetailPage("1");

    await screen.findByRole("heading", { name: "Budi Santoso" });
    await screen.findByText("Manajemen Role");

    // Customer currently has "admin" role — its checkbox should be checked.
    const adminCheckbox = screen.getByRole("checkbox", { name: /admin/i });
    expect(adminCheckbox).toBeChecked();

    // "customer" is not in current roles — unchecked.
    const customerCheckbox = screen.getByRole("checkbox", { name: /customer/i });
    expect(customerCheckbox).not.toBeChecked();
  });

  it("submits updated role names on save and shows success message", async () => {
    const user = userEvent.setup();
    useAuthStore.setState({ permissions: ["customer:read", "user:manage"], userId: 99 });

    let capturedBody: unknown;
    server.use(
      http.get(`${config.apiBaseUrl}/admin/roles`, () =>
        HttpResponse.json({ success: true, data: availableRoles }),
      ),
      http.put(`${config.apiBaseUrl}/admin/customers/1/roles`, async ({ request }) => {
        capturedBody = await request.json();
        return HttpResponse.json({ success: true, data: { ...sampleCustomer, roles: [availableRoles[1]] } });
      }),
    );

    renderDetailPage("1");

    await screen.findByText("Manajemen Role");
    const adminCheckbox = screen.getByRole("checkbox", { name: /admin/i });
    await user.click(adminCheckbox); // uncheck admin

    const customerCheckbox = screen.getByRole("checkbox", { name: /customer/i });
    await user.click(customerCheckbox); // check customer

    await user.click(screen.getByRole("button", { name: /simpan role/i }));

    await waitFor(() => expect(screen.getByRole("status")).toHaveTextContent(/berhasil/i));
    expect(capturedBody).toEqual({ role_names: ["customer"] });
  });

  it("shows conflict error when admin tries to remove own admin role", async () => {
    // userId == customerId (1) — self-update scenario
    useAuthStore.setState({ permissions: ["customer:read", "user:manage"], userId: 1 });

    server.use(
      http.get(`${config.apiBaseUrl}/admin/roles`, () =>
        HttpResponse.json({ success: true, data: availableRoles }),
      ),
      http.put(`${config.apiBaseUrl}/admin/customers/1/roles`, () =>
        HttpResponse.json(
          { success: false, error: { code: "CONFLICT", message: "Admin tidak dapat menghapus role admin dari diri sendiri", details: null } },
          { status: 409 },
        ),
      ),
    );

    renderDetailPage("1");

    await screen.findByText("Manajemen Role");
    // Self-lockout warning should be visible
    expect(screen.getByText(/tidak dapat menghapus role/i)).toBeInTheDocument();

    // Admin checkbox should be disabled for self
    const adminCheckbox = screen.getByRole("checkbox", { name: /admin/i });
    expect(adminCheckbox).toBeDisabled();
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
