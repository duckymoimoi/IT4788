package service

import (
	"errors"
	"strings"
	"time"

	"hospital/repository"
	"hospital/schema"

	"golang.org/x/crypto/bcrypt"
)

// Các lỗi nghiệp vụ cho UserService
var (
	ErrInvalidProfileInput = errors.New("invalid profile input")
	ErrInvalidSettingInput = errors.New("invalid setting input")
	ErrInvalidPlatform     = errors.New("invalid platform")
	ErrVersionNotFound     = errors.New("version not found")
)

// UserService xử lý logic nghiệp vụ cho 6 API user + 1 API sys.
// Nhận dữ liệu từ handler, gọi repository để truy vấn DB,
// xử lý logic, trả kết quả hoặc lỗi.
type UserService struct {
	repo *repository.UserRepo
}

func NewUserService(repo *repository.UserRepo) *UserService {
	return &UserService{repo: repo}
}

// ========================================
// RETURN TYPES — đúng theo spec API
// ========================================

// ProfileResult là output cho get_profile và set_profile.
// Trả đúng các trường theo spec: user_id, full_name, phone_number, dob, gender, avatar.
type ProfileResult struct {
	UserID      uint64  `json:"user_id"`
	FullName    string  `json:"full_name"`
	PhoneNumber string  `json:"phone_number"`
	DOB         *string `json:"dob"`    // "YYYY-MM-DD" hoặc null
	Gender      *int    `json:"gender"` // 0: nữ, 1: nam
	Avatar      *string `json:"avatar"`
}

// SettingsResult là output cho get_settings.
// Trả đúng các trường theo spec: language, theme, notification.
type SettingsResult struct {
	Language     string `json:"language"`
	Theme        string `json:"theme"`
	Notification bool   `json:"notification"`
}

// DeleteAccountResult là output cho delete_account.
// Trả id của tài khoản vừa bị xóa.
type DeleteAccountResult struct {
	ID uint64 `json:"id"`
}

// VersionCheckResult là output cho check_version.
// Trả latest_version, force_update, download_url.
type VersionCheckResult struct {
	LatestVersion string `json:"latest_version"`
	ForceUpdate   bool   `json:"force_update"`
	DownloadURL   string `json:"download_url"`
}

// ========================================
// INPUT TYPES
// ========================================

// UpdateProfileInput là input cho set_profile.
// Các trường đều là pointer để phân biệt "không gửi" vs "gửi rỗng".
type UpdateProfileInput struct {
	FullName *string `json:"full_name"`
	DOB      *string `json:"dob"`    // "YYYY-MM-DD"
	Gender   *int    `json:"gender"` // 0: nữ, 1: nam
	Avatar   *string `json:"avatar"` // URL ảnh (handler xử lý upload file, truyền URL xuống)
}

// UpdateSettingInput là input cho set_settings.
// Theo spec: language, theme, notification.
type UpdateSettingInput struct {
	Language     *string `json:"language"`
	Theme        *string `json:"theme"`
	Notification *bool   `json:"notification"`
}

// ========================================
// PRIVATE HELPER
// ========================================

// buildProfileResult chuyển đổi schema.User sang ProfileResult đúng spec.
// - user_id: chuyển uint64 → string
// - gender: chuyển "F"→0, "M"→1
// - dob: chuyển time.Time → "YYYY-MM-DD"
// - avatar: giữ nguyên *string
func (s *UserService) buildProfileResult(user *schema.User) *ProfileResult {
	result := &ProfileResult{
		UserID:      user.UserID,
		FullName:    user.FullName,
		PhoneNumber: user.PhoneNumber,
		Avatar:      user.AvatarURL,
	}

	// Chuyển DateOfBirth time.Time → string "YYYY-MM-DD"
	if user.DateOfBirth != nil {
		dob := user.DateOfBirth.Format("2006-01-02")
		result.DOB = &dob
	}

	// Chuyển Gender string → int (0: nữ, 1: nam)
	if user.Gender != nil {
		var g int
		switch *user.Gender {
		case schema.GenderFemale:
			g = 0
		case schema.GenderMale:
			g = 1
		default:
			g = 2 // other
		}
		result.Gender = &g
	}

	return result
}

// ========================================
// PUBLIC METHODS — 7 HÀM THEO SPEC
// ========================================

// 1. GetProfile lấy thông tin cá nhân user.
// Input: token (lấy userID từ middleware)
// Output: ProfileResult { user_id, full_name, phone_number, dob, gender, avatar }
func (s *UserService) GetProfile(userID uint64) (*ProfileResult, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return s.buildProfileResult(user), nil
}

