# 🕊️ Berita Duka WebApp (Lelayu & Obituari Digital)

Aplikasi web portal dan pembuatan *microsite* berita duka cita / lelayu digital berbasis **Go** dan **Neon PostgreSQL** (Serverless). Dirancang untuk siap dideploy ke **Vercel** sebagai Serverless Functions maupun dijalankan sebagai *standalone server* (Docker / Fly.io / Railway / VPS lokal).

---

## 🌟 Fitur Utama

1. **Self-Service Input Keluarga (Mobile-First)**:
   - Formulir mandiri yang sederhana, santun, dan minim friksi untuk keluarga berduka.
   - Input profil almarhum/almarhumah (nama lengkap, nama panggilan, usia, tanggal wafat, agama, foto).
   - Susunan keluarga besar yang ditinggalkan (peran & daftar nama).
   - Jadwal & lokasi prosesi lengkap (rumah duka, ibadah pelepasan, pemakaman/kremasi).
   - Rekening tanda kasih duka cita digital (opsional).

2. **Integrasi Cloudinary (Upload Foto Otomatis)**:
   - Pemilihan foto langsung dari galeri atau kamera smartphone.
   - Pengunggahan aman ke CDN Cloudinary melalui endpoint backend `/api/upload` dengan penandatanganan kriptografis (*signed upload*).
   - Pratinjau instan (*live thumbnail preview*).
   - Mendukung format JPG, PNG, WEBP dengan kompresi dan optimasi otomatis.
   - Opsi cadangan input URL langsung jika foto sudah tersedia online.

3. **Microsite Publik Elegan (`/m/:slug`)**:
   - URL khusus yang rapi dan mudah dibagikan (contoh: `/m/alm-bambang-soekoco-2026-10-18-1234`).
   - Dilengkapi **OpenGraph & Twitter Card Meta Tags**: Saat tautan dibagikan ke grup WhatsApp/Telegram, pratinjau kartu foto almarhum dari Cloudinary dan ringkasan jadwal pemakaman langsung muncul rapi.
   - Tombol cepat navigasi: 📍 *Google Maps Rute*, 📲 *Bagikan ke WhatsApp*, 💳 *Salin Nomor Rekening Duka*.

4. **Layanan Pembersihan Otomatis (*Purging Service*)**:
   - Menghapus otomatis data berita duka yang telah kadaluwarsa (`funeral_time` lewat dari `PURGE_RETENTION_DAYS`, default 30 hari).
   - Otomatis dipicu oleh **Vercel Cron Jobs** setiap hari pukul 02:00 WIB via endpoint `/api/purge`.
   - Dilindungi oleh token rahasia `CRON_SECRET` untuk mencegah eksekusi liar.

5. **100% Zero Hardcoding (.env Driven)**:
   - Semua parameter (koneksi database, port, nama situs, masa retensi purging, secret token, URL dasar, kredensial Cloudinary) dikonfigurasi melalui variabel lingkungan (`.env`).

---

## 📁 Struktur Direktori

```text
repos/berita-duka-webapp/
├── api/
│   └── index.go            # Entrypoint Vercel Serverless Function
├── internal/
│   ├── config/
│   │   └── config.go       # Parser konfigurasi .env (Zero hardcoding)
│   ├── db/
│   │   └── db.go           # Koneksi Neon Postgres (pgxpool) & query purging
│   ├── cloudinary/
│   │   └── cloudinary.go   # Integrasi Cloudinary Image Upload API
│   ├── models/
│   │   └── obituary.go     # Model data, validasi, dan generator slug
│   ├── handlers/
│   │   └── handlers.go     # Router HTTP, API JSON, dan handler upload & purge
│   └── views/
│       └── templates.go    # Template HTML Microsite, Form Input, & Beranda
├── .env.example            # Template variabel lingkungan
├── go.mod                  # Definisi Go module (Go 1.22)
├── main.go                 # Entrypoint Standalone Server (Local / Docker / VPS)
├── vercel.json             # Konfigurasi runtime Vercel Go & Cron Purge
└── README.md
```

---

## ⚙️ Variabel Lingkungan (`.env`)

Salin `.env.example` ke `.env`:

```bash
cp .env.example .env
```

| Variabel | Default | Deskripsi |
| :--- | :--- | :--- |
| `PORT` | `8080` | Port listen server standalone |
| `APP_ENV` | `development` | Environment (`development` / `production`) |
| `APP_BASE_URL` | `http://localhost:8080` | Base URL aplikasi (untuk link WA & OpenGraph) |
| `SITE_NAME` | `"Portal Berita Duka & Lelayu Digital"` | Nama portal / brand |
| `DATABASE_URL` | *(wajib diisi)* | Connection string Neon Postgres (`sslmode=require`) |
| `AUTO_MIGRATE` | `true` | Otomatis membuat tabel `obituaries` saat startup |
| `MAX_DB_CONNS` | `10` | Maksimum koneksi pool database |
| `PURGE_RETENTION_DAYS` | `30` | Jumlah hari simpan data setelah tanggal pemakaman |
| `CRON_SECRET` | *(string rahasia)* | Bearer token pengaman endpoint `/api/purge` |
| `DEFAULT_OBITUARY_PHOTO_URL` | *(url gambar)* | Placeholder foto jika keluarga belum melampirkan foto |
| `CLOUDINARY_CLOUD_NAME` | *(opsional)* | Cloud name akun Cloudinary Anda |
| `CLOUDINARY_API_KEY` | *(opsional)* | API Key Cloudinary |
| `CLOUDINARY_API_SECRET` | *(opsional)* | API Secret Cloudinary |
| `CLOUDINARY_FOLDER` | `berita-duka` | Nama folder target penyimpanan aset di Cloudinary |

---

## 🚀 Menjalankan Secara Lokal

1. **Jalankan Aplikasi**:
   ```bash
   go run main.go
   ```
2. **Akses di Browser**:
   * Halaman Utama: `http://localhost:8080/`
   * Form Input Mandiri & Upload Foto: `http://localhost:8080/create`
   * Test Trigger Purge:
     ```bash
     curl -X POST http://localhost:8080/api/purge \
       -H "Authorization: Bearer <CRON_SECRET_ANDA>"
     ```

---

## ☁️ Panduan Deploy ke Vercel

1. **Push ke GitHub**:
   Pastikan repositori ini telah dipush ke branch `main`.
2. **Import Project di Vercel Dashboard**:
   * Pilih repositori `berita-duka-webapp`.
   * Framework Preset: **Other**.
3. **Tambahkan Environment Variables di Vercel**:
   Masuk ke **Project Settings** $\rightarrow$ **Environment Variables**, lalu tambahkan:
   * `DATABASE_URL` = DSN dari dashboard Neon Anda (`postgresql://user:pass@ep-xyz.neon.tech/neondb?sslmode=require`).
   * `APP_BASE_URL` = URL domain produksi Vercel Anda (misal `https://berita-duka.vercel.app`).
   * `PURGE_RETENTION_DAYS` = `30` (atau sesuai kebutuhan).
   * `CRON_SECRET` = String acak aman untuk memvalidasi Vercel Cron.
   * `CLOUDINARY_CLOUD_NAME`, `CLOUDINARY_API_KEY`, `CLOUDINARY_API_SECRET` = Kredensial dari dashboard Cloudinary.
4. **Deploy**:
   Vercel akan otomatis mendeteksi `vercel.json`, mengompilasi `api/index.go`, dan mengaktifkan jadwal cron purge harian secara otomatis.
