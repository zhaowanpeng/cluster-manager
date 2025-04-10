package group

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"zhaowanpeng/cluster-manager/internal/crud"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	groupShowName string
)

var groupShowCmd = &cobra.Command{
	Use:   "show",
	Short: "显示组",
	Run:   groupShowFunc,
}

func init() {
	groupShowCmd.Flags().StringVarP(&groupShowName, "name", "n", "", "组名称")
}

func groupShowFunc(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		groupShowName = args[0]
	}

	reader := bufio.NewReader(os.Stdin)

	// 如果命令行没有提供组名，交互式获取
	if groupShowName == "" {
		fmt.Print("Group Name: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			color.Red("Read input failed: %v", err)
			return
		}
		groupShowName = strings.TrimSpace(input)
		if groupShowName == "" {
			color.Red("Group Name cannot be empty")
			return
		}
	}

	// 获取组信息
	group, err := crud.GetGroup(groupShowName)
	if err != nil {
		color.Red("Get group info failed: %v", err)
		return
	}

	// 获取组中的节点
	nodes, err := crud.GetNodesInGroup(groupShowName)
	if err != nil {
		color.Red("Get nodes info failed: %v", err)
		return
	}

	// 显示组信息
	color.Green("Group: %s", group.Name)
	fmt.Printf("Description: %s\n", group.Description)
	fmt.Printf("Created at: %s\n", group.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated at: %s\n", group.UpdatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Node count: %d\n", len(nodes))

	if len(nodes) > 0 {
		fmt.Println("\nNodes list:")
		fmt.Println("----------------------------------------")
		for i, node := range nodes {
			statusSymbol := "✓"
			statusColor := color.New(color.FgGreen)
			if !node.Usable {
				statusSymbol = "✗"
				statusColor = color.New(color.FgRed)
			}

			statusColor.Printf("%s %s", statusSymbol, node.IP)
			fmt.Printf(" (Port: %d, User: %s)\n", node.Port, node.User)

			if i < len(nodes)-1 {
				fmt.Println("----------------------------------------")
			}
		}
	}
}
