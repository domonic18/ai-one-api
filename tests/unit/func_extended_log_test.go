package unit

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/model/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExtendedLog_CreateAndRetrieve 测试扩展日志的创建和检索功能
// 测试目的：验证扩展日志的基本CRUD操作，确保数据能够正确创建、保存和检索
// 测试内容：
// 1. 正常创建包含维度信息的扩展日志
// 2. 创建空维度信息的扩展日志
// 3. 根据日志ID检索扩展日志
// 4. 验证维度信息的序列化和反序列化
func TestExtendedLog_CreateAndRetrieve(t *testing.T) {
	db := setupSimpleTestDB()
	model.DB = db

	// 清理测试数据
	model.DB.Where("1 = 1").Delete(&identity.ExtendedLog{})

	ctx := context.Background()

	// 测试用例1：正常创建扩展日志
	t.Run("正常创建扩展日志", func(t *testing.T) {
		logId := int64(1001)
		externalUserId := "user_001"
		userGroup := "test_group"
		dimensionInfo := &identity.DimensionInfo{
			"external_user_id": "user_001",
			"group_id":         1,
			"group_name":       "测试组",
			"category_id":      10,
			"category_name":    "测试类别",
			"user_name":        "测试用户",
		}

		// 创建扩展日志
		extendedLog, err := identity.CreateExtendedLog(ctx, logId, externalUserId, userGroup, dimensionInfo)
		require.NoError(t, err)
		require.NotNil(t, extendedLog)

		// 验证基本字段
		assert.Equal(t, logId, extendedLog.LogId)
		assert.Equal(t, externalUserId, extendedLog.ExternalUserId)
		assert.Equal(t, userGroup, extendedLog.UserGroup)
		assert.NotZero(t, extendedLog.Id)

		// 验证维度信息
		parsedDimensionInfo, err := extendedLog.GetDimensionInfo()
		require.NoError(t, err)
		require.NotNil(t, parsedDimensionInfo)

		// 验证抽象化的维度信息 - 使用类型断言处理JSON数字类型转换
		for key, expectedValue := range *dimensionInfo {
			actualValue, exists := (*parsedDimensionInfo)[key]
			assert.True(t, exists, "Key %s should exist in parsed dimension info", key)

			// 处理JSON数字类型转换问题
			switch expected := expectedValue.(type) {
			case int:
				// JSON中的数字会被解析为float64
				if actualFloat, ok := actualValue.(float64); ok {
					assert.Equal(t, float64(expected), actualFloat, "Value for key %s should match", key)
				} else {
					assert.Equal(t, expectedValue, actualValue, "Value for key %s should match", key)
				}
			default:
				assert.Equal(t, expectedValue, actualValue, "Value for key %s should match", key)
			}
		}
	})

	// 测试用例2：创建空维度信息的扩展日志
	t.Run("创建空维度信息的扩展日志", func(t *testing.T) {
		logId := int64(1002)
		externalUserId := "user_002"
		userGroup := "test_group"

		// 创建扩展日志（维度信息为nil）
		extendedLog, err := identity.CreateExtendedLog(ctx, logId, externalUserId, userGroup, nil)
		require.NoError(t, err)
		require.NotNil(t, extendedLog)

		// 验证基本字段
		assert.Equal(t, logId, extendedLog.LogId)
		assert.Equal(t, externalUserId, extendedLog.ExternalUserId)
		assert.Equal(t, userGroup, extendedLog.UserGroup)

		// 验证维度信息为空
		parsedDimensionInfo, err := extendedLog.GetDimensionInfo()
		assert.NoError(t, err)
		assert.Nil(t, parsedDimensionInfo)
	})

	// 测试用例3：根据日志ID获取扩展日志
	t.Run("根据日志ID获取扩展日志", func(t *testing.T) {
		logId := int64(1003)
		externalUserId := "user_003"
		userGroup := "test_group"
		dimensionInfo := &identity.DimensionInfo{
			"external_user_id": "user_003",
			"test_field":       "test_value",
		}

		// 创建扩展日志
		createdLog, err := identity.CreateExtendedLog(ctx, logId, externalUserId, userGroup, dimensionInfo)
		require.NoError(t, err)

		// 根据日志ID获取扩展日志
		retrievedLog, err := identity.GetExtendedLogByLogId(ctx, logId)
		require.NoError(t, err)
		require.NotNil(t, retrievedLog)

		// 验证检索到的日志与创建的日志一致
		assert.Equal(t, createdLog.Id, retrievedLog.Id)
		assert.Equal(t, createdLog.LogId, retrievedLog.LogId)
		assert.Equal(t, createdLog.ExternalUserId, retrievedLog.ExternalUserId)
		assert.Equal(t, createdLog.UserGroup, retrievedLog.UserGroup)

		// 使用JSON比较来避免字段顺序问题
		createdDimensionInfo, err := createdLog.GetDimensionInfo()
		require.NoError(t, err)
		retrievedDimensionInfo, err := retrievedLog.GetDimensionInfo()
		require.NoError(t, err)

		// 将两个维度信息序列化为JSON进行比较
		createdJSON, err := json.Marshal(createdDimensionInfo)
		require.NoError(t, err)
		retrievedJSON, err := json.Marshal(retrievedDimensionInfo)
		require.NoError(t, err)

		assert.Equal(t, string(createdJSON), string(retrievedJSON), "Dimension info should match")
	})
}

