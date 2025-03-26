package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"zhaowanpeng/cluster-manager/internal/crud"
	"zhaowanpeng/cluster-manager/internal/utils"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "管理节点组",
	Long:  "管理节点组，包括添加、删除、列出和编辑组",
}

var groupListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有组",
	Run:   groupListFunc,
}

var groupAddCmd = &cobra.Command{
	Use:   "add",
	Short: "添加新组",
	Run:   groupAddFunc,
}

var groupRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "删除组",
	Run:   groupRemoveFunc,
}

var groupShowCmd = &cobra.Command{
	Use:   "show",
	Short: "显示组详情",
	Run:   groupShowFunc,
}

// 命令行参数
var (
	groupName        string
	groupNodes       string
	groupPort        int
	groupUser        string
	groupPassword    bool
	groupDescription string
	groupForce       bool
)

func init() {
	// 添加子命令到 group 命令
	groupCmd.AddCommand(groupListCmd)
	groupCmd.AddCommand(groupAddCmd)
	groupCmd.AddCommand(groupRemoveCmd)
	groupCmd.AddCommand(groupShowCmd)

	// 添加 group 命令到根命令
	rootCmd.AddCommand(groupCmd)

	// 设置 add 命令的参数
	groupAddCmd.Flags().StringVarP(&groupName, "name", "n", "", "组名称")
	groupAddCmd.Flags().StringVarP(&groupNodes, "nodes", "i", "", "节点列表，支持范围表示法，如 192.168.1.1-5,192.168.1.10")
	groupAddCmd.Flags().IntVarP(&groupPort, "port", "p", 22, "SSH 端口")
	groupAddCmd.Flags().StringVarP(&groupUser, "user", "u", "root", "SSH 用户名")
	groupAddCmd.Flags().BoolVarP(&groupPassword, "password", "P", false, "提示输入密码")
	groupAddCmd.Flags().StringVarP(&groupDescription, "description", "d", "", "组描述")

	// 设置 remove 命令的参数
	groupRemoveCmd.Flags().StringVarP(&groupName, "name", "n", "", "组名称")
	groupRemoveCmd.Flags().BoolVarP(&groupForce, "force", "f", false, "强制删除，不提示确认")

	// 设置 show 命令的参数
	groupShowCmd.Flags().StringVarP(&groupName, "name", "n", "", "组名称")
}

// groupListFunc 列出所有组
func groupListFunc(cmd *cobra.Command, args []string) {
	groups, err := crud.ListGroups()
	if err != nil {
		color.Red("列出组失败: %v", err)
		return
	}

	if len(groups) == 0 {
		fmt.Println("没有找到任何组")
		return
	}

	// 打印组列表
	fmt.Println("组列表:")
	fmt.Println("----------------------------------------")
	for _, group := range groups {
		nodeCount, err := crud.CountNodesInGroup(group.Name)
		if err != nil {
			nodeCount = 0
		}
		color.Green("%s (%d 个节点)", group.Name, nodeCount)
		if group.Description != "" {
			fmt.Printf("  描述: %s\n", group.Description)
		}
		fmt.Printf("  创建时间: %s\n", group.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Println("----------------------------------------")
	}
}

// groupAddFunc 添加新组
func groupAddFunc(cmd *cobra.Command, args []string) {
	reader := bufio.NewReader(os.Stdin)

	// 如果命令行没有提供组名，交互式获取
	if groupName == "" {
		fmt.Print("组名称: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			color.Red("读取输入失败: %v", err)
			return
		}
		groupName = strings.TrimSpace(input)
		if groupName == "" {
			color.Red("组名称不能为空")
			return
		}
	}

	// 如果命令行没有提供节点列表，交互式获取
	if groupNodes == "" {
		fmt.Print("节点列表 [逗号分隔或范围，例如 192.168.1.1-5,192.168.1.10]: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			color.Red("读取输入失败: %v", err)
			return
		}
		groupNodes = strings.TrimSpace(input)
		if groupNodes == "" {
			color.Red("节点列表不能为空")
			return
		}
	}

	// 交互式获取端口
	if !cmd.Flags().Changed("port") {
		fmt.Printf("端口 [%d]: ", groupPort)
		input, err := reader.ReadString('\n')
		if err != nil {
			color.Red("读取输入失败: %v", err)
			return
		}
		input = strings.TrimSpace(input)
		if input != "" {
			fmt.Sscanf(input, "%d", &groupPort)
		}
	}

	// 交互式获取用户名
	if !cmd.Flags().Changed("user") {
		fmt.Printf("用户名 [%s]: ", groupUser)
		input, err := reader.ReadString('\n')
		if err != nil {
			color.Red("读取输入失败: %v", err)
			return
		}
		input = strings.TrimSpace(input)
		if input != "" {
			groupUser = input
		}
	}

	// 获取密码
	var password string
	if !groupPassword {
		fmt.Print("密码 [输入以安全方式输入]: ")
		bytePwd, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			color.Red("读取密码失败: %v", err)
			return
		}
		password = string(bytePwd)
	} else {
		fmt.Print("密码: ")
		bytePwd, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			color.Red("读取密码失败: %v", err)
			return
		}
		password = string(bytePwd)
	}

	// 交互式获取描述
	if !cmd.Flags().Changed("description") {
		fmt.Print("描述: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			color.Red("读取输入失败: %v", err)
			return
		}
		groupDescription = strings.TrimSpace(input)
	}

	// 解析节点列表
	ips, err := utils.ParseIPList(groupNodes)
	if err != nil {
		color.Red("解析节点列表失败: %v", err)
		return
	}

	// 创建组
	err = crud.AddGroup(groupName, groupDescription)
	if err != nil {
		color.Red("创建组失败: %v", err)
		return
	}

	// 添加节点到组
	results, err := crud.AddNodesToGroup(groupName, ips, groupPort, groupUser, password, groupDescription)
	if err != nil {
		color.Red("添加节点失败: %v", err)
		return
	}

	// 统计结果
	successCount := 0
	failureMap := make(map[string][]string)

	for _, result := range results {
		if result.Success {
			successCount++
		} else {
			failureMap[result.Msg] = append(failureMap[result.Msg], result.IP)
		}
	}

	// 显示结果
	fmt.Println("验证连接...")
	color.Green("✓ 成功连接到 %d/%d 个节点", successCount, len(ips))

	for errMsg, failedIPs := range failureMap {
		color.Red("! 连接失败: %s", errMsg)
		for _, ip := range failedIPs {
			fmt.Printf("  - %s\n", ip)
		}
	}

	// 询问是否为失败的节点创建临时组
	if len(failureMap) > 0 {
		fmt.Print("为失败的节点创建临时组? [Y/n]: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			color.Red("读取输入失败: %v", err)
			return
		}
		input = strings.ToLower(strings.TrimSpace(input))
		if input == "y" || input == "yes" || input == "" {
			for errMsg, failedIPs := range failureMap {
				// 创建临时组名称
				tmpGroupName := fmt.Sprintf("tmp-%s-%s",
					strings.ReplaceAll(strings.ToLower(errMsg[:10]), " ", "-"),
					utils.GenerateShortID())

				err = crud.AddGroup(tmpGroupName, fmt.Sprintf("自动创建的临时组: %s", errMsg))
				if err != nil {
					color.Red("创建临时组失败: %v", err)
					continue
				}

				// 添加失败的节点到临时组
				_, err = crud.AddNodesToGroup(tmpGroupName, failedIPs, groupPort, groupUser, password, fmt.Sprintf("自动添加: %s", errMsg))
				if err != nil {
					color.Red("添加节点到临时组失败: %v", err)
					continue
				}

				color.Yellow("临时组 '%s' 已创建，包含 %d 个节点", tmpGroupName, len(failedIPs))
			}
		}
	}

	color.Green("组 '%s' 已创建，包含 %d 个节点", groupName, successCount)
}

