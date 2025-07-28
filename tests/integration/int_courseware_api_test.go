package integration

import (
	"context"
	"testing"
	"time"

	"github.com/songquanpeng/one-api/common/client"
	"github.com/songquanpeng/one-api/model/smart"
	"github.com/songquanpeng/one-api/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCoursewarePlatformAPI 测试课件平台API调用
func TestCoursewarePlatformAPI(t *testing.T) {
	// 设置测试环境
	_, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 启动模拟课件平台服务
	mockServer := mocks.NewCoursePlatformMock()
	defer mockServer.Close()

	// 配置客户端使用模拟服务器
	coursewareClient := client.GetCoursewareClient()
	coursewareClient.SetBaseURL(mockServer.URL())
	coursewareClient.SetAPIKey("test-api-key")

	// 测试上下文
	ctx := context.Background()

	// 清理所有缓存，确保测试环境干净
	cleanupCache := func() {
		smart.InvalidateTeacherInfoCache(ctx, "teacher1")
		smart.InvalidateTeacherInfoCache(ctx, "teacher2")
		smart.InvalidateSubjectInfoCache(ctx, 1)
		smart.InvalidateSubjectInfoCache(ctx, 2)
		smart.InvalidateUserConfigCache(ctx, "teacher1")
		smart.InvalidateUserConfigCache(ctx, "teacher2")
		smart.InvalidateUserConfigCache(ctx, "unknown_teacher")
	}

	// 测试用例
	t.Run("获取老师信息测试", func(t *testing.T) {
		// 测试目的：验证GetTeacherInfoWithCache能够正确调用API并缓存结果
		// 测试内容：调用GetTeacherInfoWithCache，验证返回的老师信息是否正确

		// 清理缓存
		cleanupCache()

		// 从API获取老师信息
		teacherInfo, err := smart.GetTeacherInfoWithCache(ctx, "teacher1")
		require.NoError(t, err)
		require.NotNil(t, teacherInfo)

		// 验证老师信息
		assert.Equal(t, "teacher1", teacherInfo.TeacherId)
		assert.Equal(t, "张老师", teacherInfo.TeacherName)
		assert.Equal(t, 1, teacherInfo.SchoolId)
		assert.Equal(t, "示例学校", teacherInfo.SchoolName)
		assert.Equal(t, 1, teacherInfo.SubjectId)
		assert.Equal(t, "语文组", teacherInfo.SubjectName)

		// 再次获取，应该从缓存中获取
		teacherInfo2, err := smart.GetTeacherInfoWithCache(ctx, "teacher1")
		require.NoError(t, err)
		require.NotNil(t, teacherInfo2)

		// 验证老师信息一致
		assert.Equal(t, teacherInfo.TeacherId, teacherInfo2.TeacherId)
		assert.Equal(t, teacherInfo.SchoolId, teacherInfo2.SchoolId)
		assert.Equal(t, teacherInfo.SubjectId, teacherInfo2.SubjectId)
	})

	t.Run("获取学科组信息测试", func(t *testing.T) {
		// 测试目的：验证GetSubjectInfoWithCache能够正确调用API并缓存结果
		// 测试内容：调用GetSubjectInfoWithCache，验证返回的学科组信息是否正确

		// 清理缓存
		cleanupCache()

		// 从API获取学科组信息
		subjectInfo, err := smart.GetSubjectInfoWithCache(ctx, 1)
		require.NoError(t, err)
		require.NotNil(t, subjectInfo)

		// 验证学科组信息
		assert.Equal(t, 1, subjectInfo.SubjectId)
		assert.Equal(t, "语文组", subjectInfo.SubjectName)
		assert.Equal(t, 1, subjectInfo.SchoolId)
		assert.Equal(t, "示例学校", subjectInfo.SchoolName)
		assert.Equal(t, "gpt-3.5-turbo", subjectInfo.DefaultModel)
		assert.Equal(t, "chinese", subjectInfo.GroupName)

		// 再次获取，应该从缓存中获取
		subjectInfo2, err := smart.GetSubjectInfoWithCache(ctx, 1)
		require.NoError(t, err)
		require.NotNil(t, subjectInfo2)

		// 验证学科组信息一致
		assert.Equal(t, subjectInfo.SubjectId, subjectInfo2.SubjectId)
		assert.Equal(t, subjectInfo.DefaultModel, subjectInfo2.DefaultModel)
		assert.Equal(t, subjectInfo.GroupName, subjectInfo2.GroupName)
	})

	t.Run("获取用户模型配置测试", func(t *testing.T) {
		// 测试目的：验证GetUserConfigWithCache能够正确调用API并缓存结果
		// 测试内容：调用GetUserConfigWithCache，验证返回的用户模型配置是否正确

		// 清理缓存
		cleanupCache()

		// 从API获取用户模型配置
		userConfig, err := smart.GetUserConfigWithCache(ctx, "teacher1")
		require.NoError(t, err)
		require.NotNil(t, userConfig)

		// 验证用户模型配置
		assert.Equal(t, "teacher1", userConfig.UserId)
		assert.Equal(t, "gpt-4", userConfig.ModelName)

		// 再次获取，应该从缓存中获取
		userConfig2, err := smart.GetUserConfigWithCache(ctx, "teacher1")
		require.NoError(t, err)
		require.NotNil(t, userConfig2)

		// 验证用户模型配置一致
		assert.Equal(t, userConfig.UserId, userConfig2.UserId)
		assert.Equal(t, userConfig.ModelName, userConfig2.ModelName)
	})

	t.Run("智能模型选择测试", func(t *testing.T) {
		// 测试目的：验证GetModelByTeacherId能够正确选择模型
		// 测试内容：调用GetModelByTeacherId，验证返回的模型是否符合预期

		// 清理缓存
		cleanupCache()

		// 测试用户偏好模型（teacher1有用户配置gpt-4）
		model, err := smart.GetModelByTeacherId(ctx, "teacher1")
		require.NoError(t, err)
		assert.Equal(t, "gpt-4", model)

		// 测试学科组默认模型（teacher2使用数学组默认模型gpt-4）
		model, err = smart.GetModelByTeacherId(ctx, "teacher2")
		require.NoError(t, err)
		// teacher2没有用户配置，应该使用学科组默认模型gpt-4
		assert.Equal(t, "gpt-4", model)

		// 测试未知用户（应该返回错误）
		_, err = smart.GetModelByTeacherId(ctx, "unknown_teacher")
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "未找到老师信息")
		}
	})

	t.Run("缓存失效测试", func(t *testing.T) {
		// 测试目的：验证缓存失效机制能够正确工作
		// 测试内容：调用InvalidateXXXCache，验证缓存是否被正确失效

		// 清理缓存
		cleanupCache()

		// 先获取一次，确保缓存已经存在
		teacherInfo, err := smart.GetTeacherInfoWithCache(ctx, "teacher1")
		require.NoError(t, err)
		require.NotNil(t, teacherInfo)

		// 使缓存失效
		err = smart.InvalidateTeacherInfoCache(ctx, "teacher1")
		require.NoError(t, err)

		// 添加新的老师信息到模拟服务器
		mockServer.AddTeacher(&mocks.TeacherInfo{
			TeacherId:   "teacher1",
			TeacherName: "张老师（已更新）",
			SchoolId:    1,
			SchoolName:  "示例学校",
			SubjectId:   1,
			SubjectName: "语文组",
			CreatedAt:   time.Now().Unix(),
			UpdatedAt:   time.Now().Unix() + 100,
		})

		// 再次获取，应该从API获取新的数据
		teacherInfo2, err := smart.GetTeacherInfoWithCache(ctx, "teacher1")
		require.NoError(t, err)
		require.NotNil(t, teacherInfo2)

		// 验证老师信息已更新
		assert.Equal(t, "张老师（已更新）", teacherInfo2.TeacherName)
	})

	t.Run("预加载测试", func(t *testing.T) {
		// 测试目的：验证预加载功能能够正确工作
		// 测试内容：调用PreloadTeacherInfos，验证缓存是否被正确预加载

		// 清理缓存
		cleanupCache()

		// 预加载老师信息
		err := smart.PreloadTeacherInfos(ctx, []string{"teacher1", "teacher2"})
		require.NoError(t, err)

		// 获取老师信息，应该从缓存中获取
		teacherInfo1, err := smart.GetTeacherInfoWithCache(ctx, "teacher1")
		require.NoError(t, err)
		require.NotNil(t, teacherInfo1)

		teacherInfo2, err := smart.GetTeacherInfoWithCache(ctx, "teacher2")
		require.NoError(t, err)
		require.NotNil(t, teacherInfo2)

		// 验证老师信息
		assert.Equal(t, "张老师（已更新）", teacherInfo1.TeacherName)
		assert.Equal(t, "李老师", teacherInfo2.TeacherName)
	})
}
