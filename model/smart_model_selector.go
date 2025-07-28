package model

import (
	"context"
	"strings"

	"github.com/songquanpeng/one-api/common/logger"
)

// SmartModelSelector 智能模型选择器
type SmartModelSelector struct {
	UserId        string // 用户ID（老师ID）
	OriginalModel string // 原始请求中的模型名称
	Context       context.Context
}

// NewSmartModelSelector 创建智能模型选择器
func NewSmartModelSelector(ctx context.Context, userId string, originalModel string) *SmartModelSelector {
	return &SmartModelSelector{
		UserId:        userId,
		OriginalModel: originalModel,
		Context:       ctx,
	}
}

// SelectModel 选择合适的模型
// 返回选择的模型名称和是否进行了模型替换
func (s *SmartModelSelector) SelectModel() (string, bool) {
	if s.UserId == "" {
		logger.Debugf(s.Context, "未提供用户ID，使用原始模型: %s", s.OriginalModel)
		return s.OriginalModel, false
	}

	// 根据用户ID获取推荐模型
	modelName, err := GetModelByTeacherId(s.Context, s.UserId)
	if err != nil || modelName == "" {
		logger.Debugf(s.Context, "未找到用户推荐模型，使用原始模型: userId=%s, model=%s",
			s.UserId, s.OriginalModel)
		return s.OriginalModel, false
	}

	// 检查推荐模型是否可用
	if !IsModelAvailable(modelName) {
		logger.Warnf(s.Context, "推荐模型不可用，使用原始模型: userId=%s, recommendedModel=%s, originalModel=%s",
			s.UserId, modelName, s.OriginalModel)
		return s.OriginalModel, false
	}

	logger.Infof(s.Context, "智能模型选择: userId=%s, originalModel=%s -> selectedModel=%s",
		s.UserId, s.OriginalModel, modelName)
	return modelName, true
}

// IsModelAvailable 检查模型是否可用
func IsModelAvailable(modelName string) bool {
	// 这里应该检查模型是否在系统中配置并可用
	// 简单实现：检查模型名称是否为空
	return modelName != "" && strings.TrimSpace(modelName) != ""
}

// GetModelGroup 获取模型所属的组
func GetModelGroup(modelName string) string {
	// 根据模型名称获取其所属的组
	// 例如：gpt-4-turbo 属于 gpt-4 组
	// 这里简单实现，实际应该根据系统中的模型配置来判断
	if strings.HasPrefix(modelName, "gpt-4") {
		return "gpt-4"
	} else if strings.HasPrefix(modelName, "gpt-3.5") {
		return "gpt-3.5"
	} else if strings.HasPrefix(modelName, "claude") {
		return "claude"
	} else if strings.HasPrefix(modelName, "qwen") {
		return "qwen"
	} else if strings.HasPrefix(modelName, "deepseek") {
		return "deepseek"
	}

	// 默认返回模型名称本身作为组
	return modelName
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
	// 这里简单实现，实际应该从数据库或配置中获取
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
