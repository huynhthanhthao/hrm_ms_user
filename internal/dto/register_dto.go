package dto

type RegisterRequest struct {
	Username  string `json:"username" binding:"required,alphanum,min=3,max=20"`
	Password  string `json:"password" binding:"required,min=6,max=20"`
	FirstName string `json:"first_name" binding:"required,max=50"`
	LastName  string `json:"last_name" binding:"required,max=50"`
	Email     string `json:"email" binding:"required,email"`
	Phone     string `json:"phone" binding:"required,numeric,min=10,max=15"`
	WardCode  string `json:"ward_code" binding:"required,numeric,min=3,max=10"`
	Address   string `json:"address" binding:"required,max=200"`
	Gender    string `json:"gender" binding:"required,oneof=other female male"`
}

type AccountResponse struct {
	Username string `json:"username"`
}

type UserResponse struct {
	ID        int              `json:"id"`
	FirstName string           `json:"first_name"`
	LastName  string           `json:"last_name"`
	Email     string           `json:"email"`
	Phone     string           `json:"phone"`
	Address   string           `json:"address"`
	WardCode  string           `json:"ward_code"`
	Gender    string           `json:"gender"`
	Account   *AccountResponse `json:"account"`
}