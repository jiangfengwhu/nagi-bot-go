package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"jiangfengwhu/nagi-bot-go/database"

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
	url := "https://expensive-dolphin-32.deno.dev/customsearch/v1?key=" + s.getGoogleSearchApiKey() + "&cx=92240cc770b9e442b&q=" + url.QueryEscape(prompt)

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

type SpiritualRoot struct {
	RootName string `json:"root_name"`
	Afinity  int    `json:"affinity"`
}

// CreatePlayerParams 创建玩家的参数结构
type CreatePlayerParams struct {
	PlayerName      string          `json:"player_name"`
	SpiritualRoots  []SpiritualRoot `json:"spiritual_roots"`
	Physique        int             `json:"physique"`
	Comprehension   int             `json:"comprehension"`
	Luck            int             `json:"luck"`
	SpiritSense     int             `json:"spirit_sense"`
	MaxHP           int             `json:"max_hp"`
	MaxMP           int             `json:"max_mp"`
	Attack          int             `json:"attack"`
	Defense         int             `json:"defense"`
	Speed           int             `json:"speed"`
	Lifespan        int             `json:"lifespan"`
	BackgroundStory string          `json:"background_story"`
}

// CreatePlayer 创建新的修仙者角色
func (s *LLMService) CreatePlayer(db *database.DB, userID int, args string) (*database.CharacterStats, error) {
	ctx := context.Background()
	// 检查用户是否已经有角色
	existingStats, err := db.GetCharacterStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("检查现有角色失败: %v", err)
	}
	if existingStats != nil {
		return nil, fmt.Errorf("您已经拥有角色: %s", existingStats.Name)
	}

	// 使用 JSON 序列化进行类型转换
	var params CreatePlayerParams
	err = json.Unmarshal([]byte(args), &params)
	if err != nil {
		return nil, fmt.Errorf("解析参数失败: %v", err)
	}

	// 创建角色属性 - 使用大模型计算的所有属性
	spiritualRoots := database.SpiritualRoots{}
	for _, root := range params.SpiritualRoots {
		spiritualRoots[root.RootName] = root.Afinity
	}
	stats := &database.CharacterStats{
		UserID:         userID,
		Name:           params.PlayerName,
		Realm:          "练气期",
		RealmLevel:     1,
		SpiritualRoots: &spiritualRoots,
		SpiritSense:    params.SpiritSense,
		Physique:       params.Physique,
		DemonicAura:    0,
		TaoistName:     nil, // 练气期没有道号
		HP:             params.MaxHP,
		MaxHP:          params.MaxHP,
		MP:             params.MaxMP,
		MaxMP:          params.MaxMP,
		Attack:         params.Attack,
		Defense:        params.Defense,
		Speed:          params.Speed,
		Luck:           params.Luck,
		Experience:     0,
		Comprehension:  params.Comprehension,
		Age:            1,
		Lifespan:       params.Lifespan,
		Location:       "新手村",
		Status:         "健康",
	}

	// 保存到数据库
	err = db.CreateCharacterStats(ctx, stats)
	if err != nil {
		return nil, fmt.Errorf("创建角色失败: %v", err)
	}

	return stats, nil
}
