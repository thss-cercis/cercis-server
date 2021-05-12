package activity

import (
	"errors"
	"github.com/thss-cercis/cercis-server/db/user"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
	"time"
)

type ActivityComment struct {
	ID          int64  `gorm:"primarykey" json:"id"`
	ActivityID  int64  `json:"activity_id"`
	CommenterID int64  `json:"commenter_id"`
	Content     string `gorm:"type:text not null" json:"content"`

	Commenter user.User `gorm:"foreignKey:CommenterID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`

	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"-"`
	DeletedAt soft_delete.DeletedAt `gorm:"index" json:"deleted_at"`
}

// CreateActivityComment 创建新的动态评论
func CreateActivityComment(db *gorm.DB, userID int64, content string, activityID int64) (*ActivityComment, error) {
	ac := &ActivityComment{
		ActivityID:  activityID,
		CommenterID: userID,
		Content:     content,
	}
	return ac, db.Save(ac).Error
}

// GetActivityCommentByID 根据 id 获得动态
func GetActivityCommentByID(db *gorm.DB, commentID int64) (*ActivityComment, error) {
	ac := &ActivityComment{}
	return ac, db.First(ac, commentID).Error
}

// GetActivityComments 获得动态的所有评论，至少返回一个空数组
func GetActivityComments(db *gorm.DB, activityID int64) ([]ActivityComment, error) {
	acs := make([]ActivityComment, 0)
	err := db.Where("activity_id = ?", activityID).Order("created_at asc").Find(&acs).Error
	if err != nil {
		return nil, err
	}
	return acs, nil
}

// DeleteActivityComment 删除动态评论，只有动态主人和评论主人可以删除
func DeleteActivityComment(db *gorm.DB, execID int64, commentID int64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		ac, err := GetActivityCommentByID(tx, commentID)
		if err != nil {
			return err
		}
		activity, err := GetActivity(tx, ac.ActivityID)
		if err != nil {
			return err
		}
		if execID != ac.CommenterID && execID != activity.SenderID {
			return errors.New("you have no permission to delete this comment")
		}
		return db.Delete(ac).Error
	})
}
