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
	SelfID   int64  `gorm:"uniqueIndex:idx_composited_id;index:idx_self" json:"self_id"`
	FriendID int64  `gorm:"uniqueIndex:idx_composited_id;index:idx_friend" json:"friend_id"`
	Alias    string `gorm:"type:varChar(127) not null" json:"alias"`
}

// CreateFriendEntry 创建一个单向好友列表项目
//
// `alias` 可以为 empty string
func CreateFriendEntry(db *gorm.DB, selfID int64, friendID int64, alias string) error {
	newEntry := FriendEntry{SelfID: selfID, FriendID: friendID}
	return db.Create(&newEntry).Error
}

// GetFriendEntryBi 查找两个好友
func GetFriendEntryBi(db *gorm.DB, selfID int64, friendID int64) (*FriendEntry, error) {
	entry := new(FriendEntry)
	err := db.Where("self_id = ? AND friend_id = ?", selfID, friendID).First(entry).Error
	return entry, err
}

// GetFriendEntryByUserID 获得用户的好友列表，因为是双向维护的，因此只需要获取一份
func GetFriendEntryByUserID(db *gorm.DB, userID int64) ([]FriendEntry, error) {
	var arr []FriendEntry
	err := db.Model(&User{Model: base.Model{ID: userID}}).Association("FriendEntrySelf").Find(&arr)
	return arr, err
}

// DeleteFriendEntryByID 删除一个单项好友项目
func DeleteFriendEntryByID(db *gorm.DB, entryID int64) error {
	return db.Delete(&FriendEntry{}, entryID).Error
}

// DeleteFriendEntry 删除一个单项好友项目
func DeleteFriendEntry(db *gorm.DB, entry *FriendEntry) error {
	return db.Delete(entry).Error
}
