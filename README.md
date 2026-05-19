# Dompet Mandiri - Personal Finance Tracker & Dashboard

<img width="1345" height="601" alt="image" src="https://github.com/user-attachments/assets/dbbf1f13-7111-4459-89e3-ec7027b93a2d" />

**Dompet Mandiri** adalah aplikasi berbasis web mandiri yang dirancang untuk membantu Anda mencatat, melacak, dan menganalisis arus kas keuangan pribadi (Pemasukan dan Pengeluaran). Aplikasi ini dibangun dengan performa tinggi, sistem autentikasi multi-user yang aman, dashboard ringkasan harian/bulanan, serta fitur ekspor laporan langsung ke format Excel.

## 🚀 Fitur Utama

*   **Autentikasi Multi-User:** Registrasi dan login aman menggunakan enkripsi password `bcrypt` dan manajemen session berbasis cookie.
*   **Ringkasan Dashboard Dinamis:** Menampilkan total saldo saat ini, statistik pemasukan/pengeluaran harian, serta akumulasi bulanan secara *real-time*.
*   **Analisis Arus Kas Bulanan:** Agregasi data otomatis menggunakan query PostgreSQL canggih untuk menyajikan perbandingan *Income*, *Expense*, dan *Net Profit* dalam 6 bulan terakhir.
*   **Manajemen Transaksi:** Form input transaksi yang mudah digunakan dengan pembagian berdasarkan kategori dan deskripsi yang fleksibel.
*   **Riwayat Transaksi:** Halaman histori penuh yang diurutkan dari transaksi terbaru.
*   **Ekspor Laporan Excel:** Fitur *on-the-fly streaming* data transaksi langsung menjadi file laporan `.xlsx` yang rapi dan siap cetak menggunakan library `excelize`.

## 🛠️ Tech Stack

*   **Backend:** [Go (Golang)](https://go.dev/) dengan [Gin Web Framework](https://gin-gonic.com/)
*   **Database ORM:** [GORM](https://gorm.io/)
*   **Database Server:** [PostgreSQL](https://www.postgresql.org/) (Hosted on Supabase)
*   **Frontend:** HTML5 Templates, Tailwind CSS (Mobile-First Design)
*   **Excel Engine:** [Excelize v2](https://github.com/xuri/excelize)

## 📁 Struktur Proyek

```text
├── main.go               # Logika utama aplikasi Dompet Mandiri, routing, database & controller
├── views/                # Folder template tampilan HTML
│   ├── home.html         # Halaman awal / Landing page
│   ├── login.html        # Halaman masuk akun
│   ├── register.html     # Halaman pendaftaran akun
│   ├── dashboard.html    # Halaman ringkasan utama & analisa bulanan
│   ├── history.html      # Halaman riwayat transaksi lengkap
│   ├── transaction_add.html # Form input transaksi baru
│   └── settings.html     # Halaman modifikasi profil & password
├── go.mod                # File dependency manager Go
└── go.sum                # Checksum dependency Go
