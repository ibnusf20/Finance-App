# 💰 Dompet Mandiri — Mobile-First Financial Tracker

Dompet Mandiri adalah aplikasi pencatatan keuangan mandiri berbasis web yang dirancang khusus dengan pendekatan **Mobile-First**. Aplikasi ini mempermudah pengguna untuk melacak arus kas harian (pemasukan dan pengeluaran), memantau saldo bersih secara instan, serta mengekspor data riwayat transaksi langsung ke format spreadsheet (MS Excel).

---

## ✨ Fitur Utama

*   **📱 Desain Mobile-First & Modern**: Tampilan antarmuka bersih dan responsif menggunakan utilitas penuh dari Tailwind CSS.
*   **🔒 Autentikasi Keamanan Sesi**: Dilengkapi sistem pendaftaran (`register.php`) dan masuk (`login.php`) menggunakan enkripsi password satu arah standar industri (`password_hash` BCrypt).
*   **⚡ Pencatatan Transaksi Kilat**: Alur tambah transaksi (`transaction_add.php`) intuitif dengan pemisahan kategori dinamis berbasis JavaScript.
*   **📊 Ekspor Laporan Excel**: Unduh seluruh riwayat pembukuan kas dalam format berkas `.xls` siap pakai secara instan (`export_excel.php`).
*   **🧩 Proteksi Akses Global**: Gerbang validasi terpusat (`auth.php`) memastikan halaman internal tidak dapat diakses tanpa sesi login yang valid.

---

## 🛠️ Teknologi & Stack Utama

*   **Backend**: PHP (Native, Object-Oriented PDO)
*   **Database**: PostgreSQL
*   **Frontend**: Tailwind CSS (via CDN), Vanilla JavaScript, Google Fonts (Inter)
*   **Arsitektur**: Mobile-First Container (Max-width: `md`)

---

## 📂 Struktur File Proyek

```text
├── config.php             # Konfigurasi koneksi database PDO PostgreSQL
├── auth.php               # Script gerbang keamanan / proteksi sesi user
├── landing.php            # Halaman pengenalan utama (Welcome Screen)
├── login.php              # Halaman masuk log akun pengguna
├── register.php           # Halaman pembuatan akun pengguna baru
├── index.php              # Dashboard utama (Ringkasan saldo & grafik)
├── transaction_add.php    # Formulir pencatatan transaksi kas baru
├── export_excel.php       # Generator file unduhan laporan MS Excel
└── logout.php             # Penghancur sesi & pembersih cookie login
