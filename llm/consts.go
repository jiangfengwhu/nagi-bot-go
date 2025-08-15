package llm

import "google.golang.org/genai"

// tool枚举
type ToolEnum string

const (
	ToolGenerateImage ToolEnum = "generate_image"
)

var ToolsDescMap = map[ToolEnum]*genai.FunctionDeclaration{
	ToolGenerateImage: {
		Name:        string(ToolGenerateImage),
		Description: "生成图片",
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"prompt": {
					Type:        genai.TypeString,
					Description: "图片描述",
				},
			},
			Required: []string{"prompt"},
		},
	},
}
