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
