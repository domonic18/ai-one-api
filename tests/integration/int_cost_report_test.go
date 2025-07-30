package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/model/identity"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// setupTestDB 设置测试数据库
func setupTestDB() *gorm.DB {
	// 设置测试环境变量
	if os.Getenv("SQL_DSN") == "" {
		os.Setenv("SQL_DSN", "testuser:testpass@tcp(localhost:3306)/oneapi_test?charset=utf8mb4&parseTime=True&loc=Local")
	}

	// 连接MySQL数据库
	db, err := gorm.Open(mysql.Open(os.Getenv("SQL_DSN")), &gorm.Config{
		PrepareStmt:                              true,
		DisableForeignKeyConstraintWhenMigrating: false,
	})
	if err != nil {
		panic("failed to connect to test MySQL database: " + err.Error())
	}

	// 设置数据库连接
	model.DB = db
	model.LOG_DB = db

	// 先删除可能存在的表，避免冲突
	db.Migrator().DropTable(&identity.ExtendedLog{})

	// 执行迁移
	err = db.AutoMigrate(
		&model.User{},
		&model.Token{},
		&model.Channel{},
		&model.Log{},
		&model.Ability{},
		&model.Option{},
		&model.Redemption{},
		&identity.ExtendedLog{},
	)
	if err != nil {
		panic("failed to migrate database: " + err.Error())
	}

	return db
}

