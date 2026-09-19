import { describe, expect, it } from "vitest";

import { productSchema } from "./product";

const validInput = { name: "Kopi Susu", sku: "KOPI-001", price: "18000", stock: "10", categoryId: "1" };

describe("productSchema", () => {
  it("accepts a valid product with an SKU", () => {
    const result = productSchema.safeParse(validInput);
    expect(result.success).toBe(true);
  });

  it("rejects a missing SKU", () => {
    const result = productSchema.safeParse({ ...validInput, sku: "" });
    expect(result.success).toBe(false);
  });

  it("rejects an SKU longer than 64 characters", () => {
    const result = productSchema.safeParse({ ...validInput, sku: "A".repeat(65) });
    expect(result.success).toBe(false);
  });
});
