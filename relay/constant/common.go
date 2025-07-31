package constant

var StopFinishReason = "stop"
var StreamObject = "chat.completion.chunk"
var NonStreamObject = "chat.completion"

const (
	// HTTP请求头常量
	// 用户身份相关
	UserIdHeader = "X-User-ID"

	// 智能模型选择相关
	SmartModelSelectionHeader = "X-Smart-Model-Selection"

	// 智能模型选择的值
	SmartModelSelectionEnabled  = "true"
	SmartModelSelectionDisabled = "false"

	// 其他系统相关header（预留扩展）
	// RequestIdHeader = "X-Request-ID"
	// TraceIdHeader = "X-Trace-ID"
)
