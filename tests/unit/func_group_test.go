package unit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/random"
	"github.com/songquanpeng/one-api/controller"
	"github.com/songquanpeng/one-api/model"
	billingratio "github.com/songquanpeng/one-api/relay/billing/ratio"
	"github.com/stretchr/testify/assert"
)

// setupTestDB 设置测试数据库
func setupTestDB() {
	// 确保环境变量已设置
	if os.Getenv("SQL_DSN") == "" {
		os.Setenv("SQL_DSN", "testuser:testpass@tcp(localhost:3306)/oneapi_test?charset=utf8mb4&parseTime=True&loc=Local")
	}

	// 初始化数据库
	model.InitDB()

	// 自动迁移表结构
	model.DB.AutoMigrate(&model.User{})
}

// TestGroup_Create_创建用户组 测试用户组创建功能
func TestGroup_Create_创建用户组(t *testing.T) {
	// 设置测试数据库
	setupTestDB()

	// 设置测试环境
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/api/group", controller.CreateGroup)

	// 准备测试数据
	groupName := "test_group_" + random.GetRandomString(6)
	description := "测试用户组描述"

	// 构造创建请求
	createData := map[string]interface{}{
		"name":        groupName,
		"description": description,
	}
	jsonData, err := json.Marshal(createData)
	assert.NoError(t, err)

	// 发送请求
	req, err := http.NewRequest("POST", "/api/group", bytes.NewBuffer(jsonData))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.Equal(t, "用户组创建成功", response["message"])

	// 验证数据库中创建了用户组代表用户
	var groupUser model.User
	err = model.DB.Where("username = ?", "group_"+groupName).First(&groupUser).Error
	assert.NoError(t, err)
	assert.Equal(t, description, groupUser.DisplayName)
	assert.Equal(t, groupName, groupUser.Group)

	// 验证GroupRatio中添加了用户组
	_, exists := billingratio.GroupRatio[groupName]
	assert.True(t, exists)

	// 清理测试数据
	defer func() {
		model.DB.Unscoped().Delete(&groupUser)
		delete(billingratio.GroupRatio, groupName)
	}()
}

// TestGroup_Update_更新用户组描述 测试用户组描述更新功能
func TestGroup_Update_更新用户组描述(t *testing.T) {
	// 设置测试数据库
	setupTestDB()

	// 设置测试环境
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.PUT("/api/group/:id", controller.UpdateGroup)

	// 准备测试数据
	groupName := "test_group_" + random.GetRandomString(6)
	description := "测试用户组"
	newDescription := "北京光华学校AI组"

	// 创建测试用户组
	groupUser := &model.User{
		Username:    "group_" + groupName,
		Password:    "test_password",
		DisplayName: description,
		Status:      model.UserStatusEnabled,
		Group:       groupName,
		AccessToken: random.GetUUID(),
		AffCode:     random.GetRandomString(4),
	}
	err := model.DB.Create(groupUser).Error
	assert.NoError(t, err)

	// 清理测试数据
	defer func() {
		model.DB.Unscoped().Delete(groupUser)
	}()

	// 构造更新请求
	updateData := map[string]interface{}{
		"name":        groupName,
		"description": newDescription,
	}
	jsonData, err := json.Marshal(updateData)
	assert.NoError(t, err)

	// 发送请求
	req, err := http.NewRequest("PUT", "/api/group/"+groupName, bytes.NewBuffer(jsonData))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))

	// 验证数据库中的描述已更新
	var updatedUser model.User
	err = model.DB.Where("username = ?", "group_"+groupName).First(&updatedUser).Error
	assert.NoError(t, err)
	assert.Equal(t, newDescription, updatedUser.DisplayName)
}

