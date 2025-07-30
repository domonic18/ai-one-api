-- OneAPI扩展日志表v3.0版本迁移脚本
-- 将扩展日志表结构从v1.0迁移到v3.0版本，使用JSON字段存储维度信息

-- 1. 检查是否存在旧版本扩展日志表
SET @table_exists = 0;
SELECT COUNT(*) INTO @table_exists 
FROM information_schema.tables 
WHERE table_schema = DATABASE() 
  AND table_name = 'extended_logs';

-- 2. 如果存在旧表，备份数据
SET @backup_sql = CASE 
    WHEN @table_exists > 0 THEN 'CREATE TABLE extended_logs_backup AS SELECT * FROM extended_logs'
    ELSE 'SELECT "No existing table to backup" as message'
END;
PREPARE backup_stmt FROM @backup_sql;
EXECUTE backup_stmt;
DEALLOCATE PREPARE backup_stmt;

-- 3. 删除旧表（如果存在）
DROP TABLE IF EXISTS extended_logs;

-- 4. 创建新的扩展日志表结构（v3.0版本）
CREATE TABLE IF NOT EXISTS extended_logs (
    id bigint PRIMARY KEY AUTO_INCREMENT,
    
    -- 关联原始日志
    log_id bigint NOT NULL,                          -- 关联OneAPI原始日志ID
    
    -- 抽象化的身份标识信息（不耦合具体业务）
    external_user_id varchar(100) DEFAULT '',       -- 外部用户ID（如teacher_id）
    user_group varchar(100) DEFAULT '',             -- OneAPI用户组
    
    -- 多维度统计维度信息（JSON格式，灵活扩展）
    dimension_info JSON DEFAULT NULL,               -- 维度信息，如：{"school_id": 1, "school_name": "北京中学", "subject_id": 10, "subject_name": "数学组"}
    
    -- 管理字段
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- 索引优化
    INDEX idx_log_id (log_id),
    INDEX idx_external_user_id (external_user_id),
    INDEX idx_user_group (user_group),
    INDEX idx_created_at (created_at),
    
    -- JSON字段索引（MySQL 5.7+支持）
    INDEX idx_school_id ((CAST(dimension_info->'$.school_id' AS UNSIGNED))),
    INDEX idx_subject_id ((CAST(dimension_info->'$.subject_id' AS UNSIGNED))),
    
    -- 外键约束
    FOREIGN KEY (log_id) REFERENCES logs(id) ON DELETE CASCADE
);

-- 5. 迁移现有数据（如果备份表存在）
SET @migrate_sql = CASE 
    WHEN @table_exists > 0 THEN 
        'INSERT INTO extended_logs (log_id, external_user_id, user_group, dimension_info, created_at, updated_at)
         SELECT 
             LogId as log_id,
             TeacherId as external_user_id,
             GroupName as user_group,
             JSON_OBJECT(
                 "school_id", SchoolId,
                 "school_name", SchoolName,
                 "subject_id", SubjectId,
                 "subject_name", SubjectName,
                 "teacher_name", TeacherName
             ) as dimension_info,
             FROM_UNIXTIME(CreatedAt) as created_at,
             FROM_UNIXTIME(CreatedAt) as updated_at
         FROM extended_logs_backup
         WHERE LogId IS NOT NULL'
    ELSE 'SELECT "No data to migrate" as message'
END;
PREPARE migrate_stmt FROM @migrate_sql;
EXECUTE migrate_stmt;
DEALLOCATE PREPARE migrate_stmt;

-- 6. 验证迁移结果
SET @verify_sql = CASE 
    WHEN @table_exists > 0 THEN 
        'SELECT 
             (SELECT COUNT(*) FROM extended_logs_backup) as backup_count,
             (SELECT COUNT(*) FROM extended_logs) as new_count,
             CASE 
                 WHEN (SELECT COUNT(*) FROM extended_logs_backup) = (SELECT COUNT(*) FROM extended_logs) 
                 THEN "迁移成功" 
                 ELSE "迁移失败，数据量不匹配" 
             END as migration_status'
    ELSE 'SELECT "新安装，无需迁移" as migration_status'
END;
PREPARE verify_stmt FROM @verify_sql;
EXECUTE verify_stmt;
DEALLOCATE PREPARE verify_stmt;

-- 7. 清理备份表（可选，建议在确认迁移成功后手动执行）
-- DROP TABLE IF EXISTS extended_logs_backup; 