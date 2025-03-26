package crud

import (
	"fmt"
	"sync"
	"time"
	"zhaowanpeng/cluster-manager/internal/types"
	"zhaowanpeng/cluster-manager/internal/utils"
	"zhaowanpeng/cluster-manager/model"
)

// ListGroups 列出所有组
func ListGroups() ([]model.Group, error) {
	var groups []model.Group
	result := model.DB.Find(&groups)
	if result.Error != nil {
		return nil, result.Error
	}
	return groups, nil
}

// AddGroup 添加新组
func AddGroup(name, desc string) error {
	var group model.Group

	// 检查组是否已存在
	result := model.DB.Where("`name` = ?", name).First(&group)
	if result.RowsAffected > 0 {
		return fmt.Errorf("组 '%s' 已存在", name)
	}

	// 创建新组
	now := time.Now()
	newGroup := model.Group{
		Name:        name,
		Description: desc,
		CreatedAt:   now,
		UpdatedAt:   now,
		User:        "",
		Tmp:         false,
	}

	result = model.DB.Create(&newGroup)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// RemoveGroup 删除组
func RemoveGroup(name string) error {
	var group model.Group

	// 检查组是否存在
	result := model.DB.Where("`name` = ?", name).First(&group)
	if result.RowsAffected == 0 {
		return fmt.Errorf("组 '%s' 不存在", name)
	}

	// 删除组中的所有节点
	result = model.DB.Where("`group` = ?", name).Delete(&model.Node{})
	if result.Error != nil {
		return result.Error
	}

	// 删除组
	result = model.DB.Delete(&group)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// GetGroup 获取组信息
func GetGroup(name string) (model.Group, error) {
	var group model.Group

	result := model.DB.Where("`name` = ?", name).First(&group)
	if result.RowsAffected == 0 {
		return group, fmt.Errorf("组 '%s' 不存在", name)
	}

	return group, nil
}

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
func AddNodesToGroup(groupName string, ips []string, port int, user, password, description string) ([]types.Result, error) {
	// 检查组是否存在
	var group model.Group
	result := model.DB.Where("`name` = ?", groupName).First(&group)
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("组 '%s' 不存在", groupName)
	}

	var wg sync.WaitGroup
	resultChan := make(chan types.Result, len(ips))
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
				ID:          fmt.Sprintf("%d", time.Now().UnixNano()),
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

			model.DB.Create(&newNode)

			resultChan <- types.Result{
				IP:      ip,
				Msg:     status,
				Success: isConnected,
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
