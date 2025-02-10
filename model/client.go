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
