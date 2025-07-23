package constant

var StopFinishReason = "stop"
var StreamObject = "chat.completion.chunk"
var NonStreamObject = "chat.completion"

const (
	// SmartSelect 智能模型选择的模型名称
	SmartSelect = "smart_select"

	// UserIdHeader 用户ID的请求头
	UserIdHeader = "X-User-ID"

	// UserModelConfigKeyPrefix Redis中用户模型配置的键前缀
	UserModelConfigKeyPrefix = "user_model_config:"
)
