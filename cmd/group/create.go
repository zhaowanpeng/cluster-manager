package group

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"zhaowanpeng/cluster-manager/internal/crud"
	"zhaowanpeng/cluster-manager/internal/utils/ip_util"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	groupName        string
	groupNodes       string
	groupPort        int
	groupUser        string
	groupPassword    bool
	groupDescription string
)
var groupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "创建节点组",
	Run:   groupCreateFunc,
}

func init() {
	groupCreateCmd.Flags().StringVarP(&groupName, "name", "n", "", "组名称")
	groupCreateCmd.Flags().StringVarP(&groupNodes, "nodes", "N", "", "节点列表")
	groupCreateCmd.Flags().IntVarP(&groupPort, "port", "p", 22, "端口")
	groupCreateCmd.Flags().StringVarP(&groupUser, "user", "u", "root", "用户名")
	groupCreateCmd.Flags().BoolVarP(&groupPassword, "password", "P", false, "是否使用密码")
	groupCreateCmd.Flags().StringVarP(&groupDescription, "description", "d", "", "组描述")
}

func groupCreateFunc(cmd *cobra.Command, args []string) {
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
	ips, err := ip_util.ParseIPRange(groupNodes)
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
