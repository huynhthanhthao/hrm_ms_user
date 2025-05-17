package service

import (
	"context"
	"fmt"

	"github.com/huynhthanhthao/hrm_user_service/ent"
	"github.com/huynhthanhthao/hrm_user_service/ent/user"
	"github.com/huynhthanhthao/hrm_user_service/internal/dto"
)

type UserService struct {
	client    *ent.Client
	hrClients *HRServiceClients
	perClients *PermissionServiceClients
}

func NewUserService(
	client *ent.Client,
	hrClients *HRServiceClients,
	perClients *PermissionServiceClients,
) (*UserService, error) {
	return &UserService{
		client:     client,
		hrClients:  hrClients,
		perClients: perClients,
	}, nil
}

func (s *UserService) BeginTx(ctx context.Context) (*ent.Tx, error) {
	return s.client.Tx(ctx)
}

func (s *UserService) GetAllUsers(ctx context.Context) ([]*ent.User, error) {
	users, err := s.client.User.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve users: %w", err)
	}
	return users, nil
}

func (s *UserService) GetUser(ctx context.Context, id int) (*ent.User, error) {
	// Use the converted UUID
	user, err := s.client.User.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user: %w", err)
	}
	return user, nil
}

func (s *UserService) GetUsersByIDs(ctx context.Context, params dto.UserParams) ([]*ent.User, int, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 20
	}

	offset := (params.Page - 1) * params.PageSize

	// Convert string IDs to int
	intIDs := make([]int, len(params.IDs))
	for i, id := range params.IDs {
		intIDs[i] = int(id)
	}

	// Query users by IDs with pagination
	users, err := s.client.User.Query().
		Where(user.IDIn(intIDs...)).
		Offset(offset).
		Limit(params.PageSize).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve users: %w", err)
	}

	// Get total count
	totalCount, err := s.client.User.Query().
		Where(user.IDIn(intIDs...)).
		Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	return users, totalCount, nil
}
func (s *UserService) CreateUser(ctx context.Context, input *dto.CreateUserDTO) (*ent.User, error) {
	tx, err := s.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	user, err := tx.User.Create().
		SetFirstName(input.FirstName).
		SetLastName(input.LastName).
		SetGender(user.Gender(input.Gender)).
		SetEmail(input.Email).
		SetPhone(input.Phone).
		SetWardCode(input.WardCode).
		SetAddress(input.Address).
		SetAvatar(input.Avatar).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	/*
		Call grpc to permission service here 
	*/

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) UpdateUserByID(ctx context.Context, tx *ent.Tx, userID int, input *dto.CreateUserDTO) (*ent.User, error) {
	/* 
		Call grpc to permission service  
	*/

	user, err := tx.User.UpdateOneID(userID).
		SetFirstName(input.FirstName).
		SetLastName(input.LastName).
		SetGender(user.Gender(input.Gender)).
		SetEmail(input.Email).
		SetPhone(input.Phone).
		SetWardCode(input.WardCode).
		SetAddress(input.Address).
		SetAvatar(input.Avatar).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) DeleteUserByID(ctx context.Context, id int) error {
	tx, err := s.client.Tx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := tx.User.DeleteOneID(id).Exec(ctx); err != nil {
		return err
	}

	if err := tx.Account.DeleteOneID(id).Exec(ctx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	/* 
		Call grpc to permission service 
	*/

	return nil
}
