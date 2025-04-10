package crud

import (
	"fmt"
	"sync"
	"time"
	"zhaowanpeng/cluster-manager/internal/types"
	"zhaowanpeng/cluster-manager/internal/utils"
	"zhaowanpeng/cluster-manager/model"
)

// CountNodesInGroup 统计组中的节点数量
func CountNodesInGroup(groupName string) (int, error) {
	var count int64
	result := model.DB.Model(&model.Node{}).Where("`group` = ?", groupName).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}
	return int(count), nil
}

// GetNodesInGroup 获取组中的所有节点
func GetNodesInGroup(groupName string) ([]model.Node, error) {
	var nodes []model.Node
	result := model.DB.Where("`group` = ?", groupName).Find(&nodes)
	if result.Error != nil {
		return nil, result.Error
	}
	return nodes, nil
}

// AddNodesToGroup 添加节点到组
func AddOrUpdateNodes(groupName string, ips []string, port int, user, password, description string) ([]types.Result, error) {
	// 检查组是否存在
	var group model.Group
	result := model.DB.Where("`name` = ?", groupName).First(&group)
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("group '%s' not found", groupName)
	}

	var wg sync.WaitGroup
	// 结果集通道, make中len(ips)代表最大容量
	resultChan := make(chan types.Result, len(ips))

	// 结果集, make中0代表初始容量, len(ips)代表最大容量
	results := make([]types.Result, 0, len(ips))

	// 为每个 IP 创建一个 goroutine
	for _, ip := range ips {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()

			// 检查节点是否已存在
			var existingNode model.Node
			result := model.DB.Where("`group` = ? AND ip = ?", groupName, ip).First(&existingNode)

			now := time.Now()

			// 检查 SSH 连接
			sshClient, status := utils.SSH_Check(
				ip,
				port,
				user,
				password,
				30*time.Second,
			)

			isConnected := sshClient != nil
			if sshClient != nil {
				sshClient.Close() // 确保关闭客户端
			}

			// 如果节点已存在，更新它
			if result.RowsAffected > 0 {
				existingNode.Password = password
				existingNode.LastCheckAt = now
				existingNode.Usable = isConnected
				existingNode.Description = description

				model.DB.Save(&existingNode)

				resultChan <- types.Result{
					IP:      ip,
					Msg:     status,
					Success: isConnected,
				}
				return
			}

			// 创建新节点
			newNode := model.Node{
				ID:          fmt.Sprintf("%s-%s", groupName, ip),
				IP:          ip,
				Port:        port,
				User:        user,
				Password:    password,
				Group:       groupName,
				AddAt:       now,
				LastCheckAt: now,
				Usable:      isConnected,
				Description: description,
			}

			result = model.DB.Create(&newNode)

			if result.Error != nil {
				resultChan <- types.Result{
					IP:      ip,
					Msg:     fmt.Sprintf("Database error: %v", result.Error),
					Success: false,
				}

			} else {
				resultChan <- types.Result{
					IP:      ip,
					Msg:     status,
					Success: true,
				}
			}

		}(ip)
	}

	// 等待所有 goroutine 完成
	wg.Wait()
	close(resultChan)

	// 收集结果
	for result := range resultChan {
		results = append(results, result)
	}

	return results, nil
}
