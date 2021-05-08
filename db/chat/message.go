package chat

import (
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
	"time"
)

type Message struct {
	ID int64 `gorm:"primaryKey" json:"id"`
	// ChatID 消息所属的聊天
	ChatID int64 `gorm:"uniqueIndex:idx_chat_message_delete" json:"chat_id"`
	// MessageID 单个聊天中，MessageID 从 0 开始计数且独立
	MessageID int64 `gorm:"uniqueIndex:idx_chat_message_delete" json:"message_id"`

	Type int64 `gorm:"type:smallint not null;check:type >= 0" json:"type"`
	// SenderID 消息所属的用户，外键
	SenderID int64 `gorm:"type:bigint not null" json:"sender_id"`
	// IsWithdrawn 消息是否撤回
	IsWithdrawn bool `gorm:"type:boolean not null;default:false" json:"is_withdrawn"`

	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	DeletedAt soft_delete.DeletedAt `gorm:"uniqueIndex:idx_chat_message_delete" json:"deleted_at"`
}

func CreateMessage(db *gorm.DB, chatID int64, message Message) (*Message, error) {
	// TODO
	return nil, nil
}
