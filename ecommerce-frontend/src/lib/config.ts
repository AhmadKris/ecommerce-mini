/**
 * Runtime configuration read once from `import.meta.env`. Validated at
 * import time so a missing required env var fails fast at app start,
 * instead of surfacing as a confusing runtime error deep in a component.
 */

function requireEnv(key: string): string {
  const value = import.meta.env[key];
  if (!value) {
    throw new Error(`config: missing required env var ${key} (check .env.local)`);
  }
  return value;
}

export const config = {
  apiBaseUrl: requireEnv("VITE_API_BASE_URL"),
};
