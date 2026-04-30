package user_manager

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
)

// userStatsAggregate 用户日志统计聚合结果。
type userStatsAggregate struct {
	// StatCount 统计次数。
	StatCount int64
	// StatQuota 统计额度。
	StatQuota int
	// StatTokens 统计 Tokens。
	StatTokens int
}

// userRecentPerformanceAggregate 用户最近性能聚合结果。
type userRecentPerformanceAggregate struct {
	// RecentRPM 最近 60 秒 RPM。
	RecentRPM int64
	// RecentTPM 最近 60 秒 TPM。
	RecentTPM int64
}

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

// GetUserStats 根据用户 ID 查询统计信息。
func GetUserStats(c *gin.Context) {
	userId, err := strconv.Atoi(c.Param("id"))
	if err != nil || userId <= 0 {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)

	user, err := model.GetUserById(userId, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	stats, err := queryUserStatsAggregate(userId, startTimestamp, endTimestamp)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	recent, err := queryUserRecentPerformance(userId)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	timeDiffMinutes := float64(endTimestamp-startTimestamp) / 60
	avgRPM := 0.0
	avgTPM := 0.0
	if startTimestamp > 0 && endTimestamp > startTimestamp && timeDiffMinutes > 0 {
		avgRPM = float64(stats.StatCount) / timeDiffMinutes
		avgTPM = float64(stats.StatTokens) / timeDiffMinutes
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": UserStatsResponse{
			UserId:   user.Id,
			Username: user.Username,
			AccountData: UserStatsAccountData{
				CurrentBalance:        user.Quota,
				HistoricalConsumption: user.UsedQuota,
			},
			UsageStats: UserStatsUsageData{
				RequestCount: user.RequestCount,
				StatCount:    stats.StatCount,
			},
			ResourceConsumption: UserStatsResourceData{
				StatQuota:  stats.StatQuota,
				StatTokens: stats.StatTokens,
			},
			PerformanceMetrics: UserStatsPerformanceData{
				AvgRPM:    avgRPM,
				AvgTPM:    avgTPM,
				RecentRPM: recent.RecentRPM,
				RecentTPM: recent.RecentTPM,
			},
		},
	})
}

// GetUserLogs 根据用户 ID 分页查询个人调用记录。
func GetUserLogs(c *gin.Context) {
	userId, err := strconv.Atoi(c.Param("id"))
	if err != nil || userId <= 0 {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}
	pageInfo := common.GetPageQuery(c)
	logType, _ := strconv.Atoi(c.Query("type"))
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	tokenName := c.Query("token_name")
	modelName := c.Query("model_name")
	group := c.Query("group")
	requestId := c.Query("request_id")

	logs, total, err := model.GetUserLogs(userId, logType, startTimestamp, endTimestamp, modelName, tokenName, pageInfo.GetStartIdx(), pageInfo.GetPageSize(), group, requestId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(logs)
	common.ApiSuccess(c, pageInfo)
}

// queryUserStatsAggregate 查询用户指定时间范围内的日志统计。
func queryUserStatsAggregate(userId int, startTimestamp int64, endTimestamp int64) (userStatsAggregate, error) {
	var stats userStatsAggregate
	query := model.LOG_DB.Model(&model.Log{}).Select(
		"COUNT(*) AS stat_count, COALESCE(SUM(quota), 0) AS stat_quota, COALESCE(SUM(prompt_tokens), 0) + COALESCE(SUM(completion_tokens), 0) AS stat_tokens",
	).Where("user_id = ? AND type = ?", userId, model.LogTypeConsume)
	if startTimestamp > 0 {
		query = query.Where("created_at >= ?", startTimestamp)
	}
	if endTimestamp > 0 {
		query = query.Where("created_at <= ?", endTimestamp)
	}
	if err := query.Scan(&stats).Error; err != nil {
		common.SysError("failed to query user stats: " + err.Error())
		return stats, err
	}
	return stats, nil
}

// queryUserRecentPerformance 查询用户最近 60 秒的 RPM/TPM。
func queryUserRecentPerformance(userId int) (userRecentPerformanceAggregate, error) {
	var stats userRecentPerformanceAggregate
	err := model.LOG_DB.Model(&model.Log{}).Select(
		"COUNT(*) AS recent_rpm, COALESCE(SUM(prompt_tokens), 0) + COALESCE(SUM(completion_tokens), 0) AS recent_tpm",
	).Where("user_id = ? AND type = ? AND created_at >= ?", userId, model.LogTypeConsume, time.Now().Add(-60*time.Second).Unix()).Scan(&stats).Error
	if err != nil {
		common.SysError("failed to query user recent performance: " + err.Error())
		return stats, err
	}
	return stats, nil
}
