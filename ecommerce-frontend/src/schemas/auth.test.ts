import { describe, expect, it } from "vitest";

import { loginSchema, registerSchema } from "./auth";

describe("loginSchema", () => {
  it("accepts a valid email/password pair", () => {
    expect(loginSchema.safeParse({ email: "a@example.com", password: "x" }).success).toBe(true);
  });

  it("rejects an invalid email", () => {
    const result = loginSchema.safeParse({ email: "not-an-email", password: "x" });
    expect(result.success).toBe(false);
  });
});

describe("registerSchema", () => {
  const base = { name: "Achmad", email: "achmad@example.com" };

  it("accepts a password with uppercase, lowercase, and a digit", () => {
    expect(registerSchema.safeParse({ ...base, password: "Passw0rd" }).success).toBe(true);
  });

  it("rejects a password missing an uppercase letter", () => {
    const result = registerSchema.safeParse({ ...base, password: "password1" });
    expect(result.success).toBe(false);
  });

  it("rejects a password missing a digit", () => {
    const result = registerSchema.safeParse({ ...base, password: "Password" });
    expect(result.success).toBe(false);
  });

  it("rejects a password shorter than 8 characters", () => {
    const result = registerSchema.safeParse({ ...base, password: "Pw0" });
    expect(result.success).toBe(false);
  });

  it("rejects a name shorter than 2 characters", () => {
    const result = registerSchema.safeParse({ ...base, name: "A", password: "Passw0rd" });
    expect(result.success).toBe(false);
  });
});
