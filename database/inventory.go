package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// InventoryItem 背包物品结构
type InventoryItem struct {
	UserID       int       `json:"user_id,omitempty"`
	ItemName     string    `json:"item_name"`
	ItemType     string    `json:"item_type"`
	Quality      string    `json:"quality"`
	Level        int       `json:"level"`
	Quantity     int       `json:"quantity"`
	Properties   string    `json:"properties"`
	Description  string    `json:"description"`
	ObtainedFrom string    `json:"obtained_from"`
	ObtainedAt   time.Time `json:"obtained_at,omitzero"`
}

func (db *DB) UpdateInventory(ctx context.Context, userID int, inventoryItems []*InventoryItem) error {
	tx, err := db.GetPool().Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, item := range inventoryItems {
		item.UserID = userID
		err = db.AddInventoryItemsBatch(ctx, []*InventoryItem{item})
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// AddInventoryItemsBatch 批量添加物品到背包
func (db *DB) AddInventoryItemsBatch(ctx context.Context, items []*InventoryItem) error {
	if len(items) == 0 {
		return nil
	}

	// 开始事务
	tx, err := db.GetPool().Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 为了提高效率，先批量查询现有物品
	userItemMap := make(map[int]map[string]*InventoryItem)

	// 按用户ID分组
	userItems := make(map[int][]*InventoryItem)
	for _, item := range items {
		userItems[item.UserID] = append(userItems[item.UserID], item)
	}

	// 为每个用户查询现有物品
	for userID, userItemList := range userItems {
		itemNames := make([]string, len(userItemList))
		for i, item := range userItemList {
			itemNames[i] = item.ItemName
		}

		// 批量查询现有物品
		existingItems, err := db.getInventoryItemsByNamesInTx(ctx, tx, userID, itemNames)
		if err != nil {
			return err
		}

		// 建立用户物品映射
		if userItemMap[userID] == nil {
			userItemMap[userID] = make(map[string]*InventoryItem)
		}
		for _, existingItem := range existingItems {
			userItemMap[userID][existingItem.ItemName] = existingItem
		}
	}

	// 准备批量插入、更新和删除的数据
	var itemsToInsert []*InventoryItem
	var itemsToUpdate []*InventoryItem
	var itemsToDelete []*InventoryItem

	for _, item := range items {
		if existingItem, exists := userItemMap[item.UserID][item.ItemName]; exists {
			// 物品已存在，准备更新数量
			existingItem.Quantity += item.Quantity
			itemsToUpdate = append(itemsToUpdate, existingItem)
		} else {
			// 新物品处理
			if item.Quantity < 0 {
				// 如果是负数量的新物品，添加到删除列表（删除可能存在的记录）
				itemsToDelete = append(itemsToDelete, item)
			} else {
				// 正数量的新物品，准备插入
				// 如果ObtainedAt为零值，使用当前时间
				if item.ObtainedAt.IsZero() {
					item.ObtainedAt = time.Now()
				}
				itemsToInsert = append(itemsToInsert, item)
			}
		}
	}

	// 批量插入新物品
	if len(itemsToInsert) > 0 {
		err = db.batchInsertInventoryItems(ctx, tx, itemsToInsert)
		if err != nil {
			return err
		}
	}

	// 批量更新现有物品数量
	if len(itemsToUpdate) > 0 {
		err = db.batchUpdateInventoryItemQuantity(ctx, tx, itemsToUpdate)
		if err != nil {
			return err
		}
	}

	// 批量删除物品
	if len(itemsToDelete) > 0 {
		err = db.batchDeleteInventoryItems(ctx, tx, itemsToDelete)
		if err != nil {
			return err
		}
	}

	// 提交事务
	return tx.Commit(ctx)
}

// getInventoryItemsByNamesInTx 在事务中批量查询物品
func (db *DB) getInventoryItemsByNamesInTx(ctx context.Context, tx pgx.Tx, userID int, itemNames []string) ([]*InventoryItem, error) {
	if len(itemNames) == 0 {
		return nil, nil
	}

	query := `
		SELECT user_id, item_name, item_type, quality, level, quantity,
			properties, description, obtained_from, obtained_at
		FROM inventory
		WHERE user_id = $1 AND item_name = ANY($2)
	`

	rows, err := tx.Query(ctx, query, userID, itemNames)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*InventoryItem
	for rows.Next() {
		var item InventoryItem
		err := rows.Scan(
			&item.UserID, &item.ItemName, &item.ItemType, &item.Quality, &item.Level, &item.Quantity,
			&item.Properties, &item.Description, &item.ObtainedFrom, &item.ObtainedAt,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, &item)
	}

	return items, rows.Err()
}

// batchInsertInventoryItems 批量插入物品
func (db *DB) batchInsertInventoryItems(ctx context.Context, tx pgx.Tx, items []*InventoryItem) error {
	if len(items) == 0 {
		return nil
	}

	query := `
		INSERT INTO inventory (
			user_id, item_name, item_type, quality, level, quantity,
			properties, description, obtained_from, obtained_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	batch := &pgx.Batch{}
	for _, item := range items {
		batch.Queue(query,
			item.UserID, item.ItemName, item.ItemType, item.Quality, item.Level, item.Quantity,
			item.Properties, item.Description, item.ObtainedFrom, item.ObtainedAt,
		)
	}

	results := tx.SendBatch(ctx, batch)
	defer results.Close()

	// 处理所有批量插入结果
	for range items {
		_, err := results.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

// batchDeleteInventoryItems 批量删除物品
func (db *DB) batchDeleteInventoryItems(ctx context.Context, tx pgx.Tx, items []*InventoryItem) error {
	if len(items) == 0 {
		return nil
	}

	query := `DELETE FROM inventory WHERE user_id = $1 AND item_name = $2`

	batch := &pgx.Batch{}
	for _, item := range items {
		batch.Queue(query, item.UserID, item.ItemName)
	}

	results := tx.SendBatch(ctx, batch)
	defer results.Close()

	// 处理所有批量删除结果
	for range items {
		_, err := results.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

// batchUpdateInventoryItemQuantity 批量更新物品数量
func (db *DB) batchUpdateInventoryItemQuantity(ctx context.Context, tx pgx.Tx, items []*InventoryItem) error {
	if len(items) == 0 {
		return nil
	}

	query := `
		UPDATE inventory 
		SET quantity = $3
		WHERE user_id = $1 AND item_name = $2
	`

	batch := &pgx.Batch{}
	for _, item := range items {
		if item.Quantity <= 0 {
			// 数量为0或负数时，删除物品
			deleteQuery := `DELETE FROM inventory WHERE user_id = $1 AND item_name = $2`
			batch.Queue(deleteQuery, item.UserID, item.ItemName)
		} else {
			batch.Queue(query, item.UserID, item.ItemName, item.Quantity)
		}
	}

	results := tx.SendBatch(ctx, batch)
	defer results.Close()

	// 处理所有批量更新结果
	for range items {
		_, err := results.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

// GetInventoryItemByID 根据物品ID获取背包物品
func (db *DB) GetInventoryItemByName(ctx context.Context, userID int, itemName string) (*InventoryItem, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT user_id, item_name, item_type, quality, level, quantity,
			properties, description, obtained_from, obtained_at
		FROM inventory
		WHERE user_id = $1 AND item_name = $2
	`

	row := db.GetPool().QueryRow(timeoutCtx, query, userID, itemName)

	var item InventoryItem
	err := row.Scan(
		&item.UserID, &item.ItemName, &item.ItemType, &item.Quality, &item.Level, &item.Quantity,
		&item.Properties, &item.Description, &item.ObtainedFrom, &item.ObtainedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &item, nil
}

// GetUserInventory 获取用户完整背包
func (db *DB) GetUserInventory(ctx context.Context, userID int) ([]*InventoryItem, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT user_id, item_name, item_type, quality, level, quantity,
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
		err := rows.Scan(
			&item.UserID, &item.ItemName, &item.ItemType, &item.Quality, &item.Level, &item.Quantity,
			&item.Properties, &item.Description, &item.ObtainedFrom, &item.ObtainedAt,
		)
		if err != nil {
			return nil, err
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
		SELECT user_id, item_name, item_type, quality, level, quantity,
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
		err := rows.Scan(
			&item.UserID, &item.ItemName, &item.ItemType, &item.Quality, &item.Level, &item.Quantity,
			&item.Properties, &item.Description, &item.ObtainedFrom, &item.ObtainedAt,
		)
		if err != nil {
			return nil, err
		}

		items = append(items, &item)
	}

	return items, rows.Err()
}

// UpdateInventoryItemQuantity 更新物品数量
func (db *DB) UpdateInventoryItemQuantity(ctx context.Context, userID int, itemName string, quantity int) error {
	if quantity <= 0 {
		// 数量为0或负数时，删除物品
		return db.RemoveInventoryItem(ctx, userID, itemName)
	}

	query := `
		UPDATE inventory 
		SET quantity = $3
		WHERE user_id = $1 AND item_name = $2
	`
	_, err := db.GetPool().Exec(ctx, query, userID, itemName, quantity)
	return err
}

// RemoveInventoryItem 从背包移除物品
func (db *DB) RemoveInventoryItem(ctx context.Context, userID int, itemName string) error {
	query := `
		DELETE FROM inventory
		WHERE user_id = $1 AND item_name = $2
	`
	_, err := db.GetPool().Exec(ctx, query, userID, itemName)
	return err
}

// UpdateInventoryItemProperties 更新物品属性
func (db *DB) UpdateInventoryItemProperties(ctx context.Context, userID int, itemName string, properties string) error {
	query := `
		UPDATE inventory 
		SET properties = $3
		WHERE user_id = $1 AND item_name = $2
	`
	_, err := db.GetPool().Exec(ctx, query, userID, itemName, properties)
	return err
}

// UseItem 使用物品（减少数量）
func (db *DB) UseItem(ctx context.Context, userID int, itemName string, useQuantity int) error {
	// 先获取当前数量
	item, err := db.GetInventoryItemByName(ctx, userID, itemName)
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
	return db.UpdateInventoryItemQuantity(ctx, userID, item.ItemName, newQuantity)
}

// GetInventoryItemCount 获取特定物品的数量
func (db *DB) GetInventoryItemCount(ctx context.Context, userID int, itemName string) (int, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT quantity FROM inventory
		WHERE user_id = $1 AND item_name = $2
	`

	var quantity int
	err := db.GetPool().QueryRow(timeoutCtx, query, userID, itemName).Scan(&quantity)
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
		SELECT user_id, item_name, item_type, quality, level, quantity,
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
		err := rows.Scan(
			&item.UserID, &item.ItemName, &item.ItemType, &item.Quality, &item.Level, &item.Quantity,
			&item.Properties, &item.Description, &item.ObtainedFrom, &item.ObtainedAt,
		)
		if err != nil {
			return nil, err
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
		SELECT user_id, item_name, item_type, quality, level, quantity,
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
		err := rows.Scan(
			&item.UserID, &item.ItemName, &item.ItemType, &item.Quality, &item.Level, &item.Quantity,
			&item.Properties, &item.Description, &item.ObtainedFrom, &item.ObtainedAt,
		)
		if err != nil {
			return nil, err
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
