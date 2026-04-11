package handler

import (
	"github.com/gin-gonic/gin"

	response "hospital/pkg"
	"hospital/service"
)

type SysHandler struct {
	svc *service.UserService
}

func NewSysHandler(svc *service.UserService) *SysHandler {
	return &SysHandler{svc: svc}
}

// GET /api/sys/check_version
// Query params: platform=android|ios&app_version=1.0.0
func (h *SysHandler) CheckVersion(c *gin.Context) {
	platform := c.Query("platform")
	appVersion := c.Query("app_version")

	if platform == "" || appVersion == "" {
		response.ErrMissingParam(c)
		return
	}

	result, err := h.svc.CheckVersion(platform, appVersion)
	if err != nil {
		switch err {
		case service.ErrInvalidPlatform:
			response.Error(c, response.CodeInvalidValue, "Invalid platform, must be 'android' or 'ios'")
		case service.ErrVersionNotFound:
			response.ErrUnexpected(c)
		default:
			response.ErrUnexpected(c)
		}
		return
	}

	response.Success(c, result)
}

// [79] GET /api/sys/get_voice_key
func (h *SysHandler) GetVoiceKey(c *gin.Context) {
	response.Success(c, gin.H{
		"provider":  "google",
		"api_key":   "DEMO_KEY_FOR_DEVELOPMENT",
		"language":  "vi-VN",
		"enabled":   true,
	})
}

// [80] GET /api/sys/get_voice_files
func (h *SysHandler) GetVoiceFiles(c *gin.Context) {
	response.Success(c, gin.H{
		"language": "vi",
		"files": []gin.H{
			{"key": "turn_left", "url": "/audio/turn_left.mp3", "text": "Rẽ trái"},
			{"key": "turn_right", "url": "/audio/turn_right.mp3", "text": "Rẽ phải"},
			{"key": "go_straight", "url": "/audio/go_straight.mp3", "text": "Đi thẳng"},
			{"key": "arrived", "url": "/audio/arrived.mp3", "text": "Đã đến đích"},
			{"key": "elevator_up", "url": "/audio/elevator_up.mp3", "text": "Đi thang máy lên"},
			{"key": "elevator_down", "url": "/audio/elevator_down.mp3", "text": "Đi thang máy xuống"},
			{"key": "stairs_up", "url": "/audio/stairs_up.mp3", "text": "Đi cầu thang lên"},
			{"key": "stairs_down", "url": "/audio/stairs_down.mp3", "text": "Đi cầu thang xuống"},
		},
	})
}
