import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { afterEach, describe, expect, it } from "vitest";

import { Profile } from "./Profile";
import { config } from "../../lib/config";
import { useAuthStore } from "../../store/auth-store";
import { fakeAccessToken } from "../../test/fake-jwt";
import { server } from "../../test/msw-server";
import { renderWithProviders } from "../../test/test-utils";

const meUrl = `${config.apiBaseUrl}/auth/me`;
const addressesUrl = `${config.apiBaseUrl}/profile/addresses`;

afterEach(() => useAuthStore.getState().clearSession());

function logIn() {
  useAuthStore.getState().setSession({ access_token: fakeAccessToken(), refresh_token: "r", expires_in: 900 });
}

const sampleAddress = {
  id: 1,
  user_id: 1,
  recipient_name: "Budi",
  phone: "08123456789",
  address_line: "Jl. Sudirman No. 1",
  city: "Jakarta",
  province: "DKI Jakarta",
  postal_code: "10110",
  is_default: true,
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

describe("Profile page", () => {
  it("renders the user's name/email and their saved addresses", async () => {
    logIn();
    server.use(
      http.get(meUrl, () => HttpResponse.json({ success: true, data: { id: 1, name: "Budi", email: "budi@test.com" } })),
      http.get(addressesUrl, () => HttpResponse.json({ success: true, data: { items: [sampleAddress] } })),
    );

    renderWithProviders(<Profile />);

    expect(await screen.findByText("budi@test.com")).toBeInTheDocument();
    expect(screen.getAllByText("Budi")).toHaveLength(2); // profile info + address recipient
    expect(screen.getByText(/jl\. sudirman no\. 1/i)).toBeInTheDocument();
    expect(screen.getByText("Utama")).toBeInTheDocument();
  });

  it("adds a new address", async () => {
    logIn();
    let receivedBody: unknown;
    server.use(
      http.get(meUrl, () => HttpResponse.json({ success: true, data: { id: 1, name: "Budi", email: "budi@test.com" } })),
      http.get(addressesUrl, () => HttpResponse.json({ success: true, data: { items: [] } })),
      http.post(addressesUrl, async ({ request }) => {
        receivedBody = await request.json();
        return HttpResponse.json({ success: true, data: { ...sampleAddress, id: 2 } }, { status: 201 });
      }),
    );
    const user = userEvent.setup();

    renderWithProviders(<Profile />);

    await user.click(await screen.findByRole("button", { name: /tambah alamat/i }));
    await user.type(screen.getByLabelText(/nama penerima/i), "Siti");
    await user.type(screen.getByLabelText(/nomor telepon/i), "08129999999");
    await user.type(screen.getByLabelText(/alamat lengkap/i), "Jl. Thamrin No. 5");
    await user.type(screen.getByLabelText(/^kota$/i), "Jakarta");
    await user.type(screen.getByLabelText(/provinsi/i), "DKI Jakarta");
    await user.type(screen.getByLabelText(/kode pos/i), "10250");
    await user.click(screen.getByRole("button", { name: /simpan alamat/i }));

    await screen.findByRole("button", { name: /tambah alamat/i });
    expect(receivedBody).toMatchObject({ recipient_name: "Siti", city: "Jakarta" });
  });

  it("shows an error state when the address book fails to load", async () => {
    logIn();
    server.use(
      http.get(meUrl, () => HttpResponse.json({ success: true, data: { id: 1, name: "Budi", email: "budi@test.com" } })),
      http.get(addressesUrl, () =>
        HttpResponse.json(
          { success: false, error: { code: "INTERNAL_ERROR", message: "err", details: null } },
          { status: 500 },
        ),
      ),
    );

    renderWithProviders(<Profile />);

    expect(await screen.findByRole("alert")).toHaveTextContent(/gagal memuat buku alamat/i);
  });
});
