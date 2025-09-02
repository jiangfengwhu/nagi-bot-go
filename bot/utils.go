package bot

import (
	"fmt"
	"regexp"
	"strings"

	"jiangfengwhu/nagi-bot-go/database"
)

func escapeTelegramMarkdownV2(text string) string {
	telegramSpecialChars := "_*[]()~`>#+-=|{}.!"
	for _, char := range telegramSpecialChars {
		text = strings.ReplaceAll(text, string(char), "\\"+string(char))
	}
	return text
}

func escapeTelegramMarkdownV2CodeBlocks(text string) string {
	backslashRegex := regexp.MustCompile(`\\([^ ]|\\|$)`)

	text = backslashRegex.ReplaceAllStringFunc(text, func(match string) string {
		if len(match) > 1 {
			return fmt.Sprintf("\\\\%s", match[1:])
		}

		return "\\\\"
	})

	text = escapeTelegramMarkdownV2(text)

	return text
}

func wrapLinksInMarkdown(text string) string {
	re := regexp.MustCompile(`\bhttps?://[a-zA-Z0-9-.]+(.[a-zA-Z]{2,})(:[0-9]{1,5})?(/[a-zA-Z0-9-_.~%/?#=&+]*)?\b`)

	wrappedText := re.ReplaceAllStringFunc(text, func(link string) string {
		if strings.Contains(link, "](http") || strings.Contains(link, "[http") {
			return link
		}
		return fmt.Sprintf("[%s](%s)", link, link)
	})

	return wrappedText
}

func removePlaceholders(text string) string {
	re := regexp.MustCompile(`ELEMENTPLACEHOLDER\d+|CODEBLOCKPLACEHOLDER\d+|INLINECODEPLACEHOLDER\d+|LINKPLACEHOLDER\d`)

	result := re.ReplaceAllString(text, "")

	return result
}

func addPlaceholders(text string, elements map[string]string) (string, map[string]string) {
	codeBlockRegex := regexp.MustCompile("```(?s:.*?)```")
	inlineCodeRegex := regexp.MustCompile("`[^`]+`")
	linkRegex := regexp.MustCompile(`\[[^\]]+\]\([^\)]+\)`)

	elementsRegex := regexp.MustCompile(fmt.Sprintf("%s|%s|%s", codeBlockRegex.String(), inlineCodeRegex.String(), linkRegex.String()))

	matches := elementsRegex.FindAllString(text, -1)
	for i, match := range matches {
		placeholder := fmt.Sprintf("ELEMENTPLACEHOLDER%d", len(elements)+i)

		// Определяем тип элемента и добавляем соответствующий маркер
		switch {
		case codeBlockRegex.MatchString(match):
			placeholder = fmt.Sprintf("CODEBLOCKPLACEHOLDER%d", len(elements)+i)
		case inlineCodeRegex.MatchString(match):
			placeholder = fmt.Sprintf("INLINECODEPLACEHOLDER%d", len(elements)+i)
		case linkRegex.MatchString(match):
			placeholder = fmt.Sprintf("LINKPLACEHOLDER%d", len(elements)+i)
		}

		elements[placeholder] = match
		text = strings.Replace(text, match, placeholder, 1)
	}
	return text, elements
}

func processPlaceholders(md string, elements map[string]string) string {
	for placeholder, element := range elements {
		if strings.HasPrefix(placeholder, "LINKPLACEHOLDER") {
			parts := strings.SplitN(element, "](", 2)
			linkText := parts[0][1:]
			linkURL := parts[1][:len(parts[1])-1]

			linkText = escapeTelegramMarkdownV2(linkText)
			element = "[" + linkText + "](" + linkURL + ")"
		}

		if strings.HasPrefix(placeholder, "CODEBLOCKPLACEHOLDER") {
			re := regexp.MustCompile("(?s)```([a-zA-Z]*)\\n(.*?)\\n```")

			element = re.ReplaceAllStringFunc(element, func(block string) string {
				matches := re.FindStringSubmatch(block)
				if len(matches) > 2 {
					language := matches[1]
					return fmt.Sprintf("```%s\n%s\n```", language, escapeTelegramMarkdownV2CodeBlocks(matches[2]))
				}
				return block
			})
		}

		if strings.HasPrefix(placeholder, "INLINECODEPLACEHOLDER") {
			re := regexp.MustCompile("`([^`]+)`")

			element = re.ReplaceAllStringFunc(element, func(block string) string {
				matches := re.FindStringSubmatch(block)
				if len(matches) > 1 {
					return fmt.Sprintf("`%s`", escapeTelegramMarkdownV2CodeBlocks(matches[1]))
				}
				return block
			})
		}

		md = strings.Replace(md, placeholder, element, 1)
	}

	return md
}

