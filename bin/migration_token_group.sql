-- 为tokens表添加group字段的迁移脚本
-- 执行时间：2025-08-04

-- 为tokens表添加group字段
ALTER TABLE tokens ADD COLUMN `group` VARCHAR(255) DEFAULT 'default' COMMENT '令牌所属用户组';

-- 创建索引以提高查询性能
CREATE INDEX idx_tokens_group ON tokens(`group`);

-- 更新现有令牌的group字段为default（如果有需要的话）
UPDATE tokens SET `group` = 'default' WHERE `group` IS NULL OR `group` = '';