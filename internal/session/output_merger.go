package session

import (
	"fmt"
	"strings"
	"zhaowanpeng/cluster-manager/internal/utils/ip_util"
	"zhaowanpeng/cluster-manager/model"

	"github.com/fatih/color"
)

// displayMergedResults 显示合并后的命令执行结果
func displayMergedResults(nodesBySubnet map[string][]model.Node, results map[string]ExecResult) {
	for subnet, nodes := range nodesBySubnet {
		fmt.Print("\n")
		color.New(color.FgHiCyan).Printf("%s:\n", subnet)

		// 按输出内容分组
		successOutputGroups := make(map[string][]string) // 输出 -> IP列表
		errorGroups := make(map[string][]string)         // 错误信息 -> IP列表
		errorOutputGroups := make(map[string][]string)   // 错误输出 -> IP列表

		for _, node := range nodes {
			result, ok := results[node.IP]
			if !ok {
				continue
			}

			if !result.Success {
				errorMsg := result.Error.Error()
				errorGroups[errorMsg] = append(errorGroups[errorMsg], node.IP)

				if result.Output != "" {
					errorOutputGroups[result.Output] = append(errorOutputGroups[result.Output], node.IP)
				}
			} else {
				successOutputGroups[result.Output] = append(successOutputGroups[result.Output], node.IP)
			}
		}

		// 显示错误
		for errorMsg, ips := range errorGroups {
			displayNodeGroup(ips, color.FgRed, "错误")
			fmt.Printf("  %s\n", errorMsg)
		}

		// 显示错误输出
		for output, ips := range errorOutputGroups {
			displayNodeGroup(ips, color.FgYellow, "错误输出")
			fmt.Printf("  %s\n", output)
		}

		// 显示成功输出
		for output, ips := range successOutputGroups {
			displayNodeGroup(ips, color.FgGreen, "")
			if output != "" {
				fmt.Printf("  %s\n", output)
			}
		}
	}
	fmt.Print("\n")
}

// displayNodeGroup 显示节点组
func displayNodeGroup(ips []string, textColor color.Attribute, label string) {
	c := color.New(textColor)

	if len(ips) > 5 {
		// 显示前3个IP和总数
		ipDisplay := fmt.Sprintf("%s,%s,%s等%d个节点", ips[0], ips[1], ips[2], len(ips))
		if label != "" {
			c.Printf("[%s] %s:\n", ipDisplay, label)
		} else {
			c.Printf("[%s]\n", ipDisplay)
		}
	} else if len(ips) > 1 {
		// 显示所有IP，用逗号分隔
		ipDisplay := strings.Join(ips, ",")
		if label != "" {
			c.Printf("[%s] %s:\n", ipDisplay, label)
		} else {
			c.Printf("[%s]\n", ipDisplay)
		}
	} else if len(ips) == 1 {
		// 只有一个IP
		if label != "" {
			c.Printf("[%s] %s:\n", ips[0], label)
		} else {
			c.Printf("[%s]\n", ips[0])
		}
	}
}

// CompressIPList 将IP列表压缩为更紧凑的表示
func CompressIPList(ips []string) string {
	if len(ips) == 0 {
		return ""
	}

	// 使用IP工具压缩IP列表
	compressed, err := ip_util.CompressIPList(ips)
	if err != nil {
		// 如果压缩失败，回退到简单的表示方法
		if len(ips) > 3 {
			return fmt.Sprintf("%s等%d个节点", ips[0], len(ips))
		}
		return strings.Join(ips, ",")
	}

	return compressed
}
