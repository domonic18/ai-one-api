package fixtures

import (
	"time"

	"github.com/songquanpeng/one-api/model"
)

// TestUserData 提供测试用户数据
var TestUserData = []model.User{
	{
		Username:    "admin",
		Password:    "admin123",
		DisplayName: "系统管理员",
		Role:        100,
		Status:      1,
		Group:       "admin",
		Quota:       1000000,
		UsedQuota:   0,
		RequestCount: 0,
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	},
	{
		Username:    "testuser",
		Password:    "testpass",
		DisplayName: "测试用户",
		Role:        1,
		Status:      1,
		Group:       "default",
		Quota:       10000,
		UsedQuota:   1500,
		RequestCount: 25,
		CreatedAt:   time.Now().Unix() - 86400,
		UpdatedAt:   time.Now().Unix(),
	},
	{
		Username:    "premium",
		Password:    "premiumpass",
		DisplayName: "高级用户",
		Role:        1,
		Status:      1,
		Group:       "premium",
		Quota:       50000,
		UsedQuota:   8000,
		RequestCount: 120,
		CreatedAt:   time.Now().Unix() - 86400*7,
		UpdatedAt:   time.Now().Unix(),
	},
	{
		Username:    "disabled",
		Password:    "disabledpass",
		DisplayName: "已禁用用户",
		Role:        1,
		Status:      2,
		Group:       "default",
		Quota:       1000,
		UsedQuota:   500,
		RequestCount: 10,
		CreatedAt:   time.Now().Unix() - 86400*30,
		UpdatedAt:   time.Now().Unix(),
	},
}

// TestTokenData 提供测试令牌数据
var TestTokenData = []model.Token{
	{
		UserId:      1,
		Key:         "sk-admin-token",
		Name:        "管理员令牌",
		Status:      1,
		RemainQuota: -1,
		Unlimited:   true,
		Models:      "gpt-3.5-turbo,gpt-4,claude-3,dall-e-3",
		CreatedAt:   time.Now().Unix(),
		AccessedAt:  time.Now().Unix(),
	},
	{
		UserId:      2,
		Key:         "sk-test-token",
		Name:        "测试令牌",
		Status:      1,
		RemainQuota: 8000,
		Unlimited:   false,
		Models:      "gpt-3.5-turbo",
		Subnet:      "192.168.1.0/24",
		CreatedAt:   time.Now().Unix() - 3600,
		AccessedAt:  time.Now().Unix(),
	},
	{
		UserId:      3,
		Key:         "sk-premium-token",
		Name:        "高级令牌",
		Status:      1,
		RemainQuota: 40000,
		Unlimited:   false,
		Models:      "gpt-3.5-turbo,gpt-4",
		CreatedAt:   time.Now().Unix() - 7200,
		AccessedAt:  time.Now().Unix(),
	},
	{
		UserId:      2,
		Key:         "sk-expired-token",
		Name:        "过期令牌",
		Status:      1,
		RemainQuota: 1000,
		ExpiredAt:   time.Now().Unix() - 86400,
		CreatedAt:   time.Now().Unix() - 86400*2,
		AccessedAt:  time.Now().Unix() - 86400,
	},
	{
		UserId:      2,
		Key:         "sk-disabled-token",
		Name:        "禁用令牌",
		Status:      2,
		RemainQuota: 500,
		CreatedAt:   time.Now().Unix() - 3600,
		AccessedAt:  time.Now().Unix(),
	},
}

// TestChannelData 提供测试渠道数据
var TestChannelData = []model.Channel{
	{
		Type:     model.ChannelTypeOpenAI,
		Key:      "sk-openai-key1",
		Name:     "OpenAI GPT-3.5",
		Status:   1,
		Weight:   100,
		Group:    "default",
		Models:   "gpt-3.5-turbo",
		Balance:  100.0,
		CreatedAt: time.Now().Unix(),
	},
	{
		Type:     model.ChannelTypeOpenAI,
		Key:      "sk-openai-key2",
		Name:     "OpenAI GPT-4",
		Status:   1,
		Weight:   90,
		Group:    "premium",
		Models:   "gpt-4",
		Balance:  50.0,
		CreatedAt: time.Now().Unix(),
	},
	{
		Type:     model.ChannelTypeAnthropic,
		Key:      "sk-claude-key",
		Name:     "Anthropic Claude",
		Status:   1,
		Weight:   80,
		Group:    "default",
		Models:   "claude-3-sonnet,claude-3-opus",
		Balance:  75.0,
		CreatedAt: time.Now().Unix(),
	},
	{
		Type:     model.ChannelTypeOpenAI,
		Key:      "sk-openai-key3",
		Name:     "OpenAI DALL-E",
		Status:   1,
		Weight:   70,
		Group:    "default",
		Models:   "dall-e-2,dall-e-3",
		Balance:  30.0,
		CreatedAt: time.Now().Unix(),
	},
	{
		Type:     model.ChannelTypeOpenAI,
		Key:      "sk-openai-key4",
		Name:     "OpenAI Disabled",
		Status:   2,
		Weight:   60,
		Group:    "default",
		Models:   "gpt-3.5-turbo",
		Balance:  0.0,
		CreatedAt: time.Now().Unix(),
	},
}

