package dto

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	FullName string `json:"full_name" binding:"required"`
}

type RegisterResponse struct {
	UserID    int    `json:"user_id"`
	AccountID int    `json:"account_id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
}