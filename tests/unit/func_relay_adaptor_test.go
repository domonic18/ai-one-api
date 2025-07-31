package unit

import (
	"testing"

	"github.com/songquanpeng/one-api/relay"
	"github.com/songquanpeng/one-api/relay/apitype"
)

// TestRelay_GetAdaptor 测试GetAdaptor函数
func TestRelay_GetAdaptor(t *testing.T) {
	tests := []struct {
		name     string
		apiType  int
		expected bool // 是否期望返回非nil的adaptor
	}{
		{
			name:     "AIProxyLibrary类型",
			apiType:  apitype.AIProxyLibrary,
			expected: true,
		},
		{
			name:     "Ali类型",
			apiType:  apitype.Ali,
			expected: true,
		},
		{
			name:     "Anthropic类型",
			apiType:  apitype.Anthropic,
			expected: true,
		},
		{
			name:     "AwsClaude类型",
			apiType:  apitype.AwsClaude,
			expected: true,
		},
		{
			name:     "Baidu类型",
			apiType:  apitype.Baidu,
			expected: true,
		},
		{
			name:     "Gemini类型",
			apiType:  apitype.Gemini,
			expected: true,
		},
		{
			name:     "OpenAI类型",
			apiType:  apitype.OpenAI,
			expected: true,
		},
		{
			name:     "PaLM类型",
			apiType:  apitype.PaLM,
			expected: true,
		},
		{
			name:     "Tencent类型",
			apiType:  apitype.Tencent,
			expected: true,
		},
		{
			name:     "Xunfei类型",
			apiType:  apitype.Xunfei,
			expected: true,
		},
		{
			name:     "Zhipu类型",
			apiType:  apitype.Zhipu,
			expected: true,
		},
		{
			name:     "Ollama类型",
			apiType:  apitype.Ollama,
			expected: true,
		},
		{
			name:     "Coze类型",
			apiType:  apitype.Coze,
			expected: true,
		},
		{
			name:     "Cohere类型",
			apiType:  apitype.Cohere,
			expected: true,
		},
		{
			name:     "Cloudflare类型",
			apiType:  apitype.Cloudflare,
			expected: true,
		},
		{
			name:     "DeepL类型",
			apiType:  apitype.DeepL,
			expected: true,
		},
		{
			name:     "VertexAI类型",
			apiType:  apitype.VertexAI,
			expected: true,
		},
		{
			name:     "Proxy类型",
			apiType:  apitype.Proxy,
			expected: true,
		},
		{
			name:     "Replicate类型",
			apiType:  apitype.Replicate,
			expected: true,
		},
		{
			name:     "无效类型",
			apiType:  apitype.Dummy,
			expected: false,
		},
		{
			name:     "超出范围类型",
			apiType:  999,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adaptor := relay.GetAdaptor(tt.apiType)

			if tt.expected {
				if adaptor == nil {
					t.Errorf("GetAdaptor(%d) = nil, 期望非nil", tt.apiType)
				}
			} else {
				if adaptor != nil {
					t.Errorf("GetAdaptor(%d) = %v, 期望nil", tt.apiType, adaptor)
				}
			}
		})
	}
}

// TestRelay_GetAdaptor_边界值测试
func TestRelay_GetAdaptor_边界值测试(t *testing.T) {
	tests := []struct {
		name     string
		apiType  int
		expected bool
	}{
		{
			name:     "最小值",
			apiType:  0,
			expected: true, // OpenAI = 0，应该返回有效的adaptor
		},
		{
			name:     "最大值",
			apiType:  apitype.Dummy - 1,
			expected: true,
		},
		{
			name:     "负数",
			apiType:  -1,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adaptor := relay.GetAdaptor(tt.apiType)

			if tt.expected {
				if adaptor == nil {
					t.Errorf("GetAdaptor(%d) = nil, 期望非nil", tt.apiType)
				}
			} else {
				if adaptor != nil {
					t.Errorf("GetAdaptor(%d) = %v, 期望nil", tt.apiType, adaptor)
				}
			}
		})
	}
}
