package dto

type UserParams struct {
	IDs []int `json:"ids"`
	PaginationParams
}
type Account struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6,max=50"`
}

type CreateUserInput struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Gender    string `json:"gender"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	WardCode  string `json:"ward_code"`
	Address   string `json:"address"`
	Avatar    string `json:"avatar"`

	Account Account  `json:"account"`
	PermIDs []string `json:"perm_ids"`
	RoleIDs []string `json:"role_ids"`
}

type CreateUserDTO struct {
	FirstName string `json:"first_name" binding:"required,max=50"`
	LastName  string `json:"last_name" binding:"required,max=50"`
	Gender    string `json:"gender" binding:"required,oneof=male female other"`
	Email     string `json:"email" binding:"required,email"`
	Phone     string `json:"phone" binding:"required,numeric,min=10,max=15"`
	WardCode  string `json:"ward_code" binding:"required,numeric,min=3,max=10"`
	Address   string `json:"address" binding:"required,max=200"`
	Avatar    string `json:"avatar"`

	Account Account  `json:"account" binding:"required"`
	PermIDs []string `json:"perm_ids" binding:"omitempty,dive,required"`
	RoleIDs []string `json:"role_ids" binding:"omitempty,dive,required"`
}

type UpdateUserDTO struct {
	FirstName string `json:"first_name" binding:"omitempty,max=50"`
	LastName  string `json:"last_name" binding:"omitempty,max=50"`
	Gender    string `json:"gender" binding:"omitempty,oneof=male female other"`
	Email     string `json:"email" binding:"omitempty,email"`
	Phone     string `json:"phone" binding:"omitempty,numeric,min=10,max=15"`
	WardCode  string `json:"ward_code" binding:"omitempty,numeric,min=3,max=10"`
	Address   string `json:"address" binding:"omitempty,max=200"`
	Avatar    string `json:"avatar"`

	Account Account  `json:"account" binding:"omitempty"`
	PermIDs []string `json:"perm_ids" binding:"omitempty,dive"`
	RoleIDs []string `json:"role_ids" binding:"omitempty,dive"`
}
