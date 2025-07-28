package unit

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/model"
	"github.com/stretchr/testify/assert"
)

func TestSubjectInfo(t *testing.T) {
	t.Run("学科组信息结构体测试", func(t *testing.T) {
		info := &model.SubjectInfo{
			SubjectId:    101,
			SubjectName:  "数学组",
			SchoolId:     1,
			SchoolName:   "测试学校",
			DefaultModel: "gpt-4-turbo",
			UpdatedAt:    time.Now().Unix(),
		}

		assert.Equal(t, 101, info.SubjectId)
		assert.Equal(t, "数学组", info.SubjectName)
		assert.Equal(t, 1, info.SchoolId)
		assert.Equal(t, "测试学校", info.SchoolName)
		assert.Equal(t, "gpt-4-turbo", info.DefaultModel)
		assert.True(t, info.UpdatedAt > 0)
	})

	t.Run("学科组信息JSON序列化测试", func(t *testing.T) {
		info := &model.SubjectInfo{
			SubjectId:    102,
			SubjectName:  "英语组",
			SchoolId:     2,
			SchoolName:   "示例学校",
			DefaultModel: "claude-3-sonnet",
			UpdatedAt:    time.Now().Unix(),
		}

		jsonData, err := json.Marshal(info)
		assert.NoError(t, err)
		assert.NotEmpty(t, jsonData)

		var decodedInfo model.SubjectInfo
		err = json.Unmarshal(jsonData, &decodedInfo)
		assert.NoError(t, err)
		assert.Equal(t, info.SubjectId, decodedInfo.SubjectId)
		assert.Equal(t, info.SubjectName, decodedInfo.SubjectName)
		assert.Equal(t, info.SchoolId, decodedInfo.SchoolId)
		assert.Equal(t, info.SchoolName, decodedInfo.SchoolName)
		assert.Equal(t, info.DefaultModel, decodedInfo.DefaultModel)
		assert.Equal(t, info.UpdatedAt, decodedInfo.UpdatedAt)
	})
}

func TestTeacherInfo(t *testing.T) {
	t.Run("老师信息结构体测试", func(t *testing.T) {
		info := &model.TeacherInfo{
			TeacherId:   "teacher_001",
			TeacherName: "张老师",
			SchoolId:    1,
			SchoolName:  "测试学校",
			SubjectId:   101,
			SubjectName: "数学组",
			GroupName:   "数学组",
		}

		assert.Equal(t, "teacher_001", info.TeacherId)
		assert.Equal(t, "张老师", info.TeacherName)
		assert.Equal(t, 1, info.SchoolId)
		assert.Equal(t, "测试学校", info.SchoolName)
		assert.Equal(t, 101, info.SubjectId)
		assert.Equal(t, "数学组", info.SubjectName)
		assert.Equal(t, "数学组", info.GroupName)
	})

	t.Run("老师信息JSON序列化测试", func(t *testing.T) {
		info := &model.TeacherInfo{
			TeacherId:   "teacher_002",
			TeacherName: "李老师",
			SchoolId:    2,
			SchoolName:  "示例学校",
			SubjectId:   102,
			SubjectName: "英语组",
			GroupName:   "英语组",
		}

		jsonData, err := json.Marshal(info)
		assert.NoError(t, err)
		assert.NotEmpty(t, jsonData)

		var decodedInfo model.TeacherInfo
		err = json.Unmarshal(jsonData, &decodedInfo)
		assert.NoError(t, err)
		assert.Equal(t, info.TeacherId, decodedInfo.TeacherId)
		assert.Equal(t, info.TeacherName, decodedInfo.TeacherName)
		assert.Equal(t, info.SchoolId, decodedInfo.SchoolId)
		assert.Equal(t, info.SchoolName, decodedInfo.SchoolName)
		assert.Equal(t, info.SubjectId, decodedInfo.SubjectId)
		assert.Equal(t, info.SubjectName, decodedInfo.SubjectName)
		assert.Equal(t, info.GroupName, decodedInfo.GroupName)
	})
}

func TestGetSubjectInfoWithCache(t *testing.T) {
	ctx := context.Background()

	t.Run("Redis未启用测试", func(t *testing.T) {
		// 保存原始值
		originalRedisEnabled := common.RedisEnabled
		// 测试结束后恢复
		defer func() { common.RedisEnabled = originalRedisEnabled }()

		// 设置Redis为未启用
		common.RedisEnabled = false

		subjectId := 101
		info, err := model.GetSubjectInfoWithCache(ctx, subjectId)

		// 应该返回错误，表示Redis未启用
		assert.Error(t, err)
		assert.Nil(t, info)
		assert.Contains(t, err.Error(), "Redis未启用")
	})
}

func TestInvalidateSubjectInfoCache(t *testing.T) {
	ctx := context.Background()

	t.Run("Redis未启用测试", func(t *testing.T) {
		// 保存原始值
		originalRedisEnabled := common.RedisEnabled
		// 测试结束后恢复
		defer func() { common.RedisEnabled = originalRedisEnabled }()

		// 设置Redis为未启用
		common.RedisEnabled = false

		subjectId := 101
		err := model.InvalidateSubjectInfoCache(ctx, subjectId)

		// 应该返回错误，表示Redis未启用
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis未启用")
	})
}

func TestInvalidateTeacherInfoCache(t *testing.T) {
	ctx := context.Background()

	t.Run("Redis未启用测试", func(t *testing.T) {
		// 保存原始值
		originalRedisEnabled := common.RedisEnabled
		// 测试结束后恢复
		defer func() { common.RedisEnabled = originalRedisEnabled }()

		// 设置Redis为未启用
		common.RedisEnabled = false

		teacherId := "teacher_001"
		err := model.InvalidateTeacherInfoCache(ctx, teacherId)

		// 应该返回错误，表示Redis未启用
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis未启用")
	})
}

func TestGetTeacherInfoWithCache(t *testing.T) {
	ctx := context.Background()

	t.Run("Redis未启用测试", func(t *testing.T) {
		// 保存原始值
		originalRedisEnabled := common.RedisEnabled
		// 测试结束后恢复
		defer func() { common.RedisEnabled = originalRedisEnabled }()

		// 设置Redis为未启用
		common.RedisEnabled = false

		teacherId := "teacher_001"
		info, err := model.GetTeacherInfoWithCache(ctx, teacherId)

		// 应该返回错误，表示Redis未启用
		assert.Error(t, err)
		assert.Nil(t, info)
		assert.Contains(t, err.Error(), "Redis未启用")
	})
}

// 简化版GetModelByTeacherId测试，只测试不需要Redis的场景
func TestGetModelByTeacherId(t *testing.T) {
	ctx := context.Background()

	// 保存原始Redis状态
	originalRedisEnabled := common.RedisEnabled
	defer func() {
		common.RedisEnabled = originalRedisEnabled
	}()

	// 禁用Redis，确保单元测试不依赖外部环境
	common.RedisEnabled = false

	t.Run("空用户ID测试", func(t *testing.T) {
		teacherId := ""
		modelName, err := model.GetModelByTeacherId(ctx, teacherId)

		// 空用户ID仍然会调用Redis相关函数，所以会返回Redis未启用的错误
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis未启用")
		assert.Equal(t, "", modelName)
	})

	t.Run("Redis禁用时的行为测试", func(t *testing.T) {
		teacherId := "teacher_001"
		modelName, err := model.GetModelByTeacherId(ctx, teacherId)

		// Redis禁用时应该返回错误
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis未启用")
		assert.Equal(t, "", modelName)
	})
}
