import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { Route, Routes } from "react-router-dom";
import { describe, expect, it } from "vitest";

import { ResetPassword } from "./ResetPassword";
import { config } from "../../lib/config";
import { server } from "../../test/msw-server";
import { renderWithProviders } from "../../test/test-utils";

const resetPasswordUrl = `${config.apiBaseUrl}/auth/reset-password`;

function renderResetPage(search = "?token=validtoken") {
  return renderWithProviders(
    <Routes>
      <Route path="/reset-password" element={<ResetPassword />} />
    </Routes>,
    { route: `/reset-password${search}` },
  );
}

describe("ResetPassword page", () => {
  it("shows an invalid-link message when there is no token in the URL", () => {
    renderResetPage("");

    expect(screen.getByRole("alert")).toHaveTextContent(/tautan reset password tidak valid/i);
  });

  it("submits the token and new password, then shows success", async () => {
    let receivedBody: unknown;
    server.use(
      http.post(resetPasswordUrl, async ({ request }) => {
        receivedBody = await request.json();
        return HttpResponse.json({ success: true, data: { message: "Password berhasil direset" } });
      }),
    );
    const user = userEvent.setup();
    renderResetPage("?token=validtoken");

    await user.type(screen.getByLabelText(/password baru/i), "NewPassw0rd");
    await user.click(screen.getByRole("button", { name: /reset password/i }));

    expect(await screen.findByText(/password berhasil direset/i)).toBeInTheDocument();
    expect(receivedBody).toEqual({ token: "validtoken", new_password: "NewPassw0rd" });
  });

  it("shows the backend's error for an expired or invalid token", async () => {
    server.use(
      http.post(resetPasswordUrl, () =>
        HttpResponse.json(
          {
            success: false,
            error: {
              code: "UNAUTHORIZED",
              message: "Token reset password tidak valid atau sudah kedaluwarsa",
              details: null,
            },
          },
          { status: 401 },
        ),
      ),
    );
    const user = userEvent.setup();
    renderResetPage("?token=expiredtoken");

    await user.type(screen.getByLabelText(/password baru/i), "NewPassw0rd");
    await user.click(screen.getByRole("button", { name: /reset password/i }));

    expect(await screen.findByRole("alert")).toHaveTextContent(/sudah kedaluwarsa/i);
  });

  it("shows a client-side validation error for a weak password without calling the API", async () => {
    const user = userEvent.setup();
    renderResetPage("?token=validtoken");

    await user.type(screen.getByLabelText(/password baru/i), "weak");
    await user.click(screen.getByRole("button", { name: /reset password/i }));

    expect(await screen.findByText(/password minimal 8 karakter/i)).toBeInTheDocument();
  });
});
