package chat

import (
	mapset "github.com/deckarep/golang-set"
	"github.com/pkg/errors"
	"github.com/thss-cercis/cercis-server/db/user"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
	"time"
)

type ChatType int64

const (
	// ChatTypePrivate 私聊
	ChatTypePrivate = iota
	// ChatTypeGroup 群聊
	ChatTypeGroup
)

type MemberPermission int64

const (
	// PermNormal 狗群员
	PermNormal = iota
	// PermAdmin 狗管理
	PermAdmin
	// PermOwner 狗群主
	PermOwner
)

// Chat 聊天列表项的 dao
type Chat struct {
	ID     int64    `gorm:"primarykey" json:"id"`
	Type   ChatType `gorm:"type:smallint not null;check:type >= 0 and type <= 1" json:"type"`
	Name   string   `gorm:"type:varChar(127) not null" json:"name"`
	Avatar string   `gorm:"type:varChar(255) not null" json:"avatar"`

	Members  []user.User `gorm:"many2many:chat_users;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	Messages []Message   `gorm:"foreignKey:ChatID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`

	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	DeletedAt soft_delete.DeletedAt `gorm:"index" json:"deleted_at"`
}

type ChatUser struct {
	ID     int64 `gorm:"primaryKey" json:"-"`
	ChatID int64 `gorm:"uniqueIndex:idx_chat_user_delete" json:"chat_id"`
	UserID int64 `gorm:"uniqueIndex:idx_chat_user_delete" json:"user_id"`
	// Alias 群内备注
	Alias string `gorm:"type:varChar(127) not null" json:"alias"`
	// 群内权限
	Permission MemberPermission `gorm:"type:smallint not null;check:permission >= 0;default:0;" json:"permission"`

	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"-"`
	DeletedAt soft_delete.DeletedAt `gorm:"uniqueIndex:idx_chat_user_delete" json:"deleted_at"`
}

// CreatePrivateChat 创建私人聊天，若已经存在私聊，则返回此私聊
func CreatePrivateChat(db *gorm.DB, user1 int64, user2 int64) (*Chat, error) {
	tx := db.Begin()
	tmp := &Chat{}
	// 先确定没有已经创建的私聊
	err := db.Raw("SELECT * FROM chats WHERE chats.type = ? AND chats.deleted_at = 0 "+
		"AND EXISTS (SELECT * FROM chat_users AS cu1 WHERE cu1.chat_id = chats.id AND cu1.user_id = ? AND cu1.deleted_at = 0) "+
		"AND EXISTS (SELECT * FROM chat_users AS cu2 WHERE cu2.chat_id = chats.id AND cu2.user_id = ? AND cu2.deleted_at = 0)",
		ChatTypePrivate, user1, user2).
		Scan(tmp).Error
	if err == nil && tmp.ID != 0 {
		// 已经存在这样一个私聊
		return nil, errors.New("private chat is already existed")
	}
	// 创建新私聊
	chat := &Chat{
		Type: ChatTypePrivate,
		Members: []user.User{
			user.User{ID: user1},
			user.User{ID: user2},
		},
	}
	err = db.Create(chat).Error
	if err != nil {
		return nil, err
	}
	return chat, tx.Commit().Error
}

// CreateGroupChat 创建群聊，members 为成员 UserID 的集合，允许为空
func CreateGroupChat(db *gorm.DB, name string, ownerID int64, members mapset.Set) (*Chat, error) {
	var chat *Chat = nil
	tx := db.Begin()
	chat = &Chat{
		Type: ChatTypeGroup,
		Name: name,
	}
	// 加入狗群主
	chat.Members = append(chat.Members, user.User{ID: ownerID})
	// 加入狗群员
	if members != nil {
		for member := range members.Iter() {
			userID, ok := member.(int64)
			if !ok {
				tx.Rollback()
				return nil, errors.New("id of members of group chat is invalid")
			}
			chat.Members = append(chat.Members, user.User{ID: userID})
		}
	}
	if err := tx.Create(chat).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	// 更改狗群主的权限
	if err := tx.Model(&ChatUser{}).Where("chat_id = ? AND user_id = ?", chat.ID, ownerID).Update("permission", PermOwner).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	return chat, tx.Commit().Error
}

// CheckIfInChat 判断用户是否在属于某个聊天
func CheckIfInChat(db *gorm.DB, chatID int64, userID int64) bool {
	_, err := GetChatMember(db, chatID, userID)
	if err != nil {
		return false
	}
	return true
}

func CheckIfUserInChats(db *gorm.DB, userID int64, chatIDs []int64) bool {
	chatsLen := len(chatIDs)
	if chatIDs == nil || chatsLen == 0 {
		return true
	}
	var count int64
	if err := db.Model(&ChatUser{}).Where("user_id = ? AND chat_id IN ?", userID, chatIDs).Count(&count).Error; err != nil {
		return false
	}
	return int64(chatsLen) == count
}

// GetChat 获得聊天
func GetChat(db *gorm.DB, chatID int64) (*Chat, error) {
	chat := &Chat{ID: chatID}
	return chat, db.Model(chat).First(chat).Error
}

// GetPrivateChat 获得私聊
func GetPrivateChat(db *gorm.DB, user1 int64, user2 int64) (*Chat, error) {
	chat := &Chat{}
	// 先确定没有已经创建的私聊
	err := db.Raw("SELECT * FROM chats WHERE chats.type = ? AND chats.deleted_at = 0 "+
		"AND EXISTS (SELECT * FROM chat_users AS cu1 WHERE cu1.chat_id = chats.id AND cu1.user_id = ? AND cu1.deleted_at = 0) "+
		"AND EXISTS (SELECT * FROM chat_users AS cu2 WHERE cu2.chat_id = chats.id AND cu2.user_id = ? AND cu2.deleted_at = 0)",
		ChatTypePrivate, user1, user2).
		Scan(chat).Error
	if err == nil && chat.ID != 0 {
		// 已经存在这样一个私聊
		return chat, nil
	}
	return nil, errors.New("private chat not created yet")
}

// GetAllChats 获得一个人加入的所有聊天
func GetAllChats(db *gorm.DB, userID int64) ([]Chat, error) {
	chats := make([]Chat, 0)
	err := db.Raw("SELECT * FROM chats WHERE chats.deleted_at = 0 "+
		"AND EXISTS (SELECT * FROM chat_users AS cu WHERE cu.chat_id = chats.id AND cu.user_id = ? AND cu.deleted_at = 0)",
		userID).
		Scan(&chats).Error
	if err != nil {
		return nil, err
	}
	return chats, nil
}

func (chat *Chat) UpdateFrom(db *gorm.DB) error {
	return db.First(chat, chat.ID).Error
}

func (chat *Chat) UpdateTo(db *gorm.DB) error {
	return db.Save(chat).Error
}

// DeleteChat 删除聊天，execID 表示删除的发起人，会检查权限
func DeleteChat(db *gorm.DB, execID int64, chatID int64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		chat, err := GetChat(tx, chatID)
		if err != nil {
			return err
		}

		if chat.Type == ChatTypePrivate {
			if _, err := GetChatMember(tx, chatID, execID); err != nil {
				return err
			}
			return tx.Select("Members", "Messages").Delete(chat).Error
		} else {
			chatMember, err := GetChatMember(tx, chatID, execID)
			if err != nil {
				return err
			}
			if chatMember.Permission != PermOwner {
				return errors.New("insufficient permission")
			}
			return tx.Select("Members", "Messages").Delete(chat).Error
		}
	})
}

// AddChatMember 通过用户 UserID 和聊天 UserID 加入新成员
func AddChatMember(db *gorm.DB, chatID int64, userID int64) (*ChatUser, error) {
	tx := db.Begin()
	chat, err := GetChat(tx, chatID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if chat.Type == ChatTypePrivate {
		return nil, errors.New("could not add member to private chat")
	}
	chatUser := &ChatUser{
		ChatID:     chatID,
		UserID:     userID,
		Permission: PermNormal,
	}
	err = tx.Create(chatUser).Error
	if err != nil {
		return nil, err
	}
	return chatUser, tx.Commit().Error
}

// GetChatMember 获得某个群成员表项
func GetChatMember(db *gorm.DB, chatID int64, userID int64) (*ChatUser, error) {
	chatUser := &ChatUser{}
	return chatUser, db.Where("chat_id = ? AND user_id = ?", chatID, userID).First(chatUser).Error
}

// GetChatMembers 获得一个群的所有成员表项
func GetChatMembers(db *gorm.DB, chatID int64) ([]ChatUser, error) {
	ret := make([]ChatUser, 0)
	return ret, db.Find(&ret, "chat_id = ?", chatID).Error
}

// ModifyChatMemberPermission 修改成员权限
func ModifyChatMemberPermission(db *gorm.DB, execID int64, chatID int64, userID int64, newPerm MemberPermission) error {
	return db.Transaction(func(tx *gorm.DB) error {
		chat, err := GetChat(tx, chatID)
		if err != nil {
			return err
		} else if chat.Type == ChatTypePrivate {
			return errors.New("could not modify permission of private chat member")
		}

		execUser, err := GetChatMember(tx, chatID, execID)
		if err != nil {
			return err
		}
		modifiedUser, err := GetChatMember(tx, chatID, userID)
		if err != nil {
			return err
		}
		if modifiedUser.Permission >= execUser.Permission || newPerm >= execUser.Permission {
			return errors.New("you have no permission to do this")
		}

		modifiedUser.Permission = newPerm
		return db.Save(modifiedUser).Error
	})
}

// ModifyChatMemberAlias 修改成员名片
func ModifyChatMemberAlias(db *gorm.DB, chatID int64, userID int64, alias string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		chat, err := GetChat(tx, chatID)
		if err != nil {
			return err
		} else if chat.Type == ChatTypePrivate {
			return errors.New("could not modify alias of private chat member")
		}

		return db.Model(&ChatUser{}).
			Where("chat_id = ? AND user_id = ?", chatID, userID).
			Update("alias", alias).Error
	})
}

// ChangeGroupOwner 禅让群主
func ChangeGroupOwner(db *gorm.DB, execID int64, chatID int64, userID int64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		chat, err := GetChat(tx, chatID)
		if err != nil {
			return err
		} else if chat.Type == ChatTypePrivate {
			return errors.New("could not set new owner for private chat member")
		}

		execUser, err := GetChatMember(tx, chatID, execID)
		if err != nil {
			return err
		}
		if execUser.Permission != PermOwner {
			return errors.New("you are not allowed to set new owner")
		}
		modifiedUser, err := GetChatMember(tx, chatID, userID)
		if err != nil {
			return err
		}
		execUser.Permission = PermNormal
		modifiedUser.Permission = PermOwner
		if err := db.Save(execUser).Error; err != nil {
			return err
		}
		return db.Save(modifiedUser).Error
	})
}

// DeleteChatMember 通过用户 UserID 和聊天 UserID 删除成员，execID 为执行者的 UserID，会查询权限。可以删除自身.
func DeleteChatMember(db *gorm.DB, execID int64, chatID int64, userID int64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		chat, err := GetChat(tx, chatID)
		if err != nil {
			return err
		} else if chat.Type != ChatTypeGroup {
			return errors.New("could not delete member of group chat")
		}

		execUser, err := GetChatMember(tx, chatID, execID)
		if err != nil {
			return err
		}
		delUser, err := GetChatMember(tx, chatID, userID)
		if err != nil {
			return err
		}
		if delUser.Permission == PermOwner {
			return errors.New("could not remove the owner of the group chat")
		}
		if delUser.Permission >= execUser.Permission && execID != userID {
			return errors.New("you have no permission to do this")
		}
		return tx.Delete(&ChatUser{}, "chat_id = ? AND user_id = ?", chatID, userID).Error
	})
}
