package service

import (
	"chat-room/internal/dao/pool"
	"chat-room/pkg/common/response"
	"chat-room/pkg/errors"

	"chat-room/internal/model"

	"github.com/google/uuid"
)

type groupService struct {
}

var GroupService = new(groupService)

func (g *groupService) GetGroups(account string) ([]response.GroupResponse, error) {
	db := pool.GetDB()

	migrate := &model.Group{}
	pool.GetDB().AutoMigrate(&migrate)
	migrate2 := &model.GroupMember{}
	pool.GetDB().AutoMigrate(&migrate2)

	// var queryUser *model.User
	// db.First(&queryUser, "account = ?", account)

	// if queryUser.Id <= 0 {
	// 	return nil, errors.New("用户不存在")
	// }

	var groups []response.GroupResponse

	db.Raw("SELECT g.id AS group_id, g.uuid, g.created_at, g.name, g.notice FROM group_members AS gm LEFT JOIN `groups` AS g ON gm.group_id = g.id WHERE gm.account = ?",
		account).Scan(&groups)

	return groups, nil
}

func (g *groupService) SaveGroup(account string, group model.Group) {
	db := pool.GetDB()
	// var fromUser model.User
	// db.Find(&fromUser, "account = ?", Account)
	// if fromUser.Id <= 0 {
	// 	return
	// }

	group.Account = account
	group.Uuid = uuid.New().String()
	db.Save(&group)

	groupMember := model.GroupMember{
		Account: account,
		GroupId: group.ID,
		// Nickname: fromUser.Username,
		Mute: 0,
	}
	db.Save(&groupMember)
}

func (g *groupService) GetUserIdByGroupUuid(groupUuid string) []string {
	var group model.Group
	db := pool.GetDB()
	db.First(&group, "uuid = ?", groupUuid)
	if group.ID <= 0 {
		return nil
	}

	var accounts []string
	db.Raw("SELECT gm.account "+
		"FROM `groups` AS g "+
		"JOIN group_members AS gm ON gm.group_id = g.id "+
		"WHERE g.id = ?", group.ID).Scan(&accounts)
	return accounts
}

func (g *groupService) JoinGroup(groupUuid, userAccount string) error {
	// var user model.User
	db := pool.GetDB()
	// db.First(&user, "uuid = ?", userUuid)
	// if user.Id <= 0 {
	// 	return errors.New("用户不存在")
	// }

	var group model.Group
	db.First(&group, "uuid = ?", groupUuid)
	if group.ID <= 0 {
		return errors.New("群组不存在")
	}
	var groupMember model.GroupMember
	db.First(&groupMember, "account = ? and group_id = ?", userAccount, group.ID)
	if groupMember.ID > 0 {
		return errors.New("已经加入该群组")
	}
	// nickname := user.Nickname
	// if nickname == "" {
	// 	nickname = user.Username
	// }
	groupMemberInsert := model.GroupMember{
		Account: userAccount,
		GroupId: group.ID,
		// Nickname: nickname,
		Mute: 0,
	}
	db.Save(&groupMemberInsert)

	return nil
}
