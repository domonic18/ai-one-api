-- One-API v0.4 到 v0.5 数据库迁移脚本
-- 创建扩展日志表，支持多学校费用分别计算功能
-- 支持 MySQL、PostgreSQL 和 SQLite

-- MySQL 版本
-- 创建扩展日志表
CREATE TABLE IF NOT EXISTS `extended_logs` (
    `id` INT PRIMARY KEY AUTO_INCREMENT,
    `log_id` INT NOT NULL COMMENT '关联原日志表ID',
    `school_id` INT DEFAULT 0 COMMENT '学校ID',
    `school_name` VARCHAR(100) DEFAULT '' COMMENT '学校名称',
    `subject_id` INT DEFAULT 0 COMMENT '学科组ID',
    `subject_name` VARCHAR(100) DEFAULT '' COMMENT '学科组名称',
    `teacher_id` VARCHAR(100) DEFAULT '' COMMENT '老师ID',
    `teacher_name` VARCHAR(100) DEFAULT '' COMMENT '老师姓名',
    `group_name` VARCHAR(100) DEFAULT '' COMMENT '用户组名称',
    `created_at` BIGINT NOT NULL COMMENT '创建时间',
    
    UNIQUE KEY `idx_log_id` (`log_id`),
    INDEX `idx_school_id` (`school_id`),
    INDEX `idx_subject_id` (`subject_id`),
    INDEX `idx_teacher_id` (`teacher_id`),
    INDEX `idx_created_at` (`created_at`),
    FOREIGN KEY (`log_id`) REFERENCES `logs`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='扩展日志表，用于记录学校、学科组、老师等维度信息';

-- PostgreSQL 版本（如果需要手动执行）
/*
CREATE TABLE IF NOT EXISTS extended_logs (
    id SERIAL PRIMARY KEY,
    log_id INTEGER NOT NULL,
    school_id INTEGER DEFAULT 0,
    school_name VARCHAR(100) DEFAULT '',
    subject_id INTEGER DEFAULT 0,
    subject_name VARCHAR(100) DEFAULT '',
    teacher_id VARCHAR(100) DEFAULT '',
    teacher_name VARCHAR(100) DEFAULT '',
    group_name VARCHAR(100) DEFAULT '',
    created_at BIGINT NOT NULL,
    
    CONSTRAINT idx_log_id UNIQUE (log_id),
    CONSTRAINT fk_extended_logs_log_id FOREIGN KEY (log_id) REFERENCES logs(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_extended_logs_school_id ON extended_logs(school_id);
CREATE INDEX IF NOT EXISTS idx_extended_logs_subject_id ON extended_logs(subject_id);
CREATE INDEX IF NOT EXISTS idx_extended_logs_teacher_id ON extended_logs(teacher_id);
CREATE INDEX IF NOT EXISTS idx_extended_logs_created_at ON extended_logs(created_at);

COMMENT ON TABLE extended_logs IS '扩展日志表，用于记录学校、学科组、老师等维度信息';
COMMENT ON COLUMN extended_logs.log_id IS '关联原日志表ID';
COMMENT ON COLUMN extended_logs.school_id IS '学校ID';
COMMENT ON COLUMN extended_logs.school_name IS '学校名称';
COMMENT ON COLUMN extended_logs.subject_id IS '学科组ID';
COMMENT ON COLUMN extended_logs.subject_name IS '学科组名称';
COMMENT ON COLUMN extended_logs.teacher_id IS '老师ID';
COMMENT ON COLUMN extended_logs.teacher_name IS '老师姓名';
COMMENT ON COLUMN extended_logs.group_name IS '用户组名称';
COMMENT ON COLUMN extended_logs.created_at IS '创建时间';
*/

-- SQLite 版本（如果需要手动执行）
/*
CREATE TABLE IF NOT EXISTS extended_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    log_id INTEGER NOT NULL,
    school_id INTEGER DEFAULT 0,
    school_name TEXT DEFAULT '',
    subject_id INTEGER DEFAULT 0,
    subject_name TEXT DEFAULT '',
    teacher_id TEXT DEFAULT '',
    teacher_name TEXT DEFAULT '',
    group_name TEXT DEFAULT '',
    created_at INTEGER NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_log_id ON extended_logs(log_id);
CREATE INDEX IF NOT EXISTS idx_school_id ON extended_logs(school_id);
CREATE INDEX IF NOT EXISTS idx_subject_id ON extended_logs(subject_id);
CREATE INDEX IF NOT EXISTS idx_teacher_id ON extended_logs(teacher_id);
CREATE INDEX IF NOT EXISTS idx_created_at ON extended_logs(created_at);
*/ 