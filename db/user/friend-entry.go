package user

// Copyright 2021 AyajiLin
// 好友列表项的数据库定义

import (
	"github.com/thss-cercis/cercis-server/db/base"
	"gorm.io/gorm"
)

// FriendEntry 好友列表项的 dao
type FriendEntry struct {
	base.Model
	SelfID   int    `gorm:"uniqueIndex:idx_composited_id" json:"self_id"`
	FriendID int    `gorm:"uniqueIndex:idx_composited_id" json:"friend_id"`
	Friend   User   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Alias    string `gorm:"type:varChar(127) not null"`
}

// CreateFriendEntry 创建一个单向好友列表项目
//
// `alias` 可以为 empty string
func CreateFriendEntry(db *gorm.DB, selfID int, friendID int, alias string) error {
	newEntry := FriendEntry{SelfID: selfID, FriendID: friendID}
	return db.Create(&newEntry).Error
}

// DeleteFriendEntryByID 删除一个单项好友项目
func DeleteFriendEntryByID(db *gorm.DB, entryID int) error {
	return db.Delete(&FriendEntry{Model: base.Model{ID: entryID}}).Error
}

// DeleteFriendEntry 删除一个单项好友项目
func DeleteFriendEntry(db *gorm.DB, entry *FriendEntry) error {
	return db.Delete(entry).Error
}
