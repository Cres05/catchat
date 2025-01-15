package service

import (
	"time"

	"chat-room/internal/dao/pool"
	"chat-room/internal/model"
	"chat-room/pkg/common/request"
	"chat-room/pkg/common/response"
	"chat-room/pkg/errors"
	"chat-room/pkg/global/log"

	"github.com/google/uuid"
)

type userService struct {
}

var UserService = new(userService)

func (u *userService) Register(user *model.User) error {
	db := pool.GetDB()
	var userCount int64
	db.Model(user).Where("username", user.Username).Count(&userCount)
	if userCount > 0 {
		return errors.New("user already exists")
	}
	user.Account = uuid.New().String()
	user.CreateAt = time.Now()
	user.DeleteAt = 0

	db.Create(&user)
	return nil
}

func (u *userService) Login(user *model.User) bool {
	pool.GetDB().AutoMigrate(&user)
	log.Logger.Debug("user", log.Any("user in service", user))
	db := pool.GetDB()

	var queryUser *model.User
	db.First(&queryUser, "username = ?", user.Username)
	log.Logger.Debug("queryUser", log.Any("queryUser", queryUser))

	user.Account = queryUser.Account

	return queryUser.Password == user.Password
}

func (u *userService) ModifyUserInfo(user *model.User) error {
	var queryUser *model.User
	db := pool.GetDB()
	db.First(&queryUser, "username = ?", user.Username)
	log.Logger.Debug("queryUser", log.Any("queryUser", queryUser))
	var nullId int32 = 0
	if nullId == queryUser.Id {
		return errors.New("用户不存在")
	}
	queryUser.Nickname = user.Nickname
	queryUser.Email = user.Email
	queryUser.Password = user.Password

	db.Save(queryUser)
	return nil
}

func (u *userService) GetUserDetails(uuid string) model.User {
	var queryUser *model.User
	db := pool.GetDB()
	db.Select("account", "username", "nickname", "avatar").First(&queryUser, "account = ?", uuid)
	return *queryUser
}

// 通过名称查找群组或者用户
func (u *userService) GetUserOrGroupByName(name string) response.SearchResponse {
	var queryUser *model.User
	db := pool.GetDB()
	db.Select("uuid", "username", "nickname", "avatar").First(&queryUser, "username = ?", name)

	var queryGroup *model.Group
	db.Select("uuid", "name").First(&queryGroup, "name = ?", name)

	search := response.SearchResponse{
		User:  *queryUser,
		Group: *queryGroup,
	}
	return search
}

func (u *userService) GetUserList(uuid string) []model.User {
	db := pool.GetDB()

	var queryUser *model.User
	db.First(&queryUser, "account = ?", uuid)
	var nullId int32 = 0
	if nullId == queryUser.Id {
		return nil
	}

	var queryUsers []model.User
	db.Raw("SELECT u.username, u.account, u.avatar FROM user_friends AS uf JOIN users AS u ON uf.friend_id = u.id WHERE uf.user_id = ?", queryUser.Id).Scan(&queryUsers)

	return queryUsers
}

func (u *userService) AddFriend(userFriendRequest *request.FriendRequest) error {
	db := pool.GetDB()

	// 使用事务
	tx := db.Begin()

	// // 查询用户
	// var queryUser model.User
	// if err := tx.First(&queryUser, "uuid = ?", userFriendRequest.Uuid).Error; err != nil {
	// 	tx.Rollback()
	// 	return errors.New("用户不存在")
	// }

	// // 查询好友
	// var friend model.User
	// if err := tx.First(&friend, "username = ?", userFriendRequest.FriendUsername).Error; err != nil {
	// 	tx.Rollback()
	// 	return errors.New("好友不存在")
	// }

	// 检查好友关系
	var userFriendQuery model.UserFriend
	if err := tx.First(&userFriendQuery, "user_account = ? and friend_account = ?", userFriendRequest.Account, userFriendRequest.FriendAccount).Error; err == nil {
		tx.Rollback()
		return errors.New("该用户已经是你好友")
	}

	// 保存好友关系
	userFriend := model.UserFriend{
		UserAccount:   userFriendRequest.Account,
		FriendAccount: userFriendRequest.FriendAccount,
	}
	if err := tx.Create(&userFriend).Error; err != nil {
		tx.Rollback()
		return errors.New("添加好友失败")
	}

	// 提交事务
	tx.Commit()

	log.Logger.Debug("userFriend", log.Any("userFriend", userFriend))
	return nil
}

// 修改头像
func (u *userService) ModifyUserAvatar(avatar string, userUuid string) error {
	var queryUser *model.User
	db := pool.GetDB()
	db.First(&queryUser, "uuid = ?", userUuid)

	if NULL_ID == queryUser.Id {
		return errors.New("用户不存在")
	}

	db.Model(&queryUser).Update("avatar", avatar)
	return nil
}
