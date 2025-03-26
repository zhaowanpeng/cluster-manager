package types

// Result 表示操作结果
type Result struct {
	IP      string `json:"ip"`
	Msg     string `json:"msg"`
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
}
