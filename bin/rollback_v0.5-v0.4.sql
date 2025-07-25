-- One-API v0.5 到 v0.4 数据库回滚脚本
-- 删除扩展日志表，回滚多学校费用分别计算功能
-- 支持 MySQL、PostgreSQL 和 SQLite

-- MySQL 版本
-- 删除扩展日志表
DROP TABLE IF EXISTS `extended_logs`;

-- PostgreSQL 版本（如果需要手动执行）
/*
DROP TABLE IF EXISTS extended_logs CASCADE;
*/

-- SQLite 版本（如果需要手动执行）
/*
DROP TABLE IF EXISTS extended_logs;
*/ 