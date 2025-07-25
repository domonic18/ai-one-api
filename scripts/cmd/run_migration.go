package main

import (
	"fmt"

	"github.com/songquanpeng/one-api/model"
)

func main() {
	fmt.Println("开始执行数据库迁移...")

	// 初始化数据库连接
	model.InitDB()

	fmt.Println("✓ 数据库连接初始化完成")

	// 执行自动迁移
	fmt.Println("正在执行自动迁移...")

	// 这里会自动执行model/main.go中的migrateDB()函数
	// 该函数会创建extended_logs表

	fmt.Println("✓ 数据库迁移完成")
	fmt.Println("✓ 扩展日志表已创建")

	fmt.Println("\n🎉 数据库迁移执行成功！")
	fmt.Println("\n下一步:")
	fmt.Println("1. 运行数据完整性检查: go run scripts/cmd/data_integrity_check.go")
	fmt.Println("2. 运行迁移测试: go run scripts/cmd/db_migration_test.go")
}
