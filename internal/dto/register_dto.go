package dto

type RegisterDto struct {
	Username  string `json:"username" binding:"required,alphanum,min=3,max=20"`
	Password  string `json:"password" binding:"required,min=6,max=20"`
	FirstName string `json:"first_name" binding:"required,max=50"`
	LastName  string `json:"last_name" binding:"required,max=50"`
	Email     string `json:"email" binding:"required,email"`
	Avatar    string `json:"avatar"`
	Phone     string `json:"phone" binding:"required,numeric,min=10,max=15"`
	WardCode  string `json:"ward_code" binding:"required,numeric,min=3,max=10"`
	Address   string `json:"address" binding:"required,max=200"`
	Gender    string `json:"gender" binding:"required,oneof=other female male"`
}

type RegisterInput struct {
	Username  string
	Password  string
	FirstName string
	LastName  string
	Email     string
	Avatar    string
	Phone     string
	WardCode  string
	Address   string
	Gender    string
}
