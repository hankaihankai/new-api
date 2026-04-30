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
	// RecentRPM 最近 60 秒 RPM。
	RecentRPM int64 `json:"recent_rpm"`
	// RecentTPM 最近 60 秒 TPM。
	RecentTPM int64 `json:"recent_tpm"`
}
