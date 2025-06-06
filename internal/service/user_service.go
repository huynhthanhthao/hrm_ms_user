package service

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/google/uuid"
	"github.com/huynhthanhthao/hrm_user_service/ent"
	"github.com/huynhthanhthao/hrm_user_service/ent/account"
	user "github.com/huynhthanhthao/hrm_user_service/ent/user"
	"github.com/huynhthanhthao/hrm_user_service/internal/dto"
	userPb "github.com/huynhthanhthao/hrm_user_service/proto/user"
	permPb "github.com/longgggwwww/hrm-ms-permission/ent/proto/entpb"
)

type UserService struct {
	client     *ent.Client
	hrClients  *HRServiceClients
	perClients *PermissionServiceClients
}

type UserResponse struct {
	User        *ent.User
	Permissions []*permPb.GetUserPermsResponse
	Roles       []*permPb.GetUserRolesResponse
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

func (s *UserService) GetUserById(ctx context.Context, id int) (*ent.User, error) {
	if id <= 0 {
		return nil, errors.New("#1 GetUserById: invalid user ID")
	}

	if s.perClients.PermExt == nil {
		return nil, errors.New("#2 GetUserById: permission client is not initialized")
	}

	// Retrieve user by ID
	user, err := s.client.User.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("#3 GetUserById: failed to retrieve user: %w", err)
	}

	return user, nil
}

func (s *UserService) GetUserPermsByUserId(ctx context.Context, user_id string) (*permPb.GetUserPermsResponse, error) {
	userPerms, err := s.perClients.PermExt.GetUserPerms(ctx, &permPb.GetUserPermsRequest{
		UserId: user_id,
	})

	if err != nil {
		return nil, fmt.Errorf("#1 GetUserPermsByUserId: failed to get user perms: %w", err)
	}

	return userPerms, nil
}