// TestGroup_Delete_删除空用户组 测试删除空用户组功能
func TestGroup_Delete_删除空用户组(t *testing.T) {
	// 设置测试数据库
	setupTestDB()

	// 设置测试环境
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.DELETE("/api/group/:id", controller.DeleteGroup)

	// 准备测试数据
	groupName := "test_group_" + random.GetRandomString(6)
	description := "待删除的测试用户组"

	// 创建测试用户组
	groupUser := &model.User{
		Username:    "group_" + groupName,
		Password:    "test_password",
		DisplayName: description,
		Status:      model.UserStatusEnabled,
		Group:       groupName,
		AccessToken: random.GetUUID(),
		AffCode:     random.GetRandomString(4),
	}
	err := model.DB.Create(groupUser).Error
	assert.NoError(t, err)

	// 添加到GroupRatio
	billingratio.GroupRatio[groupName] = 1.0

	// 发送删除请求
	req, err := http.NewRequest("DELETE", "/api/group/"+groupName, nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.Equal(t, "用户组删除成功", response["message"])

	// 验证数据库中用户组代表用户已删除
	var deletedUser model.User
	err = model.DB.Where("username = ?", "group_"+groupName).First(&deletedUser).Error
	assert.Error(t, err) // 应该找不到记录

	// 验证GroupRatio中用户组已移除
	_, exists := billingratio.GroupRatio[groupName]
	assert.False(t, exists)
}

// TestGroup_Delete_有活跃用户的用户组 测试删除有活跃用户的用户组（应该失败）
func TestGroup_Delete_有活跃用户的用户组(t *testing.T) {
	// 设置测试数据库
	setupTestDB()

	// 设置测试环境
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.DELETE("/api/group/:id", controller.DeleteGroup)

	// 准备测试数据
	groupName := "test_group_" + random.GetRandomString(6)
	description := "有活跃用户的测试用户组"

	// 创建测试用户组代表用户
	groupUser := &model.User{
		Username:    "group_" + groupName,
		Password:    "test_password",
		DisplayName: description,
		Status:      model.UserStatusEnabled,
		Group:       groupName,
		AccessToken: random.GetUUID(),
		AffCode:     random.GetRandomString(4),
	}
	err := model.DB.Create(groupUser).Error
	assert.NoError(t, err)

	// 创建一个普通用户属于该用户组
	normalUser := &model.User{
		Username:    "normal_user_" + random.GetRandomString(6),
		Password:    "test_password",
		Status:      model.UserStatusEnabled,
		Group:       groupName,
		AccessToken: random.GetUUID(), // 必须设置AccessToken，因为有唯一索引
		AffCode:     random.GetRandomString(4),
	}
	err = model.DB.Create(normalUser).Error
	assert.NoError(t, err)

	// 清理测试数据
	defer func() {
		model.DB.Unscoped().Delete(groupUser)
		model.DB.Unscoped().Delete(normalUser)
	}()

	// 发送删除请求
	req, err := http.NewRequest("DELETE", "/api/group/"+groupName, nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// 验证响应 - 应该失败
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["success"].(bool))
	assert.Contains(t, response["message"].(string), "还有 1 个活跃用户")

	// 验证数据库中用户组代表用户仍然存在
	var existingUser model.User
	err = model.DB.Where("username = ?", "group_"+groupName).First(&existingUser).Error
	assert.NoError(t, err)
}

// TestGroup_GetGroupsDetail_获取用户组详情 测试获取用户组详情功能
func TestGroup_GetGroupsDetail_获取用户组详情(t *testing.T) {
	// 设置测试数据库
	setupTestDB()

	// 设置测试环境
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/api/group/detail", controller.GetGroupsDetail)

	// 准备测试数据
	groupName := "test_group_" + random.GetRandomString(6)
	description := "测试用户组详情"

	// 创建测试用户组
	groupUser := &model.User{
		Username:    "group_" + groupName,
		Password:    "test_password",
		DisplayName: description,
		Status:      model.UserStatusEnabled,
		Group:       groupName,
		AccessToken: random.GetUUID(),
		AffCode:     random.GetRandomString(4),
	}
	err := model.DB.Create(groupUser).Error
	assert.NoError(t, err)

	// 清理测试数据
	defer func() {
		model.DB.Unscoped().Delete(groupUser)
	}()

	// 发送请求
	req, err := http.NewRequest("GET", "/api/group/detail", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))

	// 验证返回的数据中包含创建的用户组
	data := response["data"].([]interface{})
	found := false
	for _, item := range data {
		group := item.(map[string]interface{})
		if group["name"].(string) == groupName {
			assert.Equal(t, description, group["description"].(string))
			assert.Equal(t, float64(0), group["user_count"].(float64)) // 排除用户组代表用户
			found = true
			break
		}
	}
	assert.True(t, found, "应该在返回的用户组列表中找到创建的用户组")
}

// TestGroup_Create_重复名称 测试创建重复名称的用户组（应该失败）
func TestGroup_Create_重复名称(t *testing.T) {
	// 设置测试数据库
	setupTestDB()

	// 设置测试环境
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/api/group", controller.CreateGroup)

	// 准备测试数据
	groupName := "test_group_" + random.GetRandomString(6)
	description := "重复名称测试用户组"

	// 先创建一个用户组
	groupUser := &model.User{
		Username:    "group_" + groupName,
		Password:    "test_password",
		DisplayName: description,
		Status:      model.UserStatusEnabled,
		Group:       groupName,
		AccessToken: random.GetUUID(),
		AffCode:     random.GetRandomString(4),
	}
	err := model.DB.Create(groupUser).Error
	assert.NoError(t, err)

	// 清理测试数据
	defer func() {
		model.DB.Unscoped().Delete(groupUser)
	}()

	// 构造创建请求（相同名称）
	createData := map[string]interface{}{
		"name":        groupName,
		"description": "另一个描述",
	}
	jsonData, err := json.Marshal(createData)
	assert.NoError(t, err)

	// 发送请求
	req, err := http.NewRequest("POST", "/api/group", bytes.NewBuffer(jsonData))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// 验证响应 - 应该失败
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["success"].(bool))
	assert.Equal(t, "用户组名称已存在", response["message"].(string))
}

// TestGroup_Delete_预定义用户组 测试删除预定义用户组（应该失败）
func TestGroup_Delete_预定义用户组(t *testing.T) {
	// 设置测试数据库
	setupTestDB()

	// 设置测试环境
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.DELETE("/api/group/:id", controller.DeleteGroup)

	predefinedGroups := []string{"default", "vip", "svip"}

	for _, groupName := range predefinedGroups {
		t.Run("删除"+groupName+"用户组", func(t *testing.T) {
			// 发送删除请求
			req, err := http.NewRequest("DELETE", "/api/group/"+groupName, nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			// 验证响应 - 应该失败
			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.False(t, response["success"].(bool))
			assert.Equal(t, "无法删除系统预定义用户组", response["message"].(string))
		})
	}
}
