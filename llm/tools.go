package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
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

func UpdatePlayer(db *database.DB, userID int, args map[string]any) (string, error) {
	ctx := context.Background()

	// 解析JSON参数到部分更新结构体
	var updateParams database.CharacterStatsUpdate
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return "", fmt.Errorf("解析更新参数失败: %v", err)
	}
	if err := json.Unmarshal(argsJSON, &updateParams); err != nil {
		return "", fmt.Errorf("解析更新参数失败: %v", err)
	}

	// 设置用户ID
	updateParams.UserID = userID

	// 调用部分更新方法
	if err := db.UpdateCharacterStatsPartial(ctx, &updateParams); err != nil {
		return "", fmt.Errorf("更新玩家信息失败: %v", err)
	}

	// 构建更新成功的消息
	updateFields := []string{}
	if updateParams.Realm != nil {
		updateFields = append(updateFields, fmt.Sprintf("境界: %s", *updateParams.Realm))
	}
	if updateParams.RealmLevel != nil {
		updateFields = append(updateFields, fmt.Sprintf("境界等级: %d", *updateParams.RealmLevel))
	}
	if updateParams.SpiritSense != nil {
		updateFields = append(updateFields, fmt.Sprintf("神识: %d", *updateParams.SpiritSense))
	}
	if updateParams.Physique != nil {
		updateFields = append(updateFields, fmt.Sprintf("根骨: %d", *updateParams.Physique))
	}
	if updateParams.DemonicAura != nil {
		updateFields = append(updateFields, fmt.Sprintf("煞气: %d", *updateParams.DemonicAura))
	}
	if updateParams.Attack != nil {
		updateFields = append(updateFields, fmt.Sprintf("攻击力: %d", *updateParams.Attack))
	}
	if updateParams.Defense != nil {
		updateFields = append(updateFields, fmt.Sprintf("防御力: %d", *updateParams.Defense))
	}
	if updateParams.Speed != nil {
		updateFields = append(updateFields, fmt.Sprintf("速度: %d", *updateParams.Speed))
	}
	if updateParams.Luck != nil {
		updateFields = append(updateFields, fmt.Sprintf("幸运值: %d", *updateParams.Luck))
	}
	if updateParams.Comprehension != nil {
		updateFields = append(updateFields, fmt.Sprintf("悟性: %d", *updateParams.Comprehension))
	}
	if updateParams.Lifespan != nil {
		updateFields = append(updateFields, fmt.Sprintf("寿命: %d", *updateParams.Lifespan))
	}
	if updateParams.Location != nil {
		updateFields = append(updateFields, fmt.Sprintf("位置: %s", *updateParams.Location))
	}
	if updateParams.Status != nil {
		updateFields = append(updateFields, fmt.Sprintf("状态: %s", *updateParams.Status))
	}
	if updateParams.Stories != nil {
		updateFields = append(updateFields, fmt.Sprintf("新增经历: %s", *updateParams.Stories))
	}

	if len(updateFields) == 0 {
		return "没有需要更新的字段", nil
	}

	return fmt.Sprintf("玩家信息更新成功，已更新: %s", strings.Join(updateFields, ", ")), nil
}

func UpdateInventory(db *database.DB, userID int, args map[string]any) (string, error) {
	ctx := context.Background()

	// 解析JSON参数到部分更新结构体
	var updateParams []*database.InventoryItem
	argsJSON, err := json.Marshal(args["items"])
	fmt.Println("argsJSON", string(argsJSON))
	if err != nil {
		return "", fmt.Errorf("解析更新参数失败: %v", err)
	}
	if err := json.Unmarshal(argsJSON, &updateParams); err != nil {
		return "", fmt.Errorf("解析更新参数失败: %v", err)
	}

	// 调用部分更新方法
	if err := db.UpdateInventory(ctx, userID, updateParams); err != nil {
		return "", fmt.Errorf("更新背包物品失败: %v", err)
	}

	return "背包物品更新成功", nil
}
