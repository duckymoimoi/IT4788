package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	response "hospital/pkg"
	"hospital/repository"
	"hospital/schema"

	"golang.org/x/crypto/bcrypt"
)

// ========================================
// STRUCT & CONSTRUCTOR
// ========================================

// AuthService xu ly logic nghiep vu cho 7 API xac thuc.
// Nhan request data tu handler, goi repository de truy van DB,
// xu ly logic (hash password, tao OTP, sinh JWT), tra ve ket qua hoac loi.
type AuthService struct {
	repo *repository.UserRepo
}

// NewAuthService khoi tao AuthService voi UserRepo da co san.
// Duoc goi 1 lan khi khoi dong server trong handler/routes.go.
func NewAuthService(repo *repository.UserRepo) *AuthService {
	return &AuthService{repo: repo}
}

// ========================================
// RETURN TYPES
// ========================================

// LoginResult chua thong tin tra ve sau khi dang nhap thanh cong.
// Cac truong tra ve khop voi dac ta API trong slide.
type LoginResult struct {
	UserID      uint64  `json:"user_id"`
	FullName    string  `json:"full_name"`
	PhoneNumber string  `json:"phone_number"`
	Token       string  `json:"token"`
	Avatar      *string `json:"avatar"`
	Active      int     `json:"active"` // 1: active, 0: inactive
	Role        string  `json:"role"`
}

// SignupResult chua thong tin tra ve sau khi dang ky.
// otp_code chi dung de debug trong MVP (khong gui SMS).
type SignupResult struct {
	UserID  uint64 `json:"user_id"`
	OTPCode string `json:"otp_code,omitempty"` // debug only
}

// ForgotResult chua OTP code debug cho forgot password.
// Slide chi tra code+message, otp_code la de debug trong MVP.
type ForgotResult struct {
	OTPCode string `json:"otp_code,omitempty"` // debug only
}

// ========================================
// CUSTOM ERRORS
// Handler dung errors.Is() de map sang response code tuong ung.
// ========================================

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrPasswordIncorrect = errors.New("password incorrect")
	ErrOTPIncorrect      = errors.New("otp incorrect")
	ErrOTPExpired        = errors.New("otp expired")
	ErrAccountBanned     = errors.New("account banned")
	ErrAccountNotActive  = errors.New("account not active")
)

// ========================================
// PRIVATE HELPERS
// ========================================

// generateOTPCode tao ma OTP 6 chu so ngau nhien.
// Trong MVP, ma nay tra ve truc tiep trong response de debug,
// khong gui SMS that.
func generateOTPCode() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

// determineRole xac dinh role cua user dua tren thong tin staff.
// Neu user co Staff (nhan vien) -> role = staff.Role (admin/coordinator/staff).
// Neu khong -> role = "patient".
func determineRole(user *schema.User) string {
	if user.Staff != nil {
		return string(user.Staff.Role)
	}
	return "patient"
}

// ========================================
// PUBLIC METHODS  - 7 HAM AUTH
// ========================================

// Login xac thuc nguoi dung bang so dien thoai va mat khau.
// Tra ve JWT token, thong tin ca nhan co ban theo dac ta slide.
// Neu co device_token va platform, tu dong luu FCM token.
//
// Flow:
//  1. Tim user theo phone (kem staff info)
//  2. Kiem tra trang thai tai khoan
//  3. So sanh password voi bcrypt
//  4. Xac dinh role va sinh JWT token
//  5. Luu FCM token neu co
func (s *AuthService) Login(phone, password, deviceToken, platform string) (*LoginResult, error) {
	// Tim user theo phone, preload Staff de biet role chi tiet
	user, err := s.repo.FindByPhoneWithStaff(phone)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Kiem tra trang thai tai khoan
	switch user.Status {
	case schema.UserStatusBanned:
		return nil, ErrAccountBanned
	case schema.UserStatusDeleted:
		return nil, ErrUserNotFound
	case schema.UserStatusPending:
		return nil, ErrAccountNotActive
	}

	// So sanh password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrPasswordIncorrect
	}

	// Xac dinh role va sinh token
	role := determineRole(user)
	token, err := response.GenerateToken(user.UserID, role)
	if err != nil {
		return nil, err
	}

	// Luu FCM token neu client gui kem
	if deviceToken != "" && platform != "" {
		fcm := &schema.FCMToken{
			UserID:         user.UserID,
			FCMToken:       deviceToken,
			DevicePlatform: schema.DevicePlatform(platform),
		}
		// Khong tra loi neu luu FCM that bai, chi log
		_ = s.repo.UpsertFCMToken(fcm)
	}

	// Tinh trang thai active (1 = active, 0 = khong active)
	active := 0
	if user.Status == schema.UserStatusActive {
		active = 1
	}

	return &LoginResult{
		UserID:      user.UserID,
		FullName:    user.FullName,
		PhoneNumber: user.PhoneNumber,
		Token:       token,
		Avatar:      user.AvatarURL,
		Active:      active,
		Role:        role,
	}, nil
}

