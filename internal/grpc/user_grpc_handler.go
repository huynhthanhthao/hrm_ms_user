package userGrpc

import (
	"context"
	"strconv"

	"github.com/huynhthanhthao/hrm_user_service/ent"
	"github.com/huynhthanhthao/hrm_user_service/internal/dto"
	"github.com/huynhthanhthao/hrm_user_service/internal/service"
	userpb "github.com/huynhthanhthao/hrm_user_service/proto/user"
	"google.golang.org/protobuf/types/known/wrapperspb"
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

func entUserToProtoUser(u *ent.User) *userpb.User {
	if u == nil {
		return nil
	}
	var (
		email, wardCode, address, avatar *wrapperspb.StringValue
	)
	if u.Email != nil {
		email = wrapperspb.String(*u.Email)
	}
	if u.WardCode != nil {
		wardCode = wrapperspb.String(*u.WardCode)
	}
	if u.Address != nil {
		address = wrapperspb.String(*u.Address)
	}
	if u.Avatar != nil {
		avatar = wrapperspb.String(*u.Avatar)
	}
	return &userpb.User{
		Id:        int32(u.ID),
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Gender:    string(u.Gender),
		Phone:     wrapperspb.String(u.Phone),
		Email:     email,
		WardCode:  wardCode,
		Address:   address,
		Avatar:    avatar,
		CreatedAt: u.CreatedAt.String(),
		UpdatedAt: u.UpdatedAt.String(),
	}
}

func (s *UserGRPCServer) ListUsers(ctx context.Context, req *userpb.ListUsersRequest) (*userpb.ListUsersResponse, error) {
	users, err := s.userService.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}

	var res []*userpb.User
	for _, u := range users {
		res = append(res, entUserToProtoUser(u))
	}

	/*
		Gọi qua lấy permission và map res
	*/

	return &userpb.ListUsersResponse{
		Users: res,
	}, nil
}

func (s *UserGRPCServer) GetUserById(ctx context.Context, req *userpb.GetUserByIdRequest) (*userpb.GetUserByIdResponse, error) {
	user, err := s.userService.GetUserById(ctx, int(req.Id))
	if err != nil {
		return nil, err
	}

	// Lấy vai trò
	rolesResp, err := s.userService.GetUserRolesByUserId(ctx, strconv.Itoa(int(req.Id)))
	if err != nil {
		return nil, err
	}

	// Lấy quyền
	permsResp, err := s.userService.GetUserPermsByUserId(ctx, strconv.Itoa(int(req.Id)))
	if err != nil {
		return nil, err
	}

	// Map roles
	var roles []*userpb.RoleExt
	for _, r := range rolesResp.Roles {
		roles = append(roles, &userpb.RoleExt{
			Id:          r.Id,
			Code:        r.Code,
			Name:        r.Name,
			Color:       r.Color,
			Description: r.Description,
			CreatedAt:   r.CreatedAt,
			UpdatedAt:   r.UpdatedAt,
		})
	}

	// Map permissions
	var perms []*userpb.PermExt
	for _, p := range permsResp.Perms {
		perms = append(perms, &userpb.PermExt{
			Id:          p.Id,
			Code:        p.Code,
			Name:        p.Name,
			Description: p.Description,
		})
	}

	return &userpb.GetUserByIdResponse{
		User:  entUserToProtoUser(user),
		Roles: roles,
		Perms: perms,
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
		res = append(res, entUserToProtoUser(u))
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
		User: entUserToProtoUser(user),
	}, nil
}

func (s *UserGRPCServer) UpdateUserByID(ctx context.Context, req *userpb.UpdateUserRequest) (*userpb.UpdateUserResponse, error) {
	tx, err := s.userService.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	user, err := s.userService.UpdateUserByID(ctx, tx, int(req.Id), req)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &userpb.UpdateUserResponse{
		User: entUserToProtoUser(user),
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
