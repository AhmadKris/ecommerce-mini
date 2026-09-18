import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { Home } from "./Home";
import { renderWithProviders } from "../test/test-utils";

describe("Home", () => {
  it("renders the landing headline and a link to the product catalog", () => {
    renderWithProviders(<Home />);

    expect(screen.getByRole("heading", { name: /belanja lebih tenang/i })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /lihat produk/i })).toHaveAttribute(
      "href",
      "/products",
    );
  });
});
