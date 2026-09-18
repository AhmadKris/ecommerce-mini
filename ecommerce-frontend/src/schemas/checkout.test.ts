import { describe, expect, it } from "vitest";

import { checkoutSchema } from "./checkout";

describe("checkoutSchema", () => {
  it("accepts an address of 10+ characters", () => {
    expect(checkoutSchema.safeParse({ shippingAddress: "Jl. Merdeka No. 1" }).success).toBe(true);
  });

  it("rejects an address shorter than 10 characters", () => {
    const result = checkoutSchema.safeParse({ shippingAddress: "Jl. A" });
    expect(result.success).toBe(false);
  });
});
