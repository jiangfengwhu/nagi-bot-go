package bot

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"jiangfengwhu/nagi-bot-go/config"
	"jiangfengwhu/nagi-bot-go/database"
	"jiangfengwhu/nagi-bot-go/llm"

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
	pref := tele.Settings{
		Token:  cfg.Bot.Token,
		Poller: &tele.LongPoller{Timeout: time.Duration(cfg.Bot.Timeout) * time.Second},
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
	needAuth.Use(Auth(b.db))
	needAuth.Handle("/start", b.handleStart)
	needAuth.Handle(tele.OnText, b.handleChat)
}

// handleStart 处理 /start 命令
func (b *Bot) handleStart(c tele.Context) error {
	user := c.Get("db_user").(*database.User)
	return c.Send(fmt.Sprintf("欢迎回来，您的余额为%d个token", user.TotalRechargedToken-user.TotalUsedToken))
}

func (b *Bot) handleChat(c tele.Context) error {
	message, err := b.Send(c.Sender(), "正在思考...")
	if err != nil {
		return c.Send(fmt.Sprintf("发送消息失败: %v", err))
	}
	streamID, err := b.llmService.Chat(c.Message().Text, "gemini-2.5-flash")
	defer b.llmService.DeleteStream(streamID)
	if err != nil {
		return c.Send(fmt.Sprintf("聊天失败: %v", err))
	}

	stream, err := b.llmService.SSE(streamID)
	if err != nil {
		return c.Send(fmt.Sprintf("获取流失败: %v", err))
	}

	llmResult := ""
	for response := range stream.Stream {
		llmResult += response.Text()
		if llmResult != "" {
			b.Edit(message, llmResult)
		}
		for _, tool := range response.FunctionCalls() {
			if tool.Name == string(llm.ToolGenerateImage) {
				c.Send("正在生成图片：" + tool.Args["prompt"].(string))
				image, err := b.llmService.GenerateImage(tool.Args["prompt"].(string))
				if err != nil {
					return c.Send(fmt.Sprintf("生成图片失败: %v", err))
				}
				c.Send(&tele.Photo{File: tele.FromReader(bytes.NewReader(image))})
			}
		}
	}

	return nil
}

// Run 启动 bot
func (b *Bot) Run() {
	log.Println("Bot 开始运行...")
	b.Start()
}
