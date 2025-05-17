package service

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/huynhthanhthao/hrm_user_service/ent"
	"github.com/huynhthanhthao/hrm_user_service/ent/account"
	"github.com/huynhthanhthao/hrm_user_service/ent/user"
	grpcClient "github.com/huynhthanhthao/hrm_user_service/generated"
	"github.com/huynhthanhthao/hrm_user_service/internal/dto"
	"github.com/huynhthanhthao/hrm_user_service/internal/helper"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	client    *ent.Client
	hrClients *HRServiceClients
	perClients *PermissionServiceClients
}

func NewAuthService(
	client *ent.Client, 
	hrClients *HRServiceClients, 
	perClients *PermissionServiceClients,
) (*AuthService, error) {
	return &AuthService{
		client: client,
		hrClients:  hrClients,
		perClients: perClients,
	}, nil
}

func (s *AuthService) Register(ctx context.Context, c *gin.Context, input dto.RegisterInput) {
	// companyUUID, err := uuid.Parse(input.CompanyId)
	// if err != nil {
	// 	helper.RespondWithError(c, http.StatusBadRequest, fmt.Errorf("invalid company ID: %w", err))
	// 	return
	// }

	// resp, err := s.hrClients.Company.Get(ctx, &grpcClient.GetCompanyRequest{
	// 	Id: companyUUID,
	// })

	// if err != nil {
	// 	helper.RespondWithError(c, http.StatusBadRequest, err)
	// 	return
	// }

	// if resp == nil {
	// 	helper.RespondWithError(c, http.StatusBadRequest, err)
	// 	return
	// }

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
		SetAvatar(input.Avatar).
		SetPhone(input.Phone).
		SetWardCode(input.WardCode).
		SetAddress(input.Address).
		SetGender(user.Gender(input.Gender)).
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

func (s *AuthService) Login(ctx context.Context, c *gin.Context, input dto.LoginInput) {
	// Tìm tài khoản theo tên đăng nhập
	acc, err := s.client.Account.
		Query().
		Where(account.UsernameEQ(input.Username)).
		Only(ctx)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	// So sánh mật khẩu
	if err := bcrypt.CompareHashAndPassword([]byte(acc.Password), []byte(input.Password)); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	// Lấy user từ account
	usr, err := acc.QueryUser().Only(ctx)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	// Gọi gRPC lấy branch theo user ID (int -> string)
	branch, err := s.hrClients.HrExt.GetBranchByUserId(ctx, &grpcClient.GetBranchByUserIdRequest{
		UserId: strconv.Itoa(usr.ID),
	})
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	branchID := string(branch.Id)

	// Parse duration
	accessDur, _ := time.ParseDuration(os.Getenv("JWT_ACCESS_TOKEN_DURATION"))
	refreshDur, _ := time.ParseDuration(os.Getenv("JWT_REFRESH_TOKEN_DURATION"))

	// Tạo token
	accessToken, err := GenerateToken(strconv.Itoa(acc.ID), strconv.Itoa(usr.ID), branchID, accessDur)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	refreshToken, err := GenerateToken(strconv.Itoa(acc.ID), strconv.Itoa(usr.ID), branchID, refreshDur)
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

func (s *AuthService) DecodeToken(ctx context.Context, token string, c *gin.Context) {
	// Parse the token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !parsedToken.Valid {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	// Extract claims
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		helper.RespondWithError(c, http.StatusBadRequest, fmt.Errorf("invalid token claims"))
		return
	}

	accountIDStr, ok := claims["account_id"].(string)
	if !ok {
		helper.RespondWithError(c, http.StatusBadRequest, fmt.Errorf("account_id not found in token claims"))
		return
	}

	// Convert accountID string to int
	accountID, err := strconv.Atoi(accountIDStr)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, fmt.Errorf("invalid account_id format in token: %v", err))
		return
	}

	// Query the account by ID
	acc, err := s.client.Account.Query().Where(account.IDEQ(accountID)).Only(ctx)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	usr, err := acc.QueryUser().Only(ctx)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	usr.Edges.Account = acc

	// (Nếu có phần lấy vị trí bằng grpc, thêm xử lý ở đây)

	c.JSON(http.StatusOK, gin.H{
		"user": usr,
	})
}

func GenerateToken(accountID, userID, branchID string, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"account_id": accountID,
		"user_id":    userID,
		"branch_id":  branchID,
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
