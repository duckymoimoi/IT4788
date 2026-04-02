package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterNotifRoutes dang ky notification endpoints (Slice 8).
// Person C so huu file nay.
func RegisterNotifRoutes(api *gin.RouterGroup, db *gorm.DB) {
	// TODO: Person C implement
	// notifRepo := repository.NewNotifRepo(db)
	// notifSvc := service.NewNotifService(notifRepo)
	// notifH := NewNotifHandler(notifSvc)
	//
	// notif := api.Group("/notification")
	// notif.Use(middleware.Auth())
	// notif.GET("/get_list", notifH.GetList)       // [71]
	// notif.POST("/set_read", notifH.SetRead)      // [72]
	// notif.DELETE("/delete", notifH.Delete)        // [73]
}
