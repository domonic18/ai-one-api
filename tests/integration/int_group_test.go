package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/random"
	"github.com/songquanpeng/one-api/controller"
	"github.com/songquanpeng/one-api/middleware"
	"github.com/songquanpeng/one-api/model"
	billingratio "github.com/songquanpeng/one-api/relay/billing/ratio"

	"github.com/stretchr/testify/assert"
)

// setupIntegrationTestDB 设置集成测试数据库
func setupIntegrationTestDB() {
	// 确保环境变量已设置
	if os.Getenv("SQL_DSN") == "" {
		os.Setenv("SQL_DSN", "testuser:testpass@tcp(localhost:3306)/oneapi_test?charset=utf8mb4&parseTime=True&loc=Local")
	}

	// 初始化数据库
	model.InitDB()

	// 自动迁移表结构
	model.DB.AutoMigrate(&model.User{})
}

// setupGroupTestRouter 设置用户组测试路由
func setupGroupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// 添加用户组管理路由
	groupRoute := r.Group("/api/group")
	groupRoute.Use(middleware.AdminAuth()) // 需要管理员权限
	{
		groupRoute.GET("/", controller.GetGroups)
		groupRoute.GET("/detail", controller.GetGroupsDetail)
		groupRoute.POST("/", controller.CreateGroup)
		groupRoute.PUT("/:id", controller.UpdateGroup)
		groupRoute.DELETE("/:id", controller.DeleteGroup)
	}

	return r
}

// createGroupTestAdmin 创建测试管理员用户
func createGroupTestAdmin(t *testing.T) (*model.User, string) {
	admin := &model.User{
		Username:    "test_admin_" + random.GetRandomString(6),
		Password:    "test_password",
		DisplayName: "测试管理员",
		Role:        model.RoleAdminUser,
		Status:      model.UserStatusEnabled,
		AccessToken: random.GetUUID(),
		AffCode:     random.GetRandomString(4),
	}
	err := model.DB.Create(admin).Error
	assert.NoError(t, err)
	return admin, admin.AccessToken
}

// TestGroupAPI_完整流程 测试用户组管理的完整API流程
func TestGroupAPI_完整流程(t *testing.T) {
	// 设置集成测试数据库
	setupIntegrationTestDB()

	r := setupGroupTestRouter()

	// 创建测试管理员
	admin, token := createGroupTestAdmin(t)
	defer func() {
		model.DB.Unscoped().Delete(admin)
	}()

	groupName := "integration_test_group_" + random.GetRandomString(6)
	description := "集成测试用户组"
	newDescription := "更新后的集成测试用户组"

	t.Run("创建用户组", func(t *testing.T) {
		createData := map[string]interface{}{
			"name":        groupName,
			"description": description,
		}
		jsonData, err := json.Marshal(createData)
		assert.NoError(t, err)

		req, err := http.NewRequest("POST", "/api/group", bytes.NewBuffer(jsonData))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
		assert.Equal(t, "用户组创建成功", response["message"])

		// 验证数据库中创建了用户组
		var groupUser model.User
		err = model.DB.Where("username = ?", "group_"+groupName).First(&groupUser).Error
		assert.NoError(t, err)
		assert.Equal(t, description, groupUser.DisplayName)
	})

	t.Run("获取用户组列表", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/group", nil)
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))

		// 验证返回的用户组列表包含创建的用户组
		data := response["data"].([]interface{})
		found := false
		for _, item := range data {
			if item.(string) == groupName {
				found = true
				break
			}
		}
		assert.True(t, found, "用户组列表应该包含创建的用户组")
	})

	t.Run("获取用户组详情", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/group/detail", nil)
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))

		// 验证返回的用户组详情包含创建的用户组
		data := response["data"].([]interface{})
		found := false
		for _, item := range data {
			group := item.(map[string]interface{})
			if group["name"].(string) == groupName {
				assert.Equal(t, description, group["description"].(string))
				assert.Equal(t, float64(0), group["user_count"].(float64))
				found = true
				break
			}
		}
		assert.True(t, found, "用户组详情列表应该包含创建的用户组")
	})

	t.Run("更新用户组", func(t *testing.T) {
		updateData := map[string]interface{}{
			"name":        groupName,
			"description": newDescription,
		}
		jsonData, err := json.Marshal(updateData)
		assert.NoError(t, err)

		req, err := http.NewRequest("PUT", "/api/group/"+groupName, bytes.NewBuffer(jsonData))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
		assert.Equal(t, "用户组更新成功", response["message"])

		// 验证数据库中的描述已更新
		var updatedUser model.User
		err = model.DB.Where("username = ?", "group_"+groupName).First(&updatedUser).Error
		assert.NoError(t, err)
		assert.Equal(t, newDescription, updatedUser.DisplayName)
	})

	t.Run("删除用户组", func(t *testing.T) {
		req, err := http.NewRequest("DELETE", "/api/group/"+groupName, nil)
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
		assert.Equal(t, "用户组删除成功", response["message"])

		// 验证数据库中用户组已删除
		var deletedUser model.User
		err = model.DB.Where("username = ?", "group_"+groupName).First(&deletedUser).Error
		assert.Error(t, err) // 应该找不到记录

		// 验证GroupRatio中用户组已移除
		_, exists := billingratio.GroupRatio[groupName]
		assert.False(t, exists)
	})
}

