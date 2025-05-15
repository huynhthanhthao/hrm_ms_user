package dto

type UserParams struct {
	IDs []string `json:"ids"`
	PaginationParams
}

type CreateUserInput struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Gender    string `json:"gender"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	WardCode  string `json:"ward_code"`
	Address   string `json:"address"`
	CompanyID string `json:"company_id"`
}
