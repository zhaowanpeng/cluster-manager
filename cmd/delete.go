package cmd

import (
	"fmt"
	"zhaowanpeng/cluster-manager/internal/utils"
	"zhaowanpeng/cluster-manager/model"

	"github.com/spf13/cobra"
)

var (
	delGroup string
	delUser  string
	delPort  int
	delNode  string
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "删除分组或删除分组中的部分记录",
	Long:  "Example：\n  clush delete -g group1 -n 192.168.108.1-100 -o 22 -u root -p",
	Args:  cobra.MaximumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {

		if delGroup == "" {
			fmt.Println("必须指定分组(-g)")
			return
		}

		query := model.DB.Where("`group` = ?", delGroup)
		if delUser != "" {
			query = query.Where("user = ?", delUser)
		}
		if delPort != 0 {
			query = query.Where("port = ?", delPort)
		}
		if delNode != "" {
			ips, err := utils.ParseIPList(delNode)
			if err != nil {
				fmt.Printf("解析节点 IP 失败: %v\n", err)
				return
			}
			query = query.Where("ip IN ?", ips)
		}
		res := query.Delete(&model.ShellClient{})
		fmt.Printf("删除了 %d 条记录\n", res.RowsAffected)
	},
}

func init() {
	deleteCmd.Flags().StringVarP(&delGroup, "group", "g", "", "分组名称 (必填)")
	deleteCmd.Flags().StringVarP(&delUser, "user", "u", "", "用户名 (可选)")
	deleteCmd.Flags().IntVarP(&delPort, "port", "o", 0, "端口号 (可选)")
	deleteCmd.Flags().StringVarP(&delNode, "node", "n", "", "节点 IP 或 IP 范围 (可选)")
}