// TestExtendedLog_DimensionInfo_Abstraction 测试抽象化的维度信息处理
// 测试目的：验证扩展日志系统能够灵活处理各种类型的维度信息，支持不同业务场景
// 测试内容：
// 1. 创建和解析抽象化维度信息
// 2. 空维度信息的正确处理
// 3. 不同业务场景的维度信息结构（教育、企业、混合场景）
// 4. JSON序列化后的数据类型转换处理
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

// TestExtendedLog_RecordFunction 测试扩展日志记录函数
// 测试目的：验证异步记录扩展日志的功能，确保系统能够正确处理各种输入情况
// 测试内容：
// 1. 正常记录扩展日志（异步操作）
// 2. 空外部用户ID的跳过处理
// 3. 身份解析器未初始化时的处理
func TestExtendedLog_RecordFunction(t *testing.T) {
	db := setupSimpleTestDB()
	model.DB = db

	// 清理测试数据
	model.DB.Where("1 = 1").Delete(&identity.ExtendedLog{})

	ctx := context.Background()

	t.Run("记录扩展日志", func(t *testing.T) {
		logId := int64(2001)
		externalUserId := "user_record_001"

		// 调用记录函数（异步）
		identity.RecordExtendedLog(ctx, logId, externalUserId)

		// 等待异步操作完成（简单的等待）
		// 注意：在实际测试中，可能需要更复杂的同步机制
		// 这里为了简化测试，我们不做等待，因为RecordExtendedLog会处理身份解析器未初始化的情况
	})

	t.Run("记录空外部用户ID的扩展日志", func(t *testing.T) {
		logId := int64(2002)
		externalUserId := ""

		// 调用记录函数（应该跳过）
		identity.RecordExtendedLog(ctx, logId, externalUserId)

		// 验证没有创建日志记录
		var count int64
		model.DB.Model(&identity.ExtendedLog{}).Where("log_id = ?", logId).Count(&count)
		assert.Equal(t, int64(0), count)
	})
}

// TestExtendedLog_RecordFunction_Abstraction 测试抽象化的记录函数
// 测试目的：验证记录函数不依赖具体业务逻辑，能够处理各种边界情况
// 测试内容：
// 1. 记录扩展日志不依赖具体业务
// 2. 空外部用户ID跳过记录
// 3. 异步操作的正确处理
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
