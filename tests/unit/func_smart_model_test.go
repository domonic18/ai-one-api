package unit

import (
	"context"
	"testing"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/model/smart"
	"github.com/stretchr/testify/assert"
)

// TestSmartModelSelector 测试智能模型选择器
type TestSmartModelSelector struct {
	userId        string
	originalModel string
	expectedModel string
	shouldReplace bool
}

// SelectModel 模拟智能模型选择逻辑
func (s *TestSmartModelSelector) SelectModel(ctx context.Context, userId string, originalModel string) (string, bool) {
	// 简单的测试逻辑：如果用户ID匹配，则返回预期模型
	if userId == s.userId && originalModel == s.originalModel {
		return s.expectedModel, s.shouldReplace
	}
	return originalModel, false
}

func TestSmartModelSelectorInterface(t *testing.T) {
	ctx := context.Background()

	t.Run("测试智能模型选择器接口", func(t *testing.T) {
		// 创建测试选择器
		selector := &TestSmartModelSelector{
			userId:        "teacher_001",
			originalModel: "gpt-3.5-turbo",
			expectedModel: "gpt-4-turbo",
			shouldReplace: true,
		}

		// 测试匹配的情况
		selectedModel, replaced := selector.SelectModel(ctx, "teacher_001", "gpt-3.5-turbo")
		assert.Equal(t, "gpt-4-turbo", selectedModel)
		assert.True(t, replaced)

		// 测试不匹配的情况
		selectedModel, replaced = selector.SelectModel(ctx, "teacher_002", "gpt-3.5-turbo")
		assert.Equal(t, "gpt-3.5-turbo", selectedModel)
		assert.False(t, replaced)
	})

	t.Run("测试默认模型选择器", func(t *testing.T) {
		// 禁用Redis以进行单元测试
		originalRedisEnabled := common.RedisEnabled
		common.RedisEnabled = false
		defer func() {
			common.RedisEnabled = originalRedisEnabled
		}()

		// 使用默认选择器
		selector := smart.NewModelSelector()

		// 测试空用户ID
		selectedModel, replaced := selector.SelectModel(ctx, "", "gpt-3.5-turbo")
		assert.Equal(t, "gpt-3.5-turbo", selectedModel)
		assert.False(t, replaced)

		// 测试空原始模型
		selectedModel, replaced = selector.SelectModel(ctx, "teacher_001", "")
		assert.Equal(t, "", selectedModel)
		assert.False(t, replaced)

		// 测试Redis禁用情况
		selectedModel, replaced = selector.SelectModel(ctx, "teacher_001", "gpt-3.5-turbo")
		assert.Equal(t, "gpt-3.5-turbo", selectedModel)
		assert.False(t, replaced)
	})
}

func TestModelAvailability(t *testing.T) {
	t.Run("测试模型可用性检查", func(t *testing.T) {
		// 测试可用模型
		assert.True(t, smart.IsModelAvailable("gpt-4-turbo"))
		assert.True(t, smart.IsModelAvailable("gpt-3.5-turbo"))
		assert.True(t, smart.IsModelAvailable("claude-3-opus"))

		// 测试不可用模型
		assert.False(t, smart.IsModelAvailable(""))
		assert.False(t, smart.IsModelAvailable("   "))
		assert.False(t, smart.IsModelAvailable("non-existent-model"))
	})
}

func TestModelGrouping(t *testing.T) {
	t.Run("测试模型分组", func(t *testing.T) {
		// 测试GPT模型分组
		assert.Equal(t, "gpt-4", smart.GetModelGroup("gpt-4-turbo"))
		assert.Equal(t, "gpt-4", smart.GetModelGroup("gpt-4"))
		assert.Equal(t, "gpt-3.5", smart.GetModelGroup("gpt-3.5-turbo"))

		// 测试Claude模型分组
		assert.Equal(t, "claude", smart.GetModelGroup("claude-3-opus"))
		assert.Equal(t, "claude", smart.GetModelGroup("claude-3-sonnet"))

		// 测试其他模型分组
		assert.Equal(t, "qwen", smart.GetModelGroup("qwen-turbo"))
		assert.Equal(t, "deepseek", smart.GetModelGroup("deepseek-chat"))

		// 测试未知模型
		assert.Equal(t, "unknown-model", smart.GetModelGroup("unknown-model"))
	})
}

func TestModelParameters(t *testing.T) {
	t.Run("测试模型参数获取", func(t *testing.T) {
		// 测试GPT-4参数
		params := smart.GetModelParameters("gpt-4-turbo")
		assert.Equal(t, 0.7, params["temperature"])
		assert.Equal(t, 4000, params["max_tokens"])

		// 测试GPT-3.5参数
		params = smart.GetModelParameters("gpt-3.5-turbo")
		assert.Equal(t, 0.8, params["temperature"])
		assert.Equal(t, 2000, params["max_tokens"])

		// 测试Claude参数
		params = smart.GetModelParameters("claude-3-opus")
		assert.Equal(t, 0.7, params["temperature"])
		assert.Equal(t, 4000, params["max_tokens"])

		// 测试默认参数
		params = smart.GetModelParameters("unknown-model")
		assert.Equal(t, 0.7, params["temperature"])
		assert.Equal(t, 2000, params["max_tokens"])
	})
}

func TestGetModelByTeacherId(t *testing.T) {
	// 禁用Redis以进行单元测试
	originalRedisEnabled := common.RedisEnabled
	common.RedisEnabled = false
	defer func() {
		common.RedisEnabled = originalRedisEnabled
	}()

	ctx := context.Background()

	t.Run("空用户ID测试", func(t *testing.T) {
		modelName, err := smart.GetModelByTeacherId(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "teacherId不能为空")
		assert.Equal(t, "", modelName)
	})

	t.Run("Redis禁用测试", func(t *testing.T) {
		modelName, err := smart.GetModelByTeacherId(ctx, "teacher_001")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis未启用")
		assert.Equal(t, "", modelName)
	})
}

func TestCacheConstants(t *testing.T) {
	t.Run("缓存常量测试", func(t *testing.T) {
		// 验证缓存前缀
		assert.Equal(t, "teacher_info:", smart.TeacherInfoCachePrefix)
		assert.Equal(t, "subject_info:", smart.SubjectInfoCachePrefix)
	})
}

func TestCacheOperations(t *testing.T) {
	// 禁用Redis以进行单元测试
	originalRedisEnabled := common.RedisEnabled
	common.RedisEnabled = false
	defer func() {
		common.RedisEnabled = originalRedisEnabled
	}()

	ctx := context.Background()

	t.Run("获取老师信息缓存测试", func(t *testing.T) {
		teacherId := "teacher_001"
		info, err := smart.GetTeacherInfoWithCache(ctx, teacherId)
		assert.Error(t, err)
		assert.Nil(t, info)
		assert.Contains(t, err.Error(), "Redis未启用")
	})

	t.Run("获取学科组信息缓存测试", func(t *testing.T) {
		subjectId := 1
		info, err := smart.GetSubjectInfoWithCache(ctx, subjectId)
		assert.Error(t, err)
		assert.Nil(t, info)
		assert.Contains(t, err.Error(), "Redis未启用")
	})

	t.Run("失效老师信息缓存测试", func(t *testing.T) {
		teacherId := "teacher_001"
		err := smart.InvalidateTeacherInfoCache(ctx, teacherId)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis未启用")
	})

	t.Run("失效学科组信息缓存测试", func(t *testing.T) {
		subjectId := 1
		err := smart.InvalidateSubjectInfoCache(ctx, subjectId)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis未启用")
	})
}
