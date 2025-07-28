package unit

import (
	"context"
	"fmt"
	"testing"

	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtendedLog(t *testing.T) {
	// 设置测试环境
	model.InitDB()
	ctx := context.Background()

	// 清理测试数据
	defer func() {
		model.DB.Exec("DELETE FROM extended_logs")
		model.DB.Exec("DELETE FROM logs")
	}()

	t.Run("创建扩展日志", func(t *testing.T) {
		// 清理可能存在的测试数据
		model.DB.Exec("DELETE FROM extended_logs")
		model.DB.Exec("DELETE FROM logs")

		// 创建测试日志记录
		testLog := &model.Log{
			UserId:            1,
			CreatedAt:         helper.GetTimestamp(),
			Type:              model.LogTypeConsume,
			Content:           "测试日志",
			Username:          "test_user",
			TokenName:         "test_token",
			ModelName:         "gpt-3.5-turbo",
			Quota:             100,
			PromptTokens:      50,
			CompletionTokens:  50,
			ChannelId:         1,
			RequestId:         "test_request_001",
			ElapsedTime:       1000,
			IsStream:          false,
			SystemPromptReset: false,
		}

		err := model.DB.Create(testLog).Error
		require.NoError(t, err)

		// 创建扩展日志信息
		testInfo := &model.ExtendedLogInfo{
			SchoolId:    1,
			SchoolName:  "测试学校",
			SubjectId:   101,
			SubjectName: "数学组",
			TeacherId:   "teacher_001",
			TeacherName: "张老师",
			GroupName:   "数学组",
		}

		// 创建扩展日志
		err = model.CreateExtendedLog(ctx, testLog.Id, testInfo)
		require.NoError(t, err)

		// 查询扩展日志
		extendedLog, err := model.GetExtendedLogByLogId(ctx, testLog.Id)
		require.NoError(t, err)
		require.NotNil(t, extendedLog)

		// 验证字段
		assert.Equal(t, testLog.Id, extendedLog.LogId)
		assert.Equal(t, testInfo.SchoolId, extendedLog.SchoolId)
		assert.Equal(t, testInfo.SchoolName, extendedLog.SchoolName)
		assert.Equal(t, testInfo.SubjectId, extendedLog.SubjectId)
		assert.Equal(t, testInfo.SubjectName, extendedLog.SubjectName)
		assert.Equal(t, testInfo.TeacherId, extendedLog.TeacherId)
		assert.Equal(t, testInfo.TeacherName, extendedLog.TeacherName)
		assert.Equal(t, testInfo.GroupName, extendedLog.GroupName)
		assert.Greater(t, extendedLog.CreatedAt, int64(0))
	})

	t.Run("查询完整日志信息", func(t *testing.T) {
		// 清理可能存在的测试数据
		model.DB.Exec("DELETE FROM extended_logs")
		model.DB.Exec("DELETE FROM logs")

		// 创建测试日志记录
		testLog := &model.Log{
			UserId:            1,
			CreatedAt:         helper.GetTimestamp(),
			Type:              model.LogTypeConsume,
			Content:           "测试日志",
			Username:          "test_user",
			TokenName:         "test_token",
			ModelName:         "gpt-3.5-turbo",
			Quota:             100,
			PromptTokens:      50,
			CompletionTokens:  50,
			ChannelId:         1,
			RequestId:         "test_request_002",
			ElapsedTime:       1000,
			IsStream:          false,
			SystemPromptReset: false,
		}

		err := model.DB.Create(testLog).Error
		require.NoError(t, err)

		// 创建扩展日志信息
		testInfo := &model.ExtendedLogInfo{
			SchoolId:    2,
			SchoolName:  "测试学校2",
			SubjectId:   102,
			SubjectName: "语文组",
			TeacherId:   "teacher_002",
			TeacherName: "李老师",
			GroupName:   "语文组",
		}

		// 创建扩展日志
		err = model.CreateExtendedLog(ctx, testLog.Id, testInfo)
		require.NoError(t, err)

		// 查询完整日志信息
		completeLog, err := model.GetCompleteLogInfo(ctx, testLog.Id)
		require.NoError(t, err)
		require.NotNil(t, completeLog)

		// 验证原日志信息
		assert.Equal(t, testLog.Id, completeLog.Log.Id)
		assert.Equal(t, testLog.UserId, completeLog.Log.UserId)
		assert.Equal(t, testLog.ModelName, completeLog.Log.ModelName)

		// 验证扩展日志信息
		assert.Equal(t, testInfo.SchoolId, completeLog.ExtendedLog.SchoolId)
		assert.Equal(t, testInfo.SchoolName, completeLog.ExtendedLog.SchoolName)
		assert.Equal(t, testInfo.SubjectId, completeLog.ExtendedLog.SubjectId)
		assert.Equal(t, testInfo.TeacherId, completeLog.ExtendedLog.TeacherId)
	})

	t.Run("按老师ID查询扩展日志", func(t *testing.T) {
		// 清理可能存在的测试数据
		model.DB.Exec("DELETE FROM extended_logs")
		model.DB.Exec("DELETE FROM logs")

		// 创建多个测试数据
		teacherId := "teacher_003"
		for i := 1; i <= 3; i++ {
			testLog := &model.Log{
				UserId:            1,
				CreatedAt:         helper.GetTimestamp(),
				Type:              model.LogTypeConsume,
				Content:           "测试日志",
				Username:          "test_user",
				TokenName:         "test_token",
				ModelName:         "gpt-3.5-turbo",
				Quota:             100,
				PromptTokens:      50,
				CompletionTokens:  50,
				ChannelId:         1,
				RequestId:         fmt.Sprintf("test_request_003_%d", i),
				ElapsedTime:       1000,
				IsStream:          false,
				SystemPromptReset: false,
			}

			err := model.DB.Create(testLog).Error
			require.NoError(t, err)

			testInfo := &model.ExtendedLogInfo{
				SchoolId:    3,
				SchoolName:  "测试学校3",
				SubjectId:   103,
				SubjectName: "英语组",
				TeacherId:   teacherId,
				TeacherName: "王老师",
				GroupName:   "英语组",
			}

			err = model.CreateExtendedLog(ctx, testLog.Id, testInfo)
			require.NoError(t, err)
		}

		// 按老师ID查询
		logs, err := model.GetExtendedLogsByTeacherId(ctx, teacherId, 0, 10)
		require.NoError(t, err)
		assert.Len(t, logs, 3)

		// 验证所有记录都属于同一个老师
		for _, log := range logs {
			assert.Equal(t, teacherId, log.TeacherId)
		}
	})

	t.Run("按学科组ID查询扩展日志", func(t *testing.T) {
		// 清理可能存在的测试数据
		model.DB.Exec("DELETE FROM extended_logs")
		model.DB.Exec("DELETE FROM logs")

		// 创建多个测试数据
		subjectId := 104
		for i := 1; i <= 2; i++ {
			testLog := &model.Log{
				UserId:            1,
				CreatedAt:         helper.GetTimestamp(),
				Type:              model.LogTypeConsume,
				Content:           "测试日志",
				Username:          "test_user",
				TokenName:         "test_token",
				ModelName:         "gpt-3.5-turbo",
				Quota:             100,
				PromptTokens:      50,
				CompletionTokens:  50,
				ChannelId:         1,
				RequestId:         fmt.Sprintf("test_request_004_%d", i),
				ElapsedTime:       1000,
				IsStream:          false,
				SystemPromptReset: false,
			}

			err := model.DB.Create(testLog).Error
			require.NoError(t, err)

			testInfo := &model.ExtendedLogInfo{
				SchoolId:    4,
				SchoolName:  "测试学校4",
				SubjectId:   subjectId,
				SubjectName: "历史组",
				TeacherId:   fmt.Sprintf("teacher_004_%d", i),
				TeacherName: "赵老师",
				GroupName:   "历史组",
			}

			err = model.CreateExtendedLog(ctx, testLog.Id, testInfo)
			require.NoError(t, err)
		}

		// 按学科组ID查询
		logs, err := model.GetExtendedLogsBySubjectId(ctx, subjectId, 0, 10)
		require.NoError(t, err)
		assert.Len(t, logs, 2)

		// 验证所有记录都属于同一个学科组
		for _, log := range logs {
			assert.Equal(t, subjectId, log.SubjectId)
		}
	})

	t.Run("删除扩展日志", func(t *testing.T) {
		// 清理可能存在的测试数据
		model.DB.Exec("DELETE FROM extended_logs")
		model.DB.Exec("DELETE FROM logs")

		// 创建测试日志记录
		testLog := &model.Log{
			UserId:            1,
			CreatedAt:         helper.GetTimestamp(),
			Type:              model.LogTypeConsume,
			Content:           "测试日志",
			Username:          "test_user",
			TokenName:         "test_token",
			ModelName:         "gpt-3.5-turbo",
			Quota:             100,
			PromptTokens:      50,
			CompletionTokens:  50,
			ChannelId:         1,
			RequestId:         "test_request_005",
			ElapsedTime:       1000,
			IsStream:          false,
			SystemPromptReset: false,
		}

		err := model.DB.Create(testLog).Error
		require.NoError(t, err)

		// 创建扩展日志信息
		testInfo := &model.ExtendedLogInfo{
			SchoolId:    5,
			SchoolName:  "测试学校5",
			SubjectId:   105,
			SubjectName: "化学组",
			TeacherId:   "teacher_005",
			TeacherName: "孙老师",
			GroupName:   "化学组",
		}

		// 创建扩展日志
		err = model.CreateExtendedLog(ctx, testLog.Id, testInfo)
		require.NoError(t, err)

		// 验证扩展日志存在
		extendedLog, err := model.GetExtendedLogByLogId(ctx, testLog.Id)
		require.NoError(t, err)
		require.NotNil(t, extendedLog)

		// 删除扩展日志
		err = model.DeleteExtendedLogsByLogIds(ctx, []int{int(testLog.Id)})
		require.NoError(t, err)

		// 验证扩展日志已删除
		extendedLog, err = model.GetExtendedLogByLogId(ctx, testLog.Id)
		assert.NoError(t, err) // 查询不存在的记录不应该返回错误
		assert.Nil(t, extendedLog)
	})

	t.Run("按学校ID查询扩展日志", func(t *testing.T) {
		// 清理可能存在的测试数据
		model.DB.Exec("DELETE FROM extended_logs")
		model.DB.Exec("DELETE FROM logs")

		// 创建测试日志记录
		testLog := &model.Log{
			UserId:            1,
			CreatedAt:         helper.GetTimestamp(),
			Type:              model.LogTypeConsume,
			Content:           "测试日志",
			Username:          "test_user",
			TokenName:         "test_token",
			ModelName:         "gpt-3.5-turbo",
			Quota:             100,
			PromptTokens:      50,
			CompletionTokens:  50,
			ChannelId:         1,
			RequestId:         "test_request_006",
			ElapsedTime:       1000,
			IsStream:          false,
			SystemPromptReset: false,
		}

		err := model.DB.Create(testLog).Error
		require.NoError(t, err)

		// 创建扩展日志信息
		testInfo := &model.ExtendedLogInfo{
			SchoolId:    6,
			SchoolName:  "测试学校6",
			SubjectId:   106,
			SubjectName: "物理组",
			TeacherId:   "teacher_006",
			TeacherName: "赵老师",
			GroupName:   "物理组",
		}

		// 创建扩展日志
		err = model.CreateExtendedLog(ctx, testLog.Id, testInfo)
		require.NoError(t, err)

		// 按学校ID查询
		logs, err := model.GetExtendedLogsBySchoolId(ctx, testInfo.SchoolId, 0, 10)
		require.NoError(t, err)
		assert.Len(t, logs, 1)
		assert.Equal(t, testInfo.SchoolId, logs[0].SchoolId)
	})
}
