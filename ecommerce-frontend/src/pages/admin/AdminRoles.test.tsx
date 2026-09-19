import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { Route, Routes } from "react-router-dom";
import { beforeEach, describe, expect, it } from "vitest";

import { config } from "../../lib/config";
import { queryClient } from "../../lib/query-client";
import { useAuthStore } from "../../store/auth-store";
import { server } from "../../test/msw-server";
import { renderWithProviders } from "../../test/test-utils";
import { AdminRoles } from "./AdminRoles";

const availableRoles = [
  {
    id: 1,
    name: "admin",
    permissions: [
      { id: 10, code: "customer:read" },
      { id: 11, code: "user:manage" },
    ],
  },
  {
    id: 2,
    name: "customer",
    permissions: [{ id: 12, code: "product:read" }],
  },
];

const customerList = {
  items: [
    {
      id: 1,
      name: "Budi Santoso",
      email: "budi@example.com",
      roles: [{ id: 1, name: "admin" }],
      created_at: "2026-01-01T00:00:00Z",
    },
    {
      id: 2,
      name: "Siti Rahma",
      email: "siti@example.com",
      roles: [{ id: 2, name: "customer" }],
      created_at: "2026-01-02T00:00:00Z",
    },
  ],
  meta: {
    page: 1,
    limit: 10,
    total_items: 2,
    total_pages: 1,
  },
};

function renderRolesPage(route = "/admin/roles") {
  return renderWithProviders(
    <Routes>
      <Route path="/admin/roles" element={<AdminRoles />} />
      <Route path="/admin" element={<div>Admin Home Page</div>} />
    </Routes>,
    { route },
  );
}

describe("AdminRoles page", () => {
  beforeEach(() => {
    queryClient.clear();
    useAuthStore.setState({
      accessToken: "mock.jwt.token",
      roles: ["admin"],
      permissions: ["user:manage"],
      userId: 1,
    });

    server.use(
      http.get(`${config.apiBaseUrl}/admin/roles`, () => HttpResponse.json({ success: true, data: availableRoles })),
      http.get(`${config.apiBaseUrl}/admin/customers`, () => HttpResponse.json({ success: true, data: customerList })),
    );
  });

  it("renders role list and user table when caller has user:manage permission", async () => {
    renderRolesPage();

    await screen.findByText("customer:read");
    expect(screen.getByText("Daftar Role & Permission Sistem")).toBeInTheDocument();
    expect(screen.getByText("user:manage")).toBeInTheDocument();

    await screen.findByText("Budi Santoso");
    expect(screen.getByText("Siti Rahma")).toBeInTheDocument();
  });

  it("opens edit modal, updates user roles, and displays success message", async () => {
    const user = userEvent.setup();

    let capturedBody: unknown;
    server.use(
      http.put(`${config.apiBaseUrl}/admin/customers/2/roles`, async ({ request }) => {
        capturedBody = await request.json();
        return HttpResponse.json({
          success: true,
          data: {
            id: 2,
            name: "Siti Rahma",
            email: "siti@example.com",
            roles: [
              { id: 1, name: "admin" },
              { id: 2, name: "customer" },
            ],
            created_at: "2026-01-02T00:00:00Z",
          },
        });
      }),
    );

    renderRolesPage();

    await screen.findByText("Siti Rahma");

    // Click "Ubah Role" for Siti Rahma (index 1 button)
    const editButtons = screen.getAllByRole("button", { name: /ubah role/i });
    await user.click(editButtons[1]);

    expect(screen.getByText("Kelola Role — Siti Rahma")).toBeInTheDocument();

    // Check admin checkbox for Siti Rahma
    const adminCheckbox = screen.getByRole("checkbox", { name: /admin/i });
    await user.click(adminCheckbox);

    // Click save
    await user.click(screen.getByRole("button", { name: /simpan role/i }));

    await waitFor(() => expect(screen.getByRole("status")).toHaveTextContent(/berhasil diperbarui/i));
    expect(capturedBody).toEqual({ role_names: ["customer", "admin"] });
  });

  it("disables admin role checkbox when editing own account (self-lockout protection)", async () => {
    const user = userEvent.setup();

    renderRolesPage();

    await screen.findByText("Budi Santoso");

    // Click "Ubah Role" for Budi Santoso (index 0 button)
    const editButtons = screen.getAllByRole("button", { name: /ubah role/i });
    await user.click(editButtons[0]);

    expect(screen.getByText(/tidak dapat menghapus role/i)).toBeInTheDocument();

    const adminCheckbox = screen.getByRole("checkbox", { name: /admin/i });
    expect(adminCheckbox).toBeDisabled();
  });

  it("redirects to /admin if caller lacks user:manage permission", async () => {
    useAuthStore.setState({ permissions: ["customer:read"] }); // lacks user:manage

    renderRolesPage();

    await screen.findByText("Admin Home Page");
  });
});
