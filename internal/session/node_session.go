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
	// 如果已经有一个会话，先关闭它
	if ns.shellSession != nil {
		ns.shellSession.Close()
		ns.shellSession = nil
	}

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

	// 请求伪终端，使用更大的尺寸以适应更多输出
	if err := session.RequestPty("xterm", 1000, 1000, ssh.TerminalModes{
		ssh.ECHO:          0,     // 禁用回显
		ssh.TTY_OP_ISPEED: 14400, // 输入速度
		ssh.TTY_OP_OSPEED: 14400, // 输出速度
	}); err != nil {
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

	// 清除初始提示和欢迎消息
	ns.stdout.Reset()

	// 设置更友好的shell环境
	setupCmds := []string{
		"export TERM=xterm",       // 设置终端类型
		"export LANG=en_US.UTF-8", // 设置语言环境
		"export PS1='> '",         // 设置简单提示符
		"stty -echo",              // 禁用终端回显
		"unalias ls 2>/dev/null",  // 移除ls别名（如果有）
	}

	for _, cmd := range setupCmds {
		ns.stdin.WriteString(cmd + "\n")
	}

	// 再次等待这些设置命令执行完成
	time.Sleep(300 * time.Millisecond)
	ns.stdout.Reset()

	return nil
}

// ExecuteCommand 在长会话中执行命令
func (ns *NodeSession) ExecuteCommand(command string, timeout time.Duration) (string, error) {
	// 如果会话不存在，尝试初始化
	if ns.shellSession == nil {
		if err := ns.initShellSession(); err != nil {
			return "", fmt.Errorf("无法初始化会话: %v", err)
		}
	}

	// 使用管道分隔符实现更可靠的命令执行和返回
	// 这将确保命令输出与状态的分离
	execID := fmt.Sprintf("CMD_END_%d", time.Now().UnixNano())
	execCmd := fmt.Sprintf("{ %s; } 2>&1; echo -e \"\\n%s:$?:CMD_END\"\n", command, execID)

	// 清空输出缓冲区
	ns.stdout.Reset()
	ns.stderr.Reset()

	// 写入命令
	_, err := ns.stdin.WriteString(execCmd)
	if err != nil {
		return "", fmt.Errorf("写入命令失败: %v", err)
	}

	// 等待命令执行完成
	doneChan := make(chan struct{})
	var output string
	var cmdErr error

	go func() {
		startTime := time.Now()
		checkInterval := 100 * time.Millisecond

		for {
			// 避免过度消耗CPU
			time.Sleep(checkInterval)

			currentOutput := ns.stdout.String()
			endMarker := fmt.Sprintf("%s:", execID)

			// 检查输出中是否包含我们的结束标记
			if idx := strings.Index(currentOutput, endMarker); idx >= 0 {
				// 提取命令的真实输出
				commandOutput := currentOutput[:idx]

				// 提取退出码
				markerWithCode := currentOutput[idx:]
				parts := strings.Split(markerWithCode, ":")
				if len(parts) >= 3 {
					exitCode := parts[1]
					if exitCode != "0" {
						cmdErr = fmt.Errorf("命令退出码: %s", exitCode)
					}
				}

				output = strings.TrimSpace(commandOutput)
				close(doneChan)
				return
			}

			// 检查是否已超时
			if time.Since(startTime) > timeout {
				return // 让select处理超时
			}
		}
	}()

	// 等待命令完成或超时
	select {
	case <-doneChan:
		return output, cmdErr
	case <-time.After(timeout):
		// 获取当前已收集的输出
		partialOutput := ns.stdout.String()
		if strings.Contains(partialOutput, fmt.Sprintf("%s:", execID)) {
			// 如果已经有结束标记，但goroutine还没处理完，给它更多时间
			time.Sleep(500 * time.Millisecond)
			currentOutput := ns.stdout.String()
			return strings.TrimSpace(currentOutput), fmt.Errorf("命令已完成但处理超时")
		}
		return strings.TrimSpace(partialOutput), fmt.Errorf("命令执行超时")
	}
}

// 在会话中执行shell文件内容
func (ns *NodeSession) ExecuteShellFile(shellContent string) (string, error) {
	cmd := fmt.Sprintf("cat <<EOF | bash\n%s\nEOF", shellContent)
	return ns.ExecuteCommand(cmd, 10*time.Second)
}

// 恢复会话
// func (ns *NodeSession) RestoreSession() error {

// }

// SetEnvironmentVariable 设置环境变量
// 注意：环境变量只会在当前会话中有效，不会跨命令保留
func (ns *NodeSession) SetEnvironmentVariable(name, value string) error {
	cmd := fmt.Sprintf("export %s=%s", name, value)
	_, err := ns.ExecuteCommand(cmd, 3*time.Second)

	// 保存到本地记录，作为参考
	if err == nil {
		ns.environmentVars[name] = value
	}
	return err
}

// GetEnvironmentVariable 获取环境变量值
func (ns *NodeSession) GetEnvironmentVariable(name string) (string, error) {
	cmd := fmt.Sprintf("echo $%s", name)
	return ns.ExecuteCommand(cmd, 3*time.Second)
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

// Ping 检查会话是否仍然有效
func (ns *NodeSession) Ping() error {
	if ns.client == nil || ns.shellSession == nil {
		return fmt.Errorf("会话未初始化")
	}

	// 尝试执行一个简单命令来测试会话
	_, err := ns.ExecuteCommand("echo ping", 2*time.Second)
	return err
}
