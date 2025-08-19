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
}
