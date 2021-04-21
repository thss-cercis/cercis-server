package user

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thss-cercis/cercis-server/api"
	"github.com/thss-cercis/cercis-server/db"
	users "github.com/thss-cercis/cercis-server/db/user"
	"github.com/thss-cercis/cercis-server/middleware"
)

// Current 当前用户
func Current(c *fiber.Ctx) error {
	sess := middleware.GetSession(c)

	userId, ok := sess.Get("user_id").(int)
	if ok {
		user, err := users.GetUserByID(db.GetDB(), userId)
		if err == nil {
			return c.JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess, Payload: user})
		}
	}
	return c.Status(fiber.StatusUnauthorized).JSON(api.BaseRes{Code: api.CodeNotLogin, Msg: "未登录"})
}
