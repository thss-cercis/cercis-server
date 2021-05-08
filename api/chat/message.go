package chat

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thss-cercis/cercis-server/api"
	"github.com/thss-cercis/cercis-server/db"
	"github.com/thss-cercis/cercis-server/db/chat"
	"github.com/thss-cercis/cercis-server/middleware"
	"github.com/thss-cercis/cercis-server/util"
)

// AddMessage 添加新消息 api
func AddMessage(c *fiber.Ctx) error {
	// TODO: type 限制
	req := new(struct {
		ChatID  int64        `json:"chat_id" validate:"required"`
		Type    chat.MsgType `json:"type" validate:"gte=0"`
		Message string       `json:"message" validate:"min=1"`
	})

	if ok, err := api.ParamParserWrap(c, req); !ok {
		return err
	}

	if ok, err := api.ValidateWrap(c, req); !ok {
		return err
	}

	userID, ok := middleware.GetUserIDFromSession(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(api.BaseRes{Code: api.CodeNotLogin, Msg: api.MsgNotLogin})
	}

	msg, err := chat.CreateMessage(db.GetDB(), req.ChatID, userID, req.Type, req.Message)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeChatError, Msg: util.MsgWithError(api.MsgChatError, err)})
	}

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess, Payload: msg})
}

// GetMessage 查询一条消息 api
func GetMessage(c *fiber.Ctx) error {
	req := new(struct {
		ChatID    int64 `json:"chat_id" query:"chat_id" validate:"required"`
		MessageID int64 `json:"message_id" query:"message_id" validate:"required"`
	})

	if ok, err := api.ParamParserWrap(c, req); !ok {
		return err
	}

	if ok, err := api.ValidateWrap(c, req); !ok {
		return err
	}

	userID, ok := middleware.GetUserIDFromSession(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(api.BaseRes{Code: api.CodeNotLogin, Msg: api.MsgNotLogin})
	}

	msg, err := chat.GetMessage(db.GetDB(), req.ChatID, userID, req.MessageID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeChatError, Msg: util.MsgWithError(api.MsgChatError, err)})
	}

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess, Payload: msg})
}

// GetMessages 查询一堆消息 api
func GetMessages(c *fiber.Ctx) error {
	req := new(struct {
		ChatID int64 `json:"chat_id" query:"chat_id" validate:"required"`
		FromID int64 `json:"from_id" query:"from_id"`
		ToID   int64 `json:"to_id" query:"to_id"`
	})

	if ok, err := api.ParamParserWrap(c, req); !ok {
		return err
	}

	if ok, err := api.ValidateWrap(c, req); !ok {
		return err
	}

	userID, ok := middleware.GetUserIDFromSession(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(api.BaseRes{Code: api.CodeNotLogin, Msg: api.MsgNotLogin})
	}

	messages, err := chat.GetMessages(db.GetDB(), req.ChatID, userID, req.FromID, req.ToID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeChatError, Msg: util.MsgWithError(api.MsgChatError, err)})
	}

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess, Payload: messages})
}

// WithdrawMessage 撤回一条消息 api
func WithdrawMessage(c *fiber.Ctx) error {
	req := new(struct {
		ChatID    int64 `json:"chat_id" validate:"required"`
		MessageID int64 `json:"message_id" validate:"required"`
	})

	if ok, err := api.ParamParserWrap(c, req); !ok {
		return err
	}

	if ok, err := api.ValidateWrap(c, req); !ok {
		return err
	}

	userID, ok := middleware.GetUserIDFromSession(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(api.BaseRes{Code: api.CodeNotLogin, Msg: api.MsgNotLogin})
	}

	err := chat.WithdrawMessage(db.GetDB(), req.ChatID, userID, req.MessageID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeChatError, Msg: util.MsgWithError(api.MsgChatError, err)})
	}

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess})
}
