package mobile

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thss-cercis/cercis-server/api"
	"github.com/thss-cercis/cercis-server/db"
	"github.com/thss-cercis/cercis-server/db/user"
	"github.com/thss-cercis/cercis-server/redis"
	"github.com/thss-cercis/cercis-server/util"
	"github.com/thss-cercis/cercis-server/util/sms"
	"time"
)

type SMSReq struct {
	Mobile string `json:"mobile" validate:"required,phone_number"`
}

func SendSMSRegister(c *fiber.Ctx) error {
	// 已经注册过的不可再发短信
	req := &SMSReq{}

	if ok, err := api.ParamParserWrap(c, req); !ok {
		return err
	}

	if ok, err := api.ValidateWrap(c, req); !ok {
		return err
	}

	if _, err := user.GetUserByMobile(db.GetDB(), req.Mobile); err == nil {
		// 用户已经存在
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeUserAlreadyExist, Msg: api.MsgUserAlreadyExist})
	}

	return SendSMSTemplate(
		req, redis.TagSMSSignUp, redis.ExpSMSSignUp, redis.TagSMSSignUpRetry, redis.ExpSMSSignUpRetry,
	)(c)
}

func SendSMSRecover(c *fiber.Ctx) error {
	// 已经注册过的不可再发短信
	req := &SMSReq{}

	if ok, err := api.ParamParserWrap(c, req); !ok {
		return err
	}

	if ok, err := api.ValidateWrap(c, req); !ok {
		return err
	}

	if _, err := user.GetUserByMobile(db.GetDB(), req.Mobile); err != nil {
		// 用户不存在
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeUserIDNotFound, Msg: util.MsgWithError(api.MsgUserNotFound, err)})
	}

	return SendSMSTemplate(
		req, redis.TagSMSSignUp, redis.ExpSMSSignUp, redis.TagSMSSignUpRetry, redis.ExpSMSSignUpRetry,
	)(c)
}

func SendSMSTemplate(req *SMSReq, tag string, exp time.Duration, tagRetry string, expRetry time.Duration) func(ctx *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		// 冷却期仍未过
		if _, err := redis.GetKV(tagRetry, req.Mobile); err == nil {
			return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeSMSTooOften, Msg: api.MsgSMSTooOften})
		}

		client, ok := sms.GetClient()
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeSMSError, Msg: api.MsgSMSError})
		}

		code := sms.NewRandomCode()
		res, err := sms.SendSMS(client, req.Mobile, code)
		if err != nil || res.Code != "OK" {
			return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeSMSError, Msg: util.MsgWithError(api.MsgSMSError, err)})
		}

		err = redis.PutKV(tag, req.Mobile, code, exp)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeFailure, Msg: util.MsgWithError(api.MsgUnknown, err)})
		}
		// 1 分钟内禁止再索要短信
		err = redis.PutKV(tagRetry, req.Mobile, code, expRetry)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeFailure, Msg: util.MsgWithError(api.MsgUnknown, err)})
		}

		return c.JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess})
	}
}
