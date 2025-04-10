package ip_util

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
)

// ParseIPRange 解析 IP 范围字符串，返回包含的所有 IP 地址
// 支持以下格式:
// - 单个IP: 192.168.1.1
// - IP范围(最后一段): 192.168.1.1-5
// - 完整IP范围: 192.168.1.1-192.168.1.5
// - 混合格式: 192.168.1.1-5,192.168.1.10,10.0.0.1
func ParseIPRange(ipRange string) ([]string, error) {
	var ips []string

	// 按逗号分割
	parts := strings.Split(ipRange, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// 检查是否是范围表示法
		if strings.Contains(part, "-") {
			rangeIPs, err := parseIPRangePart(part)
			if err != nil {
				return nil, err
			}
			ips = append(ips, rangeIPs...)
		} else {
			// 单个 IP
			if net.ParseIP(part) == nil {
				return nil, fmt.Errorf("无效的 IP 地址: %s", part)
			}
			ips = append(ips, part)
		}
	}

	return ips, nil
}

// parseIPRangePart 解析单个IP范围部分
func parseIPRangePart(part string) ([]string, error) {
	// var ips []string

	ipRange := strings.Split(part, "-")
	if len(ipRange) != 2 {
		return nil, fmt.Errorf("无效的 IP 范围: %s", part)
	}

	// 解析起始 IP
	startIP := ipRange[0]
	startIPParts := strings.Split(startIP, ".")
	if len(startIPParts) != 4 {
		return nil, fmt.Errorf("无效的起始 IP: %s", startIP)
	}

	// 解析结束范围
	endRange := ipRange[1]

	// 如果结束范围只是一个数字，则假定它是最后一个八位字节
	if !strings.Contains(endRange, ".") {
		return parseLastOctetRange(startIPParts, endRange)
	} else {
		// 完整的结束 IP
		return parseFullIPRange(startIP, endRange)
	}
}

// parseLastOctetRange 解析最后一个八位字节范围
func parseLastOctetRange(startIPParts []string, endRange string) ([]string, error) {
	var ips []string

	endNum, err := strconv.Atoi(endRange)
	if err != nil {
		return nil, fmt.Errorf("无效的结束范围: %s", endRange)
	}

	// 构建完整的结束 IP
	startNum, err := strconv.Atoi(startIPParts[3])
	if err != nil {
		return nil, fmt.Errorf("无效的起始 IP 最后一个八位字节: %s", startIPParts[3])
	}

	if endNum < startNum {
		return nil, fmt.Errorf("结束范围 (%d) 小于起始范围 (%d)", endNum, startNum)
	}

	// 生成 IP 范围
	for i := startNum; i <= endNum; i++ {
		ip := fmt.Sprintf("%s.%s.%s.%d", startIPParts[0], startIPParts[1], startIPParts[2], i)
		ips = append(ips, ip)
	}

	return ips, nil
}

// parseFullIPRange 解析完整的IP范围
func parseFullIPRange(startIP, endIP string) ([]string, error) {
	var ips []string

	endIPParts := strings.Split(endIP, ".")
	if len(endIPParts) != 4 {
		return nil, fmt.Errorf("无效的结束 IP: %s", endIP)
	}

	// 将 IP 转换为整数进行比较
	startIPInt, err := ipToInt(startIP)
	if err != nil {
		return nil, err
	}

	endIPInt, err := ipToInt(endIP)
	if err != nil {
		return nil, err
	}

	if endIPInt < startIPInt {
		return nil, fmt.Errorf("结束 IP (%s) 小于起始 IP (%s)", endIP, startIP)
	}

	// 生成 IP 范围
	for i := startIPInt; i <= endIPInt; i++ {
		ip := intToIP(i)
		ips = append(ips, ip)
	}

	return ips, nil
}

// CompressIPList 将IP列表压缩为紧凑的字符串表示
// 例如: ["192.168.1.1", "192.168.1.2", "192.168.1.3", "192.168.1.5", "10.0.0.1"]
// 将压缩为: "192.168.1.1-3,192.168.1.5,10.0.0.1"
func CompressIPList(ips []string) (string, error) {
	if len(ips) == 0 {
		return "", nil
	}

	// 验证所有IP并转换为整数
	ipInts := make([]uint32, 0, len(ips))
	ipMap := make(map[uint32]string)

	for _, ip := range ips {
		ipInt, err := ipToInt(ip)
		if err != nil {
			return "", err
		}
		ipInts = append(ipInts, ipInt)
		ipMap[ipInt] = ip
	}

	// 排序IP
	sort.Slice(ipInts, func(i, j int) bool {
		return ipInts[i] < ipInts[j]
	})

	// 查找连续范围
	var result []string
	start := 0

	for i := 1; i <= len(ipInts); i++ {
		// 检查是否需要结束当前范围
		if i == len(ipInts) || ipInts[i] != ipInts[i-1]+1 {
			// 处理范围
			if start == i-1 {
				// 单个IP
				result = append(result, ipMap[ipInts[start]])
			} else {
				// 检查是否可以使用简化表示法
				startIP := ipMap[ipInts[start]]
				endIP := ipMap[ipInts[i-1]]

				startParts := strings.Split(startIP, ".")
				endParts := strings.Split(endIP, ".")

				// 如果只有最后一个八位字节不同，使用简化表示法
				if startParts[0] == endParts[0] &&
					startParts[1] == endParts[1] &&
					startParts[2] == endParts[2] {
					result = append(result, fmt.Sprintf("%s-%s", startIP, endParts[3]))
				} else {
					// 否则使用完整表示法
					result = append(result, fmt.Sprintf("%s-%s", startIP, endIP))
				}
			}

			// 开始新范围
			if i < len(ipInts) {
				start = i
			}
		}
	}

	return strings.Join(result, ","), nil
}

// ipToInt 将 IP 地址转换为整数
func ipToInt(ipStr string) (uint32, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return 0, fmt.Errorf("无效的 IP 地址: %s", ipStr)
	}

	ip = ip.To4()
	if ip == nil {
		return 0, fmt.Errorf("不是 IPv4 地址: %s", ipStr)
	}

	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3]), nil
}

// intToIP 将整数转换为 IP 地址
func intToIP(ipInt uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		ipInt>>24,
		(ipInt>>16)&0xFF,
		(ipInt>>8)&0xFF,
		ipInt&0xFF)
}

// GenerateShortID 生成短 ID
func GenerateShortID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}
