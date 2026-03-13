package database

import (
	"log"
	"time"

	model "hospital/schema"

	"golang.org/x/crypto/bcrypt"
)

// Seed tao du lieu mau cho demo.
// Chi chay khi database rong (kiem tra bang users truoc).
// Du lieu nay dung de test API va demo san pham.
func Seed() error {
	var count int64
	DB.Model(&model.User{}).Count(&count)
	if count > 0 {
		log.Println("Database da co du lieu, bo qua seed")
		return nil
	}

	log.Println("Bat dau seed du lieu demo...")

	// Tat foreign key khi seed de xu ly circular dependency
	// giua staffs va wards (ward.head_staff_id -> staffs)
	DB.Exec("PRAGMA foreign_keys = OFF")
	defer DB.Exec("PRAGMA foreign_keys = ON")

	// --- BUOC 1: Tao cac khoa/vien (wards) ---
	// head_staff_id de NULL truoc, cap nhat sau khi co staff
	wards := []model.Ward{
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
	genderMale := model.GenderMale
	genderFemale := model.GenderFemale

	users := []model.User{
		{
			PhoneNumber:  "0900000001",
			PasswordHash: testPassword,
			FullName:     "Nguyen Van Admin",
			UserType:     model.UserTypeStaff,
			Gender:       &genderMale,
			Status:       model.UserStatusActive,
		},
		{
			PhoneNumber:  "0900000002",
			PasswordHash: testPassword,
			FullName:     "Tran Thi Coordinator",
			UserType:     model.UserTypeStaff,
			Gender:       &genderFemale,
			Status:       model.UserStatusActive,
		},
		{
			PhoneNumber:  "0900000003",
			PasswordHash: testPassword,
			FullName:     "Le Van Staff",
			UserType:     model.UserTypeStaff,
			Gender:       &genderMale,
			Status:       model.UserStatusActive,
		},
		{
			PhoneNumber:  "0900000004",
			PasswordHash: testPassword,
			FullName:     "Pham Thi Benh Nhan",
			UserType:     model.UserTypePatient,
			DateOfBirth:  &dob,
			Gender:       &genderFemale,
			Status:       model.UserStatusActive,
		},
		{
			PhoneNumber:  "0900000005",
			PasswordHash: testPassword,
			FullName:     "Hoang Van Test",
			UserType:     model.UserTypePatient,
			Gender:       &genderMale,
			Status:       model.UserStatusActive,
		},
	}

	if err := DB.Create(&users).Error; err != nil {
		return err
	}
	log.Printf("Da tao %d users", len(users))

	// --- BUOC 3: Tao user_settings cho tung user ---
	// Row nay tu dong tao khi signup, o day seed thu cong
	for _, u := range users {
		setting := model.UserSetting{
			UserID:               u.UserID,
			VoiceGuidanceEnabled: true,
			NotificationEnabled:  true,
			TravelMode:           model.TravelModeWalk,
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

	staffs := []model.Staff{
		{
			UserID:    users[0].UserID,
			StaffCode: "NV001",
			Role:      model.StaffRoleAdmin,
			IsActive:  true,
			// Admin khong can ca truc cu the, de NULL
		},
		{
			UserID:     users[1].UserID,
			StaffCode:  "NV002",
			Role:       model.StaffRoleCoordinator,
			WardID:     &wardNI,
			IsActive:   true,
			ShiftStart: &s1,
			ShiftEnd:   &e1,
		},
		{
			UserID:     users[2].UserID,
			StaffCode:  "NV003",
			Role:       model.StaffRoleStaff,
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
