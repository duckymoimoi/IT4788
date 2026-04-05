package service

import (
	"fmt"
	"time"

	"hospital/repository"
	"hospital/schema"
)

// SOSService xu ly logic nghiep vu cho module SOS.
// Person E so huu file nay.
type SOSService struct {
	repo *repository.SupportRepo
}

func NewSOSService(repo *repository.SupportRepo) *SOSService {
	return &SOSService{repo: repo}
}

// ========================================
// API #96 POST create_sos
// ========================================

// CreateSOS tao yeu cau cap cuu moi.
// Benh nhan bam SOS -> tao record voi status = "received".
func (s *SOSService) CreateSOS(userID uint64, gridLocation int, posX, posY float64, note string) (*schema.SOSRequest, error) {
	sos := &schema.SOSRequest{
		UserID:       userID,
		GridLocation: gridLocation,
		PosX:         posX,
		PosY:         posY,
		Note:         note,
		Status:       schema.SOSStatusReceived,
	}

	if err := s.repo.CreateSOS(sos); err != nil {
		return nil, fmt.Errorf("cannot create SOS request: %w", err)
	}

	return sos, nil
}

// ========================================
// API #97 GET get_sos_list
// ========================================

// GetSOSList lay danh sach SOS cho staff xem (pagination).
func (s *SOSService) GetSOSList(page, limit int) ([]schema.SOSRequest, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	return s.repo.FindSOSList(page, limit)
}

// ========================================
// API #98 POST respond_sos
// ========================================

// RespondSOS staff nhan xu ly SOS.
// Nhan userID tu handler, tu dong chuyen sang staffID.
// Chi cho phep khi status = "received" (chua ai nhan).
func (s *SOSService) RespondSOS(sosID uint64, userID uint64) error {
	sos, err := s.repo.FindSOSByID(sosID)
	if err != nil {
		return fmt.Errorf("SOS request not found")
	}

	if sos.Status != schema.SOSStatusReceived {
		return fmt.Errorf("SOS already assigned or resolved")
	}

	staff, err := s.repo.FindStaffByUserID(userID)
	if err != nil {
		return fmt.Errorf("staff record not found")
	}

	updates := map[string]interface{}{
		"status":            schema.SOSStatusAssigned,
		"assigned_staff_id": staff.StaffID,
	}
	return s.repo.UpdateSOS(sosID, updates)
}

// ========================================
// API #99 POST resolve_sos
// ========================================

// ResolveSOS dong vu viec SOS.
// Nhan userID tu handler, tu dong chuyen sang staffID.
// Chi staff da nhan (assigned_staff_id) moi duoc dong.
func (s *SOSService) ResolveSOS(sosID uint64, userID uint64) error {
	sos, err := s.repo.FindSOSByID(sosID)
	if err != nil {
		return fmt.Errorf("SOS request not found")
	}

	if sos.Status != schema.SOSStatusAssigned {
		return fmt.Errorf("SOS is not in assigned status")
	}

	staff, err := s.repo.FindStaffByUserID(userID)
	if err != nil {
		return fmt.Errorf("staff record not found")
	}

	if sos.AssignedStaff == nil || *sos.AssignedStaff != staff.StaffID {
		return fmt.Errorf("only the assigned staff can resolve this SOS")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":      schema.SOSStatusResolved,
		"resolved_at": &now,
	}
	return s.repo.UpdateSOS(sosID, updates)
}

// ========================================
// API #100 GET get_sos_detail
// ========================================

// GetSOSDetail lay chi tiet 1 SOS case.
func (s *SOSService) GetSOSDetail(sosID uint64) (*schema.SOSRequest, error) {
	sos, err := s.repo.FindSOSByID(sosID)
	if err != nil {
		return nil, fmt.Errorf("SOS request not found")
	}
	return sos, nil
}
