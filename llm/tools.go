package llm

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

func (s *LLMService) GenerateImage(prompt string) ([]byte, error) {
	fmt.Println("GenerateImage", prompt)
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:      s.getApiKey(),
		HTTPOptions: genai.HTTPOptions{BaseURL: s.baseURL},
	})

	if err != nil {
		return nil, fmt.Errorf("创建LLMClient失败: %v", err)
	}

	config := &genai.GenerateContentConfig{
		ResponseModalities: []string{"TEXT", "IMAGE"},
	}
	result, err := client.Models.GenerateContent(ctx, "gemini-2.0-flash-preview-image-generation", genai.Text(prompt), config)

	if err != nil {
		return nil, fmt.Errorf("生成图片失败: %v", err)
	}

	for _, part := range result.Candidates[0].Content.Parts {
		if part.InlineData != nil {
			return part.InlineData.Data, nil
		}
	}

	return nil, fmt.Errorf("生成图片失败")
}

func (s *LLMService) GetTool(tool ToolEnum) func(prompt string) ([]byte, error) {
	switch tool {
	case ToolGenerateImage:
		return s.GenerateImage
	default:
		return nil
	}
}
