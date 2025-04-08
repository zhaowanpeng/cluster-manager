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
	// rootCmd.AddCommand(groupCmd)

	// 设置 add 命令的参数
	groupAddCmd.Flags().StringVarP(&groupName, "name", "n", "", "组名称")
	groupAddCmd.Flags().StringVarP(&groupNodes, "nodes", "i", "", "节点列表，支持范围表示法，如 192.168.1.1-5,192.168.1.10")
	groupAddCmd.Flags().IntVarP(&groupPort, "port", "p", 22, "SSH 端口")
	groupAddCmd.Flags().StringVarP(&groupUser, "user", "u", "root", "SSH 用户名")
	groupAddCmd.Flags().BoolVarP(&groupPassword, "password", "P", false, "提示输入密码")
	groupAddCmd.Flags().StringVarP(&groupDescription, "description", "d", "", "组描述")
	// group位置参数第一个就是groupName

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
		color.Red("List groups failed: %v", err)
		return
	}

	if len(groups) == 0 {
		fmt.Println("No groups found")
		return
	}

	// 打印组列表
	fmt.Println("Groups list:")
	fmt.Println("----------------------------------------")
	for _, group := range groups {
		nodeCount, err := crud.CountNodesInGroup(group.Name)
		if err != nil {
			nodeCount = 0
		}
		color.Green("%s (%d nodes)", group.Name, nodeCount)
		// fmt.Printf("   User: %s\n", group.User)
		// fmt.Printf("   Port: %d\n", group.Port)
		if group.Description != "" {
			fmt.Printf("   Description: %s\n", group.Description)
		}

		fmt.Printf("   Created at: %s\n", group.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Println("----------------------------------------")
	}
}

// groupAddFunc 添加新组
func groupAddFunc(cmd *cobra.Command, args []string) {

	// 如果命令行有位置参数，则使用位置参数
	if len(args) > 0 {
		groupName = args[0]
	}

	reader := bufio.NewReader(os.Stdin)

	// 如果命令行没有提供组名，交互式获取
	if groupName == "" {
		fmt.Print("Group Name: ")
		// color.Green("Group Name: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			color.Red("Read input failed: %v", err)
			return
		}
		groupName = strings.TrimSpace(input)
		if groupName == "" {
			color.Red("Group name cannot be empty")
			return
		}
	} else {
		fmt.Println("Group Name: " + groupName)
	}

	// 如果命令行没有提供节点列表，交互式获取
	if groupNodes == "" {
		//灰色的提示
		fmt.Println("\nplease input nodes list, support the following formats:")
		fmt.Println("  - single IP: 192.168.1.1")
		fmt.Println("  - IP range: 192.168.1.1-5")
		fmt.Println("  - mixed format: 192.168.1.1-5,192.168.1.10,10.0.0.1")
		fmt.Print("Nodes: ")

		//fmt.Print("节点列表 [逗号分隔或范围，例如 192.168.1.1-5,192.168.1.10]: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			color.Red("读取输入失败: %v", err)
			return
		}
		groupNodes = strings.TrimSpace(input)
		if groupNodes == "" {
			color.Red("Nodes list cannot be empty")
			return
		}
	} else {
		fmt.Println("Nodes: " + groupNodes)
	}

	// 交互式获取端口
	if !cmd.Flags().Changed("port") {
		fmt.Printf("Port [%d]: ", groupPort)
		input, err := reader.ReadString('\n')
		if err != nil {
			color.Red("Read input failed: %v", err)
			return
		}
		input = strings.TrimSpace(input)
		if input != "" {
			fmt.Sscanf(input, "%d", &groupPort)
		}
	}

	// 交互式获取用户名
	if !cmd.Flags().Changed("user") {
		fmt.Printf("User [%s]: ", groupUser)
		input, err := reader.ReadString('\n')
		if err != nil {
			color.Red("Read input failed: %v", err)
			return
		}
		input = strings.TrimSpace(input)
		if input != "" {
			groupUser = input
		}
	}

	// 获取密码
	var password string
	//密码不能为空
	if !groupPassword {
		fmt.Print("Password: ")
		bytePwd, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			color.Red("Read password failed: %v", err)
			return
		}
		password = string(bytePwd)
	} else {
		//
		fmt.Print("Password: ******")
	}

	// 交互式获取描述
	if !cmd.Flags().Changed("description") {
		fmt.Print("Description: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			color.Red("Read input failed: %v", err)
			return
		}
		groupDescription = strings.TrimSpace(input)
	}

	// 解析节点列表
	ips, err := utils.ParseIPList(groupNodes)
	if err != nil {
		color.Red("Parse nodes list failed: %v", err)
		return
	}

	// 创建组
	err = crud.AddGroup(groupName, groupDescription, "default", false)
	if err != nil {
		color.Red("Create group failed: %v", err)
		return
	}

	// 显示结果
	fmt.Println("Verifying connection...")
	// 添加节点到组
	results, err := crud.AddOrUpdateNodes(groupName, ips, groupPort, groupUser, password, groupDescription)
	if err != nil {
		color.Red("Add nodes to group failed: %v", err)
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

	color.Green("✓ Successfully connected to %d/%d nodes", successCount, len(ips))

	for errMsg, failedIPs := range failureMap {
		color.Red("! Connection failed: %s", errMsg)
		for _, ip := range failedIPs {
			fmt.Printf("  - %s\n", ip)
		}
	}

	color.Green("Group '%s' created, contains %d nodes", groupName, successCount)
}

