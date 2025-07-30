package unit

import (
	"context"
	"sync"
	"testing"

	"github.com/songquanpeng/one-api/middleware/identity"
)

// MockIdentityResolver 用于测试的模拟身份解析器
type MockIdentityResolver struct {
	groupMap map[string]string
	modelMap map[string]string
}

func NewMockIdentityResolver() *MockIdentityResolver {
	return &MockIdentityResolver{
		groupMap: map[string]string{
			"teacher_001": "beijing_math_group",
			"teacher_002": "beijing_chinese_group",
		},
		modelMap: map[string]string{
			"teacher_001": "gpt-4",
			"teacher_002": "gemini-pro",
		},
	}
}

func (m *MockIdentityResolver) ResolveGroup(ctx context.Context, externalIdentity string) string {
	if group, exists := m.groupMap[externalIdentity]; exists {
		return group
	}
	return "default"
}

func (m *MockIdentityResolver) ResolveModel(ctx context.Context, externalIdentity string, requestModel string) string {
	if model, exists := m.modelMap[externalIdentity]; exists {
		return model
	}
	return requestModel
}

func TestIdentityResolver_DefaultImplementation_正常场景(t *testing.T) {
	resolver := &identity.DefaultIdentityResolver{}
	ctx := context.Background()

	// 测试ResolveGroup
	group := resolver.ResolveGroup(ctx, "teacher_001")
	if group != "" {
		t.Errorf("Expected empty string, got %s", group)
	}

	// 测试ResolveModel
	originalModel := "gpt-3.5-turbo"
	model := resolver.ResolveModel(ctx, "teacher_001", originalModel)
	if model != originalModel {
		t.Errorf("Expected %s, got %s", originalModel, model)
	}
}

func TestIdentityResolver_全局管理器_设置和获取(t *testing.T) {
	// 保存原始解析器
	originalResolver := identity.GetIdentityResolver()
	defer identity.SetIdentityResolver(originalResolver)

	// 测试设置和获取模拟解析器
	mockResolver := NewMockIdentityResolver()
	identity.SetIdentityResolver(mockResolver)

	retrievedResolver := identity.GetIdentityResolver()
	if retrievedResolver != mockResolver {
		t.Error("Retrieved resolver is not the same as set resolver")
	}

	// 测试解析功能
	ctx := context.Background()
	group := retrievedResolver.ResolveGroup(ctx, "teacher_001")
	if group != "beijing_math_group" {
		t.Errorf("Expected beijing_math_group, got %s", group)
	}

	model := retrievedResolver.ResolveModel(ctx, "teacher_001", "gpt-3.5-turbo")
	if model != "gpt-4" {
		t.Errorf("Expected gpt-4, got %s", model)
	}
}

func TestIdentityResolver_并发访问_线程安全(t *testing.T) {
	mockResolver := NewMockIdentityResolver()
	identity.SetIdentityResolver(mockResolver)

	ctx := context.Background()
	var wg sync.WaitGroup
	results := make(chan string, 100)

	// 并发测试
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resolver := identity.GetIdentityResolver()
			group := resolver.ResolveGroup(ctx, "teacher_001")
			results <- group
		}()
	}

	wg.Wait()
	close(results)

	// 验证所有结果都是一致的
	expectedGroup := "beijing_math_group"
	for group := range results {
		if group != expectedGroup {
			t.Errorf("Expected %s, got %s", expectedGroup, group)
		}
	}
}

func TestIdentityResolver_接口实现_类型检查(t *testing.T) {
	// 确保DefaultIdentityResolver实现了IdentityResolver接口
	var _ identity.IdentityResolver = &identity.DefaultIdentityResolver{}

	// 确保MockIdentityResolver实现了IdentityResolver接口
	var _ identity.IdentityResolver = &MockIdentityResolver{}
}

func TestIdentityResolver_边界值_空值处理(t *testing.T) {
	resolver := &identity.DefaultIdentityResolver{}
	ctx := context.Background()

	// 测试空字符串输入
	group := resolver.ResolveGroup(ctx, "")
	if group != "" {
		t.Errorf("Expected empty string for empty input, got %s", group)
	}

	model := resolver.ResolveModel(ctx, "", "")
	if model != "" {
		t.Errorf("Expected empty string for empty input, got %s", model)
	}

	// 测试nil上下文（虽然不推荐，但要确保不崩溃）
	group = resolver.ResolveGroup(nil, "teacher_001")
	if group != "" {
		t.Errorf("Expected empty string for nil context, got %s", group)
	}
}

func BenchmarkIdentityResolver_DefaultResolver_性能测试(b *testing.B) {
	resolver := &identity.DefaultIdentityResolver{}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resolver.ResolveGroup(ctx, "teacher_001")
		resolver.ResolveModel(ctx, "teacher_001", "gpt-3.5-turbo")
	}
}

func BenchmarkIdentityResolver_GlobalGetter_性能测试(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		identity.GetIdentityResolver()
	}
}
