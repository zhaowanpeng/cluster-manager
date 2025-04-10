package session

import (
	"fmt"
	"strings"
	"sync"
	"time"
	"zhaowanpeng/cluster-manager/internal/crud"
	"zhaowanpeng/cluster-manager/internal/utils/ip_util"
	"zhaowanpeng/cluster-manager/model"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
)

// ExecOptions 表示执行命令的选项
type ExecOptions struct {
	GroupName    string
	Timeout      time.Duration
	ExcludeNodes string
	AddNodes     string
	MergeOutput  bool
	Port         int
	User         string
	Password     string
}

// ExecResult 表示命令执行结果
type ExecResult struct {
	Node    model.Node
	Output  string
	Error   error
	Success bool
}

// StartGroupExec 启动组执行会话
func StartGroupExec(options ExecOptions) error {
	// 1. 获取组中的节点
	group, err := crud.GetGroup(options.GroupName)
	if err != nil {
		return fmt.Errorf("获取组信息失败: %v", err)
	}

	// 如果组中没有节点，则从组中获取节点
	nodes, err := crud.GetNodesInGroup(options.GroupName)
	if err != nil {
		return fmt.Errorf("获取组节点失败: %v", err)
	}

	// if len(nodes) == 0 {
	// 	return fmt.Errorf("组 '%s' 中没有可用节点", options.GroupName)
	// }

	// 2. 处理添加节点
	if options.AddNodes != "" {
		addIPs, err := ip_util.ParseIPRange(options.AddNodes)
		if err != nil {
			return fmt.Errorf("解析添加节点失败: %v", err)
		}

		// 添加新节点
		for _, ip := range addIPs {
			// 检查是否已存在
			exists := false
			for _, node := range nodes {
				if node.IP == ip {
					exists = true
					break
				}
			}

			if !exists {
				// 使用组的默认连接信息
				newNode := model.Node{
					IP:       ip,
					Port:     options.Port,
					User:     options.User,
					Password: options.Password,
					Group:    group.Name,
				}
				nodes = append(nodes, newNode)
			}
		}
	}

	// 3. 处理排除节点
	if options.ExcludeNodes != "" {
		excludeIPs, err := ip_util.ParseIPRange(options.ExcludeNodes)
		if err != nil {
			return fmt.Errorf("解析排除节点失败: %v", err)
		}

		// 过滤掉排除的节点
		var filteredNodes []model.Node
		for _, node := range nodes {
			if !contains(excludeIPs, node.IP) {
				filteredNodes = append(filteredNodes, node)
			}
		}
		nodes = filteredNodes

		if len(nodes) == 0 {
			return fmt.Errorf("所有节点都被排除了")
		}
	}

	// 4. 创建会话管理器
	sessionManager := NewSessionManager()
	defer sessionManager.CloseAll()

	// 5. 预连接所有节点
	color.Yellow("正在建立SSH连接到所有节点...")
	var wg sync.WaitGroup
	var mutex sync.Mutex
	var failedNodes []string

	for _, node := range nodes {
		wg.Add(1)
		go func(node model.Node) {
			defer wg.Done()
			_, err := sessionManager.GetOrCreateSession(node)
			if err != nil {
				mutex.Lock()
				failedNodes = append(failedNodes, node.IP)
				color.Red("连接节点 %s 失败: %v", node.IP, err)
				mutex.Unlock()
			}
		}(node)
	}
	wg.Wait()

	// 移除连接失败的节点
	if len(failedNodes) > 0 {
		var connectedNodes []model.Node
		for _, node := range nodes {
			if !contains(failedNodes, node.IP) {
				connectedNodes = append(connectedNodes, node)
			}
		}
		nodes = connectedNodes
	}

	if len(nodes) == 0 {
		return fmt.Errorf("所有连接都失败了")
	}

	// 6. 显示连接信息
	color.Green("已连接到组 '%s' 的 %d 个节点", options.GroupName, len(nodes))
	// if group.Description != "" {
	// 	fmt.Printf("描述: %s\n", group.Description)
	// }

	// 7. 分组显示节点信息
	// eg:
	// 192.168.1.1
	// 192.168.1.2
	// 192.168.1.3
	// 192.168.1.4
	// 192.168.1.5
	// 192.168.1.6
	nodesBySubnet := groupNodesBySubnet(nodes)
	color.Cyan("已连接节点:")
	for subnet, subnetNodes := range nodesBySubnet {
		color.Yellow("%s:", subnet)
		for _, node := range subnetNodes {
			fmt.Printf("  - %s\n", node.IP)
		}
	}
	fmt.Println()

	// 8. 创建交互式会话
	rl, err := readline.New(color.GreenString(options.GroupName) + " > ")
	if err != nil {
		return fmt.Errorf("创建交互式会话失败: %v", err)
	}
	defer rl.Close()

	// 9. 交互式循环
	for {
		line, err := rl.Readline()
		// 如果输入为空，则退出
		if err != nil {
			break
		}

		// 去掉命令前后空白
		command := strings.TrimSpace(line)
		if command == "" {
			continue
		}

		// 处理特殊命令，只有ctrl+q退出
		if command == "ctrl+q" {
			color.Green("退出会话")
			break
		}
		// if command == "exit" || command == "quit" {
		// 	color.Green("退出会话")
		// 	break
		// }

		if command == "nodes" {
			// 显示当前连接的节点
			displayConnectedNodes(nodesBySubnet)
			continue
		}

		// 执行命令并收集结果
		results := executeCommandOnNodes(sessionManager, nodes, command, options.Timeout)

		// 显示结果
		if options.MergeOutput {
			displayMergedResults(nodesBySubnet, results)
		} else {
			displayResults(nodesBySubnet, results)
		}
	}

	return nil
}

