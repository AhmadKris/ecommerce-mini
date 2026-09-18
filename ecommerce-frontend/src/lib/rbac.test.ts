import { describe, expect, it } from "vitest";

import { hasPermission } from "./rbac";

describe("hasPermission", () => {
  it("returns true when the code is present", () => {
    expect(hasPermission(["order:read_own", "cart:manage"], "cart:manage")).toBe(true);
  });

  it("returns false when the code is absent", () => {
    expect(hasPermission(["order:read_own"], "product:create")).toBe(false);
  });

  it("returns false for an empty permission list", () => {
    expect(hasPermission([], "product:create")).toBe(false);
  });
});
