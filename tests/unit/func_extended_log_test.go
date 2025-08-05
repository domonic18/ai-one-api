package unit

import (
	"context"
	"testing"

	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/model/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExtendedLog_DimensionInfo_Abstraction 测试抽象化的维度信息
func TestExtendedLog_DimensionInfo_Abstraction(t *testing.T) {
	db := setupSimpleTestDB()
	model.DB = db

	// 清理测试数据
	model.DB.Where("1 = 1").Delete(&identity.ExtendedLog{})

	ctx := context.Background()

	t.Run("创建和解析抽象化维度信息", func(t *testing.T) {
		// 创建抽象化的维度信息
		dimensionInfo := &identity.DimensionInfo{
			"external_user_id": "user_001",
			"group_id":         float64(1), // JSON序列化后数字会变成float64
			"group_name":       "测试组",
			"category_id":      float64(10), // JSON序列化后数字会变成float64
			"category_name":    "测试类别",
			"user_name":        "测试用户",
			"custom_field":     "自定义值",
		}

		// 创建扩展日志
		extendedLog, err := identity.CreateExtendedLog(ctx, 1001, "user_001", "test_group", dimensionInfo)
		require.NoError(t, err)
		require.NotNil(t, extendedLog)

		// 验证基本字段
		assert.Equal(t, int64(1001), extendedLog.LogId)
		assert.Equal(t, "user_001", extendedLog.ExternalUserId)
		assert.Equal(t, "test_group", extendedLog.UserGroup)

		// 解析维度信息
		parsedDimensionInfo, err := extendedLog.GetDimensionInfo()
		require.NoError(t, err)
		require.NotNil(t, parsedDimensionInfo)

		// 验证所有维度信息都被正确保存和解析
		for key, expectedValue := range *dimensionInfo {
			actualValue, exists := (*parsedDimensionInfo)[key]
			assert.True(t, exists, "Key %s should exist", key)
			assert.Equal(t, expectedValue, actualValue, "Value for key %s should match", key)
		}
	})

	t.Run("空维度信息处理", func(t *testing.T) {
		// 创建没有维度信息的扩展日志
		extendedLog, err := identity.CreateExtendedLog(ctx, 1002, "user_002", "test_group", nil)
		require.NoError(t, err)
		require.NotNil(t, extendedLog)

		// 验证基本字段
		assert.Equal(t, int64(1002), extendedLog.LogId)
		assert.Equal(t, "user_002", extendedLog.ExternalUserId)
		assert.Equal(t, "test_group", extendedLog.UserGroup)

		// 解析维度信息应该返回nil
		parsedDimensionInfo, err := extendedLog.GetDimensionInfo()
		assert.NoError(t, err)
		assert.Nil(t, parsedDimensionInfo)
	})

	t.Run("灵活的维度信息结构", func(t *testing.T) {
		// 测试不同结构的维度信息
		testCases := []struct {
			name          string
			dimensionInfo *identity.DimensionInfo
		}{
			{
				name: "教育场景维度",
				dimensionInfo: &identity.DimensionInfo{
					"school_id":    float64(1), // JSON序列化后数字会变成float64
					"school_name":  "测试学校",
					"subject_id":   float64(10), // JSON序列化后数字会变成float64
					"subject_name": "数学",
				},
			},
			{
				name: "企业场景维度",
				dimensionInfo: &identity.DimensionInfo{
					"department_id":   "IT",
					"department_name": "信息技术部",
					"project_id":      "PRJ001",
					"project_name":    "AI项目",
				},
			},
			{
				name: "混合场景维度",
				dimensionInfo: &identity.DimensionInfo{
					"region":   "北京",
					"level":    float64(5), // JSON序列化后数字会变成float64
					"active":   true,
					"metadata": map[string]interface{}{"nested": "value"},
				},
			},
		}

		for i, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				logId := int64(2000 + i)
				externalUserId := "user_" + string(rune('A'+i))

				// 创建扩展日志
				extendedLog, err := identity.CreateExtendedLog(ctx, logId, externalUserId, "test_group", tc.dimensionInfo)
				require.NoError(t, err)

				// 解析并验证维度信息
				parsedDimensionInfo, err := extendedLog.GetDimensionInfo()
				require.NoError(t, err)
				require.NotNil(t, parsedDimensionInfo)

				// 验证所有字段
				for key, expectedValue := range *tc.dimensionInfo {
					actualValue, exists := (*parsedDimensionInfo)[key]
					assert.True(t, exists, "Key %s should exist in %s", key, tc.name)
					assert.Equal(t, expectedValue, actualValue, "Value for key %s should match in %s", key, tc.name)
				}
			})
		}
	})
}

// TestExtendedLog_RecordFunction_Abstraction 测试抽象化的记录函数
func TestExtendedLog_RecordFunction_Abstraction(t *testing.T) {
	db := setupSimpleTestDB()
	model.DB = db

	// 清理测试数据
	model.DB.Where("1 = 1").Delete(&identity.ExtendedLog{})

	ctx := context.Background()

	t.Run("记录扩展日志不依赖具体业务", func(t *testing.T) {
		logId := int64(3001)
		externalUserId := "abstract_user_001"

		// 调用记录函数（异步）
		identity.RecordExtendedLog(ctx, logId, externalUserId)

		// 注意：RecordExtendedLog是异步的，并且在身份解析器未初始化时会使用基本信息
		// 这个测试主要验证函数不会崩溃，而不是验证具体的记录结果
	})

	t.Run("空外部用户ID跳过记录", func(t *testing.T) {
		logId := int64(3002)
		externalUserId := ""

		// 调用记录函数（应该直接返回）
		identity.RecordExtendedLog(ctx, logId, externalUserId)

		// 验证没有创建任何记录
		var count int64
		db.Model(&identity.ExtendedLog{}).Where("log_id = ?", logId).Count(&count)
		assert.Equal(t, int64(0), count)
	})
}
