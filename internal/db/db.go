package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	// 替换 SQLite 驱动
	_ "modernc.org/sqlite"
)

var DB *sql.DB

// InitDB 初始化数据库连接
func InitDB() error {
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("无法获取用户主目录: %v", err)
	}

	// 创建应用数据目录
	dataDir := filepath.Join(homeDir, ".clush")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("创建数据目录失败: %v", err)
	}

	// 数据库文件路径
	dbPath := filepath.Join(dataDir, "clush.db")

	// 连接数据库
	// 注意这里将 "sqlite3" 改为 "sqlite"
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("无法链接数据库: %v", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("无法链接数据库: %v", err)
	}

	DB = db

	// 初始化表结构
	if err := initTables(); err != nil {
		DB.Close()
		return fmt.Errorf("初始化表结构失败: %v", err)
	}

	return nil
}

// 初始化表结构
func initTables() error {
	// 创建组表
	_, err := DB.Exec(`
	CREATE TABLE IF NOT EXISTS groups (
		id TEXT PRIMARY KEY,
		name TEXT UNIQUE NOT NULL,
		description TEXT
	)`)
	if err != nil {
		return err
	}

	// 创建节点表
	_, err = DB.Exec(`
	CREATE TABLE IF NOT EXISTS nodes (
		id TEXT PRIMARY KEY,
		ip TEXT NOT NULL,
		port INTEGER NOT NULL,
		user TEXT NOT NULL,
		password TEXT NOT NULL,
		group_name TEXT NOT NULL,
		UNIQUE(ip, group_name)
	)`)
	if err != nil {
		return err
	}

	return nil
}

// CloseDB 关闭数据库连接
func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}
