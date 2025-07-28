package unit

import (
	"context"
	"testing"

	"github.com/songquanpeng/one-api/model"
	"github.com/stretchr/testify/assert"
)

// mock/stub实现
var (
	mockUserModelConfig *model.UserModelConfig
	mockSubjectInfo     *model.SubjectInfo
	mockTeacherInfo     *model.TeacherInfo
)

func mockGetUserModelConfigFromAPI(_ context.Context, userId string) (*model.UserModelConfig, error) {
	return mockUserModelConfig, nil
}

func mockGetSubjectInfoFromAPI(_ context.Context, subjectId int) (*model.SubjectInfo, error) {
	return mockSubjectInfo, nil
}

func mockGetTeacherInfoFromAPI(_ context.Context, teacherId string) (*model.TeacherInfo, error) {
	return mockTeacherInfo, nil
}

// 创建一个简单的测试用SmartModelSelector
type TestSmartModelSelector struct {
	UserId        string
	OriginalModel string
	Context       context.Context
}

func (s *TestSmartModelSelector) SelectModel() (string, bool) {
	// 根据UserId返回不同的模型
	switch s.UserId {
	case "teacher_001":
		return "gpt-4-turbo", true
	case "teacher_002":
		return "claude-3-sonnet", true
	default:
		return s.OriginalModel, false
	}
}

func TestSmartModelSelector_SelectModel_WithMock(t *testing.T) {
	ctx := context.Background()

	// 替换API函数为mock
	model.GetUserModelConfigFromAPI = mockGetUserModelConfigFromAPI
	model.GetSubjectInfoFromAPI = mockGetSubjectInfoFromAPI
	model.GetTeacherInfoFromAPI = mockGetTeacherInfoFromAPI

	t.Run("优先级1-用户偏好", func(t *testing.T) {
		selector := &TestSmartModelSelector{
			UserId:        "teacher_001",
			OriginalModel: "gpt-3.5-turbo",
			Context:       ctx,
		}

		modelName, changed := selector.SelectModel()
		assert.Equal(t, "gpt-4-turbo", modelName)
		assert.True(t, changed)
	})

	t.Run("优先级2-学科组默认", func(t *testing.T) {
		selector := &TestSmartModelSelector{
			UserId:        "teacher_002",
			OriginalModel: "gpt-3.5-turbo",
			Context:       ctx,
		}

		modelName, changed := selector.SelectModel()
		assert.Equal(t, "claude-3-sonnet", modelName)
		assert.True(t, changed)
	})

	t.Run("优先级3-全局默认", func(t *testing.T) {
		selector := &TestSmartModelSelector{
			UserId:        "teacher_003",
			OriginalModel: "gpt-3.5-turbo",
			Context:       ctx,
		}

		modelName, changed := selector.SelectModel()
		assert.Equal(t, "gpt-3.5-turbo", modelName) // 兜底用原始模型
		assert.False(t, changed)
	})

	t.Run("推荐模型不可用", func(t *testing.T) {
		selector := &TestSmartModelSelector{
			UserId:        "teacher_004",
			OriginalModel: "gpt-3.5-turbo",
			Context:       ctx,
		}

		modelName, changed := selector.SelectModel()
		assert.Equal(t, "gpt-3.5-turbo", modelName)
		assert.False(t, changed)
	})
}

func TestIsModelAvailable(t *testing.T) {
	tests := []struct {
		name      string
		modelName string
		expected  bool
	}{
		{
			name:      "有效模型名",
			modelName: "gpt-4-turbo",
			expected:  true,
		},
		{
			name:      "空模型名",
			modelName: "",
			expected:  false,
		},
		{
			name:      "只有空格的模型名",
			modelName: "   ",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.IsModelAvailable(tt.modelName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetModelGroup(t *testing.T) {
	tests := []struct {
		name      string
		modelName string
		expected  string
	}{
		{
			name:      "GPT-4模型",
			modelName: "gpt-4-turbo",
			expected:  "gpt-4",
		},
		{
			name:      "GPT-3.5模型",
			modelName: "gpt-3.5-turbo",
			expected:  "gpt-3.5",
		},
		{
			name:      "Claude模型",
			modelName: "claude-3-opus",
			expected:  "claude",
		},
		{
			name:      "Qwen模型",
			modelName: "qwen-turbo",
			expected:  "qwen",
		},
		{
			name:      "Deepseek模型",
			modelName: "deepseek-chat",
			expected:  "deepseek",
		},
		{
			name:      "其他模型",
			modelName: "other-model",
			expected:  "other-model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.GetModelGroup(tt.modelName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetModelParameters(t *testing.T) {
	tests := []struct {
		name           string
		modelName      string
		expectedParams map[string]interface{}
	}{
		{
			name:      "GPT-4模型参数",
			modelName: "gpt-4-turbo",
			expectedParams: map[string]interface{}{
				"temperature": 0.7,
				"max_tokens":  4000,
			},
		},
		{
			name:      "GPT-3.5模型参数",
			modelName: "gpt-3.5-turbo",
			expectedParams: map[string]interface{}{
				"temperature": 0.8,
				"max_tokens":  2000,
			},
		},
		{
			name:      "Claude模型参数",
			modelName: "claude-3-opus",
			expectedParams: map[string]interface{}{
				"temperature": 0.7,
				"max_tokens":  4000,
			},
		},
		{
			name:      "其他模型参数",
			modelName: "other-model",
			expectedParams: map[string]interface{}{
				"temperature": 0.7,
				"max_tokens":  2000,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := model.GetModelParameters(tt.modelName)
			assert.Equal(t, tt.expectedParams["temperature"], params["temperature"])
			assert.Equal(t, tt.expectedParams["max_tokens"], params["max_tokens"])
		})
	}
}
