package session

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
		fmt.Print("密码: ")
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

	color.Green("组 '%s' 已创建，包含 %d 个节点", groupName, successCount)
}
