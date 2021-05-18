package chat

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/thss-cercis/cercis-server/api"
	"github.com/thss-cercis/cercis-server/db"
	"github.com/thss-cercis/cercis-server/db/chat"
	logger2 "github.com/thss-cercis/cercis-server/logger"
	"github.com/thss-cercis/cercis-server/middleware"
	"github.com/thss-cercis/cercis-server/util"
	"github.com/thss-cercis/cercis-server/ws"
)

var logMsgFields = logrus.Fields{
	"module": "message",
	"api":    true,
}

// AddMessage 添加新消息 api
func AddMessage(c *fiber.Ctx) error {
	req := new(struct {
		ChatID  int64        `json:"chat_id" validate:"required"`
		Type    chat.MsgType `json:"type" validate:"gte=0,lte=5"`
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

	// websocket
	go func() {
		chatMembers, err := chat.GetChatMembers(db.GetDB(), req.ChatID)
		if err != nil {
			logger := logger2.GetLogger()
			logger.WithFields(logMsgFields).Errorf("websocket to send msg notification fail for chat %v", req.ChatID)
			return
		}
		sum := util.FirstNCharOfString(req.Message, 30)
		for _, chatMember := range chatMembers {
			err := ws.WriteToUser(chatMember.UserID, &struct {
				Type int64 `json:"type"`
				Msg  struct {
					ChatID int64        `json:"chat_id"`
					MsgID  int64        `json:"msg_id"`
					Type   chat.MsgType `json:"type"`
					Sum    string       `json:"sum"`
				}
			}{
				Type: api.TypeAddNewMessage,
				Msg: struct {
					ChatID int64        `json:"chat_id"`
					MsgID  int64        `json:"msg_id"`
					Type   chat.MsgType `json:"type"`
					Sum    string       `json:"sum"`
				}{ChatID: msg.ChatID, MsgID: msg.MessageID, Type: msg.Type, Sum: sum},
			})
			if err != nil {
				continue
			}
		}
	}()

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

// GetLatestMessages 获得所有获得某个用户给定某些聊天的最新消息 api
func GetLatestMessages(c *fiber.Ctx) error {
	req := new(struct {
		ChatIDs []int64 `json:"chat_ids"`
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

	msgs, err := chat.GetLatestMessages(db.GetDB(), userID, req.ChatIDs)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess, Payload: msgs})
}

// GetAllChatsLatestMessageID 获得某个用户所有的聊天的最新消息 id 的 api
func GetAllChatsLatestMessageID(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserIDFromSession(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(api.BaseRes{Code: api.CodeNotLogin, Msg: api.MsgNotLogin})
	}

	rets, err := chat.GetAllChatsLatestMessageID(db.GetDB(), userID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess, Payload: rets})
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

	msg, err := chat.WithdrawMessage(db.GetDB(), req.ChatID, userID, req.MessageID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeChatError, Msg: util.MsgWithError(api.MsgChatError, err)})
	}

	// websocket
	go func() {
		chatMembers, err := chat.GetChatMembers(db.GetDB(), req.ChatID)
		if err != nil {
			logger := logger2.GetLogger()
			logger.WithFields(logMsgFields).Errorf("websocket to send msg notification fail for chat %v", req.ChatID)
			return
		}
		for _, chatMember := range chatMembers {
			err := ws.WriteToUser(chatMember.UserID, &struct {
				Type int64 `json:"type"`
				Msg  struct {
					ChatID int64        `json:"chat_id"`
					MsgID  int64        `json:"msg_id"`
					Type   chat.MsgType `json:"type"`
					Sum    string       `json:"sum"`
				}
			}{
				Type: api.TypeAddNewMessage,
				Msg: struct {
					ChatID int64        `json:"chat_id"`
					MsgID  int64        `json:"msg_id"`
					Type   chat.MsgType `json:"type"`
					Sum    string       `json:"sum"`
				}{ChatID: msg.ChatID, MsgID: msg.MessageID, Type: msg.Type, Sum: msg.Message},
			})
			if err != nil {
				continue
			}
		}
	}()

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess, Payload: msg})
}
