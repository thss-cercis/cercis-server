package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/sirupsen/logrus"
	"github.com/thss-cercis/cercis-server/api"
	logger2 "github.com/thss-cercis/cercis-server/logger"
	"github.com/thss-cercis/cercis-server/ws"
	"time"
)

var logFieldsWS = logrus.Fields{
	"module":     "websocket",
	"middleware": true,
}

func WebsocketGetSession(c *fiber.Ctx) error {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeBadParam, Msg: api.MsgWrongParam})
	}
	c.Cookie(&fiber.Cookie{Name: "session_id", Value: sessionID})
	userID, ok := GetUserIDFromSession(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(api.BaseRes{Code: api.CodeNotLogin, Msg: api.MsgNotLogin})
	}
	if websocket.IsWebSocketUpgrade(c) {
		c.Locals("session_id", sessionID)
		c.Locals("user_id", userID)
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

func WebsocketConnect() fiber.Handler {
	return websocket.New(func(conn *websocket.Conn) {
		logger := logger2.GetLogger()
		sessionID := conn.Locals("session_id").(string)
		userID := conn.Locals("user_id").(int64)
		logger.WithFields(logFieldsWS).Infof("Create new ws conn of user %v for session %v", userID, sessionID)
		// 存入当前的 websocket 连接
		conn.SetPingHandler(func(appData string) error {
			return conn.WriteControl(websocket.PongMessage, []byte("pong"), time.Now().Add(5*time.Second))
		})
		c := ws.PutConn(sessionID, userID, conn.Conn)
		c.Start()
	})
}
