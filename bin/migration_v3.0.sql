-- OneAPI v3.0 综合数据库迁移脚本（增强版）
-- 包含令牌用户组功能和扩展日志表功能
-- 执行时间：2025-08-04
-- 版本：v3.0
-- 特性：事务支持、状态检查

-- 开启事务支持
START TRANSACTION;

-- ========================================
-- 第一部分：令牌用户组功能迁移
-- ========================================

-- 检查tokens表是否存在
SELECT COUNT(*) INTO @tokens_table_exists
FROM INFORMATION_SCHEMA.TABLES 
WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'tokens';

-- 如果tokens表存在，执行迁移
SELECT IF(@tokens_table_exists > 0, 'tokens表存在，开始迁移', 'tokens表不存在，跳过令牌用户组功能迁移') as status;

-- 检查group字段是否已存在
SELECT COUNT(*) INTO @group_column_exists
FROM INFORMATION_SCHEMA.COLUMNS 
WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'tokens' 
    AND COLUMN_NAME = 'group';

-- 如果group字段不存在，则添加
SET @add_group_sql = IF(@group_column_exists = 0, 
    'ALTER TABLE tokens ADD COLUMN `group` VARCHAR(255) DEFAULT "default" COMMENT "令牌所属用户组"',
    'SELECT "tokens表group字段已存在，跳过添加" as message'
);

PREPARE stmt FROM @add_group_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 检查索引是否已存在
SELECT COUNT(*) INTO @group_index_exists
FROM INFORMATION_SCHEMA.STATISTICS 
WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'tokens' 
    AND INDEX_NAME = 'idx_tokens_group';

-- 如果索引不存在，则创建
SET @create_index_sql = IF(@group_index_exists = 0, 
    'CREATE INDEX idx_tokens_group ON tokens(`group`)',
    'SELECT "tokens表group字段索引已存在，跳过创建" as message'
);

PREPARE stmt FROM @create_index_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 更新现有令牌的group字段为default（如果有需要的话）
UPDATE tokens SET `group` = 'default' WHERE `group` IS NULL OR `group` = '';
SELECT CONCAT('更新了 ', ROW_COUNT(), ' 条令牌记录的group字段') as message;

-- ========================================
-- 第二部分：扩展日志表功能迁移
-- ========================================

-- 检查extended_logs表是否已存在
SELECT COUNT(*) INTO @extended_logs_exists
FROM INFORMATION_SCHEMA.TABLES 
WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'extended_logs';

-- 如果extended_logs表不存在，则创建
SET @create_table_sql = IF(@extended_logs_exists = 0, 
    'CREATE TABLE extended_logs (
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
        
        -- 基础索引
        INDEX idx_log_id (log_id),
        INDEX idx_external_user_id (external_user_id),
        INDEX idx_user_group (user_group),
        INDEX idx_created_at (created_at)
    )',
    'SELECT "extended_logs表已存在，跳过创建" as message'
);

PREPARE stmt FROM @create_table_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ========================================
-- 第三部分：迁移完成状态检查
-- ========================================

-- 检查tokens表的group字段状态
SELECT 
    'tokens表group字段状态' as check_type,
    COLUMN_NAME,
    COLUMN_DEFAULT,
    IS_NULLABLE,
    COLUMN_COMMENT
FROM INFORMATION_SCHEMA.COLUMNS 
WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'tokens' 
    AND COLUMN_NAME = 'group';

-- 检查extended_logs表状态
SELECT 
    'extended_logs表状态' as check_type,
    TABLE_NAME,
    TABLE_COMMENT
FROM INFORMATION_SCHEMA.TABLES 
WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'extended_logs';

-- 检查相关索引状态
SELECT 
    '索引状态' as check_type,
    TABLE_NAME,
    INDEX_NAME,
    COLUMN_NAME
FROM INFORMATION_SCHEMA.STATISTICS 
WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME IN ('tokens', 'extended_logs')
    AND INDEX_NAME IN ('idx_tokens_group', 'idx_log_id', 'idx_external_user_id', 'idx_user_group', 'idx_created_at')
ORDER BY TABLE_NAME, INDEX_NAME;

-- 提交事务
COMMIT;

-- 迁移完成提示
SELECT 'OneAPI v3.0 综合数据库迁移完成，所有更改已提交' as status; 