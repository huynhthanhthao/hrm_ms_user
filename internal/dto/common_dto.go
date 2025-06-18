package dto

type PaginationParams struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
