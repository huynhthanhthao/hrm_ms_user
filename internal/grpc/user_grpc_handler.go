package userGrpc

import (
	"context"

	userpb "github.com/huynhthanhthao/hrm_user_service/generated"
	"github.com/huynhthanhthao/hrm_user_service/internal/dto"
	"github.com/huynhthanhthao/hrm_user_service/internal/service"
)

type UserGRPCServer struct {
	userpb.UnimplementedUserServiceServer
	userService *service.UserService
}

func NewUserGRPCServer(us *service.UserService) *UserGRPCServer {
	return &UserGRPCServer{
		userService: us,
	}
}
func (s *UserGRPCServer) ListUsers(ctx context.Context, req *userpb.ListUsersRequest) (*userpb.ListUsersResponse, error) {
	users, err := s.userService.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}

	var res []*userpb.User
	for _, u := range users {
		res = append(res, &userpb.User{
			Id:        int32(u.ID), // đổi từ string sang int32
			FirstName: u.FirstName,
			LastName:  u.LastName,
			Gender:    string(u.Gender),
			Email:     u.Email,
			Phone:     u.Phone,
			WardCode:  u.WardCode,
			Address:   u.Address,
			CompanyId: u.CompanyID,
			CreatedAt: u.CreatedAt.String(),
			UpdatedAt: u.UpdatedAt.String(),
		})
	}

	return &userpb.ListUsersResponse{
		Users: res,
	}, nil
}

func (s *UserGRPCServer) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
	user, err := s.userService.GetUser(ctx, int(req.Id)) // Id là int32 trong proto
	if err != nil {
		return nil, err
	}

	return &userpb.GetUserResponse{
		User: &userpb.User{
			Id:        int32(user.ID),
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Gender:    string(user.Gender),
			Email:     user.Email,
			Phone:     user.Phone,
			WardCode:  user.WardCode,
			Address:   user.Address,
			CompanyId: user.CompanyID,
			CreatedAt: user.CreatedAt.String(),
			UpdatedAt: user.UpdatedAt.String(),
		},
	}, nil
}

func (s *UserGRPCServer) GetUsersByIDs(ctx context.Context, req *userpb.GetUsersByIDsRequest) (*userpb.GetUsersByIDsResponse, error) {
	intIDs := make([]int, len(req.Ids))
	for i, id := range req.Ids {
		intIDs[i] = int(id)
	}

	params := dto.UserParams{
		IDs: intIDs,
		PaginationParams: dto.PaginationParams{
			Page:     int(req.Page),
			PageSize: int(req.PageSize),
		},
	}

	users, totalCount, err := s.userService.GetUsersByIDs(ctx, params)
	if err != nil {
		return nil, err
	}

	var res []*userpb.User
	for _, u := range users {
		res = append(res, &userpb.User{
			Id:        int32(u.ID),
			FirstName: u.FirstName,
			LastName:  u.LastName,
			Gender:    string(u.Gender),
			Email:     u.Email,
			Phone:     u.Phone,
			WardCode:  u.WardCode,
			Address:   u.Address,
			CompanyId: u.CompanyID,
			CreatedAt: u.CreatedAt.String(),
			UpdatedAt: u.UpdatedAt.String(),
		})
	}

	return &userpb.GetUsersByIDsResponse{
		Users:      res,
		TotalCount: int32(totalCount),
	}, nil
}
