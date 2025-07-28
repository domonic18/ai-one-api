package smart

import (
	"context"
	"strings"

	"github.com/songquanpeng/one-api/common/logger"
)

// ModelSelector 智能模型选择器接口
type ModelSelector interface {
	// SelectModel 根据用户ID和原始模型选择最合适的模型
	// 返回选择的模型名称和是否进行了模型替换
	SelectModel(ctx context.Context, userId string, originalModel string) (string, bool)
}

// DefaultModelSelector 默认智能模型选择器实现
type DefaultModelSelector struct{}

// SelectModel 实现智能模型选择逻辑
func (s *DefaultModelSelector) SelectModel(ctx context.Context, userId string, originalModel string) (string, bool) {
	// 如果没有用户ID，则无法进行智能选择
	if userId == "" {
		logger.Debugf(ctx, "智能模型选择: 未提供用户ID，使用原始模型: %s", originalModel)
		return originalModel, false
	}

	// 如果没有原始模型，则无法进行智能选择
	if originalModel == "" {
		logger.Debugf(ctx, "智能模型选择: 未提供原始模型，无法进行智能选择")
		return originalModel, false
	}

	// 1. 首先尝试获取用户模型配置（用户偏好）
	selectedModel, err := GetModelByTeacherId(ctx, userId)
	if err != nil {
		logger.Warnf(ctx, "智能模型选择失败: userId=%s, error=%v, 使用原始模型: %s", userId, err, originalModel)
		return originalModel, false
	}

	// 2. 检查推荐模型是否可用
	if !IsModelAvailable(selectedModel) {
		logger.Warnf(ctx, "推荐模型不可用，使用原始模型: userId=%s, recommendedModel=%s, originalModel=%s",
			userId, selectedModel, originalModel)
		return originalModel, false
	}

	// 3. 如果选择的模型与原始模型相同，则不进行替换
	if selectedModel == originalModel {
		logger.Debugf(ctx, "智能模型选择: 选择的模型与原始模型相同: %s", originalModel)
		return originalModel, false
	}

	// 4. 返回选择的模型
	logger.Infof(ctx, "智能模型选择: 用户=%s, 原始模型=%s, 选择模型=%s", userId, originalModel, selectedModel)
	return selectedModel, true
}

// NewModelSelector 创建新的智能模型选择器
func NewModelSelector() ModelSelector {
	return &DefaultModelSelector{}
}

// IsModelAvailable 检查模型是否可用
func IsModelAvailable(modelName string) bool {
	// 检查模型名称是否为空或只包含空白字符
	if modelName == "" || strings.TrimSpace(modelName) == "" {
		return false
	}

	// 这里应该检查模型是否在系统中配置并可用
	// 简单实现：检查是否在可用模型列表中
	availableModels := GetAvailableModels()
	for _, available := range availableModels {
		if available == modelName {
			return true
		}
	}

	return false
}

// GetModelGroup 获取模型所属的组
func GetModelGroup(modelName string) string {
	// 根据模型名称获取其所属的组
	// 例如：gpt-4-turbo 属于 gpt-4 组
	switch {
	case strings.HasPrefix(modelName, "gpt-4"):
		return "gpt-4"
	case strings.HasPrefix(modelName, "gpt-3.5"):
		return "gpt-3.5"
	case strings.HasPrefix(modelName, "claude"):
		return "claude"
	case strings.HasPrefix(modelName, "qwen"):
		return "qwen"
	case strings.HasPrefix(modelName, "deepseek"):
		return "deepseek"
	default:
		// 默认返回模型名称本身作为组
		return modelName
	}
}

// GetAvailableModels 获取当前可用的所有模型
func GetAvailableModels() []string {
	// 从系统中获取所有可用的模型
	// 这里简单实现，实际应该从数据库或配置中获取
	return []string{
		"gpt-4-turbo",
		"gpt-4",
		"gpt-3.5-turbo",
		"claude-3-opus",
		"claude-3-sonnet",
		"qwen-turbo",
		"deepseek-chat",
	}
}

// GetModelParameters 获取模型的默认参数
func GetModelParameters(modelName string) map[string]interface{} {
	// 根据模型名称获取其默认参数
	params := make(map[string]interface{})

	// 根据不同模型设置不同的默认参数
	switch {
	case strings.HasPrefix(modelName, "gpt-4"):
		params["temperature"] = 0.7
		params["max_tokens"] = 4000
	case strings.HasPrefix(modelName, "gpt-3.5"):
		params["temperature"] = 0.8
		params["max_tokens"] = 2000
	case strings.HasPrefix(modelName, "claude"):
		params["temperature"] = 0.7
		params["max_tokens"] = 4000
	case strings.HasPrefix(modelName, "qwen"):
		params["temperature"] = 0.8
		params["max_tokens"] = 2000
	case strings.HasPrefix(modelName, "deepseek"):
		params["temperature"] = 0.7
		params["max_tokens"] = 2000
	default:
		// 默认参数
		params["temperature"] = 0.7
		params["max_tokens"] = 2000
	}

	return params
}
