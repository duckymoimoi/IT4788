package handler

import (
	"encoding/json"
	"hospital/middleware"
	"hospital/repository"
	response "hospital/pkg"
	"hospital/schema"
	"hospital/service"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type UtilHandler struct {
	svc    *service.UtilService
	mapRepo *repository.MapRepo
}

func NewUtilHandler(svc *service.UtilService, mapRepo *repository.MapRepo) *UtilHandler {
	return &UtilHandler{svc: svc, mapRepo: mapRepo}
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

// [99] GET /api/util/pharmacy
func (h *UtilHandler) GetPharmacy(c *gin.Context) {
	pois, err := h.mapRepo.FindPOIsByType(schema.POITypePharmacy, 0)
	if err != nil {
		response.ErrInternalError(c)
		return
	}
	response.Success(c, pois)
}

// [100] GET /api/util/canteen
func (h *UtilHandler) GetCanteen(c *gin.Context) {
	pois, err := h.mapRepo.FindPOIsByType(schema.POITypeCanteen, 0)
	if err != nil {
		response.ErrInternalError(c)
		return
	}
	response.Success(c, pois)
}

// [101] GET /api/util/parking
func (h *UtilHandler) GetParking(c *gin.Context) {
	pois, err := h.mapRepo.FindPOIsByType(schema.POITypeParking, 0)
	if err != nil {
		response.ErrInternalError(c)
		return
	}
	response.Success(c, pois)
}

// [102] GET /api/util/wifi
func (h *UtilHandler) GetWifi(c *gin.Context) {
	pois, err := h.mapRepo.FindPOIsByType(schema.POITypeWifi, 0)
	if err != nil {
		response.ErrInternalError(c)
		return
	}
	response.Success(c, pois)
}

// [106] GET /api/util/weather
// Sử dụng wttr.in free API - không cần API key
func (h *UtilHandler) GetWeather(c *gin.Context) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://wttr.in/Hanoi?format=j1")
	if err != nil {
		response.Error(c, 5000, "Khong ket noi duoc toi weather API")
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var weather map[string]interface{}
	json.Unmarshal(body, &weather)

	// Trích xuất thông tin chính
	result := gin.H{"city": "Hanoi", "raw": weather}
	if cc, ok := weather["current_condition"].([]interface{}); ok && len(cc) > 0 {
		current := cc[0].(map[string]interface{})
		result = gin.H{
			"city":         "Hanoi",
			"temp_c":       current["temp_C"],
			"feels_like_c": current["FeelsLikeC"],
			"humidity":     current["humidity"],
			"description":  current["weatherDesc"],
			"wind_speed":   current["windspeedKmph"],
		}
	}
	response.Success(c, result)
}