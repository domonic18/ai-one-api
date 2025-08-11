package models

// TeacherInfo 教师信息结构体
type TeacherInfo struct {
	TeacherId      string `json:"teacher_id"`
	TeacherName    string `json:"teacher_name"`
	SchoolId       int    `json:"school_id"`
	SchoolName     string `json:"school_name"`
	SubjectId      int    `json:"subject_id"`
	SubjectName    string `json:"subject_name"`
	GroupName      string `json:"group_name"`
	PreferredModel string `json:"preferred_model"`
}

// TeacherInfoResponse 获取教师信息响应结构体
type TeacherInfoResponse struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Data    *TeacherInfo `json:"data"`
}

// TeacherIdsResponse 获取所有教师ID列表响应结构体
type TeacherIdsResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		TeacherIds []string `json:"teacher_ids"`
		Total      int      `json:"total"`
	} `json:"data"`
}

// BatchTeacherRequest 批量获取教师信息请求结构体
type BatchTeacherRequest struct {
	TeacherIds []string `json:"teacher_ids"`
}

// BatchTeacherResponse 批量获取教师信息响应结构体
type BatchTeacherResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Users        []*TeacherInfo `json:"users"`
		SuccessCount int            `json:"success_count"`
		ErrorCount   int            `json:"error_count"`
	} `json:"data"`
}

// ServerConfig 服务器配置结构体
type ServerConfig struct {
	Server       ServerInfo     `json:"server"`
	API          APIInfo        `json:"api"`
	DefaultUsers []*TeacherInfo `json:"default_users"`
}

// ServerInfo 服务器信息结构体
type ServerInfo struct {
	Port int    `json:"port"`
	Host string `json:"host"`
}

// APIInfo API信息结构体
type APIInfo struct {
	ResponseDelay       string  `json:"response_delay"`
	ErrorRate           float64 `json:"error_rate"`
	APIKey              string  `json:"api_key"`
	OneAPIBaseURL       string  `json:"oneapi_base_url"`
	OneAPIWebhookSecret string  `json:"oneapi_webhook_secret"`
}

// WebUser 用于Web界面的用户结构体
type WebUser struct {
	TeacherId      string `json:"teacher_id" form:"teacher_id"`
	TeacherName    string `json:"teacher_name" form:"teacher_name"`
	SchoolId       int    `json:"school_id" form:"school_id"`
	SchoolName     string `json:"school_name" form:"school_name"`
	SubjectId      int    `json:"subject_id" form:"subject_id"`
	SubjectName    string `json:"subject_name" form:"subject_name"`
	GroupName      string `json:"group_name" form:"group_name"`
	PreferredModel string `json:"preferred_model" form:"preferred_model"`
}

// WebConfig 用于Web界面的配置结构体
type WebConfig struct {
	Server ServerInfo `json:"server" form:"server"`
	API    APIInfo    `json:"api" form:"api"`
}
