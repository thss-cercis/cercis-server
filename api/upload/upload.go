package upload

import (
	"github.com/gofiber/fiber/v2"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
	"github.com/thss-cercis/cercis-server/api"
	"github.com/thss-cercis/cercis-server/config"
	"github.com/thss-cercis/cercis-server/middleware"
)

func GetUploadToken(c *fiber.Ctx) error {
	_, ok := middleware.GetUserIDFromSession(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(api.BaseRes{Code: api.CodeNotLogin, Msg: api.MsgNotLogin})
	}

	conf := config.GetConfig()
	putPolicy := storage.PutPolicy{
		Scope:   conf.Qiniu.Bucket,
		Expires: 1800,
	}
	mac := qbox.NewMac(conf.Qiniu.AccessKey, conf.Qiniu.SecretKey)
	upToken := putPolicy.UploadToken(mac)

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess, Payload: struct {
		UploadToken string `json:"upload_token"`
	}{
		UploadToken: upToken,
	}})
}
