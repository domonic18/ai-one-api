package ctxkey

const (
	Config            = "config"
	Id                = "id"
	Username          = "username"
	Role              = "role"
	Status            = "status"
	Channel           = "channel"
	ChannelId         = "channel_id"
	SpecificChannelId = "specific_channel_id"
	RequestModel      = "request_model"
	ConvertedRequest  = "converted_request"
	Group             = "group"
	ModelMapping      = "model_mapping"
	ChannelName       = "channel_name"
	TokenId           = "token_id"
	TokenName         = "token_name"
	BaseURL           = "base_url"
	AvailableModels   = "available_models"
	KeyRequestBody    = "key_request_body"
	SystemPrompt      = "system_prompt"
	LogId             = "log_id" // 日志ID

	// 扩展日志相关上下文键
	SchoolId    = "school_id"    // 学校ID
	SchoolName  = "school_name"  // 学校名称
	SubjectId   = "subject_id"   // 学科组ID
	SubjectName = "subject_name" // 学科组名称
	TeacherId   = "teacher_id"   // 老师ID
	TeacherName = "teacher_name" // 老师姓名

	// 智能模型选择相关上下文键
	OriginalModel = "original_model" // 原始请求中的模型名称
	SelectedModel = "selected_model" // 智能选择的模型名称
)
