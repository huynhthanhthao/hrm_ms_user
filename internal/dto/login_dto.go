package dto

import "github.com/huynhthanhthao/hrm_user_service/ent"

type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	User         *ent.User `json:"user"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required,alphanum,min=3,max=20"`
	Password string `json:"password" binding:"required"`
}
