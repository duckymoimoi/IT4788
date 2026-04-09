package handler

import (
    "hospital/middleware"
    "hospital/repository"
    "hospital/service"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

// RegisterNotifRoutes dang ky notification endpoints (Slice 8).
// Person C so huu file nay[cite: 26].
func RegisterNotifRoutes(api *gin.RouterGroup, db *gorm.DB) {
    // 1. Khoi tao Repository (Tang truy van DB) [cite: 150]
    notifRepo := repository.NewNotifRepository(db)
    
    // 2. Khoi tao Service (Tang logic nghiep vu) [cite: 150]
    notifSvc := service.NewNotifService(notifRepo)
    
    // 3. Khoi tao Handler (Tang xu ly HTTP) [cite: 150]
    notifH := NewNotifHandler(notifSvc)

    // 4. Nhom cac API Notification [cite: 70]
    notif := api.Group("/notification")
    
    // Tat ca API thong bao yeu cau phai dang nhap (Auth) [cite: 73, 248]
    notif.Use(middleware.AuthMiddleware())
    {
        notif.GET("/get_list", notifH.GetList)      // [71] Lay danh sach thong bao [cite: 30, 246]
        notif.POST("/set_read", notifH.SetRead)     // [72] Danh dau da doc [cite: 30, 246]
        notif.DELETE("/delete", notifH.Delete)      // [73] Xoa thong bao [cite: 30, 246]
    }
}