package unit

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/tests/common"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupTokenMultiGroupsTestDB() *gorm.DB {
	return common.SetupMySQLTestDB()
}

// TestToken_MultiUserGroups_基本功能 测试令牌多用户组的基本功能
// 测试目的：验证令牌能够正确配置和使用多个用户组
// 测试内容：
// 1. 测试设置多个用户组
// 2. 测试获取用户组列表
// 3. 测试用户组优先级顺序
// 4. 测试向后兼容性
func TestToken_MultiUserGroups_基本功能(t *testing.T) {
	db := setupTokenMultiGroupsTestDB()
	model.DB = db
	common.CleanupTestDB(db)

	// 创建测试用户
	user := model.User{
		Username: "testuser",
		Password: "password123",
		Status:   1,
		Role:     1,
		Group:    "default",
		Quota:    100000,
	}
	err := user.Insert(nil, 0)
	assert.NoError(t, err)

	// 创建测试令牌
	token := model.Token{
		UserId:         user.Id,
		Key:            "test-multi-groups-token",
		Name:           "多用户组测试令牌",
		Status:         1,
		ExpiredTime:    -1,
		RemainQuota:    1000,
		UnlimitedQuota: false,
		CreatedTime:    time.Now().Unix(),
		AccessedTime:   time.Now().Unix(),
	}

	tests := []struct {
		name           string
		userGroups     []string
		expectedGroups []string
		primaryGroup   string
	}{
		{
			name:           "单个用户组",
			userGroups:     []string{"beijing_math_group"},
			expectedGroups: []string{"beijing_math_group"},
			primaryGroup:   "beijing_math_group",
		},
		{
			name:           "多个用户组",
			userGroups:     []string{"beijing_math_group", "default"},
			expectedGroups: []string{"beijing_math_group", "default"},
			primaryGroup:   "beijing_math_group",
		},
		{
			name:           "三个用户组",
			userGroups:     []string{"beijing_math_group", "beijing_ai_group", "default"},
			expectedGroups: []string{"beijing_math_group", "beijing_ai_group", "default"},
			primaryGroup:   "beijing_math_group",
		},
		{
			name:           "空用户组列表",
			userGroups:     []string{},
			expectedGroups: []string{"default"},
			primaryGroup:   "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置用户组
			err := token.SetUserGroups(tt.userGroups)
			assert.NoError(t, err)

			// 插入令牌
			err = token.Insert()
			assert.NoError(t, err)

			// 验证用户组列表
			userGroups := token.GetUserGroups()
			assert.Equal(t, tt.expectedGroups, userGroups)

			// 验证主要用户组
			primaryGroup := token.GetPrimaryGroup()
			assert.Equal(t, tt.primaryGroup, primaryGroup)

			// 验证每个用户组是否在列表中
			for _, group := range tt.expectedGroups {
				assert.True(t, token.HasGroup(group))
			}

			// 验证不存在的用户组
			assert.False(t, token.HasGroup("non_existent_group"))

			// 清理测试数据
			model.DB.Where("1 = 1").Delete(&model.Token{})
		})
	}
}

// TestToken_MultiUserGroups_用户组选择 测试令牌根据解析的用户组进行智能选择
// 测试目的：验证令牌能够根据解析的用户组智能选择合适的用户组
// 测试内容：
// 1. 测试解析的用户组在令牌组列表中的情况
// 2. 测试解析的用户组不在令牌组列表中的情况
// 3. 测试无用户组解析的情况
// 4. 测试优先级顺序的正确性
func TestToken_MultiUserGroups_用户组选择(t *testing.T) {
	db := setupTokenMultiGroupsTestDB()
	model.DB = db
	common.CleanupTestDB(db)

	// 创建测试用户
	user := model.User{
		Username: "testuser",
		Password: "password123",
		Status:   1,
		Role:     1,
		Group:    "default",
		Quota:    100000,
	}
	err := user.Insert(nil, 0)
	assert.NoError(t, err)

	// 创建测试令牌
	token := model.Token{
		UserId:         user.Id,
		Key:            "test-group-selection-token",
		Name:           "用户组选择测试令牌",
		Status:         1,
		ExpiredTime:    -1,
		RemainQuota:    1000,
		UnlimitedQuota: false,
		CreatedTime:    time.Now().Unix(),
		AccessedTime:   time.Now().Unix(),
	}

	// 设置多个用户组
	userGroups := []string{"beijing_math_group", "beijing_ai_group", "default"}
	err = token.SetUserGroups(userGroups)
	assert.NoError(t, err)

	err = token.Insert()
	assert.NoError(t, err)

	tests := []struct {
		name          string
		resolvedGroup string
		expectedGroup string
		description   string
	}{
		{
			name:          "解析到第一个用户组",
			resolvedGroup: "beijing_math_group",
			expectedGroup: "beijing_math_group",
			description:   "解析的用户组在令牌组列表中，应使用该组",
		},
		{
			name:          "解析到第二个用户组",
			resolvedGroup: "beijing_ai_group",
			expectedGroup: "beijing_ai_group",
			description:   "解析的用户组在令牌组列表中，应使用该组",
		},
		{
			name:          "解析到第三个用户组",
			resolvedGroup: "default",
			expectedGroup: "default",
			description:   "解析的用户组在令牌组列表中，应使用该组",
		},
		{
			name:          "解析的用户组不在列表中",
			resolvedGroup: "shanghai_math_group",
			expectedGroup: "beijing_math_group",
			description:   "解析的用户组不在令牌组列表中，应使用最高优先级组",
		},
		{
			name:          "无用户组解析",
			resolvedGroup: "",
			expectedGroup: "beijing_math_group",
			description:   "无用户组解析时，应使用最高优先级组",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selectedGroup := token.SelectGroupByUserGroup(tt.resolvedGroup)
			assert.Equal(t, tt.expectedGroup, selectedGroup, tt.description)
		})
	}
}

