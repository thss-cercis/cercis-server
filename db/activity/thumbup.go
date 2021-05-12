package activity

import (
	"github.com/thss-cercis/cercis-server/db/user"
	"gorm.io/gorm"
)

type ActivityThumbUp struct {
	ActivityID int64      `gorm:"primaryKey" json:"activity_id"`
	UserID     int64      `gorm:"primaryKey" json:"user_id"`
	User       *user.User `gorm:"foreignKey:UserID" json:"-"`
}

// AddActivityThumbUp 动态点赞
func AddActivityThumbUp(db *gorm.DB, activityID int64, userID int64) error {
	return db.Model(&ActivityThumbUp{}).Create(&ActivityThumbUp{
		ActivityID: activityID,
		UserID:     userID,
	}).Error
}

// DeleteActivityThumbUp 取消动态点赞
func DeleteActivityThumbUp(db *gorm.DB, activityID int64, userID int64) error {
	return db.Model(&ActivityThumbUp{}).Delete(&ActivityThumbUp{
		ActivityID: activityID,
		UserID:     userID,
	}).Error
}
