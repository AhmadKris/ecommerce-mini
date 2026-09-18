# Ecommerce Mini

Platform e-commerce skala kecil: REST API (Go) + storefront SPA (React), dengan
autentikasi berbasis JWT, role-based access control, katalog produk, keranjang
belanja, dan checkout yang aman dari race condition (stok tidak akan oversell
walau ada banyak checkout bersamaan untuk produk yang sama).

Monorepo, dua bagian independen:

```
ecommerce-mini/
├── ecommerce-backend/    Go REST API
├── ecommerce-frontend/   React SPA (storefront)
└── docker-compose.yml    Menjalankan keduanya + Postgres + Redis sekaligus
```

## Fitur

- **Autentikasi & RBAC** — register/login dengan JWT (access + refresh token),
  role `admin`/`customer`, permission per-endpoint (mis. hanya admin yang bisa
  membuat/mengubah produk).
- **Katalog produk** — list dengan pagination, detail produk, kategori.
- **Keranjang belanja** — tambah/ubah/hapus item, tervalidasi terhadap stok
  yang tersedia.
- **Checkout & order** — mengubah keranjang jadi order secara atomic: stok
  dikunci dan dikurangi dalam satu transaction database (`SELECT ... FOR
  UPDATE`), sehingga dua checkout bersamaan untuk stok terakhir yang sama
  tidak akan menyebabkan oversell. Harga produk di-snapshot saat checkout,
  ongkos kirim dihitung dan dipersist di sisi server.
- **Riwayat pesanan** — daftar order milik user yang login, dengan status dan
  rincian item.
- **Admin panel** — kelola produk & kategori, lihat semua pesanan dari
  seluruh customer.
- **Hardening** — rate limiting (endpoint auth), idempotency key di
  checkout, audit log untuk action sensitif.

## Tech Stack

| Layer | Teknologi |
|---|---|
| Backend | Go, Gin, GORM, PostgreSQL, Redis, JWT (`golang-jwt/v5`), `golang-migrate` |
| Frontend | React 19, TypeScript, Vite, Tailwind CSS v4, React Query, Zustand, React Router, React Hook Form + Zod |
| Testing | Go standard testing (backend), Vitest + React Testing Library + MSW (frontend) |
| Infra | Docker, Docker Compose, Nginx (serve frontend statis) |

## Menjalankan Project

### Semua sekaligus lewat Docker (cara tercepat)

```
docker compose up -d --build
```

Menjalankan Postgres, Redis, migration database (sekali jalan lalu exit), API,
dan frontend dalam satu jaringan Docker:

| Service | URL | Keterangan |
|---|---|---|
| Frontend | http://localhost:3000 | Storefront SPA |
| API | http://localhost:8080 | `GET /health`, `GET /ready`, `GET /api/products`, dst |
| PostgreSQL | localhost:5433 | |
| Redis | localhost:6379 | |

Cek status: `docker compose ps` — semua service harus `healthy` (`migrate`
`Exited (0)` itu normal, dia memang sekali jalan). Matikan dengan
`docker compose down` (tambahkan `-v` untuk ikut hapus data Postgres).

### Development aktif (hot reload)

Untuk development sehari-hari (tanpa rebuild image Docker tiap ubah kode),
jalankan backend dan frontend secara native — lihat instruksi detail di
masing-masing:

- [`ecommerce-backend/README.md`](ecommerce-backend/README.md)
- [`ecommerce-frontend/README.md`](ecommerce-frontend/README.md)

## API

Spesifikasi lengkap ada di [`ecommerce-backend/docs/openapi.yaml`](ecommerce-backend/docs/openapi.yaml).
Endpoint utama:

| Endpoint | Keterangan |
|---|---|
| `POST /api/auth/register`, `POST /api/auth/login` | Registrasi & login |
| `POST /api/auth/refresh` | Perpanjang access token |
| `GET /api/auth/me` | Profil user yang login |
| `GET /api/products`, `GET /api/products/:slug` | Katalog produk |
| `POST/PUT /api/admin/products` | Kelola produk (admin) |
| `GET /api/categories` | List kategori |
| `POST/PUT/DELETE /api/admin/categories` | Kelola kategori (admin) |
| `GET/POST/PUT/DELETE /api/cart` | Keranjang belanja |
| `POST /api/orders` | Checkout (header opsional `Idempotency-Key`) |
| `GET /api/orders` | Riwayat pesanan sendiri |
| `GET /api/admin/orders` | Semua pesanan (admin) |

## Status

**Fase 1 (Foundation) — selesai**, di kedua sisi. **Fase 2 (Production
Hardening) — hampir selesai**:

- Backend: scaffold, auth + RBAC, CRUD produk & kategori, keranjang,
  checkout/order (idempotency key, audit log), rate limiting, `/auth/me`,
  admin order list, integration test (testcontainers-go).
- Frontend: setup project, API client + auth store, halaman login/register,
  katalog produk, keranjang, checkout, riwayat pesanan, admin panel (produk,
  kategori, semua pesanan).

Belum ada: circuit breaker (menunggu integrasi payment gateway eksternal
sungguhan), observability (Fase 3 — Prometheus/Grafana, tracing).
