package dto

import "user/ent"

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         *ent.User    `json:"user"`
	Account      *ent.Account `json:"account"`
}


type LoginRequest struct {
	Username  string `json:"username" binding:"required,alphanum,min=3,max=20"`
	Password  string `json:"password" binding:"required"`
}