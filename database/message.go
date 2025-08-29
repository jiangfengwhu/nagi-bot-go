package database

import (
	"context"
	"encoding/json"
	"time"

	"google.golang.org/genai"
)

// Message 表示一条消息
type Message struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	Role       string    `json:"role"`
	Content    any       `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
	LLMAPIType string    `json:"llm_api_type"`
}

// AddMessage 添加新消息
func (db *DB) AddMessage(ctx context.Context, userID int, role string, content any) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 序列化content为JSONB
	contentBytes, err := json.Marshal(content)
	if err != nil {
		return err
	}

	// 插入新消息
	_, err = tx.Exec(ctx, `
		INSERT INTO messages (user_id, role, content)
		VALUES ($1, $2, $3)
	`, userID, role, contentBytes)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// GetUserMessages 获取用户的所有消息，按创建时间排序
func (db *DB) GetUserMessages(ctx context.Context, userID int) ([]Message, error) {
	query := `
		SELECT id, user_id, role, content, created_at, llm_api_type
		FROM messages
		WHERE user_id = $1
		ORDER BY created_at ASC
	`

	rows, err := db.GetPool().Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		var contentBytes []byte

		err := rows.Scan(
			&msg.ID,
			&msg.UserID,
			&msg.Role,
			&contentBytes,
			&msg.CreatedAt,
			&msg.LLMAPIType,
		)
		if err != nil {
			return nil, err
		}

		// 反序列化content
		var parts []*genai.Part
		err = json.Unmarshal(contentBytes, &parts)
		if err != nil {
			return nil, err
		}

		msg.Content = parts

		messages = append(messages, msg)
	}

	return messages, rows.Err()
}

// GetRecentMessages 获取用户最近的N条消息
func (db *DB) GetRecentMessages(ctx context.Context, userID int, limit int) ([]Message, error) {
	query := `
		SELECT id, user_id, role, content, created_at, llm_api_type
		FROM messages
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := db.GetPool().Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		var contentBytes []byte

		err := rows.Scan(
			&msg.ID,
			&msg.UserID,
			&msg.Role,
			&contentBytes,
			&msg.CreatedAt,
			&msg.LLMAPIType,
		)
		if err != nil {
			return nil, err
		}

		// 反序列化content
		var parts []*genai.Part
		err = json.Unmarshal(contentBytes, &parts)
		if err != nil {
			return nil, err
		}

		msg.Content = parts

		messages = append(messages, msg)
	}

	// 反转顺序，使其按时间正序
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, rows.Err()
}

// ClearUserMessages 清空用户的所有消息
func (db *DB) ClearUserMessages(ctx context.Context, userID int) error {
	_, err := db.GetPool().Exec(ctx, "DELETE FROM messages WHERE user_id = $1", userID)
	return err
}

// GetMessageCount 获取用户的消息数量
func (db *DB) GetMessageCount(ctx context.Context, userID int) (int, error) {
	var count int
	err := db.GetPool().QueryRow(ctx, "SELECT COUNT(*) FROM messages WHERE user_id = $1", userID).Scan(&count)
	return count, err
}
