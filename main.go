package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// --- MODEL DATABASE ---
type User struct {
	ID        uint   `gorm:"primaryKey"`
	Username  string `gorm:"unique;not null"`
	Email     string `gorm:"unique;not null"`
	Password  string `gorm:"not null"`
	CreatedAt time.Time
}

type Transaction struct {
	ID          uint      `gorm:"primaryKey"`
	UserID      uint      `gorm:"not null"`
	Type        string    `gorm:"type:varchar(10);not null"` // "income" atau "expense"
	Amount      float64   `gorm:"type:decimal(15,2);not null"`
	Category    string    `gorm:"type:varchar(50)"`
	Description string    `gorm:"type:text"`
	Date        time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

type MonthlyReport struct {
	MonthName string // "Januari", "Februari", dll
	Year      int
	Income    int64
	Expense   int64
	Net       int64 // Income - Expense
}

var DB *gorm.DB

// --- KONEKSI DATABASE ---
func connectDatabase() {
	dsn := "postgresql://postgres.jknwjhfwuhxgrnsciwks:lupapaswode@aws-1-ap-northeast-1.pooler.supabase.com:5432/postgres"
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println(err)
		log.Fatal("Gagal koneksi ke database:", err)
	}

	database.AutoMigrate(&User{}, &Transaction{})
	DB = database
}

// --- MIDDLEWARE AUTH ---
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		if userID == nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		c.Next()
	}
}

func formatRupiah(n float64) string {
	s := fmt.Sprintf("%.0f", n)
	if len(s) <= 3 {
		return s
	}
	var res []string
	for len(s) > 3 {
		res = append([]string{s[len(s)-3:]}, res...)
		s = s[:len(s)-3]
	}
	res = append([]string{s}, res...)
	return strings.Join(res, ".")
}