// TestToken_MultiUserGroups_JSON序列化 测试令牌多用户组的JSON序列化和反序列化
// 测试目的：验证令牌多用户组功能的JSON兼容性
// 测试内容：
// 1. 测试多用户组令牌的JSON序列化
// 2. 测试JSON反序列化到令牌对象
// 3. 测试向后兼容性（group字段）
// 4. 验证序列化后的数据完整性
func TestToken_MultiUserGroups_JSON序列化(t *testing.T) {
	db := setupTokenMultiGroupsTestDB()
	model.DB = db
	common.CleanupTestDB(db)

	// 创建测试用户
	user := model.User{
		Username: "testuser",
		Password: "password123",
		Status:   1,
		Role:     1,
		Group:    "default",
		Quota:    100000,
	}
	err := user.Insert(nil, 0)
	assert.NoError(t, err)

	// 创建测试令牌
	token := model.Token{
		UserId:         user.Id,
		Key:            "test-json-token",
		Name:           "JSON序列化测试令牌",
		Status:         1,
		ExpiredTime:    -1,
		RemainQuota:    1000,
		UnlimitedQuota: false,
		CreatedTime:    time.Now().Unix(),
		AccessedTime:   time.Now().Unix(),
	}

	// 设置多个用户组
	userGroups := []string{"beijing_math_group", "default"}
	err = token.SetUserGroups(userGroups)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		token       *model.Token
		description string
	}{
		{
			name:        "多用户组令牌序列化",
			token:       &token,
			description: "测试多用户组令牌的JSON序列化",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 序列化令牌
			jsonData, err := json.Marshal(tt.token)
			assert.NoError(t, err)

			// 不再校验兼容字段 group

			// 反序列化令牌
			var newToken model.Token
			err = json.Unmarshal(jsonData, &newToken)
			assert.NoError(t, err)

			// 验证反序列化后的用户组
			userGroups := newToken.GetUserGroups()
			assert.Equal(t, []string{"beijing_math_group", "default"}, userGroups)

			// 验证主要用户组
			primaryGroup := newToken.GetPrimaryGroup()
			assert.Equal(t, "beijing_math_group", primaryGroup)
		})
	}
}

// 兼容测试已移除：不再支持旧的 group 字段反序列化

