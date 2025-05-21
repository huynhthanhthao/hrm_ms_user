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
			Id:        int32(u.ID),
			FirstName: u.FirstName,
			LastName:  u.LastName,
			Gender:    string(u.Gender),
			Phone:     u.Phone,
			Email:     *u.Email,
			WardCode:  *u.WardCode,
			Address:   *u.Address,
			Avatar:    *u.Avatar,
			CreatedAt: u.CreatedAt.String(),
			UpdatedAt: u.UpdatedAt.String(),
		})
	}

	/*
		Gọi qua lấy permission và map res
	*/

	return &userpb.ListUsersResponse{
		Users: res,
	}, nil
}

func (s *UserGRPCServer) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
	user, err := s.userService.GetUser(ctx, int(req.Id))

	if err != nil {
		return nil, err
	}

	/*
		Gọi qua lấy permission và map res
	*/

	return &userpb.GetUserResponse{
		User: &userpb.User{
			Id:        int32(user.ID),
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Gender:    string(user.Gender),
			Email:     *user.Email,
			Phone:     user.Phone,
			WardCode:  *user.WardCode,
			Address:   *user.Address,
			Avatar:    *user.Avatar,
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

	users, err := s.userService.GetUsersByIDs(ctx, params)
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
			Email:     *u.Email,
			Phone:     u.Phone,
			WardCode:  *u.WardCode,
			Address:   *u.Address,
			Avatar:    *u.Avatar,
			CreatedAt: u.CreatedAt.String(),
			UpdatedAt: u.UpdatedAt.String(),
		})
	}

	return &userpb.GetUsersByIDsResponse{
		Users: res,
	}, nil
}

func (s *UserGRPCServer) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.CreateUserResponse, error) {
	user, err := s.userService.CreateUser(ctx, req)

	if err != nil {
		return nil, err
	}

	return &userpb.CreateUserResponse{
		User: &userpb.User{
			Id:        int32(user.ID),
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Gender:    string(user.Gender),
			Phone:     user.Phone,
			Email:     *user.Email,
			WardCode:  *user.WardCode,
			Address:   *user.Address,
			Avatar:    *user.Avatar,
			CreatedAt: user.CreatedAt.String(),
			UpdatedAt: user.UpdatedAt.String(),
		},
	}, nil
}

func (s *UserGRPCServer) UpdateUserByID(ctx context.Context, req *userpb.UpdateUserRequest) (*userpb.UpdateUserResponse, error) {
	input := &dto.UpdateUserDTO{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Gender:    req.Gender,
		Email:     req.Email,
		Phone:     req.Phone,
		WardCode:  req.WardCode,
		Address:   req.Address,
		Avatar:    req.Avatar,
		PermIDs:   req.PermIds,
		RoleIDs:   req.RoleIds,
	}

	tx, err := s.userService.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	user, err := s.userService.UpdateUserByID(ctx, tx, int(req.Id), input)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &userpb.UpdateUserResponse{
		User: &userpb.User{
			Id:        int32(user.ID),
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Gender:    string(user.Gender),
			Email:     *user.Email,
			Phone:     user.Phone,
			WardCode:  *user.WardCode,
			Address:   *user.Address,
			Avatar:    *user.Avatar,
			CreatedAt: user.CreatedAt.String(),
			UpdatedAt: user.UpdatedAt.String(),
		},
	}, nil
}

func (s *UserGRPCServer) DeleteUserByID(ctx context.Context, req *userpb.DeleteUserRequest) (*userpb.DeleteUserResponse, error) {
	err := s.userService.DeleteUserByID(ctx, int(req.Id))
	if err != nil {
		return nil, err
	}

	return &userpb.DeleteUserResponse{
		Success: true,
	}, nil
}
