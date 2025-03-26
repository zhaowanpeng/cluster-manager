package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd 是整个应用的根命令
var rootCmd = &cobra.Command{
	Use:   "go-clush",
	Short: "A parallel command tool for multi-host operations",
	Long:  "go-clush 是一个用于在多个远程节点上并行执行命令、拷贝文件等操作的命令行工具，仿照 clush 的风格设计。",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("请使用 --help 查看可用命令")
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
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(execCmd)
	rootCmd.AddCommand(scpCmd)
	rootCmd.AddCommand(groupCmd)
}
