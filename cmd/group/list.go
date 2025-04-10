package group

import (
	"fmt"
	"zhaowanpeng/cluster-manager/internal/crud"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var groupListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出组",
	Run:   groupListFunc,
}

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
