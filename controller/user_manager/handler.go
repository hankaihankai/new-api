package user_manager

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
)

// userStatsAggregate 用户用量数据聚合结果。
type userStatsAggregate struct {
	// StatCount 统计次数。
	StatCount int64
	// StatQuota 统计额度。
	StatQuota int
	// StatTokens 统计 Tokens。
	StatTokens int
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

	stats, err := queryUserQuotaDataAggregate(userId, startTimestamp, endTimestamp)
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

	totalQuota := user.Quota + user.UsedQuota
	quotaPerUnit := common.QuotaPerUnit
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": UserStatsResponse{
			UserId:   user.Id,
			Username: user.Username,
			AccountData: UserStatsAccountData{
				CurrentBalance:        user.Quota,
				HistoricalConsumption: user.UsedQuota,
				Quota:                 user.Quota,
				UsedQuota:             user.UsedQuota,
				TotalQuota:            totalQuota,
				CurrentBalanceAmount:  quotaToAmount(user.Quota, quotaPerUnit),
				UsedQuotaAmount:       quotaToAmount(user.UsedQuota, quotaPerUnit),
				TotalQuotaAmount:      quotaToAmount(totalQuota, quotaPerUnit),
				QuotaPerUnit:          quotaPerUnit,
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
				AvgRPM: avgRPM,
				AvgTPM: avgTPM,
			},
		},
	})
}

// GetUserQuotaRecords 根据用户 ID 分页查询个人额度调用记录。
func GetUserQuotaRecords(c *gin.Context) {
	userId, err := strconv.Atoi(c.Param("id"))
	if err != nil || userId <= 0 {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}
	pageInfo := common.GetPageQuery(c)
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	modelName := c.Query("model_name")

	records, total, err := model.GetUserQuotaRecords(userId, startTimestamp, endTimestamp, modelName, pageInfo.GetStartIdx(), pageInfo.GetPageSize())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(records)
	common.ApiSuccess(c, pageInfo)
}

// queryUserQuotaDataAggregate 查询用户指定时间范围内的用量统计。
func queryUserQuotaDataAggregate(userId int, startTimestamp int64, endTimestamp int64) (userStatsAggregate, error) {
	var stats userStatsAggregate
	query := model.DB.Model(&model.QuotaData{}).Select(
		"COALESCE(SUM(count), 0) AS stat_count, COALESCE(SUM(quota), 0) AS stat_quota, COALESCE(SUM(token_used), 0) AS stat_tokens",
	).Where("user_id = ?", userId)
	if startTimestamp > 0 {
		query = query.Where("created_at >= ?", startTimestamp)
	}
	if endTimestamp > 0 {
		query = query.Where("created_at <= ?", endTimestamp)
	}
	if err := query.Scan(&stats).Error; err != nil {
		common.SysError("failed to query user quota data stats: " + err.Error())
		return stats, err
	}
	return stats, nil
}

// quotaToAmount 将数据库额度换算为金额数值。
func quotaToAmount(quota int, quotaPerUnit float64) float64 {
	if quotaPerUnit <= 0 {
		return 0
	}
	return float64(quota) / quotaPerUnit
}
