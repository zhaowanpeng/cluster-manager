package node

import (
	"github.com/spf13/cobra"
)

var NodeCmd = &cobra.Command{
	Use:   "node",
	Short: "节点管理",
	// Run:   nodeFunc,
}

func init() {
	NodeCmd.AddCommand(nodeAddCmd)
	NodeCmd.AddCommand(nodeRemoveCmd)
}