// TestCostReport_Integration 测试费用报表功能的集成测试
// 测试目的：验证费用报表功能在实际数据库环境中的完整工作流程
// 测试内容：
// 1. 测试费用报表查询功能的完整流程
// 2. 验证数据库连接和查询的正确性
// 3. 测试各种查询参数的正确处理
// 4. 验证返回数据的格式和结构
func TestCostReport_Integration(t *testing.T) {
	// 设置集成测试环境
	_ = setupTestDB()

	ctx := context.Background()
	startTime := time.Now().AddDate(0, 0, -30)
	endTime := time.Now()

	t.Run("学校维度费用报表", func(t *testing.T) {
		// 测试学校维度费用报表
		reports, err := identity.GetCostReportBySchool(ctx, 1, startTime, endTime)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, reports)

		// 如果没有数据，这是正常的（测试数据库为空）
		if len(reports) == 0 {
			t.Log("查询返回空结果，这在测试环境中是正常的")
			return
		}

		// 验证数据结构
		for _, report := range reports {
			assert.NotEmpty(t, report.Dimension)
			assert.GreaterOrEqual(t, report.TotalCost, 0.0)
			assert.GreaterOrEqual(t, report.TotalTokens, int64(0))
			assert.GreaterOrEqual(t, report.RequestCount, int64(0))
		}
	})

	t.Run("学科组维度费用报表", func(t *testing.T) {
		// 测试学科组维度费用报表
		reports, err := identity.GetCostReportBySubject(ctx, 1, startTime, endTime)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, reports)

		// 如果没有数据，这是正常的（测试数据库为空）
		if len(reports) == 0 {
			t.Log("查询返回空结果，这在测试环境中是正常的")
			return
		}

		// 验证数据结构
		for _, report := range reports {
			assert.NotEmpty(t, report.Dimension)
			assert.GreaterOrEqual(t, report.TotalCost, 0.0)
			assert.GreaterOrEqual(t, report.TotalTokens, int64(0))
			assert.GreaterOrEqual(t, report.RequestCount, int64(0))
		}
	})

	t.Run("老师维度费用报表", func(t *testing.T) {
		// 测试老师维度费用报表
		reports, err := identity.GetCostReportByTeacher(ctx, "teacher_001", startTime, endTime)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, reports)

		// 如果没有数据，这是正常的（测试数据库为空）
		if len(reports) == 0 {
			t.Log("查询返回空结果，这在测试环境中是正常的")
			return
		}

		// 验证数据结构
		for _, report := range reports {
			assert.NotEmpty(t, report.Dimension)
			assert.GreaterOrEqual(t, report.TotalCost, 0.0)
			assert.GreaterOrEqual(t, report.TotalTokens, int64(0))
			assert.GreaterOrEqual(t, report.RequestCount, int64(0))
		}
	})

	t.Run("用户组维度费用报表", func(t *testing.T) {
		// 测试用户组维度费用报表
		reports, err := identity.GetCostReportByUserGroup(ctx, "default", startTime, endTime)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, reports)

		// 如果没有数据，这是正常的（测试数据库为空）
		if len(reports) == 0 {
			t.Log("查询返回空结果，这在测试环境中是正常的")
			return
		}

		// 验证数据结构
		for _, report := range reports {
			assert.NotEmpty(t, report.Dimension)
			assert.GreaterOrEqual(t, report.TotalCost, 0.0)
			assert.GreaterOrEqual(t, report.TotalTokens, int64(0))
			assert.GreaterOrEqual(t, report.RequestCount, int64(0))
		}
	})

	t.Run("综合维度费用报表", func(t *testing.T) {
		// 测试综合维度费用报表
		filters := make(map[string]interface{})
		reports, err := identity.GetComprehensiveCostReport(ctx, filters, startTime, endTime)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, reports)

		// 如果没有数据，这是正常的（测试数据库为空）
		if len(reports) == 0 {
			t.Log("查询返回空结果，这在测试环境中是正常的")
			return
		}

		// 验证数据结构
		for _, report := range reports {
			assert.GreaterOrEqual(t, report.TotalCost, 0.0)
			assert.GreaterOrEqual(t, report.TotalTokens, int64(0))
			assert.GreaterOrEqual(t, report.RequestCount, int64(0))
		}
	})

	t.Run("费用报表汇总", func(t *testing.T) {
		// 测试费用报表汇总
		summary, err := identity.GetCostReportSummary(ctx, startTime, endTime)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, summary)

		// 验证汇总数据包含必要字段
		assert.Contains(t, summary, "total_teachers")
		assert.Contains(t, summary, "total_schools")
		assert.Contains(t, summary, "total_subjects")
		assert.Contains(t, summary, "total_groups")
		assert.Contains(t, summary, "total_cost")
		assert.Contains(t, summary, "total_tokens")
		assert.Contains(t, summary, "total_requests")
	})

	t.Run("学校费用排行", func(t *testing.T) {
		// 测试学校费用排行
		reports, err := identity.GetTopSchoolsByCost(ctx, startTime, endTime, 10)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, reports)
		assert.LessOrEqual(t, len(reports), 10)

		// 如果没有数据，这是正常的（测试数据库为空）
		if len(reports) == 0 {
			t.Log("查询返回空结果，这在测试环境中是正常的")
			return
		}

		// 验证数据结构
		for _, report := range reports {
			assert.NotEmpty(t, report.Dimension)
			assert.GreaterOrEqual(t, report.TotalCost, 0.0)
			assert.GreaterOrEqual(t, report.TotalTokens, int64(0))
			assert.GreaterOrEqual(t, report.RequestCount, int64(0))
		}
	})

	t.Run("老师费用排行", func(t *testing.T) {
		// 测试老师费用排行
		reports, err := identity.GetTopTeachersByCost(ctx, startTime, endTime, 10)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, reports)
		assert.LessOrEqual(t, len(reports), 10)

		// 如果没有数据，这是正常的（测试数据库为空）
		if len(reports) == 0 {
			t.Log("查询返回空结果，这在测试环境中是正常的")
			return
		}

		// 验证数据结构
		for _, report := range reports {
			assert.NotEmpty(t, report.Dimension)
			assert.GreaterOrEqual(t, report.TotalCost, 0.0)
			assert.GreaterOrEqual(t, report.TotalTokens, int64(0))
			assert.GreaterOrEqual(t, report.RequestCount, int64(0))
		}
	})
}

