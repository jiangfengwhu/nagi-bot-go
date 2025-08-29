package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config 配置结构体
type Config struct {
	Bot struct {
		Token      string  `json:"token"`
		Timeout    int     `json:"timeout"`
		AdminIds   []int64 `json:"admin_ids"`
		UseWebhook bool    `json:"use_webhook"`
		WebhookURL string  `json:"webhook_url"`
		ListenPort string  `json:"listen_port"`
	} `json:"bot"`
	Database struct {
		URL string `json:"url"`
	} `json:"database"`
	LLM struct {
		APIKeys             string `json:"api_keys"`
		BaseURL             string `json:"base_url"`
		GoogleSearchAPIKeys string `json:"google_search_api_keys"`
	} `json:"llm"`
	Prompts map[string]string `json:"prompts"`
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

	// 对所有Prompts中的字段进行读取文件替换
	for key, value := range config.Prompts {
		if content, err := os.ReadFile(value); err == nil {
			config.Prompts[key] = string(content)
		}
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

	// 验证webhook配置
	if c.Bot.UseWebhook {
		if c.Bot.WebhookURL == "" {
			return fmt.Errorf("启用webhook时必须设置webhook_url")
		}
		if c.Bot.ListenPort == "" {
			c.Bot.ListenPort = ":8080" // 使用HTTP默认端口，因为Cloudflare会处理HTTPS
		}
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
