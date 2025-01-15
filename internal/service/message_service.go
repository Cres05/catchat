package service

import (
	"chat-room/internal/dao/pool"
	"chat-room/pkg/common/constant"
	"chat-room/pkg/common/response"
	"chat-room/pkg/errors"
	"chat-room/pkg/global/log"
	"chat-room/pkg/protocol"
	"fmt"

	"chat-room/internal/model"
	"chat-room/pkg/common/request"

	"gorm.io/gorm"
)

const NULL_ID int32 = 0

type messageService struct {
}

var MessageService = new(messageService)

func (m *messageService) GetMessages(message request.MessageRequest) ([]response.MessageResponse, error) {
	db := pool.GetDB()

	// 自动迁移消息表结构
	migrate := &model.Message{}
	db.AutoMigrate(&migrate)

	if message.MessageType == constant.MESSAGE_TYPE_USER {
		// 查询单聊消息
		return m.fetchUserMessages(db, message)
	}

	if message.MessageType == constant.MESSAGE_TYPE_GROUP {
		// 查询群组消息
		return m.fetchGroupMessages(db, message)
	}

	return nil, errors.New("不支持查询类型")
}

func (m *messageService) fetchUserMessages(db *gorm.DB, message request.MessageRequest) ([]response.MessageResponse, error) {
	// 用 account 作为唯一标识，直接处理消息逻辑

	// 检查发送方和接收方的账户名是否存在
	if message.Account == "" || message.ToAccount == "" {
		return nil, errors.New("发送方或接收方账户不能为空")
	}

	// 开启事务
	tx := db.Begin()

	// 标记为已读（接收方为当前用户的消息）
	if err := tx.Model(&model.Message{}).
		Where("from_account = ? AND to_account = ? AND is_read = ?", message.ToAccount, message.Account, 0).
		Update("is_read", 1).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("更新消息为已读失败: %w", err)
	}

	// 查询消息记录
	var messages []response.MessageResponse
	if err := tx.Raw(`
		SELECT 
			m.id, m.from_account, m.to_account, m.content, m.content_type, m.url, m.created_at 
		FROM messages AS m 
		WHERE (m.from_account = ? AND m.to_account = ?) OR (m.from_account = ? AND m.to_account = ?) 
		ORDER BY m.created_at ASC`,
		message.Account, message.ToAccount, message.ToAccount, message.Account).Scan(&messages).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("查询消息失败: %w", err)
	}

	// 提交事务
	tx.Commit()

	return messages, nil
}

func (m *messageService) fetchGroupMessages(db *gorm.DB, message request.MessageRequest) ([]response.MessageResponse, error) {
	// 查询群组
	var group model.Group
	db.First(&group, "uuid = ?", message.ToAccount)
	if group.ID == 0 {
		return nil, errors.New("群组不存在")
	}

	var groupMember model.GroupMember
	if err := db.First(&groupMember, "group_id = ? AND account = ?", group.ID, message.Account).Error; err != nil {
		return nil, errors.New("用户不在群组中或无权限")
	}

	// 开启事务
	tx := db.Begin()

	// 查找未读消息
	// var unreadMessages []model.Message
	// if err := tx.Raw(`
	// 	SELECT m.id
	// 	FROM messages AS m
	// 	WHERE m.to_account = ? AND m.message_type = ?`, group.Uuid, constant.MESSAGE_TYPE_GROUP, message.Account).Scan(&unreadMessages).Error; err != nil {
	// 	tx.Rollback()
	// 	return nil, fmt.Errorf("查询未读消息失败: %w", err)
	// }

	// 标记未读消息为已读
	// for _, msg := range unreadMessages {
	// 	if err := tx.Exec(`
	// 		INSERT INTO message_read_status (message_id, user_id, is_read, read_at)
	// 		VALUES (?, ?, 1, NOW())
	// 		ON DUPLICATE KEY UPDATE is_read = 1, read_at = NOW()`,
	// 		msg.ID, message.Account).Error; err != nil {
	// 		tx.Rollback()
	// 		return nil, fmt.Errorf("标记群组消息为已读失败: %w", err)
	// 	}
	// }

	// 查询群组消息记录
	var messages []response.MessageResponse
	if err := tx.Raw(`
		SELECT 
			m.id, m.from_account, m.to_account, m.content, m.content_type, m.url, m.created_at 
		FROM messages AS m 
		WHERE m.message_type = ? AND m.to_account = ? 
		ORDER BY m.created_at ASC`, constant.MESSAGE_TYPE_GROUP, group.Uuid).Scan(&messages).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("查询群组消息失败: %w", err)
	}

	// 提交事务
	tx.Commit()

	return messages, nil
}

func (m *messageService) SaveMessage(message protocol.Message) {
	db := pool.GetDB()
	var fromUser model.User
	db.Find(&fromUser, "account = ?", message.From)
	if NULL_ID == fromUser.Id {
		log.Logger.Error("SaveMessage not find from user", log.Any("SaveMessage not find from user", fromUser.Id))
		return
	}

	var ToAccount string = ""

	if message.MessageType == constant.MESSAGE_TYPE_USER {
		var toUser model.User
		db.Find(&toUser, "account = ?", message.To)
		if NULL_ID == toUser.Id {
			return
		}
		ToAccount = toUser.Account
	}

	if message.MessageType == constant.MESSAGE_TYPE_GROUP {
		var group model.Group
		db.Find(&group, "uuid = ?", message.To)
		if NULL_ID == group.ID {
			return
		}
		ToAccount = group.Uuid
	}

	saveMessage := model.Message{
		FromAccount: fromUser.Account,
		ToAccount:   ToAccount,
		Content:     message.Content,
		ContentType: int16(message.ContentType),
		MessageType: int16(message.MessageType),
		Url:         message.Url,
	}
	db.Save(&saveMessage)
}
