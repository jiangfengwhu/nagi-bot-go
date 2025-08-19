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
	needAuth.Handle("/c", b.handleRecharge)
	needAuth.Handle("/my", b.handleMy)
}

func (b *Bot) handleMy(c tele.Context) error {
	user := c.Get("db_user").(*database.User)
	return c.Reply(fmt.Sprintf("您的ID为: %d\n您的余额为: %d\n您的系统提示词为: %s", user.ID, user.TotalRechargedToken-user.TotalUsedToken, user.SystemPrompt))
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
	promptToken := int32(0)
	totalToken := int32(0)

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

		for chunk, err := range stream.Stream {
			if err != nil {
				c.Reply(fmt.Sprintf("获取流失败: %v", err))
				continue
			}
			tmpResult += chunk.Text()
			if tmpResult != "" {
				b.Edit(message, llmResult+ConvertMarkdownToTelegramMarkdownV2(tmpResult), tele.ModeMarkdownV2)
			}
			toolCalls = append(toolCalls, chunk.FunctionCalls()...)
			promptToken = chunk.UsageMetadata.PromptTokenCount + chunk.UsageMetadata.ThoughtsTokenCount
			totalToken += (chunk.UsageMetadata.CandidatesTokenCount + chunk.UsageMetadata.ToolUsePromptTokenCount)
		}
		if tmpResult != "" {
			llmResult += tmpResult
			b.db.AddMessage(ctx, user.ID, "model", []*genai.Part{genai.NewPartFromText(tmpResult)})
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
	log.Println("Bot 开始运行...")
	b.Start()
}
