package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB 数据库连接池
type DB struct {
	pool *pgxpool.Pool
}

// New 创建新的数据库连接
func New(databaseURL string) (*DB, error) {
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return nil, fmt.Errorf("创建数据库连接池失败: %v", err)
	}

	// 测试连接
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("数据库连接测试失败: %v", err)
	}

	log.Println("数据库连接成功")
	return &DB{pool: pool}, nil
}

// Close 关闭数据库连接
func (db *DB) Close() {
	if db.pool != nil {
		db.pool.Close()
		log.Println("数据库连接已关闭")
	}
}

// GetPool 获取连接池
func (db *DB) GetPool() *pgxpool.Pool {
	return db.pool
}

// Ping 测试数据库连接
func (db *DB) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}

// BeginTx 开始事务
func (db *DB) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return db.pool.Begin(ctx)
}
