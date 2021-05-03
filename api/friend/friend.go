package friend

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thss-cercis/cercis-server/api"
	"github.com/thss-cercis/cercis-server/db"
	"github.com/thss-cercis/cercis-server/db/user"
	"github.com/thss-cercis/cercis-server/middleware"
	"github.com/thss-cercis/cercis-server/util"
)

// GetSendApply 获得自己发送的好友申请
func GetSendApply(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserIDFromSession(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(api.BaseRes{Code: api.CodeNotLogin, Msg: api.MsgNotLogin})
	}

	applies, err := user.GetFriendApplyFromByUserID(db.GetDB(), userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeFailure, Msg: util.MsgWithError(api.MsgUnknown, err)})
	}

	type resType struct {
		ApplyID   int64                 `json:"apply_id"`
		FromID    int64                 `json:"from_id"`
		ToID      int64                 `json:"to_id"`
		Alias     string                `json:"alias"`
		Remark    string                `json:"remark"`
		State     user.FriendApplyState `json:"state"`
		CreatedAt int64                 `json:"created_at"`
	}
	var res = make([]resType, 0)
	for _, apply := range applies {
		res = append(res, resType{
			ApplyID:   apply.ID,
			FromID:    apply.FromID,
			ToID:      apply.ToID,
			Alias:     apply.Alias,
			Remark:    apply.Remark,
			State:     apply.State,
			CreatedAt: apply.CreatedAt.UnixNano(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess, Payload: struct {
		Applies []resType `json:"applies"`
	}{
		Applies: res,
	}})
}

// GetReceiveApply 获得自己收到的好友申请
func GetReceiveApply(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserIDFromSession(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(api.BaseRes{Code: api.CodeNotLogin, Msg: api.MsgNotLogin})
	}

	applies, err := user.GetFriendApplyToByUserID(db.GetDB(), userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeFailure, Msg: util.MsgWithError(api.MsgUnknown, err)})
	}

	type resType struct {
		ApplyID   int64                 `json:"apply_id"`
		FromID    int64                 `json:"from_id"`
		ToID      int64                 `json:"to_id"`
		Remark    string                `json:"remark"`
		State     user.FriendApplyState `json:"state"`
		CreatedAt int64                 `json:"created_at"`
	}
	var res = make([]resType, 0)
	for _, apply := range applies {
		res = append(res, resType{
			ApplyID:   apply.ID,
			FromID:    apply.FromID,
			ToID:      apply.ToID,
			Remark:    apply.Remark,
			State:     apply.State,
			CreatedAt: apply.CreatedAt.UnixNano(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess, Payload: struct {
		Applies []resType `json:"applies"`
	}{
		Applies: res,
	}})
}

// SendApply 发送好友申请
func SendApply(c *fiber.Ctx) error {
	// TODO websocket
	req := new(struct {
		ToID int64 `json:"to_id" validate:"required"`
		// 申请者给接受者的预设备注
		Alias string `json:"alias" validate:"max=127"`
		// 验证消息
		Remark string `json:"remark" validate:"max=255"`
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

	// 自己不能发给自己
	if userID == req.ToID {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeFailure, Msg: "不允许向自身发送好友请求"})
	}
	_, err := user.CreateFriendApply(db.GetDB(), userID, req.ToID, req.Alias, req.Remark)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeFailure, Msg: util.MsgWithError(api.MsgUnknown, err)})
	}

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess})
}

// AcceptApply 接收好友申请
func AcceptApply(c *fiber.Ctx) error {
	// TODO websocket
	req := new(struct {
		ApplyID int64 `json:"apply_id" validate:"required"`
		// 接收者给申请者的备注
		Alias string `json:"alias" validate:"max=127"`
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

	if err := user.AcceptFriendApply(db.GetDB(), req.ApplyID, userID, req.Alias); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeFailure, Msg: util.MsgWithError(api.MsgUnknown, err)})
	}

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess})
}

// RejectApply 拒绝好友申请
func RejectApply(c *fiber.Ctx) error {
	// TODO websocket
	req := new(struct {
		ApplyID int64 `json:"apply_id" validate:"required"`
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

	if err := user.RejectFriendApply(db.GetDB(), req.ApplyID, userID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeFailure, Msg: util.MsgWithError(api.MsgUnknown, err)})
	}

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess})
}

// GetFriends 获得所有好友
func GetFriends(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserIDFromSession(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(api.BaseRes{Code: api.CodeNotLogin, Msg: api.MsgNotLogin})
	}

	entries, err := user.GetFriendEntrySelfByUserID(db.GetDB(), userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeFailure, Msg: util.MsgWithError(api.MsgUnknown, err)})
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

// ModifyAlias 修改备注名
func ModifyAlias(c *fiber.Ctx) error {
	req := new(struct {
		FriendID int64  `json:"friend_id" validate:"required"`
		Alias    string `json:"alias" validate:"required"`
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

	// 修改备注名
	if _, err := user.ModifyFriendEntryAlias(db.GetDB(), userID, req.FriendID, req.Alias); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeFailure, Msg: util.MsgWithError(api.MsgUnknown, err)})
	}

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess})
}

// DeleteFriend 双向删除好友
func DeleteFriend(c *fiber.Ctx) error {
	req := new(struct {
		FriendID int64 `json:"friend_id" validate:"required"`
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

	// 修改备注名
	if err := user.DeleteFriendEntryBi(db.GetDB(), userID, req.FriendID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeFailure, Msg: util.MsgWithError(api.MsgUnknown, err)})
	}

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess})
}
