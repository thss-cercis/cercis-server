package base

import (
	"time"

	"gorm.io/gorm"
)

// Model 数据库 dao 类型的基类
type Model struct {
	ID        int64 `gorm:"primarykey" json:"id"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
