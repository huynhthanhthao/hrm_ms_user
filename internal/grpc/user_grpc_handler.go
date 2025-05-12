package userGrpc

import (
	"context"

	userpb "github.com/huynhthanhthao/hrm_user_service/generated"
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
	// Fix the GetAllUsers call by passing context
	users, err := s.userService.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}

	var res []*userpb.User
	for _, u := range users {
		res = append(res, &userpb.User{
			Id:        u.ID.String(),
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

	return &userpb.ListUsersResponse{Users: res}, nil
}

func (s *UserGRPCServer) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
	// Fetch the user by ID using the service
	user, err := s.userService.GetUser(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	// Map the user entity to the gRPC response
	return &userpb.GetUserResponse{
		User: &userpb.User{
			Id:        user.ID.String(),
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
	// Fetch the users by IDs using the service
	users, err := s.userService.GetUsersByIDs(ctx, req.Ids)
	if err != nil {
		return nil, err
	}

	// Map the user entities to the gRPC response
	var res []*userpb.User
	for _, u := range users {
		res = append(res, &userpb.User{
			Id:        u.ID.String(),
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

	return &userpb.GetUsersByIDsResponse{Users: res}, nil
}
