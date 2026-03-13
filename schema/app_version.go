package schema

import "time"

// AppVersion luu thong tin phien ban ung dung.
// Dung de kiem tra client co dang dung phien ban moi nhat khong.
// client gui len platform + version_code, server so sanh voi
// version_code lon nhat trong bang de biet co can cap nhat khong.
// Bang: app_versions [T07]
type AppVersion struct {
	VersionID     uint64    `gorm:"primaryKey;autoIncrement;column:version_id"`
	Platform      string    `gorm:"not null;size:10;index;column:platform"`
	VersionName   string    `gorm:"not null;size:20;column:version_name"`
	VersionCode   int       `gorm:"not null;column:version_code"`
	IsForceUpdate bool      `gorm:"not null;default:false;column:is_force_update"`
	ChangeLog     string    `gorm:"size:1000;column:change_log"`
	DownloadURL   string    `gorm:"size:500;column:download_url"`
	CreatedAt     time.Time `gorm:"not null;autoCreateTime;column:created_at"`
}

func (AppVersion) TableName() string {
	return "app_versions"
}
