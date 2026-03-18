package handler

import (
	"github.com/gin-gonic/gin"
	"errors"
	"hospital/middleware"
	response "hospital/pkg"
	"hospital/service"
	"hospital/schema"
)

// ==========================================================
// 1. CÁC KHUÔN HỨNG DỮ LIỆU TỪ CLIENT (REQUEST BODY)
// ==========================================================

type LoginRequest struct {
	PhoneNumber       string `json:"phone_number" binding:"required"`
	Password    string `json:"password" binding:"required"`
	DeviceToken string `json:"device_token"` 
	Platform    string `json:"platform"`     
}

type SignupRequest struct {
	PhoneNumber    string `json:"phone_number" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	DOB      string `json:"dob"`    // Ngày sinh (định dạng YYYY-MM-DD), không bắt buộc
	Gender   *int   `json:"gender"` // 0: Nữ, 1: Nam.
}

type VerifyOTPRequest struct {
	PhoneNumber   string         `json:"phone_number" binding:"required"`
	OTP    string         `json:"otp" binding:"required"`
	OTPType schema.OTPType `json:"otp_type" ` 
}
type ResendOTPRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
}
type ForgotPasswordRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
}

type ResetPasswordRequest struct {
	PhoneNumber       string `json:"phone_number" binding:"required"`
	OTP        string `json:"otp" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type LogoutRequest struct {
	FCMToken string `json:"fcm_token"` // Không bắt buộc
}

// AuthHandler chứa AuthService (Đầu bếp) bên trong
type AuthHandler struct {
	svc *service.AuthService
}

// Hàm tạo Handler mới
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

	// Gọi Service 
	result, err := h.svc.Login(req.PhoneNumber, req.Password, req.DeviceToken, req.Platform)
	if err != nil {
		response.Error(c, 401, err.Error())
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
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, result)
}

// POST /api/auth/verify-otp
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	// Xử lý thông dịch: Nếu FE không gửi otp_type, mặc định là Signup
	otpType := req.OTPType
	if otpType == "" {
		otpType = schema.OTPTypeSignup
	}

	err := h.svc.VerifyOTP(req.PhoneNumber, req.OTP, otpType)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, nil)
}
// POST /api/auth/forgot-password
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	result, err := h.svc.ForgotPassword(req.PhoneNumber)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, result)
}

// POST /api/auth/reset-password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	// Gọi service để đổi pass mới sau khi check OTP reset_password
	err := h.svc.ResetPassword(req.PhoneNumber, req.OTP, req.NewPassword)
	if err != nil {
		response.Error(c, 400, err.Error())
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
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, nil)
}

// POST /api/auth/change-password
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
		if errors.Is(err, service.ErrPasswordIncorrect) {
			response.ErrPasswordIncorrect(c)
			return
		}
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, nil)
}