// TestToken_MultiUserGroups_边界情况 测试令牌多用户组功能的边界情况
// 测试目的：验证功能在各种边界情况下的正确性
// 测试内容：
// 1. 测试大量用户组的情况
// 2. 测试重复用户组的情况
// 3. 测试特殊字符用户组名称
// 4. 测试空字符串用户组
func TestToken_MultiUserGroups_边界情况(t *testing.T) {
	db := setupTokenMultiGroupsTestDB()
	model.DB = db
	common.CleanupTestDB(db)

	// 创建测试用户
	user := model.User{
		Username: "testuser",
		Password: "password123",
		Status:   1,
		Role:     1,
		Group:    "default",
		Quota:    100000,
	}
	err := user.Insert(nil, 0)
	assert.NoError(t, err)

	// 创建测试令牌
	token := model.Token{
		UserId:         user.Id,
		Key:            "test-edge-cases-token",
		Name:           "边界情况测试令牌",
		Status:         1,
		ExpiredTime:    -1,
		RemainQuota:    1000,
		UnlimitedQuota: false,
		CreatedTime:    time.Now().Unix(),
		AccessedTime:   time.Now().Unix(),
	}

	tests := []struct {
		name           string
		userGroups     []string
		expectedGroups []string
		description    string
	}{
		{
			name:           "大量用户组",
			userGroups:     []string{"group1", "group2", "group3", "group4", "group5", "group6", "group7", "group8", "group9", "group10"},
			expectedGroups: []string{"group1", "group2", "group3", "group4", "group5", "group6", "group7", "group8", "group9", "group10"},
			description:    "大量用户组应该正确处理",
		},
		{
			name:           "重复用户组",
			userGroups:     []string{"group1", "group2", "group1", "group3"},
			expectedGroups: []string{"group1", "group2", "group1", "group3"},
			description:    "重复用户组应该保持顺序",
		},
		{
			name:           "特殊字符用户组",
			userGroups:     []string{"group-with-dash", "group_with_underscore", "group.with.dot", "group with space"},
			expectedGroups: []string{"group-with-dash", "group_with_underscore", "group.with.dot", "group with space"},
			description:    "特殊字符用户组应该正确处理",
		},
		{
			name:           "空字符串用户组",
			userGroups:     []string{"", "group1", ""},
			expectedGroups: []string{"", "group1", ""},
			description:    "空字符串用户组应该保持",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置用户组
			err := token.SetUserGroups(tt.userGroups)
			assert.NoError(t, err)

			// 验证用户组列表
			userGroups := token.GetUserGroups()
			assert.Equal(t, tt.expectedGroups, userGroups, tt.description)

			// 验证主要用户组
			primaryGroup := token.GetPrimaryGroup()
			if len(tt.expectedGroups) > 0 {
				assert.Equal(t, tt.expectedGroups[0], primaryGroup)
			} else {
				assert.Equal(t, "default", primaryGroup)
			}

			// 验证每个用户组是否在列表中
			for _, group := range tt.expectedGroups {
				assert.True(t, token.HasGroup(group))
			}
		})
	}
}

// TestToken_MultiUserGroups_数据库操作 测试令牌多用户组功能的数据库操作
// 测试目的：验证多用户组令牌的数据库存储和检索
// 测试内容：
// 1. 测试令牌插入和更新
// 2. 测试从数据库检索令牌
// 3. 测试用户组数据的持久化
// 4. 验证数据库字段的正确性
func TestToken_MultiUserGroups_数据库操作(t *testing.T) {
	db := setupTokenMultiGroupsTestDB()
	model.DB = db
	common.CleanupTestDB(db)

	// 创建测试用户
	user := model.User{
		Username: "testuser",
		Password: "password123",
		Status:   1,
		Role:     1,
		Group:    "default",
		Quota:    100000,
	}
	err := user.Insert(nil, 0)
	assert.NoError(t, err)

	// 创建测试令牌
	token := model.Token{
		UserId:         user.Id,
		Key:            "test-db-token",
		Name:           "数据库操作测试令牌",
		Status:         1,
		ExpiredTime:    -1,
		RemainQuota:    1000,
		UnlimitedQuota: false,
		CreatedTime:    time.Now().Unix(),
		AccessedTime:   time.Now().Unix(),
	}

	// 设置多个用户组
	userGroups := []string{"beijing_math_group", "beijing_ai_group", "default"}
	err = token.SetUserGroups(userGroups)
	assert.NoError(t, err)

	// 插入令牌
	err = token.Insert()
	assert.NoError(t, err)

	// 从数据库检索令牌
	retrievedToken, err := model.GetTokenById(token.Id)
	assert.NoError(t, err)

	// 验证用户组数据
	retrievedGroups := retrievedToken.GetUserGroups()
	assert.Equal(t, userGroups, retrievedGroups)

	// 验证主要用户组
	primaryGroup := retrievedToken.GetPrimaryGroup()
	assert.Equal(t, "beijing_math_group", primaryGroup)

	// 验证用户组存在性
	assert.True(t, retrievedToken.HasGroup("beijing_math_group"))
	assert.True(t, retrievedToken.HasGroup("beijing_ai_group"))
	assert.True(t, retrievedToken.HasGroup("default"))
	assert.False(t, retrievedToken.HasGroup("non_existent_group"))

	// 测试更新令牌
	newUserGroups := []string{"shanghai_math_group", "default"}
	err = retrievedToken.SetUserGroups(newUserGroups)
	assert.NoError(t, err)

	err = retrievedToken.Update()
	assert.NoError(t, err)

	// 验证更新后的用户组
	updatedToken, err := model.GetTokenById(token.Id)
	assert.NoError(t, err)

	updatedGroups := updatedToken.GetUserGroups()
	assert.Equal(t, newUserGroups, updatedGroups)

	updatedPrimaryGroup := updatedToken.GetPrimaryGroup()
	assert.Equal(t, "shanghai_math_group", updatedPrimaryGroup)
}
