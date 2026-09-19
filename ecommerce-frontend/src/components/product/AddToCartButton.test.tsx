import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { Route, Routes } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import { AddToCartButton } from "./AddToCartButton";
import { config } from "../../lib/config";
import { useAuthStore } from "../../store/auth-store";
import { fakeAccessToken } from "../../test/fake-jwt";
import { server } from "../../test/msw-server";
import { renderWithProviders } from "../../test/test-utils";
import type { Product } from "../../types/product";

const product: Product = {
  id: 1,
  name: "Kopi Susu Gula Aren",
  slug: "kopi-susu-gula-aren",
  sku: "KOPI-001",
  description: "",
  price: 18000,
  stock: 10,
  category_id: 1,
  image_url: "",
  created_at: "2026-01-01T00:00:00Z",
};

afterEach(() => useAuthStore.getState().clearSession());

function renderButton() {
  return renderWithProviders(
    <Routes>
      <Route path="/" element={<AddToCartButton product={product} />} />
      <Route path="/login" element={<p>Halaman login</p>} />
    </Routes>,
    { route: "/" },
  );
}

describe("AddToCartButton", () => {
  it("shows a disabled button when the product is out of stock", () => {
    renderWithProviders(<AddToCartButton product={{ ...product, stock: 0 }} />);
    expect(screen.getByRole("button", { name: /stok habis/i })).toBeDisabled();
  });

  it("redirects to /login instead of calling the API when logged out", async () => {
    const user = userEvent.setup();
    renderButton();

    await user.click(screen.getByRole("button", { name: /tambah ke keranjang/i }));

    expect(await screen.findByText("Halaman login")).toBeInTheDocument();
  });

  it("adds the item and shows a confirmation when logged in", async () => {
    useAuthStore.getState().setSession({ access_token: fakeAccessToken(), refresh_token: "r", expires_in: 900 });
    server.use(
      http.post(`${config.apiBaseUrl}/cart/items`, () =>
        HttpResponse.json({ success: true, data: { items: [], total: 0 } }),
      ),
    );
    const user = userEvent.setup();
    renderButton();

    await user.click(screen.getByRole("button", { name: /tambah ke keranjang/i }));

    expect(await screen.findByRole("button", { name: /ditambahkan/i })).toBeInTheDocument();
  });
});
