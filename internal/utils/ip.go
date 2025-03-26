package utils

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// ParseIPList 将类似 "192.168.108.1-100,192.168.108.200" 的字符串解析为 IP 列表
func ParseIPList(ipStr string) ([]string, error) {
	var ips []string
	parts := strings.Split(ipStr, ",") //192.168.108.1-100,192.168.108.200

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		// 判断是否包含范围
		if strings.Contains(part, "-") { //192.168.108.1-100
			octets := strings.Split(part, ".")
			if len(octets) != 4 {
				return nil, fmt.Errorf("无效的 IP 格式: %s", part)
			}
			// 目前只支持对第四段进行范围扩展，例如 "192.168.108.1-100"
			if strings.Contains(octets[3], "-") {
				rangeParts := strings.Split(octets[3], "-")
				if len(rangeParts) != 2 {
					return nil, fmt.Errorf("无效的范围格式: %s", part)
				}
				start, err := strconv.Atoi(rangeParts[0])
				if err != nil {
					return nil, fmt.Errorf("起始范围无效: %s", rangeParts[0])
				}
				end, err := strconv.Atoi(rangeParts[1])
				if err != nil {
					return nil, fmt.Errorf("结束范围无效: %s", rangeParts[1])
				}
				if start > end {
					return nil, fmt.Errorf("起始范围大于结束范围: %s", part)
				}
				prefix := strings.Join(octets[0:3], ".")
				for i := start; i <= end; i++ {
					ips = append(ips, fmt.Sprintf("%s.%d", prefix, i))
				}
			} else {
				ips = append(ips, part)
			}
		} else {
			ips = append(ips, part)
		}
	}

	return ips, nil
}

// ipToInt 将 IPv4 地址转换为整数，便于比较和排序
func ipToInt(ipStr string) int {
	parts := strings.Split(ipStr, ".")
	if len(parts) != 4 {
		return 0
	}
	a, _ := strconv.Atoi(parts[0])
	b, _ := strconv.Atoi(parts[1])
	c, _ := strconv.Atoi(parts[2])
	d, _ := strconv.Atoi(parts[3])
	return a<<24 | b<<16 | c<<8 | d
}

// formatRange 格式化连续的 IP 段，如 "192.168.1.1" 到 "192.168.1.5" 显示为 "192.168.1.1-5"
func formatRange(start, end string) string {
	if start == end {
		return start
	}
	parts := strings.Split(start, ".")
	lastStart, _ := strconv.Atoi(strings.Split(start, ".")[3])
	lastEnd, _ := strconv.Atoi(strings.Split(end, ".")[3])
	return fmt.Sprintf("%s.%s.%s.%d-%d", parts[0], parts[1], parts[2], lastStart, lastEnd)
}

// CompressIPs 对 IP 列表进行排序并压缩显示成范围形式
func CompressIPs(ips []string) string {
	if len(ips) == 0 {
		return ""
	}
	sort.Slice(ips, func(i, j int) bool {
		return ipToInt(ips[i]) < ipToInt(ips[j])
	})
	var ranges []string
	start := ips[0]
	prev := ips[0]
	for i := 1; i < len(ips); i++ {
		current := ips[i]
		if ipToInt(current) == ipToInt(prev)+1 {
			prev = current
		} else {
			ranges = append(ranges, formatRange(start, prev))
			start = current
			prev = current
		}
	}
	ranges = append(ranges, formatRange(start, prev))
	return strings.Join(ranges, ",")
}