// TestCostReport_Performance 测试费用报表性能
// 测试目的：验证费用报表查询的性能表现，确保在大数据量下的响应时间合理
// 测试内容：
// 1. 测试不同时间范围的查询性能
// 2. 验证查询响应时间的合理性
// 3. 测试大数据量下的系统稳定性
func TestCostReport_Performance(t *testing.T) {
	// 设置集成测试环境
	_ = setupTestDB()

	ctx := context.Background()

	t.Run("短期查询性能", func(t *testing.T) {
		// 测试短期查询性能
		startTime := time.Now().AddDate(0, 0, -7)
		endTime := time.Now()

		start := time.Now()
		_, err := identity.GetCostReportSummary(ctx, startTime, endTime)
		duration := time.Since(start)

		// 验证结果
		assert.NoError(t, err)
		assert.Less(t, duration, 5*time.Second, "短期查询应该在5秒内完成")
	})

	t.Run("长期查询性能", func(t *testing.T) {
		// 测试长期查询性能
		startTime := time.Now().AddDate(0, 0, -90)
		endTime := time.Now()

		start := time.Now()
		_, err := identity.GetCostReportSummary(ctx, startTime, endTime)
		duration := time.Since(start)

		// 验证结果
		assert.NoError(t, err)
		assert.Less(t, duration, 30*time.Second, "长期查询应该在30秒内完成")
	})

	t.Run("综合查询性能", func(t *testing.T) {
		// 测试综合查询性能
		startTime := time.Now().AddDate(0, 0, -30)
		endTime := time.Now()
		filters := make(map[string]interface{})

		start := time.Now()
		reports, err := identity.GetComprehensiveCostReport(ctx, filters, startTime, endTime)
		duration := time.Since(start)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, reports)
		assert.Less(t, duration, 10*time.Second, "综合查询应该在10秒内完成")
	})
}

// TestCostReport_EdgeCases 测试费用报表边界情况
// 测试目的：验证费用报表功能在各种边界情况下的正确性和健壮性
// 测试内容：
// 1. 测试空数据场景的处理
// 2. 测试极端时间范围的处理
// 3. 测试无效参数的处理
// 4. 验证错误处理的正确性
func TestCostReport_EdgeCases(t *testing.T) {
	// 设置集成测试环境
	_ = setupTestDB()

	ctx := context.Background()

	t.Run("未来时间范围", func(t *testing.T) {
		// 测试未来时间范围
		startTime := time.Now().AddDate(0, 0, 1)
		endTime := time.Now().AddDate(0, 0, 2)

		reports, err := identity.GetCostReportBySchool(ctx, 1, startTime, endTime)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, reports)
		assert.Len(t, reports, 0, "未来时间范围应该返回空结果")
	})

	t.Run("无效学校ID", func(t *testing.T) {
		// 测试无效学校ID
		startTime := time.Now().AddDate(0, 0, -30)
		endTime := time.Now()

		reports, err := identity.GetCostReportBySchool(ctx, -1, startTime, endTime)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, reports)
		assert.Len(t, reports, 0, "无效学校ID应该返回空结果")
	})

	t.Run("空老师ID", func(t *testing.T) {
		// 测试空老师ID
		startTime := time.Now().AddDate(0, 0, -30)
		endTime := time.Now()

		reports, err := identity.GetCostReportByTeacher(ctx, "", startTime, endTime)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, reports)
		assert.Len(t, reports, 0, "空老师ID应该返回空结果")
	})

	t.Run("空用户组", func(t *testing.T) {
		// 测试空用户组
		startTime := time.Now().AddDate(0, 0, -30)
		endTime := time.Now()

		reports, err := identity.GetCostReportByUserGroup(ctx, "", startTime, endTime)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, reports)
		assert.Len(t, reports, 0, "空用户组应该返回空结果")
	})

	t.Run("零限制数量", func(t *testing.T) {
		// 测试零限制数量
		startTime := time.Now().AddDate(0, 0, -30)
		endTime := time.Now()

		reports, err := identity.GetTopSchoolsByCost(ctx, startTime, endTime, 0)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, reports)
		assert.Len(t, reports, 0, "零限制数量应该返回空结果")
	})
}
