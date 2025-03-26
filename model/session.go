package model

import (
	"time"
)

// Session 表示一次命令会话
type Session struct {
	ID          string    `gorm:"primaryKey"`
	Name        string    `gorm:"index"`
	Description string    `gorm:""`
	StartTime   time.Time `gorm:""`
	EndTime     time.Time `gorm:""`
	User        string    `gorm:""`
	GroupName   string    `gorm:"index"`
	ParentID    string    `gorm:"index"` // 用于构建会话树
}

// TableName 指定表名
func (Session) TableName() string {
	return "sessions"
}

// Command 表示会话中执行的命令，
type Command struct {
	ID        string    `gorm:"primaryKey"`
	SessionID string    `gorm:"index"`
	Command   string    `gorm:""`
	ExecTime  time.Time `gorm:""`
	Duration  int64     `gorm:""` // 执行时长（毫秒）
	ExitCode  int       `gorm:""` // 退出码
}

// TableName 指定表名
func (Command) TableName() string {
	return "commands"
}

// CommandOutput 表示命令在特定节点上的输出
type CommandOutput struct {
	ID        string `gorm:"primaryKey"`
	CommandID string `gorm:"index"`
	NodeIP    string `gorm:"index"`
	Output    string `gorm:"type:text"`
	ExitCode  int    `gorm:""`
}

// TableName 指定表名
func (CommandOutput) TableName() string {
	return "command_outputs"
}
