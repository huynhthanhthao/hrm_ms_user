package service

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/huynhthanhthao/hrm_user_service/ent"
	"github.com/huynhthanhthao/hrm_user_service/ent/account"
	"github.com/huynhthanhthao/hrm_user_service/ent/user"
	hrpb "github.com/huynhthanhthao/hrm_user_service/generated"
	"github.com/huynhthanhthao/hrm_user_service/internal/dto"
	"google.golang.org/grpc"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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

func NewUserService(client *ent.Client, hrClients *HRServiceClients) (*UserService, error) {
	if client == nil || hrClients == nil || hrClients.Company == nil || hrClients.Branch == nil {
		return nil, fmt.Errorf("client or hrClients cannot be nil")
	}
	return &UserService{
		client:    client,
		hrClients: hrClients,
	}, nil
}

func (s *UserService) Register(ctx context.Context, c *gin.Context, input dto.RegisterInput) (*ent.User, error) {
	// Call gRPC to validate company_id
	resp, err := s.hrClients.Company.Get(ctx, &hrpb.GetCompanyRequest{
		Id: []byte(input.CompanyId),
	})

	if err != nil {
		return nil, fmt.Errorf("HTTP %d: Lỗi validate ID công ty: %v",
			http.StatusInternalServerError, err)
	}

	if resp == nil {
		return nil, fmt.Errorf("HTTP %d: ID công ty không tồn tại", http.StatusNotFound)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)

	if err != nil {
		return nil, fmt.Errorf("HTTP %d: Lỗi mã hóa mật khẩu: %v", http.StatusConflict, err)
	}

	// Bắt đầu transaction
	tx, err := s.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("HTTP %d: Lỗi bắt đầu transaction: %v", http.StatusInternalServerError, err)
	}

	// Tạo người dùng mới
	usr, err := tx.User.
		Create().
		SetFirstName(input.FirstName).
		SetLastName(input.LastName).
		SetEmail(input.Email).
		SetPhone(input.Phone).
		SetWardCode(input.WardCode).
		SetAddress(input.Address).
		SetGender(user.Gender(input.Gender)).
		SetCompanyID(input.CompanyId).
		Save(ctx)

	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("HTTP %d: Lỗi tạo người dùng: %v", http.StatusBadRequest, err)
	}

	// Tạo tài khoản cho người dùng
	acc, err := tx.Account.
		Create().
		SetUsername(input.Username).
		SetPassword(string(hashedPassword)).
		SetStatus(account.StatusActive).
		SetUser(usr).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("HTTP %d: Lỗi tạo tài khoản: %v", http.StatusBadRequest, err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("HTTP %d: Lỗi commit giao dịch: %v", http.StatusInternalServerError, err)
	}

	usr.Edges.Account = acc

	return usr, nil
}

func (s *UserService) Login(ctx context.Context, c *gin.Context, input dto.LoginInput) (*dto.LoginResponse, error) {
	// Tìm tài khoản theo tên đăng nhập
	acc, err := s.client.Account.
		Query().
		Where(account.UsernameEQ(input.Username)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("HTTP %d: Tài khoản không tồn tại: %v", http.StatusNotFound, err)
	}

	// Kiểm tra mật khẩu
	err = bcrypt.CompareHashAndPassword([]byte(acc.Password), []byte(input.Password))
	if err != nil {
		return nil, fmt.Errorf("HTTP %d: Mật khẩu không đúng: %v", http.StatusUnauthorized, err)
	}

	// Lấy thông tin người dùng
	usr, err := acc.QueryUser().Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("HTTP %d: Không thể lấy thông tin người dùng: %v", http.StatusInternalServerError, err)
	}

	branch, err := s.hrClients.HrExt.GetBranchByUserId(ctx, &hrpb.GetBranchByUserIdRequest{
		UserId: usr.ID.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("HTTP %d: Không tìm thấy chi nhánh: %v", http.StatusNotFound, err)
	}

	branchID, err := uuid.FromBytes(branch.Id)
	if err != nil {
		return nil, fmt.Errorf("HTTP %d: Lỗi branch ID: %v", http.StatusInternalServerError, err)
	}

	companyID, err := uuid.FromBytes(branch.CompanyId)
	if err != nil {
		return nil, fmt.Errorf("HTTP %d: Lỗi company ID: %v", http.StatusInternalServerError, err)
	}

	accessDur, _ := time.ParseDuration(os.Getenv("JWT_ACCESS_TOKEN_DURATION"))
	refreshDur, _ := time.ParseDuration(os.Getenv("JWT_REFRESH_TOKEN_DURATION"))

	accessToken, err := GenerateToken(acc.ID.String(), usr.ID.String(), branchID.String(), companyID.String(), accessDur)
	if err != nil {
		return nil, fmt.Errorf("HTTP %d: Lỗi tạo access token: %v", http.StatusInternalServerError, err)
	}

	refreshToken, err := GenerateToken(acc.ID.String(), usr.ID.String(), branchID.String(), companyID.String(), refreshDur)
	if err != nil {
		return nil, fmt.Errorf("HTTP %d: Lỗi tạo refresh token: %v", http.StatusInternalServerError, err)
	}

	if err != nil {
		return nil, fmt.Errorf("HTTP %d: Lỗi tạo refresh token: %v", http.StatusInternalServerError, err)
	}

	usr.Edges.Account = acc

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         usr,
	}, nil
}

func (s *UserService) DecodeToken(token string) (*ent.User, error) {
	// Parse the token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !parsedToken.Valid {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	// Extract claims
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	accountIDStr, ok := claims["account_id"].(string)
	if !ok {
		return nil, fmt.Errorf("account_id not found in token claims")
	}

	// Convert accountID to uuid.UUID
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid account_id format: %v", err)
	}

	// Query the user associated with the account ID
	ctx := context.Background()
	acc, err := s.client.Account.Query().Where(account.IDEQ(accountID)).Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("account not found: %v", err)
	}

	usr, err := acc.QueryUser().Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("user not found: %v", err)
	}

	usr.Edges.Account = acc

	return usr, nil
}

func GenerateToken(accountID, userID, branchID, companyID string, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"account_id": accountID,
		"user_id":    userID,
		"branch_id":  branchID,
		"company_id": companyID,
		"exp":        time.Now().Add(duration).Unix(),
		"iss":        os.Getenv("ISS_KEY"),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", fmt.Errorf("JWT_SECRET is not set")
	}
	return token.SignedString([]byte(secret))
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

func (s *UserService) CreateUser(ctx context.Context, input dto.CreateUserInput) (*ent.User, error) {
	// Start a transaction
	tx, err := s.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}

	// Create a new user
	user, err := tx.User.
		Create().
		SetFirstName(input.FirstName).
		SetLastName(input.LastName).
		SetGender(user.Gender(input.Gender)).
		SetEmail(input.Email).
		SetPhone(input.Phone).
		SetWardCode(input.WardCode).
		SetAddress(input.Address).
		SetCompanyID(input.CompanyID).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return user, nil
}
