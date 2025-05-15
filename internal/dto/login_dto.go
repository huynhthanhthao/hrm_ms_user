package dto

type LoginDto struct {
	Username string `json:"username" binding:"required,alphanum,min=3,max=20"`
	Password string `json:"password" binding:"required"`
}

type LoginInput struct {
	Username string
	Password string
}
