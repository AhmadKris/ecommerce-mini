# Ecommerce Mini — Backend

REST API untuk platform e-commerce mini: autentikasi JWT, role-based access
control, katalog produk, keranjang belanja, dan checkout yang aman dari race
condition (stok dikunci & dikurangi dalam satu transaction database, sehingga
checkout bersamaan untuk stok terakhir yang sama tidak akan oversell).

## Tech Stack

Go 1.26, Gin, GORM + PostgreSQL, Redis, JWT (`golang-jwt/jwt/v5`), Docker,
`golang-migrate`.

## Fitur

- **Auth & RBAC** — register/login dengan access + refresh token, role
  `admin`/`customer`, permission per-endpoint lewat middleware.
- **Produk & kategori** — list dengan pagination, detail via slug, CRUD
  produk (admin only).
- **Keranjang** — tambah/ubah/hapus item, tervalidasi terhadap stok.
- **Checkout & order** — satu transaction atomic: cek stok, kurangi stok
  (row lock `SELECT ... FOR UPDATE`), buat order + order items + payment,
  kosongkan cart. Harga produk di-snapshot per item, ongkos kirim dihitung
  dan dipersist di `orders.shipping_cost`.
- **Riwayat order** — daftar order milik user yang login, dengan pagination.

## Setup

### Opsi A — semua via Docker (tercepat, cocok untuk coba-coba)

```
docker compose up -d --build
```

Ini menjalankan Postgres + Redis, lalu service `migrate` (jalan sekali, apply
semua migration, exit), baru `api` start setelah `migrate` sukses. Cek
`docker compose ps` — semua service harus `healthy`/`Exited (0)` untuk
`migrate`. Server di `http://localhost:8080`.

### Opsi B — Go native di host (untuk development aktif, hot reload lebih cepat)

1. Install Go 1.26+, Docker Desktop, dan `golang-migrate` CLI (opsional — ada
   fallback Docker image di bawah kalau tidak mau install).
2. Salin `.env.example` ke `.env`, sesuaikan `JWT_ACCESS_SECRET` /
   `JWT_REFRESH_SECRET` untuk development lokal.
3. Jalankan dependency (PostgreSQL + Redis saja, tanpa `api`):
   ```
   docker compose up -d postgres redis
   ```
4. Jalankan migration:
   ```
   make migrate-up
   ```
   Kalau `migrate` CLI belum terinstall di mesin lokal, pakai image Docker
   resminya (ganti nama container sesuai `DATABASE_URL` di `.env`):
   ```
   docker run --rm --network ecommerce-backend_default \
     -v "$(pwd)/migrations:/migrations" migrate/migrate:v4.17.0 \
     -path=/migrations -database "postgres://postgres:postgres@ecommerce-backend-postgres-1:5432/ecommerce_mini?sslmode=disable" up
   ```
5. Jalankan server:
   ```
   make run
   ```
   Server listen di `PORT` dari `.env` (default `8080`). Cek `GET /health`
   dan `GET /ready`.

> Catatan: `docker-compose.yml` di root repo (`ecommerce-mini/`) menjalankan
> backend ini bersama frontend dalam satu jaringan, untuk uji integrasi
> full-stack — pakai itu kalau ingin lihat storefront benar-benar bicara ke
> API ini.

## API

Spesifikasi lengkap ada di [`docs/openapi.yaml`](docs/openapi.yaml).

## Testing

```
make test              # semua test
make test-coverage     # test + coverage report
make lint              # golangci-lint (butuh golangci-lint terinstall)
```

## Struktur Project

```
cmd/api/            Entry point
internal/
  config/            Load & validasi env var
  database/, cache/  Wiring PostgreSQL & Redis
  model/             Struct domain + DTO
  repository/        Akses data (GORM)
  service/           Business logic
  handler/           HTTP handler (Gin)
  middleware/         Auth, RBAC, CORS, dll
  router/            Registrasi route
  apperror/          Error terstruktur (code, message, HTTP status)
migrations/          SQL migration (golang-migrate)
docs/openapi.yaml    Spesifikasi API
```

## Status

Backend sudah menyelesaikan fitur inti: scaffold, migration, auth + RBAC,
CRUD produk, keranjang, checkout/order. Belum ada: rate limiting, audit log,
integration test otomatis untuk repository layer, panel admin di sisi API
(endpoint yang sudah ada cukup untuk dikonsumsi admin panel, tapi belum ada
endpoint manajemen user/role).
