import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";

import { ForgotPassword } from "./ForgotPassword";
import { config } from "../../lib/config";
import { server } from "../../test/msw-server";
import { renderWithProviders } from "../../test/test-utils";

const forgotPasswordUrl = `${config.apiBaseUrl}/auth/forgot-password`;

describe("ForgotPassword page", () => {
  it("shows the backend's generic message after submitting, regardless of whether the email exists", async () => {
    server.use(
      http.post(forgotPasswordUrl, () =>
        HttpResponse.json({
          success: true,
          data: { message: "Jika email terdaftar, instruksi reset password telah dikirim" },
        }),
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<ForgotPassword />);

    await user.type(screen.getByLabelText(/email/i), "siapa-saja@example.com");
    await user.click(screen.getByRole("button", { name: /kirim instruksi reset/i }));

    expect(await screen.findByText(/instruksi reset password telah dikirim/i)).toBeInTheDocument();
  });

  it("shows a client-side validation error for an invalid email without calling the API", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ForgotPassword />);

    await user.type(screen.getByLabelText(/email/i), "not-an-email");
    await user.click(screen.getByRole("button", { name: /kirim instruksi reset/i }));

    expect(await screen.findByText(/email tidak valid/i)).toBeInTheDocument();
  });
});
