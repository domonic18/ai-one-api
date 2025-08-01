package handlers

import (
	"net/http"

	"mock-courseware-platform/models"

	"github.com/gin-gonic/gin"
)

// ServeWebInterface 提供Web管理界面
func ServeWebInterface(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "模拟课件平台API服务器",
	})
}

// GetUsers 获取所有用户信息（Web界面用）
func GetUsers(c *gin.Context) {
	users, err := memoryStorage.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

// GetUser 获取单个用户信息（Web界面用）
func GetUser(c *gin.Context) {
	teacherId := c.Param("teacher_id")
	user, err := memoryStorage.GetUser(teacherId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

// AddUser 添加用户（Web界面用）
func AddUser(c *gin.Context) {
	var user models.WebUser
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	teacherInfo := &models.TeacherInfo{
		TeacherId:      user.TeacherId,
		TeacherName:    user.TeacherName,
		SchoolId:       user.SchoolId,
		SchoolName:     user.SchoolName,
		SubjectId:      user.SubjectId,
		SubjectName:    user.SubjectName,
		OneapiGroup:    user.OneapiGroup,
		PreferredModel: user.PreferredModel,
	}

	if err := memoryStorage.AddUser(teacherInfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 保存到文件
	users, _ := memoryStorage.GetAllUsers()
	fileStorage.SaveUsers(users)

	c.JSON(http.StatusOK, gin.H{"message": "user added successfully"})
}

// UpdateUser 更新用户信息（Web界面用）
func UpdateUser(c *gin.Context) {
	var user models.WebUser
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	teacherInfo := &models.TeacherInfo{
		TeacherId:      user.TeacherId,
		TeacherName:    user.TeacherName,
		SchoolId:       user.SchoolId,
		SchoolName:     user.SchoolName,
		SubjectId:      user.SubjectId,
		SubjectName:    user.SubjectName,
		OneapiGroup:    user.OneapiGroup,
		PreferredModel: user.PreferredModel,
	}

	if err := memoryStorage.UpdateUser(teacherInfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 保存到文件
	users, _ := memoryStorage.GetAllUsers()
	fileStorage.SaveUsers(users)

	c.JSON(http.StatusOK, gin.H{"message": "user updated successfully"})
}

// DeleteUser 删除用户（Web界面用）
func DeleteUser(c *gin.Context) {
	teacherId := c.Param("teacher_id")
	if err := memoryStorage.DeleteUser(teacherId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 保存到文件
	users, _ := memoryStorage.GetAllUsers()
	fileStorage.SaveUsers(users)

	c.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}

// GetConfig 获取配置（Web界面用）
func GetConfig(c *gin.Context) {
	config := memoryStorage.GetConfig()
	c.JSON(http.StatusOK, config)
}

// UpdateConfig 更新配置（Web界面用）
func UpdateConfig(c *gin.Context) {
	var webConfig models.WebConfig
	if err := c.ShouldBindJSON(&webConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config := memoryStorage.GetConfig()
	if config == nil {
		config = &models.ServerConfig{}
	}

	config.Server.Port = webConfig.Port
	config.Server.Host = webConfig.Host
	config.API.ResponseDelay = webConfig.ResponseDelay
	config.API.ErrorRate = webConfig.ErrorRate
	config.API.APIKey = webConfig.APIKey

	memoryStorage.SetConfig(config)

	// 保存到文件
	fileStorage.SaveConfig(config)

	c.JSON(http.StatusOK, gin.H{"message": "config updated successfully"})
}

// ResetToDefault 重置为默认数据
func ResetToDefault(c *gin.Context) {
	// 清空内存数据
	memoryStorage.Clear()

	// 重新初始化默认数据
	if err := fileStorage.InitializeDefaultData(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 重新加载数据到内存
	users, err := fileStorage.LoadUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	memoryStorage.LoadDefaultUsers(users)

	serverConfig, err := fileStorage.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	memoryStorage.SetConfig(serverConfig)

	c.JSON(http.StatusOK, gin.H{"message": "reset to default successfully"})
}

// TestAPI 测试API接口
func TestAPI(c *gin.Context) {
	apiType := c.Param("type")
	teacherId := c.Query("teacher_id")

	var result interface{}
	var err error

	switch apiType {
	case "teacher_info":
		if teacherId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "teacher_id is required"})
			return
		}
		result, err = memoryStorage.GetUser(teacherId)
	case "teacher_ids":
		result, err = memoryStorage.GetAllUserIds()
	case "batch":
		if teacherId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "teacher_id is required"})
			return
		}
		result, err = memoryStorage.BatchGetUsers([]string{teacherId})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid api type"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"api_type": apiType,
		"result":   result,
	})
}
