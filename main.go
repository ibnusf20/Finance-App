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
	dsn := "link database anda"
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

			// =================================================================
			// LOGIKA BARU: Query Agregasi untuk Grafik Arus Kas Bulanan
			// =================================================================
			var reports []MonthlyReport

			// Raw SQL Query untuk mengelompokkan income & expense berdasarkan bulan menggunakan GORM
			err := DB.Raw(`
				SELECT 
					TO_CHAR(date, 'TMMonth') as month_name,
					EXTRACT(YEAR FROM date)::int as year,
					COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0)::bigint as income,
					COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0)::bigint as expense,
					(COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0) - COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0))::bigint as net
				FROM transactions
				WHERE user_id = ?
				GROUP BY EXTRACT(YEAR FROM date), EXTRACT(MONTH FROM date), TO_CHAR(date, 'TMMonth')
				ORDER BY EXTRACT(YEAR FROM date) ASC, EXTRACT(MONTH FROM date) ASC
				LIMIT 6
			`, userID).Scan(&reports).Error

			if err != nil {
				fmt.Println("Gagal mengambil data laporan bulanan:", err)
			}
			// =================================================================

			// Mengirimkan data ke views/dashboard.html
			c.HTML(http.StatusOK, "dashboard.html", gin.H{
				"Username":       username,
				"Balance":        totalIncome - totalExpense,
				"TotalIncome":    totalIncome,
				"TotalExpense":   totalExpense,
				"DayIncome":      dayIncome,
				"DayExpense":     dayExpense,
				"MonthIncome":    monthIncome,
				"MonthExpense":   monthExpense,
				"MonthlyReports": reports, // <--- Sekarang data ini sudah dikirim ke Chart.js!
			})
		})

		// DI SINI TEMPAT MENEMPELKAN ROUTE BARU NYA:

		// RUTE BARU 1: Menampilkan halaman form input (transaction_add.html)
		dashboard.GET("/transaction/add", func(c *gin.Context) {
			session := sessions.Default(c)
			username := session.Get("username").(string)

			c.HTML(http.StatusOK, "transaction_add.html", gin.H{
				"Username": username,
			})
		})

		// RUTE BARU 2: Menampilkan halaman riwayat penuh (history.html)
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

		// Rute Proses Submit Form (Tetap sama, diarahkan kembali ke halaman history setelah simpan)
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

			// Setelah sukses simpan, langsung redirect ke halaman riwayat log
			c.Redirect(http.StatusFound, "/dashboard/history")
		})
	}

	r.Run("0.0.0.0:8080")
}
