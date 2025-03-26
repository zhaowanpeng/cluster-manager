package model

import (
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 是全局数据库连接
var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB() error {
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("无法获取用户主目录: %v", err)
	}

	// 创建应用数据目录
	appDir := filepath.Join(homeDir, ".cluster-manager")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return fmt.Errorf("无法创建应用数据目录: %v", err)
	}

	// 数据库文件路径
	dbPath := filepath.Join(appDir, "cluster.db")

	// 连接数据库
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("无法连接数据库: %v", err)
	}

	// 自动迁移表结构
	err = db.AutoMigrate(&Node{}, &Group{}, &Session{}, &Command{}, &CommandOutput{})
	if err != nil {
		return fmt.Errorf("自动迁移表结构失败: %v", err)
	}

	DB = db
	return nil
}
