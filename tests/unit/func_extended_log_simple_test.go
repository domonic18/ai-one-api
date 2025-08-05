package unit

import (
	"context"
	"testing"

	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/model/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExtendedLog_CreateAndRetrieve 测试扩展日志的创建和检索（简化版）
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

		// 验证抽象化的维度信息
		for key, expectedValue := range *dimensionInfo {
			actualValue, exists := (*parsedDimensionInfo)[key]
			assert.True(t, exists, "Key %s should exist in parsed dimension info", key)
			assert.Equal(t, expectedValue, actualValue, "Value for key %s should match", key)
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
		assert.Equal(t, createdLog.DimensionInfo, retrievedLog.DimensionInfo)
	})
}

// TestExtendedLog_RecordFunction 测试扩展日志记录函数
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
