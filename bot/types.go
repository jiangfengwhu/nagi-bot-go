package bot

import (
	"time"

	"jiangfengwhu/nagi-bot-go/database"
)

// CreatePlayerParams 创建玩家的参数结构
type CreatePlayerParams struct {
	PlayerName      string                   `json:"player_name"`
	SpiritualRoots  *database.SpiritualRoots `json:"spiritual_roots"`
	Physique        int                      `json:"physique"`
	Comprehension   int                      `json:"comprehension"`
	Luck            int                      `json:"luck"`
	SpiritSense     int                      `json:"spirit_sense"`
	Attack          int                      `json:"attack"`
	Defense         int                      `json:"defense"`
	Speed           int                      `json:"speed"`
	Lifespan        int                      `json:"lifespan"`
	BackgroundStory string                   `json:"background_story"`
	InitInventory   []*CreateInventoryParams `json:"init_inventory"`
}

type CreateInventoryParams struct {
	ItemName     string    `json:"item_name"`
	Quantity     int       `json:"quantity"`
	ItemType     string    `json:"item_type"`
	Quality      string    `json:"quality"`
	Level        int       `json:"level"`
	Properties   string    `json:"properties"`
	Description  string    `json:"description"`
	ObtainedFrom string    `json:"obtained_from"`
	ObtainedAt   time.Time `json:"obtained_at"`
}
