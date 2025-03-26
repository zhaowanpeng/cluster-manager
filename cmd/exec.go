package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
	"zhaowanpeng/cluster-manager/internal/utils"
	"zhaowanpeng/cluster-manager/model"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

var (
	execGroup       string
	execNodes       string
	execTimeout     int
	execSaveResults bool
	execSaveDir     string
)

var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "命令分发",
	Long:  "Example：\n  clush exec -g g1 ",
	Run:   execFunc,
}

// 添加 group 子命令
var execGroupCmd = &cobra.Command{
	Use:   "group",
	Short: "向指定的组发送命令",
	Long:  "向指定组中的所有节点发送命令，并支持保存命令和响应\n  Example: clush exec group [组名]",
	Run:   execGroupFunc,
}

// SSH客户端连接池
type SSHClientPool struct {
	clients map[string]*ssh.Client
	mu      sync.Mutex
}

var globalSSHPool = &SSHClientPool{
	clients: make(map[string]*ssh.Client),
}

// 获取或创建SSH连接
func (p *SSHClientPool) GetClient(ip string, port int, user, password string) (*ssh.Client, error) {
	key := fmt.Sprintf("%s:%d:%s", ip, port, user)

	p.mu.Lock()
	defer p.mu.Unlock()

	// 检查是否已有连接且可用
	if client, ok := p.clients[key]; ok {
		// 测试连接是否仍然可用
		_, _, err := client.SendRequest("keepalive@golang", true, nil)
		if err == nil {
			return client, nil
		}
		// 连接已断开，删除并重新创建
		delete(p.clients, key)
	}

	// 创建新连接
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", ip, port), config)
	if err != nil {
		return nil, err
	}

	// 存储到池中
	p.clients[key] = client
	return client, nil
}

// 关闭所有连接
func (p *SSHClientPool) CloseAll() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for key, client := range p.clients {
		client.Close()
		delete(p.clients, key)
	}
}

func execFunc(cmd *cobra.Command, args []string) {
	if (execGroup == "" && execNodes == "") || (execGroup != "" && execNodes != "") {
		fmt.Println("必须指定 -g 或 -n 其中之一")
		return
	}

	var nodes []model.ShellClient

	if execGroup != "" {
		err := model.DB.Where("`group` = ? AND `usable` = ?", execGroup, true).Find(&nodes).Error
		if err != nil {
			fmt.Printf("查询分组 %s 失败: %v\n", execGroup, err)
			return
		}
		if len(nodes) == 0 {
			fmt.Printf("分组 %s 没有找到节点\n", execGroup)
			return
		}
	}

	if execNodes != "" {
		fmt.Println("user")
		fmt.Println("pwd")
		fmt.Println("port")
	}

	for _, clt := range nodes {
		fmt.Println(clt.IP)
	}

	// if execNodes != "" {
	// 	ips, err := utils.ParseIPList(execNodes)
	// 	if err != nil {
	// 		fmt.Printf("解析节点 IP 失败: %v\n", err)
	// 		return
	// 	}
	// 	if len(ips) == 0 {
	// 		fmt.Println("没有解析到有效的 IP")
	// 		return
	// 	}
	// 	for _, ip := range ips {
	// 		node := model.ShellClient{
	// 			IP:       ip,
	// 			Port:     execPort,
	// 			User:     execUser,
	// 			Password: pwd,
	// 		}
	// 		nodes = append(nodes, node)
	// 	}
	// }

	if len(args) > 0 {
		fmt.Println("单次模式")
	} else {
		interactiveMode(nodes)
	}
}

// 执行组命令
func execGroupFunc(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Println("用法: clush exec group [组名]")
		return
	}

	groupName := args[0]

	// 修改这里：使用正确的列名 "group" 而不是 "group_name"
	var nodes []model.Node
	err := model.DB.Where("`group` = ? AND `usable` = ?", groupName, true).Find(&nodes).Error
	if err != nil {
		color.Red("查询分组 %s 失败: %v", groupName, err)
		return
	}

	if len(nodes) == 0 {
		color.Red("分组 %s 没有找到可用节点", groupName)
		return
	}

	color.Green("已连接到 '%s' 组的 %d 个节点", groupName, len(nodes))
	color.Yellow("进入交互式模式，输入 'exit' 退出")

	// 进入交互模式，同样需要使用 Node 结构
	interactiveGroupSession(nodes, groupName)
}

