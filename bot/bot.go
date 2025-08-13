package bot

import (
	"fmt"
	"log"
	"time"

	"jiangfengwhu/nagi-bot-go/config"
	"jiangfengwhu/nagi-bot-go/database"

	tele "gopkg.in/telebot.v4"
)

// Bot Telegram bot 包装器
type Bot struct {
	*tele.Bot
	config *config.Config
	db     *database.DB
}

// New 创建新的 bot 实例
func New(cfg *config.Config, db *database.DB) (*Bot, error) {
	pref := tele.Settings{
		Token:  cfg.Bot.Token,
		Poller: &tele.LongPoller{Timeout: time.Duration(cfg.Bot.Timeout) * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("创建 bot 失败: %v", err)
	}

	bot := &Bot{
		Bot:    b,
		config: cfg,
		db:     db,
	}

	bot.setupHandlers()
	return bot, nil
}

// setupHandlers 设置处理器
func (b *Bot) setupHandlers() {
	needAuth := b.Group()
	needAuth.Use(Auth(b.db))
	needAuth.Handle("/start", b.handleStart)
}

// handleStart 处理 /start 命令
func (b *Bot) handleStart(c tele.Context) error {
	user := c.Get("db_user").(*database.User)
	return c.Send(fmt.Sprintf("欢迎回来，您的余额为%d个token", user.TotalRechargedToken-user.TotalUsedToken))
}

// Run 启动 bot
func (b *Bot) Run() {
	log.Println("Bot 开始运行...")
	b.Start()
}
