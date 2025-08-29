package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// CultivationTechnique 修炼功法结构
type CultivationTechnique struct {
	UserID         int                    `json:"user_id"`
	TechniqueName  string                 `json:"technique_name"`
	TechniqueType  string                 `json:"technique_type"`
	TechniqueLevel int                    `json:"technique_level"`
	Quality        string                 `json:"quality"`
	Progress       int                    `json:"progress"`
	Effects        map[string]interface{} `json:"effects"`
	Requirements   map[string]interface{} `json:"requirements"`
	LearnedAt      time.Time              `json:"learned_at"`
}

// LearnCultivationTechnique 学习新功法
func (db *DB) LearnCultivationTechnique(ctx context.Context, technique *CultivationTechnique) error {
	// 检查是否已经学习过这个功法
	existing, err := db.GetCultivationTechnique(ctx, technique.UserID, technique.TechniqueName)
	if err != nil {
		return err
	}

	if existing != nil {
		return &TechniqueAlreadyLearnedError{TechniqueName: technique.TechniqueName}
	}

	effectsJSON, err := json.Marshal(technique.Effects)
	if err != nil {
		return err
	}

	requirementsJSON, err := json.Marshal(technique.Requirements)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO cultivation_techniques (
			user_id, technique_name, technique_type, technique_level, quality,
			progress, effects, requirements, learned_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = db.GetPool().Exec(ctx, query,
		technique.UserID, technique.TechniqueName, technique.TechniqueType, technique.TechniqueLevel, technique.Quality,
		technique.Progress, effectsJSON, requirementsJSON, technique.LearnedAt,
	)
	return err
}

// GetCultivationTechnique 获取特定功法
func (db *DB) GetCultivationTechnique(ctx context.Context, userID int, techniqueName string) (*CultivationTechnique, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT user_id, technique_name, technique_type, technique_level, quality,
			progress, effects, requirements, learned_at
		FROM cultivation_techniques
		WHERE user_id = $1 AND technique_name = $2
	`

	row := db.GetPool().QueryRow(timeoutCtx, query, userID, techniqueName)

	var technique CultivationTechnique
	var effectsJSON, requirementsJSON []byte
	err := row.Scan(
		&technique.UserID, &technique.TechniqueName, &technique.TechniqueType, &technique.TechniqueLevel, &technique.Quality,
		&technique.Progress, &effectsJSON, &requirementsJSON, &technique.LearnedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// 解析JSON字段
	if len(effectsJSON) > 0 {
		err = json.Unmarshal(effectsJSON, &technique.Effects)
		if err != nil {
			return nil, err
		}
	}

	if len(requirementsJSON) > 0 {
		err = json.Unmarshal(requirementsJSON, &technique.Requirements)
		if err != nil {
			return nil, err
		}
	}

	return &technique, nil
}

// GetUserCultivationTechniques 获取用户所有功法
func (db *DB) GetUserCultivationTechniques(ctx context.Context, userID int) ([]*CultivationTechnique, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT user_id, technique_name, technique_type, technique_level, quality,
			progress, effects, requirements, learned_at
		FROM cultivation_techniques
		WHERE user_id = $1
		ORDER BY technique_type, quality DESC, technique_level DESC, learned_at DESC
	`

	rows, err := db.GetPool().Query(timeoutCtx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var techniques []*CultivationTechnique
	for rows.Next() {
		var technique CultivationTechnique
		var effectsJSON, requirementsJSON []byte
		err := rows.Scan(
			&technique.UserID, &technique.TechniqueName, &technique.TechniqueType, &technique.TechniqueLevel, &technique.Quality,
			&technique.Progress, &effectsJSON, &requirementsJSON, &technique.LearnedAt,
		)
		if err != nil {
			return nil, err
		}

		// 解析JSON字段
		if len(effectsJSON) > 0 {
			err = json.Unmarshal(effectsJSON, &technique.Effects)
			if err != nil {
				return nil, err
			}
		}

		if len(requirementsJSON) > 0 {
			err = json.Unmarshal(requirementsJSON, &technique.Requirements)
			if err != nil {
				return nil, err
			}
		}

		techniques = append(techniques, &technique)
	}

	return techniques, rows.Err()
}

// GetCultivationTechniquesByType 根据功法类型获取功法
func (db *DB) GetCultivationTechniquesByType(ctx context.Context, userID int, techniqueType string) ([]*CultivationTechnique, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT user_id, technique_name, technique_type, technique_level, quality,
			progress, effects, requirements, learned_at
		FROM cultivation_techniques
		WHERE user_id = $1 AND technique_type = $2
		ORDER BY quality DESC, technique_level DESC, learned_at DESC
	`

	rows, err := db.GetPool().Query(timeoutCtx, query, userID, techniqueType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var techniques []*CultivationTechnique
	for rows.Next() {
		var technique CultivationTechnique
		var effectsJSON, requirementsJSON []byte
		err := rows.Scan(
			&technique.UserID, &technique.TechniqueName, &technique.TechniqueType, &technique.TechniqueLevel, &technique.Quality,
			&technique.Progress, &effectsJSON, &requirementsJSON, &technique.LearnedAt,
		)
		if err != nil {
			return nil, err
		}

		// 解析JSON字段
		if len(effectsJSON) > 0 {
			err = json.Unmarshal(effectsJSON, &technique.Effects)
			if err != nil {
				return nil, err
			}
		}

		if len(requirementsJSON) > 0 {
			err = json.Unmarshal(requirementsJSON, &technique.Requirements)
			if err != nil {
				return nil, err
			}
		}

		techniques = append(techniques, &technique)
	}

	return techniques, rows.Err()
}

// UpdateCultivationTechniqueProgress 更新功法修炼进度
func (db *DB) UpdateCultivationTechniqueProgress(ctx context.Context, userID int, techniqueName string, progress int) error {
	// 确保进度在0-100范围内
	if progress < 0 {
		progress = 0
	} else if progress > 100 {
		progress = 100
	}

	query := `
		UPDATE cultivation_techniques 
		SET progress = $3
		WHERE user_id = $1 AND technique_name = $2
	`
	_, err := db.GetPool().Exec(ctx, query, userID, techniqueName, progress)
	return err
}

// UpgradeCultivationTechnique 升级功法
func (db *DB) UpgradeCultivationTechnique(ctx context.Context, userID int, techniqueName string) error {
	// 先检查当前进度是否达到100%
	technique, err := db.GetCultivationTechnique(ctx, userID, techniqueName)
	if err != nil {
		return err
	}
	if technique == nil {
		return &TechniqueNotFoundError{TechniqueName: techniqueName}
	}

	if technique.Progress < 100 {
		return &TechniqueNotReadyForUpgradeError{TechniqueName: techniqueName, CurrentProgress: technique.Progress}
	}

	// 升级：等级+1，进度重置为0
	query := `
		UPDATE cultivation_techniques 
		SET technique_level = technique_level + 1, progress = 0
		WHERE user_id = $1 AND technique_name = $2
	`
	_, err = db.GetPool().Exec(ctx, query, userID, techniqueName)
	return err
}

// UpdateCultivationTechniqueEffects 更新功法效果
func (db *DB) UpdateCultivationTechniqueEffects(ctx context.Context, userID int, techniqueName string, effects map[string]interface{}) error {
	effectsJSON, err := json.Marshal(effects)
	if err != nil {
		return err
	}

	query := `
		UPDATE cultivation_techniques 
		SET effects = $3
		WHERE user_id = $1 AND technique_name = $2
	`
	_, err = db.GetPool().Exec(ctx, query, userID, techniqueName, effectsJSON)
	return err
}

// AddCultivationTechniqueProgress 增加功法修炼进度
func (db *DB) AddCultivationTechniqueProgress(ctx context.Context, userID int, techniqueName string, addProgress int) error {
	query := `
		UPDATE cultivation_techniques 
		SET progress = LEAST(100, progress + $3)
		WHERE user_id = $1 AND technique_name = $2
	`
	_, err := db.GetPool().Exec(ctx, query, userID, techniqueName, addProgress)
	return err
}

// ForgetCultivationTechnique 遗忘功法
func (db *DB) ForgetCultivationTechnique(ctx context.Context, userID int, techniqueName string) error {
	query := `
		DELETE FROM cultivation_techniques
		WHERE user_id = $1 AND technique_name = $2
	`
	_, err := db.GetPool().Exec(ctx, query, userID, techniqueName)
	return err
}

// GetCultivationTechniquesByQuality 根据品质获取功法
func (db *DB) GetCultivationTechniquesByQuality(ctx context.Context, userID int, quality string) ([]*CultivationTechnique, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT user_id, technique_name, technique_type, technique_level, quality,
			progress, effects, requirements, learned_at
		FROM cultivation_techniques
		WHERE user_id = $1 AND quality = $2
		ORDER BY technique_type, technique_level DESC, learned_at DESC
	`

	rows, err := db.GetPool().Query(timeoutCtx, query, userID, quality)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var techniques []*CultivationTechnique
	for rows.Next() {
		var technique CultivationTechnique
		var effectsJSON, requirementsJSON []byte
		err := rows.Scan(
			&technique.UserID, &technique.TechniqueName, &technique.TechniqueType, &technique.TechniqueLevel, &technique.Quality,
			&technique.Progress, &effectsJSON, &requirementsJSON, &technique.LearnedAt,
		)
		if err != nil {
			return nil, err
		}

		// 解析JSON字段
		if len(effectsJSON) > 0 {
			err = json.Unmarshal(effectsJSON, &technique.Effects)
			if err != nil {
				return nil, err
			}
		}

		if len(requirementsJSON) > 0 {
			err = json.Unmarshal(requirementsJSON, &technique.Requirements)
			if err != nil {
				return nil, err
			}
		}

		techniques = append(techniques, &technique)
	}

	return techniques, rows.Err()
}

// SearchCultivationTechniquesByName 根据功法名称搜索功法
func (db *DB) SearchCultivationTechniquesByName(ctx context.Context, userID int, techniqueName string) ([]*CultivationTechnique, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT user_id, technique_name, technique_type, technique_level, quality,
			progress, effects, requirements, learned_at
		FROM cultivation_techniques
		WHERE user_id = $1 AND technique_name ILIKE '%' || $2 || '%'
		ORDER BY technique_type, quality DESC, technique_level DESC, learned_at DESC
	`

	rows, err := db.GetPool().Query(timeoutCtx, query, userID, techniqueName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var techniques []*CultivationTechnique
	for rows.Next() {
		var technique CultivationTechnique
		var effectsJSON, requirementsJSON []byte
		err := rows.Scan(
			&technique.UserID, &technique.TechniqueName, &technique.TechniqueType, &technique.TechniqueLevel, &technique.Quality,
			&technique.Progress, &effectsJSON, &requirementsJSON, &technique.LearnedAt,
		)
		if err != nil {
			return nil, err
		}

		// 解析JSON字段
		if len(effectsJSON) > 0 {
			err = json.Unmarshal(effectsJSON, &technique.Effects)
			if err != nil {
				return nil, err
			}
		}

		if len(requirementsJSON) > 0 {
			err = json.Unmarshal(requirementsJSON, &technique.Requirements)
			if err != nil {
				return nil, err
			}
		}

		techniques = append(techniques, &technique)
	}

	return techniques, rows.Err()
}

// GetMasteredCultivationTechniques 获取已完全掌握的功法（进度100%）
func (db *DB) GetMasteredCultivationTechniques(ctx context.Context, userID int) ([]*CultivationTechnique, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT user_id, technique_name, technique_type, technique_level, quality,
			progress, effects, requirements, learned_at
		FROM cultivation_techniques
		WHERE user_id = $1 AND progress = 100
		ORDER BY technique_type, quality DESC, technique_level DESC, learned_at DESC
	`

	rows, err := db.GetPool().Query(timeoutCtx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var techniques []*CultivationTechnique
	for rows.Next() {
		var technique CultivationTechnique
		var effectsJSON, requirementsJSON []byte
		err := rows.Scan(
			&technique.UserID, &technique.TechniqueName, &technique.TechniqueType, &technique.TechniqueLevel, &technique.Quality,
			&technique.Progress, &effectsJSON, &requirementsJSON, &technique.LearnedAt,
		)
		if err != nil {
			return nil, err
		}

		// 解析JSON字段
		if len(effectsJSON) > 0 {
			err = json.Unmarshal(effectsJSON, &technique.Effects)
			if err != nil {
				return nil, err
			}
		}

		if len(requirementsJSON) > 0 {
			err = json.Unmarshal(requirementsJSON, &technique.Requirements)
			if err != nil {
				return nil, err
			}
		}

		techniques = append(techniques, &technique)
	}

	return techniques, rows.Err()
}

// 错误类型定义
type TechniqueAlreadyLearnedError struct {
	TechniqueName string
}

func (e *TechniqueAlreadyLearnedError) Error() string {
	return fmt.Sprintf("功法 %s 已经学习过了", e.TechniqueName)
}

type TechniqueNotFoundError struct {
	TechniqueName string
}

func (e *TechniqueNotFoundError) Error() string {
	return fmt.Sprintf("功法 %s 未找到", e.TechniqueName)
}

type TechniqueNotReadyForUpgradeError struct {
	TechniqueName   string
	CurrentProgress int
}

func (e *TechniqueNotReadyForUpgradeError) Error() string {
	return fmt.Sprintf("功法 %s 还未修炼完成，当前进度：%d%%", e.TechniqueName, e.CurrentProgress)
}
