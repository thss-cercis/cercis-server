package middleware

// Session 结构说明
// 目前在 session.Session 中存入一个名为 `user_id` 的键值对

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/gofiber/storage/redis"
	"github.com/thss-cercis/cercis-server/api"
	"github.com/thss-cercis/cercis-server/config"
)

//var loggerOut = log.New(os.Stdout, "[redis-session] ", log.LstdFlags|log.Lshortfile)
var loggerErr = log.New(os.Stderr, "[redis-session] Error: ", log.LstdFlags|log.Lshortfile)

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

// RedisSessionAuthenticate 使用 redis 的验证用户身份的中间件
func RedisSessionAuthenticate(c *fiber.Ctx) error {
	sess, err := GetSession(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(api.BaseRes{Code: api.CodeNotLogin, Msg: api.MsgNotLogin})
	}
	// get user id
	rawUserID := sess.Get("user_id")
	_, ok := rawUserID.(int64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(api.BaseRes{Code: api.CodeNotLogin, Msg: api.MsgNotLogin})
	}
	//sess.Set("user_id", userID)
	if err := sess.Save(); err != nil {
		panic(err)
	}
	return c.Next()
}

// RedisSessionAuthorize 使用 redis 的检验用户权限的中间件
func RedisSessionAuthorize(c *fiber.Ctx) error {
	return fiber.ErrNotImplemented
}
