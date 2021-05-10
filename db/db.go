package db

import (
	"fmt"
	"github.com/thss-cercis/cercis-server/db/activity"
	"github.com/thss-cercis/cercis-server/db/chat"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/thss-cercis/cercis-server/config"
	"github.com/thss-cercis/cercis-server/db/user"
)

var dbNow *gorm.DB = nil

const connectStr = "host=%v user=%v password=%v dbname=%v port=%v sslmode=%v TimeZone=%v"

// GetDB 获得数据库
func GetDB() *gorm.DB {
	if dbNow == nil {
		cp := config.GetConfig().Postgres
		db, err := gorm.Open(
			postgres.Open(fmt.Sprintf(connectStr, cp.Host, cp.User, cp.Password, cp.Dbname, cp.Port, cp.Sslmode, cp.Timezone)),
			&gorm.Config{},
		)
		if err != nil {
			panic(err)
		}
		dbNow = db
	}
	return dbNow
}

// AutoMigrate 更新数据库
func AutoMigrate() {
	db := GetDB()
	err := db.Migrator().AutoMigrate(
		&user.User{}, &user.FriendEntry{}, &user.FriendApply{}, &chat.Chat{}, &chat.ChatUser{}, &chat.Message{},
		&activity.Activity{}, &activity.ActivityMedium{}, &activity.ActivityComment{},
	)
	if err != nil {
		panic(err)
	}
	// id of `users` start from 100001
	if cnt, err := user.GetUserCount(db); err == nil && cnt == 0 {
		db.Exec("alter sequence users_id_seq restart 100001")
	}
	//创建 join 表
	err = db.SetupJoinTable(&chat.Chat{}, "Members", &chat.ChatUser{})
	if err != nil {
		panic(err)
	}
}
