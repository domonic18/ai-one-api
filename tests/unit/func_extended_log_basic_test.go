package unit

import (
	"context"
	"testing"
	"time"

	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/model/identity"
	"github.com/songquanpeng/one-api/tests/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupExtendedLogTestDB() *gorm.DB {
	return common.SetupMySQLTestDB()
}

// TestExtendedLog_BasicCRUD 测试扩展日志的基本CRUD操作
func TestExtendedLog_BasicCRUD(t *testing.T) {
	db := setupExtendedLogTestDB()
	model.DB = db
	common.CleanupTestDB(db)

	ctx := context.Background()

	// 准备测试数据
	logId := int64(1001)
	externalUserId := "teacher_001"
	userGroup := "beijing_math_group"
	dimensionInfo := &identity.DimensionInfo{
		"school_id":    1,
		"school_name":  "北京中学",
		"subject_id":   10,
		"subject_name": "数学组",
		"teacher_name": "张老师",
	}

	// 创建扩展日志
	extendedLog, err := identity.CreateExtendedLog(ctx, logId, externalUserId, userGroup, dimensionInfo)
	require.NoError(t, err)
	assert.NotNil(t, extendedLog)
	assert.Equal(t, logId, extendedLog.LogId)
	assert.Equal(t, externalUserId, extendedLog.ExternalUserId)
	assert.Equal(t, userGroup, extendedLog.UserGroup)

	// 验证维度信息
	parsedDimensionInfo, err := extendedLog.GetDimensionInfo()
	require.NoError(t, err)
	assert.NotNil(t, parsedDimensionInfo)
	assert.Equal(t, "北京中学", (*parsedDimensionInfo)["school_name"])
	assert.Equal(t, "数学组", (*parsedDimensionInfo)["subject_name"])
	assert.Equal(t, "张老师", (*parsedDimensionInfo)["teacher_name"])

	// 根据日志ID检索
	retrievedLog, err := identity.GetExtendedLogByLogId(ctx, logId)
	require.NoError(t, err)
	assert.NotNil(t, retrievedLog)
	assert.Equal(t, extendedLog.Id, retrievedLog.Id)
	assert.Equal(t, logId, retrievedLog.LogId)
	assert.Equal(t, externalUserId, retrievedLog.ExternalUserId)
	assert.Equal(t, userGroup, retrievedLog.UserGroup)
}

// TestExtendedLog_QueryByExternalUserId 测试根据外部用户ID查询
func TestExtendedLog_QueryByExternalUserId(t *testing.T) {
	db := setupExtendedLogTestDB()
	model.DB = db
	common.CleanupTestDB(db)

	ctx := context.Background()

	// 创建多个测试数据
	testData := []struct {
		logId          int64
		externalUserId string
		userGroup      string
	}{
		{1002, "teacher_001", "beijing_math_group"},
		{1003, "teacher_001", "beijing_math_group"},
		{1004, "teacher_002", "beijing_chinese_group"},
	}

	for _, data := range testData {
		dimensionInfo := &identity.DimensionInfo{
			"school_name":  "北京中学",
			"subject_name": "测试学科",
		}
		_, err := identity.CreateExtendedLog(ctx, data.logId, data.externalUserId, data.userGroup, dimensionInfo)
		require.NoError(t, err)
	}

	// 获取teacher_001的扩展日志
	logs, total, err := identity.GetExtendedLogsByExternalUserId(ctx, "teacher_001", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, logs, 2)

	// 验证返回的日志都属于teacher_001
	for _, log := range logs {
		assert.Equal(t, "teacher_001", log.ExternalUserId)
	}
}

// TestExtendedLog_QueryByUserGroup 测试根据用户组查询
func TestExtendedLog_QueryByUserGroup(t *testing.T) {
	db := setupExtendedLogTestDB()
	model.DB = db
	common.CleanupTestDB(db)

	ctx := context.Background()

	// 创建多个测试数据
	testData := []struct {
		logId          int64
		externalUserId string
		userGroup      string
	}{
		{1005, "teacher_003", "beijing_math_group"},
		{1006, "teacher_004", "beijing_math_group"},
		{1007, "teacher_005", "beijing_chinese_group"},
	}

	for _, data := range testData {
		dimensionInfo := &identity.DimensionInfo{
			"school_name":  "北京中学",
			"subject_name": "测试学科",
		}
		_, err := identity.CreateExtendedLog(ctx, data.logId, data.externalUserId, data.userGroup, dimensionInfo)
		require.NoError(t, err)
	}

	// 获取beijing_math_group组的扩展日志
	logs, total, err := identity.GetExtendedLogsByUserGroup(ctx, "beijing_math_group", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, logs, 2)

	// 验证返回的日志都属于beijing_math_group组
	for _, log := range logs {
		assert.Equal(t, "beijing_math_group", log.UserGroup)
	}
}

