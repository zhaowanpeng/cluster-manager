package cmd

import (
	"fmt"
	"os"
	"sync"
	"time"

	"zhaowanpeng/cluster-manager/internal/types"
	"zhaowanpeng/cluster-manager/internal/utils"
	"zhaowanpeng/cluster-manager/model"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	addGroup       string
	addNodes       string
	addPort        int
	addUser        string
	addPassword    bool
	addDesc        string
	timeoutSeconds int
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "将节点添加到分组",
	Long:  "Example：\n  clush add -g group1 -n 192.168.108.1-100 -o 22 -u root -p",
	Run:   addFunc,
}

func addFunc(cmd *cobra.Command, args []string) {

	if addGroup == "" || addNodes == "" {
		fmt.Println("必须指定 -g (group) 和 -n (nodes)")
		return
	}

	if !addPassword {
		fmt.Println("请加 -p 参数来提示输入密码")
		return
	}

	fmt.Print("input password: ")
	bytePwd, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		fmt.Printf("读取密码失败: %v\n", err)
		return
	}
	pwd := string(bytePwd)

	// 解析 IP（支持单个 IP 或 IP 范围，如 "192.168.108.1-100"）
	ips, err := utils.ParseIPList(addNodes)
	if err != nil {
		fmt.Printf("解析节点 IP 失败: %v\n", err)
		return
	}

	var wg sync.WaitGroup
	resultChan := make(chan types.Result, len(ips))

	for _, ip := range ips {

		var client model.ShellClient
		res := model.DB.Where("`group` = ? AND ip = ? AND port = ? AND user = ?", addGroup, ip, addPort).First(&client)
		now := time.Now()
		//新增
		if res.Error != nil {
			newClient := model.ShellClient{
				ID:          fmt.Sprintf("%d", time.Now().UnixNano()),
				IP:          ip,
				Port:        addPort,
				User:        addUser,
				Password:    pwd,
				Group:       addGroup,
				AddAt:       now,
				Usable:      false,
				Tmp:         false,
				Description: addDesc,
			}
			model.DB.Create(&newClient)
		} else {
			//修改
			client.Password = pwd
			client.AddAt = now
			model.DB.Save(&client)
		}

		wg.Add(1)

		// 启动协程模拟登录验证
		go func(ip string) {
			defer wg.Done()
			// 模拟验证延时

			client, status := utils.SSH_Check(ip, addPort, addUser, pwd, time.Duration(timeoutSeconds)*time.Second)

			resultChan <- types.Result{IP: ip, Msg: status}

			if client != nil {
				model.DB.Model(&model.ShellClient{}).
					Where("ip = ? AND port = ? AND user = ? AND `group` = ?", ip, addPort, addUser, addGroup).
					Update("usable", true)
				fmt.Println(ip, "success")
			} else {
				fmt.Println(ip, status)
			}

		}(ip)
	}

	wg.Wait()
	close(resultChan)

	fmt.Println("所有节点添加完毕")
}

func init() {
	// fmt.Println("do add init")
	addCmd.Flags().StringVarP(&addGroup, "group", "g", "", "指定操作的组 (必填)")
	addCmd.Flags().StringVarP(&addNodes, "nodes", "n", "", "指定主机节点 (必填)，支持 IP 范围格式，如 192.168.108.1-100")
	addCmd.Flags().IntVarP(&addPort, "port", "o", 22, "指定端口 (默认 22)")
	addCmd.Flags().StringVarP(&addUser, "user", "u", "root", "指定用户 (默认 root)")
	addCmd.Flags().StringVarP(&addDesc, "description", "d", "", "说明信息")
	addCmd.Flags().IntVarP(&timeoutSeconds, "timeout", "t", 50, "等待时间，单位秒 (默认 20)")
	addCmd.Flags().BoolVarP(&addPassword, "password", "p", false, "提示输入密码")
}
