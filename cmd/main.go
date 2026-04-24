package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"hospital/database"
	"hospital/handler"
	"hospital/pkg/tts"
)

var startTime = time.Now()

func main() {
	log.Println("[BOOT] ========== SERVER STARTING ==========")
	log.Printf("[BOOT] Go version: %s, OS: %s, Arch: %s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	log.Printf("[BOOT] NumCPU: %d, GOMAXPROCS: %d\n", runtime.NumCPU(), runtime.GOMAXPROCS(0))

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	log.Printf("[BOOT] Memory: Alloc=%dMB, Sys=%dMB\n", memStats.Alloc/1024/1024, memStats.Sys/1024/1024)

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=hospital port=5432 sslmode=disable"
	}
	// Mask password for log
	log.Printf("[BOOT] DB_DSN: %s...\n", dsn[:min(40, len(dsn))])

	log.Println("[BOOT] Connecting to database...")
	if err := database.Connect(dsn); err != nil {
		log.Fatal("[BOOT] FATAL - Khong the ket noi database:", err)
	}
	defer database.Close()
	log.Println("[BOOT] Database connected OK")

	log.Println("[BOOT] Running migrations...")
	if err := database.Migrate(); err != nil {
		log.Fatal("[BOOT] FATAL - Migrate that bai:", err)
	}
	log.Println("[BOOT] Migrations OK")

	log.Println("[BOOT] Seeding data...")
	if err := database.Seed(); err != nil {
		log.Println("[BOOT] Seed bi bo qua (du lieu da ton tai):", err)
	} else {
		log.Println("[BOOT] Seed OK")
	}

	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// Tạo thư mục uploads nếu chưa có
	os.MkdirAll("uploads", 0755)
	router.Static("/uploads", "./uploads")

	// Tạo voice files (TTS) và serve static
	go func() {
		if err := tts.GenerateAll("audio"); err != nil {
			log.Println("[TTS] WARNING:", err)
		}
	}()
	router.Static("/audio", "./audio")

	// Debug endpoint - khong can DB, khong can auth
	router.GET("/debug/ping", func(c *gin.Context) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		c.JSON(200, gin.H{
			"status":     "alive",
			"go_version": runtime.Version(),
			"goroutines": runtime.NumGoroutine(),
			"mem_alloc_mb": fmt.Sprintf("%.1f", float64(m.Alloc)/1024/1024),
			"mem_sys_mb":   fmt.Sprintf("%.1f", float64(m.Sys)/1024/1024),
			"uptime":     time.Since(startTime).String(),
		})
	})

	log.Println("[BOOT] Registering routes...")
	handler.RegisterRoutes(router, database.DB)
	log.Println("[BOOT] Routes registered OK")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Graceful shutdown
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		log.Println("Server dang chay tai port:", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server loi:", err)
		}
	}()

	// Chan SIGINT/SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited cleanly")
}
