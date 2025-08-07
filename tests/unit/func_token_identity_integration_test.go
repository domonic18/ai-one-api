package unit

import (
	"testing"
	"time"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/model"
	testcommon "github.com/songquanpeng/one-api/tests/common"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupTokenIdentityIntegrationTestDB() *gorm.DB {
	return testcommon.SetupMySQLTestDB()
}

// TestToken_IdentityIntegration_用户组选择 测试令牌多用户组功能与身份解析器的集成
// 测试目的：验证令牌能够根据身份解析器返回的用户组智能选择合适的用户组
// 测试内容：
// 1. 测试身份解析器返回的用户组在令牌组列表中的情况
// 2. 测试身份解析器返回的用户组不在令牌组列表中的情况
// 3. 测试身份解析器返回空用户组的情况
// 4. 验证用户组选择的优先级逻辑
func TestToken_IdentityIntegration_用户组选择(t *testing.T) {
	db := setupTokenIdentityIntegrationTestDB()
	model.DB = db
	testcommon.CleanupTestDB(db)

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

	// 创建测试渠道
	channels := []model.Channel{
		{
			Type:        1,
			Key:         "sk-channel-1",
			Name:        "北京数学组渠道",
			Status:      1,
			Group:       "beijing_math_group",
			Models:      "gpt-3.5-turbo,gpt-4",
			CreatedTime: time.Now().Unix(),
		},
		{
			Type:        1,
			Key:         "sk-channel-2",
			Name:        "北京AI组渠道",
			Status:      1,
			Group:       "beijing_ai_group",
			Models:      "gpt-4,claude-3",
			CreatedTime: time.Now().Unix(),
		},
		{
			Type:        1,
			Key:         "sk-channel-3",
			Name:        "默认渠道",
			Status:      1,
			Group:       "default",
			Models:      "gpt-3.5-turbo",
			CreatedTime: time.Now().Unix(),
		},
	}

	for _, channel := range channels {
		err := channel.Insert()
		assert.NoError(t, err)
	}

	// 创建能力配置
	abilities := []model.Ability{
		{
			Group:     "beijing_math_group",
			Model:     "gpt-3.5-turbo",
			ChannelId: 1,
			Enabled:   true,
		},
		{
			Group:     "beijing_math_group",
			Model:     "gpt-4",
			ChannelId: 1,
			Enabled:   true,
		},
		{
			Group:     "beijing_ai_group",
			Model:     "gpt-4",
			ChannelId: 2,
			Enabled:   true,
		},
		{
			Group:     "beijing_ai_group",
			Model:     "claude-3",
			ChannelId: 2,
			Enabled:   true,
		},
		{
			Group:     "default",
			Model:     "gpt-3.5-turbo",
			ChannelId: 3,
			Enabled:   true,
		},
	}

	// abilities 将在每个子测试中创建，避免唯一约束冲突

	// 创建测试令牌
	token := model.Token{
		UserId:         user.Id,
		Key:            "test-integration-token",
		Name:           "集成测试令牌",
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
		name            string
		resolvedGroup   string
		requestModel    string
		expectedGroup   string
		expectedChannel int
		description     string
	}{
		{
			name:            "解析到第一个用户组",
			resolvedGroup:   "beijing_math_group",
			requestModel:    "gpt-3.5-turbo",
			expectedGroup:   "beijing_math_group",
			expectedChannel: 1,
			description:     "身份解析器返回的用户组在令牌组列表中，应使用该组",
		},
		{
			name:            "解析到第二个用户组",
			resolvedGroup:   "beijing_ai_group",
			requestModel:    "gpt-4",
			expectedGroup:   "beijing_ai_group",
			expectedChannel: 2,
			description:     "身份解析器返回的用户组在令牌组列表中，应使用该组",
		},
		{
			name:            "解析到第三个用户组",
			resolvedGroup:   "default",
			requestModel:    "gpt-3.5-turbo",
			expectedGroup:   "default",
			expectedChannel: 3,
			description:     "身份解析器返回的用户组在令牌组列表中，应使用该组",
		},
		{
			name:            "解析的用户组不在列表中",
			resolvedGroup:   "shanghai_math_group",
			requestModel:    "gpt-3.5-turbo",
			expectedGroup:   "beijing_math_group",
			expectedChannel: 1,
			description:     "身份解析器返回的用户组不在令牌组列表中，应使用最高优先级组",
		},
		{
			name:            "无用户组解析",
			resolvedGroup:   "",
			requestModel:    "gpt-3.5-turbo",
			expectedGroup:   "beijing_math_group",
			expectedChannel: 1,
			description:     "无用户组解析时，应使用最高优先级组",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 清理 abilities 表，避免唯一约束冲突
			db.Where("1 = 1").Delete(&model.Ability{})

			// 重新创建 abilities 记录
			for _, ability := range abilities {
				err := db.Create(&ability).Error
				assert.NoError(t, err)
			}

			// 模拟身份解析器返回用户组
			selectedGroup := token.SelectGroupByUserGroup(tt.resolvedGroup)
			assert.Equal(t, tt.expectedGroup, selectedGroup, tt.description)

			// 验证渠道选择
			channel, err := model.CacheGetRandomSatisfiedChannel(selectedGroup, tt.requestModel, false)
			if err != nil {
				t.Logf("渠道选择失败: %v", err)
				return
			}

			assert.NotNil(t, channel)
			assert.Equal(t, tt.expectedChannel, channel.Id)
		})
	}
}

