package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/thss-cercis/cercis-server/db/user"
)

var dbNow *gorm.DB = nil

const sqlString = "host=localhost user=postgres password=86132292 dbname=cercis port=5432 sslmode=disable TimeZone=Asia/Shanghai"

// GetDB 获得数据库
func GetDB() *gorm.DB {
	if dbNow == nil {
		db, err := gorm.Open(postgres.Open(sqlString), &gorm.Config{})
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
		&user.User{}, &user.FriendEntry{},
	)
	if err != nil {
		panic(err)
	}
}
