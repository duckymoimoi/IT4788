package handler

import (
	"hospital/middleware"
	response "hospital/pkg"
	"hospital/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UtilHandler struct {
	svc *service.UtilService
}

func NewUtilHandler(svc *service.UtilService) *UtilHandler {
	return &UtilHandler{svc: svc}
}

// ========================================
// REQUEST STRUCTS
// ========================================
type feedbackRequest struct {
	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
	Comment string `json:"comment"`
	Images  string `json:"images"` // JSON Array dang string
}

// ========================================
// API HANDLERS
// ========================================

// [77] GET /api/util/faq
func (h *UtilHandler) GetFAQ(c *gin.Context) {
	category := c.Query("category") // Co the loc theo /faq?category=general
	faqs, err := h.svc.GetFAQs(category)
	if err != nil {
		response.ErrInternalError(c)
		return
	}
	response.Success(c, faqs)
}

// [82] POST /api/util/feedback (Cần đăng nhập)
func (h *UtilHandler) SubmitFeedback(c *gin.Context) {
	var req feedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	userID := middleware.GetUserID(c)

	err := h.svc.SubmitFeedback(userID, req.Rating, req.Comment, req.Images)
	if err != nil {
		response.Error(c, 4000, err.Error())
		return
	}
	response.Success(c, map[string]string{"message": "Cảm ơn bạn đã đánh giá!"})
}

// [81] GET /api/util/feedback_summary
func (h *UtilHandler) GetFeedbackSummary(c *gin.Context) {
	summary, err := h.svc.GetFeedbackSummary()
	if err != nil {
		response.ErrInternalError(c)
		return
	}
	response.Success(c, summary)
}

// [95] GET /api/util/check_version?platform=android&code=10
func (h *UtilHandler) CheckVersion(c *gin.Context) {
	platform := c.Query("platform")
	codeStr := c.Query("code")

	if platform == "" || codeStr == "" {
		response.ErrMissingParam(c)
		return
	}

	clientCode, err := strconv.Atoi(codeStr)
	if err != nil {
		response.ErrInvalidValue(c)
		return
	}

	result, err := h.svc.CheckVersion(platform, clientCode)
	if err != nil {
		response.Error(c, 4004, err.Error())
		return
	}
	response.Success(c, result)
}

// [74] GET /api/util/languages
func (h *UtilHandler) GetLanguages(c *gin.Context) {
	response.Success(c, h.svc.GetLanguages())
}

// [78] GET /api/util/about
func (h *UtilHandler) GetAbout(c *gin.Context) {
	response.Success(c, h.svc.GetAboutInfo())
}

// [79] GET /api/util/contact
func (h *UtilHandler) GetContact(c *gin.Context) {
	response.Success(c, h.svc.GetContactInfo())
}