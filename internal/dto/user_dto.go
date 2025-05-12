package dto

type UserParams struct {
	IDs []string `json:"ids"`
	PaginationParams
}
