package auth

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/thss-cercis/cercis-server/api"
	"github.com/thss-cercis/cercis-server/db"
	"github.com/thss-cercis/cercis-server/db/user"
	"github.com/thss-cercis/cercis-server/middleware"
	"github.com/thss-cercis/cercis-server/util/security"
)

// Login 用户登录
func Login(c *fiber.Ctx) error {
	req := new(struct {
		ID       string `json:"id"`
		Password string `json:"password"`
	})

	if err := c.BodyParser(req); err != nil {
		return err
	}

	// 验证密码
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeBadParam, Msg: "参数 id 有误"})
	}
	u, err := user.GetUserByID(db.GetDB(), id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeUserIDNotFound, Msg: "此 id 不存在"})
	}
	if !security.CheckPasswordHash(req.Password, u.Password) {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeUserBadPassword, Msg: "密码错误"})
	}

	// 创建 session
	sess, err := middleware.GetStore().Get(c)
	if err != nil {
		panic(err)
	}
	// 已经登陆
	_, ok := sess.Get("user_id").(int)
	if ok {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeUserAlreadyLogin, Msg: "已经登陆"})
	}
	// 设置新 user_id
	sess.Set("user_id", u.ID)
	if err = sess.Save(); err != nil {
		panic(err)
	}

	return c.JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess})
}

// Logout 用户登出，销毁当前 session
func Logout(c *fiber.Ctx) error {
	sess := middleware.GetSession(c)

	// Destry session
	if err := sess.Destroy(); err != nil {
		panic(err)
	}

	// save session
	if err := sess.Save(); err != nil {
		panic(err)
	}

	return c.JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess})
}

// Signup 用户注册
func Signup(c *fiber.Ctx) error {
	req := new(struct {
		Nickname string `json:"nickname"`
		Email    string `json:"email,omitempty"`
		Mobile   string `json:"mobile"`
		Password string `json:"password"`
	})

	if err := c.BodyParser(req); err != nil {
		return err
	}

	newPwd, err := security.HashPassword(req.Password)
	if err != nil {
		return err
	}

	user, err := user.CreateUser(db.GetDB(), &user.User{
		NickName: req.Nickname,
		Email:    req.Email,
		Mobile:   req.Mobile,
		Avatar:   "",
		Bio:      "",
		Password: newPwd,
	})
	if err != nil {
		return err
	}

	type p struct {
		UserID string `json:"user_id"`
	}
	return c.JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess, Payload: p{
		UserID: strconv.Itoa(user.ID),
	}})
}
