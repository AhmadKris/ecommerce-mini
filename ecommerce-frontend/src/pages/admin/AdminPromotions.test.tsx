import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";

import { AdminPromotions } from "./AdminPromotions";
import { config } from "../../lib/config";
import { server } from "../../test/msw-server";
import { renderWithProviders } from "../../test/test-utils";

const promotionsUrl = `${config.apiBaseUrl}/admin/promotions`;

const samplePromotion = {
  id: 1,
  code: "HEMAT10",
  type: "percentage",
  value: 10,
  minimum_purchase: 10000,
  usage_limit: 5,
  used_count: 1,
  starts_at: "2026-01-01T00:00:00Z",
  ends_at: "2026-12-31T00:00:00Z",
  status: "active",
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

describe("AdminPromotions page", () => {
  it("lists existing promotions and creates a new one", async () => {
    server.use(
      http.get(promotionsUrl, () => HttpResponse.json({ success: true, data: { items: [samplePromotion] } })),
      http.post(promotionsUrl, () =>
        HttpResponse.json({ success: true, data: { ...samplePromotion, id: 2, code: "SAVE20" } }, { status: 201 }),
      ),
    );

    renderWithProviders(<AdminPromotions />);

    expect(await screen.findByText("HEMAT10")).toBeInTheDocument();
    expect(screen.getByText("1 / 5")).toBeInTheDocument();

    const user = userEvent.setup();
    await user.type(screen.getByLabelText(/^kode$/i), "SAVE20");
    await user.type(screen.getByLabelText(/nilai/i), "20");
    await user.type(screen.getByLabelText(/minimum pembelian/i), "0");
    await user.type(screen.getByLabelText(/batas pemakaian/i), "10");
    await user.type(screen.getByLabelText(/^mulai$/i), "2026-01-01T00:00");
    await user.type(screen.getByLabelText(/^berakhir$/i), "2026-12-31T00:00");
    await user.click(screen.getByRole("button", { name: /buat promo/i }));

    await screen.findByRole("button", { name: /buat promo/i });
  });

  it("shows an error state when the promotion list fails to load", async () => {
    server.use(
      http.get(promotionsUrl, () =>
        HttpResponse.json(
          { success: false, error: { code: "INTERNAL_ERROR", message: "err", details: null } },
          { status: 500 },
        ),
      ),
    );

    renderWithProviders(<AdminPromotions />);

    expect(await screen.findByRole("alert")).toHaveTextContent(/gagal memuat daftar promo/i);
  });
});
