import { describe, expect, it } from "vitest";

import { formatCurrency } from "./format";

describe("formatCurrency", () => {
  it("formats a whole number as Rupiah with thousands separators", () => {
    // Intl inserts a non-breaking space after "Rp" — match loosely on
    // whitespace rather than pin the exact character.
    expect(formatCurrency(18000)).toMatch(/^Rp\s*18\.000$/);
  });

  it("formats zero", () => {
    expect(formatCurrency(0)).toMatch(/^Rp\s*0$/);
  });

  it("rounds off fractional amounts (no sub-unit IDR)", () => {
    expect(formatCurrency(1500.75)).toMatch(/^Rp\s*1\.501$/);
  });
});