// executeCommandOnNodes 在所有节点上执行命令
func executeCommandOnNodes(sessionManager *SessionManager, nodes []model.Node, command string, timeout time.Duration) map[string]ExecResult {
	results := make(map[string]ExecResult)
	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, node := range nodes {
		wg.Add(1)
		go func(node model.Node) {
			defer wg.Done()

			// 获取节点会话
			session, err := sessionManager.GetOrCreateSession(node)
			if err != nil {
				mutex.Lock()
				results[node.IP] = ExecResult{
					Node:    node,
					Output:  "",
					Error:   fmt.Errorf("获取会话失败: %v", err),
					Success: false,
				}
				mutex.Unlock()
				return
			}

			// 执行命令
			output, err := session.ExecuteCommand(command, timeout)

			mutex.Lock()
			results[node.IP] = ExecResult{
				Node:    node,
				Output:  output,
				Error:   err,
				Success: err == nil,
			}
			mutex.Unlock()
		}(node)
	}

	wg.Wait()
	return results
}

// displayConnectedNodes 显示当前连接的节点
func displayConnectedNodes(nodesBySubnet map[string][]model.Node) {
	color.Cyan("当前连接的节点:")
	for subnet, nodes := range nodesBySubnet {
		color.Yellow("%s:", subnet)
		for _, node := range nodes {
			fmt.Printf("  - %s (用户: %s, 端口: %d)\n", node.IP, node.User, node.Port)
		}
	}
	fmt.Println()
}

// displayResults 显示命令执行结果
func displayResults(nodesBySubnet map[string][]model.Node, results map[string]ExecResult) {
	for subnet, nodes := range nodesBySubnet {
		fmt.Print("\n")
		color.New(color.FgHiCyan).Printf("%s:\n", subnet)

		for _, node := range nodes {
			result, ok := results[node.IP]
			if !ok {
				continue
			}

			if !result.Success {
				errorDesc := result.Error.Error()
				if strings.Contains(errorDesc, "exited with status 127") {
					color.New(color.FgRed).Printf("[%s] 命令未找到: %s\n", node.IP, errorDesc)
				} else if strings.Contains(errorDesc, "permission denied") {
					color.New(color.FgRed).Printf("[%s] 权限拒绝: %s\n", node.IP, errorDesc)
				} else if strings.Contains(errorDesc, "connection refused") {
					color.New(color.FgRed).Printf("[%s] 连接被拒绝: %s\n", node.IP, errorDesc)
				} else if strings.Contains(errorDesc, "timeout") {
					color.New(color.FgRed).Printf("[%s] 执行超时: %s\n", node.IP, errorDesc)
				} else {
					color.New(color.FgRed).Printf("[%s] %s\n", node.IP, errorDesc)
				}

				if result.Output != "" {
					fmt.Printf("    错误输出: %s\n", result.Output)
				}
			} else {
				color.New(color.FgGreen).Printf("[%s] ", node.IP)
				fmt.Printf("%s\n", result.Output)
			}
		}
	}
	fmt.Print("\n")
}

// 辅助函数：检查字符串是否在切片中
func contains(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

// 按子网分组节点
func groupNodesBySubnet(nodes []model.Node) map[string][]model.Node {
	groups := make(map[string][]model.Node)

	for _, node := range nodes {
		// 提取IP的前三段作为分组依据
		parts := strings.Split(node.IP, ".")
		if len(parts) >= 3 {
			groupKey := fmt.Sprintf("%s.%s.%s", parts[0], parts[1], parts[2])
			groups[groupKey] = append(groups[groupKey], node)
		} else {
			// 如果IP格式不标准，使用原始IP作为键
			groups[node.IP] = append(groups[node.IP], node)
		}
	}

	return groups
}