// TestGroupAPI_权限控制 测试用户组API的权限控制
func TestGroupAPI_权限控制(t *testing.T) {
	// 设置集成测试数据库
	setupIntegrationTestDB()

	r := setupGroupTestRouter()

	// 创建普通用户（非管理员）
	normalUser := &model.User{
		Username:    "normal_user_" + random.GetRandomString(6),
		Password:    "test_password",
		DisplayName: "普通用户",
		Role:        model.RoleCommonUser, // 普通用户
		Status:      model.UserStatusEnabled,
		AccessToken: random.GetUUID(),
		AffCode:     random.GetRandomString(4),
	}
	err := model.DB.Create(normalUser).Error
	assert.NoError(t, err)
	defer func() {
		model.DB.Unscoped().Delete(normalUser)
	}()

	groupName := "test_group_" + random.GetRandomString(6)

	testCases := []struct {
		name   string
		method string
		path   string
		body   map[string]interface{}
	}{
		{"获取用户组列表", "GET", "/api/group", nil},
		{"获取用户组详情", "GET", "/api/group/detail", nil},
		{"创建用户组", "POST", "/api/group", map[string]interface{}{
			"name":        groupName,
			"description": "测试用户组",
		}},
		{"更新用户组", "PUT", "/api/group/" + groupName, map[string]interface{}{
			"name":        groupName,
			"description": "更新后的用户组",
		}},
		{"删除用户组", "DELETE", "/api/group/" + groupName, nil},
	}

	for _, tc := range testCases {
		t.Run("普通用户"+tc.name+"应该被拒绝", func(t *testing.T) {
			var req *http.Request
			var err error

			if tc.body != nil {
				jsonData, _ := json.Marshal(tc.body)
				req, err = http.NewRequest(tc.method, tc.path, bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, err = http.NewRequest(tc.method, tc.path, nil)
			}
			assert.NoError(t, err)
			req.Header.Set("Authorization", "Bearer "+normalUser.AccessToken)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			// 普通用户应该被拒绝访问
			assert.Equal(t, http.StatusForbidden, w.Code)
		})
	}
}

// TestGroupAPI_错误场景 测试用户组API的错误场景
func TestGroupAPI_错误场景(t *testing.T) {
	// 设置集成测试数据库
	setupIntegrationTestDB()

	r := setupGroupTestRouter()

	// 创建测试管理员
	admin, token := createGroupTestAdmin(t)
	defer func() {
		model.DB.Unscoped().Delete(admin)
	}()

	t.Run("创建用户组_无效参数", func(t *testing.T) {
		createData := map[string]interface{}{
			// 缺少必需的name字段
			"description": "无效的用户组",
		}
		jsonData, err := json.Marshal(createData)
		assert.NoError(t, err)

		req, err := http.NewRequest("POST", "/api/group", bytes.NewBuffer(jsonData))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.False(t, response["success"].(bool))
		assert.Contains(t, response["message"].(string), "请求参数错误")
	})

	t.Run("更新不存在的用户组", func(t *testing.T) {
		nonExistentGroup := "non_existent_group_" + random.GetRandomString(6)
		updateData := map[string]interface{}{
			"name":        nonExistentGroup,
			"description": "不存在的用户组",
		}
		jsonData, err := json.Marshal(updateData)
		assert.NoError(t, err)

		req, err := http.NewRequest("PUT", "/api/group/"+nonExistentGroup, bytes.NewBuffer(jsonData))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.False(t, response["success"].(bool))
		assert.Equal(t, "用户组不存在", response["message"].(string))
	})

	t.Run("删除不存在的用户组", func(t *testing.T) {
		nonExistentGroup := "non_existent_group_" + random.GetRandomString(6)

		req, err := http.NewRequest("DELETE", "/api/group/"+nonExistentGroup, nil)
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.False(t, response["success"].(bool))
		assert.Equal(t, "用户组不存在", response["message"].(string))
	})

	t.Run("删除有活跃用户的用户组", func(t *testing.T) {
		// 创建用户组
		groupName := "test_group_with_users_" + random.GetRandomString(6)
		groupUser := &model.User{
			Username:    "group_" + groupName,
			Password:    "test_password",
			DisplayName: "有用户的测试用户组",
			Status:      model.UserStatusEnabled,
			Group:       groupName,
			AccessToken: random.GetUUID(),
			AffCode:     random.GetRandomString(4),
		}
		err := model.DB.Create(groupUser).Error
		assert.NoError(t, err)

		// 创建属于该用户组的普通用户
		normalUser := &model.User{
			Username:    "user_in_group_" + random.GetRandomString(6),
			Password:    "test_password",
			Status:      model.UserStatusEnabled,
			Group:       groupName,
			AccessToken: random.GetUUID(),
			AffCode:     random.GetRandomString(4),
		}
		err = model.DB.Create(normalUser).Error
		assert.NoError(t, err)

		defer func() {
			model.DB.Unscoped().Delete(groupUser)
			model.DB.Unscoped().Delete(normalUser)
		}()

		// 尝试删除用户组
		req, err := http.NewRequest("DELETE", "/api/group/"+groupName, nil)
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.False(t, response["success"].(bool))
		assert.Contains(t, response["message"].(string), "还有 1 个活跃用户")
	})
}

// TestGroupAPI_并发操作 测试用户组API的并发操作
func TestGroupAPI_并发操作(t *testing.T) {
	// 设置集成测试数据库
	setupIntegrationTestDB()

	r := setupGroupTestRouter()

	// 创建测试管理员
	admin, token := createGroupTestAdmin(t)
	defer func() {
		model.DB.Unscoped().Delete(admin)
	}()

	const concurrency = 5
	groupNames := make([]string, concurrency)

	// 并发创建多个用户组
	t.Run("并发创建用户组", func(t *testing.T) {
		done := make(chan bool, concurrency)

		for i := 0; i < concurrency; i++ {
			go func(index int) {
				defer func() { done <- true }()

				groupName := fmt.Sprintf("concurrent_group_%d_%s", index, random.GetRandomString(4))
				groupNames[index] = groupName

				createData := map[string]interface{}{
					"name":        groupName,
					"description": fmt.Sprintf("并发测试用户组 %d", index),
				}
				jsonData, err := json.Marshal(createData)
				assert.NoError(t, err)

				req, err := http.NewRequest("POST", "/api/group", bytes.NewBuffer(jsonData))
				assert.NoError(t, err)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+token)

				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)

				assert.Equal(t, http.StatusOK, w.Code)
			}(i)
		}

		// 等待所有goroutine完成
		for i := 0; i < concurrency; i++ {
			<-done
		}

		// 验证所有用户组都创建成功
		for _, groupName := range groupNames {
			var groupUser model.User
			err := model.DB.Where("username = ?", "group_"+groupName).First(&groupUser).Error
			assert.NoError(t, err)
		}
	})

	// 清理并发创建的用户组
	defer func() {
		for _, groupName := range groupNames {
			if groupName != "" {
				model.DB.Exec("DELETE FROM users WHERE username = ?", "group_"+groupName)
				delete(billingratio.GroupRatio, groupName)
			}
		}
	}()
}
