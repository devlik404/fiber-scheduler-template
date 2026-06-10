# Go Fiber Scheduler Template

Template scheduler dan API Go dengan Fiber dan clean architecture ringan. Pengguna cukup fokus mengubah logic utama di file job atau handler yang digenerate.

## Struktur

```text
cmd/api                 # entrypoint aplikasi
internal/config          # konfigurasi env
internal/http            # handler dan route HTTP
internal/platform/database # koneksi database
internal/platform/server # bootstrap Fiber
internal/scheduler       # engine scheduler
internal/task            # area utama untuk business logic scheduler
timezone                 # embedded timezone data dan helper lokasi
```

## Mulai

```bash
cp .env.example .env
go mod tidy
go run ./cmd/api
```

Log memakai JSON agar mudah dibaca oleh log collector. Level log bisa diatur lewat:

```env
APP_LOG_LEVEL=info
```

Health check:

```bash
curl http://localhost:8080/health
```

## Database

Database bersifat opsional. Default template tetap berjalan tanpa DB karena `DB_PRIMARY_ENABLED=false` dan `DB_SECONDARY_ENABLED=false`.

Template mendukung 1 atau 2 koneksi database:

- `DB_PRIMARY_*` untuk koneksi utama.
- `DB_SECONDARY_*` untuk koneksi kedua jika dibutuhkan.

Driver yang didukung:

- `postgres`
- `mysql`
- `sqlserver`

```env
DB_PRIMARY_ENABLED=true
DB_PRIMARY_DRIVER=postgres
DB_PRIMARY_DSN=postgres://postgres:postgres@localhost:5432/template_sch?sslmode=disable
DB_PRIMARY_MAX_OPEN_CONNS=10
DB_PRIMARY_MAX_IDLE_CONNS=5
DB_PRIMARY_CONN_MAX_LIFETIME=30m
DB_PRIMARY_CONN_MAX_IDLE_TIME=10m
DB_PRIMARY_PING_TIMEOUT=5s

DB_SECONDARY_ENABLED=false
DB_SECONDARY_DRIVER=mysql
DB_SECONDARY_DSN=user:password@tcp(localhost:3306)/template_sch?parseTime=true
```

Contoh DSN:

```env
# PostgreSQL
DB_PRIMARY_DRIVER=postgres
DB_PRIMARY_DSN=postgres://postgres:postgres@localhost:5432/template_sch?sslmode=disable

# MySQL
DB_PRIMARY_DRIVER=mysql
DB_PRIMARY_DSN=user:password@tcp(localhost:3306)/template_sch?parseTime=true

# SQL Server
DB_PRIMARY_DRIVER=sqlserver
DB_PRIMARY_DSN=sqlserver://sa:password@localhost:1433?database=template_sch
```

Koneksi DB dikirim ke setiap job lewat constructor di `internal/task/registry.go`. Di dalam job, gunakan `j.db.Primary` untuk DB utama dan `j.db.Secondary` untuk DB kedua.

## Mengubah Logic Utama

### Scheduler Job

Untuk membuat job baru, gunakan generator:

```bash
go run ./cmd/generate-job --name invoice-sync --schedule "*/5 * * * *"
```

Atau lewat Makefile:

```bash
make generate-job NAME=invoice-sync SCHEDULE="*/5 * * * *"
```

Generator otomatis membuat:

- file job di `internal/task/jobs`
- field schedule di `internal/config/config.go`
- env schedule di `.env.example`
- registry entry di `internal/task/registry.go`

Setelah itu pengguna cukup mengubah logic utama di method `Run` pada file job yang dibuat.

Contoh schedule di `.env`:

```env
EXAMPLE_TASK_SCHEDULE=*/1 * * * *
CASE_TASK_SCHEDULE=*/5 * * * *
```

Scheduler memakai format cron standar 5 field: `minute hour day-of-month month day-of-week`.

### HTTP API

Untuk membuat API baru, gunakan generator:

```bash
go run ./cmd/generate-api --name user-profile --method GET --path /api/user-profile
```

Atau lewat Makefile:

```bash
make generate-api NAME=user-profile METHOD=GET PATH=/api/user-profile
```

Generator otomatis membuat:

- file handler di `internal/http/handler`
- route di `internal/http/routes.go`
- logger API dengan atribut `component=api`
- akses dependency DB lewat `h.deps.DB.Primary` dan `h.deps.DB.Secondary`

Jika ingin membuat banyak API, jalankan generator berulang dengan nama berbeda. Setelah itu cukup isi logic utama di method `Handle` pada masing-masing file handler.

## Timezone

Default scheduler memakai `Asia/Jakarta` dari folder `timezone/Asia/Jakarta`.
Jika `SCHEDULER_LOCATION` diubah ke timezone lain, aplikasi akan mencoba memakai database timezone dari sistem operasi.
