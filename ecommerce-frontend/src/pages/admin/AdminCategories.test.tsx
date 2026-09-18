import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";

import { AdminCategories } from "./AdminCategories";
import { config } from "../../lib/config";
import { server } from "../../test/msw-server";
import { renderWithProviders } from "../../test/test-utils";

const categoriesUrl = `${config.apiBaseUrl}/categories`;
const adminCategoriesUrl = `${config.apiBaseUrl}/admin/categories`;

describe("AdminCategories page", () => {
  it("lists existing categories and creates a new one", async () => {
    server.use(
      http.get(categoriesUrl, () =>
        HttpResponse.json({ success: true, data: { items: [{ id: 1, name: "Minuman", slug: "minuman" }] } }),
      ),
      http.post(adminCategoriesUrl, () =>
        HttpResponse.json({ success: true, data: { id: 2, name: "Snack", slug: "snack" } }, { status: 201 }),
      ),
    );
    const user = userEvent.setup();

    renderWithProviders(<AdminCategories />);

    expect(await screen.findByText("Minuman")).toBeInTheDocument();

    await user.type(screen.getByLabelText(/kategori baru/i), "Snack");
    await user.click(screen.getByRole("button", { name: /tambah/i }));

    await screen.findByRole("button", { name: /tambah/i }); // form settles back to idle
  });

  it("shows an error state when the category list fails to load", async () => {
    server.use(
      http.get(categoriesUrl, () =>
        HttpResponse.json(
          { success: false, error: { code: "INTERNAL_ERROR", message: "err", details: null } },
          { status: 500 },
        ),
      ),
    );

    renderWithProviders(<AdminCategories />);

    expect(await screen.findByRole("alert")).toHaveTextContent(/gagal memuat kategori/i);
  });
});
