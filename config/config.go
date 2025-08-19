package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config 配置结构体
type Config struct {
	Bot struct {
		Token               string  `json:"token"`
		Timeout             int     `json:"timeout"`
		DefaultSystemPrompt string  `json:"default_system_prompt"`
		AdminIds            []int64 `json:"admin_ids"`
	} `json:"bot"`
	Database struct {
		URL string `json:"url"`
	} `json:"database"`
	LLM struct {
		APIKeys string `json:"api_keys"`
		BaseURL string `json:"base_url"`
	} `json:"llm"`
}

// Load 从配置文件加载配置
func Load(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("无法打开配置文件 %s: %v", filename, err)
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return &config, nil
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Bot.Token == "" || c.Bot.Token == "YOUR_BOT_TOKEN_HERE" {
		return fmt.Errorf("请在配置文件中设置有效的 bot token")
	}

	if c.Bot.Timeout <= 0 {
		c.Bot.Timeout = 10 // 设置默认值
	}

	// 验证数据库配置
	if c.Database.URL == "" {
		return fmt.Errorf("请在配置文件中设置数据库连接 URL")
	}

	// 验证LLM配置
	if c.LLM.APIKeys == "" {
		return fmt.Errorf("请在配置文件中设置LLM API密钥")
	}

	return nil
}
