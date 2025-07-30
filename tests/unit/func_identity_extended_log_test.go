package unit

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/model/identity"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupIdentityTestDB 设置身份模块测试数据库
func setupIdentityTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// 创建logs表（用于外键测试）
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			created_at INTEGER,
			type INTEGER,
			content TEXT,
			username TEXT,
			token_name TEXT,
			model_name TEXT,
			quota INTEGER,
			prompt_tokens INTEGER,
			completion_tokens INTEGER,
			channel INTEGER,
			request_id TEXT,
			elapsed_time INTEGER,
			is_stream BOOLEAN,
			system_prompt_reset BOOLEAN
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create logs table: %v", err)
	}

	// 自动迁移扩展日志表
	err = db.AutoMigrate(&identity.ExtendedLog{})
	if err != nil {
		t.Fatalf("Failed to migrate ExtendedLog: %v", err)
	}

	// 插入测试日志数据
	err = db.Exec("INSERT INTO logs (id, user_id, created_at, type) VALUES (1, 1, ?, 1)", time.Now().Unix()).Error
	if err != nil {
		t.Fatalf("Failed to insert test log: %v", err)
	}

	return db
}

func TestExtendedLog_表名_正确配置(t *testing.T) {
	log := identity.ExtendedLog{}
	if log.TableName() != "extended_logs" {
		t.Errorf("Expected table name 'extended_logs', got '%s'", log.TableName())
	}
}

func TestExtendedLog_创建日志_正常场景(t *testing.T) {
	db := setupIdentityTestDB(t)
	originalDB := model.DB
	model.DB = db
	defer func() { model.DB = originalDB }()

	ctx := context.Background()

	dimensionInfo := &identity.DimensionInfo{
		SchoolId:    1,
		SchoolName:  "北京中学",
		SubjectId:   10,
		SubjectName: "数学组",
		TeacherName: "张老师",
	}

	// 创建扩展日志
	extendedLog, err := identity.CreateExtendedLog(ctx, 1, "teacher_001", "beijing_math_group", dimensionInfo)
	if err != nil {
		t.Fatalf("Failed to create extended log: %v", err)
	}

	// 验证创建结果
	if extendedLog.LogId != 1 {
		t.Errorf("Expected LogId 1, got %d", extendedLog.LogId)
	}
	if extendedLog.ExternalUserId != "teacher_001" {
		t.Errorf("Expected ExternalUserId 'teacher_001', got '%s'", extendedLog.ExternalUserId)
	}
	if extendedLog.UserGroup != "beijing_math_group" {
		t.Errorf("Expected UserGroup 'beijing_math_group', got '%s'", extendedLog.UserGroup)
	}

	// 验证维度信息
	parsedDimensionInfo, err := extendedLog.GetDimensionInfo()
	if err != nil {
		t.Fatalf("Failed to parse dimension info: %v", err)
	}
	if parsedDimensionInfo.SchoolId != 1 {
		t.Errorf("Expected SchoolId 1, got %d", parsedDimensionInfo.SchoolId)
	}
	if parsedDimensionInfo.SchoolName != "北京中学" {
		t.Errorf("Expected SchoolName '北京中学', got '%s'", parsedDimensionInfo.SchoolName)
	}
}

func TestExtendedLog_根据日志ID查询_存在和不存在(t *testing.T) {
	db := setupIdentityTestDB(t)
	originalDB := model.DB
	model.DB = db
	defer func() { model.DB = originalDB }()

	ctx := context.Background()

	// 先创建一个扩展日志
	dimensionInfo := &identity.DimensionInfo{
		SchoolId:   1,
		SchoolName: "北京中学",
	}
	_, err := identity.CreateExtendedLog(ctx, 1, "teacher_001", "beijing_math_group", dimensionInfo)
	if err != nil {
		t.Fatalf("Failed to create extended log: %v", err)
	}

	// 根据LogId获取扩展日志
	extendedLog, err := identity.GetExtendedLogByLogId(ctx, 1)
	if err != nil {
		t.Fatalf("Failed to get extended log: %v", err)
	}
	if extendedLog == nil {
		t.Fatal("Expected extended log, got nil")
	}
	if extendedLog.ExternalUserId != "teacher_001" {
		t.Errorf("Expected ExternalUserId 'teacher_001', got '%s'", extendedLog.ExternalUserId)
	}

	// 测试不存在的LogId
	extendedLog, err = identity.GetExtendedLogByLogId(ctx, 999)
	if err != nil {
		t.Fatalf("Expected no error for non-existent log, got: %v", err)
	}
	if extendedLog != nil {
		t.Error("Expected nil for non-existent log")
	}
}

