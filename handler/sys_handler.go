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
