/** Builds an unsigned-but-well-formed JWT for tests that only decode claims. */
export function fakeAccessToken(
  overrides: Partial<{ user_id: number; roles: string[]; permissions: string[] }> = {},
): string {
  const base64UrlEncode = (payload: object) =>
    btoa(JSON.stringify(payload)).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");

  const header = base64UrlEncode({ alg: "HS256", typ: "JWT" });
  const body = base64UrlEncode({
    sub: "1",
    iat: 1,
    exp: 9999999999,
    user_id: 1,
    roles: [],
    permissions: [],
    ...overrides,
  });
  return `${header}.${body}.fake-signature`;
}