// TestLogData 提供测试日志数据
var TestLogData = []model.Log{
	{
		UserId:           2,
		Type:             model.LogTypeConsume,
		Username:         "testuser",
		TokenName:        "测试令牌",
		ModelName:        "gpt-3.5-turbo",
		Quota:            100,
		PromptTokens:     200,
		CompletionTokens: 100,
		ChannelId:        1,
		DurationMs:       1500,
		Content:          "模型调用成功",
		CreatedAt:        time.Now().Unix() - 3600,
	},
	{
		UserId:           3,
		Type:             model.LogTypeConsume,
		Username:         "premium",
		TokenName:        "高级令牌",
		ModelName:        "gpt-4",
		Quota:            200,
		PromptTokens:     150,
		CompletionTokens: 250,
		ChannelId:        2,
		DurationMs:       2200,
		Content:          "GPT-4调用成功",
		CreatedAt:        time.Now().Unix() - 1800,
	},
	{
		UserId:           2,
		Type:             model.LogTypeConsume,
		Username:         "testuser",
		TokenName:        "测试令牌",
		ModelName:        "dall-e-3",
		Quota:            2000,
		PromptTokens:     0,
		CompletionTokens: 0,
		ChannelId:        4,
		DurationMs:       5000,
		Content:          "图像生成成功",
		CreatedAt:        time.Now().Unix() - 900,
	},
	{
		UserId:           2,
		Type:             model.LogTypeRecharge,
		Username:         "testuser",
		TokenName:        "",
		ModelName:        "",
		Quota:            5000,
		PromptTokens:     0,
		CompletionTokens: 0,
		ChannelId:        0,
		DurationMs:       0,
		Content:          "管理员充值",
		CreatedAt:        time.Now().Unix() - 600,
	},
}

// TestCase 提供完整的测试场景数据
var TestCase = struct {
	Users    []model.User
	Tokens   []model.Token
	Channels []model.Channel
	Logs     []model.Log
}{
	Users:    TestUserData,
	Tokens:   TestTokenData,
	Channels: TestChannelData,
	Logs:     TestLogData,
}

// InsertTestData 插入测试数据到数据库
func InsertTestData(db *gorm.DB) error {
	// 插入用户数据
	if err := db.CreateInBatches(TestUserData, len(TestUserData)).Error; err != nil {
		return err
	}

	// 插入令牌数据
	if err := db.CreateInBatches(TestTokenData, len(TestTokenData)).Error; err != nil {
		return err
	}

	// 插入渠道数据
	if err := db.CreateInBatches(TestChannelData, len(TestChannelData)).Error; err != nil {
		return err
	}

	// 插入日志数据
	if err := db.CreateInBatches(TestLogData, len(TestLogData)).Error; err != nil {
		return err
	}

	return nil
}

// ClearTestData 清理测试数据
func ClearTestData(db *gorm.DB) error {
	tables := []interface{}{
		&model.Log{},
		&model.Channel{},
		&model.Token{},
		&model.User{},
	}

	for _, table := range tables {
		if err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(table).Error; err != nil {
			return err
		}
	}

	return nil
}

// GetTestUser 根据用户名获取测试用户
func GetTestUser(username string) *model.User {
	for _, user := range TestUserData {
		if user.Username == username {
			return &user
		}
	}
	return nil
}

// GetTestToken 根据key获取测试令牌
func GetTestToken(key string) *model.Token {
	for _, token := range TestTokenData {
		if token.Key == key {
			return &token
		}
	}
	return nil
}

// GetTestChannelsByGroup 根据分组获取测试渠道
func GetTestChannelsByGroup(group string) []model.Channel {
	var channels []model.Channel
	for _, channel := range TestChannelData {
		if channel.Group == group && channel.Status == 1 {
			channels = append(channels, channel)
		}
	}
	return channels
}