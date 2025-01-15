package v1

import (
	"chat-room/internal/model"
	"chat-room/internal/service"
	"chat-room/pkg/common/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 获取分组列表
func GetGroup(c *gin.Context) {
	account := c.Param("account")
	groups, err := service.GroupService.GetGroups(account)
	if err != nil {
		c.JSON(http.StatusOK, response.FailMsg(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.SuccessMsg(groups))
}

// 保存分组列表
func SaveGroup(c *gin.Context) {
	account := c.Param("account")
	var group model.Group
	c.ShouldBindJSON(&group)

	service.GroupService.SaveGroup(account, group)
	c.JSON(http.StatusOK, response.SuccessMsg(nil))
}

// 加入组别
func JoinGroup(c *gin.Context) {
	userAccount := c.Param("account")
	groupUuid := c.Param("groupUuid")
	err := service.GroupService.JoinGroup(groupUuid, userAccount)
	if err != nil {
		c.JSON(http.StatusOK, response.FailMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.SuccessMsg(nil))
}

// 获取组内成员信息
func GetGroupUsers(c *gin.Context) {
	groupUuid := c.Param("groupUuid")
	users := service.GroupService.GetUserIdByGroupUuid(groupUuid)
	c.JSON(http.StatusOK, response.SuccessMsg(users))
}
