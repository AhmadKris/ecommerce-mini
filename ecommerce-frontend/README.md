# Ecommerce Mini — Frontend

Storefront SPA untuk Ecommerce Mini, konsumsi REST API dari
`ecommerce-backend`: login/register, katalog produk, keranjang belanja,
checkout, dan riwayat pesanan.

## Tech Stack

React 19 + Vite, TypeScript, Tailwind CSS v4, React Query, Zustand, React
Router, React Hook Form + Zod, Axios, Vitest + React Testing Library + MSW.

## Fitur

- **Auth** — login/register, sesi tersimpan (access + refresh token), rotasi
  token otomatis saat access token kedaluwarsa.
- **Katalog produk** — list dengan pagination, halaman detail.
- **Keranjang** — tambah/ubah/hapus item, badge jumlah item di header.
- **Checkout** — ringkasan keranjang + ongkos kirim, form alamat pengiriman,
  konfirmasi order setelah submit.
- **Riwayat pesanan** — daftar order milik user yang login dengan status dan
  total.

## Setup

1. Install Node.js 22+.
2. Salin `.env.example` ke `.env.local`, sesuaikan `VITE_API_BASE_URL` kalau
   backend tidak jalan di `http://localhost:8080`.
3. `npm install`
4. `npm run dev` — dev server di `http://localhost:5173`.

## Design tokens

Token warna/tipografi/radius/shadow ada di `src/index.css` (Tailwind v4
CSS-first config, lihat komentar di file itu) — tambahkan token baru di sana
dulu, jangan hardcode hex di komponen.

## Testing & Quality

```
npm run test            # Vitest
npm run test:coverage   # Test + coverage report
npm run lint             # ESLint
npm run typecheck         # tsc --noEmit
npm run format              # Prettier
npm run build                # Production build
```

## Docker

Build production: Node (build) → Nginx non-root (serve static + SPA
fallback):

```
docker compose up -d --build
```

Serve di `http://localhost:3000`. `VITE_API_BASE_URL` di-bake saat build
image (env var Vite adalah compile-time, bukan runtime) — override lewat
build arg kalau perlu target API lain:

```
VITE_API_BASE_URL=https://api.staging.example.com docker compose up -d --build
```

## Struktur Project

```
src/
  pages/            Halaman (login, produk, cart, checkout, order, dll)
  components/
    ui/              Komponen presentational generik (Button, Input, dll)
    <domain>/         Komponen spesifik domain (cart, product, layout)
  hooks/             React Query hooks per resource
  lib/               api-client (Axios), config, format, query-client
  store/             Zustand store (auth)
  types/             Tipe response API
  schemas/           Validasi form (Zod)
  test/              Test utilities, MSW handlers
```

## Status

Fitur inti sudah selesai: setup project, API client + auth store, halaman
login/register, katalog produk, keranjang, checkout, riwayat pesanan. Belum
ada: admin panel, halaman profile/`/me`.
