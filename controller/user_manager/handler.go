package user_manager

import (
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
)

// CreateUserWithDefaultToken 创建普通用户并同时创建默认无限额度令牌。
func CreateUserWithDefaultToken(c *gin.Context) {
	var req CreateUserRequest
	if err := common.DecodeJson(c.Request.Body, &req); err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	if req.Username == "" || req.Password == "" {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}
	if req.DisplayName == "" {
		req.DisplayName = req.Username
	}
	if err := common.Validate.Struct(&req); err != nil {
		common.ApiErrorI18n(c, i18n.MsgUserInputInvalid, map[string]any{"Error": err.Error()})
		return
	}

	exist, err := model.CheckUserExistOrDeleted(req.Username, "")
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgDatabaseError)
		common.SysLog("CheckUserExistOrDeleted error: " + err.Error())
		return
	}
	if exist {
		common.ApiErrorI18n(c, i18n.MsgUserExists)
		return
	}

	key, err := common.GenerateKey()
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgTokenGenerateFailed)
		common.SysLog("failed to generate token key: " + err.Error())
		return
	}

	user := model.User{
		Username:    req.Username,
		Password:    req.Password,
		DisplayName: req.DisplayName,
		Role:        common.RoleCommonUser,
		Group:       "default",
	}
	tx := model.DB.Begin()
	if tx.Error != nil {
		common.ApiError(c, tx.Error)
		return
	}
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	if err := user.InsertWithTx(tx, 0); err != nil {
		common.ApiError(c, err)
		return
	}
	token := model.Token{
		UserId:             user.Id,
		Name:               "default",
		Key:                key,
		CreatedTime:        common.GetTimestamp(),
		AccessedTime:       common.GetTimestamp(),
		ExpiredTime:        -1,
		RemainQuota:        0,
		UnlimitedQuota:     true,
		ModelLimitsEnabled: false,
		ModelLimits:        "",
		Group:              "default",
	}
	if err := tx.Create(&token).Error; err != nil {
		common.ApiError(c, err)
		return
	}
	if err := tx.Commit().Error; err != nil {
		common.ApiError(c, err)
		return
	}
	committed = true
	user.FinalizeOAuthUserCreation(0)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": CreateUserResponse{
			UserId:    user.Id,
			Username:  user.Username,
			TokenName: token.Name,
			TokenKey:  "sk-" + token.GetFullKey(),
		},
	})
}