// groupRemoveFunc 删除组
func groupRemoveFunc(cmd *cobra.Command, args []string) {

	//args
	if len(args) > 0 {
		groupName = args[0]
	}

	reader := bufio.NewReader(os.Stdin)

	// 如果命令行没有提供组名，交互式获取
	if groupName == "" {
		fmt.Print("Group Name: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			color.Red("Read input failed: %v", err)
			return
		}
		groupName = strings.TrimSpace(input)
		if groupName == "" {
			color.Red("Group Name cannot be empty")
			return
		}
	}

	// 确认删除
	if !groupForce {
		fmt.Printf("Are you sure you want to delete group '%s'? [y/N]: ", groupName)
		input, err := reader.ReadString('\n')
		if err != nil {
			color.Red("Read input failed: %v", err)
			return
		}
		input = strings.ToLower(strings.TrimSpace(input))
		if input != "y" && input != "yes" {
			fmt.Println("Operation cancelled")
			return
		}
	}

	// 删除组
	err := crud.RemoveGroup(groupName)
	if err != nil {
		color.Red("Delete group failed: %v", err)
		return
	}

	color.Green("Group '%s' deleted", groupName)
}

// groupShowFunc 显示组详情
func groupShowFunc(cmd *cobra.Command, args []string) {

	if len(args) > 0 {
		groupName = args[0]
	}

	reader := bufio.NewReader(os.Stdin)

	// 如果命令行没有提供组名，交互式获取
	if groupName == "" {
		fmt.Print("Group Name: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			color.Red("Read input failed: %v", err)
			return
		}
		groupName = strings.TrimSpace(input)
		if groupName == "" {
			color.Red("Group Name cannot be empty")
			return
		}
	}

	// 获取组信息
	group, err := crud.GetGroup(groupName)
	if err != nil {
		color.Red("Get group info failed: %v", err)
		return
	}

	// 获取组中的节点
	nodes, err := crud.GetNodesInGroup(groupName)
	if err != nil {
		color.Red("Get nodes info failed: %v", err)
		return
	}

	// 显示组信息
	color.Green("Group: %s", group.Name)
	fmt.Printf("Description: %s\n", group.Description)
	fmt.Printf("Created at: %s\n", group.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated at: %s\n", group.UpdatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Node count: %d\n", len(nodes))

	if len(nodes) > 0 {
		fmt.Println("\nNodes list:")
		fmt.Println("----------------------------------------")
		for i, node := range nodes {
			statusSymbol := "✓"
			statusColor := color.New(color.FgGreen)
			if !node.Usable {
				statusSymbol = "✗"
				statusColor = color.New(color.FgRed)
			}

			statusColor.Printf("%s %s", statusSymbol, node.IP)
			fmt.Printf(" (Port: %d, User: %s)\n", node.Port, node.User)

			if i < len(nodes)-1 {
				fmt.Println("----------------------------------------")
			}
		}
	}
}
