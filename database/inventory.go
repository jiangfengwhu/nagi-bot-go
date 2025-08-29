package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// InventoryItem 背包物品结构
type InventoryItem struct {
	UserID       int                    `json:"user_id"`
	ItemID       int                    `json:"item_id"`
	ItemName     string                 `json:"item_name"`
	ItemType     string                 `json:"item_type"`
	Quality      string                 `json:"quality"`
	Level        int                    `json:"level"`
	Quantity     int                    `json:"quantity"`
	Properties   map[string]interface{} `json:"properties"`
	Description  *string                `json:"description"`
	ObtainedFrom *string                `json:"obtained_from"`
	ObtainedAt   time.Time              `json:"obtained_at"`
}

// AddInventoryItem 添加物品到背包
func (db *DB) AddInventoryItem(ctx context.Context, item *InventoryItem) error {
	// 先检查是否已存在相同物品，如果存在则增加数量
	existingItem, err := db.GetInventoryItemByID(ctx, item.UserID, item.ItemID)
	if err != nil {
		return err
	}

	if existingItem != nil {
		// 物品已存在，增加数量
		return db.UpdateInventoryItemQuantity(ctx, item.UserID, item.ItemID, existingItem.Quantity+item.Quantity)
	}

	// 新物品，直接插入
	propertiesJSON, err := json.Marshal(item.Properties)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO inventory (
			user_id, item_id, item_name, item_type, quality, level, quantity,
			properties, description, obtained_from, obtained_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err = db.GetPool().Exec(ctx, query,
		item.UserID, item.ItemID, item.ItemName, item.ItemType, item.Quality, item.Level, item.Quantity,
		propertiesJSON, item.Description, item.ObtainedFrom, item.ObtainedAt,
	)
	return err
}

// GetInventoryItemByID 根据物品ID获取背包物品
func (db *DB) GetInventoryItemByID(ctx context.Context, userID int, itemID int) (*InventoryItem, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT user_id, item_id, item_name, item_type, quality, level, quantity,
			properties, description, obtained_from, obtained_at
		FROM inventory
		WHERE user_id = $1 AND item_id = $2
	`

	row := db.GetPool().QueryRow(timeoutCtx, query, userID, itemID)

	var item InventoryItem
	var propertiesJSON []byte
	err := row.Scan(
		&item.UserID, &item.ItemID, &item.ItemName, &item.ItemType, &item.Quality, &item.Level, &item.Quantity,
		&propertiesJSON, &item.Description, &item.ObtainedFrom, &item.ObtainedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// 解析properties JSON
	if len(propertiesJSON) > 0 {
		err = json.Unmarshal(propertiesJSON, &item.Properties)
		if err != nil {
			return nil, err
		}
	}

	return &item, nil
}

// GetUserInventory 获取用户完整背包
func (db *DB) GetUserInventory(ctx context.Context, userID int) ([]*InventoryItem, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT user_id, item_id, item_name, item_type, quality, level, quantity,
			properties, description, obtained_from, obtained_at
		FROM inventory
		WHERE user_id = $1
		ORDER BY item_type, quality DESC, level DESC, obtained_at DESC
	`

	rows, err := db.GetPool().Query(timeoutCtx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*InventoryItem
	for rows.Next() {
		var item InventoryItem
		var propertiesJSON []byte
		err := rows.Scan(
			&item.UserID, &item.ItemID, &item.ItemName, &item.ItemType, &item.Quality, &item.Level, &item.Quantity,
			&propertiesJSON, &item.Description, &item.ObtainedFrom, &item.ObtainedAt,
		)
		if err != nil {
			return nil, err
		}

		// 解析properties JSON
		if len(propertiesJSON) > 0 {
			err = json.Unmarshal(propertiesJSON, &item.Properties)
			if err != nil {
				return nil, err
			}
		}

		items = append(items, &item)
	}

	return items, rows.Err()
}

// GetInventoryByType 根据物品类型获取背包物品
func (db *DB) GetInventoryByType(ctx context.Context, userID int, itemType string) ([]*InventoryItem, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT user_id, item_id, item_name, item_type, quality, level, quantity,
			properties, description, obtained_from, obtained_at
		FROM inventory
		WHERE user_id = $1 AND item_type = $2
		ORDER BY quality DESC, level DESC, obtained_at DESC
	`

	rows, err := db.GetPool().Query(timeoutCtx, query, userID, itemType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*InventoryItem
	for rows.Next() {
		var item InventoryItem
		var propertiesJSON []byte
		err := rows.Scan(
			&item.UserID, &item.ItemID, &item.ItemName, &item.ItemType, &item.Quality, &item.Level, &item.Quantity,
			&propertiesJSON, &item.Description, &item.ObtainedFrom, &item.ObtainedAt,
		)
		if err != nil {
			return nil, err
		}

		// 解析properties JSON
		if len(propertiesJSON) > 0 {
			err = json.Unmarshal(propertiesJSON, &item.Properties)
			if err != nil {
				return nil, err
			}
		}

		items = append(items, &item)
	}

	return items, rows.Err()
}

// UpdateInventoryItemQuantity 更新物品数量
func (db *DB) UpdateInventoryItemQuantity(ctx context.Context, userID int, itemID int, quantity int) error {
	if quantity <= 0 {
		// 数量为0或负数时，删除物品
		return db.RemoveInventoryItem(ctx, userID, itemID)
	}

	query := `
		UPDATE inventory 
		SET quantity = $3
		WHERE user_id = $1 AND item_id = $2
	`
	_, err := db.GetPool().Exec(ctx, query, userID, itemID, quantity)
	return err
}

// RemoveInventoryItem 从背包移除物品
func (db *DB) RemoveInventoryItem(ctx context.Context, userID int, itemID int) error {
	query := `
		DELETE FROM inventory
		WHERE user_id = $1 AND item_id = $2
	`
	_, err := db.GetPool().Exec(ctx, query, userID, itemID)
	return err
}

// UpdateInventoryItemProperties 更新物品属性
func (db *DB) UpdateInventoryItemProperties(ctx context.Context, userID int, itemID int, properties map[string]interface{}) error {
	propertiesJSON, err := json.Marshal(properties)
	if err != nil {
		return err
	}

	query := `
		UPDATE inventory 
		SET properties = $3
		WHERE user_id = $1 AND item_id = $2
	`
	_, err = db.GetPool().Exec(ctx, query, userID, itemID, propertiesJSON)
	return err
}

// UseItem 使用物品（减少数量）
func (db *DB) UseItem(ctx context.Context, userID int, itemID int, useQuantity int) error {
	// 先获取当前数量
	item, err := db.GetInventoryItemByID(ctx, userID, itemID)
	if err != nil {
		return err
	}
	if item == nil {
		return pgx.ErrNoRows
	}

	if item.Quantity < useQuantity {
		return &InsufficientItemError{ItemName: item.ItemName, Required: useQuantity, Available: item.Quantity}
	}

	newQuantity := item.Quantity - useQuantity
	return db.UpdateInventoryItemQuantity(ctx, userID, itemID, newQuantity)
}

// GetInventoryItemCount 获取特定物品的数量
func (db *DB) GetInventoryItemCount(ctx context.Context, userID int, itemID int) (int, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT quantity FROM inventory
		WHERE user_id = $1 AND item_id = $2
	`

	var quantity int
	err := db.GetPool().QueryRow(timeoutCtx, query, userID, itemID).Scan(&quantity)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}

	return quantity, nil
}

// GetInventoryByQuality 根据品质获取背包物品
func (db *DB) GetInventoryByQuality(ctx context.Context, userID int, quality string) ([]*InventoryItem, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT user_id, item_id, item_name, item_type, quality, level, quantity,
			properties, description, obtained_from, obtained_at
		FROM inventory
		WHERE user_id = $1 AND quality = $2
		ORDER BY item_type, level DESC, obtained_at DESC
	`

	rows, err := db.GetPool().Query(timeoutCtx, query, userID, quality)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*InventoryItem
	for rows.Next() {
		var item InventoryItem
		var propertiesJSON []byte
		err := rows.Scan(
			&item.UserID, &item.ItemID, &item.ItemName, &item.ItemType, &item.Quality, &item.Level, &item.Quantity,
			&propertiesJSON, &item.Description, &item.ObtainedFrom, &item.ObtainedAt,
		)
		if err != nil {
			return nil, err
		}

		// 解析properties JSON
		if len(propertiesJSON) > 0 {
			err = json.Unmarshal(propertiesJSON, &item.Properties)
			if err != nil {
				return nil, err
			}
		}

		items = append(items, &item)
	}

	return items, rows.Err()
}

// SearchInventoryByName 根据物品名称搜索背包物品
func (db *DB) SearchInventoryByName(ctx context.Context, userID int, itemName string) ([]*InventoryItem, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT user_id, item_id, item_name, item_type, quality, level, quantity,
			properties, description, obtained_from, obtained_at
		FROM inventory
		WHERE user_id = $1 AND item_name ILIKE '%' || $2 || '%'
		ORDER BY item_type, quality DESC, level DESC, obtained_at DESC
	`

	rows, err := db.GetPool().Query(timeoutCtx, query, userID, itemName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*InventoryItem
	for rows.Next() {
		var item InventoryItem
		var propertiesJSON []byte
		err := rows.Scan(
			&item.UserID, &item.ItemID, &item.ItemName, &item.ItemType, &item.Quality, &item.Level, &item.Quantity,
			&propertiesJSON, &item.Description, &item.ObtainedFrom, &item.ObtainedAt,
		)
		if err != nil {
			return nil, err
		}

		// 解析properties JSON
		if len(propertiesJSON) > 0 {
			err = json.Unmarshal(propertiesJSON, &item.Properties)
			if err != nil {
				return nil, err
			}
		}

		items = append(items, &item)
	}

	return items, rows.Err()
}

// InsufficientItemError 物品数量不足错误
type InsufficientItemError struct {
	ItemName  string
	Required  int
	Available int
}

func (e *InsufficientItemError) Error() string {
	return fmt.Sprintf("物品 %s 数量不足：需要 %d，可用 %d", e.ItemName, e.Required, e.Available)
}
