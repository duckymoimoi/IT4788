package repository

import (
	"time"

	"hospital/schema"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UserRepo la struct truy van database cho User va cac bang lien quan.
// Chi chua cac ham CRUD thuan tuy, khong chua logic nghiep vu.
// Service layer se goi cac ham nay de lay/ghi du lieu.
type UserRepo struct {
	db *gorm.DB
}

// NewUserRepo khoi tao UserRepo voi ket noi database.
// Duoc goi 1 lan khi khoi dong server, truyen vao service.
func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

// ========================================
// USER CRUD
// ========================================

// FindByPhone tim user theo so dien thoai.
// Tra ve nil, nil neu khong tim thay (khong phai loi).
// Dung khi dang nhap va kiem tra trung so dien thoai luc dang ky.
func (r *UserRepo) FindByPhone(phone string) (*schema.User, error) {
	var user schema.User
	err := r.db.Where("phone_number = ?", phone).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByID tim user theo ID.
// Tra ve nil, nil neu khong tim thay.
// Dung khi lay profile, doi mat khau, xoa tai khoan.
func (r *UserRepo) FindByID(id uint64) (*schema.User, error) {
	var user schema.User
	err := r.db.Where("user_id = ?", id).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByIDWithStaff tim user theo ID va preload thong tin Staff.
// Can thiet khi login de biet role chi tiet cua nhan vien
// (admin/coordinator/staff) va tao JWT token chua role chinh xac.
func (r *UserRepo) FindByIDWithStaff(id uint64) (*schema.User, error) {
	var user schema.User
	err := r.db.Preload("Staff").Where("user_id = ?", id).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByPhoneWithStaff tim user theo phone va preload Staff.
// Dung khi login de lay role chi tiet.
func (r *UserRepo) FindByPhoneWithStaff(phone string) (*schema.User, error) {
	var user schema.User
	err := r.db.Preload("Staff").Where("phone_number = ?", phone).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// Create tao user moi trong database.
// GORM tu dong set UserID (autoIncrement) va CreatedAt.
// Dung Omit(clause.Associations) de chi insert row users,
// khong xu ly cac relation (Staff, Setting, FCMTokens)
// tranh loi FK constraint voi SQLite.
func (r *UserRepo) Create(user *schema.User) error {
	return r.db.Omit(clause.Associations).Create(user).Error
}

// UpdateStatus cap nhat trang thai tai khoan (active/banned/deleted).
// Dung khi admin ban user hoac user xoa tai khoan (soft delete).
func (r *UserRepo) UpdateStatus(id uint64, status schema.UserStatus) error {
	return r.db.Model(&schema.User{}).
		Where("user_id = ?", id).
		Update("status", status).Error
}

// UpdateProfile cap nhat thong tin ca nhan cua user.
// Nhan map[string]interface{} de chi cap nhat cac truong duoc gui len,
// khong ghi de cac truong khac. Vi du: {"full_name": "Tran B", "gender": "F"}
func (r *UserRepo) UpdateProfile(id uint64, data map[string]interface{}) error {
	return r.db.Model(&schema.User{}).
		Where("user_id = ?", id).
		Updates(data).Error
}

// UpdatePassword cap nhat password hash.
// Dung khi doi mat khau hoac reset password.
func (r *UserRepo) UpdatePassword(id uint64, passwordHash string) error {
	return r.db.Model(&schema.User{}).
		Where("user_id = ?", id).
		Update("password_hash", passwordHash).Error
}

// DeleteSoft danh dau tai khoan da xoa bang cach doi status = "deleted".
// KHONG xoa row that su de giu du lieu cho audit/thong ke.
// User sau khi bi soft delete se khong the dang nhap lai.
func (r *UserRepo) DeleteSoft(id uint64) error {
	return r.UpdateStatus(id, schema.UserStatusDeleted)
}

// ========================================
// OTP
// ========================================

// CreateOTP tao ma OTP moi.
// Server tao OTP va luu vao database, tra lai cho client qua response
// (trong MVP khong gui SMS).
func (r *UserRepo) CreateOTP(otp *schema.OTPCode) error {
	return r.db.Create(otp).Error
}

// FindValidOTP tim OTP hop le theo so dien thoai va loai OTP.
// OTP hop le: chua duoc su dung (is_used = false) va chua het han.
// Tra ve OTP moi nhat (ORDER BY otp_id DESC) de xu ly truong hop
// user gui nhieu OTP lien tiep.
func (r *UserRepo) FindValidOTP(phone string, otpType schema.OTPType) (*schema.OTPCode, error) {
	var otp schema.OTPCode
	err := r.db.Where("phone_number = ? AND type = ? AND is_used = ? AND expired_at > ?",
		phone, otpType, false, time.Now()).
		Order("otp_id DESC").
		First(&otp).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &otp, nil
}

// MarkOTPUsed danh dau OTP da su dung.
// Sau khi verify OTP thanh cong, set is_used = true de tranh replay attack.
func (r *UserRepo) MarkOTPUsed(otpID uint64) error {
	return r.db.Model(&schema.OTPCode{}).
		Where("otp_id = ?", otpID).
		Update("is_used", true).Error
}

// ========================================
// FCM TOKEN
// ========================================

// UpsertFCMToken tao hoac cap nhat FCM token.
// Neu token da ton tai (cung fcm_token string), cap nhat thong tin
// va set is_active = true. Neu chua co, tao row moi.
// Dung khi user set device token sau login.
func (r *UserRepo) UpsertFCMToken(token *schema.FCMToken) error {
	// Tim token da ton tai theo gia tri fcm_token
	var existing schema.FCMToken
	err := r.db.Where("fcm_token = ?", token.FCMToken).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Token chua ton tai, tao moi
		return r.db.Create(token).Error
	}
	if err != nil {
		return err
	}

	// Token da ton tai, cap nhat
	now := time.Now()
	return r.db.Model(&existing).Updates(map[string]interface{}{
		"user_id":         token.UserID,
		"device_platform": token.DevicePlatform,
		"device_model":    token.DeviceModel,
		"app_version":     token.AppVersion,
		"is_active":       true,
		"last_used_at":    &now,
	}).Error
}

// DeactivateFCMToken vo hieu hoa FCM token khi user logout.
// Khong xoa row ma chi set is_active = false de giu history.
func (r *UserRepo) DeactivateFCMToken(fcmToken string) error {
	return r.db.Model(&schema.FCMToken{}).
		Where("fcm_token = ?", fcmToken).
		Update("is_active", false).Error
}

// ========================================
// USER SETTINGS
// ========================================

// FindSettingByUserID lay cau hinh ca nhan cua user.
// Tra ve nil, nil neu chua co setting (truong hop hiem gap vi seed tu tao).
func (r *UserRepo) FindSettingByUserID(userID uint64) (*schema.UserSetting, error) {
	var setting schema.UserSetting
	err := r.db.Where("user_id = ?", userID).First(&setting).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &setting, nil
}

// CreateSetting tao setting mac dinh cho user moi.
// Duoc goi khi signup de dam bao moi user luon co 1 row setting.
func (r *UserRepo) CreateSetting(setting *schema.UserSetting) error {
	return r.db.Create(setting).Error
}

// UpdateSetting cap nhat cau hinh cua user.
// Chi cap nhat cac truong duoc gui len, khong ghi de cac truong khac.
func (r *UserRepo) UpdateSetting(userID uint64, data map[string]interface{}) error {
	return r.db.Model(&schema.UserSetting{}).
		Where("user_id = ?", userID).
		Updates(data).Error
}

// ========================================
// APP VERSION
// ========================================

// FindLatestVersion lay phien ban moi nhat theo platform (android/ios).
// Sap xep theo version_code giam dan, lay row dau tien.
// Dung cho API check_version de client biet co can update khong.
func (r *UserRepo) FindLatestVersion(platform string) (*schema.AppVersion, error) {
	var version schema.AppVersion
	err := r.db.Where("platform = ?", platform).
		Order("version_code DESC").
		First(&version).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &version, nil
}
