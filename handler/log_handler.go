package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"hospital/middleware"
)

// GetRequestLogs trả về danh sách request/response log từ server.
// GET /api/admin/get_request_logs?limit=100
func GetRequestLogs(c *gin.Context) {
	logs := middleware.GetLogs()

	limit := 200
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	if len(logs) > limit {
		logs = logs[:limit]
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    1000,
		"message": "OK",
		"data": gin.H{
			"total": len(logs),
			"logs":  logs,
		},
	})
}

// ClearRequestLogs xóa toàn bộ log.
// POST /api/admin/clear_request_logs
func ClearRequestLogs(c *gin.Context) {
	middleware.ClearLogs()
	c.JSON(http.StatusOK, gin.H{
		"code":    1000,
		"message": "OK",
		"data":    nil,
	})
}
