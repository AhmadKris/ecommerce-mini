import { describe, expect, it } from "vitest";

import { decodeAccessToken } from "./jwt";
import { fakeAccessToken } from "../test/fake-jwt";

describe("decodeAccessToken", () => {
  it("decodes user_id, roles, and permissions from the payload", () => {
    const token = fakeAccessToken({
      user_id: 1,
      roles: ["customer"],
      permissions: ["cart:manage", "order:read_own"],
    });

    const claims = decodeAccessToken(token);

    expect(claims.user_id).toBe(1);
    expect(claims.roles).toEqual(["customer"]);
    expect(claims.permissions).toEqual(["cart:manage", "order:read_own"]);
  });

  it("throws when the token has no payload segment", () => {
    expect(() => decodeAccessToken("not-a-jwt")).toThrow();
  });
});
