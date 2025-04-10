package group

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"
	"zhaowanpeng/cluster-manager/internal/session"
	"zhaowanpeng/cluster-manager/internal/utils/ip_util"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	execGroupName    string
	execTimeout      int
	execExcludeNodes string
	execAddNodes     string
	execMergeOutput  bool
	execPort         int
	execUser         string
	execPassword     string
)

var groupExecCmd = &cobra.Command{
	Use:   "exec [group-name]",
	Short: "在组上执行命令",
	Long:  "在指定组的所有节点上执行命令，支持交互式会话",
	Run:   execFunc,
}

func init() {
	execPort = 22
	execUser = "root"
	execPassword = ""
	groupExecCmd.Flags().StringVarP(&execGroupName, "name", "n", "", "组名称")
	groupExecCmd.Flags().IntVarP(&execTimeout, "timeout", "t", 60, "命令执行超时时间（秒）")
	groupExecCmd.Flags().StringVarP(&execExcludeNodes, "exclude", "e", "", "排除节点，支持范围表示法，如 192.168.1.1-5,192.168.1.10")
	groupExecCmd.Flags().StringVarP(&execAddNodes, "add", "a", "", "额外添加节点，支持范围表示法")
	groupExecCmd.Flags().BoolVarP(&execMergeOutput, "merge", "m", false, "合并相同输出")
	// groupExecCmd.Flags().IntVarP(&execPort, "port", "p", 22, "SSH端口（用于额外添加的节点）")
	// groupExecCmd.Flags().StringVarP(&execUser, "user", "u", "root", "SSH用户名（用于额外添加的节点）")
	// groupExecCmd.Flags().StringVarP(&execPassword, "password", "P", "", "SSH密码（用于额外添加的节点）")
}

func execFunc(cmd *cobra.Command, args []string) {
	// 如果命令行参数提供了组名，优先使用
	if len(args) > 0 {
		execGroupName = args[0]
	}

	// 如果没有提供组名，显示错误
	if execGroupName == "" {
		color.Red("请提供组名称")
		return
	}

	// 如果添加了节点，进入会话提示输入port，账户，密码
	if execAddNodes != "" {
		_, err := ip_util.ParseIPRange(execAddNodes)
		if err != nil {
			color.Red("解析添加节点失败: %v", err)
			return
		}
		reader := bufio.NewReader(os.Stdin)
		// 交互式获取端口
		fmt.Printf("Port [%d]: ", execPort)
		input, err := reader.ReadString('\n')
		if err != nil {
			color.Red("Read input failed: %v", err)
			return
		}
		input = strings.TrimSpace(input)
		if input != "" {
			fmt.Sscanf(input, "%d", &execPort)
		}

		// 交互式获取用户名
		fmt.Printf("User [%s]: ", execUser)
		input, err = reader.ReadString('\n')
		if err != nil {
			color.Red("Read input failed: %v", err)
			return
		}
		input = strings.TrimSpace(input)
		if input != "" {
			execUser = input
		}

		// 获取密码
		fmt.Print("Password: ")
		bytePwd, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			color.Red("Read password failed: %v", err)
			return
		}
		execPassword = string(bytePwd)

	}

	// 提示用户设置的超时时间
	color.Cyan("命令执行超时设置为 %d 秒", execTimeout)

	// 创建执行选项
	options := session.ExecOptions{
		GroupName:    execGroupName,
		Timeout:      time.Duration(execTimeout) * time.Second,
		ExcludeNodes: execExcludeNodes,
		AddNodes:     execAddNodes,
		MergeOutput:  execMergeOutput,
		Port:         execPort,
		User:         execUser,
		Password:     execPassword,
	}

	// 启动组执行会话
	err := session.StartGroupExec(options)
	if err != nil {
		color.Red("执行失败: %v", err)
		return
	}
}
