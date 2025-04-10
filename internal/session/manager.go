package session

import (
	"fmt"
	"sync"
	"zhaowanpeng/cluster-manager/model"
)

// SessionManager 管理所有活跃的会话
type SessionManager struct {
	sessions map[string]*NodeSession
	mu       sync.Mutex
}

// NewSessionManager 创建新的会话管理器
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*NodeSession),
	}
}

// GetOrCreateSession 获取或创建节点会话
func (sm *SessionManager) GetOrCreateSession(node model.Node) (*NodeSession, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	key := fmt.Sprintf("%s:%d:%s", node.IP, node.Port, node.User)
	if session, exists := sm.sessions[key]; exists {
		// 验证会话是否仍然有效
		err := session.Ping()
		if err == nil {
			return session, nil
		}

		// 会话失效，需要关闭并重建
		session.Close()
		delete(sm.sessions, key)
	}

	// 创建新会话
	session, err := NewNodeSession(node)
	if err != nil {
		return nil, err
	}

	sm.sessions[key] = session
	return session, nil
}

// CloseAll 关闭所有会话
func (sm *SessionManager) CloseAll() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for key, session := range sm.sessions {
		session.Close()
		delete(sm.sessions, key)
	}
}
