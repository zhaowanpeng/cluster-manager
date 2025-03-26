package model

import (
	"time"
)

// Group 表示节点分组
type Group struct {
	Name        string    `gorm:"primaryKey"`
	Description string    `gorm:""`
	CreatedAt   time.Time `gorm:""`
	UpdatedAt   time.Time `gorm:""`
	User        string    `gorm:""`
	Tmp         bool      `gorm:"default:false"`
}

// TableName 指定表名
func (Group) TableName() string {
	return "groups"
}
