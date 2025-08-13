package llm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"jiangfengwhu/nagi-bot-go/config"

	"google.golang.org/genai"
)

// Client LLM客户端结构体
type Client struct {
	client     *genai.Client
	config     *config.Config
	maxRetries int
	retryDelay time.Duration
}

// Message 消息结构体
type Message struct {
	Role    string `json:"role"`    // "user" 或 "assistant" 或 "system"
	Content string `json:"content"` // 消息内容
}

// ChatRequest 聊天请求结构体
type ChatRequest struct {
	Messages []Message `json:"messages"` // 对话历史
	Stream   bool      `json:"stream"`   // 是否流式响应
}

// ChatResponse 聊天响应结构体
type ChatResponse struct {
	Content      string `json:"content"`       // 响应内容
	TokensUsed   int    `json:"tokens_used"`   // 使用的token数量
	Model        string `json:"model"`         // 使用的模型
	FinishReason string `json:"finish_reason"` // 完成原因
}

// NewClient 创建新的LLM客户端
func NewClient(cfg *config.Config) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 创建GenAI客户端配置
	clientConfig := &genai.ClientConfig{
		APIKey:  cfg.LLM.APIKey,
		Backend: genai.BackendGeminiAPI,
	}

	// 如果配置了BaseURL，设置为自定义后端
	if cfg.LLM.BaseURL != "" {
		clientConfig.Backend = genai.BackendVertexAI // 或其他适当的后端
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, clientConfig)
	if err != nil {
		return nil, fmt.Errorf("创建GenAI客户端失败: %v", err)
	}

	return &Client{
		client:     client,
		config:     cfg,
		maxRetries: 3,
		retryDelay: time.Second * 2,
	}, nil
}

// Close 关闭客户端连接
func (c *Client) Close() error {
	// genai.Client 可能没有 Close 方法，这里留空
	return nil
}

// Chat 发送聊天请求并获取响应
func (c *Client) Chat(ctx context.Context, request *ChatRequest) (*ChatResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("请求不能为空")
	}

	if len(request.Messages) == 0 {
		return nil, fmt.Errorf("消息列表不能为空")
	}

	// 构建提示文本（简化处理，将对话历史合并）
	var promptBuilder strings.Builder

	for _, msg := range request.Messages {
		if msg.Content == "" {
			continue
		}

		switch strings.ToLower(msg.Role) {
		case "user":
			promptBuilder.WriteString("用户: " + msg.Content + "\n")
		case "assistant", "model":
			promptBuilder.WriteString("助手: " + msg.Content + "\n")
		case "system":
			promptBuilder.WriteString("系统: " + msg.Content + "\n")
		}
	}

	prompt := promptBuilder.String()
	if strings.TrimSpace(prompt) == "" {
		return nil, fmt.Errorf("没有有效的消息内容")
	}

	// 构建内容和配置（使用默认值）
	contents := genai.Text(prompt)

	// 使用默认配置值
	temperature := float32(0.7)
	maxTokens := int32(4096)

	config := &genai.GenerateContentConfig{
		Temperature:     &temperature,
		MaxOutputTokens: maxTokens,
	}

	// 执行聊天请求（带重试机制）
	var response *genai.GenerateContentResponse
	var err error

	// 使用默认模型名称
	modelName := "gemini-1.5-flash"

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		response, err = c.client.Models.GenerateContent(ctx, modelName, contents, config)
		if err == nil {
			break
		}

		// 如果不是最后一次重试，等待后重试
		if attempt < c.maxRetries {
			time.Sleep(c.retryDelay * time.Duration(attempt+1))
			continue
		}

		return nil, fmt.Errorf("聊天请求失败，已重试%d次: %v", c.maxRetries, err)
	}

	if response == nil || len(response.Candidates) == 0 {
		return nil, fmt.Errorf("没有收到有效响应")
	}

	// 提取响应文本
	var responseText strings.Builder
	candidate := response.Candidates[0]
	if candidate.Content != nil {
		for _, part := range candidate.Content.Parts {
			if part != nil && part.Text != "" {
				responseText.WriteString(part.Text)
			}
		}
	}

	// 获取token使用信息
	tokensUsed := 0
	if response.UsageMetadata != nil {
		tokensUsed = int(response.UsageMetadata.TotalTokenCount)
	}

	return &ChatResponse{
		Content:      responseText.String(),
		TokensUsed:   tokensUsed,
		Model:        modelName,
		FinishReason: "completed",
	}, nil
}

// SimpleChat 简单的文本聊天，直接传入用户消息
func (c *Client) SimpleChat(ctx context.Context, userMessage string) (*ChatResponse, error) {
	if strings.TrimSpace(userMessage) == "" {
		return nil, fmt.Errorf("用户消息不能为空")
	}

	request := &ChatRequest{
		Messages: []Message{
			{
				Role:    "user",
				Content: userMessage,
			},
		},
		Stream: false,
	}

	return c.Chat(ctx, request)
}

// ChatWithHistory 带历史记录的聊天
func (c *Client) ChatWithHistory(ctx context.Context, messages []Message, newMessage string) (*ChatResponse, error) {
	if strings.TrimSpace(newMessage) == "" {
		return nil, fmt.Errorf("新消息不能为空")
	}

	// 复制历史消息
	allMessages := make([]Message, 0, len(messages)+1)
	allMessages = append(allMessages, messages...)

	// 添加新消息
	allMessages = append(allMessages, Message{
		Role:    "user",
		Content: newMessage,
	})

	request := &ChatRequest{
		Messages: allMessages,
		Stream:   false,
	}

	return c.Chat(ctx, request)
}
