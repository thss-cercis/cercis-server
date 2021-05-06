package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/sirupsen/logrus"
	"github.com/thss-cercis/cercis-server/api"
	logger2 "github.com/thss-cercis/cercis-server/logger"
	"github.com/thss-cercis/cercis-server/ws"
)

var logFieldsWS = logrus.Fields{
	"middleware": true,
	"module":     "websocket",
}

func WebsocketGetSession(c *fiber.Ctx) error {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeBadParam, Msg: api.MsgWrongParam})
	}
	c.Cookie(&fiber.Cookie{Name: "session_id", Value: sessionID})
	if _, err := GetSession(c); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(api.BaseRes{Code: api.CodeNotLogin, Msg: api.MsgNotLogin})
	}
	if websocket.IsWebSocketUpgrade(c) {
		c.Locals("session_id", sessionID)
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

func WebsocketConnect() fiber.Handler {
	return websocket.New(func(conn *websocket.Conn) {
		logger := logger2.GetLogger()
		sessionID := conn.Locals("session_id").(string)
		// 存入当前的 websocket 连接
		ws.PutConn(sessionID, conn.Conn)
		logger.WithFields(logFieldsWS).Infof("Create new ws conn for session %v", sessionID)
	})
}
