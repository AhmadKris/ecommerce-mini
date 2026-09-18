import { screen } from "@testing-library/react";
import { Route, Routes } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import { ProtectedRoute } from "./ProtectedRoute";
import { useAuthStore } from "../store/auth-store";
import { fakeAccessToken } from "../test/fake-jwt";
import { renderWithProviders } from "../test/test-utils";

afterEach(() => useAuthStore.getState().clearSession());

function renderGuardedRoute(route: string) {
  return renderWithProviders(
    <Routes>
      <Route path="/login" element={<p>Halaman login</p>} />
      <Route element={<ProtectedRoute />}>
        <Route path="/cart" element={<p>Halaman cart</p>} />
      </Route>
    </Routes>,
    { route },
  );
}

describe("ProtectedRoute", () => {
  it("redirects to /login when there is no session", () => {
    renderGuardedRoute("/cart");
    expect(screen.getByText("Halaman login")).toBeInTheDocument();
  });

  it("renders the guarded route when a session exists", () => {
    useAuthStore.getState().setSession({
      access_token: fakeAccessToken(),
      refresh_token: "r",
      expires_in: 900,
    });

    renderGuardedRoute("/cart");

    expect(screen.getByText("Halaman cart")).toBeInTheDocument();
  });
});
