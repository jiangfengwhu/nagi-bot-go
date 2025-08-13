#!/bin/bash

# 配置变量
DB_HOST="localhost"          # 数据库主机，默认 localhost
DB_PORT="5432"               # 默认端口
DB_NAME="nagi"         # 数据库名（需已存在）
DB_USER="fengjiang"           # 用户名
DB_PASS=""      # 密码（生产环境使用环境变量或 .pgpass）

# 执行 SQL
echo "Connecting to PostgreSQL database: $DB_NAME"
PGPASSWORD="$DB_PASS" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "scripts/table.sql"

# 检查执行结果
if [ $? -eq 0 ]; then
    echo "Table 'users' created successfully."
else
    echo "Error creating table."
fi