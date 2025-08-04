package unit

import (
	"context"
	"testing"
	"time"

	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/model/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExtendedLog_CreateExtendedLog 测试扩展日志创建功能
func TestExtendedLog_CreateExtendedLog(t *testing.T) {
	db := setupSimpleTestDB()
	model.DB = db

	// 清理测试数据
	model.DB.Where("1 = 1").Delete(&identity.ExtendedLog{})

	ctx := context.Background()

	tests := []struct {
		name           string
		logId          int64
		externalUserId string
		userGroup      string
		dimensionInfo  *identity.DimensionInfo
		expectError    bool
	}{
		{
			name:           "正常创建扩展日志",
			logId:          1001,
			externalUserId: "teacher_001",
			userGroup:      "beijing_math_group",
			dimensionInfo: &identity.DimensionInfo{
				SchoolId:    1,
				SchoolName:  "北京中学",
				SubjectId:   10,
				SubjectName: "数学组",
				TeacherName: "张老师",
			},
			expectError: false,
		},
		{
			name:           "创建空维度信息的扩展日志",
			logId:          1002,
			externalUserId: "teacher_002",
			userGroup:      "shanghai_school",
			dimensionInfo:  nil,
			expectError:    false,
		},
		{
			name:           "创建空外部用户ID的扩展日志",
			logId:          1003,
			externalUserId: "",
			userGroup:      "default",
			dimensionInfo:  nil,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建扩展日志
			extendedLog, err := identity.CreateExtendedLog(ctx, tt.logId, tt.externalUserId, tt.userGroup, tt.dimensionInfo)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, extendedLog)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, extendedLog)
				assert.Equal(t, tt.logId, extendedLog.LogId)
				assert.Equal(t, tt.externalUserId, extendedLog.ExternalUserId)
				assert.Equal(t, tt.userGroup, extendedLog.UserGroup)
				assert.NotZero(t, extendedLog.Id)
				assert.False(t, extendedLog.CreatedAt.IsZero())
				assert.False(t, extendedLog.UpdatedAt.IsZero())

				// 验证维度信息
				if tt.dimensionInfo != nil {
					parsedDimensionInfo, err := extendedLog.GetDimensionInfo()
					assert.NoError(t, err)
					assert.NotNil(t, parsedDimensionInfo)
					assert.Equal(t, tt.dimensionInfo.SchoolId, parsedDimensionInfo.SchoolId)
					assert.Equal(t, tt.dimensionInfo.SchoolName, parsedDimensionInfo.SchoolName)
					assert.Equal(t, tt.dimensionInfo.SubjectId, parsedDimensionInfo.SubjectId)
					assert.Equal(t, tt.dimensionInfo.SubjectName, parsedDimensionInfo.SubjectName)
					assert.Equal(t, tt.dimensionInfo.TeacherName, parsedDimensionInfo.TeacherName)
				}
			}
		})
	}
}

// TestExtendedLog_GetExtendedLogByLogId 测试根据日志ID获取扩展日志
func TestExtendedLog_GetExtendedLogByLogId(t *testing.T) {
	db := setupSimpleTestDB()
	model.DB = db

	// 清理测试数据
	model.DB.Where("1 = 1").Delete(&identity.ExtendedLog{})

	ctx := context.Background()

	// 创建测试数据
	logId := int64(2001)
	externalUserId := "teacher_test_001"
	userGroup := "test_group"
	dimensionInfo := &identity.DimensionInfo{
		SchoolId:    2,
		SchoolName:  "测试学校",
		SubjectId:   20,
		SubjectName: "测试学科",
		TeacherName: "测试老师",
	}

	// 创建扩展日志
	createdLog, err := identity.CreateExtendedLog(ctx, logId, externalUserId, userGroup, dimensionInfo)
	require.NoError(t, err)
	require.NotNil(t, createdLog)

	tests := []struct {
		name        string
		logId       int64
		expectFound bool
	}{
		{
			name:        "查找存在的扩展日志",
			logId:       logId,
			expectFound: true,
		},
		{
			name:        "查找不存在的扩展日志",
			logId:       99999,
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extendedLog, err := identity.GetExtendedLogByLogId(ctx, tt.logId)
			assert.NoError(t, err)

			if tt.expectFound {
				assert.NotNil(t, extendedLog)
				assert.Equal(t, tt.logId, extendedLog.LogId)
				assert.Equal(t, externalUserId, extendedLog.ExternalUserId)
				assert.Equal(t, userGroup, extendedLog.UserGroup)
			} else {
				assert.Nil(t, extendedLog)
			}
		})
	}
}

