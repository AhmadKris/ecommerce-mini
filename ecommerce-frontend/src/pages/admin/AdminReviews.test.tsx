import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";

import { AdminReviews } from "./AdminReviews";
import { config } from "../../lib/config";
import { server } from "../../test/msw-server";
import { renderWithProviders } from "../../test/test-utils";

const adminReviewsUrl = `${config.apiBaseUrl}/admin/reviews`;

const sampleReview = {
  id: 1,
  product_id: 3,
  user_id: 10,
  order_id: 9,
  rating: 5,
  title: "Bagus banget",
  body: "Kualitas produk sangat memuaskan.",
  status: "pending",
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
  product: { id: 3, name: "Produk Test", slug: "produk-test" },
};

describe("AdminReviews page", () => {
  it("lists pending reviews and approves one", async () => {
    server.use(
      http.get(adminReviewsUrl, () =>
        HttpResponse.json({
          success: true,
          data: { items: [sampleReview], meta: { page: 1, limit: 10, total: 1, total_pages: 1 } },
        }),
      ),
      http.patch(`${adminReviewsUrl}/1/status`, () =>
        HttpResponse.json({ success: true, data: { ...sampleReview, status: "approved" } }),
      ),
    );
    const user = userEvent.setup();

    renderWithProviders(<AdminReviews />);

    expect(await screen.findByText("Bagus banget")).toBeInTheDocument();
    expect(screen.getByText("Produk Test")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /setujui/i }));

    await screen.findByRole("button", { name: /setujui/i }); // settles back to idle
  });

  it("shows an error state when the review list fails to load", async () => {
    server.use(
      http.get(adminReviewsUrl, () =>
        HttpResponse.json(
          { success: false, error: { code: "INTERNAL_ERROR", message: "err", details: null } },
          { status: 500 },
        ),
      ),
    );

    renderWithProviders(<AdminReviews />);

    expect(await screen.findByRole("alert")).toHaveTextContent(/gagal memuat daftar ulasan/i);
  });
});
