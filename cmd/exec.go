package cmd

import (
	"fmt"
	"strings"
	"sync"
	"time"
	"zhaowanpeng/cluster-manager/internal/utils"
	"zhaowanpeng/cluster-manager/model"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
)

var (
	execGroup   string
	execNodes   string
	execTimeout int
)

var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "命令分发",
	Long:  "Example：\n  clush exec -g g1 ",
	Run:   execFunc,
}

func execFunc(cmd *cobra.Command, args []string) {
	if (execGroup == "" && execNodes == "") || (execGroup != "" && execNodes != "") {
		fmt.Println("必须指定 -g 或 -n 其中之一")
		return
	}

	var nodes []model.ShellClient

	if execGroup != "" {
		err := model.DB.Where("`group` = ? AND `usable` = ?", execGroup, true).Find(&nodes).Error
		if err != nil {
			fmt.Printf("查询分组 %s 失败: %v\n", execGroup, err)
			return
		}
		if len(nodes) == 0 {
			fmt.Printf("分组 %s 没有找到节点\n", execGroup)
			return
		}
	}

	if execNodes != "" {

		fmt.Println("user")
		fmt.Println("pwd")
		fmt.Println("port")
	}

	for _, clt := range nodes {
		fmt.Println(clt.IP)
	}

	// if execNodes != "" {
	// 	ips, err := utils.ParseIPList(execNodes)
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
	// 			Port:     execPort,
	// 			User:     execUser,
	// 			Password: pwd,
	// 		}
	// 		nodes = append(nodes, node)
	// 	}
	// }

	if len(args) > 0 {
		fmt.Println("单次模式")
	} else {
		interactiveMode(nodes)
	}

}

func interactiveMode(nodes []model.ShellClient) {
	rl, err := readline.New(">>> ")
	if err != nil {
		fmt.Printf("创建 readline 失败: %v\n", err)
		return
	}
	defer rl.Close()

	// 自动保存历史记录功能（readline 默认支持上下翻阅历史记录）
	for {
		line, err := rl.Readline()
		if err != nil {
			fmt.Printf("bye ~")
			break
		}
		command := strings.TrimSpace(line)
		if command == "exit" {
			break
		}
		fmt.Println(command)
		if command == "ls -l" {
			executeCommandOnNodes(nodes, command)
		}

	}
}

func executeCommandOnNodes(nodes []model.ShellClient, command string) {

	var wg sync.WaitGroup
	// 为了方便显示，我们给每个节点生成 g1(), g2() 等标识
	for i, node := range nodes {
		wg.Add(1)
		go func(i int, node model.ShellClient) {
			defer wg.Done()
			label := fmt.Sprintf("g%d()", i+1)
			output, err := utils.Exec_SSH_Command(
				node.IP,
				node.Port,
				node.User,
				node.Password,
				command,
				time.Duration(execTimeout)*time.Second,
			)
			if err != nil {
				fmt.Printf("%s error: %v\n", label, err)
			} else {
				// 如输出中有换行，自动换行显示
				fmt.Printf("%s output:\n%s\n", label, output)
			}
		}(i, node)
	}
	wg.Wait()
}

func init() {
	execCmd.Flags().StringVarP(&execGroup, "group", "g", "", "指定分组，2选1")
	execCmd.Flags().StringVarP(&execNodes, "nodes", "n", "", "指定节点，2选1")
	execCmd.Flags().IntVarP(&execTimeout, "timeout", "t", 20, "执行命令超时时间，单位秒 (默认20)")
}
