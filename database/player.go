package database

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// SpiritualRoots 灵根属性结构
type SpiritualRoots []SpiritualRoot

type SpiritualRoot struct {
	RootName string `json:"root_name"`
	Afinity  int    `json:"affinity"`
}

// CharacterStats 人物属性结构
type CharacterStats struct {
	UserID int    `json:"user_id"`
	Name   string `json:"name"`

	// 修炼境界
	Realm      string `json:"realm"`
	RealmLevel int    `json:"realm_level"`

	// 灵根属性 (JSONB存储)
	SpiritualRoots *SpiritualRoots `json:"spiritual_roots"`

	// 新增属性
	SpiritSense int    `json:"spirit_sense"` // 神识
	Physique    int    `json:"physique"`     // 根骨/体魄
	DemonicAura int    `json:"demonic_aura"` // 煞气/心魔
	TaoistName  string `json:"taoist_name"`  // 道号

	// 战斗属性
	Attack  int `json:"attack"`
	Defense int `json:"defense"`
	Speed   int `json:"speed"`
	Luck    int `json:"luck"`

	// 修炼相关
	Experience    int64 `json:"experience"`
	Comprehension int   `json:"comprehension"`

	// 寿命相关
	Age      int `json:"age"`
	Lifespan int `json:"lifespan"`

	// 位置信息
	Location string `json:"location"`

	// 状态
	Status string `json:"status"`

	// 成长经历
	Stories string `json:"stories"`
}

// CharacterStatsUpdate 用于部分更新的结构体，所有字段都是指针类型
type CharacterStatsUpdate struct {
	UserID int `json:"user_id,omitempty"` // 必须字段，不使用指针

	// 修炼境界
	Realm      *string `json:"realm,omitempty"`
	RealmLevel *int    `json:"realm_level,omitempty"`

	// 新增属性
	SpiritSense *int `json:"spirit_sense,omitempty"` // 神识
	Physique    *int `json:"physique,omitempty"`     // 根骨/体魄
	DemonicAura *int `json:"demonic_aura,omitempty"` // 煞气/心魔

	// 战斗属性
	Attack  *int `json:"attack,omitempty"`
	Defense *int `json:"defense,omitempty"`
	Speed   *int `json:"speed,omitempty"`
	Luck    *int `json:"luck,omitempty"`

	// 修炼相关
	Comprehension *int `json:"comprehension,omitempty"`

	// 寿命相关
	Lifespan *int `json:"lifespan,omitempty"`

	// 位置信息
	Location *string `json:"location,omitempty"`

	// 状态
	Status *string `json:"status,omitempty"`

	// 成长经历
	Stories *string `json:"stories,omitempty"`
}

func (c *CharacterStats) String() string {
	json, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return string(json)
}

// CreateCharacterStatsInTx 在事务中创建人物属性
func (db *DB) CreateCharacterStatsInTx(ctx context.Context, tx pgx.Tx, stats *CharacterStats) error {
	var spiritualRootsJSON []byte
	var err error

	if stats.SpiritualRoots != nil {
		spiritualRootsJSON, err = json.Marshal(stats.SpiritualRoots)
		if err != nil {
			return err
		}
	}

	query := `
		INSERT INTO character_stats (
			user_id, name, realm, realm_level,
			spiritual_roots, spirit_sense, physique, demonic_aura, taoist_name,
			attack, defense, speed, luck,
			experience, comprehension,
			age, lifespan, location, status, stories
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7, $8, $9,
			$10, $11, $12, $13,
			$14, $15, $16, $17,
			$18, $19,
			$20
		)
	`

	_, err = tx.Exec(ctx, query,
		stats.UserID, stats.Name, stats.Realm, stats.RealmLevel,
		spiritualRootsJSON, stats.SpiritSense, stats.Physique, stats.DemonicAura, stats.TaoistName,
		stats.Attack, stats.Defense, stats.Speed, stats.Luck,
		stats.Experience, stats.Comprehension,
		stats.Age, stats.Lifespan, stats.Location, stats.Status,
		stats.Stories,
	)
	return err
}