// TestExtendedLog_DimensionInfoHandling 测试维度信息处理
func TestExtendedLog_DimensionInfoHandling(t *testing.T) {
	db := setupExtendedLogTestDB()
	model.DB = db
	common.CleanupTestDB(db)

	ctx := context.Background()

	// 测试复杂的维度信息
	complexDimensionInfo := &identity.DimensionInfo{
		"school_id":    1,
		"school_name":  "北京中学",
		"subject_id":   10,
		"subject_name": "数学组",
		"teacher_name": "张老师",
		"department":   "教学部",
		"region":       "华北",
		"project":      "AI教学项目",
		"metadata": map[string]interface{}{
			"level":      "高级",
			"experience": 5,
			"active":     true,
		},
	}

	logId := int64(1008)
	extendedLog, err := identity.CreateExtendedLog(ctx, logId, "teacher_006", "beijing_math_group", complexDimensionInfo)
	require.NoError(t, err)
	assert.NotNil(t, extendedLog)

	// 解析维度信息
	parsedDimensionInfo, err := extendedLog.GetDimensionInfo()
	require.NoError(t, err)
	assert.NotNil(t, parsedDimensionInfo)

	// 验证基本字段
	assert.Equal(t, "北京中学", (*parsedDimensionInfo)["school_name"])
	assert.Equal(t, "数学组", (*parsedDimensionInfo)["subject_name"])
	assert.Equal(t, "张老师", (*parsedDimensionInfo)["teacher_name"])
	assert.Equal(t, "教学部", (*parsedDimensionInfo)["department"])
	assert.Equal(t, "华北", (*parsedDimensionInfo)["region"])
	assert.Equal(t, "AI教学项目", (*parsedDimensionInfo)["project"])

	// 验证嵌套的metadata字段
	metadata, ok := (*parsedDimensionInfo)["metadata"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "高级", metadata["level"])
	assert.Equal(t, float64(5), metadata["experience"])
	assert.Equal(t, true, metadata["active"])
}

// TestExtendedLog_EmptyDimensionInfo 测试空维度信息处理
func TestExtendedLog_EmptyDimensionInfo(t *testing.T) {
	db := setupExtendedLogTestDB()
	model.DB = db
	common.CleanupTestDB(db)

	ctx := context.Background()

	// 测试nil维度信息
	logId := int64(1009)
	extendedLog, err := identity.CreateExtendedLog(ctx, logId, "teacher_007", "default_group", nil)
	require.NoError(t, err)
	assert.NotNil(t, extendedLog)

	// 解析维度信息
	parsedDimensionInfo, err := extendedLog.GetDimensionInfo()
	require.NoError(t, err)
	assert.Nil(t, parsedDimensionInfo) // 应该返回nil
}

// TestExtendedLog_DeleteByLogIds 测试根据日志ID删除
func TestExtendedLog_DeleteByLogIds(t *testing.T) {
	db := setupExtendedLogTestDB()
	model.DB = db
	common.CleanupTestDB(db)

	ctx := context.Background()

	// 创建测试数据
	logId := int64(1010)
	dimensionInfo := &identity.DimensionInfo{
		"school_name": "测试学校",
	}
	_, err := identity.CreateExtendedLog(ctx, logId, "teacher_test", "test_group", dimensionInfo)
	require.NoError(t, err)

	// 验证数据已创建
	log, err := identity.GetExtendedLogByLogId(ctx, logId)
	require.NoError(t, err)
	assert.NotNil(t, log)

	// 删除数据
	err = identity.DeleteExtendedLogsByLogIds(ctx, []int64{logId})
	require.NoError(t, err)

	// 验证数据已删除
	log, err = identity.GetExtendedLogByLogId(ctx, logId)
	require.NoError(t, err)
	assert.Nil(t, log)
}

