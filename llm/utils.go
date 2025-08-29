package llm

import (
	"fmt"
	"strings"

	"jiangfengwhu/nagi-bot-go/database"
)

type ToolCallArgs struct {
	DB     *database.DB
	UserID int
	Args   *map[string]any
}

// GoogleSearchResult 表示Google搜索的单个结果项
type GoogleSearchResult struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
}

// GoogleSearchResponse 表示Google搜索API的响应
type GoogleSearchResponse struct {
	Items             []GoogleSearchResult `json:"items"`
	SearchInformation struct {
		SearchTime   float64 `json:"searchTime"`
		TotalResults string  `json:"totalResults"`
	} `json:"searchInformation"`
}

// formatSearchResults 格式化搜索结果，只返回关键信息
func formatSearchResults(response *GoogleSearchResponse) string {
	if len(response.Items) == 0 {
		return "未找到相关搜索结果"
	}

	var result strings.Builder

	// 添加搜索统计信息
	result.WriteString(fmt.Sprintf("搜索结果统计: 找到约%s条结果（用时%.2f秒）\n\n",
		response.SearchInformation.TotalResults,
		response.SearchInformation.SearchTime))

	for i, item := range response.Items {

		result.WriteString(fmt.Sprintf("【结果%d】\n", i+1))
		result.WriteString(fmt.Sprintf("标题: %s\n", item.Title))
		result.WriteString(fmt.Sprintf("来源: %s\n", item.Link))
		result.WriteString(fmt.Sprintf("摘要: %s\n", item.Snippet))

	}

	return result.String()
}
