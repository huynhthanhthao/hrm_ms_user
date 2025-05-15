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
	"github.com/huynhthanhthao/hrm_user_service/internal/helper"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func NewUserService(client *ent.Client, hrClients *HRServiceClients) (*UserService, error) {
	if client == nil || hrClients == nil || hrClients.Company == nil || hrClients.Branch == nil {
		return nil, fmt.Errorf("client or hrClients cannot be nil")
	}
	return &UserService{
		client:    client,
		hrClients: hrClients,
	}, nil
}

func (s *UserService) Register(ctx context.Context, c *gin.Context, input dto.RegisterInput) {
	// Call gRPC to validate company_id
	resp, err := s.hrClients.Company.Get(ctx, &hrpb.GetCompanyRequest{
		Id: []byte(input.CompanyId),
	})

	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	if resp == nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)

	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	// Bắt đầu transaction
	tx, err := s.client.Tx(ctx)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
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
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
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
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	if err := tx.Commit(); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	usr.Edges.Account = acc

	c.JSON(http.StatusOK, gin.H{
		"user": usr,
	})
}

func (s *UserService) Login(ctx context.Context, c *gin.Context, input dto.LoginInput) {
	// Tìm tài khoản theo tên đăng nhập
	acc, err := s.client.Account.
		Query().
		Where(account.UsernameEQ(input.Username)).
		Only(ctx)

	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(acc.Password), []byte(input.Password))
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	usr, err := acc.QueryUser().Only(ctx)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	branch, err := s.hrClients.HrExt.GetBranchByUserId(ctx, &hrpb.GetBranchByUserIdRequest{
		UserId: usr.ID.String(),
	})
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	branchID, err := uuid.FromBytes(branch.Id)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	companyID, err := uuid.FromBytes(branch.CompanyId)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	accessDur, _ := time.ParseDuration(os.Getenv("JWT_ACCESS_TOKEN_DURATION"))
	refreshDur, _ := time.ParseDuration(os.Getenv("JWT_REFRESH_TOKEN_DURATION"))

	accessToken, err := GenerateToken(acc.ID.String(), usr.ID.String(), branchID.String(), companyID.String(), accessDur)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	refreshToken, err := GenerateToken(acc.ID.String(), usr.ID.String(), branchID.String(), companyID.String(), refreshDur)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	usr.Edges.Account = acc

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          usr,
	})
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
