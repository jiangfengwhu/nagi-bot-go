package llm

import "google.golang.org/genai"

var TextSafetySettings = []*genai.SafetySetting{
	{
		Category:  genai.HarmCategoryHateSpeech,
		Threshold: genai.HarmBlockThresholdOff,
	},
	{
		Category:  genai.HarmCategoryDangerousContent,
		Threshold: genai.HarmBlockThresholdOff,
	},
	{
		Category:  genai.HarmCategorySexuallyExplicit,
		Threshold: genai.HarmBlockThresholdOff,
	},

	{
		Category:  genai.HarmCategoryHarassment,
		Threshold: genai.HarmBlockThresholdOff,
	},
}

// tool枚举
type ToolEnum string

const (
	ToolGenerateImage ToolEnum = "generate_image"
	ToolGetTime       ToolEnum = "get_time"
	ToolGoogleSearch  ToolEnum = "google_search"
	ToolCreatePlayer  ToolEnum = "create_player"
)

var ToolsDescMap = map[ToolEnum]*genai.FunctionDeclaration{
	ToolGenerateImage: {
		Name:        string(ToolGenerateImage),
		Description: "根据用户提供的描述，加以完善和补充，使其具有艺术气息，仅在需要生成图片的时候调用",
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"prompt": {
					Type:        genai.TypeString,
					Description: "根据用户提供的描述完善之后的提示词",
				},
			},
			Required: []string{"prompt"},
		},
	},
	ToolGetTime: {
		Name:        string(ToolGetTime),
		Description: "获取当前时间",
	},
	ToolGoogleSearch: {
		Name:        string(ToolGoogleSearch),
		Description: "根据搜索的关键词，进行谷歌搜索获取实时信息，并返回搜索结果",
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"prompt": {
					Type:        genai.TypeString,
					Description: "搜索词",
				},
			},
			Required: []string{"prompt"},
		},
	},
	ToolCreatePlayer: {
		Name:        string(ToolCreatePlayer),
		Description: "根据用户提供的名称创建新修仙者角色。创建时需要生成完整的修仙者属性，包括：基础信（姓名、年龄、出生地）。灵根属性（根据姓名和时辰生成）、基础属性（根据灵根生成生命值、法力值、攻击防御等）、修炼属性（悟性）。请根据用户名、当前时间，结合周易八卦、五行相生相克理论来生成合理的初始属性。",
		Parameters: &genai.Schema{
			Type:       genai.TypeObject,
			Properties: map[string]*genai.Schema{},
			Required:   []string{"player_name", "spiritual_roots", "physique", "comprehension", "luck", "spirit_sense", "max_hp", "max_mp", "attack", "defense", "speed", "lifespan", "background_story"},
		},
	},
}
