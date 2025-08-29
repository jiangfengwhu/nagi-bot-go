package database

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
)

// SpiritualRoots 灵根属性结构
type SpiritualRoots map[string]int

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
	SpiritSense int     `json:"spirit_sense"` // 神识
	Physique    int     `json:"physique"`     // 根骨/体魄
	DemonicAura int     `json:"demonic_aura"` // 煞气/心魔
	TaoistName  *string `json:"taoist_name"`  // 道号

	// 基础属性
	HP    int `json:"hp"`
	MaxHP int `json:"max_hp"`
	MP    int `json:"mp"`
	MaxMP int `json:"max_mp"`

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
}

func (c *CharacterStats) String() string {
	json, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return string(json)
}

// CreateCharacterStats 创建人物属性
func (db *DB) CreateCharacterStats(ctx context.Context, stats *CharacterStats) error {
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
			hp, max_hp, mp, max_mp,
			attack, defense, speed, luck,
			experience, comprehension,
			age, lifespan, location, status
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7, $8, $9,
			$10, $11, $12, $13,
			$14, $15, $16, $17,
			$18, $19,
			$20, $21, $22, $23
		)
	`

	_, err = db.GetPool().Exec(ctx, query,
		stats.UserID, stats.Name, stats.Realm, stats.RealmLevel,
		spiritualRootsJSON, stats.SpiritSense, stats.Physique, stats.DemonicAura, stats.TaoistName,
		stats.HP, stats.MaxHP, stats.MP, stats.MaxMP,
		stats.Attack, stats.Defense, stats.Speed, stats.Luck,
		stats.Experience, stats.Comprehension,
		stats.Age, stats.Lifespan, stats.Location, stats.Status,
	)
	return err
}

// GetCharacterStats 获取人物属性
func (db *DB) GetCharacterStats(ctx context.Context, userID int) (*CharacterStats, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT user_id, name, realm, realm_level,
			spiritual_roots, spirit_sense, physique, demonic_aura, taoist_name,
			hp, max_hp, mp, max_mp,
			attack, defense, speed, luck,
			experience, comprehension,
			age, lifespan, location, status
		FROM character_stats
		WHERE user_id = $1
	`

	row := db.GetPool().QueryRow(timeoutCtx, query, userID)

	var stats CharacterStats
	var spiritualRootsJSON []byte
	err := row.Scan(
		&stats.UserID, &stats.Name, &stats.Realm, &stats.RealmLevel,
		&spiritualRootsJSON, &stats.SpiritSense, &stats.Physique, &stats.DemonicAura, &stats.TaoistName,
		&stats.HP, &stats.MaxHP, &stats.MP, &stats.MaxMP,
		&stats.Attack, &stats.Defense, &stats.Speed, &stats.Luck,
		&stats.Experience, &stats.Comprehension,
		&stats.Age, &stats.Lifespan, &stats.Location, &stats.Status,
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
			hp = $10, max_hp = $11, mp = $12, max_mp = $13,
			attack = $14, defense = $15, speed = $16, luck = $17,
			experience = $18, comprehension = $19,
			age = $20, lifespan = $21, location = $22, status = $23
		WHERE user_id = $1
	`

	_, err = db.GetPool().Exec(ctx, query,
		stats.UserID, stats.Name, stats.Realm, stats.RealmLevel,
		spiritualRootsJSON, stats.SpiritSense, stats.Physique, stats.DemonicAura, stats.TaoistName,
		stats.HP, stats.MaxHP, stats.MP, stats.MaxMP,
		stats.Attack, stats.Defense, stats.Speed, stats.Luck,
		stats.Experience, stats.Comprehension,
		stats.Age, stats.Lifespan, stats.Location, stats.Status,
	)
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

// UpdateCharacterHP 更新生命值
func (db *DB) UpdateCharacterHP(ctx context.Context, userID int, hp int) error {
	query := `
		UPDATE character_stats 
		SET hp = GREATEST(0, LEAST($2, max_hp))
		WHERE user_id = $1
	`
	_, err := db.GetPool().Exec(ctx, query, userID, hp)
	return err
}

// UpdateCharacterMP 更新法力值
func (db *DB) UpdateCharacterMP(ctx context.Context, userID int, mp int) error {
	query := `
		UPDATE character_stats 
		SET mp = GREATEST(0, LEAST($2, max_mp))
		WHERE user_id = $1
	`
	_, err := db.GetPool().Exec(ctx, query, userID, mp)
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
			hp, max_hp, mp, max_mp,
			attack, defense, speed, luck,
			experience, comprehension,
			age, lifespan, location, status
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
			&stats.HP, &stats.MaxHP, &stats.MP, &stats.MaxMP,
			&stats.Attack, &stats.Defense, &stats.Speed, &stats.Luck,
			&stats.Experience, &stats.Comprehension,
			&stats.Age, &stats.Lifespan, &stats.Location, &stats.Status,
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
			hp, max_hp, mp, max_mp,
			attack, defense, speed, luck,
			experience, comprehension,
			age, lifespan, location, status
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
			&stats.HP, &stats.MaxHP, &stats.MP, &stats.MaxMP,
			&stats.Attack, &stats.Defense, &stats.Speed, &stats.Luck,
			&stats.Experience, &stats.Comprehension,
			&stats.Age, &stats.Lifespan, &stats.Location, &stats.Status,
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
			hp, max_hp, mp, max_mp,
			attack, defense, speed, luck,
			experience, comprehension,
			age, lifespan, location, status
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
			&stats.HP, &stats.MaxHP, &stats.MP, &stats.MaxMP,
			&stats.Attack, &stats.Defense, &stats.Speed, &stats.Luck,
			&stats.Experience, &stats.Comprehension,
			&stats.Age, &stats.Lifespan, &stats.Location, &stats.Status,
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
			hp, max_hp, mp, max_mp,
			attack, defense, speed, luck,
			experience, comprehension,
			age, lifespan, location, status
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
			&stats.HP, &stats.MaxHP, &stats.MP, &stats.MaxMP,
			&stats.Attack, &stats.Defense, &stats.Speed, &stats.Luck,
			&stats.Experience, &stats.Comprehension,
			&stats.Age, &stats.Lifespan, &stats.Location, &stats.Status,
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
