import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Route, Routes } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import { AdminLayout } from "./AdminLayout";
import { useAuthStore } from "../../store/auth-store";
import { fakeAccessToken } from "../../test/fake-jwt";
import { renderWithProviders } from "../../test/test-utils";

afterEach(() => useAuthStore.getState().clearSession());

function renderAdminLayout(route: string) {
  useAuthStore.getState().setSession({
    access_token: fakeAccessToken({ roles: ["admin"] }),
    refresh_token: "r",
    expires_in: 900,
  });

  return renderWithProviders(
    <Routes>
      <Route element={<AdminLayout />}>
        <Route path="/admin" element={<p>Konten dashboard</p>} />
        <Route path="/admin/products" element={<p>Konten produk</p>} />
      </Route>
    </Routes>,
    { route },
  );
}

describe("AdminLayout", () => {
  it("renders the sidebar nav and the routed page content", () => {
    renderAdminLayout("/admin");

    expect(screen.getByRole("link", { name: "Dashboard" })).toHaveAttribute("href", "/admin");
    expect(screen.getByRole("link", { name: "Produk" })).toHaveAttribute("href", "/admin/products");
    expect(screen.getByText("Konten dashboard")).toBeInTheDocument();
  });

  it("shows a breadcrumb matching the current section", () => {
    renderAdminLayout("/admin/products");

    expect(screen.getByText(/admin \/ produk/i)).toBeInTheDocument();
    expect(screen.getByText("Konten produk")).toBeInTheDocument();
  });

  it("clears the session when Keluar is clicked", async () => {
    const user = userEvent.setup();
    renderAdminLayout("/admin");

    await user.click(screen.getByRole("button", { name: /keluar/i }));

    expect(useAuthStore.getState().accessToken).toBeNull();
  });
});
