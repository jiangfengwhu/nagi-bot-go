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
	SystemPrompt        string    `json:"system_prompt"`
}

func (db *DB) CreateUser(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (tg_id, username, created_at, total_recharged_token, total_used_token, system_prompt)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := db.GetPool().Exec(ctx, query, user.TgId, user.Username, user.CreatedAt, user.TotalRechargedToken, user.TotalUsedToken, user.SystemPrompt)
	return err
}

func (db *DB) GetUser(ctx context.Context, tgId int64) (*User, error) {
	// 创建带5秒超时的context
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT id, tg_id, username, created_at, total_recharged_token, total_used_token, system_prompt
		FROM users
		WHERE tg_id = $1
	`
	row := db.GetPool().QueryRow(timeoutCtx, query, tgId)

	var user User
	err := row.Scan(&user.ID, &user.TgId, &user.Username, &user.CreatedAt, &user.TotalRechargedToken, &user.TotalUsedToken, &user.SystemPrompt)
	if err != nil {
		if err == pgx.ErrNoRows {
			// 用户不存在，返回 nil
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (db *DB) UpdateUserTotalUsedToken(ctx context.Context, id int, usedToken int64) error {
	query := `
		UPDATE users
		SET total_used_token = total_used_token + $1
		WHERE id = $2
	`
	_, err := db.GetPool().Exec(ctx, query, usedToken, id)
	return err
}

// UpdateUserTotalUsedTokenTx 在事务中更新用户总使用token数
func (db *DB) UpdateUserTotalUsedTokenTx(ctx context.Context, tx pgx.Tx, id int, usedToken int64) error {
	query := `
		UPDATE users
		SET total_used_token = total_used_token + $1
		WHERE id = $2
	`
	_, err := tx.Exec(ctx, query, usedToken, id)
	return err
}

func (db *DB) UpdateUserTotalRechargedToken(ctx context.Context, id int, rechargedToken int64) error {
	query := `
		UPDATE users
		SET total_recharged_token = total_recharged_token + $1
		WHERE id = $2
	`
	_, err := db.GetPool().Exec(ctx, query, rechargedToken, id)
	return err
}

// 查看用户现有token数
func (db *DB) GetUserTotalToken(ctx context.Context, id int) (int64, error) {
	query := `
		SELECT total_recharged_token - total_used_token FROM users WHERE id = $1
	`
	var totalToken int64
	err := db.GetPool().QueryRow(ctx, query, id).Scan(&totalToken)
	return totalToken, err
}
