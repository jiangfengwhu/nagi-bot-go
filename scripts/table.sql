CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    tg_id BIGINT NOT NULL UNIQUE,
    username VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    total_recharged_token BIGINT NOT NULL,
    total_used_token BIGINT NOT NULL,
    system_prompt TEXT NOT NULL
);

-- 消息表，存储用户的消息历史（每个用户最多100条）
CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL CHECK (role IN ('user', 'model', 'system')),
    content JSONB NOT NULL, -- 存储完整的parts数组内容，包括thoughtSignature
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    llm_api_type VARCHAR(50) NOT NULL DEFAULT 'gemini'
);

-- 创建索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_messages_user_id ON messages(user_id);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
CREATE INDEX IF NOT EXISTS idx_messages_user_created ON messages(user_id, created_at);

