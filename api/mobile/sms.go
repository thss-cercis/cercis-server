package mobile

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thss-cercis/cercis-server/api"
	"github.com/thss-cercis/cercis-server/redis"
	"github.com/thss-cercis/cercis-server/util/sms"
	"github.com/thss-cercis/cercis-server/util/validator"
	"time"
)

func SendSMSTemplate(tag string, exp time.Duration, tagRetry string, expRetry time.Duration) func(ctx *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		req := new(struct {
			Mobile string `json:"mobile" validate:"required,phone_number"`
		})

		if err := c.BodyParser(req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeBadParam, Msg: api.MsgWrongParam, Payload: err.Error()})
		}

		if err := validator.Validate(req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeBadParam, Msg: api.MsgWrongParam, Payload: err})
		}

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
			tmp := struct {
				Response interface{} `json:"response"`
				Error    interface{} `json:"error"`
			}{
				Response: res,
				Error:    err,
			}
			return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeSMSError, Msg: api.MsgSMSError, Payload: tmp})
		}

		err = redis.PutKV(tag, req.Mobile, code, exp)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeFailure, Msg: api.MsgUnknown, Payload: err.Error()})
		}
		// 1 分钟内禁止再索要短信
		err = redis.PutKV(tagRetry, req.Mobile, code, expRetry)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeFailure, Msg: api.MsgUnknown, Payload: err.Error()})
		}

		return c.JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess})
	}
}
