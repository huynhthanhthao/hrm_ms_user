package service

import (
	"context"
	"fmt"

	"github.com/huynhthanhthao/hrm_user_service/ent"
	"github.com/huynhthanhthao/hrm_user_service/ent/user"
	hrpb "github.com/huynhthanhthao/hrm_user_service/generated"
	"github.com/huynhthanhthao/hrm_user_service/internal/dto"
	"google.golang.org/grpc"

	"github.com/google/uuid"
)

type HRServiceClients struct {
	Company hrpb.CompanyServiceClient
	Branch  hrpb.BranchServiceClient
	HrExt   hrpb.ExtServiceClient
	Conn    *grpc.ClientConn
}
type UserService struct {
	client    *ent.Client
	hrClients *HRServiceClients
}

func (c *HRServiceClients) Close() {
	if c.Conn != nil {
		c.Conn.Close()
	}
}

func (s *UserService) GetAllUsers(ctx context.Context) ([]*ent.User, error) {
	users, err := s.client.User.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve users: %w", err)
	}
	return users, nil
}

func (s *UserService) GetUser(ctx context.Context, id string) (*ent.User, error) {
	// Convert the string ID to uuid.UUID
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	// Use the converted UUID
	user, err := s.client.User.Get(ctx, userID)
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

	// Convert string IDs to uuid.UUID
	uuidIDs := make([]uuid.UUID, len(params.IDs))
	for i, id := range params.IDs {
		uuidID, err := uuid.Parse(id)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid user ID format for ID %s: %w", id, err)
		}
		uuidIDs[i] = uuidID
	}

	// Query users by IDs with pagination
	users, err := s.client.User.Query().
		Where(user.IDIn(uuidIDs...)).
		Offset(offset).
		Limit(params.PageSize).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve users: %w", err)
	}

	// Get total count of users matching the IDs
	totalCount, err := s.client.User.Query().Where(user.IDIn(uuidIDs...)).Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	return users, totalCount, nil
}
