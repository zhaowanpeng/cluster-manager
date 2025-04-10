package crud

import (
	"fmt"
	"time"
	"zhaowanpeng/cluster-manager/model"
)

// AddGroup 添加新组
func AddGroup(name, desc, user string, tmp bool) error {
	var group model.Group

	// 检查组是否已存在
	result := model.DB.Where("`name` = ?", name).First(&group)
	if result.RowsAffected > 0 {
		return fmt.Errorf("group '%s' already exists", name)
	}

	// 创建新组
	now := time.Now()
	newGroup := model.Group{
		Name:        name,
		Description: desc,
		CreatedAt:   now,
		UpdatedAt:   now,
		User:        user,
		Tmp:         tmp,
	}

	result = model.DB.Create(&newGroup)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// ListGroups 列出所有组
func ListGroups() ([]model.Group, error) {
	var groups []model.Group
	result := model.DB.Find(&groups)
	if result.Error != nil {
		return nil, result.Error
	}
	return groups, nil
}

// RemoveGroup 删除组
func RemoveGroup(name string) error {
	var group model.Group

	// 检查组是否存在
	result := model.DB.Where("`name` = ?", name).First(&group)
	if result.RowsAffected == 0 {
		return fmt.Errorf("group '%s' not found", name)
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
		return group, fmt.Errorf("group '%s' not found", name)
	}

	return group, nil
}