// 交互式组会话
func interactiveGroupSession(nodes []model.Node, groupName string) {
	// 初始连接所有节点
	color.Yellow("正在建立SSH连接到所有节点...")

	// 并行预连接所有节点
	var wg sync.WaitGroup
	for _, node := range nodes {
		wg.Add(1)
		go func(node model.Node) {
			defer wg.Done()
			_, err := globalSSHPool.GetClient(node.IP, node.Port, node.User, node.Password)
			if err != nil {
				color.Red("预连接节点 %s 失败: %v", node.IP, err)
			}
		}(node)
	}
	wg.Wait()

	color.Green("所有连接已准备就绪")

	// 分组显示节点信息
	nodesBySubgroup := groupNodesByIP(nodes)
	color.Cyan("已连接节点:")
	for subgroup, nodeList := range nodesBySubgroup {
		color.Yellow("%s:", subgroup)
		for _, node := range nodeList {
			fmt.Printf("  - %s\n", node.IP)
		}
	}
	fmt.Println()

	// 创建readline
	rl, err := readline.New(">>> ")
	if err != nil {
		fmt.Printf("创建交互会话失败: %v\n", err)
		return
	}
	defer rl.Close()

	// 交互式循环
	for {
		line, err := rl.Readline()
		if err != nil {
			break
		}

		command := strings.TrimSpace(line)
		if command == "" {
			continue
		}

		if command == "exit" {
			color.Green("退出会话")
			break
		}

		// 执行命令并按子组显示结果
		executeCommandOnGroupedNodes(nodes, command, nodesBySubgroup)
	}
}

