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
	ToolGenerateImage   ToolEnum = "generate_image"
	ToolGetTime         ToolEnum = "get_time"
	ToolGoogleSearch    ToolEnum = "google_search"
	ToolUpdatePlayer    ToolEnum = "update_player"
	ToolUpdateInventory ToolEnum = "update_inventory"
)

var ToolsDescMap = map[ToolEnum]*genai.FunctionDeclaration{
	ToolGenerateImage: {
		Name:        string(ToolGenerateImage),
		Description: "根据法器或宝物描述，加以完善和补充，使其具有仙风道骨，生成对应的图片",
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"prompt": {
					Type:        genai.TypeString,
					Description: "法器或宝物的详细描述",
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
		Description: "遇到不理解的概念，知识或者需要搜索的场景，可以调用此工具进行搜索，并返回搜索结果",
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"prompt": {
					Type:        genai.TypeString,
					Description: "搜索内容，比如不理解的概念，知识，或者需要搜索的场景",
				},
			},
			Required: []string{"prompt"},
		},
	},
	ToolUpdateInventory: {
		Name:        string(ToolUpdateInventory),
		Description: "更新玩家背包中的物品",
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"items": {
					Type: genai.TypeArray,
					Items: &genai.Schema{
						Type:        genai.TypeObject,
						Description: "物品列表, 添加时quantity为正数，删除时quantity为负数",
						Properties: map[string]*genai.Schema{
							"item_name": {
								Type:        genai.TypeString,
								Description: "物品的名称",
							},
							"quantity": {
								Type:        genai.TypeInteger,
								Description: "物品的数量",
							},
							"item_type": {
								Type:        genai.TypeString,
								Description: "物品的类型，比如法宝，丹药，符箓等",
							},
							"quality": {
								Type:        genai.TypeString,
								Description: "物品的品质，比如普通，高级，稀有，史诗，传说",
							},
							"level": {
								Type:        genai.TypeInteger,
								Description: "物品的等级，比如1品，珍品，远古，玄天，通天等",
							},
							"properties": {
								Type:        genai.TypeString,
								Description: "物品的属性，比如攻击力，防御力，特殊效果等",
							},
							"description": {
								Type:        genai.TypeString,
								Description: "物品的描述",
							},
							"obtained_from": {
								Type:        genai.TypeString,
								Description: "物品的获取来源",
							},
						},
						Required: []string{"item_name", "quantity", "item_type", "quality", "level", "properties", "description", "obtained_from"},
					},
					Description: "物品列表，每个元素包含物品名称，数量，类型，品质，等级，属性，描述等",
				},
			},
		},
	},
	ToolUpdatePlayer: {
		Name:        string(ToolUpdatePlayer),
		Description: "在玩家经历冒险，战斗，奇遇，机缘等事件后，更新玩家的基础属性，包括：基础属性，修炼属性，境界提升等",
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"realm": {
					Type:        genai.TypeString,
					Description: "境界",
				},
				"realm_level": {
					Type:        genai.TypeInteger,
					Description: "境界等级",
				},
				"spirit_sense": {
					Type:        genai.TypeInteger,
					Description: "神识",
				},
				"physique": {
					Type:        genai.TypeInteger,
					Description: "根骨",
				},
				"demonic_aura": {
					Type:        genai.TypeInteger,
					Description: "煞气",
				},
				"attack": {
					Type:        genai.TypeInteger,
					Description: "攻击力",
				},
				"defense": {
					Type:        genai.TypeInteger,
					Description: "防御力",
				},
				"speed": {
					Type:        genai.TypeInteger,
					Description: "速度",
				},
				"luck": {
					Type:        genai.TypeInteger,
					Description: "幸运值",
				},
				"comprehension": {
					Type:        genai.TypeInteger,
					Description: "悟性",
				},
				"lifespan": {
					Type:        genai.TypeInteger,
					Description: "寿命",
				},
				"location": {
					Type:        genai.TypeString,
					Description: "位置",
				},
				"status": {
					Type:        genai.TypeString,
					Description: "状态",
				},
				"stories": {
					Type:        genai.TypeString,
					Description: "玩家遭遇的重大事件的描述",
				},
			},
			Required: []string{},
		},
	},
}