// TestExtendedLog_Pagination 测试分页功能
func TestExtendedLog_Pagination(t *testing.T) {
	db := setupExtendedLogTestDB()
	model.DB = db
	common.CleanupTestDB(db)

	ctx := context.Background()

	// 创建多个测试数据
	for i := 0; i < 5; i++ {
		dimensionInfo := &identity.DimensionInfo{
			"index": i,
		}
		_, err := identity.CreateExtendedLog(ctx, int64(2000+i), "teacher_pagination", "pagination_group", dimensionInfo)
		require.NoError(t, err)
	}

	// 测试第一页（每页2条）
	logs, total, err := identity.GetExtendedLogsByExternalUserId(ctx, "teacher_pagination", 1, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, logs, 2)

	// 测试第二页
	logs, total, err = identity.GetExtendedLogsByExternalUserId(ctx, "teacher_pagination", 2, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, logs, 2)

	// 测试第三页
	logs, total, err = identity.GetExtendedLogsByExternalUserId(ctx, "teacher_pagination", 3, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, logs, 1)
}

// TestExtendedLog_Ordering 测试排序功能
func TestExtendedLog_Ordering(t *testing.T) {
	db := setupExtendedLogTestDB()
	model.DB = db
	common.CleanupTestDB(db)

	ctx := context.Background()

	// 创建测试数据，使用不同的创建时间
	time1 := time.Now().Add(-2 * time.Hour)
	time2 := time.Now().Add(-1 * time.Hour)
	time3 := time.Now()

	testData := []struct {
		logId     int64
		createdAt time.Time
	}{
		{3001, time1},
		{3002, time2},
		{3003, time3},
	}

	for _, data := range testData {
		dimensionInfo := &identity.DimensionInfo{
			"created_at": data.createdAt.Format(time.RFC3339),
		}
		_, err := identity.CreateExtendedLog(ctx, data.logId, "teacher_order", "order_group", dimensionInfo)
		require.NoError(t, err)
	}

	// 获取日志（应该按创建时间倒序排列）
	logs, _, err := identity.GetExtendedLogsByExternalUserId(ctx, "teacher_order", 1, 10)
	require.NoError(t, err)
	assert.Len(t, logs, 3)

	// 验证排序（最新的在前）
	assert.True(t, logs[0].CreatedAt.After(logs[1].CreatedAt))
	assert.True(t, logs[1].CreatedAt.After(logs[2].CreatedAt))
}

// TestExtendedLog_InvalidData 测试无效数据处理
func TestExtendedLog_InvalidData(t *testing.T) {
	db := setupExtendedLogTestDB()
	model.DB = db
	common.CleanupTestDB(db)

	ctx := context.Background()

	// 测试空外部用户ID
	_, err := identity.CreateExtendedLog(ctx, 4001, "", "test_group", nil)
	require.NoError(t, err) // 应该允许空外部用户ID

	// 测试空用户组
	_, err = identity.CreateExtendedLog(ctx, 4002, "teacher_invalid", "", nil)
	require.NoError(t, err) // 应该允许空用户组

	// 测试无效的日志ID
	log, err := identity.GetExtendedLogByLogId(ctx, 99999)
	require.NoError(t, err)
	assert.Nil(t, log) // 应该返回nil而不是错误
}

// TestExtendedLog_JSONHandling 测试JSON字段处理
func TestExtendedLog_JSONHandling(t *testing.T) {
	db := setupExtendedLogTestDB()
	model.DB = db
	common.CleanupTestDB(db)

	ctx := context.Background()

	// 测试特殊字符和Unicode
	specialDimensionInfo := &identity.DimensionInfo{
		"special_chars": "测试中文 & < > \" ' \\ /",
		"unicode":       "🎉🚀💻",
		"numbers":       []int{1, 2, 3, 4, 5},
		"nested": map[string]interface{}{
			"level1": map[string]interface{}{
				"level2": "deep nested",
			},
		},
	}

	logId := int64(5001)
	extendedLog, err := identity.CreateExtendedLog(ctx, logId, "teacher_special", "special_group", specialDimensionInfo)
	require.NoError(t, err)
	assert.NotNil(t, extendedLog)

	// 解析维度信息
	parsedDimensionInfo, err := extendedLog.GetDimensionInfo()
	require.NoError(t, err)
	assert.NotNil(t, parsedDimensionInfo)

	// 验证特殊字符
	assert.Equal(t, "测试中文 & < > \" ' \\ /", (*parsedDimensionInfo)["special_chars"])
	assert.Equal(t, "🎉🚀💻", (*parsedDimensionInfo)["unicode"])

	// 验证数组
	numbers, ok := (*parsedDimensionInfo)["numbers"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, numbers, 5)
	assert.Equal(t, float64(1), numbers[0])
	assert.Equal(t, float64(5), numbers[4])

	// 验证嵌套结构
	nested, ok := (*parsedDimensionInfo)["nested"].(map[string]interface{})
	assert.True(t, ok)
	level1, ok := nested["level1"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "deep nested", level1["level2"])
}
