import { HttpResponse, http } from "msw";
import { afterEach, describe, expect, it } from "vitest";

import { apiClient } from "./api-client";
import { config } from "./config";
import { fakeAccessToken } from "../test/fake-jwt";
import { server } from "../test/msw-server";
import { useAuthStore } from "../store/auth-store";
import { ApiError } from "../types/api";

const url = (path: string) => `${config.apiBaseUrl}${path}`;

afterEach(() => {
  useAuthStore.getState().clearSession();
});

describe("apiClient", () => {
  it("unwraps the success envelope so callers get the inner data directly", async () => {
    server.use(
      http.get(url("/products"), () =>
        HttpResponse.json({
          success: true,
          data: { items: [], meta: { page: 1, limit: 10, total: 0, total_pages: 0 } },
        }),
      ),
    );

    const response = await apiClient.get("/products");

    expect(response.data).toEqual({
      items: [],
      meta: { page: 1, limit: 10, total: 0, total_pages: 0 },
    });
  });

  it("rejects with an ApiError built from the error envelope", async () => {
    server.use(
      http.get(url("/products/unknown-slug"), () =>
        HttpResponse.json(
          {
            success: false,
            error: { code: "NOT_FOUND", message: "Produk tidak ditemukan", details: null },
          },
          { status: 404 },
        ),
      ),
    );

    await expect(apiClient.get("/products/unknown-slug")).rejects.toMatchObject({
      code: "NOT_FOUND",
      message: "Produk tidak ditemukan",
      status: 404,
    });
  });

  it("attaches the bearer token from the auth store", async () => {
    useAuthStore.setState({ accessToken: "test-access-token" });
    let receivedAuth: string | null = null;
    server.use(
      http.get(url("/orders"), ({ request }) => {
        receivedAuth = request.headers.get("Authorization");
        return HttpResponse.json({ success: true, data: [] });
      }),
    );

    await apiClient.get("/orders");

    expect(receivedAuth).toBe("Bearer test-access-token");
  });

  it("on 401, refreshes the token once and retries the original request", async () => {
    useAuthStore.setState({ accessToken: "expired-token", refreshToken: "valid-refresh-token" });
    let ordersCallCount = 0;

    server.use(
      http.get(url("/orders"), () => {
        ordersCallCount += 1;
        if (ordersCallCount === 1) {
          return HttpResponse.json(
            {
              success: false,
              error: { code: "UNAUTHORIZED", message: "Token kedaluwarsa", details: null },
            },
            { status: 401 },
          );
        }
        return HttpResponse.json({ success: true, data: [{ id: 1 }] });
      }),
      http.post(url("/auth/refresh"), () =>
        HttpResponse.json({
          success: true,
          data: {
            access_token: fakeAccessToken(),
            refresh_token: "rotated-refresh-token",
            expires_in: 900,
          },
        }),
      ),
    );

    const response = await apiClient.get("/orders");

    expect(ordersCallCount).toBe(2);
    expect(response.data).toEqual([{ id: 1 }]);
    expect(useAuthStore.getState().accessToken).toBe(fakeAccessToken());
  });

  it("on 401 with a failed refresh, clears the session and rejects", async () => {
    useAuthStore.setState({
      accessToken: "expired-token",
      refreshToken: "already-used-refresh-token",
    });

    server.use(
      http.get(url("/orders"), () =>
        HttpResponse.json(
          {
            success: false,
            error: { code: "UNAUTHORIZED", message: "Token kedaluwarsa", details: null },
          },
          { status: 401 },
        ),
      ),
      http.post(url("/auth/refresh"), () =>
        HttpResponse.json(
          {
            success: false,
            error: { code: "UNAUTHORIZED", message: "Refresh token tidak valid", details: null },
          },
          { status: 401 },
        ),
      ),
    );

    await expect(apiClient.get("/orders")).rejects.toBeInstanceOf(ApiError);
    expect(useAuthStore.getState().accessToken).toBeNull();
  });

  it("does not attempt a refresh for a 401 from /auth/login itself, even with a stale session in the store", async () => {
    // Regression test: found via a real browser check — submitting a wrong
    // password on /login while a still-valid session sat in sessionStorage
    // (e.g. from another tab) silently rotated that unrelated session's
    // refresh token as a side effect.
    useAuthStore.setState({
      accessToken: "stale-token",
      refreshToken: "stale-but-still-valid-refresh-token",
    });
    let refreshCallCount = 0;

    server.use(
      http.post(url("/auth/login"), () =>
        HttpResponse.json(
          {
            success: false,
            error: { code: "UNAUTHORIZED", message: "Email atau password salah", details: null },
          },
          { status: 401 },
        ),
      ),
      http.post(url("/auth/refresh"), () => {
        refreshCallCount += 1;
        return HttpResponse.json({
          success: true,
          data: { access_token: fakeAccessToken(), refresh_token: "rotated", expires_in: 900 },
        });
      }),
    );

    await expect(
      apiClient.post("/auth/login", { email: "a@b.com", password: "wrong" }),
    ).rejects.toMatchObject({
      code: "UNAUTHORIZED",
    });

    expect(refreshCallCount).toBe(0);
    expect(useAuthStore.getState().refreshToken).toBe("stale-but-still-valid-refresh-token");
  });
});
