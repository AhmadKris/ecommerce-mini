/** Token pair returned by POST /auth/login and POST /auth/refresh. */
export interface AuthTokens {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

/** The logged-in user's own profile, from GET /auth/me. */
export interface UserProfile {
  id: number;
  name: string;
  email: string;
}

/**
 * Claims carried by the access token JWT (see internal/auth.AccessClaims on
 * the backend). Decoded client-side purely to drive UI — the backend, not
 * this decode, is the actual authorization boundary.
 */
export interface AccessTokenClaims {
  sub: string;
  iat: number;
  exp: number;
  user_id: number;
  roles: string[];
  permissions: string[];
}