// 同样修改这个函数的参数类型
func groupNodesByIP(nodes []model.Node) map[string][]model.Node {
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

// 按子组执行命令并显示结果
func executeCommandOnGroupedNodes(nodes []model.Node, command string, nodeGroups map[string][]model.Node) {
	results := make(map[string]string)
	errors := make(map[string]error)
	var mutex sync.Mutex
	var wg sync.WaitGroup

	// 执行命令并收集结果
	for _, node := range nodes {
		wg.Add(1)
		go func(node model.Node) {
			defer wg.Done()

			output, err := executeCommandWithPooledConnection(
				node,
				command,
				time.Duration(execTimeout)*time.Second,
			)

			mutex.Lock()
			if err != nil {
				// 保存详细错误信息
				errors[node.IP] = err
				// 捕获错误输出（如果有）
				if output != "" {
					results[node.IP] = output
				} else {
					results[node.IP] = ""
				}
			} else {
				results[node.IP] = strings.TrimSpace(output)
			}
			mutex.Unlock()
		}(node)
	}
	wg.Wait()

	// 按子组显示结果，使用颜色增强可读性
	for subgroup, nodeList := range nodeGroups {
		ipRange := formatIPRange(nodeList)

		// 子组标题显示为青色
		fmt.Print("\n")
		color.New(color.FgHiCyan).Printf("%s (%s):\n", subgroup, ipRange)

		// 为每个节点结果添加颜色
		for _, node := range nodeList {
			if err, hasError := errors[node.IP]; hasError {
				// 详细显示错误信息
				errorDesc := err.Error()

				// 特殊处理常见的错误类型
				if strings.Contains(errorDesc, "exited with status 127") {
					color.New(color.FgRed).Printf("[%s] 命令未找到 (127): %s\n", node.IP, errorDesc)
				} else if strings.Contains(errorDesc, "permission denied") {
					color.New(color.FgRed).Printf("[%s] 权限拒绝: %s\n", node.IP, errorDesc)
				} else if strings.Contains(errorDesc, "connection refused") {
					color.New(color.FgRed).Printf("[%s] 连接被拒绝: %s\n", node.IP, errorDesc)
				} else if strings.Contains(errorDesc, "timeout") {
					color.New(color.FgRed).Printf("[%s] 执行超时: %s\n", node.IP, errorDesc)
				} else {
					// 其他错误类型
					color.New(color.FgRed).Printf("[%s] %s\n", node.IP, errorDesc)
				}

				// 如果有错误输出，也显示它
				if output := results[node.IP]; output != "" {
					fmt.Printf("    错误输出: %s\n", output)
				}
			} else if output, ok := results[node.IP]; ok {
				// 正常输出
				color.New(color.FgGreen).Printf("[%s] ", node.IP)
				fmt.Printf("%s\n", output)
			}
		}
	}
	fmt.Print("\n")
}

// 格式化IP范围显示
func formatIPRange(nodes []model.Node) string {
	if len(nodes) == 0 {
		return ""
	}

	// 收集所有IP
	ips := make([]string, 0, len(nodes))
	for _, node := range nodes {
		ips = append(ips, node.IP)
	}

	// 简单合并显示
	if len(ips) <= 3 {
		return strings.Join(ips, ", ")
	}

	// 对于多个IP，显示范围
	return fmt.Sprintf("%s-%s", ips[0], ips[len(ips)-1])
}

// 执行命令并收集结果
func executeCommandAndCollectResults(nodes []model.Node, command string) map[string]string {
	var wg sync.WaitGroup
	var mutex sync.Mutex
	results := make(map[string]string)

	color.Yellow("开始执行命令...")

	for _, node := range nodes {
		wg.Add(1)
		go func(node model.Node) {
			defer wg.Done()

			output, err := executeCommandWithPooledConnection(node, command, time.Duration(execTimeout)*time.Second)

			// 输出格式化
			label := node.IP
			if node.Description != "" {
				label = fmt.Sprintf("%s (%s)", node.IP, node.Description)
			}

			// 保存结果
			mutex.Lock()
			if err != nil {
				errorMsg := fmt.Sprintf("执行失败: %v", err)
				results[node.IP] = errorMsg
				color.Red("[%s] %s", label, errorMsg)
			} else {
				results[node.IP] = output
				color.Green("[%s] 执行成功:", label)
				fmt.Println(output)
			}
			mutex.Unlock()
		}(node)
	}

	wg.Wait()
	return results
}

// 保存命令结果到文件
func saveCommandResults(groupName, command string, results map[string]string) {
	// 创建保存目录
	if execSaveDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			color.Red("获取用户主目录失败: %v", err)
			return
		}
		execSaveDir = fmt.Sprintf("%s/clush-results", homeDir)
	}

	// 确保目录存在
	if err := os.MkdirAll(execSaveDir, 0755); err != nil {
		color.Red("创建结果保存目录失败: %v", err)
		return
	}

	// 创建时间戳目录
	timestamp := time.Now().Format("20060102_150405")
	resultDir := fmt.Sprintf("%s/%s_%s", execSaveDir, groupName, timestamp)
	if err := os.MkdirAll(resultDir, 0755); err != nil {
		color.Red("创建结果子目录失败: %v", err)
		return
	}

	// 保存命令
	commandFile := fmt.Sprintf("%s/command.txt", resultDir)
	if err := os.WriteFile(commandFile, []byte(command), 0644); err != nil {
		color.Red("保存命令失败: %v", err)
		return
	}

	// 保存每个节点的结果
	for ip, output := range results {
		resultFile := fmt.Sprintf("%s/%s.txt", resultDir, ip)
		if err := os.WriteFile(resultFile, []byte(output), 0644); err != nil {
			color.Red("保存节点 %s 的结果失败: %v", ip, err)
			continue
		}
	}

	color.Green("命令结果已保存到: %s", resultDir)
}

func interactiveMode(nodes []model.ShellClient) {
	rl, err := readline.New(">>> ")
	if err != nil {
		fmt.Printf("创建 readline 失败: %v\n", err)
		return
	}
	defer rl.Close()

	// 自动保存历史记录功能（readline 默认支持上下翻阅历史记录）
	for {
		line, err := rl.Readline()
		if err != nil {
			fmt.Printf("bye ~")
			break
		}
		command := strings.TrimSpace(line)
		if command == "exit" {
			break
		}
		fmt.Println(command)
		if command == "ls -l" {
			executeCommandOnNodes(nodes, command)
		}

	}
}

