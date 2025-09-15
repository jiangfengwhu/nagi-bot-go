package bot

import (
	"context"
	"fmt"
	"strings"
	"time"

	"encoding/json"
	"jiangfengwhu/nagi-bot-go/database"
	"jiangfengwhu/nagi-bot-go/llm"

	"google.golang.org/genai"
	tele "gopkg.in/telebot.v4"
)

func (b *Bot) handleRegister(c tele.Context) error {
	user := c.Get("db_user").(*database.User)
	name := strings.TrimSpace(strings.TrimPrefix(c.Message().Text, "/reg"))
	if name == "" {
		return c.Reply("请输入角色名, 长度不超过10个字符")
	}
	if len(name) > 20 {
		return c.Reply("角色名长度不超过10个字符")
	}
	dbPlayer, err := b.db.GetCharacterStats(context.Background(), user.ID)
	if err != nil {
		return c.Reply(fmt.Sprintf("获取玩家信息失败: %v", err))
	}
	if dbPlayer != nil {
		return c.Reply("您已经注册了角色，/start查看角色信息")
	}
	exists, err := b.db.NameExists(context.Background(), name)
	if err != nil {
		return c.Reply(fmt.Sprintf("获取玩家信息失败: %v", err))
	}
	if exists {
		return c.Reply("该角色名已存在，请重新输入")
	}
	ctx := context.Background()
	client, err := b.llmService.NewClient(ctx)
	if err != nil {
		return c.Reply(fmt.Sprintf("创建LLMClient失败: %v", err))
	}
	config := &genai.GenerateContentConfig{
		ResponseMIMEType:  "application/json",
		SystemInstruction: genai.NewContentFromText(b.config.Prompts["system_prompt"], genai.RoleUser),
		ResponseSchema: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"player_name": {
					Type:        genai.TypeString,
					Description: "修仙者的姓名",
				},
				"spiritual_roots": {
					Type: genai.TypeArray,
					Items: &genai.Schema{
						Type:        genai.TypeObject,
						Description: "单个灵根的键值对",
						Properties: map[string]*genai.Schema{
							"root_name": {
								Type:        genai.TypeString,
								Description: "灵根的名称 (常见灵根：'金', '木', '水', '火', '土'，特殊灵根：冰、雷、风、暗、光、空间、时间、混沌等指定灵根，比较罕见)",
							},
							"affinity": {
								Type:        genai.TypeInteger,
								Description: "该灵根的资质数值 (0-100)，越大越罕见",
							},
						},
						Required: []string{"root_name", "affinity"},
					},
					Description: "灵根属性列表，每个元素包含灵根名称和其对应的资质。",
				},
				"physique": {
					Type:        genai.TypeInteger,
					Description: "根骨/体魄，影响生命值、攻击力、防御力",
				},
				"comprehension": {
					Type:        genai.TypeInteger,
					Description: "悟性，影响修炼速度和功法领悟",
				},
				"luck": {
					Type:        genai.TypeInteger,
					Description: "幸运值，影响奇遇概率",
				},
				"spirit_sense": {
					Type:        genai.TypeInteger,
					Description: "神识强度，根据灵根（特别是精神系灵根如空间、时间）和悟性计算。影响感知能力和法术威力",
				},
				"attack": {
					Type:        genai.TypeInteger,
					Description: "攻击力，主要基于金灵根、根骨，攻击型特殊灵根（雷、火）有加成",
				},
				"defense": {
					Type:        genai.TypeInteger,
					Description: "防御力，主要基于土灵根、根骨，防御型特殊灵根（冰、暗）有加成",
				},
				"speed": {
					Type:        genai.TypeInteger,
					Description: "速度，主要基于木灵根，风灵根有巨大加成，雷灵根也有提升",
				},
				"lifespan": {
					Type:        genai.TypeInteger,
					Description: "寿命（100-200），基于灵根品质和特殊灵根计算。时间灵根、混沌灵根等顶级灵根大幅延长寿命",
				},
				"background_story": {
					Type:        genai.TypeString,
					Description: "背景故事，要结合灵根特点描述角色的出身和修仙机缘",
				},
				"init_inventory": {
					Type:        genai.TypeArray,
					Items:       llm.InventoryItemSchema,
					Description: "初始背包物品列表，每个元素包含物品名称，数量，类型，品质，等级，属性，描述等，列表是根据角色的命格生成，有好有坏，全凭命理（不包含灵石）",
				},
			},
			Required: []string{"player_name", "spiritual_roots", "physique", "comprehension", "luck", "spirit_sense", "attack", "defense", "speed", "lifespan", "background_story", "init_inventory"},
		},
	}
	message, _ := b.Reply(c.Message(), "正在生成角色...")
	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		genai.Text(fmt.Sprintf("创建一个修仙者角色，角色名称为: %s，现在时间是: %s", name, time.Now().Format("2006-01-02 15:04:05"))),
		config,
	)
	if err != nil {
		b.Edit(message, fmt.Sprintf("创建角色失败: %v", err))
		return err
	}
	fmt.Println(result.Text())
	player, inventory, err := CreatePlayer(b.db, user.ID, result.Text())
	if err != nil {
		b.Edit(message, fmt.Sprintf("创建角色失败: %v", err))
		return err
	}
	b.Edit(message, fmt.Sprintf("角色创建成功: \n%s\n初始背包物品: \n%s\n", formatPlayerInfo(player), formatInventoryInfo(inventory)))
	return nil
}

// CreatePlayer 创建新的修仙者角色
func CreatePlayer(db *database.DB, userID int, args string) (*database.CharacterStats, []*database.InventoryItem, error) {
	ctx := context.Background()

	// 使用 JSON 序列化进行类型转换
	var params CreatePlayerParams
	err := json.Unmarshal([]byte(args), &params)
	if err != nil {
		return nil, nil, fmt.Errorf("解析参数失败: %v", err)
	}

	stats := &database.CharacterStats{
		UserID:         userID,
		Name:           params.PlayerName,
		Realm:          "练气期",
		RealmLevel:     1,
		SpiritualRoots: params.SpiritualRoots,
		SpiritSense:    params.SpiritSense,
		Physique:       params.Physique,
		DemonicAura:    0,
		TaoistName:     "", // 练气期没有道号
		Attack:         params.Attack,
		Defense:        params.Defense,
		Speed:          params.Speed,
		Luck:           params.Luck,
		Experience:     0,
		Comprehension:  params.Comprehension,
		Age:            1,
		Lifespan:       params.Lifespan,
		Location:       "新手村",
		Status:         "健康",
		Stories:        params.BackgroundStory,
	}

	inventory := []*database.InventoryItem{}
	for _, item := range params.InitInventory {
		inventory = append(inventory, &database.InventoryItem{
			UserID:       userID,
			ItemName:     item.ItemName,
			Quantity:     item.Quantity,
			ItemType:     item.ItemType,
			Quality:      item.Quality,
			Level:        item.Level,
			Properties:   item.Properties,
			Description:  item.Description,
			ObtainedFrom: item.ObtainedFrom,
			ObtainedAt:   time.Now(),
		})
	}

	tx, err := db.GetPool().Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback(ctx)

	for _, item := range inventory {
		item.UserID = userID
	}

	err = db.AddInventoryItemsBatchInTx(ctx, tx, inventory)
	if err != nil {
		return nil, nil, fmt.Errorf("添加物品失败: %v", err)
	}

	// 保存到数据库
	err = db.CreateCharacterStatsInTx(ctx, tx, stats)
	if err != nil {
		return nil, nil, fmt.Errorf("创建角色失败: %v", err)
	}

	return stats, inventory, tx.Commit(ctx)
}
