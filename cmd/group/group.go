package group

import (
	"zhaowanpeng/cluster-manager/cmd/group/node"

	"github.com/spf13/cobra"
)

var GroupCmd = &cobra.Command{
	Use:   "group",
	Short: "管理节点组",
	Long:  "管理节点组，包括添加、删除、列出和编辑组",
}

func init() {
	// 添加子命令
	GroupCmd.AddCommand(groupCreateCmd)
	GroupCmd.AddCommand(groupDeleteCmd)
	GroupCmd.AddCommand(groupListCmd)
	GroupCmd.AddCommand(groupExecCmd)
	GroupCmd.AddCommand(groupShowCmd)

	GroupCmd.AddCommand(node.NodeCmd)
}
