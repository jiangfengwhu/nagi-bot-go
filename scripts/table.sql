CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    tg_id BIGINT NOT NULL UNIQUE,
    username VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    total_recharged_token BIGINT NOT NULL,
    total_used_token BIGINT NOT NULL,
    system_prompt TEXT NOT NULL
);

-- 消息表，存储用户的消息历史
CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL CHECK (role IN ('user', 'model', 'system')),
    content JSONB NOT NULL, -- 存储完整的parts数组内容，包括thoughtSignature
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    llm_api_type VARCHAR(50) NOT NULL DEFAULT 'gemini'
);

-- 人物属性表，存储修仙者的基本属性（参考凡人修仙传设定）
CREATE TABLE IF NOT EXISTS character_stats (
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE UNIQUE,
    name VARCHAR(50) NOT NULL, -- 修仙者姓名
    
    -- 修炼境界 (筑基期、结丹期、元婴期等)
    realm VARCHAR(20) NOT NULL DEFAULT '练气期',
    realm_level INTEGER NOT NULL DEFAULT 1, -- 境界内的小层次 (如练气一层、二层...)
    
    -- 灵根属性
    spiritual_roots JSONB,

    -- 神识
    spirit_sense INTEGER NOT NULL DEFAULT 0,

    -- 根骨/体魄
    physique INTEGER NOT NULL DEFAULT 10,

    -- 煞气/心魔
    demonic_aura INTEGER NOT NULL DEFAULT 0,

    -- 道号
    taoist_name VARCHAR(50),

    -- 基础属性
    hp INTEGER NOT NULL DEFAULT 100,           -- 生命值
    max_hp INTEGER NOT NULL DEFAULT 100,       -- 最大生命值
    mp INTEGER NOT NULL DEFAULT 50,            -- 法力值
    max_mp INTEGER NOT NULL DEFAULT 50,        -- 最大法力值
    
    -- 战斗属性
    attack INTEGER NOT NULL DEFAULT 10,        -- 攻击力
    defense INTEGER NOT NULL DEFAULT 5,        -- 防御力
    speed INTEGER NOT NULL DEFAULT 10,         -- 速度
    luck INTEGER NOT NULL DEFAULT 5,           -- 幸运值
    
    -- 修炼相关
    experience BIGINT NOT NULL DEFAULT 0,      -- 修炼经验
    comprehension INTEGER NOT NULL DEFAULT 10, -- 悟性
    
    -- 寿命相关
    age INTEGER NOT NULL DEFAULT 16,           -- 当前年龄
    lifespan INTEGER NOT NULL DEFAULT 100,     -- 寿命上限
    
    -- 位置信息
    location VARCHAR(100) NOT NULL DEFAULT '新手村', -- 当前位置
    
    -- 状态
    status VARCHAR(20) NOT NULL DEFAULT '健康',

    -- 成长经历
    stories TEXT NOT NULL DEFAULT '' -- 成长经历
);

-- 背包系统表 - 存储物品信息
CREATE TABLE IF NOT EXISTS inventory (
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    item_name VARCHAR(100) NOT NULL, -- 物品名称
    item_type VARCHAR(50) NOT NULL,  -- 物品类型 (weapon, armor, pill, material, book, talisman)
    
    -- 物品品质和等级
    quality VARCHAR(20) NOT NULL DEFAULT '普通', -- 品质 (common, uncommon, rare, epic, legendary, immortal)
    level INTEGER NOT NULL DEFAULT 1,               -- 物品等级
    
    -- 数量和堆叠
    quantity INTEGER NOT NULL DEFAULT 1,            -- 数量
    
    -- 物品属性
    properties TEXT,  -- 存储物品的具体属性 (攻击力、防御力、特殊效果等)
    
    -- 物品描述
    description TEXT,    -- 物品描述
    
    -- 获得信息
    obtained_from VARCHAR(100), -- 获得来源 (shop, monster, quest, craft)
    obtained_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 修炼功法表
CREATE TABLE IF NOT EXISTS cultivation_techniques (
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    technique_name VARCHAR(100) NOT NULL,     -- 功法名称
    technique_type VARCHAR(50) NOT NULL,      -- 功法类型 (cultivation, combat, movement, auxiliary)
    
    -- 功法等级
    technique_level INTEGER NOT NULL DEFAULT 1, -- 功法层数
    
    -- 功法品质
    quality VARCHAR(20) NOT NULL DEFAULT '低阶', -- 品质 (mortal, spiritual, earth, heaven, immortal)
    
    -- 修炼进度
    progress INTEGER NOT NULL DEFAULT 0,        -- 当前修炼进度 (0-100)
    
    -- 功法效果
    effects JSONB, -- 功法提供的效果和加成
    
    -- 修炼要求
    requirements JSONB, -- 修炼要求 (境界、灵根等)
    
    -- 获得信息
    learned_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 消息表索引
CREATE INDEX IF NOT EXISTS idx_messages_user_id ON messages(user_id);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
CREATE INDEX IF NOT EXISTS idx_messages_user_created ON messages(user_id, created_at);

-- 人物属性表索引
CREATE INDEX IF NOT EXISTS idx_character_stats_user_id ON character_stats(user_id);
CREATE INDEX IF NOT EXISTS idx_character_stats_realm ON character_stats(realm, realm_level);
CREATE INDEX IF NOT EXISTS idx_character_stats_location ON character_stats(location);

-- 背包表索引
CREATE INDEX IF NOT EXISTS idx_inventory_user_id ON inventory(user_id);
CREATE INDEX IF NOT EXISTS idx_inventory_item_type ON inventory(item_type);
CREATE INDEX IF NOT EXISTS idx_inventory_user_type ON inventory(user_id, item_type);

-- 功法表索引
CREATE INDEX IF NOT EXISTS idx_cultivation_techniques_user_id ON cultivation_techniques(user_id);
CREATE INDEX IF NOT EXISTS idx_cultivation_techniques_type ON cultivation_techniques(technique_type);
CREATE INDEX IF NOT EXISTS idx_cultivation_techniques_quality ON cultivation_techniques(quality);

