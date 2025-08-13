package main

import (
	"log"

	"jiangfengwhu/nagi-bot-go/bot"
	"jiangfengwhu/nagi-bot-go/config"
	"jiangfengwhu/nagi-bot-go/database"
)

func main() {
	// 加载配置文件
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		log.Fatalf("配置验证失败: %v", err)
	}

	// 初始化数据库连接
	db, err := database.New(cfg.Database.URL)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()

	// 创建并启动 bot
	b, err := bot.New(cfg, db)
	if err != nil {
		log.Fatalf("创建 bot 失败: %v", err)
	}

	b.Run()
}
