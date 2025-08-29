package bot

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"jiangfengwhu/nagi-bot-go/config"
	"jiangfengwhu/nagi-bot-go/database"
	"jiangfengwhu/nagi-bot-go/llm"

	"google.golang.org/genai"
	tele "gopkg.in/telebot.v4"
)

// Bot Telegram bot 包装器
type Bot struct {
	*tele.Bot
	config     *config.Config
	db         *database.DB
	llmService *llm.LLMService
}

// New 创建新的 bot 实例
func New(cfg *config.Config, db *database.DB, llmService *llm.LLMService) (*Bot, error) {
	var pref tele.Settings

	if cfg.Bot.UseWebhook {
		// 使用webhook模式
		webhook := &tele.Webhook{
			Listen: cfg.Bot.ListenPort,
			Endpoint: &tele.WebhookEndpoint{
				PublicURL: cfg.Bot.WebhookURL,
			},
		}

		pref = tele.Settings{
			Token:  cfg.Bot.Token,
			Poller: webhook,
		}
	} else {
		// 使用长轮询模式
		pref = tele.Settings{
			Token:  cfg.Bot.Token,
			Poller: &tele.LongPoller{Timeout: time.Duration(cfg.Bot.Timeout) * time.Second},
		}
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("创建 bot 失败: %v", err)
	}

	bot := &Bot{
		Bot:        b,
		config:     cfg,
		db:         db,
		llmService: llmService,
	}

	bot.setupHandlers()
	return bot, nil
}

// setupHandlers 设置处理器
func (b *Bot) setupHandlers() {
	needAuth := b.Group()
	needAuth.Use(Auth(b.db, b.Bot))
	needAuth.Handle("/start", b.handleStart)
	needAuth.Handle(tele.OnText, b.handleChat)
	needAuth.Handle(tele.OnPhoto, b.handleFile)
	needAuth.Handle(tele.OnAudio, b.handleFile)
	needAuth.Handle(tele.OnVoice, b.handleFile)
	needAuth.Handle(tele.OnDocument, b.handleFile)
	needAuth.Handle("/c", b.handleRecharge)
	needAuth.Handle("/reg", b.handleRegister)
}

func (b *Bot) handleRegister(c tele.Context) error {
	user := c.Get("db_user").(*database.User)
	name := strings.TrimSpace(strings.TrimPrefix(c.Message().Text, "/reg"))
	if name == "" {
		return c.Reply("请输入角色名")
	}
	player, err := b.db.GetCharacterStats(context.Background(), user.ID)
	if err != nil {
		return c.Reply(fmt.Sprintf("获取玩家信息失败: %v", err))
	}
	if player != nil {
		return c.Reply("您已经注册了角色，/start查看角色信息")
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
				"max_hp": {
					Type:        genai.TypeInteger,
					Description: "生命值，主要基于根骨和体质型灵根（土、木）计算，特殊体质可能有加成",
				},
				"max_mp": {
					Type:        genai.TypeInteger,
					Description: "法力值，基于灵根总体强度和特殊灵根（如水、冰、雷）计算",
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
			},
			Required: []string{"player_name", "spiritual_roots", "physique", "comprehension", "luck", "spirit_sense", "max_hp", "max_mp", "attack", "defense", "speed", "lifespan", "background_story"},
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
	player, err = b.llmService.CreatePlayer(b.db, user.ID, result.Text())
	if err != nil {
		b.Edit(message, fmt.Sprintf("创建角色失败: %v", err))
		return err
	}
	b.Edit(message, fmt.Sprintf("角色创建成功: %s", player))
	return nil
}

func (b *Bot) handleRecharge(c tele.Context) error {
	user := c.Get("db_user").(*database.User)
	if !slices.Contains(b.config.Bot.AdminIds, user.TgId) {
		return c.Reply("您没有权限使用此命令")
	}
	args := c.Args()
	if len(args) != 2 {
		return c.Reply("请输入正确的命令，格式为: /c <id> <amount>")
	}
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return c.Reply("请输入正确的id")
	}
	amount, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return c.Reply("请输入正确的金额")
	}
	err = b.db.UpdateUserTotalRechargedToken(context.Background(), id, amount)
	if err != nil {
		return c.Reply(fmt.Sprintf("充值失败: %v", err))
	}
	return c.Reply(fmt.Sprintf("充值成功，充值金额为%d个token", amount))
}

