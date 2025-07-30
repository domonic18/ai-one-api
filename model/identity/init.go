package identity

import (
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/model"
)

// Init 初始化identity包
func Init() {
	// 自动迁移扩展日志表结构
	err := model.DB.AutoMigrate(&ExtendedLog{})
	if err != nil {
		logger.SysLogf("扩展日志表结构迁移失败: %v", err)
	} else {
		logger.SysLog("扩展日志表结构迁移完成 (v3.0)")
	}
}
