package dto

type UserParams struct {
	IDs []int `json:"ids"`
	PaginationParams
}
type CreateUserInput struct {
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Gender    string   `json:"gender"`
	Email     string   `json:"email"`
	Phone     string   `json:"phone"`
	WardCode  string   `json:"ward_code"`
	Address   string   `json:"address"`
	Avatar    string   `json:"avatar"`
	CompanyID int      `json:"company_id"`
	PermIDs   []string `json:"perm_ids"`
	RoleIDs   []string `json:"role_ids"`
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
	CompanyID int    `json:"company_id" binding:"required"`
	PermIDs   []int  `json:"perm_ids" binding:"omitempty,dive,required"`
	RoleIDs   []int  `json:"role_ids" binding:"omitempty,dive,required"`
}