// TestExtendedLog_GetExtendedLogsByExternalUserId 测试根据外部用户ID获取扩展日志
func TestExtendedLog_GetExtendedLogsByExternalUserId(t *testing.T) {
	db := setupSimpleTestDB()
	model.DB = db

	// 清理测试数据
	model.DB.Where("1 = 1").Delete(&identity.ExtendedLog{})

	ctx := context.Background()
	externalUserId := "teacher_pagination_test"
	userGroup := "pagination_test_group"

	// 创建多个测试数据
	for i := 0; i < 15; i++ {
		logId := int64(3000 + i)
		dimensionInfo := &identity.DimensionInfo{
			SchoolId:    3,
			SchoolName:  "分页测试学校",
			SubjectId:   30,
			SubjectName: "分页测试学科",
			TeacherName: "分页测试老师",
		}
		_, err := identity.CreateExtendedLog(ctx, logId, externalUserId, userGroup, dimensionInfo)
		require.NoError(t, err)
	}

	tests := []struct {
		name           string
		externalUserId string
		page           int
		pageSize       int
		expectedCount  int
		expectedTotal  int64
	}{
		{
			name:           "第一页数据",
			externalUserId: externalUserId,
			page:           1,
			pageSize:       10,
			expectedCount:  10,
			expectedTotal:  15,
		},
		{
			name:           "第二页数据",
			externalUserId: externalUserId,
			page:           2,
			pageSize:       10,
			expectedCount:  5,
			expectedTotal:  15,
		},
		{
			name:           "不存在的用户",
			externalUserId: "nonexistent_user",
			page:           1,
			pageSize:       10,
			expectedCount:  0,
			expectedTotal:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logs, total, err := identity.GetExtendedLogsByExternalUserId(ctx, tt.externalUserId, tt.page, tt.pageSize)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCount, len(logs))
			assert.Equal(t, tt.expectedTotal, total)

			// 验证数据排序（应该按创建时间倒序）
			if len(logs) > 1 {
				for i := 0; i < len(logs)-1; i++ {
					assert.True(t, logs[i].CreatedAt.After(logs[i+1].CreatedAt) || logs[i].CreatedAt.Equal(logs[i+1].CreatedAt))
				}
			}
		})
	}
}

// TestExtendedLog_GetExtendedLogsByUserGroup 测试根据用户组获取扩展日志
func TestExtendedLog_GetExtendedLogsByUserGroup(t *testing.T) {
	db := setupSimpleTestDB()
	model.DB = db

	// 清理测试数据
	model.DB.Where("1 = 1").Delete(&identity.ExtendedLog{})

	ctx := context.Background()
	userGroup := "group_test_group"

	// 创建多个测试数据
	for i := 0; i < 8; i++ {
		logId := int64(4000 + i)
		externalUserId := "teacher_group_test_" + string(rune('A'+i))
		dimensionInfo := &identity.DimensionInfo{
			SchoolId:    4,
			SchoolName:  "用户组测试学校",
			SubjectId:   40,
			SubjectName: "用户组测试学科",
			TeacherName: "用户组测试老师" + string(rune('A'+i)),
		}
		_, err := identity.CreateExtendedLog(ctx, logId, externalUserId, userGroup, dimensionInfo)
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		userGroup     string
		page          int
		pageSize      int
		expectedCount int
		expectedTotal int64
	}{
		{
			name:          "获取用户组的所有日志",
			userGroup:     userGroup,
			page:          1,
			pageSize:      10,
			expectedCount: 8,
			expectedTotal: 8,
		},
		{
			name:          "分页获取用户组日志",
			userGroup:     userGroup,
			page:          1,
			pageSize:      5,
			expectedCount: 5,
			expectedTotal: 8,
		},
		{
			name:          "不存在的用户组",
			userGroup:     "nonexistent_group",
			page:          1,
			pageSize:      10,
			expectedCount: 0,
			expectedTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logs, total, err := identity.GetExtendedLogsByUserGroup(ctx, tt.userGroup, tt.page, tt.pageSize)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCount, len(logs))
			assert.Equal(t, tt.expectedTotal, total)

			// 验证所有日志都属于指定用户组
			for _, log := range logs {
				assert.Equal(t, tt.userGroup, log.UserGroup)
			}
		})
	}
}