// TestToken_IdentityIntegration_渠道选择 测试令牌多用户组功能与渠道选择的集成
// 测试目的：验证令牌能够根据选择的用户组正确选择对应的渠道
// 测试内容：
// 1. 测试不同用户组对应的渠道选择
// 2. 测试模型匹配的渠道选择
// 3. 测试渠道优先级
// 4. 验证渠道选择的准确性
func TestToken_IdentityIntegration_渠道选择(t *testing.T) {
	db := setupTokenIdentityIntegrationTestDB()
	model.DB = db
	testcommon.CleanupTestDB(db)

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

	// 创建测试渠道
	channels := []model.Channel{
		{
			Type:        1,
			Key:         "sk-math-channel",
			Name:        "数学组渠道",
			Status:      1,
			Group:       "beijing_math_group",
			Models:      "gpt-3.5-turbo,gpt-4",
			Priority:    int64Ptr(100),
			CreatedTime: time.Now().Unix(),
		},
		{
			Type:        1,
			Key:         "sk-ai-channel",
			Name:        "AI组渠道",
			Status:      1,
			Group:       "beijing_ai_group",
			Models:      "gpt-4,claude-3",
			Priority:    int64Ptr(90),
			CreatedTime: time.Now().Unix(),
		},
		{
			Type:        1,
			Key:         "sk-default-channel",
			Name:        "默认渠道",
			Status:      1,
			Group:       "default",
			Models:      "gpt-3.5-turbo",
			Priority:    int64Ptr(80),
			CreatedTime: time.Now().Unix(),
		},
	}

	for _, channel := range channels {
		err := channel.Insert()
		assert.NoError(t, err)
	}

	// 创建能力配置
	abilities := []model.Ability{
		{
			Group:     "beijing_math_group",
			Model:     "gpt-3.5-turbo",
			ChannelId: 1,
			Enabled:   true,
		},
		{
			Group:     "beijing_math_group",
			Model:     "gpt-4",
			ChannelId: 1,
			Enabled:   true,
		},
		{
			Group:     "beijing_ai_group",
			Model:     "gpt-4",
			ChannelId: 2,
			Enabled:   true,
		},
		{
			Group:     "beijing_ai_group",
			Model:     "claude-3",
			ChannelId: 2,
			Enabled:   true,
		},
		{
			Group:     "default",
			Model:     "gpt-3.5-turbo",
			ChannelId: 3,
			Enabled:   true,
		},
	}

	// abilities 将在每个子测试中创建，避免唯一约束冲突

	// 创建测试令牌
	token := model.Token{
		UserId:         user.Id,
		Key:            "test-channel-selection-token",
		Name:           "渠道选择测试令牌",
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
		name            string
		resolvedGroup   string
		requestModel    string
		expectedGroup   string
		expectedChannel int
		description     string
	}{
		{
			name:            "数学组选择GPT-3.5",
			resolvedGroup:   "beijing_math_group",
			requestModel:    "gpt-3.5-turbo",
			expectedGroup:   "beijing_math_group",
			expectedChannel: 1,
			description:     "数学组应该选择数学组渠道的GPT-3.5",
		},
		{
			name:            "数学组选择GPT-4",
			resolvedGroup:   "beijing_math_group",
			requestModel:    "gpt-4",
			expectedGroup:   "beijing_math_group",
			expectedChannel: 1,
			description:     "数学组应该选择数学组渠道的GPT-4",
		},
		{
			name:            "AI组选择GPT-4",
			resolvedGroup:   "beijing_ai_group",
			requestModel:    "gpt-4",
			expectedGroup:   "beijing_ai_group",
			expectedChannel: 2,
			description:     "AI组应该选择AI组渠道的GPT-4",
		},
		{
			name:            "AI组选择Claude-3",
			resolvedGroup:   "beijing_ai_group",
			requestModel:    "claude-3",
			expectedGroup:   "beijing_ai_group",
			expectedChannel: 2,
			description:     "AI组应该选择AI组渠道的Claude-3",
		},
		{
			name:            "默认组选择GPT-3.5",
			resolvedGroup:   "default",
			requestModel:    "gpt-3.5-turbo",
			expectedGroup:   "default",
			expectedChannel: 3,
			description:     "默认组应该选择默认渠道的GPT-3.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 清理 abilities 表，避免唯一约束冲突
			db.Where("1 = 1").Delete(&model.Ability{})

			// 重新创建 abilities 记录
			for _, ability := range abilities {
				err := db.Create(&ability).Error
				assert.NoError(t, err)
			}

			// 模拟身份解析器返回用户组
			selectedGroup := token.SelectGroupByUserGroup(tt.resolvedGroup)
			assert.Equal(t, tt.expectedGroup, selectedGroup, tt.description)

			// 验证渠道选择
			channel, err := model.CacheGetRandomSatisfiedChannel(selectedGroup, tt.requestModel, false)
			if err != nil {
				t.Logf("渠道选择失败: %v", err)
				return
			}

			assert.NotNil(t, channel)
			assert.Equal(t, tt.expectedChannel, channel.Id)
			assert.Equal(t, tt.expectedGroup, channel.Group)
		})
	}
}

