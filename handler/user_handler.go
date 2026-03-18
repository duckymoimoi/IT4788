package handler

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"

	"hospital/middleware"
	response "hospital/pkg"
	"hospital/service"
)

type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

type setProfileRequest struct {
	FullName    *string `json:"full_name"`
	DateOfBirth *string `json:"date_of_birth"`
	Gender      *int    `json:"gender"`
	AvatarURL   *string `json:"avatar_url"`
}

type setDevTokenRequest struct {
	Token    string `json:"token"`
	Platform string `json:"platform"`
	Model    string `json:"model"`
	Version  string `json:"version"`
}

type setSettingsRequest struct {
	VoiceGuidanceEnabled *bool   `json:"voice_guidance_enabled"`
	NotificationEnabled  *bool   `json:"notification_enabled"`
	TravelMode           *string `json:"travel_mode"`
	Language             *string `json:"language"`
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.ErrNotAuthenticated(c)
		return
	}

	user, err := h.svc.GetProfile(userID)
	if err != nil {
		h.handleUserError(c, err)
		return
	}

	response.Success(c, user)
}

func (h *UserHandler) SetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.ErrNotAuthenticated(c)
		return
	}

	var req setProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	input := service.UpdateProfileInput{
		FullName:    req.FullName,
		DateOfBirth: req.DateOfBirth,
		Gender:      req.Gender,
		AvatarURL:   req.AvatarURL,
	}

	if err := h.svc.SetProfile(userID, input); err != nil {
		h.handleUserError(c, err)
		return
	}

	response.Success(c, gin.H{"updated": true})
}

func (h *UserHandler) SetDevToken(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.ErrNotAuthenticated(c)
		return
	}

	var req setDevTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	if strings.TrimSpace(req.Token) == "" || strings.TrimSpace(req.Platform) == "" {
		response.ErrMissingParam(c)
		return
	}

	if err := h.svc.SetDevToken(userID, req.Token, req.Platform, req.Model, req.Version); err != nil {
		h.handleUserError(c, err)
		return
	}

	response.Success(c, gin.H{"saved": true})
}

func (h *UserHandler) GetSettings(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.ErrNotAuthenticated(c)
		return
	}

	setting, err := h.svc.GetSettings(userID)
	if err != nil {
		h.handleUserError(c, err)
		return
	}

	response.Success(c, setting)
}

func (h *UserHandler) SetSettings(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.ErrNotAuthenticated(c)
		return
	}

	var req setSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	input := service.UpdateSettingInput{
		VoiceGuidanceEnabled: req.VoiceGuidanceEnabled,
		NotificationEnabled:  req.NotificationEnabled,
		TravelMode:           req.TravelMode,
		Language:             req.Language,
	}

	if err := h.svc.SetSettings(userID, input); err != nil {
		h.handleUserError(c, err)
		return
	}

	response.Success(c, gin.H{"updated": true})
}

func (h *UserHandler) DeleteAccount(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.ErrNotAuthenticated(c)
		return
	}

	if err := h.svc.DeleteAccount(userID); err != nil {
		h.handleUserError(c, err)
		return
	}

	response.Success(c, gin.H{"deleted": true})
}

func (h *UserHandler) handleUserError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrUserNotFound):
		response.ErrUserNotFound(c)
	case errors.Is(err, service.ErrInvalidProfileInput), errors.Is(err, service.ErrInvalidSettingInput), errors.Is(err, service.ErrInvalidPlatform):
		response.Error(c, response.CodeInvalidValue, err.Error())
	default:
		response.ErrUnexpected(c)
	}
}