func processStyles(md string) string {
	replacements := []struct {
		regex       *regexp.Regexp
		replacement string
	}{
		// bold text
		{regexp.MustCompile(`\\\*\\\*(.*?)\\\*\\\*`), `*$1*`},

		// italic text
		{regexp.MustCompile(`\\\*(.*?)\\\*`), `_${1}_`},
		{regexp.MustCompile(`\\_(.*?)\\_`), `_${1}_`},

		// strikethrough text
		{regexp.MustCompile(`\\~\\~(.*?)\\~\\~`), `~$1~`},
		{regexp.MustCompile(`\\~(.*?)\\~`), `~$1~`},
	}

	for _, r := range replacements {
		md = r.regex.ReplaceAllString(md, r.replacement)
	}

	return md
}

func ConvertMarkdownToTelegramMarkdownV2(md string) string {
	elements := make(map[string]string)

	md = removePlaceholders(md)
	md, elements = addPlaceholders(md, elements)

	md = wrapLinksInMarkdown(md)
	md, elements = addPlaceholders(md, elements)

	md = escapeTelegramMarkdownV2(md)

	md = processStyles(md)

	md = processPlaceholders(md, elements)

	return md
}

func formatPlayerInfo(player *database.CharacterStats) string {

	spiritualRoots := ""
	for _, root := range *player.SpiritualRoots {
		spiritualRoots += fmt.Sprintf("- %s: %d\n", root.RootName, root.Afinity)
	}
	// telegram markdown v2 format，emoji开头
	return fmt.Sprintf("👤 角色名称: %s\n", player.Name) + "\n" +
		fmt.Sprintf("🌟 角色等级: %d\n", player.RealmLevel) + "\n" +
		fmt.Sprintf("🌿 角色境界: %s\n", player.Realm) + "\n" +
		fmt.Sprintf("🌱 角色灵根: \n%s\n", spiritualRoots) + "\n" +
		fmt.Sprintf("🔮 角色神识: %d\n", player.SpiritSense) + "\n" +
		fmt.Sprintf("💪 角色根骨: %d\n", player.Physique) + "\n" +
		fmt.Sprintf("👹 角色煞气: %d\n", player.DemonicAura) + "\n" +
		fmt.Sprintf("👺 角色道号: %s\n", player.TaoistName) + "\n" +
		fmt.Sprintf("💜 角色生命值: %d\n", player.HP) + "\n" +
		fmt.Sprintf("💚 角色法力值: %d\n", player.MP) + "\n" +
		fmt.Sprintf("💪 角色攻击力: %d\n", player.Attack) + "\n" +
		fmt.Sprintf("🛡️ 角色防御力: %d\n", player.Defense) + "\n" +
		fmt.Sprintf("🏃 角色速度: %d\n", player.Speed) + "\n" +
		fmt.Sprintf("🍀 角色幸运值: %d\n", player.Luck) + "\n" +
		fmt.Sprintf("💪 角色修炼经验: %d\n", player.Experience) + "\n" +
		fmt.Sprintf("🤔 角色悟性: %d\n", player.Comprehension) + "\n" +
		fmt.Sprintf("👵 角色年龄: %d\n", player.Age) + "\n" +
		fmt.Sprintf("👴 角色寿命: %d\n", player.Lifespan) + "\n" +
		fmt.Sprintf("🏠 角色位置: %s\n", player.Location) + "\n" +
		fmt.Sprintf("👨‍🦰 角色状态: %s\n", player.Status) + "\n" +
		fmt.Sprintf("📚 角色成长经历: %s\n", player.Stories)
}

func formatInventoryInfo(inventory []*database.InventoryItem) string {
	inventoryInfo := ""
	for _, item := range inventory {
		inventoryInfo += fmt.Sprintf("💼 背包物品: %s\n", item.ItemName) + "\n" +
			fmt.Sprintf("🔍 物品数量: %d\n", item.Quantity) + "\n" +
			fmt.Sprintf("🎯 物品类型: %s\n", item.ItemType) + "\n" +
			fmt.Sprintf("🔮 物品品质: %s\n", item.Quality) + "\n" +
			fmt.Sprintf("🔍 物品等级: %d\n", item.Level) + "\n" +
			fmt.Sprintf("🔍 物品属性: %s\n", item.Properties) + "\n" +
			fmt.Sprintf("📝 物品描述: %s\n", item.Description)
		inventoryInfo += "\n"
	}
	return inventoryInfo
}
