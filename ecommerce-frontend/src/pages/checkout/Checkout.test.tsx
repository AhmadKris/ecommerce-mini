import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { afterEach, describe, expect, it } from "vitest";

import { Checkout } from "./Checkout";
import { config } from "../../lib/config";
import { useAuthStore } from "../../store/auth-store";
import { fakeAccessToken } from "../../test/fake-jwt";
import { server } from "../../test/msw-server";
import { renderWithProviders } from "../../test/test-utils";

const cartUrl = `${config.apiBaseUrl}/cart`;
const ordersUrl = `${config.apiBaseUrl}/orders`;

const sampleItem = {
  id: 1,
  product: {
    id: 1,
    name: "Kopi Susu Gula Aren",
    slug: "kopi-susu-gula-aren",
    description: "",
    price: 18000,
    stock: 10,
    category_id: 1,
    image_url: "",
    created_at: "2026-01-01T00:00:00Z",
  },
  quantity: 1,
  subtotal: 18000,
};

afterEach(() => useAuthStore.getState().clearSession());

function logIn() {
  useAuthStore.getState().setSession({ access_token: fakeAccessToken(), refresh_token: "r", expires_in: 900 });
}

describe("Checkout page", () => {
  it("shows an empty-cart state instead of the form when the cart has no items", async () => {
    logIn();
    server.use(http.get(cartUrl, () => HttpResponse.json({ success: true, data: { items: [], total: 0 } })));

    renderWithProviders(<Checkout />);

    expect(await screen.findByText(/keranjang anda kosong/i)).toBeInTheDocument();
  });

  it("shows a client-side validation error for a too-short address", async () => {
    logIn();
    server.use(
      http.get(cartUrl, () =>
        HttpResponse.json({ success: true, data: { items: [sampleItem], total: 18000 } }),
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<Checkout />);

    await user.type(await screen.findByLabelText(/alamat pengiriman/i), "Jl. A");
    await user.click(screen.getByRole("button", { name: /buat pesanan/i }));

    expect(await screen.findByText(/alamat pengiriman minimal 10 karakter/i)).toBeInTheDocument();
  });

  it("submits and shows the order confirmation on success", async () => {
    logIn();
    server.use(
      http.get(cartUrl, () =>
        HttpResponse.json({ success: true, data: { items: [sampleItem], total: 18000 } }),
      ),
      http.post(ordersUrl, () =>
        HttpResponse.json(
          {
            success: true,
            data: {
              id: 42,
              user_id: 1,
              status: "pending",
              total_amount: 43000,
              shipping_cost: 25000,
              shipping_address: "Jl. Merdeka No. 1",
              created_at: "2026-01-01T00:00:00Z",
              items: [],
            },
          },
          { status: 201 },
        ),
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<Checkout />);

    await user.type(await screen.findByLabelText(/alamat pengiriman/i), "Jl. Merdeka No. 1");
    await user.click(screen.getByRole("button", { name: /buat pesanan/i }));

    expect(await screen.findByText(/pesanan berhasil dibuat/i)).toBeInTheDocument();
    expect(screen.getByText(/order #42/i)).toBeInTheDocument();
  });

  it("shows the server error when checkout fails (e.g. insufficient stock)", async () => {
    logIn();
    server.use(
      http.get(cartUrl, () =>
        HttpResponse.json({ success: true, data: { items: [sampleItem], total: 18000 } }),
      ),
      http.post(ordersUrl, () =>
        HttpResponse.json(
          {
            success: false,
            error: {
              code: "CONFLICT",
              message: "Stok tidak mencukupi untuk salah satu produk di keranjang",
              details: null,
            },
          },
          { status: 409 },
        ),
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<Checkout />);

    await user.type(await screen.findByLabelText(/alamat pengiriman/i), "Jl. Merdeka No. 1");
    await user.click(screen.getByRole("button", { name: /buat pesanan/i }));

    expect(await screen.findByRole("alert")).toHaveTextContent(/stok tidak mencukupi/i);
  });
});
