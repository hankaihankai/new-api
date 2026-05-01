package user_manager

// CreateUserRequest 外部用户管理创建用户请求。
type CreateUserRequest struct {
	// Username 用户名。
	Username string `json:"username" validate:"max=20"`
	// Password 密码。
	Password string `json:"password" validate:"min=8,max=20"`
	// DisplayName 显示名称。
	DisplayName string `json:"display_name" validate:"max=20"`
}

// CreateUserResponse 外部用户管理创建用户响应。
type CreateUserResponse struct {
	// UserId 用户 ID。
	UserId int `json:"user_id"`
	// Username 用户名。
	Username string `json:"username"`
	// TokenName 默认令牌名称。
	TokenName string `json:"token_name"`
	// TokenKey 默认令牌完整密钥。
	TokenKey string `json:"token_key"`
}

// UserStatsResponse 用户统计信息响应。
type UserStatsResponse struct {
	// UserId 用户 ID。
	UserId int `json:"user_id"`
	// Username 用户名。
	Username string `json:"username"`
	// AccountData 账户数据。
	AccountData UserStatsAccountData `json:"account_data"`
	// UsageStats 使用统计。
	UsageStats UserStatsUsageData `json:"usage_stats"`
	// ResourceConsumption 资源消耗。
	ResourceConsumption UserStatsResourceData `json:"resource_consumption"`
	// PerformanceMetrics 性能指标。
	PerformanceMetrics UserStatsPerformanceData `json:"performance_metrics"`
}

// UserStatsAccountData 用户账户数据。
type UserStatsAccountData struct {
	// CurrentBalance 当前余额。
	CurrentBalance int `json:"current_balance"`
	// HistoricalConsumption 历史消耗。
	HistoricalConsumption int `json:"historical_consumption"`
	// Quota 当前剩余额度原始值。
	Quota int `json:"quota"`
	// UsedQuota 已使用额度原始值。
	UsedQuota int `json:"used_quota"`
	// TotalQuota 总额度原始值。
	TotalQuota int `json:"total_quota"`
	// CurrentBalanceAmount 当前余额金额。
	CurrentBalanceAmount float64 `json:"current_balance_amount"`
	// UsedQuotaAmount 已使用额度金额。
	UsedQuotaAmount float64 `json:"used_quota_amount"`
	// TotalQuotaAmount 总额度金额。
	TotalQuotaAmount float64 `json:"total_quota_amount"`
	// QuotaPerUnit 额度换算基数。
	QuotaPerUnit float64 `json:"quota_per_unit"`
}

// UserStatsUsageData 用户使用统计。
type UserStatsUsageData struct {
	// RequestCount 请求次数。
	RequestCount int `json:"request_count"`
	// StatCount 统计次数。
	StatCount int64 `json:"stat_count"`
}

// UserStatsResourceData 用户资源消耗。
type UserStatsResourceData struct {
	// StatQuota 统计额度。
	StatQuota int `json:"stat_quota"`
	// StatTokens 统计 Tokens。
	StatTokens int `json:"stat_tokens"`
}

// UserStatsPerformanceData 用户性能指标。
type UserStatsPerformanceData struct {
	// AvgRPM 平均 RPM。
	AvgRPM float64 `json:"avg_rpm"`
	// AvgTPM 平均 TPM。
	AvgTPM float64 `json:"avg_tpm"`
}

// SetUserQuotaRequest 设置用户额度请求。
type SetUserQuotaRequest struct {
	// Mode 操作模式：add（增加）、subtract（减少）、override（覆盖）。
	Mode string `json:"mode" validate:"required,oneof=add subtract override"`
	// Value 额度值。
	Value int `json:"value" validate:"required"`
}
