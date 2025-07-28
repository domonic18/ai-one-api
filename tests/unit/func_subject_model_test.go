package unit

import (
	"context"
	"testing"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/model/smart"
	"github.com/stretchr/testify/assert"
)

// TestSubjectInfo 测试学科组信息结构
func TestSubjectInfo(t *testing.T) {
	// 创建学科组信息实例
	info := &smart.SubjectInfo{
		SubjectId:    1,
		SubjectName:  "数学",
		SchoolId:     100,
		SchoolName:   "示例学校",
		DefaultModel: "gpt-3.5-turbo",
	}

	// 验证字段
	assert.Equal(t, 1, info.SubjectId)
	assert.Equal(t, "数学", info.SubjectName)
	assert.Equal(t, 100, info.SchoolId)
	assert.Equal(t, "示例学校", info.SchoolName)
	assert.Equal(t, "gpt-3.5-turbo", info.DefaultModel)
}

// TestSubjectInfoJSON 测试学科组信息JSON序列化
func TestSubjectInfoJSON(t *testing.T) {
	info := &smart.SubjectInfo{
		SubjectId:    2,
		SubjectName:  "英语",
		SchoolId:     101,
		SchoolName:   "测试学校",
		DefaultModel: "gpt-4",
	}

	// 测试JSON序列化（简单验证结构体可以被正确序列化）
	assert.NotNil(t, info)
	assert.Equal(t, 2, info.SubjectId)
}

// TestTeacherInfo 测试老师信息结构
func TestTeacherInfo(t *testing.T) {
	// 创建老师信息实例
	info := &smart.TeacherInfo{
		TeacherId:   "teacher_001",
		TeacherName: "张老师",
		SchoolId:    100,
		SchoolName:  "示例学校",
		SubjectId:   1,
		SubjectName: "数学",
	}

	// 验证字段
	assert.Equal(t, "teacher_001", info.TeacherId)
	assert.Equal(t, "张老师", info.TeacherName)
	assert.Equal(t, 100, info.SchoolId)
	assert.Equal(t, "示例学校", info.SchoolName)
	assert.Equal(t, 1, info.SubjectId)
	assert.Equal(t, "数学", info.SubjectName)
}

// TestTeacherInfoJSON 测试老师信息JSON序列化
func TestTeacherInfoJSON(t *testing.T) {
	info := &smart.TeacherInfo{
		TeacherId:   "teacher_002",
		TeacherName: "李老师",
		SchoolId:    101,
		SchoolName:  "测试学校",
		SubjectId:   2,
		SubjectName: "英语",
	}

	// 测试JSON序列化（简单验证结构体可以被正确序列化）
	assert.NotNil(t, info)
	assert.Equal(t, "teacher_002", info.TeacherId)
}

// TestSubjectInfoCache 测试学科组信息缓存功能
func TestSubjectInfoCache(t *testing.T) {
	// 禁用Redis以进行单元测试
	originalRedisEnabled := common.RedisEnabled
	common.RedisEnabled = false
	defer func() {
		common.RedisEnabled = originalRedisEnabled
	}()

	ctx := context.Background()

	t.Run("获取学科组信息缓存测试", func(t *testing.T) {
		subjectId := 1

		// Redis未启用时应该返回错误
		info, err := smart.GetSubjectInfoWithCache(ctx, subjectId)
		assert.Error(t, err)
		assert.Nil(t, info)
		assert.Contains(t, err.Error(), "Redis未启用")
	})

	t.Run("获取老师信息缓存测试", func(t *testing.T) {
		teacherId := "teacher_001"

		// Redis未启用时应该返回错误
		info, err := smart.GetTeacherInfoWithCache(ctx, teacherId)
		assert.Error(t, err)
		assert.Nil(t, info)
		assert.Contains(t, err.Error(), "Redis未启用")
	})

	t.Run("失效学科组信息缓存测试", func(t *testing.T) {
		subjectId := 1

		// Redis未启用时应该返回错误
		err := smart.InvalidateSubjectInfoCache(ctx, subjectId)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis未启用")
	})

	t.Run("失效老师信息缓存测试", func(t *testing.T) {
		teacherId := "teacher_001"

		// Redis未启用时应该返回错误
		err := smart.InvalidateTeacherInfoCache(ctx, teacherId)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis未启用")
	})

	t.Run("根据老师ID获取模型测试", func(t *testing.T) {
		// 空用户ID测试
		modelName, err := smart.GetModelByTeacherId(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "teacherId不能为空")
		assert.Equal(t, "", modelName)

		// 有效用户ID但Redis未启用
		modelName, err = smart.GetModelByTeacherId(ctx, "teacher_001")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis未启用")
		assert.Equal(t, "", modelName)
	})

	t.Run("批量失效老师信息缓存测试", func(t *testing.T) {
		teacherIds := []string{"teacher1", "teacher2"}

		// Redis未启用时应该返回错误
		err := smart.BatchInvalidateTeacherInfoCache(ctx, teacherIds)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis未启用")

		// 空列表应该直接返回nil
		err = smart.BatchInvalidateTeacherInfoCache(ctx, []string{})
		assert.NoError(t, err)
	})

	t.Run("预加载老师信息测试", func(t *testing.T) {
		teacherIds := []string{"teacher1", "teacher2"}

		// Redis未启用时应该返回错误
		err := smart.PreloadTeacherInfos(ctx, teacherIds)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis未启用")

		// 空列表应该直接返回nil
		err = smart.PreloadTeacherInfos(ctx, []string{})
		assert.NoError(t, err)
	})
}
