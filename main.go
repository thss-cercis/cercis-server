package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thss-cercis/cercis-server/db"
)

func main() {
	// TODO 目前此处用于 debug
	db.AutoMigrate()
	return

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})

	app.Listen("localhost:3000")
}
