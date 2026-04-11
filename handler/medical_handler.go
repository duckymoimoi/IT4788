package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"hospital/middleware"
	response "hospital/pkg"
	"hospital/service"
)

// MedicalHandler xu ly HTTP request cho module Y te (Slice 6).
type MedicalHandler struct {
	svc service.MedicalService
}

func NewMedicalHandler(svc service.MedicalService) *MedicalHandler {
	return &MedicalHandler{svc: svc}
}

// ========================================
// REQUEST STRUCTS
// ========================================

type checkinRequest struct {
	TreatmentID uint64 `json:"treatment_id" binding:"required"`
}

type prescriptionRequest struct {
	TreatmentID uint64 `json:"treatment_id" binding:"required"`
}

// ========================================
// SLICE 6 APIS - Medical
// ========================================

// [61] GET /api/medical/get_tasks
func (h *MedicalHandler) GetTasks(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.ErrNotAuthenticated(c)
		return
	}

	tasks, err := h.svc.GetMyTasks(c)
	if err != nil {
		response.Error(c, response.CodeHISUnavailable, "Cannot fetch tasks from HIS")
		return
	}

	response.Success(c, tasks)
}

// [62] GET /api/medical/get_queue?poi_id=
func (h *MedicalHandler) GetQueue(c *gin.Context) {
	poiID, _ := strconv.Atoi(c.Query("poi_id"))
	if poiID == 0 {
		response.ErrMissingParam(c)
		return
	}

	queue, err := h.svc.GetQueueStatus(uint32(poiID))
	if err != nil {
		response.ErrNotFound(c)
		return
	}

	response.Success(c, queue)
}

// [63] POST /api/medical/checkin_room
func (h *MedicalHandler) CheckinRoom(c *gin.Context) {
	var req checkinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	err := h.svc.CheckinRoom(c, req.TreatmentID)
	if err != nil {
		response.Error(c, response.CodeDBQueryFailed, err.Error())
		return
	}

	response.Success(c, gin.H{"checkin": true})
}

// [67] POST /api/medical/sync_now
func (h *MedicalHandler) SyncNow(c *gin.Context) {
	err := h.svc.SyncHIS(c)
	if err != nil {
		response.Error(c, response.CodeHISUnavailable, "Sync failed")
		return
	}

	response.Success(c, gin.H{"synced": true})
}

// [68] GET /api/medical/room_open?poi_id=
func (h *MedicalHandler) GetRoomOpen(c *gin.Context) {
	poiID, _ := strconv.Atoi(c.Query("poi_id"))
	if poiID == 0 {
		response.ErrMissingParam(c)
		return
	}

	hours, err := h.svc.GetRoomOpeningHours(uint32(poiID))
	if err != nil {
		response.ErrNotFound(c)
		return
	}

	response.Success(c, hours)
}

// [64] POST /api/medical/checkout_room
func (h *MedicalHandler) CheckoutRoom(c *gin.Context) {
	var req checkinRequest // reuse: cùng cần treatment_id
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	err := h.svc.CheckoutRoom(c, req.TreatmentID)
	if err != nil {
		response.Error(c, response.CodeDBQueryFailed, err.Error())
		return
	}

	response.Success(c, gin.H{"checkout": true})
}

// [65] GET /api/medical/result_status?treatment_id=
func (h *MedicalHandler) GetResultStatus(c *gin.Context) {
	tid, _ := strconv.ParseUint(c.Query("treatment_id"), 10, 64)
	if tid == 0 {
		response.ErrMissingParam(c)
		return
	}

	result, err := h.svc.GetResultStatus(c, tid)
	if err != nil {
		response.ErrNotFound(c)
		return
	}

	response.Success(c, result)
}

// [66] GET /api/medical/get_prescription
func (h *MedicalHandler) GetPrescription(c *gin.Context) {
	prescriptions, err := h.svc.GetPrescriptions(c)
	if err != nil {
		response.Error(c, response.CodeDBQueryFailed, err.Error())
		return
	}

	response.Success(c, prescriptions)
}

// [69] POST /api/medical/cancel_task
func (h *MedicalHandler) CancelTask(c *gin.Context) {
	var req checkinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	err := h.svc.CancelTask(c, req.TreatmentID)
	if err != nil {
		response.Error(c, response.CodeDBQueryFailed, err.Error())
		return
	}

	response.Success(c, gin.H{"cancelled": true})
}

// [70] GET /api/medical/get_history
func (h *MedicalHandler) GetHistory(c *gin.Context) {
	treatments, err := h.svc.GetHistory(c)
	if err != nil {
		response.Error(c, response.CodeDBQueryFailed, err.Error())
		return
	}

	response.Success(c, treatments)
}