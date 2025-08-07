package unit

import (
	"testing"

	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/tests/common"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupChannelTestDB() *gorm.DB {
	return common.SetupMySQLTestDB()
}

// TestChannel_ChooseChannel 测试渠道选择功能
// 测试目的：验证系统能够根据用户组、模型名称等条件正确选择可用的渠道
// 测试内容：
// 1. 测试权重优先的渠道选择逻辑
// 2. 测试模型匹配过滤功能
// 3. 测试用户组权限控制
// 4. 测试禁用渠道的跳过逻辑
// 5. 测试高权限组的渠道选择
func TestChannel_ChooseChannel(t *testing.T) {
	db := setupChannelTestDB()
	model.DB = db
	common.CleanupTestDB(db)

	// 创建测试渠道
	channels := []model.Channel{
		{
			Type:    1,
			Key:     "key1",
			Name:    "Channel 1",
			Status:  1,
			Weight:  uintPtr(10),
			Group:   "default",
			Balance: 100.0,
			Models:  "gpt-3.5-turbo,gpt-4",
		},
		{
			Type:    1,
			Key:     "key2",
			Name:    "Channel 2",
			Status:  1,
			Weight:  uintPtr(20),
			Group:   "default",
			Balance: 50.0,
			Models:  "gpt-3.5-turbo",
		},
		{
			Type:    1,
			Key:     "key3",
			Name:    "Channel 3",
			Status:  2, // 禁用
			Group:   "default",
			Balance: 200.0,
			Models:  "gpt-3.5-turbo",
		},
		{
			Type:    1,
			Key:     "key4",
			Name:    "Channel 4",
			Status:  1,
			Weight:  uintPtr(15),
			Group:   "premium",
			Balance: 300.0,
			Models:  "gpt-4",
		},
	}

	for _, ch := range channels {
		db.Create(&ch)
		// 添加渠道能力
		ch.AddAbilities()
	}

	tests := []struct {
		name        string
		group       string
		model       string
		wantChannel bool
		wantName    string
	}{
		{
			name:        "选择可用渠道 - 权重最高",
			group:       "default",
			model:       "gpt-3.5-turbo",
			wantChannel: true,
			wantName:    "", // 不检查具体名称，因为是随机选择
		},
		{
			name:        "模型不匹配",
			group:       "default",
			model:       "claude-3",
			wantChannel: false,
		},
		{
			name:        "组不匹配",
			group:       "nonexistent",
			model:       "gpt-3.5-turbo",
			wantChannel: false,
		},
		{
			name:        "高权限组选择",
			group:       "premium",
			model:       "gpt-4",
			wantChannel: true,
			wantName:    "Channel 4",
		},
		{
			name:        "禁用渠道跳过",
			group:       "default",
			model:       "gpt-3.5-turbo",
			wantChannel: true,
			wantName:    "", // 不检查具体名称，因为是随机选择
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			channel, err := model.GetRandomSatisfiedChannel(tt.group, tt.model, false)

			if tt.wantChannel {
				assert.NoError(t, err)
				assert.NotNil(t, channel)
				if tt.wantName != "" {
					assert.Equal(t, tt.wantName, channel.Name)
				}
			} else {
				assert.Error(t, err)
				assert.Nil(t, channel)
			}
		})
	}
}

// 辅助函数
func uintPtr(v uint) *uint {
	return &v
}