// TestExtendedLog_DeleteExtendedLogsByLogIds 测试批量删除扩展日志
func TestExtendedLog_DeleteExtendedLogsByLogIds(t *testing.T) {
	db := setupSimpleTestDB()
	model.DB = db

	// 清理测试数据
	model.DB.Where("1 = 1").Delete(&identity.ExtendedLog{})

	ctx := context.Background()

	// 创建测试数据
	var logIds []int64
	for i := 0; i < 5; i++ {
		logId := int64(5000 + i)
		logIds = append(logIds, logId)
		externalUserId := "teacher_delete_test_" + string(rune('A'+i))
		userGroup := "delete_test_group"
		dimensionInfo := &identity.DimensionInfo{
			SchoolId:    5,
			SchoolName:  "删除测试学校",
			SubjectId:   50,
			SubjectName: "删除测试学科",
			TeacherName: "删除测试老师" + string(rune('A'+i)),
		}
		_, err := identity.CreateExtendedLog(ctx, logId, externalUserId, userGroup, dimensionInfo)
		require.NoError(t, err)
	}

	tests := []struct {
		name             string
		logIds           []int64
		expectError      bool
		expectedAffected int
	}{
		{
			name:             "删除存在的日志",
			logIds:           logIds[:3], // 删除前3个
			expectError:      false,
			expectedAffected: 3,
		},
		{
			name:             "删除不存在的日志",
			logIds:           []int64{99990, 99991},
			expectError:      false,
			expectedAffected: 0,
		},
		{
			name:             "删除空列表",
			logIds:           []int64{},
			expectError:      false,
			expectedAffected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := identity.DeleteExtendedLogsByLogIds(ctx, tt.logIds)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// 验证删除结果
				for _, logId := range tt.logIds {
					extendedLog, err := identity.GetExtendedLogByLogId(ctx, logId)
					assert.NoError(t, err)
					if tt.expectedAffected > 0 && logId <= 5002 { // 前3个应该被删除
						assert.Nil(t, extendedLog)
					}
				}
			}
		})
	}
}

// TestExtendedLog_RecordExtendedLog 测试异步记录扩展日志功能
func TestExtendedLog_RecordExtendedLog(t *testing.T) {
	db := setupSimpleTestDB()
	model.DB = db

	// 清理测试数据
	model.DB.Where("1 = 1").Delete(&identity.ExtendedLog{})

	ctx := context.Background()

	tests := []struct {
		name           string
		logId          int64
		externalUserId string
		expectCreated  bool
	}{
		{
			name:           "正常记录扩展日志",
			logId:          6001,
			externalUserId: "teacher_record_test_001",
			expectCreated:  true,
		},
		{
			name:           "空外部用户ID",
			logId:          6002,
			externalUserId: "",
			expectCreated:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用异步记录函数
			identity.RecordExtendedLog(ctx, tt.logId, tt.externalUserId)

			// 等待异步操作完成
			time.Sleep(100 * time.Millisecond)

			// 验证结果
			extendedLog, err := identity.GetExtendedLogByLogId(ctx, tt.logId)
			assert.NoError(t, err)

			if tt.expectCreated {
				assert.NotNil(t, extendedLog)
				assert.Equal(t, tt.logId, extendedLog.LogId)
				assert.Equal(t, tt.externalUserId, extendedLog.ExternalUserId)
			} else {
				assert.Nil(t, extendedLog)
			}
		})
	}
}

// TestExtendedLog_GetDimensionInfo 测试维度信息解析功能
func TestExtendedLog_GetDimensionInfo(t *testing.T) {
	tests := []struct {
		name           string
		dimensionInfo  string
		expectedResult *identity.DimensionInfo
		expectError    bool
	}{
		{
			name:          "正常解析维度信息",
			dimensionInfo: `{"school_id":1,"school_name":"北京中学","subject_id":10,"subject_name":"数学组","teacher_name":"张老师"}`,
			expectedResult: &identity.DimensionInfo{
				SchoolId:    1,
				SchoolName:  "北京中学",
				SubjectId:   10,
				SubjectName: "数学组",
				TeacherName: "张老师",
			},
			expectError: false,
		},
		{
			name:           "空维度信息",
			dimensionInfo:  "",
			expectedResult: nil,
			expectError:    false,
		},
		{
			name:           "无效JSON格式",
			dimensionInfo:  `{"invalid": json}`,
			expectedResult: nil,
			expectError:    true,
		},
		{
			name:          "部分字段的维度信息",
			dimensionInfo: `{"school_name":"测试学校","teacher_name":"测试老师"}`,
			expectedResult: &identity.DimensionInfo{
				SchoolName:  "测试学校",
				TeacherName: "测试老师",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extendedLog := &identity.ExtendedLog{
				DimensionInfo: tt.dimensionInfo,
			}

			result, err := extendedLog.GetDimensionInfo()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.expectedResult == nil {
					assert.Nil(t, result)
				} else {
					assert.NotNil(t, result)
					assert.Equal(t, tt.expectedResult.SchoolId, result.SchoolId)
					assert.Equal(t, tt.expectedResult.SchoolName, result.SchoolName)
					assert.Equal(t, tt.expectedResult.SubjectId, result.SubjectId)
					assert.Equal(t, tt.expectedResult.SubjectName, result.SubjectName)
					assert.Equal(t, tt.expectedResult.TeacherName, result.TeacherName)
				}
			}
		})
	}
}
