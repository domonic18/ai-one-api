package identity

import (
	"context"
)

// DefaultIdentityResolver 默认身份解析器实现
// 当课件平台集成未启用时，不做任何处理，系统将完全使用OneAPI原有的用户组和模型选择逻辑
type DefaultIdentityResolver struct{}

// ResolveGroup 解析用户组
// 不做任何处理，返回空字符串，让系统使用原有逻辑
func (d *DefaultIdentityResolver) ResolveGroup(ctx context.Context, externalIdentity string) string {
	// 不做任何处理，返回空字符串，让系统使用原有逻辑
	return ""
}

// ResolveModel 解析用户偏好模型
// 不做任何处理，返回原始模型
func (d *DefaultIdentityResolver) ResolveModel(ctx context.Context, externalIdentity string, requestModel string) string {
	// 不做任何处理，返回原始模型
	return requestModel
}
