package chat

import (
	"errors"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
	"strconv"
	"time"
)

type MsgType int64

const (
	// MsgTypeText 纯文本消息种类
	MsgTypeText = 0
	// MsgTypeImage 图片消息
	MsgTypeImage = 1
	// MsgTypeAudio 音频消息
	MsgTypeAudio = 2
	// MsgTypeVideo 视频消息
	MsgTypeVideo = 3
	// MsgTypeGeo 位置消息
	MsgTypeGeo = 4
	// MsgTypeWithdraw 撤回消息
	MsgTypeWithdraw = 100
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

	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	DeletedAt soft_delete.DeletedAt `gorm:"uniqueIndex:idx_chat_message_delete" json:"-"`
}

// CreateMessage 创建一条新的信息，每个 chat 中都有自己独立的一套从 1 开始的 message_id
func CreateMessage(db *gorm.DB, chatID int64, senderID int64, typ MsgType, message string) (*Message, error) {
	msg := &Message{}
	return msg, db.Transaction(func(tx *gorm.DB) error {
		var id int64
		timeNow := time.Now()
		err := tx.Raw("INSERT INTO messages AS m1 (chat_id, message_id, type, message, sender_id, is_withdrawn, created_at, updated_at, deleted_at) "+
			"SELECT ?, COALESCE(MAX(m2.message_id),0)+1, ?, ?, ?, ?, ?, ?, ? FROM messages AS m2 WHERE m2.chat_id = ? AND m2.deleted_at = 0"+
			"RETURNING m1.id",
			chatID, typ, message, senderID, false, timeNow, timeNow, 0, chatID).Scan(&id).Error
		if err == nil && id != 0 {
			// 插入成功
			if err := tx.First(msg, id).Error; err != nil {
				return err
			}
			return nil
		} else {
			return err
		}
	})
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

// GetLatestMessages 获得某个用户给定某些聊天的最新消息
func GetLatestMessages(db *gorm.DB, userID int64, chatIDs []int64) ([]Message, error) {
	var ret = make([]Message, 0)
	if chatIDs == nil || len(chatIDs) == 0 {
		return ret, nil
	}
	if !CheckIfUserInChats(db, userID, chatIDs) {
		return nil, errors.New("user is not in some of the chats")
	}
	// 获取最新信息
	err := db.Raw("SELECT * FROM messages AS m WHERE (m.chat_id, m.message_id) IN "+
		"(SELECT chat_id, MAX(message_id) FROM messages WHERE chat_id IN ? GROUP BY chat_id)",
		chatIDs).
		Scan(&ret).Error
	return ret, err
}

// GetAllChatsLatestMessageID 获得某个用户所有的聊天的最新消息 id
func GetAllChatsLatestMessageID(db *gorm.DB, userID int64) ([]struct {
	ChatID       int64 `json:"chat_id"`
	MaxMessageID int64 `json:"max_message_id"`
}, error) {
	chats, err := GetAllChats(db, userID)
	if err != nil {
		return nil, err
	}
	chatIDs := make([]int64, 0)
	for _, chats := range chats {
		chatIDs = append(chatIDs, chats.ID)
	}
	// 获取最新信息
	var ret = make([]struct {
		ChatID       int64 `json:"chat_id"`
		MaxMessageID int64 `json:"max_message_id"`
	}, 0)
	err = db.Model(&Message{}).
		Select("chat_id, MAX(message_id) AS max_message_id").
		Where("chat_id IN ?", chatIDs).
		Group("chat_id").
		Find(&ret).Error
	return ret, err
}

// CheckIsWithdrawn 判断消息是否被撤回
func CheckIsWithdrawn(db *gorm.DB, chatID int64, messageID int64) bool {
	var count int64
	if err := db.Model(&Message{}).
		Where("chat_id = ? AND type = ? AND message = ?", chatID, MsgTypeWithdraw, strconv.Itoa(int(messageID))).
		Count(&count).Error; err != nil {
		return false
	}
	return count >= 1
}

// WithdrawMessage 使用 userID 的身份，撤回某一条消息
func WithdrawMessage(db *gorm.DB, chatID int64, userID int64, messageID int64) (*Message, error) {
	var msg *Message
	return msg, db.Transaction(func(tx *gorm.DB) error {
		message, err := GetMessage(db, chatID, userID, messageID)
		if err != nil {
			return err
		}
		if CheckIsWithdrawn(tx, chatID, messageID) {
			return errors.New("the message is already withdrawn")
		}
		if message.SenderID != userID {
			return errors.New("could not withdraw other one's message")
		}
		if message.Type == MsgTypeWithdraw {
			return errors.New("could not withdraw a withdraw message")
		}
		msg, err = CreateMessage(db, chatID, userID, MsgTypeWithdraw, strconv.Itoa(int(messageID)))
		if err != nil {
			return err
		}
		return nil
	})
}
