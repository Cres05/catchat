package model

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

type UserFriend struct {
	ID            int32                 `json:"id" gorm:"primarykey"`
	CreatedAt     time.Time             `json:"createAt"`
	UpdatedAt     time.Time             `json:"updatedAt"`
	DeletedAt     soft_delete.DeletedAt `json:"deletedAt"`
	UserAccount   string                `json:"userId" gorm:"index;comment:'用户ID'"`
	FriendAccount string                `json:"friendId" gorm:"index;comment:'好友ID'"`
}
