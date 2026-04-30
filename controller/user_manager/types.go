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