func main() {
	connectDatabase()

	r := gin.Default()
	r.SetFuncMap(template.FuncMap{
		"rupiah": formatRupiah,
	})

	store := cookie.NewStore([]byte("secret_key_sangat_rahasia"))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 24, // 1 Hari
		HttpOnly: true,
		Secure:   false,
	})
	r.Use(sessions.Sessions("finance_session", store))

	// Setup HTML Templates
	r.LoadHTMLGlob("views/*")

	// --- ROUTES ---

	// 1. Home
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home.html", nil)
	})

	// 2. Register
	r.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", nil)
	})
	r.POST("/register", func(c *gin.Context) {
		username := c.PostForm("username")
		email := c.PostForm("email")
		password := c.PostForm("password")

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		user := User{Username: username, Email: email, Password: string(hashedPassword)}
		if err := DB.Create(&user).Error; err != nil {
			c.HTML(http.StatusBadRequest, "register.html", gin.H{"error": "Email atau Username sudah terdaftar"})
			return
		}
		c.Redirect(http.StatusFound, "/login")
	})

	// 3. Login
	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})
	r.POST("/login", func(c *gin.Context) {
		email := c.PostForm("email")
		password := c.PostForm("password")

		var user User
		if err := DB.Where("email = ?", email).First(&user).Error; err != nil {
			c.HTML(http.StatusBadRequest, "login.html", gin.H{"error": "Email tidak ditemukan"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
			c.HTML(http.StatusBadRequest, "login.html", gin.H{"error": "Password salah"})
			return
		}

		session := sessions.Default(c)
		session.Set("user_id", user.ID)
		session.Set("username", user.Username)
		session.Save()

		c.Redirect(http.StatusFound, "/dashboard")
	})

	// 4. Logout
	r.GET("/logout", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()
		session.Save()
		c.Redirect(http.StatusFound, "/")
	})

	// 5. Dashboard (Protected)
	dashboard := r.Group("/dashboard")
	dashboard.Use(AuthRequired())
	{
		// Halaman Utama Ringkasan (dashboard.html)
		dashboard.GET("/", func(c *gin.Context) {
			session := sessions.Default(c)
			userID := session.Get("user_id").(uint)
			username := session.Get("username").(string)

			var transactions []Transaction
			now := time.Now()

			DB.Where("user_id = ?", userID).Find(&transactions)

			var totalIncome, totalExpense float64
			var dayIncome, dayExpense float64
			var monthIncome, monthExpense float64

			for _, t := range transactions {
				if t.Type == "income" {
					totalIncome += t.Amount
				} else {
					totalExpense += t.Amount
				}

				if t.Date.Format("2006-01-02") == now.Format("2006-01-02") {
					if t.Type == "income" {
						dayIncome += t.Amount
					} else {
						dayExpense += t.Amount
					}
				}

				if t.Date.Format("2006-01") == now.Format("2006-01") {
					if t.Type == "income" {
						monthIncome += t.Amount
					} else {
						monthExpense += t.Amount
					}
				}
			}

			// LOGIKA BARU: Query Agregasi untuk Grafik Arus Kas Bulanan
			var reports []MonthlyReport

			// Gunakan alias dengan tanda kutip dua "" agar pas dengan field Struct Go
			err := DB.Raw(`
    SELECT 
        TO_CHAR(date, 'TMMonth') as "MonthName",
        EXTRACT(YEAR FROM date)::int as "Year",
        COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0)::bigint as "Income",
        COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0)::bigint as "Expense",
        (COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0) - COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0))::bigint as "Net"
    FROM transactions
    WHERE user_id = ?
    GROUP BY EXTRACT(YEAR FROM date), EXTRACT(MONTH FROM date), TO_CHAR(date, 'TMMonth')
    ORDER BY EXTRACT(YEAR FROM date) ASC, EXTRACT(MONTH FROM date) ASC
    LIMIT 6
`, userID).Scan(&reports).Error

			if err != nil {
				fmt.Println("Gagal mengambil data laporan bulanan:", err)
			}

			c.HTML(http.StatusOK, "dashboard.html", gin.H{
				"Username":       username,
				"Balance":        totalIncome - totalExpense,
				"TotalIncome":    totalIncome,
				"TotalExpense":   totalExpense,
				"DayIncome":      dayIncome,
				"DayExpense":     dayExpense,
				"MonthIncome":    monthIncome,
				"MonthExpense":   monthExpense,
				"MonthlyReports": reports,
			})
		})

		// Menampilkan halaman form input (transaction_add.html)
		dashboard.GET("/transaction/add", func(c *gin.Context) {
			session := sessions.Default(c)
			username := session.Get("username").(string)

			c.HTML(http.StatusOK, "transaction_add.html", gin.H{
				"Username": username,
			})
		})

		// Menampilkan halaman riwayat penuh (history.html)
		dashboard.GET("/history", func(c *gin.Context) {
			session := sessions.Default(c)
			userID := session.Get("user_id").(uint)
			username := session.Get("username").(string)

			var transactions []Transaction
			DB.Where("user_id = ?", userID).Order("date desc").Find(&transactions)

			c.HTML(http.StatusOK, "history.html", gin.H{
				"Username":     username,
				"Transactions": transactions,
			})
		})

		// Rute Proses Submit Form
		dashboard.POST("/transaction", func(c *gin.Context) {
			session := sessions.Default(c)
			userID := session.Get("user_id").(uint)

			amountType := c.PostForm("type")
			amountStr := c.PostForm("amount")
			category := c.PostForm("category")
			description := c.PostForm("description")

			amountFloat, err := strconv.ParseFloat(amountStr, 64)
			if err != nil {
				amountFloat = 0
			}

			transaction := Transaction{
				UserID:      userID,
				Type:        amountType,
				Amount:      amountFloat,
				Category:    category,
				Description: description,
				Date:        time.Now(),
			}

			DB.Create(&transaction)

			c.Redirect(http.StatusFound, "/dashboard/history")
		})
		// =================================================================
		// RUTE BARU: EKSPOR DATA KE EXCEL
		// =================================================================
		dashboard.GET("/export/excel", func(c *gin.Context) {
			session := sessions.Default(c)
			userID := session.Get("user_id").(uint)
			username := session.Get("username").(string)

			// 1. Ambil data transaksi user dari database Postgres
			var transactions []Transaction
			if err := DB.Where("user_id = ?", userID).Order("date desc").Find(&transactions).Error; err != nil {
				c.String(http.StatusInternalServerError, "Gagal mengambil data transaksi")
				return
			}

			// 2. Buat file Excel baru menggunakan excelize
			f := excelize.NewFile()
			defer func() {
				if err := f.Close(); err != nil {
					fmt.Println(err)
				}
			}()

			// Ubah nama sheet default menjadi "Laporan Keuangan"
			sheetName := "Laporan Keuangan"
			f.SetSheetName("Sheet1", sheetName)

			// 3. Buat Style untuk Header (Warna Background Biru, Teks Putih & Tebal)
			headerStyle, err := f.NewStyle(&excelize.Style{
				Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
				Fill:      excelize.Fill{Type: "pattern", Color: []string{"0D6EFD"}, Pattern: 1},
				Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
			})
			if err != nil {
				fmt.Println(err)
			}

			// 4. Tulis Judul di Baris Atas
			f.SetCellValue(sheetName, "A1", "LAPORAN KEUANGAN PERSONAL")
			f.SetCellValue(sheetName, "A2", "Pengguna: "+username)
			f.SetCellValue(sheetName, "A3", "Tanggal Ekspor: "+time.Now().Format("02-01-2006 15:04"))

			// 5. Tulis Header Tabel di Baris ke-5
			headers := []string{"No", "Tanggal", "Tipe", "Kategori", "Jumlah (Rp)", "Deskripsi"}
			for i, header := range headers {
				cell, _ := excelize.CoordinatesToCellName(i+1, 5)
				f.SetCellValue(sheetName, cell, header)
				f.SetCellStyle(sheetName, cell, cell, headerStyle)
			}

			// 6. Isi Data Transaksi ke Baris-Baris Selanjutnya
			startRow := 6
			for idx, t := range transactions {
				currentRow := startRow + idx

				// Terjemahkan tipe ke bahasa Indonesia agar rapi
				tipeTransaksi := "Pemasukan"
				if t.Type == "expense" {
					tipeTransaksi = "Pengeluaran"
				}

				f.SetCellValue(sheetName, fmt.Sprintf("A%d", currentRow), idx+1)
				f.SetCellValue(sheetName, fmt.Sprintf("B%d", currentRow), t.Date.Format("02-01-2006 15:04"))
				f.SetCellValue(sheetName, fmt.Sprintf("C%d", currentRow), tipeTransaksi)
				f.SetCellValue(sheetName, fmt.Sprintf("D%d", currentRow), t.Category)
				f.SetCellValue(sheetName, fmt.Sprintf("E%d", currentRow), t.Amount) // Angka mentah agar bisa diformat di Excel
				f.SetCellValue(sheetName, fmt.Sprintf("F%d", currentRow), t.Description)
			}

			// 7. Atur Lebar Kolom Secara Otomatis Agar Tidak Terpotong
			f.SetColWidth(sheetName, "A", "A", 5)
			f.SetColWidth(sheetName, "B", "B", 20)
			f.SetColWidth(sheetName, "C", "C", 15)
			f.SetColWidth(sheetName, "D", "D", 15)
			f.SetColWidth(sheetName, "E", "E", 18)
			f.SetColWidth(sheetName, "F", "F", 30)

			// 8. Atur Header HTTP agar browser mengenali ini sebagai unduhan file Excel
			filename := fmt.Sprintf("Laporan_Keuangan_%s_%s.xlsx", username, time.Now().Format("20060102"))

			c.Header("Content-Disposition", "attachment; filename="+filename)
			c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
			c.Header("Content-Transfer-Encoding", "binary")
			c.Header("Cache-Control", "no-cache")

			// Write data langsung ke response body Gin (Stream ke Browser)
			if err := f.Write(c.Writer); err != nil {
				c.String(http.StatusInternalServerError, "Gagal menulis file Excel")
			}
		})

		// =================================================================
		// DI SINI ADALAH PERUBAHAN TAMBAHAN ROUTE PENGATURAN (SETTINGS)
		// =================================================================

		// 1. Menampilkan Halaman Pengaturan Akun
		dashboard.GET("/settings", func(c *gin.Context) {
			session := sessions.Default(c)
			userID := session.Get("user_id").(uint)

			var user User
			if err := DB.First(&user, userID).Error; err != nil {
				c.Redirect(http.StatusFound, "/login")
				return
			}

			c.HTML(http.StatusOK, "settings.html", gin.H{
				"Username": user.Username,
				"Email":    user.Email,
			})
		})

		// 2. Memproses Update Informasi Akun & Password
		dashboard.POST("/settings", func(c *gin.Context) {
			session := sessions.Default(c)
			userID := session.Get("user_id").(uint)

			usernameInput := c.PostForm("username")
			emailInput := c.PostForm("email")
			oldPassword := c.PostForm("old_password")
			newPassword := c.PostForm("new_password")

			var user User
			if err := DB.First(&user, userID).Error; err != nil {
				c.Redirect(http.StatusFound, "/login")
				return
			}

			// Validasi Password Lama wajib diisi jika berniat ganti password baru
			if oldPassword != "" || newPassword != "" {
				if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
					c.HTML(http.StatusBadRequest, "settings.html", gin.H{
						"Username": user.Username,
						"Email":    user.Email,
						"Error":    "Password lama yang dimasukkan tidak sesuai!",
					})
					return
				}

				if len(newPassword) < 6 {
					c.HTML(http.StatusBadRequest, "settings.html", gin.H{
						"Username": user.Username,
						"Email":    user.Email,
						"Error":    "Password baru minimal harus 6 karakter!",
					})
					return
				}

				// Enkripsi password baru
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
				user.Password = string(hashedPassword)
			}

			// Update field profil utama
			user.Username = usernameInput
			user.Email = emailInput

			// Simpan semua modifikasi ke database Postgres
			if err := DB.Save(&user).Error; err != nil {
				c.HTML(http.StatusBadRequest, "settings.html", gin.H{
					"Username": user.Username,
					"Email":    user.Email,
					"Error":    "Username atau Email sudah terpakai oleh pengguna lain!",
				})
				return
			}

			// Update session agar nama di pojok navbar ikut berubah seketika
			session.Set("username", user.Username)
			session.Save()

			c.HTML(http.StatusOK, "settings.html", gin.H{
				"Username": user.Username,
				"Email":    user.Email,
				"Success":  "Informasi akun Anda telah berhasil diperbarui!",
			})
		})
	}

	r.Run("0.0.0.0:8080")
}
