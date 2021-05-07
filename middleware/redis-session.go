package middleware

// Session 结构说明
// 目前在 session.Session 中存入一个名为 `user_id` 的键值对

import (
	"github.com/sirupsen/logrus"
	logger2 "github.com/thss-cercis/cercis-server/logger"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/gofiber/storage/redis"
	"github.com/thss-cercis/cercis-server/api"
	"github.com/thss-cercis/cercis-server/config"
)

var logFieldsRedis = logrus.Fields{
	"module":     "redis-session",
	"middleware": true,
}

var store *session.Store

// GetStore 获得 redis 数据库连接
func GetStore() *session.Store {
	if store == nil {
		cr := config.GetConfig().Redis
		storage := redis.New(redis.Config{
			Host:     cr.Host,
			Port:     cr.Port,
			Username: cr.Username,
			Password: cr.Password,
			Database: cr.Database,
			Reset:    cr.Reset,
		})
		store = session.New(session.Config{
			Expiration:   24 * time.Hour,
			Storage:      storage,
			CookieName:   "session_id",
			KeyGenerator: utils.UUIDv4,
		})
	}
	return store
}

// GetSession 获得当前 ctx 中的 session
func GetSession(c *fiber.Ctx) (*session.Session, error) {
	sess, err := GetStore().Get(c)
	return sess, err
}

// GetUserIDFromSession 从 ctx 中的 session 中获取当前 userID
func GetUserIDFromSession(c *fiber.Ctx) (userID int64, ok bool) {
	sess, err := GetSession(c)
	if err != nil {
		return 0, false
	}
	userID, ok = sess.Get("user_id").(int64)
	return
}

func GetSessionIDFromSession(c *fiber.Ctx) (sessionID string, ok bool) {
	_, err := GetSession(c)
	if err != nil {
		return "", false
	}
	sessionID = c.Cookies("session_id")
	if sessionID == "" {
		ok = false
	} else {
		ok = true
	}
	return
}

// RedisSessionAuthenticate 使用 redis 的验证用户身份的中间件
func RedisSessionAuthenticate(c *fiber.Ctx) error {
	logger := logger2.GetLogger()
	sess, err := GetSession(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(api.BaseRes{Code: api.CodeNotLogin, Msg: api.MsgNotLogin})
	}
	// get user id
	rawUserID := sess.Get("user_id")
	userID, ok := rawUserID.(int64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(api.BaseRes{Code: api.CodeNotLogin, Msg: api.MsgNotLogin})
	}

	if err := sess.Save(); err != nil {
		panic(err)
	}

	logger.WithFields(logFieldsRedis).Infof("User_id %v with session %v", userID, c.Cookies("session_id"))
	return c.Next()
}

// RedisSessionAuthorize 使用 redis 的检验用户权限的中间件
func RedisSessionAuthorize(c *fiber.Ctx) error {
	return fiber.ErrNotImplemented
}
