import { screen } from "@testing-library/react";
import { Route, Routes } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import { AdminRoute } from "./AdminRoute";
import { useAuthStore } from "../store/auth-store";
import { fakeAccessToken } from "../test/fake-jwt";
import { renderWithProviders } from "../test/test-utils";

afterEach(() => useAuthStore.getState().clearSession());

function renderGuardedRoute() {
  return renderWithProviders(
    <Routes>
      <Route path="/" element={<p>Halaman utama</p>} />
      <Route path="/login" element={<p>Halaman login</p>} />
      <Route element={<AdminRoute />}>
        <Route path="/admin" element={<p>Halaman admin</p>} />
      </Route>
    </Routes>,
    { route: "/admin" },
  );
}

describe("AdminRoute", () => {
  it("redirects to /login when there is no session", () => {
    renderGuardedRoute();
    expect(screen.getByText("Halaman login")).toBeInTheDocument();
  });

  it("redirects to / when logged in without the admin role", () => {
    useAuthStore.getState().setSession({
      access_token: fakeAccessToken({ roles: ["customer"] }),
      refresh_token: "r",
      expires_in: 900,
    });

    renderGuardedRoute();

    expect(screen.getByText("Halaman utama")).toBeInTheDocument();
  });

  it("renders the guarded route for the admin role", () => {
    useAuthStore.getState().setSession({
      access_token: fakeAccessToken({ roles: ["admin"] }),
      refresh_token: "r",
      expires_in: 900,
    });

    renderGuardedRoute();

    expect(screen.getByText("Halaman admin")).toBeInTheDocument();
  });
});
