package cache

// 导出缓存功能的初始化文件

// Init 初始化所有缓存组件
func Init() {
	// 初始化缓存管理器
	InitManager()

	// 初始化缓存失效管理器
	InitInvalidationManager()

	// 初始化缓存监控器
	InitMonitor()
}
