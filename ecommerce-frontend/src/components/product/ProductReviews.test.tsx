import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { afterEach, describe, expect, it } from "vitest";

import { ProductReviews } from "./ProductReviews";
import { config } from "../../lib/config";
import { useAuthStore } from "../../store/auth-store";
import { fakeAccessToken } from "../../test/fake-jwt";
import { server } from "../../test/msw-server";
import { renderWithProviders } from "../../test/test-utils";

const reviewsUrl = `${config.apiBaseUrl}/products/kopi-hitam/reviews`;

afterEach(() => useAuthStore.getState().clearSession());

function logIn() {
  useAuthStore.getState().setSession({ access_token: fakeAccessToken(), refresh_token: "r", expires_in: 900 });
}

describe("ProductReviews", () => {
  it("renders approved reviews and prompts a guest to log in instead of showing the form", async () => {
    server.use(
      http.get(reviewsUrl, () =>
        HttpResponse.json({
          success: true,
          data: {
            items: [
              {
                id: 1,
                product_id: 1,
                user_id: 5,
                order_id: 9,
                rating: 4,
                title: "Mantap",
                body: "Kopinya enak dan pengiriman cepat.",
                status: "approved",
                created_at: "2026-01-01T00:00:00Z",
                updated_at: "2026-01-01T00:00:00Z",
              },
            ],
            meta: { page: 1, limit: 10, total: 1, total_pages: 1 },
          },
        }),
      ),
    );

    renderWithProviders(<ProductReviews slug="kopi-hitam" />);

    expect(await screen.findByText("Mantap")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /masuk/i })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /kirim ulasan/i })).not.toBeInTheDocument();
  });

  it("lets a logged-in user submit a review and shows the pending-moderation message", async () => {
    logIn();
    server.use(
      http.get(reviewsUrl, () =>
        HttpResponse.json({
          success: true,
          data: { items: [], meta: { page: 1, limit: 10, total: 0, total_pages: 0 } },
        }),
      ),
      http.post(reviewsUrl, () =>
        HttpResponse.json(
          { success: true, data: { id: 1, product_id: 1, user_id: 5, order_id: 9, rating: 5, status: "pending" } },
          { status: 201 },
        ),
      ),
    );
    const user = userEvent.setup();

    renderWithProviders(<ProductReviews slug="kopi-hitam" />);

    await user.type(await screen.findByLabelText(/^judul$/i), "Bagus banget");
    await user.type(screen.getByLabelText(/^ulasan$/i), "Kualitas produk sangat memuaskan sekali.");
    await user.click(screen.getByRole("button", { name: /kirim ulasan/i }));

    expect(await screen.findByText(/menunggu persetujuan admin/i)).toBeInTheDocument();
  });

  it("surfaces the backend's forbidden error when the caller never received the product", async () => {
    logIn();
    server.use(
      http.get(reviewsUrl, () =>
        HttpResponse.json({
          success: true,
          data: { items: [], meta: { page: 1, limit: 10, total: 0, total_pages: 0 } },
        }),
      ),
      http.post(reviewsUrl, () =>
        HttpResponse.json(
          {
            success: false,
            error: {
              code: "FORBIDDEN",
              message: "Kamu hanya bisa memberi ulasan untuk produk yang sudah diterima",
              details: null,
            },
          },
          { status: 403 },
        ),
      ),
    );
    const user = userEvent.setup();

    renderWithProviders(<ProductReviews slug="kopi-hitam" />);

    await user.type(await screen.findByLabelText(/^judul$/i), "Bagus banget");
    await user.type(screen.getByLabelText(/^ulasan$/i), "Kualitas produk sangat memuaskan sekali.");
    await user.click(screen.getByRole("button", { name: /kirim ulasan/i }));

    expect(await screen.findByRole("alert")).toHaveTextContent(/sudah diterima/i);
  });
});
