package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"hospital/schema"
)

// DB la bien global giu ket noi database.
// Duoc khoi tao 1 lan khi start server, dung chung cho toan bo ung dung.
var DB *gorm.DB

// Connect khoi tao ket noi den PostgreSQL va chay auto-migrate.
// dsn co dang: "host=localhost user=postgres password=secret dbname=hospital port=5432 sslmode=disable"
//
// Environment variable: DB_DSN hoac truyen truc tiep.
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
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		// Tat FK constraint khi AutoMigrate de tranh loi circular dependency
		// (users -> routes -> users). FK van duoc enforce luc runtime.
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return fmt.Errorf("ket noi database that bai: %w", err)
	}

	// Lay sql.DB de cau hinh connection pool.
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("lay sql.DB that bai: %w", err)
	}

	// PostgreSQL ho tro concurrent connections tot,
	// dat connection pool phu hop voi production.
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(10)

	log.Println("Ket noi PostgreSQL thanh cong")
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
func Migrate() error {
	// Chay auto-migrate theo thu tu phu thuoc
	err := DB.AutoMigrate(
		// --- Core ---
		&schema.Ward{},
		&schema.User{},
		&schema.Staff{},
		&schema.OTPCode{},
		&schema.UserSetting{},
		&schema.FCMToken{},
		&schema.AppVersion{},

		// --- Map module ---
		&schema.GridMap{},
		&schema.GridPOI{},

		// --- Route module ---
		&schema.TravelMode{},
		&schema.Route{},
		&schema.RoutePath{},
		&schema.RouteHistoryNode{},
		&schema.RouteShare{},
		&schema.RouteFeedback{},

		// --- Flow module ---
		&schema.UserPing{},
		&schema.ObstacleReport{},
		&schema.HeatmapSnapshot{},
		&schema.PriorityRoute{},

		// --- Simulation ---
		&schema.SimulationRun{},
		&schema.PatientAgent{},

		// --- Medical ---
		&schema.Treatment{},
		&schema.Prescription{},
		&schema.Queue{},

		// --- Device ---
		&schema.DeviceStation{},
		&schema.Device{},
		&schema.DeviceBooking{},
		&schema.DeviceBrokenReport{},

		// --- Support ---
		&schema.Notification{},
		&schema.SOSRequest{},
		&schema.Conversation{},
		&schema.Message{},

		// --- Util ---
		&schema.Feedback{},
		&schema.FAQ{},
	)
	if err != nil {
		return fmt.Errorf("auto-migrate that bai: %w", err)
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
