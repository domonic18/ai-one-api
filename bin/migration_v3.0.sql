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

-- 检查user_groups字段是否已存在
SELECT COUNT(*) INTO @user_groups_column_exists
FROM INFORMATION_SCHEMA.COLUMNS 
WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'tokens' 
    AND COLUMN_NAME = 'user_groups';

-- 如果user_groups字段不存在，则添加
SET @add_user_groups_sql = IF(@user_groups_column_exists = 0, 
    'ALTER TABLE tokens ADD COLUMN `user_groups` TEXT DEFAULT NULL COMMENT "令牌所属用户组列表（JSON数组，按优先级）"',
    'SELECT "tokens表user_groups字段已存在，跳过添加" as message'
);

PREPARE stmt FROM @add_user_groups_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 为新创建的令牌设置默认user_groups值（如果user_groups字段为空）
UPDATE tokens 
SET `user_groups` = JSON_ARRAY('default') 
WHERE `user_groups` IS NULL OR `user_groups` = '' OR `user_groups` = '[]';
SELECT CONCAT('更新了 ', ROW_COUNT(), ' 条令牌记录的user_groups字段为默认值') as message;

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
        log_id bigint NOT NULL,
        external_user_id varchar(100) DEFAULT "",
        user_group varchar(100) DEFAULT "",
        dimension_info JSON DEFAULT NULL,
        created_at datetime(3) DEFAULT NULL,
        updated_at datetime(3) DEFAULT NULL
    )',
    'SELECT "extended_logs表已存在，跳过创建" as message'
);

PREPARE stmt FROM @create_table_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 为extended_logs表添加索引（如果表是新创建的）
SET @add_indexes_sql = IF(@extended_logs_exists = 0, 
    'ALTER TABLE extended_logs 
     ADD UNIQUE INDEX idx_log_id_unique (log_id),
     ADD INDEX idx_external_user_id (external_user_id),
     ADD INDEX idx_user_group (user_group),
     ADD INDEX idx_created_at (created_at)',
    'SELECT "extended_logs表已存在，跳过索引创建" as message'
);

PREPARE stmt FROM @add_indexes_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ========================================
-- 第三部分：迁移完成状态检查
-- ========================================

-- 检查tokens表的user_groups字段状态
SELECT 
    'tokens表user_groups字段状态' as check_type,
    COLUMN_NAME,
    COLUMN_DEFAULT,
    IS_NULLABLE,
    COLUMN_COMMENT
FROM INFORMATION_SCHEMA.COLUMNS 
WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'tokens' 
    AND COLUMN_NAME = 'user_groups';

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
    AND TABLE_NAME = 'extended_logs'
    AND INDEX_NAME IN ('idx_log_id_unique', 'idx_external_user_id', 'idx_user_group', 'idx_created_at')
ORDER BY TABLE_NAME, INDEX_NAME;

-- 提交事务
COMMIT;

-- 迁移完成提示
SELECT 'OneAPI v3.0 综合数据库迁移完成，所有更改已提交' as status; 