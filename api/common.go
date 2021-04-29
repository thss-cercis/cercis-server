package api

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/thss-cercis/cercis-server/util"
	"github.com/thss-cercis/cercis-server/util/validator"
)

// ParamParserWrap 简化 BodyParser 的过程
func ParamParserWrap(c *fiber.Ctx, req interface{}) (ok bool, err error) {
	// 先尝试 BodyParser
	if e := c.BodyParser(req); e == nil {
		ok = true
		err = nil
		return
	}
	// 再尝试 QueryParser
	if e := c.QueryParser(req); e == nil {
		ok = true
		err = nil
		return
	}
	ok = false
	err = c.Status(fiber.StatusBadRequest).JSON(BaseRes{Code: CodeBadParam, Msg: util.MsgWithError(MsgWrongParam, errors.New("反序列化失败"))})
	return
}

// ValidateWrap 简化 Validate 的过程
func ValidateWrap(c *fiber.Ctx, req interface{}) (ok bool, err error) {
	if e := validator.Validate(req); e != nil {
		ok = false
		err = c.Status(fiber.StatusBadRequest).JSON(BaseRes{Code: CodeBadParam, Msg: util.MsgWithError(MsgWrongParam, e)})
		return
	}
	ok = true
	err = nil
	return
}
