package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"google.golang.org/genai"
)

func (s *LLMService) GenerateImage(prompt string) ([]byte, int32, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:      s.getApiKey(),
		HTTPOptions: genai.HTTPOptions{BaseURL: s.baseURL},
	})

	if err != nil {
		return nil, 0, fmt.Errorf("创建LLMClient失败: %v", err)
	}

	config := &genai.GenerateContentConfig{
		ResponseModalities: []string{"TEXT", "IMAGE"},
		SafetySettings:     TextSafetySettings,
	}
	result, err := client.Models.GenerateContent(ctx, "gemini-2.0-flash-preview-image-generation", genai.Text(prompt), config)

	if err != nil {
		return nil, 0, fmt.Errorf("生成图片失败: %v", err)
	}

	for _, part := range result.Candidates[0].Content.Parts {
		if part.InlineData != nil {
			return part.InlineData.Data, result.UsageMetadata.TotalTokenCount, nil
		}
	}
	return nil, 0, fmt.Errorf("生成图片失败")
}

func (s *LLMService) GetTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func (s *LLMService) GoogleSearch(prompt string) (string, error) {
	url := "https://www.googleapis.com/customsearch/v1?key=" + s.getGoogleSearchApiKey() + "&cx=92240cc770b9e442b&q=" + url.QueryEscape(prompt)

	response, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("获取谷歌搜索结果失败: %v", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("读取谷歌搜索结果失败: %v", err)
	}

	// 解析JSON响应
	var searchResponse GoogleSearchResponse
	if err := json.Unmarshal(body, &searchResponse); err != nil {
		return "", fmt.Errorf("解析搜索结果失败: %v", err)
	}

	// 整合有用信息
	return formatSearchResults(&searchResponse), nil
}
