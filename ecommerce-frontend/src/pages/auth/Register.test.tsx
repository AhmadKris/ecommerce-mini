import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { Route, Routes } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import { Register } from "./Register";
import { config } from "../../lib/config";
import { useAuthStore } from "../../store/auth-store";
import { fakeAccessToken } from "../../test/fake-jwt";
import { server } from "../../test/msw-server";
import { renderWithProviders } from "../../test/test-utils";

const registerUrl = `${config.apiBaseUrl}/auth/register`;
const loginUrl = `${config.apiBaseUrl}/auth/login`;

afterEach(() => useAuthStore.getState().clearSession());

function renderRegisterPage() {
  return renderWithProviders(
    <Routes>
      <Route path="/register" element={<Register />} />
      <Route path="/" element={<p>Beranda</p>} />
    </Routes>,
    { route: "/register" },
  );
}

async function fillValidForm(user: ReturnType<typeof userEvent.setup>) {
  await user.type(screen.getByLabelText(/nama/i), "Achmad Crizdianto");
  await user.type(screen.getByLabelText(/email/i), "achmad@example.com");
  await user.type(screen.getByLabelText(/password/i), "Passw0rd");
}

describe("Register page", () => {
  it("registers, auto-logs in, and navigates home", async () => {
    server.use(
      http.post(registerUrl, () =>
        HttpResponse.json(
          {
            success: true,
            data: { id: 1, name: "Achmad Crizdianto", email: "achmad@example.com" },
          },
          { status: 201 },
        ),
      ),
      http.post(loginUrl, () =>
        HttpResponse.json({
          success: true,
          data: { access_token: fakeAccessToken(), refresh_token: "r", expires_in: 900 },
        }),
      ),
    );
    const user = userEvent.setup();
    renderRegisterPage();

    await fillValidForm(user);
    await user.click(screen.getByRole("button", { name: /daftar/i }));

    expect(await screen.findByText("Beranda")).toBeInTheDocument();
    expect(useAuthStore.getState().accessToken).not.toBeNull();
  });

  it("shows the server error when the email is already taken", async () => {
    server.use(
      http.post(registerUrl, () =>
        HttpResponse.json(
          {
            success: false,
            error: { code: "DUPLICATE_ENTRY", message: "Email sudah terdaftar", details: null },
          },
          { status: 409 },
        ),
      ),
    );
    const user = userEvent.setup();
    renderRegisterPage();

    await fillValidForm(user);
    await user.click(screen.getByRole("button", { name: /daftar/i }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Email sudah terdaftar");
  });

  it("shows a client-side validation error for a weak password without calling the API", async () => {
    const user = userEvent.setup();
    renderRegisterPage();

    await user.type(screen.getByLabelText(/nama/i), "Achmad Crizdianto");
    await user.type(screen.getByLabelText(/email/i), "achmad@example.com");
    await user.type(screen.getByLabelText(/password/i), "password");
    await user.click(screen.getByRole("button", { name: /daftar/i }));

    expect(await screen.findByText(/huruf besar/i)).toBeInTheDocument();
  });
});
