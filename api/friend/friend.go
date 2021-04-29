package friend

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thss-cercis/cercis-server/api"
	"github.com/thss-cercis/cercis-server/db"
	"github.com/thss-cercis/cercis-server/db/user"
	"github.com/thss-cercis/cercis-server/middleware"
	"github.com/thss-cercis/cercis-server/util/validator"
)

// GetSendApply 获得自己发送的好友申请
func GetSendApply(c *fiber.Ctx) error {
	userId, ok := middleware.GetUserIDFromSession(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(api.BaseRes{Code: api.CodeNotLogin, Msg: api.MsgNotLogin})
	}

	applies, err := user.GetFriendApplyFromByUserID(db.GetDB(), userId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeFailure, Msg: api.MsgUnknown, Payload: err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess, Payload: struct {
		Applies []user.FriendApply `json:"applies"`
	}{
		Applies: applies,
	}})
}

// GetReceiveApply 获得自己收到的好友申请
func GetReceiveApply(c *fiber.Ctx) error {
	userId, ok := middleware.GetUserIDFromSession(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(api.BaseRes{Code: api.CodeNotLogin, Msg: api.MsgNotLogin})
	}

	applies, err := user.GetFriendApplyToByUserID(db.GetDB(), userId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeFailure, Msg: api.MsgUnknown, Payload: err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess, Payload: struct {
		Applies []user.FriendApply `json:"applies"`
	}{
		Applies: applies,
	}})
}

// SendApply 发送好友申请
func SendApply(c *fiber.Ctx) error {
	// TODO websocket
	req := new(struct {
		ToID int64 `json:"to_id" validate:"required"`
	})

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeBadParam, Msg: api.MsgWrongParam, Payload: err.Error()})
	}

	if err := validator.Validate(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeBadParam, Msg: api.MsgWrongParam, Payload: err})
	}

	userId, ok := middleware.GetUserIDFromSession(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(api.BaseRes{Code: api.CodeNotLogin, Msg: api.MsgNotLogin})
	}

	// 自己不能发给自己
	if userId == req.ToID {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeFailure, Msg: "不允许向自身发送好友请求"})
	}
	_, err := user.CreateFriendApply(db.GetDB(), userId, req.ToID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeFailure, Msg: api.MsgUnknown, Payload: err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess})
}

// AcceptApply 接收好友申请
func AcceptApply(c *fiber.Ctx) error {
	// TODO websocket
	req := new(struct {
		ApplyID int64 `json:"apply_id" validate:"required"`
	})

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeBadParam, Msg: api.MsgWrongParam, Payload: err.Error()})
	}

	if err := validator.Validate(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeBadParam, Msg: api.MsgWrongParam, Payload: err})
	}

	if err := user.AcceptFriendApply(db.GetDB(), req.ApplyID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeFailure, Msg: api.MsgUnknown, Payload: err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess})
}

// RejectApply 拒绝好友申请
func RejectApply(c *fiber.Ctx) error {
	// TODO websocket
	req := new(struct {
		ApplyID int64 `json:"apply_id" validate:"required"`
	})

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeBadParam, Msg: api.MsgWrongParam, Payload: err.Error()})
	}

	if err := validator.Validate(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeBadParam, Msg: api.MsgWrongParam, Payload: err})
	}

	if err := user.RejectFriendApply(db.GetDB(), req.ApplyID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeFailure, Msg: api.MsgUnknown, Payload: err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess})
}

// GetFriends 获得所有好友
func GetFriends(c *fiber.Ctx) error {
	userId, ok := middleware.GetUserIDFromSession(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(api.BaseRes{Code: api.CodeNotLogin, Msg: api.MsgNotLogin})
	}

	entries, err := user.GetFriendEntryByUserID(db.GetDB(), userId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeFailure, Msg: api.MsgUnknown, Payload: err.Error()})
	}

	type retType struct {
		FriendID int64  `json:"friend_id"`
		Alias    string `json:"alias"`
	}
	var ret []retType = make([]retType, 0)
	for _, entry := range entries {
		ret = append(ret, retType{
			FriendID: entry.FriendID,
			Alias:    entry.Alias,
		})
	}

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess, Payload: struct {
		Friends []retType `json:"friends"`
	}{Friends: ret}})
}
