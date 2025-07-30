-- OneAPI扩展日志表v3.0版本创建脚本
-- 创建扩展日志表，使用JSON字段存储维度信息，支持灵活的多维度统计

-- 创建扩展日志表结构（v3.0版本）
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
    
    -- 基础索引
    INDEX idx_log_id (log_id),
    INDEX idx_external_user_id (external_user_id),
    INDEX idx_user_group (user_group),
    INDEX idx_created_at (created_at)
);

-- 创建成功提示
SELECT 'OneAPI扩展日志表v3.0版本创建完成' as status; 