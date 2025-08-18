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
}
