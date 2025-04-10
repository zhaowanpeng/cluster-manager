package cmd

import (
	"fmt"
	"os"

	"zhaowanpeng/cluster-manager/cmd/group"

	"github.com/spf13/cobra"
)

// rootCmd 是整个应用的根命令
var rootCmd = &cobra.Command{
	Use:   "talko",
	Short: "智能Shell助手与集群管理工具",
	Long: `
OctoShell (octosh) - 智能化多节点管理工具
如同章鱼通过一个大脑控制多个触手，OctoShell让您可以同时控制多个远程节点，
高效完成集群命令执行、文件分发、状态监控等运维工作。`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("请使用 --help 查看可用命令")
	},
	Version: "0.0.1",
	// 禁用自动生成的命令
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

// Execute 执行根命令
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// rootCmd.AddCommand(addCmd)
	// rootCmd.AddCommand(listCmd)
	// rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(group.GroupCmd)
	// rootCmd.AddCommand(execCmd)
	// rootCmd.AddCommand(scpCmd)

}
