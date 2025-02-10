package types

import "time"

type ShellClient struct {
	ID       string
	IP       string
	Port     int
	User     string
	Password string
	Group    string
	AddAt    time.Time
	Usable   bool
}

/*
*spf13/viper
*spf13/cobra
 */

// type PassWord struct {
// 	ID      string
// 	Content string
// }

// type GroupInfo struct {
// 	ID   string
// 	Name string
// }

// type UserInfo struct {
// 	ID       string
// 	Name     string
// 	Password string
// 	Pub      string
// 	Node     string
// 	Deactive bool
// }

// type CmdGroup struct {
// }

// type CmdCallBack struct {
// 	Cmd  string
// 	Node string
// 	// StartAt  datetime
// 	// EndAt    datetime
// 	CallBack string
// }
