import type { AccessTokenClaims } from "../types/auth";

/**
 * Decodes (without verifying) a JWT's payload segment. Used only to read
 * non-sensitive claims (user_id, roles, permissions) for UI purposes — the
 * backend is the party that actually verifies the signature.
 */
export function decodeAccessToken(token: string): AccessTokenClaims {
  const payload = token.split(".")[1];
  if (!payload) {
    throw new Error("jwt: token has no payload segment");
  }
  return JSON.parse(base64UrlDecode(payload)) as AccessTokenClaims;
}

function base64UrlDecode(input: string): string {
  const base64 = input.replace(/-/g, "+").replace(/_/g, "/");
  const padded = base64.padEnd(base64.length + ((4 - (base64.length % 4)) % 4), "=");
  return atob(padded);
}
