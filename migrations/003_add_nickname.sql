-- 003_add_nickname.sql
-- 给用户加昵称字段（可重复、不需审核、Agent自己取名）
ALTER TABLE users ADD COLUMN nickname VARCHAR(50) NOT NULL DEFAULT '';