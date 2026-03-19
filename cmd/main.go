package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"hospital/database"
	"hospital/handler"
)

func main() {
	// Lay duong dan file SQLite tu bien moi truong.
	// Mac dinh la "hospital.db" neu khong set.
	dsn := os.Getenv("DB_PATH")
	if dsn == "" {
		dsn = "hospital.db"
	}

	// Ket noi database
	if err := database.Connect(dsn); err != nil {
		log.Fatal("Khong the ket noi database:", err)
	}
	defer database.Close()

	// Chay auto-migrate tao/cap nhat cac bang
	if err := database.Migrate(); err != nil {
		log.Fatal("Migrate that bai:", err)
	}

	// Seed du lieu demo neu database dang rong
	if err := database.Seed(); err != nil {
		log.Println("Seed bi bo qua (du lieu da ton tai):", err)
	}

	// Cau hinh Gin mode (mac dinh debug, production dung APP_ENV=production)
	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Khoi tao router va dang ky tat ca routes
	router := gin.Default()
	handler.RegisterRoutes(router, database.DB)

	// Lay port tu bien moi truong, mac dinh 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server dang chay tai port:", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Khong the khoi dong server:", err)
	}
}
