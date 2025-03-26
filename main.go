package main

import (
	"fmt"
	"os"
	"zhaowanpeng/cluster-manager/cmd"
	"zhaowanpeng/cluster-manager/model"
)

func main() {
	// 初始化数据库
	err := model.InitDB()
	if err != nil {
		fmt.Printf("初始化数据库失败: %v\n", err)
		os.Exit(1)
	}

	// 执行命令
	cmd.Execute()
}
