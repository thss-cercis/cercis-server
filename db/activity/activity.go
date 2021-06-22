package activity

import (
	"errors"
	"github.com/thss-cercis/cercis-server/db/user"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
	"time"
)

type Activity struct {
	ID       int64  `gorm:"primarykey" json:"id"`
	Text     string `gorm:"type:text not null" json:"text"`
	SenderID int64  `json:"sender_id"`

	Media    []ActivityMedium  `gorm:"foreignKey:ActivityID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"media"`
	Comments []ActivityComment `gorm:"foreignKey:ActivityID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"comments"`
	// ThumbUps 点赞者
	ThumbUps []ActivityThumbUp `gorm:"foreignKey:ActivityID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"thumb_ups"`
	Sender   user.User         `gorm:"foreignKey:SenderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`

	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"-"`
	DeletedAt soft_delete.DeletedAt `gorm:"index" json:"deleted_at"`
}

type MediumType int64

const (
	// MediumTypeImageURL 图片类型 url
	MediumTypeImageURL = 0
	// MediumTypeVideoURL 视频类型 url
	MediumTypeVideoURL = 1
	// MediumTypeGeo 地理位置 url
	MediumTypeGeo = 2
)

type ActivityMedium struct {
	ID         int64      `gorm:"primarykey" json:"id"`
	ActivityID int64      `json:"activity_id"`
	Type       MediumType `gorm:"type:smallint not null" json:"type"`
	Content    string     `gorm:"type:text not null" json:"content"`

	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"-"`
	DeletedAt soft_delete.DeletedAt `gorm:"index" json:"deleted_at"`
}

type MediumCapsule struct {
	Type    MediumType `json:"type" validate:"required,gte=0"`
	Content string     `json:"content" validate:"required"`
}

// CreateActivity 创建新动态，media 可以为 nil
func CreateActivity(db *gorm.DB, userID int64, text string, media []MediumCapsule) (*Activity, error) {
	// 生成 media
	m := make([]ActivityMedium, 0)
	if media != nil {
		for _, medium := range media {
			m = append(m, ActivityMedium{
				Type:    medium.Type,
				Content: medium.Content,
			})
		}
	}
	// 生成动态
	activity := &Activity{
		SenderID: userID,
		Text:     text,
		Media:    m,
	}
	return activity, db.Save(activity).Error
}

// GetActivity 获得动态，并且 preload 评论和 media
func GetActivity(db *gorm.DB, activityID int64) (*Activity, error) {
	activity := &Activity{}
	return activity, db.Preload("Media").Preload("Comments").Preload("ThumbUps").
		First(activity, activityID).Error
}

// GetActivitiesBefore 获得某个用户能够接受到的，在某个 startID 之前的所有动态，如果 count 为零值，则获取所有
func GetActivitiesBefore(db *gorm.DB, userID int64, activityID int64, count int64) ([]Activity, error) {
	acs := make([]Activity, 0)
	if count < 0 {
		return acs, nil
	}
	friends, err := user.GetFriendEntrySelfByUserID(db, userID)
	if err != nil {
		return nil, err
	}
	friendIDs := make([]int64, 0)
	for _, friend := range friends {
		friendIDs = append(friendIDs, friend.FriendID)
	}
	// 加入自己
	friendIDs = append(friendIDs, userID)
	tmp := db.Model(&Activity{}).Where("id < ? AND sender_id IN ?", activityID, friendIDs).
		Preload("Media").Preload("Comments").Preload("ThumbUps").Order("id desc")
	if count != 0 {
		tmp = tmp.Limit(int(count))
	}
	if err := tmp.Find(&acs).Error; err != nil {
		return nil, err
	}

	return acs, nil
}

// GetActivitiesAfter 获得某个用户能够接受到的，在某个 ID 之后的所有动态，如果 count 为零值，则获取所有
func GetActivitiesAfter(db *gorm.DB, userID int64, activityID int64, count int64) ([]Activity, error) {
	acs := make([]Activity, 0)
	if count < 0 {
		return acs, nil
	}
	friends, err := user.GetFriendEntrySelfByUserID(db, userID)
	if err != nil {
		return nil, err
	}
	friendIDs := make([]int64, 0)
	for _, friend := range friends {
		friendIDs = append(friendIDs, friend.FriendID)
	}
	// 加入自己
	friendIDs = append(friendIDs, userID)
	tmp := db.Model(&Activity{}).Where("id > ? AND sender_id IN ?", activityID, friendIDs).
		Preload("Media").Preload("Comments").Preload("ThumbUps").Order("id asc")
	if count != 0 {
		tmp = tmp.Limit(int(count))
	}
	if err := tmp.Find(&acs).Error; err != nil {
		return nil, err
	}

	return acs, nil
}

// DeleteActivity 删除动态
func DeleteActivity(db *gorm.DB, execID int64, activityID int64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		activity, err := GetActivity(tx, activityID)
		if err != nil {
			return err
		}
		if activity.SenderID != execID {
			return errors.New("you have insufficient permission to delete this activity")
		}
		return tx.Select("Media", "Comments").Delete(activity).Error
	})
}
