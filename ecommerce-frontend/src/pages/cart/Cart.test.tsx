import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { afterEach, describe, expect, it } from "vitest";

import { Cart } from "./Cart";
import { config } from "../../lib/config";
import { useAuthStore } from "../../store/auth-store";
import { fakeAccessToken } from "../../test/fake-jwt";
import { server } from "../../test/msw-server";
import { renderWithProviders } from "../../test/test-utils";

const cartUrl = `${config.apiBaseUrl}/cart`;

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
  quantity: 2,
  subtotal: 36000,
};

afterEach(() => useAuthStore.getState().clearSession());

function logIn() {
  useAuthStore.getState().setSession({ access_token: fakeAccessToken(), refresh_token: "r", expires_in: 900 });
}

describe("Cart page", () => {
  it("shows an empty state when the cart has no items", async () => {
    logIn();
    server.use(
      http.get(cartUrl, () => HttpResponse.json({ success: true, data: { items: [], total: 0 } })),
    );

    renderWithProviders(<Cart />);

    expect(await screen.findByText(/keranjang anda kosong/i)).toBeInTheDocument();
  });

  it("renders items with subtotal and the order summary total (items + flat shipping)", async () => {
    logIn();
    server.use(
      http.get(cartUrl, () =>
        HttpResponse.json({ success: true, data: { items: [sampleItem], total: 36000 } }),
      ),
    );

    renderWithProviders(<Cart />);

    expect(await screen.findByText("Kopi Susu Gula Aren")).toBeInTheDocument();
    // Subtotal (36.000) + flat shipping (25.000) = 61.000
    expect(screen.getByText(/61\.000/)).toBeInTheDocument();
  });

  it("removes an item when Hapus is clicked", async () => {
    logIn();
    let removeCalled = false;
    server.use(
      http.get(cartUrl, () =>
        HttpResponse.json({ success: true, data: { items: [sampleItem], total: 36000 } }),
      ),
      http.delete(`${cartUrl}/items/1`, () => {
        removeCalled = true;
        return HttpResponse.json({ success: true, data: { items: [], total: 0 } });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<Cart />);

    await user.click(await screen.findByRole("button", { name: /hapus/i }));

    expect(removeCalled).toBe(true);
    expect(await screen.findByText(/keranjang anda kosong/i)).toBeInTheDocument();
  });

  it("disables the decrease button at quantity 1", async () => {
    logIn();
    server.use(
      http.get(cartUrl, () =>
        HttpResponse.json({
          success: true,
          data: { items: [{ ...sampleItem, quantity: 1, subtotal: 18000 }], total: 18000 },
        }),
      ),
    );

    renderWithProviders(<Cart />);

    expect(await screen.findByRole("button", { name: /kurangi jumlah/i })).toBeDisabled();
  });

  it("disables the increase button once quantity reaches stock", async () => {
    logIn();
    server.use(
      http.get(cartUrl, () =>
        HttpResponse.json({
          success: true,
          data: { items: [{ ...sampleItem, quantity: 10, subtotal: 180000 }], total: 180000 },
        }),
      ),
    );

    renderWithProviders(<Cart />);

    expect(await screen.findByRole("button", { name: /tambah jumlah/i })).toBeDisabled();
  });
});
