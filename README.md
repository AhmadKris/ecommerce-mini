# Ecommerce Mini — Production-Ready Architecture

Platform E-commerce Mini skala kecil dengan standar **production-ready** sungguhan (REST API Go + React SPA Storefront & Admin Panel). Dibangun dengan prinsip high availability, resilience, observability, dan granular RBAC.

[![CI/CD Pipeline](https://github.com/AhmadKris/ecommerce-mini/actions/workflows/ci.yml/badge.svg)](https://github.com/AhmadKris/ecommerce-mini/actions/workflows/ci.yml)

---

## 🌟 Fitur & 5 Pilar Production-Ready

### 1. Granular RBAC (Role-Based Access Control)
- Multi-role per pengguna (`admin`, `customer`, dsb.).
- Permission code berformat `resource:action` (`product:create`, `order:read_all`, `user:manage`).
- Permission disuntikkan langsung ke dalam **JWT Access Token claims** untuk verifikasi middleware berkecepatan tinggi tanpa DB lookup overhead (~0ms).
- Halaman **Manajemen Role & Akses** tersendiri untuk mengalokasikan role pengguna dengan proteksi *Self-Lockout Prevention*.

### 2. High Resilience & Atomic Checkout
- **Strict Race Condition Prevention**: Pengurangan stok saat checkout dikunci secara atomic di database (`SELECT ... FOR UPDATE`), mencegah *overselling* saat ada lonjakan request bersamaan.
- **Idempotency Key**: Endpoint checkout mendukung header `Idempotency-Key` untuk mencegah *double-charge* saat client melakukan retry.
- **Audit Logging**: Mencatat setiap perubahan sensitif (mutasi stok, pembaruan role, update status order) ke dalam tabel `audit_logs`.

### 3. Reliable Outbox Pattern & Background Worker
- Penulisan event `order.created` dieksekusi **atomic** di dalam transaksi DB checkout yang sama ke tabel `outbox_events`.
- Process **Background Worker** terpisah memproses tugas asinkron (notifikasi email, pemicuan fulfillment queue) secara andal (*At-Least-Once Delivery*) dengan *graceful shutdown* saat menerima `SIGTERM`.

### 4. Redis Cache-Aside Pattern
- Pembacaan katalog produk (`GET /api/products`, `GET /api/products/:slug`) dan kategori menggunakan strategi **Cache-Aside** dengan Redis TTL (5–10 menit).
- Auto-invalidation proaktif (`products:*`, `categories:*`) saat terjadi mutasi produk/kategori oleh admin.

### 5. Full Observability (Metrics & Tracing)
- **Prometheus Exporter**: Endpoint `/metrics` mengukur HTTP request counter, status code, latency histogram (`http_request_duration_seconds`), dan in-flight requests.
- **OpenTelemetry Tracing**: Middleware menyuntikkan header response `X-Trace-ID` dan mempropagasi context span tracer.
- **Dashboard Grafana**: Config preset dashboard Grafana ter-provisioning otomatis untuk memantau RPS dan p95/p99 latency.

---

## 📐 Arsitektur Sistem

```
                    ┌─────────────────────────┐
                    │      Browser Client     │
                    └────────────┬────────────┘
                                 │
                                 ▼
                     ┌───────────────────────┐
                     │     Nginx Web SPA     │ (Port 3000)
                     └───────────┬───────────┘
                                 │
                                 ▼
                     ┌───────────────────────┐
                     │    Go REST API App    │ (Port 8080)
                     └───┬───────────┬───────┘
                         │           │
           ┌─────────────┴─┐       ┌─┴─────────────┐
           │ PostgreSQL 16 │       │    Redis 7    │
           │  (DB + Outbox)│       │(Cache-Aside)  │
           └───────┬───────┘       └───────────────┘
                   │
                   ▼
         ┌──────────────────┐
         │ Background Worker│ (Internal Goroutine Loop)
         └──────────────────┘
```

---

## 📑 Architecture Decision Records (ADRs)

Arsitektur dan pilihan desain penting telah didokumentasikan dalam format ADR:
- 📄 **[ADR 001: Custom Granular RBAC vs Casbin](docs/adr/001-custom-rbac-vs-casbin.md)**
- 📄 **[ADR 002: Outbox Pattern & Background Worker untuk Async Order Events](docs/adr/002-outbox-pattern-for-async-events.md)**
- 📄 **[ADR 003: Redis Cache-Aside Pattern & Proactive Invalidation](docs/adr/003-redis-cache-aside-and-invalidation.md)**

---

## 🚀 Menjalankan Aplikasi

### Option A: Full Production Stack (Docker Compose)

Jalankan seluruh stack (PostgreSQL, Redis, Migration, API Backend, Web Frontend, Prometheus, dan Grafana) sekaligus:

```bash
docker compose -f docker-compose.prod.yml up -d --build
```

| Service | Endpoint | Deskripsi |
|---|---|---|
| **Storefront & Admin SPA** | http://localhost:3000 | Frontend Web React SPA |
| **Go REST API** | http://localhost:8080 | Endpoint API (`/api/products`, `/api/orders`, `/metrics`, `/health`) |
| **Prometheus Server** | http://localhost:9090 | Metrics Scraping Engine |
| **Grafana Dashboard** | http://localhost:3001 | Dashboard Monitoring (Login: `admin` / `admin`) |

### Option B: Local Development

Untuk pengembangan lokal (hot reload):
1. **Backend**:
   ```bash
   cd ecommerce-backend
   go run ./cmd/api
   ```
2. **Frontend**:
   ```bash
   cd ecommerce-frontend
   npm run dev
   ```

---

## 🧪 Pengujian & Quality Assurance

### 1. Backend & Frontend Tests
```bash
# Backend unit & integration tests
cd ecommerce-backend && go test -v ./...

# Frontend component & schema tests
cd ecommerce-frontend && npm test -- --run
```

### 2. Load Testing dengan k6
Eksekusi pengujian beban (*load test*) untuk menguji performa endpoint katalog dan checkout:

```bash
# Install k6 atau gunakan npx / docker run k6
k6 run tests/load/k6_test.js
```

Threshold SLA yang diuji:
- `http_req_duration`: p95 < 200ms, p99 < 500ms
- `http_req_failed`: error rate < 1%

---

## 🔑 Akun Demo (Default Seed Data)

- **Admin Account**:
  - Email: `admin@example.com`
  - Password: `Password123!`
  - Role: `admin` (Hak akses penuh ke Admin Panel & Manajemen Role)
- **Customer Account**:
  - Email: `customer@example.com`
  - Password: `Password123!`
  - Role: `customer` (Akses Storefront, Cart, Checkout, & Orders)
