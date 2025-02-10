package main

import (
	"zhaowanpeng/cluster-manager/cmd"
	"zhaowanpeng/cluster-manager/model"
)

func main() {
	model.InitDB()
	cmd.Execute()
}
