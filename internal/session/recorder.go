package session

import (
	"time"

	"zhaowanpeng/cluster-manager/model"
)

// Recorder 用于记录会话
type Recorder struct {
	session     *model.Session
	commands    []*model.Command
	currentCmd  *model.Command
	isRecording bool
}

// NewRecorder 创建新的会话记录器
func NewRecorder(name, description, user, groupName string) *Recorder {
	session := &model.Session{
		ID:          time.Now().Format("20060102150405"),
		Name:        name,
		Description: description,
		StartTime:   time.Now(),
		User:        user,
		GroupName:   groupName,
	}

	return &Recorder{
		session:     session,
		commands:    make([]*model.Command, 0),
		isRecording: false,
	}
}

// Start 开始记录会话
func (r *Recorder) Start() error {
	if r.isRecording {
		return nil
	}

	// 保存会话信息到数据库
	if err := model.DB.Create(r.session).Error; err != nil {
		return err
	}

	r.isRecording = true
	return nil
}

// RecordCommand 记录命令
func (r *Recorder) RecordCommand(cmdStr string) {
	if !r.isRecording {
		return
	}

	r.currentCmd = &model.Command{
		ID:        time.Now().Format("20060102150405.000"),
		SessionID: r.session.ID,
		Command:   cmdStr,
		ExecTime:  time.Now(),
	}

	// 保存命令到数据库
	model.DB.Create(r.currentCmd)
	r.commands = append(r.commands, r.currentCmd)
}

// RecordOutput 记录命令输出
func (r *Recorder) RecordOutput(nodeIP, output string, exitCode int) {
	if !r.isRecording || r.currentCmd == nil {
		return
	}

	cmdOutput := &model.CommandOutput{
		ID:        time.Now().Format("20060102150405.000"),
		CommandID: r.currentCmd.ID,
		NodeIP:    nodeIP,
		Output:    output,
		ExitCode:  exitCode,
	}

	// 保存命令输出到数据库
	model.DB.Create(cmdOutput)
}

// FinishCommand 完成命令记录
func (r *Recorder) FinishCommand(exitCode int, duration time.Duration) {
	if !r.isRecording || r.currentCmd == nil {
		return
	}

	r.currentCmd.ExitCode = exitCode
	r.currentCmd.Duration = duration.Milliseconds()

	// 更新命令信息
	model.DB.Save(r.currentCmd)
	r.currentCmd = nil
}

// Stop 停止记录会话
func (r *Recorder) Stop() {
	if !r.isRecording {
		return
	}

	r.session.EndTime = time.Now()
	model.DB.Save(r.session)
	r.isRecording = false
}