func (db *DB) NameExists(ctx context.Context, name string) (bool, error) {
	query := `
		SELECT EXISTS(SELECT 1 FROM character_stats WHERE name = $1)
	`
	var exists bool
	err := db.GetPool().QueryRow(ctx, query, name).Scan(&exists)
	return exists, err
}

// GetCharacterStats 获取人物属性
func (db *DB) GetCharacterStats(ctx context.Context, userID int) (*CharacterStats, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT user_id, name, realm, realm_level,
			spiritual_roots, spirit_sense, physique, demonic_aura, taoist_name,
			attack, defense, speed, luck,
			experience, comprehension,
			age, lifespan, location, status, stories
		FROM character_stats
		WHERE user_id = $1
	`

	row := db.GetPool().QueryRow(timeoutCtx, query, userID)

	var stats CharacterStats
	var spiritualRootsJSON []byte
	err := row.Scan(
		&stats.UserID, &stats.Name, &stats.Realm, &stats.RealmLevel,
		&spiritualRootsJSON, &stats.SpiritSense, &stats.Physique, &stats.DemonicAura, &stats.TaoistName,
		&stats.Attack, &stats.Defense, &stats.Speed, &stats.Luck,
		&stats.Experience, &stats.Comprehension,
		&stats.Age, &stats.Lifespan, &stats.Location, &stats.Status, &stats.Stories,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// 解析灵根JSON
	if len(spiritualRootsJSON) > 0 {
		err = json.Unmarshal(spiritualRootsJSON, &stats.SpiritualRoots)
		if err != nil {
			return nil, err
		}
	}

	return &stats, nil
}

// UpdateCharacterStats 更新人物属性
func (db *DB) UpdateCharacterStats(ctx context.Context, stats *CharacterStats) error {
	var spiritualRootsJSON []byte
	var err error

	if stats.SpiritualRoots != nil {
		spiritualRootsJSON, err = json.Marshal(stats.SpiritualRoots)
		if err != nil {
			return err
		}
	}

	query := `
		UPDATE character_stats SET
			name = $2, realm = $3, realm_level = $4,
			spiritual_roots = $5, spirit_sense = $6, physique = $7, demonic_aura = $8, taoist_name = $9,
			attack = $10, defense = $11, speed = $12, luck = $13,
			experience = $14, comprehension = $15,
			age = $16, lifespan = $17, location = $18, status = $19, stories = $20
		WHERE user_id = $1
	`

	_, err = db.GetPool().Exec(ctx, query,
		stats.UserID, stats.Name, stats.Realm, stats.RealmLevel,
		spiritualRootsJSON, stats.SpiritSense, stats.Physique, stats.DemonicAura, stats.TaoistName,
		stats.Attack, stats.Defense, stats.Speed, stats.Luck,
		stats.Experience, stats.Comprehension,
		stats.Age, stats.Lifespan, stats.Location, stats.Status, stats.Stories,
	)
	return err
}

// UpdateCharacterStatsPartial 部分更新人物属性，只更新非nil的字段
func (db *DB) UpdateCharacterStatsPartial(ctx context.Context, update *CharacterStatsUpdate) error {
	setParts := []string{}
	args := []interface{}{update.UserID}
	argIndex := 2

	// 动态构建SET子句
	if update.Realm != nil {
		setParts = append(setParts, fmt.Sprintf("realm = $%d", argIndex))
		args = append(args, *update.Realm)
		argIndex++
	}
	if update.RealmLevel != nil {
		setParts = append(setParts, fmt.Sprintf("realm_level = $%d", argIndex))
		args = append(args, *update.RealmLevel)
		argIndex++
	}
	if update.SpiritSense != nil {
		setParts = append(setParts, fmt.Sprintf("spirit_sense = $%d", argIndex))
		args = append(args, *update.SpiritSense)
		argIndex++
	}
	if update.Physique != nil {
		setParts = append(setParts, fmt.Sprintf("physique = $%d", argIndex))
		args = append(args, *update.Physique)
		argIndex++
	}
	if update.DemonicAura != nil {
		setParts = append(setParts, fmt.Sprintf("demonic_aura = $%d", argIndex))
		args = append(args, *update.DemonicAura)
		argIndex++
	}
	if update.Attack != nil {
		setParts = append(setParts, fmt.Sprintf("attack = $%d", argIndex))
		args = append(args, *update.Attack)
		argIndex++
	}
	if update.Defense != nil {
		setParts = append(setParts, fmt.Sprintf("defense = $%d", argIndex))
		args = append(args, *update.Defense)
		argIndex++
	}
	if update.Speed != nil {
		setParts = append(setParts, fmt.Sprintf("speed = $%d", argIndex))
		args = append(args, *update.Speed)
		argIndex++
	}
	if update.Luck != nil {
		setParts = append(setParts, fmt.Sprintf("luck = $%d", argIndex))
		args = append(args, *update.Luck)
		argIndex++
	}
	if update.Comprehension != nil {
		setParts = append(setParts, fmt.Sprintf("comprehension = $%d", argIndex))
		args = append(args, *update.Comprehension)
		argIndex++
	}
	if update.Lifespan != nil {
		setParts = append(setParts, fmt.Sprintf("lifespan = $%d", argIndex))
		args = append(args, *update.Lifespan)
		argIndex++
	}
	if update.Location != nil {
		setParts = append(setParts, fmt.Sprintf("location = $%d", argIndex))
		args = append(args, *update.Location)
		argIndex++
	}
	if update.Status != nil {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *update.Status)
		argIndex++
	}
	if update.Stories != nil {
		setParts = append(setParts, fmt.Sprintf("stories = CASE WHEN stories IS NULL OR stories = '' THEN $%d ELSE stories || '\n' || $%d END", argIndex, argIndex))
		args = append(args, *update.Stories)
		argIndex++
	}

	// 如果没有要更新的字段，直接返回
	if len(setParts) == 0 {
		return nil
	}

	// 构建完整的UPDATE语句
	query := fmt.Sprintf(`
		UPDATE character_stats SET %s
		WHERE user_id = $1
	`, strings.Join(setParts, ", "))

	_, err := db.GetPool().Exec(ctx, query, args...)
	return err
}

// UpdateCharacterRealm 提升境界
func (db *DB) UpdateCharacterRealm(ctx context.Context, userID int, realm string, realmLevel int) error {
	query := `
		UPDATE character_stats 
		SET realm = $2, realm_level = $3
		WHERE user_id = $1
	`
	_, err := db.GetPool().Exec(ctx, query, userID, realm, realmLevel)
	return err
}

// UpdateSpiritualRoots 更新灵根属性
func (db *DB) UpdateSpiritualRoots(ctx context.Context, userID int, spiritualRoots *SpiritualRoots) error {
	spiritualRootsJSON, err := json.Marshal(spiritualRoots)
	if err != nil {
		return err
	}

	query := `
		UPDATE character_stats 
		SET spiritual_roots = $2
		WHERE user_id = $1
	`
	_, err = db.GetPool().Exec(ctx, query, userID, spiritualRootsJSON)
	return err
}

// UpdateSpiritSense 更新神识
func (db *DB) UpdateSpiritSense(ctx context.Context, userID int, spiritSense int) error {
	query := `
		UPDATE character_stats 
		SET spirit_sense = $2
		WHERE user_id = $1
	`
	_, err := db.GetPool().Exec(ctx, query, userID, spiritSense)
	return err
}

// UpdatePhysique 更新根骨/体魄
func (db *DB) UpdatePhysique(ctx context.Context, userID int, physique int) error {
	query := `
		UPDATE character_stats 
		SET physique = $2
		WHERE user_id = $1
	`
	_, err := db.GetPool().Exec(ctx, query, userID, physique)
	return err
}

// UpdateDemonicAura 更新煞气/心魔
func (db *DB) UpdateDemonicAura(ctx context.Context, userID int, demonicAura int) error {
	query := `
		UPDATE character_stats 
		SET demonic_aura = $2
		WHERE user_id = $1
	`
	_, err := db.GetPool().Exec(ctx, query, userID, demonicAura)
	return err
}

// SetTaoistName 设置道号
func (db *DB) SetTaoistName(ctx context.Context, userID int, taoistName string) error {
	query := `
		UPDATE character_stats 
		SET taoist_name = $2
		WHERE user_id = $1
	`
	_, err := db.GetPool().Exec(ctx, query, userID, taoistName)
	return err
}

// AddExperience 增加修炼经验
func (db *DB) AddExperience(ctx context.Context, userID int, exp int64) error {
	query := `
		UPDATE character_stats 
		SET experience = experience + $2
		WHERE user_id = $1
	`
	_, err := db.GetPool().Exec(ctx, query, userID, exp)
	return err
}

// UpdateCharacterLocation 更新位置
func (db *DB) UpdateCharacterLocation(ctx context.Context, userID int, location string) error {
	query := `
		UPDATE character_stats 
		SET location = $2
		WHERE user_id = $1
	`
	_, err := db.GetPool().Exec(ctx, query, userID, location)
	return err
}

// UpdateCharacterStatus 更新状态
func (db *DB) UpdateCharacterStatus(ctx context.Context, userID int, status string) error {
	query := `
		UPDATE character_stats 
		SET status = $2
		WHERE user_id = $1
	`
	_, err := db.GetPool().Exec(ctx, query, userID, status)
	return err
}

// GetCharactersByRealm 根据境界查询人物
func (db *DB) GetCharactersByRealm(ctx context.Context, realm string) ([]*CharacterStats, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT user_id, name, realm, realm_level,
			spiritual_roots, spirit_sense, physique, demonic_aura, taoist_name,
			attack, defense, speed, luck,
			experience, comprehension,
			age, lifespan, location, status, stories
		FROM character_stats
		WHERE realm = $1
		ORDER BY realm_level DESC, experience DESC
	`

	rows, err := db.GetPool().Query(timeoutCtx, query, realm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var characters []*CharacterStats
	for rows.Next() {
		var stats CharacterStats
		var spiritualRootsJSON []byte
		err := rows.Scan(
			&stats.UserID, &stats.Name, &stats.Realm, &stats.RealmLevel,
			&spiritualRootsJSON, &stats.SpiritSense, &stats.Physique, &stats.DemonicAura, &stats.TaoistName,
			&stats.Attack, &stats.Defense, &stats.Speed, &stats.Luck,
			&stats.Experience, &stats.Comprehension,
			&stats.Age, &stats.Lifespan, &stats.Location, &stats.Status, &stats.Stories,
		)
		if err != nil {
			return nil, err
		}

		// 解析灵根JSON
		if len(spiritualRootsJSON) > 0 {
			err = json.Unmarshal(spiritualRootsJSON, &stats.SpiritualRoots)
			if err != nil {
				return nil, err
			}
		}

		characters = append(characters, &stats)
	}

	return characters, rows.Err()
}

// GetCharactersByLocation 根据位置查询人物
func (db *DB) GetCharactersByLocation(ctx context.Context, location string) ([]*CharacterStats, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT user_id, name, realm, realm_level,
			spiritual_roots, spirit_sense, physique, demonic_aura, taoist_name,
			attack, defense, speed, luck,
			experience, comprehension,
			age, lifespan, location, status, stories
		FROM character_stats
		WHERE location = $1
		ORDER BY realm_level DESC, experience DESC
	`

	rows, err := db.GetPool().Query(timeoutCtx, query, location)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var characters []*CharacterStats
	for rows.Next() {
		var stats CharacterStats
		var spiritualRootsJSON []byte
		err := rows.Scan(
			&stats.UserID, &stats.Name, &stats.Realm, &stats.RealmLevel,
			&spiritualRootsJSON, &stats.SpiritSense, &stats.Physique, &stats.DemonicAura, &stats.TaoistName,
			&stats.Attack, &stats.Defense, &stats.Speed, &stats.Luck,
			&stats.Experience, &stats.Comprehension,
			&stats.Age, &stats.Lifespan, &stats.Location, &stats.Status, &stats.Stories,
		)
		if err != nil {
			return nil, err
		}

		// 解析灵根JSON
		if len(spiritualRootsJSON) > 0 {
			err = json.Unmarshal(spiritualRootsJSON, &stats.SpiritualRoots)
			if err != nil {
				return nil, err
			}
		}

		characters = append(characters, &stats)
	}

	return characters, rows.Err()
}

// GetCharactersByTaoistName 根据道号查询人物
func (db *DB) GetCharactersByTaoistName(ctx context.Context, taoistName string) ([]*CharacterStats, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT user_id, name, realm, realm_level,
			spiritual_roots, spirit_sense, physique, demonic_aura, taoist_name,
			attack, defense, speed, luck,
			experience, comprehension,
			age, lifespan, location, status, stories
		FROM character_stats
		WHERE taoist_name ILIKE '%' || $1 || '%'
		ORDER BY realm_level DESC, experience DESC
	`

	rows, err := db.GetPool().Query(timeoutCtx, query, taoistName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var characters []*CharacterStats
	for rows.Next() {
		var stats CharacterStats
		var spiritualRootsJSON []byte
		err := rows.Scan(
			&stats.UserID, &stats.Name, &stats.Realm, &stats.RealmLevel,
			&spiritualRootsJSON, &stats.SpiritSense, &stats.Physique, &stats.DemonicAura, &stats.TaoistName,
			&stats.Attack, &stats.Defense, &stats.Speed, &stats.Luck,
			&stats.Experience, &stats.Comprehension,
			&stats.Age, &stats.Lifespan, &stats.Location, &stats.Status, &stats.Stories,
		)
		if err != nil {
			return nil, err
		}

		// 解析灵根JSON
		if len(spiritualRootsJSON) > 0 {
			err = json.Unmarshal(spiritualRootsJSON, &stats.SpiritualRoots)
			if err != nil {
				return nil, err
			}
		}

		characters = append(characters, &stats)
	}

	return characters, rows.Err()
}

// AddDemonicAura 增加煞气/心魔
func (db *DB) AddDemonicAura(ctx context.Context, userID int, amount int) error {
	query := `
		UPDATE character_stats 
		SET demonic_aura = GREATEST(0, demonic_aura + $2)
		WHERE user_id = $1
	`
	_, err := db.GetPool().Exec(ctx, query, userID, amount)
	return err
}

// AddSpiritSense 增加神识
func (db *DB) AddSpiritSense(ctx context.Context, userID int, amount int) error {
	query := `
		UPDATE character_stats 
		SET spirit_sense = spirit_sense + $2
		WHERE user_id = $1
	`
	_, err := db.GetPool().Exec(ctx, query, userID, amount)
	return err
}

// GetCharactersBySpiritualRoot 根据特定灵根查询人物
func (db *DB) GetCharactersBySpiritualRoot(ctx context.Context, rootType string, minValue int) ([]*CharacterStats, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT user_id, name, realm, realm_level,
			spiritual_roots, spirit_sense, physique, demonic_aura, taoist_name,
			attack, defense, speed, luck,
			experience, comprehension,
			age, lifespan, location, status, stories
		FROM character_stats
		WHERE spiritual_roots ->> $1 IS NOT NULL 
		AND CAST(spiritual_roots ->> $1 AS INTEGER) >= $2
		ORDER BY CAST(spiritual_roots ->> $1 AS INTEGER) DESC, realm_level DESC
	`

	rows, err := db.GetPool().Query(timeoutCtx, query, rootType, minValue)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var characters []*CharacterStats
	for rows.Next() {
		var stats CharacterStats
		var spiritualRootsJSON []byte
		err := rows.Scan(
			&stats.UserID, &stats.Name, &stats.Realm, &stats.RealmLevel,
			&spiritualRootsJSON, &stats.SpiritSense, &stats.Physique, &stats.DemonicAura, &stats.TaoistName,
			&stats.Attack, &stats.Defense, &stats.Speed, &stats.Luck,
			&stats.Experience, &stats.Comprehension,
			&stats.Age, &stats.Lifespan, &stats.Location, &stats.Status, &stats.Stories,
		)
		if err != nil {
			return nil, err
		}

		// 解析灵根JSON
		if len(spiritualRootsJSON) > 0 {
			err = json.Unmarshal(spiritualRootsJSON, &stats.SpiritualRoots)
			if err != nil {
				return nil, err
			}
		}

		characters = append(characters, &stats)
	}

	return characters, rows.Err()
}
