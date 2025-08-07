package common

import (
	"fmt"
	"log"
	"os"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/model/identity"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// SetupMySQLTestDB 设置 MySQL 测试数据库
func SetupMySQLTestDB() *gorm.DB {
	// 设置测试环境变量
	os.Setenv("SQL_DSN", "testuser:testpass@tcp(localhost:3306)/oneapi_test?charset=utf8mb4&parseTime=True&loc=Local")
	os.Setenv("REDIS_CONN_STRING", "redis://localhost:6379")
	os.Setenv("SESSION_SECRET", "test-secret-key")
	os.Setenv("SYNC_FREQUENCY", "60")

	// 设置 SQLite 标志为 false，使用 MySQL
	common.UsingSQLite = false

	// 初始化 Redis 客户端
	err := common.InitRedisClient()
	if err != nil {
		log.Printf("Warning: Failed to initialize Redis client: %v", err)
		common.RedisEnabled = false
	}

	// 连接 MySQL 数据库
	dsn := "testuser:testpass@tcp(localhost:3306)/oneapi_test?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to MySQL test database: %v", err)
	}

	// 自动迁移所有表结构
	err = db.AutoMigrate(
		&model.User{},
		&model.Token{},
		&model.Channel{},
		&model.Log{},
		&model.Ability{},
		&identity.ExtendedLog{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	fmt.Println("MySQL test database setup completed")
	return db
}

// CleanupTestDB 清理测试数据库
func CleanupTestDB(db *gorm.DB) {
	// 清理所有测试数据
	db.Exec("SET FOREIGN_KEY_CHECKS = 0")
	db.Exec("TRUNCATE TABLE users")
	db.Exec("TRUNCATE TABLE tokens")
	db.Exec("TRUNCATE TABLE channels")
	db.Exec("TRUNCATE TABLE logs")
	db.Exec("TRUNCATE TABLE abilities")
	db.Exec("TRUNCATE TABLE extended_logs")
	db.Exec("SET FOREIGN_KEY_CHECKS = 1")
}
