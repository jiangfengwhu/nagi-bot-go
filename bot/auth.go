package bot

import (
	"context"
	"jiangfengwhu/nagi-bot-go/database"
	"strings"
	"time"

	tele "gopkg.in/telebot.v4"
)

func Auth(db *database.DB, b *tele.Bot) tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			userId := c.Sender().ID
			user, err := db.GetUser(context.Background(), userId)
			if err != nil {
				return c.Send("获取用户信息失败" + err.Error())
			}
			if user == nil {
				user = &database.User{
					TgId:                userId,
					Username:            c.Sender().Username,
					CreatedAt:           time.Now(),
					TotalRechargedToken: 10000000,
					TotalUsedToken:      0,
					SystemPrompt:        "",
				}
				db.CreateUser(context.Background(), user)
				c.Send("欢迎来到《凡尘仙途》。您已获得1000000个灵石，请使用/reg 角色名 注册之后进行游戏")
			}
			c.Set("db_user", user)
			if c.Message().Chat.Type != "private" {
				isMention := false
				msg := c.Message()
				for _, entity := range msg.Entities {
					if entity.Type == tele.EntityMention { // 普通 @xxx
						mention := msg.Text[entity.Offset : entity.Offset+entity.Length]
						// 判断是不是提到自己
						if mention == "@"+b.Me.Username {
							isMention = true
							msg.Text = strings.Replace(msg.Text, mention, "", 1)
							break
						}
					}
				}
				if !isMention {
					return nil
				}
			}
			return next(c)
		}
	}
}