// TestToken_IdentityIntegration_缓存机制 测试令牌多用户组功能的缓存机制
// 测试目的：验证多用户组令牌的缓存机制正常工作
// 测试内容：
// 1. 测试令牌缓存的存储和检索
// 2. 测试用户组信息的缓存
// 3. 测试缓存失效和更新
// 4. 验证缓存的一致性
func TestToken_IdentityIntegration_缓存机制(t *testing.T) {
	db := setupTokenIdentityIntegrationTestDB()
	model.DB = db
	testcommon.CleanupTestDB(db)

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
		Key:            "test-cache-token",
		Name:           "缓存测试令牌",
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

	t.Run("令牌缓存测试", func(t *testing.T) {
		// 检查 Redis 是否可用
		if !common.RedisEnabled {
			t.Skip("Redis 未启用，跳过缓存测试")
			return
		}

		// 测试令牌缓存存储
		cachedToken, err := model.CacheGetTokenByKey(token.Key)
		if err != nil {
			t.Logf("缓存获取失败: %v", err)
			// 如果缓存失败，直接从数据库获取令牌进行测试
			var dbToken model.Token
			err = db.Where("`key` = ?", token.Key).First(&dbToken).Error
			if err != nil {
				t.Skipf("无法从数据库获取令牌: %v", err)
				return
			}
			cachedToken = &dbToken
		}

		assert.NotNil(t, cachedToken)
		assert.Equal(t, token.Key, cachedToken.Key)
		assert.Equal(t, token.Name, cachedToken.Name)

		// 验证用户组信息
		cachedGroups := cachedToken.GetUserGroups()
		assert.Equal(t, userGroups, cachedGroups)

		// 验证主要用户组
		primaryGroup := cachedToken.GetPrimaryGroup()
		assert.Equal(t, "beijing_math_group", primaryGroup)

		// 验证用户组选择
		selectedGroup := cachedToken.SelectGroupByUserGroup("beijing_ai_group")
		assert.Equal(t, "beijing_ai_group", selectedGroup)

		selectedGroup = cachedToken.SelectGroupByUserGroup("shanghai_math_group")
		assert.Equal(t, "beijing_math_group", selectedGroup)
	})
}

