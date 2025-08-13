package bot

import (
	"context"
	"jiangfengwhu/nagi-bot-go/database"
	"time"

	tele "gopkg.in/telebot.v4"
)

func Auth(db *database.DB) tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			userId := c.Sender().ID
			user, err := db.GetUser(context.Background(), userId)
			if err != nil {
				return c.Send("获取用户信息失败")
			}
			if user == nil {
				user = &database.User{
					TgId:                userId,
					Username:            c.Sender().Username,
					CreatedAt:           time.Now(),
					TotalRechargedToken: 10000000,
					TotalUsedToken:      0,
				}
				db.CreateUser(context.Background(), user)
				c.Send("注册成功，欢迎使用 Nagi Bot！您已获得10000000个token")
			}
			c.Set("db_user", user)
			return next(c)
		}
	}
}
