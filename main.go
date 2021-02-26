package main

import (
	"errors"
	"flag"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/thss-cercis/cercis-server/api/auth"
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

	// 自动迁移数据库
	db.AutoMigrate()

	app := fiber.New()
	app.Use(logger.New())

	v1 := app.Group("/api/v1")

	// auth
	v1.Post("/auth/login", auth.Login)
	v1.Post("/auth/logout", middleware.RedisSessionAuthenticate, auth.Logout)
	v1.Post("/auth/signup", auth.Signup)

	// chat
	// chat := v1.Group("/chat", middleware.RedisSessionAuthenticate)

	app.Listen("localhost:9191")
}
