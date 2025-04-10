package node

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	nodeRemoveGroupName string
	nodeRemoveNodes     string
	nodeRemoveForce     bool
)

var nodeRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "删除节点",
	Run:   nodeRemoveFunc,
}

func init() {
	nodeRemoveCmd.Flags().StringVarP(&nodeRemoveGroupName, "group", "g", "", "组名称")
	nodeRemoveCmd.Flags().StringVarP(&nodeRemoveNodes, "nodes", "N", "", "节点列表")
	nodeRemoveCmd.Flags().BoolVarP(&nodeRemoveForce, "force", "f", false, "强制删除")
}

func nodeRemoveFunc(cmd *cobra.Command, args []string) {
	fmt.Println("删除节点,开发中...")
}
