package handler

import (
	"errors"

	"github.com/gin-gonic/gin"

	"hospital/middleware"
	response "hospital/pkg"
	"hospital/schema"
	"hospital/service"
)

// ==========================================================
// REQUEST STRUCTS
// ==========================================================

type LoginRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	Password    string `json:"password" binding:"required"`
	DeviceToken string `json:"device_token"`
	Platform    string `json:"platform"`
}

type SignupRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	Password    string `json:"password" binding:"required,min=6"`
	FullName    string `json:"full_name" binding:"required"`
	DOB         string `json:"dob"`
	Gender      *int   `json:"gender"`
}

type VerifyOTPRequest struct {
	PhoneNumber string         `json:"phone_number" binding:"required"`
	OTP         string         `json:"otp" binding:"required"`
	OTPType     schema.OTPType `json:"otp_type"`
}

type ForgotPasswordRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
}

type ResetPasswordRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	OTP         string `json:"otp" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type LogoutRequest struct {
	FCMToken string `json:"fcm_token"`
}

// ==========================================================
// AUTH HANDLER
// ==========================================================

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	result, err := h.svc.Login(req.PhoneNumber, req.Password, req.DeviceToken, req.Platform)
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	response.Success(c, result)
}

// POST /api/auth/signup
func (h *AuthHandler) Signup(c *gin.Context) {
	var req SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	result, err := h.svc.Signup(req.PhoneNumber, req.Password, req.FullName, req.DOB, req.Gender)
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	response.Success(c, result)
}

// POST /api/auth/verify_otp
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	otpType := req.OTPType
	if otpType == "" {
		otpType = schema.OTPTypeSignup
	}

	err := h.svc.VerifyOTP(req.PhoneNumber, req.OTP, otpType)
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	response.Success(c, nil)
}

// POST /api/auth/forgot_password
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	result, err := h.svc.ForgotPassword(req.PhoneNumber)
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	response.Success(c, result)
}

// POST /api/auth/reset_password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	err := h.svc.ResetPassword(req.PhoneNumber, req.OTP, req.NewPassword)
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	response.Success(c, nil)
}

// POST /api/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.ErrNotAuthenticated(c)
		return
	}

	var req LogoutRequest
	_ = c.ShouldBindJSON(&req)

	err := h.svc.Logout(userID, req.FCMToken)
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	response.Success(c, nil)
}

// POST /api/auth/change_password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.ErrNotAuthenticated(c)
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	err := h.svc.ChangePassword(userID, req.OldPassword, req.NewPassword)
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	response.Success(c, nil)
}

// ==========================================================
// ERROR HANDLER — map lỗi service sang đúng response code
// ==========================================================

func (h *AuthHandler) handleAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrUserNotFound):
		response.ErrUserNotFound(c)
	case errors.Is(err, service.ErrPasswordIncorrect):
		response.ErrPasswordIncorrect(c)
	case errors.Is(err, service.ErrUserAlreadyExists):
		response.ErrUserAlreadyExists(c)
	case errors.Is(err, service.ErrOTPIncorrect):
		response.ErrOTPIncorrect(c)
	case errors.Is(err, service.ErrOTPExpired):
		response.ErrOTPExpired(c)
	case errors.Is(err, service.ErrAccountBanned),
		errors.Is(err, service.ErrAccountNotActive):
		response.ErrNotAuthenticated(c)
	default:
		response.ErrUnexpected(c)
	}
}