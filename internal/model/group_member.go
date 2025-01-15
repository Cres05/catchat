package model

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

type GroupMember struct {
	ID        int32                 `json:"id" gorm:"primarykey"`
	CreatedAt time.Time             `json:"createAt"`
	UpdatedAt time.Time             `json:"updatedAt"`
	DeletedAt soft_delete.DeletedAt `json:"deletedAt"`
	Account   string                `json:"userId" gorm:"index;comment:'用户ID'"`
	GroupId   int32                 `json:"groupId" gorm:"index;comment:'群组ID'"`
	// Nickname  string                `json:"nickname" gorm:"type:varchar(350);comment:'昵称"`
	Mute int16 `json:"mute" gorm:"comment:'是否禁言'"`
}
