import { setupServer } from "msw/node";

/**
 * Shared MSW server for tests. Individual test files add their handlers via
 * `server.use(...)` in the test itself, rather than declaring app-wide
 * handlers here — keeps each test's expectations next to the test.
 */
export const server = setupServer();
