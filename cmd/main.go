package main

import (
	"log"
	"os"

	"hospital/database"
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
		log.Fatal("Seed that bai:", err)
	}

	log.Println("Database san sang. Server se duoc them vao day sau.")

}
