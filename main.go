package main

import (
	"errors"
	"flag"
	"fmt"

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

	// 自动迁移数据库
	db.AutoMigrate()

	app := fiber.New()
	app.Use(logger.New())

	v1 := app.Group("/api/v1")

	// auth
	v1.Post("/auth/login", auth.Login)
	v1.Post("/auth/logout", middleware.RedisSessionAuthenticate, auth.Logout)
	v1.Post("/auth/signup", auth.Signup)

	// user
	user := v1.Group("/user", middleware.RedisSessionAuthenticate)
	user.Get("/current", userApi.Current)

	// chat
	// chat := v1.Group("/chat", middleware.RedisSessionAuthenticate)

	err := app.Listen(fmt.Sprintf("%s:%d", cf.Server.Host, cf.Server.Port))
	if err != nil {
		panic(err)
	}
}