// Signup dang ky tai khoan moi bang so dien thoai.
// Tao user voi status = "pending", sinh OTP de xac thuc.
// Trong MVP tra OTP code truc tiep (khong gui SMS).
//
// Input theo slide: phone_number, password, full_name, dob (date), gender (int)
//   gender: 0 = nu (F), 1 = nam (M)
//
// Flow:
//  1. Kiem tra trung so dien thoai
//  2. Hash password va tao user (kem dob, gender)
//  3. Tao OTP 6 so va luu vao DB
//  4. Tao setting mac dinh cho user
func (s *AuthService) Signup(phone, password, fullName, dob string, gender *int) (*SignupResult, error) {
	// Kiem tra so dien thoai da ton tai chua
	existing, err := s.repo.FindByPhone(phone)
	if err != nil {
		return nil, err
	}
	// Cho phep dang ky lai neu tai khoan cu da bi xoa (soft delete)
	if existing != nil && existing.Status != schema.UserStatusDeleted {
		return nil, ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Tao user moi voi status = pending (cho verify OTP)
	user := &schema.User{
		PhoneNumber:  phone,
		PasswordHash: string(hashedPassword),
		FullName:     fullName,
		UserType:     schema.UserTypePatient,
		Status:       schema.UserStatusPending,
	}

	// Parse dob neu co (format: "YYYY-MM-DD")
	if dob != "" {
		parsedDOB, err := time.Parse("2006-01-02", dob)
		if err == nil {
			user.DateOfBirth = &parsedDOB
		}
	}

	// Map gender: 0 = nu (F), 1 = nam (M)
	if gender != nil {
		var g schema.Gender
		switch *gender {
		case 0:
			g = schema.GenderFemale
		case 1:
			g = schema.GenderMale
		default:
			g = schema.GenderOther
		}
		user.Gender = &g
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	// Tao OTP 6 chu so
	otpCode := generateOTPCode()
	otp := &schema.OTPCode{
		PhoneNumber: phone,
		OTPCode:     otpCode,
		Type:        schema.OTPTypeSignup,
		ExpiredAt:   time.Now().Add(5 * time.Minute),
	}
	if err := s.repo.CreateOTP(otp); err != nil {
		return nil, err
	}

	// Tao setting mac dinh cho user moi
	setting := &schema.UserSetting{
		UserID: user.UserID,
	}
	if err := s.repo.CreateSetting(setting); err != nil {
		return nil, err
	}

	return &SignupResult{
		UserID:  user.UserID,
		OTPCode: otpCode,
	}, nil
}

// VerifyOTP xac thuc ma OTP da gui cho nguoi dung.
// Nhan vao otpType de tim dung loai OTP, tranh verify nham.
// Neu OTP la loai signup, chuyen user status tu pending -> active.
//
// Flow:
//  1. Tim OTP hop le theo phone + otpType (chua dung, chua het han)
//  2. So sanh ma OTP
//  3. Danh dau OTP da su dung
//  4. Cap nhat status user (neu la OTP signup)
func (s *AuthService) VerifyOTP(phone, code string, otpType schema.OTPType) error {
	// Tim OTP hop le theo dung loai
	otp, err := s.repo.FindValidOTP(phone, otpType)
	if err != nil {
		return err
	}
	if otp == nil {
		return ErrOTPExpired
	}

	// So sanh ma OTP
	if otp.OTPCode != code {
		return ErrOTPIncorrect
	}

	// Danh dau OTP da su dung
	if err := s.repo.MarkOTPUsed(otp.OTPID); err != nil {
		return err
	}

	// Neu la OTP signup, chuyen status tu pending -> active
	if otpType == schema.OTPTypeSignup {
		user, err := s.repo.FindByPhone(phone)
		if err != nil {
			return err
		}
		if user == nil {
			return ErrUserNotFound
		}
		if err := s.repo.UpdateStatus(user.UserID, schema.UserStatusActive); err != nil {
			return err
		}
	}

	return nil
}

// ResendOTP gui lai ma OTP cho so dien thoai.
// Slide: input = phone_number, output = code + message.
// Tao OTP moi loai signup va luu vao DB.
func (s *AuthService) ResendOTP(phone string) (*ForgotResult, error) {
	// Kiem tra user ton tai
	user, err := s.repo.FindByPhone(phone)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Tao OTP moi
	otpCode := generateOTPCode()
	otp := &schema.OTPCode{
		PhoneNumber: phone,
		OTPCode:     otpCode,
		Type:        schema.OTPTypeSignup,
		ExpiredAt:   time.Now().Add(5 * time.Minute),
	}
	if err := s.repo.CreateOTP(otp); err != nil {
		return nil, err
	}

	// Tra otp_code de debug (MVP khong gui SMS)
	return &ForgotResult{
		OTPCode: otpCode,
	}, nil
}

// Logout huy phien dang nhap va vo hieu hoa FCM token.
// Trong MVP khong co Redis blacklist token, chi deactivate FCM.
func (s *AuthService) Logout(userID uint64, fcmToken string) error {
	if fcmToken != "" {
		return s.repo.DeactivateFCMToken(fcmToken)
	}
	return nil
}

// ForgotPassword gui ma OTP de khoi phuc mat khau.
// Kiem tra user ton tai, tao OTP loai reset_password.
// Trong MVP tra OTP truc tiep (khong gui SMS).
func (s *AuthService) ForgotPassword(phone string) (*ForgotResult, error) {
	// Kiem tra user ton tai
	user, err := s.repo.FindByPhone(phone)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Tao OTP reset password
	otpCode := generateOTPCode()
	otp := &schema.OTPCode{
		PhoneNumber: phone,
		OTPCode:     otpCode,
		Type:        schema.OTPTypeResetPassword,
		ExpiredAt:   time.Now().Add(5 * time.Minute),
	}
	if err := s.repo.CreateOTP(otp); err != nil {
		return nil, err
	}

	return &ForgotResult{
		OTPCode: otpCode,
	}, nil
}

// ResetPassword dat lai mat khau moi sau khi xac thuc OTP.
// Verify OTP reset_password, hash password moi va cap nhat DB.
func (s *AuthService) ResetPassword(phone, code, newPassword string) error {
	// Tim OTP hop le
	otp, err := s.repo.FindValidOTP(phone, schema.OTPTypeResetPassword)
	if err != nil {
		return err
	}
	if otp == nil {
		return ErrOTPExpired
	}

	// So sanh ma OTP
	if otp.OTPCode != code {
		return ErrOTPIncorrect
	}

	// Danh dau OTP da su dung
	if err := s.repo.MarkOTPUsed(otp.OTPID); err != nil {
		return err
	}

	// Hash password moi
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Tim user de lay ID
	user, err := s.repo.FindByPhone(phone)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Cap nhat password
	return s.repo.UpdatePassword(user.UserID, string(hashedPassword))
}

// ChangePassword doi mat khau khi nguoi dung da dang nhap.
// Xac thuc mat khau cu, hash mat khau moi va cap nhat DB.
func (s *AuthService) ChangePassword(userID uint64, oldPass, newPass string) error {
	// Tim user theo ID
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Xac thuc mat khau cu
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPass)); err != nil {
		return ErrPasswordIncorrect
	}

	// Hash mat khau moi
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Cap nhat password moi
	return s.repo.UpdatePassword(userID, string(hashedPassword))
}
