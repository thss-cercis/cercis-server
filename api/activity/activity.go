package activity

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/thss-cercis/cercis-server/api"
	"github.com/thss-cercis/cercis-server/db"
	"github.com/thss-cercis/cercis-server/db/activity"
	"github.com/thss-cercis/cercis-server/db/user"
	logger2 "github.com/thss-cercis/cercis-server/logger"
	"github.com/thss-cercis/cercis-server/middleware"
	"github.com/thss-cercis/cercis-server/util"
	"github.com/thss-cercis/cercis-server/ws"
)

var logActivityFields = logrus.Fields{
	"module": "activity",
	"api":    true,
}

// AddActivity 新建动态 api
func AddActivity(c *fiber.Ctx) error {
	req := new(struct {
		Text  string                   `json:"text" validate:"required_without=Media"`
		Media []activity.MediumCapsule `json:"media" validate:"omitempty,min=1"`
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

	ac, err := activity.CreateActivity(db.GetDB(), userID, req.Text, req.Media)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeActivityCreateFail, Msg: util.MsgWithError(api.MsgActivityCreateFail, err)})
	}

	// websocket
	go func() {
		members, err := user.GetFriendEntrySelfByUserID(db.GetDB(), userID)
		if err != nil {
			logger := logger2.GetLogger()
			logger.WithFields(logActivityFields).Errorf("websocket to send msg notification fail for activity %v", ac.ID)
			return
		}
		for _, member := range members {
			err := ws.WriteToUser(member.ID, &struct {
				Type int64 `json:"type"`
				Msg  struct {
					ActivityID int64 `json:"activity_id"`
					UserID     int64 `json:"user_id"`
				} `json:"msg"`
			}{
				Type: api.TypeNewActivity,
				Msg: struct {
					ActivityID int64 `json:"activity_id"`
					UserID     int64 `json:"user_id"`
				}{ActivityID: ac.ID, UserID: userID},
			})
			if err != nil {
				continue
			}
		}
	}()

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess, Payload: ac})
}

// GetActivity 获得动态 api
func GetActivity(c *fiber.Ctx) error {
	req := new(struct {
		ActivityID int64 `query:"activity_id" validate:"required"`
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

	ac, err := activity.GetActivity(db.GetDB(), req.ActivityID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeActivityError, Msg: util.MsgWithError(api.MsgActivityError, err)})
	}

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess, Payload: ac})
}

// GetActivitiesBefore 获得动态 api
func GetActivitiesBefore(c *fiber.Ctx) error {
	req := new(struct {
		ActivityID int64 `query:"activity_id" validate:"required"`
		Count      int64 `query:"count" validate:"omitempty,gte=0"`
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

	acs, err := activity.GetActivitiesBefore(db.GetDB(), userID, req.ActivityID, req.Count)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeActivityError, Msg: util.MsgWithError(api.MsgActivityError, err)})
	}

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess, Payload: acs})
}

// GetActivitiesAfter 获得动态 api
func GetActivitiesAfter(c *fiber.Ctx) error {
	req := new(struct {
		ActivityID int64 `query:"activity_id" validate:"required"`
		Count      int64 `query:"count" validate:"omitempty,gte=0"`
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

	acs, err := activity.GetActivitiesAfter(db.GetDB(), userID, req.ActivityID, req.Count)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeActivityError, Msg: util.MsgWithError(api.MsgActivityError, err)})
	}

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess, Payload: acs})
}

// DeleteActivity 删除动态 api
func DeleteActivity(c *fiber.Ctx) error {
	req := new(struct {
		ActivityID int64 `json:"activity_id" validate:"required"`
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

	if err := activity.DeleteActivity(db.GetDB(), userID, req.ActivityID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeActivityDeleteFail, Msg: util.MsgWithError(api.MsgActivityDeleteFail, err)})
	}

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess})
}

// CommentActivity 评论动态 api
func CommentActivity(c *fiber.Ctx) error {
	req := new(struct {
		ActivityID int64  `json:"activity_id" validate:"required"`
		Content    string `json:"content" validate:"required"`
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

	comment, err := activity.CreateActivityComment(db.GetDB(), userID, req.Content, req.ActivityID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeActivityCommentCreateFail, Msg: util.MsgWithError(api.MsgActivityCommentCreateFail, err)})
	}

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess, Payload: comment})
}

// DeleteActivityComment 删除动态评论 api
func DeleteActivityComment(c *fiber.Ctx) error {
	req := new(struct {
		CommentID int64 `json:"comment_id" validate:"required"`
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

	err := activity.DeleteActivityComment(db.GetDB(), userID, req.CommentID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeActivityCommentCreateFail, Msg: util.MsgWithError(api.MsgActivityCommentCreateFail, err)})
	}

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess})
}

// ThumbUpActivity 点赞动态
func ThumbUpActivity(c *fiber.Ctx) error {
	req := new(struct {
		ActivityID int64 `json:"activity_id" validate:"required"`
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

	if err := activity.AddActivityThumbUp(db.GetDB(), req.ActivityID, userID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeActivityError, Msg: util.MsgWithError(api.MsgActivityError, err)})
	}

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess})
}

// ThumbDownActivity 取消点赞动态
func ThumbDownActivity(c *fiber.Ctx) error {
	req := new(struct {
		ActivityID int64 `json:"activity_id" validate:"required"`
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

	if err := activity.DeleteActivityThumbUp(db.GetDB(), req.ActivityID, userID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeActivityError, Msg: util.MsgWithError(api.MsgActivityError, err)})
	}

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess})
}
