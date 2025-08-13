package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type User struct {
	ID                  int       `json:"id"`
	TgId                int64     `json:"tg_id"`
	Username            string    `json:"username"`
	CreatedAt           time.Time `json:"created_at"`
	TotalRechargedToken int64     `json:"total_recharged_token"`
	TotalUsedToken      int64     `json:"total_used_token"`
}

func (db *DB) CreateUser(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (tg_id, username, created_at, total_recharged_token, total_used_token)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := db.GetPool().Exec(ctx, query, user.TgId, user.Username, user.CreatedAt, user.TotalRechargedToken, user.TotalUsedToken)
	return err
}

func (db *DB) GetUser(ctx context.Context, tgId int64) (*User, error) {
	// 创建带5秒超时的context
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT id, tg_id, username, created_at, total_recharged_token, total_used_token
		FROM users
		WHERE tg_id = $1
	`
	row := db.GetPool().QueryRow(timeoutCtx, query, tgId)

	var user User
	err := row.Scan(&user.ID, &user.TgId, &user.Username, &user.CreatedAt, &user.TotalRechargedToken, &user.TotalUsedToken)
	if err != nil {
		if err == pgx.ErrNoRows {
			// 用户不存在，返回 nil
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}
