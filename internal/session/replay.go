package session

import (
	"fmt"

	"zhaowanpeng/cluster-manager/model"
)

// SessionTree 表示会话树结构
type SessionTree struct {
	Session  model.Session
	Children []*SessionTree
}

// GetSessionTree 获取会话树
func GetSessionTree(rootID string) (*SessionTree, error) {
	var root model.Session
	if err := model.DB.First(&root, "id = ?", rootID).Error; err != nil {
		return nil, err
	}

	tree := &SessionTree{
		Session:  root,
		Children: make([]*SessionTree, 0),
	}

	// 递归构建树
	buildSessionTree(tree)
	return tree, nil
}

// buildSessionTree 递归构建会话树
func buildSessionTree(node *SessionTree) {
	var children []model.Session
	model.DB.Find(&children, "parent_id = ?", node.Session.ID)

	for _, child := range children {
		childNode := &SessionTree{
			Session:  child,
			Children: make([]*SessionTree, 0),
		}
		node.Children = append(node.Children, childNode)
		buildSessionTree(childNode)
	}
}

// ReplaySession 重放会话
func ReplaySession(sessionID string) error {
	var session model.Session
	if err := model.DB.First(&session, "id = ?", sessionID).Error; err != nil {
		return fmt.Errorf("找不到会话: %v", err)
	}

	fmt.Printf("重放会话: %s (%s)\n", session.Name, session.Description)
	fmt.Printf("开始时间: %s\n", session.StartTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("结束时间: %s\n", session.EndTime.Format("2006-01-02 15:04:05"))
	fmt.Println("-----------------------------------")

	var commands []model.Command
	if err := model.DB.Find(&commands, "session_id = ?", sessionID).Error; err != nil {
		return fmt.Errorf("获取命令失败: %v", err)
	}

	for _, cmd := range commands {
		fmt.Printf("[%s] $ %s\n", cmd.ExecTime.Format("15:04:05"), cmd.Command)

		var outputs []model.CommandOutput
		model.DB.Find(&outputs, "command_id = ?", cmd.ID)

		for _, output := range outputs {
			fmt.Printf("[%s] 输出:\n%s\n", output.NodeIP, output.Output)
		}

		fmt.Printf("退出码: %d, 耗时: %dms\n", cmd.ExitCode, cmd.Duration)
		fmt.Println("-----------------------------------")
	}

	return nil
}

// GetRecentSessions 获取最近的会话
func GetRecentSessions(limit int) ([]model.Session, error) {
	var sessions []model.Session
	if err := model.DB.Order("start_time desc").Limit(limit).Find(&sessions).Error; err != nil {
		return nil, err
	}
	return sessions, nil
}
