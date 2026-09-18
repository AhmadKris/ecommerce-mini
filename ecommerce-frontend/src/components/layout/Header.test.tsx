import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { afterEach, describe, expect, it } from "vitest";

import { Header } from "./Header";
import { config } from "../../lib/config";
import { useAuthStore } from "../../store/auth-store";
import { fakeAccessToken } from "../../test/fake-jwt";
import { server } from "../../test/msw-server";
import { renderWithProviders } from "../../test/test-utils";

afterEach(() => useAuthStore.getState().clearSession());

describe("Header", () => {
  it("shows Sign in when logged out", () => {
    renderWithProviders(<Header />);
    expect(screen.getByRole("link", { name: /masuk/i })).toHaveAttribute("href", "/login");
    expect(screen.queryByText(/keranjang/i)).not.toBeInTheDocument();
    expect(screen.queryByRole("link", { name: /pesanan/i })).not.toBeInTheDocument();
  });

  it("shows the cart link with item count when logged in", async () => {
    useAuthStore.getState().setSession({ access_token: fakeAccessToken(), refresh_token: "r", expires_in: 900 });
    server.use(
      http.get(`${config.apiBaseUrl}/cart`, () =>
        HttpResponse.json({
          success: true,
          data: { items: [{ id: 1, product: { id: 1 }, quantity: 1, subtotal: 1000 }], total: 1000 },
        }),
      ),
    );

    renderWithProviders(<Header />);

    // The cart link's aria-label ("Lihat keranjang, 1 item") is its
    // accessible name — it overrides the visible "Keranjang (1)" text per
    // the ARIA accessible-name computation, so that's what findByRole
    // matches against, not the visible label.
    const cartLink = await screen.findByRole("link", { name: /lihat keranjang, 1 item/i });
    expect(cartLink).toHaveAttribute("href", "/cart");
    expect(cartLink).toHaveTextContent("Keranjang (1)");
    expect(screen.getByRole("button", { name: /keluar/i })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Pesanan" })).toHaveAttribute("href", "/orders");
  });

  it("clears the session when Keluar is clicked", async () => {
    useAuthStore.getState().setSession({ access_token: fakeAccessToken(), refresh_token: "r", expires_in: 900 });
    server.use(
      http.get(`${config.apiBaseUrl}/cart`, () =>
        HttpResponse.json({ success: true, data: { items: [], total: 0 } }),
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<Header />);

    await user.click(await screen.findByRole("button", { name: /keluar/i }));

    expect(useAuthStore.getState().accessToken).toBeNull();
  });
});
