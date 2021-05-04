package user

// Copyright 2021 AyajiLin
// 好友列表项的数据库定义

import (
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
	"time"
)

// FriendEntry 好友列表项的 dao
type FriendEntry struct {
	ID       int64  `gorm:"primarykey" json:"id"`
	SelfID   int64  `gorm:"uniqueIndex:idx_composited_id;index:idx_self" json:"self_id"`
	FriendID int64  `gorm:"uniqueIndex:idx_composited_id;index:idx_friend" json:"friend_id"`
	Alias    string `gorm:"type:varChar(127) not null" json:"alias"`

	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	DeletedAt soft_delete.DeletedAt `gorm:"uniqueIndex:idx_composited_id" json:"deleted_at"`
}

// CreateFriendEntry 创建一个单向好友列表项目
//
// `alias` 可以为 empty string
func CreateFriendEntry(db *gorm.DB, selfID int64, friendID int64, alias string) error {
	newEntry := FriendEntry{SelfID: selfID, FriendID: friendID}
	return db.Create(&newEntry).Error
}

// GetFriendEntry 查找两个好友
func GetFriendEntry(db *gorm.DB, selfID int64, friendID int64) (*FriendEntry, error) {
	entry := new(FriendEntry)
	err := db.Where("self_id = ? AND friend_id = ?", selfID, friendID).First(entry).Error
	return entry, err
}

// GetFriendEntrySelfByUserID 获得用户自己的好友列表
func GetFriendEntrySelfByUserID(db *gorm.DB, userID int64) ([]FriendEntry, error) {
	var arr []FriendEntry
	err := db.Model(&User{ID: userID}).Association("FriendEntrySelf").Find(&arr)
	return arr, err
}

// ModifyFriendEntryAlias 修改好友备注名
func ModifyFriendEntryAlias(db *gorm.DB, selfID int64, friendID int64, newAlias string) (*FriendEntry, error) {
	tx := db.Begin()
	entry, err := GetFriendEntry(tx, selfID, friendID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	entry.Alias = newAlias
	if err := tx.Save(entry).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	return entry, tx.Commit().Error
}

// DeleteFriendEntryByID 删除一个单项好友项目
func DeleteFriendEntryByID(db *gorm.DB, entryID int64) error {
	return db.Delete(&FriendEntry{}, entryID).Error
}

// DeleteFriendEntryBi 双向删除好友
func DeleteFriendEntryBi(db *gorm.DB, userID1 int64, userID2 int64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		entry1, err := GetFriendEntry(db, userID1, userID2)
		if err != nil {
			return err
		}
		entry2, err := GetFriendEntry(db, userID2, userID1)
		if err != nil {
			return err
		}
		if err := db.Delete(entry1).Error; err != nil {
			return err
		}
		if err := db.Delete(entry2).Error; err != nil {
			return err
		}
		return nil
	})
}

// DeleteFriendEntry 删除一个单项好友项目
func DeleteFriendEntry(db *gorm.DB, entry *FriendEntry) error {
	return db.Delete(entry).Error
}
