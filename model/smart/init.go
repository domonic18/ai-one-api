package smart

import (
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/model"
)

// 全局变量
var (
	// 默认智能模型选择器
	DefaultSelector ModelSelector
)

// Init 初始化智能模型选择包
func Init() {
	// 初始化智能模型选择器
	DefaultSelector = NewModelSelector()
	logger.SysLog("智能模型选择器初始化完成")

	// 自动迁移扩展日志表结构
	err := model.DB.AutoMigrate(&ExtendedLog{})
	if err != nil {
		logger.SysLogf("扩展日志表结构迁移失败: %v", err)
	} else {
		logger.SysLog("扩展日志表结构迁移完成")
	}
}