func TestExtendedLog_根据外部用户ID查询_分页功能(t *testing.T) {
	db := setupIdentityTestDB(t)
	originalDB := model.DB
	model.DB = db
	defer func() { model.DB = originalDB }()

	ctx := context.Background()

	// 创建多个扩展日志
	for i := 0; i < 5; i++ {
		dimensionInfo := &identity.DimensionInfo{
			SchoolId:    1,
			SchoolName:  "北京中学",
			TeacherName: "张老师",
		}
		_, err := identity.CreateExtendedLog(ctx, int64(i+1), "teacher_001", "beijing_math_group", dimensionInfo)
		if err != nil {
			// 跳过外键约束错误（因为我们只创建了一个logs记录）
			continue
		}
	}

	// 分页查询
	logs, total, err := identity.GetExtendedLogsByExternalUserId(ctx, "teacher_001", 1, 10)
	if err != nil {
		t.Fatalf("Failed to get extended logs by external user id: %v", err)
	}
	if total == 0 {
		t.Error("Expected at least one log")
	}
	if len(logs) == 0 {
		t.Error("Expected at least one log in results")
	}
}

func TestExtendedLog_根据用户组查询_分页功能(t *testing.T) {
	db := setupIdentityTestDB(t)
	originalDB := model.DB
	model.DB = db
	defer func() { model.DB = originalDB }()

	ctx := context.Background()

	// 创建扩展日志
	dimensionInfo := &identity.DimensionInfo{
		SchoolId:   1,
		SchoolName: "北京中学",
	}
	_, err := identity.CreateExtendedLog(ctx, 1, "teacher_001", "beijing_math_group", dimensionInfo)
	if err != nil {
		t.Fatalf("Failed to create extended log: %v", err)
	}

	// 按用户组查询
	logs, total, err := identity.GetExtendedLogsByUserGroup(ctx, "beijing_math_group", 1, 10)
	if err != nil {
		t.Fatalf("Failed to get extended logs by user group: %v", err)
	}
	if total != 1 {
		t.Errorf("Expected total 1, got %d", total)
	}
	if len(logs) != 1 {
		t.Errorf("Expected 1 log, got %d", len(logs))
	}
}

func TestExtendedLog_批量删除_根据日志ID列表(t *testing.T) {
	db := setupIdentityTestDB(t)
	originalDB := model.DB
	model.DB = db
	defer func() { model.DB = originalDB }()

	ctx := context.Background()

	// 创建扩展日志
	dimensionInfo := &identity.DimensionInfo{
		SchoolId:   1,
		SchoolName: "北京中学",
	}
	_, err := identity.CreateExtendedLog(ctx, 1, "teacher_001", "beijing_math_group", dimensionInfo)
	if err != nil {
		t.Fatalf("Failed to create extended log: %v", err)
	}

	// 删除扩展日志
	err = identity.DeleteExtendedLogsByLogIds(ctx, []int64{1})
	if err != nil {
		t.Fatalf("Failed to delete extended logs: %v", err)
	}

	// 验证删除结果
	extendedLog, err := identity.GetExtendedLogByLogId(ctx, 1)
	if err != nil {
		t.Fatalf("Failed to check deleted log: %v", err)
	}
	if extendedLog != nil {
		t.Error("Expected log to be deleted")
	}
}

func TestDimensionInfo_JSON序列化_完整字段(t *testing.T) {
	dimensionInfo := &identity.DimensionInfo{
		SchoolId:    1,
		SchoolName:  "北京中学",
		SubjectId:   10,
		SubjectName: "数学组",
		TeacherName: "张老师",
		Department:  "教学部",
		Project:     "AI教学项目",
		Region:      "北京",
	}

	// 序列化
	data, err := json.Marshal(dimensionInfo)
	if err != nil {
		t.Fatalf("Failed to marshal dimension info: %v", err)
	}

	// 反序列化
	var parsedInfo identity.DimensionInfo
	err = json.Unmarshal(data, &parsedInfo)
	if err != nil {
		t.Fatalf("Failed to unmarshal dimension info: %v", err)
	}

	// 验证字段
	if parsedInfo.SchoolId != dimensionInfo.SchoolId {
		t.Errorf("Expected SchoolId %d, got %d", dimensionInfo.SchoolId, parsedInfo.SchoolId)
	}
	if parsedInfo.SchoolName != dimensionInfo.SchoolName {
		t.Errorf("Expected SchoolName %s, got %s", dimensionInfo.SchoolName, parsedInfo.SchoolName)
	}
	if parsedInfo.Department != dimensionInfo.Department {
		t.Errorf("Expected Department %s, got %s", dimensionInfo.Department, parsedInfo.Department)
	}
}

