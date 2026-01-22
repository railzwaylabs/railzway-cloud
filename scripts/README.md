# Migration Script

Script untuk menjalankan database migration dengan mudah menggunakan `golang-migrate`.

## Prerequisites

Install `golang-migrate`:
```bash
brew install golang-migrate
```

## Setup

Tambahkan `DB_SOURCE` ke file `.env`:
```bash
DB_SOURCE=postgres://user:password@localhost:5432/dbname?sslmode=disable
```

Atau gunakan format dari environment variables yang ada:
```bash
DB_SOURCE=postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSL_MODE}
```

## Usage

### Menjalankan semua migration (up)
```bash
./scripts/migrate.sh up
```

### Rollback semua migration (down)
```bash
./scripts/migrate.sh down
```

### Menjalankan N migration
```bash
./scripts/migrate.sh up 1
./scripts/migrate.sh down 2
```

### Membuat migration baru
```bash
./scripts/migrate.sh create add_new_table
```

### Force version (jika migration stuck)
```bash
./scripts/migrate.sh force 3
```

### Cek versi migration saat ini
```bash
./scripts/migrate.sh version
```

## Migration Files

Migration files berada di `sql/migrations/`:
- `NNNNNN_name.up.sql` - untuk apply migration
- `NNNNNN_name.down.sql` - untuk rollback migration

## Troubleshooting

Jika ada error "dirty database", gunakan:
```bash
./scripts/migrate.sh force [version]
```
