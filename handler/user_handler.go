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

// ========================================
// REQUEST STRUCTS  - đúng theo spec API
// ========================================

// set_profile: full_name, dob, gender, avatar
type setProfileRequest struct {
	FullName *string `json:"full_name"`
	DOB      *string `json:"dob"`
	Gender   *int    `json:"gender"`
	Avatar   *string `json:"avatar"`
}

// set_devtoken: device_token, platform
type setDevTokenRequest struct {
	DeviceToken string `json:"device_token"`
	Platform    string `json:"platform"`
}

// set_settings: language, theme, notification
type setSettingsRequest struct {
	Language     *string `json:"language"`
	Theme        *string `json:"theme"`
	Notification *bool   `json:"notification"`
}

// delete_account: password
type deleteAccountRequest struct {
	Password string `json:"password"`
}

// ========================================
// HANDLER METHODS
// ========================================

// 1. GetProfile  - GET /api/user/get_profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.ErrNotAuthenticated(c)
		return
	}

	result, err := h.svc.GetProfile(userID)
	if err != nil {
		h.handleUserError(c, err)
		return
	}

	response.Success(c, result)
}

// 2. SetProfile  - POST /api/user/set_profile
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
		FullName: req.FullName,
		DOB:      req.DOB,
		Gender:   req.Gender,
		Avatar:   req.Avatar,
	}

	result, err := h.svc.SetProfile(userID, input)
	if err != nil {
		h.handleUserError(c, err)
		return
	}

	response.Success(c, result)
}

// 3. SetDevToken  - POST /api/user/set_devtoken
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

	if strings.TrimSpace(req.DeviceToken) == "" || strings.TrimSpace(req.Platform) == "" {
		response.ErrMissingParam(c)
		return
	}

	if err := h.svc.SetDevToken(userID, req.DeviceToken, req.Platform); err != nil {
		h.handleUserError(c, err)
		return
	}

	response.Success(c, nil)
}

// 4. GetSettings  - GET /api/user/get_settings
func (h *UserHandler) GetSettings(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.ErrNotAuthenticated(c)
		return
	}

	result, err := h.svc.GetSettings(userID)
	if err != nil {
		h.handleUserError(c, err)
		return
	}

	response.Success(c, result)
}

// 5. SetSettings  - POST /api/user/set_settings
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
		Language:     req.Language,
		Theme:        req.Theme,
		Notification: req.Notification,
	}

	if err := h.svc.SetSettings(userID, input); err != nil {
		h.handleUserError(c, err)
		return
	}

	response.Success(c, nil)
}

// 6. DeleteAccount  - DELETE /api/user/delete_account
func (h *UserHandler) DeleteAccount(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.ErrNotAuthenticated(c)
		return
	}

	var req deleteAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	if strings.TrimSpace(req.Password) == "" {
		response.ErrMissingParam(c)
		return
	}

	result, err := h.svc.DeleteAccount(userID, req.Password)
	if err != nil {
		h.handleUserError(c, err)
		return
	}

	response.Success(c, result)
}

// ========================================
// ERROR HANDLER
// ========================================

func (h *UserHandler) handleUserError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrUserNotFound):
		response.ErrUserNotFound(c)
	case errors.Is(err, service.ErrPasswordIncorrect):
		response.ErrPasswordIncorrect(c)
	case errors.Is(err, service.ErrInvalidProfileInput),
		errors.Is(err, service.ErrInvalidSettingInput),
		errors.Is(err, service.ErrInvalidPlatform):
		response.Error(c, response.CodeInvalidValue, err.Error())
	default:
		response.ErrUnexpected(c)
	}
}