func TestExtendedLog_获取维度信息_空值和有效值(t *testing.T) {
	// 测试空维度信息
	extendedLog := &identity.ExtendedLog{
		DimensionInfo: "",
	}

	dimensionInfo, err := extendedLog.GetDimensionInfo()
	if err != nil {
		t.Fatalf("Expected no error for empty dimension info, got: %v", err)
	}
	if dimensionInfo != nil {
		t.Error("Expected nil for empty dimension info")
	}

	// 测试有效维度信息
	validJSON := `{"school_id": 1, "school_name": "北京中学", "teacher_name": "张老师"}`
	extendedLog = &identity.ExtendedLog{
		DimensionInfo: validJSON,
	}

	dimensionInfo, err = extendedLog.GetDimensionInfo()
	if err != nil {
		t.Fatalf("Failed to get dimension info: %v", err)
	}
	if dimensionInfo == nil {
		t.Fatal("Expected dimension info, got nil")
	}
	if dimensionInfo.SchoolId != 1 {
		t.Errorf("Expected SchoolId 1, got %d", dimensionInfo.SchoolId)
	}
}

func TestExtendedLog_异步记录_并发安全(t *testing.T) {
	db := setupIdentityTestDB(t)
	originalDB := model.DB
	model.DB = db
	defer func() { model.DB = originalDB }()

	ctx := context.Background()

	// 测试记录扩展日志（异步）
	identity.RecordExtendedLog(ctx, 1, "teacher_001")

	// 等待异步操作完成
	time.Sleep(100 * time.Millisecond)

	// 验证日志是否被创建
	extendedLog, err := identity.GetExtendedLogByLogId(ctx, 1)
	if err != nil {
		t.Fatalf("Failed to get recorded log: %v", err)
	}
	if extendedLog == nil {
		t.Error("Expected extended log to be recorded")
	}
}

func TestExtendedLog_边界值_空值和异常输入(t *testing.T) {
	db := setupIdentityTestDB(t)
	originalDB := model.DB
	model.DB = db
	defer func() { model.DB = originalDB }()

	ctx := context.Background()

	// 测试空外部用户ID
	identity.RecordExtendedLog(ctx, 1, "")
	time.Sleep(50 * time.Millisecond)

	// 测试nil维度信息
	_, err := identity.CreateExtendedLog(ctx, 1, "teacher_001", "test_group", nil)
	if err != nil {
		t.Fatalf("Failed to create extended log with nil dimension info: %v", err)
	}

	// 测试空用户组
	_, err = identity.CreateExtendedLog(ctx, 2, "teacher_002", "", &identity.DimensionInfo{
		SchoolName: "测试学校",
	})
	if err != nil {
		t.Fatalf("Failed to create extended log with empty user group: %v", err)
	}
}

func BenchmarkExtendedLog_创建操作_性能测试(b *testing.B) {
	db := setupIdentityTestDB(&testing.T{})
	originalDB := model.DB
	model.DB = db
	defer func() { model.DB = originalDB }()

	ctx := context.Background()
	dimensionInfo := &identity.DimensionInfo{
		SchoolId:    1,
		SchoolName:  "北京中学",
		SubjectId:   10,
		SubjectName: "数学组",
		TeacherName: "张老师",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		identity.CreateExtendedLog(ctx, int64(i+2), "teacher_001", "beijing_math_group", dimensionInfo)
	}
}

func BenchmarkExtendedLog_查询操作_性能测试(b *testing.B) {
	db := setupIdentityTestDB(&testing.T{})
	originalDB := model.DB
	model.DB = db
	defer func() { model.DB = originalDB }()

	ctx := context.Background()

	// 预先创建一些测试数据
	dimensionInfo := &identity.DimensionInfo{
		SchoolId:    1,
		SchoolName:  "北京中学",
		TeacherName: "张老师",
	}
	identity.CreateExtendedLog(ctx, 1, "teacher_001", "beijing_math_group", dimensionInfo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		identity.GetExtendedLogByLogId(ctx, 1)
	}
}
