package llm

import (
	"context"
	"fmt"
	"iter"
	"log"
	"math/rand"
	"strings"

	"jiangfengwhu/nagi-bot-go/config"

	"github.com/google/uuid"
	"google.golang.org/genai"
)

// StreamData 流数据结构，包含流和已完成的responses
type StreamData struct {
	Stream             iter.Seq2[*genai.GenerateContentResponse, error]
	CompletedResponses []*genai.GenerateContentResponse
}

// LLMService LLM服务结构 - 极简设计
type LLMService struct {
	streams map[string]*StreamData
	baseURL string
	apiKeys string
}

// NewLLMService 创建新的LLM服务实例
func NewLLMService(config *config.Config) *LLMService {
	service := &LLMService{
		streams: make(map[string]*StreamData),
		baseURL: config.LLM.BaseURL,
		apiKeys: config.LLM.APIKeys,
	}

	return service
}

func (s *LLMService) getApiKey() string {
	apiKeys := strings.Split(s.apiKeys, ",")
	return apiKeys[rand.Intn(len(apiKeys))]
}

// Chat 发起对话并生成流，返回流ID
func (s *LLMService) Chat(prompt string, model string) (string, error) {
	if model == "" {
		model = "gemini-2.5-flash"
	}

	// 创建内容
	contents := genai.Text(prompt)

	// 创建生成内容的配置
	config := &genai.GenerateContentConfig{
		Tools: []*genai.Tool{
			{
				FunctionDeclarations: []*genai.FunctionDeclaration{
					ToolsDescMap[ToolGenerateImage],
				},
			},
		},
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:      s.getApiKey(),
		HTTPOptions: genai.HTTPOptions{BaseURL: s.baseURL},
	})

	if err != nil {
		return "", fmt.Errorf("创建LLMClient失败: %v", err)
	}

	// 创建生成内容的流
	stream := client.Models.GenerateContentStream(ctx, model, contents, config)

	// 生成唯一ID
	id := uuid.New().String()

	// 创建StreamData并存储
	streamData := &StreamData{
		Stream:             stream,
		CompletedResponses: make([]*genai.GenerateContentResponse, 0),
	}

	s.streams[id] = streamData

	log.Printf("创建聊天流，ID: %s, 模型: %s", id, model)
	return id, nil
}

// SSE 根据ID获取对应的流数据
func (s *LLMService) SSE(id string) (*StreamData, error) {
	streamData, exists := s.streams[id]
	if !exists {
		return nil, fmt.Errorf("未找到ID为 %s 的流", id)
	}

	return streamData, nil
}

// DeleteStream 删除指定的流
func (s *LLMService) DeleteStream(id string) error {
	_, exists := s.streams[id]
	if !exists {
		return fmt.Errorf("未找到ID为 %s 的流", id)
	}

	delete(s.streams, id)
	log.Printf("删除流: %s", id)
	return nil
}

// ListStreams 列出所有流ID
func (s *LLMService) ListStreams() []string {
	var streamIDs []string
	for id := range s.streams {
		streamIDs = append(streamIDs, id)
	}

	return streamIDs
}

// Close 关闭LLM服务，清空所有流
func (s *LLMService) Close() error {
	// 清空streams map
	s.streams = make(map[string]*StreamData)

	log.Println("LLM服务已关闭，已清空所有流")
	return nil
}