// groupRemoveFunc 删除组
func groupRemoveFunc(cmd *cobra.Command, args []string) {
	reader := bufio.NewReader(os.Stdin)

	// 如果命令行没有提供组名，交互式获取
	if groupName == "" {
		fmt.Print("组名称: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			color.Red("读取输入失败: %v", err)
			return
		}
		groupName = strings.TrimSpace(input)
		if groupName == "" {
			color.Red("组名称不能为空")
			return
		}
	}

	// 确认删除
	if !groupForce {
		fmt.Printf("确定要删除组 '%s' 吗? [y/N]: ", groupName)
		input, err := reader.ReadString('\n')
		if err != nil {
			color.Red("读取输入失败: %v", err)
			return
		}
		input = strings.ToLower(strings.TrimSpace(input))
		if input != "y" && input != "yes" {
			fmt.Println("操作已取消")
			return
		}
	}

	// 删除组
	err := crud.RemoveGroup(groupName)
	if err != nil {
		color.Red("删除组失败: %v", err)
		return
	}

	color.Green("组 '%s' 已删除", groupName)
}

// groupShowFunc 显示组详情
func groupShowFunc(cmd *cobra.Command, args []string) {
	reader := bufio.NewReader(os.Stdin)

	// 如果命令行没有提供组名，交互式获取
	if groupName == "" {
		fmt.Print("组名称: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			color.Red("读取输入失败: %v", err)
			return
		}
		groupName = strings.TrimSpace(input)
		if groupName == "" {
			color.Red("组名称不能为空")
			return
		}
	}

	// 获取组信息
	group, err := crud.GetGroup(groupName)
	if err != nil {
		color.Red("获取组信息失败: %v", err)
		return
	}

	// 获取组中的节点
	nodes, err := crud.GetNodesInGroup(groupName)
	if err != nil {
		color.Red("获取节点信息失败: %v", err)
		return
	}

	// 显示组信息
	color.Green("组: %s", group.Name)
	fmt.Printf("描述: %s\n", group.Description)
	fmt.Printf("创建时间: %s\n", group.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("更新时间: %s\n", group.UpdatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("节点数量: %d\n", len(nodes))

	if len(nodes) > 0 {
		fmt.Println("\n节点列表:")
		fmt.Println("----------------------------------------")
		for i, node := range nodes {
			statusSymbol := "✓"
			statusColor := color.New(color.FgGreen)
			if !node.Usable {
				statusSymbol = "✗"
				statusColor = color.New(color.FgRed)
			}

			statusColor.Printf("%s %s", statusSymbol, node.IP)
			fmt.Printf(" (端口: %d, 用户: %s)\n", node.Port, node.User)

			if i < len(nodes)-1 {
				fmt.Println("----------------------------------------")
			}
		}
	}
}
