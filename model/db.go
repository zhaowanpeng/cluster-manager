package model

import (
	"log"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB   *gorm.DB
	once sync.Once
)

func InitDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open("data.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 禁用日志输出
	})

	if err != nil {
		panic("failed to connect database")
	}

	// 自动迁移模型
	err = DB.AutoMigrate(&ShellClient{})
	if err != nil {
		panic("failed to migrate database")
	}
}

// GetDB 返回数据库单例
func GetDBSingle() *gorm.DB {
	once.Do(func() {
		var err error
		DB, err = gorm.Open(sqlite.Open("shell_client.db"), &gorm.Config{})
		if err != nil {
			log.Fatalf("数据库连接失败: %v", err)
		}
		if err := DB.AutoMigrate(&ShellClient{}); err != nil {
			log.Fatalf("数据库迁移失败: %v", err)
		}
	})
	return DB
}
