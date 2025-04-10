package node

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	nodeAddGroupName string
	nodeAddNodes     string
)

var nodeAddCmd = &cobra.Command{
	Use:   "add",
	Short: "添加节点",
	Run:   nodeAddFunc,
}

func init() {
	nodeAddCmd.Flags().StringVarP(&nodeAddGroupName, "group", "g", "", "组名称")
	nodeAddCmd.Flags().StringVarP(&nodeAddNodes, "nodes", "N", "", "节点列表")
}

func nodeAddFunc(cmd *cobra.Command, args []string) {
	fmt.Println("添加节点,开发中...")
}
