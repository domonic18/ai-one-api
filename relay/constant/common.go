package constant

var StopFinishReason = "stop"
var StreamObject = "chat.completion.chunk"
var NonStreamObject = "chat.completion"

const (
	// 用户ID请求头
	UserIdHeader = "X-User-ID"

	// 智能模型选择控制头
	SmartModelSelectionHeader = "X-Smart-Model-Selection"

	// Redis键前缀
	UserModelConfigKeyPrefix = "user:model:config:"
)
