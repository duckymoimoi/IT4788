package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"hospital/database"
	"hospital/handler"
	"hospital/middleware"
	"hospital/repository"
	"hospital/service"
)

func main() {
	dsn := os.Getenv("DB_PATH")
	if dsn == "" {
		dsn = "hospital.db"
	}
	if err := database.Connect(dsn); err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	if err := database.Migrate(); err != nil {
		log.Fatal(err)
	}
	if err := database.Seed(); err != nil {
		log.Println("Seed skipped (data exists):", err)
	}

	// Khoi tao cac tang
	userRepo := repository.NewUserRepo(database.DB)
	authSvc := service.NewAuthService(userRepo)
	userSvc := service.NewUserService(userRepo)
	authH := handler.NewAuthHandler(authSvc)
	userH := handler.NewUserHandler(userSvc)
	sysH := handler.NewSysHandler(userSvc)

	r := gin.Default()
	api := r.Group("/api")

	// === AUTH — public ===
	auth := api.Group("/auth")
	auth.POST("/login", authH.Login)
	auth.POST("/signup", authH.Signup)
	auth.POST("/verify_otp", authH.VerifyOTP)
	auth.POST("/forgot_password", authH.ForgotPassword)
	auth.POST("/reset_password", authH.ResetPassword)

	// === AUTH — private ===
	authPriv := api.Group("/auth")
	authPriv.Use(middleware.Auth())
	authPriv.POST("/logout", authH.Logout)
	authPriv.POST("/change_password", authH.ChangePassword)

	// === USER — private ===
	user := api.Group("/user")
	user.Use(middleware.Auth())
	user.GET("/get_profile", userH.GetProfile)
	user.POST("/set_profile", userH.SetProfile)
	user.POST("/set_devtoken", userH.SetDevToken)
	user.GET("/get_settings", userH.GetSettings)
	user.POST("/set_settings", userH.SetSettings)
	user.DELETE("/delete_account", userH.DeleteAccount)

	// === SYS — public ===
	api.GET("/sys/check_version", sysH.CheckVersion)

	fmt.Println("================================================")
	fmt.Println("TEST SERVER — http://localhost:8080")
	fmt.Println("Chay xong thi xoa file nay (cmd/test_full/main.go)")
	fmt.Println("================================================")
	r.Run(":8080")
}