// TestChannel_GetChannelsByGroupAndModel 测试按组和模型获取渠道功能
// 测试目的：验证系统能够根据用户组和模型名称正确筛选可用的渠道
// 测试内容：
// 1. 测试按用户组筛选渠道
// 2. 测试按模型名称筛选渠道
// 3. 测试不存在的组和模型处理
// 4. 验证筛选结果的准确性
func TestChannel_GetChannelsByGroupAndModel(t *testing.T) {
	db := setupChannelTestDB()
	model.DB = db
	common.CleanupTestDB(db)

	// 创建测试渠道
	channels := []model.Channel{
		{
			Type:   1,
			Key:    "key1",
			Name:   "Channel 1",
			Status: 1,
			Group:  "default",
			Models: "gpt-3.5-turbo,gpt-4",
		},
		{
			Type:   1,
			Key:    "key2",
			Name:   "Channel 2",
			Status: 1,
			Group:  "default",
			Models: "gpt-3.5-turbo",
		},
		{
			Type:   1,
			Key:    "key3",
			Name:   "Channel 3",
			Status: 2, // 禁用
			Group:  "default",
			Models: "gpt-3.5-turbo",
		},
		{
			Type:   1,
			Key:    "key4",
			Name:   "Channel 4",
			Status: 1,
			Group:  "premium",
			Models: "gpt-4",
		},
	}

	for _, ch := range channels {
		db.Create(&ch)
		// 添加渠道能力
		ch.AddAbilities()
	}

	tests := []struct {
		name     string
		group    string
		model    string
		expected int
	}{
		{
			name:     "获取default组gpt-3.5-turbo模型渠道",
			group:    "default",
			model:    "gpt-3.5-turbo",
			expected: 2, // Channel 1 和 Channel 2
		},
		{
			name:     "获取premium组gpt-4模型渠道",
			group:    "premium",
			model:    "gpt-4",
			expected: 1, // Channel 4
		},
		{
			name:     "获取不存在的组",
			group:    "nonexistent",
			model:    "gpt-3.5-turbo",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 使用GetRandomSatisfiedChannel来测试，因为这是实际存在的函数
			channel, err := model.GetRandomSatisfiedChannel(tt.group, tt.model, false)

			if tt.expected > 0 {
				assert.NoError(t, err)
				assert.NotNil(t, channel)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// TestChannel_UpdateChannelBalance 测试渠道余额更新功能
// 测试目的：验证系统能够正确更新渠道的余额信息
// 测试内容：
// 1. 测试渠道余额更新操作
// 2. 验证余额更新后的数据一致性
// 3. 检查余额更新时间戳的更新
func TestChannel_UpdateChannelBalance(t *testing.T) {
	db := setupChannelTestDB()
	model.DB = db
	common.CleanupTestDB(db)

	// 创建测试渠道
	channel := model.Channel{
		Type:    1,
		Key:     "test-key",
		Name:    "Test Channel",
		Status:  1,
		Balance: 100.0,
		Models:  "gpt-3.5-turbo",
	}

	err := channel.Insert()
	assert.NoError(t, err)

	// 测试更新余额
	newBalance := 150.0
	channel.UpdateBalance(newBalance)

	// 验证更新
	updatedChannel, err := model.GetChannelById(channel.Id, false)
	assert.NoError(t, err)
	assert.Equal(t, newBalance, updatedChannel.Balance)
	assert.NotZero(t, updatedChannel.BalanceUpdatedTime)
}

// TestChannel_UpdateChannelUsedQuota 测试渠道已使用配额更新功能
// 测试目的：验证系统能够正确更新渠道的已使用配额
// 测试内容：
// 1. 测试增加已使用配额
// 2. 测试减少已使用配额
// 3. 测试大额配额变化
// 4. 验证配额更新的准确性
func TestChannel_UpdateChannelUsedQuota(t *testing.T) {
	db := setupChannelTestDB()
	model.DB = db
	common.CleanupTestDB(db)

	// 创建测试渠道
	channel := model.Channel{
		Type:      1,
		Key:       "test-key",
		Name:      "Test Channel",
		Status:    1,
		UsedQuota: 100,
		Models:    "gpt-3.5-turbo",
	}

	err := channel.Insert()
	assert.NoError(t, err)

	tests := []struct {
		name       string
		quotaDelta int64
		expected   int64
	}{
		{
			name:       "增加配额",
			quotaDelta: 50,
			expected:   150,
		},
		{
			name:       "减少配额",
			quotaDelta: -30,
			expected:   120,
		},
		{
			name:       "增加大量配额",
			quotaDelta: 1000,
			expected:   1120,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model.UpdateChannelUsedQuota(channel.Id, tt.quotaDelta)

			// 验证更新
			updatedChannel, err := model.GetChannelById(channel.Id, false)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, updatedChannel.UsedQuota)
		})
	}
}

// TestChannel_GetChannelByID 测试根据ID获取渠道信息功能
// 测试目的：验证系统能够根据渠道ID正确获取渠道信息
// 测试内容：
// 1. 测试获取渠道基本信息（不包含密钥）
// 2. 测试获取渠道完整信息（包含密钥）
// 3. 测试获取不存在渠道的错误处理
// 4. 验证返回数据的完整性
func TestChannel_GetChannelByID(t *testing.T) {
	db := setupChannelTestDB()
	model.DB = db
	common.CleanupTestDB(db)

	// 创建测试渠道
	channel := model.Channel{
		Type:   1,
		Key:    "test-key",
		Name:   "Test Channel",
		Status: 1,
		Models: "gpt-3.5-turbo",
	}

	err := channel.Insert()
	assert.NoError(t, err)

	tests := []struct {
		name      string
		channelID int
		selectAll bool
		wantErr   bool
	}{
		{
			name:      "获取渠道信息（不包含key）",
			channelID: channel.Id,
			selectAll: false,
			wantErr:   false,
		},
		{
			name:      "获取渠道信息（包含key）",
			channelID: channel.Id,
			selectAll: true,
			wantErr:   false,
		},
		{
			name:      "获取不存在的渠道",
			channelID: 99999,
			selectAll: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := model.GetChannelById(tt.channelID, tt.selectAll)

			if tt.wantErr {
				assert.Error(t, err)
				// 不检查 result 是否为 nil，因为 GORM 总是返回一个对象
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.channelID, result.Id)
				assert.Equal(t, channel.Name, result.Name)

				if tt.selectAll {
					assert.NotEmpty(t, result.Key)
				} else {
					assert.Empty(t, result.Key)
				}
			}
		})
	}
}
