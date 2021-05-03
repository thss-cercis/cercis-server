package user

import (
	"github.com/pkg/errors"
	"github.com/thss-cercis/cercis-server/db/base"
	"gorm.io/gorm"
)

type FriendApplyState int64

const (
	StateReject    FriendApplyState = -1
	StateUncertain FriendApplyState = 0
	StateAccept    FriendApplyState = 1
)

// FriendApply 好友申请项的 dao
type FriendApply struct {
	base.Model
	FromID int64            `gorm:"index:idx_from;" json:"from_id"`
	ToID   int64            `gorm:"index:idx_to;" json:"to_id"`
	State  FriendApplyState `gorm:"type:smallint;check:state >= -1 and state <= 1" json:"state"`
}

// GetFriendApplyByID 根据 id 获取好友申请
func GetFriendApplyByID(db *gorm.DB, applyID int64) (*FriendApply, error) {
	entry := new(FriendApply)
	err := db.First(entry, applyID).Error
	return entry, err
}

// GetUncertainFriendApply 获取一个待确定的好友申请
func GetUncertainFriendApply(db *gorm.DB, fromID int64, toID int64) (*FriendApply, error) {
	entry := new(FriendApply)
	err := db.First(entry, "from_id = ? AND to_id = ? AND state = ?", fromID, toID, StateUncertain).Error
	return entry, err
}

// GetFriendApplyFromByUserID 获得用户自己发送的好友申请列表
func GetFriendApplyFromByUserID(db *gorm.DB, userID int64) ([]FriendApply, error) {
	var arr []FriendApply
	err := db.Model(&User{Model: base.Model{ID: userID}}).Association("FriendApplyFrom").Find(&arr)
	return arr, err
}

// GetFriendApplyToByUserID 获得用户收到的好友申请列表
func GetFriendApplyToByUserID(db *gorm.DB, userID int64) ([]FriendApply, error) {
	var arr []FriendApply
	err := db.Model(&User{Model: base.Model{ID: userID}}).Association("FriendApplyTo").Find(&arr)
	return arr, err
}

// CreateFriendApply 创建一个新的待确定的好友申请
func CreateFriendApply(db *gorm.DB, fromID int64, toID int64) (*FriendApply, error) {
	tx := db.Begin()
	// 先检查是否已经为好友
	if _, err := GetFriendEntry(tx, fromID, toID); err == nil {
		tx.Rollback()
		return nil, errors.New("已经成为好友")
	}
	// 再检查是否已经存在未确认的申请
	if _, err := GetUncertainFriendApply(tx, fromID, toID); err == nil {
		tx.Rollback()
		return nil, errors.New("已经存在待确认的申请")
	}
	// 申请好友
	entry := &FriendApply{
		FromID: fromID,
		ToID:   toID,
		State:  StateUncertain,
	}
	if err := tx.Create(entry).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return entry, nil
}

// AcceptFriendApply 接受一个待确定的好友申请
func AcceptFriendApply(db *gorm.DB, applyID int64, userID int64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		entry, err := GetFriendApplyByID(tx, applyID)
		if err != nil || entry.State != StateUncertain || entry.ToID != userID {
			return errors.Errorf("Error: %v, %v", err, "获取待确定的好友申请失败")
		}
		// 设置 apply state
		entry.State = StateAccept
		if err := tx.Save(entry).Error; err != nil {
			return errors.Wrap(err, "更新好友申请状态失败")
		}
		// 插入双向好友项
		if err := tx.Create(&FriendEntry{
			SelfID:   entry.FromID,
			FriendID: entry.ToID,
		}).Error; err != nil {
			return errors.Wrap(err, "创建新好友项失败")
		}
		if err := tx.Create(&FriendEntry{
			SelfID:   entry.ToID,
			FriendID: entry.FromID,
		}).Error; err != nil {
			return errors.Wrap(err, "创建新好友项失败")
		}
		return nil
	})
}

// RejectFriendApply 拒绝一个待确定的好友申请
func RejectFriendApply(db *gorm.DB, applyID int64, userID int64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		entry, err := GetFriendApplyByID(tx, applyID)
		if err != nil || entry.State != StateUncertain || entry.ToID != userID {
			return errors.Errorf("Error: %v, %v", err, "获取待确定的好友申请失败")
		}
		// 设置 apply state
		entry.State = StateReject
		if err := tx.Save(entry).Error; err != nil {
			return errors.Wrap(err, "更新好友申请状态失败")
		}
		return nil
	})
}
