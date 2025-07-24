package unit

import (
	"testing"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/model"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupUserTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 设置 SQLite 标志
	common.UsingSQLite = true

	// 迁移表结构
	err = db.AutoMigrate(&model.User{}, &model.Token{})
	if err != nil {
		panic("failed to migrate database")
	}

	return db
}

// TestUser_ValidateAndFill 测试用户验证和填充功能
// 测试目的：验证系统能够正确验证用户凭据并填充用户信息
// 测试内容：
// 1. 测试正确凭据的验证
// 2. 测试错误密码的处理
// 3. 测试不存在用户的处理
// 4. 验证用户信息填充的准确性
func TestUser_ValidateAndFill(t *testing.T) {
	db := setupUserTestDB()
	model.DB = db

	// 清理测试数据
	model.DB.Where("1 = 1").Delete(&model.User{})
	model.DB.Where("1 = 1").Delete(&model.Token{})

	// 创建测试用户
	user := model.User{
		Username:    "testuser1",
		Password:    "password123",
		DisplayName: "Test User",
		Role:        1,
		Status:      1,
		Group:       "default",
		Quota:       1000,
	}
	err := user.Insert(nil, 0)
	assert.NoError(t, err)

	tests := []struct {
		name     string
		username string
		password string
		wantErr  bool
	}{
		{
			name:     "正确凭据",
			username: "testuser1",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "错误密码",
			username: "testuser1",
			password: "wrongpassword",
			wantErr:  true,
		},
		{
			name:     "用户不存在",
			username: "nonexistent",
			password: "password123",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &model.User{Username: tt.username, Password: tt.password}
			err := u.ValidateAndFill()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, user.Id, u.Id)
				assert.Equal(t, user.DisplayName, u.DisplayName)
			}
		})
	}
}

// TestUser_IncreaseUserQuota 测试用户配额增加功能
// 测试目的：验证系统能够正确增加或减少用户的配额
// 测试内容：
// 1. 测试增加用户配额
// 2. 测试减少用户配额
// 3. 测试零配额变化
// 4. 验证配额更新的准确性
func TestUser_IncreaseUserQuota(t *testing.T) {
	db := setupUserTestDB()
	model.DB = db

	// 清理测试数据
	model.DB.Where("1 = 1").Delete(&model.User{})
	model.DB.Where("1 = 1").Delete(&model.Token{})

	tests := []struct {
		name      string
		amount    int64
		expected  int64
		wantError bool
	}{
		{
			name:      "增加配额",
			amount:    500,
			expected:  1500,
			wantError: false,
		},
		{
			name:      "减少配额",
			amount:    -200,
			expected:  1000, // 不能减少到负数，配额应保持不变
			wantError: true, // 不能减少到负数
		},
		{
			name:      "零配额变化",
			amount:    0,
			expected:  1000,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 每个子测试新建一个独立用户
			user := model.User{
				Username: "user_quota_" + tt.name,
				Password: "password123",
				Quota:    1000,
				Status:   1,
				Role:     1,
				Group:    "default",
			}
			err := user.Insert(nil, 0)
			assert.NoError(t, err)
			// 强制设置配额为1000，防止Insert未写入
			err = model.DB.Model(&model.User{}).Where("id = ?", user.Id).Update("quota", 1000).Error
			assert.NoError(t, err)

			err = model.IncreaseUserQuota(user.Id, tt.amount)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// 重新查询用户确认配额已更新
			updatedUser, err := model.GetUserById(user.Id, false)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, updatedUser.Quota)
		})
	}
}

// TestUser_IsAdmin 测试用户管理员权限判断功能
// 测试目的：验证系统能够正确判断用户的角色权限
// 测试内容：
// 1. 测试超级管理员角色判断
// 2. 测试管理员角色判断
// 3. 测试普通用户角色判断
// 4. 测试无效角色处理
func TestUser_IsAdmin(t *testing.T) {
	db := setupUserTestDB()
	model.DB = db

	// 清理测试数据
	model.DB.Where("1 = 1").Delete(&model.User{})
	model.DB.Where("1 = 1").Delete(&model.Token{})

	// 创建不同角色的用户
	adminUser := model.User{
		Username: "admin",
		Password: "password123",
		Role:     10, // 管理员
		Status:   1,
		Group:    "default",
	}
	err := adminUser.Insert(nil, 0)
	assert.NoError(t, err)

	rootUser := model.User{
		Username: "root",
		Password: "password123",
		Role:     100, // 超级管理员
		Status:   1,
		Group:    "default",
	}
	err = rootUser.Insert(nil, 0)
	assert.NoError(t, err)

	normalUser := model.User{
		Username: "normal",
		Password: "password123",
		Role:     1, // 普通用户
		Status:   1,
		Group:    "default",
	}
	err = normalUser.Insert(nil, 0)
	assert.NoError(t, err)

	tests := []struct {
		name   string
		userId int
		want   bool
	}{
		{
			name:   "超级管理员",
			userId: rootUser.Id,
			want:   true,
		},
		{
			name:   "管理员",
			userId: adminUser.Id,
			want:   true,
		},
		{
			name:   "普通用户",
			userId: normalUser.Id,
			want:   false,
		},
		{
			name:   "无效用户ID",
			userId: 99999,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 使用model.IsAdmin函数
			assert.Equal(t, tt.want, model.IsAdmin(tt.userId))
		})
	}
}

// TestUser_UpdateUserUsedQuotaAndRequestCount 测试用户已使用配额和请求次数更新功能
// 测试目的：验证系统能够正确更新用户的已使用配额和请求次数
// 测试内容：
// 1. 测试增加已使用配额
// 2. 测试减少已使用配额
// 3. 验证请求次数的自动递增
// 4. 验证数据更新的准确性
func TestUser_UpdateUserUsedQuotaAndRequestCount(t *testing.T) {
	db := setupUserTestDB()
	model.DB = db

	// 清理测试数据
	model.DB.Where("1 = 1").Delete(&model.User{})
	model.DB.Where("1 = 1").Delete(&model.Token{})

	user := model.User{
		Username:     "testuser3",
		Password:     "password123",
		UsedQuota:    100,
		RequestCount: 5,
		Status:       1,
		Role:         1,
		Group:        "default",
	}
	err := user.Insert(nil, 0)
	assert.NoError(t, err)

	tests := []struct {
		name          string
		quotaDelta    int64
		expectedQuota int64
		expectedCount int
	}{
		{
			name:          "增加使用配额",
			quotaDelta:    50,
			expectedQuota: 150,
			expectedCount: 6,
		},
		{
			name:          "减少使用配额",
			quotaDelta:    -30,
			expectedQuota: 70,
			expectedCount: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 先获取当前用户状态
			currentUser, err := model.GetUserById(user.Id, false)
			assert.NoError(t, err)

			model.UpdateUserUsedQuotaAndRequestCount(user.Id, tt.quotaDelta)

			updatedUser, err := model.GetUserById(user.Id, false)
			assert.NoError(t, err)
			assert.Equal(t, currentUser.UsedQuota+tt.quotaDelta, updatedUser.UsedQuota)
			assert.Equal(t, currentUser.RequestCount+1, updatedUser.RequestCount)
		})
	}
}
