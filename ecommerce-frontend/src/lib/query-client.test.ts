import { describe, expect, it } from "vitest";

import { shouldRetryQuery } from "./query-client";
import { ApiError } from "../types/api";

describe("shouldRetryQuery", () => {
  it("does not retry a 404", () => {
    expect(shouldRetryQuery(0, new ApiError("Not found", "NOT_FOUND", [], 404))).toBe(false);
  });

  it("does not retry a 400", () => {
    expect(shouldRetryQuery(0, new ApiError("Bad request", "VALIDATION_ERROR", [], 400))).toBe(false);
  });

  it("retries a 500 once", () => {
    expect(shouldRetryQuery(0, new ApiError("Server error", "INTERNAL_ERROR", [], 500))).toBe(true);
    expect(shouldRetryQuery(1, new ApiError("Server error", "INTERNAL_ERROR", [], 500))).toBe(false);
  });

  it("retries a non-ApiError (e.g. network failure) once", () => {
    expect(shouldRetryQuery(0, new Error("network down"))).toBe(true);
    expect(shouldRetryQuery(1, new Error("network down"))).toBe(false);
  });
});
