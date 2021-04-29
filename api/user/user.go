package user

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thss-cercis/cercis-server/api"
	"github.com/thss-cercis/cercis-server/db"
	userDB "github.com/thss-cercis/cercis-server/db/user"
	"github.com/thss-cercis/cercis-server/middleware"
	"github.com/thss-cercis/cercis-server/redis"
	"github.com/thss-cercis/cercis-server/util"
	"github.com/thss-cercis/cercis-server/util/security"
	"github.com/thss-cercis/cercis-server/util/validator"
)

// CurrentUser 查询当前用户信息的 api
func CurrentUser(c *fiber.Ctx) error {
	userId, ok := middleware.GetUserIDFromSession(c)
	if ok {
		user, err := userDB.GetUserByID(db.GetDB(), userId)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeUserIDNotFound, Msg: util.MsgWithError(api.MsgUserNotFound, err)})
		}
		return c.JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess, Payload: user})
	}
	return c.Status(fiber.StatusUnauthorized).JSON(api.BaseRes{Code: api.CodeNotLogin, Msg: api.MsgNotLogin})
}

// ModifyUser 修改用户个人信息的 api
func ModifyUser(c *fiber.Ctx) error {
	req := new(struct {
		NickName string `json:"nickname" validate:"omitempty"`
		Email    string `json:"email" validate:"omitempty,email"`
		Mobile   string `json:"mobile" validate:"omitempty,phone_number"`
		Avatar   string `json:"avatar" validate:"omitempty,url"`
		Bio      string `json:"bio"`
	})

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeBadParam, Msg: util.MsgWithError(api.MsgWrongParam, err)})
	}

	if err := validator.Validate(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeBadParam, Msg: util.MsgWithError(api.MsgWrongParam, err)})
	}

	userId, ok := middleware.GetUserIDFromSession(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(api.BaseRes{Code: api.CodeNotLogin, Msg: api.MsgNotLogin})
	}

	user, err := userDB.GetUserByID(db.GetDB(), userId)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(api.BaseRes{Code: api.CodeUserIDNotFound, Msg: api.MsgUserNotFound})
	}

	// TODO: 暂时不做更改内容的校验
	if req.NickName != "" {
		user.NickName = req.NickName
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Mobile != "" {
		user.Mobile = req.Mobile
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Bio != "" {
		user.Bio = req.Bio
	}

	err = user.UpdateTo(db.GetDB())
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(api.BaseRes{Code: api.CodeFailure, Msg: util.MsgWithError("更新用户信息失败", err)})
	}

	rep := struct {
		NickName string `json:"nickname"`
		Email    string `json:"email"`
		Mobile   string `json:"mobile"`
		Avatar   string `json:"avatar"`
		Bio      string `json:"bio"`
	}{
		NickName: user.NickName,
		Email:    user.Email,
		Mobile:   user.Mobile,
		Avatar:   user.Avatar,
		Bio:      user.Bio,
	}

	return c.JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess, Payload: rep})
}

// ModifyPassword 修改用户密码
func ModifyPassword(c *fiber.Ctx) error {
	req := new(struct {
		OldPwd string `json:"old_pwd" validate:"required"`
		NewPwd string `json:"new_pwd" validate:"required,password"`
	})

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeBadParam, Msg: util.MsgWithError(api.MsgWrongParam, err)})
	}

	if err := validator.Validate(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeBadParam, Msg: util.MsgWithError(api.MsgWrongParam, err)})
	}

	userId, ok := middleware.GetUserIDFromSession(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(api.BaseRes{Code: api.CodeNotLogin, Msg: api.MsgNotLogin})
	}

	user, err := userDB.GetUserByID(db.GetDB(), userId)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(api.BaseRes{Code: api.CodeUserIDNotFound, Msg: util.MsgWithError(api.MsgUserNotFound, err)})
	}

	if !security.CheckPasswordHash(req.OldPwd, user.Password) {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeUserBadPassword, Msg: "密码错误"})
	}

	// 修改密码
	user.Password, err = security.HashPassword(req.NewPwd)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeFailure, Msg: util.MsgWithError("修改密码失败", err)})
	}

	err = user.UpdateTo(db.GetDB())
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeFailure, Msg: util.MsgWithError(api.MsgUnknown, err)})
	}

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess})
}

// RecoverPassword 找回用户密码
func RecoverPassword(c *fiber.Ctx) error {
	req := new(struct {
		Mobile string `json:"mobile" validate:"required,phone_number"`
		NewPwd string `json:"new_pwd" validate:"required,password"`
		Code   string `json:"code" validate:"required"`
	})

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeBadParam, Msg: util.MsgWithError(api.MsgWrongParam, err)})
	}

	if err := validator.Validate(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeBadParam, Msg: util.MsgWithError(api.MsgWrongParam, err)})
	}

	// 检验 code
	code, err := redis.GetKV(redis.TagSMSRecover, req.Mobile)
	if err != nil || code != req.Code {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeSMSWrong, Msg: util.MsgWithError(api.MsgSMSWrong, err)})
	}

	newPwd, err := security.HashPassword(req.NewPwd)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeBadParam, Msg: util.MsgWithError("密码 Hash 异常", err)})
	}

	user, err := userDB.GetUserByMobile(db.GetDB(), req.Mobile)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(api.BaseRes{Code: api.CodeUserIDNotFound, Msg: util.MsgWithError(api.MsgUserNotFound, err)})
	}

	// 更改密码
	user.Password = newPwd

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess})
}
