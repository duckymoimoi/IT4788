package database

import (
	"log"
	"time"

	"hospital/schema"

	"golang.org/x/crypto/bcrypt"
)

// Seed tao du lieu mau cho demo.
// Chi chay khi database rong (kiem tra bang users truoc).
// Du lieu nay dung de test API va demo san pham.
func Seed() error {
	var count int64
	DB.Model(&schema.User{}).Count(&count)
	if count > 0 {
		log.Println("Database da co du lieu, bo qua seed")
		return nil
	}

	log.Println("Bat dau seed du lieu demo...")

	// FK da duoc tat trong Connect/Migrate, khong can tat/bat lai.
	// Circular dependency staffs <-> wards duoc xu ly o tang application.

	// --- BUOC 1: Tao cac khoa/vien (wards) ---
	// head_staff_id de NULL truoc, cap nhat sau khi co staff
	wards := []schema.Ward{
		{WardCode: "XN", WardName: "Khoa Xet Nghiem", IsActive: true},
		{WardCode: "CDHA", WardName: "Khoa Chan Doan Hinh Anh", IsActive: true},
		{WardCode: "NI", WardName: "Khoa Noi", IsActive: true},
		{WardCode: "NGOAI", WardName: "Khoa Ngoai", IsActive: true},
		{WardCode: "UTIL", WardName: "Tien Ich Benh Vien", IsActive: true},
	}

	if err := DB.Create(&wards).Error; err != nil {
		return err
	}
	log.Printf("Da tao %d wards", len(wards))

	// --- BUOC 2: Tao tai khoan nguoi dung ---
	// Dung password "password123" cho tat ca tai khoan test
	testPassword := hashPassword("password123")
	dob := time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC)
	genderMale := schema.GenderMale
	genderFemale := schema.GenderFemale

	users := []schema.User{
		// === 5 user goc (KHONG duoc thay doi - test phu thuoc) ===
		{
			PhoneNumber:  "0900000001",
			PasswordHash: testPassword,
			FullName:     "Nguyen Van Admin",
			UserType:     schema.UserTypeStaff,
			Gender:       &genderMale,
			Status:       schema.UserStatusActive,
		},
		{
			PhoneNumber:  "0900000002",
			PasswordHash: testPassword,
			FullName:     "Tran Thi Coordinator",
			UserType:     schema.UserTypeStaff,
			Gender:       &genderFemale,
			Status:       schema.UserStatusActive,
		},
		{
			PhoneNumber:  "0900000003",
			PasswordHash: testPassword,
			FullName:     "Le Van Staff",
			UserType:     schema.UserTypeStaff,
			Gender:       &genderMale,
			Status:       schema.UserStatusActive,
		},
		{
			PhoneNumber:  "0900000004",
			PasswordHash: testPassword,
			FullName:     "Pham Thi Benh Nhan",
			UserType:     schema.UserTypePatient,
			DateOfBirth:  &dob,
			Gender:       &genderFemale,
			Status:       schema.UserStatusActive,
		},
		{
			PhoneNumber:  "0900000005",
			PasswordHash: testPassword,
			FullName:     "Hoang Van Test",
			UserType:     schema.UserTypePatient,
			Gender:       &genderMale,
			Status:       schema.UserStatusActive,
		},
		// === Them 10 benh nhan ===
		{PhoneNumber: "0912000001", PasswordHash: testPassword, FullName: "Dao Minh Tuan", UserType: schema.UserTypePatient, DateOfBirth: dobPtr(1985, 3, 20), Gender: &genderMale, Status: schema.UserStatusActive},
		{PhoneNumber: "0912000002", PasswordHash: testPassword, FullName: "Vu Thi Lan", UserType: schema.UserTypePatient, DateOfBirth: dobPtr(1978, 7, 10), Gender: &genderFemale, Status: schema.UserStatusActive},
		{PhoneNumber: "0912000003", PasswordHash: testPassword, FullName: "Bui Duc Manh", UserType: schema.UserTypePatient, DateOfBirth: dobPtr(1960, 12, 5), Gender: &genderMale, Status: schema.UserStatusActive},
		{PhoneNumber: "0912000004", PasswordHash: testPassword, FullName: "Nguyen Thi Hoa", UserType: schema.UserTypePatient, DateOfBirth: dobPtr(1995, 1, 15), Gender: &genderFemale, Status: schema.UserStatusActive},
		{PhoneNumber: "0912000005", PasswordHash: testPassword, FullName: "Tran Quoc Viet", UserType: schema.UserTypePatient, DateOfBirth: dobPtr(1972, 8, 22), Gender: &genderMale, Status: schema.UserStatusActive},
		{PhoneNumber: "0912000006", PasswordHash: testPassword, FullName: "Le Thi Mai", UserType: schema.UserTypePatient, DateOfBirth: dobPtr(1988, 4, 8), Gender: &genderFemale, Status: schema.UserStatusActive},
		{PhoneNumber: "0912000007", PasswordHash: testPassword, FullName: "Phan Van Son", UserType: schema.UserTypePatient, DateOfBirth: dobPtr(1955, 11, 30), Gender: &genderMale, Status: schema.UserStatusActive},
		{PhoneNumber: "0912000008", PasswordHash: testPassword, FullName: "Ngo Thanh Thuy", UserType: schema.UserTypePatient, DateOfBirth: dobPtr(2000, 6, 18), Gender: &genderFemale, Status: schema.UserStatusActive},
		{PhoneNumber: "0912000009", PasswordHash: testPassword, FullName: "Vo Quang Hieu", UserType: schema.UserTypePatient, DateOfBirth: dobPtr(1968, 2, 25), Gender: &genderMale, Status: schema.UserStatusActive},
		{PhoneNumber: "0912000010", PasswordHash: testPassword, FullName: "Dang Thi Ngoc", UserType: schema.UserTypePatient, DateOfBirth: dobPtr(1992, 9, 3), Gender: &genderFemale, Status: schema.UserStatusActive},
		// === Them 5 nhan vien ===
		{PhoneNumber: "0900000006", PasswordHash: testPassword, FullName: "Trinh Van Bac Si", UserType: schema.UserTypeStaff, Gender: &genderMale, Status: schema.UserStatusActive},
		{PhoneNumber: "0900000007", PasswordHash: testPassword, FullName: "Luong Thi Y Ta", UserType: schema.UserTypeStaff, Gender: &genderFemale, Status: schema.UserStatusActive},
		{PhoneNumber: "0900000008", PasswordHash: testPassword, FullName: "Mai Van Ky Thuat", UserType: schema.UserTypeStaff, Gender: &genderMale, Status: schema.UserStatusActive},
		{PhoneNumber: "0900000009", PasswordHash: testPassword, FullName: "Ha Thi Le Tan", UserType: schema.UserTypeStaff, Gender: &genderFemale, Status: schema.UserStatusActive},
		{PhoneNumber: "0900000010", PasswordHash: testPassword, FullName: "Do Van Bao Ve", UserType: schema.UserTypeStaff, Gender: &genderMale, Status: schema.UserStatusActive},
	}

	if err := DB.Create(&users).Error; err != nil {
		return err
	}
	log.Printf("Da tao %d users", len(users))

	// --- BUOC 3: Tao user_settings cho tung user ---
	// Row nay tu dong tao khi signup, o day seed thu cong
	for _, u := range users {
		setting := schema.UserSetting{
			UserID:               u.UserID,
			VoiceGuidanceEnabled: true,
			NotificationEnabled:  true,
			TravelMode:           schema.TravelModeWalk,
			Language:             "vi",
		}
		if err := DB.Create(&setting).Error; err != nil {
			return err
		}
	}
	log.Println("Da tao user_settings cho tat ca users")

	// --- BUOC 4: Tao staffs ---
	// users[0] = admin, users[1] = coordinator, users[2] = staff
	// Ca truc luu dang chuoi "HH:MM", NULL = chua phan ca
	s1 := "07:00"
	e1 := "13:00"
	s2 := "07:00"
	e2 := "13:00"
	wardNI := uint32(wards[2].WardID) // Khoa Noi
	wardXN := uint32(wards[0].WardID) // Khoa Xet Nghiem

	staffs := []schema.Staff{
		{
			UserID:    users[0].UserID,
			StaffCode: "NV001",
			Role:      schema.StaffRoleAdmin,
			IsActive:  true,
			// Admin khong can ca truc cu the, de NULL
		},
		{
			UserID:     users[1].UserID,
			StaffCode:  "NV002",
			Role:       schema.StaffRoleCoordinator,
			WardID:     &wardNI,
			IsActive:   true,
			ShiftStart: &s1,
			ShiftEnd:   &e1,
		},
		{
			UserID:     users[2].UserID,
			StaffCode:  "NV003",
			Role:       schema.StaffRoleStaff,
			WardID:     &wardXN,
			IsActive:   true,
			ShiftStart: &s2,
			ShiftEnd:   &e2,
		},
	}

	if err := DB.Create(&staffs).Error; err != nil {
		return err
	}
	log.Printf("Da tao %d staffs", len(staffs))

	// --- BUOC 5: Cap nhat head_staff_id cho wards ---
	// Phai lam sau khi da co staff_id, tranh circular dependency khi INSERT
	DB.Model(&wards[2]).Update("head_staff_id", staffs[1].StaffID) // Khoa Noi -> coordinator
	DB.Model(&wards[0]).Update("head_staff_id", staffs[2].StaffID) // Khoa XN  -> staff
	log.Println("Da cap nhat head_staff_id cho wards")

	log.Println("Seed du lieu demo hoan thanh")
	log.Println("-------------------------------------------")
	log.Println("Tai khoan test (mat khau: password123):")
	log.Println("  Admin      : 0900000001")
	log.Println("  Coordinator: 0900000002")
	log.Println("  Staff      : 0900000003")
	log.Println("  Benh nhan  : 0900000004 / 0900000005")
	log.Println("-------------------------------------------")

	// --- Seed app versions ---
	versions := []schema.AppVersion{
		{
			Platform:      "android",
			VersionName:   "1.0.0",
			VersionCode:   1,
			IsForceUpdate: false,
			ChangeLog:     "Phien ban ra mat",
			DownloadURL:   "https://play.google.com/store/apps/details?id=com.hospital",
		},
		{
			Platform:      "ios",
			VersionName:   "1.0.0",
			VersionCode:   1,
			IsForceUpdate: false,
			ChangeLog:     "Phien ban ra mat",
			DownloadURL:   "https://apps.apple.com/app/hospital",
		},
	}
	if err := DB.Create(&versions).Error; err != nil {
		return err
	}
	log.Println("Da tao app_versions")

	// --- BUOC 7: Seed bản đồ grid 2D từ file .map ---
	if err := SeedMap(DB); err != nil {
		log.Printf("CANH BAO: Khong seed duoc map data: %v", err)
		// Không return error, vẫn tiếp tục vì map data là optional
	}

	// --- BUOC 8: Seed travel_modes ---
	SeedRoute(DB)

	// --- BUOC 9: Seed du lieu y te (Slice 6) ---
	// Goi ham seed cua Person C de tao treatments va queues
	SeedMedical(DB)

	// --- BUOC 10: Seed du lieu thiet bi (Slice 7) ---
	SeedDevices(DB)

	// --- BUOC 11: Seed du lieu tien ich (Slice 10) ---
	SeedUtils(DB)

	// --- BUOC 12: Seed du lieu flow (Slice 5) ---
	SeedFlow(DB)

	return nil
}

// hashPassword ma hoa mat khau bang BCrypt cost=12.
// Ham nay chi dung noi bo trong seed, khong export.
func hashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		log.Fatal("Loi khi hash password trong seed:", err)
	}
	return string(hash)
}

// dobPtr tao pointer time.Time cho DateOfBirth, tien dung trong seed.
func dobPtr(year, month, day int) *time.Time {
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return &t
}

