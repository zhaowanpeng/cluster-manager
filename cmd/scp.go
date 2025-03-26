package cmd

import (
	"fmt"
	"zhaowanpeng/cluster-manager/model"

	"github.com/spf13/cobra"
)

var (
	scpGroup     string
	scpNodes     string
	scpTimeout   int
	scpRecursive bool
)

var scpCmd = &cobra.Command{
	Use:   "scp",
	Short: "文件分发",
	Long:  "Example：\n  clush scp -g g1 ",
	Run:   scpFunc,
}

func scpFunc(cmd *cobra.Command, args []string) {
	if (scpGroup == "" && scpNodes == "") || (scpGroup != "" && scpNodes != "") {
		fmt.Println("必须指定 -g 或 -n 其中之一")
		return
	}

	var nodes []model.ShellClient

	if scpGroup != "" {
		err := model.DB.Where("`group` = ? AND `usable` = ?", scpGroup, true).Find(&nodes).Error
		if err != nil {
			fmt.Printf("查询分组 %s 失败: %v\n", scpGroup, err)
			return
		}
		if len(nodes) == 0 {
			fmt.Printf("分组 %s 没有找到节点\n", scpGroup)
			return
		}
	}

	if scpNodes != "" {

		fmt.Println("user")
		fmt.Println("pwd")
		fmt.Println("port")
	}

	for _, clt := range nodes {
		fmt.Println(clt.IP)
	}

	if scpRecursive {
		//文件夹
	} else {
		//文件
	}

	// if scpNodes != "" {
	// 	ips, err := utils.ParseIPList(scpNodes)
	// 	if err != nil {
	// 		fmt.Printf("解析节点 IP 失败: %v\n", err)
	// 		return
	// 	}
	// 	if len(ips) == 0 {
	// 		fmt.Println("没有解析到有效的 IP")
	// 		return
	// 	}
	// 	for _, ip := range ips {
	// 		node := model.ShellClient{
	// 			IP:       ip,
	// 			Port:     scpPort,
	// 			User:     scpUser,
	// 			Password: pwd,
	// 		}
	// 		nodes = append(nodes, node)
	// 	}
	// }

	// if len(args) > 0 {
	// 	fmt.Println("单次模式")
	// } else {
	// 	interactiveMode(nodes)
	// }

}

func init() {
	scpCmd.Flags().StringVarP(&scpGroup, "group", "g", "", "指定分组，2选1")
	scpCmd.Flags().StringVarP(&scpNodes, "nodes", "n", "", "指定节点，2选1")
	scpCmd.Flags().IntVarP(&scpTimeout, "timeout", "t", 20, "执行命令超时时间，单位秒 (默认20)")
	scpCmd.Flags().BoolVarP(&scpRecursive, "recursive", "r", false, "Copy directories recursively")

}