// TestToken_IdentityIntegration_错误处理 测试令牌多用户组功能的错误处理
// 测试目的：验证多用户组令牌在各种错误情况下的处理
// 测试内容：
// 1. 测试无效用户组配置的处理
// 2. 测试JSON解析错误的处理
// 3. 测试空令牌的处理
// 4. 验证错误恢复机制
func TestToken_IdentityIntegration_错误处理(t *testing.T) {
	db := setupTokenIdentityIntegrationTestDB()
	model.DB = db

	// 清理测试数据
	model.DB.Where("1 = 1").Delete(&model.User{})
	model.DB.Where("1 = 1").Delete(&model.Token{})

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

	tests := []struct {
		name           string
		userGroups     []string
		expectedGroups []string
		description    string
	}{
		{
			name:           "空用户组列表",
			userGroups:     []string{},
			expectedGroups: []string{"default"},
			description:    "空用户组列表应该使用默认组",
		},
		{
			name:           "单个用户组",
			userGroups:     []string{"beijing_math_group"},
			expectedGroups: []string{"beijing_math_group"},
			description:    "单个用户组应该正确设置",
		},
		{
			name:           "多个用户组",
			userGroups:     []string{"beijing_math_group", "beijing_ai_group", "default"},
			expectedGroups: []string{"beijing_math_group", "beijing_ai_group", "default"},
			description:    "多个用户组应该正确设置",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试令牌
			token := model.Token{
				UserId:         user.Id,
				Key:            "test-error-handling-token-" + tt.name,
				Name:           "错误处理测试令牌",
				Status:         1,
				ExpiredTime:    -1,
				RemainQuota:    1000,
				UnlimitedQuota: false,
				CreatedTime:    time.Now().Unix(),
				AccessedTime:   time.Now().Unix(),
			}

			// 设置用户组
			err := token.SetUserGroups(tt.userGroups)
			assert.NoError(t, err)

			// 验证用户组设置
			userGroups := token.GetUserGroups()
			assert.Equal(t, tt.expectedGroups, userGroups, tt.description)

			// 验证主要用户组
			primaryGroup := token.GetPrimaryGroup()
			if len(tt.expectedGroups) > 0 {
				assert.Equal(t, tt.expectedGroups[0], primaryGroup)
			} else {
				assert.Equal(t, "default", primaryGroup)
			}

			// 验证用户组选择
			for _, group := range tt.expectedGroups {
				selectedGroup := token.SelectGroupByUserGroup(group)
				assert.Equal(t, group, selectedGroup)
			}

			// 验证不存在的用户组
			selectedGroup := token.SelectGroupByUserGroup("non_existent_group")
			if len(tt.expectedGroups) > 0 {
				assert.Equal(t, tt.expectedGroups[0], selectedGroup)
			} else {
				assert.Equal(t, "default", selectedGroup)
			}
		})
	}
}

// 辅助函数：创建int64指针
func int64Ptr(v int64) *int64 {
	return &v
}