// handleStart 处理 /start 命令
func (b *Bot) handleStart(c tele.Context) error {
	user := c.Get("db_user").(*database.User)
	player, err := b.db.GetCharacterStats(context.Background(), user.ID)
	if err != nil {
		return c.Reply(fmt.Sprintf("获取玩家信息失败: %v", err))
	}
	if player == nil {
		return c.Reply("您还没有注册角色，请使用/reg 角色名 注册之后进行游戏")
	} else {
		return c.Reply(fmt.Sprintf("欢迎回来：\n\n灵石: %d\n\n角色信息: %s", user.TotalRechargedToken-user.TotalUsedToken, player))
	}
}

func (b *Bot) handleFile(c tele.Context) error {
	var file *tele.File
	var filePath string
	if c.Message().Photo != nil {
		file = &c.Message().Photo.File
		filePath = c.Message().Photo.FileID + ".jpg"
	} else if c.Message().Audio != nil {
		file = &c.Message().Audio.File
		filePath = c.Message().Audio.FileID + ".mp3"
	} else if c.Message().Voice != nil {
		file = &c.Message().Voice.File
		filePath = c.Message().Voice.FileID + ".ogg"
	} else if c.Message().Document != nil {
		file = &c.Message().Document.File
		filePath = c.Message().Document.FileID + c.Message().Document.FileName
	}
	if file != nil {
		user := c.Get("db_user").(*database.User)
		message, err := b.Reply(c.Message(), "正在下载...")
		if err != nil {
			return c.Reply(fmt.Sprintf("发送消息失败: %v", err))
		}
		err = b.Download(file, filePath)
		if err != nil {
			return c.Reply(fmt.Sprintf("从Telegram下载文件失败: %v", err))
		}
		defer os.Remove(filePath)
		ctx := context.Background()
		client, err := b.llmService.NewClient(ctx)
		if err != nil {
			return c.Reply(fmt.Sprintf("创建LLMClient失败: %v", err))
		}
		b.Edit(message, "正在上传...")
		uploadedFile, err := client.Files.UploadFromPath(ctx, filePath, nil)
		if err != nil {
			return c.Reply(fmt.Sprintf("上传到LLM失败: %v", err))
		}
		b.db.AddMessage(ctx, user.ID, "user", []*genai.Part{genai.NewPartFromURI(uploadedFile.URI, uploadedFile.MIMEType)})
		if c.Message().Voice != nil {
			b.Edit(message, "上传成功")
			c.Message().Text = "请回复这条语音消息"
			b.handleChat(c)
		} else {
			_, err = b.Edit(message, "上传成功，请继续您的对话")
		}
		return err
	}
	return nil
}

