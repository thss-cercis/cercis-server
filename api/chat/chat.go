package chat

import (
	"errors"
	mapset "github.com/deckarep/golang-set"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/thss-cercis/cercis-server/api"
	"github.com/thss-cercis/cercis-server/db"
	chat2 "github.com/thss-cercis/cercis-server/db/chat"
	"github.com/thss-cercis/cercis-server/middleware"
	"github.com/thss-cercis/cercis-server/util"
)

var logChatFields = logrus.Fields{
	"module": "chat",
	"api":    true,
}

// AddPrivateChat 创建新私聊的 api
func AddPrivateChat(c *fiber.Ctx) error {
	req := new(struct {
		ID int64 `json:"id" validate:"required"`
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

	// 不允许发给自己
	if userID == req.ID {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeChatCreateFail, Msg: api.MsgChatCreateFail})
	}
	// 直接建立
	chat, err := chat2.CreatePrivateChat(db.GetDB(), userID, req.ID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeChatCreateFail, Msg: util.MsgWithError(api.MsgChatCreateFail, err)})
	}

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess, Payload: chat})
}

// AddGroupChat 创建群聊
func AddGroupChat(c *fiber.Ctx) error {
	req := new(struct {
		Name    string  `json:"name" validate:"required,min=1"`
		Members []int64 `json:"member_ids"`
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

	tmp := make([]interface{}, len(req.Members))
	for _, memberID := range req.Members {
		tmp = append(tmp, memberID)
	}
	chat, err := chat2.CreateGroupChat(db.GetDB(), req.Name, userID, mapset.NewSetFromSlice(tmp))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeChatCreateFail, Msg: util.MsgWithError(api.MsgChatCreateFail, err)})
	}
	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess, Payload: chat})
}

// GetPrivateChat 获得私聊 api
func GetPrivateChat(c *fiber.Ctx) error {
	req := new(struct {
		ID int64 `json:"id" validate:"required"`
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

	chat, err := chat2.GetPrivateChat(db.GetDB(), userID, req.ID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeChatError, Msg: util.MsgWithError(api.MsgChatError, err)})
	}

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess, Payload: chat})
}

// GetAllChats 获得所有的群聊
func GetAllChats(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserIDFromSession(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(api.BaseRes{Code: api.CodeNotLogin, Msg: api.MsgNotLogin})
	}

	chats, err := chat2.GetAllChats(db.GetDB(), userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeChatError, Msg: util.MsgWithError(api.MsgChatError, err)})
	}

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess, Payload: chats})
}

// ModifyGroupChat 修改群聊的信息
func ModifyGroupChat(c *fiber.Ctx) error {
	req := new(struct {
		ChatID int64  `json:"chat_id"`
		Name   string `json:"name" validate:"omitempty,min=1"`
		Avatar string `json:"avatar" validate:"omitempty,url"`
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

	chatUser, err := chat2.GetChatMember(db.GetDB(), req.ChatID, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeChatError, Msg: util.MsgWithError(api.MsgChatError, err)})
	}

	// 只有群主能改
	if chatUser.Permission != chat2.PermOwner {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeChatError, Msg: util.MsgWithError(api.MsgChatError, err)})
	}

	// 修改
	chat, err := chat2.GetChat(db.GetDB(), req.ChatID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeChatError, Msg: util.MsgWithError(api.MsgChatError, err)})
	}
	if req.Name != "" {
		chat.Name = req.Name
	}
	if req.Avatar != "" {
		chat.Avatar = req.Avatar
	}

	if err := chat.UpdateTo(db.GetDB()); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeChatError, Msg: util.MsgWithError(api.MsgChatError, err)})
	}

	return c.JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess, Payload: chat})
}

// DeleteChat 删除聊天的 api
func DeleteChat(c *fiber.Ctx) error {
	req := new(struct {
		ChatID int64 `json:"chat_id" validate:"required"`
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

	if err := chat2.DeleteChat(db.GetDB(), userID, req.ChatID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeChatDeleteFail, Msg: util.MsgWithError(api.MsgChatDeleteFail, err)})
	}

	return c.JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess})
}

