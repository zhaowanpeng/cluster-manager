package model

import "time"

type ShellClient struct {
	ID          string    `gorm:"primaryKey;type:varchar(64)"`
	IP          string    `gorm:"type:varchar(64);not null"`
	Port        int       `gorm:"not null"`
	User        string    `gorm:"type:varchar(64);not null"`
	Password    string    `gorm:"type:varchar(255)"`
	Group       string    `gorm:"type:varchar(64);not null"`
	AddAt       time.Time `gorm:"not null"`
	Usable      bool      `gorm:"not null;default:false"`
	Tmp         bool      `gorm:"default:false"`
	Description string    `gorm:"type:varchar(10240);default:''"`
}

// type Group struct {
// 	Name        string    `gorm:"primaryKey;type:varchar(64);not null"` // 组名
// 	Description string    `gorm:"type:varchar(10240);default:''"`
// 	CreatedAt   time.Time `gorm:"type:timestamp;"` //默认创建时间
// 	UpdatedAt   time.Time `gorm:"type:timestamp;"`
// 	User        string    `gorm:"type:varchar(64);"`
// 	Tmp         bool      `gorm:"default:false"`
// }

// type Node struct {
// 	ID          string    `gorm:"primaryKey;type:varchar(64)"` // 组ID
// 	IP          string    `gorm:"type:varchar(64);not null"`
// 	Port        int       `gorm:"not null"`
// 	User        string    `gorm:"type:varchar(64);not null"`
// 	Password    string    `gorm:"type:varchar(255)"`
// 	GroupName   string    `gorm:"type:varchar(64);not null"`
// 	AddAt       time.Time `gorm:"not null"`
// 	Usable      bool      `gorm:"not null;default:false"`
// 	Tmp         bool      `gorm:"default:false"`
// 	Description string    `gorm:"type:varchar(10240);default:''"`
// }
