package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/sirupsen/logrus"
	chatApi "github.com/thss-cercis/cercis-server/api/chat"
	friendApi "github.com/thss-cercis/cercis-server/api/friend"
	mobileApi "github.com/thss-cercis/cercis-server/api/mobile"
	searchApi "github.com/thss-cercis/cercis-server/api/search"
	logger2 "github.com/thss-cercis/cercis-server/logger"
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
	logger2.Init(logrus.Level(cf.Server.Logger.Level))

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

	// ! websocket
	v1.Use("/ws", middleware.WebsocketGetSession, middleware.WebsocketConnect())

	// user
	user := v1.Group("/user", middleware.RedisSessionAuthenticate)
	user.Get("/current", userApi.CurrentUser)
	user.Put("/modify", userApi.ModifyUser)
	user.Put("/password", userApi.ModifyPassword)
	user.Get("/info", userApi.UserInfo)

	// friend
	friend := v1.Group("/friend", middleware.RedisSessionAuthenticate)
	friend.Get("/send", friendApi.GetSendApply)
	friend.Get("/receive", friendApi.GetReceiveApply)
	friend.Post("/send", friendApi.SendApply)
	friend.Post("/accept", friendApi.AcceptApply)
	friend.Post("/reject", friendApi.RejectApply)
	friend.Get("/", friendApi.GetFriends)
	friend.Put("/", friendApi.ModifyAlias)
	friend.Delete("/", friendApi.DeleteFriend)

	// mobile
	v1.Post("/mobile/signup", mobileApi.SendSMSRegister)
	v1.Post("/mobile/recover", mobileApi.SendSMSRecover)

	// search
	search := v1.Group("/search", middleware.RedisSessionAuthenticate)
	search.Get("/users", searchApi.SearchUser)

	// chat
	chat := v1.Group("/chat", middleware.RedisSessionAuthenticate)
	chat.Get("/private", chatApi.GetPrivateChat)
	chat.Post("/private", chatApi.AddPrivateChat)
	chat.Post("/group", chatApi.AddGroupChat)
	chat.Get("/all", chatApi.GetAllChats)
	chat.Put("/group", chatApi.ModifyGroupChat)
	chat.Get("/", chatApi.GetAllChatMembers)
	chat.Delete("/", chatApi.DeleteChat)
	chat.Post("/group/member", chatApi.InviteChatMember)
	chat.Put("/group/member/alias", chatApi.ModifyChatMemberAlias)
	chat.Put("/group/member/perm", chatApi.ModifyChatMemberPerm)
	chat.Put("/group/member/owner", chatApi.ChangeGroupOwner)
	chat.Delete("/group/member", chatApi.DeleteChatMember)
	// chat - message
	chat.Post("/message", chatApi.AddMessage)
	chat.Get("/message", chatApi.GetMessage)
	chat.Get("/messages", chatApi.GetMessages)
	chat.Post("/messages/latest", chatApi.GetLatestMessages) // Get 方法不好解析数组
	chat.Get("/messages/all-latest", chatApi.GetAllChatsLatestMessageID)
	chat.Post("/message/withdraw", chatApi.WithdrawMessage)

	err := app.Listen(fmt.Sprintf("%v:%v", cf.Server.Host, cf.Server.Port))
	if err != nil {
		panic(err)
	}
}
