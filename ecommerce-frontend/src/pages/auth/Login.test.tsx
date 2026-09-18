import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { Route, Routes } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import { Login } from "./Login";
import { config } from "../../lib/config";
import { useAuthStore } from "../../store/auth-store";
import { fakeAccessToken } from "../../test/fake-jwt";
import { server } from "../../test/msw-server";
import { renderWithProviders } from "../../test/test-utils";

const loginUrl = `${config.apiBaseUrl}/auth/login`;

afterEach(() => useAuthStore.getState().clearSession());

function renderLoginPage() {
  return renderWithProviders(
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/" element={<p>Beranda</p>} />
    </Routes>,
    { route: "/login" },
  );
}

describe("Login page", () => {
  it("logs in and navigates home on success", async () => {
    server.use(
      http.post(loginUrl, () =>
        HttpResponse.json({
          success: true,
          data: { access_token: fakeAccessToken(), refresh_token: "r", expires_in: 900 },
        }),
      ),
    );
    const user = userEvent.setup();
    renderLoginPage();

    await user.type(screen.getByLabelText(/email/i), "achmad@example.com");
    await user.type(screen.getByLabelText(/password/i), "Passw0rd");
    await user.click(screen.getByRole("button", { name: /masuk/i }));

    expect(await screen.findByText("Beranda")).toBeInTheDocument();
    expect(useAuthStore.getState().accessToken).not.toBeNull();
  });

  it("shows the server error message on invalid credentials", async () => {
    server.use(
      http.post(loginUrl, () =>
        HttpResponse.json(
          {
            success: false,
            error: { code: "UNAUTHORIZED", message: "Email atau password salah", details: null },
          },
          { status: 401 },
        ),
      ),
    );
    const user = userEvent.setup();
    renderLoginPage();

    await user.type(screen.getByLabelText(/email/i), "achmad@example.com");
    await user.type(screen.getByLabelText(/password/i), "wrongpass");
    await user.click(screen.getByRole("button", { name: /masuk/i }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Email atau password salah");
  });

  it("shows a client-side validation error without calling the API", async () => {
    const user = userEvent.setup();
    renderLoginPage();

    await user.type(screen.getByLabelText(/email/i), "not-an-email");
    await user.type(screen.getByLabelText(/password/i), "x");
    await user.click(screen.getByRole("button", { name: /masuk/i }));

    expect(await screen.findByText(/email tidak valid/i)).toBeInTheDocument();
  });
});
