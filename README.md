# Dompet-Mandiri
# 💰 Simple Personal Dompet Mandiri

<img width="1345" height="601" alt="image" src="https://github.com/user-attachments/assets/dbbf1f13-7111-4459-89e3-ec7027b93a2d" />

Aplikasi web manajemen keuangan personal berbasis **Go (Golang)** yang dirancang untuk mencatat pemasukan dan pengeluaran secara mandiri. Aplikasi ini dilengkapi dengan visualisasi tren arus kas bulanan menggunakan grafik garis interaktif, serta ringkasan statistik harian dan bulanan.

## 🚀 Fitur Utama

*   **Autentikasi Pengguna**: Registrasi akun dan Login aman menggunakan enkripsi password (bcrypt) dan manajemen session berbasis cookie.
*   **Ringkasan Dashboard**: Informasi saldo real-time, statistik akumulasi harian, dan statistik bulanan.
*   **Grafik Arus Kas Interaktif**: Visualisasi data pemasukan dan pengeluaran bulanan menggunakan *Line Chart* (Chart.js) yang responsif.
*   **Pencatatan Transaksi**: Form pencatatan transaksi baru dengan kategori kustom serta deskripsi opsional.
*   **Jurnal & Riwayat Log**: Tabel riwayat seluruh transaksi yang tercatat secara kronologis, lengkap dengan formatting mata uang Rupiah (`Rp`).

---

## 🛠️ Arsitektur & Teknologi

### Backend
*   **Bahasa Pemrograman**: Go (Golang)
*   **Web Framework**: [Gin Gonic](https://gin-gonic.com/)
*   **ORM**: [GORM](https://gorm.io/) (Go Object Relational Mapping)
*   **Session Management**: `github.com/gin-contrib/sessions`

### Frontend & Visualisasi
*   **UI Framework**: Bootstrap 5 (Responsive Layout)
*   **Grafik Engine**: Chart.js (Line Chart dengan smoothing curve)
*   **Templating**: Go `html/template` engine dengan custom template functions (untuk format Rupiah)

### Database
*   **DBMS**: PostgreSQL (Menggunakan agregasi waktu native `TO_CHAR` dan `SUM CASE` untuk performa query grafik yang optimal)

---

## 📁 Struktur Direktori Proyek

```text
├── main.go               # Logika backend utama, routing, database init, & controllers
├── go.mod                # File dependency Go
├── go.sum                # Verifikasi checksum dependency Go
└── views/                # Direktori template tampilan (HTML)
    ├── dashboard.html    # Halaman utama ringkasan saldo & grafik garis
    ├── history.html      # Halaman tabel riwayat semua transaksi
    └── transaction_add.html # Halaman form input transaksi baru
