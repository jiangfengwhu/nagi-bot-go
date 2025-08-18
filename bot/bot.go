package bot

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
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
	needAuth.Use(Auth(b.db, b.Bot))
	needAuth.Handle("/start", b.handleStart)
	needAuth.Handle(tele.OnText, b.handleChat)
	needAuth.Handle(tele.OnPhoto, b.handleChat)
	needAuth.Handle("/prompt", b.handleSystemPrompt)
}

func (b *Bot) handleSystemPrompt(c tele.Context) error {
	user := c.Get("db_user").(*database.User)
	prompt := strings.TrimSpace(strings.TrimPrefix(c.Message().Text, "/prompt"))
	if prompt == "" {
		return c.Reply("您的系统提示词为: \n\n" + user.SystemPrompt)
	}
	b.db.UpdateUserSystemPrompt(context.Background(), user.TgId, prompt)
	return c.Reply("系统提示词已更新")
}

// handleStart 处理 /start 命令
func (b *Bot) handleStart(c tele.Context) error {
	user := c.Get("db_user").(*database.User)
	return c.Reply(fmt.Sprintf("欢迎回来，您的余额为%d个token\n\n使用/prompt 设置系统提示词，留空可查询当前提示词", user.TotalRechargedToken-user.TotalUsedToken))
}

func (b *Bot) handleChat(c tele.Context) error {
	user := c.Get("db_user").(*database.User)
	message, err := b.Reply(c.Message(), "正在思考...")
	if err != nil {
		return c.Reply(fmt.Sprintf("发送消息失败: %v", err))
	}
	systemPrompt := b.config.Bot.DefaultSystemPrompt + "\n\n" + user.SystemPrompt
	history := []*genai.Content{
		genai.NewContentFromText(systemPrompt, genai.RoleUser),
	}
	photo := c.Message().Photo
	caption := c.Message().Caption
	text := c.Message().Text
	nextParts := []*genai.Part{}
	llmResult := ""

	// === start llm ====
	ctx := context.Background()
	client, err := b.llmService.NewClient(ctx)
	if err != nil {
		return c.Reply(fmt.Sprintf("创建LLMClient失败: %v", err))
	}
	if photo != nil {
		b.Edit(message, "正在上传图片...")
		filePath := photo.FileID + ".jpg"
		err := b.Download(&photo.File, filePath)
		if err != nil {
			return c.Reply(fmt.Sprintf("下载图片失败: %v", err))
		}
		defer os.Remove(filePath)
		uploadedFile, err := client.Files.UploadFromPath(ctx, filePath, nil)
		if err != nil {
			return c.Reply(fmt.Sprintf("上传图片失败: %v", err))
		}
		nextParts = append(nextParts, genai.NewPartFromURI(uploadedFile.URI, uploadedFile.MIMEType))
		if caption != "" {
			nextParts = append(nextParts, genai.NewPartFromText(caption))
		} else {
			nextParts = append(nextParts, genai.NewPartFromText("请描述该图片的内容"))
		}
	}

	if text != "" {
		nextParts = append(nextParts, genai.NewPartFromText(text))
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
		for chunk := range stream.Stream {
			tmpResult += chunk.Text()
			if tmpResult != "" {
				b.Edit(message, llmResult+tmpResult)
			}
			toolCalls = append(toolCalls, chunk.FunctionCalls()...)
		}
		if tmpResult != "" {
			llmResult += tmpResult
			b.db.AddMessage(ctx, user.ID, "model", []*genai.Part{genai.NewPartFromText(tmpResult)})
		}
		for _, tool := range toolCalls {
			if tool.Name == string(llm.ToolGenerateImage) {
				b.Edit(message, llmResult+"\n\n正在生成图片："+tool.Args["prompt"].(string))
				b.db.AddMessage(ctx, user.ID, "model", []*genai.Part{genai.NewPartFromFunctionCall(tool.Name, tool.Args)})
				image, err := b.llmService.GenerateImage(tool.Args["prompt"].(string))
				if err != nil {
					b.Edit(message, fmt.Sprintf("生成图片失败: %v", err))
					return nil
				}
				c.Reply(&tele.Photo{File: tele.FromReader(bytes.NewReader(image)), Caption: tool.Args["prompt"].(string)})
				nextPart := genai.NewPartFromFunctionResponse(tool.Name, map[string]any{
					"text": "图片生成成功",
				})
				nextParts = append(nextParts, nextPart)
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