func (s *UserService) GetUserRolesByUserId(ctx context.Context, user_id string) (*permPb.GetUserRolesResponse, error) {
	userRoles, err := s.perClients.PermExt.GetUserRoles(ctx, &permPb.GetUserRolesRequest{UserId: user_id})

	if err != nil {
		return nil, fmt.Errorf("#1 GetUserRolesByUserId: failed to get user roles: %w", err)
	}

	return userRoles, nil
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

func (s *UserService) CreateUser(ctx context.Context, input *userPb.CreateUserRequest) (*ent.User, error) {
	tx, err := s.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	userCreate := tx.User.Create().
		SetFirstName(input.FirstName).
		SetLastName(input.LastName).
		SetGender(user.Gender(input.Gender)).
		SetPhone(input.Phone)

	if input.Email != nil {
		userCreate = userCreate.SetEmail(input.Email.Value)
	}
	if input.WardCode != nil {
		userCreate = userCreate.SetWardCode(input.WardCode.Value)
	}
	if input.Address != nil {
		userCreate = userCreate.SetAddress(input.Address.Value)
	}
	if input.Avatar != nil {
		userCreate = userCreate.SetAvatar(input.Avatar.Value)
	}

	user, err := userCreate.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("#1 CreateUser: failed when create user: %w", err)
	}

	if input.Account == nil {
		return nil, fmt.Errorf("#2 CreateUser: account info is required")
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(input.Account.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("#3 CreateUser: failed to hash password: %w", err)
	}
	_, err = tx.Account.Create().
		SetUsername(input.Account.Username).
		SetPassword(string(hashedPwd)).
		SetUser(user).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("#4 CreateUser: failed to create account: %w", err)
	}

	// Call grpc to permission service here
	if len(input.PermIds) > 0 {
		userPermRequests := make([]*permPb.CreateUserPermRequest, len(input.PermIds))
		for i, permID := range input.PermIds {
			parsedUUID, err := uuid.Parse(permID)
			if err != nil {
				return nil, fmt.Errorf("#5 CreateUser: invalid permID %s: %w", permID, err)
			}

			userPermRequests[i] = &permPb.CreateUserPermRequest{
				UserPerm: &permPb.UserPerm{
					UserId:    fmt.Sprintf("%d", user.ID),
					PermId:    parsedUUID[:],
					CreatedAt: timestamppb.Now(),
				},
			}
		}
		_, err := s.perClients.UserPerm.BatchCreate(ctx, &permPb.BatchCreateUserPermsRequest{
			Requests: userPermRequests,
		})
		if err != nil {
			return nil, fmt.Errorf("#6 CreateUser: failed to create user permissions: %w", err)
		}
	}

	if len(input.RoleIds) > 0 {
		userRoleRequests := make([]*permPb.CreateUserRoleRequest, len(input.RoleIds))
		for i, roleID := range input.RoleIds {
			parsedUUID, err := uuid.Parse(roleID)
			if err != nil {
				return nil, fmt.Errorf("#7 CreateUser: invalid roleID %s: %w", roleID, err)
			}
			userRoleRequests[i] = &permPb.CreateUserRoleRequest{
				UserRole: &permPb.UserRole{
					UserId:    fmt.Sprintf("%d", user.ID),
					RoleId:    parsedUUID[:],
					CreatedAt: timestamppb.Now(),
				},
			}
		}
		_, err := s.perClients.UserRole.BatchCreate(ctx, &permPb.BatchCreateUserRolesRequest{
			Requests: userRoleRequests,
		})
		if err != nil {
			return nil, fmt.Errorf("#8 CreateUser: failed to create user roles: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) UpdateUserPerms(ctx context.Context, userID int, permIDs []string) error {
	req := &permPb.UpdateUserPermsRequest{
		UserId: fmt.Sprintf("%d", userID),
	}
	req.PermIds = append(req.PermIds, permIDs...)
	_, err := s.perClients.PermExt.UpdateUserPerms(ctx, req)
	if err != nil {
		return fmt.Errorf("#1 UpdateUserPerms: failed to update user permissions: %w", err)
	}
	return nil
}

func (s *UserService) UpdateUserRoles(ctx context.Context, userID int, roleIDs []string) error {
	req := &permPb.UpdateUserRolesRequest{
		UserId: fmt.Sprintf("%d", userID),
	}
	req.RoleIds = append(req.RoleIds, roleIDs...)
	_, err := s.perClients.PermExt.UpdateUserRoles(ctx, req)
	if err != nil {
		return fmt.Errorf("#1 UpdateUserRoles: failed to update user roles: %w", err)
	}
	return nil
}

func (s *UserService) UpdateUserByID(ctx context.Context, tx *ent.Tx, userID int, input *userPb.UpdateUserRequest) (*ent.User, error) {
	userUpdate := tx.User.UpdateOneID(userID)

	if input.FirstName != "" {
		userUpdate = userUpdate.SetFirstName(input.FirstName)
	}

	if input.LastName != "" {
		userUpdate = userUpdate.SetLastName(input.LastName)
	}

	if input.Gender != "" {
		userUpdate = userUpdate.SetGender(user.Gender(input.Gender))
	}

	// Nullable fields: email, ward_code, address, avatar
	if input.Email != nil {
		userUpdate = userUpdate.SetEmail(input.Email.Value)
	}
	if input.WardCode != nil {
		userUpdate = userUpdate.SetWardCode(input.WardCode.Value)
	}
	if input.Address != nil {
		userUpdate = userUpdate.SetAddress(input.Address.Value)
	}
	if input.Avatar != nil {
		userUpdate = userUpdate.SetAvatar(input.Avatar.Value)
	}

	// Phone is not nullable, only update if not empty string
	if input.Phone != "" {
		userUpdate = userUpdate.SetPhone(input.Phone)
	}

	userCreated, err := userUpdate.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("#1 UpdateUserByID: failed to update user: %w", err)
	}

	// Tìm account theo userID
	acc, err := tx.Account.Query().Where(account.HasUserWith(user.ID(userID))).Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("#2 UpdateUserByID: account not found for userID %d", userID)
	}

	accountUpdate := tx.Account.UpdateOneID(acc.ID)
	if input.Account != nil {
		if input.Account.Status != "" {
			accountUpdate = accountUpdate.SetStatus(account.Status(input.Account.Status))
		}
		if input.Account.Password != "" {
			hashedPwd, err := bcrypt.GenerateFromPassword([]byte(input.Account.Password), bcrypt.DefaultCost)
			if err != nil {
				return nil, fmt.Errorf("#3 UpdateUserByID: failed to hash password: %w", err)
			}
			accountUpdate = accountUpdate.SetPassword(string(hashedPwd))
		}
		if input.Account.Status != "" || input.Account.Password != "" {
			_, err = accountUpdate.Save(ctx)
			if err != nil {
				return nil, fmt.Errorf("#4 UpdateUserByID: failed to update account: %w", err)
			}
		}
	}

	if input.PermIds != nil {
		if err := s.UpdateUserPerms(ctx, userID, input.PermIds); err != nil {
			return nil, fmt.Errorf("#5 UpdateUserByID: failed to update user perms: %w", err)
		}
	}

	if input.RoleIds != nil {
		if err := s.UpdateUserRoles(ctx, userID, input.RoleIds); err != nil {
			return nil, fmt.Errorf("#6 UpdateUserByID: failed to update user roles: %w", err)
		}
	}

	return userCreated, nil
}

func (s *UserService) DeleteUserByID(ctx context.Context, id int) error {
	tx, err := s.client.Tx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Tìm account liên kết với user (nếu có)
	acc, err := tx.Account.Query().Where(account.HasUserWith(user.ID(id))).Only(ctx)
	if err == nil {
		// Xóa account trước nếu tìm thấy
		if err := tx.Account.DeleteOneID(acc.ID).Exec(ctx); err != nil {
			return fmt.Errorf("#1 DeleteUserByID: failed to delete account: %w", err)
		}
	}
	// Nếu không tìm thấy account thì vẫn tiếp tục xóa user

	if err := tx.User.DeleteOneID(id).Exec(ctx); err != nil {
		return fmt.Errorf("#2 DeleteUserByID: failed to delete user: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	// Call grpc to permission service
	if s.perClients.PermExt != nil {
		userIDStr := fmt.Sprintf("%d", id)
		_, err := s.perClients.PermExt.DeleteUserPermsByUserID(ctx, &permPb.DeleteUserPermsByUserIDRequest{
			UserId: userIDStr,
		})
		if err != nil {
			return fmt.Errorf("#3 DeleteUserByID: failed to delete user permissions: %w", err)
		}
		_, err = s.perClients.PermExt.DeleteUserRolesByUserID(ctx, &permPb.DeleteUserRolesByUserIDRequest{
			UserId: userIDStr,
		})
		if err != nil {
			return fmt.Errorf("#4 DeleteUserByID: failed to delete user roles: %w", err)
		}
	}

	return nil
}
