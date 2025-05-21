package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/huynhthanhthao/hrm_user_service/ent"
	"github.com/huynhthanhthao/hrm_user_service/ent/user"
	"github.com/huynhthanhthao/hrm_user_service/generated"
	"github.com/huynhthanhthao/hrm_user_service/internal/dto"
)

type UserService struct {
	client     *ent.Client
	hrClients  *HRServiceClients
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
		return nil, fmt.Errorf("#1 GetAllUsers: failed to retrieve users: %w", err)
	}
	return users, nil
}

func (s *UserService) GetUser(ctx context.Context, id int) (*ent.User, error) {
	// Use the converted UUID
	user, err := s.client.User.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("#1 GetUser: failed to retrieve user: %w", err)
	}
	return user, nil
}

func (s *UserService) GetUsersByIDs(ctx context.Context, params dto.UserParams) ([]*ent.User, error) {
	// Convert string IDs to int
	intIDs := make([]int, len(params.IDs))
	for i, id := range params.IDs {
		intIDs[i] = int(id)
	}

	// Query users by IDs with pagination
	users, err := s.client.User.Query().
		Where(user.IDIn(intIDs...)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("#1 GetUsersByIDs: failed to retrieve users: %w", err)
	}

	return users, nil
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
		return nil, fmt.Errorf("#1 CreateUser: failed when create user: %w", err)
	}

	// Call grpc to permission service here
	if len(input.PermIDs) > 0 {
		userPermRequests := make([]*generated.CreateUserPermRequest, len(input.PermIDs))
		for i, permID := range input.PermIDs {
			parsedUUID, err := uuid.Parse(permID)
			if err != nil {
				return nil, fmt.Errorf("#2 CreateUser: invalid permID %s: %w", permID, err)
			}

			userPermRequests[i] = &generated.CreateUserPermRequest{
				UserPerm: &generated.UserPerm{
					UserId: fmt.Sprintf("%d", user.ID),
					PermId: parsedUUID[:],
				},
			}
		}
		_, err := s.perClients.UserPerm.BatchCreate(ctx, &generated.BatchCreateUserPermsRequest{
			Requests: userPermRequests,
		})
		if err != nil {
			return nil, fmt.Errorf("#3 CreateUser: failed to create user permissions: %w", err)
		}
	}

	if len(input.RoleIDs) > 0 {
		userRoleRequests := make([]*generated.CreateUserRoleRequest, len(input.RoleIDs))
		for i, roleID := range input.RoleIDs {
			parsedUUID, err := uuid.Parse(roleID)
			if err != nil {
				return nil, fmt.Errorf("#4 CreateUser: invalid roleID %s: %w", roleID, err)
			}
			userRoleRequests[i] = &generated.CreateUserRoleRequest{
				UserRole: &generated.UserRole{
					UserId: fmt.Sprintf("%d", user.ID),
					RoleId: parsedUUID[:],
				},
			}
		}
		_, err := s.perClients.UserRole.BatchCreate(ctx, &generated.BatchCreateUserRolesRequest{
			Requests: userRoleRequests,
		})
		if err != nil {
			return nil, fmt.Errorf("#5 CreateUser: failed to create user roles: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) UpdateUserPerms(ctx context.Context, userID int, permIDs []string) error {
	if len(permIDs) == 0 {
		return fmt.Errorf("#1 UpdateUserPerms: permIDs is empty")
	}

	req := &generated.UpdateUserPermsRequest{
		UserId: fmt.Sprintf("%d", userID),
	}
	req.PermIds = append(req.PermIds, permIDs...)
	_, err := s.perClients.PermExt.UpdateUserPerms(ctx, req)
	if err != nil {
		return fmt.Errorf("#2 UpdateUserPerms: failed to update user permissions: %w", err)
	}
	return nil
}

func (s *UserService) UpdateUserRoles(ctx context.Context, userID int, roleIDs []string) error {
	req := &generated.UpdateUserRolesRequest{
		UserId: fmt.Sprintf("%d", userID),
	}
	req.RoleIds = append(req.RoleIds, roleIDs...)
	_, err := s.perClients.PermExt.UpdateUserRoles(ctx, req)
	if err != nil {
		return fmt.Errorf("#1 UpdateUserRoles: failed to update user roles: %w", err)
	}
	return nil
}

func (s *UserService) UpdateUserByID(ctx context.Context, tx *ent.Tx, userID int, input *dto.UpdateUserDTO) (*ent.User, error) {
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
		return nil, fmt.Errorf("#1 UpdateUserByID: failed to update user: %w", err)
	}

	if err := s.UpdateUserPerms(ctx, userID, input.PermIDs); err != nil {
		return nil, fmt.Errorf("#2 UpdateUserByID: failed to update user perms: %w", err)
	}

	if err := s.UpdateUserRoles(ctx, userID, input.RoleIDs); err != nil {
		return nil, fmt.Errorf("#3 UpdateUserByID: failed to update user roles: %w", err)
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

	// Call grpc to permission service
	if s.perClients.PermExt != nil {
		userIDStr := fmt.Sprintf("%d", id)
		_, err := s.perClients.PermExt.DeleteUserPermsByUserID(ctx, &generated.DeleteUserPermsByUserIDRequest{
			UserId: userIDStr,
		})
		if err != nil {
			return fmt.Errorf("#1 DeleteUserByID: failed to delete user permissions: %w", err)
		}
		_, err = s.perClients.PermExt.DeleteUserRolesByUserID(ctx, &generated.DeleteUserRolesByUserIDRequest{
			UserId: userIDStr,
		})
		if err != nil {
			return fmt.Errorf("#2 DeleteUserByID: failed to delete user roles: %w", err)
		}
	}

	return nil
}
