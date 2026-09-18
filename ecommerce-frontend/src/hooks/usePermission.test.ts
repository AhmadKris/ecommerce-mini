import { renderHook } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";

import { usePermission } from "./usePermission";
import { useAuthStore } from "../store/auth-store";

afterEach(() => {
  useAuthStore.getState().clearSession();
});

describe("usePermission", () => {
  it("returns false when the session has no permissions", () => {
    const { result } = renderHook(() => usePermission("product:create"));
    expect(result.current).toBe(false);
  });

  it("returns true once the store has the matching permission", () => {
    useAuthStore.setState({ permissions: ["product:create"] });

    const { result } = renderHook(() => usePermission("product:create"));

    expect(result.current).toBe(true);
  });
});
