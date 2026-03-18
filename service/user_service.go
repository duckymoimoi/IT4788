package service

import (
	"errors"
	"strings"
	"time"

	"hospital/repository"
	"hospital/schema"
)

// Các lỗi nghiệp vụ cho UserService
var (
	ErrInvalidProfileInput = errors.New("invalid profile input")
	ErrInvalidSettingInput = errors.New("invalid setting input")
	ErrInvalidPlatform     = errors.New("invalid platform")
	ErrVersionNotFound     = errors.New("version not found")
)

// UserService xử lý logic nghiệp vụ cho 6 API user + 1 API sys
type UserService struct {
	repo *repository.UserRepo
}

func NewUserService(repo *repository.UserRepo) *UserService {
	return &UserService{repo: repo}
}

// UpdateProfileInput là input cho SetProfile
type UpdateProfileInput struct {
	FullName    *string `json:"full_name"`
	DateOfBirth *string `json:"date_of_birth"` // "YYYY-MM-DD"
	Gender      *int    `json:"gender"`       // 0: nữ, 1: nam, khác: other
	AvatarURL   *string `json:"avatar_url"`
}

// UpdateSettingInput là input cho SetSettings
type UpdateSettingInput struct {
	VoiceGuidanceEnabled *bool   `json:"voice_guidance_enabled"`
	NotificationEnabled  *bool   `json:"notification_enabled"`
	TravelMode           *string `json:"travel_mode"`
	Language             *string `json:"language"`
}

// VersionCheckResult là output cho CheckVersion
type VersionCheckResult struct {
	Platform          string `json:"platform"`
	CurrentVersion    int    `json:"current_version_code"`
	LatestVersionCode int    `json:"latest_version_code"`
	LatestVersionName string `json:"latest_version_name"`
	NeedUpdate        bool   `json:"need_update"`
	IsForceUpdate     bool   `json:"is_force_update"`
	DownloadURL       string `json:"download_url"`
	ChangeLog         string `json:"change_log"`
}

// GetProfile trả về thông tin user theo userID
func (s *UserService) GetProfile(userID uint64) (*schema.User, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// SetProfile cập nhật thông tin cá nhân user
func (s *UserService) SetProfile(userID uint64, data UpdateProfileInput) error {
	updates := make(map[string]interface{})

	if data.FullName != nil {
		name := strings.TrimSpace(*data.FullName)
		if name == "" {
			return ErrInvalidProfileInput
		}
		updates["full_name"] = name
	}

	if data.DateOfBirth != nil {
		dob := strings.TrimSpace(*data.DateOfBirth)
		if dob != "" {
			parsed, err := time.Parse("2006-01-02", dob)
			if err != nil {
				return ErrInvalidProfileInput
			}
			updates["date_of_birth"] = parsed
		}
	}

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

	if data.AvatarURL != nil {
		avatar := strings.TrimSpace(*data.AvatarURL)
		if avatar == "" {
			updates["avatar_url"] = nil
		} else {
			updates["avatar_url"] = avatar
		}
	}

	if len(updates) == 0 {
		return nil
	}

	user, err := s.repo.FindByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	return s.repo.UpdateProfile(userID, updates)
}

// SetDevToken lưu hoặc cập nhật FCM token cho user
func (s *UserService) SetDevToken(userID uint64, token, platform, model, version string) error {
	token = strings.TrimSpace(token)
	platform = strings.ToLower(strings.TrimSpace(platform))
	if token == "" || platform == "" {
		return ErrInvalidSettingInput
	}
	if platform != string(schema.DevicePlatformAndroid) && platform != string(schema.DevicePlatformIOS) {
		return ErrInvalidPlatform
	}

	payload := &schema.FCMToken{
		UserID:         userID,
		FCMToken:       token,
		DevicePlatform: schema.DevicePlatform(platform),
	}

	if m := strings.TrimSpace(model); m != "" {
		payload.DeviceModel = &m
	}
	if v := strings.TrimSpace(version); v != "" {
		payload.AppVersion = &v
	}

	return s.repo.UpsertFCMToken(payload)
}

// GetSettings trả về cấu hình cá nhân của user
func (s *UserService) GetSettings(userID uint64) (*schema.UserSetting, error) {
	setting, err := s.repo.FindSettingByUserID(userID)
	if err != nil {
		return nil, err
	}
	if setting != nil {
		return setting, nil
	}

	created := &schema.UserSetting{UserID: userID}
	if err := s.repo.CreateSetting(created); err != nil {
		return nil, err
	}
	return s.repo.FindSettingByUserID(userID)
}

// SetSettings cập nhật cấu hình cá nhân user
func (s *UserService) SetSettings(userID uint64, data UpdateSettingInput) error {
	updates := make(map[string]interface{})

	if data.VoiceGuidanceEnabled != nil {
		updates["voice_guidance_enabled"] = *data.VoiceGuidanceEnabled
	}
	if data.NotificationEnabled != nil {
		updates["notification_enabled"] = *data.NotificationEnabled
	}
	if data.TravelMode != nil {
		mode := strings.ToLower(strings.TrimSpace(*data.TravelMode))
		switch schema.TravelMode(mode) {
		case schema.TravelModeWalk, schema.TravelModeWheelchair, schema.TravelModeStretcher:
			updates["travel_mode"] = mode
		default:
			return ErrInvalidSettingInput
		}
	}
	if data.Language != nil {
		lang := strings.TrimSpace(*data.Language)
		if lang == "" {
			return ErrInvalidSettingInput
		}
		updates["language"] = lang
	}

	if len(updates) == 0 {
		return nil
	}

	return s.repo.UpdateSetting(userID, updates)
}

// DeleteAccount đánh dấu tài khoản đã xóa (soft delete)
func (s *UserService) DeleteAccount(userID uint64) error {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}
	return s.repo.DeleteSoft(userID)
}

// CheckVersion kiểm tra phiên bản app client
func (s *UserService) CheckVersion(platform string, versionCode int) (*VersionCheckResult, error) {
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
		Platform:          platform,
		CurrentVersion:    versionCode,
		LatestVersionCode: latest.VersionCode,
		LatestVersionName: latest.VersionName,
		NeedUpdate:        versionCode < latest.VersionCode,
		IsForceUpdate:     latest.IsForceUpdate,
		DownloadURL:       latest.DownloadURL,
		ChangeLog:         latest.ChangeLog,
	}, nil
}
