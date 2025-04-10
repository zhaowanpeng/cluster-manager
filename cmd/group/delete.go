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
	groupDeleteName  string
	groupDeleteForce bool
)

var groupDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "删除组",
	Run:   groupDeleteFunc,
}

func init() {
	groupDeleteCmd.Flags().StringVarP(&groupDeleteName, "name", "n", "", "组名称")
	groupDeleteCmd.Flags().BoolVarP(&groupDeleteForce, "force", "f", false, "强制删除")
}

func groupDeleteFunc(cmd *cobra.Command, args []string) {
	//args
	if len(args) > 0 {
		groupDeleteName = args[0]
	}

	reader := bufio.NewReader(os.Stdin)

	// 如果命令行没有提供组名，交互式获取
	if groupDeleteName == "" {
		fmt.Print("Group Name: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			color.Red("Read input failed: %v", err)
			return
		}
		groupName = strings.TrimSpace(input)
		if groupName == "" {
			color.Red("Group Name cannot be empty")
			return
		}
	}

	// 确认删除
	if !groupDeleteForce {
		fmt.Printf("Are you sure you want to delete group '%s'? [y/N]: ", groupDeleteName)
		input, err := reader.ReadString('\n')
		if err != nil {
			color.Red("Read input failed: %v", err)
			return
		}
		input = strings.ToLower(strings.TrimSpace(input))
		if input != "y" && input != "yes" {
			fmt.Println("Operation cancelled")
			return
		}
	}

	// 删除组
	err := crud.RemoveGroup(groupDeleteName)
	if err != nil {
		color.Red("Delete group failed: %v", err)
		return
	}

	color.Green("Group '%s' deleted", groupDeleteName)
}
