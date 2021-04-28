package user

// Copyright 2021 AyajiLin
// 用户在数据表的定义

import (
	"github.com/thss-cercis/cercis-server/db/base"
	"gorm.io/gorm"
)

// Meta 额外的用户信息
type Meta struct {
	// TODO
}

// User 用户的 dao
type User struct {
	base.Model
	NickName string `gorm:"type:varChar(255) not null" json:"nickname"`
	Email    string `gorm:"type:varChar(255) not null" json:"email"`
	Mobile   string `gorm:"type:varChar(31) not null;uniqueIndex:idx_mobile" json:"mobile"`
	Avatar   string `gorm:"type:varChar(255) not null" json:"avatar"`
	Bio      string `gorm:"type:text not null" json:"bio"`
	Password string `gorm:"type:text not null" json:"-"`
	Meta
	// 好友列表项
	FriendEntrys []FriendEntry `gorm:"foreignKey:SelfID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
}

// CreateUser 创建一个新用户
func CreateUser(db *gorm.DB, user *User) (*User, error) {
	return user, db.Create(user).Error
}

// GetUserByID 通过 ID 查找一个用户
//
// Throw: gorm.ErrRecordNotFound
func GetUserByID(db *gorm.DB, userID int) (*User, error) {
	u := new(User)
	err := db.First(u, userID).Error
	return u, err
}

// GetUserByMobile 通过 Mobile 查找一个用户
//
// Throw: gorm.ErrRecordNotFound
func GetUserByMobile(db *gorm.DB, mobile string) (*User, error) {
	u := new(User)
	err := db.Where("mobile = ?", mobile).First(u).Error
	return u, err
}

func GetUserCount(db *gorm.DB) (int64, error) {
	var cnt int64
	err := db.Model(&User{}).Count(&cnt).Error
	return cnt, err
}

// DeleteUser 软删除一个用户，带 cascade
func DeleteUser(db *gorm.DB, userID int) error {
	return db.Select("FriendEntrys").Delete(&User{Model: base.Model{ID: userID}}).Error
}

// UpdateFrom 根据主键，从数据库中获取数据
func (user *User) UpdateFrom(db *gorm.DB) error {
	return db.First(user, user.ID).Error
}

// UpdateTo 根据主键，写入到数据库
func (user *User) UpdateTo(db *gorm.DB) error {
	return db.Save(user).Error
}