// 2. SetProfile cập nhật thông tin cá nhân user.
// Input: token, full_name, dob, gender, avatar (file → URL do handler xử lý upload)
// Output: ProfileResult (thông tin user sau khi cập nhật)
func (s *UserService) SetProfile(userID uint64, data UpdateProfileInput) (*ProfileResult, error) {
	updates := make(map[string]interface{})

	if data.FullName != nil {
		name := strings.TrimSpace(*data.FullName)
		if name == "" {
			return nil, ErrInvalidProfileInput
		}
		updates["full_name"] = name
	}

	// Parse dob nếu có (format: "YYYY-MM-DD")
	if data.DOB != nil {
		dob := strings.TrimSpace(*data.DOB)
		if dob != "" {
			parsed, err := time.Parse("2006-01-02", dob)
			if err != nil {
				return nil, ErrInvalidProfileInput
			}
			updates["date_of_birth"] = parsed
		}
	}

	// Map gender: 0 = nữ (F), 1 = nam (M)
	if data.Gender != nil {
		var g schema.Gender
		switch *data.Gender {
		case 0:
			g = schema.GenderFemale
		case 1:
			g = schema.GenderMale
		default:
			g = schema.GenderOther
		}
		updates["gender"] = g
	}

	// Avatar URL (handler đã upload file và truyền URL xuống)
	if data.Avatar != nil {
		avatar := strings.TrimSpace(*data.Avatar)
		if avatar == "" {
			updates["avatar_url"] = nil
		} else {
			updates["avatar_url"] = avatar
		}
	}

	// Kiểm tra user tồn tại
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Cập nhật nếu có trường nào thay đổi
	if len(updates) > 0 {
		if err := s.repo.UpdateProfile(userID, updates); err != nil {
			return nil, err
		}
	}

	// Lấy lại user sau khi cập nhật để trả về đúng data mới nhất
	updated, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	return s.buildProfileResult(updated), nil
}

// 3. SetDevToken đăng ký nhận thông báo đẩy (FCM).
// Input: device_token, platform (android/ios)
// Output: chỉ code + message (không có data)
func (s *UserService) SetDevToken(userID uint64, deviceToken, platform string) error {
	deviceToken = strings.TrimSpace(deviceToken)
	platform = strings.ToLower(strings.TrimSpace(platform))
	if deviceToken == "" || platform == "" {
		return ErrInvalidSettingInput
	}
	if platform != string(schema.DevicePlatformAndroid) && platform != string(schema.DevicePlatformIOS) {
		return ErrInvalidPlatform
	}

	payload := &schema.FCMToken{
		UserID:         userID,
		FCMToken:       deviceToken,
		DevicePlatform: schema.DevicePlatform(platform),
	}

	return s.repo.UpsertFCMToken(payload)
}

// 4. GetSettings lấy cấu hình giao diện/ngôn ngữ.
// Input: token
// Output: SettingsResult { language, theme, notification }
func (s *UserService) GetSettings(userID uint64) (*SettingsResult, error) {
	setting, err := s.repo.FindSettingByUserID(userID)
	if err != nil {
		return nil, err
	}

	// Tự động tạo setting mặc định nếu chưa có
	if setting == nil {
		created := &schema.UserSetting{UserID: userID}
		if err := s.repo.CreateSetting(created); err != nil {
			return nil, err
		}
		setting, err = s.repo.FindSettingByUserID(userID)
		if err != nil {
			return nil, err
		}
	}

	return &SettingsResult{
		Language:     setting.Language,
		Theme:        setting.Theme,
		Notification: setting.NotificationEnabled,
	}, nil
}

// 5. SetSettings lưu cấu hình người dùng.
// Input: token, language, theme, notification
// Output: chỉ code + message (không có data)
func (s *UserService) SetSettings(userID uint64, data UpdateSettingInput) error {
	updates := make(map[string]interface{})

	if data.Language != nil {
		lang := strings.TrimSpace(*data.Language)
		if lang == "" {
			return ErrInvalidSettingInput
		}
		updates["language"] = lang
	}

	if data.Theme != nil {
		theme := strings.TrimSpace(*data.Theme)
		if theme == "" {
			return ErrInvalidSettingInput
		}
		updates["theme"] = theme
	}

	if data.Notification != nil {
		updates["notification_enabled"] = *data.Notification
	}

	if len(updates) == 0 {
		return nil
	}

	return s.repo.UpdateSetting(userID, updates)
}

// 6. DeleteAccount xóa tài khoản (soft delete) sau khi xác nhận mật khẩu.
// Input: token, password (xác nhận)
// Output: DeleteAccountResult { id }
func (s *UserService) DeleteAccount(userID uint64, password string) (*DeleteAccountResult, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Xác nhận mật khẩu trước khi xóa
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrPasswordIncorrect
	}

	if err := s.repo.DeleteSoft(userID); err != nil {
		return nil, err
	}

	return &DeleteAccountResult{
		ID: userID,
	}, nil
}

// 7. CheckVersion kiểm tra cập nhật ứng dụng.
// Input: app_version (string, version hiện tại), platform (android/ios)
// Output: VersionCheckResult { latest_version, force_update, download_url }
func (s *UserService) CheckVersion(platform, appVersion string) (*VersionCheckResult, error) {
	platform = strings.ToLower(strings.TrimSpace(platform))
	if platform != "android" && platform != "ios" {
		return nil, ErrInvalidPlatform
	}

	latest, err := s.repo.FindLatestVersion(platform)
	if err != nil {
		return nil, err
	}
	if latest == nil {
		return nil, ErrVersionNotFound
	}

	return &VersionCheckResult{
		LatestVersion: latest.VersionName,
		ForceUpdate:   latest.IsForceUpdate,
		DownloadURL:   latest.DownloadURL,
	}, nil
}
