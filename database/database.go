package database

import (
	"fmt"
	"log"
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"hospital/schema"
)

// DB la bien global giu ket noi database.
// Duoc khoi tao 1 lan khi start server, dung chung cho toan bo ung dung.
var DB *gorm.DB

// Connect khoi tao ket noi den SQLite va chay auto-migrate.
// dsn la duong dan file SQLite, vi du: "hospital.db"
//
// De chuyen sang PostgreSQL ve sau, chi can thay 2 dong:
//
//	import "gorm.io/driver/postgres"
//	gorm.Open(postgres.Open(dsn), &gorm.Config{})
//
// Toan bo code query, schema, handler khong can thay doi gi.
func Connect(dsn string) error {
	// Cau hinh logger: hien thi SQL query ra console khi dev,
	// tat di khi production de tranh lo du lieu nhay cam.
	logLevel := logger.Info
	if os.Getenv("APP_ENV") == "production" {
		logLevel = logger.Error
	}

	gormLogger := logger.New(
		log.New(os.Stdout, "\n", log.LstdFlags),
		logger.Config{
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	var err error
	DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return fmt.Errorf("ket noi database that bai: %w", err)
	}

	// Bat foreign key constraint cho SQLite.
	// Mac dinh SQLite tat foreign key, can bat len de GORM enforce FK.
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("lay sql.DB that bai: %w", err)
	}
	if _, err := sqlDB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("bat foreign key that bai: %w", err)
	}

	// Gioi han connection pool cho SQLite.
	// SQLite khong ho tro concurrent write nen dat MaxOpenConns = 1
	// de tranh loi "database is locked".
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	log.Println("Ket noi database thanh cong:", dsn)
	return nil
}

// Migrate chay auto-migrate cho tat ca cac schema.
// Thu tu quan trong de tranh loi foreign key:
//  1. wards    (chua co FK phuc tap)
//  2. users    (chua co FK)
//  3. staffs   (FK -> users, FK -> wards)
//  4. otp_codes
//  5. user_settings (FK -> users)
//  6. fcm_tokens    (FK -> users)
//
// Luu y voi circular dependency giua staffs va wards:
// GORM se tao column head_staff_id trong wards nhung khong tao FK constraint
// tu dong vi de tranh loi. FK nay duoc quan ly o tang application.
func Migrate() error {
	// Tam tat foreign key khi migrate de tranh loi circular dependency
	// giua staffs va wards.
	if err := DB.Exec("PRAGMA foreign_keys = OFF").Error; err != nil {
		return fmt.Errorf("tat foreign key truoc migrate that bai: %w", err)
	}

	// Chay auto-migrate theo thu tu phu thuoc
	err := DB.AutoMigrate(
		&schema.Ward{},
		&schema.User{},
		&schema.Staff{},
		&schema.OTPCode{},
		&schema.UserSetting{},
		&schema.FCMToken{},
		&schema.AppVersion{},
	)
	if err != nil {
		return fmt.Errorf("auto-migrate that bai: %w", err)
	}

	// Bat lai foreign key sau khi migrate xong
	if err := DB.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		return fmt.Errorf("bat lai foreign key that bai: %w", err)
	}

	log.Println("Auto-migrate hoan thanh")
	return nil
}

// Close dong ket noi database, goi khi shutdown server.
func Close() {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err == nil {
			sqlDB.Close()
			log.Println("Dong ket noi database")
		}
	}
}
