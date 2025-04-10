package session

import (
	"bytes"
	"fmt"
	"strings"
	"time"
	"zhaowanpeng/cluster-manager/model"

	"golang.org/x/crypto/ssh"
)

// NodeSession 表示与单个节点的长会话
type NodeSession struct {
	Node            model.Node
	client          *ssh.Client
	shellSession    *ssh.Session
	stdin           *bytes.Buffer
	stdout          *bytes.Buffer
	stderr          *bytes.Buffer
	environmentVars map[string]string
}

// NewNodeSession 创建新的节点会话
func NewNodeSession(node model.Node) (*NodeSession, error) {
	// 创建SSH客户端连接
	config := &ssh.ClientConfig{
		User: node.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(node.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", node.IP, node.Port), config)
	if err != nil {
		return nil, fmt.Errorf("连接到节点 %s 失败: %v", node.IP, err)
	}

	// 创建会话对象
	session := &NodeSession{
		Node:            node,
		client:          client,
		environmentVars: make(map[string]string),
	}

	// 初始化shell会话
	if err := session.initShellSession(); err != nil {
		client.Close()
		return nil, err
	}

	return session, nil
}

// initShellSession 初始化交互式shell会话
func (ns *NodeSession) initShellSession() error {
	session, err := ns.client.NewSession()
	if err != nil {
		return fmt.Errorf("创建SSH会话失败: %v", err)
	}

	// 设置I/O
	ns.stdin = &bytes.Buffer{}
	ns.stdout = &bytes.Buffer{}
	ns.stderr = &bytes.Buffer{}

	session.Stdin = ns.stdin
	session.Stdout = ns.stdout
	session.Stderr = ns.stderr

	// 请求伪终端
	if err := session.RequestPty("xterm", 80, 40, ssh.TerminalModes{}); err != nil {
		session.Close()
		return fmt.Errorf("请求伪终端失败: %v", err)
	}

	// 启动shell
	if err := session.Shell(); err != nil {
		session.Close()
		return fmt.Errorf("启动shell失败: %v", err)
	}

	ns.shellSession = session

	// 等待shell准备就绪
	time.Sleep(500 * time.Millisecond)
	ns.stdout.Reset()

	return nil
}

// ExecuteCommand 在长会话中执行命令
func (ns *NodeSession) ExecuteCommand(command string, timeout time.Duration) (string, error) {
	// 清空缓冲区
	ns.stdout.Reset()
	ns.stderr.Reset()

	// 写入命令
	commandWithEcho := fmt.Sprintf("%s; echo \"COMMAND_COMPLETED_$?\"\n", command)
	_, err := ns.stdin.WriteString(commandWithEcho)
	if err != nil {
		return "", fmt.Errorf("写入命令失败: %v", err)
	}

	// 等待命令完成
	completed := make(chan bool, 1)
	var result string
	var cmdErr error

	go func() {
		// 等待命令完成标记
		for {
			time.Sleep(100 * time.Millisecond)
			output := ns.stdout.String()
			if strings.Contains(output, "COMMAND_COMPLETED_") {
				// 提取命令输出和退出码
				parts := strings.Split(output, "COMMAND_COMPLETED_")
				if len(parts) > 1 {
					result = strings.TrimSpace(parts[0])
					exitCode := strings.TrimSpace(parts[1])
					if exitCode != "0" {
						cmdErr = fmt.Errorf("命令退出码: %s", exitCode)
					}
				}
				completed <- true
				return
			}
		}
	}()

	// 处理超时
	select {
	case <-completed:
		return result, cmdErr
	case <-time.After(timeout):
		return "", fmt.Errorf("命令执行超时")
	}
}

// 在会话中执行shell文件内容
func (ns *NodeSession) ExecuteShellFile(shellContent string) (string, error) {
	cmd := fmt.Sprintf("cat <<EOF | bash\n%s\nEOF", shellContent)
	return ns.ExecuteCommand(cmd, 5*time.Second)
}

// 恢复会话
// func (ns *NodeSession) RestoreSession() error {

// }

// SetEnvironmentVariable 设置环境变量
func (ns *NodeSession) SetEnvironmentVariable(name, value string) error {
	cmd := fmt.Sprintf("export %s=%s", name, value)
	_, err := ns.ExecuteCommand(cmd, 5*time.Second)
	if err != nil {
		return err
	}

	// 保存到本地记录
	ns.environmentVars[name] = value
	return nil
}

// GetEnvironmentVariable 获取环境变量值
func (ns *NodeSession) GetEnvironmentVariable(name string) (string, error) {
	cmd := fmt.Sprintf("echo $%s", name)
	return ns.ExecuteCommand(cmd, 5*time.Second)
}

// Close 关闭会话
func (ns *NodeSession) Close() {
	if ns.shellSession != nil {
		ns.shellSession.Close()
	}
	if ns.client != nil {
		ns.client.Close()
	}
}
