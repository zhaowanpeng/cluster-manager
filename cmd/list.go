package cmd

import (
	"fmt"
	"zhaowanpeng/cluster-manager/internal/utils"
	"zhaowanpeng/cluster-manager/model"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "查看分组列表",
	Long:  "clush list",
	Run:   listFunc,
}

func listFunc(cmd *cobra.Command, args []string) {
	var groups []string
	if err := model.DB.Model(&model.ShellClient{}).Distinct().Pluck("`group`", &groups).Error; err != nil {
		fmt.Printf("获取分组失败: %v\n", err)
		return
	}
	for _, grp := range groups {
		var clients []model.ShellClient
		model.DB.Where("`group` = ?", grp).Find(&clients)
		var ipList []string
		for _, c := range clients {
			ipList = append(ipList, c.IP)
		}
		compressed := utils.CompressIPs(ipList)
		fmt.Printf("%s(%s)\n", grp, compressed)
	}

}