func (b *Bot) handleChat(c tele.Context) error {
	user := c.Get("db_user").(*database.User)
	player, _ := b.db.GetCharacterStats(context.Background(), user.ID)
	if player == nil {
		return c.Reply("您还没有注册角色，请使用/reg 角色名 注册之后进行游戏")
	}
	message, err := b.Reply(c.Message(), "正在思考...")
	if err != nil {
		return c.Reply(fmt.Sprintf("发送消息失败: %v", err))
	}
	systemPrompt := b.config.Prompts["system_prompt"] + fmt.Sprintf("\n\n玩家%s的信息如下：\n\n%s\n\n", player.Name, player)
	history := []*genai.Content{
		genai.NewContentFromText(systemPrompt, genai.RoleUser),
	}
	username := "[" + player.Name + "]"
	text := fmt.Sprintf("%s: %s", username, c.Message().Text)
	nextParts := []*genai.Part{genai.NewPartFromText(text)}
	llmResult := ""
	promptToken := int32(0)
	totalToken := int32(0)

	// === start llm ====
	ctx := context.Background()
	client, err := b.llmService.NewClient(ctx)
	if err != nil {
		return c.Reply(fmt.Sprintf("创建LLMClient失败: %v", err))
	}

	historyMsgs, err := b.db.GetRecentMessages(ctx, user.ID, 10)
	if err != nil {
		c.Reply(fmt.Sprintf("获取消息失败: %v", err))
	}
	for _, msg := range historyMsgs {
		history = append(history, genai.NewContentFromParts(msg.Content.([]*genai.Part), genai.Role(msg.Role)))
	}

	chatClient, err := b.llmService.CreateConversation(ctx, client, "gemini-2.5-flash", history)
	if err != nil {
		return c.Reply(fmt.Sprintf("创建聊天失败: %v", err))
	}
	for len(nextParts) > 0 {
		streamID, err := b.llmService.Chat(ctx, chatClient, nextParts)
		defer b.llmService.DeleteStream(streamID)
		if err != nil {
			return c.Reply(fmt.Sprintf("聊天失败: %v", err))
		}

		stream, err := b.llmService.SSE(streamID)
		if err != nil {
			return c.Reply(fmt.Sprintf("获取流失败: %v", err))
		}

		b.db.AddMessage(ctx, user.ID, "user", nextParts)

		toolCalls := []*genai.FunctionCall{}
		nextParts = []*genai.Part{}
		tmpResult := ""
		thoughtSignature := []byte{}

		for chunk, err := range stream.Stream {
			if err != nil {
				c.Reply(fmt.Sprintf("获取流失败: %v", err))
				continue
			}
			part := chunk.Candidates[0].Content.Parts[0]
			tmpResult += part.Text
			thoughtSignature = append(thoughtSignature, part.ThoughtSignature...)
			if tmpResult != "" {
				b.Edit(message, llmResult+ConvertMarkdownToTelegramMarkdownV2(tmpResult), tele.ModeMarkdownV2)
			}
			toolCalls = append(toolCalls, chunk.FunctionCalls()...)
			promptToken = chunk.UsageMetadata.PromptTokenCount + chunk.UsageMetadata.ThoughtsTokenCount
			totalToken += (chunk.UsageMetadata.CandidatesTokenCount + chunk.UsageMetadata.ToolUsePromptTokenCount)
		}
		if tmpResult != "" {
			llmResult += tmpResult
			part := genai.NewPartFromText(tmpResult)
			if len(thoughtSignature) > 0 {
				part.ThoughtSignature = thoughtSignature
			}
			b.db.AddMessage(ctx, user.ID, "model", []*genai.Part{part})
		}
		totalToken += promptToken
		for _, tool := range toolCalls {
			if tool.Name == string(llm.ToolGenerateImage) {
				b.Edit(message, llmResult+"\n\n正在生成图片："+tool.Args["prompt"].(string))
				image, token, err := b.llmService.GenerateImage(tool.Args["prompt"].(string))
				totalToken += token
				if err != nil {
					b.Edit(message, fmt.Sprintf("生成图片失败: %v", err))
					return nil
				}
				b.Edit(message, llmResult+"\n\n图片生成完毕，正在发送图片...")
				c.Reply(&tele.Photo{File: tele.FromReader(bytes.NewReader(image)), Caption: tool.Args["prompt"].(string)})
				b.db.AddMessage(ctx, user.ID, "model", []*genai.Part{genai.NewPartFromFunctionCall(tool.Name, tool.Args)})
				nextPart := genai.NewPartFromFunctionResponse(tool.Name, map[string]any{
					"text": "图片生成成功",
				})
				nextParts = append(nextParts, nextPart)
			} else if tool.Name == string(llm.ToolGetTime) {
				b.Edit(message, llmResult+"\n\n正在获取时间")
				time := b.llmService.GetTime()
				b.db.AddMessage(ctx, user.ID, "model", []*genai.Part{genai.NewPartFromFunctionCall(tool.Name, tool.Args)})
				nextPart := genai.NewPartFromFunctionResponse(tool.Name, map[string]any{
					"text": time,
				})
				nextParts = append(nextParts, nextPart)
			} else if tool.Name == string(llm.ToolGoogleSearch) {
				b.Edit(message, llmResult+"\n\n正在Google搜索："+tool.Args["prompt"].(string))
				searchResult, err := b.llmService.GoogleSearch(tool.Args["prompt"].(string))
				if err != nil {
					b.Edit(message, llmResult+fmt.Sprintf("Google搜索失败: %v", err))
					return nil
				}
				b.Edit(message, llmResult+"\n\nGoogle搜索结果："+searchResult)
				b.db.AddMessage(ctx, user.ID, "model", []*genai.Part{genai.NewPartFromFunctionCall(tool.Name, tool.Args)})
				nextPart := genai.NewPartFromFunctionResponse(tool.Name, map[string]any{
					"text": searchResult,
				})
				nextParts = append(nextParts, nextPart)
			}
		}
	}
	b.db.UpdateUserTotalUsedToken(ctx, user.TgId, int64(totalToken))

	return nil
}

// Run 启动 bot
func (b *Bot) Run() {
	if b.config.Bot.UseWebhook {
		log.Printf("Bot 开始运行 (Webhook 模式)...")
		log.Printf("监听端口: %s", b.config.Bot.ListenPort)
		log.Printf("Webhook URL: %s", b.config.Bot.WebhookURL)
	} else {
		log.Println("Bot 开始运行 (长轮询模式)...")
	}
	b.Start()
}
