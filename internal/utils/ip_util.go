package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"strings"
)

// ParseIPList 解析 IP 列表，支持范围表示法
func ParseIPList(ipList string) ([]string, error) {
	var ips []string

	// 按逗号分割
	parts := strings.Split(ipList, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// 检查是否是范围表示法
		if strings.Contains(part, "-") {
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
			} else {
				// 完整的结束 IP
				endIP := endRange
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
			}
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