// InviteChatMember 邀请新聊天成员的 api
func InviteChatMember(c *fiber.Ctx) error {
	// 邀请新的群组人员
	// 目前不需要对方同意
	req := new(struct {
		ChatID int64 `json:"chat_id" validate:"required"`
		UserID int64 `json:"user_id" validate:"required"`
	})

	if ok, err := api.ParamParserWrap(c, req); !ok {
		return err
	}

	if ok, err := api.ValidateWrap(c, req); !ok {
		return err
	}

	_, ok := middleware.GetUserIDFromSession(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(api.BaseRes{Code: api.CodeNotLogin, Msg: api.MsgNotLogin})
	}

	if _, err := chat2.AddChatMember(db.GetDB(), req.ChatID, req.UserID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeChatMemberAddFail, Msg: util.MsgWithError(api.MsgChatMemberAddFail, err)})
	}

	return c.JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess})
}

// GetAllChatMembers 获得所有聊天成员
func GetAllChatMembers(c *fiber.Ctx) error {
	req := new(struct {
		ID int64 `json:"id" validate:"required"`
	})

	if ok, err := api.ParamParserWrap(c, req); !ok {
		return err
	}

	if ok, err := api.ValidateWrap(c, req); !ok {
		return err
	}

	_, ok := middleware.GetUserIDFromSession(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(api.BaseRes{Code: api.CodeNotLogin, Msg: api.MsgNotLogin})
	}

	members, err := chat2.GetChatMembers(db.GetDB(), req.ID)
	if err != nil {
		return err
	}

	return c.JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess, Payload: members})
}

// ModifyChatMemberAlias 修改成员备注名
func ModifyChatMemberAlias(c *fiber.Ctx) error {
	req := new(struct {
		ChatID int64  `json:"chat_id" validate:"required"`
		UserID int64  `json:"user_id" validate:"required"`
		Alias  string `json:"alias"`
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

	if userID != req.UserID {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeChatError, Msg: util.MsgWithError(api.MsgChatError, errors.New("could only modify your own alias"))})
	}

	if err := chat2.ModifyChatMemberAlias(db.GetDB(), req.ChatID, req.UserID, req.Alias); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeChatError, Msg: util.MsgWithError(api.MsgChatError, err)})
	}

	return c.JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess})
}

// ModifyChatMemberPerm 修改成员权限
func ModifyChatMemberPerm(c *fiber.Ctx) error {
	req := new(struct {
		ChatID     int64                  `json:"chat_id" validate:"required"`
		UserID     int64                  `json:"user_id" validate:"required"`
		Permission chat2.MemberPermission `json:"permission" validate:"gte=0,lte=2"`
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

	if userID == req.UserID {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeChatError, Msg: api.MsgChatError})
	}

	if err := chat2.ModifyChatMemberPermission(db.GetDB(), userID, req.ChatID, req.UserID, req.Permission); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeChatError, Msg: util.MsgWithError(api.MsgChatError, err)})
	}

	return c.JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess})
}

// ChangeGroupOwner 禅让群主的 api
func ChangeGroupOwner(c *fiber.Ctx) error {
	// 禅让群主，需要自己是群主，否则数据库报错
	req := new(struct {
		ChatID int64 `json:"chat_id" validate:"required"`
		UserID int64 `json:"user_id" validate:"required"`
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

	if err := chat2.ChangeGroupOwner(db.GetDB(), userID, req.ChatID, req.UserID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeChatError, Msg: util.MsgWithError(api.MsgChatError, err)})
	}

	return c.JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess})
}

// DeleteChatMember 删除聊天成员的 api
func DeleteChatMember(c *fiber.Ctx) error {
	// 删除聊天成员，需要权限大于被删者，或者是删除自己
	req := new(struct {
		ChatID int64 `json:"chat_id" validate:"required"`
		UserID int64 `json:"user_id" validate:"required"`
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

	if err := chat2.DeleteChatMember(db.GetDB(), userID, req.ChatID, req.UserID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeChatError, Msg: util.MsgWithError(api.MsgChatError, err)})
	}

	return c.JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess})
}
