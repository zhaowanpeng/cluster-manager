package model

import (
	"time"
)

// Node 表示集群中的一个节点
type Node struct {
	ID          string    `gorm:"primaryKey"`
	IP          string    `gorm:"uniqueIndex"`
	Port        int       `gorm:"default:22"`
	User        string    `gorm:"default:root"`
	Password    string    `gorm:""`
	Group       string    `gorm:"index"`
	AddAt       time.Time `gorm:""`
	LastCheckAt time.Time `gorm:""`
	Usable      bool      `gorm:"default:false"`
	Description string    `gorm:""`
}

// TableName 指定表名
func (Node) TableName() string {
	return "nodes"
}