func executeCommandOnNodes(nodes []model.ShellClient, command string) {
	var wg sync.WaitGroup
	// 为了方便显示，我们给每个节点生成 g1(), g2() 等标识
	for i, node := range nodes {
		wg.Add(1)
		go func(i int, node model.ShellClient) {
			defer wg.Done()
			label := fmt.Sprintf("g%d()", i+1)
			output, err := utils.Exec_SSH_Command(
				node.IP,
				node.Port,
				node.User,
				node.Password,
				command,
				time.Duration(execTimeout)*time.Second,
			)
			if err != nil {
				fmt.Printf("%s error: %v\n", label, err)
			} else {
				// 如输出中有换行，自动换行显示
				fmt.Printf("%s output:\n%s\n", label, output)
			}
		}(i, node)
	}
	wg.Wait()
}

// 执行命令使用连接池中的连接
func executeCommandWithPooledConnection(node model.Node, command string, timeout time.Duration) (string, error) {
	// 从连接池获取连接
	client, err := globalSSHPool.GetClient(node.IP, node.Port, node.User, node.Password)
	if err != nil {
		return "", fmt.Errorf("建立SSH连接失败: %v", err)
	}

	// 创建会话
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("创建SSH会话失败: %v", err)
	}
	defer session.Close()

	// 准备接收输出
	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	// 执行命令
	err = session.Run(command)
	if err != nil {
		// 如果有错误输出，返回它
		if stderrBuf.Len() > 0 {
			return stderrBuf.String(), err
		}
		return "", err
	}

	return stdoutBuf.String(), nil
}

func init() {
	execCmd.Flags().StringVarP(&execGroup, "group", "g", "", "指定分组，2选1")
	execCmd.Flags().StringVarP(&execNodes, "nodes", "n", "", "指定节点，2选1")
	execCmd.Flags().IntVarP(&execTimeout, "timeout", "t", 20, "执行命令超时时间，单位秒 (默认20)")

	// 为group子命令添加保存结果的参数
	execGroupCmd.Flags().BoolVarP(&execSaveResults, "save", "s", false, "保存命令和响应")
	execGroupCmd.Flags().StringVar(&execSaveDir, "save-dir", "", "保存结果的目录 (默认: ~/clush-results)")
	execGroupCmd.Flags().IntVarP(&execTimeout, "timeout", "t", 20, "执行命令超时时间，单位秒 (默认20)")

	// 添加group子命令到exec命令
	execCmd.AddCommand(execGroupCmd)
}

// 优化输出格式，避免大量重复输出拖慢速度
func displayGroupedResults(nodeGroups map[string][]model.Node, results map[string]string, errors map[string]error) {
	for subgroup, nodeList := range nodeGroups {
		// 检查该子组中的输出是否都相同
		var uniqueOutputs []string
		outputMap := make(map[string][]string) // 输出内容 -> IP列表

		for _, node := range nodeList {
			output := results[node.IP]
			if _, exists := outputMap[output]; !exists {
				uniqueOutputs = append(uniqueOutputs, output)
			}
			outputMap[output] = append(outputMap[output], node.IP)
		}

		// 如果所有节点输出相同，只显示一次
		if len(uniqueOutputs) == 1 && errors[nodeList[0].IP] == nil {
			ipRange := formatIPRange(nodeList)
			color.Green("%s (%s):", subgroup, ipRange)
			fmt.Printf("%s\n\n", uniqueOutputs[0])
		} else {
			// 否则按不同输出分组显示
			ipRange := formatIPRange(nodeList)
			color.Green("%s (%s):", subgroup, ipRange)

			for output, ips := range outputMap {
				if len(ips) > 3 {
					// 多个节点有相同输出
					fmt.Printf("[%s等%d个节点] %s\n", ips[0], len(ips), output)
				} else {
					// 少量节点，单独显示
					for _, ip := range ips {
						if err := errors[ip]; err != nil {
							color.Red("[%s] 错误: %v\n", ip, err)
						} else {
							fmt.Printf("[%s] %s\n", ip, output)
						}
					}
				}
			}
			fmt.Println()
		}
	}
}
