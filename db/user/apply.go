package user

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
	"time"
)

type FriendApplyState int64

const (
	StateReject    FriendApplyState = -1
	StateUncertain FriendApplyState = 0
	StateAccept    FriendApplyState = 1
)

// FriendApply 好友申请项的 dao
type FriendApply struct {
	ID     int64 `gorm:"primarykey" json:"id"`
	FromID int64 `gorm:"type:bigint not null;index:idx_from;" json:"from_id"`
	ToID   int64 `gorm:"type:bigint not null;index:idx_to;" json:"to_id"`
	// Alias 表示申请人给接受者预设的备注
	Alias string `gorm:"type:varChar(127) not null" json:"alias"`
	// Remark
	Remark string           `gorm:"type:varChar(255) not null" json:"remark"`
	State  FriendApplyState `gorm:"type:smallint not null;check:state >= -1 and state <= 1" json:"state"`

	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	DeletedAt soft_delete.DeletedAt `gorm:"index" json:"deleted_at"`
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
	err := db.Model(&User{ID: userID}).Association("FriendApplyFrom").Find(&arr)
	return arr, err
}

// GetFriendApplyToByUserID 获得用户收到的好友申请列表
func GetFriendApplyToByUserID(db *gorm.DB, userID int64) ([]FriendApply, error) {
	var arr []FriendApply
	err := db.Model(&User{ID: userID}).Association("FriendApplyTo").Find(&arr)
	return arr, err
}

// CreateFriendApply 创建一个新的待确定的好友申请
// alias 表示发送者给接受者的预设备注
func CreateFriendApply(db *gorm.DB, fromID int64, toID int64, alias string, remark string) (*FriendApply, error) {
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
		Alias:  alias,
		Remark: remark,
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

// AcceptFriendApply 接受一个待确定的好友申请, alias 表示接受者给申请人的备注
func AcceptFriendApply(db *gorm.DB, applyID int64, userID int64, alias string) error {
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
		// 删除对方给自己的待定好友申请
		entryOpposite, err := GetUncertainFriendApply(db, entry.ToID, entry.FromID)
		if err == nil && entryOpposite != nil {
			entryOpposite.State = StateAccept
			db.Save(entryOpposite)
		}
		// 插入双向好友项
		if err := tx.Create(&FriendEntry{
			SelfID:   entry.FromID,
			FriendID: entry.ToID,
			Alias:    entry.Alias,
		}).Error; err != nil {
			return errors.Wrap(err, "创建新好友项失败")
		}
		if err := tx.Create(&FriendEntry{
			SelfID:   entry.ToID,
			FriendID: entry.FromID,
			Alias:    alias,
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
