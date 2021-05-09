package chat

import (
	"errors"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
	"time"
)

type MsgType int64

const (
	// MsgTypeText 纯文本消息种类
	MsgTypeText = iota
)

type Message struct {
	ID int64 `gorm:"primaryKey" json:"-"`
	// ChatID 消息所属的聊天
	ChatID int64 `gorm:"uniqueIndex:idx_chat_message_delete" json:"chat_id"`
	// MessageID 单个聊天中，MessageID 从 0 开始计数且独立
	MessageID int64 `gorm:"uniqueIndex:idx_chat_message_delete" json:"message_id"`

	Type    MsgType `gorm:"type:smallint not null;check:type >= 0" json:"type"`
	Message string  `gorm:"text not null" json:"message"`
	// SenderID 消息所属的用户，外键
	SenderID int64 `gorm:"type:bigint not null" json:"sender_id"`
	// IsWithdrawn 消息是否撤回
	IsWithdrawn bool `gorm:"type:boolean not null;default:false" json:"is_withdrawn"`

	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	DeletedAt soft_delete.DeletedAt `gorm:"uniqueIndex:idx_chat_message_delete" json:"-"`
}

// CreateMessage 创建一条新的信息，每个 chat 中都有自己独立的一套从 1 开始的 message_id
func CreateMessage(db *gorm.DB, chatID int64, senderID int64, typ MsgType, message string) (*Message, error) {
	tx := db.Begin()
	msg := &Message{}
	var id int64
	timeNow := time.Now()
	err := tx.Raw("INSERT INTO messages AS m1 (chat_id, message_id, type, message, sender_id, is_withdrawn, created_at, updated_at, deleted_at) "+
		"SELECT ?, COALESCE(MAX(m2.message_id),0)+1, ?, ?, ?, ?, ?, ?, ? FROM messages AS m2 WHERE m2.chat_id = ? AND m2.deleted_at = 0"+
		"RETURNING m1.id",
		chatID, typ, message, senderID, false, timeNow, timeNow, 0, chatID).Scan(&id).Error
	if err == nil && id != 0 {
		// 插入成功
		if err := tx.First(msg, id).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		return msg, tx.Commit().Error
	} else {
		tx.Rollback()
		return nil, err
	}
}

// GetMessage 使用 userID 的身份，获取从 chat 中某条消息.
func GetMessage(db *gorm.DB, chatID int64, userID int64, messageID int64) (*Message, error) {
	if !CheckIfInChat(db, chatID, userID) {
		return nil, errors.New("user is not in the chat")
	}
	msg := &Message{}
	if err := db.Where("chat_id = ? AND message_id = ?", chatID, messageID).First(msg).Error; err != nil {
		return nil, err
	}
	return msg, nil
}

// GetMessages 使用 userID 的身份，获取从 chat 中从 fromID(inclusive) 到 toID(exclusive) 的信息。
// 如果 fromID 为 0，表示从头开始；如果 toID 为 0，表示到末尾为止.
func GetMessages(db *gorm.DB, chatID int64, userID int64, fromID int64, toID int64) ([]Message, error) {
	tx := db.Begin()
	if !CheckIfInChat(tx, chatID, userID) {
		tx.Rollback()
		return nil, errors.New("user is not in the chat")
	}
	messages := make([]Message, 0)
	var tmp *gorm.DB
	if fromID == 0 && toID == 0 {
		tmp = tx.Where("chat_id = ?", chatID)
	} else if fromID != 0 && toID == 0 {
		tmp = tx.Where("chat_id = ? AND message_id >= ? ", chatID, fromID)
	} else if fromID == 0 && toID != 0 {
		tmp = tx.Where("chat_id = ? AND message_id < ?", chatID, toID)
	} else {
		tmp = tx.Where("chat_id = ? AND message_id >= ? AND message_id < ?", chatID, fromID, toID)
	}
	if err := tmp.Find(&messages).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	return messages, tx.Commit().Error
}

// GetLatestMessageID 获得某个 chat 的最新消息 id
func GetLatestMessageID(db *gorm.DB, chatID int64) (int64, error) {
	var ret int64
	if err := db.Model(&Message{}).Where("chat_id = ?", chatID).Select("MAX(message_id)").First(&ret).Error; err != nil {
		return 0, err
	}
	return ret, nil
}

// WithdrawMessage 使用 userID 的身份，撤回某一条消息
func WithdrawMessage(db *gorm.DB, chatID int64, userID int64, messageID int64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		message, err := GetMessage(tx, chatID, userID, messageID)
		if err != nil {
			return err
		}
		message.IsWithdrawn = true
		return tx.Save(message).Error
	})
}
