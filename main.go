package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	mobileApi "github.com/thss-cercis/cercis-server/api/mobile"
	"github.com/thss-cercis/cercis-server/redis"
	"github.com/thss-cercis/cercis-server/util/sms"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/thss-cercis/cercis-server/api/auth"
	userApi "github.com/thss-cercis/cercis-server/api/user"
	"github.com/thss-cercis/cercis-server/config"
	"github.com/thss-cercis/cercis-server/db"
	"github.com/thss-cercis/cercis-server/middleware"
)

func main() {
	// 命令行参数解析
	configPath := flag.String("c", "", "config path")
	flag.Parse()
	if *configPath == "" {
		panic(errors.New("ConfigPath must not be empty. Type --help"))
	}
	// 初始化
	config.Init(*configPath)
	cf := config.GetConfig()
	sms.Init(cf.SMS.Region, cf.SMS.AccessKey, cf.SMS.Secret, cf.SMS.SignName, cf.SMS.TemplateCode)

	// 自动迁移数据库
	db.AutoMigrate()

	app := fiber.New()
	// 日志中间件
	app.Use(logger.New())
	// 请求序号中间件
	app.Use(requestid.New())

	v1 := app.Group("/api/v1")

	// auth
	v1.Post("/auth/login", auth.Login)
	v1.Post("/auth/logout", middleware.RedisSessionAuthenticate, auth.Logout)
	v1.Post("/auth/signup", auth.Signup)
	v1.Post("/auth/recover", userApi.RecoverPassword)

	// user
	user := v1.Group("/user", middleware.RedisSessionAuthenticate)
	user.Get("/current", userApi.CurrentUser)
	user.Put("/modify", userApi.ModifyUser)
	user.Put("/password", userApi.ModifyPassword)

	// mobile
	v1.Post("/mobile/signup", mobileApi.SendSMSTemplate(
		redis.TagSMSSignUp, redis.ExpSMSSignUp, redis.TagSMSSignUpRetry, redis.ExpSMSSignUpRetry,
	),
	)
	v1.Post("/mobile/recover", mobileApi.SendSMSTemplate(
		redis.TagSMSRecover, redis.ExpSMSRecover, redis.TagSMSRecoverRetry, redis.ExpSMSRecoverRetry,
	),
	)

	// chat
	// chat := v1.Group("/chat", middleware.RedisSessionAuthenticate)

	err := app.Listen(fmt.Sprintf("%v:%v", cf.Server.Host, cf.Server.Port))
	if err != nil {
		panic(err)
	}
}
