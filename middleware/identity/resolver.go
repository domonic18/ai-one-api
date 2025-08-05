package identity

import (
	"context"
	"sync"
)

// IdentityResolver 身份解析器接口
// 极简化的身份解析器接口 - 支持用户组和模型解析
type IdentityResolver interface {
	// ResolveGroup 将外部身份ID解析为OneAPI用户组
	ResolveGroup(ctx context.Context, externalIdentity string) string

	// ResolveModel 将外部身份ID解析为用户偏好模型（可选）
	ResolveModel(ctx context.Context, externalIdentity string, requestModel string) string

	// GetUserDetails 获取用户详细信息用于扩展日志记录（可选）
	// 返回的map包含任意维度信息，如：school_id, school_name, subject_id, subject_name, teacher_name等
	// 如果无法获取详细信息，返回nil
	GetUserDetails(ctx context.Context, externalIdentity string) map[string]interface{}
}

// 全局解析器实例管理
var (
	globalIdentityResolver IdentityResolver = &DefaultIdentityResolver{}
	resolverMutex          sync.RWMutex
)

// GetIdentityResolver 获取当前的身份解析器
func GetIdentityResolver() IdentityResolver {
	resolverMutex.RLock()
	defer resolverMutex.RUnlock()
	return globalIdentityResolver
}

// SetIdentityResolver 设置身份解析器 (可选配置)
func SetIdentityResolver(resolver IdentityResolver) {
	resolverMutex.Lock()
	defer resolverMutex.Unlock()
	globalIdentityResolver = resolver
}
