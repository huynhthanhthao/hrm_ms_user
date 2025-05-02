package mapper

import (
	"user/ent"
	"user/internal/dto"
)

func MapUserToResponse(u *ent.User) *dto.UserResponse {
    return &dto.UserResponse{
        ID:        u.ID,
        FirstName: u.FirstName,
        LastName:  u.LastName,
        Email:     u.Email,
        Phone:     u.Phone,
        Address:   u.Address,
        WardCode:  u.WardCode,
        Gender:    string(u.Gender),
        Account: &dto.AccountResponse{
            Username: u.Edges.Account.Username,
        },
    }
}


func MapLoginToResponse(access_token String, refresh_token string, u *ent.User) *dto.UserResponse {
    return {
        access_token,       
        refresh_token,
    }
}

