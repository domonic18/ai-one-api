package storage

import (
	"fmt"
	"sync"

	"mock-courseware-platform/models"
)

// MemoryStorage 内存存储结构体
type MemoryStorage struct {
	users  map[string]*models.TeacherInfo
	config *models.ServerConfig
	mu     sync.RWMutex
}

// NewMemoryStorage 创建新的内存存储实例
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		users:  make(map[string]*models.TeacherInfo),
		config: &models.ServerConfig{},
	}
}

// GetUser 获取单个用户信息
func (m *MemoryStorage) GetUser(teacherId string) (*models.TeacherInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if user, exists := m.users[teacherId]; exists {
		return user, nil
	}
	return nil, fmt.Errorf("user not found: %s", teacherId)
}

// GetAllUserIds 获取所有用户ID列表
func (m *MemoryStorage) GetAllUserIds() ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.users))
	for id := range m.users {
		ids = append(ids, id)
	}
	return ids, nil
}

// BatchGetUsers 批量获取用户信息
func (m *MemoryStorage) BatchGetUsers(teacherIds []string) ([]*models.TeacherInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	users := make([]*models.TeacherInfo, 0, len(teacherIds))
	for _, id := range teacherIds {
		if user, exists := m.users[id]; exists {
			users = append(users, user)
		}
	}
	return users, nil
}

// AddUser 添加用户
func (m *MemoryStorage) AddUser(user *models.TeacherInfo) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if user.TeacherId == "" {
		return fmt.Errorf("teacher_id cannot be empty")
	}

	m.users[user.TeacherId] = user
	return nil
}

// UpdateUser 更新用户信息
func (m *MemoryStorage) UpdateUser(user *models.TeacherInfo) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if user.TeacherId == "" {
		return fmt.Errorf("teacher_id cannot be empty")
	}

	if _, exists := m.users[user.TeacherId]; !exists {
		return fmt.Errorf("user not found: %s", user.TeacherId)
	}

	m.users[user.TeacherId] = user
	return nil
}

// DeleteUser 删除用户
func (m *MemoryStorage) DeleteUser(teacherId string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.users[teacherId]; !exists {
		return fmt.Errorf("user not found: %s", teacherId)
	}

	delete(m.users, teacherId)
	return nil
}

// GetAllUsers 获取所有用户信息
func (m *MemoryStorage) GetAllUsers() ([]*models.TeacherInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	users := make([]*models.TeacherInfo, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, user)
	}
	return users, nil
}

// LoadDefaultUsers 加载默认用户数据
func (m *MemoryStorage) LoadDefaultUsers(users []*models.TeacherInfo) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, user := range users {
		if user.TeacherId != "" {
			m.users[user.TeacherId] = user
		}
	}
	return nil
}

// GetConfig 获取配置
func (m *MemoryStorage) GetConfig() *models.ServerConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// SetConfig 设置配置
func (m *MemoryStorage) SetConfig(config *models.ServerConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = config
}

// Clear 清空所有数据
func (m *MemoryStorage) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users = make(map[string]*models.TeacherInfo)
}
