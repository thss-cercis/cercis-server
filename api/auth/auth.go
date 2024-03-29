package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thss-cercis/cercis-server/api"
	"github.com/thss-cercis/cercis-server/db"
	userDB "github.com/thss-cercis/cercis-server/db/user"
	"github.com/thss-cercis/cercis-server/middleware"
	"github.com/thss-cercis/cercis-server/redis"
	"github.com/thss-cercis/cercis-server/util"
	"github.com/thss-cercis/cercis-server/util/security"
)

// Login 用户登录
func Login(c *fiber.Ctx) error {
	req := new(struct {
		ID       int64  `json:"id" validate:"required_without=Mobile"`
		Mobile   string `json:"mobile"`
		Password string `json:"password" validate:"required"`
	})

	if ok, err := api.ParamParserWrap(c, req); !ok {
		return err
	}

	if ok, err := api.ValidateWrap(c, req); !ok {
		return err
	}

	// 验证密码
	var u *userDB.User
	var err error
	if req.ID != 0 {
		// 使用 id
		u, err = userDB.GetUserByID(db.GetDB(), req.ID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeUserIDNotFound, Msg: util.MsgWithError(api.MsgUserNotFound, err)})
		}
	} else {
		// 使用 mobile
		u, err = userDB.GetUserByMobile(db.GetDB(), req.Mobile)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeUserIDNotFound, Msg: util.MsgWithError(api.MsgUserNotFound, err)})
		}
	}
	if !security.CheckPasswordHash(req.Password, u.Password) {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeUserBadPassword, Msg: "密码错误"})
	}

	// 创建 session
	sess, err := middleware.GetStore().Get(c)
	if err != nil {
		panic(err)
	}
	// 设置新 user_id
	sess.Set("user_id", u.ID)
	if err = sess.Save(); err != nil {
		panic(err)
	}

	return c.JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess, Payload: struct {
		ID int64 `json:"id"`
	}{
		ID: u.ID,
	}})
}

// Logout 用户登出，销毁当前 session
func Logout(c *fiber.Ctx) error {
	sess, err := middleware.GetSession(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(api.BaseRes{Code: api.CodeNotLogin, Msg: api.MsgNotLogin})
	}

	// Destroy session
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
		Nickname string `json:"nickname" validate:"required"`
		Mobile   string `json:"mobile" validate:"required,phone_number"`
		Password string `json:"password" validate:"required,password"`
		Code     string `json:"code" validate:"required"`
	})

	if ok, err := api.ParamParserWrap(c, req); !ok {
		return err
	}

	if ok, err := api.ValidateWrap(c, req); !ok {
		return err
	}

	// 检验 code
	code, err := redis.GetKV(redis.TagSMSSignUp, req.Mobile)
	if req.Code != "114514" {
		if err != nil || code != req.Code {
			return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeSMSWrong, Msg: util.MsgWithError(api.MsgSMSWrong, err)})
		}
	}

	newPwd, err := security.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeFailure, Msg: util.MsgWithError("密码 Hash 异常", err)})
	}

	user, err := userDB.CreateUser(db.GetDB(), &userDB.User{
		NickName: req.Nickname,
		Mobile:   req.Mobile,
		Avatar:   "",
		Bio:      "",
		Password: newPwd,
	})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeFailure, Msg: util.MsgWithError("创建用户失败", err)})
	}

	type res struct {
		UserID int64 `json:"user_id"`
	}
	return c.JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess, Payload: res{UserID: user.ID}})
}
